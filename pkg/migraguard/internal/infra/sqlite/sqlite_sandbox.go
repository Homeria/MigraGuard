package sqlite

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// WorkloadProfiler is the interface for generating realistic traffic patterns.
type WorkloadProfiler interface {
	CalculateTPS(t time.Time, profile types.SQLiteHistoryProfile) float64
}

// DefaultWorkloadProfiler implements a rich daily/weekly pattern with sine waves and noise.
type DefaultWorkloadProfiler struct{}

func (p *DefaultWorkloadProfiler) CalculateTPS(t time.Time, profile types.SQLiteHistoryProfile) float64 {
	// A. Time-based normalization (0.0 - 1.0)
	hour := float64(t.Hour()) + float64(t.Minute())/60.0
	
	// 1. Daily Deterministic Fluctuation Factors (Seed based on Year-Month-Day)
	shift := 0.0
	volumeScale := 1.0
	skew := profile.AsymmetricSkew

	if profile.PeakShiftHours > 0 || profile.AsymmetricSkew != 0 {
		daySeed := int64(t.Year()*10000 + int(t.Month())*100 + t.Day())
		r := rand.New(rand.NewSource(daySeed))
		
		// Peak Shift Offset
		if profile.PeakShiftHours > 0 {
			shift = (r.Float64()*2.0 - 1.0) * profile.PeakShiftHours
		}
		
		// Daily Volume Scale: Modulates peak amplitude by +-15% per day
		volumeScale = 0.85 + r.Float64()*0.30 // [0.85, 1.15]
		
		// Daily Skewness Modulation: Modulates skewness by +-20% per day
		if profile.AsymmetricSkew != 0 {
			skew = profile.AsymmetricSkew * (0.8 + r.Float64()*0.4) // [0.8 * skew, 1.2 * skew]
		}
	}

	skewedHour := hour - shift
	for skewedHour < 0 {
		skewedHour += 24.0
	}
	for skewedHour >= 24.0 {
		skewedHour -= 24.0
	}

	// 2. Asymmetric Load Curve (Time Warping) with modulated daily skew
	warpedHour := skewedHour
	if skew != 0 {
		theta := skewedHour * math.Pi / 12.0
		thetaSkewed := theta + skew*math.Sin(theta)
		warpedHour = thetaSkewed * 12.0 / math.Pi
		for warpedHour < 0 {
			warpedHour += 24.0
		}
		for warpedHour >= 24.0 {
			warpedHour -= 24.0
		}
	}

	// 3. Daily Sine Wave (Peak at 14:00 and 20:00, modified by shifted & warped time)
	dailyPattern := 0.4*math.Sin((warpedHour-9)*math.Pi/12.0) + 0.3*math.Sin((warpedHour-18)*math.Pi/6.0) + 0.5
	
	// 4. Weekly Pattern (Weekend traffic is 40% lower)
	weeklyMult := 1.0
	if profile.WeeklyPattern && (t.Weekday() == time.Saturday || t.Weekday() == time.Sunday) {
		weeklyMult = 0.6
	}

	// 5. Special Events (Flash Sales, Maintenance)
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

	// 6. Combined Calculation with daily volumeScale
	finalTPS := (profile.BaseTPS + dailyPattern*(profile.PeakTPS-profile.BaseTPS)) * weeklyMult * eventMult * volumeScale
	
	// 7. Gaussian-like Noise
	noise := (rand.Float64()*2 - 1) * profile.NoiseVariance * finalTPS
	
	return math.Max(1.0, finalTPS + noise)
}

type SandboxEngine struct {
	adapter  *SQLiteAdapter
	profiler WorkloadProfiler
}

func NewSandboxEngine(adapter *SQLiteAdapter) *SandboxEngine {
	return &SandboxEngine{
		adapter:  adapter,
		profiler: &DefaultWorkloadProfiler{},
	}
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

	metricStmt, _ := tx.Prepare(`INSERT INTO table_metrics (timestamp, table_name, table_size, replication_lag, active_connections, p99_time, tps, shared_blks_hit, shared_blks_read) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	workloadStmt, _ := tx.Prepare(`INSERT INTO workload_snapshots (timestamp, query_id, query, calls, total_time, shared_blks_hit, shared_blks_read) VALUES (?, ?, ?, ?, ?, ?, ?)`)

	// Define tables to seed: TargetTable + auxiliary fintech tables for realistic background noise
	tables := []string{"account_balances", "inventory_stocks", "orders", "order_event_logs", "users", "products", "logs"}
	if scenario.TargetTable != "" {
		found := false
		for _, t := range tables {
			if t == scenario.TargetTable {
				found = true
				break
			}
		}
		if !found {
			tables = append(tables, scenario.TargetTable)
		}
	}

	for i := 0; i <= totalPoints; i++ {
		t := startTime.Add(time.Duration(i) * interval)
		isLastPoint := (i == totalPoints)

		// 1. Calculate Base Traffic using the profiler strategy
		tps := e.profiler.CalculateTPS(t, scenario.History)

		// 2. Correlation-based Metrics
		p99 := scenario.PGState.P99TimeMS * math.Exp(tps/scenario.History.PeakTPS-1.0)
		conns := int(float64(scenario.PGState.ActiveConnections) * (tps / scenario.History.PeakTPS))
		currentSize := scenario.PGState.TableSizeMB*1024*1024 + int64(i*1024)

		// 3. I/O Simulation (Cache Hit Ratio)
		// Base hit ratio 98%, drops slightly as TPS increases towards Peak
		hitRatio := 0.98 - 0.05*(tps/scenario.History.PeakTPS)
		if hitRatio < 0.85 { hitRatio = 0.85 }

		// For the last point (Now), we force it to match the requested PGState exactly
		if isLastPoint {
			t = now // Use actual current time for the final snapshot
			tps = scenario.PGState.CurrentTPS
			p99 = scenario.PGState.P99TimeMS
			conns = scenario.PGState.ActiveConnections
			currentSize = scenario.PGState.TableSizeMB * 1024 * 1024
		}

		for _, tableName := range tables {
			// Each table has a different load share
			tableTPS := tps
			if tableName == "order_event_logs" {
				tableTPS *= 1.5
			} else if tableName == "account_balances" {
				tableTPS *= 0.3
			} else if tableName != scenario.TargetTable {
				tableTPS *= 0.5 // Secondary tables have less load
			}

			// Generate Blocks (Hit/Read)
			// Assume each transaction touches ~20 blocks on average
			totalBlocks := int64(tableTPS * 60 * 20)
			hitBlocks := int64(float64(totalBlocks) * hitRatio)
			readBlocks := totalBlocks - hitBlocks

			if _, err := metricStmt.Exec(t, tableName, currentSize, scenario.PGState.ReplicationLagS, conns, p99, tableTPS, hitBlocks, readBlocks); err != nil {
				return err
			}

			// Add variety to queries in workload_snapshots
			queries := []string{
				fmt.Sprintf("SELECT * FROM %s WHERE id = ?", tableName),
				fmt.Sprintf("UPDATE %s SET updated_at = now() WHERE id = ?", tableName),
				fmt.Sprintf("INSERT INTO %s_audit (table_name, action) VALUES ('%s', 'change')", tableName, tableName),
			}

			for qIdx, qText := range queries {
				queryID := int64(1000 + (len(tables) * qIdx) + rand.Intn(10))
				// Split total table TPS across these 3 queries (50%, 30%, 20% distribution)
				share := 0.5
				if qIdx == 1 { share = 0.3 } else if qIdx == 2 { share = 0.2 }
				
				calls := int64(tableTPS * 60 * share) // calls per interval
				totalTime := float64(calls) * p99      // total time in ms
				
				qHit := int64(float64(calls*20) * hitRatio)
				qRead := int64(calls*20) - qHit
				
				if _, err := workloadStmt.Exec(t, queryID, qText, calls, totalTime, qHit, qRead); err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}
