package weatherprovider

import (
	"context"
	"go/parser"
	"go/token"
	"strconv"
	"testing"

	domainweather "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/domain/weather"
)

const forbiddenOpenMeteoImport = "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/integrations/openmeteo"

type domainWeatherDelegate struct{}

func (domainWeatherDelegate) GetCurrentWeather(
	context.Context,
	domainweather.CurrentWeatherRequest,
) (domainweather.CurrentSnapshot, error) {
	return domainweather.CurrentSnapshot{}, nil
}

var _ Delegate = domainWeatherDelegate{}

func TestWeatherProviderOrchestrationDoesNotImportOpenMeteo(
	t *testing.T,
) {
	for _, path := range []string{
		"client.go",
		"request_key.go",
	} {
		assertWeatherProviderFileDoesNotImport(
			t,
			path,
			forbiddenOpenMeteoImport,
		)
	}
}

func TestCurrentWeatherRequestKeyAcceptsDomainRequest(
	t *testing.T,
) {
	key := CurrentWeatherRequestKey(
		domainweather.CurrentWeatherRequest{
			Latitude:  40.4093,
			Longitude: 49.8671,
		},
	)
	if key != "current:40.4093:49.8671" {
		t.Fatalf("request key = %q", key)
	}
}

func assertWeatherProviderFileDoesNotImport(
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
