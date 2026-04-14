package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/internal/analyzer"
	"github.com/Homeria/MigraGuard/internal/app"
	"github.com/Homeria/MigraGuard/internal/infra/postgres"
	"github.com/Homeria/MigraGuard/internal/infra/sqlite"
	"github.com/Homeria/MigraGuard/internal/shared/reporter"
	"github.com/spf13/cobra"
)

var (
	analyzeDbString   string
	analyzeSqlitePath string
	analyzeOutput     string
)

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVar(&analyzeDbString, "db", "", "대상 PostgreSQL 접속 문자열 (필수)")
	analyzeCmd.Flags().StringVar(&analyzeSqlitePath, "sqlite", "./migraguard.db", "로컬 메트릭 저장소(SQLite) 경로")
	analyzeCmd.Flags().StringVarP(&analyzeOutput, "output", "o", "console", "결과 출력 형식 (console, markdown)")
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze [migration_file.sql]",
	Short: "마이그레이션 SQL 파일을 즉시 분석하여 위험도를 평가합니다",
	Long:  `제공된 DDL 스크립트를 파싱하고 실시간 및 과거 트래픽 데이터를 기반으로 잠재적 장애 리스크를 수치화합니다.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		ctx := context.Background()

		// 설정 파일 또는 플래그로부터 DB 주소 로드
		if analyzeDbString == "" && GlobalConfig.Database.Postgres != "" {
			analyzeDbString = GlobalConfig.Database.Postgres
		}

		if analyzeDbString == "" {
			fmt.Println("❌ 에러: PostgreSQL 접속 문자열이 필요합니다. --db 플래그나 설정 파일을 확인하세요.")
			os.Exit(1)
		}

		// 1. 인프라 어댑터 초기화
		pool, err := postgres.ConnectPostgres(ctx, analyzeDbString)
		if err != nil {
			fmt.Printf("❌ PostgreSQL 연결 실패: %v\n", err)
			os.Exit(1)
		}
		pg := postgres.NewPostgresAdapter(pool)
		defer pg.Close()

		sl, err := sqlite.NewSQLiteAdapter(analyzeSqlitePath)
		if err != nil {
			fmt.Printf("❌ SQLite 저장소 초기화 실패: %v\n", err)
			os.Exit(1)
		}
		defer sl.Close()

		// 2. 리스크 분석 서비스 기동
		constants := analyzer.RiskConstants{
			DiskIO:   GlobalConfig.Risk.DiskIO,
			MuMax:    GlobalConfig.Risk.MuMax,
			CMax:     GlobalConfig.Risk.CMax,
			TTimeout: GlobalConfig.Risk.TTimeout,
			TMeta:    GlobalConfig.Risk.TMeta,
		}
		if constants.CMax == 0 {
			constants = analyzer.DefaultRiskConstants()
		}

		svc := app.NewAnalyzeService(pg, sl, constants, Verbose)

		// 3. 분석 실행
		resp, err := svc.Run(ctx, app.AnalysisTask{SQLPath: filePath})
		if err != nil {
			fmt.Printf("❌ 분석 실패: %v\n", err)
			os.Exit(1)
		}

		// 4. 결과 리포팅
		var rpt reporter.Reporter
		if analyzeOutput == "markdown" {
			rpt = reporter.NewMarkdownReporter()
		} else {
			rpt = reporter.NewConsoleReporter()
		}

		if err := rpt.Write(resp.Results, resp.Reports); err != nil {
			fmt.Printf("❌ 리포트 생성 실패: %v\n", err)
			os.Exit(1)
		}

		// 5. 게이트키핑 (Danger 감지 시 종료 코드 1 반환)
		for _, r := range resp.Reports {
			if r.RiskLevel == "Danger" {
				fmt.Println("\n🛑 위험: 고위험 마이그레이션이 감지되었습니다. 배포 파이프라인을 차단합니다.")
				os.Exit(1)
			}
		}
	},
}
