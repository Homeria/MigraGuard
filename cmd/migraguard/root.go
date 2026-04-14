package main

import (
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/internal/infra/config"
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
	Long: `MigraGuard는 데이터베이스 스키마 변경(DDL) 시 발생할 수 있는 락 경합과 서비스 장애를
실제 운영 트래픽 데이터를 기반으로 교차 검증하여 차단하는 DevSecOps CLI 도구입니다.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute는 루트 명령을 실행하고 플래그를 적절히 설정합니다.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// 전역 영속성 플래그 설정
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "설정 파일 경로 (기본값: ./migraguard.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "디버깅을 위한 상세 로그 출력 활성화")
}

// initConfig는 설정 파일이나 환경 변수를 읽어옵니다.
func initConfig() {
	loadedConfig, err := config.LoadConfig(cfgFile)
	if err != nil {
		// 설정 파일이 없는 경우 기본값 사용
	} else {
		GlobalConfig = loadedConfig
	}
}
