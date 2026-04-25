package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Homeria/MigraGuard/pkg/migraguard/internal/collector"
	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// AgentService manages background workload monitoring.
type AgentService struct {
	pg            types.PostgresClient
	sqlite        types.SQLiteClient
	interval      time.Duration
	retentionDays int
	targetTables  []string
}

// NewAgentService initializes the agent service component.
func NewAgentService(pg types.PostgresClient, sqlite types.SQLiteClient, interval time.Duration, retentionDays int) *AgentService {
	return &AgentService{
		pg:            pg,
		sqlite:        sqlite,
		interval:      interval,
		retentionDays: retentionDays,
		targetTables:  []string{},
	}
}

// SetTargetTables sets the tables for monitoring.
func (s *AgentService) SetTargetTables(tables string) {
	if tables == "" {
		return
	}
	rawList := strings.Split(tables, ",")
	for _, t := range rawList {
		s.targetTables = append(s.targetTables, strings.TrimSpace(t))
	}
}

// Run executes the agent metrics collection loop.
func (s *AgentService) Run(ctx context.Context) error {
	col := collector.NewCollector(s.pg, s.sqlite, s.interval)
	col.SetRetentionDays(s.retentionDays)

	for _, table := range s.targetTables {
		col.AddTargetTable(table)
	}

	col.Start(ctx)

	fmt.Printf("[OK] MigraGuard Agent service active. (Interval: %v, Retention: %d days)\n", s.interval, s.retentionDays)
	if len(s.targetTables) > 0 {
		fmt.Printf("[INFO] Monitoring tables: %v\n", s.targetTables)
	}
	fmt.Println("[INFO] Collecting real-time metrics... (Ctrl+C to exit)")

	<-ctx.Done()

	fmt.Println("\n[STOP] Gracefully shutting down Agent service...")
	col.Stop()

	time.Sleep(1 * time.Second)
	return nil
}
