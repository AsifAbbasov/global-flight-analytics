# Document 65 — Stage 14.24 Explicit Altitude Integer Policy

Status: Implementation Baseline v1.0
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
