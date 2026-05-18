package app

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

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
	var forecastReports []*types.ForecastReport
	var validResults []types.AnalysisResult

	for _, res := range results {
		if err := s.pg.CheckTableSchemaPresence(ctx, res.TableName, res.Columns); err != nil {
			fmt.Printf("[INFO] Skipping schema validation [%s]: %v\n", res.TableName, err)
			continue
		}

		// 1. Current Risk Analysis
		report, err := riskEngine.AnalyzeRisk(ctx, res)
		if err != nil {
			return nil, migraErrors.WrapWithTable(err, "AnalyzeService.Run", res.TableName, "analysis failed")
		}

		// 2. Predictive Forecast Analysis (if requested)
		if task.Forecast && s.sqlite != nil {
			forecastData, err := s.sqlite.Get24HourTrafficForecast(ctx, res.TableName)
			if err == nil && len(forecastData) > 0 {
				fReport, err := riskEngine.AnalyzeForecast(ctx, res, forecastData)
				if err == nil {
					forecastReports = append(forecastReports, fReport)
				}
			}
		}

		validResults = append(validResults, res)
		finalReports = append(finalReports, report)
	}

	if len(validResults) == 0 {
		return nil, migraErrors.Wrap(migraErrors.ErrTableNotFound, "AnalyzeService.Run", "no valid targets")
	}

	// Always export CSV if forecast was successful for visualization
	if len(forecastReports) > 0 {
		_ = s.exportForecastCSV(forecastReports)
	}

	return &types.AnalysisResponse{
		Results:         validResults,
		Reports:         finalReports,
		ForecastReports: forecastReports,
	}, nil
}

func (s *AnalyzeService) exportForecastCSV(forecasts []*types.ForecastReport) error {
	f, err := os.Create("predictive_forecast.csv")
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// Header
	writer.Write([]string{"Hour", "ExpectedTPS", "MinTPS", "MaxTPS", "ExpectedP99", "RiskScore", "RiskLevel", "IsSafeWindow", "IsBestHour"})

	for _, report := range forecasts {
		for _, slot := range report.Timeline {
			row := []string{
				strconv.Itoa(slot.Hour),
				strconv.FormatFloat(slot.ExpectedTPS, 'f', 2, 64),
				strconv.FormatFloat(slot.MinTPS, 'f', 2, 64),
				strconv.FormatFloat(slot.MaxTPS, 'f', 2, 64),
				strconv.FormatFloat(slot.ExpectedP99, 'f', 2, 64),
				strconv.FormatFloat(slot.RiskScore, 'f', 2, 64),
				slot.RiskLevel,
				strconv.FormatBool(slot.IsSafeWindow),
				strconv.FormatBool(slot.Hour == report.BestHour),
			}
			writer.Write(row)
		}
	}
	return writer.Error()
}
