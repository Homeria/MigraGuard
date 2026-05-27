package types

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
	AsymmetricSkew  float64         `yaml:"asymmetric_skew"` // Skewness factor for asymmetric load curve (e.g. -0.5 to 0.5)
	PeakShiftHours  float64         `yaml:"peak_shift_hours"` // Max random peak shift in hours
}

// TimelineEvent represents a specific time window with unusual traffic patterns.
type TimelineEvent struct {
	Name       string  `yaml:"name"`
	StartDay   int     `yaml:"start_day"`  // 0-indexed from start
	StartHour  int     `yaml:"start_hour"` // 0-23
	DurationH  int     `yaml:"duration_h"`
	Multiplier float64 `yaml:"multiplier"`
}
