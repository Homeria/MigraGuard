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
}

// NewSimulateService initializes the simulation service.
func NewSimulateService(verbose bool) *SimulateService {
	return &SimulateService{Verbose: verbose}
}

// Run executes the simulation: loading scenario, creating sandbox, and seeding data.
func (s *SimulateService) Run(ctx context.Context, scenarioPath string) (string, error) {
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
	if s.Verbose {
		fmt.Printf("[DEBUG] Creating sandbox database: %s\n", dbPath)
	}

	// 3. Initialize Fresh Sandbox Database
	// If exists, delete to ensure clean experiment (Portability & Reproducibility)
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

	fmt.Printf("[OK] Simulation Sandbox created: %s (%s)\n", dbPath, scenario.Description)
	return dbPath, nil
}
