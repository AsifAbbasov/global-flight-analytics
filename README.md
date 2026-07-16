# Global Flight Analytics

Global Flight Analytics is an open-data aviation research and analytics platform.

The project is not a flight tracker clone, not regulated aviation software, not a flight planning system, and not a commercial aviation data platform.

The platform is centered on trajectory quality, feature engineering, historical patterns, context-aware analytics, confidence, explainability, and map-based visualization.

## Architecture Baseline

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

## Analytical Core

```text
Trajectory Intelligence Core
Route Intelligence
Historical Trajectory Similarity Engine
Historical Pattern-Based Trajectory Intelligence
Weather-Aware Trajectory Intelligence
Estimated Time of Arrival and Projection Intelligence
Multi-Aircraft Context Intelligence
Separation Risk and Airspace Interaction Intelligence
Airport and Region Intelligence
Confidence and Explainability Engine
```

## MVP Focus

The first implementation focuses on a reliable trajectory pipeline:

```text
OpenSky or compatible provider
Canonical FlightState
Data Quality
Track Builder
TrajectorySegment
FlightTrajectory
API
MapLibre frontend
Aircraft detail panel
```

Advanced prediction, machine learning, satellite fusion, FLARM, fuel analytics, emission analytics, and climate routing are deferred to later versions or the research backlog.

## Documentation

The current documentation baseline is in:

```text
docs/DOCUMENT_INDEX.md
docs/22_RESEARCH_AUDIT_DEDUPLICATION.md
docs/23_ANALYTICAL_CORE_ARCHITECTURE.md
docs/24_MVP_VERSION_ROADMAP.md
docs/25_IMPLEMENTATION_SEQUENCE.md
docs/26_RESEARCH_BACKLOG_AND_SCOPE_GUARDS.md
docs/27_ENGINEERING_PRINCIPLES.md
docs/28_RESEARCH_AND_ANALYTICAL_DECISION_METHOD.md
docs/29_REPRODUCIBLE_DOCKER.md
docs/30_AIRPORT_INTELLIGENCE_IMPLEMENTATION_ALIGNMENT.md
docs/31_STAGE_8_HISTORICAL_INTELLIGENCE_COMPLETION.md
docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md
docs/33_STAGE_10_WEATHER_CONTEXT_COMPLETION.md
```

Existing foundation documents remain in `docs/01_*` through `docs/21_ENGINEERING_AMENDMENTS_v1.1.md`.

## First Coding Slice

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
