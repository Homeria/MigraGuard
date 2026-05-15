package app

import (
	"context"
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/pkg/migraguard/internal/analyzer"
	migraErrors "github.com/Homeria/MigraGuard/pkg/migraguard/internal/shared/errors"
	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// AnalyzeService orchestrates the migration analysis pipeline.
type AnalyzeService struct {
	pg        types.PostgresClient
	sqlite    types.SQLiteClient
	constants types.RiskConstants
	Verbose   bool
}

// NewAnalyzeService initializes the analyze service.
func NewAnalyzeService(pg types.PostgresClient, sqlite types.SQLiteClient, constants types.RiskConstants, verbose bool) *AnalyzeService {
	return &AnalyzeService{
		pg:        pg,
		sqlite:    sqlite,
		constants: constants,
		Verbose:   verbose,
	}
}

// AnalysisTask defines the input for an analysis job.
type AnalysisTask struct {
	SQLPath  string
	Forecast bool
}

// Run executes the analysis workflow.
func (s *AnalyzeService) Run(ctx context.Context, task AnalysisTask) (*types.AnalysisResponse, error) {
	sqlContent, err := os.ReadFile(task.SQLPath)
	if err != nil {
		return nil, migraErrors.Wrap(err, "AnalyzeService.Run", "failed to read SQL file")
	}

	results, err := analyzer.ParseSQL(string(sqlContent))
	if err != nil {
		return nil, migraErrors.Wrap(err, "AnalyzeService.Run", "SQL parsing error")
	}
	if len(results) == 0 {
		return nil, migraErrors.ErrInvalidSQL
	}

	riskEngine := analyzer.NewRiskEngine(s.pg, s.sqlite, s.constants)
	riskEngine.Verbose = s.Verbose
	var finalReports []*types.RiskAnalysisReport
	var validResults []types.AnalysisResult

	for _, res := range results {
		if err := s.pg.CheckTableSchemaPresence(ctx, res.TableName, res.Columns); err != nil {
			fmt.Printf("[INFO] Skipping schema validation [%s]: %v\n", res.TableName, err)
			continue
		}

		report, err := riskEngine.AnalyzeRisk(ctx, res)
		if err != nil {
			return nil, migraErrors.WrapWithTable(err, "AnalyzeService.Run", res.TableName, "analysis failed")
		}

		validResults = append(validResults, res)
		finalReports = append(finalReports, report)
	}

	if len(validResults) == 0 {
		return nil, migraErrors.Wrap(migraErrors.ErrTableNotFound, "AnalyzeService.Run", "no valid targets")
	}

	return &types.AnalysisResponse{
		Results: validResults,
		Reports: finalReports,
	}, nil
}
