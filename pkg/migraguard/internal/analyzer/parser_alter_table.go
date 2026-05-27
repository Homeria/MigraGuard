package analyzer

import (
	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
	"github.com/pganalyze/pg_query_go/v5"
)

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
