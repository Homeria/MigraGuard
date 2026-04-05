package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/internal/db"
	"github.com/Homeria/MigraGuard/internal/engine"
	"github.com/Homeria/MigraGuard/internal/errors"
	"github.com/Homeria/MigraGuard/internal/reporter"
	"github.com/Homeria/MigraGuard/internal/service"
	"github.com/spf13/cobra"
)

var (
	analyzeDbString   string
	analyzeSqlitePath string
	analyzeOutput     string
)

func init() {
	rootCmd.AddCommand(analyzeCmd)

	// Analyze configuration flags with GlobalConfig defaults
	analyzeCmd.Flags().StringVar(&analyzeDbString, "db", GlobalConfig.Database.URL, "PostgreSQL connection string")
	analyzeCmd.Flags().StringVar(&analyzeSqlitePath, "sqlite", GlobalConfig.Database.SQLitePath, "Path to the local SQLite storage file")
	analyzeCmd.Flags().StringVarP(&analyzeOutput, "output", "o", "console", "Output format (console, markdown)")
}

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze [migration_file.sql]",
	Short: "Analyze a migration SQL file to evaluate risk instantly",
	Long: `Parses the given DDL migration script (AST), evaluates the potential risk
using the MigraGuard v3.1 baseline model, and reports the findings.
This command relies on data collected by the 'migraguard agent'.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		ctx := context.Background()

		// 1. Initialize Infrastructure (Adapters)

		// 타겟 DB 연결 검사
		if analyzeDbString == "" {
			fmt.Println("❌ Error: --db flag is required to connect to PostgreSQL.")
			os.Exit(1)
		}

		// 타겟 DB 연결 어댑터 초기화 및 연결 여부 검사
		pg := db.NewPostgresAdapter(analyzeDbString)
		if err := pg.Connect(ctx); err != nil {
			fmt.Printf("❌ Failed to connect to PostgreSQL: %v\n", err)
			os.Exit(1)
		}
		defer pg.Close()

		// 타겟 DB에 대한 지표를 저장할 SQLite DB 어댑터 초기화 여부 검사
		sqlite, err := db.NewSQLiteAdapter(analyzeSqlitePath)
		if err != nil {
			fmt.Printf("❌ Failed to initialize SQLite: %v\n", err)
			os.Exit(1)
		}
		defer sqlite.Close()

		// 2. Initialize Domain Service

		// 생성된 타겟 DB 어댑터 및 SQLite 어댑터를 서비스 레이어에 주입하여 분석 서비스 인스턴스 생성
		svc := service.NewAnalyzeService(pg, sqlite, GlobalConfig.Engine, Verbose)

		if analyzeOutput == "console" {
			fmt.Printf("🔍 Starting instant risk analysis for: %s\n", filePath)
			fmt.Println("📡 Fetching baseline metrics from SQLite...")
		}

		// 3. Execute Analysis

		// 서비스 레이어의 Run 메서드를 호출하여 분석 실행, 파라미터로 분석할 SQL 파일 경로 전달
		resp, err := svc.Run(ctx, service.AnalysisTask{SQLPath: filePath})
		if err != nil {
			handleAnalysisError(err)
			os.Exit(1)
		}

		// 4. Handle Output (Presentation Layer)
		// 서비스 레이어에서 반환된 분석 결과를 사용자가 선택한 출력 형식에 맞게 포맷팅하여 출력
		var rpt reporter.Reporter
		switch analyzeOutput {
		case "markdown":
			rpt = reporter.NewMarkdownReporter()
		default:
			rpt = reporter.NewConsoleReporter()
		}

		// 분석 결과를 선택한 리포터로 출력, 리포터의 Write 메서드에 분석 결과와 보고서 전달
		if err := rpt.Write(resp.Results, resp.Reports); err != nil {
			fmt.Printf("❌ Reporting Failed: %v\n", err)
			os.Exit(1)
		}

		// 5. Final Gatekeeping (Exit Code 1 for Danger)
		// 분석 결과 보고서에서 위험 수준이 "Danger"인 항목이 있는지 검사하여, 위험이 감지된 경우 사용자에게 경고 메시지를 출력하고 프로세스를 종료
		if hasDanger(resp.Reports) {
			if analyzeOutput == "console" {
				fmt.Println("🛑 Danger detected! Migration blocked.")
			}
			os.Exit(1)
		} else if analyzeOutput == "console" {
			fmt.Println("✅ Analysis complete. No critical risks found.")
		}
	},
}

// handleAnalysisError provides user-friendly error messages.
func handleAnalysisError(err error) {
	if errors.Is(err, errors.ErrInvalidSQL) {
		fmt.Println("⚠️  No valid DDL operations found in the provided SQL file.")
	} else if errors.Is(err, errors.ErrTableNotFound) {
		fmt.Println("❌  Error: The target table(s) could not be found in the database.")
	} else if errors.Is(err, errors.ErrDatabaseConn) {
		fmt.Println("❌  Error: Database connection lost or failed.")
	} else {
		fmt.Printf("❌ Analysis Failed: %v\n", err)
	}
}

// hasDanger checks if any report indicates a Danger level.
func hasDanger(reports []*engine.RiskAnalysisReport) bool {
	for _, r := range reports {
		if r.RiskLevel == "Danger" {
			return true
		}
	}
	return false
}
