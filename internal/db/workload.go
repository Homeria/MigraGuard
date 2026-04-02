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
