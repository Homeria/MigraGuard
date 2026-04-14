package types

import (
	"context"
)

// PostgresClient는 대상 PostgreSQL 데이터베이스와 통신하기 위한 인터페이스입니다.
type PostgresClient interface {
	// FetchCurrentWorkloadSnapshot은 현재 시점의 누적 쿼리 통계 데이터를 가져옵니다.
	FetchCurrentWorkloadSnapshot(ctx context.Context) ([]WorkloadSnapshot, error)
	// FetchTableDynamicMetrics는 분석에 필요한 테이블별 동적 지표(크기, 연결 수 등)를 수집합니다.
	FetchTableDynamicMetrics(ctx context.Context, tableName string) (*TableDynamicMetrics, error)
	// CheckTableSchemaPresence는 DDL 대상 테이블/컬럼의 실제 존재 여부를 검증합니다.
	CheckTableSchemaPresence(ctx context.Context, tableName string, columns []string) error

	Close()
}

// SQLiteClient는 로컬 지표 저장소인 SQLite와 통신하기 위한 인터페이스입니다.
type SQLiteClient interface {
	// GetRecentTPSByDelta는 델타 스냅샷 기반의 실시간 TPS를 조회합니다.
	GetRecentTPSByDelta(tableName string) (float64, error)
	// GetTableBaselineStatistics는 과거 1시간 평균 및 24시간 피크 통계를 가져옵니다.
	GetTableBaselineStatistics(tableName string) (*BaselineStats, error)
	// IdentifySafestDeploymentWindow는 가장 부하가 낮은 안전 배포 시간대를 추천합니다.
	IdentifySafestDeploymentWindow() (string, float64, error)
	// GetLatestTableMetrics는 특정 테이블의 가장 최근 수집된 지표를 가져옵니다.
	GetLatestTableMetrics(tableName string) (*TableDynamicMetrics, error)
	// GetTopHeavyQueries는 DB 부하를 많이 점유하고 있는 상위 쿼리들을 조회합니다.
	GetTopHeavyQueries(limit int) ([]TopQueryInfo, error)

	// RecordDeltaSnapshots는 계산된 델타 메트릭을 저장합니다.
	RecordDeltaSnapshots(snapshots []WorkloadSnapshot) error
	// RecordTableDynamicMetrics는 테이블별 동적 지표를 저장합니다.
	RecordTableDynamicMetrics(m *TableDynamicMetrics) error
	// FetchLastOriginalSnapshots는 마지막으로 기록된 누적 통계 원본을 조회합니다.
	FetchLastOriginalSnapshots() (map[int64]WorkloadSnapshot, error)
	// SynchronizeOriginalSnapshots는 누적 통계 원본을 최신으로 업데이트합니다.
	SynchronizeOriginalSnapshots(snapshots []WorkloadSnapshot) error

	// MaintenancePurgeData는 보관 주기가 지난 데이터를 삭제합니다.
	MaintenancePurgeData(retentionDays int) error
	Close() error
}
