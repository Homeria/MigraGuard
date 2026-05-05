package types

import "time"

// PostgreSQL Lock Levels (1 to 8)
const (
	LockLevelNone            = 0
	LockLevelAccessShare     = 1 // SELECT
	LockLevelRowShare        = 2 // SELECT FOR UPDATE
	LockLevelRowExclusive    = 3 // INSERT, UPDATE, DELETE
	LockLevelShareUpdateExcl = 4 // VACUUM, CREATE INDEX CONCURRENTLY
	LockLevelShare           = 5 // CREATE INDEX
	LockLevelShareRowExcl    = 6 // EXCLUSIVE
	LockLevelExclusive       = 7 // Block all but Access Share
	LockLevelAccessExclusive = 8 // ALTER TABLE, DROP, TRUNCATE (Full Block)
)

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

// AnalysisResult contains the results of a static SQL analysis.
type AnalysisResult struct {
	TableName       string
	IsIndex         bool
	Operation       string
	SubOperation    string
	Columns         []string
	RewriteRequired bool
	MetadataOnly    bool
	LockLevel       int
	RawQuery        string
}

// RiskConstants defines thresholds and weights for risk calculation.
type RiskConstants struct {
	// Performance
	DiskIO   int64   `mapstructure:"disk_io"`
	TMeta    float64 `mapstructure:"t_meta"`
	MuMax    float64 `mapstructure:"mu_max"`
	CMax     int     `mapstructure:"c_max"`
	TTimeout float64 `mapstructure:"t_timeout"`

	// Decision Thresholds
	ThresholdDanger  float64 `mapstructure:"threshold_danger"`
	ThresholdWarning float64 `mapstructure:"threshold_warning"`

	// Algorithm Weights
	AvgMultiplier    float64 `mapstructure:"avg_multiplier"`
	PeakMultiplier   float64 `mapstructure:"peak_multiplier"`
	ConcurrentImpact float64 `mapstructure:"concurrent_impact"`
	MiddleImpact     float64 `mapstructure:"middle_impact"`

	// Base Risk Values
	BaseAccessExclusiveMeta float64 `mapstructure:"base_access_exclusive_meta"`
	BaseAccessExclusiveFull float64 `mapstructure:"base_access_exclusive_full"`
	BaseExclusive           float64 `mapstructure:"base_exclusive"`
	BaseShare               float64 `mapstructure:"base_share"`
}

// RiskAnalysisReport bundles the risk model results.
type RiskAnalysisReport struct {
	RiskScore        float64
	RiskLevel        string
	EstimatedDDLTime float64
	BlockingTime     float64
	PeakConnections  int
	RecoveryTime     float64
	PermanentFailure bool
	BaseTPS          float64
	TPSSource        string
	CurrentTPS       float64
	AvgTPS1h         float64
	PeakTPS24h       float64
	ActiveConns      int
	TableSize        int64
	TopQueries       []TopQueryInfo
	SafeWindow       string
	SafeWindowTPS    float64
}

// AnalysisResponse bundles all results.
type AnalysisResponse struct {
	Results []AnalysisResult
	Reports []*RiskAnalysisReport
}

// SimulationScenario defines a research experiment setup via YAML.
type SimulationScenario struct {
	ExperimentName string               `yaml:"experiment_name"`
	Description    string               `yaml:"description"`
	TargetTable    string               `yaml:"target_table"`
	PGState        VirtualPGState       `yaml:"pg_state"`
	History        SQLiteHistoryProfile `yaml:"sqlite_history"`
}

// VirtualPGState represents the mocked state of a PostgreSQL instance.
type VirtualPGState struct {
	TableSizeMB       int64   `yaml:"table_size_mb"`
	ActiveConnections int     `yaml:"active_connections"`
	P99TimeMS         float64 `yaml:"p99_time_ms"`
	CurrentTPS        float64 `yaml:"current_tps"`
	ReplicationLagS   float64 `yaml:"replication_lag_s"`
}

// SQLiteHistoryProfile defines the rules for generating realistic 7-day time-series data.
type SQLiteHistoryProfile struct {
	Days            int             `yaml:"days"`
	IntervalMinutes int             `yaml:"interval_minutes"`
	BaseTPS         float64         `yaml:"base_tps"`
	PeakTPS         float64         `yaml:"peak_tps"`
	WeeklyPattern   bool            `yaml:"weekly_pattern"` // Lower traffic on weekends
	Events          []TimelineEvent `yaml:"events"`         // Specific marketing/spike events
	NoiseVariance   float64         `yaml:"noise_variance"`
}

// TimelineEvent represents a specific time window with unusual traffic patterns.
type TimelineEvent struct {
	Name       string  `yaml:"name"`
	StartDay   int     `yaml:"start_day"`  // 0-indexed from start
	StartHour  int     `yaml:"start_hour"` // 0-23
	DurationH  int     `yaml:"duration_h"`
	Multiplier float64 `yaml:"multiplier"`
}
