package db

import (
	"context"
)

// PostgresClient defines the interface for PostgreSQL operations.
type PostgresClient interface {
	GetTableDynamicMetrics(ctx context.Context, tableName string) (*TableDynamicMetrics, error)
	ValidateSchema(ctx context.Context, tableName string, columns []string) error
	FetchWorkload(ctx context.Context) ([]WorkloadSnapshot, error)
	Close()
}

// SQLiteClient defines the interface for SQLite operations.
type SQLiteClient interface {
	GetRecentTPSDelta(tableName string) (float64, error)
	GetTableBaselineStats(tableName string) (*BaselineStats, error)
	GetSafeWindow() (string, float64, error)
	SaveSnapshots(snapshots []WorkloadSnapshot) error
	SaveTableMetrics(m *TableDynamicMetrics) error
	PurgeOldSnapshots(retentionDays int) error
	GetLastOriginalSnapshots() (map[int64]WorkloadSnapshot, error)
	UpsertOriginalSnapshots(snapshots []WorkloadSnapshot) error
	Close() error
}
