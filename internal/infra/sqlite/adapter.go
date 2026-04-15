package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// SQLiteAdapter는 로컬 메트릭 저장소인 SQLite DB 파일과의 모든 상호작용을 담당하는 어댑터입니다.
type SQLiteAdapter struct {
	db *sql.DB // 표준 SQL 핸들러
}

// NewSQLiteAdapter는 SQLite 파일을 연결하고 시스템 가동에 필요한 초기 테이블 구조를 생성합니다.
//
// Args:
//   - path: SQLite 데이터베이스 파일 경로
//
// Returns:
//   - *SQLiteAdapter: 초기화 완료된 어댑터 객체
//   - error: 파일 연결 또는 스키마 생성 실패 에러
func NewSQLiteAdapter(path string) (*SQLiteAdapter, error) {
	// 1. CGO 없는 순수 Go 기반 SQLite 드라이버 연결
	db, err := sql.Open("sqlite", path)
	if err != nil { return nil, fmt.Errorf("SQLite 연결 실패: %w", err) }

	adapter := &SQLiteAdapter{db: db}
	
	// 2. 필수 테이블 및 인덱스 자동 생성
	if err := adapter.InitializeSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("SQLite 초기화 실패: %w", err)
	}

	return adapter, nil
}

// InitializeSchema는 MigraGuard 가동에 필수적인 시계열 및 통계 테이블들을 정의합니다.
//
// Returns:
//   - error: 스키마 생성 쿼리 실행 에러
func (a *SQLiteAdapter) InitializeSchema() error {
	query := `
	-- 1. [워크로드 스냅샷] 시점별 쿼리 변화량(Delta) 기록용
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

	-- 2. [테이블 동적 지표] 분석 리포트의 히스토리 보관용
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

	-- 3. [누적 원본 보관함] 다음 주기 델타 계산을 위한 기준점 저장용
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
	// 1. 일괄 쿼리 실행
	_, err := a.db.Exec(query)
	return err
}

// Close는 SQLite 커넥션을 안전하게 닫습니다.
//
// Returns:
//   - error: 리소스 해제 에러
func (a *SQLiteAdapter) Close() error {
	// 1. DB 핸들 폐쇄
	if a.db != nil { return a.db.Close() }
	return nil
}

// MaintenancePurgeData는 저장소 가용성 확보를 위해 보존 기간이 지난 데이터를 영구 삭제합니다.
//
// Args:
//   - retentionDays: 데이터 보존 일수
//
// Returns:
//   - error: 데이터 삭제 또는 진공(VACUUM) 에러
func (a *SQLiteAdapter) MaintenancePurgeData(retentionDays int) error {
	// 1. 시계열 데이터 삭제
	_, _ = a.db.Exec(`DELETE FROM workload_snapshots WHERE timestamp < datetime('now', '-' || ? || ' days')`, retentionDays)
	_, _ = a.db.Exec(`DELETE FROM table_metrics WHERE timestamp < datetime('now', '-' || ? || ' days')`, retentionDays)

	// 2. 물리적 디스크 공간 최적화
	_, err := a.db.Exec("VACUUM")
	return err
}
