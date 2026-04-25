package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/pkg/migraguard"
	"github.com/Homeria/MigraGuard/pkg/migraguard/reporter"
	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
	"github.com/spf13/cobra"
)

// analyze command flags
var (
	analyzeDbString   string // Postgres connection string
	analyzeSqlitePath string // SQLite file path
	analyzeOutput     string // Output format (console/markdown)
)

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVar(&analyzeDbString, "db", "", "Target PostgreSQL connection string (required)")
	analyzeCmd.Flags().StringVar(&analyzeSqlitePath, "sqlite", "", "Local metric storage (SQLite) path")
	analyzeCmd.Flags().StringVarP(&analyzeOutput, "output", "o", "console", "Output format (console, markdown)")
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze [migration_file.sql]",
	Short: "Immediately analyze migration SQL file and evaluate risk",
	Long: `Parses the provided DDL script and evaluates the potential deployment risk 
based on real-time and historical traffic data.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		ctx := context.Background()

		// 1. Resolve PostgreSQL DSN (Priority: Flag > GlobalConfig > Default)
		finalDB := analyzeDbString
		if finalDB == "" {
			finalDB = GlobalConfig.Database.Postgres
		}
		if finalDB == "" {
			finalDB = "postgres://user:pass@localhost:5432/postgres"
		}

		// 2. Resolve SQLite Path
		finalSQLite := analyzeSqlitePath
		if finalSQLite == "" {
			finalSQLite = GlobalConfig.Database.SQLite
		}
		if finalSQLite == "" {
			finalSQLite = "migraguard.db"
		}

		// 3. [Black Box] MigraGuard client initialization
		mg, err := migraguard.New(migraguard.Config{
			PostgresDSN: finalDB,
			SQLitePath:  finalSQLite,
			Verbose:     Verbose,
			Risk: types.RiskConstants{
				DiskIO:                  GlobalConfig.Risk.DiskIO,
				MuMax:                   GlobalConfig.Risk.MuMax,
				CMax:                    GlobalConfig.Risk.CMax,
				TTimeout:                GlobalConfig.Risk.TTimeout,
				TMeta:                   GlobalConfig.Risk.TMeta,
				ThresholdDanger:         GlobalConfig.Risk.Thresholds.Danger,
				ThresholdWarning:        GlobalConfig.Risk.Thresholds.Warning,
				AvgMultiplier:           GlobalConfig.Risk.Weights.AvgMultiplier,
				PeakMultiplier:          GlobalConfig.Risk.Weights.PeakMultiplier,
				ConcurrentImpact:        GlobalConfig.Risk.Weights.ConcurrentImpact,
				MiddleImpact:            GlobalConfig.Risk.Weights.MiddleImpact,
				BaseAccessExclusiveMeta: GlobalConfig.Risk.Weights.BaseAccessExclusiveMeta,
				BaseAccessExclusiveFull: GlobalConfig.Risk.Weights.BaseAccessExclusiveFull,
				BaseExclusive:           GlobalConfig.Risk.Weights.BaseExclusive,
				BaseShare:               GlobalConfig.Risk.Weights.BaseShare,
			},
		})
		if err != nil {
			fmt.Printf("[ERROR] MigraGuard initialization failed: %v\n", err)
			os.Exit(1)
		}
		defer mg.Close()

		// 4. [Black Box] Run analysis
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
