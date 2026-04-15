package analyzer

import (
	"context"
	"fmt"
	"math"

	"github.com/Homeria/MigraGuard/internal/shared/types"
)

// RiskConstants는 리스크 계산 공식에 대입될 인프라의 성능 물리량 및 임계값 설정입니다.
type RiskConstants struct {
	DiskIO   int64   `mapstructure:"disk_io"`   // 디스크 I/O 처리 성능 (Bytes/sec)
	TMeta    float64 `mapstructure:"t_meta"`    // 메타데이터 변경 시 소요되는 기본 지연 시간 (ms)
	MuMax    float64 `mapstructure:"mu_max"`    // 시스템 최대 처리량 (TPS)
	CMax     int     `mapstructure:"c_max"`     // 데이터베이스 최대 동시 커넥션 제한 수
	TTimeout float64 `mapstructure:"t_timeout"` // 쿼리 실행 허용 최대 시간 (ms)
}

// DefaultRiskConstants는 표준적인 운영 환경을 가정한 기본 설정값을 반환합니다.
//
// Returns:
//   - RiskConstants: 초기화된 기본 설정값 구조체
func DefaultRiskConstants() RiskConstants {
	return RiskConstants{
		DiskIO:   100 * 1024 * 1024,
		TMeta:    100.0,
		MuMax:    5000.0,
		CMax:     100,
		TTimeout: 5000.0,
	}
}

// RiskEngine은 수집된 지표와 큐잉 이론을 결합하여 DDL의 위험도를 정량적으로 산출하는 핵심 엔진입니다.
type RiskEngine struct {
	pg        types.PostgresClient
	sqlite    types.SQLiteClient
	constants RiskConstants
	Verbose   bool
}

// RiskAnalysisReport는 5단계 리스크 모델의 결과물과 판단 근거 데이터를 집대성한 리포트 구조체입니다.
type RiskAnalysisReport struct {
	RiskScore        float64 // 종합 위험 점수 (0 ~ 100%)
	RiskLevel        string  // 위험 등급 (Safe, Warning, Danger)
	EstimatedDDLTime float64 // Step 1: 예상 작업 시간 (T_ddl, ms)
	BlockingTime     float64 // Step 2: 예상 블로킹 시간 (T_block, ms)
	PeakConnections  int     // Step 3: 예측 최대 커넥션 수 (C_peak)
	RecoveryTime     float64 // Step 4: 예상 회복 시간 (T_rec, ms)
	PermanentFailure bool    // Step 4: 시스템 마비 여부
	BaseTPS          float64 // 분석 기준 TPS (Lambda)
	TPSSource        string  // TPS 데이터 소스
	CurrentTPS       float64 // 실시간 TPS
	AvgTPS1h         float64 // 1시간 평균 TPS
	PeakTPS24h       float64 // 24시간 피크 TPS
	ActiveConns      int     // 현재 활성 커넥션
	TableSize        int64   // 대상 테이블 크기
	TopQueries       []types.TopQueryInfo
	SafeWindow       string
	SafeWindowTPS    float64
}

// NewRiskEngine은 의존성을 주입받아 새로운 리스크 분석 엔진을 생성합니다.
//
// Args:
//   - pg: Postgres 지표 수집 인터페이스
//   - sqlite: 과거 패턴 조회 인터페이스
//   - constants: 인프라 임계값 설정
//
// Returns:
//   - *RiskEngine: 초기화된 엔진 인스턴스
func NewRiskEngine(pg types.PostgresClient, sqlite types.SQLiteClient, constants RiskConstants) *RiskEngine {
	// 1. 엔진 인스턴스 초기화 및 의존성 바인딩
	return &RiskEngine{
		pg:        pg,
		sqlite:    sqlite,
		constants: constants,
		Verbose:   false,
	}
}

// AnalyzeRisk는 5단계 정밀 리스크 모델을 실행하여 특정 DDL의 배포 안전성을 진단합니다.
//
// Args:
//   - ctx: 실행 컨텍스트
//   - analysis: SQL 파서로부터 전달받은 DDL 분석 결과
//
// Returns:
//   - *RiskAnalysisReport: 수치화된 리스크 지표가 포함된 상세 보고서
//   - error: 지표 수집 실패 등 처리 중 발생한 에러
func (e *RiskEngine) AnalyzeRisk(ctx context.Context, analysis AnalysisResult) (*RiskAnalysisReport, error) {
	// 1. 실시간 데이터 수집: 테이블 크기, 응답시간, 커넥션 로드
	metrics, err := e.pg.FetchTableDynamicMetrics(ctx, analysis.TableName)
	if err != nil {
		return nil, fmt.Errorf("PostgreSQL 실시간 지표 수집 실패: %w", err)
	}

	report := &RiskAnalysisReport{
		ActiveConns: metrics.ActiveConnections,
		TableSize:   metrics.TableSize,
	}

	// 2. 트래픽 패턴 분석: 실시간/평균/피크 데이터 대조 및 기준 Lambda 선정
	if e.sqlite != nil {
		report.CurrentTPS, _ = e.sqlite.GetRecentTPSByDelta(analysis.TableName)
		baseline, _ := e.sqlite.GetTableBaselineStatistics(analysis.TableName)
		if baseline != nil {
			report.AvgTPS1h = baseline.AvgTPS_1h
			report.PeakTPS24h = baseline.PeakTPS_24h
		}
		report.SafeWindow, report.SafeWindowTPS, _ = e.sqlite.IdentifySafestDeploymentWindow()
		report.TopQueries, _ = e.sqlite.GetTopHeavyQueries(3)

		// 보수적 TPS 산출: 현재 vs 평균(+20%) vs 피크(-20%) 중 최댓값
		weightedAvg := report.AvgTPS1h * 1.2
		weightedPeak := report.PeakTPS24h * 0.8
		maxTPS := report.CurrentTPS
		report.TPSSource = "Real-time"
		if weightedAvg > maxTPS { maxTPS = weightedAvg; report.TPSSource = "1h-Avg (+20%)" }
		if weightedPeak > maxTPS { maxTPS = weightedPeak; report.TPSSource = "24h-Peak (-20%)" }
		report.BaseTPS = maxTPS
		metrics.TPS = maxTPS
	}

	// 3. [Step 1] T_ddl 계산: Rewrite 여부에 따른 예상 수행 시간 산출
	if analysis.RewriteRequired {
		report.EstimatedDDLTime = (float64(metrics.TableSize) / float64(e.constants.DiskIO)) * 1000.0
	} else {
		report.EstimatedDDLTime = e.constants.TMeta
	}

	// 4. [Step 2] T_block 계산: Lock 레벨 영향도(LockImpact) 적용
	// LockLevel 8 (AccessExclusive)은 100% 블로킹, 그 이하는 영향도 비례 축소
	lockImpact := 1.0
	if analysis.LockLevel <= LockLevelShareUpdateExcl {
		lockImpact = 0.1 // CONCURRENTLY 등은 서비스 영향 최소화 (약 10% 수준으로 가정)
	} else if analysis.LockLevel < LockLevelAccessExclusive {
		lockImpact = 0.5 // 중간 단계 Lock은 50% 수준 영향
	}

	report.BlockingTime = (metrics.P99Time + report.EstimatedDDLTime + (metrics.ReplicationLag * 1000.0)) * lockImpact

	// 5. [Step 3] C_peak 계산: Lambda * T_block 기반 대기 세션 예측
	lambdaPerMs := metrics.TPS / 1000.0
	report.PeakConnections = metrics.ActiveConnections + int(lambdaPerMs*report.BlockingTime)

	// 6. [Step 4] T_rec 계산: Mu_max 임계치 대조 및 해소 시간 산출
	if metrics.TPS >= e.constants.MuMax {
		report.PermanentFailure = true
		report.RecoveryTime = math.Inf(1)
	} else {
		excessiveConns := float64(report.PeakConnections - e.constants.CMax)
		if excessiveConns > 0 {
			recoveryRatePerMs := (e.constants.MuMax - metrics.TPS) / 1000.0
			report.RecoveryTime = excessiveConns / recoveryRatePerMs
		} else {
			report.RecoveryTime = 0
		}
	}

	// 7. [Step 5] 최종 점수 및 등급 확정
	report.RiskScore = (float64(report.PeakConnections) / float64(e.constants.CMax)) * 100.0

	// Lock Level에 따른 최소 위험 점수 보정 (수학적 수치가 낮아도 위험성 반영)
	baseRisk := 0.0
	switch analysis.LockLevel {
	case LockLevelAccessExclusive:
		if analysis.MetadataOnly {
			baseRisk = 30.0 // 단순 이름 변경, DEFAULT 변경 등은 Warning 수준 미만으로
		} else {
			baseRisk = 85.0 // TRUNCATE, DROP 등 무거운 작업은 Danger
		}
	case LockLevelExclusive, LockLevelShareRowExcl:
		baseRisk = 50.0 // Warning 수준
	case LockLevelShare:
		baseRisk = 20.0 // 기본 인덱스 등은 최소 점수 부여
	}

	if report.RiskScore < baseRisk {
		report.RiskScore = baseRisk
	}

	report.RiskLevel = EvaluateLevel(report.RiskScore, report.PermanentFailure)

	return report, nil
}
