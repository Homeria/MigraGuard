package reporter

import (
	"fmt"

	"github.com/Homeria/MigraGuard/internal/engine"
	"github.com/Homeria/MigraGuard/internal/parser"
)

// ConsoleReporter outputs analysis results to the terminal with emojis and colors.
type ConsoleReporter struct{}

func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{}
}

func (r *ConsoleReporter) Write(results []parser.AnalysisResult, reports []*engine.RiskAnalysisReport) error {
	fmt.Println("\n--- 🛡️ MigraGuard v3.1 Risk Analysis Report ---")

	for i, res := range results {
		report := reports[i]
		fmt.Printf("\n[Target Table: %s | Operation: %s]\n", res.TableName, res.Operation)
		fmt.Printf("  📊 Traffic Stats: Current=%.1f, Avg(1h)=%.1f, Peak(24h)=%.1f TPS\n",
			report.CurrentTPS, report.AvgTPS1h, report.PeakTPS24h)
		fmt.Printf("  ✅ Physical DDL Time (T_ddl): %.2f ms\n", report.EstimatedDDLTime)
		fmt.Printf("  ✅ Estimated Block Time (T_block): %.2f ms\n", report.BlockingTime)
		fmt.Printf("  📡 Peak Connections (C_peak): %d\n", report.PeakConnections)
		fmt.Printf("  🚦 RISK LEVEL: [%s]\n", report.RiskLevel)

		if report.SafeWindow != "" && report.SafeWindowTPS < report.AvgTPS1h {
			safetyGain := (1.0 - (report.SafeWindowTPS / report.AvgTPS1h)) * 100.0
			fmt.Printf("  💡 Tip: 배포를 %s로 미루면 현재보다 약 %.0f%% 더 안전합니다. (예상 TPS: %.1f)\n",
				report.SafeWindow, safetyGain, report.SafeWindowTPS)
		}
		if report.PermanentFailure {
			fmt.Println("  🚨 CRITICAL: Permanent system failure predicted! Check your mu_max and connection limits.")
		}
	}
	fmt.Println("\n------------------------------------------------")
	return nil
}
