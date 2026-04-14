package reporter

import (
	"github.com/Homeria/MigraGuard/internal/analyzer"
)

// Reporter는 리스크 분석 결과를 다양한 포맷으로 출력하기 위한 인터페이스입니다.
type Reporter interface {
	Write(results []analyzer.AnalysisResult, reports []*analyzer.RiskAnalysisReport) error
}
