package weather

import (
	"context"
	"go/parser"
	"go/token"
	"strconv"
	"testing"

	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
)

const openMeteoIntegrationImport = "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/openmeteo"

type domainCurrentWeatherClient struct{}

func (domainCurrentWeatherClient) GetCurrentWeather(
	context.Context,
	domainweather.CurrentWeatherRequest,
) (domainweather.CurrentSnapshot, error) {
	return domainweather.CurrentSnapshot{}, nil
}

var _ CurrentWeatherClient = domainCurrentWeatherClient{}

func TestWeatherServiceDoesNotImportOpenMeteo(
	t *testing.T,
) {
	assertGoFileDoesNotImport(
		t,
		"service.go",
		openMeteoIntegrationImport,
	)
}

func assertGoFileDoesNotImport(
	t *testing.T,
	path string,
	forbidden string,
) {
	t.Helper()

	parsed, err := parser.ParseFile(
		token.NewFileSet(),
		path,
		nil,
		parser.ImportsOnly,
	)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	for _, imported := range parsed.Imports {
		importPath, err := strconv.Unquote(
			imported.Path.Value,
		)
		if err != nil {
			t.Fatalf(
				"unquote import %s in %s: %v",
				imported.Path.Value,
				path,
				err,
			)
		}
		if importPath == forbidden {
			t.Fatalf(
				"%s must not import provider integration %s",
				path,
				forbidden,
			)
		}
	}
}
