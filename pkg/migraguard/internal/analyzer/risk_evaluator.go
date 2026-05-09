package analyzer

import (
	"context"
	"math"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// StepEvaluator is the strategy interface for each risk assessment stage.
type StepEvaluator interface {
	Evaluate(ctx context.Context, analysis types.AnalysisResult, metrics types.TableDynamicMetrics, report *types.RiskAnalysisReport, constants types.RiskConstants) error
}

// DDLTimeEvaluator (Step 1): Predicts the duration of the DDL execution.
type DDLTimeEvaluator struct{}

func (e *DDLTimeEvaluator) Evaluate(ctx context.Context, analysis types.AnalysisResult, metrics types.TableDynamicMetrics, report *types.RiskAnalysisReport, constants types.RiskConstants) error {
	if analysis.RewriteRequired {
		report.EstimatedDDLTime = (float64(metrics.TableSize) / float64(constants.DiskIO)) * 1000.0
	} else {
		report.EstimatedDDLTime = constants.TMeta
	}
	return nil
}

// BlockingTimeEvaluator (Step 2): Calculates potential service blocking time.
type BlockingTimeEvaluator struct{}

func (e *BlockingTimeEvaluator) Evaluate(ctx context.Context, analysis types.AnalysisResult, metrics types.TableDynamicMetrics, report *types.RiskAnalysisReport, constants types.RiskConstants) error {
	lockImpact := 1.0
	if analysis.LockLevel <= types.LockLevelShareUpdateExcl {
		lockImpact = constants.ConcurrentImpact
	} else if analysis.LockLevel < types.LockLevelAccessExclusive {
		lockImpact = constants.MiddleImpact
	}

	report.BlockingTime = (metrics.P99Time + report.EstimatedDDLTime + (metrics.ReplicationLag * 1000.0)) * lockImpact
	return nil
}

// PeakConnectionEvaluator (Step 3): Predicts the maximum influx of connections.
type PeakConnectionEvaluator struct{}

func (e *PeakConnectionEvaluator) Evaluate(ctx context.Context, analysis types.AnalysisResult, metrics types.TableDynamicMetrics, report *types.RiskAnalysisReport, constants types.RiskConstants) error {
	lambdaPerMs := metrics.TPS / 1000.0
	report.PeakConnections = metrics.ActiveConnections + int(lambdaPerMs*report.BlockingTime)
	return nil
}

// RecoveryTimeEvaluator (Step 4): Evaluates cost for system normalization.
type RecoveryTimeEvaluator struct{}

func (e *RecoveryTimeEvaluator) Evaluate(ctx context.Context, analysis types.AnalysisResult, metrics types.TableDynamicMetrics, report *types.RiskAnalysisReport, constants types.RiskConstants) error {
	if metrics.TPS >= constants.MuMax {
		report.PermanentFailure = true
		report.RecoveryTime = math.Inf(1)
	} else {
		excessiveConns := float64(report.PeakConnections - constants.CMax)
		if excessiveConns > 0 {
			recoveryRatePerMs := (constants.MuMax - metrics.TPS) / 1000.0
			report.RecoveryTime = excessiveConns / recoveryRatePerMs
		} else {
			report.RecoveryTime = 0
		}
	}
	return nil
}

// RiskScoreEvaluator (Step 5): Determines the final risk score and level.
type RiskScoreEvaluator struct{}

func (e *RiskScoreEvaluator) Evaluate(ctx context.Context, analysis types.AnalysisResult, metrics types.TableDynamicMetrics, report *types.RiskAnalysisReport, constants types.RiskConstants) error {
	report.RiskScore = (float64(report.PeakConnections) / float64(constants.CMax)) * 100.0

	baseRisk := 0.0
	switch analysis.LockLevel {
	case types.LockLevelAccessExclusive:
		if analysis.MetadataOnly {
			baseRisk = constants.BaseAccessExclusiveMeta
		} else {
			baseRisk = constants.BaseAccessExclusiveFull
		}
	case types.LockLevelExclusive:
		baseRisk = constants.BaseExclusive
	case types.LockLevelShare:
		baseRisk = constants.BaseShare
	}

	if report.RiskScore < baseRisk {
		report.RiskScore = baseRisk
	}

	// Determine final level
	if report.PermanentFailure || report.RiskScore >= constants.ThresholdDanger {
		report.RiskLevel = "Danger"
	} else if report.RiskScore >= constants.ThresholdWarning {
		report.RiskLevel = "Warning"
	} else {
		report.RiskLevel = "Safe"
	}
	return nil
}
