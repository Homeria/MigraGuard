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
	analyzeForecast   bool
	noHeader          bool
)

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVar(&analyzeDbString, "db", "", "Target PostgreSQL connection string")
	analyzeCmd.Flags().StringVar(&analyzeSqlitePath, "sqlite", "", "Local metric storage (SQLite) path")
	analyzeCmd.Flags().StringVarP(&analyzeSandbox, "sandbox", "s", "", "Path to SQLite simulation sandbox (Offline Mode)")
	analyzeCmd.Flags().StringVarP(&analyzeOutput, "output", "o", "console", "Output format (console, markdown, csv)")
	analyzeCmd.Flags().BoolVar(&analyzeForecast, "forecast", false, "Enable 24-hour predictive risk forecasting")
	analyzeCmd.Flags().BoolVar(&noHeader, "no-header", false, "Do not print CSV header (only for --output csv)")
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

		// 2. Initialize client based on mode
		var mg *migraguard.Client
		var err error

		mgCfg := migraguard.Config{
			PostgresDSN: finalDB,
			SQLitePath:  finalSQLite,
			Verbose:     Verbose,
			Risk:        GlobalConfig.Risk.ToRiskConstants(),
		}

		if analyzeSandbox != "" {
			mg, err = migraguard.NewSandboxClient(analyzeSandbox, mgCfg)
			if err == nil {
				fmt.Fprintf(os.Stderr, "[INFO] Offline Mode: Using sandbox data from %s\n", analyzeSandbox)
			}
		} else {
			mg, err = migraguard.NewLiveClient(mgCfg)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] MigraGuard initialization failed: %v\n", err)
			os.Exit(1)
		}
		defer mg.Close()

		// 4. Run analysis
		resp, err := mg.Analyze(ctx, filePath, analyzeForecast)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Analysis failed: %v\n", err)
			os.Exit(1)
		}

		// 5. Output results
		var rpt reporter.Reporter
		switch analyzeOutput {
		case "markdown":
			rpt = reporter.NewMarkdownReporter()
		case "csv":
			csvRpt := reporter.NewCSVReporter()
			csvRpt.SQLFile = filePath
			if analyzeSandbox != "" {
				csvRpt.Scenario = analyzeSandbox
			} else {
				csvRpt.Scenario = "live"
			}

			if !noHeader {
				// Use the writer from CSVReporter to print header to the same output
				csvRpt.Writer.Write(csvRpt.GetHeader())
				csvRpt.Writer.Flush()
			}
			rpt = csvRpt
		default:
			rpt = reporter.NewConsoleReporter()
		}

		if err := rpt.Write(resp.Results, resp.Reports); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Report generation failed: %v\n", err)
			os.Exit(1)
		}

		// 6. Gatekeeping
		for _, r := range resp.Reports {
			if r.RiskLevel == "Danger" {
				fmt.Fprintln(os.Stderr, "\n[DANGER] High-risk migration detected. Deployment pipeline forcibly blocked.")
				os.Exit(1)
			}
		}
	},
}
