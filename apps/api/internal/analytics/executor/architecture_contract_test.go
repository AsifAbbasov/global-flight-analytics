package executor

import (
	"reflect"
	"strings"
	"testing"
)

func TestExecutorDoesNotExposeInternalDependencies(
	t *testing.T,
) {
	executorType := reflect.TypeOf(
		&Executor{},
	)

	for _, methodName := range []string{
		"Calculator",
		"ScopeGuard",
		"ConfidenceEvaluator",
	} {
		if _, exists := executorType.MethodByName(
			methodName,
		); exists {
			t.Fatalf(
				"executor must not expose %s",
				methodName,
			)
		}
	}
}

func TestExecutorDoesNotRetainLegacyCalculator(
	t *testing.T,
) {
	executorType := reflect.TypeOf(
		Executor{},
	)

	for index := 0; index < executorType.NumField(); index++ {
		field := executorType.Field(index)
		if strings.Contains(
			field.Type.String(),
			"calculator.Calculator",
		) {
			t.Fatalf(
				"executor must not retain legacy calculator field %s",
				field.Name,
			)
		}
	}
}
