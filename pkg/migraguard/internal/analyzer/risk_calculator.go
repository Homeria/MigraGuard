package analyzer

import (
	"context"
	"fmt"
	"math"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// DefaultRiskConstants returns standard threshold values for the risk engine.
func DefaultRiskConstants() types.RiskConstants {
	return types.RiskConstants{
		// Performance
		DiskIO:   100 * 1024 * 1024,
		TMeta:    100.0,
		MuMax:    5000.0,
		CMax:     100,
		TTimeout: 5000.0,

		// Thresholds
		ThresholdDanger:  80.0,
		ThresholdWarning: 50.0,

		// Algorithm Weights
		AvgMultiplier:    1.2,
		PeakMultiplier:   0.8,
		ConcurrentImpact: 0.1,
		MiddleImpact:     0.5,

		// Base Risk Values
		BaseAccessExclusiveMeta: 30.0,
		BaseAccessExclusiveFull: 85.0,
		BaseExclusive:           50.0,
		BaseShare:               20.0,
	}
}

// RiskEngine calculates DDL risk scores based on workload metrics.
type RiskEngine struct {
	pg        types.PostgresClient
	sqlite    types.SQLiteClient
	constants types.RiskConstants
	Verbose   bool
}

// NewRiskEngine initializes a new RiskEngine.
func NewRiskEngine(pg types.PostgresClient, sqlite types.SQLiteClient, constants types.RiskConstants) *RiskEngine {
	return &RiskEngine{
		pg:        pg,
		sqlite:    sqlite,
		constants: constants,
		Verbose:   false,
	}
}

// AnalyzeRisk performs the 5-step risk analysis using configurable weights.
func (e *RiskEngine) AnalyzeRisk(ctx context.Context, analysis types.AnalysisResult) (*types.RiskAnalysisReport, error) {
	metrics, err := e.pg.FetchTableDynamicMetrics(ctx, analysis.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch real-time metrics: %w", err)
	}

	report := &types.RiskAnalysisReport{
		ActiveConns: metrics.ActiveConnections,
		TableSize:   metrics.TableSize,
	}

	if e.sqlite != nil {
		report.CurrentTPS, _ = e.sqlite.GetRecentTPSByDelta(analysis.TableName)
		baseline, _ := e.sqlite.GetTableBaselineStatistics(analysis.TableName)
		if baseline != nil {
			report.AvgTPS1h = baseline.AvgTPS_1h
			report.PeakTPS24h = baseline.PeakTPS_24h
		}
		report.SafeWindow, report.SafeWindowTPS, _ = e.sqlite.IdentifySafestDeploymentWindow()
		report.TopQueries, _ = e.sqlite.GetTopHeavyQueries(3)

		// Use configurable multipliers for conservative TPS estimation
		weightedAvg := report.AvgTPS1h * e.constants.AvgMultiplier
		weightedPeak := report.PeakTPS24h * e.constants.PeakMultiplier
		maxTPS := report.CurrentTPS
		report.TPSSource = "Real-time"
		if weightedAvg > maxTPS {
			maxTPS = weightedAvg
			report.TPSSource = fmt.Sprintf("1h-Avg (x%.1f)", e.constants.AvgMultiplier)
		}
		if weightedPeak > maxTPS {
			maxTPS = weightedPeak
			report.TPSSource = fmt.Sprintf("24h-Peak (x%.1f)", e.constants.PeakMultiplier)
		}
		report.BaseTPS = maxTPS
		metrics.TPS = maxTPS
	}

	if analysis.RewriteRequired {
		report.EstimatedDDLTime = (float64(metrics.TableSize) / float64(e.constants.DiskIO)) * 1000.0
	} else {
		report.EstimatedDDLTime = e.constants.TMeta
	}

	// Use configurable lock impact factors
	lockImpact := 1.0
	if analysis.LockLevel <= types.LockLevelShareUpdateExcl {
		lockImpact = e.constants.ConcurrentImpact
	} else if analysis.LockLevel < types.LockLevelAccessExclusive {
		lockImpact = e.constants.MiddleImpact
	}

	report.BlockingTime = (metrics.P99Time + report.EstimatedDDLTime + (metrics.ReplicationLag * 1000.0)) * lockImpact

	lambdaPerMs := metrics.TPS / 1000.0
	report.PeakConnections = metrics.ActiveConnections + int(lambdaPerMs*report.BlockingTime)

	if metrics.TPS >= e.constants.MuMax {
		report.PermanentFailure = true
		report.RecoveryTime = math.Inf(1)
	} else {
		excessiveConns := float64(report.PeakConnections - e.constants.CMax)
		if excessiveConns > 0 {
			recoveryRatePerMs := (e.constants.MuMax - metrics.TPS) / 1000.0
			report.RecoveryTime = excessiveConns / recoveryRatePerMs
		} else {
			report.RecoveryTime = 0
		}
	}

	report.RiskScore = (float64(report.PeakConnections) / float64(e.constants.CMax)) * 100.0

	// Apply configurable base risk scores
	baseRisk := 0.0
	switch analysis.LockLevel {
	case types.LockLevelAccessExclusive:
		if analysis.MetadataOnly {
			baseRisk = e.constants.BaseAccessExclusiveMeta
		} else {
			baseRisk = e.constants.BaseAccessExclusiveFull
		}
	case types.LockLevelExclusive:
		baseRisk = e.constants.BaseExclusive
	case types.LockLevelShare:
		baseRisk = e.constants.BaseShare
	}

	if report.RiskScore < baseRisk {
		report.RiskScore = baseRisk
	}

	// Evaluate level using configurable thresholds
	report.RiskLevel = e.EvaluateLevel(report.RiskScore, report.PermanentFailure)

	return report, nil
}
