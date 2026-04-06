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

// GetLastOriginalSnapshots retrieves the last recorded cumulative stats for delta calculation.
func (a *SQLiteAdapter) GetLastOriginalSnapshots() (map[int64]WorkloadSnapshot, error) {
	query := `SELECT query_id, query, calls, total_time, rows_affected, shared_blks_hit, shared_blks_read FROM original_pg_stat_statements`
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query original stats: %w", err)
	}
	defer rows.Close()

	snapshots := make(map[int64]WorkloadSnapshot)
	for rows.Next() {
		var s WorkloadSnapshot
		err := rows.Scan(&s.QueryID, &s.Query, &s.Calls, &s.TotalTime, &s.Rows, &s.SharedBlksHit, &s.SharedBlksRead)
		if err != nil {
			return nil, fmt.Errorf("failed to scan original stat row: %w", err)
		}
		snapshots[s.QueryID] = s
	}
	return snapshots, nil
}

// UpsertOriginalSnapshots updates the last recorded cumulative stats in SQLite.
func (a *SQLiteAdapter) UpsertOriginalSnapshots(snapshots []WorkloadSnapshot) error {
	tx, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO original_pg_stat_statements (
			query_id, query, calls, total_time, rows_affected, shared_blks_hit, shared_blks_read, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(query_id) DO UPDATE SET
			calls = excluded.calls,
			total_time = excluded.total_time,
			rows_affected = excluded.rows_affected,
			shared_blks_hit = excluded.shared_blks_hit,
			shared_blks_read = excluded.shared_blks_read,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare upsert: %w", err)
	}
	defer stmt.Close()

	for _, s := range snapshots {
		_, err := stmt.Exec(s.QueryID, s.Query, s.Calls, s.TotalTime, s.Rows, s.SharedBlksHit, s.SharedBlksRead)
		if err != nil {
			return fmt.Errorf("failed to execute upsert: %w", err)
		}
	}
	return tx.Commit()
}

// SaveSnapshots persists multiple workload snapshots to the SQLite database.
// 여러 워크로드 스냅샷을 트랜잭션을 통해 SQLite 데이터베이스에 효율적으로 저장합니다.
func (a *SQLiteAdapter) SaveSnapshots(snapshots []WorkloadSnapshot) error {

	tx, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// collector.collect()에서 pg_stat_statements 쿼리를 통해 가져온 Snapshot을 SQLite에 저장하기 시작
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

// BaselineStats represents statistical traffic data for a table.
type BaselineStats struct {
	AvgTPS_1h   float64
	PeakTPS_24h float64
}

// GetTableBaselineStats retrieves the average and peak TPS for a table from SQLite.
// [L32] SQLite에 저장된 과거 데이터를 분석하여 최근 1시간 평균 및 24시간 최대 TPS를 산출합니다.
func (a *SQLiteAdapter) GetTableBaselineStats(tableName string) (*BaselineStats, error) {
	// 1시간 평균 TPS 조회
	avgQuery := `
		SELECT AVG(tps) FROM table_metrics 
		WHERE table_name = ? AND timestamp > datetime('now', '-1 hour')
	`
	// 24시간 최대 TPS 조회
	peakQuery := `
		SELECT MAX(tps) FROM table_metrics 
		WHERE table_name = ? AND timestamp > datetime('now', '-24 hours')
	`

	var stats BaselineStats
	var avg, peak sql.NullFloat64

	if err := a.db.QueryRow(avgQuery, tableName).Scan(&avg); err != nil {
		return nil, fmt.Errorf("failed to get avg tps: %w", err)
	}
	if err := a.db.QueryRow(peakQuery, tableName).Scan(&peak); err != nil {
		return nil, fmt.Errorf("failed to get peak tps: %w", err)
	}

	if avg.Valid {
		stats.AvgTPS_1h = avg.Float64
	}
	if peak.Valid {
		stats.PeakTPS_24h = peak.Float64
	}

	return &stats, nil
}

// GetSafeWindow finds the best time to deploy (lowest traffic) within the last 24 hours.
// [L33] 지난 24시간 동안의 트래픽 패턴을 분석하여 배포하기 가장 안전한 시간대를 찾습니다.
func (a *SQLiteAdapter) GetSafeWindow() (string, float64, error) {
	query := `
		SELECT strftime('%H:00', timestamp) as hour, AVG(tps) as avg_tps
		FROM table_metrics
		WHERE timestamp > datetime('now', '-24 hours')
		GROUP BY hour
		ORDER BY avg_tps ASC
		LIMIT 1
	`
	var hour string
	var avgTPS float64
	err := a.db.QueryRow(query).Scan(&hour, &avgTPS)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", 0, nil
		}
		return "", 0, fmt.Errorf("failed to find safe window: %w", err)
	}
	return hour, avgTPS, nil
}

// GetRecentTPSDelta calculates the TPS using the most recent delta snapshot.
// [L22] 최근 스냅샷의 델타(변화량) 호출 수를 시간 간격으로 나누어 실제 초당 트랜잭션 수(TPS)를 계산합니다.
func (a *SQLiteAdapter) GetRecentTPSDelta(tableName string) (float64, error) {
	query := `
		WITH recent AS (
			SELECT timestamp, SUM(calls) as delta_calls
			FROM workload_snapshots
			WHERE query LIKE ?
			GROUP BY timestamp
			ORDER BY timestamp DESC
			LIMIT 2
		)
		SELECT 
			CAST(MAX(delta_calls) AS DOUBLE) / 
			NULLIF(MAX(strftime('%s', timestamp)) - MIN(strftime('%s', timestamp)), 0) as tps
		FROM recent;
	`
	// 참고: LIMIT 2를 가져오는 이유는 두 스냅샷 사이의 실제 시간 간격(strftime 차이)을 정확히 알기 위함입니다.
	// 결과적으로 가장 최근(MAX)의 delta_calls를 두 스냅샷 사이의 초 단위 시간으로 나눕니다.

	pattern := "%" + tableName + "%"
	var tps sql.NullFloat64
	err := a.db.QueryRow(query, pattern).Scan(&tps)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate tps delta: %w", err)
	}

	if !tps.Valid {
		return 0, nil
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

// PurgeOldSnapshots deletes snapshots older than the retention days and optimizes SQLite.
// 설정된 보존 기간(retentionDays)보다 오래된 스냅샷 데이터를 삭제하고 SQLite 공간을 최적화합니다.
func (a *SQLiteAdapter) PurgeOldSnapshots(retentionDays int) error {

	// 1. Delete old workload snapshots
	// 오래된 워크로드 스냅샷을 삭제합니다.
	workloadQuery := `DELETE FROM workload_snapshots WHERE timestamp < datetime('now', '-' || ? || ' days')`
	_, err := a.db.Exec(workloadQuery, retentionDays)
	if err != nil {
		return fmt.Errorf("failed to purge old workload snapshots: %w", err)
	}

	// 2. Delete old table metrics
	// 오래된 테이블 지표를 삭제합니다.
	metricsQuery := `DELETE FROM table_metrics WHERE timestamp < datetime('now', '-' || ? || ' days')`
	_, err = a.db.Exec(metricsQuery, retentionDays)
	if err != nil {
		return fmt.Errorf("failed to purge old table metrics: %w", err)
	}

	// 3. VACUUM the database to reclaim space
	// SQLite의 VACUUM 명령을 실행하여 삭제된 공간을 회수하고 최적화합니다.
	_, err = a.db.Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("failed to vacuum sqlite: %w", err)
	}

	return nil
}
