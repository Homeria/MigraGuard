package main

import (
	"context"
	"fmt"

	"github.com/Homeria/MigraGuard/pkg/migraguard"
	"github.com/spf13/cobra"
)

var (
	scenarioPath string
	forceSeed    bool
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
		})
		if err != nil {
			return err
		}

		_, err = mg.Simulate(context.Background(), scenarioPath, forceSeed)
		return err
	},
}

func init() {
	rootCmd.AddCommand(simulateCmd)
	simulateCmd.Flags().StringVarP(&scenarioPath, "scenario", "s", "", "Path to the simulation scenario YAML")
	simulateCmd.Flags().BoolVarP(&forceSeed, "force", "f", false, "Force overwrite existing sandbox database")
}
