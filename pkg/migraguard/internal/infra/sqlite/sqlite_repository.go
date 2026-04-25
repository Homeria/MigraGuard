package sqlite

import (
	"fmt"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// RecordDeltaSnapshots stores calculated delta metrics in SQLite.
func (a *SQLiteAdapter) RecordDeltaSnapshots(snapshots []types.WorkloadSnapshot) error {
	tx, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("transaction start failed: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO workload_snapshots (timestamp, query_id, query, calls, total_time, rows_affected, shared_blks_hit, shared_blks_read) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range snapshots {
		if _, err := stmt.Exec(s.Timestamp, s.QueryID, s.Query, s.Calls, s.TotalTime, s.Rows, s.SharedBlksHit, s.SharedBlksRead); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// RecordTableDynamicMetrics persists table-specific metrics.
func (a *SQLiteAdapter) RecordTableDynamicMetrics(m *types.TableDynamicMetrics) error {
	query := `INSERT INTO table_metrics (timestamp, table_name, table_size, replication_lag, active_connections, p99_time, tps) VALUES (CURRENT_TIMESTAMP, ?, ?, ?, ?, ?, ?)`
	_, err := a.db.Exec(query, m.TableName, m.TableSize, m.ReplicationLag, m.ActiveConnections, m.P99Time, m.TPS)
	return err
}

// FetchLastOriginalSnapshots retrieves the last recorded raw statistics.
func (a *SQLiteAdapter) FetchLastOriginalSnapshots() (map[int64]types.WorkloadSnapshot, error) {
	query := `SELECT query_id, query, calls, total_time, rows_affected, shared_blks_hit, shared_blks_read FROM original_pg_stat_statements`
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	snapshots := make(map[int64]types.WorkloadSnapshot)
	for rows.Next() {
		var s types.WorkloadSnapshot
		if err := rows.Scan(&s.QueryID, &s.Query, &s.Calls, &s.TotalTime, &s.Rows, &s.SharedBlksHit, &s.SharedBlksRead); err != nil {
			continue
		}
		snapshots[s.QueryID] = s
	}
	return snapshots, nil
}

// SynchronizeOriginalSnapshots updates raw statistics in the baseline table.
func (a *SQLiteAdapter) SynchronizeOriginalSnapshots(snapshots []types.WorkloadSnapshot) error {
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO original_pg_stat_statements (query_id, query, calls, total_time, rows_affected, shared_blks_hit, shared_blks_read, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP) ON CONFLICT(query_id) DO UPDATE SET calls = excluded.calls, total_time = excluded.total_time, rows_affected = excluded.rows_affected, shared_blks_hit = excluded.shared_blks_hit, shared_blks_read = excluded.shared_blks_read, updated_at = CURRENT_TIMESTAMP`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range snapshots {
		if _, err := stmt.Exec(s.QueryID, s.Query, s.Calls, s.TotalTime, s.Rows, s.SharedBlksHit, s.SharedBlksRead); err != nil {
			return err
		}
	}
	return tx.Commit()
}
