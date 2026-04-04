package main

import (
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	Verbose      bool
	GlobalConfig = config.DefaultConfig()
)

var rootCmd = &cobra.Command{
	Use:   "migraguard",
	Short: "MigraGuard - DB Migration Gatekeeper",
	Long: `MigraGuard is a DevSecOps CLI tool that prevents lock contention
and service outages during database schema changes (DDL) by cross-validating
migration scripts with actual runtime traffic.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./migraguard.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enable verbose output for debugging")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in current directory with name "migraguard" (without extension).
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("migraguard")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Printf("📂 Using config file: %s\n", viper.ConfigFileUsed())
		if err := viper.Unmarshal(GlobalConfig); err != nil {
			fmt.Printf("⚠️ Unable to decode config into struct: %v\n", err)
		}
	}
}
