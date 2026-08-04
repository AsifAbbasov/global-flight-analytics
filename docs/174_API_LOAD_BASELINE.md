# Document 174 — API Load Baseline

Status: IMPLEMENTED
Project: Global Flight Analytics
Scope: reproducible containerized HTTP performance evidence for the public read path

## 1. Purpose

This increment adds a bounded Grafana k6 workload that measures the Go API and PostgreSQL
inside one isolated Docker network on a GitHub-hosted Ubuntu runner. It provides repeatable
latency, failure-rate, throughput and dropped-iteration evidence for review and regression
control.

The result is not a Render capacity claim. Public free-tier cold starts, shared platform
contention, network distance and production data volume are separate deployment concerns.

## 2. Workload

The baseline performs ten warm-up iterations and then maintains eight complete API iterations
per second for thirty seconds. Each iteration sends six requests, which is forty-eight HTTP
requests per second during the steady interval.

The request mix is:

```text
GET /api/v1/health
GET /api/v1/ready
GET /api/v1/version
GET /api/v1/regions
GET /api/v1/airports?limit=20
GET /api/v1/traffic/current?limit=100
```

The test runs against a freshly migrated PostgreSQL 16 database and the exact backend image
built for the Continuous Integration source SHA.

## 3. Initial Objectives

```text
minimum requests:             1400
failed request rate:          below 0.5%
check success rate:           above 99.9%
overall p95 latency:          below 750 milliseconds
overall p99 latency:          below 1500 milliseconds
lifecycle p95 latency:        below 300 milliseconds
database-read p95 latency:    below 750 milliseconds
dropped iterations:           zero
```

These objectives are conservative regression guardrails. They are not a statement of maximum
throughput or production autoscaling capacity.

## 4. Evidence

Each successful workflow run uploads:

```text
artifacts/performance/k6-summary.json
artifacts/performance/api-load-baseline.json
artifacts/performance/api-load-baseline.md
```

`api-load-baseline.json` records the exact source SHA, pinned container images, workload profile,
request count, request rate, checks, failures, overall percentiles, endpoint-class percentiles,
dropped iterations and every threshold decision.

Required markers are:

```text
API_LOAD_BASELINE_TARGET=PASS
API_LOAD_BASELINE_K6=PASS
API_LOAD_BASELINE_SUMMARY=PASS
API_LOAD_BASELINE_EVIDENCE=PASS
API_LOAD_BASELINE=PASS
```

## 5. Reproduction

Build the backend image and run:

```bash
docker build \
  --build-arg APP_VERSION=performance-local \
  --build-arg RENDER_GIT_COMMIT="$(git rev-parse HEAD)" \
  --build-arg BUILD_DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --file apps/api/Dockerfile \
  --tag global-flight-analytics-api:performance \
  .

pnpm run run:api-load-baseline
```

The runner creates and removes its own Docker network and containers. It does not publish the
API container to a host port and does not contact the public Render deployment.

## 6. Interpretation Boundary

A passing run proves that the tested source revision met the stated objectives in the recorded
single-runner container environment. Trend claims require multiple comparable runs. Production
capacity claims require a separately approved test against production-like infrastructure and
data volume.
