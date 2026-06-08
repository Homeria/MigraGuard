package analyzer

import (
	"context"
	"fmt"
	"math"

	"github.com/Homeria/MigraGuard/pkg/migraguard/internal/analyzer/evaluators"
	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// DefaultRiskConstants returns standard threshold values for the risk engine.
func DefaultRiskConstants() types.RiskConstants {
	return types.RiskConstants{
		// Performance
		DiskIO:   100 * 1024 * 1024,
		TMeta:    100.0,
		MuMax:    5000.0,
		CMax:     100,
		TTimeout: 5000.0,

		// Thresholds
		ThresholdDanger:  80.0,
		ThresholdWarning: 50.0,

		// Algorithm Weights
		AvgMultiplier:    1.2,
		PeakMultiplier:   0.8,
		ConcurrentImpact: 0.1,
		MiddleImpact:     0.5,

		// Base Risk Values
		BaseAccessExclusiveMeta: 30.0,
		BaseAccessExclusiveFull: 85.0,
		BaseExclusive:           50.0,
		BaseShare:               20.0,
	}
}

// RiskEngine calculates DDL risk scores based on workload metrics.
type RiskEngine struct {
	pg         types.PostgresClient
	sqlite     types.SQLiteClient
	constants  types.RiskConstants
	evaluators []evaluators.StepEvaluator
	Verbose    bool
}

// NewRiskEngine initializes a new RiskEngine with default evaluators.
func NewRiskEngine(pg types.PostgresClient, sqlite types.SQLiteClient, constants types.RiskConstants) *RiskEngine {
	return &RiskEngine{
		pg:        pg,
		sqlite:    sqlite,
		constants: constants,
		evaluators: []evaluators.StepEvaluator{
			&evaluators.DDLTimeEvaluator{},
			&evaluators.BlockingTimeEvaluator{},
			&evaluators.PeakConnectionEvaluator{},
			&evaluators.RecoveryTimeEvaluator{},
			&evaluators.RiskScoreEvaluator{},
		},
		Verbose: false,
	}
}

// AnalyzeRisk performs the 5-step risk analysis using configurable weights and strategies.
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
		currentTPS, err := e.sqlite.GetRecentTPSByDelta(ctx, analysis.TableName)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate recent TPS: %w", err)
		}
		report.CurrentTPS = currentTPS

		baseline, err := e.sqlite.GetTableBaselineStatistics(ctx, analysis.TableName)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch table baseline statistics: %w", err)
		}
		if baseline != nil {
			report.AvgTPS1h = baseline.AvgTPS_1h
			report.PeakTPS24h = baseline.PeakTPS_24h
		}

		safeWindow, safeWindowTPS, err := e.sqlite.IdentifySafestDeploymentWindow(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to identify safe deployment window: %w", err)
		}
		report.SafeWindow = safeWindow
		report.SafeWindowTPS = safeWindowTPS

		topQueries, err := e.sqlite.GetTopHeavyQueries(ctx, 3)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch top heavy queries: %w", err)
		}
		report.TopQueries = topQueries

		// Use configurable multipliers for conservative TPS estimation
		weightedAvg := report.AvgTPS1h * e.constants.AvgMultiplier
		weightedPeak := report.PeakTPS24h * e.constants.PeakMultiplier
		maxTPS := report.CurrentTPS
		report.TPSSource = "Real-time"
		if weightedAvg > maxTPS {
			maxTPS = weightedAvg
			report.TPSSource = fmt.Sprintf("1h-Avg (x%.1f)", e.constants.AvgMultiplier)
		}
		if weightedPeak > maxTPS {
			maxTPS = weightedPeak
			report.TPSSource = fmt.Sprintf("24h-Peak (x%.1f)", e.constants.PeakMultiplier)
		}
		report.BaseTPS = maxTPS
		metrics.TPS = maxTPS
	}

	// Execute evaluation strategies (Phase 2 Architectural Refinement)
	for _, evaluator := range e.evaluators {
		if err := evaluator.Evaluate(ctx, analysis, *metrics, report, e.constants); err != nil {
			return nil, err
		}
	}

	return report, nil
}

// AnalyzeForecast simulates DDL risk across a 24-hour forecasted traffic profile.
func (e *RiskEngine) AnalyzeForecast(ctx context.Context, analysis types.AnalysisResult, forecast []types.ForecastTimeSlot) (*types.ForecastReport, error) {
	// Fetch static metrics (table size) once from PG to avoid repeated I/O
	metrics, err := e.pg.FetchTableDynamicMetrics(ctx, analysis.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch base metrics for forecast: %w", err)
	}

	report := &types.ForecastReport{
		TableName: analysis.TableName,
		Timeline:  make([]types.ForecastTimeSlot, 0, len(forecast)),
		BestHour:  -1,
	}

	minScore := 9999.0
	minTPS := 999999.0

	for _, slot := range forecast {
		// Create a virtual snapshot for this hour
		virtualMetrics := *metrics
		virtualMetrics.TPS = slot.ExpectedTPS
		virtualMetrics.P99Time = slot.ExpectedP99

		// Run the 5-step risk model in-memory
		tempReport := &types.RiskAnalysisReport{
			TableSize:   metrics.TableSize,
			ActiveConns: metrics.ActiveConnections,
			BaseTPS:     slot.ExpectedTPS,
		}

		for _, evaluator := range e.evaluators {
			if err := evaluator.Evaluate(ctx, analysis, virtualMetrics, tempReport, e.constants); err != nil {
				return nil, err
			}
		}

		// Update slot results
		slot.RiskScore = tempReport.RiskScore
		slot.RiskLevel = tempReport.RiskLevel
		slot.IsSafeWindow = tempReport.RiskScore < e.constants.ThresholdWarning

		// Best Hour Selection with TPS Tie-breaker
		// 1. If lower risk score found, update best hour
		// 2. If risk scores are equal (using epsilon for float stability), choose lower TPS
		isLowerScore := slot.RiskScore < (minScore - 0.001)
		isEqualScore := math.Abs(slot.RiskScore-minScore) < 0.001
		isLowerTPS := slot.ExpectedTPS < minTPS

		if isLowerScore || (isEqualScore && isLowerTPS) {
			minScore = slot.RiskScore
			minTPS = slot.ExpectedTPS
			report.BestHour = slot.Hour
		}

		report.Timeline = append(report.Timeline, slot)
	}

	return report, nil
}
