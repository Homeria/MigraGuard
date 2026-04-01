package parser

import (
	pg_query "github.com/pganalyze/pg_query_go/v5"
)

// LockLevel represents the PostgreSQL lock level.
type LockLevel string

const (
	AccessExclusiveLock LockLevel = "AccessExclusiveLock"
	ExclusiveLock       LockLevel = "ExclusiveLock"
	ShareLock           LockLevel = "ShareLock"
	RowExclusiveLock    LockLevel = "RowExclusiveLock"
	// TODO: 필요한 다른 락 레벨들을 추가할 예정
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
		return nil, err
	}

	// TODO: 1.2 AST 파싱 단계에서 상세 구현 예정
	_ = result

	return []AnalysisResult{}, nil
}
