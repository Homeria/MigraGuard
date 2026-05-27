package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Homeria/MigraGuard/pkg/migraguard/internal/shared/errors"
	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresAdapter collects real-time metrics from PostgreSQL.
type PostgresAdapter struct {
	pool *pgxpool.Pool
}

// NewPostgresAdapter creates a new adapter with an existing pool.
func NewPostgresAdapter(pool *pgxpool.Pool) *PostgresAdapter {
	return &PostgresAdapter{pool: pool}
}

// NewAdapter creates a new pool and adapter from a DSN.
func NewAdapter(dsn string) (*PostgresAdapter, error) {
	pool, err := ConnectPostgres(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres connection failed: %w", err)
	}
	return &PostgresAdapter{pool: pool}, nil
}

// FetchCurrentWorkloadSnapshot collects query statistics from pg_stat_statements.
func (a *PostgresAdapter) FetchCurrentWorkloadSnapshot(ctx context.Context) ([]types.WorkloadSnapshot, error) {
	if a.pool == nil {
		return nil, errors.ErrDatabaseConn
	}

	query := `SELECT queryid, query, calls, total_exec_time, rows, shared_blks_hit, shared_blks_read FROM pg_stat_statements ORDER BY total_exec_time DESC LIMIT 100;`
	rows, err := a.pool.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "FetchCurrentWorkloadSnapshot", "query failed")
	}
	defer rows.Close()

	var snapshots []types.WorkloadSnapshot
	now := time.Now()
	for rows.Next() {
		var s types.WorkloadSnapshot
		s.Timestamp = now
		if err := rows.Scan(&s.QueryID, &s.Query, &s.Calls, &s.TotalTime, &s.Rows, &s.SharedBlksHit, &s.SharedBlksRead); err != nil {
			continue
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, nil
}

// FetchTableDynamicMetrics collects current metrics for a table.
func (a *PostgresAdapter) FetchTableDynamicMetrics(ctx context.Context, tableName string) (*types.TableDynamicMetrics, error) {
	tables := strings.Split(tableName, ",")
	metrics := &types.TableDynamicMetrics{TableName: tableName}

	for _, t := range tables {
		t = strings.TrimSpace(t)
		var size int64
		var conns int
		
		// Use NullFloat64 to safely handle potentially NULL query results under low traffic
		var p99, tps sql.NullFloat64

		if err := a.pool.QueryRow(ctx, "SELECT pg_total_relation_size($1)", t).Scan(&size); err != nil {
			return nil, fmt.Errorf("failed to fetch table relation size: %w", err)
		}
		metrics.TableSize += size

		activeQuery := `SELECT count(*) FROM pg_stat_activity WHERE query LIKE '%' || $1 || '%' AND state = 'active' AND pid <> pg_backend_pid();`
		if err := a.pool.QueryRow(ctx, activeQuery, t).Scan(&conns); err != nil {
			return nil, fmt.Errorf("failed to query active connections: %w", err)
		}
		if conns > metrics.ActiveConnections {
			metrics.ActiveConnections = conns
		}

		statsQuery := `SELECT PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY max_exec_time) as p99_time, SUM(calls) / GREATEST(EXTRACT(EPOCH FROM (now() - (SELECT stats_reset FROM pg_stat_statements_info))), 1) as tps FROM pg_stat_statements WHERE query LIKE '%' || $1 || '%';`
		if err := a.pool.QueryRow(ctx, statsQuery, t).Scan(&p99, &tps); err != nil {
			return nil, fmt.Errorf("failed to query pg_stat_statements metrics: %w", err)
		}
		
		if p99.Valid && p99.Float64 > metrics.P99Time {
			metrics.P99Time = p99.Float64
		}
		if tps.Valid && tps.Float64 > metrics.TPS {
			metrics.TPS = tps.Float64
		}
	}

	return metrics, nil
}

// CheckTableSchemaPresence verifies if tables or indexes exist in the DB.
func (a *PostgresAdapter) CheckTableSchemaPresence(ctx context.Context, name string, columns []string) error {
	names := strings.Split(name, ",")

	for _, n := range names {
		n = strings.TrimSpace(n)
		var exists bool

		tableQuery := `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE LOWER(table_name) = LOWER($1))`
		if err := a.pool.QueryRow(ctx, tableQuery, n).Scan(&exists); err != nil {
			return fmt.Errorf("failed to query information_schema for table: %w", err)
		}
		if exists {
			continue
		}

		indexQuery := `SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE LOWER(indexname) = LOWER($1))`
		if err := a.pool.QueryRow(ctx, indexQuery, n).Scan(&exists); err != nil {
			return fmt.Errorf("failed to query pg_indexes for index: %w", err)
		}
		if exists {
			continue
		}

		return errors.WrapWithTable(errors.ErrTableNotFound, "PostgresAdapter.CheckTableSchemaPresence", n, "validation failed")
	}
	return nil
}

// Close closes the database connection pool.
func (a *PostgresAdapter) Close() {
	if a.pool != nil {
		a.pool.Close()
	}
}
