package reporter

import (
	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// Reporter is the interface for outputting analysis results in various formats.
type Reporter interface {
	Write(results []types.AnalysisResult, reports []*types.RiskAnalysisReport) error
}
