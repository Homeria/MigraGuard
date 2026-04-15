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

// analyze 관련 플래그 변수
var (
	analyzeDbString   string // Postgres 접속 문자열
	analyzeSqlitePath string // SQLite 파일 경로
	analyzeOutput     string // 출력 형식 (console/markdown)
)

func init() {
	// analyze 서브 명령어를 등록합니다.
	rootCmd.AddCommand(analyzeCmd)

	// 명령줄 플래그들을 정의하고 기본값을 설정합니다.
	analyzeCmd.Flags().StringVar(&analyzeDbString, "db", "", "대상 PostgreSQL 접속 문자열 (필수)")
	analyzeCmd.Flags().StringVar(&analyzeSqlitePath, "sqlite", "./migraguard.db", "로컬 메트릭 저장소(SQLite) 경로")
	analyzeCmd.Flags().StringVarP(&analyzeOutput, "output", "o", "console", "결과 출력 형식 (console, markdown)")
}

// analyzeCmd는 'analyze' 명령어의 정의 및 실행 로직을 담고 있습니다.
var analyzeCmd = &cobra.Command{
	Use:   "analyze [migration_file.sql]",
	Short: "마이그레이션 SQL 파일을 즉시 분석하여 위험도를 평가합니다",
	Long: `제공된 DDL 스크립트를 파싱하고 실시간 및 과거 트래픽 데이터를 기반으로 
잠재적 장애 리스크를 정량적으로 수치화하여 배포 가능 여부를 판단합니다.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		ctx := context.Background()

		// 1. [접속 정보 로드] 설정 파일이 있다면 로드하고, 플래그가 우선시됩니다.
		if analyzeDbString == "" && GlobalConfig.Database.Postgres != "" {
			analyzeDbString = GlobalConfig.Database.Postgres
		}

		if analyzeDbString == "" {
			fmt.Println("❌ 에러: PostgreSQL 접속 문자열이 필요합니다. --db 플래그나 설정 파일을 확인하세요.")
			os.Exit(1)
		}

		// 2. [인프라 어댑터 초기화]
		// 2-1. 운영 Postgres에 연결하여 커넥션 풀을 생성합니다.
		pool, err := postgres.ConnectPostgres(ctx, analyzeDbString)
		if err != nil {
			fmt.Printf("❌ PostgreSQL 연결 실패: %v\n", err)
			os.Exit(1)
		}
		pg := postgres.NewPostgresAdapter(pool)
		defer pg.Close()

		// 2-2. 로컬 SQLite 저장소를 엽니다.
		sl, err := sqlite.NewSQLiteAdapter(analyzeSqlitePath)
		if err != nil {
			fmt.Printf("❌ SQLite 저장소 초기화 실패: %v\n", err)
			os.Exit(1)
		}
		defer sl.Close()

		// 3. [서비스 레이어 기동] 
		// 설정값으로부터 리스크 상수를 구성합니다.
		constants := analyzer.RiskConstants{
			DiskIO:   GlobalConfig.Risk.DiskIO,
			MuMax:    GlobalConfig.Risk.MuMax,
			CMax:     GlobalConfig.Risk.CMax,
			TTimeout: GlobalConfig.Risk.TTimeout,
			TMeta:    GlobalConfig.Risk.TMeta,
		}
		// 설정이 비어있을 경우 시스템 기본값을 사용합니다.
		if constants.CMax == 0 {
			constants = analyzer.DefaultRiskConstants()
		}

		// AnalyzeService를 생성하여 비즈니스 로직을 실행할 준비를 합니다.
		svc := app.NewAnalyzeService(pg, sl, constants, Verbose)

		// 4. [분석 실행] 서비스의 Run 메서드를 호출하여 전체 파이프라인을 구동합니다.
		resp, err := svc.Run(ctx, app.AnalysisTask{SQLPath: filePath})
		if err != nil {
			fmt.Printf("❌ 분석 실패: %v\n", err)
			os.Exit(1)
		}

		// 5. [결과 출력] 선택된 형식에 맞는 리포터를 생성하여 분석 보고서를 출력합니다.
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

		// 6. [게이트키핑] 'Danger' 등급이 발견될 경우 종료 코드 1을 반환하여 CI/CD 배포를 중단시킵니다.
		for _, r := range resp.Reports {
			if r.RiskLevel == "Danger" {
				fmt.Println("\n🛑 위험: 고위험 마이그레이션이 감지되었습니다. 배포 파이프라인을 강제 차단합니다.")
				os.Exit(1)
			}
		}
	},
}
