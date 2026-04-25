package types

import (
	"context"
)

// PostgresClient is the interface for interacting with the target PostgreSQL database.
type PostgresClient interface {
	// FetchCurrentWorkloadSnapshot collects raw query statistics.
	FetchCurrentWorkloadSnapshot(ctx context.Context) ([]WorkloadSnapshot, error)
	// FetchTableDynamicMetrics collects table-specific metrics like size and active connections.
	FetchTableDynamicMetrics(ctx context.Context, tableName string) (*TableDynamicMetrics, error)
	// CheckTableSchemaPresence verifies if the target tables and columns exist.
	CheckTableSchemaPresence(ctx context.Context, tableName string, columns []string) error

	Close()
}

// SQLiteClient is the interface for interacting with the local SQLite metrics database.
type SQLiteClient interface {
	// GetRecentTPSByDelta calculates real-time TPS based on snapshots.
	GetRecentTPSByDelta(tableName string) (float64, error)
	// GetTableBaselineStatistics retrieves 1h average and 24h peak statistics.
	GetTableBaselineStatistics(tableName string) (*BaselineStats, error)
	// IdentifySafestDeploymentWindow recommends the safest window for deployment.
	IdentifySafestDeploymentWindow() (string, float64, error)
	// GetLatestTableMetrics retrieves the most recent table metrics.
	GetLatestTableMetrics(tableName string) (*TableDynamicMetrics, error)
	// GetTopHeavyQueries retrieves the top resource-consuming queries.
	GetTopHeavyQueries(limit int) ([]TopQueryInfo, error)

	// RecordDeltaSnapshots stores calculated delta metrics.
	RecordDeltaSnapshots(snapshots []WorkloadSnapshot) error
	// RecordTableDynamicMetrics stores table-specific metrics.
	RecordTableDynamicMetrics(m *TableDynamicMetrics) error
	// FetchLastOriginalSnapshots retrieves the last recorded raw statistics.
	FetchLastOriginalSnapshots() (map[int64]WorkloadSnapshot, error)
	// SynchronizeOriginalSnapshots updates raw statistics.
	SynchronizeOriginalSnapshots(snapshots []WorkloadSnapshot) error

	// MaintenancePurgeData deletes expired data.
	MaintenancePurgeData(retentionDays int) error
	Close() error
}
