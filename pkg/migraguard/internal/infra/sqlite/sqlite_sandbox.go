package sqlite

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

type SandboxEngine struct {
	adapter *SQLiteAdapter
}

func NewSandboxEngine(adapter *SQLiteAdapter) *SandboxEngine {
	return &SandboxEngine{adapter: adapter}
}

func (e *SandboxEngine) SeedScenario(scenario types.SimulationScenario) error {
	now := time.Now()
	// All research data starts exactly 7 days ago at midnight for consistency
	startTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -scenario.History.Days)
	
	interval := time.Duration(scenario.History.IntervalMinutes) * time.Minute
	totalPoints := (scenario.History.Days * 24 * 60) / scenario.History.IntervalMinutes

	tx, err := e.adapter.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	metricStmt, _ := tx.Prepare(`INSERT INTO table_metrics (timestamp, table_name, table_size, replication_lag, active_connections, p99_time, tps) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	workloadStmt, _ := tx.Prepare(`INSERT INTO workload_snapshots (timestamp, query_id, query, calls, total_time) VALUES (?, ?, ?, ?, ?)`)

	tables := []string{"account_balances", "inventory_stocks", "orders", "order_event_logs"}

	for i := 0; i <= totalPoints; i++ {
		t := startTime.Add(time.Duration(i) * interval)
		
		// 1. Calculate Base Traffic (Daily + Weekly Cycle)
		tps := e.calculateRichTPS(t, scenario.History)

		// 2. Correlation-based Metrics
		// P99 grows exponentially with TPS: Base_P99 * e^(TPS/Peak)
		p99 := scenario.PGState.P99TimeMS * math.Exp(tps/scenario.History.PeakTPS-1.0)
		// Active Conns grows linearly with TPS
		conns := int(float64(scenario.PGState.ActiveConnections) * (tps / scenario.History.PeakTPS))
		// Table Size grows slowly over time
		currentSize := scenario.PGState.TableSizeMB*1024*1024 + int64(i*1024) 

		for _, tableName := range tables {
			// Each table has a different load share
			tableTPS := tps
			if tableName == "order_event_logs" { tableTPS *= 1.5 } // Logs are more active
			if tableName == "account_balances" { tableTPS *= 0.3 } // Balances are hit less than browses

			if _, err := metricStmt.Exec(t, tableName, currentSize, scenario.PGState.ReplicationLagS, conns, p99, tableTPS); err != nil {
				return err
			}

			// Add corresponding queries to workload_snapshots
			queryID := int64(1000 + rand.Intn(100))
			if _, err := workloadStmt.Exec(t, queryID, fmt.Sprintf("UPDATE %s SET updated_at = now()", tableName), int64(tableTPS*60), tableTPS*p99); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (e *SandboxEngine) calculateRichTPS(t time.Time, profile types.SQLiteHistoryProfile) float64 {
	// A. Time-based normalization (0.0 - 1.0)
	hour := float64(t.Hour()) + float64(t.Minute())/60.0
	
	// 1. Daily Sine Wave (Peak at 14:00 and 20:00)
	dailyPattern := 0.4*math.Sin((hour-9)*math.Pi/12.0) + 0.3*math.Sin((hour-18)*math.Pi/6.0) + 0.5
	
	// 2. Weekly Pattern (Weekend traffic is 40% lower)
	weeklyMult := 1.0
	if profile.WeeklyPattern && (t.Weekday() == time.Saturday || t.Weekday() == time.Sunday) {
		weeklyMult = 0.6
	}

	// 3. Special Events (Flash Sales, Maintenance)
	eventMult := 1.0
	for _, event := range profile.Events {
		eventStart := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).
			AddDate(0, 0, -profile.Days+event.StartDay).
			Add(time.Duration(event.StartHour) * time.Hour)
		eventEnd := eventStart.Add(time.Duration(event.DurationH) * time.Hour)

		if t.After(eventStart) && t.Before(eventEnd) {
			eventMult = event.Multiplier
			break
		}
	}

	// 4. Combined Calculation
	finalTPS := (profile.BaseTPS + dailyPattern*(profile.PeakTPS-profile.BaseTPS)) * weeklyMult * eventMult
	
	// 5. Gaussian-like Noise (10% variance)
	noise := (rand.Float64()*2 - 1) * profile.NoiseVariance * finalTPS
	
	return math.Max(1.0, finalTPS + noise)
}
