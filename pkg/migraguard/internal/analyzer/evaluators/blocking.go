package evaluators

import (
	"context"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

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
