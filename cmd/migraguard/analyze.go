package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/internal/db"
	"github.com/Homeria/MigraGuard/internal/reporter"
	"github.com/Homeria/MigraGuard/internal/service"
	"github.com/spf13/cobra"
)

var (
	analyzeDbString   string
	analyzeSqlitePath string
	analyzeOutput     string
)

func init() {
	rootCmd.AddCommand(analyzeCmd)

	// Analyze configuration flags with GlobalConfig defaults
	analyzeCmd.Flags().StringVar(&analyzeDbString, "db", GlobalConfig.Database.URL, "PostgreSQL connection string")
	analyzeCmd.Flags().StringVar(&analyzeSqlitePath, "sqlite", GlobalConfig.Database.SQLitePath, "Path to the local SQLite storage file")
	analyzeCmd.Flags().StringVarP(&analyzeOutput, "output", "o", "console", "Output format (console, markdown)")
}

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze [migration_file.sql]",
	Short: "Analyze a migration SQL file to evaluate risk instantly",
	Long: `Parses the given DDL migration script (AST), evaluates the potential risk
using the MigraGuard v3.1 baseline model, and reports the findings.
This command relies on data collected by the 'migraguard agent'.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		if analyzeOutput == "console" {
			fmt.Printf("🔍 Starting instant risk analysis for: %s\n", filePath)
		}

		if analyzeDbString == "" {
			fmt.Println("❌ Error: --db flag is required to connect to PostgreSQL.")
			os.Exit(1)
		}

		ctx := context.Background()

		// 1. Initialize Adapters (Infrastructure Layer)
		pg := db.NewPostgresAdapter(analyzeDbString)
		if err := pg.Connect(ctx); err != nil {
			fmt.Printf("❌ Failed to connect to PostgreSQL: %v\n", err)
			os.Exit(1)
		}
		defer pg.Close()

		sqlite, err := db.NewSQLiteAdapter(analyzeSqlitePath)
		if err != nil {
			fmt.Printf("❌ Failed to initialize SQLite: %v\n", err)
			os.Exit(1)
		}
		defer sqlite.Close()

		if analyzeOutput == "console" {
			fmt.Println("📡 Fetching baseline metrics from SQLite...")
		}

		// 2. Initialize and Run Service (Domain Layer)
		svc := service.NewAnalyzeService(pg, sqlite, GlobalConfig.Engine, Verbose)
		resp, err := svc.Run(ctx, service.AnalysisTask{SQLPath: filePath})
		if err != nil {
			fmt.Printf("❌ Analysis Failed: %v\n", err)
			os.Exit(1)
		}

		// 3. Report Results (Presentation Layer)
		if analyzeOutput == "markdown" {
			fmt.Println(reporter.MarkdownReport(resp.Results, resp.Reports))
		} else {
			fmt.Println("\n--- MigraGuard v3.1 Risk Analysis Report ---")
			for i, res := range resp.Results {
				report := resp.Reports[i]
				fmt.Printf("\n[Target Table: %s | Operation: %s]\n", res.TableName, res.Operation)
				fmt.Printf("  📊 Traffic Stats: Current=%.1f, Avg(1h)=%.1f, Peak(24h)=%.1f TPS\n", 
					report.CurrentTPS, report.AvgTPS1h, report.PeakTPS24h)
				fmt.Printf("  ✅ Physical DDL Time (T_ddl): %.2f ms\n", report.EstimatedDDLTime)
				fmt.Printf("  ✅ Estimated Block Time (T_block): %.2f ms\n", report.BlockingTime)
				fmt.Printf("  📡 Peak Connections (C_peak): %d\n", report.PeakConnections)
				fmt.Printf("  🚦 RISK LEVEL: [%s]\n", report.RiskLevel)

				if report.SafeWindow != "" {
					fmt.Printf("  💡 Tip: Deployment is 80%% safer at %s (Avg: %.1f TPS)\n", 
						report.SafeWindow, report.SafeWindowTPS)
				}
				if report.PermanentFailure {
					fmt.Println("  🚨 CRITICAL: Permanent system failure predicted! Check your mu_max and connection limits.")
				}
			}
			fmt.Println("\n------------------------------------------------")
		}

		// 4. Final Gatekeeping
		hasDanger := false
		for _, r := range resp.Reports {
			if r.RiskLevel == "Danger" {
				hasDanger = true
				break
			}
		}

		if hasDanger {
			if analyzeOutput == "console" {
				fmt.Println("🛑 Danger detected! Migration blocked.")
			}
			os.Exit(1)
		} else {
			if analyzeOutput == "console" {
				fmt.Println("✅ Analysis complete. No critical risks found.")
			}
		}
	},
}
