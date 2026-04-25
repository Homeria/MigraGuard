package analyzer

// EvaluateLevel determines the final risk level based on the risk score.
func EvaluateLevel(score float64, permanentFailure bool) string {
	if permanentFailure || score >= 80.0 {
		return "Danger"
	}
	if score >= 50.0 {
		return "Warning"
	}
	return "Safe"
}
