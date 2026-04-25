package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Homeria/MigraGuard/pkg/migraguard"
	"github.com/spf13/cobra"
)

var (
	dbURL         string
	sqlitePath    string
	agentInterval int
	retentionDays int
	targetTables  string
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run real-time metric collection agent",
	Long:  `Periodically queries PostgreSQL statistics and stores workload patterns in SQLite.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		// 1. Resolve PostgreSQL DSN
		finalDB := dbURL
		if finalDB == "" {
			finalDB = GlobalConfig.Database.Postgres
		}
		if finalDB == "" {
			finalDB = "postgres://user:pass@localhost:5432/postgres"
		}

		// 2. Resolve SQLite Path
		finalSQLite := sqlitePath
		if finalSQLite == "" {
			finalSQLite = GlobalConfig.Database.SQLite
		}
		if finalSQLite == "" {
			finalSQLite = "migraguard.db"
		}

		// 3. Resolve Interval (Priority: Flag > Config > Default 60s)
		finalInterval := time.Duration(agentInterval) * time.Second
		// If flag is at default (60) and config has a value, use config
		if agentInterval == 60 && GlobalConfig.Agent.Interval != "" {
			if d, err := time.ParseDuration(GlobalConfig.Agent.Interval); err == nil {
				finalInterval = d
			}
		}

		// 4. Resolve Retention (Priority: Flag > Config > Default 7d)
		finalRetention := retentionDays
		if retentionDays == 7 && GlobalConfig.Agent.RetentionDays != 0 {
			finalRetention = GlobalConfig.Agent.RetentionDays
		}

		// 5. Initialize client
		mg, err := migraguard.New(migraguard.Config{
			PostgresDSN:   finalDB,
			SQLitePath:    finalSQLite,
			Interval:      finalInterval,
			RetentionDays: finalRetention,
			Verbose:       Verbose,
		})
		if err != nil {
			return fmt.Errorf("[ERROR] MigraGuard initialization failed: %w", err)
		}
		defer mg.Close()

		if Verbose {
			fmt.Printf("[DEBUG] Interval: %v, Retention: %d days\n", finalInterval, finalRetention)
		}

		// 6. Run agent
		return mg.StartAgent(ctx, targetTables)
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)

	agentCmd.Flags().StringVar(&dbURL, "db", "", "PostgreSQL connection URL")
	agentCmd.Flags().StringVar(&sqlitePath, "sqlite", "", "SQLite file path for metrics")
	agentCmd.Flags().IntVar(&agentInterval, "interval", 60, "Metric collection interval (seconds)")
	agentCmd.Flags().IntVar(&retentionDays, "retention", 7, "Data retention period (days)")
	agentCmd.Flags().StringVar(&targetTables, "tables", "", "Target tables for monitoring (comma separated)")
}
