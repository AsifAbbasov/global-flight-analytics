# Recruiter Demo Script

## Goal

Show a coherent engineering product in seven minutes. Do not attempt to demonstrate every
module or scroll through the complete documentation history.

## Seven-minute walkthrough

### Minute 0–1: Product and boundary

Open the landing page. State:

> Global Flight Analytics is an open-data aviation research platform. It is not air
> traffic control or authoritative flight status. The product makes data quality,
> confidence and limitations visible instead of hiding them behind polished charts.

Point to the startup status, architecture summary and research boundary.

### Minute 1–2: Live traffic workspace

Open the Live workspace and choose Azerbaijan, the Caucasus, Türkiye or World.

Show:

- synchronized map and aircraft explorer;
- search and deterministic sorting;
- shared aircraft selection;
- current, aging and stale snapshot semantics;
- pause, resume and retry behavior;
- shareable URL state.

### Minute 2–3: Aircraft and route evidence

Select an aircraft and open Intelligence.

Explain that Route Intelligence is an inference with evidence, confidence, provenance
and limitations. It does not claim to possess a filed flight plan.

### Minute 3–4: Airport Intelligence

Open Airport Intelligence.

Show:

- global ranking and airport search;
- digital airport passport;
- completed-day statistics;
- history, trends, continuity and published limitations.

Mention that the frontend validates the API response but does not recreate the backend
ranking formula.

### Minute 4–5: Historical Intelligence

Open Historical Analytics.

Demonstrate one global metric, one airport metric and one route-pair metric. Show bucket
status, coverage, confidence, previous-period comparison and persisted-record comparison.
Unavailable values remain unavailable instead of becoming synthetic zeros.

### Minute 5–6: Data quality and export

Return to live traffic. Open the Data Quality Lens and explain the independent structural
checks for identifiers, coordinates, motion, time, altitude and attribution.

Export CSV or GeoJSON. Point out deterministic schema, snapshot provenance and invalid
coordinate exclusion accounting.

### Minute 6–7: Engineering decisions

Open the repository and show:

1. `apps/api/internal` bounded contexts;
2. `database/migrations` and PostgreSQL contracts;
3. `apps/web/lib/api` runtime response validation;
4. `apps/web/tests` dependency-free contract tests;
5. `.github/workflows/backend-ci.yml` and `frontend-ci.yml`;
6. `pnpm verify:release` and `pnpm smoke:production`.

## Strong engineering decisions to discuss

- modular monolith rather than premature microservices;
- PostgreSQL as the durable analytical source of truth;
- explicit open-data evidence boundaries;
- nullable telemetry preserved end to end;
- read-only repeatable-read snapshots for multi-query analytics;
- exact keyset pagination and deterministic fingerprints;
- mutation authorization separated from public research reads;
- bounded retries, provider budgets and fallback evidence;
- backend-owned analytical semantics with frontend transport validation;
- rollback-safe incremental delivery and permanent audit tooling.

## Questions a reviewer may ask

### Why Go?

The backend needs predictable concurrency, explicit context ownership, efficient polling,
strong static types and a small deployable runtime. Go fits those constraints without
forcing a distributed architecture.

### Why PostgreSQL instead of a time-series database?

The MVP needs transactions, constraints, repeatable reads, keyset pagination and rich
relational joins more than specialized high-volume time-series infrastructure. PostgreSQL
keeps operational complexity bounded while remaining extensible.

### Why no microservices?

The domains are separated in code, but deployment remains one backend because the team,
traffic volume and operational budget do not justify network boundaries, distributed
transactions or independent service ownership.

### How is bad source data handled?

Provider parsing, canonical normalization, data-quality checks, nullable availability,
confidence, limitations and provenance are explicit. Invalid evidence is rejected or
published as partial; it is not silently converted into plausible defaults.

### What would come next?

Only after real usage evidence: production telemetry, observed bottlenecks, improved demo
data, broader historical materialization and selected domain depth. Authentication,
billing, mobile clients or service extraction would require a concrete product need.

## Demo discipline

- Do not call inferred routes confirmed routes.
- Do not call transponder evidence a confirmed emergency.
- Do not claim a live deployment unless the exact URL and revision passed the smoke test.
- Do not hide empty, partial or stale states.
- Do not spend the demo reading documentation line by line.
