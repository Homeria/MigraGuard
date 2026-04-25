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

		// 1. [Black Box] Initialize client
		mg, err := migraguard.New(migraguard.Config{
			PostgresDSN:   dbURL,
			SQLitePath:    sqlitePath,
			Interval:      time.Duration(agentInterval) * time.Second,
			RetentionDays: retentionDays,
			Verbose:       Verbose,
		})
		if err != nil {
			return fmt.Errorf("[ERROR] MigraGuard initialization failed: %w", err)
		}
		defer mg.Close()

		// 2. [Black Box] Run agent
		return mg.StartAgent(ctx, targetTables)
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)

	agentCmd.Flags().StringVar(&dbURL, "db", "postgres://user:pass@localhost:5432/postgres", "PostgreSQL connection URL")
	agentCmd.Flags().StringVar(&sqlitePath, "sqlite", "migraguard.db", "SQLite file path for metrics")
	agentCmd.Flags().IntVar(&agentInterval, "interval", 60, "Metric collection interval (seconds)")
	agentCmd.Flags().IntVar(&retentionDays, "retention", 7, "Data retention period (days)")
	agentCmd.Flags().StringVar(&targetTables, "tables", "", "Target tables for monitoring (comma separated)")
}
