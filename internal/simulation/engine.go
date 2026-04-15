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

// TrafficProfile은 부하 생성 시나리오의 성격을 정의합니다.
type TrafficProfile string

const (
	ProfileSteady    TrafficProfile = "steady"
	ProfileFlashSale TrafficProfile = "flash-sale"
	ProfileReadHeavy TrafficProfile = "read-heavy"
)

// LoadGenerator는 운영 환경의 트래픽을 모사하여 DB에 실제 부하를 가하는 엔진입니다.
type LoadGenerator struct {
	db       *sql.DB
	conns    int
	profile  TrafficProfile
	stopChan chan struct{}
	browseCount uint64 // 실시간 탐색 건수 집계
	orderCount  uint64 // 실시간 주문 건수 집계
}

// NewLoadGenerator는 부하 생성기 본체를 초기화합니다.
//
// Args:
//   - dsn: 타겟 DB DSN
//   - conns: 병렬 워커 수
//   - profile: 부하 시나리오 유형
//
// Returns:
//   - *LoadGenerator: 엔진 인스턴스
//   - error: 연결 실패 에러
func NewLoadGenerator(dsn string, conns int, profile TrafficProfile) (*LoadGenerator, error) {
	// 1. DB 커넥션 오픈
	db, err := sql.Open("pgx", dsn)
	if err != nil { return nil, fmt.Errorf("DB 연결 실패: %w", err) }
	db.SetMaxOpenConns(conns)
	return &LoadGenerator{db: db, conns: conns, profile: profile, stopChan: make(chan struct{})}, nil
}

// Run은 병렬 워커들을 기동하여 부하 생성을 개시하고 통계 루프를 관리합니다.
//
// Args:
//   - ctx: 실행 컨텍스트
func (g *LoadGenerator) Run(ctx context.Context) {
	log.Printf("🚀 부하 생성기 가동! 프로파일: [%s], 동시성: %d", g.profile, g.conns)
	
	// 1. 통계 리포팅 루프 기동
	go g.reportingLoop(ctx)

	// 2. 워커 풀 생성 및 기동
	for i := 0; i < g.conns; i++ { go g.worker(ctx, i) }

	// 3. 종료 대기
	<-ctx.Done()
	log.Println("🛑 부하 생성 시뮬레이션을 중단합니다...")
}

// reportingLoop은 10초마다 현재의 트래픽 생성 현황을 로그로 출력합니다.
//
// Args:
//   - ctx: 종료 제어 컨텍스트
func (g *LoadGenerator) reportingLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done(): return
		case <-ticker.C:
			// 1. 카운터 원자적 교체 및 현재 강도 산출
			browses := atomic.SwapUint64(&g.browseCount, 0)
			orders := atomic.SwapUint64(&g.orderCount, 0)
			intensity := g.calculateHourlyIntensity()
			log.Printf("📊 [부하 요약] 10초간: 탐색 %d건, 주문 %d건 | 강도: %.2f", browses, orders, intensity)
		}
	}
}

// worker는 개별 사용자 행동을 모사하여 지속적으로 쿼리를 발생시킵니다.
func (g *LoadGenerator) worker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done(): return
		default:
			// 1. 현재 시간대 강도 및 시나리오 실행
			intensity := g.calculateHourlyIntensity()
			g.executeScenario(ctx, intensity)

			// 2. 부하 강도 기반 대기 시간 조절
			sleepBase := 100.0
			if g.profile == ProfileFlashSale { sleepBase = 10.0 }
			actualSleep := math.Max(10.0, sleepBase/intensity)
			time.Sleep(time.Duration(actualSleep) * time.Millisecond)
		}
	}
}

// calculateHourlyIntensity는 사인파 공식에 따라 시간대별 유동적 부하 강도를 산출합니다.
func (g *LoadGenerator) calculateHourlyIntensity() float64 {
	hour := time.Now().Hour()
	val := math.Sin(float64(hour-9) * math.Pi / 12.0)
	return (val + 1.5) / 2.5 
}

// executeScenario는 시나리오 비율에 따라 탐색 또는 주문 트랜잭션을 선택 실행합니다.
func (g *LoadGenerator) executeScenario(ctx context.Context, intensity float64) {
	r := rand.Float64()
	switch g.profile {
	case ProfileReadHeavy: if r < 0.95 { g.browse(ctx) } else { g.order(ctx) }
	case ProfileFlashSale: if r < 0.3 { g.browse(ctx) } else { g.order(ctx) }
	default: if r < 0.7 { g.browse(ctx) } else { g.order(ctx) }
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
	userID, productID, qty := rand.Intn(100)+1, rand.Intn(50)+1, rand.Intn(2)+1
	var orderID int
	err = tx.QueryRowContext(ctx, "INSERT INTO orders (user_id, total_amount, status) VALUES ($1, $2, 'paid') RETURNING id", userID, float64(qty)*100.0).Scan(&orderID)
	if err != nil { return }
	_, _ = tx.ExecContext(ctx, "INSERT INTO order_items (order_id, product_id, quantity, unit_price) VALUES ($1, $2, $3, 100.0)", orderID, productID, qty)
	_, _ = tx.ExecContext(ctx, "UPDATE products SET stock_quantity = stock_quantity - $1 WHERE id = $2 AND stock_quantity >= $1", qty, productID)
	if err := tx.Commit(); err == nil { atomic.AddUint64(&g.orderCount, 1) }
}
