package types

import "time"

// WorkloadSnapshot represents a point-in-time snapshot of query statistics.
type WorkloadSnapshot struct {
	Timestamp      time.Time
	QueryID        int64
	Query          string
	Calls          int64
	TotalTime      float64
	Rows           int64
	SharedBlksHit  int64
	SharedBlksRead int64
}

// TableDynamicMetrics represents the real-time state of a table.
type TableDynamicMetrics struct {
	Timestamp         time.Time
	TableName         string
	TableSize         int64
	ReplicationLag    float64
	ActiveConnections int
	P99Time           float64
	TPS               float64
	SharedBlksHit     int64
	SharedBlksRead    int64
}

// BaselineStats represents historical workload patterns.
type BaselineStats struct {
	AvgTPS_1h   float64
	PeakTPS_24h float64
}

// TopQueryInfo represents a high-impact query.
type TopQueryInfo struct {
	QueryID   int64
	QueryText string
	Calls     int64
	TotalTime float64
	Impact    float64
}
