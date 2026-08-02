# Backend Observability and Service-Level Objective Closure

Status: repository-side observability foundation closed
Baseline: `76e8744513bc1ad7d185692ea9ad3222bc961e0c`
Date: 2026-08-02

## Scope

This increment adds a dependency-free Prometheus-compatible observability foundation without changing the modular-monolith architecture or introducing Redis, Kafka, Kubernetes, an external metrics database, or paid infrastructure.

The API exposes `GET /internal/metrics`. The endpoint is protected by a separately configured SHA-256 digest through `METRICS_KEY_SHA256` and the existing `X-Internal-API-Key` request header. When the key is not configured, the endpoint returns service unavailable instead of publishing metrics anonymously.

The long-running traffic ingestion command can expose the same protected format on `INGEST_METRICS_ADDRESS`. The address remains disabled when omitted. Setting an address without a metrics key is rejected during configuration loading.

## Metric families

The bounded metric families cover:

- API request count, in-flight requests, status class and duration histograms;
- external provider outcomes and latency histograms for Open-Meteo, airplanes.live and OpenSky;
- traffic ingestion cycle result, duration, consecutive failure count and next delay;
- PostgreSQL pool acquisition, idle, total and maximum connection state;
- persisted ingestion run lifecycle counts and freshness;
- persisted reconciliation task lifecycle counts and oldest pending age;
- collector health and collection failure count;
- build version and exact revision.

## Cardinality and confidentiality boundary

HTTP metrics use the Fiber route template, never the raw request path or query string. Provider, outcome, status and result values are normalized to finite allowlists.

The following values are deliberately prohibited as metric labels:

- request identifiers;
- client IP addresses;
- aircraft ICAO24 identifiers;
- trajectory or task identifiers;
- raw URLs or query strings;
- error messages and error types.

This prevents high-cardinality growth and avoids exposing operational or user-derived data through the metrics endpoint.

## Initial service-level objectives

These are initial portfolio-production objectives, not claims about historical performance. They become measurable only after a public deployment and an external Prometheus-compatible scraper exist.

| Signal | Initial objective | Evaluation window |
| --- | ---: | ---: |
| API availability excluding client errors | at least 99.5% | rolling 30 days |
| API p95 request latency | at most 2 seconds | rolling 15 minutes |
| API server-error ratio | below 1% | rolling 15 minutes |
| Latest finished ingestion age | below 120 seconds during scheduled operation | rolling 10 minutes |
| Consecutive ingestion failures | fewer than 3 | current state |
| PostgreSQL acquired-to-maximum pool ratio | below 80% | rolling 10 minutes |
| Oldest pending reconciliation task age | below 300 seconds | rolling 15 minutes |
| Observability collector scrape success | 1 | every scrape |

## Initial alert thresholds

The deployment runbook should configure alerts for:

1. API availability below 99.5% for 15 minutes;
2. p95 API latency above 2 seconds for 15 minutes;
3. server-error ratio at or above 1% for 10 minutes;
4. latest finished ingestion age above 120 seconds for 10 minutes;
5. three or more consecutive ingestion failures;
6. PostgreSQL pool utilization above 80% for 10 minutes;
7. oldest pending reconciliation task age above 300 seconds for 15 minutes;
8. any observability collector with a failed most-recent scrape.

Alerts are intentionally based on stable, low-cardinality metric families. No alert depends on a single aircraft, trajectory, request identifier or raw error message.

## Evidence boundary

Repository-side metrics, tests, authorization and Continuous Integration gates can be closed by source evidence. Real availability, latency, alert delivery and retention cannot be claimed until the API and ingestion worker are deployed and scraped by an external monitoring system.
