package reporter

import (
	"github.com/Homeria/MigraGuard/internal/engine"
	"github.com/Homeria/MigraGuard/internal/parser"
)

// Reporter defines the interface for displaying analysis results.
type Reporter interface {
	Write(results []parser.AnalysisResult, reports []*engine.RiskAnalysisReport) error
}
