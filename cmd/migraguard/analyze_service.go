package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/pkg/migraguard"
	"github.com/Homeria/MigraGuard/pkg/migraguard/reporter"
	"github.com/spf13/cobra"
)

var (
	analyzeDbString   string
	analyzeSqlitePath string
	analyzeSandbox    string
	analyzeOutput     string
)

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVar(&analyzeDbString, "db", "", "Target PostgreSQL connection string")
	analyzeCmd.Flags().StringVar(&analyzeSqlitePath, "sqlite", "", "Local metric storage (SQLite) path")
	analyzeCmd.Flags().StringVarP(&analyzeSandbox, "sandbox", "s", "", "Path to SQLite simulation sandbox (Offline Mode)")
	analyzeCmd.Flags().StringVarP(&analyzeOutput, "output", "o", "console", "Output format (console, markdown)")
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze [migration_file.sql]",
	Short: "Analyze migration SQL file and evaluate risk",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		ctx := context.Background()

		// 1. Resolve connection info
		finalDB := analyzeDbString
		if analyzeSandbox == "" {
			if finalDB == "" {
				finalDB = GlobalConfig.Database.Postgres
			}
		} else {
			// In sandbox mode, we don't need a real PG connection during New()
			finalDB = ""
		}

		finalSQLite := analyzeSqlitePath
		if analyzeSandbox != "" {
			finalSQLite = analyzeSandbox
		} else {
			if finalSQLite == "" {
				finalSQLite = GlobalConfig.Database.SQLite
			}
			if finalSQLite == "" {
				finalSQLite = "migraguard.db"
			}
		}

		// 2. Initialize client
		mg, err := migraguard.New(migraguard.Config{
			PostgresDSN: finalDB,
			SQLitePath:  finalSQLite,
			Verbose:     Verbose,
			Risk:        GlobalConfig.Risk.ToRiskConstants(),
		})
		if err != nil {
			fmt.Printf("[ERROR] MigraGuard initialization failed: %v\n", err)
			os.Exit(1)
		}
		defer mg.Close()

		// 3. Switch to Sandbox Mode if requested
		if analyzeSandbox != "" {
			if err := mg.UseSandbox(analyzeSandbox); err != nil {
				fmt.Printf("[ERROR] Failed to load sandbox: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("[INFO] Offline Mode: Using sandbox data from %s\n", analyzeSandbox)
		}

		// 4. Run analysis
		resp, err := mg.Analyze(ctx, filePath)
		if err != nil {
			fmt.Printf("[ERROR] Analysis failed: %v\n", err)
			os.Exit(1)
		}

		// 5. Output results
		var rpt reporter.Reporter
		if analyzeOutput == "markdown" {
			rpt = reporter.NewMarkdownReporter()
		} else {
			rpt = reporter.NewConsoleReporter()
		}

		if err := rpt.Write(resp.Results, resp.Reports); err != nil {
			fmt.Printf("[ERROR] Report generation failed: %v\n", err)
			os.Exit(1)
		}

		// 6. Gatekeeping
		for _, r := range resp.Reports {
			if r.RiskLevel == "Danger" {
				fmt.Println("\n[DANGER] High-risk migration detected. Deployment pipeline forcibly blocked.")
				os.Exit(1)
			}
		}
	},
}
