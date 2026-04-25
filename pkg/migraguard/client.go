package migraguard

import (
	"context"
	"fmt"
	"time"

	"github.com/Homeria/MigraGuard/pkg/migraguard/internal/analyzer"
	"github.com/Homeria/MigraGuard/pkg/migraguard/internal/app"
	"github.com/Homeria/MigraGuard/pkg/migraguard/internal/infra/postgres"
	"github.com/Homeria/MigraGuard/pkg/migraguard/internal/infra/sqlite"
	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
)

// Config is the unified configuration for the MigraGuard client.
type Config struct {
	PostgresDSN   string
	SQLitePath    string
	Interval      time.Duration
	RetentionDays int
	Verbose       bool
	Risk          types.RiskConstants
}

// Client is the main entry point for all MigraGuard features.
type Client struct {
	config *Config
	pg     *postgres.PostgresAdapter
	sqlite *sqlite.SQLiteAdapter
}

// New creates a new MigraGuard client and initializes database connections.
func New(cfg Config) (*Client, error) {
	if cfg.Risk.CMax == 0 {
		cfg.Risk = analyzer.DefaultRiskConstants()
	}

	pgAdapter, err := postgres.NewAdapter(cfg.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	sqliteRepo, err := sqlite.NewRepository(cfg.SQLitePath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to sqlite: %w", err)
	}

	return &Client{
		config: &cfg,
		pg:     pgAdapter,
		sqlite: sqliteRepo,
	}, nil
}

// Close releases all resources used by the client.
func (c *Client) Close() error {
	if c.pg != nil {
		c.pg.Close()
	}
	if c.sqlite != nil {
		c.sqlite.Close()
	}
	return nil
}

// StartAgent starts the background metric collection process.
func (c *Client) StartAgent(ctx context.Context, targetTables string) error {
	agent := app.NewAgentService(c.pg, c.sqlite, c.config.Interval, c.config.RetentionDays)
	agent.SetTargetTables(targetTables)
	return agent.Run(ctx)
}

// Analyze analyzes a migration SQL file and returns a risk report.
func (c *Client) Analyze(ctx context.Context, sqlPath string) (*types.AnalysisResponse, error) {
	analyzeService := app.NewAnalyzeService(c.pg, c.sqlite, c.config.Risk, c.config.Verbose)
	return analyzeService.Run(ctx, app.AnalysisTask{SQLPath: sqlPath})
}
