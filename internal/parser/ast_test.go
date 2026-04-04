package parser

import (
	"testing"
)

func TestParseSQL_FullVerification(t *testing.T) {
	tests := []struct {
		name              string
		sql               string
		expectedOp        string
		expectedTableName string
		expectedLock      LockLevel
		expectedColumns   []string
	}{
		{
			name:              "Alter Table Add Column",
			sql:               "ALTER TABLE users ADD COLUMN age INT;",
			expectedOp:        "ALTER",
			expectedTableName: "users",
			expectedLock:      AccessExclusiveLock,
			expectedColumns:   []string{"age"},
		},
		{
			name:              "Create Table with Columns",
			sql:               "CREATE TABLE orders (id SERIAL PRIMARY KEY, price NUMERIC);",
			expectedOp:        "CREATE",
			expectedTableName: "orders",
			expectedLock:      AccessExclusiveLock,
			expectedColumns:   []string{"id", "price"},
		},
		{
			name:              "Standard Index",
			sql:               "CREATE INDEX idx_user_email ON profile(email);",
			expectedOp:        "CREATE INDEX",
			expectedTableName: "profile",
			expectedLock:      ShareLock,
			expectedColumns:   nil,
		},
		{
			name:              "Concurrent Index",
			sql:               "CREATE INDEX CONCURRENTLY idx_user_email ON profile(email);",
			expectedOp:        "CREATE INDEX",
			expectedTableName: "profile",
			expectedLock:      ShareUpdateExclusiveLock,
			expectedColumns:   nil,
		},
		{
			name:              "Alter Table Add Column (Rewrite Required if default)",
			sql:               "ALTER TABLE users ADD COLUMN age INT DEFAULT 0;",
			expectedOp:        "ALTER",
			expectedTableName: "users",
			expectedLock:      AccessExclusiveLock,
			expectedColumns:   []string{"age"},
		},
		{
			name:              "Alter Column Type (Always Rewrite)",
			sql:               "ALTER TABLE products ALTER COLUMN price TYPE NUMERIC;",
			expectedOp:        "ALTER",
			expectedTableName: "products",
			expectedLock:      AccessExclusiveLock,
			expectedColumns:   []string{"price"},
		},
		{
			name:              "Set Not Null (Rewrite Required)",
			sql:               "ALTER TABLE posts ALTER COLUMN content SET NOT NULL;",
			expectedOp:        "ALTER",
			expectedTableName: "posts",
			expectedLock:      AccessExclusiveLock,
			expectedColumns:   []string{"content"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := ParseSQL(tt.sql)
			if err != nil {
				t.Fatalf("ParseSQL() error = %v", err)
			}

			if len(results) == 0 {
				t.Errorf("Expected 1 result, got 0")
				return
			}

			res := results[0]
			if res.Operation != tt.expectedOp {
				t.Errorf("Expected op %s, got %s", tt.expectedOp, res.Operation)
			}
			if res.TableName != tt.expectedTableName {
				t.Errorf("Expected table %s, got %s", tt.expectedTableName, res.TableName)
			}
			if res.LockLevel != tt.expectedLock {
				t.Errorf("Expected lock %s, got %s", tt.expectedLock, res.LockLevel)
			}

			// F_rewrite check (v3.0 Core)
			if tt.name == "Alter Column Type (Always Rewrite)" || tt.name == "Set Not Null (Rewrite Required)" {
				if !res.RewriteRequired {
					t.Errorf("[%s] Expected RewriteRequired = true, got false", tt.name)
				}
			}
			if len(tt.expectedColumns) > 0 {
				if len(res.Columns) != len(tt.expectedColumns) {
					t.Errorf("Expected %d columns, got %d", len(tt.expectedColumns), len(res.Columns))
				}
			}
		})
	}
}

func TestParseSQL_MultipleStatements(t *testing.T) {
	sql := "ALTER TABLE t1 ADD COLUMN c1 INT; CREATE TABLE t2 (c2 TEXT);"
	results, err := ParseSQL(sql)
	if err != nil {
		t.Fatalf("ParseSQL() error = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}
