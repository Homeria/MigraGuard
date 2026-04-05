package db

import (
	"context"
	"time"

	"github.com/Homeria/MigraGuard/internal/errors"
)

// WorkloadSnapshot represents a snapshot of the database workload.
type WorkloadSnapshot struct {
	Timestamp      time.Time `json:"timestamp"`
	QueryID        int64     `json:"query_id"`
	Query          string    `json:"query"`
	Calls          int64     `json:"calls"`
	TotalTime      float64   `json:"total_time"`
	Rows           int64     `json:"rows"`
	SharedBlksHit  int64     `json:"shared_blks_hit"`
	SharedBlksRead int64     `json:"shared_blks_read"`
}

// TableDynamicMetrics represents dynamic metrics of a table for Risk Model.
type TableDynamicMetrics struct {
	TableName         string  `json:"table_name"`
	TableSize         int64   `json:"table_size"`
	ReplicationLag    float64 `json:"replication_lag"`
	ActiveConnections int     `json:"active_connections"`
	P99Time           float64 `json:"p99_time"`
	TPS               float64 `json:"tps"`
}

// ValidateSchema checks if the given table and columns exist in the database.
func (a *PostgresAdapter) ValidateSchema(ctx context.Context, tableName string, columns []string) error {

	// Database connection pool이 초기화되지 않은 경우
	if a.pool == nil {
		return errors.ErrDatabaseConn
	}

	// 1. Check if Table exists
	// information_schema.tables 뷰를 조회하여 public 스키마에 지정된 테이블이 존재하는지 확인
	var tableExists bool
	tableQuery := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = $1 AND table_schema = 'public'
		)
	`

	// 쿼리 실행 및 결과 스캔하여 tableExists 변수에 저장
	err := a.pool.QueryRow(ctx, tableQuery, tableName).Scan(&tableExists)
	// 쿼리 실행 중 오류 발생 시
	if err != nil {
		return errors.Wrap(err, "ValidateSchema", "failed to check table existence")
	}
	// 테이블이 존재하지 않는 경우
	if !tableExists {
		return errors.ErrTableNotFound
	}

	// 2. Check if Columns exist (if provided)
	if len(columns) > 0 {
		for _, col := range columns {
			// 컬럼이 존재하는지 확인하기 위해 information_schema.columns 뷰를 조회
			// public 스키마에 지정된 테이블과 컬럼이 존재하는지 확인
			var colExists bool
			colQuery := `
				SELECT EXISTS (
					SELECT 1 FROM information_schema.columns 
					WHERE table_name = $1 AND column_name = $2 AND table_schema = 'public'
				)
			`
			err := a.pool.QueryRow(ctx, colQuery, tableName, col).Scan(&colExists)
			if err != nil {
				return errors.WrapWithTable(err, "ValidateSchema", tableName, "failed to check column existence")
			}
			if !colExists {
				return errors.ErrColumnNotFound
			}
		}
	}

	return nil
}

// FetchWorkload retrieves the current snapshot from pg_stat_statements.
func (a *PostgresAdapter) FetchWorkload(ctx context.Context) ([]WorkloadSnapshot, error) {
	if a.pool == nil {
		return nil, errors.ErrDatabaseConn
	}

	// pg_stat_statements 쿼리
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

	// pg_stat_statements 쿼리 전송 및 응답
	rows, err := a.pool.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "FetchWorkload", "failed to query pg_stat_statements")
	}
	defer rows.Close()

	// snapshot 변수 생성
	var snapshots []WorkloadSnapshot
	now := time.Now()

	// pg_stat_statements 쿼리 응답을 WorkloadSnapshot 구조체로 매핑
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
			return nil, errors.Wrap(err, "FetchWorkload", "failed to scan workload row")
		}
		snapshots = append(snapshots, s)
	}

	return snapshots, nil
}

// GetTableDynamicMetrics collects all dynamic metrics required for Risk Score calculation.
func (a *PostgresAdapter) GetTableDynamicMetrics(ctx context.Context, tableName string) (*TableDynamicMetrics, error) {
	if a.pool == nil {
		return nil, errors.ErrDatabaseConn
	}

	metrics := &TableDynamicMetrics{TableName: tableName}

	// [L21] Table Size
	// pg_total_relation_size: 특정 테이블이 디스크에서 차지하고 있는 전체 용량을 바이트 단위로 변환한 값
	err := a.pool.QueryRow(ctx, "SELECT pg_total_relation_size($1)", tableName).Scan(&metrics.TableSize)
	if err != nil {
		return nil, errors.WrapWithTable(err, "GetTableDynamicMetrics", tableName, "failed to get table size")
	}

	// [L24] Replication Lag
	// pg_stat_replication : Primary(Master) 서버에 연결된 Replica(standby) 서버들의 목록, 동기화 상태, 데이터 전송 및 적용 위치(LSN - Log Squence Number)
	lagQuery := `
		SELECT COALESCE(EXTRACT(EPOCH FROM (now() - reply_time)), 0)
		FROM pg_stat_replication
		ORDER BY reply_time ASC LIMIT 1;
	`
	_ = a.pool.QueryRow(ctx, lagQuery).Scan(&metrics.ReplicationLag)

	// [L23] Active Connections
	// pg_stat_activity : 현재 타겟 DB에 접속해 있는 모든 연결(세션)의 실시간 활동 상태
	activeQuery := `
		SELECT count(*)
		FROM pg_stat_activity
		WHERE query LIKE '%' || $1 || '%'
		AND state = 'active'
		AND pid <> pg_backend_pid();
	`
	err = a.pool.QueryRow(ctx, activeQuery, tableName).Scan(&metrics.ActiveConnections)
	if err != nil {
		return nil, errors.WrapWithTable(err, "GetTableDynamicMetrics", tableName, "failed to get active connections")
	}

	// [L22] P99 & TPS
	statsQuery := `
		SELECT 
			PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY max_exec_time) as p99_time,
			SUM(calls) / GREATEST(EXTRACT(EPOCH FROM (now() - min(stats_reset))), 1) as tps
		FROM pg_stat_statements
		WHERE query LIKE '%' || $1 || '%';
	`
	_ = a.pool.QueryRow(ctx, statsQuery, tableName).Scan(&metrics.P99Time, &metrics.TPS)

	return metrics, nil
}
