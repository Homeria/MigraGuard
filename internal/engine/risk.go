package engine

import (
	"context"
	"fmt"
	"math"

	"github.com/Homeria/MigraGuard/internal/db"
	"github.com/Homeria/MigraGuard/internal/parser"
)

// RiskConstants는 리스크 계산에 사용되는 인프라 성능 지표 및 임계값 설정입니다.
type RiskConstants struct {
	DiskIO   int64   `mapstructure:"disk_io"`   // 디스크 I/O 속도 (Bytes/sec) - 재작성(Rewrite) 시간 예측용
	TMeta    float64 `mapstructure:"t_meta"`    // 메타데이터 변경 기본 시간 (ms)
	MuMax    float64 `mapstructure:"mu_max"`    // 시스템의 최대 처리 가능 TPS (Mu_max)
	CMax     int     `mapstructure:"c_max"`     // DB 최대 허용 연결 수 (Connection Limit)
	TTimeout float64 `mapstructure:"t_timeout"` // 쿼리 타임아웃 임계치 (ms)
}

// DefaultRiskConstants는 일반적인 운영 환경을 위한 표준 권장값을 제공합니다.
func DefaultRiskConstants() RiskConstants {
	return RiskConstants{
		DiskIO:   100 * 1024 * 1024, // 100MB/s
		TMeta:    100.0,             // 100ms
		MuMax:    5000.0,            // 5000 TPS
		CMax:     100,               // 테스트를 위해 1000 -> 100으로 하향 조정
		TTimeout: 5000.0,            // 5s
	}
}

// RiskEngine은 MigraGuard v3.2 정밀 리스크 모델을 기반으로 DDL 작업의 위험도를 산출합니다.
type RiskEngine struct {
	pg        db.PostgresClient
	sqlite    db.SQLiteClient
	constants RiskConstants
	Verbose   bool
}

// RiskAnalysisReport는 위험도 평가의 상세 결과와 근거 데이터를 담고 있는 최종 리포트입니다.
type RiskAnalysisReport struct {
	RiskScore        float64 // 최종 위험 점수 (%)
	EstimatedDDLTime float64 // 예상 DDL 실행 시간 (T_ddl, ms)
	BlockingTime     float64 // 예상 전체 블로킹 시간 (T_block, ms)
	PeakConnections  int     // 예측되는 최대 동시 연결 수 (C_peak)
	RecoveryTime     float64 // 예측되는 시스템 회복 시간 (T_rec, ms)
	PermanentFailure bool    // 시스템 영구 마비(Deadlock/Overload) 여부
	RiskLevel        string  // 최종 위험 등급 (Danger, Warning, Safe)

	// 트래픽 상황 분석 (Baseline Analytics)
	CurrentTPS    float64 // 실시간 수집된 현재 TPS
	AvgTPS1h      float64 // 최근 1시간 평균 TPS
	PeakTPS24h    float64 // 최근 24시간 최대 피크 TPS
	SafeWindow    string  // 추천 배포 시간대 (예: "03:00")
	SafeWindowTPS float64 // 추천 시간대의 예상 평균 TPS
}

// NewRiskEngine은 필요한 의존성을 주입하여 RiskEngine 인스턴스를 생성합니다.
func NewRiskEngine(pg db.PostgresClient, sqlite db.SQLiteClient, constants RiskConstants) *RiskEngine {
	return &RiskEngine{
		pg:        pg,
		sqlite:    sqlite,
		constants: constants,
		Verbose:   false,
	}
}

// AnalyzeRisk는 5단계 리스크 평가 모델을 실행하여 DDL의 안전성을 진단합니다.
func (e *RiskEngine) AnalyzeRisk(ctx context.Context, analysis parser.AnalysisResult) (*RiskAnalysisReport, error) {

	if e.Verbose {
		fmt.Printf("\n[DEBUG] 🔍 리스크 분석 시작: %s (작업: %s)\n", analysis.TableName, analysis.Operation)
	}

	// 1. PostgreSQL 실시간 상태 정보 수집
	metrics, err := e.pg.FetchTableDynamicMetrics(ctx, analysis.TableName)
	if err != nil {
		return nil, fmt.Errorf("PostgreSQL 실시간 지표 수집 실패: %w", err)
	}

	report := &RiskAnalysisReport{}

	// 2. SQLite 기반 과거 트래픽 패턴 분석 (v3.2 Baseline Model)
	if e.sqlite != nil {
		// A. 실시간 TPS (가장 최근 수집된 델타값 기반)
		report.CurrentTPS, _ = e.sqlite.GetRecentTPSByDelta(analysis.TableName)

		// B. 과거 통계 (1시간 평균 및 24시간 피크)
		baseline, _ := e.sqlite.GetTableBaselineStatistics(analysis.TableName)
		if baseline != nil {
			report.AvgTPS1h = baseline.AvgTPS_1h
			report.PeakTPS24h = baseline.PeakTPS_24h
		}

		// C. 최적의 배포 윈도우 추천
		report.SafeWindow, report.SafeWindowTPS, _ = e.sqlite.IdentifySafestDeploymentWindow()

		// [Lambda 산출] 보수적 접근을 위해 현재, 평균(+20%), 피크(-20%) 중 최대값을 기준 트래픽으로 선정
		metrics.TPS = math.Max(report.CurrentTPS, math.Max(report.AvgTPS1h*1.2, report.PeakTPS24h*0.8))

		if e.Verbose {
			fmt.Printf("[DEBUG] 📡 트래픽 분석 결과: 현재=%.1f, 평균=%.1f, 피크=%.1f -> 분석 기준 Lambda=%.2f TPS\n",
				report.CurrentTPS, report.AvgTPS1h, report.PeakTPS24h, metrics.TPS)
		}
	}

	// [Step 1] 예상 DDL 실행 시간 (T_ddl)
	// 테이블 재작성(Full Rewrite) 여부에 따라 실행 시간 예측 방식 차별화
	if analysis.RewriteRequired {
		// 테이블 크기 / 디스크 I/O 속도 기반 예측
		report.EstimatedDDLTime = (float64(metrics.TableSize) / float64(e.constants.DiskIO)) * 1000.0
	} else {
		// 단순 메타데이터 변경 시간 적용
		report.EstimatedDDLTime = e.constants.TMeta
	}

	// [Step 2] 총 예상 블로킹 시간 (T_block)
	// T_block = P99 응답 시간 + DDL 실행 시간 + 복제 지연 시간
	report.BlockingTime = metrics.P99Time + report.EstimatedDDLTime + (metrics.ReplicationLag * 1000.0)

	// [Step 3] 예측 최대 동시 연결 수 (C_peak)
	// DDL 수행 중 대기하게 될 신규 요청 수(Lambda * T_block) + 기존 활성 연결 수
	lambdaPerMs := metrics.TPS / 1000.0
	report.PeakConnections = metrics.ActiveConnections + int(lambdaPerMs*report.BlockingTime)

	// [Step 4] 예측 시스템 회복 시간 (T_rec)
	// 쌓인 요청(C_peak - C_max)을 시스템 여유 대역폭(Mu_max - Lambda)으로 처리하는 데 걸리는 시간
	if metrics.TPS >= e.constants.MuMax {
		// 유입 트래픽이 시스템 한계를 넘어서면 영구 마비로 간주
		report.PermanentFailure = true
		report.RecoveryTime = math.Inf(1)
	} else {
		excessiveConns := float64(report.PeakConnections - e.constants.CMax)
		recoveryRatePerMs := (e.constants.MuMax - metrics.TPS) / 1000.0 
		if excessiveConns > 0 {
			report.RecoveryTime = excessiveConns / recoveryRatePerMs
		} else {
			report.RecoveryTime = 0
		}
	}

	// [Step 5] 최종 리스크 점수 및 등급 산출
	// 최대 연결 수(C_peak)가 시스템 한계(C_max)를 얼마나 위협하는지를 백분율로 산출
	report.RiskScore = (float64(report.PeakConnections) / float64(e.constants.CMax)) * 100.0

	if report.RiskScore >= 90.0 || report.PermanentFailure {
		report.RiskLevel = "Danger"
	} else if report.RiskScore >= 60.0 {
		report.RiskLevel = "Warning"
	} else {
		report.RiskLevel = "Safe"
	}

	return report, nil
}
