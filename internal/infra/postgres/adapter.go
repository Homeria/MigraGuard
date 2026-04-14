package postgres

import (
	"context"
	"time"

	"github.com/Homeria/MigraGuard/internal/shared/errors"
	"github.com/Homeria/MigraGuard/internal/shared/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresAdapter는 PostgreSQL 데이터베이스와의 통신 및 지표 수집을 담당하는 어댑터입니다.
type PostgresAdapter struct {
	pool *pgxpool.Pool
}

// NewPostgresAdapter는 새로운 PostgresAdapter 인스턴스를 생성합니다.
func NewPostgresAdapter(pool *pgxpool.Pool) *PostgresAdapter {
	return &PostgresAdapter{pool: pool}
}

// FetchCurrentWorkloadSnapshot은 pg_stat_statements 뷰를 스캔하여 현재 시점의 누적 쿼리 통계 데이터를 가져옵니다.
func (a *PostgresAdapter) FetchCurrentWorkloadSnapshot(ctx context.Context) ([]types.WorkloadSnapshot, error) {
	if a.pool == nil {
		return nil, errors.ErrDatabaseConn
	}

	query := `
		SELECT 
			queryid, 
			query, 
			calls, 
			total_exec_time, 
			rows, 
			shared_blks_hit, 
			shared_blks_read
		FROM pg_stat_statements
		ORDER BY total_exec_time DESC
		LIMIT 100;
	`

	rows, err := a.pool.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "FetchCurrentWorkloadSnapshot", "pg_stat_statements 조회 도중 오류가 발생했습니다.")
	}
	defer rows.Close()

	var snapshots []types.WorkloadSnapshot
	now := time.Now()

	for rows.Next() {
		var s types.WorkloadSnapshot
		s.Timestamp = now
		err := rows.Scan(
			&s.QueryID,
			&s.Query,
			&s.Calls,
			&s.TotalTime,
			&s.Rows,
			&s.SharedBlksHit,
			&s.SharedBlksRead,
		)
		if err != nil {
			return nil, errors.Wrap(err, "FetchCurrentWorkloadSnapshot", "결과 데이터 스캔 중 오류가 발생했습니다.")
		}
		snapshots = append(snapshots, s)
	}

	return snapshots, nil
}

// FetchTableDynamicMetrics는 리스크 엔진 분석에 필요한 테이블의 실시간 동적 지표(크기, 응답시간, 커넥션 등)를 수집합니다.
func (a *PostgresAdapter) FetchTableDynamicMetrics(ctx context.Context, tableName string) (*types.TableDynamicMetrics, error) {
	if a.pool == nil {
		return nil, errors.ErrDatabaseConn
	}

	metrics := &types.TableDynamicMetrics{TableName: tableName}

	// 1. 테이블의 물리적 전체 크기 조회 (인덱스 용량 포함)
	err := a.pool.QueryRow(ctx, "SELECT pg_total_relation_size($1)", tableName).Scan(&metrics.TableSize)
	if err != nil {
		return nil, errors.WrapWithTable(err, "FetchTableDynamicMetrics", tableName, "테이블 크기 정보를 조회할 수 없습니다.")
	}

	// 2. 현재 DB의 복제 지연(Replication Lag) 상태 확인
	lagQuery := `
		SELECT COALESCE(EXTRACT(EPOCH FROM (now() - reply_time)), 0)
		FROM pg_stat_replication
		ORDER BY reply_time ASC LIMIT 1;
	`
	_ = a.pool.QueryRow(ctx, lagQuery).Scan(&metrics.ReplicationLag)

	// 3. 해당 테이블을 타겟으로 실행 중인 실시간 활성 세션 수 파악
	activeQuery := `
		SELECT count(*)
		FROM pg_stat_activity
		WHERE query LIKE '%' || $1 || '%'
		AND state = 'active'
		AND pid <> pg_backend_pid();
	`
	err = a.pool.QueryRow(ctx, activeQuery, tableName).Scan(&metrics.ActiveConnections)
	if err != nil {
		return nil, errors.WrapWithTable(err, "FetchTableDynamicMetrics", tableName, "활성 연결 수 조회에 실패했습니다.")
	}

	// 4. pg_stat_statements 기반의 P99 응답 시간 및 근사 TPS 계산 (PostgreSQL 14+ 대응)
	statsQuery := `
		SELECT 
			PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY max_exec_time) as p99_time,
			SUM(calls) / GREATEST(EXTRACT(EPOCH FROM (now() - (SELECT stats_reset FROM pg_stat_statements_info))), 1) as tps
		FROM pg_stat_statements
		WHERE query LIKE '%' || $1 || '%';
	`
	_ = a.pool.QueryRow(ctx, statsQuery, tableName).Scan(&metrics.P99Time, &metrics.TPS)

	return metrics, nil
}

// CheckTableSchemaPresence는 DDL 대상 테이블과 컬럼이 실제 운영 환경에 존재하는지 사전 검증을 수행합니다.
func (a *PostgresAdapter) CheckTableSchemaPresence(ctx context.Context, tableName string, columns []string) error {
	if a.pool == nil {
		return errors.ErrDatabaseConn
	}

	var tableExists bool
	tableQuery := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = $1 AND table_schema = 'public'
		)
	`
	err := a.pool.QueryRow(ctx, tableQuery, tableName).Scan(&tableExists)
	if err != nil {
		return errors.Wrap(err, "CheckTableSchemaPresence", "테이블 스키마 정보 조회 중 오류가 발생했습니다.")
	}
	if !tableExists {
		return errors.ErrTableNotFound
	}

	// 요청된 컬럼들이 모두 존재하는지 순차 확인
	if len(columns) > 0 {
		for _, col := range columns {
			var colExists bool
			colQuery := `
				SELECT EXISTS (
					SELECT 1 FROM information_schema.columns 
					WHERE table_name = $1 AND column_name = $2 AND table_schema = 'public'
				)
			`
			err := a.pool.QueryRow(ctx, colQuery, tableName, col).Scan(&colExists)
			if err != nil {
				return errors.WrapWithTable(err, "CheckTableSchemaPresence", tableName, "컬럼 메타데이터 조회에 실패했습니다.")
			}
			if !colExists {
				return errors.ErrColumnNotFound
			}
		}
	}

	return nil
}

// Close는 DB 연결 풀을 안전하게 닫습니다.
func (a *PostgresAdapter) Close() {
	if a.pool != nil {
		a.pool.Close()
	}
}
