package simulation

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sync/atomic"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// TrafficProfile defines the character of the load generation.
type TrafficProfile string

const (
	ProfileSteady    TrafficProfile = "steady"
	ProfileFlashSale TrafficProfile = "flash-sale"
	ProfileReadHeavy TrafficProfile = "read-heavy"
)

// LoadGenerator simulates production traffic to generate load on the DB.
type LoadGenerator struct {
	db          *sql.DB
	conns       int
	profile     TrafficProfile
	stopChan    chan struct{}
	browseCount uint64
	orderCount  uint64
	failCount   uint64
}

// NewLoadGenerator initializes the generator.
func NewLoadGenerator(dsn string, conns int, profile TrafficProfile) (*LoadGenerator, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("DB connection failed: %w", err)
	}
	db.SetMaxOpenConns(conns)
	return &LoadGenerator{db: db, conns: conns, profile: profile, stopChan: make(chan struct{})}, nil
}

// Run starts the workers and reporting loop.
func (g *LoadGenerator) Run(ctx context.Context) {
	log.Printf("[START] Starting Fintech Load Generator. Profile: [%s], Concurrency: %d", g.profile, g.conns)

	go g.reportingLoop(ctx)

	for i := 0; i < g.conns; i++ {
		go g.worker(ctx, i)
	}

	<-ctx.Done()
	log.Println("[STOP] Stopping load simulation...")
}

func (g *LoadGenerator) reportingLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			browses := atomic.SwapUint64(&g.browseCount, 0)
			orders := atomic.SwapUint64(&g.orderCount, 0)
			fails := atomic.SwapUint64(&g.failCount, 0)
			intensity := g.calculateHourlyIntensity()
			log.Printf("[STAT] [Load Summary] 10s: Browse %d, Success %d, Fail %d | Intensity: %.2f", browses, orders, fails, intensity)
		}
	}
}

func (g *LoadGenerator) worker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			intensity := g.calculateHourlyIntensity()
			g.executeScenario(ctx, intensity)

			sleepBase := 100.0
			if g.profile == ProfileFlashSale {
				sleepBase = 10.0
			}
			actualSleep := math.Max(10.0, sleepBase/intensity)
			time.Sleep(time.Duration(actualSleep) * time.Millisecond)
		}
	}
}

func (g *LoadGenerator) calculateHourlyIntensity() float64 {
	hour := time.Now().Hour()
	val := math.Sin(float64(hour-9) * math.Pi / 12.0)
	return (val + 1.5) / 2.5
}

func (g *LoadGenerator) executeScenario(ctx context.Context, intensity float64) {
	r := rand.Float64()
	switch g.profile {
	case ProfileReadHeavy:
		if r < 0.95 {
			g.browse(ctx)
		} else {
			g.orderTransaction(ctx)
		}
	case ProfileFlashSale:
		if r < 0.3 {
			g.browse(ctx)
		} else {
			g.orderTransaction(ctx)
		}
	default:
		if r < 0.7 {
			g.browse(ctx)
		} else {
			g.orderTransaction(ctx)
		}
	}
}

func (g *LoadGenerator) browse(ctx context.Context) {
	userID := rand.Intn(1000) + 1
	var username string
	_ = g.db.QueryRowContext(ctx, "SELECT username FROM users WHERE id = $1", userID).Scan(&username)
	
	category := ARRAY_CATEGORIES[rand.Intn(len(ARRAY_CATEGORIES))]
	rows, _ := g.db.QueryContext(ctx, "SELECT p.name, p.price, s.stock_quantity FROM products p JOIN inventory_stocks s ON p.id = s.product_id WHERE p.category = $1 LIMIT 10", category)
	if rows != nil {
		rows.Close()
	}
	atomic.AddUint64(&g.browseCount, 1)
}

var ARRAY_CATEGORIES = []string{"Electronics", "Clothing", "Books", "Home"}

// orderTransaction represents a complex fintech transaction flow.
func (g *LoadGenerator) orderTransaction(ctx context.Context) {
	tx, err := g.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		atomic.AddUint64(&g.failCount, 1)
		return
	}
	defer tx.Rollback()

	userID := rand.Intn(1000) + 1
	productID := rand.Intn(500) + 1
	qty := 1
	price := 100.0

	// 1. Check & Update Inventory (Contention Point)
	res, err := tx.ExecContext(ctx, "UPDATE inventory_stocks SET stock_quantity = stock_quantity - $1 WHERE product_id = $2 AND stock_quantity >= $1", qty, productID)
	if err != nil {
		atomic.AddUint64(&g.failCount, 1)
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		atomic.AddUint64(&g.failCount, 1)
		return
	}

	// 2. Update Balance (Hot Spot)
	res, err = tx.ExecContext(ctx, "UPDATE account_balances SET balance = balance - $1, last_updated_at = CURRENT_TIMESTAMP WHERE user_id = $2 AND balance >= $1", price, userID)
	if err != nil {
		atomic.AddUint64(&g.failCount, 1)
		return
	}
	affected, _ = res.RowsAffected()
	if affected == 0 {
		atomic.AddUint64(&g.failCount, 1)
		return
	}

	// 3. Create Order (Heavy Body)
	orderNo := fmt.Sprintf("ORD-%d-%d", time.Now().UnixNano(), idCounter)
	atomic.AddUint64(&idCounter, 1)
	
	var orderID int64
	err = tx.QueryRowContext(ctx, "INSERT INTO orders (order_no, user_id, total_amount, status, order_details) VALUES ($1, $2, $3, 'PAID', '{\"source\": \"sim\"}') RETURNING id", orderNo, userID, price).Scan(&orderID)
	if err != nil {
		atomic.AddUint64(&g.failCount, 1)
		return
	}

	// 4. Record Log (Massive Log)
	_, _ = tx.ExecContext(ctx, "INSERT INTO order_event_logs (order_id, event_type, raw_payload) VALUES ($1, 'PAYMENT_SUCCESS', '{\"sim\": true}')", orderID)

	if err := tx.Commit(); err == nil {
		atomic.AddUint64(&g.orderCount, 1)
	} else {
		atomic.AddUint64(&g.failCount, 1)
	}
}

var idCounter uint64
