package engine

import (
	"context"
	"fmt"
	"math"

	"github.com/Homeria/MigraGuard/internal/db"
	"github.com/Homeria/MigraGuard/internal/parser"
)

// RiskConstants represents infrastructure-specific constants for risk calculation.
type RiskConstants struct {
	DiskIO    int64   `mapstructure:"disk_io"`    // Disk_IO: Bytes/sec
	TMeta     float64 `mapstructure:"t_meta"`     // T_meta (ms)
	MuMax     float64 `mapstructure:"mu_max"`     // Mu_max: Max system TPS
	CMax      int     `mapstructure:"c_max"`      // C_max: Connection limit
	TTimeout  float64 `mapstructure:"t_timeout"`  // T_timeout (ms)
}

// DefaultRiskConstants provides standard values for general environments.
func DefaultRiskConstants() RiskConstants {
	return RiskConstants{
		DiskIO:   100 * 1024 * 1024, // 100MB/s
		TMeta:    100.0,            // 100ms
		MuMax:    5000.0,           // 5000 TPS
		CMax:     1000,             // 1000 Conns
		TTimeout: 5000.0,           // 5s
	}
}

// RiskEngine computes the risk of a DDL operation using the MigraGuard v3.1 baseline model.
type RiskEngine struct {
	pg        *db.PostgresAdapter
	sqlite    *db.SQLiteAdapter
	constants RiskConstants
}

// RiskAnalysisReport contains the detailed results of the risk evaluation.
type RiskAnalysisReport struct {
	RiskScore        float64 // Final risk percentage (%)
	EstimatedDDLTime float64 // T_ddl (ms)
	BlockingTime     float64 // T_block (ms)
	PeakConnections  int     // C_peak
	RecoveryTime     float64 // T_rec (ms)
	PermanentFailure bool    // Lambda >= Mu_max
	RiskLevel        string  // Danger, Warning, Safe

	// v3.1 Additional Context
	CurrentTPS    float64
	AvgTPS1h      float64
	PeakTPS24h    float64
	SafeWindow    string  // Recommended deployment hour (e.g., "03:00")
	SafeWindowTPS float64
}

// NewRiskEngine creates a new RiskEngine instance.
func NewRiskEngine(pg *db.PostgresAdapter, sqlite *db.SQLiteAdapter, constants RiskConstants) *RiskEngine {
	return &RiskEngine{
		pg:        pg,
		sqlite:    sqlite,
		constants: constants,
	}
}

// AnalyzeRisk performs the 5-step risk assessment for a given DDL analysis result.
// [L-B02, L-B03] Uses SQLite baseline metrics to evaluate risk instantly.
func (e *RiskEngine) AnalyzeRisk(ctx context.Context, analysis parser.AnalysisResult) (*RiskAnalysisReport, error) {
	// 1. Get Dynamic Metrics from Postgres (Size, Conns, etc.)
	metrics, err := e.pg.GetTableDynamicMetrics(ctx, analysis.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dynamic metrics: %w", err)
	}

	report := &RiskAnalysisReport{}

	// [Step 0] v3.1 Baseline Analytics & Weighted TPS
	if e.sqlite != nil {
		// Real-time TPS (delta from last 1-min snapshots)
		realtimeTPS, _ := e.sqlite.GetRecentTPSDelta(analysis.TableName)
		report.CurrentTPS = realtimeTPS
		
		// Historical stats from table_metrics
		baseline, _ := e.sqlite.GetTableBaselineStats(analysis.TableName)
		if baseline != nil {
			report.AvgTPS1h = baseline.AvgTPS_1h
			report.PeakTPS24h = baseline.PeakTPS_24h
		}

		// Deployment Window Recommendation
		hour, avgTPS, _ := e.sqlite.GetSafeWindow()
		report.SafeWindow = hour
		report.SafeWindowTPS = avgTPS

		// [L34] Multi-Weighted TPS Calculation (Conservative Approach)
		// Lambda = Max(Current, Avg_1h * 1.2, Peak_24h * 0.8)
		metrics.TPS = math.Max(report.CurrentTPS, math.Max(report.AvgTPS1h*1.2, report.PeakTPS24h*0.8))
	}

	// [Step 1] Estimated DDL Time (T_ddl)
	if analysis.RewriteRequired {
		report.EstimatedDDLTime = (float64(metrics.TableSize) / float64(e.constants.DiskIO)) * 1000.0
	} else {
		report.EstimatedDDLTime = e.constants.TMeta
	}

	// [Step 2] Total Blocking Time (T_block)
	report.BlockingTime = metrics.P99Time + report.EstimatedDDLTime + (metrics.ReplicationLag * 1000.0)

	// [Step 3] Peak Connections (C_peak)
	lambdaPerMs := metrics.TPS / 1000.0
	report.PeakConnections = metrics.ActiveConnections + int(lambdaPerMs * report.BlockingTime)

	// [Step 4] Recovery Time (T_rec)
	if metrics.TPS >= e.constants.MuMax {
		report.PermanentFailure = true
		report.RecoveryTime = math.Inf(1)
	} else {
		recoveryNumerator := float64(report.PeakConnections - e.constants.CMax)
		recoveryDenominator := (e.constants.MuMax - metrics.TPS) / 1000.0 // per ms
		if recoveryNumerator > 0 {
			report.RecoveryTime = recoveryNumerator / recoveryDenominator
		} else {
			report.RecoveryTime = 0
		}
	}

	// [Step 5] Risk Score Calculation
	report.RiskScore = (float64(report.PeakConnections) / float64(e.constants.CMax)) * 100.0

	// Risk Level Classification
	if report.RiskScore >= 90.0 || report.PermanentFailure {
		report.RiskLevel = "Danger"
	} else if report.RiskScore >= 60.0 {
		report.RiskLevel = "Warning"
	} else {
		report.RiskLevel = "Safe"
	}

	return report, nil
}
