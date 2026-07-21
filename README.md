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

The Stage 14 migration closure also executes the complete production catalog against clean PostgreSQL. Migration identities must be canonical, unique, and contiguous; Data Quality parent integrity owns version 019 while the earlier Flight State observation metadata migration retains version 016.

<!-- STAGE-14-2-DEAD-CODE-CLASSIFICATION:README -->

## Dead Code Classification

Stage 14.2 removes the obsolete `analytics/query` and `analytics/window` foundation packages after importer proofs. Every remaining analytical package outside production runtime now requires an explicit disposition and next action; unknown non-runtime packages fail strict project audit.

<!-- STAGE-14-3-AIRPORT-INTELLIGENCE-PRODUCTION:README -->

## Airport Intelligence Production API

Airport Passport, Statistics, Ranking, Overview, History, and Trends are composed through a PostgreSQL-backed read-only production service. The API uses completed Coordinated Universal Time days and exposes explicit open-data limitations.

<!-- STAGE-14-4-FEATURE-MATERIALIZATION:README -->

## Flight Feature Materialization

Persisted trajectories can now be processed through the complete Feature Pipeline with `materialize-flight-features`. The command uses real PostgreSQL trajectory and aircraft data and stores idempotent snapshots in `flight_feature_snapshots`. The isolated in-memory dataset profiler was removed rather than falsely connected.

<!-- STAGE-14-5-MUTATION-ENDPOINT-PROTECTION:README -->

## Mutation Endpoint Protection

Public read routes remain unauthenticated. Every state-changing or computation-triggering HTTP route requires the backend-only `X-Internal-API-Key` header. The backend stores only `API_MUTATION_KEY_SHA256`, compares presented credentials in constant time, and refuses database-backed production configuration without the digest.

<!-- STAGE-14-6-FORMULA-BENCHMARK:README -->

## Formula Benchmark and Calibration Gate

Projection formulas can now be evaluated through a bounded offline command that consumes an approved research dataset manifest and an immutable Projection Evaluation aggregate. Reports distinguish insufficient evidence, failed gates, and passed gates while permanently prohibiting automatic formula changes or calibration claims.

<!-- STAGE-14-7-FRONTEND-DEPENDENCY-SECURITY:README -->

## Frontend Dependency Security

The pnpm workspace now redirects PostCSS versions below 8.5.10 to 8.5.15, verifies the committed dependency graph without network access, and blocks moderate or more severe production dependency findings in frontend continuous integration.

<!-- STAGE-14-8-SERVER-COMPOSITION-ROOT-DECOMPOSITION:README -->

## Server Composition Root Decomposition

Database-backed server wiring is now organized by bounded context. The coordinator describes startup order, composition files construct dependencies, and route files register HTTP topology. Existing methods, paths, contracts, and mutation protection remain unchanged.

<!-- STAGE-14-9-HTTP-QUERY-CONTRACT-BOUNDARY:README -->

## Historical Intelligence Contract Boundary

Historical Intelligence HTTP handlers and DTO conversion now depend on a pure aggregate store contract rather than the PostgreSQL implementation. Latest and history query parsing use separate intent-revealing functions without boolean mode arguments.

<!-- STAGE-14-10-TRANSPONDER-EVIDENCE-PRODUCTION:README -->

## Transponder Evidence Production API

The production server now exposes the latest persisted special transponder code as read-only research evidence. Responses explicitly state that the observation is evidence-only, does not confirm an emergency, and is not an operational alert.

<!-- STAGE-14-11-TARGETED-LARGE-MODULE-HARDENING:README -->

## Targeted Large-Module Hardening

Historical and Route Intelligence validation are divided by contract responsibility. Projection continuation and estimated-arrival public methods are narrow coordinators, while detailed preparation, computation, fallback, confidence, provenance, and result construction remain isolated and testable.

<!-- STAGE-14-12-PROJECTION-READ-SNAPSHOT-CONSISTENCY:README -->

## Projection Read Snapshot Consistency

One Projection Intelligence result now loads its current trajectory, route, historical candidates, and route-frequency history through one PostgreSQL read-only repeatable-read snapshot. Concurrent ingestion or materialization cannot make different input queries observe different committed database states.

<!-- STAGE-14-13-NULLABLE-TELEMETRY-INTEGRITY:README -->

## Nullable Telemetry Integrity

Projection Intelligence no longer converts absent coordinates, motion telemetry, or on-ground state into plausible zero or false values. Only complete required kinematic observations become analytical trajectory points; legitimate stored zero values remain valid.

<!-- STAGE-14-14-COMPOSITE-HISTORICAL-PAGINATION-V3:README -->

## Lossless Historical Pagination

Historical Intelligence history now uses a versioned opaque `cursor` token that carries the complete PostgreSQL ordering boundary: window end, window start, as-of time, and record identifier. Store, HTTP response, handler parsing, and runtime verification use the same lossless keyset contract.

<!-- STAGE-14-15-WEATHER-COMPOSITION-BOUNDARY:README -->

## Weather Composition Boundary

Weather production wiring now separates provider governance and integration, PostgreSQL-backed application construction, dependency coordination, and Fiber route registration. The current weather endpoint and runtime dependency graph remain unchanged.

<!-- BACKEND-FINAL-CORRECTNESS-AUDIT:README -->

## Backend Final Correctness Audit

The repository now includes a permanent backend correctness gate for Projection read snapshot consistency, nullable telemetry integrity, Historical pagination, and Weather composition. Run `scripts/verify-backend-final-correctness.sh` before backend release or architecture-sensitive changes.

<!-- STAGE-14-16-END-TO-END-TELEMETRY-AVAILABILITY:README -->

## End-to-End Telemetry Availability

Flight State now preserves velocity, heading, vertical-rate, and on-ground availability from provider mapping through PostgreSQL persistence and downstream analytical reads. Missing provider telemetry remains `NULL`; real zero values remain valid observations.

<!-- STAGE-14-29-MIGRATION-CATALOG-INTEGRITY:README -->

## Stage 14 Migration Catalog Integrity

The duplicate PostgreSQL migration version `016` is removed by assigning Data Quality
Parent Integrity to canonical migration `019`. The production `cmd/migrate` path is now
part of the permanent PostgreSQL gate. Stage 14 remains reopened while the remaining
correctness and maintainability debts are addressed through separate verified increments.

<!-- STAGE-14-30-POSTGRES-CORRECTNESS-HARDENING:README -->

## Stage 14 PostgreSQL Correctness Hardening

Migration 020 now enforces Ingestion Run evidence consistency and exact timestamp
mirror contracts for Route and Historical results. Transactional repository writes use
an independent bounded rollback context. Stage 14 remains reopened for the separate
maintainability and Clean Code backlog recorded in Document 72.

<!-- STAGE-14-31-POSTGRES-WRITE-REPOSITORY-DECOMPOSITION:README -->

## Stage 14 PostgreSQL Write Repository Decomposition

Airport Import and Flight State PostgreSQL write paths now keep transaction coordination
separate from staging, merge, validation, mapping, and insert ownership. Permanent Go parser
and Stage 14 audit gates prevent the coordinator methods from becoming monoliths again.
Stage 14 remains reopened for the separate Airport pagination contract.

<!-- STAGE-14-32-AIRPORT-KEYSET-PAGINATION:README -->

## Stage 14 Airport Keyset Pagination

Airport reads now expose bounded keyset pages ordered by the complete `(name, id)` key.
The legacy complete-list method delegates to those pages, and `ListPage` plus `GetByICAO`
share one canonical row scanner. Stage 14 remains reopened for the remaining recorded scope.

<!-- STAGE-14-33-EXPLICIT-REPOSITORY-CONTEXT-AND-TRAJECTORY-WRITE-MODE:README -->

## Stage 14 Explicit Repository Context and Trajectory Write Mode

Database-reaching Airport, Flight State, and Trajectory repository paths now reject a nil
caller context instead of silently inventing `context.Background()`. Trajectory persistence
uses an explicit live or reconciled write request rather than an empty task identifier as a
hidden mode switch. Stage 14 remains reopened for the remaining recorded scope.

<!-- STAGE-14-34-POSTGRESQL-CONTRACT-CONSOLIDATION:README -->

## Stage 14 PostgreSQL Contract Consolidation

Migration repair now derives its anchor checksum and later-version boundary from a repair
plan. Repository nullable arguments are concrete driver values, missing source evidence fails
closed instead of becoming `unknown`, and UUID array queries keep UUID columns typed. Stage 14
remains reopened for trajectory profiling and the final closure audit.
