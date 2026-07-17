# Documentation Index — Global Flight Analytics

Status: Documentation Index v1.9
Project: Global Flight Analytics

---

## Purpose

This index records the documentation structure for Global Flight Analytics.

The project documentation is divided into two groups:

```text
Documents 01–21: existing product, system, data, architecture foundation, and engineering amendments
Documents 22–35: research audit, analytical architecture, roadmap, engineering rules, decision method, container operations, implementation alignment, and completion evidence
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
including the first coding slice and formal completion boundaries for implemented stages.
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

### Document 29 — Reproducible Docker

Path:

```text
docs/29_REPRODUCIBLE_DOCKER.md
```

Purpose:

```text
Defines the pinned container build, scratch runtime,
non-root execution, healthcheck, local PostgreSQL Compose environment,
migration startup order, and continuous integration verification contract.
```

---

### Document 30 — Airport Intelligence Implementation Alignment

Path:

```text
docs/30_AIRPORT_INTELLIGENCE_IMPLEMENTATION_ALIGNMENT.md
```

Purpose:

```text
Records the implemented Airport Intelligence domain contracts,
the corrected Activity Score and Data Confidence separation,
historical and trends baselines, limitations, and next integration steps.
```

### Document 31 — Stage 8 Historical Intelligence Completion

Path:

```text
docs/31_STAGE_8_HISTORICAL_INTELLIGENCE_COMPLETION.md
```

Purpose:

```text
Records the completed production Historical Intelligence foundation,
scope alignment, acceptance matrix, PostgreSQL and HTTP runtime evidence,
production materialization and replay idempotency, known limitations,
deferred prediction work, and the formal Stage 8 completion statement.
```

### Document 32 — Stage 9 Projection and Estimated Time of Arrival Completion

Path:

```text
docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md
```

Purpose:

```text
Records the completed research-only Production Projection Intelligence foundation,
contract and horizon policy, kinematic and historical continuation strategies,
Estimated Arrival, prediction guards, replay evaluation, PostgreSQL and HTTP
runtime evidence, deterministic fallback behavior, known limitations,
deferred weather and airspace work, and the formal Stage 9 completion statement.
```

### Document 33 — Stage 10 Weather Context Completion

Path:

```text
docs/33_STAGE_10_WEATHER_CONTEXT_COMPLETION.md
```

Purpose:

```text
Records the completed research-only Production Weather Context foundation,
canonical weather contract, Open-Meteo adapter, Weather Trust Gate,
four-dimensional alignment, Weather Encounter Profile, policy-controlled
uncertainty preservation or widening, PostgreSQL and HTTP runtime evidence,
future-evidence protection, known limitations, and the formal Stage 10
completion statement.
```

### Document 34 — Stage 11 Airspace Intelligence Completion

Path:

```text
docs/34_STAGE_11_AIRSPACE_INTELLIGENCE_COMPLETION.md
```

Purpose:

```text
Records the completed research-only Production Airspace Intelligence foundation,
interaction graph, radius policy, local traffic scenes, proximity scanning,
separation-risk context, temporal occupancy, synthetic-sector complexity,
regional analytics, PostgreSQL and HTTP runtime evidence, deterministic replay,
scope guards, known limitations, and the formal Stage 11 completion statement.
```

### Document 35 — Stage 12 Stability and Explainability Completion

Path:

```text
docs/35_STAGE_12_STABILITY_AND_EXPLAINABILITY_COMPLETION.md
```

Purpose:

```text
Records the completed research-only Production Stability and Explainability
foundation, deterministic forecast versions, Decision Stability, multi-version
Forecast Stability Analysis, Confidence Propagation, Failure Explanation,
Unknown Intervention and Scope Guard protection, standardized HTTP output,
PostgreSQL and Fiber runtime evidence, limitations, and formal Stage 12 closure.
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
```

<!-- SOURCE-CONSTRAINTS-OPENSKY-V1 -->
## Free Data Source and Evidence Boundary

```text
docs/36_FREE_DATA_SOURCE_AND_EVIDENCE_BOUNDARIES.md
```

This document is authoritative for free-source-only operation, absence of first-party collection infrastructure, absence of satellite access, absence of commercial aviation data, OpenSky evidence semantics, and prohibited analytical claims.

<!-- OPENSKY-PRODUCTION-PROVIDER-V1 -->
## OpenSky production provider selection

```text
docs/37_OPENSKY_PRODUCTION_PROVIDER_SELECTION.md
```

Document 37 records the controlled production selection boundary for the two free regional traffic providers.

<!-- TRAFFIC-PROVIDER-AUTOMATIC-FALLBACK-V1 -->
## Traffic provider automatic fallback

- `38_TRAFFIC_PROVIDER_AUTOMATIC_FALLBACK.md` — ordered free-provider fallback,
  recoverable triggers, actual-source provenance, decision evidence, and
  non-recoverable failure boundaries.
