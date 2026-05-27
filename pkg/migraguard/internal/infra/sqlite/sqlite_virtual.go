package sqlite

import (
	"context"

	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// VirtualPGAdapter implements types.PostgresClient by reading from an SQLite sandbox.
type VirtualPGAdapter struct {
	sqlite types.SQLiteClient
}

// NewVirtualPGAdapter creates a new virtual adapter.
func NewVirtualPGAdapter(sqlite types.SQLiteClient) *VirtualPGAdapter {
	return &VirtualPGAdapter{sqlite: sqlite}
}

// FetchCurrentWorkloadSnapshot retrieves the latest simulated workload from SQLite.
func (a *VirtualPGAdapter) FetchCurrentWorkloadSnapshot(ctx context.Context) ([]types.WorkloadSnapshot, error) {
	adapter, ok := a.sqlite.(*SQLiteAdapter)
	if !ok {
		// Fallback safe snapshot mock if non-SQLiteAdapter (e.g. Mock client in testing) is injected
		return []types.WorkloadSnapshot{
			{
				QueryID:   9999,
				Query:     "SELECT 1",
				Calls:     100,
				TotalTime: 10.0,
			},
		}, nil
	}

	// For simulation, we return the latest unique queries (up to 100)
	query := `
		SELECT query_id, query, MAX(calls), MAX(total_time), 0, 0, 0 
		FROM workload_snapshots 
		GROUP BY query_id 
		ORDER BY timestamp DESC 
		LIMIT 100
	`
	rows, err := adapter.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []types.WorkloadSnapshot
	for rows.Next() {
		var s types.WorkloadSnapshot
		if err := rows.Scan(&s.QueryID, &s.Query, &s.Calls, &s.TotalTime, &s.Rows, &s.SharedBlksHit, &s.SharedBlksRead); err != nil {
			continue
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, nil
}

// FetchTableDynamicMetrics retrieves the latest simulated table state from SQLite.
func (a *VirtualPGAdapter) FetchTableDynamicMetrics(ctx context.Context, tableName string) (*types.TableDynamicMetrics, error) {
	m, err := a.sqlite.GetLatestTableMetrics(ctx, tableName)
	if err != nil {
		return nil, err
	}
	if m == nil {
		// If not found in metrics, return a default simulated state to prevent analysis failure
		return &types.TableDynamicMetrics{
			TableName:         tableName,
			TableSize:         1024 * 1024 * 100, // 100MB
			ActiveConnections: 10,
			P99Time:           5.0,
			TPS:               100.0,
		}, nil
	}
	return m, nil
}

// CheckTableSchemaPresence always returns nil in virtual mode as we assume simulation tables exist.
func (a *VirtualPGAdapter) CheckTableSchemaPresence(ctx context.Context, tableName string, columns []string) error {
	return nil
}

// Close is a no-op as the SQLite connection is managed by the client.
func (a *VirtualPGAdapter) Close() {}
