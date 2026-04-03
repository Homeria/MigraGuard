package engine

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Homeria/MigraGuard/internal/db"
	"github.com/Homeria/MigraGuard/internal/parser"
)

// RiskConstants represents infrastructure-specific constants for risk calculation.
// 위험도 산출을 위한 인프라 종속적 상수들입니다.
type RiskConstants struct {
	DiskIO    int64   // Disk_IO: 초당 디스크 처리 속도 (Bytes/sec)
	TMeta     float64 // T_meta: 메타데이터 처리 평균 시간 (ms)
	MuMax     float64 // Mu_max: 시스템 최대 초당 처리량 (TPS)
	CMax      int     // C_max: 최대 커넥션 풀 크기
	TTimeout  float64 // T_timeout: API 타임아웃 임계치 (ms)
}

// DefaultRiskConstants provides standard values for general environments.
// 일반적인 환경을 위한 표준 위험도 상수값들을 제공합니다.
func DefaultRiskConstants() RiskConstants {
	return RiskConstants{
		DiskIO:   100 * 1024 * 1024, // 100MB/s
		TMeta:    100.0,            // 100ms
		MuMax:    5000.0,           // 5000 TPS
		CMax:     1000,             // 1000 Conns
		TTimeout: 5000.0,           // 5s
	}
}

// RiskEngine computes the risk of a DDL operation using the MigraGuard v3.0 model.
// MigraGuard v3.0 모델을 사용하여 DDL 작업의 위험도를 계산하는 엔진입니다.
type RiskEngine struct {
	pg        *db.PostgresAdapter
	sqlite    *db.SQLiteAdapter
	constants RiskConstants
}

// RiskAnalysisReport contains the detailed results of the risk evaluation.
// 위험도 평가의 상세 결과를 담고 있는 리포트 구조체입니다.
type RiskAnalysisReport struct {
	RiskScore        float64 // 최종 위험도 점수 (%)
	EstimatedDDLTime float64 // T_ddl (ms)
	BlockingTime     float64 // T_block (ms)
	PeakConnections  int     // C_peak
	RecoveryTime     float64 // T_rec (ms)
	PermanentFailure bool    // 영구 장애 여부 (Lambda >= Mu_max)
	RiskLevel        string  // Danger, Warning, Safe
}

// NewRiskEngine creates a new RiskEngine instance.
func NewRiskEngine(pg *db.PostgresAdapter, sqlite *db.SQLiteAdapter, constants RiskConstants) *RiskEngine {
	return &RiskEngine{
		pg:        pg,
		sqlite:    sqlite,
		constants: constants,
	}
}

// AnalyzeRisk performs the 5-step risk assessment for a given DDL analysis result.
// 주어진 DDL 분석 결과를 바탕으로 5단계 위험도 평가를 수행합니다.
func (e *RiskEngine) AnalyzeRisk(ctx context.Context, analysis parser.AnalysisResult) (*RiskAnalysisReport, error) {
	// 1. Get Dynamic Metrics
	// 운영 DB로부터 최신 동적 지표를 가져옵니다.
	metrics, err := e.pg.GetTableDynamicMetrics(ctx, analysis.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dynamic metrics: %w", err)
	}

	report := &RiskAnalysisReport{}

	// Step 1: DDL 물리적 소요 시간 추정 (T_ddl)
	// F_rewrite 플래그와 테이블 크기(S_table)를 사용하여 계산합니다.
	if analysis.RewriteRequired {
		report.EstimatedDDLTime = (float64(metrics.TableSize) / float64(e.constants.DiskIO)) * 1000.0
	} else {
		report.EstimatedDDLTime = e.constants.TMeta
	}

	// Step 2: 총 블로킹 시간 산출 (T_block)
	// T_block = T_p99 + T_ddl + Lag_repl
	report.BlockingTime = metrics.P99Time + report.EstimatedDDLTime + (metrics.ReplicationLag * 1000.0)

	// Step 3: 락 해제 직후 큐 스파이크량 산출 (C_peak)
	// C_peak = C_active + (Lambda * T_block)
	// Lambda (TPS)를 ms 단위로 변환하여 계산
	lambdaPerMs := metrics.TPS / 1000.0
	report.PeakConnections = metrics.ActiveConnections + int(lambdaPerMs * report.BlockingTime)

	// Step 4: 시스템 회복 소요 시간 산출 (T_rec)
	// T_rec = (C_peak - C_max) / (Mu_max - Lambda)
	if metrics.TPS >= e.constants.MuMax {
		report.PermanentFailure = true
		report.RecoveryTime = math.Inf(1)
	} else {
		recoveryNumerator := float64(report.PeakConnections - e.constants.CMax)
		recoveryDenominator := (e.constants.MuMax - metrics.TPS) / 1000.0 // per ms
		if recoveryNumerator > 0 {
			report.RecoveryTime = recoveryNumerator / recoveryDenominator
		} else {
			report.RecoveryTime = 0
		}
	}

	// Step 5: 최종 커넥션 고갈 위험도 (RiskScore)
	// RiskScore = (C_peak / C_max) * 100
	report.RiskScore = (float64(report.PeakConnections) / float64(e.constants.CMax)) * 100.0

	// Determine Risk Level
	if report.RiskScore >= 90.0 || report.PermanentFailure {
		report.RiskLevel = "Danger"
	} else if report.RiskScore >= 60.0 {
		report.RiskLevel = "Warning"
	} else {
		report.RiskLevel = "Safe"
	}

	return report, nil
}
