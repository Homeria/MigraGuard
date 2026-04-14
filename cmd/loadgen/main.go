package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Homeria/MigraGuard/internal/simulation"
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
		Short: "이커머스 시나리오 기반의 부하 생성기",
		Run: func(cmd *cobra.Command, args []string) {
			if loadgenDSN == "" {
				fmt.Println("❌ 에러: 대상 DB의 DSN 문자열이 필요합니다. (--db)")
				os.Exit(1)
			}

			// 부하 생성기 엔진 초기화
			gen, err := simulation.NewLoadGenerator(loadgenDSN, loadgenConns, simulation.TrafficProfile(loadgenProfile))
			if err != nil {
				log.Fatalf("❌ 생성기 초기화 실패: %v", err)
			}

			// 시뮬레이션 시작
			gen.Run(cmd.Context())
		},
	}

	rootCmd.Flags().StringVar(&loadgenDSN, "db", "", "대상 PostgreSQL DSN (필수)")
	rootCmd.Flags().IntVar(&loadgenConns, "conns", 15, "동시 워커(고루틴) 수")
	rootCmd.Flags().StringVar(&loadgenProfile, "profile", "steady", "부하 프로파일 (steady, flash-sale, read-heavy)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
