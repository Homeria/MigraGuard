package db

import (
	"time"
)

// WorkloadSnapshot은 특정 시점의 데이터베이스 쿼리 실행 통계를 나타냅니다.
// pg_stat_statements에서 수집된 데이터를 기반으로 합니다.
type WorkloadSnapshot struct {
	Timestamp      time.Time `json:"timestamp"`        // 수집 시점
	QueryID        int64     `json:"query_id"`         // 쿼리 고유 식별자
	Query          string    `json:"query"`            // 정규화된 쿼리 텍스트
	Calls          int64     `json:"calls"`            // 실행 횟수 (Delta 또는 Cumulative)
	TotalTime      float64   `json:"total_time"`       // 총 실행 시간 (ms)
	Rows           int64     `json:"rows"`             // 처리된 총 행 수
	SharedBlksHit  int64     `json:"shared_blks_hit"`  // 공유 버퍼 히트 수
	SharedBlksRead int64     `json:"shared_blks_read"` // 공유 버퍼 읽기 수
}

// TableDynamicMetrics는 리스크 분석 엔진에서 사용하는 테이블별 동적 지표입니다.
// 실시간 상태와 최근 트래픽 경향을 모두 포함합니다.
type TableDynamicMetrics struct {
	TableName         string  `json:"table_name"`         // 대상 테이블명
	TableSize         int64   `json:"table_size"`         // 테이블 전체 용량 (Bytes)
	ReplicationLag    float64 `json:"replication_lag"`    // 복제 지연 시간 (Seconds)
	ActiveConnections int     `json:"active_connections"` // 해당 테이블 관련 활성 세션 수
	P99Time           float64 `json:"p99_time"`           // 99백분위수 실행 시간 (ms)
	TPS               float64 `json:"tps"`                // 최근 초당 트랜잭션 수
}

// BaselineStats는 분석 대상 테이블의 과거 트래픽 통계 정보를 담고 있습니다.
// 현재 트래픽이 평소(평균)나 최악(피크) 대비 어느 정도인지 비교할 때 사용합니다.
type BaselineStats struct {
	AvgTPS_1h   float64 // 최근 1시간 평균 TPS
	PeakTPS_24h float64 // 최근 24시간 최대 피크 TPS
}
