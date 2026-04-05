package db

import (
	"context"
	"log"
	"time"
)

// Collector orchestrates the background data collection from Postgres to SQLite.
type Collector struct {
	pg            PostgresClient
	sqlite        SQLiteClient
	interval      time.Duration
	retentionDays int
	targetTables  []string
	stopChan      chan struct{}
}

// NewCollector creates a new Collector instance.
func NewCollector(pg PostgresClient, sqlite SQLiteClient, interval time.Duration) *Collector {
	return &Collector{
		pg:            pg,
		sqlite:        sqlite,
		interval:      interval,
		retentionDays: 7,          // Default retention: 7 days
		targetTables:  []string{}, // Initialize with an empty list
		stopChan:      make(chan struct{}),
	}
}

// SetRetentionDays sets the data retention period in days.
// 데이터 보존 기간(일 단위)을 설정합니다.
func (c *Collector) SetRetentionDays(days int) {
	c.retentionDays = days
}

// AddTargetTable adds a table to the monitoring list for dynamic metrics collection.
// 동적 지표 수집을 위해 모니터링 대상 테이블을 추가합니다.
func (c *Collector) AddTargetTable(tableName string) {
	// Check for duplicates
	// 중복 여부를 확인합니다.
	for _, t := range c.targetTables {
		if t == tableName {
			return
		}
	}
	c.targetTables = append(c.targetTables, tableName)
}

// Start begins the background collection process in a separate goroutine.
// 별도의 고루틴에서 백그라운드 수집 프로세스를 시작합니다.
func (c *Collector) Start(ctx context.Context) {
	ticker := time.NewTicker(c.interval)

	go func() {
		log.Printf("Background collector started with interval: %v", c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Perform collection
				// 데이터 수집을 수행합니다.
				if err := c.collect(ctx); err != nil {
					log.Printf("Collection error: %v", err)
				}
			case <-c.stopChan:
				log.Println("Background collector stopping...")
				return
			case <-ctx.Done():
				log.Println("Background collector context cancelled")
				return
			}
		}
	}()
}

// collect performs a single round of data fetching and saving.
// 데이터 가져오기 및 저장의 단일 라운드를 수행합니다.
func (c *Collector) collect(ctx context.Context) error {

	// 1. Fetch from Postgres (Workload Snapshots)
	// internal/db/workload.go - PostgreSQL에 쿼리를 통해 pg_stat_statements 조회 후 []WorkloadSnapshot 형태로 반환
	snapshots, err := c.pg.FetchWorkload(ctx)
	if err != nil {
		return err
	}

	// 받은 Snapshot이 있을 때만 SQLite에 저장
	if len(snapshots) > 0 {
		// 2. Save Snapshots to SQLite
		// internal/db/activity.go - 가져온 스냅샷을 로컬 SQLite에 저장
		if err := c.sqlite.SaveSnapshots(snapshots); err != nil {
			return err
		}
		log.Printf("Successfully collected %d workload snapshots", len(snapshots))
	}

	// 3. Fetch Dynamic Metrics for Target Tables (v3.0 Model)
	// 대상 테이블들에 대해 v3.0용 동적 지표를 수집합니다.
	for _, table := range c.targetTables {
		// internal/db/workload.go - 대상 테이블에 대한 지표(pg_total_relation_size, pg_stat_replication, pg_stat_activity)를 수집
		// pg_total_relation_size: 특정 테이블이 디스크에서 차지하고 있는 전체 용량을 바이트 단위로 변환한 값
		// pg_stat_replication : Primary(Master) 서버에 연결된 Replica(standby) 서버들의 목록, 동기화 상태, 데이터 전송 및 적용 위치(LSN - Log Squence Number)
		// pg_stat_activity : 현재 타겟 DB에 접속해 있는 모든 연결(세션)의 실시간 활동 상태
		metrics, err := c.pg.GetTableDynamicMetrics(ctx, table)
		if err != nil {
			log.Printf("Error fetching metrics for table %s: %v", table, err)
			continue
		}

		// 4. Save Table Metrics to SQLite
		// internal/db/activity.go - 수집된 지표를 SQLite에 저장.
		if err := c.sqlite.SaveTableMetrics(metrics); err != nil {
			log.Printf("Error saving metrics for table %s: %v", table, err)
		} else {
			log.Printf("Successfully collected dynamic metrics for table: %s", table)
		}
	}

	// 5. Purge old snapshots (Retention Policy)
	// internal/db/activity.go - 설정된 보존 기간을 초과한 오래된 데이터를 정리하고 VACUUM 호출로 물리 공간 회수.
	if err := c.sqlite.PurgeOldSnapshots(c.retentionDays); err != nil {
		log.Printf("Error purging old snapshots: %v", err)
	}

	return nil
}

// Stop signals the background collector to stop.
// 백그라운드 컬렉터에 중단 신호를 보냅니다.
func (c *Collector) Stop() {
	close(c.stopChan)
}
