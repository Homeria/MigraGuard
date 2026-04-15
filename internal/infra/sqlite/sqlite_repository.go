package sqlite

import (
	"fmt"

	"github.com/Homeria/MigraGuard/internal/shared/types"
)

// RecordDeltaSnapshots는 수집된 차이분(Delta) 데이터를 SQLite 시계열 테이블에 트랜잭션 방식으로 안전하게 저장합니다.
//
// Args:
//   - snapshots: 영속화할 델타 데이터 슬라이스
//
// Returns:
//   - error: 트랜잭션 또는 삽입 오류
func (a *SQLiteAdapter) RecordDeltaSnapshots(snapshots []types.WorkloadSnapshot) error {
	// 1. 원자적 처리를 위한 트랜잭션 시작
	tx, err := a.db.Begin()
	if err != nil {
		return fmt.Errorf("트랜잭션 시작 실패: %w", err)
	}
	defer tx.Rollback()

	// 2. PreparedStatement 준비
	stmt, err := tx.Prepare(`INSERT INTO workload_snapshots (timestamp, query_id, query, calls, total_time, rows_affected, shared_blks_hit, shared_blks_read) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// 3. 각 레코드 반복 삽입
	for _, s := range snapshots {
		if _, err := stmt.Exec(s.Timestamp, s.QueryID, s.Query, s.Calls, s.TotalTime, s.Rows, s.SharedBlksHit, s.SharedBlksRead); err != nil {
			return err
		}
	}

	// 4. 커밋
	return tx.Commit()
}

// RecordTableDynamicMetrics는 분석 리포트 생성을 위해 특정 시점의 테이블 상태 지표를 영속화합니다.
//
// Args:
//   - m: 저장할 동적 지표 객체
//
// Returns:
//   - error: DB 실행 에러
func (a *SQLiteAdapter) RecordTableDynamicMetrics(m *types.TableDynamicMetrics) error {
	// 1. 현재 시점 기준 인서트
	query := `INSERT INTO table_metrics (timestamp, table_name, table_size, replication_lag, active_connections, p99_time, tps) VALUES (CURRENT_TIMESTAMP, ?, ?, ?, ?, ?, ?)`
	_, err := a.db.Exec(query, m.TableName, m.TableSize, m.ReplicationLag, m.ActiveConnections, m.P99Time, m.TPS)
	return err
}

// FetchLastOriginalSnapshots는 델타 계산의 비교군이 될 마지막 누적 원본 데이터들을 조회합니다.
//
// Returns:
//   - map[int64]WorkloadSnapshot: QueryID 기반 매핑 데이터
//   - error: 조회 실패 에러
func (a *SQLiteAdapter) FetchLastOriginalSnapshots() (map[int64]types.WorkloadSnapshot, error) {
	// 1. 원본 저장 테이블 조회
	query := `SELECT query_id, query, calls, total_time, rows_affected, shared_blks_hit, shared_blks_read FROM original_pg_stat_statements`
	rows, err := a.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 2. 맵 구조로 결과 매핑
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

// SynchronizeOriginalSnapshots는 새로 수집된 누적치를 원본 보관함에 업데이트하여 다음 주기의 기준으로 삼습니다.
//
// Args:
//   - snapshots: 최신 누적 통계 목록
//
// Returns:
//   - error: UPSERT 실행 실패 에러
func (a *SQLiteAdapter) SynchronizeOriginalSnapshots(snapshots []types.WorkloadSnapshot) error {
	// 1. 원자적 업데이트를 위한 트랜잭션
	tx, err := a.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 2. ON CONFLICT(Upsert) 구문 준비
	stmt, err := tx.Prepare(`INSERT INTO original_pg_stat_statements (query_id, query, calls, total_time, rows_affected, shared_blks_hit, shared_blks_read, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP) ON CONFLICT(query_id) DO UPDATE SET calls = excluded.calls, total_time = excluded.total_time, rows_affected = excluded.rows_affected, shared_blks_hit = excluded.shared_blks_hit, shared_blks_read = excluded.shared_blks_read, updated_at = CURRENT_TIMESTAMP`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// 3. 데이터 동기화 수행
	for _, s := range snapshots {
		if _, err := stmt.Exec(s.QueryID, s.Query, s.Calls, s.TotalTime, s.Rows, s.SharedBlksHit, s.SharedBlksRead); err != nil {
			return err
		}
	}
	return tx.Commit()
}
