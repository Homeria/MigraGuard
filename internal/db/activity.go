package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteAdapter manages the local time-series storage for workload data.
// 워크로드 데이터를 위한 로컬 시계열 저장소를 관리합니다.
type SQLiteAdapter struct {
	db *sql.DB
}

// NewSQLiteAdapter creates a new SQLiteAdapter and initializes the database.
// 새로운 SQLiteAdapter를 생성하고 데이터베이스를 초기화합니다.
func NewSQLiteAdapter(path string) (*SQLiteAdapter, error) {
	// Open SQLite database with WAL mode for better concurrency
	// 동시성 성능 향상을 위해 WAL 모드를 사용하여 SQLite 데이터베이스를 엽니다.
	dsn := fmt.Sprintf("%s?_journal_mode=WAL", path)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	adapter := &SQLiteAdapter{db: db}
	if err := adapter.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize sqlite schema: %w", err)
	}

	return adapter, nil
}

// initSchema creates the necessary tables for storing workload data.
// 워크로드 데이터 저장을 위한 테이블과 인덱스를 생성합니다.
func (a *SQLiteAdapter) initSchema() error {
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
	CREATE INDEX IF NOT EXISTS idx_workload_query_id ON workload_snapshots(query_id);
	`
	_, err := a.db.Exec(query)
	return err
}

// SaveSnapshots persists multiple workload snapshots to the SQLite database.
// 여러 워크로드 스냅샷을 트랜잭션을 통해 SQLite 데이터베이스에 효율적으로 저장합니다.
func (a *SQLiteAdapter) SaveSnapshots(snapshots []WorkloadSnapshot) error {
	tx, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO workload_snapshots (
			timestamp, query_id, query, calls, total_time, rows_affected, shared_blks_hit, shared_blks_read
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, s := range snapshots {
		_, err := stmt.Exec(
			s.Timestamp,
			s.QueryID,
			s.Query,
			s.Calls,
			s.TotalTime,
			s.Rows,
			s.SharedBlksHit,
			s.SharedBlksRead,
		)
		if err != nil {
			return fmt.Errorf("failed to execute insert: %w", err)
		}
	}

	return tx.Commit()
}

// Close closes the SQLite database connection.
// SQLite 데이터베이스 연결을 닫습니다.
func (a *SQLiteAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}
