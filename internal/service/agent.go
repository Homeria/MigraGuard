package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Homeria/MigraGuard/internal/db"
)

// AgentService는 백그라운드 워크로드 수집 프로세스를 관리하는 서비스 레이어입니다.
type AgentService struct {
	pg            db.PostgresClient
	sqlite        db.SQLiteClient
	interval      time.Duration
	retentionDays int
	targetTables  []string // 감시 대상 테이블 목록
}

// NewAgentService는 AgentService의 인스턴스를 생성합니다.
func NewAgentService(pg db.PostgresClient, sqlite db.SQLiteClient, interval time.Duration, retentionDays int) *AgentService {
	return &AgentService{
		pg:            pg,
		sqlite:        sqlite,
		interval:      interval,
		retentionDays: retentionDays,
		targetTables:  []string{},
	}
}

// SetTargetTables는 쉼표로 구분된 문자열을 받아 감시 대상 테이블 목록을 설정합니다.
func (s *AgentService) SetTargetTables(tables string) {
	if tables == "" {
		return
	}
	rawList := strings.Split(tables, ",")
	for _, t := range rawList {
		s.targetTables = append(s.targetTables, strings.TrimSpace(t))
	}
}

// Run은 지속적인 수집 루프를 시작합니다.
func (s *AgentService) Run(ctx context.Context) error {
	collector := db.NewCollector(s.pg, s.sqlite, s.interval)
	collector.SetRetentionDays(s.retentionDays)

	for _, table := range s.targetTables {
		collector.AddTargetTable(table)
	}

	collector.Start(ctx)

	fmt.Printf("✅ MigraGuard 에이전트 서비스 시작됨. (간격: %v, 보존: %d일)\n", s.interval, s.retentionDays)
	if len(s.targetTables) > 0 {
		fmt.Printf("🔍 감시 대상 테이블: %v\n", s.targetTables)
	}
	fmt.Println("📡 워크로드 지표를 수집 중입니다... 중단하려면 Ctrl+C를 누르세요.")

	<-ctx.Done()

	fmt.Println("\n🛑 에이전트 서비스를 안전하게 종료하는 중...")
	collector.Stop()

	time.Sleep(1 * time.Second)
	return nil
}
