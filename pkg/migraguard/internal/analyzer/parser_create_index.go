package analyzer

import (
	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
	"github.com/pganalyze/pg_query_go/v5"
)

func analyzeCreateIndex(stmt *pg_query.IndexStmt) types.AnalysisResult {
	res := types.AnalysisResult{
		TableName:    NormalizeTableName(stmt.Relation.Relname),
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
