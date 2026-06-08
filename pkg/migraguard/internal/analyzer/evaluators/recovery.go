package evaluators

import (
	"context"
	"math"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

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
