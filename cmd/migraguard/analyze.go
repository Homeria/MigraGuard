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

var dbString string

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{

	Use: "analyze [migration_file.sql]",

	Short: "Analyze a migration SQL file to evaluate risk and report results",

	Long: `Parses the given DDL migration script (AST), evaluates the potential risk
using the MigraGuard v3.0 queuing model, and reports the findings to the terminal.
Exits with status 1 if the risk level is 'Danger'.`,

	Args: cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {

		// filePath : command argument로 받은 Migration(DDL) 파일 경로
		filePath := args[0]
		fmt.Printf("🚀 Starting static and dynamic analysis for: %s\n", filePath)

		if dbString == "" {
			fmt.Println("❌ Error: --db flag is required to fetch real-time metrics.")
			os.Exit(1)
		}

		// [L01] SQL 로드: Read SQL File
		// sqlContent : Migration(DDL) 파일의 내용을 문자열로 읽어옴
		// err : 파일 읽기 과정에서 발생할 수 있는 오류
		sqlContent, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("❌ Failed to read file: %v\n", err)
			os.Exit(1)
		}

		// [L02] 정적 분석 호출: Parse SQL (AST Analysis)
		// ast.go - ParseSQL 함수를 호출하여 SQL 문자열을 복사하고, AST로 변환하여 AST 분석 결과를 받음
		// results : AST 분석 결과, DDL 작업 유형과 영향을 받는 테이블/컬럼 정보 등을 포함하는 AnalysisResult 구조체의 슬라이스
		// err : AST 분석 과정에서 발생할 수 있는 오류
		results, err := parser.ParseSQL(string(sqlContent))
		if err != nil {
			fmt.Printf("❌ Parser Error: %v\n", err)
			os.Exit(1)
		}
		if len(results) == 0 {
			fmt.Println("⚠️ No DDL operations found in the file.")
			return
		}

		// [L03] DB 동적 지표 수집 준비: Connect to Database
		// ctx : 데이터베이스 연결 및 작업 수행에 사용할 컨텍스트(url)
		ctx := context.Background()

		// pg : internal/db/connection.go - NewPostgresAdapter 함수를 통해 PostgreSQL 데이터베이스에 연결하기 위한 Adapter Instance 생성
		// Connect 메서드를 호출하여 실제로 데이터베이스에 연결을 시도하고, 연결 오류가 발생할 경우 오류 메시지를 출력하고 프로세스 종료
		pg := db.NewPostgresAdapter(dbString)
		if err := pg.Connect(ctx); err != nil {
			fmt.Printf("❌ Database Connection Error: %v\n", err)
			os.Exit(1)
		}

		// defer : 해당 함수(analyzeCmd.Run)가 종료될 때 pg.Close()가 호출되어 데이터베이스 연결이 적절히 종료되도록 보장
		defer pg.Close()

		// 4. Initialize Engines
		riskEngine := engine.NewRiskEngine(pg, nil, engine.DefaultRiskConstants())

		fmt.Println("\n--- MigraGuard v3.0 Risk Analysis Report ---")

		hasDanger := false
		for _, res := range results {
			fmt.Printf("\n[Target Table: %s | Operation: %s]\n", res.TableName, res.Operation)

			// [L04] 리스크 계산: Calculate Risk Score using Model v3.0
			report, err := riskEngine.AnalyzeRisk(ctx, res)
			if err != nil {
				fmt.Printf("  ❌ Risk Analysis Error: %v\n", err)
				continue
			}

			// [L41] 결과 리포팅: Report Results
			fmt.Printf("  📊 Physical DDL Time (T_ddl): %.2f ms\n", report.EstimatedDDLTime)
			fmt.Printf("  🔒 Estimated Block Time (T_block): %.2f ms\n", report.BlockingTime)
			fmt.Printf("  📈 Peak Connections (C_peak): %d\n", report.PeakConnections)
			fmt.Printf("  ♻️ System Recovery Time (T_rec): %.2f ms\n", report.RecoveryTime)
			fmt.Printf("  ⚠️ RISK SCORE: %.2f%%\n", report.RiskScore)
			fmt.Printf("  🚦 RISK LEVEL: [%s]\n", report.RiskLevel)

			if report.PermanentFailure {
				fmt.Println("  🚨 CRITICAL: Permanent system failure predicted!")
			}

			if report.RiskLevel == "Danger" {
				hasDanger = true
			}
		}

		fmt.Println("\n------------------------------------------------")

		// [L05/L42] 최종 판단 및 프로세스 제어: Gatekeeping
		if hasDanger {
			fmt.Println("🚫 STATUS: REJECTED (High Risk Detected)")
			os.Exit(1)
		} else {
			fmt.Println("✅ STATUS: APPROVED (Safe to Deploy)")
		}
	},
}

func init() {
	analyzeCmd.Flags().StringVarP(&dbString, "db", "d", "", "PostgreSQL connection string (required)")
	rootCmd.AddCommand(analyzeCmd)
}
