package main

import (
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/pkg/migraguard"
	"github.com/spf13/cobra"
)

var checkDbURL string

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check target database connection and agent status",
	Run: func(cmd *cobra.Command, args []string) {
		// decision on connection string
		if checkDbURL == "" && GlobalConfig.Database.Postgres != "" {
			checkDbURL = GlobalConfig.Database.Postgres
		}

		if checkDbURL == "" {
			fmt.Println("[ERROR] PostgreSQL connection string is required.")
			os.Exit(1)
		}

		// Initialize client
		mg, err := migraguard.New(migraguard.Config{
			PostgresDSN: checkDbURL,
			SQLitePath:  "./migraguard.db",
		})
		if err != nil {
			fmt.Printf("[ERROR] Connection failed: %v\n", err)
			os.Exit(1)
		}
		defer mg.Close()

		fmt.Println("[OK] Successfully connected to PostgreSQL.")
		fmt.Println("[INFO] MigraGuard SDK is ready to use.")
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringVar(&checkDbURL, "db", "", "Target PostgreSQL URL")
}
