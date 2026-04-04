package config

import (
	"github.com/Homeria/MigraGuard/internal/engine"
)

// Config represents the overall configuration for MigraGuard.
type Config struct {
	Database DatabaseConfig       `mapstructure:"database"`
	Engine   engine.RiskConstants `mapstructure:"engine"`
	Agent    AgentConfig          `mapstructure:"agent"`
}

// DatabaseConfig holds DB connection defaults.
type DatabaseConfig struct {
	URL        string `mapstructure:"url"`
	SQLitePath string `mapstructure:"sqlite_path"`
}

// AgentConfig holds default settings for the background agent.
type AgentConfig struct {
	Interval  int `mapstructure:"interval"`
	Retention int `mapstructure:"retention"`
}

// DefaultConfig returns the default configuration values.
func DefaultConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			SQLitePath: "migraguard.db",
		},
		Engine: engine.DefaultRiskConstants(),
		Agent: AgentConfig{
			Interval:  60,
			Retention: 7,
		},
	}
}
