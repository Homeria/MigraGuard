package main

import (
	"context"
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/internal/db"
	"github.com/Homeria/MigraGuard/internal/engine"
	"github.com/Homeria/MigraGuard/internal/parser"
	"github.com/spf13/cobra"
)

var (
	analyzeDbString  string
	analyzeSqlitePath string
)

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
		fmt.Printf("🔍 Starting instant risk analysis for: %s\n", filePath)

		if analyzeDbString == "" {
			fmt.Println("❌ Error: --db flag is required to connect to PostgreSQL.")
			os.Exit(1)
		}

		ctx := context.Background()

		// [L01] Load SQL file
		// 마이그레이션 SQL 파일을 읽습니다.
		sqlContent, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("❌ Failed to read file: %v\n", err)
			os.Exit(1)
		}

		// [L02] Static Analysis (AST)
		// SQL을 파싱하여 타겟 테이블과 작업을 식별합니다.
		results, err := parser.ParseSQL(string(sqlContent))
		if err != nil {
			fmt.Printf("❌ Parser Error: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Println("⚠️ No DDL operations found in the provided SQL file.")
			return
		}

		// [L03] Initialize Adapters
		pg, err := db.NewPostgresAdapter(analyzeDbString)
		if err != nil {
			fmt.Printf("❌ Failed to connect to PostgreSQL: %v\n", err)
			os.Exit(1)
		}
		defer pg.Close()

		sqlite, err := db.NewSQLiteAdapter(analyzeSqlitePath)
		if err != nil {
			fmt.Printf("❌ Failed to initialize SQLite: %v\n", err)
			os.Exit(1)
		}
		defer sqlite.Close()

		// [L04] Check for baseline data
		// 에이전트가 수집한 데이터가 있는지 확인합니다.
		// 데이터가 없으면 분석의 정확도가 떨어질 수 있음을 경고합니다.
		fmt.Println("📊 Fetching baseline metrics from SQLite...")

		// [L05] Initialize Risk Engine
		// v3.1 큐잉 모델을 기반으로 리스크 엔진을 초기화합니다.
		riskEngine := engine.NewRiskEngine(pg, sqlite, engine.DefaultRiskConstants())

		fmt.Println("\n--- MigraGuard v3.1 Risk Analysis Report ---")

		hasDanger := false
		for _, res := range results {
			fmt.Printf("\n[Target Table: %s | Operation: %s]\n", res.TableName, res.Operation)

			// [L31~L35] Analyze Risk using Baseline Data
			// 에이전트가 쌓아둔 시계열 데이터를 사용하여 즉각 분석을 수행합니다.
			report, err := riskEngine.AnalyzeRisk(ctx, res)
			if err != nil {
				fmt.Printf("  ❌ Risk Analysis Error: %v\n", err)
				continue
			}

			// [L41] Report results
			// 산출된 정량적 지표를 리포팅합니다.
			fmt.Printf("  ✅ Physical DDL Time (T_ddl): %.2f ms\n", report.EstimatedDDLTime)
			fmt.Printf("  ⌛ Estimated Block Time (T_block): %.2f ms\n", report.BlockingTime)
			fmt.Printf("  📈 Peak Connections (C_peak): %d\n", report.PeakConnections)
			fmt.Printf("  🚦 RISK LEVEL: [%s]\n", report.RiskLevel)

			if report.PermanentFailure {
				fmt.Println("  🔥 CRITICAL: Permanent system failure predicted! Check your mu_max and connection limits.")
			}

			if report.RiskLevel == "Danger" {
				hasDanger = true
			}
		}

		fmt.Println("\n------------------------------------------------")

		// [L42] Gatekeeping (Exit Code 1 for Danger)
		// 리스크 레벨이 Danger인 경우 배포를 차단하기 위해 종료 코드 1을 반환합니다.
		if hasDanger {
			fmt.Println("🛑 DEPLOYMENT BLOCKED: High risk detected. Please review the report above.")
			os.Exit(1)
		} else {
			fmt.Println("🚀 DEPLOYMENT SAFE: No critical risks identified.")
		}
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVar(&analyzeDbString, "db", "", "PostgreSQL connection string (required)")
	analyzeCmd.Flags().StringVar(&analyzeSqlitePath, "sqlite", "migraguard.db", "Path to local SQLite database")
}
