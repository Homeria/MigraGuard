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

// TrafficProfile은 생성될 부하의 특성을 정의하는 타입입니다.
type TrafficProfile string

const (
	ProfileSteady    TrafficProfile = "steady"
	ProfileFlashSale TrafficProfile = "flash-sale"
	ProfileReadHeavy TrafficProfile = "read-heavy"
)

// LoadGenerator는 운영 환경의 트래픽을 모사하여 DB 부하를 생성하는 엔진입니다.
type LoadGenerator struct {
	db       *sql.DB
	conns    int
	profile  TrafficProfile
	stopChan chan struct{}

	// 통계 수집용 카운터 (atomic 사용)
	browseCount uint64
	orderCount  uint64
}

// NewLoadGenerator는 새로운 부하 생성기 인스턴스를 생성합니다.
func NewLoadGenerator(dsn string, conns int, profile TrafficProfile) (*LoadGenerator, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("부하 생성기 DB 연결 실패: %w", err)
	}
	db.SetMaxOpenConns(conns)

	return &LoadGenerator{
		db:       db,
		conns:    conns,
		profile:  profile,
		stopChan: make(chan struct{}),
	}, nil
}

// Run은 설정된 워커 수만큼 고루틴을 생성하여 부하 생성을 시작합니다.
func (g *LoadGenerator) Run(ctx context.Context) {
	log.Printf("🚀 부하 생성기 가동! 프로파일: [%s], 동시성: %d", g.profile, g.conns)

	// 리포팅 루프 시작 (10초마다 요약 보고)
	go g.reportingLoop(ctx)

	for i := 0; i < g.conns; i++ {
		go g.worker(ctx, i)
	}

	<-ctx.Done()
	log.Println("🛑 부하 생성 시뮬레이션을 중단합니다...")
}

// reportingLoop은 주기적으로 부하 발생 현황을 요약 출력합니다.
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
			intensity := g.calculateHourlyIntensity()
			
			log.Printf("📊 [부하 요약] 지난 10초간 트랜잭션: 탐색(Browse) %d건, 주문(Order) %d건 | 현재 강도: %.2f", 
				browses, orders, intensity)
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
		if r < 0.95 { g.browse(ctx) } else { g.order(ctx) }
	case ProfileFlashSale:
		if r < 0.3 { g.browse(ctx) } else { g.order(ctx) }
	default: // Steady
		if r < 0.7 { g.browse(ctx) } else { g.order(ctx) }
	}
}

func (g *LoadGenerator) browse(ctx context.Context) {
	userID := rand.Intn(100) + 1
	var username string
	_ = g.db.QueryRowContext(ctx, "SELECT username FROM users WHERE id = $1", userID).Scan(&username)

	category := ARRAY_CATEGORIES[rand.Intn(len(ARRAY_CATEGORIES))]
	rows, _ := g.db.QueryContext(ctx, "SELECT name, price FROM products WHERE category = $1 LIMIT 10", category)
	if rows != nil { rows.Close() }

	atomic.AddUint64(&g.browseCount, 1)
}

var ARRAY_CATEGORIES = []string{"Electronics", "Clothing", "Books", "Home"}

func (g *LoadGenerator) order(ctx context.Context) {
	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil { return }
	defer tx.Rollback()

	userID := rand.Intn(100) + 1
	productID := rand.Intn(50) + 1
	qty := rand.Intn(2) + 1

	var orderID int
	err = tx.QueryRowContext(ctx, 
		"INSERT INTO orders (user_id, total_amount, status) VALUES ($1, $2, 'paid') RETURNING id", 
		userID, float64(qty)*100.0).Scan(&orderID)
	if err != nil { return }

	_, _ = tx.ExecContext(ctx, 
		"INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES ($1, $2, $3, 100.0)", 
		orderID, productID, qty)

	_, _ = tx.ExecContext(ctx, 
		"UPDATE products SET stock_quantity = stock_quantity - $1 WHERE id = $2 AND stock_quantity >= $1", 
		qty, productID)

	if err := tx.Commit(); err == nil {
		atomic.AddUint64(&g.orderCount, 1)
	}
}
