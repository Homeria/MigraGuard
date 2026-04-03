package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// SQLiteAdapter manages the local time-series storage for workload data.
// 워크로드 데이터를 위한 로컬 시계열 저장소를 관리합니다.
type SQLiteAdapter struct {
	db *sql.DB
}

// NewSQLiteAdapter creates a new SQLiteAdapter and initializes the database.
// 새로운 SQLiteAdapter를 생성하고 데이터베이스를 초기화합니다.
func NewSQLiteAdapter(path string) (*SQLiteAdapter, error) {
	// Open SQLite database
	db, err := sql.Open("sqlite", path)
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

	CREATE TABLE IF NOT EXISTS table_metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		table_name TEXT,
		table_size BIGINT,
		replication_lag DOUBLE,
		active_connections INTEGER,
		p99_time DOUBLE,
		tps DOUBLE
	);
	CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON table_metrics(timestamp);
	CREATE INDEX IF NOT EXISTS idx_metrics_table_name ON table_metrics(table_name);
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

// SaveTableMetrics persists dynamic table metrics to the SQLite database.
// 테이블의 동적 지표를 SQLite 데이터베이스에 저장합니다.
func (a *SQLiteAdapter) SaveTableMetrics(m *TableDynamicMetrics) error {
	query := `
		INSERT INTO table_metrics (
			timestamp, table_name, table_size, replication_lag, active_connections, p99_time, tps
		) VALUES (CURRENT_TIMESTAMP, ?, ?, ?, ?, ?, ?)
	`
	_, err := a.db.Exec(query, 
		m.TableName, 
		m.TableSize, 
		m.ReplicationLag, 
		m.ActiveConnections, 
		m.P99Time, 
		m.TPS,
	)
	if err != nil {
		return fmt.Errorf("failed to save table metrics: %w", err)
	}
	return nil
}

// GetLatestTableMetrics retrieves the most recent metrics for a specific table.
// 특정 테이블에 대한 가장 최신의 동적 지표를 조회합니다.
func (a *SQLiteAdapter) GetLatestTableMetrics(tableName string) (*TableDynamicMetrics, error) {
	query := `
		SELECT table_name, table_size, replication_lag, active_connections, p99_time, tps
		FROM table_metrics
		WHERE table_name = ?
		ORDER BY timestamp DESC
		LIMIT 1;
	`
	var m TableDynamicMetrics
	err := a.db.QueryRow(query, tableName).Scan(
		&m.TableName,
		&m.TableSize,
		&m.ReplicationLag,
		&m.ActiveConnections,
		&m.P99Time,
		&m.TPS,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest table metrics: %w", err)
	}
	return &m, nil
}

// TableStats represents aggregated traffic statistics for a table from local storage.
// 로컬 저장소에서 집계된 테이블 트래픽 통계를 나타냅니다.
type TableStats struct {
	TableName   string
	AvgCalls    float64
	MaxCalls    int64
	TotalTime   float64
	SnapshotCnt int
}

// GetRecentTPSDelta calculates the TPS by comparing the two most recent snapshots for a table.
// [L22] 최근 두 스냅샷의 누적 호출 수 차이(Delta)를 이용해 실제 초당 트랜잭션 수(TPS)를 계산합니다.
func (a *SQLiteAdapter) GetRecentTPSDelta(tableName string) (float64, error) {
	query := `
		WITH recent_snapshots AS (
			SELECT timestamp, SUM(calls) as total_calls
			FROM workload_snapshots
			WHERE query LIKE ?
			GROUP BY timestamp
			ORDER BY timestamp DESC
			LIMIT 2
		)
		SELECT 
			(MAX(total_calls) - MIN(total_calls)) / 
			(MAX(strftime('%s', timestamp)) - MIN(strftime('%s', timestamp))) as tps
		FROM recent_snapshots;
	`
	
	pattern := "%" + tableName + "%"
	var tps sql.NullFloat64
	err := a.db.QueryRow(query, pattern).Scan(&tps)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate tps delta: %w", err)
	}

	if !tps.Valid {
		return 0, nil // 데이터가 부족한 경우 0 반환
	}

	return tps.Float64, nil
}


// Close closes the SQLite database connection.
// SQLite 데이터베이스 연결을 닫습니다.
func (a *SQLiteAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}
