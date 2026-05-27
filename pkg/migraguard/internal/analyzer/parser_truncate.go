package analyzer

import (
	"strings"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
	"github.com/pganalyze/pg_query_go/v5"
)

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
