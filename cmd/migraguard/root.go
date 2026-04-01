package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "migraguard",
	Short: "MigraGuard - DB Migration Gatekeeper",
	Long: `MigraGuard is a DevSecOps CLI tool that prevents lock contention 
and service outages during database schema changes (DDL) by cross-validating 
migration scripts with actual runtime traffic.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 기본 명령어 실행 시 도움말 출력
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// 향후 글로벌 플래그 설정 (예: --config, --verbose 등)을 여기에 추가합니다.
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.migraguard.yaml)")
}
