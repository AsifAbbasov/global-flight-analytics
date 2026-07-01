# Documentation Index — Global Flight Analytics

Status: Documentation Index v1.1  
Project: Global Flight Analytics

---

## Purpose

This index records the documentation structure for Global Flight Analytics.

The project documentation is divided into two groups:

```text
Documents 01–20: existing product, system, data, and architecture foundation
Documents 21–25: research audit deduplication, analytical core, roadmap, implementation sequence, and research scope guards
```

---

## Existing Foundation Documents

The existing documentation foundation is retained. The new analytical core documents do not replace the earlier product and system architecture work. They extend it.

```text
Documents 01–20
Existing product vision, system architecture, data source planning,
domain model, database planning, implementation planning,
and final architecture blueprint.
```

If any existing document conflicts with Documents 21–25, the newer analytical core documents should be treated as the updated architecture baseline for analytical modules, MVP scope, version planning, and implementation order.

---

## New Analytical Architecture Documents

### Document 21 — Research Audit Deduplication

Path:

```text
docs/21_RESEARCH_AUDIT_DEDUPLICATION.md
```

Purpose:

```text
Consolidates all research audit outputs into deduplicated architecture layers,
removes repeated module names, and defines the final accepted architecture ideas.
```

### Document 22 — Analytical Core Architecture

Path:

```text
docs/22_ANALYTICAL_CORE_ARCHITECTURE.md
```

Purpose:

```text
Defines the analytical core of Global Flight Analytics:
Trajectory Intelligence, Route Intelligence, Historical Similarity,
Historical Patterns, Weather-Aware Intelligence, Projection,
Multi-Aircraft Context, Airspace Interaction, Airport Intelligence,
and Confidence and Explainability.
```

### Document 23 — MVP and Version Roadmap

Path:

```text
docs/23_MVP_VERSION_ROADMAP.md
```

Purpose:

```text
Defines MVP, Version 1, Version 2, release boundaries,
capabilities, tables, frontend scope, and success criteria.
```

### Document 24 — Implementation Sequence

Path:

```text
docs/24_IMPLEMENTATION_SEQUENCE.md
```

Purpose:

```text
Defines the exact implementation order from data foundation to advanced analytics,
including the first coding slice.
```

### Document 25 — Research Backlog and Scope Guards

Path:

```text
docs/25_RESEARCH_BACKLOG_AND_SCOPE_GUARDS.md
```

Purpose:

```text
Defines deferred research topics, MVP forbidden scope,
version promotion rules, prediction scope guards,
weather scope guards, and open-data limitations.
```

---

## Current Architecture Baseline

The active architecture baseline is:

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

## Current MVP Baseline

The current MVP baseline is:

```text
OpenSky or compatible aircraft ingestion
OurAirports import
Canonical FlightState
TrackPoint4D
Data normalization
Duplicate removal
Gap and jump detection
Motion plausibility check
Track Builder
TrajectorySegment
FlightTrajectory
CoverageGap
Track Quality Score
Basic Airport Context
Basic Route Intelligence
Basic Flight Phase Detection
MapLibre frontend
Aircraft Detail Panel
Data Quality Explanation
Source Limitation Guard
```

---

## First Implementation Slice

The first coding slice is:

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

## Documentation Rule

New architecture changes must not silently overwrite the earlier documents.

Future changes should be added as:

```text
new numbered documents
or explicit amendments
or clearly marked updates to the relevant existing document
```

---

## Final Documentation Statement

Global Flight Analytics documentation now recognizes the project as:

```text
an open-data aviation research and analytics platform
centered on trajectory quality, feature engineering,
historical patterns, context-aware analytics,
confidence, explainability, and visualization.
```
