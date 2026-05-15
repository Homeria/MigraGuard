package reporter

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// CSVReporter outputs analysis results in a flat CSV format for research data processing.
type CSVReporter struct {
	Writer   *csv.Writer
	Scenario string
	SQLFile  string
}

// NewCSVReporter initializes a new CSV reporter.
func NewCSVReporter() *CSVReporter {
	return &CSVReporter{
		Writer: csv.NewWriter(os.Stdout),
	}
}

// Write outputs the quantitative risk metrics to the configured writer.
func (r *CSVReporter) Write(results []types.AnalysisResult, reports []*types.RiskAnalysisReport, forecasts []*types.ForecastReport) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	// Standard Report
	for i, res := range results {
		report := reports[i]
		
		row := []string{
			now,                                      // Timestamp
			r.Scenario,                               // Scenario (Metadata)
			r.SQLFile,                                // SQLFile (Metadata)
			res.TableName,                            // TableName
			res.Operation,                            // Operation
			strconv.Itoa(res.LockLevel),               // LockLevel
			strconv.FormatBool(res.RewriteRequired),  // RewriteRequired
			strconv.FormatFloat(report.RiskScore, 'f', 2, 64), // RiskScore
			report.RiskLevel,                         // RiskLevel
			strconv.FormatFloat(report.EstimatedDDLTime, 'f', 2, 64), // T_ddl
			strconv.FormatFloat(report.BlockingTime, 'f', 2, 64),     // T_block
			strconv.Itoa(report.PeakConnections),     // C_peak
			strconv.FormatFloat(report.RecoveryTime, 'f', 2, 64),     // T_rec
			strconv.FormatFloat(report.BaseTPS, 'f', 2, 64),          // BaseTPS
			report.TPSSource,                          // TPSSource
			strconv.FormatInt(report.TableSize, 10),   // TableSize
		}

		if err := r.Writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}
	r.Writer.Flush()

	// Forecast Report (Exported to a separate file for visualization script)
	if len(forecasts) > 0 {
		if err := r.writeForecastCSV(forecasts); err != nil {
			return fmt.Errorf("failed to write forecast CSV: %w", err)
		}
	}

	return r.Writer.Error()
}

func (r *CSVReporter) writeForecastCSV(forecasts []*types.ForecastReport) error {
	f, err := os.Create("predictive_forecast.csv")
	if err != nil {
		return err
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// Header
	writer.Write([]string{"Hour", "ExpectedTPS", "ExpectedP99", "RiskScore", "RiskLevel", "IsSafeWindow", "IsBestHour"})

	for _, report := range forecasts {
		for _, slot := range report.Timeline {
			row := []string{
				strconv.Itoa(slot.Hour),
				strconv.FormatFloat(slot.ExpectedTPS, 'f', 2, 64),
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

// GetHeader returns the standard CSV header for research data.
func (r *CSVReporter) GetHeader() []string {
	return []string{
		"Timestamp", "Scenario", "SQLFile", "TableName", "Operation", "LockLevel", "RewriteRequired",
		"RiskScore", "RiskLevel", "T_ddl", "T_block", "C_peak", "T_rec",
		"BaseTPS", "TPSSource", "TableSize",
	}
}
