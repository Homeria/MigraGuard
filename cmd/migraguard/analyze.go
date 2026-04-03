package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Homeria/MigraGuard/internal/db"
	"github.com/Homeria/MigraGuard/internal/engine"
	"github.com/Homeria/MigraGuard/internal/parser"
	"github.com/spf13/cobra"
)

var dbString string

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze [migration_file.sql]",
	Short: "Analyze a migration SQL file to evaluate risk and report results",
	Long: `Parses the given DDL migration script (AST), evaluates the potential risk
using the MigraGuard v3.0 queuing model, and reports the findings to the terminal.
Exits with status 1 if the risk level is 'Danger'.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		fmt.Printf("🚀 Starting static and dynamic analysis for: %s\n", filePath)

		if dbString == "" {
			fmt.Println("❌ Error: --db flag is required to fetch real-time metrics.")
			os.Exit(1)
		}

		// [L01] SQL 로드
		sqlContent, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("❌ Failed to read file: %v\n", err)
			os.Exit(1)
		}

		// [L02] 정적 분석 (AST)
		results, err := parser.ParseSQL(string(sqlContent))
		if err != nil {
			fmt.Printf("❌ Parser Error: %v\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Println("⚠️ No DDL operations found in the file.")
			return
		}

		// [L03] DB 연결
		ctx := context.Background()
		pg := db.NewPostgresAdapter(dbString)
		if err := pg.Connect(ctx); err != nil {
			fmt.Printf("❌ Database Connection Error: %v\n", err)
			os.Exit(1)
		}
		defer pg.Close()

		// [L21] SQLite 및 백그라운드 수집기(Collector) 활성화
		// 누적치 차이 분석(Delta Analysis)을 위해 SQLite를 저장소로 사용
		sqlite, err := db.NewSQLiteAdapter("migraguard.db")
		if err != nil {
			fmt.Printf("⚠️ SQLite Initialization Warning: %v\n", err)
		} else {
			defer sqlite.Close()
			// 1초 주기로 스냅샷 수집 시작 (누적치의 차이를 얻기 위함)
			collector := db.NewCollector(pg, sqlite, 1000*time.Millisecond)
			for _, res := range results {
				collector.AddTargetTable(res.TableName)
			}
			collector.Start(ctx)

			fmt.Println("📊 Collecting real-time workload snapshots (Deltas)...")
			time.Sleep(3 * time.Second) // 최소 2개 이상의 스냅샷 확보를 위해 대기
		}

		// [L04] 리스크 엔진 초기화 (SQLite 주입)
		riskEngine := engine.NewRiskEngine(pg, sqlite, engine.DefaultRiskConstants())

		fmt.Println("\n--- MigraGuard v3.0 Risk Analysis Report ---")

		hasDanger := false
		for _, res := range results {
			fmt.Printf("\n[Target Table: %s | Operation: %s]\n", res.TableName, res.Operation)

			report, err := riskEngine.AnalyzeRisk(ctx, res)
			if err != nil {
				fmt.Printf("  ❌ Risk Analysis Error: %v\n", err)
				continue
			}

			// [L41] 결과 리포팅
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

		// [L05/L42] 게이트키핑
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
