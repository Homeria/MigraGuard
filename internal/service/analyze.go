package service

import (
	"context"
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/internal/db"
	"github.com/Homeria/MigraGuard/internal/engine"
	"github.com/Homeria/MigraGuard/internal/errors"
	"github.com/Homeria/MigraGuard/internal/parser"
)

// AnalyzeService handles the end-to-end migration analysis pipeline.
type AnalyzeService struct {
	pg        db.PostgresClient
	sqlite    db.SQLiteClient
	constants engine.RiskConstants
	Verbose   bool
}

// NewAnalyzeService creates a new instance of AnalyzeService.
func NewAnalyzeService(pg db.PostgresClient, sqlite db.SQLiteClient, constants engine.RiskConstants, verbose bool) *AnalyzeService {
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
		return nil, errors.Wrap(err, "AnalyzeService.Run", "failed to read SQL file")
	}

	// 2. Static Analysis (AST Parsing)
	results, err := parser.ParseSQL(string(sqlContent))
	if err != nil {
		return nil, errors.Wrap(err, "AnalyzeService.Run", "SQL parser error")
	}

	if len(results) == 0 {
		return nil, errors.ErrInvalidSQL
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
			if errors.Is(err, errors.ErrTableNotFound) || errors.Is(err, errors.ErrColumnNotFound) {
				if s.Verbose {
					fmt.Printf("[DEBUG] Validation failed for table '%s': %v. Skipping.\n", res.TableName, err)
				}
				continue
			}
			return nil, errors.WrapWithTable(err, "AnalyzeService.Run", res.TableName, "schema validation error")
		}

		// [L31~L35] Risk Analysis
		report, err := riskEngine.AnalyzeRisk(ctx, res)
		if err != nil {
			return nil, errors.WrapWithTable(err, "AnalyzeService.Run", res.TableName, "risk analysis execution failed")
		}

		validResults = append(validResults, res)
		finalReports = append(finalReports, report)
	}

	if len(validResults) == 0 {
		return nil, errors.Wrap(errors.ErrTableNotFound, "AnalyzeService.Run", "no valid tables found for analysis")
	}

	return &AnalysisResponse{
		Results: validResults,
		Reports: finalReports,
	}, nil
}
