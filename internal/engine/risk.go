package engine

import (
	"context"
	"fmt"
	"math"

	"github.com/Homeria/MigraGuard/internal/db"
	"github.com/Homeria/MigraGuard/internal/parser"
)

// RiskConstants represents infrastructure-specific constants for risk calculation.
type RiskConstants struct {
	DiskIO   int64   `mapstructure:"disk_io"`   // Disk_IO: Bytes/sec
	TMeta    float64 `mapstructure:"t_meta"`    // T_meta (ms)
	MuMax    float64 `mapstructure:"mu_max"`    // Mu_max: Max system TPS
	CMax     int     `mapstructure:"c_max"`     // C_max: Connection limit
	TTimeout float64 `mapstructure:"t_timeout"` // T_timeout (ms)
}

// DefaultRiskConstants provides standard values for general environments.
func DefaultRiskConstants() RiskConstants {
	return RiskConstants{
		DiskIO:   100 * 1024 * 1024, // 100MB/s
		TMeta:    100.0,             // 100ms
		MuMax:    5000.0,            // 5000 TPS
		CMax:     1000,              // 1000 Conns
		TTimeout: 5000.0,            // 5s
	}
}

// RiskEngine computes the risk of a DDL operation using the MigraGuard v3.1 baseline model.
type RiskEngine struct {
	pg        db.PostgresClient
	sqlite    db.SQLiteClient
	constants RiskConstants
	Verbose   bool
}

// RiskAnalysisReport contains the detailed results of the risk evaluation.
type RiskAnalysisReport struct {
	RiskScore        float64 // Final risk percentage (%)
	EstimatedDDLTime float64 // T_ddl (ms)
	BlockingTime     float64 // T_block (ms)
	PeakConnections  int     // C_peak
	RecoveryTime     float64 // T_rec (ms)
	PermanentFailure bool    // Lambda >= Mu_max
	RiskLevel        string  // Danger, Warning, Safe

	// v3.1 Additional Context
	CurrentTPS    float64
	AvgTPS1h      float64
	PeakTPS24h    float64
	SafeWindow    string // Recommended deployment hour (e.g., "03:00")
	SafeWindowTPS float64
}

// NewRiskEngine creates a new RiskEngine instance.
func NewRiskEngine(pg db.PostgresClient, sqlite db.SQLiteClient, constants RiskConstants) *RiskEngine {
	return &RiskEngine{
		pg:        pg,
		sqlite:    sqlite,
		constants: constants,
		Verbose:   false,
	}
}

// AnalyzeRisk performs the 5-step risk assessment for a given DDL analysis result.
// [L-B02, L-B03] Uses SQLite baseline metrics to evaluate risk instantly.
func (e *RiskEngine) AnalyzeRisk(ctx context.Context, analysis parser.AnalysisResult) (*RiskAnalysisReport, error) {

	// 디버그 모드인 경우 분석 대상 테이블과 작업 유형을 출력하여 분석 진행 상황을 표시
	if e.Verbose {
		fmt.Printf("\n[DEBUG] 🔍 Analyzing risk for table: %s (Operation: %s)\n", analysis.TableName, analysis.Operation)
	}

	// 1. Get Dynamic Metrics from Postgres (Size, Conns, etc.)

	// internal/db/workload.go - GetTableDynamicMetrics 메서드를 호출하여 분석 대상 테이블의 현재 크기, 활성 연결 수, P99 응답 시간, 실시간 TPS 등을 조회
	metrics, err := e.pg.GetTableDynamicMetrics(ctx, analysis.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch dynamic metrics: %w", err)
	}

	// 디버그 모드인 경우 조회된 Postgres 지표를 출력하여 분석에 사용된 실시간 데이터 확인
	if e.Verbose {
		fmt.Printf("[DEBUG] 🐘 Postgres Metrics: Size=%d bytes, Conns=%d, P99=%.2fms, BaselineTPS=%.2f\n",
			metrics.TableSize, metrics.ActiveConnections, metrics.P99Time, metrics.TPS)
	}

	// 2. Initialize Risk Analysis Report

	// RiskAnalysisReport 구조체를 초기화하여 분석 결과를 저장할 준비
	report := &RiskAnalysisReport{}

	// [Step 0] v3.1 Baseline Analytics & Weighted TPS

	// GetTableBaselineStats 메서드를 호출하여 SQLite에 저장된 과거 데이터를 분석하여 최근 1시간 평균 및 24시간 최대 TPS를 산출
	if e.sqlite != nil {
		// Real-time TPS (delta from last 1-min snapshots)
		// internal/db/activity.go - GetRecentTPSDelta 메서드를 호출하여 최근 1분 간격으로 수집된 TPS 스냅샷을 기반으로 실시간 TPS 변화를 계산하여 보고서에 저장
		realtimeTPS, _ := e.sqlite.GetRecentTPSDelta(analysis.TableName)
		report.CurrentTPS = realtimeTPS

		// Historical stats from table_metrics
		// internal/db/activity.go - GetTableBaselineStats 메서드를 호출하여 SQLite에 저장된 과거 데이터를 분석하여 최근 1시간 평균 및 24시간 최대 TPS를 산출하여 보고서에 저장
		baseline, _ := e.sqlite.GetTableBaselineStats(analysis.TableName)
		if baseline != nil {
			report.AvgTPS1h = baseline.AvgTPS_1h
			report.PeakTPS24h = baseline.PeakTPS_24h
		}

		// Deployment Window Recommendation
		// internal/db/activity.go - GetSafeWindow 메서드를 호출하여 SQLite에 저장된 과거 데이터를 분석하여 최근 24시간 동안 가장 트래픽이 낮았던 시간대를 찾아 안전한 배포 시간으로 추천하여 보고서에 저장
		hour, avgTPS, _ := e.sqlite.GetSafeWindow()
		report.SafeWindow = hour
		report.SafeWindowTPS = avgTPS

		// [L34] Multi-Weighted TPS Calculation (Conservative Approach)
		// Lambda = Max(Current, Avg_1h * 1.2, Peak_24h * 0.8)
		// 실시간 TPS, 최근 1시간 평균 TPS의 20% 증가값, 24시간 최대 TPS의 20% 감소값 중 가장 높은 값을 최종 Lambda로 사용하여 위험 분석에 반영
		metrics.TPS = math.Max(report.CurrentTPS, math.Max(report.AvgTPS1h*1.2, report.PeakTPS24h*0.8))

		// 디버그 모드인 경우 SQLite에서 조회된 지표와 최종 Lambda 값을 출력하여 분석에 사용된 트래픽 데이터와 가중치 적용 결과 확인
		if e.Verbose {
			fmt.Printf("[DEBUG] 📡 SQLite Analytics: Current=%.1f, Avg_1h=%.1f, Peak_24h=%.1f -> Final Lambda=%.2f TPS\n",
				report.CurrentTPS, report.AvgTPS1h, report.PeakTPS24h, metrics.TPS)
		}
	}

	// [Step 1] Estimated DDL Time (T_ddl)

	// 분석 결과의 RewriteRequired 필드 값에 따라 DDL 작업이 테이블 재작성(예: ALTER TABLE)으로 인해 긴 다운타임이 예상되는지 여부를 판단
	// 테이블 재작성 작업인 경우, 테이블 크기와 Disk_IO 상수를 기반으로 DDL 작업에 필요한 시간을 밀리초 단위로 계산하여 보고서에 저장
	// 테이블 재작성 작업이 아닌 경우, T_meta 상수값을 DDL 시간으로 사용하여 보고서에 저장
	if analysis.RewriteRequired {
		report.EstimatedDDLTime = (float64(metrics.TableSize) / float64(e.constants.DiskIO)) * 1000.0
	} else {
		report.EstimatedDDLTime = e.constants.TMeta
	}

	// 디버그 모드인 경우 DDL 시간 계산에 사용된 테이블 크기, Disk_IO 상수, T_meta 상수, 그리고 최종 계산된 T_ddl 값을 출력하여 분석에 사용된 데이터와 계산 결과 확인
	if e.Verbose {
		fmt.Printf("[DEBUG] ⚙️ Step 1 (T_ddl): %.2f ms (RewriteRequired: %v, DiskIO: %d)\n",
			report.EstimatedDDLTime, analysis.RewriteRequired, e.constants.DiskIO)
	}

	// [Step 2] Total Blocking Time (T_block)
	// T_block = P99 + T_ddl + (ReplicationLag * 1000)
	// 분석 대상 테이블의 P99 응답 시간, 계산된 DDL 시간, 그리고 Replication Lag을 기반으로 총 예상 블로킹 시간을 밀리초 단위로 계산하여 보고서에 저장
	// Replication Lag은 초 단위로 제공되므로 밀리초로 변환하여 계산에 사용
	report.BlockingTime = metrics.P99Time + report.EstimatedDDLTime + (metrics.ReplicationLag * 1000.0)

	// 디버그 모드인 경우 T_block 계산에 사용된 P99 시간, DDL 시간, Replication Lag, 그리고 최종 계산된 T_block 값을 출력하여 분석에 사용된 데이터와 계산 결과 확인
	if e.Verbose {
		fmt.Printf("[DEBUG] ⚙️ Step 2 (T_block): %.2f ms (P99: %.1f, Lag: %.1f)\n",
			report.BlockingTime, metrics.P99Time, metrics.ReplicationLag)
	}

	// [Step 3] Peak Connections (C_peak)
	// LambdaPerMs = TPS / 1000
	// 분석 대상 테이블의 실시간 TPS를 기반으로 초당 트랜잭션 수를 밀리초 단위로 변환하여 LambdaPerMs를 계산
	// C_peak = ActiveConnections + (LambdaPerMs * T_block)
	// 현재 활성 연결 수에 LambdaPerMs와 T_block을 곱한 값을 더하여 DDL 작업 중 예상되는 최대 연결 수를 계산하여 보고서에 저장
	lambdaPerMs := metrics.TPS / 1000.0
	report.PeakConnections = metrics.ActiveConnections + int(lambdaPerMs*report.BlockingTime)

	// 디버그 모드인 경우 C_peak 계산에 사용된 활성 연결 수, LambdaPerMs, T_block, 그리고 최종 계산된 C_peak 값을 출력하여 분석에 사용된 데이터와 계산 결과 확인
	if e.Verbose {
		fmt.Printf("[DEBUG] ⚙️ Step 3 (C_peak): %d (Active: %d, Incoming during block: %.2f)\n",
			report.PeakConnections, metrics.ActiveConnections, lambdaPerMs*report.BlockingTime)
	}

	// [Step 4] Recovery Time (T_rec)
	// Lambda >= Mu_max인 경우, 시스템이 영구적으로 과부하 상태에 빠질 것으로 간주하여 T_rec을 무한대로 설정하고 PermanentFailure 플래그를 true로 설정
	// Lambda < Mu_max인 경우, C_peak가 C_max를 초과하는지 여부에 따라 T_rec을 계산
	// C_peak > C_max인 경우, T_rec = (C_peak - C_max) / ((Mu_max - Lambda) / 1000)
	// C_peak <= C_max인 경우, T_rec = 0 (즉시 회복 가능)
	// 계산된 T_rec 값을 보고서에 저장
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

	// 디버그 모드인 경우 T_rec 계산에 사용된 Lambda, Mu_max, C_peak, C_max, 그리고 최종 계산된 T_rec 값을 출력하여 분석에 사용된 데이터와 계산 결과 확인
	if e.Verbose {
		fmt.Printf("[DEBUG] ⚙️ Step 4 (T_rec): %.2f ms (MuMax: %.1f, RecoveryNeeded: %v)\n",
			report.RecoveryTime, e.constants.MuMax, report.PeakConnections > e.constants.CMax)
	}

	// [Step 5] Risk Score Calculation
	// Risk Score는 C_peak가 C_max를 초과하는 정도에 따라 계산
	// Risk Score = (C_peak / C_max) * 100, 단 C_peak가 C_max를 초과하는 경우에만 계산, 그렇지 않으면 0으로 설정
	report.RiskScore = (float64(report.PeakConnections) / float64(e.constants.CMax)) * 100.0

	// 디버그 모드인 경우 Risk Score 계산에 사용된 C_peak, C_max, 그리고 최종 계산된 Risk Score 값을 출력하여 분석에 사용된 데이터와 계산 결과 확인
	if e.Verbose {
		fmt.Printf("[DEBUG] ⚙️ Step 5 (RiskScore): %.2f%% (C_max: %d)\n", report.RiskScore, e.constants.CMax)
	}

	// Risk Level Classification
	// Risk Score와 PermanentFailure 플래그를 기반으로 위험 수준을 분류하여 보고서에 저장
	if report.RiskScore >= 90.0 || report.PermanentFailure {
		report.RiskLevel = "Danger"
	} else if report.RiskScore >= 60.0 {
		report.RiskLevel = "Warning"
	} else {
		report.RiskLevel = "Safe"
	}

	return report, nil
}
