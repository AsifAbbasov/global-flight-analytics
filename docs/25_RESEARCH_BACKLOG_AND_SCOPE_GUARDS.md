# Document 25 — Research Backlog and Scope Guards

Status: Architecture Baseline v1.1  
Project: Global Flight Analytics  
Scope: Deferred research topics, version boundaries, and product scope guards

---

## 1. Purpose

This document protects Global Flight Analytics from uncontrolled scope growth.

The research audit produced many useful ideas, but not all of them belong in the first implementation. Some require historical data. Some require stronger evaluation. Some require specialized data sources that are outside the current budget and open-data scope.

---

## 2. Core Scope Rule

The project must build in this order:

```text
Reliable open-data trajectory first.
Validated features second.
Historical analytics third.
Prediction fourth.
Advanced research last.
```

Any feature that violates this order must be deferred.

---

## 3. Research Backlog

The following topics are allowed as future research, but they are not part of the MVP:

```text
1. CNN-LSTM 4D Trajectory Prediction
2. B-STAR Bayesian Spatio-Temporal Graph Transformer
3. Sobolev Functional Regression
4. Wavelet-Based Trajectory Compression
5. Exact Fréchet Verification
6. C++ Similarity Benchmark Tool
7. Full Weather-Aware Deep Prediction
8. Full Contrail Climate Model
9. CocipGrid-style Precomputed Climate Grid
10. Climate-Optimal Routing
11. Fuel Burn Prediction
12. Emission Prediction
13. Actual Take-Off Weight Prediction
14. ADS-C Ingestion
15. Satellite Data Fusion
16. FLARM Low-Altitude Traffic Layer
17. Object Classification for Non-Airliner Traffic
18. Regulated Operational Aviation Modules
19. Advanced Optimization Engine
20. Experimental Reinforcement Learning
```

---

## 4. Deferred Machine Learning Topics

Deferred items:

```text
CNN-LSTM 4D Trajectory Prediction
B-STAR Bayesian Spatio-Temporal Graph Transformer
Sobolev Functional Regression
Experimental Reinforcement Learning
```

Reason:

These methods require mature historical datasets, evaluation pipelines, and stable feature engineering. Implementing them before the MVP would create complexity without reliable value.

Entry condition:

```text
historical trajectory storage exists
feature store exists
replay evaluation exists
baseline projection exists
baseline error metrics exist
```

---

## 5. Deferred Advanced Similarity Topics

Deferred items:

```text
Full Fréchet Similarity Pipeline
Exact Fréchet Verification
C++ Similarity Benchmark Tool
Advanced Spatial Similarity Index
```

Reason:

The project first needs accumulated historical segments and a simple similarity baseline. Advanced similarity optimization is premature before historical data exists.

Entry condition:

```text
historical trajectory segment table is populated
basic similar trajectory selector exists
route pattern library exists
simple similarity metrics are measured
```

---

## 6. Deferred Weather and Climate Topics

Deferred items:

```text
Full Weather-Aware Deep Prediction
Full Weather Grid Analytics
Contrail Climate Model
CocipGrid-style Precomputed Climate Grid
Climate-Optimal Routing
```

Reason:

Weather can be used as an uncertainty modifier earlier, but full climate-aware routing and contrail modeling require specialized datasets, stronger validation, and careful scientific framing.

Entry condition:

```text
weather provider exists
weather trust gate exists
weather-trajectory alignment exists
projection evaluation exists
scientific assumptions are documented
```

---

## 7. Deferred Source Expansion Topics

Deferred items:

```text
ADS-C Ingestion
Satellite Data Fusion
FLARM Low-Altitude Traffic Layer
Object Classification for Non-Airliner Traffic
```

Reason:

The MVP should work with the available open data sources first. Additional source types require source-specific parsing, coverage rules, provenance logic, and new quality models.

Entry condition:

```text
provider interface is stable
source provenance is implemented
trajectory segments support source status
coverage gap model is implemented
fusion architecture is documented
```

---

## 8. Deferred Fuel, Emission, and Weight Topics

Deferred items:

```text
Fuel Burn Prediction
Emission Prediction
Actual Take-Off Weight Prediction
```

Reason:

These models require trustworthy aircraft features, route features, weather context, and often ground truth that open sources may not provide.

Entry condition:

```text
aircraft feature provider exists
feature store exists
trajectory quality is stable
route distance and flown distance are computed
weather context is available
model limitations are documented
```

---

## 9. Product Scope Guard

The project must not be positioned as regulated aviation software.

Forbidden positioning:

```text
regulated aviation product
flight planning product
dispatch product
official aeronautical source
commercial aviation data replacement
```

Allowed positioning:

```text
open-data research platform
aviation analytics demo
trajectory quality visualization
historical pattern analysis
non-operational airspace research dashboard
portfolio-grade engineering project
```

---

## 10. Airspace Research Scope Guard

Airspace interaction analytics may be implemented only as research visualization.

Allowed:

```text
research-level proximity analysis
historical proximity visualization
sector complexity estimate
temporal occupancy index
non-operational what-if visualization
```

Forbidden:

```text
regulated operational recommendations
real-world aviation instructions
official safety assurances
claims of certified aviation-grade correctness
```

---

## 11. Prediction Scope Guard

Prediction outputs must be conservative.

Allowed prediction formats:

```text
probable corridor
confidence-scored projection
estimated range
low-confidence refusal
historical pattern continuation
```

Forbidden prediction formats:

```text
guaranteed future position
official route claim
exact operator intent claim
exact cause of maneuver without supporting data
certified aviation-grade forecast claim
```

---

## 12. Weather Scope Guard

Weather may modify uncertainty. Weather must not be used as automatic proof of cause.

Allowed:

```text
weather encounter profile
weather-adjusted uncertainty
weather context display
confidence reduction due to weather mismatch
```

Forbidden:

```text
claiming exact maneuver cause from weather alone
claiming operator intent from weather alone
claiming official rerouting reason without official data
```

---

## 13. Open Data Scope Guard

Every analytical output must respect open data limitations.

Mandatory limitations:

```text
No official flight plan access
No operational instruction access
No commercial aviation feed access
No guaranteed global coverage
No guaranteed completeness
No regulated aviation use
```

These limitations must be visible in the API and frontend when relevant.

---

## 14. MVP Forbidden List

The following items are explicitly forbidden in the MVP:

```text
machine learning prediction
advanced weather analytics
advanced historical similarity
ADS-C ingestion
FLARM ingestion
fuel prediction
emission prediction
contrail optimization
airspace interaction dashboard
regulated operational modules
```

If any of these appears in MVP planning, it must be removed or moved to backlog.

---

## 15. Backlog Promotion Rules

A backlog item can move into an active version only when all of the following are true:

```text
required data exists
baseline implementation exists
evaluation method exists
limitations are documented
frontend explanation is designed
implementation does not break MVP stability
```

---

## 16. Final Scope Statement

Global Flight Analytics must remain honest about what open data can and cannot support.

The platform should be ambitious, but not misleading.

The correct strategy is:

```text
build reliable trajectory analytics first
add historical intelligence after data accumulates
add prediction only after evaluation exists
add advanced research only after baselines are proven
never claim regulated aviation authority
```
