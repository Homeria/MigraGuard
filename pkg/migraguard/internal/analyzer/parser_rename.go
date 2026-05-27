package analyzer

import (
	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
	"github.com/pganalyze/pg_query_go/v5"
)

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
