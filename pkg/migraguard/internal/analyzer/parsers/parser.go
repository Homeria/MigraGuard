package parsers

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
			res = AnalyzeAlterTable(node.GetAlterTableStmt())
			found = true
		} else if node.GetIndexStmt() != nil {
			res = AnalyzeCreateIndex(node.GetIndexStmt())
			found = true
		} else if node.GetDropStmt() != nil {
			res = AnalyzeDrop(node.GetDropStmt())
			found = true
		} else if node.GetTruncateStmt() != nil {
			res = AnalyzeTruncate(node.GetTruncateStmt())
			found = true
		} else if node.GetRenameStmt() != nil {
			res = AnalyzeRename(node.GetRenameStmt())
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

func NormalizeTableName(name string) string {
	name = strings.ReplaceAll(name, "\"", "")
	return strings.ToLower(strings.TrimSpace(name))
}
