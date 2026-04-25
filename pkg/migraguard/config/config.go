package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// System default constants
const (
	DefaultDiskIO   = 104857600 // 100MB/s
	DefaultMuMax    = 5000.0    // 5000 TPS
	DefaultCMax     = 500       // 500 Connections
	DefaultTTimeout = 5000.0    // 5000 ms
	DefaultTMeta    = 100.0     // 100 ms
)

// Config represents the unified configuration structure.
type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Agent    AgentConfig    `mapstructure:"agent"`
	Risk     RiskConfig     `mapstructure:"risk"`
}

type DatabaseConfig struct {
	Postgres string `mapstructure:"postgres"`
	SQLite   string `mapstructure:"sqlite"`
}

type AgentConfig struct {
	Interval      string `mapstructure:"interval"`
	RetentionDays int    `mapstructure:"retention_days"`
}

type RiskConfig struct {
	DiskIO   int64   `mapstructure:"disk_io"`
	MuMax    float64 `mapstructure:"mu_max"`
	CMax     int     `mapstructure:"c_max"`
	TTimeout float64 `mapstructure:"t_timeout"`
	TMeta    float64 `mapstructure:"t_meta"`

	Thresholds RiskThresholds `mapstructure:"thresholds"`
	Weights    RiskWeights    `mapstructure:"weights"`
}

type RiskThresholds struct {
	Danger  float64 `mapstructure:"danger"`
	Warning float64 `mapstructure:"warning"`
}

type RiskWeights struct {
	AvgMultiplier           float64 `mapstructure:"avg_multiplier"`
	PeakMultiplier          float64 `mapstructure:"peak_multiplier"`
	ConcurrentImpact        float64 `mapstructure:"concurrent_impact"`
	MiddleImpact            float64 `mapstructure:"middle_impact"`
	BaseAccessExclusiveMeta float64 `mapstructure:"base_access_exclusive_meta"`
	BaseAccessExclusiveFull float64 `mapstructure:"base_access_exclusive_full"`
	BaseExclusive           float64 `mapstructure:"base_exclusive"`
	BaseShare               float64 `mapstructure:"base_share"`
}

// ToRiskConstants maps the hierarchical configuration to the flat SDK structure.
func (rc *RiskConfig) ToRiskConstants() types.RiskConstants {
	return types.RiskConstants{
		DiskIO:                  rc.DiskIO,
		MuMax:                   rc.MuMax,
		CMax:                    rc.CMax,
		TTimeout:                rc.TTimeout,
		TMeta:                   rc.TMeta,
		ThresholdDanger:         rc.Thresholds.Danger,
		ThresholdWarning:        rc.Thresholds.Warning,
		AvgMultiplier:           rc.Weights.AvgMultiplier,
		PeakMultiplier:          rc.Weights.PeakMultiplier,
		ConcurrentImpact:        rc.Weights.ConcurrentImpact,
		MiddleImpact:            rc.Weights.MiddleImpact,
		BaseAccessExclusiveMeta: rc.Weights.BaseAccessExclusiveMeta,
		BaseAccessExclusiveFull: rc.Weights.BaseAccessExclusiveFull,
		BaseExclusive:           rc.Weights.BaseExclusive,
		BaseShare:               rc.Weights.BaseShare,
	}
}

// LoadConfig loads configuration and ensures environment variables are mapped correctly.
func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("migraguard")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("/etc/migraguard/")
	}

	v.SetEnvPrefix("MIGRAGUARD")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 1. Set internal defaults to help Unmarshal recognize keys
	setInternalDefaults(v)

	// 2. Try to read config file (ignore error if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Real error (e.g. permission or syntax)
			return nil, err
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 3. Manual override for nested structs if Unmarshal missed them
	// This is a known limitation of Viper when config file is missing
	if config.Database.Postgres == "" {
		config.Database.Postgres = v.GetString("database.postgres")
	}
	if config.Database.SQLite == "" {
		config.Database.SQLite = v.GetString("database.sqlite")
	}

	return &config, nil
}

func setInternalDefaults(v *viper.Viper) {
	v.SetDefault("database.postgres", "")
	v.SetDefault("database.sqlite", "./migraguard.db")
	v.SetDefault("agent.interval", "1m")
	v.SetDefault("agent.retention_days", 7)
	v.SetDefault("risk.disk_io", DefaultDiskIO)
	v.SetDefault("risk.mu_max", DefaultMuMax)
	v.SetDefault("risk.c_max", DefaultCMax)
	v.SetDefault("risk.t_timeout", DefaultTTimeout)
	v.SetDefault("risk.t_meta", DefaultTMeta)
	v.SetDefault("risk.thresholds.danger", 80.0)
	v.SetDefault("risk.thresholds.warning", 50.0)
	v.SetDefault("risk.weights.avg_multiplier", 1.2)
	v.SetDefault("risk.weights.peak_multiplier", 0.8)
	v.SetDefault("risk.weights.concurrent_impact", 0.1)
	v.SetDefault("risk.weights.middle_impact", 0.5)
	v.SetDefault("risk.weights.base_access_exclusive_meta", 30.0)
	v.SetDefault("risk.weights.base_access_exclusive_full", 85.0)
	v.SetDefault("risk.weights.base_exclusive", 50.0)
	v.SetDefault("risk.weights.base_share", 20.0)
}
