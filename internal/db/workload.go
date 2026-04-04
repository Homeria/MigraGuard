package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// WorkloadSnapshot represents a snapshot of the database workload.
// 데이터베이스 워크로드의 스냅샷을 나타냅니다.
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

// TableDynamicMetrics represents dynamic metrics of a table for Risk Model v3.0.
// 위험도 모델 v3.0을 위한 테이블의 동적 지표를 나타냅니다.
type TableDynamicMetrics struct {
	TableName         string  `json:"table_name"`
	TableSize         int64   `json:"table_size"`          // S_table
	ReplicationLag    float64 `json:"replication_lag"`    // Lag_repl (seconds)
	ActiveConnections int     `json:"active_connections"` // C_active
	P99Time           float64 `json:"p99_time"`           // T_p99 (ms)
	TPS               float64 `json:"tps"`                // Lambda (queries per second)
}

// ValidateSchema checks if the given table and columns exist in the database.
// [L61] information_schema를 조회하여 테이블과 컬럼의 실제 존재 여부를 검증합니다.
func (a *PostgresAdapter) ValidateSchema(ctx context.Context, tableName string, columns []string) error {
	if a.pool == nil {
		return fmt.Errorf("database connection is not established")
	}

	// 1. Check if Table exists
	var tableExists bool
	tableQuery := `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = $1 AND table_schema = 'public'
		)
	`
	err := a.pool.QueryRow(ctx, tableQuery, tableName).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check table existence: %w", err)
	}
	if !tableExists {
		return fmt.Errorf("table '%s' does not exist in the database", tableName)
	}

	// 2. Check if Columns exist (if provided)
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
				return fmt.Errorf("failed to check column existence: %w", err)
			}
			if !colExists {
				return fmt.Errorf("column '%s' does not exist in table '%s'", col, tableName)
			}
		}
	}

	return nil
}

// FetchWorkload retrieves the current snapshot from pg_stat_statements.
// pg_stat_statements에서 현재 워크로드 스냅샷을 가져옵니다.
func (a *PostgresAdapter) FetchWorkload(ctx context.Context) ([]WorkloadSnapshot, error) {
	if a.pool == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	// PostgreSQL 13+ pg_stat_statements query
	// pg_stat_statements 뷰를 조회하는 쿼리입니다.
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
		return nil, fmt.Errorf("failed to query pg_stat_statements: %w", err)
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
			return nil, fmt.Errorf("failed to scan workload row: %w", err)
		}
		snapshots = append(snapshots, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	return snapshots, nil
}

// GetTableDynamicMetrics collects all dynamic metrics required for v3.0 Risk Score calculation.
// 위험도 점수 산출(v3.0)에 필요한 모든 동적 지표를 수집합니다.
func (a *PostgresAdapter) GetTableDynamicMetrics(ctx context.Context, tableName string) (*TableDynamicMetrics, error) {
	if a.pool == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	metrics := &TableDynamicMetrics{TableName: tableName}

	// [L21] 테이블 크기 ($S_table): Get Table Size
	// pg_total_relation_size()를 사용하여 테이블의 물리적 크기를 가져옵니다.
	err := a.pool.QueryRow(ctx, "SELECT pg_total_relation_size($1)", tableName).Scan(&metrics.TableSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get table size: %w", err)
	}

	// [L24] 복제 지연 ($Lag_repl): Get Replication Lag
	// pg_stat_replication에서 현재 복제 지연 시간을 초 단위로 가져옵니다.
	// 지연이 없거나 마스터 단독 환경이면 0을 반환합니다.
	lagQuery := `
		SELECT COALESCE(EXTRACT(EPOCH FROM (now() - reply_time)), 0)
		FROM pg_stat_replication
		ORDER BY reply_time ASC LIMIT 1;
	`
	_ = a.pool.QueryRow(ctx, lagQuery).Scan(&metrics.ReplicationLag)

	// [L23] 활성 커넥션 ($C_active): Get Active Connections
	// 현재 해당 테이블을 쿼리 중이거나 락을 대기 중인 활성 세션 수를 조회합니다.
	activeQuery := `
		SELECT count(*)
		FROM pg_stat_activity
		WHERE query LIKE '%' || $1 || '%'
		AND state = 'active'
		AND pid <> pg_backend_pid();
	`
	err = a.pool.QueryRow(ctx, activeQuery, tableName).Scan(&metrics.ActiveConnections)
	if err != nil {
		return nil, fmt.Errorf("failed to get active connections: %w", err)
	}

	// [L22] 트래픽 처리량 ($\lambda$): Get P99 Time (T_p99) and TPS (Lambda)
	// pg_stat_statements를 활용하여 테이블 대상 쿼리의 P99 지연시간과 TPS를 추정합니다.
	statsQuery := `
		SELECT 
			PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY max_exec_time) as p99_time,
			SUM(calls) / GREATEST(EXTRACT(EPOCH FROM (now() - min(stats_reset))), 1) as tps
		FROM pg_stat_statements
		WHERE query LIKE '%' || $1 || '%';
	`
	_ = a.pool.QueryRow(ctx, statsQuery, tableName).Scan(&metrics.P99Time, &metrics.TPS)

	return metrics, nil}

// GetTableStats retrieves traffic statistics for a specific table.
// 특정 테이블에 대한 트래픽 통계를 조회합니다.
func (a *PostgresAdapter) GetTableStats(ctx context.Context, tableName string) (int64, error) {
	if a.pool == nil {
		return 0, fmt.Errorf("database connection is not established")
	}

	query := `
		SELECT seq_scan + idx_scan as total_scans
		FROM pg_stat_user_tables
		WHERE relname = $1;
	`

	var totalScans int64
	err := a.pool.QueryRow(ctx, query, tableName).Scan(&totalScans)
	if err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get table stats: %w", err)
	}

	return totalScans, nil
}
