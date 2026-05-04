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

// Option is a functional option for configuring the client.
type Option func(*Config)

// Client is the main entry point for all MigraGuard features.
type Client struct {
	config *Config
	pg     types.PostgresClient
	sqlite types.SQLiteClient
}

// New creates a new MigraGuard client with optional overrides.
func New(cfg Config, opts ...Option) (*Client, error) {
	// 1. Fill missing risk parameters with system defaults (Surgical approach)
	defaults := analyzer.DefaultRiskConstants()
	fillMissingRiskParams(&cfg.Risk, &defaults)

	// 2. Apply functional options (highest priority)
	for _, opt := range opts {
		opt(&cfg)
	}

	// 3. Initialize Adapters
	var pgAdapter *postgres.PostgresAdapter
	var err error
	if cfg.PostgresDSN != "" {
		pgAdapter, err = postgres.NewAdapter(cfg.PostgresDSN)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to postgres: %w", err)
		}
	}

	sqliteRepo, err := sqlite.NewRepository(cfg.SQLitePath)
	if err != nil {
		if pgAdapter != nil {
			pgAdapter.Close()
		}
		return nil, fmt.Errorf("failed to connect to sqlite: %w", err)
	}

	return &Client{
		config: &cfg,
		pg:     pgAdapter,
		sqlite: sqliteRepo,
	}, nil
}

// fillMissingRiskParams ensures that user settings are preserved while filling in gaps.
func fillMissingRiskParams(target *types.RiskConstants, def *types.RiskConstants) {
	if target.DiskIO == 0 { target.DiskIO = def.DiskIO }
	if target.TMeta == 0 { target.TMeta = def.TMeta }
	if target.MuMax == 0 { target.MuMax = def.MuMax }
	if target.CMax == 0 { target.CMax = def.CMax }
	if target.TTimeout == 0 { target.TTimeout = def.TTimeout }
	if target.ThresholdDanger == 0 { target.ThresholdDanger = def.ThresholdDanger }
	if target.ThresholdWarning == 0 { target.ThresholdWarning = def.ThresholdWarning }
	if target.AvgMultiplier == 0 { target.AvgMultiplier = def.AvgMultiplier }
	if target.PeakMultiplier == 0 { target.PeakMultiplier = def.PeakMultiplier }
	if target.ConcurrentImpact == 0 { target.ConcurrentImpact = def.ConcurrentImpact }
	if target.MiddleImpact == 0 { target.MiddleImpact = def.MiddleImpact }
	if target.BaseAccessExclusiveMeta == 0 { target.BaseAccessExclusiveMeta = def.BaseAccessExclusiveMeta }
	if target.BaseAccessExclusiveFull == 0 { target.BaseAccessExclusiveFull = def.BaseAccessExclusiveFull }
	if target.BaseExclusive == 0 { target.BaseExclusive = def.BaseExclusive }
	if target.BaseShare == 0 { target.BaseShare = def.BaseShare }
}

// --- Functional Options ---

func WithVerbose(v bool) Option {
	return func(c *Config) { c.Verbose = v }
}

func WithDangerThreshold(t float64) Option {
	return func(c *Config) { c.Risk.ThresholdDanger = t }
}

func WithWarningThreshold(t float64) Option {
	return func(c *Config) { c.Risk.ThresholdWarning = t }
}

// UseSandbox switches the client to use a VirtualPGAdapter backed by the provided SQLite path.
func (c *Client) UseSandbox(sandboxPath string) error {
	sandboxRepo, err := sqlite.NewRepository(sandboxPath)
	if err != nil {
		return fmt.Errorf("failed to load sandbox: %w", err)
	}
	
	// Replace PG adapter with Virtual adapter
	c.pg = sqlite.NewVirtualPGAdapter(sandboxRepo)
	
	// Close current sqlite and replace with sandbox sqlite
	if c.sqlite != nil {
		c.sqlite.Close()
	}
	c.sqlite = sandboxRepo
	return nil
}

// Close releases all resources.
func (c *Client) Close() error {
	if c.pg != nil { c.pg.Close() }
	if c.sqlite != nil { c.sqlite.Close() }
	return nil
}

// StartAgent starts background collection.
func (c *Client) StartAgent(ctx context.Context, targetTables string) error {
	if c.pg == nil {
		return fmt.Errorf("agent service requires a valid PostgreSQL connection")
	}
	agent := app.NewAgentService(c.pg, c.sqlite, c.config.Interval, c.config.RetentionDays)
	agent.SetTargetTables(targetTables)
	return agent.Run(ctx)
}

// Analyze performs the analysis.
func (c *Client) Analyze(ctx context.Context, sqlPath string) (*types.AnalysisResponse, error) {
	if c.pg == nil {
		return nil, fmt.Errorf("analysis requires a valid PostgreSQL connection or --sandbox mode")
	}
	analyzeService := app.NewAnalyzeService(c.pg, c.sqlite, c.config.Risk, c.config.Verbose)
	return analyzeService.Run(ctx, app.AnalysisTask{SQLPath: sqlPath})
}

// Simulate creates a sandbox environment and seeds data based on a scenario.
func (c *Client) Simulate(ctx context.Context, scenarioPath string, force bool) (string, error) {
	simulateService := app.NewSimulateService(c.config.Verbose)
	return simulateService.Run(ctx, scenarioPath, force)
}
