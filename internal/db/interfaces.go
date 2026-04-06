package db

import (
	"context"
)

// PostgresClient는 대상 PostgreSQL 데이터베이스와 통신하기 위한 인터페이스입니다.
type PostgresClient interface {
	// 수집(Collection) 관련
	FetchCurrentWorkloadSnapshot(ctx context.Context) ([]WorkloadSnapshot, error)
	FetchTableDynamicMetrics(ctx context.Context, tableName string) (*TableDynamicMetrics, error)

	// 검증(Validation) 관련
	CheckTableSchemaPresence(ctx context.Context, tableName string, columns []string) error

	Close()
}

// SQLiteClient는 로컬 지표 저장소인 SQLite와 통신하기 위한 인터페이스입니다.
type SQLiteClient interface {
	// 분석(Analysis) 전용 조회
	GetRecentTPSByDelta(tableName string) (float64, error)
	GetTableBaselineStatistics(tableName string) (*BaselineStats, error)
	IdentifySafestDeploymentWindow() (string, float64, error)
	GetLatestTableMetrics(tableName string) (*TableDynamicMetrics, error)

	// 수집(Collection) 및 저장
	RecordDeltaSnapshots(snapshots []WorkloadSnapshot) error
	RecordTableDynamicMetrics(m *TableDynamicMetrics) error
	FetchLastOriginalSnapshots() (map[int64]WorkloadSnapshot, error)
	SynchronizeOriginalSnapshots(snapshots []WorkloadSnapshot) error

	// 유지보수(Maintenance)
	MaintenancePurgeData(retentionDays int) error
	Close() error
}
