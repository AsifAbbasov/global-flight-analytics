# Document 23 — Analytical Core Architecture

Status: Architecture Baseline v1.1  
Project: Global Flight Analytics  
Scope: Analytical core, domain boundaries, architecture layers, and module map

---

## 1. Purpose

This document defines the analytical core of Global Flight Analytics.

The analytical core is the part of the platform that transforms clean trajectory data into explainable aviation intelligence.

It is separate from the infrastructure architecture. Infrastructure moves, stores, normalizes, and serves data. The analytical core interprets data.

---

## 2. Project Definition

Global Flight Analytics is an open-data aviation research and analytics platform.

It is not:

```text
a flight tracker clone
regulated aviation software
a flight planning system
a commercial aviation data platform
a regulated operational decision product
```

It is:

```text
a trajectory-centered aviation analytics platform
a map-based research interface
a system for open-data route intelligence
a system for trajectory quality and confidence analysis
a system for historical pattern analytics
```

---

## 3. Core Principle

```text
Raw aircraft positions are not the analytical object.
The analytical object is the reconstructed, quality-scored, source-aware flight trajectory.
```

Therefore:

```text
FlightState = atomic observation
TrackPoint4D = normalized movement point
TrajectorySegment = meaningful movement interval
FlightTrajectory = primary analytical object
FlightFeatures = extracted analytical features
AnalyticalResult = output of the analytical core
ConfidenceReport = explanation of certainty and limitations
```

---

## 4. High-Level Analytical Pipeline

```text
Open Data Sources
↓
Source Adapters
↓
Canonical Flight State
↓
Data Quality and Provenance Layer
↓
Track Builder
↓
Trajectory Segment
↓
Flight Trajectory
↓
Feature Engineering Layer
↓
Context Enrichment Layer
↓
Analytical Core
↓
Confidence and Explainability Layer
↓
API
↓
Frontend
```

---

## 5. Analytical Core Map

```text
Analytical Core

1. Trajectory Intelligence Core
2. Route Intelligence
3. Historical Trajectory Similarity Engine
4. Historical Pattern-Based Trajectory Intelligence
5. Weather-Aware Trajectory Intelligence
6. Estimated Time of Arrival and Projection Intelligence
7. Multi-Aircraft Context Intelligence
8. Separation Risk and Airspace Interaction Intelligence
9. Airport and Region Intelligence
10. Confidence and Explainability Engine
```

---

## 6. Trajectory Intelligence Core

Purpose: build and evaluate the primary analytical object — the flight trajectory.

Components:

```text
FlightTrajectory
TrajectorySegment
CoverageGap
Resampled Trajectory
Trajectory Quality
Segment Status Model
Source Provenance
```

Responsibilities:

```text
build movement history from normalized points
split tracks conservatively when continuity is broken
mark observed, interpolated, estimated, and invalid segments
detect coverage gaps
score trajectory quality
prevent weak trajectories from producing strong analytics
```

MVP status: required in MVP.

---

## 7. Route Intelligence

Purpose: infer probable route context from open observed data.

Components:

```text
Probable Origin Detection
Probable Destination Detection
Route Confidence Score
Phase-Based Route Pattern
Representative Route Profile
Route Deviation Analyzer
```

Responsibilities:

```text
infer likely origin airport
infer likely destination airport
classify route confidence
compare trajectory against known route patterns
detect deviation from expected corridor
explain route uncertainty
```

MVP status: basic route inference is required in MVP. Phase-based route patterns and deviation analysis move to Version 1.

---

## 8. Historical Trajectory Similarity Engine

Purpose: find historical trajectories with similar shape, endpoints, corridor behavior, and movement profile.

Components:

```text
Route Shape Similarity Score
Trajectory Endpoint Filter
Trajectory Corridor Filter
Discrete Fréchet Similarity Filter
Spatial Index
Similarity Threshold Policy
Similarity Candidate Decision States
```

MVP status: not required in MVP. Basic historical matching starts in Version 1. Advanced similarity, spatial indexing, and Fréchet-style filtering move to Version 2.

---

## 9. Historical Pattern-Based Trajectory Intelligence

Purpose: use historical observed trajectories to understand and cautiously project current movement.

Components:

```text
Historical Route Pattern Library
Trajectory Shape Object
Time-Normalized Trajectory Representation
Similar Historical Trajectory Selector
Local Neighbor-Based Continuation Baseline
Pattern Confidence Engine
Pattern Freshness Guard
Low-Frequency Route Failure Guard
```

Responsibilities:

```text
store historical route patterns
normalize trajectory shape for comparison
find similar historical trajectories
build probable continuation corridors
score pattern confidence
reject forecasts on low-frequency routes
mark stale historical patterns
```

MVP status: not fully available on day one because historical data must be accumulated first.

---

## 10. Weather-Aware Trajectory Intelligence

Purpose: use weather as context for uncertainty, not as proof of pilot intent or maneuver cause.

Components:

```text
Weather Feature Contract
Weather Trust Gate
Four-Dimensional Weather-Trajectory Alignment
Weather Encounter Profile
Weather-Adjusted Uncertainty Modifier
```

MVP status: not required in MVP. Compact version starts in Version 1 or Version 2 depending on available free weather source quality.

---

## 11. Estimated Time of Arrival and Projection Intelligence

Purpose: provide conservative short-horizon projections and explain their uncertainty.

Components:

```text
Short-Horizon Projection Baseline
Estimated Time of Arrival Feature Set
Estimated Time of Arrival Confidence Score
Estimated Time of Arrival Evolution Analyzer
Projection Horizon Policy
Probabilistic Projection Output
```

MVP status: not required in the first MVP. Short-horizon baseline belongs to Version 1.

---

## 12. Multi-Aircraft Context Intelligence

Purpose: analyze aircraft not only as isolated tracks, but as part of a local traffic scene.

Components:

```text
Airborne Interaction Graph
Interaction Radius Policy
Local Traffic Scene Builder
Arrival and Departure Complexity Split
Interaction-Aware Projection Context
```

MVP status: Version 2.

---

## 13. Separation Risk and Airspace Interaction Intelligence

Purpose: analyze proximity, occupancy, and airspace complexity for research visualization.

Components:

```text
Separation Risk Intelligence
Separation Envelope Engine
Pairwise Separation Risk Detector
Temporal Airspace Occupancy Index
Multi-Aircraft Proximity Scanner
Sector Complexity Score
What-If Separation Analysis
```

Scope guard: this module must never be presented as regulated operational aviation software or certified safety logic.

MVP status: Version 2 or later.

---

## 14. Airport and Region Intelligence

Purpose: connect trajectories to airports, regions, infrastructure, and traffic statistics.

Components:

```text
Airport Digital Passport
Airport Traffic Statistics
Popular Route Detection
Region Crossing Detector
Airspace Region Analytics
Airport Congestion Score
```

MVP status: basic airport context is required in MVP. Airport and region analytics expand in Version 1 and Version 2.

---

## 15. Confidence and Explainability Engine

Purpose: make every analytical result honest, inspectable, and limited by data quality.

Components:

```text
Confidence Propagation
Source Limitation Explanation
Data Quality Explanation
Low-Confidence Reasoning
Forecast Stability Analysis
Decision Stability Analysis
Unknown Cause Protection
```

MVP status: a basic version is required in MVP. Full confidence propagation and stability analysis move to Version 1 and Version 2.

---

## 16. Mandatory Analytical Output Contract

Every analytical output must include:

```text
result_type
result_value
confidence_level
confidence_score
input_data_quality
source_limitations
explanation
scope_guard
created_at
```

No analytical module may return a strong conclusion without confidence and explanation.

---

## 17. Scope Guards

The platform must always expose its limitations:

```text
No official flight plan access
No operational instruction access
No commercial aviation feed access
No guaranteed global coverage
No regulated aviation use
```

---

## 18. Final Statement

The analytical core turns Global Flight Analytics from a map application into a research-grade portfolio platform.

The core must be implemented gradually:

```text
MVP: reliable trajectory and basic route intelligence
Version 1: features, historical patterns, replay, and confidence
Version 2: airspace complexity, multi-aircraft context, stability, and advanced similarity
Research Backlog: heavy models, satellites, FLARM, emissions, contrails, and regulated operational systems
```
