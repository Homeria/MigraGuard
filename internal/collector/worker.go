package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/Homeria/MigraGuard/internal/shared/types"
)

// Collector는 주기적으로 PostgreSQL의 지표를 수집하고 델타 값을 계산하여 저장소에 기록하는 객체입니다.
type Collector struct {
	pg            types.PostgresClient
	sqlite        types.SQLiteClient
	interval      time.Duration
	retentionDays int
	targetTables  []string
	stopChan      chan struct{}
}

// NewCollector는 Collector의 새로운 인스턴스를 생성합니다.
func NewCollector(pg types.PostgresClient, sqlite types.SQLiteClient, interval time.Duration) *Collector {
	return &Collector{
		pg:           pg,
		sqlite:       sqlite,
		interval:     interval,
		targetTables: []string{},
		stopChan:     make(chan struct{}),
	}
}

// SetRetentionDays는 수집된 데이터의 보관 기간(일 단위)을 설정합니다.
func (c *Collector) SetRetentionDays(days int) {
	c.retentionDays = days
}

// AddTargetTable은 집중 모니터링이 필요한 특정 테이블을 수집 대상에 추가합니다.
func (c *Collector) AddTargetTable(table string) {
	c.targetTables = append(c.targetTables, table)
}

// Start는 지표 수집 루프를 백그라운드 고루틴으로 실행합니다.
func (c *Collector) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := c.CollectOnce(ctx); err != nil {
					fmt.Printf("[%s] ❌ 수집 오류: %v\n", time.Now().Format("15:04:05"), err)
				}
			case <-c.stopChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// CollectOnce는 1회성 지표 수집 및 델타 동기화 작업을 수행합니다.
func (c *Collector) CollectOnce(ctx context.Context) error {
	timestamp := time.Now().Format("15:04:05")

	// 1. PostgreSQL에서 현재 누적 통계 조회
	current, err := c.pg.FetchCurrentWorkloadSnapshot(ctx)
	if err != nil {
		return err
	}

	// 2. SQLite에서 마지막으로 기록된 누적 통계 로드
	previous, _ := c.sqlite.FetchLastOriginalSnapshots()

	// 3. 델타(Delta) 계산
	deltas := c.computeDelta(current, previous)

	// 4. 결과 영속화
	if len(deltas) > 0 {
		_ = c.sqlite.RecordDeltaSnapshots(deltas)
	}
	_ = c.sqlite.SynchronizeOriginalSnapshots(current)

	// 5. 집중 모니터링 대상 테이블의 실시간 동적 지표 수집
	updatedTables := []string{}
	for _, table := range c.targetTables {
		metrics, err := c.pg.FetchTableDynamicMetrics(ctx, table)
		if err == nil {
			_ = c.sqlite.RecordTableDynamicMetrics(metrics)
			updatedTables = append(updatedTables, table)
		}
	}

	// 6. 지표 수집 요약 로그 출력 (사용자 요청 사항)
	fmt.Printf("[%s] 📡 지표 수집 완료: 쿼리 변화량 %d건 기록 | 대상 테이블: %v\n", 
		timestamp, len(deltas), updatedTables)

	// 7. 설정된 주기에 따른 데이터 정리(Retention)
	if c.retentionDays > 0 {
		_ = c.sqlite.MaintenancePurgeData(c.retentionDays)
	}

	return nil
}

// computeDelta는 현재 누적 통계와 이전 통계를 비교하여 순수 변화량(Delta)을 산출합니다.
func (c *Collector) computeDelta(curr []types.WorkloadSnapshot, prev map[int64]types.WorkloadSnapshot) []types.WorkloadSnapshot {
	var deltas []types.WorkloadSnapshot
	for _, s := range curr {
		p, exists := prev[s.QueryID]
		if !exists {
			deltas = append(deltas, s)
			continue
		}

		deltaCalls := s.Calls - p.Calls
		if deltaCalls < 0 {
			deltas = append(deltas, s)
			continue
		}

		if deltaCalls > 0 {
			delta := s
			delta.Calls = deltaCalls
			delta.TotalTime = s.TotalTime - p.TotalTime
			delta.Rows = s.Rows - p.Rows
			delta.SharedBlksHit = s.SharedBlksHit - p.SharedBlksHit
			delta.SharedBlksRead = s.SharedBlksRead - p.SharedBlksRead
			deltas = append(deltas, delta)
		}
	}
	return deltas
}

// Stop은 백그라운드 수집 루프를 안전하게 중단합니다.
func (c *Collector) Stop() {
	close(c.stopChan)
}
