package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// SQLiteAdapter는 로컬 시계열 데이터 및 수집기 상태를 저장하는 SQLite 라이브러리 관리 객체입니다.
type SQLiteAdapter struct {
	db *sql.DB
}

// NewSQLiteAdapter는 지정된 경로에 SQLite 데이터베이스 연결을 생성하고 초기 스키마를 구성합니다.
func NewSQLiteAdapter(path string) (*SQLiteAdapter, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("SQLite 데이터베이스 열기 실패: %w", err)
	}

	adapter := &SQLiteAdapter{db: db}
	if err := adapter.InitializeSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("SQLite 스키마 초기화 실패: %w", err)
	}

	return adapter, nil
}

// InitializeSchema는 MigraGuard 작동에 필요한 모든 테이블과 인덱스를 생성합니다.
func (a *SQLiteAdapter) InitializeSchema() error {
	query := `
	-- 델타(차이값) 워크로드 스냅샷 저장소
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

	-- 테이블별 동적 지표 이력 저장소
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

	-- 차이값 계산을 위한 마지막 누적 원본 데이터 보관소
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

// Close는 SQLite 데이터베이스와의 연결을 안전하게 닫습니다.
func (a *SQLiteAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// MaintenancePurgeData는 설정된 보존 기간을 넘긴 오래된 데이터를 삭제하고 DB 공간을 최적화(VACUUM)합니다.
func (a *SQLiteAdapter) MaintenancePurgeData(retentionDays int) error {
	// 1. 오래된 스냅샷 삭제
	_, err := a.db.Exec(`DELETE FROM workload_snapshots WHERE timestamp < datetime('now', '-' || ? || ' days')`, retentionDays)
	if err != nil {
		return fmt.Errorf("워크로드 스냅샷 정리 실패: %w", err)
	}

	// 2. 오래된 테이블 지표 삭제
	_, err = a.db.Exec(`DELETE FROM table_metrics WHERE timestamp < datetime('now', '-' || ? || ' days')`, retentionDays)
	if err != nil {
		return fmt.Errorf("테이블 지표 정리 실패: %w", err)
	}

	// 3. 물리적 공간 회수 및 조각 모음
	_, err = a.db.Exec("VACUUM")
	return err
}
