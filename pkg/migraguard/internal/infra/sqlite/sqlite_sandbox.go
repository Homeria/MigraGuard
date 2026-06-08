package sqlite

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

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

	// Initialize local pseudo-random generator to avoid global math/rand resource locks
	rng := rand.New(rand.NewSource(now.UnixNano()))

	tx, err := e.adapter.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	metricStmt, err := tx.Prepare(`INSERT INTO table_metrics (timestamp, table_name, table_size, replication_lag, active_connections, p99_time, tps, shared_blks_hit, shared_blks_read) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare table metrics statement: %w", err)
	}
	defer metricStmt.Close()

	workloadStmt, err := tx.Prepare(`INSERT INTO workload_snapshots (timestamp, query_id, query, calls, total_time, shared_blks_hit, shared_blks_read) VALUES (?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare workload snapshots statement: %w", err)
	}
	defer workloadStmt.Close()

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
		tps := e.profiler.CalculateTPS(t, startTime, scenario.History)

		// 2. Correlation-based Metrics
		p99 := scenario.PGState.P99TimeMS * math.Exp(tps/scenario.History.PeakTPS-1.0)
		if p99 > 5000.0 {
			p99 = 5000.0 // Upper-bound latency clipping to simulate realistic client-side timeouts
		}
		conns := int(float64(scenario.PGState.ActiveConnections) * (tps / scenario.History.PeakTPS))
		currentSize := scenario.PGState.TableSizeMB*1024*1024 + int64(i*1024)

		// 3. I/O Simulation (Cache Hit Ratio)
		// Base hit ratio 98%, drops slightly as TPS increases towards Peak
		hitRatio := 0.98 - 0.05*(tps/scenario.History.PeakTPS)
		if hitRatio < 0.85 {
			hitRatio = 0.85
		}

		// For the last point (Now), we force it to match the requested PGState exactly
		if isLastPoint {
			t = now // Use actual current time for the final snapshot
			tps = scenario.PGState.CurrentTPS
			p99 = scenario.PGState.P99TimeMS
			conns = scenario.PGState.ActiveConnections
			currentSize = scenario.PGState.TableSizeMB * 1024 * 1024
		}
		timestamp := t.Format("2006-01-02 15:04:05")

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
			// Assume each transaction touches ~20 blocks on average, scaled by the interval duration
			intervalSeconds := float64(scenario.History.IntervalMinutes * 60)
			totalBlocks := int64(tableTPS * intervalSeconds * 20)
			hitBlocks := int64(float64(totalBlocks) * hitRatio)
			readBlocks := totalBlocks - hitBlocks

			if _, err := metricStmt.Exec(timestamp, tableName, currentSize, scenario.PGState.ReplicationLagS, conns, p99, tableTPS, hitBlocks, readBlocks); err != nil {
				return err
			}

			// Add variety to queries in workload_snapshots
			queries := []string{
				fmt.Sprintf("SELECT * FROM %s WHERE id = ?", tableName),
				fmt.Sprintf("UPDATE %s SET updated_at = now() WHERE id = ?", tableName),
				fmt.Sprintf("INSERT INTO %s_audit (table_name, action) VALUES ('%s', 'change')", tableName, tableName),
			}

			for qIdx, qText := range queries {
				queryID := int64(1000 + (len(tables) * qIdx) + rng.Intn(10))
				// Split total table TPS across these 3 queries (50%, 30%, 20% distribution)
				share := 0.5
				if qIdx == 1 {
					share = 0.3
				} else if qIdx == 2 {
					share = 0.2
				}

				intervalSeconds := float64(scenario.History.IntervalMinutes * 60)
				calls := int64(tableTPS * intervalSeconds * share) // calls accumulated over the interval
				totalTime := float64(calls) * p99                  // total time in ms

				qHit := int64(float64(calls*20) * hitRatio)
				qRead := int64(calls*20) - qHit

				if _, err := workloadStmt.Exec(timestamp, queryID, qText, calls, totalTime, qHit, qRead); err != nil {
					return err
				}
			}
		}
	}

	return tx.Commit()
}
