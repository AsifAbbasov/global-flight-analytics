# Document 35 — Stage 12 Stability and Explainability Completion

Status: COMPLETED
Completion date: 2026-07-17
Project: Global Flight Analytics
Stage: 12 — Stability and Explainability
Completion classification: Production Stability and Explainability Foundation
Evidence baseline commit: `750140d`

---

## 1. Purpose

This document closes Stage 12 with an evidence-based implementation record.

Stage 12 is complete as a bounded, deterministic, explainable, research-only Production Stability and Explainability foundation.

The completed foundation can:

```text
create immutable deterministic forecast versions
reuse identical forecast replays without duplicate version identity
compare consecutive forecast decisions
analyze a bounded sequence of forecast versions
propagate confidence through explicit analytical dependencies
rank failure and limitation conditions
preserve unknown causes instead of inventing intent
withhold pilot-intent, air traffic control instruction, and exact-cause claims
enforce research-only publication boundaries
compose all Stage 12 capabilities over the PostgreSQL Projection Intelligence reader
expose a stable read-only HTTP response
verify the full PostgreSQL and Fiber path with deterministic fixture cleanup
```

This completion statement does not claim:

```text
forecast accuracy certification
calibrated probability
operational flight prediction
pilot-intent detection
air traffic control instruction reconstruction
exact maneuver-cause attribution
flight planning suitability
safety-critical decision support
regulated aviation software
automatic operational authorization
long-term persistent forecast-version storage
frontend Stability Intelligence visualization
```

The Stage 12 result remains a research and portfolio-grade analytical output.

---

## 2. Stage Boundary

Stage 9 produces bounded future estimates.

Stage 10 adds weather context without converting association into causation.

Stage 11 adds multi-aircraft and airspace interaction context without creating operational separation logic.

Stage 12 evaluates whether analytical decisions remain stable and whether their explanations stay honest.

The boundary is:

```text
Projection Intelligence
produces an estimated forecast
↓
Forecast Versioning
records deterministic immutable forecast identity
↓
Decision Stability
compares two versions
↓
Forecast Stability Analysis
evaluates a bounded version sequence
↓
Confidence Propagation
limits output confidence by required dependencies
↓
Failure Explanation
ranks detected limitations and unknown conditions
↓
Unknown Intervention Guard
prevents unsupported intent or exact-cause claims
↓
Scope Enforcement
prevents operational, directive, certain, and safety-critical publication
↓
Stability Intelligence API
publishes the bounded research result
```

Stage 12 does not rewrite source observations or projections.

It evaluates and explains analytical change around them.

---

## 3. Implemented Architecture

The production read path is:

```text
HTTP GET request
↓
Stability Intelligence handler
↓
production Stability Intelligence service
↓
existing PostgreSQL Projection Intelligence reader
↓
multiple bounded as-of projections
↓
immutable forecast version registration
↓
consecutive Decision Stability evaluations
↓
Forecast Stability Analysis
↓
Confidence Propagation graph
↓
Failure Explanation Engine
↓
Unknown Intervention Guard
↓
Scope Guard Enforcement
↓
standardized Stability Intelligence data transfer object
↓
JSON response
```

The endpoint is:

```text
GET /api/v1/trajectories/:id/stability-intelligence
```

Required query parameters are:

```text
as_of_times
duration_seconds
```

`as_of_times` is a comma-separated ordered list of two to eight RFC 3339 timestamps.

The timestamps must be:

```text
valid
unique
strictly increasing
not in the future
within the existing Projection Intelligence read boundary
```

The endpoint is read-only.

It does not write forecast versions to PostgreSQL.

---

## 4. Acceptance Matrix

| Capability | Implementation status | Verification status |
|---|---:|---:|
| Forecast Versioning | Implemented | Unit, race, regression, and deterministic replay verified |
| Idempotent identical replay | Implemented | Unit verified |
| Decision Stability Evaluator | Implemented | Unit and production composition verified |
| Multi-version Forecast Stability Analysis | Implemented | Unit and production composition verified |
| Confidence Propagation | Implemented | Unit and production composition verified |
| Weakest required dependency confidence cap | Implemented | Unit verified |
| Failure Explanation Engine | Implemented | Unit and production composition verified |
| Unknown cause preservation | Implemented | Unit and production composition verified |
| Unknown Intervention Guard | Implemented | Unit and production composition verified |
| Pilot-intent claim withholding | Implemented | Unit verified |
| Air traffic control instruction claim withholding | Implemented | Unit verified |
| Exact-cause claim withholding | Implemented | Unit verified |
| Scope Guard Enforcement | Implemented | Unit and production composition verified |
| Operational and safety-critical scope blocking | Implemented | Unit verified |
| Explanation API standardization | Implemented | HTTP and JSON contract verified |
| Production PostgreSQL composition | Implemented | Runtime verified |
| Fiber server wiring | Implemented | Route registration and HTTP runtime verified |
| Bounded runtime HTTP execution | Implemented | Fiber synthetic one-second test timeout disabled; production service bounded by a ninety-second verifier context |
| Deterministic production fingerprint | Implemented | Direct replay verified |
| Runtime fixture cleanup | Implemented | Zero persistent rows verified |
| New database migration | Not required | Existing PostgreSQL projection path reused |
| Persistent forecast-version history | Not implemented | Deferred |
| Frontend visualization | Not implemented | Deferred |

---

## 5. Forecast Versioning

Forecast Versioning is implemented in:

```text
apps/api/internal/stabilityintelligence/forecaststability
```

Each version contains:

```text
schema version
deterministic version identifier
ordinal
trajectory identifier
projection schema version
method identity
policy version
implementation version
input fingerprint
output fingerprint
decision fingerprint
parent version identifier
immutable projection snapshot
created-at time
research-only scope guard
```

Supported registration decisions are:

```text
initial_version_created
identical_version_reused
successor_version_created
```

The version identity is deterministic.

Equivalent normalized input produces the same:

```text
output fingerprint
decision fingerprint
version identifier
```

A successor version records explicit changes in:

```text
projection schema
method
policy
implementation
input
output
horizon
```

Forecast Versioning does not claim persistence by itself.

The Stage 12 production service creates request-scoped immutable records from PostgreSQL-backed projections.

---

## 6. Decision Stability Evaluator

Decision Stability compares two immutable forecast versions.

Supported levels are:

```text
unchanged
stable
changed
material_change
indeterminate
```

The evaluator measures:

```text
aligned forecast point count
aligned point share
mean horizontal shift
maximum horizontal shift
point-confidence change
aggregate-confidence change
relative uncertainty change
Estimated Arrival change
projection status change
method change
policy change
implementation change
input change
output change
```

The first policy version is explicitly experimental:

```text
decision-stability-policy-v1-experimental
```

Its thresholds are project policy, not certified aviation boundaries.

They require future historical replay calibration.

A critical limitation is always preserved:

```text
stability_is_not_accuracy
```

A stable forecast may still be inaccurate.

A changed forecast may be an improvement.

---

## 7. Forecast Stability Analysis

Forecast Stability Analysis is implemented in:

```text
apps/api/internal/stabilityintelligence/forecastanalysis
```

The module consumes a bounded ordered history of immutable versions.

It creates consecutive transitions:

```text
version 1 → version 2
version 2 → version 3
version 3 → version 4
```

It calculates:

```text
version count
transition count
comparable transition count
unchanged count
stable count
changed count
material-change count
indeterminate count
stable-transition share
comparable-transition share
material-change share
mean stability score
minimum stability score
score standard deviation
longest stable run
method-change count
policy-change count
implementation-change count
input-change count
output-change count
mean horizontal shift
maximum horizontal shift
latest stability level
```

Supported trends are:

```text
insufficient_history
steady
improving
degrading
volatile
```

Supported health states are:

```text
insufficient_evidence
stable
watch
unstable
```

The analysis does not claim forecast correctness.

It describes consistency and change across versions.

---

## 8. Confidence Propagation

Confidence Propagation is implemented in:

```text
apps/api/internal/stabilityintelligence/confidencepropagation
```

The production dependency graph is:

```text
projection confidence at as-of time 1
projection confidence at as-of time 2
projection confidence at as-of time 3
↓
forecast history stability
↓
Stability Intelligence output confidence
```

Each node records:

```text
node identifier
label
node kind
evidence classification
local score
dependencies
dependency weights
required dependency flag
source fingerprint
```

Supported evidence classifications are:

```text
observed
openly_sourced
derived
estimated
unknown
```

The propagation rules include:

```text
weighted dependency score
local and dependency contribution
weakest required dependency cap
estimated-evidence confidence cap
unknown-evidence confidence cap
cycle rejection
missing-dependency rejection
duplicate-node rejection
duplicate-dependency rejection
deterministic graph fingerprint
```

The final confidence score is not a probability.

The required limitation is:

```text
confidence_is_not_probability
```

---

## 9. Failure Explanation Engine

Failure Explanation is implemented in:

```text
apps/api/internal/stabilityintelligence/failureexplanation
```

The engine consumes normalized signals from:

```text
Decision Stability
Forecast Stability Analysis
Confidence Propagation
```

Signals are separated into:

```text
observed_condition
derived_condition
unknown_cause
```

Supported severities are:

```text
information
warning
blocking
```

Signals are ranked deterministically by:

```text
severity
blocking effect
unknown-cause priority
category
code
source
```

The engine publishes:

```text
primary failure code
ordered failure list
blocking count
warning count
information count
unknown-cause count
explanation confidence
evidence fingerprints
limitations
explanations
research-only scope guard
```

The engine does not turn a detected limitation into proof of operational cause.

The required limitation is:

```text
explanation_not_causation
```

---

## 10. Unknown Intervention Guard

Unknown Intervention Guard is implemented in:

```text
apps/api/internal/stabilityintelligence/unknownintervention
```

The guard protects against unsupported claims about:

```text
pilot intent
air traffic control intent
air traffic control instruction
exact maneuver cause
exact route-change cause
exact plan change
operational intervention
```

Supported claim kinds are:

```text
contextual_association
causal_attribution
intent_attribution
operational_instruction
```

Supported decisions are:

```text
allowed_context_only
limited_context
withheld
```

The production Stability Intelligence service requests only:

```text
contextual_association
```

Its evidence includes:

```text
forecast history analysis
propagated confidence
latest estimated projection
```

Estimated evidence remains explicitly classified and limited.

Causal, intent, and operational-instruction claims are withheld.

The required limitation is:

```text
association_not_causation
```

---

## 11. Scope Guard Enforcement

Scope Guard Enforcement is implemented in:

```text
apps/api/internal/stabilityintelligence/scopeenforcement
```

Allowed publication scopes are:

```text
research_analysis
research_visualization
```

Forbidden publication scopes are:

```text
operational_decision
air_traffic_control
flight_planning
safety_critical
```

Supported claim strengths are:

```text
descriptive
analytical
causal
directive
certain
```

Rules are:

```text
descriptive research claims may be allowed
analytical research claims may be allowed
causal research claims are limited
directive claims are blocked
certainty claims are blocked
operational claims are blocked
air traffic control claims are blocked
flight-planning claims are blocked
safety-critical claims are blocked
missing or unknown source guards are blocked
```

Passing the guard is not operational authorization.

The required limitation is:

```text
enforcement_not_authorization
```

---

## 12. Production Stability Intelligence Composition

The production composition is implemented in:

```text
apps/api/internal/stabilityintelligence/stabilityproduction
```

The service accepts:

```text
trajectory identifier
two to eight ordered as-of timestamps
projection duration
```

For every as-of timestamp it requests a projection from the existing production Projection Intelligence reader.

Therefore the production data source remains:

```text
PostgreSQL flight trajectories
PostgreSQL flight states
PostgreSQL Route Intelligence evidence when available
```

The Stage 12 composition does not introduce an alternative source of aviation truth.

The service then performs:

```text
projection validation
forecast version registration
history analysis
confidence propagation
failure explanation
unknown-intervention evaluation
scope enforcement
production result validation
deterministic fingerprint generation
```

The production result includes all component fingerprints and scope guards.

---

## 13. Explanation API Standardization

The standardized data transfer object is implemented in:

```text
apps/api/internal/http/dto/stability_intelligence.go
```

The response includes:

```text
production version
trajectory identifier
ordered as-of timestamps
full Projection Intelligence responses
forecast-version summaries
pairwise transition summaries
forecast analysis
propagated confidence
failure explanation
unknown-intervention decision
scope-enforcement decision
declared scope guards
production input fingerprint
generated-at time
```

The response does not hide limited or blocking evidence.

The client can inspect:

```text
which version changed
how much aligned points moved
whether uncertainty changed
whether confidence changed
whether the method changed
whether the history is stable or volatile
which dependency limited confidence
which failure condition ranked first
whether intervention attribution was withheld
whether publication claims were allowed, limited, or blocked
```

---

## 14. HTTP Contract

The route is registered in:

```text
apps/api/internal/server/stability_intelligence_routes.go
```

The database production wiring is registered in:

```text
apps/api/internal/server/database_routes.go
```

Endpoint:

```text
GET /api/v1/trajectories/:id/stability-intelligence
```

Example query shape:

```text
?as_of_times=2026-07-17T00:00:00Z,2026-07-17T00:00:30Z,2026-07-17T00:01:00Z
&duration_seconds=300
```

Validation errors include:

```text
invalid trajectory UUID
missing as-of timestamps
fewer than two timestamps
more than eight timestamps
invalid RFC 3339 timestamp
duplicate timestamp
non-increasing timestamp order
invalid duration
future analytical timestamp
projection-policy rejection
```

HTTP responses include:

```text
200 success
400 invalid request
404 trajectory not found
408 request canceled
500 invalid service contract or load failure
503 service unavailable
504 timeout
```

---

## 15. PostgreSQL Runtime Verification

Runtime verification is implemented in:

```text
apps/api/cmd/verify-postgres-stability-intelligence-http-api
```

The verifier creates a deterministic temporary fixture:

```text
one trajectory
six flight-state observations
zero route-result rows
one callsign
one ICAO 24-bit address
five-minute observed trajectory
three Stability Intelligence as-of timestamps
five-minute projection duration
```

The as-of timestamps are:

```text
latest observed time minus one minute
latest observed time minus thirty seconds
latest observed time
```

This produces:

```text
three PostgreSQL-backed projections
three immutable forecast versions
two consecutive stability transitions
one forecast-history analysis
one propagated-confidence result
one failure explanation
one unknown-intervention result
one scope-enforcement result
one standardized HTTP response
```

Before the persistent runtime scenario, the verifier inserts the complete trajectory and flight-state fixture inside a PostgreSQL transaction, validates row counts, rolls the transaction back, and confirms that zero rows remain. This preflight checks the canonical trajectory identity format and all active database constraints before the HTTP path runs.

The HTTP test transport does not use Fiber's synthetic one-second default. Every database-backed Stability Intelligence service call receives a real ninety-second context deadline, while Fiber's shorter test-only timeout is disabled. This keeps the runtime check both tolerant of free-tier database latency and bounded against a stalled production call.

The verifier checks:

```text
required schema objects
canonical trajectory identity
transactional fixture compatibility preflight
fixture counts
PostgreSQL projection hydration
as-of boundaries
multi-version identity
parent-version lineage
transition lineage
forecast analysis
confidence propagation
failure explanation
unknown-intervention protection
scope enforcement
direct production composition
deterministic replay
bounded Fiber HTTP execution beyond one second
Fiber endpoint
JSON response
validation errors
not-found response
fixture cleanup
zero persistent verification rows
```

No Stage 12 table is created.

The verifier deletes all temporary rows.

---

## 16. Determinism

Deterministic identity exists at multiple levels:

```text
projection input fingerprint
forecast output fingerprint
forecast decision fingerprint
forecast version identifier
pairwise stability fingerprint
forecast-analysis fingerprint
confidence graph fingerprint
failure-explanation fingerprint
unknown-intervention fingerprint
scope-enforcement fingerprint
production Stability Intelligence fingerprint
```

The production fingerprint excludes incidental execution order.

It is built from:

```text
trajectory identifier
ordered as-of timestamps
forecast version identifiers
projection fingerprints
analysis fingerprint
confidence fingerprint
failure fingerprint
intervention fingerprint
scope fingerprint
declared guards
```

A replay with equivalent normalized inputs must return the same production fingerprint.

---

## 17. Confidence and Honesty Rules

Stage 12 enforces the following interpretation rules:

```text
stability is not accuracy
confidence is not probability
explanation is not causation
association is not causation
scope enforcement is not operational authorization
unknown cause must remain unknown
estimated evidence must remain estimated
weak required evidence must limit final confidence
```

These rules are part of the production contract.

They are not optional frontend wording.

---

## 18. Security and Operational Scope

The endpoint is read-only.

It does not:

```text
control aircraft
contact pilots
contact air traffic control
issue route instructions
change flight plans
produce separation commands
produce collision-avoidance commands
authorize operational decisions
```

The production service consumes existing internal readers.

The frontend does not call aviation providers directly.

The endpoint remains protected by the existing server middleware:

```text
request identification
request logging
security headers
Cross-Origin Resource Sharing policy
rate limiting
body limits
read timeout
write timeout
idle timeout
panic recovery
```

---

## 19. Database Decision

Stage 12 does not add a migration.

Reason:

```text
the production feature is a bounded read-only analysis
the existing PostgreSQL projection reader already provides source evidence
request-scoped deterministic version records are sufficient for the current stage
persistent version history is not required to prove the production algorithm
```

Persistent forecast-version storage remains a future capability.

It should be added only when the product requires:

```text
cross-request history
longitudinal dashboards
version audit retention
offline calibration datasets
large-scale replay analysis
```

That work must define retention, indexing, idempotency, and storage budgets before adding a table.

---

## 20. Known Limitations

The completed Stage 12 foundation has explicit limitations:

```text
forecast stability is not forecast accuracy
confidence is not calibrated probability
thresholds remain project-derived and require replay calibration
request-scoped history is limited to two through eight versions
version records are not persisted across requests
projection quality remains limited by open-data coverage
historical route evidence may be absent
weather and airspace context are not automatically converted into causal claims
pilot intent is unavailable
air traffic control instruction is unavailable
official flight-plan intent may be unavailable
frontend visualization is not implemented
no operational aviation use is authorized
```

---

## 21. Deferred Work

Deferred work includes:

```text
persistent forecast-version store
retention and archival policy
longitudinal stability dashboards
historical calibration of stability thresholds
forecast stability versus observed accuracy correlation
calibrated reliability curves
cross-method benchmark reports
frontend version timeline
frontend confidence dependency graph
frontend failure explanation panel
frontend scope-guard panel
large-scale replay materialization
alerting on material forecast change
```

Deferred work must preserve the Stage 12 scope guards.

---

## 22. Test Evidence

The final installer executes:

```text
targeted Stability Intelligence production tests
targeted handler tests
targeted server route tests
targeted command compilation
targeted race detector
complete Stability Intelligence regression tests
HTTP and server regression tests
targeted static analysis
PostgreSQL and Fiber runtime verification
complete backend regression tests
complete backend static analysis
exact file-delta verification
pre-existing worktree preservation verification
```

Required runtime result:

```text
Persistent verification rows: 0
Result: PASS
```

---

## 23. Stage 12 Commit Evidence

Stage 12 foundation commits before final closure:

```text
b7454ac
feat: add forecast versioning and decision stability

2c97ed0
feat: add forecast stability analysis and confidence propagation

750140d
feat: add failure explanation and scope enforcement
```

The final closure increment adds:

```text
production Stability Intelligence composition
standardized Stability Intelligence data transfer object
read-only HTTP handler
Fiber route
database production wiring
PostgreSQL runtime verifier
Stage 12 completion documentation
documentation index alignment
implementation sequence alignment
```

---

## 24. Formal Completion Statement

Stage 12 — Stability and Explainability is complete.

The completed production foundation provides:

```text
immutable deterministic forecast versions
pairwise Decision Stability
multi-version Forecast Stability Analysis
dependency-aware Confidence Propagation
ranked Failure Explanation
Unknown Intervention protection
Scope Guard Enforcement
standardized read-only API output
PostgreSQL-backed production composition
Fiber HTTP integration
deterministic runtime evidence
zero-row fixture cleanup
official documentation closure
```

The formal scope remains:

```text
research-only
open-data aware
deterministic
confidence-limited
explainable
non-causal unless independently proven
non-operational
not safety-critical
```

Stage 12 has zero remaining implementation increments after the final closure commit.
