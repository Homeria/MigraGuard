package analyzer

import (
	"context"
	"fmt"
	"math"

	"github.com/Homeria/MigraGuard/internal/shared/types"
)

// RiskConstants는 리스크 계산에 사용되는 인프라 성능 지표 및 임계값 설정입니다.
type RiskConstants struct {
	DiskIO   int64   `mapstructure:"disk_io"`   // 디스크 I/O 속도 (Bytes/sec)
	TMeta    float64 `mapstructure:"t_meta"`    // 메타데이터 변경 기본 시간 (ms)
	MuMax    float64 `mapstructure:"mu_max"`    // 시스템 최대 처리 TPS
	CMax     int     `mapstructure:"c_max"`     // DB 최대 커넥션 수
	TTimeout float64 `mapstructure:"t_timeout"` // 쿼리 타임아웃 임계치 (ms)
}

func DefaultRiskConstants() RiskConstants {
	return RiskConstants{
		DiskIO:   100 * 1024 * 1024,
		TMeta:    100.0,
		MuMax:    5000.0,
		CMax:     100,
		TTimeout: 5000.0,
	}
}

type RiskEngine struct {
	pg        types.PostgresClient
	sqlite    types.SQLiteClient
	constants RiskConstants
	Verbose   bool
}

// RiskAnalysisReport는 더욱 상세해진 리스크 분석 결과 리포트입니다.
type RiskAnalysisReport struct {
	// 1. 종합 요약
	RiskScore float64 // 종합 위험 점수 (0-100)
	RiskLevel string  // 위험 등급 (Safe, Warning, Danger)

	// 2. 5단계 상세 지표 ($ Step-by-Step Metrics $)
	EstimatedDDLTime float64 // Step 1: 예상 작업 시간 (T_ddl, ms)
	BlockingTime     float64 // Step 2: 예상 블로킹 시간 (T_block, ms)
	PeakConnections  int     // Step 3: 예측 최대 커넥션 (C_peak)
	RecoveryTime     float64 // Step 4: 예상 시스템 회복 시간 (T_rec, ms)
	PermanentFailure bool    // Step 4: 시스템 마비 여부

	// 3. 트래픽 분석 근거
	BaseTPS       float64 // 분석에 사용된 기준 TPS (Lambda)
	TPSSource     string  // 기준 TPS 출처 (Real-time, 1h-Avg, 24h-Peak)
	CurrentTPS    float64 // 실시간 TPS
	AvgTPS1h      float64 // 1시간 평균 TPS
	PeakTPS24h    float64 // 24시간 피크 TPS
	ActiveConns   int     // 현재 활성 커넥션 수
	TableSize     int64   // 대상 테이블 크기 (Bytes)

	// 4. 상위 부하 쿼리 현황 (Top Heavy Queries)
	TopQueries []types.TopQueryInfo

	// 5. 추천 사항
	SafeWindow    string  // 가장 안전한 배포 시간대
	SafeWindowTPS float64 // 해당 시간대의 예상 TPS
}

func NewRiskEngine(pg types.PostgresClient, sqlite types.SQLiteClient, constants RiskConstants) *RiskEngine {
	return &RiskEngine{
		pg:        pg,
		sqlite:    sqlite,
		constants: constants,
		Verbose:   false,
	}
}

func (e *RiskEngine) AnalyzeRisk(ctx context.Context, analysis AnalysisResult) (*RiskAnalysisReport, error) {
	// 1. 실시간 지표 수집
	metrics, err := e.pg.FetchTableDynamicMetrics(ctx, analysis.TableName)
	if err != nil {
		return nil, fmt.Errorf("PostgreSQL 지표 수집 실패: %w", err)
	}

	report := &RiskAnalysisReport{
		ActiveConns: metrics.ActiveConnections,
		TableSize:   metrics.TableSize,
	}

	// 2. 과거 트래픽 분석 및 기준 TPS 결정
	if e.sqlite != nil {
		report.CurrentTPS, _ = e.sqlite.GetRecentTPSByDelta(analysis.TableName)
		baseline, _ := e.sqlite.GetTableBaselineStatistics(analysis.TableName)
		if baseline != nil {
			report.AvgTPS1h = baseline.AvgTPS_1h
			report.PeakTPS24h = baseline.PeakTPS_24h
		}
		report.SafeWindow, report.SafeWindowTPS, _ = e.sqlite.IdentifySafestDeploymentWindow()
		report.TopQueries, _ = e.sqlite.GetTopHeavyQueries(3) // 상위 3개 부하 쿼리

		// 보수적 TPS 산출 로직
		weightedAvg := report.AvgTPS1h * 1.2
		weightedPeak := report.PeakTPS24h * 0.8
		
		maxTPS := report.CurrentTPS
		report.TPSSource = "Real-time"
		
		if weightedAvg > maxTPS {
			maxTPS = weightedAvg
			report.TPSSource = "1h-Avg (+20%)"
		}
		if weightedPeak > maxTPS {
			maxTPS = weightedPeak
			report.TPSSource = "24h-Peak (-20%)"
		}
		report.BaseTPS = maxTPS
		metrics.TPS = maxTPS
	}

	// [Step 1] T_ddl 계산
	if analysis.RewriteRequired {
		report.EstimatedDDLTime = (float64(metrics.TableSize) / float64(e.constants.DiskIO)) * 1000.0
	} else {
		report.EstimatedDDLTime = e.constants.TMeta
	}

	// [Step 2] T_block 계산
	report.BlockingTime = metrics.P99Time + report.EstimatedDDLTime + (metrics.ReplicationLag * 1000.0)

	// [Step 3] C_peak 계산
	lambdaPerMs := metrics.TPS / 1000.0
	report.PeakConnections = metrics.ActiveConnections + int(lambdaPerMs*report.BlockingTime)

	// [Step 4] T_rec 계산
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

	// [Step 5] 최종 점수 및 등급
	report.RiskScore = (float64(report.PeakConnections) / float64(e.constants.CMax)) * 100.0
	report.RiskLevel = EvaluateLevel(report.RiskScore, report.PermanentFailure)

	return report, nil
}
