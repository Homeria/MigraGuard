package db

import (
	"context"
	"fmt"
	"time"

	"github.com/Homeria/MigraGuard/internal/errors"
	"github.com/jackc/pgx/v5"
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
	if a.pool == nil {
		return errors.ErrDatabaseConn
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
		return errors.Wrap(err, "ValidateSchema", "failed to check table existence")
	}
	if !tableExists {
		return errors.ErrTableNotFound
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
		return nil, errors.Wrap(err, "FetchWorkload", "failed to query pg_stat_statements")
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
	err := a.pool.QueryRow(ctx, "SELECT pg_total_relation_size($1)", tableName).Scan(&metrics.TableSize)
	if err != nil {
		return nil, errors.WrapWithTable(err, "GetTableDynamicMetrics", tableName, "failed to get table size")
	}

	// [L24] Replication Lag
	lagQuery := `
		SELECT COALESCE(EXTRACT(EPOCH FROM (now() - reply_time)), 0)
		FROM pg_stat_replication
		ORDER BY reply_time ASC LIMIT 1;
	`
	_ = a.pool.QueryRow(ctx, lagQuery).Scan(&metrics.ReplicationLag)

	// [L23] Active Connections
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
