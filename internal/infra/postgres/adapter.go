package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/Homeria/MigraGuard/internal/shared/errors"
	"github.com/Homeria/MigraGuard/internal/shared/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresAdapter는 PostgreSQL 데이터베이스와 통신하여 실시간 운영 지표를 수집하는 어댑터입니다.
type PostgresAdapter struct {
	pool *pgxpool.Pool
}

// NewPostgresAdapter는 PostgresAdapter 인스턴스가 생성합니다.
func NewPostgresAdapter(pool *pgxpool.Pool) *PostgresAdapter {
	return &PostgresAdapter{pool: pool}
}

// FetchCurrentWorkloadSnapshot은 pg_stat_statements 뷰를 조회하여 누적 쿼리 통계를 수집합니다.
func (a *PostgresAdapter) FetchCurrentWorkloadSnapshot(ctx context.Context) ([]types.WorkloadSnapshot, error) {
	if a.pool == nil { return nil, errors.ErrDatabaseConn }

	query := `SELECT queryid, query, calls, total_exec_time, rows, shared_blks_hit, shared_blks_read FROM pg_stat_statements ORDER BY total_exec_time DESC LIMIT 100;`
	rows, err := a.pool.Query(ctx, query)
	if err != nil { return nil, errors.Wrap(err, "FetchCurrentWorkloadSnapshot", "조회 실패") }
	defer rows.Close()

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

// FetchTableDynamicMetrics는 분석 대상 테이블(들)의 실시간 상태를 수집합니다.
// 다중 테이블일 경우 수치를 합산 또는 최댓값으로 집계합니다.
func (a *PostgresAdapter) FetchTableDynamicMetrics(ctx context.Context, tableName string) (*types.TableDynamicMetrics, error) {
	tables := strings.Split(tableName, ",")
	metrics := &types.TableDynamicMetrics{TableName: tableName}

	for _, t := range tables {
		t = strings.TrimSpace(t)
		var size int64
		var conns int
		var p99 float64
		var tps float64

		// 1. 테이블 물리 크기 합산
		_ = a.pool.QueryRow(ctx, "SELECT pg_total_relation_size($1)", t).Scan(&size)
		metrics.TableSize += size

		// 2. 실시간 활성 연결 수 (최댓값 유지)
		activeQuery := `SELECT count(*) FROM pg_stat_activity WHERE query LIKE '%' || $1 || '%' AND state = 'active' AND pid <> pg_backend_pid();`
		_ = a.pool.QueryRow(ctx, activeQuery, t).Scan(&conns)
		if conns > metrics.ActiveConnections {
			metrics.ActiveConnections = conns
		}

		// 3. P99 및 TPS (최댓값 유지)
		statsQuery := `SELECT PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY max_exec_time) as p99_time, SUM(calls) / GREATEST(EXTRACT(EPOCH FROM (now() - (SELECT stats_reset FROM pg_stat_statements_info))), 1) as tps FROM pg_stat_statements WHERE query LIKE '%' || $1 || '%';`
		_ = a.pool.QueryRow(ctx, statsQuery, t).Scan(&p99, &tps)
		if p99 > metrics.P99Time { metrics.P99Time = p99 }
		if tps > metrics.TPS { metrics.TPS = tps }
	}

	return metrics, nil
}

// CheckTableSchemaPresence는 테이블 또는 인덱스가 운영 DB에 실제 존재하는지 사전 검증합니다.
// 쉼표로 구분된 다중 대상에 대응합니다.
func (a *PostgresAdapter) CheckTableSchemaPresence(ctx context.Context, name string, columns []string) error {
	names := strings.Split(name, ",")
	
	for _, n := range names {
		n = strings.TrimSpace(n)
		var exists bool
		
		// 테이블 확인
		tableQuery := `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE LOWER(table_name) = LOWER($1))`
		_ = a.pool.QueryRow(ctx, tableQuery, n).Scan(&exists)
		if exists { continue }

		// 인덱스 확인
		indexQuery := `SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE LOWER(indexname) = LOWER($1))`
		_ = a.pool.QueryRow(ctx, indexQuery, n).Scan(&exists)
		if exists { continue }

		// 하나라도 없으면 에러
		return errors.Wrap(errors.ErrTableNotFound, "CheckTableSchemaPresence", n)
	}
	return nil
}

// Close는 데이터베이스 연결 풀을 정상적으로 닫습니다.
func (a *PostgresAdapter) Close() {
	if a.pool != nil { a.pool.Close() }
}
