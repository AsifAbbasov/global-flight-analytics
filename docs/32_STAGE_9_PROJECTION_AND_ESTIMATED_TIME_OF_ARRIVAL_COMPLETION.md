# Document 32 — Stage 9 Projection and Estimated Time of Arrival Completion

Status: COMPLETED
Completion date: 2026-07-16
Project: Global Flight Analytics
Stage: 9 — Projection and Estimated Time of Arrival
Completion classification: Production Projection Intelligence Foundation
Evidence baseline commit: `c2927c7f6411c33f702244f4c9277cd91e44b753`

---

## 1. Purpose

This document closes Stage 9 with an evidence-based implementation record.

Stage 9 is complete as a bounded, explainable, research-only Production Projection Intelligence foundation. It can build short-horizon projections from observed trajectory state, select similar historical trajectories, evaluate historical pattern confidence and freshness, reject weak route history, generate local historical continuation, attach an Estimated Arrival when the route contract permits it, fall back deterministically to a conservative kinematic baseline, expose a read-only HTTP endpoint, and execute both prediction paths through the production PostgreSQL composition.

This completion statement does not claim:

```text
operational flight prediction
air traffic control suitability
flight planning suitability
safety-critical Estimated Arrival
weather-adjusted forecasting
airspace conflict prediction
machine-learning calibration
large-scale real-world forecast accuracy
frontend prediction visualization
```

Every result remains protected by:

```text
research-only scope guard
explicit method identity
bounded horizon
confidence
uncertainty
limitations
explanations
provenance
deterministic input fingerprint
auditable fallback reason
```

---

## 2. Scope Alignment

Stage 8 computes historical facts.

Stage 9 computes bounded future estimates.

The boundary is:

```text
Stage 8
historical windows
historical series
historical comparison
historical aggregate persistence
historical similarity evidence

Stage 9
future projection contract
short-horizon prediction
historical-neighbor continuation
prediction-specific freshness policy
prediction-specific route-frequency policy
Estimated Arrival
prediction replay evaluation
production fallback behavior
```

Historical source records are not rewritten as observed future facts.

Projection output is always identified as estimated, derived, experimental, or physics-derived evidence.

---

## 3. Implemented Production Architecture

The read path is:

```text
HTTP request
↓
Projection Intelligence handler
↓
Production Projection Read Service
↓
PostgreSQL Data Source
↓
Current FlightTrajectory
↓
Route Intelligence result
↓
Historical candidate trajectories
↓
Route history summary
↓
Production Projection Composer
↓
Projection contract result
↓
Stable HTTP data transfer object
```

The decision path is:

```text
Projection Horizon Policy
↓
Route contract validation
↓
Historical Neighbor Selection
↓
Pattern Confidence
↓
Pattern Freshness Guard
↓
Low-Frequency Route Guard
↓
Historical Neighbor Continuation
↓
Estimated Arrival
```

The conservative fallback path is:

```text
missing or unusable historical evidence
↓
auditable fallback reason
↓
Short-Horizon Kinematic Baseline
↓
Estimated Arrival withheld when route evidence is unavailable
```

---

## 4. Acceptance Matrix

| Capability | Implementation status | Runtime status |
|---|---:|---:|
| Projection contract and validation | Implemented | Unit, race, and static analysis verified |
| Research-only scope guard | Implemented | HTTP and runtime verified |
| Projection horizon policy | Implemented | Unit and production runtime verified |
| Short-horizon kinematic baseline | Implemented | PostgreSQL and HTTP runtime verified |
| Historical Neighbor Selection | Implemented | PostgreSQL and HTTP runtime verified |
| Pattern Confidence | Implemented | PostgreSQL and HTTP runtime verified |
| Pattern Freshness Guard | Implemented | PostgreSQL and HTTP runtime verified |
| Low-Frequency Route Guard | Implemented | PostgreSQL and HTTP runtime verified |
| Local Historical Neighbor Continuation | Implemented | PostgreSQL and HTTP runtime verified |
| Estimated Arrival baseline | Implemented | PostgreSQL and HTTP runtime verified |
| Projection replay evaluation | Implemented | Unit, race, and static analysis verified |
| Production composition | Implemented | Both strategies runtime verified |
| PostgreSQL production read source | Implemented | Runtime verified |
| Read-only HTTP endpoint | Implemented | Runtime verified |
| Production server wiring | Implemented | Route registration and runtime verified |
| Deterministic fingerprints | Implemented | Contract and runtime verified |
| Explicit fallback reason | Implemented | Kinematic fallback runtime verified |
| Runtime fixture cleanup | Implemented | Zero persistent rows verified |
| New database migration | Not required | No schema change |
| Projection persistence | Not implemented | Deferred |
| Frontend projection interface | Not implemented | Deferred |
| Weather-adjusted projection | Not implemented | Stage 10 |
| Airspace interaction projection | Not implemented | Stage 11 |
| Forecast stability analysis | Not implemented | Stage 12 |

---

## 5. Prediction Contract

The prediction-specific contract is implemented in:

```text
apps/api/internal/projectionintelligence/projectioncontract
```

The contract includes:

```text
schema version
result status
trajectory identity
flight identity when available
aircraft identity when available
ICAO 24-bit address
callsign
method name
method version
decision class
projection horizon
forecast points
position
horizontal uncertainty
optional vertical uncertainty
confidence
limitations
explanations
scope guard
provenance
generated-at time
optional Estimated Arrival
```

Supported result statuses are:

```text
unavailable
limited
complete
```

Supported decision classes are:

```text
source_backed
research_adapted
physics_derived
project_derived
experimental
```

The mandatory scope guard is:

```text
research_only_not_for_operational_use
```

The contract rejects structurally invalid projections, invalid confidence, invalid uncertainty, invalid timestamps, invalid fingerprints, missing scope protection, and inconsistent Estimated Arrival output.

---

## 6. Projection Horizon Policy

The horizon policy is implemented in:

```text
apps/api/internal/projectionintelligence/projectionhorizon
```

The production policy is:

```text
minimum duration: 1 minute
default duration: 5 minutes
maximum duration: 15 minutes
step: 30 seconds
maximum forecast point count: 30
```

The runtime historical verification used:

```text
requested duration: 3 minutes
forecast step: 30 seconds
forecast point count: 6
```

The hardened horizon planner enforces:

```text
omitted HTTP duration resolves to the configured default
negative requested duration is rejected
positive requested duration below the configured minimum is rejected
requests above the configured maximum carry explicit truncation evidence
effective duration is exactly divisible by Step
forecast points occupy every fixed-step slot through EndTime
point count is bounded before schedule allocation
all timestamps are canonical UTC
consumer modules reject invalid planner output
the complete plan has a deterministic SHA-256 fingerprint
```

The permanent audit is implemented in:

```text
apps/api/tools/projectionhorizonreviewaudit
```

```text
PROJECTION_HORIZON_POLICY_VERSION=projection-horizon-policy-v2
PROJECTION_HORIZON_FIXED_STEP_GRID=ENFORCED
PROJECTION_HORIZON_PLAN_VALIDATION=ENFORCED
PROJECTION_HORIZON_PLAN_FINGERPRINT=SHA256
PROJECTION_HORIZON_DEFAULT_HTTP_DURATION=SUPPORTED
PROJECTION_HORIZON_POINT_COUNT_BOUNDED=ENFORCED
```

---

## 7. Short-Horizon Kinematic Baseline

The conservative baseline is implemented in:

```text
apps/api/internal/projectionintelligence/projectionbaseline
```

The method identity is:

```text
short_horizon_kinematic_baseline
```

It uses the latest usable observed motion state and propagates a short bounded trajectory.

The baseline includes:

```text
geodesic position propagation
altitude propagation when supported
bounded projection horizon
horizontal uncertainty growth
vertical uncertainty growth
confidence degradation over time
observed-input provenance
deterministic fingerprint
limitations and explanations
```

The baseline is selected when historical continuation cannot be authorized.

It is not represented as a route-aware or weather-aware forecast.

---

## 8. Historical Neighbor Selection

Historical Neighbor Selection is implemented in:

```text
apps/api/internal/projectionintelligence/projectionneighbors
```

The selector:

```text
requires a current trajectory
applies an explicit as-of boundary
requires a positive continuation duration
rejects the current trajectory as its own neighbor
rejects duplicate candidate identifiers
rejects candidates that are not historical
rejects candidates that are too old
rejects insufficient candidate tracks
finds a local spatial anchor
requires enough continuation after the anchor
enforces maximum anchor distance
compares the current prefix with the historical prefix
enforces minimum similarity
orders results deterministically
caps the selected neighbor count
publishes rejection evidence
publishes deterministic fingerprints
```

The production target is:

```text
minimum current point count: 5
maximum candidate count: 50
selection limit: 5
minimum similarity score: 0.60
maximum anchor distance: 100 kilometers
maximum candidate age: 90 days
```

The final PostgreSQL runtime verification selected exactly five historical neighbors and produced a complete selection.

---

## 9. Pattern Confidence

Pattern Confidence is implemented in:

```text
apps/api/internal/projectionintelligence/projectionpatternconfidence
```

The confidence score combines:

```text
mean similarity
neighbor support
candidate freshness
anchor proximity
```

Production weights are:

```text
similarity: 0.45
support: 0.20
freshness: 0.20
anchor proximity: 0.15
```

Production thresholds are:

```text
minimum neighbor count: 2
target neighbor count: 5
minimum usable score: 0.55
medium confidence minimum: 0.60
high confidence minimum: 0.80
```

The result publishes:

```text
status
usable flag
neighbor count
target neighbor count
mean similarity score
mean candidate age
mean anchor distance
component scores
overall score
confidence level
selected trajectory identifiers
limitations
input fingerprint
```

The final runtime evidence produced a complete usable Pattern Confidence result for five selected neighbors.

---

## 10. Pattern Freshness Guard

The prediction-specific freshness guard is implemented in:

```text
apps/api/internal/projectionintelligence/projectionfreshness
```

The guard checks:

```text
newest selected neighbor age
mean selected neighbor age
oldest selected neighbor age
recent-neighbor support
overall freshness score
selection completeness
Pattern Confidence completeness
```

Production limits are:

```text
maximum newest-neighbor age: 30 days
maximum mean-neighbor age: 60 days
maximum oldest-neighbor age: 90 days
recent-neighbor age limit: 30 days
minimum recent-neighbor count: 1
target recent-neighbor count: 3
minimum usable score: 0.45
complete score minimum: 0.70
```

Production policy rejects limited freshness evidence.

The final runtime fixture supplies five recent historical neighbors. The resulting decision is:

```text
allowed
```

This guard is separate from generic source freshness. It protects the use of historical patterns for future continuation.

---

## 11. Low-Frequency Route Guard

The route-frequency guard is implemented in:

```text
apps/api/internal/projectionintelligence/projectionroutefrequency
```

The guard requires:

```text
complete origin and destination
route confidence above the minimum
minimum historical observation count
minimum distinct-day count
recent route observations
latest observation within the maximum age
usable aggregate score
```

Production thresholds are:

```text
minimum observations: 3
target observations: 10
minimum distinct days: 2
target distinct days: 7
recent window: 30 days
minimum recent observations: 1
target recent observations: 4
maximum latest-observation age: 30 days
minimum route confidence: 0.60
minimum usable score: 0.45
complete score minimum: 0.75
```

The final PostgreSQL runtime fixture provides:

```text
5 independent historical route flights
5 distinct historical UTC days
recent route support
complete ZAAA to ZBBB route
high route confidence
```

The resulting decision is:

```text
allowed
```

The guard prevents a rare or weakly supported route from being treated as a strong historical prediction pattern. Route-history evidence excludes the current trajectory and current flight, then deduplicates multiple trajectories of one logical flight before counts, recent support, UTC-day support, and latest-observation age are calculated.

---

## 12. Local Historical Neighbor Continuation

Historical continuation is implemented in:

```text
apps/api/internal/projectionintelligence/projectioncontinuation
```

The method identity is:

```text
local_historical_neighbor_continuation
```

The method:

```text
uses selected historical neighbors
starts from each local anchor
samples observed continuation after the anchor
aligns continuation with the requested horizon
interpolates positions at forecast timestamps
aggregates supported positions
uses neighbor spread in uncertainty
requires minimum point support
requires altitude support before publishing altitude
degrades confidence across the horizon
publishes evidence and limitations
creates a deterministic fingerprint
```

The method is local and bounded. It does not claim a global route model or a persistent trajectory-shape index.

The final PostgreSQL and HTTP runtime verification produced:

```text
strategy: historical_neighbor_continuation
method: local_historical_neighbor_continuation
historical neighbors: 5
forecast points: 6
fallback reason: none
```

---

## 13. Estimated Arrival

Estimated Arrival is implemented in:

```text
apps/api/internal/projectionintelligence/projectionarrival
```

The estimator requires:

```text
a complete destination airport
destination confidence above the minimum
usable projected position
sufficient speed samples
ground speed above the minimum
valid distance to destination
bounded estimated duration
```

Production policy includes:

```text
arrival radius: 10 kilometers
minimum destination confidence: 0.60
minimum speed samples: 3
maximum speed samples: 8
minimum ground speed: 30 meters per second
minimum arrival interval: 2 minutes
maximum estimated duration: 8 hours
```

The result includes:

```text
destination airport ICAO code
earliest arrival time
estimated arrival time
latest arrival time
confidence
limitations
```

Estimated Arrival is attached only when its prerequisites are satisfied.

When Route Intelligence is missing or unusable, the production composition may still produce a kinematic projection but withholds Estimated Arrival.

The final historical runtime verification attached Estimated Arrival to:

```text
ZBBB
```

---

## 14. Projection Replay Evaluation

Projection replay evaluation is implemented in:

```text
apps/api/internal/projectionintelligence/projectionevaluation
```

The evaluator supports:

```text
immutable canonical projection snapshot
future observed truth with explicit point-level system-availability evidence
strict event-time and knowledge-time replay cutoffs
deterministic equal-timestamp conflict rejection
horizontal and vertical error when altitude is supported
explicit endpoint error
point coverage and uncertainty coverage
lead-time error buckets
point micro-average and trajectory macro-average aggregation
bounded confidence-to-normalized-accuracy comparison
arrival prediction recall and airport accuracy
matched Estimated Arrival error and interval coverage
aggregation by method, version, DecisionClass, horizon, step, and evaluation policy
deterministic projection, truth, evaluation, and aggregate fingerprints
bounded physical interpolation configuration
derived-metric recomputation before aggregation
```

The replay knowledge boundary is not inferred from `ObservedAt` alone. Every usable
truth point supplies an explicit `AvailableAt` value, and actual arrival evidence also
proves availability by `EvaluatedAt`.

This is the foundation for later model comparison and calibration. Confidence
comparison remains a bounded engineering diagnostic and is not represented as
scientific probability calibration.

Stage 9 does not claim that a large real-world benchmark dataset has already been evaluated.

The evaluation engine is implemented and tested, while broader empirical calibration remains future work.

---

## 15. Production Composition

Production composition is implemented in:

```text
apps/api/internal/projectionintelligence/projectionproduction
```

The production composer coordinates:

```text
horizon planning
route validation
historical neighbor selection
Pattern Confidence
Pattern Freshness Guard
Low-Frequency Route Guard
historical continuation
kinematic fallback
Estimated Arrival attachment
result validation
fallback notices
production fingerprint
```

Supported strategies are:

```text
historical_neighbor_continuation
kinematic_baseline
```

The historical strategy requires complete usable evidence from:

```text
Neighbor Selection
Pattern Confidence
Pattern Freshness
Route Frequency
```

The kinematic strategy requires:

```text
explicit fallback reason
kinematic method identity
at least one auditable notice
```

The production result cannot silently switch strategies.

---

## 16. PostgreSQL Production Read Service

The PostgreSQL composition is implemented in:

```text
apps/api/internal/projectionintelligence/projectionread
```

It loads:

```text
current trajectory metadata
observed flight-state points up to the as-of boundary
latest Route Intelligence result at or before the as-of time
route-scoped historical candidate identifiers
complete historical candidate trajectories
route history summary
```

Production source controls include:

```text
maximum trajectory points: 10000
maximum historical candidates: 50
historical candidate lookback: 90 days
route history window: 180 days
recent route window: 30 days
```

The source excludes:

```text
future observations
the current trajectory from candidate history
candidate trajectories ending after the current trajectory starts
candidates exceeding configured limits
missing or invalid source records
```

Stage 9 adds no new database table and no migration.

Projection results are computed through the read service and are not persisted as a new prediction record in this stage.

---

## 17. Read-Only HTTP API

The read-only endpoint is:

```text
GET /api/v1/trajectories/{trajectory_id}/projection-intelligence
```

Required query parameters are:

```text
as_of_time
duration_seconds
```

Example request shape:

```text
GET /api/v1/trajectories/{uuid}/projection-intelligence
    ?as_of_time={RFC3339 timestamp}
    &duration_seconds={positive integer}
```

The HTTP layer provides:

```text
trajectory UUID validation
RFC 3339 as-of validation
positive duration validation
stable snake_case JSON
production result validation before response
not-found mapping
validation-error mapping
service-unavailable mapping
database-error mapping
request-timeout mapping
```

The response includes:

```text
production version
selected strategy
fallback reason when applicable
arrival status
projection contract
historical evidence
notices
production fingerprint
generated-at time
```

No public Projection Intelligence write endpoint is part of Stage 9.

---

## 18. Runtime Verification

Two independent production runtime verifiers are included.

### 18.1 Kinematic fallback verifier

Command:

```text
go run ./cmd/verify-postgres-projection-intelligence-http-api
```

Verified:

```text
PostgreSQL schema objects
deterministic trajectory fixture
production PostgreSQL reader
observed flight-state hydration
as-of boundary
missing Route Intelligence fallback
kinematic projection endpoint
projection uncertainty
projection confidence
not-found contract
validation-error contract
JSON response contract
fixture cleanup
zero persistent rows
```

Final result:

```text
Persistent verification rows: 0
Result: PASS
```

### 18.2 Historical continuation verifier

Command:

```text
go run ./cmd/verify-postgres-projection-intelligence-historical-http-api
```

Verified:

```text
PostgreSQL schema objects
deterministic six-flight fixture
production route-record identifier contract
production policy coverage
direct production service result
Route Intelligence history loading
historical candidate loading
complete five-neighbor selection
complete Pattern Confidence
allowed Pattern Freshness
allowed Low-Frequency Route Guard
Local Historical Neighbor Continuation
Estimated Arrival attachment
read-only HTTP contract
fixture cleanup
zero persistent rows
```

Final runtime evidence:

```text
Production composition: projection-production-composition-v2
Command timeout: 5m0s
Direct service duration: 2.865s
HTTP verification duration: 2.057s
Projection method: local_historical_neighbor_continuation
Direct strategy: historical_neighbor_continuation
Required historical neighbors: 5
Historical neighbors: 5
Forecast points: 6
Arrival airport: ZBBB
Persistent verification rows: 0
Result: PASS
```

---

## 19. Runtime Defects Found and Corrected Before Closure

Stage closure was not declared after compilation alone.

The PostgreSQL verifier exposed three runtime defects in the verification fixture and harness.

### 19.1 Route-record identifier contract

Initial fixture records used UUID identifiers for `flight_route_results`.

The production Route Store requires:

```text
route-record- + SHA-256(
    trajectory identifier
    + schema version
    + as-of time
    + input fingerprint
)
```

The verifier now uses the exact production identifier contract and no UUID cast for the record identifier.

### 19.2 Fiber test timeout

The initial historical HTTP verification used Fiber's default one-second test timeout.

The final verifier uses:

```text
production read timeout: 60 seconds
Fiber HTTP test timeout: 65 seconds
command timeout floor: 5 minutes
cleanup timeout: 60 seconds
```

Timeouts remain finite and explicit.

### 19.3 Incomplete historical evidence

The initial fixture supplied four historical candidates while production policy requires a selection target of five.

The final fixture supplies:

```text
5 historical candidates
5 selected neighbors
5 recent neighbors
5 independent historical route flights
5 distinct historical UTC days
```

The verifier now proves that the historical path satisfies current production policy rather than only a simplified test policy.

---

## 20. Determinism and Provenance

Deterministic fingerprints are implemented for:

```text
projection contract input
horizon plan
kinematic baseline
historical neighbor selection
Pattern Confidence
Pattern Freshness
Route Frequency
historical continuation
Estimated Arrival
replay evaluation
production composition
```

The Route Intelligence fixture uses production-compatible deterministic route-record identifiers.

Inputs are normalized before fingerprints are produced.

The horizon plan fingerprint is owned by `projectionhorizon` and covers the
policy version and name, canonical UTC times, fixed step, requested and
effective durations, truncation evidence, and the complete canonical forecast
time grid. Kinematic and historical continuation fingerprints consume this
single horizon fingerprint instead of rebuilding a partial horizon identity.

Published provenance identifies:

```text
input name
input classification
source name
observed time
retrieved time
latest observed input
input fingerprint
```

This makes projection output reproducible for the same normalized input and policy.

---

## 21. Failure and Fallback Behavior

The production composer records why historical continuation was not selected.

Representative fallback reasons include:

```text
route_contract_invalid
historical_neighbors_unavailable
historical_pattern_not_usable
pattern_freshness_guard_blocked
complete_route_unavailable
route_history_unavailable
route_frequency_guard_blocked
historical_projection_failed
```

A fallback is not hidden.

When fallback succeeds:

```text
strategy = kinematic_baseline
method = short_horizon_kinematic_baseline
fallback_reason is non-empty
auditable notices are present
```

Estimated Arrival may be:

```text
attached
withheld
failed
skipped
```

A non-attached status must not contain an arrival estimate.

---

## 22. Safety and Product Boundary

Projection Intelligence is research software.

It must not be used for:

```text
air traffic control
aircraft separation
navigation
dispatch
flight planning
safety decisions
regulatory compliance
passenger-facing guaranteed arrival time
```

The system has no authoritative surveillance feed and no operational flight-plan feed.

Open-data gaps, delayed observations, inferred routes, missing interventions, weather absence, and aircraft intent remain material limitations.

The explicit scope guard must remain visible through every output layer.

---

## 23. Known Limitations

The completed Stage 9 foundation has the following known limitations:

```text
historical continuation is heuristic
historical continuation is local to selected anchors
there is no persistent trajectory-shape index
there is no machine-learning model
there is no model registry
there is no trained route-specific model
there is no large held-out real-world benchmark
there is no weather adjustment
there is no wind correction
there is no airspace restriction context
there is no controller-intervention model
there is no flight-plan intent source
there is no persistent projection store
there is no frontend projection visualization
```

The deterministic PostgreSQL runtime evidence uses synthetic fixtures designed to exercise production data paths.

Synthetic runtime evidence proves system integration and policy behavior. It does not prove global prediction accuracy.

---

## 24. Deferred Work

The next implementation stages remain:

```text
Stage 10 — Weather Context
Stage 11 — Airspace Intelligence
Stage 12 — Stability and Explainability
```

Important future work includes:

```text
weather observation trust gate
wind and weather alignment
weather-adjusted uncertainty
airspace and interaction context
forecast versioning
forecast stability
confidence calibration
large-scale replay benchmark
compound empirical evaluation
frontend projection visualization
optional bounded projection persistence
```

These items are not required to close the Stage 9 foundation.

---

## 25. Commit Evidence

The Stage 9 implementation chain is:

```text
5290de1 — projection contract and horizon foundation
9fe7b50 — short-horizon kinematic baseline
ae42ae5 — historical neighbor selection
c1eccd6 — local historical continuation
14f2f32 — Estimated Arrival baseline
88e1e08 — projection replay evaluation
f8b0483 — prediction freshness and route-frequency guards
de35bce — production composition
fc26e3e — read-only Projection Intelligence HTTP API
fc8f5ac — PostgreSQL production source and server wiring
32e0a8e — kinematic fallback HTTP runtime verifier
bdfa325 — historical continuation runtime verifier
f978074 — production route-record identifier alignment
c2927c7 — complete historical production runtime verification
```

The final evidence baseline is:

```text
c2927c7f6411c33f702244f4c9277cd91e44b753
```

---

## 26. Verification Commands

Package and integration verification:

```text
go test ./...
go vet ./...
```

Kinematic fallback runtime verification:

```text
go run ./cmd/verify-postgres-projection-intelligence-http-api
```

Historical continuation runtime verification:

```text
go run ./cmd/verify-postgres-projection-intelligence-historical-http-api
```

The final runtime verifier cleans all deterministic fixture rows and requires:

```text
Persistent verification rows: 0
Result: PASS
```

---

## 27. Completion Statement

Stage 9 is complete as the Production Projection Intelligence Foundation.

The completed system can:

```text
build a bounded short-horizon projection contract
produce a conservative kinematic projection
read current and historical evidence from PostgreSQL
select complete historical neighbor support
evaluate pattern confidence
enforce prediction freshness
enforce route-frequency support
produce local historical continuation
attach Estimated Arrival when evidence permits
fall back deterministically when evidence does not permit
publish confidence, uncertainty, limitations, explanations, and provenance
serve the result through a read-only HTTP endpoint
verify both production strategies against PostgreSQL
clean every runtime fixture without persistent residue
```

The project may now proceed to Stage 10 — Weather Context.

This completion does not convert research estimates into operational aviation claims.

---

## 28. Projection Baseline Review Hardening

The conservative kinematic fallback has completed a dedicated static-review
hardening cycle. The review closes historical cutoff leakage, cutoff-safe quality
recomputation, PostgreSQL hydration alignment, observation-age confidence,
physical kinematic bounds, altitude-reference provenance, collaborator validation,
explicit horizontal fallback, deterministic latest-observation handling, and a
stationary limited on-ground model.

Permanent review enforcement is implemented in:

```text
apps/api/tools/projectionbaselinereviewaudit
```

The authoritative review record is:

```text
docs/136_PROJECTION_BASELINE_REVIEW_HARDENING.md
```

```text
PROJECTION_BASELINE_METHOD_VERSION=short-horizon-kinematic-baseline-v3
PROJECTION_BASELINE_INPUT_FINGERPRINT_VERSION=projection-baseline-input-fingerprint-v4
PROJECTION_BASELINE_CUTOFF_ISOLATION=ENFORCED
PROJECTION_BASELINE_QUALITY_RECOMPUTATION=ENFORCED
PROJECTION_BASELINE_POSTGRES_CUTOFF_ALIGNMENT=ENFORCED
PROJECTION_BASELINE_UNAVAILABLE_PROVENANCE=ENFORCED
PROJECTION_BASELINE_OBSERVATION_AGE_CONFIDENCE=ENFORCED
PROJECTION_BASELINE_PHYSICAL_BOUNDS=ENFORCED
PROJECTION_BASELINE_ALTITUDE_REFERENCE=EXPLICIT
PROJECTION_BASELINE_LATEST_OBSERVATION_AMBIGUITY=REJECTED
PROJECTION_BASELINE_ON_GROUND_MODEL=STATIONARY_LIMITED
PROJECTION_BASELINE_HORIZONTAL_FALLBACK=EXPLICIT_POLICY
PROJECTION_BASELINE_ELIGIBILITY_OUTPUT_VALIDATION=ENFORCED
PROJECTION_BASELINE_ELIGIBILITY_POLICY_PROVENANCE=ENFORCED
PROJECTION_BASELINE_PERMANENT_AUDIT_COMMIT=51476c427f77b5a7375cd30b6f9a81d446c1c3f2
PROJECTION_BASELINE_PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30408617024
PROJECTION_BASELINE_REVIEW_STATUS=CLOSED
```

---

## 29. Projection Neighbors Review Hardening

Historical-neighbor selection has completed its engineering hardening cycle.
The review closes pre-budget candidate integrity, whole-input duplicate handling,
canonical equal-timestamp ordering, similarity subsystem failure classification,
continuous continuation evidence, source-attested route scope, selector pipeline
decomposition, and explicit limiting semantics.

Permanent review enforcement is implemented in:

```text
apps/api/tools/projectionneighborsreviewaudit
```

The authoritative review record is:

```text
docs/137_PROJECTION_NEIGHBORS_REVIEW_HARDENING.md
```

```text
PROJECTION_NEIGHBORS_SELECTION_VERSION=projection-historical-neighbor-selection-v5
PROJECTION_NEIGHBORS_FINGERPRINT_VERSION=projection-historical-neighbor-selection-fingerprint-v5
PROJECTION_NEIGHBORS_CANDIDATE_ELIGIBILITY_BEFORE_BUDGET=ENFORCED
PROJECTION_NEIGHBORS_WHOLE_INPUT_DUPLICATE_INTEGRITY=ENFORCED
PROJECTION_NEIGHBORS_CANONICAL_POINT_ORDERING=ENFORCED
PROJECTION_NEIGHBORS_SIMILARITY_SYSTEM_FAILURES=PROPAGATED
PROJECTION_NEIGHBORS_CONTINUATION_GAP_POLICY=ENFORCED
PROJECTION_NEIGHBORS_ROUTE_SCOPE=SOURCE_ATTESTED
PROJECTION_NEIGHBORS_CROSS_ROUTE_REJECTION=BEFORE_SIMILARITY
PROJECTION_NEIGHBORS_SELECTOR_PIPELINE=DECOMPOSED
PROJECTION_NEIGHBORS_CANDIDATE_EVALUATION_TRUNCATION=EXPLICIT
PROJECTION_NEIGHBORS_QUALIFIED_SELECTION_LIMITING=EXPLICIT
PROJECTION_NEIGHBORS_PERMANENT_AUDIT_COMMIT=c409cc171507050625524af1a0b8b8a6f38b7a75
PROJECTION_NEIGHBORS_PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30452465009
PROJECTION_NEIGHBORS_PERMANENT_AUDIT_BACKEND_QUALITY_JOB=90577613283
PROJECTION_NEIGHBORS_PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=90577613277
PROJECTION_NEIGHBORS_PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=90577613384
PROJECTION_NEIGHBORS_PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=90577905997
PROJECTION_NEIGHBORS_REVIEW_STATUS=CLOSED
```

---

## 30. Projection Pattern Confidence Review Hardening

Projection Pattern Confidence has completed its engineering hardening cycle. The
review closes semantic evidence fingerprinting, strict configuration contracts,
similarity distribution evidence, freshness separation, continuation agreement,
mandatory continuation-aware production interfaces, and complete cross-field result
validation.

Permanent review enforcement is implemented in:

```text
apps/api/tools/projectionpatternconfidencereviewaudit
```

The authoritative review record is:

```text
docs/138_PROJECTION_PATTERN_CONFIDENCE_REVIEW_HARDENING.md
```

```text
PROJECTION_PATTERN_CONFIDENCE_VERSION=projection-pattern-confidence-v4
PROJECTION_PATTERN_CONFIDENCE_FINGERPRINT_VERSION=projection-pattern-confidence-fingerprint-v4
PROJECTION_PATTERN_CONFIDENCE_CANONICAL_COMPONENT_COUNT=5
PROJECTION_PATTERN_CONFIDENCE_SEMANTIC_FINGERPRINT=ENFORCED
PROJECTION_PATTERN_CONFIDENCE_POSITIVE_WEIGHTS=ENFORCED
PROJECTION_PATTERN_CONFIDENCE_FRESHNESS_SEPARATION=ENFORCED
PROJECTION_PATTERN_CONFIDENCE_SIMILARITY_DISTRIBUTION=ENFORCED
PROJECTION_PATTERN_CONFIDENCE_CONTINUATION_AGREEMENT=ENFORCED
PROJECTION_PATTERN_CONFIDENCE_MAXIMUM_DIVERGENCE_GUARD=ENFORCED
PROJECTION_PATTERN_CONFIDENCE_RESULT_RECONSTRUCTION=ENFORCED
PROJECTION_PATTERN_CONFIDENCE_MANDATORY_CONTINUATION_INTERFACE=ENFORCED
PROJECTION_PATTERN_CONFIDENCE_LEGACY_PRODUCTION_FALLBACK=ABSENT
PROJECTION_PATTERN_CONFIDENCE_PERMANENT_AUDIT_COMMIT=cd8f114bfef698c51cfc6008ecd2ed01f9c1cc42
PROJECTION_PATTERN_CONFIDENCE_PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30497703314
PROJECTION_PATTERN_CONFIDENCE_PERMANENT_AUDIT_BACKEND_QUALITY_JOB=90730254967
PROJECTION_PATTERN_CONFIDENCE_PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=90730255221
PROJECTION_PATTERN_CONFIDENCE_PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=90730255044
PROJECTION_PATTERN_CONFIDENCE_PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=90730452053
PROJECTION_PATTERN_CONFIDENCE_REVIEW_STATUS=CLOSED
```

---

## 31. Projection Freshness Review Hardening

Projection Freshness has completed its engineering hardening and corrective
contract-consistency cycle. The review closes Pattern Confidence usability
enforcement, exact selection lineage, timestamp-derived age evidence, overflow-safe
duration averaging, ordered positive configuration thresholds, strictly positive
component weights, complete hard-violation reporting, policy and upstream state
snapshots, result reconstruction, and production fixture contract drift.

Permanent review enforcement is implemented in:

```text
apps/api/tools/projectionfreshnessreviewaudit
```

The authoritative review record is:

```text
docs/139_PROJECTION_FRESHNESS_REVIEW_HARDENING.md
```

```text
PROJECTION_FRESHNESS_VERSION=projection-pattern-freshness-guard-v3
PROJECTION_FRESHNESS_FINGERPRINT_VERSION=projection-pattern-freshness-fingerprint-v3
PROJECTION_FRESHNESS_CANONICAL_COMPONENT_COUNT=4
PROJECTION_FRESHNESS_PATTERN_USABILITY_GUARD=ENFORCED
PROJECTION_FRESHNESS_EXACT_SELECTION_LINEAGE=ENFORCED
PROJECTION_FRESHNESS_TIMESTAMP_DERIVED_AGE=ENFORCED
PROJECTION_FRESHNESS_OVERFLOW_SAFE_MEAN=ENFORCED
PROJECTION_FRESHNESS_ORDERED_AGE_THRESHOLDS=ENFORCED
PROJECTION_FRESHNESS_POSITIVE_SCORE_THRESHOLDS=ENFORCED
PROJECTION_FRESHNESS_POSITIVE_COMPONENT_WEIGHTS=ENFORCED
PROJECTION_FRESHNESS_ALL_HARD_VIOLATIONS=REPORTED
PROJECTION_FRESHNESS_POLICY_SNAPSHOT=ENFORCED
PROJECTION_FRESHNESS_UPSTREAM_STATE_SNAPSHOT=ENFORCED
PROJECTION_FRESHNESS_COMPONENT_RECONSTRUCTION=ENFORCED
PROJECTION_FRESHNESS_DECISION_RECONSTRUCTION=ENFORCED
PROJECTION_FRESHNESS_PRODUCTION_FIXTURE_DRIFT=PREVENTED
PROJECTION_FRESHNESS_PERMANENT_AUDIT_COMMIT=619e24878a5025decf6fe21abddba537ce195560
PROJECTION_FRESHNESS_PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30523502590
PROJECTION_FRESHNESS_PERMANENT_AUDIT_BACKEND_QUALITY_JOB=90809046060
PROJECTION_FRESHNESS_PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=90809046046
PROJECTION_FRESHNESS_PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=90809046013
PROJECTION_FRESHNESS_PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=90809225151
PROJECTION_FRESHNESS_WEIGHT_POLICY_CORRECTION_COMMIT=e3e99758d6f654db12ccce32ec55ad1339fb518f
PROJECTION_FRESHNESS_WEIGHT_POLICY_CORRECTION_GITHUB_ACTIONS_RUN=30527541240
PROJECTION_FRESHNESS_WEIGHT_POLICY_CORRECTION_BACKEND_QUALITY_JOB=90821894564
PROJECTION_FRESHNESS_WEIGHT_POLICY_CORRECTION_BACKEND_RACE_SAFETY_JOB=90821894465
PROJECTION_FRESHNESS_WEIGHT_POLICY_CORRECTION_POSTGRESQL_16_INTEGRATION_JOB=90821894536
PROJECTION_FRESHNESS_WEIGHT_POLICY_CORRECTION_BACKEND_CONTAINER_JOB=90822090361
PROJECTION_FRESHNESS_WEIGHT_POLICY_CONSISTENCY=CLOSED
PROJECTION_FRESHNESS_REVIEW_STATUS=CLOSED
```
---

## 32. Projection Route Frequency Review Hardening

Projection Route Frequency has completed the confirmed production-code hardening for
historical evidence isolation and policy integrity. The module now excludes the
current prediction target, deduplicates historical evidence by logical flight,
enforces fixed full and recent exposure windows, rejects incoherent policy targets,
reports all simultaneous hard violations, reconstructs weighted scores, and protects
decision-relevant fingerprint identity.

Permanent review enforcement is implemented in:

```text
apps/api/tools/projectionroutefrequencyreviewaudit
```

The authoritative review record is:

```text
docs/140_PROJECTION_ROUTE_FREQUENCY_REVIEW_HARDENING.md
```

```text
PROJECTION_ROUTE_FREQUENCY_VERSION=projection-low-frequency-route-guard-v3
PROJECTION_ROUTE_FREQUENCY_FINGERPRINT_VERSION=projection-low-frequency-route-fingerprint-v3
PROJECTION_ROUTE_FREQUENCY_EVIDENCE_ISOLATION=ENFORCED
PROJECTION_ROUTE_FREQUENCY_LOGICAL_FLIGHT_DEDUPLICATION=ENFORCED
PROJECTION_ROUTE_FREQUENCY_DISTINCT_FLIGHT_SCORING=ENFORCED
PROJECTION_ROUTE_FREQUENCY_FIXED_HISTORY_WINDOW=ENFORCED
PROJECTION_ROUTE_FREQUENCY_FIXED_RECENT_WINDOW=ENFORCED
PROJECTION_ROUTE_FREQUENCY_POLICY_TARGET_RELATIONSHIPS=ENFORCED
PROJECTION_ROUTE_FREQUENCY_POSITIVE_SCORE_THRESHOLDS=ENFORCED
PROJECTION_ROUTE_FREQUENCY_POSITIVE_COMPONENT_WEIGHTS=ENFORCED
PROJECTION_ROUTE_FREQUENCY_ALL_HARD_VIOLATIONS=REPORTED
PROJECTION_ROUTE_FREQUENCY_WEIGHTED_SCORE_RECONSTRUCTION=ENFORCED
PROJECTION_ROUTE_FREQUENCY_DETERMINISTIC_LIMITATIONS=ENFORCED
PROJECTION_ROUTE_FREQUENCY_SEMANTIC_FINGERPRINT=ENFORCED
PROJECTION_ROUTE_FREQUENCY_PRODUCTION_FIXTURE_SCORE=0.94
PROJECTION_ROUTE_FREQUENCY_EVIDENCE_ISOLATION_COMMIT=c6fff15f8d0c770197db40a69d54f8856044d8d2
PROJECTION_ROUTE_FREQUENCY_EVIDENCE_ISOLATION_GITHUB_ACTIONS_RUN=30534243693
PROJECTION_ROUTE_FREQUENCY_POLICY_DECISION_INTEGRITY_COMMIT=ee7c79bc8213dc030ce0d98f13d1065c9bb96275
PROJECTION_ROUTE_FREQUENCY_POLICY_DECISION_INTEGRITY_GITHUB_ACTIONS_RUN=30544636679
PROJECTION_ROUTE_FREQUENCY_POLICY_DECISION_INTEGRITY_POSTGRESQL_16_INTEGRATION_JOB=90877578926
PROJECTION_ROUTE_FREQUENCY_POLICY_DECISION_INTEGRITY_BACKEND_RACE_SAFETY_JOB=90877578928
PROJECTION_ROUTE_FREQUENCY_POLICY_DECISION_INTEGRITY_BACKEND_QUALITY_JOB=90877579007
PROJECTION_ROUTE_FREQUENCY_POLICY_DECISION_INTEGRITY_BACKEND_CONTAINER_JOB=90877915808
PROJECTION_ROUTE_FREQUENCY_PERMANENT_AUDIT_COMMIT=6f039b33c96cdb67370158b0eda5d0fc87593de5
PROJECTION_ROUTE_FREQUENCY_PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30548438062
PROJECTION_ROUTE_FREQUENCY_PERMANENT_AUDIT_BACKEND_QUALITY_JOB=90890525039
PROJECTION_ROUTE_FREQUENCY_PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=90890525126
PROJECTION_ROUTE_FREQUENCY_PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=90890525150
PROJECTION_ROUTE_FREQUENCY_PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=90890829745
PROJECTION_ROUTE_FREQUENCY_FORMAL_CLOSURE=COMPLETE
PROJECTION_ROUTE_FREQUENCY_REVIEW_STATUS=CLOSED
```

The review is formally closed with exact Continuous Integration recorded for the
policy-and-decision-integrity commit and the permanent-audit commit. The closure commit
must itself pass the same four Backend Continuous Integration jobs before the external
formal-module-closure verdict is declared.

## 33. Projection Continuation Review Hardening

Projection Continuation review hardening began at baseline
`a9b72001f1358af06a06f3a16212850daceef553`. Approved Evidence Integrity removed
production authorization/projection identity drift and bound Pattern, Selection,
anchors, candidates and source provenance to the consumed evidence chain.

```text
PROJECTION_CONTINUATION_REVIEW_BASELINE=a9b72001f1358af06a06f3a16212850daceef553
PROJECTION_CONTINUATION_APPROVED_EVIDENCE_CONTRACT=ENFORCED
PROJECTION_CONTINUATION_DOUBLE_EVALUATION_IN_PRODUCTION=REMOVED
PROJECTION_CONTINUATION_PATTERN_SELECTION_FINGERPRINT=ENFORCED
PROJECTION_CONTINUATION_ANCHOR_CANDIDATE_BINDING=ENFORCED
PROJECTION_CONTINUATION_OBSERVED_SOURCE_PROVENANCE=PRESERVED
PROJECTION_CONTINUATION_APPROVED_EVIDENCE_COMMIT=23ecf72a0700b5a96459bc4a8618c72951a4e6aa
PROJECTION_CONTINUATION_APPROVED_EVIDENCE_GITHUB_ACTIONS_RUN=30573655172
PROJECTION_CONTINUATION_REVIEW_STATUS=CLOSED
```

The authoritative review record is:

```text
docs/141_PROJECTION_CONTINUATION_REVIEW_HARDENING.md
```

## 34. Projection Continuation Interpolation Plausibility

The second hardening increment added a centrally normalized and fingerprinted
plausibility policy and passed exact Backend Continuous Integration.

```text
PROJECTION_CONTINUATION_PLAUSIBILITY_BASELINE=23ecf72a0700b5a96459bc4a8618c72951a4e6aa
MAXIMUM_INTERPOLATION_GAP=5m
MAXIMUM_HORIZONTAL_SPEED_MPS=400
MAXIMUM_VERTICAL_SPEED_MPS=100
EXACT_ENDPOINT_SEGMENT_VALIDATION=ENFORCED
PLAUSIBILITY_FILTERED_OUTPUT_STATUS=LIMITED
PLAUSIBILITY_SUPPORT_FALLBACK=ENFORCED
INTERPOLATION_PLAUSIBILITY=CI_CONFIRMED_COMMIT_739073de31e4c1da2aa105d495bc789a294cb3c9_RUN_30576928637
PROJECTION_CONTINUATION_REVIEW_STATUS=CLOSED
```

## 35. Projection Continuation Engineering and Formal Review Closure

The remaining confirmed engineering findings were closed in commit `13838c4273a3be6bde63835e1d8f51af6f6daa21`.
Configured model uncertainty and neighbor disagreement are conservatively added,
effective weighted support replaces raw sample count, disagreement directly reduces
confidence, terminal confidence loss cannot reach one, zero-confidence points force
limited status, near-antipodal means are rejected, standalone candidate evidence is
bound, and fallback causes remain inspectable through `errors.Is`.

Permanent review enforcement is implemented in:

```text
apps/api/tools/projectioncontinuationreviewaudit
```

```text
PROJECTION_CONTINUATION_VERSION=local-historical-neighbor-continuation-v3
PROJECTION_CONTINUATION_FINGERPRINT_VERSION=local-historical-neighbor-continuation-fingerprint-v3
PROJECTION_CONTINUATION_FALLBACK_FINGERPRINT_VERSION=local-historical-neighbor-fallback-fingerprint-v3
PROJECTION_CONTINUATION_APPROVED_EVIDENCE_INTEGRITY=CI_CONFIRMED
PROJECTION_CONTINUATION_INTERPOLATION_PLAUSIBILITY=CI_CONFIRMED
PROJECTION_CONTINUATION_ADDITIVE_UNCERTAINTY_COMPOSITION=ENFORCED
PROJECTION_CONTINUATION_EFFECTIVE_WEIGHTED_SUPPORT=ENFORCED
PROJECTION_CONTINUATION_DISAGREEMENT_CONFIDENCE_PENALTY=ENFORCED
PROJECTION_CONTINUATION_MAXIMUM_CONFIDENCE_LOSS_BELOW_ONE=ENFORCED
PROJECTION_CONTINUATION_ZERO_CONFIDENCE_STATUS=LIMITED
PROJECTION_CONTINUATION_NEAR_ANTIPODAL_GUARD=ENFORCED
PROJECTION_CONTINUATION_STANDALONE_CANDIDATE_BINDING=ENFORCED
PROJECTION_CONTINUATION_FALLBACK_CAUSE_PRESERVATION=ENFORCED
PROJECTION_CONTINUATION_TRUNCATED_PLAN_FINGERPRINT_REGRESSION=ENFORCED
PROJECTION_CONTINUATION_IRREGULAR_FORECAST_GRID_REJECTION=ENFORCED
PROJECTION_CONTINUATION_EQUAL_TIMESTAMP_CANONICAL_ORDER=ENFORCED
PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_COMMIT=13838c4273a3be6bde63835e1d8f51af6f6daa21
PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN=30593549087
PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB=91040848886
PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_BACKEND_RACE_SAFETY_JOB=91040848927
PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_BACKEND_QUALITY_JOB=91040848967
PROJECTION_CONTINUATION_ENGINEERING_CLOSURE_BACKEND_CONTAINER_JOB=91041042383
PROJECTION_CONTINUATION_PERMANENT_REVIEW_AUDIT=CI_CONFIRMED
PROJECTION_CONTINUATION_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_CONTINUATION_ENGINEERING_DEBT=CLOSED
PROJECTION_CONTINUATION_OPEN_CONFIRMED_FINDINGS=0
PROJECTION_CONTINUATION_UNCLASSIFIED_FINDINGS=0
PROJECTION_CONTINUATION_DEFERRED_FINDINGS=0
PROJECTION_CONTINUATION_ADDITIONAL_CODE_FIXES_REQUIRED=NO
PROJECTION_CONTINUATION_FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_CONTINUATION_FORMAL_CLOSURE=COMPLETE
PROJECTION_CONTINUATION_REVIEW_STATUS=CLOSED
```

The formal-closure commit must pass the same four Backend Continuous Integration jobs
before the external final closure report is issued.

## 36. Projection Arrival Engineering and Formal Review Closure

Projection Arrival engineering review hardening was completed in commit
`65311c066aebbc278b63e2d25558f79f57584ca3` and passed exact Backend Continuous
Integration run `30614617800`. The module separates physical ground speed from signed
destination closing speed, preserves slow and receding samples, enforces a maximum
physical segment speed, uses radial closing speed for radius-entry uncertainty, bounds
the complete extrapolated interval, and fingerprints the exact canonical position
evidence consumed by the calculation.

Permanent review enforcement is implemented in:

```text
apps/api/tools/projectionarrivalreviewaudit
```

The authoritative review record is:

```text
docs/142_PROJECTION_ARRIVAL_REVIEW_HARDENING.md
```

```text
PROJECTION_ARRIVAL_VERSION=estimated-arrival-boundary-v2
PROJECTION_ARRIVAL_FINGERPRINT_VERSION=estimated-arrival-boundary-fingerprint-v2
PROJECTION_ARRIVAL_UNAVAILABLE_FINGERPRINT_VERSION=estimated-arrival-unavailable-fingerprint-v2
PROJECTION_ARRIVAL_POSITION_SAMPLE_FINGERPRINT_VERSION=estimated-arrival-position-samples-v1
PROJECTION_ARRIVAL_DIRECTIONAL_CLOSING_SPEED=ENFORCED
PROJECTION_ARRIVAL_MAXIMUM_GROUND_SPEED_MPS=400
PROJECTION_ARRIVAL_LOW_SPEED_SAMPLE_PRESERVATION=ENFORCED
PROJECTION_ARRIVAL_RADIAL_RADIUS_ENTRY_UNCERTAINTY=ENFORCED
PROJECTION_ARRIVAL_COMPLETE_INTERVAL_BOUND=ENFORCED
PROJECTION_ARRIVAL_DURATION_ROUNDING=CEILING_TO_NANOSECOND
PROJECTION_ARRIVAL_DURATION_OVERFLOW_GUARD=ENFORCED
PROJECTION_ARRIVAL_CURRENT_TRAJECTORY_IDENTITY=ENFORCED
PROJECTION_ARRIVAL_CURRENT_ENDPOINT_PROVENANCE=ENFORCED
PROJECTION_ARRIVAL_POSITION_SAMPLE_LINEAGE=ENFORCED
PROJECTION_ARRIVAL_CONFIDENCE_REASON_RECONSTRUCTION=ENFORCED
PROJECTION_ARRIVAL_DURATION_POLICY_COHERENCE=ENFORCED
PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_COMMIT=65311c066aebbc278b63e2d25558f79f57584ca3
PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN=30614617800
PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB=91104833141
PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_BACKEND_RACE_SAFETY_JOB=91104833127
PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_BACKEND_QUALITY_JOB=91104833181
PROJECTION_ARRIVAL_ENGINEERING_CLOSURE_BACKEND_CONTAINER_JOB=91105067522
PROJECTION_ARRIVAL_PERMANENT_REVIEW_AUDIT=CI_CONFIRMED
PROJECTION_ARRIVAL_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_ARRIVAL_ENGINEERING_DEBT=CLOSED
PROJECTION_ARRIVAL_OPEN_CONFIRMED_FINDINGS=0
PROJECTION_ARRIVAL_UNCLASSIFIED_FINDINGS=0
PROJECTION_ARRIVAL_DEFERRED_FINDINGS=0
PROJECTION_ARRIVAL_ADDITIONAL_CODE_FIXES_REQUIRED=NO
PROJECTION_ARRIVAL_FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_ARRIVAL_FORMAL_CLOSURE=COMPLETE
PROJECTION_ARRIVAL_REVIEW_STATUS=CLOSED
```


The formal-closure commit must pass the same four Backend Continuous Integration jobs
before the external final closure report is issued.

## 37. Projection Evaluation Engineering and Formal Review Closure

Projection Evaluation engineering review hardening is implemented against baseline
`61e1696b16e39f49a3850530312555c3593acfc5`. The evaluator now binds the exact
projection output, canonical truth snapshot, availability evidence, altitude statuses,
actual arrival evidence, and immutable evaluation policy.

Permanent review enforcement is implemented in:

```text
apps/api/tools/projectionevaluationreviewaudit
```

The authoritative review record is:

```text
docs/143_PROJECTION_EVALUATION_REVIEW_HARDENING.md
```

```text
PROJECTION_EVALUATION_ENGINEERING_CLOSURE_COMMIT=279d60543bbbb8c204fab60e442f00a56d1f3bbe
PROJECTION_EVALUATION_ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN=30619973772
PROJECTION_EVALUATION_ENGINEERING_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB=91121986195
PROJECTION_EVALUATION_ENGINEERING_CLOSURE_BACKEND_RACE_SAFETY_JOB=91121986123
PROJECTION_EVALUATION_ENGINEERING_CLOSURE_BACKEND_QUALITY_JOB=91121986134
PROJECTION_EVALUATION_ENGINEERING_CLOSURE_BACKEND_CONTAINER_JOB=91122255083
PROJECTION_EVALUATION_VERSION=projection-replay-evaluation-v2
PROJECTION_EVALUATION_FINGERPRINT_VERSION=projection-replay-evaluation-fingerprint-v2
PROJECTION_EVALUATION_PROJECTION_SNAPSHOT_VERSION=projection-replay-projection-snapshot-v2
PROJECTION_EVALUATION_TRUTH_SNAPSHOT_VERSION=projection-replay-truth-snapshot-v2
PROJECTION_EVALUATION_AGGREGATE_VERSION=projection-replay-aggregate-v2
PROJECTION_EVALUATION_AGGREGATE_FINGERPRINT_VERSION=projection-replay-aggregate-fingerprint-v2
PROJECTION_EVALUATION_POLICY_VERSION=projection-replay-evaluation-policy-v2
PROJECTION_EVALUATION_TRUTH_KNOWLEDGE_CUTOFF=POINT_AVAILABILITY_EVIDENCE
PROJECTION_EVALUATION_DUPLICATE_TIMESTAMP_POLICY=CONFLICT_REJECTION
PROJECTION_EVALUATION_PROJECTION_OUTPUT_LINEAGE=ENFORCED
PROJECTION_EVALUATION_ALTITUDE_STATUS_LINEAGE=ENFORCED
PROJECTION_EVALUATION_INTERPOLATION_PLAUSIBILITY=ENFORCED
PROJECTION_EVALUATION_ENDPOINT_METRICS=IMPLEMENTED
PROJECTION_EVALUATION_LEAD_TIME_METRICS=IMPLEMENTED
PROJECTION_EVALUATION_CONFIDENCE_COMPARISON=IMPLEMENTED
PROJECTION_EVALUATION_AGGREGATION_IDENTITY=METHOD_VERSION_CLASS_HORIZON_STEP_POLICY
PROJECTION_EVALUATION_UNAVAILABLE_ACCURACY_ISOLATION=ENFORCED
PROJECTION_EVALUATION_ARRIVAL_SELECTIVE_PREDICTION_ACCOUNTING=ENFORCED
PROJECTION_EVALUATION_MICRO_AND_MACRO_AGGREGATION=IMPLEMENTED
PROJECTION_EVALUATION_STATISTICAL_MEDIAN=CONVENTIONAL
PROJECTION_EVALUATION_DERIVED_METRIC_RECOMPUTATION=ENFORCED
PROJECTION_EVALUATION_AGGREGATE_INPUT_FINGERPRINT=GENERATED_AT_INDEPENDENT
PROJECTION_EVALUATION_ACTUAL_ARRIVAL_ICAO_PATTERN=^[A-Z0-9]{4}$
PROJECTION_EVALUATION_PERMANENT_REVIEW_AUDIT=CI_CONFIRMED
PROJECTION_EVALUATION_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_EVALUATION_ENGINEERING_DEBT=CLOSED
PROJECTION_EVALUATION_OPEN_CONFIRMED_FINDINGS=0
PROJECTION_EVALUATION_UNCLASSIFIED_FINDINGS=0
PROJECTION_EVALUATION_DEFERRED_FINDINGS=0
PROJECTION_EVALUATION_ADDITIONAL_CODE_FIXES_REQUIRED=NO
PROJECTION_EVALUATION_FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_EVALUATION_FORMAL_CLOSURE=COMPLETE
PROJECTION_EVALUATION_REVIEW_STATUS=CLOSED
```

The exact engineering commit passed Backend Quality, Backend Race Safety, PostgreSQL
16 Integration, and Backend Container in run `30619973772`. The formal-closure commit
must pass the same four jobs before the external final closure report is issued.

## 38. Projection Production Corrective Engineering and Formal Reclosure

Projection Production was first hardened against baseline
`298d3fdb2d11b1797ce3728b116702b0a978d870`, with engineering commit
`c01b6ee0affff185adeda8e7fb0e1c39681cbe8c` and formal-closure commit
`0f1a31f56f4baf232e978d240216068a001a184e`. Both exact push-triggered Backend
Continuous Integration runs succeeded.

The formal review was later reopened for one dependency-boundary finding: a
substitutable Historical Projector could return a projection with correct identity,
horizon, method, status, and generation time while failing to prove that its output
was produced from the Selection and Pattern published by the composer.

The corrective implementation introduces:

```text
HistoricalProjectionAdapter
HistoricalProjectionOutcome
ApprovedProjectionLineage
independent continuation fingerprint reconstruction
exact selected-neighbor provenance reconstruction
composer comparison with authorized Plan, Selection, Pattern, and selected IDs
controlled fallback or typed error on output-lineage drift
```

The existing single Horizon Plan, immutable request snapshot, route binding,
cross-contract evidence binding, projector identity postconditions, Estimated
Arrival-only adapter, limited-evidence notices, error-chain preservation, and request
and composition fingerprints remain unchanged and protected by their regression tests.

Permanent review enforcement is implemented in:

```text
apps/api/tools/projectionproductionreviewaudit
```

The authoritative review record is:

```text
docs/144_PROJECTION_PRODUCTION_REVIEW_HARDENING.md
```

```text
PROJECTION_PRODUCTION_ENGINEERING_CLOSURE_COMMIT=c01b6ee0affff185adeda8e7fb0e1c39681cbe8c
PROJECTION_PRODUCTION_ENGINEERING_CLOSURE_GITHUB_ACTIONS_RUN=30624533886
PROJECTION_PRODUCTION_ENGINEERING_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB=91136606689
PROJECTION_PRODUCTION_ENGINEERING_CLOSURE_BACKEND_RACE_SAFETY_JOB=91136606649
PROJECTION_PRODUCTION_ENGINEERING_CLOSURE_BACKEND_QUALITY_JOB=91136606715
PROJECTION_PRODUCTION_ENGINEERING_CLOSURE_BACKEND_CONTAINER_JOB=91136827987
PROJECTION_PRODUCTION_PRIOR_FORMAL_CLOSURE_COMMIT=0f1a31f56f4baf232e978d240216068a001a184e
PROJECTION_PRODUCTION_PRIOR_FORMAL_CLOSURE_GITHUB_ACTIONS_RUN=30626948379
PROJECTION_PRODUCTION_PRIOR_FORMAL_CLOSURE_BACKEND_RACE_SAFETY_JOB=91144310170
PROJECTION_PRODUCTION_PRIOR_FORMAL_CLOSURE_BACKEND_QUALITY_JOB=91144310191
PROJECTION_PRODUCTION_PRIOR_FORMAL_CLOSURE_POSTGRESQL_16_INTEGRATION_JOB=91144310201
PROJECTION_PRODUCTION_PRIOR_FORMAL_CLOSURE_BACKEND_CONTAINER_JOB=91144541785
PROJECTION_PRODUCTION_CORRECTIVE_ENGINEERING_COMMIT=2f352821f7ef5d1a26bbb0899bad7fc431d6363c
PROJECTION_PRODUCTION_CORRECTIVE_ENGINEERING_GITHUB_ACTIONS_RUN=30629428359
PROJECTION_PRODUCTION_CORRECTIVE_ENGINEERING_BACKEND_QUALITY_JOB=91152099577
PROJECTION_PRODUCTION_CORRECTIVE_ENGINEERING_POSTGRESQL_16_INTEGRATION_JOB=91152099674
PROJECTION_PRODUCTION_CORRECTIVE_ENGINEERING_BACKEND_RACE_SAFETY_JOB=91152099675
PROJECTION_PRODUCTION_CORRECTIVE_ENGINEERING_BACKEND_CONTAINER_JOB=91152326899
PROJECTION_PRODUCTION_FORMAL_CLOSURE_REOPENED_BASELINE=0f1a31f56f4baf232e978d240216068a001a184e
PROJECTION_PRODUCTION_VERSION=projection-production-composition-v2
PROJECTION_PRODUCTION_REQUEST_FINGERPRINT_VERSION=projection-production-request-fingerprint-v2
PROJECTION_PRODUCTION_COMPOSITION_FINGERPRINT_VERSION=projection-production-composition-fingerprint-v2
PROJECTION_PRODUCTION_APPROVED_LINEAGE_VERSION=historical-approved-projection-lineage-v1
PROJECTION_PRODUCTION_SINGLE_HORIZON_PLAN=CI_CONFIRMED
PROJECTION_PRODUCTION_IMMUTABLE_REQUEST_SNAPSHOT=CI_CONFIRMED
PROJECTION_PRODUCTION_ROUTE_REQUEST_BINDING=CI_CONFIRMED
PROJECTION_PRODUCTION_CROSS_CONTRACT_EVIDENCE_BINDING=CI_CONFIRMED
PROJECTION_PRODUCTION_PROJECTOR_POSTCONDITIONS=CI_CONFIRMED
PROJECTION_PRODUCTION_ARRIVAL_ONLY_MUTATION_BOUNDARY=CI_CONFIRMED
PROJECTION_PRODUCTION_UNAVAILABLE_HISTORICAL_REJECTION=CI_CONFIRMED
PROJECTION_PRODUCTION_LIMITED_EVIDENCE_NOTICES=CI_CONFIRMED
PROJECTION_PRODUCTION_DEPENDENCY_ERROR_CHAIN=CI_CONFIRMED
PROJECTION_PRODUCTION_REQUEST_AND_COMPOSITION_FINGERPRINTS=CI_CONFIRMED
PROJECTION_PRODUCTION_HISTORICAL_PROJECTOR_OUTPUT_LINEAGE_BINDING=CI_CONFIRMED
PROJECTION_PRODUCTION_HISTORICAL_PROJECTION_PROVENANCE_RECONSTRUCTION=CI_CONFIRMED
PROJECTION_PRODUCTION_HISTORICAL_SELECTED_NEIGHBOR_PROVENANCE=CI_CONFIRMED
PROJECTION_PRODUCTION_PERMANENT_REVIEW_AUDIT=CI_CONFIRMED
PROJECTION_PRODUCTION_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_PRODUCTION_ENGINEERING_DEBT=CLOSED
PROJECTION_PRODUCTION_OPEN_CONFIRMED_FINDINGS=0
PROJECTION_PRODUCTION_UNCLASSIFIED_FINDINGS=0
PROJECTION_PRODUCTION_DEFERRED_FINDINGS=0
PROJECTION_PRODUCTION_ADDITIONAL_CODE_FIXES_REQUIRED=NO
PROJECTION_PRODUCTION_FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_PRODUCTION_FORMAL_CLOSURE=COMPLETE
PROJECTION_PRODUCTION_REVIEW_STATUS=CLOSED
```

The previous exact Continuous Integration evidence remains authoritative for the
previously closed findings. Corrective commit
`2f352821f7ef5d1a26bbb0899bad7fc431d6363c` passed exact push-triggered Backend
Continuous Integration run `30629428359`, including Backend Quality, Backend Race
Safety, PostgreSQL 16 Integration, Backend Container, and the permanent Projection
Production review audit. The formal-reclosure commit containing this record must pass
the same four jobs before the external final closure report is issued.

## 39. Projection Read Engineering Completion and Permanent Audit

Projection Read completed two correctness-hardening increments. The first established strict
snapshot, route-row, Composer-output, context, generation-time, and duration-boundary contracts.
The second added bounded historical-candidate backfill and exact contributing-record lineage for
route-history fingerprints.

Permanent review enforcement is implemented in:

```text
apps/api/tools/projectionreadreviewaudit
```

The authoritative review record is:

```text
docs/145_PROJECTION_READ_REVIEW_HARDENING.md
```

```text
PROJECTION_READ_REVIEW_BASELINE=87b853e5a74bc5b8e0cd9bcb3f1e8e13eec8df0e
PROJECTION_READ_CONTRACT_HARDENING_COMMIT=4eeff2b9f5b5c17dd6b7ebe5d0be4a7bd836fb37
PROJECTION_READ_CONTRACT_HARDENING_GITHUB_ACTIONS_RUN=30638188394
PROJECTION_READ_CONTRACT_HARDENING_POSTGRESQL_16_INTEGRATION_JOB=91181076159
PROJECTION_READ_CONTRACT_HARDENING_BACKEND_RACE_SAFETY_JOB=91181076172
PROJECTION_READ_CONTRACT_HARDENING_BACKEND_QUALITY_JOB=91181076240
PROJECTION_READ_CONTRACT_HARDENING_BACKEND_CONTAINER_JOB=91181409362
PROJECTION_READ_EVIDENCE_HARDENING_COMMIT=9dda4b102497028b59280143b86bf84564afb136
PROJECTION_READ_EVIDENCE_HARDENING_GITHUB_ACTIONS_RUN=30648605652
PROJECTION_READ_EVIDENCE_HARDENING_BACKEND_QUALITY_JOB=91216081636
PROJECTION_READ_EVIDENCE_HARDENING_POSTGRESQL_16_INTEGRATION_JOB=91216081729
PROJECTION_READ_EVIDENCE_HARDENING_BACKEND_RACE_SAFETY_JOB=91216081733
PROJECTION_READ_EVIDENCE_HARDENING_BACKEND_CONTAINER_JOB=91216395238
PROJECTION_READ_ATOMIC_REPEATABLE_READ_SNAPSHOT=CI_CONFIRMED
PROJECTION_READ_COMPOSER_OUTPUT_POSTCONDITIONS=CI_CONFIRMED
PROJECTION_READ_ROUTE_ROW_PAYLOAD_BINDING=CI_CONFIRMED
PROJECTION_READ_HISTORICAL_CANDIDATE_BACKFILL=CI_CONFIRMED
PROJECTION_READ_ROUTE_HISTORY_RECORD_LINEAGE=CI_CONFIRMED
PROJECTION_READ_PERMANENT_REVIEW_AUDIT=IMPLEMENTED_PENDING_EXACT_CI
PROJECTION_READ_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_READ_ENGINEERING_DEBT=CLOSED
PROJECTION_READ_OPEN_CONFIRMED_FINDINGS=0
PROJECTION_READ_UNCLASSIFIED_FINDINGS=0
PROJECTION_READ_DEFERRED_FINDINGS=0
PROJECTION_READ_ADDITIONAL_CODE_FIXES_REQUIRED=NO
PROJECTION_READ_FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=YES
PROJECTION_READ_FORMAL_CLOSURE=OPEN_PENDING_EXACT_CI
PROJECTION_READ_REVIEW_STATUS=OPEN_PENDING_EXACT_CI_AND_FORMAL_CLOSURE
```

The permanent-audit commit must pass Backend Quality, Backend Race Safety, PostgreSQL 16
Integration, and Backend Container. A separate final documentation increment will record that
exact evidence and change the formal review status to closed.
