package analyzer

// EvaluateLevel은 위험 점수(Risk Score)를 기준으로 최종 리스크 등급을 결정합니다.
func EvaluateLevel(score float64, permanentFailure bool) string {
	if permanentFailure || score >= 80.0 {
		return "Danger"
	}
	if score >= 50.0 {
		return "Warning"
	}
	return "Safe"
}
