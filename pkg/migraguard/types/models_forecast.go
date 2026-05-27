package types

// ForecastTimeSlot represents a predicted risk state for a specific hour.
type ForecastTimeSlot struct {
	Hour         int     `json:"hour"`
	ExpectedTPS  float64 `json:"expected_tps"`
	MinTPS       float64 `json:"min_tps"`
	MaxTPS       float64 `json:"max_tps"`
	ExpectedP99  float64 `json:"expected_p99"`
	RiskScore    float64 `json:"risk_score"`
	RiskLevel    string  `json:"risk_level"`
	IsSafeWindow bool    `json:"is_safe_window"`
}

// ForecastReport bundles 24-hour predictive results for a specific DDL.
type ForecastReport struct {
	TableName string              `json:"table_name"`
	Timeline  []ForecastTimeSlot  `json:"timeline"`
	BestHour  int                 `json:"best_hour"`
}
