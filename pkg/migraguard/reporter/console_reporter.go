package reporter

import (
	"fmt"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// ConsoleReporter outputs analysis results to the console.
type ConsoleReporter struct{}

// NewConsoleReporter creates a new ConsoleReporter instance.
func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{}
}

// Write prints analysis reports to the console with ANSI colors.
func (r *ConsoleReporter) Write(results []types.AnalysisResult, reports []*types.RiskAnalysisReport, forecasts []*types.ForecastReport) error {
	for i, res := range results {
		report := reports[i]

		fmt.Println("\n" + r.drawHeader(fmt.Sprintf("[REPORT] MigraGuard Risk Analysis: %s", res.TableName)))

		fmt.Printf(" [CODE] Raw SQL: %s\n", r.truncate(res.RawQuery, 500))
		fmt.Printf(" [STATUS] Level=%s | Risk Score=%.2f%%\n", r.colorLevel(report.RiskLevel), report.RiskScore)

		opInfo := res.Operation
		if res.SubOperation != "" {
			opInfo = fmt.Sprintf("%s (%s)", res.Operation, res.SubOperation)
		}
		fmt.Printf(" [TASK] Type=%s | Rewrite=%v | Columns=%v\n", opInfo, res.RewriteRequired, res.Columns)

		fmt.Println("\n [TRAFFIC] Traffic Data")
		fmt.Printf("  - Baseline TPS (Lambda): %.2f (%s)\n", report.BaseTPS, report.TPSSource)
		fmt.Printf("  - Real-time TPS: %.2f | 1h Avg: %.2f | 24h Peak: %.2f\n",
			report.CurrentTPS, report.AvgTPS1h, report.PeakTPS24h)
		fmt.Printf("  - Active Conns: %d | Table Size: %.2f MB\n",
			report.ActiveConns, float64(report.TableSize)/(1024*1024))

		fmt.Println("\n [METRICS] 5-Step Risk Metrics")
		fmt.Printf("  - Step 1 [T_ddl]   Estimated DDL Time: %.2f ms\n", report.EstimatedDDLTime)
		fmt.Printf("  - Step 2 [T_block] Estimated Blocking Time: %.2f ms\n", report.BlockingTime)
		fmt.Printf("  - Step 3 [C_peak]  Predicted Peak Connections: %d\n", report.PeakConnections)
		fmt.Printf("  - Step 4 [T_rec]   Estimated Recovery Time: %.2f ms (Failure: %v)\n",
			report.RecoveryTime, report.PermanentFailure)

		if len(report.TopQueries) > 0 {
			fmt.Println("\n [QUERIES] Top Resource Heavy Queries")
			for _, q := range report.TopQueries {
				fmt.Printf("  - [%.1f%%] %s (ID: %d)\n", q.Impact, r.truncate(q.QueryText, 60), q.QueryID)
			}
		}

		fmt.Println("\n [ADVICE] Recommendation")
		if report.RiskLevel == "Safe" {
			fmt.Println("  [SAFE] Deployment looks safe under current workload.")
		} else {
			fmt.Printf("  [WARNING] Consider delaying until the golden window: [%s] (Est. %.1f TPS)\n",
				report.SafeWindow, report.SafeWindowTPS)
		}

		// Display forecast summary if available
		for _, f := range forecasts {
			if f.TableName == res.TableName {
				fmt.Printf("\n [FORECAST] 24-Hour Prediction Summary\n")
				fmt.Printf("  - Recommended Golden Window: %02d:00\n", f.BestHour)
				fmt.Printf("  - Predictive CSV generated for visualization.\n")
			}
		}

		fmt.Println(r.drawFooter())
	}
	return nil
}

func (r *ConsoleReporter) drawHeader(title string) string {
	line := "================================================================================"
	return fmt.Sprintf("%s\n %s\n%s", line, title, line)
}

func (r *ConsoleReporter) drawFooter() string {
	return "================================================================================"
}

func (r *ConsoleReporter) colorLevel(level string) string {
	switch level {
	case "Danger":
		return "\033[31m" + level + "\033[0m"
	case "Warning":
		return "\033[33m" + level + "\033[0m"
	case "Safe":
		return "\033[32m" + level + "\033[0m"
	default:
		return level
	}
}

func (r *ConsoleReporter) truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit-3] + "..."
}
