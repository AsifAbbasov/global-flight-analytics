# Documentation Index — Global Flight Analytics

Status: Documentation Index v1.2  
Project: Global Flight Analytics

---

## Purpose

This index records the documentation structure for Global Flight Analytics.

The project documentation is divided into two groups:

```text
Documents 01–21: existing product, system, data, architecture foundation, and engineering amendments
Documents 22–28: research audit deduplication, analytical core, roadmap, implementation sequence, research scope guards, engineering principles, and research-to-implementation decision method
```

---

## Existing Foundation Documents

The existing documentation foundation is retained. The new analytical core documents do not replace the earlier product and system architecture work. They extend it.

```text
01_PRODUCT_VISION.md
02_SYSTEM_ARCHITECTURE.md
03_DOMAIN_MODEL.md
04_DATABASE_DESIGN.md
05_DATA_SOURCES.md
06_DATA_COLLECTION_PIPELINE.md
07_ROUTE_DETECTION_ENGINE.md
08_AIRPORT_INTELLIGENCE_MODULE.md
09_TRAFFIC_ANALYTICS_MODULE.md
10_API_SPECIFICATION.md
11_FRONTEND_ARCHITECTURE.md
12_INFRASTRUCTURE_AND_DEPLOYMENT.md
13_SECURITY_SPECIFICATION.md
14_PERFORMANCE_AND_SCALABILITY.md
15_DEVELOPMENT_ROADMAP.md
16_MVP_SCOPE.md
17_FUTURE_VERSIONS.md
18_TECHNICAL_DECISIONS_RECORD.md
19_RISK_ANALYSIS.md
20_FINAL_ARCHITECTURE_BLUEPRINT.md
21_ENGINEERING_AMENDMENTS_v1.1.md
```

---

## New Analytical Architecture Documents

### Document 22 — Research Audit Deduplication

Path:

```text
docs/22_RESEARCH_AUDIT_DEDUPLICATION.md
```

Purpose:

```text
Consolidates all research audit outputs into deduplicated architecture layers,
removes repeated module names, and defines the final accepted architecture ideas.
```

### Document 23 — Analytical Core Architecture

Path:

```text
docs/23_ANALYTICAL_CORE_ARCHITECTURE.md
```

Purpose:

```text
Defines the analytical core of Global Flight Analytics:
Trajectory Intelligence, Route Intelligence, Historical Similarity,
Historical Patterns, Weather-Aware Intelligence, Projection,
Multi-Aircraft Context, Airspace Interaction, Airport Intelligence,
and Confidence and Explainability.
```

### Document 24 — MVP and Version Roadmap

Path:

```text
docs/24_MVP_VERSION_ROADMAP.md
```

Purpose:

```text
Defines MVP, Version 1, Version 2, release boundaries,
capabilities, tables, frontend scope, and success criteria.
```

### Document 25 — Implementation Sequence

Path:

```text
docs/25_IMPLEMENTATION_SEQUENCE.md
```

Purpose:

```text
Defines the exact implementation order from data foundation to advanced analytics,
including the first coding slice.
```

### Document 26 — Research Backlog and Scope Guards

Path:

```text
docs/26_RESEARCH_BACKLOG_AND_SCOPE_GUARDS.md
```

Purpose:

```text
Defines deferred research topics, MVP forbidden scope,
version promotion rules, prediction scope guards,
weather scope guards, and open-data limitations.
```

### Document 27 — Engineering Principles

Path:

```text
docs/27_ENGINEERING_PRINCIPLES.md
```

Purpose:

```text
Defines the project engineering rules for simple-first implementation,
controlled complexity, magic number avoidance, analytical policy visibility,
unit testing, smoke testing, and documentation alignment.
```

### Document 28 — Research and Analytical Decision Method

Path:

```text
docs/28_RESEARCH_AND_ANALYTICAL_DECISION_METHOD.md
```

Purpose:

```text
Defines the mandatory research-to-code decision method,
the three hard constraints, decision classification labels,
open research expansion rules, physics and mathematics rules,
baseline-first analytics, threshold derivation, historical replay,
metrics, confidence, limitations, and scope protection.
```

---

## Superseded Duplicate Notice

The file below is superseded and must not be used as the active baseline:

```text
docs/21_RESEARCH_AUDIT_DEDUPLICATION.md
```

It was created with the wrong number before the existing local document `21_ENGINEERING_AMENDMENTS_v1.1.md` was accounted for. The active replacement is:

```text
docs/22_RESEARCH_AUDIT_DEDUPLICATION.md
```

---

## Current Architecture Baseline

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

## Mandatory Research-to-Code Rule

Research ideas do not automatically become implementation scope.

Every non-trivial analytical proposal must follow:

```text
current documentation baseline
↓
research digest
↓
three hard constraints
↓
free data availability
↓
source-versus-hypothesis separation
↓
simplest measurable baseline
↓
tests
↓
historical replay when applicable
↓
metrics
↓
confidence and limitations
↓
only then additional complexity
```

The authoritative method is defined in:

```text
docs/28_RESEARCH_AND_ANALYTICAL_DECISION_METHOD.md
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
physics-informed reasoning where inputs support it,
confidence, explainability, and visualization.
```
