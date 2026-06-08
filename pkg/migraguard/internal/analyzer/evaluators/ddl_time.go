package evaluators

import (
	"context"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

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
