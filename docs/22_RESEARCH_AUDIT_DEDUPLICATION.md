# Document 22 — Research Audit Deduplication

Status: Architecture Baseline v1.1  
Project: Global Flight Analytics  
Scope: Deduplicated research audit results for architecture and analytical core planning

---

## 1. Purpose

This document consolidates the useful engineering ideas extracted from the research audit work. The goal is not to preserve every article note, every formula, or every experimental method. The goal is to convert the research audit into a clean, deduplicated architecture for Global Flight Analytics.

The project must not become a collection of unrelated research fragments. Every extracted idea must either strengthen the platform architecture, strengthen the analytical core, or be deferred to the research backlog.

---

## 2. Main Deduplicated Principle

Global Flight Analytics is not centered around a map, a raw message, or a single aircraft point.

The central analytical object is:

```text
FlightTrajectory
```

The core pipeline is:

```text
Open Data
↓
Clean Trajectory
↓
Reliable Features
↓
Historical Patterns
↓
Context-Aware Analytics
↓
Confidence
↓
Explanation
↓
Visualization
```

This means that raw positions are only input material. They are not the final domain model.

---

## 3. Deduplicated System Pipeline

All research ideas converge into the following platform pipeline:

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

This pipeline replaces the weaker model:

```text
Provider
↓
Raw Point
↓
Map
```

The project must not build analytics directly on raw provider responses.

---

## 4. Deduplicated Architecture Layers

### 4.1 Source Adapter Layer

Purpose: isolate external data source formats from the internal domain model.

Included modules:

```text
OpenSky Provider
Airplanes.live Provider
OurAirports Provider
OpenStreetMap Provider
Wikidata Provider
Future Provider Interface
```

Decision: each provider is a source of partial observations, not a source of truth.

### 4.2 Canonical Data Layer

Purpose: convert raw provider data into stable domain entities.

Canonical entities:

```text
FlightState
TrackPoint4D
Waypoint
TrajectorySegment
FlightTrajectory
CoverageGap
```

Decision: `FlightState` is an atomic observation. `FlightTrajectory` is the primary analytical object.

### 4.3 Data Quality and Provenance Layer

Purpose: determine whether data can safely be used for analytics.

Consolidated concerns:

```text
Raw Source Isolation
Unit and Field Normalization
Duplicate Point Removal
Gap and Jump Detection
Motion Plausibility Check
Freshness Score
Field Completeness Score
Sampling Density Score
Track Quality Score
Source Reliability Score
Segment Status Model
Analytics Permission Flags
```

Decision: analytics must be permissioned by quality. Weak data must not produce strong conclusions.

### 4.4 Trajectory Construction Layer

Purpose: build meaningful movement objects from noisy observations.

Included modules:

```text
Track Builder
Flight Splitter
Conservative Track Splitting
Trajectory Resampler
Coverage Gap Detector
Trajectory Quality Evaluator
```

Decision: trajectory construction is not a frontend concern and not a provider concern. It is a backend analytical foundation.

### 4.5 Feature Engineering Layer

Purpose: transform trajectories and reference data into analytical features.

Included modules:

```text
Feature Extractor
Feature Validator
Feature Store
Aircraft Feature Provider
Temporal Feature Builder
Geographical Feature Builder
Operational Feature Builder
Trajectory Feature Builder
Dataset Profiler
```

Decision: serious analytics must run on validated features, not directly on raw database rows.

### 4.6 Context Enrichment Layer

Purpose: attach external and historical context to trajectories.

Included modules:

```text
Airport Context Engine
Region Context Engine
Weather Context Engine
Airspace Context Engine
Historical Pattern Context Engine
```

Decision: context is not decoration. It modifies confidence, interpretation, and analytical output.

### 4.7 Analytical Core

Purpose: produce route, trajectory, airspace, weather, airport, and projection intelligence.

Included modules:

```text
Route Intelligence
Historical Similarity Engine
Historical Pattern-Based Trajectory Intelligence
Weather-Aware Trajectory Intelligence
ETA and Projection Intelligence
Multi-Aircraft Context Intelligence
Separation Risk Intelligence
Airspace Complexity Intelligence
Airport Intelligence
```

Decision: the analytical core must be modular, confidence-aware, and explainable.

### 4.8 Confidence and Explainability Layer

Purpose: explain what the system knows, what it does not know, and why confidence is high or low.

Included modules:

```text
Confidence Engine
Uncertainty Engine
Failure Explanation Engine
Scope Guard
Unknown Intervention Guard
Forecast Versioning
Decision Stability Evaluator
```

Decision: every analytical result must expose its confidence and limitations.

### 4.9 Evaluation Layer

Purpose: test analytical outputs against historical data and replay scenarios.

Included modules:

```text
Replay Engine
Prediction Evaluation Metrics
Similarity Pipeline Evaluation
Weather Alignment Evaluation
Route Confidence Evaluation
Dataset Quality Evaluation
```

Decision: predictions and route inferences must be tested through replay, not only visually inspected.

---

## 5. Deduplicated Naming Decisions

The audit contained many overlapping terms. They are consolidated as follows.

| Repeated names in audit notes | Final architecture name |
|---|---|
| Track Quality, Trajectory Quality, Data Quality, Source Quality, Position Quality | Data Quality and Provenance Layer |
| Coverage Score, Coverage Gap, Data Availability, Sampling Density | Coverage and Continuity Model |
| Feature Extractor, Feature Store, Feature Repository, Feature Validator | Feature Engineering Layer |
| Weather Engine, Weather Feature Engine, Weather Trust, Weather Source Tiering | Weather-Aware Trajectory Intelligence |
| Forecast Uncertainty, Prediction Confidence, Confidence Score, Confidence Propagation | Confidence and Uncertainty Engine |
| Similarity Pipeline, Fréchet Engine, Historical Similarity, Pattern Selector | Historical Trajectory Similarity Engine |
| Pattern Library, Historical Route Patterns, Neighbor-Based Forecast | Historical Pattern-Based Trajectory Intelligence |
| Flight Plan Guard, Scope Guard, OpenSky Guard, Unknown Intervention Guard | Scope and Limitation Guard |
| Replay Testing, Evaluation Metrics, Dataset Evaluation, Model Evaluation | Evaluation and Replay Engine |
| Decision Stability, Forecast Stability, Forecast Versioning | Forecast and Decision Stability Layer |

---

## 6. What Is Accepted Into the Core Architecture

```text
Trajectory-centered domain model
Source-aware canonical states
Data quality before analytics
Coverage gaps as first-class domain facts
Trajectory segments with status and provenance
Feature engineering as a separate layer
Historical patterns as a separate analytical layer
Weather as uncertainty context, not proof of cause
Confidence and explanation as mandatory outputs
Replay and evaluation before advanced prediction claims
```

---

## 7. What Is Deferred

```text
CNN-LSTM prediction
Bayesian graph transformer prediction
Sobolev functional regression
Wavelet regression
Full Fréchet verification
ADS-C fusion
FLARM ingestion
Fuel burn prediction
Emission prediction
Contrail climate optimization
Regulated operational aviation modules
```

They belong to the research backlog until the basic trajectory pipeline, feature layer, historical data, and evaluation layer exist.

---

## 8. Final Deduplicated Architecture Statement

Global Flight Analytics is an open-data aviation research and analytics platform.

It is not a flight tracker clone.

Its value comes from:

```text
trajectory quality
feature extraction
historical pattern analysis
route intelligence
context-aware uncertainty
confidence scoring
explainable analytics
map-based visualization
```

Implementation must follow this order:

```text
Clean data first.
Reliable trajectory second.
Features third.
Analytics fourth.
Prediction last.
```
