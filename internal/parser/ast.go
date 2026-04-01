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

// handleNode identifies the statement type and extracts table names.
func handleNode(node *pg_query.Node) *AnalysisResult {
	if stmt := node.GetAlterTableStmt(); stmt != nil {
		return &AnalysisResult{
			Operation: "ALTER",
			TableName: stmt.Relation.Relname,
			LockLevel: UnknownLock,
		}
	}

	if stmt := node.GetCreateStmt(); stmt != nil {
		return &AnalysisResult{
			Operation: "CREATE",
			TableName: stmt.Relation.Relname,
			LockLevel: UnknownLock,
		}
	}

	if stmt := node.GetDropStmt(); stmt != nil {
		// DROP 구문은 여러 객체를 가질 수 있으나, 첫 번째 대상을 주 타겟으로 잡습니다.
		tableName := "unknown"
		if len(stmt.Objects) > 0 {
			// DROP TABLE의 경우 객체 리스트의 첫 번째 요소에서 이름을 추출합니다.
			// 실제로는 List 구조를 더 파싱해야 할 수도 있으나 기초 구현을 우선합니다.
			tableName = "multiple or complex drop" 
		}
		return &AnalysisResult{
			Operation: "DROP",
			TableName: tableName,
			LockLevel: UnknownLock,
		}
	}

	if stmt := node.GetIndexStmt(); stmt != nil {
		return &AnalysisResult{
			Operation: "CREATE INDEX",
			TableName: stmt.Relation.Relname,
			LockLevel: UnknownLock,
		}
	}

	return nil
}
