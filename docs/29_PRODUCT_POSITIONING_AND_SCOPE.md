# Document 29 — Product Positioning and Scope

Status: Product and Architecture Baseline Candidate v1.0
Product: Open Aviation Metrics API
Repository Lineage: Global Flight Analytics
Scope: Canonical product identity, MVP boundaries, pilot scope, metric surface, confidence requirements, geographic expansion rules, and positioning constraints

---

## 1. Purpose

This document defines the canonical product positioning and scope boundaries of the project.

The project is intentionally narrowed from a broad aviation analytics platform toward a more focused backend and data engineering product.

The current product direction is:

```text
Open Aviation Metrics API
```

The product transforms incomplete, uneven, delayed, and provider-dependent open aviation observations into explainable, confidence-aware airport and regional metrics.

This document exists to prevent the project from drifting into:

```text
a worldwide flight tracker
a Flightradar24 replacement
a map-first aviation clone
an undefined collection of aviation features
a regulated aviation product
a claim of authoritative airport operations data
```

This document is authoritative for:

```text
product positioning
MVP product boundaries
public metric surface
pilot geography
airport selection policy
geographic expansion policy
API-first delivery
dashboard boundaries
global coverage claims
proxy metric semantics
confidence requirements
```

Analytical implementation decisions remain governed by:

```text
docs/28_RESEARCH_AND_ANALYTICAL_DECISION_METHOD.md
```

---

## 2. Decision Classification

The strategic repositioning defined by this document is classified as:

```text
PROJECT-DERIVED PRODUCT DECISION
```

It is based on:

```text
available free data
free infrastructure constraints
existing backend architecture
existing analytical pipeline
provider limitations
portfolio engineering goals
scope control
need for measurable outputs
need for explainable uncertainty
```

This classification does not mean that every future metric formula is automatically accepted as project policy.

Every non-trivial analytical method must still follow Document 28.

The candidate pilot airport set defined later in this document is classified as:

```text
EXPERIMENTAL
```

until coverage benchmarking and suitability review are completed.

---

## 3. Strategic Repositioning

The project must not be positioned as a new global flight tracker.

The primary objective is not:

```text
draw as many aircraft as possible
cover as many countries as possible
make a world map look populated
copy commercial tracking products
```

The primary objective is:

```text
transform imperfect open aviation observations
into useful analytical metrics
without hiding uncertainty,
stale data,
coverage weakness,
trajectory discontinuity,
or insufficient evidence
```

Worldwide visual coverage is not proof of analytical quality.

Large geographic scope is not proof of engineering maturity.

The product becomes broader only after its outputs become trustworthy.

---

## 4. Canonical Product Identity

The canonical product positioning is:

```text
Open Aviation Metrics API
```

Preferred descriptive statement:

```text
A backend and data engineering platform that transforms
imperfect open aviation observations into explainable,
confidence-aware airport and regional traffic metrics.
```

Short positioning statement:

```text
Confidence-aware aviation metrics from imperfect open data.
```

The primary product is the analytical backend and API.

The dashboard is a consumer of the API.

The map is an optional explanatory interface.

Neither the dashboard nor the map is the architectural center of the system.

---

## 5. Repository Lineage and Naming

The repository may temporarily retain historical naming during migration.

Historical repository lineage:

```text
Global Flight Analytics
```

Current product positioning:

```text
Open Aviation Metrics API
```

The historical word `Global` must not be interpreted as a promise of:

```text
worldwide operational coverage
global real-time tracking
complete global aviation visibility
global provider reliability
global analytical readiness
```

Repository renaming is outside the scope of this document and may be performed separately.

Product documentation should progressively prefer the current product positioning where the document describes the active product rather than repository history.

---

## 6. Problem Definition

Open aviation observations have structural limitations.

Possible limitations include:

```text
uneven geographic coverage
changing provider availability
stale observations
incomplete fields
inconsistent sampling density
temporary gaps
trajectory discontinuities
missing states
duplicate observations
implausible movement jumps
provider-specific semantics
delayed observations
weak arrival evidence
weak departure evidence
```

A raw provider response is not automatically a reliable product metric.

The platform therefore implements a controlled transformation:

```text
Raw Open Observations
        ↓
Provider Isolation
        ↓
Canonical State
        ↓
Normalization
        ↓
Deduplication
        ↓
Validation
        ↓
Quality Assessment
        ↓
Trajectory Evidence
        ↓
Completeness Analysis
        ↓
Freshness Analysis
        ↓
Coverage Evaluation
        ↓
Confidence-Aware Metrics
        ↓
API
        ↓
Optional Dashboard
```

---

## 7. Core Value Proposition

The platform does not compete primarily on map coverage.

The platform competes on:

```text
analytical honesty
data quality visibility
confidence visibility
coverage visibility
method transparency
explicit limitations
backend reliability
API usability
```

The core value proposition is:

```text
Take free and imperfect aviation observations
and convert them into metrics
whose quality, confidence, freshness,
sample size, provenance, and limitations are explicit.
```

The system must distinguish:

```text
observed fact
normalized value
derived value
inferred event
proxy metric
confidence estimate
quality estimate
unknown state
```

---

## 8. What the Product Is

The product is:

```text
an open-data aviation metrics API
a backend engineering project
a data engineering project
a trajectory quality system
a coverage quality system
an airport activity analytics service
a regional traffic trend service
a confidence-aware analytical platform
a research-oriented non-operational analytics system
a portfolio-grade engineering project
```

The product may expose:

```text
airport activity metrics
selected regional traffic metrics
trajectory completeness evidence
data freshness evidence
coverage quality
confidence metadata
method versions
source provenance
limitations
historical trends after sufficient data exists
```

---

## 9. What the Product Is Not

The product is not:

```text
a global flight tracker
a Flightradar24 replacement
a worldwide aircraft visibility product
an air traffic control system
a flight planning system
a dispatch system
an official aeronautical source
an official airport operations source
an authoritative arrivals database
an authoritative departures database
a certified aviation system
a commercial aviation data replacement
a cockpit navigation application
a safety-critical aviation product
```

The product must never imply otherwise.

---

## 10. MVP Product Goal

The MVP goal is:

```text
Build a reliable API that derives a small set of
confidence-aware airport activity and data quality metrics
from imperfect open aviation observations
for a limited validated pilot airport set.
```

The MVP must prove:

```text
open observation
        ↓
canonical state
        ↓
validated evidence
        ↓
trajectory evidence
        ↓
quality assessment
        ↓
metric calculation
        ↓
confidence and limitation metadata
        ↓
API response
```

The MVP does not require:

```text
worldwide coverage
a global live map
advanced prediction
large feature count
complex frontend
commercial data
```

---

## 11. MVP Pilot Model

The preferred initial MVP model is:

```text
5 pilot airports
```

The pilot airport set is deliberately small.

Its purpose is to:

```text
validate ingestion
measure provider behavior
evaluate freshness
evaluate trajectory continuity
test metric semantics
measure coverage quality
bound infrastructure cost
validate API contracts
```

The pilot set is not automatically permanent.

An airport may be removed or replaced when the available evidence is too weak for meaningful metrics.

---

## 12. Candidate Pilot Airport Set

The initial candidate set is:

```text
GYD — Heydar Aliyev International Airport — Baku
IST — Istanbul Airport — Istanbul
TBS — Tbilisi International Airport — Tbilisi
DXB — Dubai International Airport — Dubai
DOH — Hamad International Airport — Doha
```

The leading three-letter airport codes in this candidate list are descriptive references for human readability.

They must not be treated as canonical internal entity identifiers.

The system must distinguish identifier types such as:

```text
internal airport identifier
IATA code when available
ICAO code when available
source dataset identifier
source-specific external identifier
```

The canonical internal airport identity must be defined by the domain and persistence contracts rather than inferred from one external code system.

A provider-facing or user-facing airport code must not silently become the database identity of an airport.

Cross-source airport matching must preserve identifier provenance and must not assume that one code namespace is sufficient for every integration.

This set is classified as:

```text
EXPERIMENTAL
```

It provides candidate diversity across:

```text
South Caucasus
Turkey
Gulf
```

The list is not a guarantee of final active MVP support.

Every candidate must pass the Pilot Airport Promotion Gate.

---

## 13. Pilot Airport Lifecycle

A pilot airport may have one of these states:

```text
candidate
benchmarking
active
limited
suspended
retired
```

Meaning:

### candidate

Selected for possible evaluation.

### benchmarking

Currently being measured for data suitability.

### active

Passed the required promotion gate.

### limited

Supported with explicit known limitations.

### suspended

Temporarily removed from active metric claims.

### retired

No longer part of the active pilot scope.

The product must distinguish candidate geography from validated active geography.

---

## 14. Pilot Airport Promotion Gate

An airport must not become active only because it is:

```text
famous
large
strategically attractive
visually interesting
useful for marketing
```

Promotion must consider measurable evidence.

Required evaluation areas:

```text
observation availability
observation freshness
usable-state ratio
field completeness
sampling density
continuity around airport scope
trajectory reconstruction viability
inbound pattern observability
outbound pattern observability
provider stability
repeatability of results
bounded infrastructure cost
```

Possible gate outcomes:

```text
promote to active
promote to limited
continue benchmarking
suspend
retire
```

---

## 15. Five Core MVP Metrics

The MVP contains exactly five core public product metrics:

```text
1. Active Aircraft
2. Arrivals Proxy
3. Departures Proxy
4. Data Freshness
5. Coverage Score
```

Additional internal analytical values may exist.

They must not expand the public MVP metric surface without explicit documentation review.

Confidence Score is not a sixth product metric.

---

## 16. Metric 1 — Active Aircraft

### Definition

`Active Aircraft` represents the number of unique aircraft considered observably active within a defined spatial scope and calculation window after required freshness, validity, and deduplication rules are applied.

It is not:

```text
the number of raw provider records
the number of all aircraft physically present in reality
an authoritative surveillance count
```

The method must define at least:

```text
aircraft identity rule
spatial scope
calculation window
freshness rule
validity rule
deduplication rule
calculation timestamp
method version
```

Conceptual output:

```json
{
  "metric": "active_aircraft",
  "value": 42,
  "window_minutes": 15,
  "confidence_score": 0.87,
  "coverage_score": 0.79,
  "sample_size": 311,
  "method_version": "v1",
  "calculated_at": "..."
}
```

The exact public contract belongs in API documentation.

---

## 17. Metric 2 — Arrivals Proxy

### Definition

`Arrivals Proxy` represents the estimated number of observed inbound trajectory patterns compatible with approach toward a selected airport.

It is explicitly a proxy.

It must not be described as:

```text
actual arrivals
official arrivals
confirmed landings
```

Possible evidence may include:

```text
decreasing distance to airport
compatible trajectory direction
decreasing altitude
sufficient observation continuity
acceptable data freshness
acceptable trajectory quality
entry into a defined airport proximity area
```

No single signal is automatically sufficient.

The exact method must be versioned.

Weak evidence must reduce confidence.

Insufficient evidence must permit refusal.

---

## 18. Metric 3 — Departures Proxy

### Definition

`Departures Proxy` represents the estimated number of observed outbound trajectory patterns compatible with movement away from a selected airport.

It is explicitly a proxy.

It must not be described as:

```text
actual departures
official departures
confirmed takeoffs
```

Possible evidence may include:

```text
prior proximity to airport
increasing distance from airport
compatible trajectory direction
increasing altitude
sufficient observation continuity
acceptable data freshness
acceptable trajectory quality
```

No single signal is automatically sufficient.

The exact method must be versioned.

Weak evidence must reduce confidence.

Insufficient evidence must permit refusal.

---

## 19. Metric 4 — Data Freshness

### Definition

`Data Freshness` measures how current the underlying observations are relative to calculation time and metric window.

Possible internal components include:

```text
median observation age
high-percentile observation age
fresh-state ratio
stale-state ratio
latest usable observation age
observation age distribution
```

Conceptual output:

```json
{
  "metric": "data_freshness",
  "score": 0.82,
  "median_age_seconds": 11,
  "high_percentile_age_seconds": 47,
  "method_version": "v1",
  "calculated_at": "..."
}
```

The selected percentile and normalization method must be documented before becoming a stable public contract.

---

## 20. Metric 5 — Coverage Score

### Definition

`Coverage Score` represents the quality and analytical usability of observations available to this platform for a selected scope and metric window.

Coverage Score must not pretend to measure universal real-world surveillance completeness.

It measures data quality under a documented project method.

Possible components include:

```text
freshness
trajectory continuity
sampling density
field completeness
gap frequency
usable-state ratio
spatial consistency
```

Conceptual output:

```json
{
  "metric": "coverage_score",
  "value": 0.74,
  "components": {
    "freshness": 0.83,
    "continuity": 0.69,
    "sampling_density": 0.77,
    "completeness": 0.71
  },
  "method_version": "v1"
}
```

Coverage Score must be:

```text
explainable
versioned
tested
bounded
traceable to evidence
```

Hidden scoring magic is forbidden.

---

## 21. Confidence Is Not a Sixth Metric

`Confidence Score` is metadata describing the strength of evidence behind a derived result.

Correct model:

```text
Arrivals Proxy
        +
Confidence Score
```

Incorrect model:

```text
Arrivals Proxy
        +
Confidence Score as independent airport activity metric
```

Confidence may depend on:

```text
sample size
freshness
continuity
coverage
field completeness
trajectory quality
classification ambiguity
evidence agreement
```

Confidence semantics must be documented and tested.

### Acyclic Confidence Dependency Rule

Metric confidence computation must be acyclic.

A metric must not directly or indirectly depend on its own confidence result.

Forbidden analytical dependency:

```text
metric value
        ↓
metric confidence
        ↓
same metric value
```

Also forbidden:

```text
Coverage Score
        ↓
Coverage Score Confidence
        ↓
Coverage Score
```

The preferred dependency direction is:

```text
raw evidence
        ↓
independent evidence components
        ↓
metric calculation
        ↓
metric result
        ↓
metric confidence
```

A confidence calculation may use independently computed evidence such as:

```text
sample size
observation freshness
continuity evidence
field completeness
classification ambiguity
source availability
```

When confidence uses a quality component that is also exposed through another public metric, the implementation must define the dependency explicitly and prevent self-reference.

`Coverage Score` must not depend on its own `Confidence Score`.

`Confidence Score` for `Coverage Score` must not use the final `Coverage Score` value as circular evidence unless a separately documented, non-circular method proves the dependency is valid.

Analytical dependency graphs must remain explainable and testable.

---

## 22. Mandatory Metric Evidence Contract

A derived public metric is incomplete when it exposes only a number.

Where applicable, outputs should expose:

```text
value
confidence_score
coverage_score
sample_size
calculation_window
source_provenance
calculated_at
limitations
method_version
```

Not every metric must expose identical fields.

However, a derived metric must expose enough evidence for a consumer to understand its analytical strength.

Known weaknesses must not be hidden.

### Zero, Unknown, and Insufficient Evidence Rule

A numeric zero must not be used as a substitute for missing, unavailable, or insufficient evidence.

The system must distinguish:

```text
observed zero
calculated zero
unknown
unavailable
insufficient evidence
partial evidence
calculation failure
```

These states have different analytical meanings.

For example:

```text
arrivals_proxy = 0
```

is valid only when the calculation method had sufficient evidence to conclude that zero qualifying arrival-like patterns were observed in the defined scope and window.

It must not mean:

```text
the provider returned no usable data
the observation window was unavailable
coverage was too weak
trajectory evidence was insufficient
the calculation failed
```

Likewise:

```text
departures_proxy = 0
```

must not silently represent unknown or insufficient evidence.

The exact API representation may be defined later through:

```text
nullable value
explicit result status
typed availability state
limitation metadata
error contract
```

This document does not mandate one transport representation.

It mandates the semantic distinction:

```text
zero
!=
unknown
!=
unavailable
!=
insufficient evidence
```

A dashboard must preserve the same distinction.

User-facing presentation must not convert unavailable or insufficient evidence into a visible zero.

---

## 23. Proxy Honesty Rule

Proxy metrics must remain explicitly labeled as proxies.

Forbidden transformations:

```text
arrivals_proxy → arrivals
departures_proxy → departures
observed inbound pattern → confirmed landing
observed outbound pattern → confirmed takeoff
```

This rule applies to:

```text
API
dashboard
README
documentation
logs when user-facing
examples
marketing language
```

A frontend label must not be stronger than the underlying analytical contract.

---

## 24. API-First Rule

The analytical API is the primary MVP product.

Preferred implementation order:

```text
metric contract
        ↓
calculation semantics
        ↓
evidence requirements
        ↓
quality and confidence semantics
        ↓
deterministic tests
        ↓
API endpoint
        ↓
dashboard consumer
```

The project must not build a visual feature before the underlying metric contract is sufficiently defined.

---

## 25. Dashboard Boundary

The MVP dashboard must remain simple.

Its purpose is to inspect and demonstrate metrics.

Possible capabilities:

```text
select pilot airport
show five core metrics
show confidence
show freshness
show coverage quality
show calculation window
show limitations
show recent trend when historical data exists
```

The dashboard must not become:

```text
a global tracking clone
a map-first product
an uncontrolled visualization project
a substitute for analytical quality
```

A map may exist when it supports explanation.

A map is not the primary success criterion.

---

## 26. Geographic Scope Rule

The product is geographically limited by evidence.

The MVP does not claim global coverage.

The MVP does not claim regional completeness.

The MVP operates only on explicitly declared active or limited pilot scope.

Future public scope states may include:

```text
supported
experimental
limited
unsupported
```

The exact external contract for geographic support must be defined before public stabilization.

### Internal Lifecycle Versus Public Support Status

Internal pilot lifecycle status and public geographic support status are separate concepts.

Internal lifecycle:

```text
candidate
benchmarking
active
limited
suspended
retired
```

Public support status:

```text
supported
experimental
limited
unsupported
```

The two vocabularies must not be silently treated as aliases.

Conceptual distinction:

```text
internal lifecycle status
        ↓
describes operational and evaluation progression

public support status
        ↓
describes the product claim exposed to consumers
```

An internal state does not automatically determine a public state without an explicit mapping policy.

For example:

```text
internal = benchmarking
```

does not automatically imply:

```text
public = experimental
```

unless the mapping is explicitly defined.

Likewise:

```text
internal = active
```

does not automatically imply:

```text
public = supported
```

because public support may depend on additional requirements such as:

```text
stable metric behavior
documented limitations
operational observability
API contract maturity
repeatable coverage evidence
```

Before geographic support status becomes part of a stable public API, the system must define and test an explicit mapping between internal lifecycle and public support semantics.

The same word `limited` appearing in both vocabularies must not be assumed to have identical meaning without an explicit contract.

---

## 27. Regional Expansion Strategy

Regional expansion is allowed only after the pilot airport model is proven.

Directional rollout:

```text
Phase 1
Selected pilot airports

Phase 2
Turkey and South Caucasus regional trends

Phase 3
Gulf regional trends

Phase 4
Selected Europe Core areas

Phase 5
Selected United States regions

Phase 6
Additional evidence-qualified regions
```

This sequence is directional.

It is not a guaranteed release promise.

A region must not be added only to increase visual coverage.

---

## 28. Regional Expansion Gate

A region may enter active scope only when sufficient evidence exists.

The gate should consider:

```text
stable ingestion
measured provider behavior
acceptable freshness
measured usable-state ratio
trajectory viability
documented coverage limitations
bounded infrastructure cost
operational observability
repeatable configuration
metric validation
limitation messaging
```

Failure to satisfy the gate means:

```text
do not promote the region
```

Correct responses to weak evidence include:

```text
lower confidence
limited scope
continued benchmarking
temporary suspension
refusal
```

Hidden degradation is forbidden.

---

## 29. Global Coverage Policy

Global coverage is not:

```text
an MVP goal
a Version 1 promise
a marketing claim
a success criterion
an implication of repository naming
```

Future worldwide expansion is permitted only as a distant capability resulting from repeated evidence-qualified regional maturity.

The architecture may remain geographically portable.

The product must not claim worldwide analytical readiness before evidence exists.

Preferred rule:

```text
Geographically portable by architecture.
Evidence-limited by operation.
```

---

## 30. Geographic Neutrality Rule

Core analytical logic must not hardcode pilot airports, countries, or rollout phases.

Forbidden generic analytical patterns include:

```text
if country == "Azerbaijan"
if airport == "GYD"
if region == "Gulf"
```

unless a documented domain rule specifically requires such behavior.

Geographic scope should enter through explicit mechanisms such as:

```text
configuration
airport definitions
region contracts
spatial boundaries
query parameters
runtime settings
deployment configuration
```

Pilot geography must not contaminate generic:

```text
normalization
validation
deduplication
trajectory construction
quality semantics
metric semantics
```

---

## 31. Existing Architecture Preservation Rule

This repositioning does not authorize a rewrite.

The existing analytical foundation remains strategically relevant.

### Canonical FlightState

Supports:

```text
provider-independent metric inputs
normalized observation semantics
```

### Normalizer

Supports:

```text
unit consistency
field consistency
metric comparability
```

### Deduplicator

Supports:

```text
reliable active-aircraft counting
sampling integrity
```

### Validator

Supports:

```text
usable observation decisions
input evidence quality
```

### Gap Detector

Supports:

```text
trajectory completeness
continuity analysis
coverage evidence
```

### Track Builder

Supports:

```text
inbound trajectory evidence
outbound trajectory evidence
completeness analysis
```

### Trajectory Quality

Supports:

```text
confidence inputs
proxy classification strength
coverage evidence
```

### Shared Snapshot Runtime

Supports:

```text
consistent calculation cycles
aligned provider observations
```

### Weather Context

Remains a future contextual capability.

Weather is not required to prove the initial five-metric MVP.

---

## 32. Metrics-Oriented Analytical Core Direction

Future analytical development should move incrementally toward:

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

Potential concepts include:

```text
metric window
metric evidence
metric result
metric confidence
airport scope
region scope
method version
limitation metadata
```

These concepts must be introduced only when implementation requires them.

No speculative rewrite is authorized.

No abstraction is justified only because it may be useful later.

---

## 33. MVP Non-Goals

The following are not MVP goals:

```text
global live tracking
worldwide airport coverage
Flightradar24 replacement
official arrivals
official departures
advanced route prediction
machine learning prediction
global historical replay
advanced weather analytics
complex multi-aircraft interaction intelligence
large visualization platform
mobile applications
commercial aviation data replacement
regulated aviation functionality
```

---

## 34. Competitive Positioning Boundary

The project is intentionally positioned between broad categories of aviation tooling:

```text
raw open-data providers
visualization-first tracking applications
research-oriented aviation tooling
cockpit-oriented applications
commercial aviation intelligence products
```

The project does not attempt to reproduce all capabilities of those categories.

Its narrow position is:

```text
backend and data engineering for explainable,
confidence-aware metrics derived from imperfect open aviation data
```

Specific competitor comparisons must not be treated as canonical facts unless separately researched and documented.

---

## 35. Success Criteria

The first product milestone is successful when the system can:

```text
1. ingest open aviation observations
2. normalize them into canonical state
3. reject or qualify weak observations
4. construct usable trajectory evidence
5. calculate Active Aircraft
6. calculate Arrivals Proxy
7. calculate Departures Proxy
8. calculate Data Freshness
9. calculate Coverage Score
10. attach confidence and limitations
11. expose results through an API
12. operate for a small validated pilot airport set
```

The milestone does not require worldwide coverage.

The milestone does not require a sophisticated dashboard.

---

## 36. Product Decision Rules

Prefer future work that improves:

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

Defer work whose primary purpose is:

```text
making the map look larger
adding countries for marketing
adding features without metric value
copying commercial trackers
increasing complexity without measured need
```

---

## 37. Documentation Alignment Rule

This document changes the active product positioning.

Existing documents must not be assumed to be automatically consistent with it.

Affected documentation must be reviewed and classified as:

```text
compatible
requires amendment
requires direct update
partially superseded
deferred for later review
```

Updates must be incremental and reviewable.

This document must not trigger uncontrolled bulk rewriting.

---

## 38. Final Product Statement

The canonical product direction is:

```text
Open Aviation Metrics API
```

The product takes:

```text
free
incomplete
uneven
provider-dependent
open aviation observations
```

and transforms them into:

```text
airport activity metrics
regional traffic trends
trajectory completeness evidence
data freshness evidence
coverage quality
confidence-aware analytical outputs
```

The defining principle is:

```text
Do not pretend the data is better than it is.
Measure the weakness.
Expose the uncertainty.
Return the metric with evidence.
```

The correct development sequence is:

```text
small pilot scope
        ↓
reliable evidence
        ↓
five stable metrics
        ↓
API
        ↓
simple dashboard
        ↓
regional trends
        ↓
controlled expansion
```

Worldwide coverage is not the starting point.

Trustworthy metrics are.
