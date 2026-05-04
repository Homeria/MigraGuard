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
	Writer *csv.Writer
}

// NewCSVReporter initializes a new CSV reporter.
func NewCSVReporter() *CSVReporter {
	return &CSVReporter{
		Writer: csv.NewWriter(os.Stdout),
	}
}

// Write outputs the quantitative risk metrics to the configured writer.
func (r *CSVReporter) Write(results []types.AnalysisResult, reports []*types.RiskAnalysisReport) error {
	// 1. Header (Only written if needed, usually handled by CLI or scripts to avoid duplicates in append mode)
	// For raw SDK use, we'll provide a way to include it or just write data rows.
	
	now := time.Now().Format("2006-01-02 15:04:05")

	for i, res := range results {
		report := reports[i]
		
		row := []string{
			now,                                      // Timestamp
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
	return r.Writer.Error()
}

// GetHeader returns the standard CSV header for research data.
func (r *CSVReporter) GetHeader() []string {
	return []string{
		"Timestamp", "TableName", "Operation", "LockLevel", "RewriteRequired",
		"RiskScore", "RiskLevel", "T_ddl", "T_block", "C_peak", "T_rec",
		"BaseTPS", "TPSSource", "TableSize",
	}
}
