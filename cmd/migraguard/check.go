package main

import (
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/pkg/migraguard"
	"github.com/spf13/cobra"
)

var (
	checkDbURL      string
	checkSqlitePath string
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check target database connection and agent status",
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Resolve PostgreSQL DSN
		finalDB := checkDbURL
		if finalDB == "" {
			finalDB = GlobalConfig.Database.Postgres
		}
		if finalDB == "" {
			finalDB = "postgres://user:pass@localhost:5432/postgres"
		}

		// 2. Resolve SQLite Path
		finalSQLite := checkSqlitePath
		if finalSQLite == "" {
			finalSQLite = GlobalConfig.Database.SQLite
		}
		if finalSQLite == "" {
			finalSQLite = "migraguard.db"
		}

		// 3. Initialize client
		mg, err := migraguard.New(migraguard.Config{
			PostgresDSN: finalDB,
			SQLitePath:  finalSQLite,
		})
		if err != nil {
			fmt.Printf("[ERROR] Connection failed: %v\n", err)
			os.Exit(1)
		}
		defer mg.Close()

		fmt.Println("[OK] Successfully connected to PostgreSQL.")
		fmt.Printf("[INFO] Using DB: %s\n", finalDB)
		fmt.Printf("[INFO] Using SQLite: %s\n", finalSQLite)
		fmt.Println("[INFO] MigraGuard SDK is ready to use.")
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringVar(&checkDbURL, "db", "", "Target PostgreSQL URL")
	checkCmd.Flags().StringVar(&checkSqlitePath, "sqlite", "", "Target SQLite path")
}
