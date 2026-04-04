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
	collector := db.NewCollector(s.pg, s.sqlite, s.interval)
	collector.SetRetentionDays(s.retentionDays)

	// Start the collector (this usually starts a goroutine or a loop)
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
