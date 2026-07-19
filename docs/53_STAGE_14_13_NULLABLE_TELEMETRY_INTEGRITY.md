# Document 53 — Stage 14.13 Nullable Telemetry Integrity

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: prevent absent flight-state telemetry from becoming plausible values

## 1. Problem

The Projection Intelligence trajectory queries previously used:

```text
COALESCE(latitude, 0)
COALESCE(longitude, 0)
COALESCE(velocity_mps, 0)
COALESCE(heading_degrees, 0)
COALESCE(vertical_rate_mps, 0)
COALESCE(on_ground, false)
```

The PostgreSQL columns permit `NULL`.

These expressions destroyed the difference between:

```text
a measured zero value
and
an unavailable value
```

This could turn an observation without a position into the valid geographic
coordinate `(0, 0)`. It could also turn unavailable motion telemetry into a
stationary aircraft with zero heading, zero vertical rate, and an explicit
airborne state.

Range validation could not detect this because zero is valid for every listed
numeric field.

## 2. Production Decision

Projection Intelligence now accepts only flight-state rows with complete
required kinematic telemetry:

```text
latitude
longitude
velocity
heading
vertical rate
on-ground state
```

The SQL boundary explicitly requires each field to be non-null.

The selected values remain nullable during scanning through PostgreSQL nullable
types. The scanner performs a second completeness check before constructing a
domain trajectory point.

This provides defense in depth:

```text
SQL excludes incomplete observations
scanner rejects any incomplete row that still reaches the application
```

## 3. Why Incomplete Rows Are Omitted

`trajectory.TrackPoint4D` currently represents a usable analytical point. It
does not represent a partially observed database row.

Placing fabricated defaults into this type would create false evidence.
Expanding every projection algorithm and every trajectory consumer with
partially available motion fields would be a larger domain-contract migration.

For the current projection contract, conservative omission is the correct
boundary:

```text
complete observation → usable TrackPoint4D
incomplete observation → no TrackPoint4D
```

If no usable points remain, existing eligibility and projection contracts
produce an unavailable result rather than a plausible false projection.

## 4. Legitimate Zero Values

The implementation does not classify values as missing by comparing them with
zero.

Availability is taken only from PostgreSQL nullability metadata.

Therefore all of the following remain valid when explicitly stored:

```text
latitude = 0
longitude = 0
velocity = 0
heading = 0
vertical rate = 0
on_ground = false
```

This is essential because the equator, prime meridian, stationary motion,
north heading, level flight, and an airborne state are all legitimate values.

## 5. Altitude Semantics

Altitude was already read through nullable PostgreSQL values and separate
altitude-status fields.

This increment preserves that design:

```text
horizontal and motion telemetry must be complete
altitude may remain unavailable
altitude status continues to describe its evidence state
```

A horizontal projection can therefore remain available with an explicit
altitude limitation.

## 6. Limit and Ordering Semantics

The non-null predicates are applied inside PostgreSQL before:

```text
ORDER BY observed_at, id
LIMIT
```

Incomplete rows do not consume the trajectory point limit and cannot hide later
complete rows.

Ordering and deterministic tie-breaking remain unchanged.

## 7. Preserved Behavior

This increment does not change:

```text
projection formulas
confidence weights
trajectory freshness policy
historical-neighbor policy
arrival calculations
PostgreSQL schema
migrations
HTTP contracts
provider ingestion
frontend behavior
```

It changes only which persisted observations qualify as analytical trajectory
points.

## 8. Regression Gates

Automated tests require:

```text
no telemetry COALESCE to zero or false
all required telemetry columns to have IS NOT NULL predicates
scanner destinations to retain PostgreSQL nullability
each missing required field to reject the row
legitimate zero values to remain usable
hydration to omit incomplete rows without returning false data
```

Static architecture tests protect the SQL and scanner boundary from returning
to default-value fabrication.

## 9. Acceptance

The increment is accepted only after:

```text
focused Projection Intelligence tests
nullable telemetry behavior tests
architecture regression tests
race detector
strict project architecture audit
complete Go build
go vet
complete Go test suite
frontend dependency security verification
frontend production dependency audit
ESLint
TypeScript validation
Next.js production build
backend Docker image build
git diff check
```
