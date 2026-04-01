package parser

import (
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

// LockLevel represents the PostgreSQL lock level.
type LockLevel string

const (
	AccessExclusiveLock LockLevel = "AccessExclusiveLock"
	ExclusiveLock       LockLevel = "ExclusiveLock"
	ShareLock           LockLevel = "ShareLock"
	RowExclusiveLock    LockLevel = "RowExclusiveLock"
	UnknownLock         LockLevel = "UnknownLock"
)

// AnalysisResult holds the result of the SQL static analysis.
type AnalysisResult struct {
	TableName string
	LockLevel LockLevel
	Operation string // 작업 유형: CREATE, ALTER, DROP 등
	Columns   []string
}

// ParseSQL analyzes the provided SQL string and returns a slice of AnalysisResult.
func ParseSQL(sql string) ([]AnalysisResult, error) {
	result, err := pg_query.Parse(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL: %v", err)
	}

	var results []AnalysisResult
	for _, stmt := range result.Stmts {
		res := handleNode(stmt.Stmt)
		if res != nil {
			results = append(results, *res)
		}
	}

	return results, nil
}

// handleNode identifies the statement type and extracts basic information.
// 이 함수는 1.3(타겟 추출) 및 1.4(락 매핑) 단계에서 구체화될 예정입니다.
func handleNode(node *pg_query.Node) *AnalysisResult {
	// PostgreSQL AST의 루트 노드에서 실제 구문 타입을 확인합니다.
	if node.GetAlterTableStmt() != nil {
		return &AnalysisResult{Operation: "ALTER", LockLevel: UnknownLock}
	}
	if node.GetCreateStmt() != nil {
		return &AnalysisResult{Operation: "CREATE", LockLevel: UnknownLock}
	}
	if node.GetDropStmt() != nil {
		return &AnalysisResult{Operation: "DROP", LockLevel: UnknownLock}
	}
	if node.GetIndexStmt() != nil {
		return &AnalysisResult{Operation: "CREATE INDEX", LockLevel: UnknownLock}
	}

	return nil
}
