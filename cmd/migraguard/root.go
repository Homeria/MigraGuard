package main

import (
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/pkg/migraguard/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile      string
	Verbose      bool
	GlobalConfig = &config.Config{}
)

var rootCmd = &cobra.Command{
	Use:   "migraguard",
	Short: "MigraGuard - DB Migration Gatekeeper",
	Long: `MigraGuard is a DevSecOps CLI tool that cross-verifies database schema 
changes (DDL) using real-world traffic data to prevent lock contention 
and service outages.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute runs the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Set global persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (default is ./migraguard.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "enable verbose logging for debugging")
}

// initConfig reads the configuration file or environment variables.
func initConfig() {
	loadedConfig, err := config.LoadConfig(cfgFile)
	if err != nil {
		// Use defaults if config file is missing
	} else {
		GlobalConfig = loadedConfig
	}
}
