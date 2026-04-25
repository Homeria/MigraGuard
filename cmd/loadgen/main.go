package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Homeria/MigraGuard/pkg/migraguard/simulation"
	"github.com/spf13/cobra"
)

var (
	loadgenDSN     string
	loadgenConns   int
	loadgenProfile string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "loadgen",
		Short: "Traffic generator based on e-commerce scenarios",
		Run: func(cmd *cobra.Command, args []string) {
			if loadgenDSN == "" {
				fmt.Println("[ERROR] Target database DSN is required (--db)")
				os.Exit(1)
			}

			// Initialize load generator engine
			gen, err := simulation.NewLoadGenerator(loadgenDSN, loadgenConns, simulation.TrafficProfile(loadgenProfile))
			if err != nil {
				log.Fatalf("[ERROR] Initialization failed: %v", err)
			}

			// Start simulation
			gen.Run(cmd.Context())
		},
	}

	rootCmd.Flags().StringVar(&loadgenDSN, "db", "", "Target PostgreSQL DSN (required)")
	rootCmd.Flags().IntVar(&loadgenConns, "conns", 15, "Number of concurrent workers")
	rootCmd.Flags().StringVar(&loadgenProfile, "profile", "steady", "Traffic profile (steady, flash-sale, read-heavy)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
