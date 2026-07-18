# Global Flight Analytics

Global Flight Analytics is an open-data aviation research and analytics platform.

The project is not a flight tracker clone, not regulated aviation software, not a flight planning system, and not a commercial aviation data platform.

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
docs/22_RESEARCH_AUDIT_DEDUPLICATION.md
docs/23_ANALYTICAL_CORE_ARCHITECTURE.md
docs/24_MVP_VERSION_ROADMAP.md
docs/25_IMPLEMENTATION_SEQUENCE.md
docs/26_RESEARCH_BACKLOG_AND_SCOPE_GUARDS.md
docs/27_ENGINEERING_PRINCIPLES.md
docs/28_RESEARCH_AND_ANALYTICAL_DECISION_METHOD.md
docs/29_REPRODUCIBLE_DOCKER.md
docs/30_AIRPORT_INTELLIGENCE_IMPLEMENTATION_ALIGNMENT.md
docs/31_STAGE_8_HISTORICAL_INTELLIGENCE_COMPLETION.md
docs/32_STAGE_9_PROJECTION_AND_ESTIMATED_TIME_OF_ARRIVAL_COMPLETION.md
docs/33_STAGE_10_WEATHER_CONTEXT_COMPLETION.md
docs/34_STAGE_11_AIRSPACE_INTELLIGENCE_COMPLETION.md
docs/35_STAGE_12_STABILITY_AND_EXPLAINABILITY_COMPLETION.md
```

Existing foundation documents remain in `docs/01_*` through `docs/21_ENGINEERING_AMENDMENTS_v1.1.md`.

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

<!-- SOURCE-CONSTRAINTS-OPENSKY-V1 -->
## Free Data and Evidence Boundaries

The immutable project constraints and OpenSky integration boundary are documented in:

```text
docs/36_FREE_DATA_SOURCE_AND_EVIDENCE_BOUNDARIES.md
```

Executable enforcement lives in `apps/api/internal/analytics/sourceconstraints`.
The bounded OpenSky REST contract foundation lives in `apps/api/internal/integrations/opensky`.

<!-- OPENSKY-VALIDITY-ATTRIBUTION-V1 -->
### OpenSky temporal validity and publication boundary

OpenSky is an optional external research provider, not project-owned surveillance infrastructure. Public outputs using OpenSky data must preserve the required provider citation and non-commercial research scope. State Vector fields may have different source timestamps; a position is accepted as provider-valid only within the documented fifteen-second reuse window. Access from large cloud-hosting IP ranges is not guaranteed, so OpenSky must remain behind provider health, budget, and fallback controls.

<!-- OPENSKY-PRODUCTION-PROVIDER-V1 -->
## Selectable production traffic provider

The ingestion daemon can use either `airplanes.live` or OpenSky through the same provider budget, request coalescing, health, data quality, and trajectory pipeline. `airplanes.live` remains the default. OpenSky is enabled explicitly with `TRAFFIC_PROVIDER=opensky` and remains bounded by the free-data and non-commercial research constraints.

<!-- TRAFFIC-PROVIDER-AUTOMATIC-FALLBACK-V1 -->
### Automatic free-provider fallback

`TRAFFIC_PROVIDER=auto` enables an ordered, budget-aware fallback from
`airplanes.live` to OpenSky. The secondary provider is called only after a
recoverable primary failure or access denial. Successful ingestion runs and
canonical states preserve the provider that actually supplied the data.
See `docs/38_TRAFFIC_PROVIDER_AUTOMATIC_FALLBACK.md`.

<!-- OPENSKY-REST-COMPATIBILITY-V1 -->
## OpenSky REST compatibility

Production OpenSky State Vector requests include `extended=1`. The parser accepts both the seventeen-field base representation and the eighteen-field extended representation without inventing a provider category.

<!-- OPEN-AVIATION-RESEARCH-EVIDENCE-V1-2:README -->

## Open Aviation Research Evidence Foundation

The backend now preserves transponder and OpenSky observation metadata in the canonical `FlightState`, provides research-only Transponder Alert Evidence, and enforces bounded offline manifests for selected external aviation research datasets. Satellite ADS-C evidence, automatic bulk imports, confirmed-incident claims, and production dependencies remain blocked.

<!-- STAGE-14-1-ARCHITECTURE-CONSOLIDATION-V1-1:README -->

## Architecture Consolidation

Stage 14 establishes one shared confidence vocabulary, checks Go and TypeScript trajectory contracts, audits analytical package reachability from real runtime roots, and adds supply-chain quality gates. Packages are not deleted until the reachability report distinguishes production, operational, verification, offline research, and genuinely obsolete code.

<!-- STAGE-14-2-DEAD-CODE-CLASSIFICATION:README -->

## Dead Code Classification

Stage 14.2 removes the obsolete `analytics/query` and `analytics/window` foundation packages after importer proofs. Every remaining analytical package outside production runtime now requires an explicit disposition and next action; unknown non-runtime packages fail strict project audit.
