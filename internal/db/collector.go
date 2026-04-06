package db

import (
	"context"
	"log"
	"time"
)

// Collector는 PostgreSQL에서 워크로드 지표를 주기적으로 수집하여 SQLite에 저장하는 프로세스를 관리합니다.
type Collector struct {
	pg            PostgresClient
	sqlite        SQLiteClient
	interval      time.Duration
	retentionDays int
	targetTables  []string
	stopChan      chan struct{}
}

// NewCollector는 새로운 Collector 인스턴스를 생성합니다.
func NewCollector(pg PostgresClient, sqlite SQLiteClient, interval time.Duration) *Collector {
	return &Collector{
		pg:            pg,
		sqlite:        sqlite,
		interval:      interval,
		retentionDays: 7,          // 기본 데이터 보존 기간: 7일
		targetTables:  []string{}, // 모니터링 대상 테이블 목록 초기화
		stopChan:      make(chan struct{}),
	}
}

// SetRetentionDays는 수집된 데이터의 보존 기간을 설정합니다.
func (c *Collector) SetRetentionDays(days int) {
	c.retentionDays = days
}

// AddTargetTable은 실시간 지표 수집을 위한 대상 테이블을 추가합니다.
func (c *Collector) AddTargetTable(tableName string) {
	for _, t := range c.targetTables {
		if t == tableName {
			return
		}
	}
	c.targetTables = append(c.targetTables, tableName)
}

// Start는 별도의 고루틴에서 백그라운드 수집 루프를 시작합니다.
func (c *Collector) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)

	go func() {
		log.Printf("백그라운드 수집기 시작 (수집 간격: %v)", c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := c.collect(ctx); err != nil {
					log.Printf("수집 중 오류 발생: %v", err)
				}
			case <-c.stopChan:
				log.Println("백그라운드 수집기 중단 중...")
				return
			case <-ctx.Done():
				log.Println("컨텍스트 취소로 인한 수집기 중단")
				return
			}
		}
	}()
}

// collect는 단일 수집 사이클을 수행합니다: 누적치 조회 -> 델타 계산 -> 저장 -> 유지보수
func (c *Collector) collect(ctx context.Context) error {

	// 1. PostgreSQL에서 현재 누적 통계 데이터를 가져옵니다.
	currentSnapshots, err := c.pg.FetchCurrentWorkloadSnapshot(ctx)
	if err != nil {
		return err
	}

	if len(currentSnapshots) == 0 {
		return nil
	}

	// 2. 이전 주기의 누적 데이터를 SQLite에서 가져와 차이값(Delta) 계산을 준비합니다.
	previousSnapshots, err := c.sqlite.FetchLastOriginalSnapshots()
	if err != nil {
		log.Printf("이전 상태 로드 실패 (초기 구동 가능성): %v", err)
		previousSnapshots = make(map[int64]WorkloadSnapshot)
	}

	// 3. 델타 엔진: 현재값 - 이전값을 계산하여 해당 주기 동안의 순수 부하량을 산출합니다.
	var deltaSnapshots []WorkloadSnapshot
	for _, current := range currentSnapshots {
		delta := current 

		if prev, ok := previousSnapshots[current.QueryID]; ok {
			delta.Calls = current.Calls - prev.Calls
			delta.TotalTime = current.TotalTime - prev.TotalTime
			delta.Rows = current.Rows - prev.Rows
			delta.SharedBlksHit = current.SharedBlksHit - prev.SharedBlksHit
			delta.SharedBlksRead = current.SharedBlksRead - prev.SharedBlksRead

			// 통계가 초기화(Reset)된 경우 현재 값을 그대로 사용
			if delta.Calls < 0 {
				delta = current
			}
		}

		// 활동이 있었던 쿼리만 시계열 데이터로 기록
		if delta.Calls > 0 {
			deltaSnapshots = append(deltaSnapshots, delta)
		}
	}

	// 4. 계산된 델타 스냅샷을 SQLite에 영구 저장합니다.
	if len(deltaSnapshots) > 0 {
		if err := c.sqlite.RecordDeltaSnapshots(deltaSnapshots); err != nil {
			return err
		}
		log.Printf("%d개의 델타 워크로드 스냅샷 수집 완료", len(deltaSnapshots))
	}

	// 5. 다음 주기의 델타 계산을 위해 현재 누적 원본 데이터를 동기화합니다.
	if err := c.sqlite.SynchronizeOriginalSnapshots(currentSnapshots); err != nil {
		log.Printf("원본 데이터 동기화 실패: %v", err)
	}

	// 6. 대상 테이블들의 실시간 동적 지표(용량, 세션 등)를 수집합니다.
	for _, table := range c.targetTables {
		metrics, err := c.pg.FetchTableDynamicMetrics(ctx, table)
		if err != nil {
			log.Printf("테이블 %s 지표 수집 중 오류: %v", table, err)
			continue
		}

		// 7. 수집된 테이블 지표를 저장합니다.
		if err := c.sqlite.RecordTableDynamicMetrics(metrics); err != nil {
			log.Printf("테이블 %s 지표 저장 중 오류: %v", table, err)
		}
	}

	// 8. 데이터 보존 정책에 따라 오래된 데이터를 정리합니다.
	if err := c.sqlite.MaintenancePurgeData(c.retentionDays); err != nil {
		log.Printf("데이터 유지보수(Purge) 중 오류: %v", err)
	}

	return nil
}

// Stop은 백그라운드 수집기에 중단 신호를 보냅니다.
func (c *Collector) Stop() {
	close(c.stopChan)
}
