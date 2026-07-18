# Document 25 — Implementation Sequence

Status: Implementation Baseline v1.6
Project: Global Flight Analytics
Scope: Ordered implementation stages and first coding slice

---

## 1. Purpose

This document defines the order in which Global Flight Analytics must be implemented.

The project now has a large analytical architecture. That does not mean everything should be coded at once. Implementation must follow a strict sequence so the project does not become an unfinished architecture exercise.

---

## 2. Non-Negotiable Implementation Rule

The first implementation goal is a reliable trajectory pipeline.

The project must not start implementation with:

```text
machine learning
weather intelligence
Fréchet distance
Bayesian graph transformers
Sobolev regression
wavelet regression
contrail modeling
fuel prediction
emission prediction
separation risk intelligence
```

The correct first path is:

```text
Open Data Source
↓
Canonical FlightState
↓
Data Quality
↓
Track Builder
↓
TrajectorySegment
↓
FlightTrajectory
↓
API
↓
MapLibre Frontend
```

---

## 3. Stage 1 — Data Foundation

Goal: convert open aircraft data into canonical internal state.

Tasks:

```text
1. OpenSky or compatible provider integration
2. OurAirports import
3. Canonical FlightState model
4. TrackPoint4D model
5. Source metadata model
6. Unit normalization
7. Field normalization
8. Provider response isolation
9. Repository for aircraft states
10. Live aircraft API endpoint
```

Expected result: the backend can fetch aircraft states, normalize them, persist them, and return clean aircraft state data through the API.

---

## 4. Stage 2 — Data Quality Foundation

Goal: prevent bad data from entering analytical outputs.

Tasks:

```text
1. Duplicate point detection
2. Missing field detection
3. Gap detection
4. Jump detection
5. Motion plausibility check
6. Freshness score
7. Field completeness score
8. Sampling density score
9. Track quality score
10. Analytics permission flags
```

Expected result: the backend can explain whether an aircraft state or track is suitable for analytics.

---

## 5. Stage 3 — Trajectory Foundation

Goal: build the primary analytical object: FlightTrajectory.

Tasks:

```text
1. TrajectorySegment model
2. CoverageGap model
3. FlightTrajectory model
4. Track Builder service
5. Conservative Track Splitting
6. Coverage Gap Detector
7. Segment status model
8. Trajectory quality evaluator
9. Repository for trajectory segments
10. Repository for flight trajectories
```

Expected result: the system can build a short trajectory from raw observations and mark its quality, continuity, and gaps.

---

## 6. Stage 4 — Basic Analytics

Goal: add the first analytical value above trajectory construction.

Tasks:

```text
1. Basic airport context
2. Basic route candidate generation
3. Probable origin detection
4. Probable destination detection
5. Basic route confidence score
6. Basic flight phase detection
7. Source limitation explanation
8. Data quality explanation
9. Aircraft detail API endpoint
10. Aircraft trajectory API endpoint
```

Expected result: the frontend can show aircraft position, route context, phase, confidence, and data limitations.

---

## 7. Stage 5 — Frontend MVP

Goal: expose the MVP as a usable product interface.

Tasks:

```text
1. Next.js application setup
2. MapLibre map screen
3. Live aircraft markers
4. Aircraft detail panel
5. Basic trajectory line
6. Track quality indicator
7. Route confidence indicator
8. Data limitation block
9. Airport context display
10. Loading and error states
```

Expected result: a user can open the application, see live aircraft, click an aircraft, and understand its route context and data quality.

---

## 8. Stage 6 — Feature Foundation

Goal: prepare the project for Version 1 analytics.

Tasks:

```text
1. FlightFeatures model
2. Feature Extractor
3. Feature Validator
4. Feature Store
5. Aircraft Feature Provider
6. Temporal Feature Builder
7. Geographical Feature Builder
8. Operational Feature Builder
9. Trajectory Feature Builder
10. Dataset Profiler
```

---

## 9. Stage 7 — Route Intelligence

Tasks:

```text
1. Route Pattern Library
2. Representative Route Profile
3. Phase-Based Route Pattern Engine
4. Route Confidence Score
5. Route Deviation Analyzer
6. Route candidate history
7. Route explanation output
8. Route confidence evaluation
```

---

## 10. Stage 8 — Historical Intelligence

Status: COMPLETED on 2026-07-15.

Completed production scope:

```text
1. Historical contract and validation
2. Historical time-window planner
3. Bounded Historical Read Repository
4. Traffic Historical Intelligence
5. Airport Historical Intelligence
6. Route Historical Intelligence
7. Current versus previous period comparison
8. Historical trajectory similarity baseline
9. Deterministic PostgreSQL aggregate store
10. Historical materialization
11. Bounded historical replay
12. Read-only aggregate HTTP API
13. Production materialization and replay command
14. PostgreSQL, HTTP, idempotency, and production read-back evidence
```

Completion evidence:

```text
docs/31_STAGE_8_HISTORICAL_INTELLIGENCE_COMPLETION.md
```

Scope alignment:

```text
Historical fact computation remains in Stage 8.
Future trajectory continuation, prediction freshness,
low-frequency prediction guards, and estimated time of arrival
belong to Stage 9.
```

A persistent trajectory shape index and predictive continuation are not claimed as completed Stage 8 capabilities.

---

## 11. Stage 9 — Projection and Estimated Time of Arrival

Status: COMPLETED on 2026-07-16.

Completed production scope:

```text
1. Prediction-specific contract and validation
2. Projection horizon policy
3. Conservative short-horizon kinematic baseline
4. Historical neighbor selection
5. Pattern Confidence evaluation
6. Pattern Freshness Guard
7. Low-Frequency Route Guard
8. Local Historical Neighbor Continuation
9. Estimated Arrival baseline
10. Projection replay evaluation
11. Production composition and fallback policy
12. PostgreSQL-backed production read service
13. Read-only Projection Intelligence HTTP API
14. Production server wiring
15. Kinematic fallback PostgreSQL and HTTP runtime evidence
16. Historical continuation PostgreSQL and HTTP runtime evidence
17. Deterministic fingerprints, confidence, limitations, and scope guards
```

Completion evidence:

```text
docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md
```

Completion boundary:

```text
Stage 9 is complete as a bounded, explainable,
research-only Production Projection Intelligence foundation.

It does not claim operational flight prediction,
weather-adjusted forecasting, machine-learning calibration,
airspace conflict prediction, or frontend visualization.
```

Stage 9 consumes the completed Stage 8 historical foundation without changing historical facts into forecast claims.

---

## 12. Stage 10 — Weather Context

Status: COMPLETED on 2026-07-16.

Completed production scope:

```text
1. Canonical Weather Feature Contract and validation
2. Open-Meteo current snapshot adapter
3. Weather Trust Gate and usage scopes
4. Four-Dimensional Weather-Trajectory Alignment
5. Weather Encounter Profile
6. Weather-Adjusted Uncertainty Modifier
7. Weather Context HTTP contract and data transfer object
8. Production PostgreSQL Weather Context composition
9. Bounded production trajectory hydration
10. PostgreSQL weather snapshot selection
11. Production server wiring
12. PostgreSQL and HTTP runtime verification
13. Future trajectory and weather evidence protection
14. Deterministic fingerprints, confidence, limitations, and scope guards
```

Completion evidence:

```text
docs/33_STAGE_10_WEATHER_CONTEXT_COMPLETION.md
```

Completion boundary:

```text
Stage 10 is complete as a bounded, explainable,
research-only Production Weather Context foundation.

It does not claim operational aviation weather support,
flight-level weather from a surface snapshot, causal inference,
weather radar intelligence, turbulence or icing prediction,
or frontend visualization.
```

Stage 10 consumes the completed Stage 9 Projection Intelligence foundation without changing weather context into proof of cause or intent.

---

## 13. Stage 11 — Airspace Intelligence

Status: COMPLETED on 2026-07-17.

Completed production scope:

```text
1. Airborne Interaction Graph Foundation
2. Interaction Radius Policy
3. Local Traffic Scene Builder
4. Multi-Aircraft Proximity Scanner
5. Automatic Interaction Graph composition
6. Separation Risk Intelligence and named risk policy
7. Temporal Airspace Occupancy Index
8. Multidimensional Sector Complexity Score
9. Airspace Region Analytics and Airspace Pressure Index
10. Deterministic confidence, limitations, explanations, and provenance
11. PostgreSQL flight-state observation reader
12. Successful-ingestion-run, future-evidence, regional, and capacity boundaries
13. Production multi-aircraft analytical composition
14. Read-only Airspace Region Analytics HTTP contract and data transfer object
15. Production Fiber server wiring
16. PostgreSQL and HTTP runtime verification with deterministic fixture cleanup
17. Research-only and synthetic-sector scope guards
18. Unit, race, regression, static-analysis, and deterministic replay evidence
```

Completion evidence:

```text
docs/34_STAGE_11_AIRSPACE_INTELLIGENCE_COMPLETION.md
```

Completion boundary:

```text
Stage 11 is complete as a bounded, deterministic, explainable,
research-only Production Airspace Intelligence foundation.

It does not claim operational air traffic control support,
certified aircraft separation monitoring, collision avoidance,
regulatory separation minima, official airspace sectors,
controller workload measurement, or frontend visualization.
```

Stage 11 consumes bounded canonical flight-state evidence without changing proximity, complexity, or separation-risk context into an operational aviation claim.

---

## 14. Stage 12 — Stability and Explainability

Status: COMPLETED on 2026-07-17.

Completed production scope:

```text
1. Immutable deterministic Forecast Versioning
2. Idempotent identical forecast replay
3. Pairwise Decision Stability Evaluator
4. Bounded multi-version Forecast Stability Analysis
5. Explicit confidence dependency graph and Confidence Propagation
6. Weakest required dependency, estimated evidence, and unknown evidence caps
7. Failure Explanation Engine with observed, derived, and unknown-cause separation
8. Unknown Intervention Guard for pilot intent, air traffic control instruction, and exact cause
9. Scope Guard Enforcement for operational, directive, certain, and safety-critical claims
10. Production Stability Intelligence composition over the PostgreSQL Projection Intelligence reader
11. Standardized Stability Intelligence HTTP data transfer object
12. Read-only Fiber HTTP endpoint and database route wiring
13. PostgreSQL and HTTP runtime verification
14. Deterministic replay fingerprint and zero-row fixture cleanup
15. Unit, race, regression, static-analysis, and documentation evidence
```

Completion evidence:

```text
docs/35_STAGE_12_STABILITY_AND_EXPLAINABILITY_COMPLETION.md
```

Completion boundary:

```text
Stage 12 is complete as a bounded, deterministic, explainable,
research-only Production Stability and Explainability foundation.

It does not claim forecast accuracy certification, calibrated probability,
pilot-intent detection, air traffic control instruction reconstruction,
exact operational causation, flight-planning suitability,
safety-critical decision support, or regulated aviation use.
```

Stage 12 evaluates analytical stability and explanation honesty without converting consistency into accuracy, association into causation, or scope enforcement into operational authorization.

---

## 15. Stage 13 — Frontend Analytics Integration

Status: IN PROGRESS from 2026-07-17.

Stage 13 is a proposed continuation after the completed official analytical sequence. It exposes existing backend intelligence without creating a new analytical engine.

Frontend integration slices:

```text
Stage 13.1 — Projection Intelligence Frontend Foundation
selected aircraft → persisted trajectory → Projection Intelligence HTTP API
→ validated TypeScript contract → TanStack Query → projection and arrival panel

Stage 13.2 — Projection Map Visualization
separate estimated GeoJSON source → projected path → forecast points
→ horizontal uncertainty footprints → explicit map legend

Stage 13.3 — Weather Context Frontend Foundation
persisted trajectory → Weather Context HTTP API → validated TypeScript contract
→ TanStack Query → trust, alignment, encounter and uncertainty panel

Stage 13.4 — Stability and Explainability Frontend Foundation
persisted trajectory → bounded as-of forecast history → Stability Intelligence HTTP API
→ validated TypeScript contract → TanStack Query → consistency, confidence,
failure explanation, intervention guard and scope enforcement panel
```

Observed trajectory geometry and estimated projection geometry remain separate MapLibre sources and layers. The interface must never render estimated coordinates as observed flight history.

---

## 16. First Implementation Slice

The first coding slice must be small and complete.

```text
1. OpenSky or compatible provider
2. Canonical FlightState model
3. aircraft_states table
4. data normalization
5. duplicate removal
6. gap detection
7. motion plausibility check
8. trajectory_segments table
9. Track Builder
10. track_quality_score
11. /api/aircraft/live
12. /api/aircraft/{icao24}
13. MapLibre frontend
14. Aircraft detail panel
```

---

## 17. Completion Criteria for MVP Code

The MVP is complete when:

```text
backend starts locally
frontend starts locally
live aircraft data is visible on the map
aircraft detail panel works
trajectory segment is built from recent points
track quality is calculated
data limitations are displayed
basic route candidate is shown when possible
```

---

## 18. Final Implementation Statement

The implementation order is:

```text
Data Foundation
↓
Data Quality
↓
Trajectory Foundation
↓
Basic Analytics
↓
Frontend MVP
↓
Feature Foundation
↓
Route Intelligence
↓
Historical Intelligence
↓
Projection and ETA
↓
Weather Context
↓
Airspace Intelligence
↓
Stability and Explainability
```

<!-- SOURCE-CONSTRAINTS-OPENSKY-V1 -->
## Architecture Amendment — Source Constraints and OpenSky Contract Foundation

Status: IMPLEMENTED as a non-activating architecture and analytical-core foundation.

Implemented scope:

```text
1. Immutable free-source-only project constraints
2. Explicit absence of first-party receiver infrastructure
3. Explicit absence of satellite and commercial aviation-data access
4. Executable allowed, limited, and blocked capability decisions
5. OpenSky OAuth2 Client Credentials foundation
6. Anonymous reduced-capability fallback
7. Regional bounding-box request contract and credit-cost estimate
8. Complete eighteen-field OpenSky State Vector parser
9. Position-source and aircraft-category semantics
10. Rate-limit evidence contract
11. Bounded historical-flight contract
12. Estimated-airport disclosure contract
13. Experimental-track disclosure and retention guard
14. Verification command and unit coverage
```

This amendment does not activate OpenSky in the production ingestion path. Activation requires separate provider-policy, persistence, migration, runtime, and frontend-disclosure verification.

<!-- OPENSKY-VALIDITY-ATTRIBUTION-V1 -->
## OpenSky Activation Prerequisites Amendment

Before OpenSky can become an active ingestion provider, implementation must verify:

```text
fifteen-second position-validity enforcement
snapshot-time and field-time persistence
stale and missing position rejection
required public attribution
non-commercial research-use disclosure
large-cloud-IP connectivity behavior
provider health, budget, and fallback execution
```

OpenSky activation remains a separate increment and must not be coupled to frontend visual work.

<!-- OPENSKY-PRODUCTION-PROVIDER-V1 -->
## OpenSky Production Provider Selection Increment

Status: IMPLEMENTED PENDING RUNTIME INSTALLATION EVIDENCE.

The production ingestion command can select `airplanes.live` or OpenSky while reusing the existing provider policy, budget, request coalescing, provider health, canonical state, data quality, trajectory, and PostgreSQL path. Automatic cross-provider fallback remains a separate increment because ingestion-run provenance must identify the provider that actually served the request.

<!-- TRAFFIC-PROVIDER-AUTOMATIC-FALLBACK-V1 -->
## Free traffic provider automatic fallback

Status: Completed in implementation baseline.

The production ingestion daemon now supports the explicit `auto` provider mode.
The ordered chain is `airplanes.live` followed by OpenSky. Fallback is limited
to budget denial, rate limiting, provider cooldown, server failure, and network
unavailability. Actual selected-provider provenance is preserved in ingestion
runs, canonical states, health evidence, and fallback decision evidence.

<!-- OPENSKY-REST-COMPATIBILITY-V1 -->
## OpenSky REST compatibility hardening

The production provider requests extended aircraft categories, while the State Vector parser remains compatible with base responses that omit the optional category field.

<!-- OPEN-AVIATION-RESEARCH-EVIDENCE-V1-2:IMPLEMENTATION -->

## Open Aviation Research Evidence Foundation

Status: COMPLETED when the implementation verifier, focused tests, full backend tests, migration sequence checks, and PostgreSQL migration application evidence pass.

This foundation preserves transponder and provider observation metadata, adds bounded Transponder Alert Evidence, creates executable research dataset manifests, and defines offline-only benchmark plans. It does not download datasets, train models, create operational alerts, or introduce satellite ADS-C evidence.

<!-- STAGE-14-1-ARCHITECTURE-CONSOLIDATION-V1-1:IMPLEMENTATION -->

## Stage 14 — Architecture Consolidation and Dead Code Elimination

Status: IN PROGRESS from 2026-07-18.

Stage 14 begins with a backend feature freeze. The first slice centralizes the ordinal confidence vocabulary, closes the confirmed trajectory contract drift, adds production reachability evidence, and strengthens dependency and race analysis. Package deletion follows only after factual classification of every non-runtime-reachable package.

<!-- STAGE-14-2-DEAD-CODE-CLASSIFICATION:IMPLEMENTATION -->

### Stage 14.2 — Dead Code Classification and Removal

Status: IN PROGRESS from 2026-07-18.

The first removal slice deletes two confirmed obsolete Analytical Core foundation packages and establishes mandatory classification for every remaining non-runtime analytical package. Airport Intelligence and the feature pipeline must be integrated or removed before release.

<!-- STAGE-14-3-AIRPORT-INTELLIGENCE-PRODUCTION:IMPLEMENTATION -->

### Stage 14.3 — Airport Intelligence Production Integration

Status: IMPLEMENTED.

The six Airport Intelligence domain packages are composed into the production server through PostgreSQL-backed read-only endpoints. Strict reachability audit must report the complete context as runtime-connected.

<!-- STAGE-14-4-FEATURE-MATERIALIZATION:IMPLEMENTATION -->

### Stage 14.4 — Feature Materialization and Profiler Removal

Status: IMPLEMENTED.

The complete remaining Feature Pipeline is reachable through an operational PostgreSQL materializer. The disconnected dataset profiler was deleted after importer proof because it had no bounded persisted-dataset read path.
