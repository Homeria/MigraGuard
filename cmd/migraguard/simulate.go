package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/Homeria/MigraGuard/pkg/migraguard"
	"github.com/spf13/cobra"
)

var (
	scenarioPath string
	forceSeed    bool
	exportCSV    string
	noDB         bool
)

var simulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "Create a simulation sandbox from a YAML scenario",
	Long: `Loads a declarative scenario from YAML, creates a fresh SQLite sandbox, 
and seeds it with mathematical time-series data for research validation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if scenarioPath == "" {
			return fmt.Errorf("scenario path is required (--scenario)")
		}

		// Initialize client (PG connection is NOT required for simulation setup)
		mg, err := migraguard.New(migraguard.Config{
			Verbose: Verbose,
		}, migraguard.WithDangerThreshold(0)) // Example option, dummy DSN allowed in New()
		if err != nil {
			return err
		}

		dbPath, err := mg.Simulate(context.Background(), scenarioPath, forceSeed)
		if err != nil {
			return err
		}

		if exportCSV != "" {
			// Connect to the newly created sandbox to export
			mg, err := migraguard.NewSandboxClient(dbPath, migraguard.Config{Verbose: Verbose})
			if err != nil {
				return err
			}
			defer mg.Close()

			var writer io.Writer
			if exportCSV == "-" {
				writer = os.Stdout
			} else {
				f, err := os.Create(exportCSV)
				if err != nil {
					return fmt.Errorf("failed to create CSV file: %w", err)
				}
				defer f.Close()
				writer = f
			}

			if err := mg.ExportSandboxMetricsToWriter(context.Background(), writer); err != nil {
				return fmt.Errorf("failed to export metrics: %w", err)
			}

			if exportCSV != "-" {
				fmt.Fprintf(os.Stderr, "[OK] Metrics exported to %s\n", exportCSV)
			}
		}

		if noDB && exportCSV != "" {
			mg.Close() // Close handle to allow deletion
			if err := os.Remove(dbPath); err != nil {
				fmt.Fprintf(os.Stderr, "[WARNING] Failed to remove sandbox DB: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "[INFO] Temporary sandbox database removed (%s)\n", dbPath)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(simulateCmd)
	simulateCmd.Flags().StringVarP(&scenarioPath, "scenario", "s", "", "Path to the simulation scenario YAML")
	simulateCmd.Flags().BoolVarP(&forceSeed, "force", "f", false, "Force overwrite existing sandbox database")
	simulateCmd.Flags().StringVar(&exportCSV, "csv", "", "Export generated metrics to CSV file (use '-' for Stdout)")
	simulateCmd.Flags().BoolVar(&noDB, "no-db", false, "Remove the SQLite DB file after CSV export (only works with --csv)")
}
