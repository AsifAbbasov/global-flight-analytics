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

The horizon planner prevents:

```text
zero or negative duration
duration below the configured minimum
duration above the configured maximum
point-count overflow
invalid as-of time
inconsistent horizon end time
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
6 route observations
6 distinct days
recent route support
complete ZAAA to ZBBB route
high route confidence
```

The resulting decision is:

```text
allowed
```

The guard prevents a rare or weakly supported route from being treated as a strong historical prediction pattern.

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
prediction snapshot
future observed truth
horizontal error by horizon
vertical error when altitude is supported
endpoint error
coverage
confidence comparison
aggregate evaluation across replay results
deterministic fingerprints
bounded configuration
```

This is the foundation for later model comparison and calibration.

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
Production composition: projection-production-composition-v1
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
6 route observations
6 distinct route days
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
