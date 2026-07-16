# Document 34 — Stage 11 Airspace Intelligence Completion

Status: COMPLETED
Completion date: 2026-07-17
Project: Global Flight Analytics
Stage: 11 — Airspace Intelligence
Completion classification: Production Airspace Intelligence Foundation
Implementation baseline commit: `f98a1563b6fdeed3c2a08686fc42b1d7dfed7823`

---

## 1. Purpose

This document closes Stage 11 with an evidence-based implementation record.

Stage 11 is complete as a bounded, deterministic, explainable, research-only Production Airspace Intelligence foundation. It can construct a local airborne traffic scene, calculate policy-controlled interaction radii, scan every eligible aircraft pair, build an airborne interaction graph, classify research separation-risk context, index temporal occupancy, calculate multidimensional synthetic-sector complexity, aggregate region-level airspace analytics, read bounded evidence from PostgreSQL, expose a read-only HTTP endpoint, and verify the complete production path against a deterministic database fixture.

This completion statement does not claim:

```text
operational air traffic control support
certified aircraft separation monitoring
collision avoidance
conflict alerting
controller workload measurement
official airspace sector modeling
regulatory separation minima
pilot-intent inference
controller-intent inference
maneuver-cause inference
complete surveillance coverage
radar fusion
multilateration fusion
satellite surveillance fusion
safety-of-life availability
real-time safety guarantees
frontend Airspace Intelligence visualization
```

Every Stage 11 result remains protected by:

```text
bounded region and time window
explicit observation age policy
explicit pair-time policy
explicit horizontal and vertical radius policy
altitude-reference awareness
unknown-altitude preservation
confidence
limitations
explanations
provenance
source names
deterministic input fingerprints
research-only scope guards
synthetic-sector scope guard
capacity limits
future-evidence protection
failed-ingestion-run exclusion
```

---

## 2. Scope Alignment

Stage 10 adds Weather Context around one trajectory and its projection.

Stage 11 adds multi-aircraft and regional context without rewriting proximity, complexity, or risk classification as an operational aviation fact.

The boundary is:

```text
Stage 10
single-trajectory weather context
weather trust
four-dimensional weather alignment
weather encounter summary
policy-controlled uncertainty adjustment

Stage 11
multiple-aircraft scene construction
pairwise proximity candidates
interaction graph
research separation-risk classification
temporal occupancy
synthetic-sector complexity
regional airspace analytics
```

Stage 11 does not modify Stage 8 historical facts, Stage 9 projection claims, or Stage 10 Weather Context claims.

The mandatory regional scope guard is:

```text
research_only_not_for_operational_separation_or_air_traffic_control_use
```

The mandatory synthetic-sector limitation is:

```text
synthetic_grid_not_official_sectors
```

---

## 3. Implemented Production Architecture

The production read path is:

```text
HTTP request
↓
Airspace Region Analytics handler
↓
Airspace Production service
↓
PostgreSQL observation reader
↓
flight_states joined to successful ingestion_runs
↓
region and time-window filtering
↓
deterministic minute-snapshot assembly
↓
Local Traffic Scene Builder
↓
Multi-Aircraft Proximity Scanner
↓
Airborne Interaction Graph
↓
Separation Risk Intelligence
↓
Temporal Airspace Occupancy Index
↓
Sector Complexity Score
↓
Airspace Region Analytics
↓
stable read-only HTTP response
```

The production PostgreSQL evidence path is:

```text
ingestion_runs
+
flight_states
↓
successful-ingestion-run boundary
↓
as-of-time boundary
↓
regional coordinate boundary
↓
observation capacity boundary
↓
production domain services
↓
read-only regional analytics result
```

The production endpoint is:

```text
GET /api/v1/airspace/regions/:code/analytics
```

Required query parameter:

```text
as_of_time
RFC 3339 timestamp
```

Optional query parameter:

```text
window_seconds
whole number of minutes
minimum: 60 seconds
maximum: 3600 seconds
default: 300 seconds
```

---

## 4. Acceptance Matrix

| Capability | Implementation status | Verification status |
|---|---:|---:|
| Airborne Interaction Graph Foundation | Implemented | Unit, race, static analysis, composition, and runtime path verified |
| Interaction Radius Policy | Implemented | Unit, boundary, confidence, and composition verified |
| Local Traffic Scene Builder | Implemented | Unit, deterministic selection, exclusion, and composition verified |
| Multi-Aircraft Proximity Scanner | Implemented | Unit, pair enumeration, geometry, motion, and composition verified |
| Interaction Graph composition | Implemented | Unit and candidate-to-edge consistency verified |
| Separation Risk Intelligence | Implemented | Unit, multidimensional classification, and scope protection verified |
| Temporal Airspace Occupancy Index | Implemented | Unit, minute bucket, spatial cell, altitude band, and runtime verified |
| Sector Complexity Score | Implemented | Unit, multidimensional scoring, and runtime verified |
| Airspace Region Analytics | Implemented | Unit, regional rollup, pressure, trend, and runtime verified |
| PostgreSQL observation reader | Implemented | Successful-run, time, region, altitude, and capacity boundaries verified |
| Production composition service | Implemented | Direct deterministic replay and HTTP runtime verified |
| Read-only HTTP endpoint | Implemented | HTTP 200, HTTP 400, and HTTP 404 verified |
| Production server wiring | Implemented | Unit, regression, static analysis, and route runtime verified |
| Deterministic fingerprints | Implemented | Unit and runtime replay verified |
| Research-only scope guards | Implemented | Contract, data transfer object, HTTP, and runtime verified |
| Synthetic-sector limitation | Implemented | Contract, HTTP, and runtime verified |
| Runtime fixture cleanup | Implemented | Zero persistent verification rows verified |
| New database migration | Not required | Existing ingestion_runs and flight_states reused |
| Official sector boundaries | Not implemented | Deferred and explicitly not claimed |
| Operational conflict alert | Not implemented | Prohibited by scope guard |
| Frontend Airspace Intelligence interface | Not implemented | Deferred |

---

## 5. Airborne Interaction Graph Foundation

The graph foundation is implemented in:

```text
apps/api/internal/airspaceintelligence/interactiongraph
```

The graph represents a bounded research snapshot.

Nodes represent prepared airborne aircraft observations.

Edges represent accepted pairwise interaction candidates.

The graph publishes:

```text
schema version
result status
region code
as-of time
nodes
edges
metrics
confidence
limitations
explanations
scope guard
provenance
generated-at time
```

Node evidence includes:

```text
canonical node identity
trajectory identity when available
flight identity when available
aircraft identity when available
ICAO 24-bit address
callsign
latitude
longitude
optional altitude
altitude reference
velocity
heading
vertical rate
observation time
source
quality score
```

Edge evidence includes:

```text
source node
target node
interaction kind
horizontal distance
optional vertical distance
observation-time difference
closing rate
confidence
limitations
explanations
```

The graph is deterministic for equivalent normalized input.

It is not an operational traffic-control graph.

---

## 6. Interaction Radius Policy

The radius policy is implemented in:

```text
apps/api/internal/airspaceintelligence/interactionradius
```

Policy version:

```text
interaction-radius-policy-v1
```

The policy controls:

```text
minimum horizontal radius
base horizontal radius
maximum horizontal radius
horizontal lookahead
quality uncertainty expansion
minimum vertical radius
base vertical radius
maximum vertical radius
vertical lookahead
maximum observation age
maximum pair-time difference
minimum usable quality
minimum allowed quality
motion thresholds
confidence thresholds
confidence weights
```

Default horizontal bounds are:

```text
minimum: 10 kilometers
base: 20 kilometers
maximum: 80 kilometers
lookahead: 2 minutes
```

Default vertical bounds are:

```text
minimum: 500 meters
base: 750 meters
maximum: 3000 meters
lookahead: 1 minute
```

Default temporal bounds are:

```text
maximum observation age: 90 seconds
maximum pair-time difference: 30 seconds
```

The policy may allow, limit, or block an observation.

A limited decision does not silently become complete evidence.

---

## 7. Local Traffic Scene Builder

The scene builder is implemented in:

```text
apps/api/internal/airspaceintelligence/localtrafficscene
```

The builder creates one bounded airborne scene for one region and one as-of time.

It performs:

```text
identity normalization
callsign normalization
source normalization
UTC normalization
region filtering
on-ground exclusion
future-evidence exclusion
one-observation-per-aircraft selection
quality-aware duplicate resolution
Interaction Radius Policy composition
blocked-observation exclusion
graph-ready node preparation
```

Deterministic duplicate selection order is:

```text
newest observation
then higher quality
then lexicographically smaller source name
```

Explicit exclusion reasons include:

```text
on_ground
outside_region
future_evidence
superseded_duplicate
radius_policy_blocked
```

The scene publishes coverage metrics and does not hide rejected material evidence.

---

## 8. Multi-Aircraft Proximity Scanner

The scanner is implemented in:

```text
apps/api/internal/airspaceintelligence/proximityscanner
```

The scanner evaluates every unique aircraft pair once.

For `n` aircraft, the maximum possible pair count is:

```text
n × (n - 1) / 2
```

The scanner applies:

```text
pair-time boundary
Haversine horizontal distance
policy-derived horizontal radius
comparable-altitude test
policy-derived vertical radius
vertical separation when comparable
signed closing-rate calculation
relative-motion classification
candidate confidence
candidate limitations
candidate explanations
```

Supported relative-motion kinds include:

```text
nearby
converging
parallel
diverging
```

The scanner preserves unknown vertical evidence.

When altitude cannot be compared honestly, vertical filtering is withheld and the candidate is marked limited.

The scanner also creates graph edges directly from accepted candidates.

---

## 9. Separation Risk Intelligence

The risk evaluator is implemented in:

```text
apps/api/internal/airspaceintelligence/separationrisk
```

Policy version:

```text
separation-risk-policy-v1
```

Supported research classifications are:

```text
indeterminate
contextual
elevated
high
```

The score combines:

```text
horizontal proximity
vertical proximity
closing motion
temporal alignment
evidence confidence
```

A determinate level requires comparable vertical evidence.

Without comparable vertical evidence, the evaluator returns:

```text
indeterminate
```

It does not infer a safe or unsafe operational separation state.

It does not implement regulatory separation minima.

It does not implement collision-avoidance logic.

---

## 10. Temporal Airspace Occupancy Index

The occupancy index is implemented in:

```text
apps/api/internal/airspaceintelligence/airspaceregionanalytics
```

Default grid policy:

```text
time bucket: 60 seconds
latitude cell: 1 degree
longitude cell: 1 degree
altitude band: 3000 meters
```

The index preserves three dimensions plus time:

```text
latitude cell
longitude cell
altitude band or unknown-altitude band
time bucket
```

The same aircraft is counted at most once in one time bucket.

Deterministic selection order is:

```text
newest observation
then higher quality
then lexicographically smaller source name
```

The occupancy index publishes:

```text
bucket count
expected bucket count
occupied cell count
aircraft observation count
unique aircraft count
unknown altitude count
peak aircraft per bucket
peak occupied cells
mean aircraft per bucket
temporal coverage
```

An unknown altitude is not converted into altitude zero.

---

## 11. Sector Complexity Score

The complexity score is implemented in:

```text
apps/api/internal/airspaceintelligence/airspaceregionanalytics
```

Density is only one component.

Default complexity components are:

```text
density
pair interaction
determinate risk
heading dispersion
speed variability
altitude mixing
```

Default weights are:

```text
density: 0.24
pair interaction: 0.20
determinate risk: 0.24
heading dispersion: 0.12
speed variability: 0.08
altitude mixing: 0.12
```

Supported complexity levels are:

```text
none
low
moderate
high
severe
```

The score is calculated for synthetic grid sectors.

A synthetic grid sector is not an official airspace sector.

The score is not a controller workload score.

---

## 12. Airspace Region Analytics

The regional aggregate is implemented in:

```text
apps/api/internal/airspaceintelligence/airspaceregionanalytics
```

The aggregate publishes:

```text
snapshot count
bucket count
unique aircraft count
aircraft observation count
occupied cell count
sector report count
current aircraft count
peak aircraft per bucket
mean aircraft per bucket
mean complexity score
peak complexity score
airspace pressure index
peak airspace pressure index
moderate sector count
high sector count
severe sector count
contextual risk count
elevated risk count
high risk count
indeterminate risk count
unknown altitude count
temporal coverage
occupancy trend
highest complexity level
```

Supported occupancy trends are:

```text
unavailable
falling
stable
rising
```

The Airspace Pressure Index combines bucket density and mean sector complexity.

It is a research index, not an operational traffic-flow metric.

---

## 13. PostgreSQL Production Composition

The production composition is implemented in:

```text
apps/api/internal/airspaceintelligence/airspaceproduction
```

Production version:

```text
airspace-production-composition-v1
```

The PostgreSQL reader uses:

```text
ingestion_runs
flight_states
```

The query requires:

```text
successful ingestion-run status
observation time inside the bounded query window
latitude inside the requested region
longitude inside the requested region
bounded result count
stable ordering by observation time, ICAO 24-bit address, and state identity
```

Altitude selection order is:

```text
observed geometric altitude
then observed barometric altitude
otherwise unknown altitude
```

The production service builds deterministic minute snapshots.

Each snapshot uses the newest still-fresh observation per aircraft identity.

The production limits are:

```text
default request window: 5 minutes
minimum request window: 1 minute
maximum request window: 1 hour
maximum PostgreSQL observations: 250000
maximum scene observations: 5000
maximum scanner aircraft: 1000
```

No new Stage 11 database table was required.

---

## 14. HTTP Contract

The handler is implemented in:

```text
apps/api/internal/http/handlers/airspace_region_analytics.go
```

The data transfer object is implemented in:

```text
apps/api/internal/http/dto/airspace_region_analytics.go
```

The route is registered in:

```text
apps/api/internal/server/airspace_region_analytics_routes.go
apps/api/internal/server/database_routes.go
```

Successful responses publish:

```text
version
schema version
status
region code
window start
window end
occupancy index
sector complexity reports
regional metrics
confidence
limitations
explanations
scope guard
provenance
generated-at time
```

Verified error contracts include:

```text
HTTP 400 — invalid as-of time
HTTP 400 — invalid window
HTTP 404 — unknown region
HTTP 408 — canceled request mapping
HTTP 422 — observation capacity exceeded
HTTP 500 — production load failure
HTTP 503 — unavailable reader
HTTP 504 — request timeout
```

---

## 15. PostgreSQL and HTTP Runtime Verification

The runtime verifier is implemented in:

```text
apps/api/cmd/verify-postgres-airspace-region-analytics-http-api
```

The command is:

```text
go run ./cmd/verify-postgres-airspace-region-analytics-http-api
```

The deterministic fixture uses a fixed future-neutral analytical clock and does not depend on current live traffic.

Fixture composition:

```text
2 ingestion runs
1 successful ingestion run
1 failed ingestion run
23 stored flight-state rows
22 rows attached to the successful run
1 row attached to the failed run
20 selected in-region, non-future, successful-run observations
4 selected aircraft
5 minute snapshots
5 observations with deliberately unknown altitude
1 future successful-run row that must be excluded
1 out-of-region successful-run row that must be excluded
1 in-region failed-run row that must be excluded
```

The verifier proves:

```text
required schema objects exist
successful ingestion-run boundary works
failed ingestion runs are excluded
future evidence is excluded
out-of-region evidence is excluded
geometric altitude is preferred
unknown altitude is preserved
20 expected observations are selected
five deterministic snapshots are built
four aircraft are present per snapshot
Local Traffic Scene executes
Proximity Scanner executes
Interaction Graph composition executes
Separation Risk Intelligence executes
temporal occupancy executes
sector complexity executes
regional aggregation executes
direct replay fingerprints are identical
HTTP 200 response is valid
HTTP 400 validation errors are stable
HTTP 404 unknown-region response is stable
research-only scope guard survives serialization
synthetic-sector limitation survives serialization
fixture cleanup leaves zero persistent rows
```

Expected terminal completion marker:

```text
Result: PASS
```

Expected cleanup marker:

```text
Persistent verification rows: 0
```

---

## 16. Determinism and Provenance

Stage 11 uses deterministic ordering and fingerprinting at every major boundary.

Deterministic inputs include:

```text
normalized region code
normalized identity
UTC timestamps
sorted aircraft
sorted candidates
sorted graph nodes and edges
sorted occupancy buckets and cells
sorted sector reports
sorted source names
normalized metrics
policy version
schema version
upstream fingerprints
```

The regional result preserves:

```text
scene fingerprints
scan fingerprints
risk fingerprints
source names
latest observed time
regional input fingerprint
```

Equivalent normalized evidence under the same policy produces the same regional fingerprint.

---

## 17. Confidence and Explainability

Confidence is not a decorative field.

Stage 11 confidence depends on:

```text
observation quality
observation freshness
motion plausibility
vertical evidence
scene coverage
radius-decision confidence
temporal proximity
pair evidence completeness
scan confidence
risk-assessment confidence
data quality
temporal coverage
```

Explanations identify how a score was formed.

Limitations identify what the score does not prove.

Unknown altitude reduces confidence and creates explicit limited or indeterminate output instead of silent precision.

---

## 18. Safety and Scope Guards

Stage 11 uses mandatory safety boundaries.

Regional scope guard:

```text
research_only_not_for_operational_separation_or_air_traffic_control_use
```

Separation Risk scope guard:

```text
research_only_not_for_operational_separation_or_collision_avoidance_use
```

Interaction and scene scope guard:

```text
research_only_not_for_operational_separation_use
```

Mandatory regional limitations include:

```text
research_only_not_operational_airspace_management
synthetic_grid_not_official_sectors
historical_complexity_baseline_unavailable
```

The API must preserve these values unchanged.

---

## 19. Known Limitations

The completed foundation still has material limitations.

```text
Open surveillance data may be incomplete or delayed.
Coverage varies by geography, altitude, receiver density, and provider availability.
One-minute buckets may hide faster short-lived changes.
One-degree cells are coarse analytical grid cells.
Three-thousand-meter altitude bands are coarse analytical layers.
The proximity radius is a research policy, not a regulatory minimum.
The separation-risk score is not a safety classification.
The Airspace Pressure Index has no certified operational meaning.
The complexity score has no validated controller-workload calibration.
No official sector boundaries are loaded.
No historical complexity baseline is established.
No prediction of future conflicts is implemented.
No maneuver-intent model is implemented.
No radar or multilateration fusion is implemented.
No frontend visualization is implemented.
```

These limitations do not invalidate Stage 11 completion.

They define the honest boundary of what is complete.

---

## 20. Deferred Work

Deferred work belongs to later stages or the research backlog.

```text
official sector-boundary ingestion
historical complexity baselines
complexity calibration against external operational evidence
forecast stability analysis
confidence propagation across future forecasts
future interaction prediction
unknown intervention modeling
failure explanation standardization
frontend Airspace Intelligence visualization
map heat layers
sector drill-down interface
historical replay interface
```

Stage 12 may consume Stage 11 outputs for stability and explainability work.

Stage 12 must not weaken Stage 11 scope guards.

---

## 21. Implemented Package Inventory

Stage 11 package inventory:

```text
apps/api/internal/airspaceintelligence/interactiongraph
apps/api/internal/airspaceintelligence/interactionradius
apps/api/internal/airspaceintelligence/localtrafficscene
apps/api/internal/airspaceintelligence/proximityscanner
apps/api/internal/airspaceintelligence/separationrisk
apps/api/internal/airspaceintelligence/airspaceregionanalytics
apps/api/internal/airspaceintelligence/airspaceproduction
apps/api/internal/http/dto/airspace_region_analytics.go
apps/api/internal/http/handlers/airspace_region_analytics.go
apps/api/internal/server/airspace_region_analytics_routes.go
apps/api/cmd/verify-postgres-airspace-region-analytics-http-api
```

The production server composition is registered in:

```text
apps/api/internal/server/database_routes.go
```

The generic success-response registration is in:

```text
apps/api/internal/http/response/response.go
```

---

## 22. Completion Evidence

Implementation commits before closure:

```text
ab704df1237ae0db1e575bf0820d66c07d685954
feat: add airborne interaction graph foundation

cffd3d11afdc48a54b93ed528592ca972539054c
feat: add interaction radius policy

8dd318d48c5e5bfc1bc4b31e55e308b0aaa24917
feat: add local traffic scene builder

3f8e5f423fa8fdc77e1ab6bc4681a284d3a65cf8
feat: add proximity scanner and interaction graph composition

91d29d2dd4c559754fc6971b80195892f16314d9
feat: add separation risk intelligence and policy

e5b50788d2f4a5faab60388cbcc7d2de4218f97e
feat: add regional airspace analytics

f98a1563b6fdeed3c2a08686fc42b1d7dfed7823
feat: add airspace analytics production http api
```

Final runtime evidence is produced by:

```text
apps/api/cmd/verify-postgres-airspace-region-analytics-http-api
```

The closure installer requires:

```text
targeted unit tests
race detector
combined Airspace Intelligence regression tests
HTTP and server regression tests
targeted static analysis
PostgreSQL runtime verification
HTTP runtime verification
complete backend regression tests
complete backend static analysis
git diff validation
exact file-delta validation
fixture cleanup validation
```

---

## 23. Formal Completion Statement

Stage 11 is complete.

The completed capability is:

```text
bounded
PostgreSQL-backed
multi-aircraft
region-aware
time-aware
altitude-aware
deterministic
confidence-bearing
explainable
read-only
HTTP-accessible
runtime-verified
research-only
```

The completed Stage 11 foundation can answer:

```text
Which eligible aircraft were present in a bounded regional window?
Which aircraft pairs were close enough for research interaction analysis?
Which pair relationships were converging, parallel, nearby, or diverging?
Which pair assessments were contextual, elevated, high, or indeterminate?
How was airspace occupancy distributed through time and synthetic grid cells?
Which synthetic sectors had higher multidimensional complexity?
What was the region-level Airspace Pressure Index and occupancy trend?
How confident is the result?
Which evidence and limitations produced the result?
```

It cannot answer operational questions such as:

```text
Are two aircraft legally separated?
Will a collision occur?
Should a controller issue an instruction?
Is a synthetic grid cell an official control sector?
Is a complexity score an operational workload measurement?
```

Those claims remain outside the project scope.

Stage 12 — Stability and Explainability may now begin from this completed baseline.
