package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Homeria/MigraGuard/internal/collector"
	"github.com/Homeria/MigraGuard/internal/shared/types"
)

// AgentService는 백그라운드 지표 수집 프로세스를 관리하고 운용하는 어플리케이션 레이어 서비스입니다.
type AgentService struct {
	pg            types.PostgresClient
	sqlite        types.SQLiteClient
	interval      time.Duration
	retentionDays int
	targetTables  []string
}

// NewAgentService는 AgentService의 새로운 인스턴스를 생성합니다.
func NewAgentService(pg types.PostgresClient, sqlite types.SQLiteClient, interval time.Duration, retentionDays int) *AgentService {
	return &AgentService{
		pg:            pg,
		sqlite:        sqlite,
		interval:      interval,
		retentionDays: retentionDays,
		targetTables:  []string{},
	}
}

// SetTargetTables는 콤마로 구분된 문자열을 파싱하여 집중 모니터링 테이블 목록을 설정합니다.
func (s *AgentService) SetTargetTables(tables string) {
	if tables == "" {
		return
	}
	rawList := strings.Split(tables, ",")
	for _, t := range rawList {
		s.targetTables = append(s.targetTables, strings.TrimSpace(t))
	}
}

// Run은 수집기(Collector)를 초기화하고 지속적인 지표 수집 루프를 시작합니다.
func (s *AgentService) Run(ctx context.Context) error {
	col := collector.NewCollector(s.pg, s.sqlite, s.interval)
	col.SetRetentionDays(s.retentionDays)

	for _, table := range s.targetTables {
		col.AddTargetTable(table)
	}

	col.Start(ctx)

	fmt.Printf("✅ MigraGuard 에이전트 서비스 가동됨. (수집 주기: %v, 보존 기간: %d일)\n", s.interval, s.retentionDays)
	if len(s.targetTables) > 0 {
		fmt.Printf("🔍 집중 모니터링 대상 테이블: %v\n", s.targetTables)
	}
	fmt.Println("📡 실시간 워크로드 지표를 수집 중입니다... (종료하시려면 Ctrl+C를 누르세요)")

	<-ctx.Done()

	fmt.Println("\n🛑 에이전트 서비스를 안전하게 종료하는 중...")
	col.Stop()

	return nil
}
