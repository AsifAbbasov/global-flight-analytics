# Global Flight Analytics

Global Flight Analytics is an open-data aviation research and analytics platform.

The project is not a flight tracker clone, not an air traffic control system, not a flight planning system, and not a commercial aviation data platform.

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
docs/21_RESEARCH_AUDIT_DEDUPLICATION.md
docs/22_ANALYTICAL_CORE_ARCHITECTURE.md
docs/23_MVP_VERSION_ROADMAP.md
docs/24_IMPLEMENTATION_SEQUENCE.md
docs/25_RESEARCH_BACKLOG_AND_SCOPE_GUARDS.md
```

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
