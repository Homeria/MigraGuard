package sqlite

import (
	"database/sql"

	"github.com/Homeria/MigraGuard/internal/shared/types"
)

// GetRecentTPSByDelta는 델타 스냅샷을 전수 조사하여 최근 1~2분 사이의 실질 TPS를 도출합니다.
//
// Args:
//   - tableName: 분석 대상 테이블명
//
// Returns:
//   - float64: 산출된 TPS 수치
//   - error: 조회 실패 에러
func (a *SQLiteAdapter) GetRecentTPSByDelta(tableName string) (float64, error) {
	// 1. 최근 2개 수집 주기의 데이터를 CTE로 요약 추출
	query := `
		WITH recent_delta AS (
			SELECT timestamp, SUM(calls) as delta_calls
			FROM workload_snapshots WHERE LOWER(query) LIKE LOWER(?)
			GROUP BY timestamp ORDER BY timestamp DESC LIMIT 2
		)
		SELECT CAST(MAX(delta_calls) AS DOUBLE) / NULLIF(MAX(strftime('%s', timestamp)) - MIN(strftime('%s', timestamp)), 0) as tps
		FROM recent_delta;
	`
	var tps sql.NullFloat64
	// 2. 테이블명 포함 여부 필터링 및 쿼리 실행
	if err := a.db.QueryRow(query, "%"+tableName+"%").Scan(&tps); err != nil {
		return 0, err
	}
	if !tps.Valid {
		return 0, nil
	}
	return tps.Float64, nil
}

// GetTableBaselineStatistics는 과거 24시간 이력을 분석하여 평균 및 피크 부하 지표를 제공합니다.
//
// Args:
//   - tableName: 테이블명
//
// Returns:
//   - *BaselineStats: 1시간 평균 및 24시간 피크 데이터
//   - error: 통계 산출 에러
func (a *SQLiteAdapter) GetTableBaselineStatistics(tableName string) (*types.BaselineStats, error) {
	// 1. 평균 및 피크 동시 집계 쿼리 정의
	avgQuery := `SELECT AVG(tps) FROM table_metrics WHERE table_name = ? AND timestamp > datetime('now', '-1 hour')`
	peakQuery := `SELECT MAX(tps) FROM table_metrics WHERE table_name = ? AND timestamp > datetime('now', '-24 hours')`

	var stats types.BaselineStats
	var avg, peak sql.NullFloat64

	// 2. 순차적 데이터 로드
	_ = a.db.QueryRow(avgQuery, tableName).Scan(&avg)
	_ = a.db.QueryRow(peakQuery, tableName).Scan(&peak)

	if avg.Valid {
		stats.AvgTPS_1h = avg.Float64
	}
	if peak.Valid {
		stats.PeakTPS_24h = peak.Float64
	}
	return &stats, nil
}

// IdentifySafestDeploymentWindow는 24시간 트래픽 추이를 분석하여 가장 배포하기 안전한 최적의 시간대를 추천합니다.
//
// Returns:
//   - string: 추천 시간 (예: "03:00")
//   - float64: 해당 시간대 예상 TPS
//   - error: 분석 실패 에러
func (a *SQLiteAdapter) IdentifySafestDeploymentWindow() (string, float64, error) {
	// 1. 시간대별 그룹화 및 최소 평균 부하 구간 탐색
	query := `SELECT strftime('%H:00', timestamp) as hour, AVG(tps) as avg_tps FROM table_metrics WHERE timestamp > datetime('now', '-24 hours') GROUP BY hour ORDER BY avg_tps ASC LIMIT 1`
	var hour string
	var avgTPS float64
	if err := a.db.QueryRow(query).Scan(&hour, &avgTPS); err != nil {
		if err == sql.ErrNoRows {
			return "데이터 부족", 0, nil
		}
		return "", 0, err
	}
	return hour, avgTPS, nil
}

// GetLatestTableMetrics는 특정 테이블의 가장 마지막에 수집된 동적 지표를 가져옵니다.
//
// Args:
//   - tableName: 테이블명
//
// Returns:
//   - *TableDynamicMetrics: 최신 상태 객체
//   - error: 데이터 조회 에러
func (a *SQLiteAdapter) GetLatestTableMetrics(tableName string) (*types.TableDynamicMetrics, error) {
	// 1. 타임스탬프 기준 최신 레코드 1건 조회
	query := `SELECT table_name, table_size, replication_lag, active_connections, p99_time, tps FROM table_metrics WHERE table_name = ? ORDER BY timestamp DESC LIMIT 1;`
	var m types.TableDynamicMetrics
	if err := a.db.QueryRow(query, tableName).Scan(&m.TableName, &m.TableSize, &m.ReplicationLag, &m.ActiveConnections, &m.P99Time, &m.TPS); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// GetTopHeavyQueries는 현재 시스템 리소스를 가장 많이 점유하고 있는 상위 쿼리들을 추출합니다.
//
// Args:
//   - limit: 추출할 쿼리 개수
//
// Returns:
//   - []TopQueryInfo: 상위 부하 쿼리 상세 목록
//   - error: 분석 실패 에러
func (a *SQLiteAdapter) GetTopHeavyQueries(limit int) ([]types.TopQueryInfo, error) {
	// 1. 최근 1시간 실행 시간 비율(Impact) 계산 복합 쿼리
	query := `WITH total_stat AS (SELECT SUM(total_time) as grand_total FROM workload_snapshots WHERE timestamp > datetime('now', '-1 hour')) SELECT query_id, query, SUM(calls), SUM(total_time), (SUM(total_time) / (SELECT grand_total FROM total_stat)) * 100 as impact FROM workload_snapshots WHERE timestamp > datetime('now', '-1 hour') GROUP BY query_id ORDER BY SUM(total_time) DESC LIMIT ?`
	rows, err := a.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topQueries []types.TopQueryInfo
	for rows.Next() {
		var q types.TopQueryInfo
		if err := rows.Scan(&q.QueryID, &q.QueryText, &q.Calls, &q.TotalTime, &q.Impact); err != nil {
			continue
		}
		topQueries = append(topQueries, q)
	}
	return topQueries, nil
}
