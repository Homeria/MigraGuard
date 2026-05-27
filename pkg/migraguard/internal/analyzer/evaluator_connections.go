package analyzer

import (
	"context"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// PeakConnectionEvaluator (Step 3): Predicts the maximum influx of connections.
type PeakConnectionEvaluator struct{}

func (e *PeakConnectionEvaluator) Evaluate(ctx context.Context, analysis types.AnalysisResult, metrics types.TableDynamicMetrics, report *types.RiskAnalysisReport, constants types.RiskConstants) error {
	lambdaPerMs := metrics.TPS / 1000.0
	report.PeakConnections = metrics.ActiveConnections + int(lambdaPerMs*report.BlockingTime)
	return nil
}
