# Document 24 — MVP and Version Roadmap

Status: Architecture Baseline v1.1  
Project: Global Flight Analytics  
Scope: MVP, Version 1, Version 2, and release boundaries

---

## 1. Purpose

This document defines what Global Flight Analytics must build first, what belongs to later versions, and what must not enter the MVP.

The goal is to prevent architectural overreach. The project is technically ambitious, but the first implementation must be a focused vertical slice.

---

## 2. Product Positioning

Global Flight Analytics is an open-data aviation research and analytics platform.

It is not a commercial flight tracker clone. It is not regulated aviation software. It is not a flight planning system.

The platform should demonstrate:

```text
open aviation data ingestion
trajectory construction
data quality evaluation
route intelligence
historical pattern analysis
confidence-aware analytics
explainable map visualization
```

---

## 3. MVP Goal

```text
Build a working open-data aviation map with real-time aircraft states,
basic trajectory construction, data quality evaluation,
basic route intelligence, and clear limitation explanations.
```

The MVP must prove that the platform can reliably move from raw open data to a quality-scored trajectory and a useful frontend view.

---

## 4. MVP Scope

The MVP includes only the following items:

```text
1. OpenSky or compatible live aircraft ingestion
2. OurAirports import
3. Canonical FlightState
4. TrackPoint4D
5. Raw Source Isolation
6. Unit and Field Normalization
7. Duplicate Point Removal
8. Gap and Jump Detection
9. Motion Plausibility Check
10. Track Builder
11. TrajectorySegment
12. FlightTrajectory
13. CoverageGap
14. Track Quality Score
15. Segment Status Model
16. Analytics Permission Flags
17. Basic Airport Context
18. Basic Route Intelligence
19. Basic Flight Phase Detection
20. Live Map with MapLibre
21. Aircraft Detail Panel
22. Source Limitation Guard
23. Data Quality Explanation
```

---

## 5. MVP Capabilities

The MVP must be able to:

```text
show aircraft on a live map
fetch live open aircraft data through the backend
normalize raw provider data into Canonical FlightState
build short aircraft tracks
detect gaps in data
detect unrealistic movement jumps
remove duplicate points
score basic track quality
show probable origin when enough data exists
show probable destination when enough data exists
show a basic flight phase
show why confidence is high, medium, or low
show data limitations instead of pretending to know more than the data allows
```

---

## 6. MVP Non-Goals

The MVP must not include:

```text
CNN-LSTM
Sobolev regression
wavelet regression
Bayesian spatio-temporal graph transformer
full Fréchet similarity
advanced weather grid analytics
ADS-C fusion
FLARM ingestion
contrail optimization
fuel prediction
emission prediction
regulated operational aviation modules
commercial aviation data
satellite data
```

The MVP must not begin with machine learning or advanced prediction. It must begin with a reliable trajectory pipeline.

---

## 7. MVP Tables

```text
aircraft_states
trajectory_segments
flight_trajectories
coverage_gaps
airports
aircraft_metadata
route_candidates
data_quality_reports
```

---

## 8. MVP Backend Packages

```text
apps/api/internal/sources/opensky
apps/api/internal/sources/ourairports
apps/api/internal/normalization
apps/api/internal/quality
apps/api/internal/tracks
apps/api/internal/trajectory
apps/api/internal/airports
apps/api/internal/routes
apps/api/internal/analytics
apps/api/internal/api
```

Existing packages may be reused and renamed gradually. The architecture should not be rewritten just for naming.

---

## 9. MVP Frontend Scope

```text
MapLibre map
live aircraft markers
aircraft detail panel
basic trajectory line
track quality indicator
route confidence indicator
data limitation block
airport context panel
```

The frontend must make data quality visible. It must not only draw aircraft positions.

---

## 10. Version 1 Goal

Version 1 turns the MVP from a reliable map into a first analytical platform.

```text
Add feature engineering, historical route patterns, basic replay evaluation,
route deviation, short-horizon projection, compact weather context,
and stronger confidence explanations.
```

---

## 11. Version 1 Scope

```text
1. Feature Engineering Layer
2. Feature Store
3. Aircraft Feature Provider
4. Dataset Profiler
5. Route Intelligence
6. Phase-Based Route Pattern Engine
7. Representative Route Profile
8. Route Deviation Analyzer
9. Historical Route Pattern Library
10. Historical Similar Trajectory Selector
11. Basic Historical Pattern-Based Continuation
12. Replay Engine
13. Evaluation Metrics
14. Compact Weather-Aware Trajectory Intelligence
15. Short-Horizon Projection Baseline
16. Estimated Time of Arrival Confidence Score
17. Confidence and Explainability Engine
```

---

## 12. Version 1 Capabilities

```text
build features from trajectories
store validated flight features
profile dataset quality
compare current tracks with historical patterns
find similar observed historical routes
show route confidence
show route deviation
produce a short projection corridor
run replay evaluation
explain confidence and low-confidence failures
adjust uncertainty using compact weather context
```

---

## 13. Version 1 Tables

```text
flight_features
aircraft_features
route_patterns
historical_shape_index
similarity_candidates
projection_results
replay_runs
evaluation_metrics
weather_contexts
confidence_reports
```

---

## 14. Version 2 Goal

Version 2 turns the platform into a deeper research-grade analytical system.

```text
Add advanced historical similarity, local traffic scene analysis,
airspace interaction intelligence, forecast stability, decision stability,
and region-level analytical dashboards.
```

---

## 15. Version 2 Scope

```text
1. Historical Trajectory Similarity Engine
2. Discrete Fréchet Similarity Filter
3. Trajectory Similarity Spatial Index
4. Similarity Threshold Policy
5. Multi-Aircraft Context Intelligence
6. Airborne Interaction Graph
7. Local Traffic Scene Builder
8. Separation Risk Intelligence
9. Sector Complexity Score
10. Temporal Airspace Occupancy Index
11. Weather Grid Context
12. Forecast Versioning
13. Forecast Stability Analysis
14. Decision Stability Evaluator
15. Airspace Region Analytics
16. Airport Congestion Score
17. Estimated Time of Arrival Evolution Analyzer
18. Unknown Intervention Guard
```

---

## 16. Version 2 Capabilities

```text
search historical trajectories efficiently
use advanced shape similarity
build local traffic scenes around an aircraft
measure airspace density
measure multi-aircraft proximity
estimate sector complexity
show airspace pressure
compare forecast versions
evaluate forecast stability
detect context shift without inventing the cause
build airport and region analytical dashboards
```

---

## 17. Version 2 Tables

```text
airspace_sectors
traffic_scenes
separation_events
forecast_versions
decision_stability_reports
airspace_complexity_reports
advanced_similarity_results
region_analytics_snapshots
airport_congestion_reports
```

---

## 18. Release Boundary Rules

```text
No advanced analytics before reliable trajectory construction.
No historical intelligence before historical data exists.
No projection claims before replay evaluation exists.
No weather-based explanation without weather trust scoring.
No strong confidence without data quality evidence.
No regulated aviation claims anywhere in the product.
```

---

## 19. Final Roadmap Statement

```text
MVP = reliable trajectory and basic route intelligence
Version 1 = features, historical patterns, replay, projection, confidence
Version 2 = advanced similarity, airspace intelligence, multi-aircraft context, stability
Research Backlog = heavy models, satellite fusion, climate models, regulated operational systems
```
