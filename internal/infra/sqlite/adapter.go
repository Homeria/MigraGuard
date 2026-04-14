package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// SQLiteAdapter는 로컬 메트릭 저장소 및 시계열 데이터 관리를 담당하는 어댑터입니다.
type SQLiteAdapter struct {
	db *sql.DB
}

// NewSQLiteAdapter는 지정된 경로에 SQLite 연결을 생성하고 시스템 가동에 필요한 초기 스키마를 구성합니다.
func NewSQLiteAdapter(path string) (*SQLiteAdapter, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("SQLite 데이터베이스 파일 열기 실패: %w", err)
	}

	adapter := &SQLiteAdapter{db: db}
	if err := adapter.InitializeSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("SQLite 초기 스키마 구성 실패: %w", err)
	}

	return adapter, nil
}

// InitializeSchema는 MigraGuard 지표 수집 및 분석에 필요한 테이블들과 인덱스를 생성합니다.
func (a *SQLiteAdapter) InitializeSchema() error {
	query := `
	-- 1. 델타(변화량) 기반의 워크로드 스냅샷 저장소
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

	-- 2. 테이블별 동적 지표(TPS, 응답시간 등) 이력 저장소
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

	-- 3. 델타 계산을 위해 가장 최근의 누적 통계 원본을 보관하는 테이블
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

// Close는 SQLite 데이터베이스 연결을 종료합니다.
func (a *SQLiteAdapter) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// MaintenancePurgeData는 설정된 보관 주기를 초과한 과거 데이터를 삭제하고 VACUUM을 통해 저장 공간을 최적화합니다.
func (a *SQLiteAdapter) MaintenancePurgeData(retentionDays int) error {
	// 과거 스냅샷 정리
	_, err := a.db.Exec(`DELETE FROM workload_snapshots WHERE timestamp < datetime('now', '-' || ? || ' days')`, retentionDays)
	if err != nil {
		return fmt.Errorf("워크로드 스냅샷 데이터 정리 실패: %w", err)
	}

	// 과거 테이블 지표 정리
	_, err = a.db.Exec(`DELETE FROM table_metrics WHERE timestamp < datetime('now', '-' || ? || ' days')`, retentionDays)
	if err != nil {
		return fmt.Errorf("테이블 지표 데이터 정리 실패: %w", err)
	}

	// 물리적 공간 회수
	_, err = a.db.Exec("VACUUM")
	return err
}
