package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/Homeria/MigraGuard/internal/shared/types"
)

// Collector는 백그라운드에서 PostgreSQL의 운영 지표를 상시 수집하고 가공하는 도메인 객체입니다.
type Collector struct {
	pg            types.PostgresClient // 원시 지표 추출 소스
	sqlite        types.SQLiteClient   // 가공 지표 저장소
	interval      time.Duration        // 수집 실행 간격
	retentionDays int                  // 데이터 보관 기간
	targetTables  []string             // 정밀 감시 테이블 목록
	stopChan      chan struct{}        // 서비스 중단 제어 채널
}

// NewCollector는 필요한 의존성을 주입받아 수집기 인스턴스를 초기화합니다.
//
// Args:
//   - pg: Postgres 지표 추출 인터페이스
//   - sqlite: 로컬 영속성 인터페이스
//   - interval: 수집 실행 간격
//
// Returns:
//   - *Collector: 생성된 수집기 객체
func NewCollector(pg types.PostgresClient, sqlite types.SQLiteClient, interval time.Duration) *Collector {
	// 1. 객체 생성 및 기본 필드 할당
	return &Collector{
		pg:           pg,
		sqlite:       sqlite,
		interval:     interval,
		targetTables: []string{},
		stopChan:     make(chan struct{}),
	}
}

// SetRetentionDays는 지표 데이터의 물리적 보관 기간을 설정합니다.
//
// Args:
//   - days: 데이터 보관 일수
func (c *Collector) SetRetentionDays(days int) {
	// 1. 보존 기간 필드 갱신
	c.retentionDays = days
}

// AddTargetTable은 테이블 통계 수집 대상이 될 특정 테이블을 추가합니다.
//
// Args:
//   - table: 추가할 테이블명
func (c *Collector) AddTargetTable(table string) {
	// 1. 대상 목록 슬라이스에 추가
	c.targetTables = append(c.targetTables, table)
}

// Start는 지표 수집 루프를 별도의 고루틴에서 비동기적으로 실행합니다.
//
// Args:
//   - ctx: 어플리케이션 전역 컨텍스트
func (c *Collector) Start(ctx context.Context) {
	go func() {
		// 1. 설정된 주기에 따른 티커 생성
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 2. 주기별 1회 수집 작업 실행
				if err := c.CollectOnce(ctx); err != nil {
					fmt.Printf("[%s] ❌ 수집 실패: %v\n", time.Now().Format("15:04:05"), err)
				}
			case <-c.stopChan:
				// 3. 중단 채널 신호 수신 시 종료
				return
			case <-ctx.Done():
				// 4. 전역 컨텍스트 종료 시 종료
				return
			}
		}
	}()
}

// CollectOnce는 1주기 동안의 지표 수집, 델타 계산, 데이터 영속화 과정을 일괄 처리합니다.
//
// Args:
//   - ctx: 실행 컨텍스트
//
// Returns:
//   - error: 단계별 처리 중 발생한 에러
func (c *Collector) CollectOnce(ctx context.Context) error {
	timestamp := time.Now().Format("15:04:05")

	// 1. Postgres 조회: 원시 누적 통계 추출
	current, err := c.pg.FetchCurrentWorkloadSnapshot(ctx)
	if err != nil { return err }

	// 2. SQLite 조회: 직전 주기 누적치 로드
	previous, _ := c.sqlite.FetchLastOriginalSnapshots()

	// 3. 델타 계산: (현재 - 이전) 차이분 도출
	deltas := c.computeDelta(current, previous)

	// 4. 데이터 영속화: 델타 기록 및 원본 최신화
	if len(deltas) > 0 {
		_ = c.sqlite.RecordDeltaSnapshots(deltas)
	}
	_ = c.sqlite.SynchronizeOriginalSnapshots(current)

	// 5. 테이블 통계 수집: 감시 대상 테이블 상태 기록
	updatedTables := []string{}
	for _, table := range c.targetTables {
		metrics, err := c.pg.FetchTableDynamicMetrics(ctx, table)
		if err == nil {
			_ = c.sqlite.RecordTableDynamicMetrics(metrics)
			updatedTables = append(updatedTables, table)
		}
	}

	// 6. 상태 로그 출력: 수집 현황 요약
	fmt.Printf("[%s] 📡 지표 수집 완료: 쿼리 변화량 %d건 기록 | 대상 테이블: %v\n", 
		timestamp, len(deltas), updatedTables)

	// 7. 저장소 관리: 보존 기간 만료 데이터 삭제
	if c.retentionDays > 0 {
		_ = c.sqlite.MaintenancePurgeData(c.retentionDays)
	}

	return nil
}

// computeDelta는 두 누적 통계 셋을 비교하여 변화량(Delta) 데이터만 추출합니다.
//
// Args:
//   - curr: 현재 수집된 누적 스냅샷 목록
//   - prev: 이전 주기의 누적 데이터 맵
//
// Returns:
//   - []WorkloadSnapshot: 순수 변화량만 담긴 스냅샷 슬라이스
func (c *Collector) computeDelta(curr []types.WorkloadSnapshot, prev map[int64]types.WorkloadSnapshot) []types.WorkloadSnapshot {
	var deltas []types.WorkloadSnapshot
	for _, s := range curr {
		p, exists := prev[s.QueryID]
		if !exists {
			// 신규 쿼리는 현재 전체를 델타로 간주
			deltas = append(deltas, s)
			continue
		}

		// 1. 호출 횟수 차이 계산
		deltaCalls := s.Calls - p.Calls
		if deltaCalls < 0 {
			// 통계 리셋 시 현재치를 신규 기준으로 설정
			deltas = append(deltas, s)
			continue
		}

		// 2. 유의미한 변화(호출) 발생 시 델타 생성
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
	// 1. 제어 채널 폐쇄
	close(c.stopChan)
}
