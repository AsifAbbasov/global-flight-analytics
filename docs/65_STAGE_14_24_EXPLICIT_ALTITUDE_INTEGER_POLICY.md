# Document 65 — Stage 14.24 Explicit Altitude Integer Policy

Status: Remediation History v1.1
Project: Global Flight Analytics
Scope: make conversion from provider altitude precision to PostgreSQL whole-meter storage explicit, deterministic, and testable

## 1. Correctness problem

The domain and provider boundaries carry altitude in metres as `float64`, while
`flight_states.barometric_altitude_m` and
`flight_states.geometric_altitude_m` use PostgreSQL `integer` columns.

The former repository passed floating-point values into SQL and used:

```sql
CAST($value::double precision AS integer)
```

That delegated rounding and overflow behavior to PostgreSQL. The Go code did not
state the conversion policy, could not classify non-finite values before the
insert, and had no permanent test protecting the exact whole-metre behavior.

## 2. Canonical conversion contract

The only supported observed-altitude conversion is now:

```text
finite float64 metres
↓
round to the nearest whole metre
↓
exact half values round away from zero
↓
verify PostgreSQL integer range
↓
persist int32
```

The implementation uses `math.Round` and performs the range check before the
conversion to `int32`.

Examples:

```text
9753.49 m  →  9753 m
9753.50 m  →  9754 m
-12.49 m   →  -12 m
-12.50 m   →  -13 m
```

Negative altitude remains valid because an aircraft or airport can be below
mean sea level. This policy enforces the storage representation, not an
aviation-domain minimum or maximum.

## 3. Rejected observed values

An observed altitude is rejected when it is:

```text
NaN
positive infinity
negative infinity
outside PostgreSQL integer range after rounding
```

The repository returns a typed error. `SaveFlightStates` does not commit a
partially written batch when conversion fails.

## 4. Status semantics remain unchanged

The altitude status remains authoritative:

```text
observed     → apply the explicit whole-metre conversion

ground       → persist integer zero

unknown      → persist NULL
unavailable  → persist NULL
invalid      → persist NULL
```

A non-finite numeric placeholder attached to a non-value status does not enter
the database because the numeric column remains `NULL`.

## 5. SQL responsibility

SQL now receives an already validated `pgtype.Int4` value. The flight-state
insert no longer contains a floating-point-to-integer cast for either altitude
column.

PostgreSQL remains responsible for enforcing the column type. It is no longer
responsible for choosing the application rounding policy.

## 6. Schema and migration impact

No PostgreSQL migration is required.

The existing columns already store whole metres as `integer`, and existing rows
already contain integer values. This increment changes only how future observed
values are prepared before insertion.

## 7. Regression protection

Permanent tests protect:

```text
positive and negative rounding behavior
exact half-value behavior
observed zero
integer boundaries
NaN and infinity rejection
post-rounding overflow rejection
ground and unavailable status semantics
real PostgreSQL persistence of rounded values
batch rollback after invalid observed altitude
absence of SQL-owned altitude casting
```

## 8. Acceptance commands

From `apps/api`:

```bash
gofmt -w internal/repository/postgres/flightstate_repository.go \
  internal/repository/postgres/altitude_meter_policy.go \
  internal/repository/postgres/altitude_meter_policy_test.go \
  internal/repository/postgres/altitude_meter_policy_ownership_test.go \
  internal/repository/postgres/altitude_meter_policy_integration_test.go

go test -count=1 ./internal/repository/postgres
go test -count=1 ./...
go vet ./...
```

From the repository root:

```bash
git diff --check
git status --short
```

## 9. Completion boundary

This increment closes the hidden altitude precision and integer conversion
debt. It does not change provider altitude measurement precision, analytical
altitude semantics, or the separate Traffic handling of altitude availability
status.

## 10. Finding history, root cause, and failure scenario

### Finding

Observed provider altitude arrived as floating-point metres while PostgreSQL persisted
whole-meter integers, but the application had no declared conversion rule.

### Root cause

The repository delegated an application-level semantic decision to an inline SQL cast.
That made rounding, overflow, and non-finite handling implicit database behavior instead
of an owned domain/persistence policy.

### Failure scenario

```text
provider supplies floating-point observed altitude
↓
repository passes float64 directly into SQL CAST(... AS integer)
↓
rounding/overflow semantics are chosen implicitly by PostgreSQL
↓
Go validation/tests cannot prove the exact persisted whole-meter value
```

A non-finite or out-of-range observed value could also reach the SQL boundary before the
application classified it.

### Impact

Implicit conversion could make persisted trajectory/traffic altitude differ from the
application's intended value and make behavior difficult to reproduce across refactors.
It also weakened failure diagnostics for invalid provider evidence.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P2 data
semantics** because the issue could alter persisted measurement values but was bounded to
whole-meter conversion rather than destroying identity or lifecycle state.

### Existing guarantees violated

```text
persistence conversion policy is explicit and deterministic
invalid observed evidence fails before durable write
observed zero and negative altitude remain valid measurements
non-value statuses do not fabricate numeric altitude
batch persistence does not partially commit after conversion failure
```

## 11. Considered and rejected alternatives

### Keep SQL casting and document PostgreSQL behavior

Rejected because the application still would not own or test the conversion before the
write boundary, and typed errors for non-finite/out-of-range values would remain weak.

### Change PostgreSQL altitude columns to floating point

Rejected because the existing storage contract intentionally uses whole meters and there
was no product need for a destructive schema change merely to avoid defining conversion.

### Truncate toward zero in Go

Rejected because truncation creates a different error profile and was not the intended
nearest-meter representation.

### Clamp values into integer range

Rejected because clamping would fabricate a valid-looking altitude from invalid evidence.
Fail closed is safer.

### Chosen remediation

Define one Go-owned conversion helper using nearest whole meter, `math.Round` half-away-
from-zero semantics, explicit finite/range validation, and typed integer output passed to
SQL.

## 12. Why this solution and trade-offs

The persistence adapter owns the representation change because it knows both the domain
input type and database storage type.

Trade-offs:

```text
+ deterministic/testable whole-meter semantics
+ invalid observed evidence rejected before SQL
+ SQL no longer chooses application rounding policy
- sub-meter precision is intentionally discarded at persistence boundary
- future storage-precision changes require an explicit migration/policy revision
```

The precision loss is not new; the remediation makes the already-existing integer storage
choice explicit and reproducible.

## 13. Adversarial review and remediation iterations

### Iteration 1 — identify hidden SQL policy

Review found that the SQL cast was silently deciding rounding and overflow semantics.

### Iteration 2 — explicit conversion helper

Implementation commit `467d6bf5f6e66febbb83944664735ce26e7054c3`
(`fix: enforce explicit altitude integer policy`) moved the policy into typed Go code.

### Iteration 3 — edge-value challenge

Tests cover exact half values, negative altitude, integer boundaries, NaN/infinity,
post-rounding overflow, observed zero, and batch rollback to ensure the helper is not only
a happy-path refactor.

### Iteration 4 — semantic separation challenge

The remediation deliberately leaves altitude availability/status semantics to the
separate Traffic and provider contracts rather than treating integer conversion as proof
that an altitude value exists.

## 14. Residual risks and limitations

This policy does not improve provider measurement accuracy or recover sub-meter precision
after persistence. It does not impose aviation-specific physical altitude limits; such a
policy would require separate evidence and domain justification.

## 15. Operational/deployment consequences

No schema migration is required. Invalid observed altitude now fails earlier and may
surface typed persistence errors where PostgreSQL previously performed an implicit cast.
Operational handling should treat those as provider/data correctness findings rather than
silently clamping or retrying unchanged input.

## 16. Exact evidence

```text
implementation commit:
467d6bf5f6e66febbb83944664735ce26e7054c3

regression coverage:
internal/repository/postgres/altitude_meter_policy_test.go
internal/repository/postgres/altitude_meter_policy_ownership_test.go
internal/repository/postgres/altitude_meter_policy_integration_test.go
```

## 17. Final canonical status

```text
FINDING_GFA_DB_008_EXPLICIT_ALTITUDE_INTEGER_POLICY=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/65_STAGE_14_24_EXPLICIT_ALTITUDE_INTEGER_POLICY.md
IMPLEMENTATION_COMMIT=467d6bf5f6e66febbb83944664735ce26e7054c3
```

Historical PR/reviewer identifiers are not invented when unavailable.
