#!/usr/bin/env python3
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def replace_once(relative_path: str, old: str, new: str) -> None:
    path = ROOT / relative_path
    content = path.read_text()
    count = content.count(old)
    if count != 1:
        raise SystemExit(
            f"OBSERVABILITY_FIX_ERROR={relative_path} replacement count {count}, expected 1"
        )
    path.write_text(content.replace(old, new, 1))


replace_once(
    "apps/api/internal/observability/metrics_server.go",
    '\tif ctx == nil {\n\t\tctx = context.Background()\n\t}\n\treturn server.server.Shutdown(ctx)\n',
    '\tif ctx == nil {\n\t\treturn fmt.Errorf("metrics server shutdown context is required")\n\t}\n\treturn server.server.Shutdown(ctx)\n',
)
replace_once(
    "apps/api/internal/observability/postgres_collector.go",
    '\tif ctx == nil {\n\t\tctx = context.Background()\n\t}\n\n\tcollectionContext, cancel := context.WithTimeout(ctx, 2*time.Second)\n',
    '\tif ctx == nil {\n\t\treturn fmt.Errorf("postgres metrics context is required")\n\t}\n\n\tcollectionContext, cancel := context.WithTimeout(ctx, 2*time.Second)\n',
)
replace_once(
    "apps/api/internal/observability/registry.go",
    '\tif ctx == nil {\n\t\tctx = context.Background()\n\t}\n\n\tregistry.mu.RLock()\n',
    '\tif ctx == nil {\n\t\treturn ""\n\t}\n\n\tregistry.mu.RLock()\n',
)
replace_once(
    "apps/api/internal/observability/http_test.go",
    'import (\n\t"net/http/httptest"\n',
    'import (\n\t"context"\n\t"net/http/httptest"\n',
)
replace_once(
    "apps/api/internal/observability/http_test.go",
    "registry.Render(nil)",
    "registry.Render(context.Background())",
)
replace_once(
    "apps/api/internal/observability/provider_recorder_test.go",
    "registry.Render(nil)",
    "registry.Render(context.Background())",
)
replace_once(
    "apps/api/internal/server/server_test.go",
    '\t\t\t\terr := registerWeatherRoute(\n\t\t\t\t\tv1,\n\t\t\t\t\tnil,\n\t\t\t\t\ttest.timeout,\n\t\t\t\t)\n',
    '\t\t\t\terr := registerWeatherRoute(\n\t\t\t\t\tv1,\n\t\t\t\t\tnil,\n\t\t\t\t\ttest.timeout,\n\t\t\t\t\tnil,\n\t\t\t\t)\n',
)
replace_once(
    "apps/api/internal/server/server_test.go",
    '\terr := registerWeatherRoute(\n\t\tv1,\n\t\tnil,\n\t\t5*time.Second,\n\t)\n',
    '\terr := registerWeatherRoute(\n\t\tv1,\n\t\tnil,\n\t\t5*time.Second,\n\t\tnil,\n\t)\n',
)

replace_once(
    "apps/api/internal/server/database_routes.go",
    '\t"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/observability"\n',
    '',
)
replace_once(
    "apps/api/internal/server/database_routes.go",
    '\tmetricsRegistry *observability.Registry,\n',
    '\tmetricsRecorder weatherProviderRecorder,\n',
)
replace_once(
    "apps/api/internal/server/database_routes.go",
    '\t\t\tmetricsRegistry,\n\t\t\tmutationAuthorization,\n',
    '\t\t\tmetricsRecorder,\n\t\t\tmutationAuthorization,\n',
)
replace_once(
    "apps/api/internal/server/database_routes.go",
    '\tmetricsRegistry *observability.Registry,\n',
    '\tmetricsRecorder weatherProviderRecorder,\n',
)
replace_once(
    "apps/api/internal/server/database_routes.go",
    '\t\t\t\t\tmetricsRegistry,\n',
    '\t\t\t\t\tmetricsRecorder,\n',
)

replace_once(
    "apps/api/internal/server/weather_route.go",
    '\t"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/observability"\n',
    '',
)
replace_once(
    "apps/api/internal/server/weather_route.go",
    '\tmetricsRegistry *observability.Registry,\n',
    '\tmetricsRecorder weatherProviderRecorder,\n',
)
replace_once(
    "apps/api/internal/server/weather_route.go",
    '\t\t\t\tmetricsRegistry:  metricsRegistry,\n',
    '\t\t\t\tmetricsRecorder:  metricsRecorder,\n',
)

replace_once(
    "apps/api/internal/server/weather_composition.go",
    '\t"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/observability"\n',
    '',
)
replace_once(
    "apps/api/internal/server/weather_composition.go",
    '\tmetricsRegistry  *observability.Registry\n',
    '\tmetricsRecorder  weatherProviderRecorder\n',
)
replace_once(
    "apps/api/internal/server/weather_composition.go",
    '''\tclient, err := composeWeatherProvider(
\t\tconfig.openMeteoTimeout,
\t\tobservability.NewProviderRecorder(
\t\t\tconfig.metricsRegistry,
\t\t\tnil,
\t\t),
\t)
''',
    '''\tclient, err := composeWeatherProvider(
\t\tconfig.openMeteoTimeout,
\t\tconfig.metricsRecorder,
\t)
''',
)

replace_once(
    "apps/api/internal/server/weather_provider_composition.go",
    '\t"github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/observability"\n',
    '',
)
replace_once(
    "apps/api/internal/server/weather_provider_composition.go",
    '''\tweatherservice "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/weather"
)

func composeWeatherProvider(
''',
    '''\tweatherservice "github.com/AsifAbbasov/global-flight-analytics/apps/api/internal/services/weather"
)

type weatherProviderRecorder interface {
\tproviderresponse.ObservationRecorder
}

func composeWeatherProvider(
''',
)
replace_once(
    "apps/api/internal/server/weather_provider_composition.go",
    '\tmetricsRecorder *observability.ProviderRecorder,\n',
    '\tmetricsRecorder weatherProviderRecorder,\n',
)
replace_once(
    "apps/api/internal/server/weather_provider_composition.go",
    '\trecorder *observability.ProviderRecorder,\n',
    '\trecorder weatherProviderRecorder,\n',
)
replace_once(
    "apps/api/internal/server/weather_provider_composition.go",
    '''\tobserver, err :=
\t\tproviderresponse.
\t\t\tNewIntegrationObserverWithRecorder(
\t\t\t\tcontroller,
\t\t\t\trecorder,
\t\t\t)
''',
    '''\tobserver, err :=
\t\tproviderresponse.
\t\t\tNewIntegrationObserver(
\t\t\t\tcontroller,
\t\t\t)
\tif recorder != nil {
\t\tobserver, err =
\t\t\tproviderresponse.
\t\t\t\tNewIntegrationObserverWithRecorder(
\t\t\t\t\tcontroller,
\t\t\t\t\trecorder,
\t\t\t\t)
\t}
''',
)

replace_once(
    "apps/api/internal/server/server.go",
    '''\t\t\tnormalizedConfig.ObservabilityRegistry,
\t\t\tmutationAuthorization,
''',
    '''\t\t\tobservability.NewProviderRecorder(
\t\t\t\tnormalizedConfig.ObservabilityRegistry,
\t\t\t\tnil,
\t\t\t),
\t\t\tmutationAuthorization,
''',
)

print("OBSERVABILITY_KNOWN_FAILURE_FIXES=PASS")
