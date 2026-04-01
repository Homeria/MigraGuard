package parser

import (
	"fmt"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

// LockLevel represents the PostgreSQL lock level.
type LockLevel string

const (
	AccessExclusiveLock      LockLevel = "Access Exclusive"
	ShareUpdateExclusiveLock LockLevel = "Share Update Exclusive"
	ShareRowExclusiveLock    LockLevel = "Share Row Exclusive"
	ExclusiveLock            LockLevel = "Exclusive"
	ShareLock                LockLevel = "Share"
	RowExclusiveLock         LockLevel = "Row Exclusive"
	RowShareLock             LockLevel = "Row Share"
	AccessShareLock          LockLevel = "Access Share"
	UnknownLock              LockLevel = "Unknown"
)

// AnalysisResult holds the result of the SQL static analysis.
type AnalysisResult struct {
	TableName string
	LockLevel LockLevel
	Operation string   // 작업 유형: CREATE, ALTER, DROP 등
	Columns   []string // 영향받는 컬럼 목록
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

// handleNode identifies the statement type and populates AnalysisResult.
func handleNode(node *pg_query.Node) *AnalysisResult {
	if stmt := node.GetAlterTableStmt(); stmt != nil {
		res := &AnalysisResult{
			Operation: "ALTER",
			TableName: stmt.Relation.Relname,
			LockLevel: AccessExclusiveLock,
		}
		// ALTER TABLE 명령들에서 컬럼명 추출 시도
		for _, cmd := range stmt.Cmds {
			if sub := cmd.GetAlterTableCmd(); sub != nil {
				if sub.Name != "" {
					res.Columns = append(res.Columns, sub.Name)
				}
			}
		}
		return res
	}

	if stmt := node.GetCreateStmt(); stmt != nil {
		res := &AnalysisResult{
			Operation: "CREATE",
			TableName: stmt.Relation.Relname,
			LockLevel: AccessExclusiveLock,
		}
		// CREATE TABLE의 컬럼 정의 추출
		for _, el := range stmt.TableElts {
			if def := el.GetColumnDef(); def != nil {
				res.Columns = append(res.Columns, def.Colname)
			}
		}
		return res
	}

	if stmt := node.GetDropStmt(); stmt != nil {
		tableName := "unknown"
		// DROP TABLE users -> Objects[0]에서 이름 추출 시도
		if len(stmt.Objects) > 0 {
			// 실제 AST 구조는 List 내의 List 형태로 복잡하므로 기초적인 추출만 수행
			tableName = "target object"
		}
		return &AnalysisResult{
			Operation: "DROP",
			TableName: tableName,
			LockLevel: AccessExclusiveLock,
		}
	}

	if stmt := node.GetIndexStmt(); stmt != nil {
		lockLevel := ShareLock
		if stmt.Concurrent {
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
