package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// SQLiteAdapter handles all interactions with the local SQLite metrics database.
type SQLiteAdapter struct {
	db *sql.DB
}

// Repository is an alias for SQLiteAdapter for public use.
type Repository = SQLiteAdapter

// NewSQLiteAdapter connects to the SQLite file and initializes the schema.
func NewSQLiteAdapter(path string) (*SQLiteAdapter, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	adapter := &SQLiteAdapter{db: db}

	if err := adapter.InitializeSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize SQLite schema: %w", err)
	}

	return adapter, nil
}

// NewRepository creates a new repository for the client.
func NewRepository(path string) (*SQLiteAdapter, error) {
	return NewSQLiteAdapter(path)
}

// InitializeSchema defines the essential tables for MigraGuard.
func (a *SQLiteAdapter) InitializeSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS workload_snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		query_id BIGINT,
		query TEXT,
		calls BIGINT,
		total_time DOUBLE,
		rows_affected BIGINT,
		shared_blks_hit BIGINT,
		shared_blks_read BIGINT
	);
	CREATE INDEX IF NOT EXISTS idx_workload_timestamp ON workload_snapshots(timestamp);

	CREATE TABLE IF NOT EXISTS table_metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		table_name TEXT,
		table_size BIGINT,
		replication_lag DOUBLE,
		active_connections INTEGER,
		p99_time DOUBLE,
		tps DOUBLE,
		shared_blks_hit BIGINT,
		shared_blks_read BIGINT
	);
	CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON table_metrics(timestamp);

	CREATE TABLE IF NOT EXISTS original_pg_stat_statements (
		query_id BIGINT PRIMARY KEY,
		query TEXT,
		calls BIGINT,
		total_time DOUBLE,
		rows_affected BIGINT,
		shared_blks_hit BIGINT,
		shared_blks_read BIGINT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := a.db.Exec(query)
	return err
}

// Close closes the SQLite connection.
func (a *SQLiteAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// MaintenancePurgeData deletes data older than the retention period.
func (a *SQLiteAdapter) MaintenancePurgeData(retentionDays int) error {
	_, _ = a.db.Exec(`DELETE FROM workload_snapshots WHERE timestamp < datetime('now', '-' || ? || ' days')`, retentionDays)
	_, _ = a.db.Exec(`DELETE FROM table_metrics WHERE timestamp < datetime('now', '-' || ? || ' days')`, retentionDays)

	_, err := a.db.Exec("VACUUM")
	return err
}
