# Document 28 — Research and Analytical Decision Method

Status: Architecture Baseline v1.0  
Project: Global Flight Analytics  
Scope: Mandatory method for converting research, physics, mathematics, open data, and project-specific reasoning into implementation decisions

---

## 1. Purpose

This document defines the mandatory decision method for analytical and research-driven implementation in Global Flight Analytics.

It exists to prevent four failures:

```text
1. copying research ideas without checking project constraints
2. inventing formulas or thresholds without evidence
3. silently presenting project hypotheses as source-backed facts
4. adding complexity or infrastructure that the project cannot operate
```

This document is authoritative for research-to-code decisions.

---

## 2. Three Hard Constraints

Every research idea, formula, algorithm, model, engineering pattern, and implementation proposal must first be tested against these constraints:

```text
1. only free data sources
2. no proprietary data-collection infrastructure
3. no dependency on satellite systems or commercial aviation data
```

An idea that violates one of these constraints must not enter the active implementation path unless a free and operationally realistic substitute is identified.

Possible outcomes:

```text
accepted
research-adapted
deferred
rejected
```

---

## 3. Mandatory Decision Sequence

When the research digest does not provide a complete implementation, use this sequence:

```text
1. Check the current documentation baseline.
2. Check the research digest.
3. Apply the three hard constraints.
4. Identify the free data that is actually available.
5. Separate source-backed fact from project hypothesis.
6. Build the simplest correct baseline.
7. Do not use unexplained numbers.
8. Add deterministic tests.
9. Evaluate through historical replay when history is required.
10. Measure explicit metrics.
11. Add confidence or uncertainty where output is inferential.
12. Expose limitations.
13. Only then increase complexity.
```

This sequence is mandatory.

---

## 4. Decision Classification Labels

Every non-trivial analytical decision must be classified with one of these labels in design notes, implementation comments when necessary, review descriptions, or analytical documentation.

### SOURCE-BACKED

```text
Directly supported by a published source,
official documentation,
formal standard,
or openly documented model.
```

The implementation must not claim more than the source supports.

### RESEARCH-ADAPTED

```text
Derived from a published research method,
but adapted to Global Flight Analytics constraints,
canonical models,
available free data,
or implementation boundaries.
```

The adaptation must be explicitly identified as project-specific.

### PHYSICS-DERIVED

```text
Derived from established physical or mathematical laws,
but dependent on whether the required input variables
can be observed or estimated from available free data.
```

The project must distinguish exact inputs from estimated inputs.

### PROJECT-DERIVED

```text
Derived internally from project data,
architecture,
constraints,
measured behavior,
or explicit logical reasoning.
```

The derivation must be explainable and testable.

### EXPERIMENTAL

```text
A hypothesis that requires measurement,
historical replay,
benchmarking,
calibration,
or comparison against a baseline.
```

Experimental logic must not be presented as verified fact.

### DEFERRED

```text
Potentially valuable,
but outside the current implementation stage,
current data availability,
or current operational constraints.
```

Deferred work belongs in the research backlog or a later roadmap version.

---

## 5. No Silent Mixing Rule

The project must never silently mix:

```text
what a source states
and
what the project infers
```

Required distinction:

```text
source statement
↓
project adaptation
↓
project-specific assumption
↓
experimental hypothesis
```

A project-derived rule must not be described as if it were published by the source.

---

## 6. Open Research Expansion Rule

The research digest is not the final boundary of available knowledge.

When the digest lacks a formula, algorithm, threshold, or implementation pattern, the project should search open sources before inventing a solution.

Preferred source order:

```text
1. official public documentation
2. public government or intergovernmental technical material
3. peer-reviewed or primary research papers
4. openly documented scientific models
5. university research and technical publications
6. reputable open-source implementations with traceable methodology
```

Research must still pass the three hard constraints before entering implementation.

---

## 7. Physics and Mathematics Rule

Aviation analytics should use physical and mathematical structure when appropriate.

Candidate domains include:

```text
4D kinematics
geodesic distance
velocity and acceleration
turn rate
climb and descent rates
wind-vector correction
atmospheric density
lift and drag relationships
point-mass aircraft dynamics
energy-state models
fuel-burn estimation
uncertainty propagation
trajectory similarity
probabilistic continuation
```

However, known equations do not imply known inputs.

For every physics-derived model, explicitly classify each input as:

```text
observed
openly sourced
estimated
derived
unknown
```

If a required variable is unknown, the output must not be presented as exact.

---

## 8. No Pseudo-Precision Rule

The project must not claim exact values for quantities that are not directly observed or reliably available.

Examples include:

```text
actual aircraft mass
actual payload
actual fuel remaining
actual thrust setting
actual drag coefficient for current configuration
actual Air Traffic Control intent
actual commercial flight plan when unavailable
```

Allowed outputs may include:

```text
estimate
range
interval
relative score
candidate ranking
confidence level
uncertainty band
limitation statement
```

---

## 9. Baseline-First Analytical Rule

When no complete method is available, start with the simplest baseline that is:

```text
correct enough to test
readable
measurable
replaceable
explainable
```

Examples:

```text
rule-based baseline before machine learning
endpoint and corridor filtering before advanced similarity
strict duplicate detection before fuzzy deduplication
simple historical neighbors before deep prediction
named static policy before unmeasured adaptive thresholds
```

Complexity must be earned by measured failure of the simpler baseline.

---

## 10. Threshold Derivation Rule

No analytical threshold may be introduced only because it appears reasonable.

A threshold must come from at least one of:

```text
physical boundary
formal specification
provider semantics
published research
measured project data
historical replay
explicit product policy
```

If a threshold is temporary, it must be classified as EXPERIMENTAL and have a measurement plan.

---

## 11. Historical Replay Rule

Historical replay is the primary validation method for project-derived and experimental analytics when historical observations are available.

The required pattern is:

```text
historical observations
↓
reconstruct state available at time T
↓
run analytical method without future leakage
↓
compare output with later observed reality
↓
measure metrics
↓
calibrate confidence
```

Future data must not leak into a past-time evaluation.

---

## 12. Metrics Rule

An analytical method must define what success means before it is considered mature.

Possible metrics include:

```text
precision
recall
false-positive rate
false-negative rate
coverage rate
endpoint error
trajectory distance error
calibration error
confidence reliability
latency
missing-data sensitivity
source sensitivity
```

The metric must match the analytical question.

---

## 13. Confidence and Limitation Rule

Inferential outputs must include confidence or limitations when the system does not possess complete ground truth.

Examples:

```text
probable route
estimated continuation
historical pattern match
weather-aware deviation interpretation
estimated performance state
fuel-burn estimate
```

The user-facing system must distinguish:

```text
observed
calculated
estimated
inferred
unknown
```

---

## 14. Data Quality Processing Order

For trajectory analytics, the preferred processing order is:

```text
external provider response
↓
source adapter
↓
Canonical FlightState
↓
normalization
↓
duplicate point handling
↓
validation
↓
data quality evaluation
↓
trust gate
↓
repository and processing services
↓
Track Builder
↓
FlightTrajectory
↓
feature engineering
↓
context enrichment
↓
analytical core
↓
confidence and explainability
```

Duplicate handling must not silently destroy distinct valid observations.

Strict rules should precede fuzzy rules.

---

## 15. Repeated Data Versus Repeated Behavior

The project must distinguish duplicate observations from repeated historical patterns.

```text
same repeated observation
→ data-quality problem
→ remove or consolidate conservatively
```

```text
similar route or behavior repeated across history
→ analytical signal
→ preserve for historical similarity and pattern intelligence
```

Repeated behavior is not duplicate data.

---

## 16. Research-to-Implementation Gate

Before implementation begins, answer:

```text
1. Which documentation requirement does this support?
2. Which research source or reasoning class supports it?
3. Does it satisfy the three hard constraints?
4. What free input data is available?
5. Which inputs are observed, derived, estimated, or unknown?
6. Is the design SOURCE-BACKED, RESEARCH-ADAPTED, PHYSICS-DERIVED, PROJECT-DERIVED, EXPERIMENTAL, or DEFERRED?
7. What is the simplest baseline?
8. What tests prove the rule?
9. What metrics evaluate it?
10. What limitations must be exposed?
```

If these questions cannot be answered, implementation should pause.

---

## 17. Scope Protection Rule

Research must strengthen the project, not silently expand it.

A research idea should be classified by the capability it strengthens:

```text
architecture
Data Quality and Provenance
Trajectory Foundation
Feature Engineering
Context Enrichment
Analytical Core
Confidence and Explainability
```

A useful idea does not automatically belong in the current MVP stage.

---

## 18. Final Principle

Global Flight Analytics should use open research, physics, mathematics, historical evidence, and project-specific reasoning aggressively, but honestly.

The project standard is:

```text
search before inventing,
measure before trusting,
label assumptions,
expose uncertainty,
and never claim knowledge the data does not support.
```
