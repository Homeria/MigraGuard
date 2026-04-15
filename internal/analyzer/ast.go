package analyzer

import (
	"fmt"
	"strings"

	"github.com/pganalyze/pg_query_go/v5"
)

// PostgreSQL Lock Levels (1 to 8)
const (
	LockLevelNone            = 0
	LockLevelAccessShare     = 1 // SELECT
	LockLevelRowShare        = 2 // SELECT FOR UPDATE
	LockLevelRowExclusive    = 3 // INSERT, UPDATE, DELETE
	LockLevelShareUpdateExcl = 4 // VACUUM, CREATE INDEX CONCURRENTLY
	LockLevelShare           = 5 // CREATE INDEX
	LockLevelShareRowExcl    = 6 // EXCLUSIVE
	LockLevelExclusive       = 7 // Block all but Access Share
	LockLevelAccessExclusive = 8 // ALTER TABLE, DROP, TRUNCATE (Full Block)
)

// AnalysisResult는 SQL 정적 분석을 통해 도출된 결과 데이터를 담는 구조체입니다.
type AnalysisResult struct {
	TableName       string   // 대상 테이블명 (또는 인덱스명)
	IsIndex         bool     // 대상이 인덱스인지 여부
	Operation       string   // SQL 작업 유형 (ALTER, CREATE INDEX 등)
	SubOperation    string   // 세부 작업 (ADD COLUMN, CHANGE TYPE 등)
	Columns         []string // 관련 컬럼 목록
	RewriteRequired bool     // Table Rewrite 유발 여부
	MetadataOnly    bool     // 메타데이터만 변경하는 가벼운 작업 여부
	LockLevel       int      // PostgreSQL Lock 레벨 (1~8)
	RawQuery        string   // 원본 SQL 문장
}

// ParseSQL은 입력받은 SQL 문자열을 AST로 파싱하여 리스크 분석 정보를 추출합니다.
func ParseSQL(sqlText string) ([]AnalysisResult, error) {
	// 1. Postgres 파서 기동
	result, err := pg_query.Parse(sqlText)
	if err != nil {
		return nil, fmt.Errorf("SQL 구문 분석 실패: %w", err)
	}

	var analyses []AnalysisResult

	// 2. 문장별 순회 및 DDL 특징 추출
	for _, stmt := range result.Stmts {
		node := stmt.Stmt
		var res AnalysisResult
		found := false

		if node.GetAlterTableStmt() != nil {
			res = analyzeAlterTable(node.GetAlterTableStmt())
			found = true
		} else if node.GetIndexStmt() != nil {
			res = analyzeCreateIndex(node.GetIndexStmt())
			found = true
		} else if node.GetDropStmt() != nil {
			res = analyzeDrop(node.GetDropStmt())
			found = true
		} else if node.GetTruncateStmt() != nil {
			res = analyzeTruncate(node.GetTruncateStmt())
			found = true
		} else if node.GetRenameStmt() != nil {
			res = analyzeRename(node.GetRenameStmt())
			found = true
		}

		if !found {
			continue
		}

		// 3. 원문 텍스트 추출 (주석 제외, 순수 쿼리만 추출하기 위해 Deparse 활용)
		deparsed, err := pg_query.Deparse(&pg_query.ParseResult{Stmts: []*pg_query.RawStmt{stmt}})
		if err == nil {
			res.RawQuery = deparsed
		} else {
			// Deparse 실패 시 기존 위치 기반 추출 (백업)
			res.RawQuery = extractRawQuery(sqlText, stmt)
		}

		analyses = append(analyses, res)
	}

	return analyses, nil
}

// extractRawQuery는 전체 SQL 텍스트에서 특정 문장의 위치를 찾아 해당 부분만 반환합니다.
func extractRawQuery(fullText string, stmt *pg_query.RawStmt) string {
	start := stmt.StmtLocation
	if stmt.StmtLen <= 0 {
		return strings.TrimSpace(fullText[start:])
	}
	end := start + stmt.StmtLen
	if end > int32(len(fullText)) {
		end = int32(len(fullText))
	}
	return strings.TrimSpace(fullText[start:end])
}

func analyzeAlterTable(stmt *pg_query.AlterTableStmt) AnalysisResult {
	tableName := NormalizeTableName(stmt.Relation.Relname)
	res := AnalysisResult{
		TableName:    tableName,
		Operation:    "ALTER TABLE",
		LockLevel:    LockLevelAccessExclusive,
		MetadataOnly: true,
	}

	for _, cmd := range stmt.Cmds {
		subCmd := cmd.GetAlterTableCmd()
		if subCmd == nil { continue }

		switch subCmd.Subtype {
		case pg_query.AlterTableType_AT_AlterColumnType:
			res.RewriteRequired = true
			res.MetadataOnly = false
			res.SubOperation = "CHANGE COLUMN TYPE"
		case pg_query.AlterTableType_AT_AddConstraint:
			res.SubOperation = "ADD CONSTRAINT"
			if subCmd.Def != nil && subCmd.Def.GetConstraint() != nil {
				if subCmd.Def.GetConstraint().SkipValidation {
					res.RewriteRequired = false
					res.MetadataOnly = true
					res.SubOperation += " (NOT VALID)"
				} else {
					res.RewriteRequired = true
					res.MetadataOnly = false
				}
			}
		case pg_query.AlterTableType_AT_SetNotNull:
			res.RewriteRequired = true
			res.MetadataOnly = false
			res.SubOperation = "SET NOT NULL"
		case pg_query.AlterTableType_AT_ValidateConstraint:
			res.SubOperation = "VALIDATE CONSTRAINT"
			res.LockLevel = LockLevelShareUpdateExcl
			res.MetadataOnly = false
		case pg_query.AlterTableType_AT_AddColumn:
			res.SubOperation = "ADD COLUMN"
		}

		if subCmd.Name != "" {
			res.Columns = append(res.Columns, subCmd.Name)
		} else if subCmd.Def != nil {
			if colDef := subCmd.Def.GetColumnDef(); colDef != nil {
				res.Columns = append(res.Columns, colDef.Colname)
			}
		}
	}

	return res
}

func analyzeCreateIndex(stmt *pg_query.IndexStmt) AnalysisResult {
	res := AnalysisResult{
		TableName:    NormalizeTableName(stmt.Relation.Relname),
		Operation:    "CREATE INDEX",
		LockLevel:    LockLevelShare,
		MetadataOnly: false,
	}
	if stmt.Concurrent {
		res.LockLevel = LockLevelShareUpdateExcl
		res.SubOperation = "CREATE INDEX CONCURRENTLY"
	}
	return res
}

func analyzeDrop(stmt *pg_query.DropStmt) AnalysisResult {
	res := AnalysisResult{
		Operation:    "DROP",
		LockLevel:    LockLevelAccessExclusive,
		MetadataOnly: false,
	}
	if stmt.RemoveType == pg_query.ObjectType_OBJECT_INDEX {
		res.IsIndex = true
	}

	var tableNames []string
	for _, obj := range stmt.Objects {
		if list := obj.GetList(); list != nil && len(list.Items) > 0 {
			rawName := list.Items[len(list.Items)-1].GetString_().Sval
			tableNames = append(tableNames, NormalizeTableName(rawName))
		} else if rawName := obj.GetString_(); rawName != nil {
			tableNames = append(tableNames, NormalizeTableName(rawName.Sval))
		}
	}
	res.TableName = strings.Join(tableNames, ", ")
	return res
}

func analyzeTruncate(stmt *pg_query.TruncateStmt) AnalysisResult {
	res := AnalysisResult{
		Operation:    "TRUNCATE",
		LockLevel:    LockLevelAccessExclusive,
		MetadataOnly: false,
	}
	var tableNames []string
	for _, rel := range stmt.Relations {
		if rv := rel.GetRangeVar(); rv != nil {
			tableNames = append(tableNames, NormalizeTableName(rv.Relname))
		}
	}
	res.TableName = strings.Join(tableNames, ", ")
	return res
}

func analyzeRename(stmt *pg_query.RenameStmt) AnalysisResult {
	res := AnalysisResult{
		TableName:    NormalizeTableName(stmt.Relation.Relname),
		Operation:    "RENAME",
		SubOperation: "RENAME OBJECT",
		LockLevel:    LockLevelAccessExclusive,
		MetadataOnly: true,
	}
	return res
}

func NormalizeTableName(name string) string {
	name = strings.ReplaceAll(name, "\"", "")
	return strings.ToLower(strings.TrimSpace(name))
}
