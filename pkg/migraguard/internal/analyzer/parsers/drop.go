package parsers

import (
	"strings"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
	"github.com/pganalyze/pg_query_go/v5"
)

func AnalyzeDrop(stmt *pg_query.DropStmt) types.AnalysisResult {
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
