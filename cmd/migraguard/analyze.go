package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/internal/db"
	"github.com/Homeria/MigraGuard/internal/engine"
	"github.com/Homeria/MigraGuard/internal/parser"
	"github.com/spf13/cobra"
)

var (
	analyzeDbString  string
	analyzeSqlitePath string
)

func init() {
	rootCmd.AddCommand(analyzeCmd)

	// Analyze configuration flags
	analyzeCmd.Flags().StringVar(&analyzeDbString, "db", "", "PostgreSQL connection string")
	analyzeCmd.Flags().StringVar(&analyzeSqlitePath, "sqlite", "migraguard.db", "Path to the local SQLite storage file")
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
		fmt.Printf("🔍 Starting instant risk analysis for: %s\n", filePath)

		if analyzeDbString == "" {
			fmt.Println("❌ Error: --db flag is required to connect to PostgreSQL.")
			os.Exit(1)
		}

		ctx := context.Background()

		// [L01] Load SQL file
		sqlContent, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("❌ Failed to read file: %v\n", err)
			os.Exit(1)
		}

		// [L02] Static Analysis (AST)
		results, err := parser.ParseSQL(string(sqlContent))
		if err != nil {
			fmt.Printf("❌ Parser Error: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Println("⚠️ No valid DDL operations found in the provided SQL file.")
			return
		}

		// [L03] Initialize Adapters
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

		fmt.Println("📡 Fetching baseline metrics from SQLite...")

		// [L05] Initialize Risk Engine
		riskEngine := engine.NewRiskEngine(pg, sqlite, engine.DefaultRiskConstants())

		fmt.Println("\n--- MigraGuard v3.1 Risk Analysis Report ---")

		hasDanger := false
		for _, res := range results {
			fmt.Printf("\n[Target Table: %s | Operation: %s]\n", res.TableName, res.Operation)

			// [L31~L35] Analyze Risk using Baseline Data
			report, err := riskEngine.AnalyzeRisk(ctx, res)
			if err != nil {
				fmt.Printf("  ❌ Risk Analysis Error: %v\n", err)
				continue
			}

			// [L41] Report results
			fmt.Printf("  ✅ Physical DDL Time (T_ddl): %.2f ms\n", report.EstimatedDDLTime)
			fmt.Printf("  ✅ Estimated Block Time (T_block): %.2f ms\n", report.BlockingTime)
			fmt.Printf("  📡 Peak Connections (C_peak): %d\n", report.PeakConnections)
			fmt.Printf("  🚦 RISK LEVEL: [%s]\n", report.RiskLevel)

			if report.PermanentFailure {
				fmt.Println("  🚨 CRITICAL: Permanent system failure predicted! Check your mu_max and connection limits.")
			}

			if report.RiskLevel == "Danger" {
				hasDanger = true
			}
		}

		fmt.Println("\n------------------------------------------------")

		// [L42] Gatekeeping (Exit Code 1 for Danger)
		if hasDanger {
			fmt.Println("🛑 Danger detected! Migration blocked.")
			os.Exit(1)
		} else {
			fmt.Println("✅ Analysis complete. No critical risks found.")
		}
	},
}
