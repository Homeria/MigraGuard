package analyzer

// EvaluateLevel determines the final risk level based on the configurable risk thresholds.
func (e *RiskEngine) EvaluateLevel(score float64, permanentFailure bool) string {
	if permanentFailure || score >= e.constants.ThresholdDanger {
		return "Danger"
	}
	if score >= e.constants.ThresholdWarning {
		return "Warning"
	}
	return "Safe"
}
