package db

import (
	"context"
	"time"

	"github.com/Homeria/MigraGuard/internal/errors"
)

// FetchCurrentWorkloadSnapshot은 pg_stat_statements에서 현재의 누적 통계 데이터를 가져옵니다.
// 이 데이터는 나중에 Delta(차이값) 계산을 위한 기반 데이터로 사용됩니다.
func (a *PostgresAdapter) FetchCurrentWorkloadSnapshot(ctx context.Context) ([]WorkloadSnapshot, error) {
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
		return nil, errors.Wrap(err, "FetchCurrentWorkloadSnapshot", "pg_stat_statements 조회 중 오류 발생")
	}
	defer rows.Close()

	var snapshots []WorkloadSnapshot
	now := time.Now()

	for rows.Next() {
		var s WorkloadSnapshot
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
			return nil, errors.Wrap(err, "FetchCurrentWorkloadSnapshot", "데이터 로우 스캔 실패")
		}
		snapshots = append(snapshots, s)
	}

	return snapshots, nil
}

// FetchTableDynamicMetrics는 분석 엔진(Risk Engine)이 필요한 실시간 지표들을 수집합니다.
// 테이블 크기, 복제 지연, 활성 연결 수, P99 실행 시간 등을 한 번에 조회합니다.
func (a *PostgresAdapter) FetchTableDynamicMetrics(ctx context.Context, tableName string) (*TableDynamicMetrics, error) {
	if a.pool == nil {
		return nil, errors.ErrDatabaseConn
	}

	metrics := &TableDynamicMetrics{TableName: tableName}

	// 1. 테이블의 물리적 전체 크기(인덱스 포함) 조회
	err := a.pool.QueryRow(ctx, "SELECT pg_total_relation_size($1)", tableName).Scan(&metrics.TableSize)
	if err != nil {
		return nil, errors.WrapWithTable(err, "FetchTableDynamicMetrics", tableName, "테이블 크기 조회 실패")
	}

	// 2. 현재 복제 지연(Replication Lag) 상태 확인 (Replica 환경용)
	lagQuery := `
		SELECT COALESCE(EXTRACT(EPOCH FROM (now() - reply_time)), 0)
		FROM pg_stat_replication
		ORDER BY reply_time ASC LIMIT 1;
	`
	_ = a.pool.QueryRow(ctx, lagQuery).Scan(&metrics.ReplicationLag)

	// 3. 해당 테이블을 대상으로 쿼리를 수행 중인 실시간 활성 세션 수 파악
	activeQuery := `
		SELECT count(*)
		FROM pg_stat_activity
		WHERE query LIKE '%' || $1 || '%'
		AND state = 'active'
		AND pid <> pg_backend_pid();
	`
	err = a.pool.QueryRow(ctx, activeQuery, tableName).Scan(&metrics.ActiveConnections)
	if err != nil {
		return nil, errors.WrapWithTable(err, "FetchTableDynamicMetrics", tableName, "활성 연결 수 조회 실패")
	}

	// 4. pg_stat_statements 기반의 P99 응답 시간 및 근사치 TPS 계산
	// Postgres 14+ 버전에서는 stats_reset 정보가 pg_stat_statements_info 뷰에 있습니다.
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

// CheckTableSchemaPresence는 특정 테이블과 그 컬럼들이 실제 운영 DB에 존재하는지 유효성 검사를 수행합니다.
func (a *PostgresAdapter) CheckTableSchemaPresence(ctx context.Context, tableName string, columns []string) error {
	if a.pool == nil {
		return errors.ErrDatabaseConn
	}

	// 1. 테이블 존재 여부 확인
	var tableExists bool
	tableQuery := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = $1 AND table_schema = 'public'
		)
	`
	err := a.pool.QueryRow(ctx, tableQuery, tableName).Scan(&tableExists)
	if err != nil {
		return errors.Wrap(err, "CheckTableSchemaPresence", "테이블 메타데이터 조회 중 오류 발생")
	}
	if !tableExists {
		return errors.ErrTableNotFound
	}

	// 2. 컬럼이 명시된 경우 각 컬럼의 존재 여부 순차 확인
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
				return errors.WrapWithTable(err, "CheckTableSchemaPresence", tableName, "컬럼 메타데이터 조회 실패")
			}
			if !colExists {
				return errors.ErrColumnNotFound
			}
		}
	}

	return nil
}
