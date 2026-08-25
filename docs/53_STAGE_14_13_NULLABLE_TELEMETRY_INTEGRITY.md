# Document 53 — Stage 14.13 Nullable Telemetry Integrity

Status: Remediation History v1.1
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

## 10. Canonical remediation history

### Finding / symptom

Projection reads fabricated valid-looking telemetry from database absence by using numeric and boolean `COALESCE` defaults.

### Root cause

The SQL read boundary collapsed PostgreSQL nullability into non-nullable analytical values before domain validation. Because zero and `false` are legitimate values, downstream range checks could not recover the lost availability information.

### Failure scenario

```text
persisted observation contains NULL telemetry
↓
Projection query COALESCEs NULL to 0 / false
↓
scanner receives syntactically valid values
↓
TrackPoint4D is constructed as if telemetry had been observed
↓
projection consumes fabricated movement/position evidence
```

### Impact

The defect could produce plausible but unsupported trajectories and projections, including the valid coordinate `(0,0)` and apparently stationary or level-flight states. This is an analytical correctness and evidence-integrity failure rather than a presentation-only issue.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P1 data correctness** because missing source evidence could be transformed into valid analytical evidence and influence Projection Intelligence outputs without an observable validation failure.

### Existing guarantees violated

```text
missing telemetry must remain distinguishable from observed zero
analytical inputs must not fabricate provider/database evidence
TrackPoint4D must represent usable observations
query limits must count usable rows rather than fabricated placeholders
```

### Considered solutions

1. Keep `COALESCE` and detect zeros later.
2. Make every `TrackPoint4D` kinematic field optional immediately.
3. Preserve SQL nullability and conservatively omit incomplete rows for the current Projection contract.

### Chosen remediation

Option 3: remove telemetry default fabrication, add `IS NOT NULL` eligibility predicates, scan through nullable PostgreSQL types, and enforce a second completeness check before domain construction.

### Why this solution was selected

It restores evidence integrity without forcing a cross-project optional-telemetry redesign in the same increment. It also preserves legitimate zero values exactly.

### Rejected alternatives

Zero-based missing detection was rejected because zero is a valid latitude, longitude, velocity, heading, vertical rate, and boolean false is a valid on-ground value. A repository-wide partial-TrackPoint migration was rejected as broader than the confirmed defect and unnecessary for the current Projection eligibility contract.

### Trade-offs

```text
+ absent telemetry cannot become plausible analytical evidence
+ legitimate zero values remain valid
+ incomplete rows do not consume point limits
- Projection may have fewer usable points and therefore return unavailable more often
- partial observations remain unavailable to algorithms that could theoretically use subsets of telemetry
```

### Regression tests / protection

The protected boundary includes SQL anti-`COALESCE` checks, required `IS NOT NULL` predicates, nullable scanner destinations, per-field missing-value tests, explicit legitimate-zero tests, and architecture audits.

### Adversarial review findings

The important adversarial case was that `0` cannot be treated as a sentinel: `(0,0)`, zero speed, zero heading, zero vertical rate and `false` are all legitimate. Review also required predicates to run before `ORDER BY ... LIMIT`, so incomplete rows cannot consume the result budget.

Historical PR/reviewer evidence for the original July remediation is unavailable; this reconstruction is limited to repository source, tests, commits, and the stage document.

### Remediation iterations

The direct Projection read fix landed in implementation commit `1f30bae8bb8a9e4e27d634b44362dcf7547e54ff` (`fix: preserve nullable telemetry semantics`). Stage 14.16 later proved that read-side protection alone was insufficient if provider mapping had already destroyed availability; Document 57 records the end-to-end follow-up.

### Residual risks and limitations

This remediation protects Projection reads from persisted NULL fabrication. By itself it cannot recover absence already converted to zero before persistence, and it does not make partially observed rows usable for algorithms with weaker telemetry requirements. Those limitations motivated the later end-to-end availability contract.

### Operational or deployment consequences

No schema migration was required. Deployments must preserve nullable telemetry in PostgreSQL and must treat reductions in eligible Projection points as evidence-quality behavior, not as a reason to restore synthetic defaults.

### Exact evidence

```text
implementation commit:
1f30bae8bb8a9e4e27d634b44362dcf7547e54ff

permanent evidence:
Projection PostgreSQL query tests
nullable scanner/completeness tests
backendfinalaudit protected SQL invariants
strict project architecture audit
```

### Final canonical status

```text
FINDING_GFA_DATA_051_NULLABLE_TELEMETRY_FABRICATION=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/53_STAGE_14_13_NULLABLE_TELEMETRY_INTEGRITY.md
IMPLEMENTATION_COMMIT=1f30bae8bb8a9e4e27d634b44362dcf7547e54ff
```

### Prevention / future guard

Any analytical SQL that reads nullable observations into non-nullable domain values must prove explicit availability semantics. Numerical/boolean `COALESCE` defaults for evidence-bearing telemetry are prohibited unless the default is itself part of a documented domain contract. Tests must continue to include both missing values and legitimate zero values.
