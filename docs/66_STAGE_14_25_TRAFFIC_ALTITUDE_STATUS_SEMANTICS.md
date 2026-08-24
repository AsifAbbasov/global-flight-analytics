# Document 66 — Stage 14.25 Traffic Altitude Status Semantics

Status: Remediation History v1.1
Project: Global Flight Analytics
Scope: preserve typed altitude meaning from PostgreSQL current-traffic reads through the public HTTP contract and frontend presentation

## 1. Correctness problem

The former current-traffic query reduced geometric and barometric altitude into one
number through this shape:

```text
geometric altitude unless it equals zero
otherwise barometric altitude
otherwise zero
```

That logic introduced two correctness failures:

1. a legitimate observed geometric altitude of zero was treated as missing;
2. unknown, unavailable, or invalid altitude evidence was published as the
   numeric value zero.

The HTTP and frontend contracts could not distinguish an observed zero, an
aircraft on the ground, and a missing altitude.

## 2. Typed current-traffic altitude contract

Each current-traffic item now carries:

```text
altitude_m       nullable numeric value
altitude_status  observed | ground | unknown | unavailable | invalid
altitude_source  geometric | barometric | ground | none
```

A missing altitude is represented by `null`, never by a fabricated zero.

## 3. Selection policy

The current-traffic altitude resolver uses this deterministic order:

```text
on_ground = true
    → altitude 0, status ground, source ground

usable observed geometric altitude
    → preserve the value, including observed zero

usable observed barometric altitude
    → use barometric fallback

no usable observed altitude
    → null value and the strongest available status
```

For absent evidence, invalid dominates unknown, and unknown dominates
unavailable. A ground altitude status without independent `on_ground` evidence is
invalid.

## 4. PostgreSQL read behavior

Both unbounded and bounded current-traffic queries load the geometric and
barometric values and statuses independently. They no longer use `NULLIF`,
`COALESCE`, or zero as altitude-selection logic.

The repository scans nullable values and delegates selection to the domain
resolver. This keeps SQL responsible for data retrieval and the domain layer
responsible for meaning.

## 5. Public HTTP behavior

The endpoint remains:

```text
GET /api/v1/traffic/current
GET /api/v1/traffic/current?region=<code>
```

Existing fields remain available. The altitude value is now nullable and two
explicit fields are added:

```json
{
  "altitude_m": null,
  "altitude_status": "unknown",
  "altitude_source": "none"
}
```

An observed zero remains numeric:

```json
{
  "altitude_m": 0,
  "altitude_status": "observed",
  "altitude_source": "geometric"
}
```

## 6. Frontend behavior

The current traffic map and aircraft detail panel render altitude meaning
directly:

```text
Ground (0 m)
0 m (geometric)
2400 m (barometric)
Unknown
Unavailable
Invalid altitude evidence
```

The frontend no longer renders missing altitude as `0 m`.

## 7. Regression protection

Permanent tests protect:

```text
observed geometric zero preservation
barometric fallback
ground representation
unknown and unavailable nullability
invalid evidence handling
both current-traffic query variants
HTTP DTO propagation
removal of zero-sentinel SQL
frontend type checking, linting, and production build
```

The PostgreSQL integration test is enabled when `TEST_DATABASE_URL` is set and
uses an isolated temporary schema.

## 8. Acceptance commands

From `apps/api`:

```bash
gofmt -w internal/domain/traffic/altitude.go internal/domain/traffic/altitude_test.go internal/domain/traffic/model.go internal/repository/postgres/traffic_repository.go internal/repository/postgres/traffic_altitude_semantics_ownership_test.go internal/repository/postgres/traffic_altitude_semantics_integration_test.go internal/http/dto/traffic.go internal/http/handlers/traffic.go internal/http/handlers/traffic_altitude_semantics_test.go
go test -count=1 ./internal/domain/traffic ./internal/repository/postgres ./internal/http/handlers
go test -count=1 ./...
go vet ./...
```

From the repository root:

```bash
pnpm --dir apps/web typecheck
pnpm --dir apps/web lint
pnpm --dir apps/web build
git diff --check
git status --short
```

## 9. Completion boundary

This increment closes Traffic altitude-status semantics.

It does not close:

```text
Airport elevation semantics
timestamp and Unix-nanosecond consistency
large PostgreSQL repository decomposition
```

## 10. Finding history, root cause, and failure scenario

### Finding

The current-traffic read path collapsed altitude value, availability, status, and source
into a single numeric field with zero used as both real data and a missing-value sentinel.

### Root cause

SQL query convenience (`NULLIF`/`COALESCE`-style fallback) was allowed to choose domain
meaning. The HTTP contract then exposed only a number, so downstream code had no way to
recover whether zero meant observed sea level, ground, unavailable evidence, or fallback.

### Failure scenario

```text
geometric altitude is legitimately observed as 0 m
↓
query treats zero as missing and falls back
or
both altitude sources are missing/invalid
↓
query publishes numeric 0
↓
HTTP/frontend render a factual altitude that was never observed
```

### Impact

Incorrect altitude meaning affects map display, aircraft detail UX, traffic analytics,
route context, and any consumer that interprets zero as physical evidence. The defect was
plausible-looking rather than obviously invalid, increasing the chance of silent
analytical error.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P1 data
semantics** because the public traffic contract could fabricate a valid-looking physical
measurement and erase the distinction between observed zero and missing evidence.

### Existing guarantees violated

```text
missing evidence is not converted into factual numeric data
observed zero is preserved as a legitimate measurement
ground is explicit and requires independent on-ground evidence
SQL retrieves evidence; domain code owns selection semantics
HTTP/frontend preserve availability and source provenance
```

## 11. Considered and rejected alternatives

### Keep numeric-only API and reserve a sentinel value

Rejected because every numeric sentinel collides with possible physical values or creates
an undocumented out-of-domain convention.

### Prefer barometric altitude unconditionally

Rejected because it would avoid the zero bug but discard available geometric evidence and
still fail to represent missing/invalid status explicitly.

### Fix only frontend formatting

Rejected because the semantic loss happened before HTTP serialization. The frontend
cannot reconstruct information already collapsed by SQL/repository logic.

### Encode missing altitude as zero plus a boolean

Rejected because status and source have more than two states and the project already owns
typed altitude semantics.

### Chosen remediation

Carry nullable altitude value together with explicit `status` and `source`, resolve
geometric/barometric/ground precedence in domain code, and preserve those semantics
through PostgreSQL, HTTP, TypeScript, and UI.

## 12. Why this solution and trade-offs

The chosen model makes information loss explicit and keeps meaning out of SQL shorthand.

Trade-offs:

```text
+ observed zero remains truthful
+ missing/unknown/invalid evidence remains distinguishable
+ frontend and API become provenance-aware
+ one resolver owns precedence rules
- HTTP/TypeScript contract becomes slightly richer
- consumers must handle nullable altitude and statuses
```

The added contract fields are justified because they remove ambiguity rather than add
ceremony.

## 13. Adversarial review and remediation iterations

### Iteration 1 — detect zero-sentinel corruption

Review identified both failure modes: legitimate zero being discarded and missing altitude
being fabricated as zero.

### Iteration 2 — end-to-end typed remediation

Implementation commit `18545100dda0e9852927dc4a93dafcc25394b1e1`
(`fix: preserve traffic altitude semantics`) changed domain resolution, repository reads,
HTTP DTOs/handlers, and frontend presentation together.

### Iteration 3 — query ownership challenge

Regression tests prohibit zero-sentinel SQL so a future query refactor cannot reintroduce
semantic selection below the domain resolver.

### Iteration 4 — later live-traffic hardening

A later change (`89c9b13292efe60c7924af750ab444d21dc4536a`, PR #60) preserved nullable altitude and
valid-zero semantics while improving selection of the latest displayable traffic snapshot.
This is evidence that subsequent traffic work retained, rather than bypassed, the typed
altitude contract.

## 14. Residual risks and limitations

The remediation does not validate whether provider altitude is physically accurate; it
preserves the provider/storage evidence honestly. Source disagreements and stale altitude
are separate quality/freshness problems.

## 15. Operational/deployment consequences

Clients must accept nullable `altitude_m` and use status/source for presentation. No schema
migration is required for the semantic change, but backend and frontend contracts should
be deployed compatibly.

## 16. Exact evidence

```text
implementation commit:
18545100dda0e9852927dc4a93dafcc25394b1e1

later preservation evidence:
89c9b13292efe60c7924af750ab444d21dc4536a (PR #60)

regression coverage:
internal/domain/traffic/altitude_test.go
internal/repository/postgres/traffic_altitude_semantics_ownership_test.go
internal/repository/postgres/traffic_altitude_semantics_integration_test.go
internal/http/handlers/traffic_altitude_semantics_test.go
frontend typecheck/lint/build contracts
```

## 17. Final canonical status

```text
FINDING_GFA_DATA_009_TRAFFIC_ALTITUDE_STATUS_SEMANTICS=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/66_STAGE_14_25_TRAFFIC_ALTITUDE_STATUS_SEMANTICS.md
IMPLEMENTATION_COMMIT=18545100dda0e9852927dc4a93dafcc25394b1e1
```
