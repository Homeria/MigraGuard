package app

import (
	"context"
	"fmt"
	"os"

	"github.com/Homeria/MigraGuard/pkg/migraguard/internal/infra/sqlite"
	"github.com/Homeria/MigraGuard/pkg/migraguard/types"
	"gopkg.in/yaml.v3"
)

// SimulateService manages the declarative simulation workflow.
type SimulateService struct {
	Verbose bool
	logger  types.Logger
}

// NewSimulateService initializes the simulation service.
func NewSimulateService(verbose bool, logger types.Logger) *SimulateService {
	return &SimulateService{
		Verbose: verbose,
		logger:  logger,
	}
}

// Run executes the simulation: loading scenario, creating sandbox, and seeding data.
func (s *SimulateService) Run(ctx context.Context, scenarioPath string, force bool) (string, error) {
	// 1. Load Scenario YAML
	content, err := os.ReadFile(scenarioPath)
	if err != nil {
		return "", fmt.Errorf("failed to read scenario file: %w", err)
	}

	var scenario types.SimulationScenario
	if err := yaml.Unmarshal(content, &scenario); err != nil {
		return "", fmt.Errorf("failed to parse scenario YAML: %w", err)
	}

	// 2. Determine Sandbox DB Path
	dbPath := fmt.Sprintf("%s.db", scenario.ExperimentName)

	// 3. Conditional Seeding (Persistence Logic)
	if _, err := os.Stat(dbPath); err == nil && !force {
		if s.Verbose {
			s.logger.Info("[INFO] Sandbox database already exists: %s. Skipping seeding (use --force to overwrite).", dbPath)
		}
		return dbPath, nil
	}

	if s.Verbose {
		s.logger.Debug("[DEBUG] Initializing fresh sandbox database: %s", dbPath)
	}

	// If force or not exists, ensure clean start
	os.Remove(dbPath)

	adapter, err := sqlite.NewRepository(dbPath)
	if err != nil {
		return "", fmt.Errorf("failed to initialize sandbox db: %w", err)
	}
	defer adapter.Close()

	// 4. Seed Data via Sandbox Engine
	engine := sqlite.NewSandboxEngine(adapter)
	if err := engine.SeedScenario(scenario); err != nil {
		return "", fmt.Errorf("failed to seed scenario data: %w", err)
	}

	s.logger.Info("[OK] Simulation Sandbox seeded: %s (%s)", dbPath, scenario.Description)
	return dbPath, nil
}
