package simulation

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// TrafficProfile은 시뮬레이션의 성격을 정의합니다.
type TrafficProfile string

const (
	ProfileSteady    TrafficProfile = "steady"     // 평상시: 일정한 읽기/쓰기 비율
	ProfileFlashSale TrafficProfile = "flash-sale" // 이벤트: 쓰기 폭증 및 락 경합 극대화
	ProfileReadHeavy TrafficProfile = "read-heavy" // 조회 위주: SELECT 부하 중심
)

// LoadGenerator는 실서비스 모사 트래픽을 생성하는 엔진입니다.
type LoadGenerator struct {
	db       *sql.DB
	conns    int             // 동시 실행 고루틴 수
	profile  TrafficProfile  // 현재 시뮬레이션 프로파일
	stopChan chan struct{}
}

// NewLoadGenerator는 지정된 DB와 설정으로 제네레이터를 생성합니다.
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

// Run은 백그라운드에서 트래픽 생성을 시작합니다.
func (g *LoadGenerator) Run(ctx context.Context) {
	log.Printf("🚀 부하 생성기 시작! 프로파일: %s, Concurrency: %d", g.profile, g.conns)

	for i := 0; i < g.conns; i++ {
		go g.worker(ctx, i)
	}

	<-ctx.Done()
	log.Println("🛑 부하 생성기 중단 중...")
}

// worker는 개별 고루틴에서 지정된 프로파일에 따라 쿼리를 수행합니다.
func (g *LoadGenerator) worker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 1. 현재 시간대 기반 부하 강도(Intensity) 계산
			intensity := g.calculateHourlyIntensity()
			
			// 2. 프로파일에 따른 작업 실행
			g.executeScenario(ctx, intensity)

			// 3. 지연 시간 (Intensity가 높을수록 지연 시간을 짧게 가져감)
			sleepBase := 100.0
			if g.profile == ProfileFlashSale {
				sleepBase = 10.0
			}
			actualSleep := math.Max(10.0, sleepBase/intensity)
			time.Sleep(time.Duration(actualSleep) * time.Millisecond)
		}
	}
}

// calculateHourlyIntensity는 24시간 트래픽 곡선을 생성합니다.
func (g *LoadGenerator) calculateHourlyIntensity() float64 {
	hour := time.Now().Hour()
	val := math.Sin(float64(hour-9) * math.Pi / 12.0)
	return (val + 1.5) / 2.5 
}

// executeScenario는 프로파일 비율에 따라 쿼리를 선택 실행합니다.
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

// --- 실제 쿼리 시나리오 (로그 출력 추가) ---

// browse는 사용자 및 상품 정보를 조회합니다.
func (g *LoadGenerator) browse(ctx context.Context) {
	userID := rand.Intn(100) + 1
	var username string
	_ = g.db.QueryRowContext(ctx, "SELECT username FROM users WHERE id = $1", userID).Scan(&username)

	category := ARRAY_CATEGORIES[rand.Intn(len(ARRAY_CATEGORIES))]
	rows, _ := g.db.QueryContext(ctx, "SELECT name, price FROM products WHERE category = $1 LIMIT 10", category)
	if rows != nil { rows.Close() }

	// 🔍 조회 로그 출력
	log.Printf("[INFO] Browse: 유저 %d번이 '%s' 카테고리 상품들을 둘러봅니다.", userID, category)
}

var ARRAY_CATEGORIES = []string{"Electronics", "Clothing", "Books", "Home"}

// order는 주문을 생성하고 재고를 차감합니다.
func (g *LoadGenerator) order(ctx context.Context) {
	tx, err := g.db.BeginTx(ctx, nil)
	if err != nil { return }
	defer tx.Rollback()

	userID := rand.Intn(100) + 1
	productID := rand.Intn(50) + 1
	qty := rand.Intn(2) + 1

	// 1. 주문 생성
	var orderID int
	err = tx.QueryRowContext(ctx, 
		"INSERT INTO orders (user_id, total_amount, status) VALUES ($1, $2, 'paid') RETURNING id", 
		userID, float64(qty)*100.0).Scan(&orderID)
	if err != nil { return }

	// 2. 주문 상세 기록
	_, err = tx.ExecContext(ctx, 
		"INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES ($1, $2, $3, 100.0)", 
		orderID, productID, qty)
	if err != nil { return }

	// 3. 재고 차감
	_, err = tx.ExecContext(ctx, 
		"UPDATE products SET stock_quantity = stock_quantity - $1 WHERE id = $2 AND stock_quantity >= $1", 
		qty, productID)
	if err != nil { return }

	if err := tx.Commit(); err == nil {
		// 💰 주문 완료 로그 출력
		log.Printf("[SUCCESS] Order: 유저 %d번이 상품 %d번을 %d개 주문했습니다! (Order: %d)", userID, productID, qty, orderID)
	}
}
