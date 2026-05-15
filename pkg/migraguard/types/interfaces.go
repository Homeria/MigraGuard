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
	GetRecentTPSByDelta(ctx context.Context, tableName string) (float64, error)
	// GetTableBaselineStatistics retrieves 1h average and 24h peak statistics.
	GetTableBaselineStatistics(ctx context.Context, tableName string) (*BaselineStats, error)
	// IdentifySafestDeploymentWindow recommends the safest window for deployment.
	IdentifySafestDeploymentWindow(ctx context.Context) (string, float64, error)
	// GetLatestTableMetrics retrieves the most recent table metrics.
	GetLatestTableMetrics(ctx context.Context, tableName string) (*TableDynamicMetrics, error)
	// GetTopHeavyQueries retrieves the top resource-consuming queries.
	GetTopHeavyQueries(ctx context.Context, limit int) ([]TopQueryInfo, error)
	// Get24HourTrafficForecast generates a 24-hour baseline profile.
	Get24HourTrafficForecast(ctx context.Context, tableName string) ([]ForecastTimeSlot, error)

	// RecordDeltaSnapshots stores calculated delta metrics.
	RecordDeltaSnapshots(ctx context.Context, snapshots []WorkloadSnapshot) error
	// RecordTableDynamicMetrics stores table-specific metrics.
	RecordTableDynamicMetrics(ctx context.Context, m *TableDynamicMetrics) error
	// FetchLastOriginalSnapshots retrieves the last recorded raw statistics.
	FetchLastOriginalSnapshots(ctx context.Context) (map[int64]WorkloadSnapshot, error)
	// SynchronizeOriginalSnapshots updates raw statistics.
	SynchronizeOriginalSnapshots(ctx context.Context, snapshots []WorkloadSnapshot) error

	// FetchAllTableMetrics retrieves all table metrics for export.
	FetchAllTableMetrics(ctx context.Context) ([]TableDynamicMetrics, error)

	// MaintenancePurgeData deletes expired data.
	MaintenancePurgeData(ctx context.Context, retentionDays int) error
	Close() error
}

// Logger is the interface for system-wide logging, allowing for different implementations.
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}
