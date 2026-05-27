package analyzer

import (
	"context"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// StepEvaluator is the strategy interface for each risk assessment stage.
type StepEvaluator interface {
	Evaluate(ctx context.Context, analysis types.AnalysisResult, metrics types.TableDynamicMetrics, report *types.RiskAnalysisReport, constants types.RiskConstants) error
}
