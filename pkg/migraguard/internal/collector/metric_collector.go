package collector

import (
	"context"
	"fmt"
	"time"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// Collector periodically collects and processes PostgreSQL metrics.
type Collector struct {
	pg            types.PostgresClient
	sqlite        types.SQLiteClient
	interval      time.Duration
	retentionDays int
	targetTables  []string
	stopChan      chan struct{}
}

// NewCollector initializes a new Collector.
func NewCollector(pg types.PostgresClient, sqlite types.SQLiteClient, interval time.Duration) *Collector {
	return &Collector{
		pg:           pg,
		sqlite:       sqlite,
		interval:     interval,
		targetTables: []string{},
		stopChan:     make(chan struct{}),
	}
}

func (c *Collector) SetRetentionDays(days int) {
	c.retentionDays = days
}

func (c *Collector) AddTargetTable(table string) {
	c.targetTables = append(c.targetTables, table)
}

func (c *Collector) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := c.CollectOnce(ctx); err != nil {
					fmt.Printf("[%s] collection failed: %v\n", time.Now().Format("15:04:05"), err)
				}
			case <-c.stopChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (c *Collector) CollectOnce(ctx context.Context) error {
	timestamp := time.Now().Format("15:04:05")

	current, err := c.pg.FetchCurrentWorkloadSnapshot(ctx)
	if err != nil {
		return err
	}

	previous, _ := c.sqlite.FetchLastOriginalSnapshots(ctx)

	deltas := c.computeDelta(current, previous)

	if len(deltas) > 0 {
		_ = c.sqlite.RecordDeltaSnapshots(ctx, deltas)
	}
	_ = c.sqlite.SynchronizeOriginalSnapshots(ctx, current)

	updatedTables := []string{}
	for _, table := range c.targetTables {
		metrics, err := c.pg.FetchTableDynamicMetrics(ctx, table)
		if err == nil {
			_ = c.sqlite.RecordTableDynamicMetrics(ctx, metrics)
			updatedTables = append(updatedTables, table)
		}
	}

	fmt.Printf("[%s] Metric collection complete: %d queries recorded | Tables: %v\n",
		timestamp, len(deltas), updatedTables)

	if c.retentionDays > 0 {
		_ = c.sqlite.MaintenancePurgeData(ctx, c.retentionDays)
	}

	return nil
}

func (c *Collector) computeDelta(curr []types.WorkloadSnapshot, prev map[int64]types.WorkloadSnapshot) []types.WorkloadSnapshot {
	var deltas []types.WorkloadSnapshot
	for _, s := range curr {
		p, exists := prev[s.QueryID]
		if !exists {
			deltas = append(deltas, s)
			continue
		}

		deltaCalls := s.Calls - p.Calls
		if deltaCalls < 0 {
			deltas = append(deltas, s)
			continue
		}

		if deltaCalls > 0 {
			delta := s
			delta.Calls = deltaCalls
			delta.TotalTime = s.TotalTime - p.TotalTime
			delta.Rows = s.Rows - p.Rows
			delta.SharedBlksHit = s.SharedBlksHit - p.SharedBlksHit
			delta.SharedBlksRead = s.SharedBlksRead - p.SharedBlksRead
			deltas = append(deltas, delta)
		}
	}
	return deltas
}

func (c *Collector) Stop() {
	close(c.stopChan)
}
