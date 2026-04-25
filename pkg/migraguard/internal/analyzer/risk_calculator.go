package analyzer

import (
	"context"
	"fmt"
	"math"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// DefaultRiskConstants returns standard threshold values.
func DefaultRiskConstants() types.RiskConstants {
	return types.RiskConstants{
		DiskIO:   100 * 1024 * 1024,
		TMeta:    100.0,
		MuMax:    5000.0,
		CMax:     100,
		TTimeout: 5000.0,
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

// AnalyzeRisk performs the 5-step risk analysis.
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

		weightedAvg := report.AvgTPS1h * 1.2
		weightedPeak := report.PeakTPS24h * 0.8
		maxTPS := report.CurrentTPS
		report.TPSSource = "Real-time"
		if weightedAvg > maxTPS {
			maxTPS = weightedAvg
			report.TPSSource = "1h-Avg (+20%)"
		}
		if weightedPeak > maxTPS {
			maxTPS = weightedPeak
			report.TPSSource = "24h-Peak (-20%)"
		}
		report.BaseTPS = maxTPS
		metrics.TPS = maxTPS
	}

	if analysis.RewriteRequired {
		report.EstimatedDDLTime = (float64(metrics.TableSize) / float64(e.constants.DiskIO)) * 1000.0
	} else {
		report.EstimatedDDLTime = e.constants.TMeta
	}

	lockImpact := 1.0
	if analysis.LockLevel <= types.LockLevelShareUpdateExcl {
		lockImpact = 0.1
	} else if analysis.LockLevel < types.LockLevelAccessExclusive {
		lockImpact = 0.5
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

	baseRisk := 0.0
	switch analysis.LockLevel {
	case types.LockLevelAccessExclusive:
		if analysis.MetadataOnly {
			baseRisk = 30.0
		} else {
			baseRisk = 85.0
		}
	case types.LockLevelExclusive:
		baseRisk = 50.0
	case types.LockLevelShare:
		baseRisk = 20.0
	}

	if report.RiskScore < baseRisk {
		report.RiskScore = baseRisk
	}

	report.RiskLevel = EvaluateLevel(report.RiskScore, report.PermanentFailure)

	return report, nil
}
