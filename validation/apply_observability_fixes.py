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

print("OBSERVABILITY_KNOWN_FAILURE_FIXES=PASS")
