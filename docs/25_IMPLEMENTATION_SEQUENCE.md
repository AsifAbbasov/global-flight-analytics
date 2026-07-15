# Document 25 — Implementation Sequence

Status: Implementation Baseline v1.2
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

Tasks:

```text
1. Prediction-specific contract and result model
2. Local Neighbor-Based Continuation Baseline
3. Short-Horizon Projection Baseline
4. Projection Horizon Policy
5. Estimated Time of Arrival Feature Set
6. Estimated Time of Arrival Confidence Score
7. Pattern Freshness Guard for prediction
8. Low-Frequency Route Failure Guard for prediction
9. Projection replay and evaluation metrics
10. Projection explanation and API output
```

Stage 9 must consume the completed Stage 8 historical foundation without changing historical facts into forecast claims.

---

## 12. Stage 10 — Weather Context

Tasks:

```text
1. Weather Feature Contract
2. Weather Provider Adapter
3. Weather Trust Gate
4. Four-Dimensional Weather-Trajectory Alignment
5. Weather Encounter Profile
6. Weather-Adjusted Uncertainty Modifier
7. Weather context API output
```

---

## 13. Stage 11 — Airspace Intelligence

Tasks:

```text
1. Interaction Graph
2. Local Traffic Scene Builder
3. Interaction Radius Policy
4. Multi-Aircraft Proximity Scanner
5. Separation Risk Intelligence
6. Sector Complexity Score
7. Temporal Airspace Occupancy Index
8. Airspace Region Analytics
```

---

## 14. Stage 12 — Stability and Explainability

Tasks:

```text
1. Forecast Versioning
2. Forecast Stability Analysis
3. Decision Stability Evaluator
4. Confidence Propagation
5. Failure Explanation Engine
6. Unknown Intervention Guard
7. Scope Guard enforcement
8. Explanation API standardization
```

---

## 15. First Implementation Slice

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

## 16. Completion Criteria for MVP Code

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

## 17. Final Implementation Statement

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
