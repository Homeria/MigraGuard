package parser

import (
	"testing"
)

func TestParseSQL(t *testing.T) {
	tests := []struct {
		name      string
		sql       string
		expectedOp string
	}{
		{
			name:       "Alter Table",
			sql:        "ALTER TABLE users ADD COLUMN age INT;",
			expectedOp: "ALTER",
		},
		{
			name:       "Create Table",
			sql:        "CREATE TABLE orders (id SERIAL PRIMARY KEY);",
			expectedOp: "CREATE",
		},
		{
			name:       "Drop Table",
			sql:        "DROP TABLE customers;",
			expectedOp: "DROP",
		},
		{
			name:       "Create Index",
			sql:        "CREATE INDEX idx_user_email ON users(email);",
			expectedOp: "CREATE INDEX",
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

			if results[0].Operation != tt.expectedOp {
				t.Errorf("Expected operation %s, got %s", tt.expectedOp, results[0].Operation)
			}
		})
	}
}
