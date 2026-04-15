package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Homeria/MigraGuard/internal/app"
	"github.com/Homeria/MigraGuard/internal/infra/postgres"
	"github.com/Homeria/MigraGuard/internal/infra/sqlite"
	"github.com/spf13/cobra"
)

var (
	dbURL         string
	sqlitePath    string
	agentInterval int
	retentionDays int
	targetTables  string
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "실시간 메트릭 수집 에이전트 구동",
	Long:  `PostgreSQL의 통계 뷰를 주기적으로 조회하여 워크로드 패턴을 SQLite에 시계열 데이터로 저장합니다.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. 인프라 어댑터 초기화 (Postgres)
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		pgPool, err := postgres.ConnectPostgres(ctx, dbURL)
		if err != nil {
			return fmt.Errorf("PostgreSQL 연결 실패: %w", err)
		}
		defer pgPool.Close()
		pgAdapter := postgres.NewPostgresAdapter(pgPool)

		// 2. 인프라 어댑터 초기화 (SQLite)
		sqliteAdapter, err := sqlite.NewSQLiteAdapter(sqlitePath)
		if err != nil {
			return fmt.Errorf("SQLite 초기화 실패: %w", err)
		}
		defer sqliteAdapter.Close()

		// 3. 어플리케이션 서비스 조립
		interval := time.Duration(agentInterval) * time.Second
		agentService := app.NewAgentService(pgAdapter, sqliteAdapter, interval, retentionDays)
		agentService.SetTargetTables(targetTables)

		// 4. 에이전트 실행
		return agentService.Run(ctx)
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)

	// 명령줄 플래그 정의
	agentCmd.Flags().StringVar(&dbURL, "db", "postgres://user:pass@localhost:5432/postgres", "PostgreSQL 연결 URL")
	agentCmd.Flags().StringVar(&sqlitePath, "sqlite", "migraguard.db", "메트릭 저장용 SQLite 파일 경로")
	agentCmd.Flags().IntVar(&agentInterval, "interval", 60, "지표 수집 주기 (초)")
	agentCmd.Flags().IntVar(&retentionDays, "retention", 7, "데이터 보존 기간 (일)")
	agentCmd.Flags().StringVar(&targetTables, "tables", "", "집중 모니터링 대상 테이블 (콤마 구분)")
}
