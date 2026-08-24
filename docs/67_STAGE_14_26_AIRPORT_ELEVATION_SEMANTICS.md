# Document 67 — Stage 14.26 Airport Elevation Semantics

Status: Remediation History v1.1
Project: Global Flight Analytics
Scope: preserve the distinction between unknown airport elevation and observed sea-level elevation

## 1. Correctness problem

The airport repository previously selected `COALESCE(a.elevation_ft, 0)`.
A missing OurAirports elevation therefore became a factual zero-foot elevation.
The domain and HTTP contracts had no availability or status field, so the
incorrect value propagated into Airport profiles, Airport Intelligence,
preliminary route context, production Route Intelligence, and the frontend.

## 2. Canonical semantics

Airport elevation now has three public states:

```text
observed — a finite source value exists, including exactly zero
unknown  — the source column is NULL and no elevation is claimed
invalid  — an explicitly supplied value is non-finite
```

A non-zero legacy in-memory value remains observed for compatibility. A zero
value is observed only when availability is explicit.

## 3. PostgreSQL boundary

Both airport repository queries select nullable `elevation_ft` directly into
`pgtype.Int4`. NULL maps to unknown. Any valid integer, including zero and
negative values, is converted from international feet to metres and marked
observed. No schema migration is required because the column is already
nullable.

## 4. Propagation boundary

The semantics are preserved through:

```text
Airport domain
Airport profile HTTP API
Airport Intelligence passport
preliminary aircraft route context
production Route Intelligence contract and HTTP API
TypeScript route contracts
aircraft route-context panel
```

Public JSON uses nullable `elevation_m` plus `elevation_status`. Unknown or
invalid evidence is never serialized as a factual zero.

## 5. Regression protection

The permanent tests protect:

```text
NULL versus observed zero PostgreSQL mapping
negative and legacy non-zero elevation compatibility
non-finite value rejection
absence of COALESCE(a.elevation_ft, 0)
route catalog availability fingerprint input
nullable HTTP conversion
frontend nullable elevation formatting
```

## 6. Acceptance commands

From `apps/api`:

```bash
gofmt -w internal/domain/airport internal/repository/postgres internal/airportintelligence/passport internal/routeintelligence/airportresolver internal/routeintelligence/routecontract internal/http/dto internal/http/handlers
go test -count=1 ./internal/domain/airport ./internal/repository/postgres ./internal/airportintelligence/passport ./internal/routeintelligence/airportresolver ./internal/routeintelligence/routecontract ./internal/http/dto ./internal/http/handlers
go test -count=1 ./...
go vet ./...
```

From `apps/web`:

```bash
pnpm typecheck
pnpm lint
pnpm build
```

From the repository root:

```bash
git diff --check
git status --short
```

## 7. Completion boundary

This increment closes Airport elevation semantics. It does not close timestamp
and Unix-nanosecond consistency or large PostgreSQL repository decomposition.

## 8. Finding history, root cause, and failure scenario

### Finding

The airport read path converted `NULL` source elevation into numeric zero, making an
unknown measurement indistinguishable from a real airport at mean sea level.

### Root cause

SQL convenience used `COALESCE` to satisfy a non-null application field while the domain
model lacked explicit availability/status semantics. Once the repository fabricated zero,
HTTP, analytics, route context, and frontend layers could not recover whether zero had
actually been observed.

### Failure scenario

```text
OurAirports elevation_ft is NULL
↓
repository query applies COALESCE(..., 0)
↓
Airport domain receives 0 ft as factual data
↓
Airport Intelligence / Route Intelligence / frontend propagate sea-level elevation
↓
consumer cannot distinguish missing evidence from observed zero
```

### Impact

The defect could bias airport profiles, route context, elevation-aware reasoning, and UI
presentation with plausible but unsupported data. It also corrupted provenance by
turning absence into a measurement.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P2 data
semantics/provenance** because the problem fabricated a field value but did not corrupt
entity identity or database lifecycle state.

### Existing guarantees violated

```text
missing source evidence remains missing
observed zero remains a legitimate value
negative airport elevation remains representable
SQL retrieves source data rather than inventing domain defaults
HTTP and frontend preserve source availability
```

## 9. Considered and rejected alternatives

### Keep `COALESCE(..., 0)` and add an explanatory UI label

Rejected because the evidence was already lost before the UI boundary and analytics would
still consume fabricated zero.

### Use a numeric sentinel such as a negative impossible value

Rejected because negative elevation is physically valid for airports below mean sea level,
and sentinels would mix representation concerns with domain values.

### Make every consumer query the raw nullable database field independently

Rejected because it would duplicate availability semantics across Airport Intelligence,
Route Intelligence, HTTP, and frontend code.

### Remove airport elevation entirely

Rejected because observed source elevation is useful evidence; the problem was absence
handling, not the field itself.

### Chosen remediation

Represent elevation as value plus explicit status, map PostgreSQL NULL to `unknown`, and
propagate nullable/status semantics end to end.

## 10. Why this solution and trade-offs

The chosen contract preserves information rather than guessing it.

Trade-offs:

```text
+ no fabricated sea-level airports
+ legitimate observed zero and negative elevation remain valid
+ one semantic contract reaches analytics and UI
- domain/HTTP/TypeScript contracts become slightly richer
- consumers must handle unknown elevation explicitly
```

The added status field is justified because it removes a real ambiguity shared by several
production paths.

## 11. Adversarial review and remediation iterations

### Iteration 1 — repository semantic finding

Review located `COALESCE(a.elevation_ft, 0)` and followed the fabricated value through
profiles, intelligence, route context, and frontend output.

### Iteration 2 — end-to-end remediation

Implementation commit `75247fd242b293de65fa85f164fd594c64343b9a`
(`fix: preserve airport elevation semantics`) introduced nullable PostgreSQL mapping and
status-aware propagation through backend and frontend contracts.

### Iteration 3 — zero/negative compatibility challenge

Tests explicitly distinguish NULL from observed zero and preserve negative elevation so
the remediation cannot replace one sentinel bug with another.

### Iteration 4 — route-cache/fingerprint challenge

Regression coverage includes route catalog availability fingerprint input so changing
known/unknown elevation evidence invalidates the relevant derived context instead of
silently reusing stale semantics.

## 12. Residual risks and limitations

The remediation does not verify the accuracy or freshness of the upstream OurAirports
elevation itself. It preserves whether evidence exists and prevents the application from
inventing a value. Source corrections and dataset refresh policy remain separate concerns.

## 13. Operational/deployment consequences

No schema migration is required. API/frontend deployments must tolerate nullable
`elevation_m` and explicit `elevation_status`. Unknown values should remain absent/unknown
rather than being replaced by display defaults that look factual.

## 14. Exact evidence

```text
implementation commit:
75247fd242b293de65fa85f164fd594c64343b9a

regression coverage:
PostgreSQL NULL vs zero mapping tests
absence of COALESCE(a.elevation_ft, 0) ownership check
Airport/Route Intelligence propagation tests
frontend nullable elevation formatting checks
```

## 15. Final canonical status

```text
FINDING_GFA_DATA_010_AIRPORT_ELEVATION_SEMANTICS=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/67_STAGE_14_26_AIRPORT_ELEVATION_SEMANTICS.md
IMPLEMENTATION_COMMIT=75247fd242b293de65fa85f164fd594c64343b9a
```

Historical PR/reviewer identifiers are not invented when unavailable.
