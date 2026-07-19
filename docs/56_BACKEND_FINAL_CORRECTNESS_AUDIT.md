# Document 56 — Backend Final Correctness Audit

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: final reproducible backend correctness gate before Stage 15

## 1. Purpose

Stage 14 corrected four high-risk backend boundaries:

```text
Projection Intelligence read snapshot consistency
nullable telemetry integrity
Historical Intelligence pagination integrity
Weather production composition
```

Passing tests once is not enough. The repository needs a permanent,
reproducible gate that prevents these defects from returning while later work
focuses on frontend visual design.

This document defines that gate.

## 2. Added Verification Assets

The repository now contains:

```text
apps/api/tools/backendfinalaudit
scripts/verify-backend-final-correctness.sh
```

`backendfinalaudit` performs source-level invariant checks.

`verify-backend-final-correctness.sh` combines those invariant checks with
existing architecture audits, focused package tests, race detection, complete
Go compilation, static analysis, and the complete backend test suite.

## 3. Projection Snapshot Consistency Gate

The audit requires:

```text
DataSource exposes only LoadSnapshot
Service performs one LoadSnapshot call
Service does not coordinate four independent reads
PostgreSQL uses REPEATABLE READ
PostgreSQL transaction is READ ONLY
trajectory repository and SQL client use the same transaction
successful reads commit
failed reads roll back with a bounded uncancelled context
```

This protects the corrected invariant that current trajectory, route,
historical candidates, and route history belong to one PostgreSQL snapshot.

## 4. Nullable Telemetry Integrity Gate

The audit requires both trajectory point queries to:

```text
select nullable telemetry without zero COALESCE defaults
filter incomplete required telemetry with IS NOT NULL
preserve explicit numerical zero values
```

The scanner must continue to use:

```text
pgtype.Float8
pgtype.Bool
completeRequiredTelemetry
```

Rows with missing required telemetry remain excluded. Valid zero latitude,
longitude, velocity, heading, vertical rate, or `on_ground=false` remain
distinguishable from missing values.

## 5. Historical Pagination Integrity Gate

The audit requires the cursor to contain:

```text
WindowEnd
WindowStart
AsOfTime
ID
```

The PostgreSQL keyset predicate must match:

```text
window_end DESC
window_start DESC
as_of_time DESC
id ASC
```

The HTTP boundary must continue to use:

```text
cursor
next_cursor
historical-aggregate-cursor-v1
URL-safe unpadded Base64
strict JSON decoding
complete cursor normalization
```

Legacy single-field cursor names are forbidden from the Historical
Intelligence production contract.

## 6. Weather Composition Boundary Gate

The audit requires:

```text
weather_route.go remains a narrow coordinator
provider composition owns provider governance and integration
application composition owns repository, service, and handler
route registration owns only the HTTP boundary
the established Open-Meteo timeout error remains unchanged
GET /api/v1/weather/current remains unchanged
```

The audit blocks provider, persistence, service, or handler constructors from
returning to route registration.

## 7. Existing Architecture and Security Gates

The final verification script also runs:

```text
tools/projectaudit -mode all -strict
```

This preserves existing checks for:

```text
shared confidence vocabulary
Go and TypeScript trajectory contract alignment
mutation route authorization
formula benchmark isolation
analytical production reachability
```

The new correctness audit complements these checks rather than duplicating
their implementation.

## 8. Runtime Verification Coverage

Focused compilation and tests include:

```text
Projection Intelligence PostgreSQL HTTP verifier
Historical Intelligence PostgreSQL HTTP verifier
Historical Aggregate Store verifier
Historical materialization replay verifier
Weather Context PostgreSQL HTTP verifier
Stability Intelligence PostgreSQL HTTP verifier
Airspace Region Analytics PostgreSQL HTTP verifier
source constraint verifier
OpenSky REST compatibility verifier
```

Database-dependent verifier commands are compiled and their package tests run.
They are not forced to connect to an external PostgreSQL database during an
offline source audit.

## 9. Race Detection

The race detector covers the corrected and concurrency-sensitive boundaries:

```text
Projection read
Historical Aggregate Store and cursor contract
Historical HTTP cursor
server composition
Weather service and provider orchestration
provider budget and response state
PostgreSQL repositories
internal API-key authorization
```

The complete backend test suite then runs without caching assumptions imposed
by this document.

## 10. Reproducible Command

From the repository root:

```bash
scripts/verify-backend-final-correctness.sh
```

Successful completion ends with:

```text
BACKEND_FINAL_CORRECTNESS_AUDIT=PASS
```

## 11. Non-Goals

This audit does not claim:

```text
production load-test evidence
external PostgreSQL availability
deployed Render or Neon health
frontend visual completion
browser-level end-to-end coverage
```

Those belong to later deployment, performance, and frontend acceptance work.

## 12. Acceptance

The backend final correctness audit is accepted only after:

```text
backendfinalaudit unit tests
backendfinalaudit strict repository scan
projectaudit strict scan
focused corrected-boundary tests
focused race detector
all command packages build
go vet
complete go test
frontend dependency security verification
frontend lint
frontend TypeScript validation
frontend production build
backend Docker image build
git diff check
```

After this gate is committed, Stage 14 backend correction work is closed and
the project may proceed to Stage 15 Frontend Visual System and Application
Layout.
