package projectionevaluation

import (
	"fmt"
	"strings"
)

func evaluationGroupKey(result Result) string {
	return strings.Join([]string{
		strings.TrimSpace(result.ProjectionMethod.Name),
		strings.TrimSpace(result.ProjectionMethod.Version),
		string(result.ProjectionMethod.DecisionClass),
		fmt.Sprintf("%d", result.ProjectionHorizonEndTime.Sub(result.ProjectionAsOfTime)),
		fmt.Sprintf("%d", result.ForecastStep),
		result.Policy.Version,
		result.Policy.InputFingerprint,
	}, "\x00")
}

func methodSummaryKey(method MethodSummary) string {
	return strings.Join([]string{
		strings.TrimSpace(method.MethodName),
		strings.TrimSpace(method.MethodVersion),
		string(method.DecisionClass),
		fmt.Sprintf("%d", method.ProjectionHorizonDuration),
		fmt.Sprintf("%d", method.ForecastStep),
		method.EvaluationPolicyVersion,
		method.EvaluationPolicyFingerprint,
	}, "\x00")
}
