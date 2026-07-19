package server

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWeatherRouteFileRemainsCoordinatorOnly(
	t *testing.T,
) {
	content := readWeatherServerSource(
		t,
		"weather_route.go",
	)

	for _, required := range []string{
		"composeWeatherRouteDependencies(",
		"registerCurrentWeatherRoute(",
	} {
		if !strings.Contains(
			content,
			required,
		) {
			t.Fatalf(
				"weather route coordinator is missing %q",
				required,
			)
		}
	}

	for _, forbidden := range []string{
		"openmeteo",
		"providerbudget",
		"providerresponse",
		"weatherprovider",
		"repository/postgres",
		"services/weather",
		"NewWeatherRepository",
		"NewWeatherHandler",
		"v1.Get(",
	} {
		if strings.Contains(
			content,
			forbidden,
		) {
			t.Fatalf(
				"weather route coordinator owns forbidden responsibility %q",
				forbidden,
			)
		}
	}
}

func TestWeatherCompositionResponsibilitiesRemainSeparated(
	t *testing.T,
) {
	provider := readWeatherServerSource(
		t,
		"weather_provider_composition.go",
	)
	application := readWeatherServerSource(
		t,
		"weather_application_composition.go",
	)
	registration := readWeatherServerSource(
		t,
		"weather_route_registration.go",
	)

	for _, required := range []string{
		"providerbudget.New(",
		"providerresponse.New(",
		"NewIntegrationObserver(",
		"NewDefault[",
		"openmeteo.New(",
		"weatherprovider.New(",
	} {
		if !strings.Contains(
			provider,
			required,
		) {
			t.Fatalf(
				"weather provider composition is missing %q",
				required,
			)
		}
	}
	for _, forbidden := range []string{
		"NewWeatherRepository",
		"NewWeatherHandler",
		"v1.Get(",
	} {
		if strings.Contains(
			provider,
			forbidden,
		) {
			t.Fatalf(
				"weather provider composition owns forbidden responsibility %q",
				forbidden,
			)
		}
	}

	for _, required := range []string{
		"postgres.NewWeatherRepository(",
		"weatherservice.New(",
		"handlers.NewWeatherHandler(",
	} {
		if !strings.Contains(
			application,
			required,
		) {
			t.Fatalf(
				"weather application composition is missing %q",
				required,
			)
		}
	}
	for _, forbidden := range []string{
		"providerbudget.New(",
		"providerresponse.New(",
		"openmeteo.New(",
		"weatherprovider.New(",
		"v1.Get(",
	} {
		if strings.Contains(
			application,
			forbidden,
		) {
			t.Fatalf(
				"weather application composition owns forbidden responsibility %q",
				forbidden,
			)
		}
	}

	for _, required := range []string{
		`CurrentWeatherPath = "/weather/current"`,
		"v1.Get(",
		"handler.GetCurrent",
	} {
		if !strings.Contains(
			registration,
			required,
		) {
			t.Fatalf(
				"weather route registration is missing %q",
				required,
			)
		}
	}
	for _, forbidden := range []string{
		"providerbudget.New(",
		"providerresponse.New(",
		"openmeteo.New(",
		"weatherprovider.New(",
		"NewWeatherRepository",
		"weatherservice.New(",
	} {
		if strings.Contains(
			registration,
			forbidden,
		) {
			t.Fatalf(
				"weather route registration owns forbidden responsibility %q",
				forbidden,
			)
		}
	}
}

func readWeatherServerSource(
	t *testing.T,
	name string,
) string {
	t.Helper()

	_, currentFile, _, ok :=
		runtime.Caller(0)
	if !ok {
		t.Fatal(
			"resolve weather server source path",
		)
	}

	content, err := os.ReadFile(
		filepath.Join(
			filepath.Dir(currentFile),
			name,
		),
	)
	if err != nil {
		t.Fatalf(
			"read %s: %v",
			name,
			err,
		)
	}

	return string(content)
}
