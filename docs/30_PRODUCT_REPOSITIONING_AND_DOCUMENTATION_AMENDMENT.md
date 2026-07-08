# Document 30 — Product Repositioning and Documentation Amendment

Status: Product and Architecture Amendment v1.0
Authority: Applies Document 29 across the existing documentation baseline
Product: Open Aviation Metrics API
Repository Lineage: Global Flight Analytics
Scope: Documentation precedence, supersession mapping, retained architecture, amended MVP interpretation, and migration consequences

---

## 1. Purpose

This document defines how Document 29 changes the interpretation of the existing documentation baseline.

Document 29 established a new canonical product direction:

```text
Open Aviation Metrics API
```

The project is now centered on:

```text
confidence-aware airport activity metrics
regional traffic trends
trajectory completeness evidence
data freshness evidence
coverage quality
explicit uncertainty
API-first delivery
evidence-limited geographic scope
```

This amendment exists to prevent two opposite failures:

```text
1. silently keeping earlier documents as if the product had not changed
2. destructively rewriting the entire documentation history
```

The amendment preserves valid engineering work while explicitly superseding conflicting product claims, MVP priorities, geographic assumptions, and implementation ordering.

---

## 2. Authority

This document applies Document 29 to the earlier documentation baseline.

Document 29 is authoritative for:

```text
product positioning
MVP product boundaries
pilot airport model
public metric surface
proxy semantics
confidence requirements
geographic scope policy
regional expansion gates
API-first delivery
dashboard boundaries
global coverage claims
```

Document 30 is authoritative for:

```text
how earlier Documents 01–28 must be interpreted
which earlier claims remain valid
which earlier claims are amended
which earlier claims are partially superseded
which documents require only a precedence notice
```

Analytical research-to-code decisions remain governed by:

```text
docs/28_RESEARCH_AND_ANALYTICAL_DECISION_METHOD.md
```

---

## 3. Precedence Rule

When a conflict exists, use the following precedence:

```text
Document 29
        ↓
Document 30
        ↓
earlier Documents 01–28
```

The newer decision takes precedence only for the conflicting scope.

Unrelated sections remain valid.

A conflict must not be resolved by silently deleting earlier reasoning.

---

## 4. Non-Destructive Amendment Rule

Existing documentation remains part of the project history.

Earlier documents must not be treated as wholly invalid only because the product has been repositioned.

The correct interpretation is:

```text
retain valid engineering foundation
        ↓
supersede conflicting product claims
        ↓
amend MVP and rollout priorities
        ↓
preserve historical rationale
```

Mass replacement of product names without semantic review is forbidden.

Mass rewriting of Documents 01–28 is not required by this amendment.

---

## 5. Canonical Product Direction

The active product direction is:

```text
Open Aviation Metrics API
```

The product is not primarily:

```text
a global flight tracker
a map-first aviation product
a Flightradar24 replacement
a worldwide coverage claim
an authoritative airport operations source
```

The product is primarily:

```text
a backend and data engineering platform
for explainable,
confidence-aware metrics
derived from imperfect open aviation data
```

---

## 6. MVP Authority

The active MVP is defined by Document 29.

The MVP centers on five public metrics:

```text
1. Active Aircraft
2. Arrivals Proxy
3. Departures Proxy
4. Data Freshness
5. Coverage Score
```

Confidence Score is not a sixth public activity metric.

Confidence is evidence metadata.

The MVP uses a small evidence-qualified pilot airport set.

The candidate set is experimental until benchmarking and promotion gates are satisfied.

The MVP is API-first.

A simple dashboard follows stable metric contracts.

---

## 7. Metric Honesty Authority

The following semantics are mandatory:

```text
arrivals_proxy != actual arrivals
departures_proxy != actual departures
zero != unknown
zero != unavailable
zero != insufficient evidence
```

Derived public metrics must expose sufficient evidence to communicate analytical strength.

Where applicable, evidence may include:

```text
confidence_score
coverage_score
sample_size
calculation_window
source_provenance
calculated_at
limitations
method_version
```

The API, dashboard, README, examples, and product language must preserve these distinctions.

---

## 8. Geographic Authority

The active geographic rule is:

```text
Geographically portable by architecture.
Evidence-limited by operation.
```

The MVP does not claim global coverage.

The MVP does not claim regional completeness.

Core analytical logic must not hardcode pilot countries, airports, or rollout phases without a documented domain reason.

Geographic scope should enter through explicit contracts and configuration.

---

## 9. Existing Architecture Preservation

The repositioning does not authorize a rewrite.

The following existing foundations remain strategically valid:

```text
Canonical FlightState
provider isolation
normalization
deduplication
validation
data quality evaluation
gap detection
Track Builder
Trajectory Segment
Flight Trajectory
trajectory quality
shared snapshot runtime
PostgreSQL persistence
API boundary
observability
testing discipline
```

These capabilities are reinterpreted as foundations for metric evidence and confidence-aware outputs.

---

## 10. Analytical Core Reinterpretation

The active analytical direction is:

```text
Provider Observations
        ↓
Canonical State Pipeline
        ↓
Observation Quality
        ↓
Trajectory Evidence
        ↓
Metric Evidence Windows
        ↓
Airport Metric Calculators
        ↓
Metric Confidence
        ↓
Metric API
        ↓
Regional Aggregation
```

Earlier advanced analytical capabilities remain possible future work.

They are not automatically MVP priorities.

---

## 11. Implementation Ordering Amendment

The preferred product delivery order is now:

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

An earlier sequence that promotes frontend or map work before stable metric contracts is amended by this rule.

Technical foundation stages remain valid where they support this sequence.

---

## 12. Document Impact Classification

Each earlier document is classified with one of these statuses:

```text
COMPATIBLE
VALID WITH REINTERPRETATION
REQUIRES PRECEDENCE NOTICE
PARTIALLY SUPERSEDED
MVP SCOPE SUPERSEDED
SEQUENCE AMENDED
DEFERRED REVIEW
NO CHANGE REQUIRED
```

These classifications apply only to conflicts introduced by Documents 29 and 30.

---

## 13. Document 01 — Product Vision

Path:

```text
docs/01_PRODUCT_VISION.md
```

Status:

```text
PARTIALLY SUPERSEDED
REQUIRES PRECEDENCE NOTICE
```

Still valid:

```text
open aviation data foundation
aviation analytics purpose
airport and regional analysis
research orientation
data quality concerns
```

Superseded where conflicting:

```text
broad global platform positioning
undefined worldwide ambition
map-centered interpretation
product identity inconsistent with Open Aviation Metrics API
```

Active replacement authority:

```text
Document 29 Sections 3–10
Document 30 Sections 5–6
```

---

## 14. Document 02 — System Architecture

Path:

```text
docs/02_SYSTEM_ARCHITECTURE.md
```

Status:

```text
VALID WITH REINTERPRETATION
```

Still valid:

```text
backend boundary
frontend boundary
database boundary
integration structure
modular architecture
```

Amended interpretation:

```text
analytical backend and API are the primary product
frontend is a consumer
geographic scope must remain explicit and configurable
pilot geography must not contaminate generic core logic
```

No immediate full rewrite is required.

---

## 15. Document 03 — Domain Model

Path:

```text
docs/03_DOMAIN_MODEL.md
```

Status:

```text
DEFERRED REVIEW
```

Reason:

Future implementation may require explicit concepts such as:

```text
MetricWindow
MetricEvidence
MetricResult
MetricConfidence
AirportScope
RegionScope
MethodVersion
LimitationMetadata
```

These concepts must not be added speculatively.

Review only when a vertical implementation slice requires them.

---

## 16. Document 04 — Database Design

Path:

```text
docs/04_DATABASE_DESIGN.md
```

Status:

```text
DEFERRED REVIEW
```

Required future audit:

```text
arrivals terminology
departures terminology
metric storage
method version storage
confidence metadata
availability state
zero versus unknown semantics
```

Existing database design remains valid until a concrete conflicting schema decision is implemented.

---

## 17. Document 05 — Data Sources

Path:

```text
docs/05_DATA_SOURCES.md
```

Status:

```text
COMPATIBLE
```

Document 29 strengthens the need to expose:

```text
provider limitations
coverage limitations
freshness limitations
provenance
source-specific semantics
```

No immediate rewrite is required unless factual source descriptions are outdated.

---

## 18. Document 06 — Data Collection Pipeline

Path:

```text
docs/06_DATA_COLLECTION_PIPELINE.md
```

Status:

```text
VALID WITH REINTERPRETATION
DEFERRED REVIEW
```

The collection pipeline remains valid as a foundation.

Future review must ensure that:

```text
arrivals
departures
```

are not presented as authoritative facts when the implementation only supports proxies.

---

## 19. Document 07 — Route Detection Engine

Path:

```text
docs/07_ROUTE_DETECTION_ENGINE.md
```

Status:

```text
VALID FUTURE CAPABILITY
PARTIALLY SUPERSEDED AS MVP PRIORITY
```

Route intelligence is not the center of the active MVP.

The module may remain future analytical capability.

Trajectory evidence required by MVP metrics remains valid.

---

## 20. Document 08 — Airport Intelligence Module

Path:

```text
docs/08_AIRPORT_INTELLIGENCE_MODULE.md
```

Status:

```text
VALID WITH REINTERPRETATION
```

This module becomes strategically important for:

```text
airport scope
pilot airport benchmarking
Active Aircraft
Arrivals Proxy
Departures Proxy
Data Freshness
Coverage Score
```

Detailed update should occur together with concrete metric contracts.

---

## 21. Document 09 — Traffic Analytics Module

Path:

```text
docs/09_TRAFFIC_ANALYTICS_MODULE.md
```

Status:

```text
VALID WITH REINTERPRETATION
```

Traffic analytics should progressively orient toward:

```text
metric windows
airport activity metrics
regional trends
freshness evidence
coverage evidence
confidence-aware aggregation
```

Detailed update should follow implementation contracts.

---

## 22. Document 10 — API Specification

Path:

```text
docs/10_API_SPECIFICATION.md
```

Status:

```text
DEFERRED REVIEW
```

The active MVP requires a metrics API.

However, the exact public contract must not be invented before metric semantics stabilize.

Future review must preserve:

```text
proxy naming
zero versus unknown
confidence metadata
coverage metadata
method version
limitations
availability state
```

---

## 23. Document 11 — Frontend Architecture

Path:

```text
docs/11_FRONTEND_ARCHITECTURE.md
```

Status:

```text
PARTIALLY SUPERSEDED AS MVP PRIORITY
```

Frontend remains valid as an API consumer.

The map is not the architectural center.

The preferred order is:

```text
metric contract
        ↓
metric implementation
        ↓
API
        ↓
simple dashboard
```

---

## 24. Document 12 — Infrastructure and Deployment

Path:

```text
docs/12_INFRASTRUCTURE_AND_DEPLOYMENT.md
```

Status:

```text
COMPATIBLE
```

Free-tier and operational constraints remain valid.

Regional and pilot scope should reduce rather than increase infrastructure assumptions.

No immediate change is required.

---

## 25. Document 13 — Security Specification

Path:

```text
docs/13_SECURITY_SPECIFICATION.md
```

Status:

```text
NO CHANGE REQUIRED
```

No direct product-positioning conflict identified.

---

## 26. Document 14 — Performance and Scalability

Path:

```text
docs/14_PERFORMANCE_AND_SCALABILITY.md
```

Status:

```text
VALID WITH REINTERPRETATION
```

Performance planning must prioritize:

```text
pilot reliability
bounded metric windows
regional portability
measured expansion
```

Worldwide scale is not an active product requirement.

No immediate rewrite is required.

---

## 27. Document 15 — Development Roadmap

Path:

```text
docs/15_DEVELOPMENT_ROADMAP.md
```

Status:

```text
PARTIALLY SUPERSEDED
DEFERRED REVIEW
```

Earlier roadmap items remain historical planning context.

Active delivery priority follows Documents 29 and 30.

---

## 28. Document 16 — MVP Scope

Path:

```text
docs/16_MVP_SCOPE.md
```

Status:

```text
MVP SCOPE SUPERSEDED
REQUIRES PRECEDENCE NOTICE
```

Where conflicts exist, the active MVP is:

```text
small candidate pilot airport set
five public metrics
confidence-aware outputs
explicit limitations
API-first delivery
simple dashboard later
```

Still-valid technical foundations may remain in force.

Active replacement authority:

```text
Document 29 Sections 10–25
Document 30 Sections 6–11
```

---

## 29. Document 17 — Future Versions

Path:

```text
docs/17_FUTURE_VERSIONS.md
```

Status:

```text
PARTIALLY SUPERSEDED
DEFERRED REVIEW
```

Future capabilities remain hypotheses, not release promises.

Regional expansion must pass evidence gates.

Global coverage is not a guaranteed future version.

---

## 30. Document 18 — Technical Decisions Record

Path:

```text
docs/18_TECHNICAL_DECISIONS_RECORD.md
```

Status:

```text
COMPATIBLE
```

Documents 29 and 30 add newer product and documentation decisions.

Existing technical decision history remains valid unless explicitly contradicted.

---

## 31. Document 19 — Risk Analysis

Path:

```text
docs/19_RISK_ANALYSIS.md
```

Status:

```text
COMPATIBLE
```

The existing risk around attempting to compete with Flightradar24 is strengthened by Document 29.

Future updates may add:

```text
proxy promotion risk
hidden coverage weakness
zero versus unknown confusion
map-driven expansion
confidence circularity
```

No immediate full rewrite is required.

---

## 32. Document 20 — Final Architecture Blueprint

Path:

```text
docs/20_FINAL_ARCHITECTURE_BLUEPRINT.md
```

Status:

```text
PARTIALLY SUPERSEDED
REQUIRES PRECEDENCE NOTICE
```

Still valid:

```text
technical architecture
integration boundaries
data flow
backend and database structure
analytical components
```

Superseded where conflicting:

```text
broad observation-system positioning
global-platform interpretation
prediction-centered product identity
map-centered product identity
```

Active replacement authority:

```text
Document 29
Document 30 Sections 5, 9, 10, 11
```

---

## 33. Document 21 — Engineering Amendments v1.1

Path:

```text
docs/21_ENGINEERING_AMENDMENTS_v1.1.md
```

Status:

```text
COMPATIBLE
```

Engineering amendments remain valid unless a direct conflict is identified during implementation.

No immediate rewrite is required.

---

## 34. Document 22 — Research Audit Deduplication

Path:

```text
docs/22_RESEARCH_AUDIT_DEDUPLICATION.md
```

Status:

```text
VALID WITH REINTERPRETATION
```

The deduplicated analytical architecture remains valuable.

Advanced capabilities are not automatically MVP priorities.

---

## 35. Document 23 — Analytical Core Architecture

Path:

```text
docs/23_ANALYTICAL_CORE_ARCHITECTURE.md
```

Status:

```text
VALID WITH REINTERPRETATION
```

Active reinterpretation:

```text
trajectory intelligence
        ↓
trajectory evidence foundation

route intelligence
        ↓
future capability unless required by metrics

confidence and explainability
        ↓
first-class metric evidence responsibility

analytical core
        ↓
metrics-oriented analytical core direction
```

No rewrite is authorized.

---

## 36. Document 24 — MVP and Version Roadmap

Path:

```text
docs/24_MVP_VERSION_ROADMAP.md
```

Status:

```text
PARTIALLY SUPERSEDED
REQUIRES PRECEDENCE NOTICE
```

Still valid:

```text
staged delivery
progressive maturity
explicit version boundaries
```

Superseded where conflicting:

```text
route intelligence as immediate MVP center
flight phase as mandatory immediate product output
frontend before stable metric contracts
map-driven milestone sequencing
```

Active replacement authority:

```text
Document 29
Document 30 Sections 6 and 11
```

---

## 37. Document 25 — Implementation Sequence

Path:

```text
docs/25_IMPLEMENTATION_SEQUENCE.md
```

Status:

```text
SEQUENCE AMENDED
REQUIRES PRECEDENCE NOTICE
```

Still valid:

```text
data foundation
data quality
trajectory foundation
incremental implementation
```

The sequence is amended after trajectory evidence:

```text
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
```

An earlier frontend-first continuation is no longer authoritative.

---

## 38. Document 26 — Research Backlog and Scope Guards

Path:

```text
docs/26_RESEARCH_BACKLOG_AND_SCOPE_GUARDS.md
```

Status:

```text
COMPATIBLE
```

Document 29 adds stronger product guards:

```text
no global tracker positioning
no Flightradar24 replacement claim
no proxy-to-fact promotion
no hidden low coverage
no map-driven geographic expansion
```

These guards are active through Document 29 even before direct textual synchronization.

---

## 39. Document 27 — Engineering Principles

Path:

```text
docs/27_ENGINEERING_PRINCIPLES.md
```

Status:

```text
COMPATIBLE
```

Documents 29 and 30 add active principles:

```text
metrics before visualization
confidence-aware derived outputs
geographic neutrality
zero is not unknown
acyclic analytical dependencies
```

Existing engineering principles remain valid.

---

## 40. Document 28 — Research and Analytical Decision Method

Path:

```text
docs/28_RESEARCH_AND_ANALYTICAL_DECISION_METHOD.md
```

Status:

```text
COMPATIBLE
```

Document 28 remains authoritative for research-to-code decisions.

The new product direction strengthens its rules around:

```text
metrics
confidence
limitations
baseline-first implementation
no pseudo-precision
source versus hypothesis separation
```

No rewrite is required.

---

## 41. Document 29 — Product Positioning and Scope

Path:

```text
docs/29_PRODUCT_POSITIONING_AND_SCOPE.md
```

Status:

```text
CANONICAL PRODUCT AUTHORITY
```

Document 29 defines the active product direction and scope.

---

## 42. Required Precedence Notices

The following documents require a short amendment notice because they contain high-risk conflicting product or MVP claims:

```text
docs/01_PRODUCT_VISION.md
docs/16_MVP_SCOPE.md
docs/20_FINAL_ARCHITECTURE_BLUEPRINT.md
docs/24_MVP_VERSION_ROADMAP.md
docs/25_IMPLEMENTATION_SEQUENCE.md
```

The notice must state that Documents 29 and 30 take precedence where conflicts exist.

No larger rewrite is required by this amendment.

---

## 43. Documentation Index Amendment

`docs/DOCUMENT_INDEX.md` must register:

```text
Document 29 — Product Positioning and Scope
Document 30 — Product Repositioning and Documentation Amendment
```

The index must also state the active precedence rule.

The index may retain historical baseline sections if they are clearly marked as historical or amended.

---

## 44. Implementation Consequences

After this amendment, future implementation should prefer vertical slices that improve:

```text
metric correctness
metric explainability
confidence quality
coverage honesty
trajectory evidence
freshness evidence
API usability
pilot reliability
regional portability
```

Work should be deferred when its primary purpose is:

```text
making the map look larger
adding countries for marketing
copying commercial trackers
adding features without metric value
increasing complexity without measured need
```

---

## 45. Technical Debt Continuity

This amendment does not cancel existing technical debt.

Current bug fixing and technical-debt closure remain mandatory.

Product repositioning must not be used as justification to leave known regressions unresolved.

The correct sequence after documentation migration is:

```text
close documentation amendment
        ↓
return to active technical-debt branch
        ↓
finish current Open-Meteo timeout invariant work
        ↓
resolve remaining known blockers
        ↓
continue architecture and analytical core
        ↓
implement metrics-oriented vertical slices
```

---

## 46. Final Active Baseline

The active documentation interpretation is:

```text
Documents 01–28
        ↓
historical and engineering foundation

Document 29
        ↓
canonical product positioning and scope

Document 30
        ↓
amendment and supersession mapping

Document 28
        ↓
research-to-code decision authority
```

When there is no conflict, earlier documents remain valid.

When there is a conflict in product positioning, MVP scope, geographic policy, metric semantics, or rollout order:

```text
Document 29
        ↓
Document 30
        ↓
earlier conflicting text
```

The project must preserve valid engineering foundations while building the new product direction incrementally.

The final rule is:

```text
Do not erase the history.
Do not preserve contradictions.
Make precedence explicit.
Keep the architecture.
Narrow the product.
Fix the known bugs.
Then continue building.
```
