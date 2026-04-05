package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Homeria/MigraGuard/internal/db"
)

// AgentService manages the background workload collection process.
type AgentService struct {
	pg            db.PostgresClient
	sqlite        db.SQLiteClient
	interval      time.Duration
	retentionDays int
}

// NewAgentService creates a new instance of AgentService.
func NewAgentService(pg db.PostgresClient, sqlite db.SQLiteClient, interval time.Duration, retentionDays int) *AgentService {
	return &AgentService{
		pg:            pg,
		sqlite:        sqlite,
		interval:      interval,
		retentionDays: retentionDays,
	}
}

// Run starts the continuous collection loop.
func (s *AgentService) Run(ctx context.Context) error {

	// internal/db/collector.go - 컬렉터 객체 생성
	// Collector는 PostgreSQL의 상황을 주기적으로 수집하여 SQLite에 저장하는 역할.
	collector := db.NewCollector(s.pg, s.sqlite, s.interval)

	// 데이터 보존 기간 설정
	collector.SetRetentionDays(s.retentionDays)

	// 주기적 수집 루프 시작
	collector.Start(ctx)

	fmt.Printf("✅ Agent Service started. Interval: %v, Retention: %d days\n", s.interval, s.retentionDays)
	fmt.Println("📡 Collecting workload snapshots... Press Ctrl+C to stop.")

	// Wait for context cancellation
	<-ctx.Done()

	fmt.Println("\n🛑 Stopping Agent Service gracefully...")
	collector.Stop()

	// Final cleanup
	time.Sleep(1 * time.Second)
	return nil
}
