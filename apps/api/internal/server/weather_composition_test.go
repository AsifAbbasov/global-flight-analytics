package server

import (
	"strings"
	"testing"
	"time"
)

func TestComposeWeatherRouteDependenciesRejectsInvalidTimeout(
	t *testing.T,
) {
	_, err := composeWeatherRouteDependencies(
		weatherCompositionConfig{},
	)
	if err == nil {
		t.Fatal(
			"composeWeatherRouteDependencies() error = nil",
		)
	}
	if !strings.Contains(
		err.Error(),
		"initialize open-meteo client",
	) ||
		!strings.Contains(
			err.Error(),
			"open-meteo timeout must be greater than zero",
		) {
		t.Fatalf(
			"unexpected invalid timeout error: %v",
			err,
		)
	}
}

func TestComposeWeatherRouteDependenciesBuildsCompleteGraph(
	t *testing.T,
) {
	dependencies, err :=
		composeWeatherRouteDependencies(
			weatherCompositionConfig{
				openMeteoTimeout: time.Second,
			},
		)
	if err != nil {
		t.Fatalf(
			"composeWeatherRouteDependencies() error = %v",
			err,
		)
	}

	if dependencies.client == nil {
		t.Fatal(
			"weather client was not composed",
		)
	}
	if dependencies.repository == nil {
		t.Fatal(
			"weather repository was not composed",
		)
	}
	if dependencies.service == nil {
		t.Fatal(
			"weather service was not composed",
		)
	}
	if dependencies.handler == nil {
		t.Fatal(
			"weather handler was not composed",
		)
	}
}
