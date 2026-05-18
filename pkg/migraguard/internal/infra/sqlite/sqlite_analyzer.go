package sqlite

import (
	"context"
	"database/sql"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// GetRecentTPSByDelta calculates real-time TPS based on workload snapshots.
func (a *SQLiteAdapter) GetRecentTPSByDelta(ctx context.Context, tableName string) (float64, error) {
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
	if err := a.db.QueryRowContext(ctx, query, "%"+tableName+"%").Scan(&tps); err != nil {
		return 0, err
	}
	if !tps.Valid {
		return 0, nil
	}
	return tps.Float64, nil
}

// GetTableBaselineStatistics retrieves 1h average and 24h peak statistics.
func (a *SQLiteAdapter) GetTableBaselineStatistics(ctx context.Context, tableName string) (*types.BaselineStats, error) {
	avgQuery := `SELECT AVG(tps) FROM table_metrics WHERE table_name = ? AND timestamp > datetime('now', '-1 hour')`
	peakQuery := `SELECT MAX(tps) FROM table_metrics WHERE table_name = ? AND timestamp > datetime('now', '-24 hours')`

	var stats types.BaselineStats
	var avg, peak sql.NullFloat64

	_ = a.db.QueryRowContext(ctx, avgQuery, tableName).Scan(&avg)
	_ = a.db.QueryRowContext(ctx, peakQuery, tableName).Scan(&peak)

	if avg.Valid {
		stats.AvgTPS_1h = avg.Float64
	}
	if peak.Valid {
		stats.PeakTPS_24h = peak.Float64
	}
	return &stats, nil
}

// IdentifySafestDeploymentWindow identifies the low-traffic window for deployment.
func (a *SQLiteAdapter) IdentifySafestDeploymentWindow(ctx context.Context) (string, float64, error) {
	query := `SELECT strftime('%H:00', timestamp) as hour, AVG(tps) as avg_tps FROM table_metrics WHERE timestamp > datetime('now', '-24 hours') GROUP BY hour ORDER BY avg_tps ASC LIMIT 1`
	var hour string
	var avgTPS float64
	if err := a.db.QueryRowContext(ctx, query).Scan(&hour, &avgTPS); err != nil {
		if err == sql.ErrNoRows {
			return "Insufficient Data", 0, nil
		}
		return "", 0, err
	}
	return hour, avgTPS, nil
}

// GetLatestTableMetrics retrieves the most recent table metrics.
func (a *SQLiteAdapter) GetLatestTableMetrics(ctx context.Context, tableName string) (*types.TableDynamicMetrics, error) {
	query := `SELECT table_name, table_size, replication_lag, active_connections, p99_time, tps FROM table_metrics WHERE table_name = ? ORDER BY timestamp DESC LIMIT 1;`
	var m types.TableDynamicMetrics
	if err := a.db.QueryRowContext(ctx, query, tableName).Scan(&m.TableName, &m.TableSize, &m.ReplicationLag, &m.ActiveConnections, &m.P99Time, &m.TPS); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

// GetTopHeavyQueries identifies queries with high performance impact.
func (a *SQLiteAdapter) GetTopHeavyQueries(ctx context.Context, limit int) ([]types.TopQueryInfo, error) {
	query := `WITH total_stat AS (SELECT SUM(total_time) as grand_total FROM workload_snapshots WHERE timestamp > datetime('now', '-1 hour')) SELECT query_id, query, SUM(calls), SUM(total_time), (SUM(total_time) / (SELECT grand_total FROM total_stat)) * 100 as impact FROM workload_snapshots WHERE timestamp > datetime('now', '-1 hour') GROUP BY query_id ORDER BY SUM(total_time) DESC LIMIT ?`
	rows, err := a.db.QueryContext(ctx, query, limit)
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

// Get24HourTrafficForecast generates a 24-hour baseline profile by averaging historical data by hour.
func (a *SQLiteAdapter) Get24HourTrafficForecast(ctx context.Context, tableName string) ([]types.ForecastTimeSlot, error) {
	// Use substr(timestamp, 12, 2) to reliably extract HH from ISO-8601 strings (e.g., 2026-05-15T14:30:00+09:00)
	query := `
		SELECT 
			CAST(substr(timestamp, 12, 2) AS INTEGER) as hour, 
			AVG(tps) as avg_tps, 
			MIN(tps) as min_tps,
			MAX(tps) as max_tps,
			AVG(p99_time) as avg_p99 
		FROM table_metrics 
		WHERE LOWER(table_name) = LOWER(?) 
		GROUP BY hour 
		ORDER BY hour ASC
	`
	rows, err := a.db.QueryContext(ctx, query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	forecast := make([]types.ForecastTimeSlot, 0, 24)
	for rows.Next() {
		var f types.ForecastTimeSlot
		if err := rows.Scan(&f.Hour, &f.ExpectedTPS, &f.MinTPS, &f.MaxTPS, &f.ExpectedP99); err != nil {
			continue
		}
		forecast = append(forecast, f)
	}

	// Ensure we have 24 hours even if some are missing data
	if len(forecast) > 0 && len(forecast) < 24 {
		fullForecast := make([]types.ForecastTimeSlot, 24)
		for i := 0; i < 24; i++ {
			fullForecast[i].Hour = i
		}
		for _, f := range forecast {
			fullForecast[f.Hour] = f
		}
		forecast = fullForecast
	}

	return forecast, nil
}
