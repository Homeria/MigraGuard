package parser

import (
	"testing"
)

func TestParseSQL_LockMapping(t *testing.T) {
	tests := []struct {
		name             string
		sql              string
		expectedOp       string
		expectedLock     LockLevel
	}{
		{
			name:         "Alter Table Lock",
			sql:          "ALTER TABLE users ADD COLUMN age INT;",
			expectedOp:   "ALTER",
			expectedLock: AccessExclusiveLock,
		},
		{
			name:         "Standard Index Lock",
			sql:          "CREATE INDEX idx_user_email ON profile(email);",
			expectedOp:   "CREATE INDEX",
			expectedLock: ShareLock,
		},
		{
			name:         "Concurrent Index Lock",
			sql:          "CREATE INDEX CONCURRENTLY idx_user_email ON profile(email);",
			expectedOp:   "CREATE INDEX",
			expectedLock: ShareUpdateExclusiveLock,
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

			if results[0].LockLevel != tt.expectedLock {
				t.Errorf("Expected lock %s, got %s", tt.expectedLock, results[0].LockLevel)
			}
		})
	}
}
