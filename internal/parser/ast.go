package parser

import (
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

// LockLevel represents the PostgreSQL lock level.
type LockLevel string

const (
	AccessExclusiveLock       LockLevel = "Access Exclusive"
	ShareUpdateExclusiveLock  LockLevel = "Share Update Exclusive"
	ShareRowExclusiveLock     LockLevel = "Share Row Exclusive"
	ExclusiveLock             LockLevel = "Exclusive"
	ShareLock                 LockLevel = "Share"
	RowExclusiveLock          LockLevel = "Row Exclusive"
	RowShareLock              LockLevel = "Row Share"
	AccessShareLock           LockLevel = "Access Share"
	UnknownLock               LockLevel = "Unknown"
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

// handleNode identifies the statement type and determines the lock level.
func handleNode(node *pg_query.Node) *AnalysisResult {
	if stmt := node.GetAlterTableStmt(); stmt != nil {
		return &AnalysisResult{
			Operation: "ALTER",
			TableName: stmt.Relation.Relname,
			// ALTER TABLE은 기본적으로 Access Exclusive Lock을 필요로 합니다.
			// (일부 서브 명령은 낮을 수 있으나 안전을 위해 보수적으로 설정)
			LockLevel: AccessExclusiveLock,
		}
	}

	if stmt := node.GetCreateStmt(); stmt != nil {
		return &AnalysisResult{
			Operation: "CREATE",
			TableName: stmt.Relation.Relname,
			// 새 테이블 생성은 Access Exclusive Lock을 필요로 하지만 대상이 새 테이블이므로 영향도가 낮습니다.
			LockLevel: AccessExclusiveLock,
		}
	}

	if stmt := node.GetDropStmt(); stmt != nil {
		tableName := "unknown"
		if len(stmt.Objects) > 0 {
			tableName = "complex drop"
		}
		return &AnalysisResult{
			Operation: "DROP",
			TableName: tableName,
			LockLevel: AccessExclusiveLock,
		}
	}

	if stmt := node.GetIndexStmt(); stmt != nil {
		lockLevel := ShareLock // 기본 CREATE INDEX는 Share Lock
		if stmt.Concurrent {
			// CONCURRENTLY 옵션이 붙으면 Share Update Exclusive Lock으로 완화됩니다.
			lockLevel = ShareUpdateExclusiveLock
		}
		return &AnalysisResult{
			Operation: "CREATE INDEX",
			TableName: stmt.Relation.Relname,
			LockLevel: lockLevel,
		}
	}

	return nil
}
