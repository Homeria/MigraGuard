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

// DatabaseConfig defines database connection paths.
type DatabaseConfig struct {
	Postgres string `mapstructure:"postgres"`
	SQLite   string `mapstructure:"sqlite"`
}

// AgentConfig defines parameters for the metric collection agent.
type AgentConfig struct {
	Interval      string `mapstructure:"interval"`
	RetentionDays int    `mapstructure:"retention_days"`
}

// RiskConfig defines infrastructure performance variables for risk analysis.
type RiskConfig struct {
	DiskIO   int64   `mapstructure:"disk_io"`
	MuMax    float64 `mapstructure:"mu_max"`
	CMax     int     `mapstructure:"c_max"`
	TTimeout float64 `mapstructure:"t_timeout"`
	TMeta    float64 `mapstructure:"t_meta"`
}

// LoadConfig loads configuration from a file or environment variables.
func LoadConfig(path string) (*Config, error) {
	if path != "" {
		viper.SetConfigFile(path)
	} else {
		viper.SetConfigName("migraguard")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/migraguard/")
	}

	viper.SetEnvPrefix("MIGRAGUARD")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults defines system default values when config is missing.
func setDefaults() {
	viper.SetDefault("database.sqlite", "./migraguard.db")
	viper.SetDefault("agent.interval", "1m")
	viper.SetDefault("agent.retention_days", 7)
	viper.SetDefault("risk.disk_io", DefaultDiskIO)
	viper.SetDefault("risk.mu_max", DefaultMuMax)
	viper.SetDefault("risk.c_max", DefaultCMax)
	viper.SetDefault("risk.t_timeout", DefaultTTimeout)
	viper.SetDefault("risk.t_meta", DefaultTMeta)
}
