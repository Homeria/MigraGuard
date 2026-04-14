package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config는 어플리케이션 전반에서 사용하는 통합 설정 구조체입니다.
type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Agent    AgentConfig    `mapstructure:"agent"`
	Risk     RiskConfig     `mapstructure:"risk"`
}

// DatabaseConfig는 각 데이터베이스의 접속 경로를 정의합니다.
type DatabaseConfig struct {
	Postgres string `mapstructure:"postgres"`
	SQLite   string `mapstructure:"sqlite"`
}

// AgentConfig는 지표 수집 에이전트의 동작 파라미터를 정의합니다.
type AgentConfig struct {
	Interval      string `mapstructure:"interval"`
	RetentionDays int    `mapstructure:"retention_days"`
}

// RiskConfig는 리스크 엔진 분석에 사용되는 인프라 성능 상수를 정의합니다.
type RiskConfig struct {
	DiskIO   int64   `mapstructure:"disk_io"`
	MuMax    float64 `mapstructure:"mu_max"`
	CMax     int     `mapstructure:"c_max"`
	TTimeout float64 `mapstructure:"t_timeout"`
	TMeta    float64 `mapstructure:"t_meta"`
}

// LoadConfig는 설정 파일이나 환경 변수로부터 시스템 설정을 읽어와 구조체로 매핑합니다.
func LoadConfig(path string) (*Config, error) {
	if path != "" {
		viper.SetConfigFile(path)
	} else {
		// 기본 설정 파일명 검색 (.yaml, .json 등 지원)
		viper.SetConfigName("migraguard")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/migraguard/")
	}

	// 환경 변수 연동 (예: MIGRAGUARD_DATABASE_POSTGRES)
	viper.SetEnvPrefix("MIGRAGUARD")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// 기본값 설정 로직 호출
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("설정 파일 로드 실패: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("설정 데이터 매핑 실패: %w", err)
	}

	return &config, nil
}

// setDefaults는 설정 파일이 없을 경우 적용될 시스템 기본값을 정의합니다.
func setDefaults() {
	viper.SetDefault("database.sqlite", "./migraguard.db")
	viper.SetDefault("agent.interval", "1m")
	viper.SetDefault("agent.retention_days", 7)
	viper.SetDefault("risk.disk_io", 104857600) // 100MB/s
	viper.SetDefault("risk.mu_max", 5000.0)
	viper.SetDefault("risk.c_max", 500)
	viper.SetDefault("risk.t_timeout", 5000.0)
	viper.SetDefault("risk.t_meta", 100.0)
}
