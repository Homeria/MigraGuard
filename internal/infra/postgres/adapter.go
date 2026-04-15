package postgres

import (
	"context"
	"time"

	"github.com/Homeria/MigraGuard/internal/shared/errors"
	"github.com/Homeria/MigraGuard/internal/shared/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresAdapter는 PostgreSQL 데이터베이스와 통신하여 실시간 운영 지표를 수집하는 어댑터입니다.
type PostgresAdapter struct {
	pool *pgxpool.Pool
}

// NewPostgresAdapter는 PostgresAdapter 인스턴스를 생성합니다.
//
// Args:
//   - pool: 활성화된 pgxpool 객체
//
// Returns:
//   - *PostgresAdapter: 초기화된 PostgresAdapter 인스턴스
func NewPostgresAdapter(pool *pgxpool.Pool) *PostgresAdapter {
	// 1. 커넥션 풀 할당
	return &PostgresAdapter{pool: pool}
}

// FetchCurrentWorkloadSnapshot은 pg_stat_statements 뷰를 조회하여 누적 쿼리 통계를 수집합니다.
//
// Args:
//   - ctx: 실행 컨텍스트
//
// Returns:
//   - []WorkloadSnapshot: 쿼리 통계 데이터 슬라이스
//   - error: 쿼리 실행 에러
func (a *PostgresAdapter) FetchCurrentWorkloadSnapshot(ctx context.Context) ([]types.WorkloadSnapshot, error) {
	// 1. 풀 유효성 검사
	if a.pool == nil { return nil, errors.ErrDatabaseConn }

	// 2. 누적 통계 조회 쿼리 실행 (상위 100개)
	query := `SELECT queryid, query, calls, total_exec_time, rows, shared_blks_hit, shared_blks_read FROM pg_stat_statements ORDER BY total_exec_time DESC LIMIT 100;`
	rows, err := a.pool.Query(ctx, query)
	if err != nil { return nil, errors.Wrap(err, "FetchCurrentWorkloadSnapshot", "조회 실패") }
	defer rows.Close()

	// 3. 결과 데이터 매핑
	var snapshots []types.WorkloadSnapshot
	now := time.Now()
	for rows.Next() {
		var s types.WorkloadSnapshot
		s.Timestamp = now
		if err := rows.Scan(&s.QueryID, &s.Query, &s.Calls, &s.TotalTime, &s.Rows, &s.SharedBlksHit, &s.SharedBlksRead); err != nil { continue }
		snapshots = append(snapshots, s)
	}
	return snapshots, nil
}

// FetchTableDynamicMetrics는 분석 대상 테이블의 실시간 상태(크기, 연결 수, 응답성)를 수집합니다.
//
// Args:
//   - ctx: 실행 컨텍스트
//   - tableName: 대상 테이블명
//
// Returns:
//   - *TableDynamicMetrics: 실시간 지표 객체
//   - error: 메타데이터 조회 에러
func (a *PostgresAdapter) FetchTableDynamicMetrics(ctx context.Context, tableName string) (*types.TableDynamicMetrics, error) {
	// 1. 기본 구조체 생성
	metrics := &types.TableDynamicMetrics{TableName: tableName}

	// 2. 테이블 물리 크기 조회
	if err := a.pool.QueryRow(ctx, "SELECT pg_total_relation_size($1)", tableName).Scan(&metrics.TableSize); err != nil {
		return nil, errors.WrapWithTable(err, "FetchTableDynamicMetrics", tableName, "크기 로드 실패")
	}

	// 3. 실시간 활성 연결 수 집계
	activeQuery := `SELECT count(*) FROM pg_stat_activity WHERE query LIKE '%' || $1 || '%' AND state = 'active' AND pid <> pg_backend_pid();`
	if err := a.pool.QueryRow(ctx, activeQuery, tableName).Scan(&metrics.ActiveConnections); err != nil {
		return nil, errors.WrapWithTable(err, "FetchTableDynamicMetrics", tableName, "연결 수 로드 실패")
	}

	// 4. P99 응답 시간 및 누적 TPS 계산
	statsQuery := `SELECT PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY max_exec_time) as p99_time, SUM(calls) / GREATEST(EXTRACT(EPOCH FROM (now() - (SELECT stats_reset FROM pg_stat_statements_info))), 1) as tps FROM pg_stat_statements WHERE query LIKE '%' || $1 || '%';`
	_ = a.pool.QueryRow(ctx, statsQuery, tableName).Scan(&metrics.P99Time, &metrics.TPS)

	return metrics, nil
}

// CheckTableSchemaPresence는 테이블 및 컬럼이 운영 DB에 실제 존재하는지 사전 검증합니다.
//
// Args:
//   - ctx: 실행 컨텍스트
//   - tableName: 테이블명
//   - columns: 검증할 컬럼 목록
//
// Returns:
//   - error: 존재하지 않을 경우 에러 반환
func (a *PostgresAdapter) CheckTableSchemaPresence(ctx context.Context, tableName string, columns []string) error {
	// 1. 테이블 존재 여부 확인
	var tableExists bool
	tableQuery := `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = $1 AND table_schema = 'public')`
	if err := a.pool.QueryRow(ctx, tableQuery, tableName).Scan(&tableExists); err != nil { return err }
	if !tableExists { return errors.ErrTableNotFound }

	// 2. 각 컬럼 존재 여부 전수 조사
	for _, col := range columns {
		var colExists bool
		colQuery := `SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = $1 AND column_name = $2 AND table_schema = 'public')`
		_ = a.pool.QueryRow(ctx, colQuery, tableName, col).Scan(&colExists)
		if !colExists { return errors.ErrColumnNotFound }
	}
	return nil
}

// Close는 데이터베이스 연결 풀을 정상적으로 닫습니다.
func (a *PostgresAdapter) Close() {
	if a.pool != nil { a.pool.Close() }
}
