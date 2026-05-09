package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/pkg/migraguard"
	"github.com/spf13/cobra"
)

var (
	exportInputPath  string
	exportOutputPath string
)

var exportCmd = &cobra.Command{
	Use:   "export-sandbox",
	Short: "Export metrics from an existing SQLite sandbox to CSV",
	Run: func(cmd *cobra.Command, args []string) {
		if exportInputPath == "" {
			fmt.Fprintln(os.Stderr, "[ERROR] Input SQLite path is required (--input)")
			os.Exit(1)
		}
		if exportOutputPath == "" {
			fmt.Fprintln(os.Stderr, "[ERROR] Output CSV path is required (--output)")
			os.Exit(1)
		}

		mg, err := migraguard.NewSandboxClient(exportInputPath, migraguard.Config{
			Verbose: Verbose,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Initialization failed: %v\n", err)
			os.Exit(1)
		}
		defer mg.Close()

		if err := mg.ExportSandboxMetrics(context.Background(), exportOutputPath); err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Export failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("[OK] Successfully exported %s to %s\n", exportInputPath, exportOutputPath)
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.Flags().StringVarP(&exportInputPath, "input", "i", "", "Path to the input SQLite sandbox database")
	exportCmd.Flags().StringVarP(&exportOutputPath, "output", "o", "", "Path to the output CSV file")
}
