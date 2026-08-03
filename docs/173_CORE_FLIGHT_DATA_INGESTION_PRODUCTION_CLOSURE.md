# Document 173 — Core Flight Data Ingestion Production Closure

Status: Implementation prepared for installation and runtime activation
Project: Global Flight Analytics
Scope: free production scheduling, bounded ingestion execution, freshness evidence, and honest runtime boundaries

---

## 1. Audit Finding

The source implementation already contained a mature ingestion pipeline:

```text
provider policy and budget
provider health and fallback
canonical FlightState mapping
normalization and validation
exact duplicate removal
track construction
PostgreSQL persistence
durable ingestion runs
partial and failed terminal status
stale-run recovery
reconciliation for derived-write failures
```

The production audit on 2026-08-03 found a deployment gap instead of a missing domain implementation.

```text
PRODUCTION_API_REVISION=5c1c0862581842a78c323f5581c1425641b2b363
PRODUCTION_API_SOURCE_REVISION_MATCH=YES
PRODUCTION_TRAFFIC_SAMPLE_1 count=3
PRODUCTION_TRAFFIC_SAMPLE_2 count=3
PRODUCTION_TRAFFIC_ADVANCEMENT=NOT_OBSERVED
RENDER_INGEST_SERVICE_DECLARED=NO
STAGE_5_PRODUCTION_INGESTION_RUNTIME=NOT_DECLARED
```

The newest production observation was from 2026-07-14, approximately nineteen and one-half days before the audit. The API was healthy and revision-correct, but the database was not receiving new traffic observations.

---

## 2. Root Cause

The production container includes `/app/ingest`, but the Render Blueprint starts only `/app/server`.

A second continuously running Render service would require a non-free service type. The project permanently excludes paid production infrastructure from this portfolio closure, so the fix does not add a Render background worker or paid cron job.

---

## 3. Free Production Runtime

The production ingestion runtime is:

```text
GitHub Actions schedule
↓
serialized production ingestion job
↓
go run ./cmd/ingest --once
↓
free external traffic provider
↓
canonical processing pipeline
↓
Neon PostgreSQL
↓
Render read-only API
↓
public freshness verification
```

The scheduled workflow runs every ten minutes and may also be dispatched manually.

Only one production ingestion run may execute at a time. An active database write is never cancelled by a newer scheduled event.

---

## 4. Bounded Command Contract

The existing `ingest` command keeps daemon mode as its default behavior.

```text
go run ./cmd/ingest
```

The production scheduler uses explicit one-shot mode:

```text
go run ./cmd/ingest --once
```

One-shot mode:

```text
loads the same production configuration
uses the same provider policy, budget and health controls
recovers stale durable runs
executes exactly one ingestion cycle
records the same cycle and provider evidence
returns a non-zero exit code on cycle failure
schedules no retry delay and starts no long-running metrics listener
exits after the cycle reaches a terminal result
```

It does not create an alternate ingestion implementation.

---

## 5. Secret Boundary

The scheduled workflow requires one GitHub Actions repository secret:

```text
PRODUCTION_INGESTION_DATABASE_URL
```

The value is the owner-controlled Neon PostgreSQL connection string used for bounded production writes. It must never be committed, printed, stored in artifacts, or passed through pull-request events.

The workflow has read-only repository permissions. It runs only from `schedule` and `workflow_dispatch`; it is not exposed to untrusted pull-request code.

---

## 6. Freshness Gate

After one-shot ingestion completes, the workflow queries the public production endpoint:

```text
GET /api/v1/traffic/current
```

The gate requires:

```text
an HTTP success response
a valid application response envelope
at least one aircraft record
a valid observed_at timestamp on every returned item
the newest observation to be no older than thirty minutes
future timestamp skew to remain bounded
```

Stable success evidence:

```text
PRODUCTION_TRAFFIC_FRESHNESS=PASS
```

A successful ingestion process without fresh public data is not accepted as production closure.

---

## 7. Schedule Limitations

GitHub scheduled workflows are best-effort infrastructure. They may start later than the nominal cron time, and inactive public repositories may have scheduled workflows disabled by the platform.

Therefore the product must not claim:

```text
continuous ten-minute execution
guaranteed real-time coverage
guaranteed provider availability
operational surveillance continuity
safety-critical freshness
```

The thirty-minute freshness threshold is an explicit portfolio-health boundary, not an operational aviation service-level agreement.

---

## 8. Verification

Source verification:

```bash
pnpm run test:production-ingestion-contract
pnpm run verify:production-ingestion-contract
```

Focused Go verification:

```bash
cd apps/api
go test ./cmd/ingest
```

Runtime activation sequence:

```text
install and validate the source patch
commit and push through protected main
configure PRODUCTION_INGESTION_DATABASE_URL
manually dispatch Production Traffic Ingestion
verify the workflow conclusion is success
verify PRODUCTION_TRAFFIC_FRESHNESS=PASS
verify a second scheduled run advances observed_at
```

---

## 9. Completion Boundary

This increment closes the missing free production execution path only after runtime activation evidence passes.

It does not claim:

```text
paid worker availability
continuous provider access
global coverage
satellite coverage
commercial aviation data
official flight status
operational tracking
certified freshness
```

Core Flight Data Ingestion is fully closed only when source validation, protected-main CI, workflow activation, fresh PostgreSQL evidence, and advancing public traffic timestamps are all verified against the intended commit.
