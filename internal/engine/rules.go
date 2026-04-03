package engine

import "github.com/Homeria/MigraGuard/internal/parser"

// Rule represents a static analysis rule for DDL operations.
// DDL 작업에 대한 정적 분석 규칙을 나타냅니다.
type Rule struct {
	Name        string
	Description string
	Check       func(analysis parser.AnalysisResult) bool
}

// RuleEngine evaluates static rules against DDL analysis results.
// DDL 분석 결과에 대해 정적 규칙을 평가하는 엔진입니다.
type RuleEngine struct {
	rules []Rule
}

// NewRuleEngine creates a new RuleEngine with default rules.
// 기본 규칙들이 포함된 새로운 RuleEngine 인스턴스를 생성합니다.
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{
		rules: []Rule{
			// Future: Add static rules here (e.g., "Table must have primary key")
		},
	}
}
