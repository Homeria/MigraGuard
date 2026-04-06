package db

import (
	"database/sql"
	"fmt"
)

// GetRecentTPSByDelta는 저장된 델타 스냅샷을 기반으로 특정 테이블의 최근 실질 TPS를 계산합니다.
func (a *SQLiteAdapter) GetRecentTPSByDelta(tableName string) (float64, error) {
	query := `
		WITH recent_delta AS (
			SELECT timestamp, SUM(calls) as delta_calls
			FROM workload_snapshots
			WHERE query LIKE ?
			GROUP BY timestamp
			ORDER BY timestamp DESC
			LIMIT 2
		)
		SELECT 
			CAST(MAX(delta_calls) AS DOUBLE) / 
			NULLIF(MAX(strftime('%s', timestamp)) - MIN(strftime('%s', timestamp)), 0) as tps
		FROM recent_delta;
	`
	pattern := "%" + tableName + "%"
	var tps sql.NullFloat64
	err := a.db.QueryRow(query, pattern).Scan(&tps)
	if err != nil {
		return 0, fmt.Errorf("델타 기반 TPS 계산 실패: %w", err)
	}

	if !tps.Valid {
		return 0, nil
	}
	return tps.Float64, nil
}

// GetTableBaselineStatistics는 과거 이력 데이터를 분석하여 평균(1h) 및 최대(24h) TPS를 산출합니다.
func (a *SQLiteAdapter) GetTableBaselineStatistics(tableName string) (*BaselineStats, error) {
	avgQuery := `SELECT AVG(tps) FROM table_metrics WHERE table_name = ? AND timestamp > datetime('now', '-1 hour')`
	peakQuery := `SELECT MAX(tps) FROM table_metrics WHERE table_name = ? AND timestamp > datetime('now', '-24 hours')`

	var stats BaselineStats
	var avg, peak sql.NullFloat64

	if err := a.db.QueryRow(avgQuery, tableName).Scan(&avg); err != nil {
		return nil, fmt.Errorf("평균 TPS 통계 조회 실패: %w", err)
	}
	if err := a.db.QueryRow(peakQuery, tableName).Scan(&peak); err != nil {
		return nil, fmt.Errorf("피크 TPS 통계 조회 실패: %w", err)
	}

	if avg.Valid {
		stats.AvgTPS_1h = avg.Float64
	}
	if peak.Valid {
		stats.PeakTPS_24h = peak.Float64
	}

	return &stats, nil
}

// IdentifySafestDeploymentWindow는 지난 24시간의 트래픽을 시간대별로 분석하여 가장 부하가 적은 '안전 시간대'를 추천합니다.
func (a *SQLiteAdapter) IdentifySafestDeploymentWindow() (string, float64, error) {
	query := `
		SELECT strftime('%H:00', timestamp) as hour, AVG(tps) as avg_tps
		FROM table_metrics
		WHERE timestamp > datetime('now', '-24 hours')
		GROUP BY hour
		ORDER BY avg_tps ASC
		LIMIT 1
	`
	var hour string
	var avgTPS float64
	err := a.db.QueryRow(query).Scan(&hour, &avgTPS)
	if err != nil {
		if err == sql.ErrNoRows {
			return "데이터 부족", 0, nil
		}
		return "", 0, fmt.Errorf("안전 시간대 분석 실패: %w", err)
	}
	return hour, avgTPS, nil
}

// GetLatestTableMetrics는 특정 테이블의 가장 최근 수집된 동적 지표를 가져옵니다.
func (a *SQLiteAdapter) GetLatestTableMetrics(tableName string) (*TableDynamicMetrics, error) {
	query := `
		SELECT table_name, table_size, replication_lag, active_connections, p99_time, tps
		FROM table_metrics
		WHERE table_name = ?
		ORDER BY timestamp DESC
		LIMIT 1;
	`
	var m TableDynamicMetrics
	err := a.db.QueryRow(query, tableName).Scan(
		&m.TableName,
		&m.TableSize,
		&m.ReplicationLag,
		&m.ActiveConnections,
		&m.P99Time,
		&m.TPS,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("최신 테이블 지표 조회 실패: %w", err)
	}
	return &m, nil
}
