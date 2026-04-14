package sqlite

import (
	"fmt"

	"github.com/Homeria/MigraGuard/internal/shared/types"
)

// RecordDeltaSnapshots는 수집된 차이분(Delta) 데이터를 SQLite 시계열 테이블에 트랜잭션 방식으로 저장합니다.
func (a *SQLiteAdapter) RecordDeltaSnapshots(snapshots []types.WorkloadSnapshot) error {
	tx, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("스냅샷 저장 트랜잭션 시작 실패: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO workload_snapshots (
			timestamp, query_id, query, calls, total_time, rows_affected, shared_blks_hit, shared_blks_read
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("스냅샷 저장 구문 준비 실패: %w", err)
	}
	defer stmt.Close()

	for _, s := range snapshots {
		_, err := stmt.Exec(s.Timestamp, s.QueryID, s.Query, s.Calls, s.TotalTime, s.Rows, s.SharedBlksHit, s.SharedBlksRead)
		if err != nil {
			return fmt.Errorf("스냅샷 데이터 삽입 실패: %w", err)
		}
	}

	return tx.Commit()
}

// RecordTableDynamicMetrics는 분석 엔진용 실시간 지표(크기, 연결 수 등)를 저장합니다.
func (a *SQLiteAdapter) RecordTableDynamicMetrics(m *types.TableDynamicMetrics) error {
	query := `
		INSERT INTO table_metrics (
			timestamp, table_name, table_size, replication_lag, active_connections, p99_time, tps
		) VALUES (CURRENT_TIMESTAMP, ?, ?, ?, ?, ?, ?)
	`
	_, err := a.db.Exec(query, m.TableName, m.TableSize, m.ReplicationLag, m.ActiveConnections, m.P99Time, m.TPS)
	if err != nil {
		return fmt.Errorf("테이블 동적 지표 저장 실패: %w", err)
	}
	return nil
}

// FetchLastOriginalSnapshots는 델타 계산을 위해 마지막으로 저장된 누적 원본 데이터들을 맵 형태로 조회합니다.
func (a *SQLiteAdapter) FetchLastOriginalSnapshots() (map[int64]types.WorkloadSnapshot, error) {
	query := `SELECT query_id, query, calls, total_time, rows_affected, shared_blks_hit, shared_blks_read FROM original_pg_stat_statements`
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("원본 데이터 조회 실패: %w", err)
	}
	defer rows.Close()

	snapshots := make(map[int64]types.WorkloadSnapshot)
	for rows.Next() {
		var s types.WorkloadSnapshot
		err := rows.Scan(&s.QueryID, &s.Query, &s.Calls, &s.TotalTime, &s.Rows, &s.SharedBlksHit, &s.SharedBlksRead)
		if err != nil {
			return nil, fmt.Errorf("원본 데이터 스캔 실패: %w", err)
		}
		snapshots[s.QueryID] = s
	}
	return snapshots, nil
}

// SynchronizeOriginalSnapshots는 새로운 누적 통계치를 원본 저장소에 UPSERT 하여 다음 수집 주기의 기준점으로 삼습니다.
func (a *SQLiteAdapter) SynchronizeOriginalSnapshots(snapshots []types.WorkloadSnapshot) error {
	tx, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("원본 데이터 동기화 트랜잭션 시작 실패: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO original_pg_stat_statements (
			query_id, query, calls, total_time, rows_affected, shared_blks_hit, shared_blks_read, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(query_id) DO UPDATE SET
			calls = excluded.calls,
			total_time = excluded.total_time,
			rows_affected = excluded.rows_affected,
			shared_blks_hit = excluded.shared_blks_hit,
			shared_blks_read = excluded.shared_blks_read,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return fmt.Errorf("원본 데이터 동기화 구문 준비 실패: %w", err)
	}
	defer stmt.Close()

	for _, s := range snapshots {
		_, err := stmt.Exec(s.QueryID, s.Query, s.Calls, s.TotalTime, s.Rows, s.SharedBlksHit, s.SharedBlksRead)
		if err != nil {
			return fmt.Errorf("원본 데이터 업데이트 삽입 실패: %w", err)
		}
	}
	return tx.Commit()
}
