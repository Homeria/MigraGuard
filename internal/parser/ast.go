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
// SQL 정적 분석 결과를 보유하는 구조체입니다.
type AnalysisResult struct {
	TableName       string
	LockLevel       LockLevel
	Operation       string   // 작업 유형: CREATE, ALTER, DROP 등
	Columns         []string // 영향받는 컬럼 목록
	RewriteRequired bool     // F_rewrite: 테이블 재기록(Rewrite) 발생 여부
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
		// ALTER TABLE 명령들을 순회하며 컬럼 추출 및 Rewrite 여부 판별
		for _, cmd := range stmt.Cmds {
			if sub := cmd.GetAlterTableCmd(); sub != nil {
				if sub.Name != "" {
					res.Columns = append(res.Columns, sub.Name)
				}

				// PostgreSQL에서 테이블 재기록(F_rewrite)을 유발하는 주요 작업들
				switch sub.Subtype {
				case pg_query.AlterTableType_AT_AlterColumnType: // ALTER TYPE
					res.RewriteRequired = true
				case pg_query.AlterTableType_AT_SetNotNull: // SET NOT NULL (v12 미만은 재기록 발생)
					res.RewriteRequired = true
				case pg_query.AlterTableType_AT_AddColumn: // ADD COLUMN (DEFAULT가 있거나 VOLATILE일 때 발생할 수 있음)
					if sub.Def != nil {
						res.RewriteRequired = true
					}
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
			RewriteRequired: false, // 신규 테이블 생성은 기존 데이터 재기록이 없음
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
