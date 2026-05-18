package reporter

import (
	"fmt"
	"strings"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// MarkdownReporter generates detailed markdown reports for migration risk.
type MarkdownReporter struct{}

// NewMarkdownReporter creates a new MarkdownReporter instance.
func NewMarkdownReporter() *MarkdownReporter {
	return &MarkdownReporter{}
}

// Write outputs analysis results in markdown format.
func (r *MarkdownReporter) Write(results []types.AnalysisResult, reports []*types.RiskAnalysisReport, forecasts []*types.ForecastReport) error {
	fmt.Println("# [REPORT] MigraGuard Risk Analysis Report")

	for i, res := range results {
		report := reports[i]

		fmt.Printf("\n## [TARGET] Analysis for `%s`\n", res.TableName)

		fmt.Println("\n### [CODE] Raw SQL Statement")
		fmt.Printf("```sql\n%s\n```\n", res.RawQuery)

		fmt.Println("\n### [SUMMARY] Analysis Summary")
		fmt.Printf("- **Risk Level**: **%s**\n", report.RiskLevel)
		fmt.Printf("- **Risk Score**: `%.2f / 100`\n", report.RiskScore)
		fmt.Printf("- **Operation**: `%s` (Rewrite required: `%v`)\n", res.Operation, res.RewriteRequired)

		fmt.Println("\n### [METRICS] Queuing Model Metrics")
		fmt.Println("| Metric | Value | Description |")
		fmt.Println("| :--- | :--- | :--- |")
		fmt.Printf("| Est. DDL Time (T_ddl) | %.2f ms | Estimated time for DDL execution |\n", report.EstimatedDDLTime)
		fmt.Printf("| Est. Blocking Time (T_block) | %.2f ms | Time services may be blocked |\n", report.BlockingTime)
		fmt.Printf("| Predicted Peak Conns (C_peak) | %d | Predicted max connections during DDL |\n", report.PeakConnections)
		fmt.Printf("| Est. Recovery Time (T_rec) | %.2f ms | Time for system recovery |\n", report.RecoveryTime)

		fmt.Println("\n### [TRAFFIC] Baseline Analysis")
		fmt.Printf("- **Baseline TPS (Lambda)**: `%.2f` (Source: %s)\n", report.BaseTPS, report.TPSSource)
		fmt.Println("| Field | Value |")
		fmt.Println("| :--- | :--- |")
		fmt.Printf("| Current TPS | %.2f |\n", report.CurrentTPS)
		fmt.Printf("| 1h Average TPS | %.2f |\n", report.AvgTPS1h)
		fmt.Printf("| 24h Peak TPS | %.2f |\n", report.PeakTPS24h)
		fmt.Printf("| Active Connections | %d |\n", report.ActiveConns)
		fmt.Printf("| Table Size | %.2f MB |\n", float64(report.TableSize)/(1024*1024))

		if len(report.TopQueries) > 0 {
			fmt.Println("\n### [QUERIES] Top Heavy Queries")
			fmt.Println("| Impact | Query Preview | ID |")
			fmt.Println("| :--- | :--- | :--- |")
			for _, q := range report.TopQueries {
				cleanQuery := strings.ReplaceAll(q.QueryText, "\n", " ")
				if len(cleanQuery) > 80 {
					cleanQuery = cleanQuery[:77] + "..."
				}
				fmt.Printf("| `%.1f%%` | `%s` | %d |\n", q.Impact, cleanQuery, q.QueryID)
			}
		}

		// Display forecast if available
		for _, f := range forecasts {
			if f.TableName == res.TableName {
				fmt.Println("\n### [FORECAST] Predictive 24-Hour Analysis")
				fmt.Printf("- **Optimal Execution Window**: **%02d:00**\n", f.BestHour)
				fmt.Println("- **Visualization**: Generated predictive risk heatmap CSV.")
			}
		}

		fmt.Println("\n### [ADVICE] Conclusion and Recommendation")
		if report.RiskLevel == "Safe" {
			fmt.Println("> [SAFE] Deployment is expected to have minimal impact.")
		} else {
			fmt.Printf("> [WARNING] Recommend deploying during the Golden window: **%s** (Est. %.1f TPS).\n", report.SafeWindow, report.SafeWindowTPS)
		}
		fmt.Println("\n---")
	}
	return nil
}
