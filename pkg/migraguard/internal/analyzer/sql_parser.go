package analyzer

import (
	"fmt"
	"strings"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
	"github.com/pganalyze/pg_query_go/v5"
)

// ParseSQL parses the input SQL string into AST and extracts DDL information.
func ParseSQL(sqlText string) ([]types.AnalysisResult, error) {
	result, err := pg_query.Parse(sqlText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL: %w", err)
	}

	var analyses []types.AnalysisResult

	for _, stmt := range result.Stmts {
		node := stmt.Stmt
		var res types.AnalysisResult
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

		deparsed, err := pg_query.Deparse(&pg_query.ParseResult{Stmts: []*pg_query.RawStmt{stmt}})
		if err == nil {
			res.RawQuery = deparsed
		} else {
			res.RawQuery = extractRawQuery(sqlText, stmt)
		}

		analyses = append(analyses, res)
	}

	return analyses, nil
}

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

func analyzeAlterTable(stmt *pg_query.AlterTableStmt) types.AnalysisResult {
	tableName := NormalizeTableName(stmt.Relation.Relname)
	res := types.AnalysisResult{
		TableName:    tableName,
		Operation:    "ALTER TABLE",
		LockLevel:    types.LockLevelAccessExclusive,
		MetadataOnly: true,
	}

	for _, cmd := range stmt.Cmds {
		subCmd := cmd.GetAlterTableCmd()
		if subCmd == nil {
			continue
		}

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
			res.LockLevel = types.LockLevelShareUpdateExcl
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

func analyzeCreateIndex(stmt *pg_query.IndexStmt) types.AnalysisResult {
	res := types.AnalysisResult{
		TableName:    NormalizeTableName(stmt.Relation.Relname),
		IsIndex:      true,
		Operation:    "CREATE INDEX",
		LockLevel:    types.LockLevelShare,
		MetadataOnly: false,
	}
	if stmt.Concurrent {
		res.LockLevel = types.LockLevelShareUpdateExcl
		res.SubOperation = "CREATE INDEX CONCURRENTLY"
	}
	return res
}

func analyzeDrop(stmt *pg_query.DropStmt) types.AnalysisResult {
	res := types.AnalysisResult{
		Operation:    "DROP",
		LockLevel:    types.LockLevelAccessExclusive,
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

func analyzeTruncate(stmt *pg_query.TruncateStmt) types.AnalysisResult {
	res := types.AnalysisResult{
		Operation:    "TRUNCATE",
		LockLevel:    types.LockLevelAccessExclusive,
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

func analyzeRename(stmt *pg_query.RenameStmt) types.AnalysisResult {
	res := types.AnalysisResult{
		TableName:    NormalizeTableName(stmt.Relation.Relname),
		Operation:    "RENAME",
		SubOperation: "RENAME OBJECT",
		LockLevel:    types.LockLevelAccessExclusive,
		MetadataOnly: true,
	}
	return res
}

func NormalizeTableName(name string) string {
	name = strings.ReplaceAll(name, "\"", "")
	return strings.ToLower(strings.TrimSpace(name))
}
