package reporter

import (
	"fmt"

	"github.com/Homeria/MigraGuard/internal/analyzer"
)

// ConsoleReporter는 분석 결과를 터미널에 상세하고 미려하게 출력하는 구현체입니다.
type ConsoleReporter struct{}

func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{}
}

func (r *ConsoleReporter) Write(results []analyzer.AnalysisResult, reports []*analyzer.RiskAnalysisReport) error {
	for i, res := range results {
		report := reports[i]

		fmt.Println("\n" + r.drawHeader(fmt.Sprintf("🛡️ MigraGuard 리스크 분석 보고서: %s", res.TableName)))

		// 1. 종합 요약 (Summary)
		fmt.Printf(" [SQL문] 원문: %s\n", r.truncate(res.RawQuery, 100))
		fmt.Printf(" [상태] 등급: %s | 위험 점수: %.2f%%\n", r.colorLevel(report.RiskLevel), report.RiskScore)
		fmt.Printf(" [작업] 유형: %s | 재작성 필요: %v | 영향 컬럼: %v\n", res.Operation, res.RewriteRequired, res.Columns)

		// 2. 트래픽 분석 근거 (Traffic Context)
		fmt.Println("\n 📊 트래픽 분석 데이터")
		fmt.Printf("  - 분석 기준 TPS (Lambda): %.2f (%s)\n", report.BaseTPS, report.TPSSource)
		fmt.Printf("  - 실시간 TPS: %.2f | 1시간 평균: %.2f | 24시간 피크: %.2f\n", 
			report.CurrentTPS, report.AvgTPS1h, report.PeakTPS24h)
		fmt.Printf("  - 현재 활성 커넥션: %d | 대상 테이블 크기: %.2f MB\n", 
			report.ActiveConns, float64(report.TableSize)/(1024*1024))

		// 3. 5단계 리스크 상세 지표 (Risk Metrics)
		fmt.Println("\n ⚙️ 5단계 수치 상세 (Queuing Model)")
		fmt.Printf("  - Step 1 [T_ddl]   예상 작업 시간: %.2f ms\n", report.EstimatedDDLTime)
		fmt.Printf("  - Step 2 [T_block] 총 블로킹 시간: %.2f ms\n", report.BlockingTime)
		fmt.Printf("  - Step 3 [C_peak]  예측 최대 연결: %d\n", report.PeakConnections)
		fmt.Printf("  - Step 4 [T_rec]   예상 회복 시간: %.2f ms (마비 여부: %v)\n", 
			report.RecoveryTime, report.PermanentFailure)

		// 4. 상위 부하 쿼리 현황 (Top Heavy Queries)
		if len(report.TopQueries) > 0 {
			fmt.Println("\n 🔥 현재 DB 주요 부하 쿼리 (Top 3)")
			for _, q := range report.TopQueries {
				fmt.Printf("  - [%.1f%%] %s (ID: %d)\n", q.Impact, r.truncate(q.QueryText, 60), q.QueryID)
			}
		}

		// 5. 배포 추천 사항 (Recommendation)
		fmt.Println("\n 💡 분석 결과 및 제안")
		if report.RiskLevel == "Safe" {
			fmt.Println("  ✅ 현재 트래픽 상황에서 배포가 매우 안전합니다.")
		} else {
			fmt.Printf("  ⚠️  가급적 배포를 미루고, 가장 저부하 시간대인 [%s](예상 %.1f TPS)에 수행하는 것을 권장합니다.\n", 
				report.SafeWindow, report.SafeWindowTPS)
		}
		fmt.Println(r.drawFooter())
	}
	return nil
}

func (r *ConsoleReporter) drawHeader(title string) string {
	line := "================================================================================"
	return fmt.Sprintf("%s\n %s\n%s", line, title, line)
}

func (r *ConsoleReporter) drawFooter() string {
	return "================================================================================"
}

func (r *ConsoleReporter) colorLevel(level string) string {
	switch level {
	case "Danger":
		return "\033[31m" + level + "\033[0m" // Red
	case "Warning":
		return "\033[33m" + level + "\033[0m" // Yellow
	case "Safe":
		return "\033[32m" + level + "\033[0m" // Green
	default:
		return level
	}
}

func (r *ConsoleReporter) truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit-3] + "..."
}
