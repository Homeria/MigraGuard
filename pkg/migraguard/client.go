package migraguard

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Homeria/MigraGuard/pkg/migraguard/internal/analyzer"
	"github.com/Homeria/MigraGuard/pkg/migraguard/internal/app"
	"github.com/Homeria/MigraGuard/pkg/migraguard/internal/infra/postgres"
	"github.com/Homeria/MigraGuard/pkg/migraguard/internal/infra/sqlite"
	migraErrors "github.com/Homeria/MigraGuard/pkg/migraguard/internal/shared/errors"
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
	logger types.Logger
}

// NewLiveClient creates a MigraGuard client for live database monitoring and analysis.
func NewLiveClient(cfg Config, opts ...Option) (*Client, error) {
	defaults := analyzer.DefaultRiskConstants()
	fillMissingRiskParams(&cfg.Risk, &defaults)

	for _, opt := range opts {
		opt(&cfg)
	}

	var pgAdapter types.PostgresClient
	var err error
	if cfg.PostgresDSN != "" {
		pgAdapter, err = postgres.NewAdapter(cfg.PostgresDSN)
		if err != nil {
			return nil, migraErrors.Wrap(err, "Client.NewLiveClient", "postgres connection failed")
		}
	}

	sqliteRepo, err := sqlite.NewRepository(cfg.SQLitePath)
	if err != nil {
		if pgAdapter != nil {
			pgAdapter.Close()
		}
		return nil, migraErrors.Wrap(err, "Client.NewLiveClient", "sqlite initialization failed")
	}

	return &Client{
		config: &cfg,
		pg:     pgAdapter,
		sqlite: sqliteRepo,
		logger: &defaultLogger{}, // Default no-op logger
	}, nil
}

// NewSandboxClient creates a MigraGuard client for offline simulation and research.
func NewSandboxClient(sandboxPath string, cfg Config, opts ...Option) (*Client, error) {
	defaults := analyzer.DefaultRiskConstants()
	fillMissingRiskParams(&cfg.Risk, &defaults)

	for _, opt := range opts {
		opt(&cfg)
	}

	sandboxRepo, err := sqlite.NewRepository(sandboxPath)
	if err != nil {
		return nil, migraErrors.Wrap(err, "Client.NewSandboxClient", "sandbox loading failed")
	}

	return &Client{
		config: &cfg,
		pg:     sqlite.NewVirtualPGAdapter(sandboxRepo),
		sqlite: sandboxRepo,
		logger: &defaultLogger{},
	}, nil
}

// WithLogger sets a custom logger for the client.
func (c *Client) WithLogger(l types.Logger) *Client {
	c.logger = l
	return c
}

// defaultLogger is a no-op implementation of types.Logger.
type defaultLogger struct{}

func (l *defaultLogger) Debug(msg string, args ...interface{}) {}
func (l *defaultLogger) Info(msg string, args ...interface{})  {}
func (l *defaultLogger) Warn(msg string, args ...interface{})  {}
func (l *defaultLogger) Error(msg string, args ...interface{}) {}

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

// --- Legacy Support ---

// New is kept for backward compatibility but calls NewLiveClient internally.
func New(cfg Config, opts ...Option) (*Client, error) {
	return NewLiveClient(cfg, opts...)
}

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

// --- Methods ---

// Close releases all resources.
func (c *Client) Close() error {
	if c.pg != nil {
		c.pg.Close()
	}
	if c.sqlite != nil {
		c.sqlite.Close()
	}
	return nil
}

// StartAgent starts background collection.
func (c *Client) StartAgent(ctx context.Context, targetTables string) error {
	if c.pg == nil {
		return migraErrors.New(migraErrors.ErrCodeDBConn, "Client.StartAgent", "agent requires PostgreSQL connection")
	}
	agent := app.NewAgentService(c.pg, c.sqlite, c.config.Interval, c.config.RetentionDays)
	agent.SetTargetTables(targetTables)
	return agent.Run(ctx)
}

// Analyze performs the analysis.
func (c *Client) Analyze(ctx context.Context, sqlPath string) (*types.AnalysisResponse, error) {
	if c.pg == nil {
		return nil, migraErrors.New(migraErrors.ErrCodeDBConn, "Client.Analyze", "analysis requires connection")
	}
	analyzeService := app.NewAnalyzeService(c.pg, c.sqlite, c.config.Risk, c.config.Verbose)
	return analyzeService.Run(ctx, app.AnalysisTask{SQLPath: sqlPath})
}

// Simulate creates a sandbox environment and seeds data based on a scenario.
func (c *Client) Simulate(ctx context.Context, scenarioPath string, force bool) (string, error) {
	simulateService := app.NewSimulateService(c.config.Verbose)
	return simulateService.Run(ctx, scenarioPath, force)
}

// ExportSandboxMetrics exports all metrics from the current sandbox/sqlite to a CSV file.
func (c *Client) ExportSandboxMetrics(ctx context.Context, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return migraErrors.Wrap(err, "Client.ExportSandboxMetrics", "failed to create file")
	}
	defer f.Close()

	return c.ExportSandboxMetricsToWriter(ctx, f)
}

// ExportSandboxMetricsToWriter exports all metrics from the current sandbox/sqlite to an io.Writer.
func (c *Client) ExportSandboxMetricsToWriter(ctx context.Context, w io.Writer) error {
	if c.sqlite == nil {
		return migraErrors.New(migraErrors.ErrCodeDBConn, "Client.ExportSandboxMetricsToWriter", "no sqlite connected")
	}

	metrics, err := c.sqlite.FetchAllTableMetrics(ctx)
	if err != nil {
		return migraErrors.Wrap(err, "Client.ExportSandboxMetricsToWriter", "fetch failed")
	}

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Header
	writer.Write([]string{"Timestamp", "TableName", "TableSize", "ReplicationLag", "ActiveConnections", "P99Time", "TPS", "SharedBlksHit", "SharedBlksRead"})

	for _, m := range metrics {
		row := []string{
			m.Timestamp.Format("2006-01-02 15:04:05"),
			m.TableName,
			fmt.Sprintf("%d", m.TableSize),
			fmt.Sprintf("%.2f", m.ReplicationLag),
			fmt.Sprintf("%d", m.ActiveConnections),
			fmt.Sprintf("%.2f", m.P99Time),
			fmt.Sprintf("%.2f", m.TPS),
			fmt.Sprintf("%d", m.SharedBlksHit),
			fmt.Sprintf("%d", m.SharedBlksRead),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return writer.Error()
}
