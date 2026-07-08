# Documentation Index — Global Flight Analytics

Status: Documentation Index v1.3
Project Lineage: Global Flight Analytics
Active Product Positioning: Open Aviation Metrics API

---

## Purpose

This index records the documentation structure and active authority model for the project.

The documentation is divided into three groups:

```text
Documents 01–21:
existing product, system, data, architecture foundation, and engineering amendments

Documents 22–28:
research audit deduplication, analytical core, roadmap, implementation sequence,
research scope guards, engineering principles, and research-to-implementation method

Documents 29–30:
canonical product repositioning and explicit documentation amendment
```

---

## Active Authority and Precedence

The active product authority is:

```text
docs/29_PRODUCT_POSITIONING_AND_SCOPE.md
```

The active documentation amendment is:

```text
docs/30_PRODUCT_REPOSITIONING_AND_DOCUMENTATION_AMENDMENT.md
```

The authoritative research-to-code method remains:

```text
docs/28_RESEARCH_AND_ANALYTICAL_DECISION_METHOD.md
```

When a conflict exists in product positioning, MVP scope, geographic policy, metric semantics, or rollout order, use:

```text
Document 29
        ↓
Document 30
        ↓
earlier conflicting text
```

When no conflict exists, earlier documentation remains valid.

---

## Existing Foundation Documents

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

## Analytical Architecture Documents

### Document 22 — Research Audit Deduplication

Path:

```text
docs/22_RESEARCH_AUDIT_DEDUPLICATION.md
```

Purpose:

```text
Consolidates research audit outputs into deduplicated architecture layers
and defines accepted analytical architecture ideas.
```

### Document 23 — Analytical Core Architecture

Path:

```text
docs/23_ANALYTICAL_CORE_ARCHITECTURE.md
```

Purpose:

```text
Defines the analytical core, including trajectory intelligence,
route intelligence, historical similarity, historical patterns,
weather-aware intelligence, projection, multi-aircraft context,
airport intelligence, confidence, and explainability.
```

Interpretation note:

```text
Advanced capabilities remain valid future architecture,
but are not automatically MVP priorities after Document 29.
```

### Document 24 — MVP and Version Roadmap

Path:

```text
docs/24_MVP_VERSION_ROADMAP.md
```

Status:

```text
PARTIALLY SUPERSEDED
```

Interpret through Documents 29 and 30 where conflicts exist.

### Document 25 — Implementation Sequence

Path:

```text
docs/25_IMPLEMENTATION_SEQUENCE.md
```

Status:

```text
SEQUENCE AMENDED
```

Technical foundation stages remain valid.

After trajectory evidence, the active direction prioritizes:

```text
metric evidence
five MVP metrics
metrics API
pilot validation
simple dashboard
regional trends
```

### Document 26 — Research Backlog and Scope Guards

Path:

```text
docs/26_RESEARCH_BACKLOG_AND_SCOPE_GUARDS.md
```

Purpose:

```text
Defines deferred research topics, MVP forbidden scope,
promotion rules, prediction guards, weather guards,
and open-data limitations.
```

### Document 27 — Engineering Principles

Path:

```text
docs/27_ENGINEERING_PRINCIPLES.md
```

Purpose:

```text
Defines simple-first implementation, controlled complexity,
magic number avoidance, policy visibility, testing,
and documentation alignment.
```

### Document 28 — Research and Analytical Decision Method

Path:

```text
docs/28_RESEARCH_AND_ANALYTICAL_DECISION_METHOD.md
```

Status:

```text
AUTHORITATIVE RESEARCH-TO-CODE METHOD
```

Purpose:

```text
Defines mandatory decision classification,
open research expansion, physics and mathematics rules,
baseline-first analytics, threshold derivation,
historical replay, metrics, confidence, limitations,
and scope protection.
```

---

## Product Repositioning Documents

### Document 29 — Product Positioning and Scope

Path:

```text
docs/29_PRODUCT_POSITIONING_AND_SCOPE.md
```

Status:

```text
CANONICAL PRODUCT AUTHORITY
```

Purpose:

```text
Defines Open Aviation Metrics API positioning,
MVP boundaries, candidate pilot scope,
five public metrics, confidence requirements,
proxy honesty, API-first delivery,
geographic scope, and expansion gates.
```

### Document 30 — Product Repositioning and Documentation Amendment

Path:

```text
docs/30_PRODUCT_REPOSITIONING_AND_DOCUMENTATION_AMENDMENT.md
```

Status:

```text
AUTHORITATIVE DOCUMENTATION AMENDMENT
```

Purpose:

```text
Defines how Documents 01–28 are interpreted after Document 29,
records supersession and amendment status,
preserves valid engineering foundations,
and makes precedence explicit.
```

---

## Active Product Baseline

```text
Open Aviation Metrics API

small evidence-qualified pilot scope
        ↓
five public metrics
        ↓
confidence and limitation metadata
        ↓
metrics API
        ↓
pilot validation
        ↓
simple dashboard
        ↓
regional trends
        ↓
controlled expansion
```

Five public MVP metrics:

```text
1. Active Aircraft
2. Arrivals Proxy
3. Departures Proxy
4. Data Freshness
5. Coverage Score
```

Confidence Score is evidence metadata, not a sixth activity metric.

---

## Active Architecture Baseline

```text
Open Data Sources
        ↓
Source Adapters
        ↓
Canonical Flight State
        ↓
Normalization
        ↓
Deduplication
        ↓
Validation
        ↓
Data Quality
        ↓
Trajectory Evidence
        ↓
Metric Evidence Windows
        ↓
Airport Metric Calculators
        ↓
Metric Confidence
        ↓
Metrics API
        ↓
Regional Aggregation
        ↓
Simple Dashboard
```

This baseline does not authorize a rewrite.

Existing working architecture must be evolved incrementally.

---

## Active Implementation Direction

Preferred order:

```text
data foundation
        ↓
data quality
        ↓
trajectory evidence
        ↓
metric evidence
        ↓
five MVP metrics
        ↓
metrics API
        ↓
pilot validation
        ↓
simple dashboard
        ↓
regional trends
        ↓
controlled expansion
```

---

## Metric Honesty Rules

Mandatory distinctions:

```text
arrivals_proxy != actual arrivals
departures_proxy != actual departures
zero != unknown
zero != unavailable
zero != insufficient evidence
```

Metric confidence computation must remain acyclic.

Known limitations must not be hidden.

---

## Required Precedence Notices

The following documents carry high-risk conflicting product or MVP claims and must be interpreted with Documents 29 and 30:

```text
01_PRODUCT_VISION.md
16_MVP_SCOPE.md
20_FINAL_ARCHITECTURE_BLUEPRINT.md
24_MVP_VERSION_ROADMAP.md
25_IMPLEMENTATION_SEQUENCE.md
```

---

## Superseded Duplicate Notice

The file below is superseded and must not be used as the active baseline:

```text
docs/21_RESEARCH_AUDIT_DEDUPLICATION.md
```

The active replacement is:

```text
docs/22_RESEARCH_AUDIT_DEDUPLICATION.md
```

---

## Mandatory Research-to-Code Rule

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

The authoritative method is defined in Document 28.

---

## Documentation Rule

New changes must not silently overwrite earlier documents.

Future changes should use:

```text
new numbered documents
explicit amendments
clearly marked direct updates
precedence notices
```

Mass semantic replacement without review is forbidden.

---

## Final Documentation Statement

The project documentation now recognizes:

```text
Global Flight Analytics
        ↓
historical repository and documentation lineage

Open Aviation Metrics API
        ↓
active product positioning
```

The active product is a backend and data engineering platform centered on:

```text
airport activity metrics
regional traffic trends
trajectory completeness evidence
data freshness
coverage quality
confidence-aware outputs
explicit limitations
```

The final documentation rule is:

```text
preserve valid engineering foundations
make supersession explicit
keep uncertainty visible
build metrics before visualization
fix known technical debt
continue incrementally
```
