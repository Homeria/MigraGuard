package service

import (
	"context"
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/internal/db"
	"github.com/Homeria/MigraGuard/internal/engine"
	"github.com/Homeria/MigraGuard/internal/parser"
)

// AnalyzeService handles the end-to-end migration analysis pipeline.
type AnalyzeService struct {
	pg        *db.PostgresAdapter
	sqlite    *db.SQLiteAdapter
	constants engine.RiskConstants
	Verbose   bool
}

// NewAnalyzeService creates a new instance of AnalyzeService.
func NewAnalyzeService(pg *db.PostgresAdapter, sqlite *db.SQLiteAdapter, constants engine.RiskConstants, verbose bool) *AnalyzeService {
	return &AnalyzeService{
		pg:        pg,
		sqlite:    sqlite,
		constants: constants,
		Verbose:   verbose,
	}
}

// AnalysisTask represents the input for an analysis operation.
type AnalysisTask struct {
	SQLPath string
}

// AnalysisResponse holds the results of the analysis for the reporter.
type AnalysisResponse struct {
	Results []parser.AnalysisResult
	Reports []*engine.RiskAnalysisReport
}

// Run executes the full analysis pipeline: Load -> Parse -> Validate -> Analyze.
func (s *AnalyzeService) Run(ctx context.Context, task AnalysisTask) (*AnalysisResponse, error) {
	// 1. Load SQL file
	sqlContent, err := os.ReadFile(task.SQLPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read SQL file: %w", err)
	}

	// 2. Static Analysis (AST Parsing)
	results, err := parser.ParseSQL(string(sqlContent))
	if err != nil {
		return nil, fmt.Errorf("SQL parser error: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no valid DDL operations found in SQL file")
	}

	// 3. Initialize Risk Engine
	riskEngine := engine.NewRiskEngine(s.pg, s.sqlite, s.constants)
	riskEngine.Verbose = s.Verbose

	var finalReports []*engine.RiskAnalysisReport
	var validResults []parser.AnalysisResult

	// 4. Validation & Analysis Loop
	for _, res := range results {
		// [L61] Schema Validation
		if err := s.pg.ValidateSchema(ctx, res.TableName, res.Columns); err != nil {
			if s.Verbose {
				fmt.Printf("[DEBUG] Schema validation failed for table '%s': %v. Skipping.\n", res.TableName, err)
			}
			continue
		}

		// [L31~L35] Risk Analysis
		report, err := riskEngine.AnalyzeRisk(ctx, res)
		if err != nil {
			return nil, fmt.Errorf("risk analysis error for table %s: %w", res.TableName, err)
		}

		validResults = append(validResults, res)
		finalReports = append(finalReports, report)
	}

	if len(validResults) == 0 {
		return nil, fmt.Errorf("all identified tables failed schema validation")
	}

	return &AnalysisResponse{
		Results: validResults,
		Reports: finalReports,
	}, nil
}
