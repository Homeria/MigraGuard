package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Homeria/MigraGuard/internal/collector"
	"github.com/Homeria/MigraGuard/internal/shared/types"
)

// AgentService는 백그라운드 지표 수집 프로세스를 관리하는 어플리케이션 레이어 서비스입니다.
type AgentService struct {
	pg            types.PostgresClient // 원시 지표 추출 대상 (PostgreSQL)
	sqlite        types.SQLiteClient   // 수집 지표 저장소 (SQLite)
	interval      time.Duration        // 수집 주기
	retentionDays int                  // 데이터 보존 기간
	targetTables  []string             // 정밀 감시 대상 테이블 목록
}

// NewAgentService는 에이전트를 가동하기 위한 서비스 본체를 조립합니다.
//
// Args:
//   - pg: Postgres 어댑터 구현체 (인프라 레이어)
//   - sqlite: SQLite 어댑터 구현체 (인프라 레이어)
//   - interval: 지표 수집 주기
//   - retentionDays: 데이터 보관 기한
//
// Returns:
//   - *AgentService: 초기화가 완료된 에이전트 서비스 객체
func NewAgentService(pg types.PostgresClient, sqlite types.SQLiteClient, interval time.Duration, retentionDays int) *AgentService {
	// 1. 서비스 필드 바인딩
	return &AgentService{
		pg:            pg,
		sqlite:        sqlite,
		interval:      interval,
		retentionDays: retentionDays,
		targetTables:  []string{},
	}
}

// SetTargetTables는 콤마로 구분된 테이블 목록 문자열을 파싱하여 슬라이스로 변환합니다.
//
// Args:
//   - tables: 사용자로부터 입력받은 테이블 목록 문자열
func (s *AgentService) SetTargetTables(tables string) {
	if tables == "" {
		return
	}
	// 1. 문자열 분리 및 목록 할당
	rawList := strings.Split(tables, ",")
	for _, t := range rawList {
		s.targetTables = append(s.targetTables, strings.TrimSpace(t))
	}
}

// Run은 에이전트 구동을 시작하고 종료 신호가 올 때까지 전체 라이프사이클을 유지합니다.
//
// Args:
//   - ctx: 외부(OS 신호 등)로부터 종료 요청을 수신하기 위한 컨텍스트
//
// Returns:
//   - error: 서비스 실행 중 발생한 치명적인 에러
func (s *AgentService) Run(ctx context.Context) error {

	// 1. 도메인 객체 생성: 수집 핵심 로직 'Collector' 인스턴스화
	col := collector.NewCollector(s.pg, s.sqlite, s.interval)
	col.SetRetentionDays(s.retentionDays)

	// 2. 관찰 대상 등록: 집중 관리 테이블 목록 전달
	for _, table := range s.targetTables {
		col.AddTargetTable(table)
	}

	// 3. 루프 기동: Collector 내부 고루틴 실행 및 데이터 흐름 개시
	col.Start(ctx)

	// 4. 로그 출력: 가동 상태 알림
	fmt.Printf("✅ MigraGuard 에이전트 서비스 가동됨. (주기: %v, 보존 기간: %d일)\n", s.interval, s.retentionDays)
	if len(s.targetTables) > 0 {
		fmt.Printf("🔍 집중 모니터링 대상 테이블: %v\n", s.targetTables)
	}
	fmt.Println("📡 실시간 워크로드 지표를 수집 중입니다... (종료하시려면 Ctrl+C)")

	// 5. 종료 대기: 컨텍스트 중단 신호 수신 대기
	<-ctx.Done()

	// 6. 정리 로직: Collector 루프 안전 중단
	fmt.Println("\n🛑 에이전트 서비스를 안전하게 종료하는 중...")
	col.Stop()

	// 7. 리소스 해제 지연: 안정적 종료를 위한 대기 시간 부여
	time.Sleep(1 * time.Second)
	return nil
}
