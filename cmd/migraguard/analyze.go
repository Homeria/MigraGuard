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

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze [migration_file.sql]",
	Short: "Analyze a migration SQL file to evaluate risk and report results",
	Long: `Parses the given DDL migration script (AST), evaluates the potential risk
using the MigraGuard v3.0 queuing model, and reports the findings to the terminal.
Exits with status 1 if the risk level is 'Danger'.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		fmt.Printf("🔍 Starting static and dynamic analysis for: %s\n", filePath)

		// 1. Read SQL File
		sqlContent, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("❌ Failed to read file: %v\n", err)
			os.Exit(1)
		}

		// 2. Parse SQL (AST Analysis)
		results, err := parser.ParseSQL(string(sqlContent))
		if err != nil {
			fmt.Printf("❌ Parser Error: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Println("⚠️ No DDL statements found in the provided SQL file.")
			return
		}

		// 3. Setup DB Adapters (Using Env Vars)
		ctx := context.Background()
		pgUrl := os.Getenv("DATABASE_URL")
		if pgUrl == "" {
			fmt.Println("❌ DATABASE_URL environment variable is not set.")
			os.Exit(1)
		}

		pg := db.NewPostgresAdapter(pgUrl)
		if err := pg.Connect(ctx); err != nil {
			fmt.Printf("❌ Postgres Connection Error: %v\n", err)
			os.Exit(1)
		}
		defer pg.Close()

		sqlitePath := os.Getenv("SQLITE_PATH")
		if sqlitePath == "" {
			sqlitePath = "migraguard.db"
		}
		sqlite, err := db.NewSQLiteAdapter(sqlitePath)
		if err != nil {
			fmt.Printf("❌ SQLite Error: %v\n", err)
			os.Exit(1)
		}
		defer sqlite.Close()

		// 4. Initialize Risk Engine v3.0
		riskEngine := engine.NewRiskEngine(pg, sqlite, engine.DefaultRiskConstants())

		fmt.Println("\n--- 🛡️  MigraGuard v3.0 Risk Analysis Report ---")
		
		hasDanger := false
		for _, res := range results {
			fmt.Printf("\n[Target Table: %s | Operation: %s]\n", res.TableName, res.Operation)
			
			// 5. Calculate Risk Score using Model v3.0
			report, err := riskEngine.AnalyzeRisk(ctx, res)
			if err != nil {
				fmt.Printf("  ⚠️ Risk Calculation Failed: %v\n", err)
				continue
			}

			// 6. Quantitative Reporting
			fmt.Printf("  - Risk Level:       [%s]\n", report.RiskLevel)
			fmt.Printf("  - Risk Score:       %.2f%%\n", report.RiskScore)
			fmt.Printf("  - Estimated DDL:    %.2f ms\n", report.EstimatedDDLTime)
			fmt.Printf("  - Blocking Time:    %.2f ms\n", report.BlockingTime)
			fmt.Printf("  - Peak Connections: %d conns\n", report.PeakConnections)
			fmt.Printf("  - Recovery Time:    %.2f ms\n", report.RecoveryTime)
			
			if report.PermanentFailure {
				fmt.Println("  🚨 CRITICAL: Permanent system failure predicted!")
			}
			
			if report.RiskLevel == "Danger" {
				hasDanger = true
			}
		}

		fmt.Println("\n------------------------------------------------")
		
		// 7. Gatekeeping: Control Process Exit
		if hasDanger {
			fmt.Println("❌ STATUS: REJECTED (High Risk Detected)")
			os.Exit(1)
		} else {
			fmt.Println("✅ STATUS: APPROVED (Safe to Deploy)")
		}
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
