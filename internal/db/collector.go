package db

import (
	"context"
	"log"
	"time"
)

// Collector orchestrates the background data collection from Postgres to SQLite.
// Postgres에서 SQLite로의 백그라운드 데이터 수집을 조율합니다.
type Collector struct {
	pg       *PostgresAdapter
	sqlite   *SQLiteAdapter
	interval time.Duration
	stopChan chan struct{}
}

// NewCollector creates a new Collector instance.
// 새로운 Collector 인스턴스를 생성합니다.
func NewCollector(pg *PostgresAdapter, sqlite *SQLiteAdapter, interval time.Duration) *Collector {
	return &Collector{
		pg:       pg,
		sqlite:   sqlite,
		interval: interval,
		stopChan: make(chan struct{}),
	}
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
	// 1. Fetch from Postgres
	// PostgreSQL에서 워크로드 스냅샷을 가져옵니다.
	snapshots, err := c.pg.FetchWorkload(ctx)
	if err != nil {
		return err
	}

	if len(snapshots) == 0 {
		return nil
	}

	// 2. Save to SQLite
	// 가져온 스냅샷을 로컬 SQLite에 저장합니다.
	if err := c.sqlite.SaveSnapshots(snapshots); err != nil {
		return err
	}

	log.Printf("Successfully collected %d workload snapshots", len(snapshots))
	return nil
}

// Stop signals the background collector to stop.
// 백그라운드 컬렉터에 중단 신호를 보냅니다.
func (c *Collector) Stop() {
	close(c.stopChan)
}
