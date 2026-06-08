package evaluators

import (
	"context"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

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
