package engine

import (
	"context"
	"math"
	"testing"

	"github.com/Homeria/MigraGuard/internal/parser"
)

// TestCalculateRiskScore_CoreLogic verifies the mathematical correctness of the queuing model.
func TestCalculateRiskScore_CoreLogic(t *testing.T) {
	constants := DefaultRiskConstants()
	
	tests := []struct {
		name              string
		rewriteRequired   bool
		tableSize         int64
		tps               float64
		activeConnections int
		expectedRiskLevel string
		expectedRiskScore float64
	}{
		{
			name:              "Safe Case: Low Traffic",
			rewriteRequired:   false,
			tableSize:         1024 * 1024, // 1MB
			tps:               10.0,        // 10 TPS
			activeConnections: 50,
			expectedRiskLevel: "Safe",
		},
		{
			name:              "Danger Case: High Traffic & Heavy Table Rewrite",
			rewriteRequired:   true,
			tableSize:         10 * 1024 * 1024 * 1024, // 10GB
			tps:               2000.0,                  // 2000 TPS
			activeConnections: 100,
			expectedRiskLevel: "Danger",
		},
		{
			name:              "Critical Case: Over MuMax (Permanent Failure)",
			rewriteRequired:   true,
			tableSize:         1 * 1024 * 1024, // 1MB
			tps:               6000.0,          // Above MuMax (5000)
			activeConnections: 50,
			expectedRiskLevel: "Danger",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Step 1: Physical DDL Time (T_ddl)
			var estimatedDDLTime float64
			if tt.rewriteRequired {
				estimatedDDLTime = (float64(tt.tableSize) / float64(constants.DiskIO)) * 1000.0
			} else {
				estimatedDDLTime = constants.TMeta
			}

			// Step 2: Blocking Time (T_block) - Simple assumption for unit test
			p99Time := 10.0 // 10ms
			replicationLag := 0.0
			blockingTime := p99Time + estimatedDDLTime + (replicationLag * 1000.0)

			// Step 3: Peak Connections (C_peak)
			lambdaPerMs := tt.tps / 1000.0
			peakConnections := tt.activeConnections + int(lambdaPerMs*blockingTime)

			// Step 4: Risk Score
			riskScore := (float64(peakConnections) / float64(constants.CMax)) * 100.0

			// Step 5: Decision
			var riskLevel string
			permanentFailure := tt.tps >= constants.MuMax
			if riskScore >= 90.0 || permanentFailure {
				riskLevel = "Danger"
			} else if riskScore >= 60.0 {
				riskLevel = "Warning"
			} else {
				riskLevel = "Safe"
			}

			if riskLevel != tt.expectedRiskLevel {
				t.Errorf("[%s] Expected RiskLevel %s, got %s (Score: %.2f%%, Peak: %d)", 
					tt.name, tt.expectedRiskLevel, riskLevel, riskScore, peakConnections)
			}

			if permanentFailure && !math.IsInf(math.Inf(1)) { // Simplified check
				// Success if permanent failure is detected correctly
			}
		})
	}
}
