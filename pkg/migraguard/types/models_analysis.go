package types

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
	Results         []AnalysisResult
	Reports         []*RiskAnalysisReport
	ForecastReports []*ForecastReport
}
