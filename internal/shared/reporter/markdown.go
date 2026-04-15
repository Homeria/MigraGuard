package reporter

import (
	"fmt"
	"strings"

	"github.com/Homeria/MigraGuard/internal/analyzer"
)

// MarkdownReporter는 CI/CD 연동 및 문서화에 적합한 상세 마크다운 리포트를 생성합니다.
type MarkdownReporter struct{}

// NewMarkdownReporter는 새로운 MarkdownReporter를 생성합니다.
func NewMarkdownReporter() *MarkdownReporter {
	return &MarkdownReporter{}
}

// Write는 분석 결과를 마크다운 포맷으로 변환하여 출력합니다.
//
// Args:
//   - results: 분석 정보 목록
//   - reports: 리스크 보고서 목록
//
// Returns:
//   - error: 출력 실패 시 반환
func (r *MarkdownReporter) Write(results []analyzer.AnalysisResult, reports []*analyzer.RiskAnalysisReport) error {
	fmt.Println("# 🛡️ MigraGuard 정밀 리스크 분석 리포트")
	
	for i, res := range results {
		report := reports[i]

		fmt.Printf("\n## 📋 분석 대상: `%s`\n", res.TableName)
		
		fmt.Println("\n### 🔍 SQL 원문")
		fmt.Printf("```sql\n%s\n```\n", res.RawQuery)
		
		fmt.Println("\n### 📑 종합 요약")
		fmt.Printf("- **최종 등급**: **%s**\n", report.RiskLevel)
		fmt.Printf("- **위험 점수**: `%.2f / 100`\n", report.RiskScore)
		fmt.Printf("- **DDL 작업 유형**: `%s` (재작성 필요: `%v`)\n", res.Operation, res.RewriteRequired)

		fmt.Println("\n### ⚙️ 리스크 수치 상세 ($M/M/1$ Queuing Model)")
		fmt.Println("| 지표명 | 수치 | 설명 |")
		fmt.Println("| :--- | :--- | :--- |")
		fmt.Printf("| 예상 작업 시간 ($T_{ddl}$) | %.2f ms | DDL 자체 수행 예상 시간 |\n", report.EstimatedDDLTime)
		fmt.Printf("| 총 블로킹 시간 ($T_{block}$) | %.2f ms | 락 점유로 인한 서비스 정지 시간 |\n", report.BlockingTime)
		fmt.Printf("| 예측 최대 연결 ($C_{peak}$) | %d | 작업 중 폭증할 커넥션 수 |\n", report.PeakConnections)
		fmt.Printf("| 예상 회복 시간 ($T_{rec}$) | %.2f ms | 시스템 정상화 소요 시간 |\n", report.RecoveryTime)

		fmt.Println("\n### 📊 트래픽 분석 데이터")
		fmt.Printf("- **분석 기준 TPS**: `%.2f` (소스: %s)\n", report.BaseTPS, report.TPSSource)
		fmt.Println("| 항목 | 수치 |")
		fmt.Println("| :--- | :--- |")
		fmt.Printf("| 실시간 TPS | %.2f |\n", report.CurrentTPS)
		fmt.Printf("| 1시간 평균 TPS | %.2f |\n", report.AvgTPS1h)
		fmt.Printf("| 24시간 피크 TPS | %.2f |\n", report.PeakTPS24h)
		fmt.Printf("| 현재 활성 커넥션 | %d |\n", report.ActiveConns)
		fmt.Printf("| 테이블 크기 | %.2f MB |\n", float64(report.TableSize)/(1024*1024))

		if len(report.TopQueries) > 0 {
			fmt.Println("\n### 🔥 현재 DB 주요 부하 쿼리 (Top 3)")
			fmt.Println("| 점유율 | 쿼리 내용 (일부) | ID |")
			fmt.Println("| :--- | :--- | :--- |")
			for _, q := range report.TopQueries {
				cleanQuery := strings.ReplaceAll(q.QueryText, "\n", " ")
				if len(cleanQuery) > 80 { cleanQuery = cleanQuery[:77] + "..." }
				fmt.Printf("| `%.1f%%` | `%s` | %d |\n", q.Impact, cleanQuery, q.QueryID)
			}
		}

		fmt.Println("\n### 💡 결론 및 배포 제안")
		if report.RiskLevel == "Safe" {
			fmt.Println("> ✅ **안전**: 서비스 영향이 미미할 것으로 판단됩니다.")
		} else {
			fmt.Printf("> 🛑 **주의/위험**: 골든 타임(**%s**, 예상 %.1f TPS) 활용을 권장합니다.\n", report.SafeWindow, report.SafeWindowTPS)
		}
		fmt.Println("\n---")
	}
	return nil
}
