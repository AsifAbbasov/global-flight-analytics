# Documentation Index — Global Flight Analytics

Status: Documentation Index v2.0
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

<!-- OPENSKY-REST-COMPATIBILITY-V1 -->
## Document 39

`39_OPENSKY_REST_COMPATIBILITY_HARDENING.md` records the extended category request and backward-compatible State Vector parsing contract.

<!-- OPEN-AVIATION-RESEARCH-EVIDENCE-V1-2:DOCUMENT-INDEX -->

## Document 40

`40_OPEN_AVIATION_RESEARCH_EVIDENCE_FOUNDATION.md`

Defines canonical observation metadata preservation, bounded Transponder Alert Evidence, selected scientific dataset governance, blocked ADS-C evidence, manifest gates, and reproducible offline benchmark contracts.

<!-- STAGE-14-1-ARCHITECTURE-CONSOLIDATION-V1-1:DOCUMENT-INDEX -->

## Document 41

`41_STAGE_14_1_ARCHITECTURE_CONSOLIDATION_FOUNDATION.md`

Defines the shared confidence vocabulary, Go and TypeScript trajectory contract audit, analytical production reachability evidence, supply-chain gates, and the authentication boundary for the consolidation stage.

<!-- STAGE-14-2-DEAD-CODE-CLASSIFICATION:DOCUMENT-INDEX -->

## Document 42

`42_STAGE_14_2_DEAD_CODE_CLASSIFICATION_AND_REMOVAL.md`

Records importer-proven removal of obsolete Analytical Core foundation packages and the mandatory release disposition of every remaining non-runtime analytical package.

<!-- STAGE-14-3-AIRPORT-INTELLIGENCE-PRODUCTION:DOCUMENT-INDEX -->

## Document 43

`43_STAGE_14_3_AIRPORT_INTELLIGENCE_PRODUCTION_INTEGRATION.md`

Records PostgreSQL composition, read-only HTTP routes, completed-day window semantics, ranking limitations, security boundary, and runtime completion evidence for Airport Intelligence.

<!-- STAGE-14-4-FEATURE-MATERIALIZATION:DOCUMENT-INDEX -->

## Document 44

`44_STAGE_14_4_FEATURE_MATERIALIZATION_AND_PROFILER_REMOVAL.md`

Records the real PostgreSQL Flight Feature materialization command, deterministic selector and as-of semantics, container runtime boundary, full Feature Pipeline reachability, and importer-proven removal of the unused dataset profiler.

<!-- STAGE-14-5-MUTATION-ENDPOINT-PROTECTION:DOCUMENT-INDEX -->

## Document 45

`45_STAGE_14_5_MUTATION_ENDPOINT_PROTECTION.md`

Defines the backend-only mutation credential digest, constant-time request authorization, fail-closed production configuration, frontend separation, rotation process, and architecture gate for all state-changing HTTP methods.

<!-- STAGE-14-6-FORMULA-BENCHMARK:DOCUMENT-INDEX -->

## Document 46

`46_STAGE_14_6_FORMULA_BENCHMARK_AND_CALIBRATION_GATE.md`

Defines the bounded projection formula benchmark plan, deterministic report, evidence and performance gates, exit codes, production separation, and the prohibition on automatic calibration.

<!-- STAGE-14-7-FRONTEND-DEPENDENCY-SECURITY:DOCUMENT-INDEX -->

## Document 47

`47_STAGE_14_7_FRONTEND_DEPENDENCY_SECURITY_REMEDIATION.md`

Records the PostCSS vulnerability root cause, targeted pnpm workspace override, lockfile security policy, continuous integration threshold, compatibility checks, and prohibited unsafe remediation methods.

<!-- STAGE-14-8-SERVER-COMPOSITION-ROOT-DECOMPOSITION:DOCUMENT-INDEX -->

## Document 48

`48_STAGE_14_8_SERVER_COMPOSITION_ROOT_DECOMPOSITION.md`

Defines the bounded-context server composition structure, preserved HTTP behavior, architecture boundaries, topology regression gates, and intentionally excluded dependency-injection complexity.

<!-- STAGE-14-9-HTTP-QUERY-CONTRACT-BOUNDARY:DOCUMENT-INDEX -->

## Document 49

`49_STAGE_14_9_HTTP_QUERY_AND_CONTRACT_BOUNDARY_HARDENING.md`

Records the removal of boolean query modes, the pure Historical Intelligence aggregate store contract, HTTP error-boundary rules, compatibility strategy, regression gates, and intentionally rejected mechanical refactors.

<!-- STAGE-14-10-TRANSPONDER-EVIDENCE-PRODUCTION:DOCUMENT-INDEX -->

## Document 50

`50_STAGE_14_10_TRANSPONDER_EVIDENCE_PRODUCTION_INTEGRATION.md`

Defines the read-only production endpoint, safety semantics, freshness policy, qualitative confidence boundary, dependency wiring, reachability governance, exclusions, and acceptance gates for observed special transponder code evidence.

<!-- STAGE-14-11-TARGETED-LARGE-MODULE-HARDENING:DOCUMENT-INDEX -->

## Document 51

`51_STAGE_14_11_TARGETED_LARGE_MODULE_HARDENING.md`

Records the targeted audit findings, responsibility-based source decomposition, projection workflow extraction, preserved behavior, rejected mechanical refactors, regression gates, and final acceptance criteria for Backend Architecture Hardening.

<!-- STAGE-14-12-PROJECTION-READ-SNAPSHOT-CONSISTENCY:DOCUMENT-INDEX -->

## Document 52

`52_STAGE_14_12_PROJECTION_READ_SNAPSHOT_CONSISTENCY.md`

Defines the PostgreSQL repeatable-read snapshot boundary, transaction-scoped trajectory repository, service contract, lifecycle behavior, preserved semantics, regression gates, and acceptance evidence for reproducible Projection Intelligence input loading.

<!-- STAGE-14-13-NULLABLE-TELEMETRY-INTEGRITY:DOCUMENT-INDEX -->

## Document 53

`53_STAGE_14_13_NULLABLE_TELEMETRY_INTEGRITY.md`

Defines the nullable telemetry failure mode, conservative completeness boundary, legitimate-zero semantics, altitude handling, ordering and limit behavior, preserved contracts, regression gates, and acceptance evidence.

<!-- STAGE-14-14-COMPOSITE-HISTORICAL-PAGINATION-V3:DOCUMENT-INDEX -->

## Document 54

`54_STAGE_14_14_COMPOSITE_HISTORICAL_PAGINATION_CURSOR.md`

Defines lossless store and HTTP keyset pagination, opaque cursor encoding, recovery from the failed first installer, validation rules, removed legacy names, preserved behavior, regression gates, and acceptance evidence.

<!-- STAGE-14-15-WEATHER-COMPOSITION-BOUNDARY:DOCUMENT-INDEX -->

## Document 55

`55_STAGE_14_15_WEATHER_COMPOSITION_BOUNDARY.md`

Defines the former mixed Weather composition problem, responsibility-specific server files, preserved dependency graph and endpoint behavior, same-package decomposition rationale, regression gates, and acceptance evidence.

<!-- BACKEND-FINAL-CORRECTNESS-AUDIT:DOCUMENT-INDEX -->

## Document 56

`56_BACKEND_FINAL_CORRECTNESS_AUDIT.md`

Defines the permanent backend correctness gate, protected Stage 14 invariants, existing architecture and security checks, runtime verifier coverage, race detection scope, reproducible command, non-goals, and acceptance evidence.

<!-- STAGE-14-16-END-TO-END-TELEMETRY-AVAILABILITY:DOCUMENT-INDEX -->

## Document 57

`57_STAGE_14_16_END_TO_END_TELEMETRY_AVAILABILITY.md`

Defines the provider-to-PostgreSQL-to-analytics telemetry availability contract, legacy compatibility rule, OpenSky optional mapping, nullable persistence, Traffic and Airspace eligibility, validation behavior, expanded final audit, and acceptance evidence.

<!-- STAGE-14-17-POSTGRES-MIGRATION-ATOMICITY:DOCUMENT-INDEX -->

## Document 58

`58_STAGE_14_17_POSTGRES_MIGRATION_ATOMICITY.md`

Defines the atomic PostgreSQL migration transaction, migration-history coupling, outer transaction-envelope handling, advisory lock serialization, failure rollback behavior, regression gates, and acceptance evidence.

<!-- STAGE-14-18-POSTGRES-BASELINE-REMOVAL:DOCUMENT-INDEX -->

## Document 59

`59_STAGE_14_18_POSTGRES_BASELINE_REMOVAL.md`

Records removal of the unsafe migration baseline operation, preserved normal migration behavior, supported recovery paths, regression protection, and the completion boundary for trustworthy migration history.

<!-- STAGE-14-19-DATA-QUALITY-PARENT-INTEGRITY:DOCUMENT-INDEX -->

## Document 60

`60_STAGE_14_19_DATA_QUALITY_PARENT_INTEGRITY.md`

Defines canonical Data Quality Report parent integrity, explicit rejected-observation evidence storage, migration of legacy null-parent rows, cascade semantics, repository enforcement, regression gates, and acceptance evidence.

<!-- STAGE-14-20-TRAJECTORY-READ-SNAPSHOT-CONSISTENCY:DOCUMENT-INDEX -->

## Document 61

`61_STAGE_14_20_TRAJECTORY_READ_SNAPSHOT_CONSISTENCY.md`

Defines the repository-owned PostgreSQL read-only repeatable-read boundary for complete FlightTrajectory aggregates, caller-owned transaction compatibility, pool constructor behavior, concurrent mutation evidence, rollback semantics, regression gates, and acceptance evidence.

<!-- STAGE-14-21-INGESTION-RUN-TERMINAL-INTEGRITY:DOCUMENT-INDEX -->

## Document 62

`62_STAGE_14_21_INGESTION_RUN_TERMINAL_INTEGRITY.md`

Defines the one-way Ingestion Run completion transition, explicit transition-rejected semantics, lifecycle shape constraint, terminal-row immutability trigger, concurrent finalization behavior, PostgreSQL integration evidence, regression gates, and acceptance evidence.

<!-- STAGE-14-22-TRAJECTORY-RELATIONAL-INTEGRITY:DOCUMENT-INDEX -->

## Document 63

`63_STAGE_14_22_TRAJECTORY_RELATIONAL_INTEGRITY.md`

Defines mandatory trajectory child ownership, per-trajectory segment ordering, same-trajectory coverage-gap references, deferred stored-count verification, repository fail-fast validation, legacy preflight policy, regression gates, and acceptance evidence.

<!-- STAGE-14-23-CANONICAL-MIGRATION-FILENAME-CONTRACT:DOCUMENT-INDEX -->

## Document 64

`64_STAGE_14_23_CANONICAL_MIGRATION_FILENAME_CONTRACT.md`

Defines the single canonical PostgreSQL migration filename parser shared by execution, audit, and repair verification, strict version and name rules, removed duplicate parsers, preserved behavior, regression ownership gates, and acceptance evidence.

<!-- STAGE-14-24-EXPLICIT-ALTITUDE-INTEGER-POLICY:DOCUMENT-INDEX -->

## Document 65

`65_STAGE_14_24_EXPLICIT_ALTITUDE_INTEGER_POLICY.md`

Defines the explicit whole-metre altitude persistence policy, deterministic rounding and integer-range rules, non-finite value rejection, preserved typed altitude status semantics, removal of SQL-owned conversion, regression gates, and acceptance evidence.

<!-- STAGE-14-25-TRAFFIC-ALTITUDE-STATUS-SEMANTICS:DOCUMENT-INDEX -->

## Document 66

`66_STAGE_14_25_TRAFFIC_ALTITUDE_STATUS_SEMANTICS.md`

Defines typed current-traffic altitude selection, observed-zero preservation, geometric-to-barometric fallback, nullable absence semantics, explicit ground handling, HTTP contract propagation, frontend presentation, regression gates, and acceptance evidence.

<!-- STAGE-14-26-AIRPORT-ELEVATION-SEMANTICS:DOCUMENT-INDEX -->

## Document 67

`67_STAGE_14_26_AIRPORT_ELEVATION_SEMANTICS.md`

Defines nullable airport elevation semantics from PostgreSQL through Airport profiles, Airport Intelligence, route context, production Route Intelligence, and frontend presentation, including observed sea-level values, unknown evidence, regression gates, and acceptance evidence.

<!-- STAGE-14-27-FLIGHT-FEATURE-TIMESTAMP-CONSISTENCY:DOCUMENT-INDEX -->

## Document 68

`68_STAGE_14_27_FLIGHT_FEATURE_TIMESTAMP_CONSISTENCY.md`

Defines exact Unix-nanosecond ownership for Flight Feature snapshot identity,
PostgreSQL timestamp mirror validation, permitted sub-microsecond precision loss,
fail-closed corruption handling, regression gates, and acceptance evidence.

<!-- STAGE-14-28-POSTGRES-TRAJECTORY-REPOSITORY-DECOMPOSITION:DOCUMENT-INDEX -->

## Document 69

`69_STAGE_14_28_POSTGRES_TRAJECTORY_REPOSITORY_DECOMPOSITION.md`

Defines responsibility-based decomposition of the PostgreSQL Trajectory Repository write and read paths, preserves the public repository contract and snapshot semantics, moves relational-integrity source ownership, adds permanent anti-monolith gates, and closes the final known PostgreSQL maintainability debt.

<!-- STAGE-14-FINAL-COMPLETION-AUDIT:DOCUMENT-INDEX -->

## Document 70

`70_STAGE_14_FINAL_COMPLETION_AUDIT.md`

Defines the unified cross-stack Stage 14 acceptance gate, patched Go 1.26.5 toolchain ownership, continuous integration reachability, isolated PostgreSQL integration for repository and Flight Feature timestamp semantics, dependency security, frontend production validation, backend container health evidence, final source governance, and the formal completion marker.

<!-- STAGE-14-29-MIGRATION-CATALOG-INTEGRITY:DOCUMENT-INDEX -->

## Document 71

`71_STAGE_14_29_MIGRATION_CATALOG_INTEGRITY.md`

Records the confirmed duplicate migration-version blocker, canonical renumbering of Data
Quality Parent Integrity to version 019, real repository-catalog validation through the
production migrator, retirement of the false completion marker, and the explicit
reopening of Stage 14.

<!-- STAGE-14-30-POSTGRES-CORRECTNESS-HARDENING:DOCUMENT-INDEX -->

## Document 72

`72_STAGE_14_30_POSTGRES_CORRECTNESS_HARDENING.md`

Defines Ingestion Run processed-count and error-evidence invariants, Route and Historical
timestamp mirror integrity, independent bounded repository rollback contexts, migration
020, isolated production-catalog integration evidence, and the remaining reopened scope.

<!-- STAGE-14-31-POSTGRES-WRITE-REPOSITORY-DECOMPOSITION:DOCUMENT-INDEX -->

## Document 73

`73_STAGE_14_31_POSTGRES_WRITE_REPOSITORY_DECOMPOSITION.md`

Defines responsibility-based decomposition of Airport Import and Flight State PostgreSQL
write paths, preserved public and transactional behavior, dedicated SQL and preparation
owners, parser-backed anti-monolith gates, acceptance evidence, and the separate pagination
contract boundary.

<!-- STAGE-14-32-AIRPORT-KEYSET-PAGINATION:DOCUMENT-INDEX -->

## Document 74

`74_STAGE_14_32_AIRPORT_KEYSET_PAGINATION.md`

Defines the bounded Airport page contract, stable `(name, id)` keyset cursor, legacy
complete-list adapter, canonical row scanner, duplicate-name PostgreSQL integration,
anti-offset regression gates, acceptance evidence, and remaining reopened scope.

<!-- STAGE-14-33-EXPLICIT-REPOSITORY-CONTEXT-AND-TRAJECTORY-WRITE-MODE:DOCUMENT-INDEX -->

## Document 75

`75_STAGE_14_33_EXPLICIT_REPOSITORY_CONTEXT_AND_TRAJECTORY_WRITE_MODE.md`

Defines caller-owned PostgreSQL repository context semantics, the intentionally independent
rollback context, explicit live and reconciled Trajectory write requests, invalid-mode
validation, preserved behavior, permanent regression gates, and remaining reopened scope.

<!-- STAGE-14-34-POSTGRESQL-CONTRACT-CONSOLIDATION:DOCUMENT-INDEX -->

## Document 76

`76_STAGE_14_34_POSTGRESQL_CONTRACT_CONSOLIDATION.md`

Defines repository-derived migration repair planning, concrete nullable database arguments,
required source provenance, native UUID array membership, PostgreSQL integration evidence,
permanent regression gates, and the remaining profiling and closure scope.
77_STAGE_14_35_TRAJECTORY_QUERY_CONSOLIDATION_AND_PROFILING.md

<!-- STAGE-14-36-FINAL-CLOSURE:DOCUMENT-INDEX -->

## Document 78

`78_STAGE_14_36_FINAL_CLOSURE_AUDIT.md`

Defines the committed Stage 14.35 closure baseline, complete Documents 41–78 evidence register,
authoritative cross-stack command, mandatory final markers, anti-reopening regression gate,
preserved boundaries, evidence limitations, and formal Stage 14 closed decision.

<!-- POST-CLOSURE-MIGRATOR-CONTEXT-HARDENING:DOCUMENT-INDEX -->

## Document 79

`79_POST_CLOSURE_MIGRATOR_CONTEXT_HARDENING.md`

Defines explicit nil-context rejection for PostgreSQL migration execution and advisory locking,
preserves bounded independent cleanup contexts, adds permanent regression protection, and
records that Stage 14 remains closed.

## Domain Layer Review Final Closure

- `80_DOMAIN_LAYER_REVIEW_FINAL_CLOSURE.md` — final accepted-finding closure, value-object boundaries, constructor contracts, and explicitly rejected mechanical recommendations.

---

### Document 81 — PostgreSQL Layer Full Audit Closure

Path:

```text
docs/81_POSTGRESQL_LAYER_FULL_AUDIT_CLOSURE.md
```

Purpose:

```text
Classifies and closes every finding from the original PostgreSQL Layer audit,
records fixed, not-applicable and deliberately rejected recommendations, and
binds the closure to executable source, integration, query-plan and runtime-isolation checks.
```

<!-- CODE-REVIEW-STANDARD-V1:DOCUMENT-INDEX -->

### Document 82 — Code Review Standard

Path:

```text
docs/82_CODE_REVIEW_STANDARD.md
```

Purpose:

```text
Defines evidence-based finding severity, mandatory review evidence, explicit non-mechanical interpretation of function length, naming, nullability and engineering principles, pull request review requirements, rejection and deferral rules, and merge closure criteria.
```


<!-- INGESTION-RUN-LIFECYCLE-HARDENING-V1:DOCUMENT-INDEX -->
## Document 83

`83_INGESTION_RUN_LIFECYCLE_HARDENING.md` records bounded terminal status
contexts, stale `running` recovery, startup ownership, configuration, concurrency
safety, verification, and remaining ingestion-layer follow-up boundaries.


<!-- PROVIDER-HTTP-RESILIENCE-HARDENING-V1:DOCUMENT-INDEX -->
## Document 84 — Provider HTTP Resilience Hardening

`84_PROVIDER_HTTP_RESILIENCE_HARDENING.md` defines provider status error preservation, non-destructive successful response observation, bounded JSON and CSV response bodies, typed oversized-response errors, and fallback compatibility with joined errors.

<!-- INGESTION-RETRY-FALLBACK-EVIDENCE-V1:DOCUMENT-INDEX -->
## Document 85

`85_INGESTION_RETRY_AND_FALLBACK_EVIDENCE_HARDENING.md` records provider-directed retry scheduling, bounded exponential backoff, local-denial ingestion-run semantics, ordered fallback attempt evidence, terminal fallback recording, and OpenSky polling reservation ownership.

<!-- OURAIRPORTS-PUBLICATION-LIFECYCLE-V1:DOCUMENT-INDEX -->
## Document 86 — OurAirports Publication Lifecycle Hardening

`86_OURAIRPORTS_PUBLICATION_LIFECYCLE_HARDENING.md` records deterministic
content publication identity, durable PostgreSQL reservation ownership, lease
recovery, commit and release semantics, validator ordering, production import
wiring, concurrency protection, and retry evidence.

<!-- INGESTION-DURABILITY-REPLAY-PARTIAL-V1:DOCUMENT-INDEX -->
## Document 87 — Ingestion Durability, Replay and Partial Status Hardening

`87_INGESTION_DURABILITY_REPLAY_PARTIAL_HARDENING.md` records durable
pre-request ingestion runs, provisional-run deletion for local denials,
selected-source correction, replay-safe Flight State identity, actual insert
counts, explicit partial terminal status, migration catalog repair, verification,
and the remaining open ingestion review boundaries.

<!-- EXACT-DEDUP-AIRPLANESLIVE-TELEMETRY-V1:DOCUMENT-INDEX -->
## Document 88 — Exact Deduplication and Airplanes.live Telemetry Hardening

`88_EXACT_DEDUPLICATION_AND_AIRPLANESLIVE_TELEMETRY_HARDENING.md` records
complete canonical observation equality, persistence-identity exclusions,
nullable Airplanes.live telemetry, explicit availability semantics, bounded
provider time conversion, nil-client protection, and verification evidence.

<!-- PROVIDER-BUDGET-DURABILITY-V1:DOCUMENT-INDEX -->
## Document 89 — Provider Budget Durability and Retry Scheduling

`89_PROVIDER_BUDGET_DURABILITY_AND_RETRY_SCHEDULING.md` records PostgreSQL-owned
fixed-window counters and provider-reported remaining state, cross-process atomic
acquisition, restart-safe cooldown evidence, guaranteed retry scheduling for
exhausted budgets, production wiring, migration 024, verification, and the
remaining health-aware fallback and malformed-batch review boundaries.

<!-- HEALTH-AWARE-TRAFFIC-PROVIDER-SELECTION-V1:DOCUMENT-INDEX -->
## Document 90 — Health-Aware Traffic Provider Selection

`90_HEALTH_AWARE_TRAFFIC_PROVIDER_SELECTION.md` records stable health-ranked
traffic provider ordering, configured-primary preservation, fail-open snapshot
handling, explicit health decision evidence, production collector wiring,
verification, and the remaining malformed-provider-batch review boundary.

<!-- MALFORMED-PROVIDER-BATCH-POLICY-V1:DOCUMENT-INDEX -->
## Document 91 — Malformed Provider Batch Policy

`91_MALFORMED_PROVIDER_BATCH_POLICY.md` records item-level provider rejection,
mixed-batch partial ingestion, fully rejected batch fallback, evidence
propagation, verification and conditional review closure.

<!-- INGESTION-REVIEW-CLOSURE-REPAIR-V1:DOCUMENT-INDEX -->
## Document 92 — Ingestion Review Closure Repair

`92_INGESTION_REVIEW_CLOSURE_REPAIR.md` records bounded duration conversion,
Open-Meteo missing-value preservation, PostgreSQL NULL persistence, the typed
OurAirports atomic publication policy, isolated PostgreSQL fixture alignment,
and the exact Continuous Integration gates required for formal review closure.

<!-- SERVER-HTTP-PROTECTION-REVIEW-CLOSURE:DOCUMENT-INDEX -->

## Document 93

`93_SERVER_AND_HTTP_PROTECTION_REVIEW_CLOSURE.md`

Records the authenticated mutation boundary, explicit liveness and PostgreSQL
readiness contracts, migration-backed container health verification, and
release-blocker closure for the Server and HTTP Protection review.

<!-- SERVER-REVIEW-FULL-CLOSURE:DOCUMENT-INDEX -->

## Document 94

`94_SERVER_REVIEW_FULL_CLOSURE.md`

Records lifecycle correction, final-status request logging, sensitive-error log
protection, read-interface narrowing, rate-limit classification, deferred
deployment risks, and formal full closure of the original Server review.

<!-- TRUSTED-PROXY-BUILD-METADATA-CLOSURE:DOCUMENT-INDEX -->

## Document 95

`95_TRUSTED_PROXY_AND_BUILD_METADATA_CLOSURE.md`

Records fail-closed trusted proxy client identity, spoofing protection,
rate-limiter and logging integration, linker-derived version provenance,
Open Container Initiative labels, container verification, and final resolution
of the two deferred Server review code findings.

<!-- INGESTION-RACE-COVERAGE-CLOSURE:DOCUMENT-INDEX -->

## Document 96

`96_INGESTION_RACE_COVERAGE_CLOSURE.md`

Records permanent Backend Race Safety coverage across the critical Ingestion,
Provider Adapters and Orchestration ownership boundaries.

<!-- ANALYTICAL-CONTRIBUTOR-SEMANTICS-HARDENING:DOCUMENT-INDEX -->

## Document 97

`97_ANALYTICAL_CONTRIBUTOR_SEMANTICS_HARDENING.md`

Records eligibility-before-deduplication ordering, bounded future-observation
skew, finite Traffic Density arithmetic, regression evidence and the remaining
Analytical Core review scope.

<!-- AIRPORT-GEOGRAPHIC-METRIC-INTEGRITY:DOCUMENT-INDEX -->

## Document 98

`98_AIRPORT_AND_GEOGRAPHIC_METRIC_INTEGRITY.md`

Records PostgreSQL-owned airport lookup, server-classified airport movement
geofences, region-owned Traffic Density area, public parameter changes,
regression evidence and the remaining Analytical Core review scope.

<!-- PROVENANCE-ANALYTICAL-TRUST-HARDENING:DOCUMENT-INDEX -->

## Document 99

`99_PROVENANCE_AND_ANALYTICAL_TRUST_HARDENING.md`

Records strict source retrieval timestamps, placeholder-source rejection,
unattributed-source disclosure, default failure sanitization, honest confidence
for request-parameter snapshots, regression evidence and remaining Analytical
Core review scope.

<!-- QUERY-ARCHITECTURE-CONSOLIDATION:DOCUMENT-INDEX -->

## Document 100

`100_QUERY_AND_ARCHITECTURE_CONSOLIDATION.md`

Records zero reference-time rejection, canonical UUID identifiers, canonical
Metric IDs, removal of stored legacy calculator state, hidden executor
dependencies, narrow metric-service behavior, compatibility classification,
regression evidence and remaining Analytical Core closure scope.

<!-- SERVER-OWNED-QUALITY-METRICS:DOCUMENT-INDEX -->

## Document 101

`101_SERVER_OWNED_QUALITY_METRICS.md`

Records removal of caller-owned production snapshots, server-fixed query limits,
ten-second covered-interval Coverage Score evidence, server-derived latest
observation Data Freshness, five-minute stale policy, empty-evidence behavior,
provenance, regression protection and the remaining final closure audit.
<!-- ANALYTICAL-CORE-REVIEW-CLOSURE:DOCUMENT-INDEX -->

## Document 102

`102_ANALYTICAL_CORE_REVIEW_CLOSURE.md`

Classifies all nineteen original Analytical Core findings, records fourteen
fixed findings, three deliberately retained contracts, two rejected
non-blocking mechanical observations, zero deferred or unclassified findings,
the public precision and value-presence contracts, standard-library source
sorting, permanent source audit coverage and formal closure gates.
<!-- NEXT-16-2-11-SECURITY-CLOSURE:DOCUMENT-INDEX -->

## 103. Next.js 16.2.11 Security Closure

```text
docs/103_NEXT_16_2_11_SECURITY_CLOSURE.md
```
<!-- FEATURE-PIPELINE-CONTRACT-INTEGRITY:DOCUMENT-INDEX -->

## 104. Feature Pipeline Review Triage and Contract Integrity

```text
docs/104_FEATURE_PIPELINE_REVIEW_TRIAGE_AND_CONTRACT_INTEGRITY.md
```
<!-- FEATURE-SNAPSHOT-PROCESSING-IDENTITY:DOCUMENT-INDEX -->

## 105. Feature Snapshot Processing Identity

```text
docs/105_FEATURE_SNAPSHOT_PROCESSING_IDENTITY.md
```
<!-- FEATURE-PROCESSING-IDENTITY-POSTGRES-LIST-FIX:DOCUMENT-INDEX -->

## 106. Feature Processing Identity PostgreSQL List Fix

```text
docs/106_FEATURE_PROCESSING_IDENTITY_POSTGRES_LIST_FIX.md
```
<!-- FEATURE-PROCESSING-IDENTITY-TEST-FIXTURE:DOCUMENT-INDEX -->

## 107. Feature Processing Identity Test Fixture Alignment

```text
docs/107_FEATURE_PROCESSING_IDENTITY_TEST_FIXTURE_ALIGNMENT.md
```
<!-- FEATURE-PIPELINE-FINAL-REVIEW-CLOSURE:DOCUMENT-INDEX -->

## 108. Feature Pipeline Validation Audit and Final Closure

```text
docs/108_FEATURE_PIPELINE_VALIDATION_AUDIT_AND_FINAL_CLOSURE.md
```

<!-- EXTRACTOR-COMPOSITION-PROCESSING-IDENTITY:DOCUMENT-INDEX -->

## 109. Extractor Composition Processing Identity

```text
docs/109_EXTRACTOR_COMPOSITION_PROCESSING_IDENTITY.md
```

<!-- AIRCRAFT-METADATA-TEMPORAL-SAFETY:DOCUMENT-INDEX -->

## 110. Aircraft Metadata Temporal Safety

```text
docs/110_AIRCRAFT_METADATA_TEMPORAL_SAFETY.md
```

<!-- EXTRACTOR-COMPOSITION-EXPLICIT-CONFIG:DOCUMENT-INDEX -->

## 111. Extractor Composition Explicit Configuration

```text
docs/111_EXTRACTOR_COMPOSITION_EXPLICIT_CONFIG.md
```

<!-- EXTRACTOR-INPUT-CORRECTNESS-HARDENING:DOCUMENT-INDEX -->

## Document 112 — Extractor Input Correctness Hardening

`112_EXTRACTOR_INPUT_CORRECTNESS_HARDENING.md` records nested event-time cutoff
validation, strict context and typed-nil dependency contracts, invalid evidence
and floating-point rejection, semantic fingerprint normalization, processing
generation version 4, permanent audit coverage, and explicitly deferred
provenance and optional-completeness boundaries.

<!-- EXTRACTOR-QUALITY-PROVENANCE:DOCUMENT-INDEX -->

## Document 113 — Extractor Quality and Provenance Semantics

`113_EXTRACTOR_QUALITY_AND_PROVENANCE_SEMANTICS.md` records separation of
required completeness from optional coverage, removal of the trajectory
`EndTime` provenance fallback, explicit trajectory creation/update timestamps,
aircraft metadata source/version/retrieval provenance, validator generation 2,
processing generation 5, tests, and permanent audit enforcement.

<!-- EXTRACTOR-REVIEW-FINAL-CLOSURE:DOCUMENT-INDEX -->

## Document 114 — Extractor Review Final Closure

`114_EXTRACTOR_REVIEW_FINAL_CLOSURE.md` records centralized ICAO24 identity,
defensive trajectory cloning, schema-derived aircraft field counts, canonical
fingerprint mirror protection, permanent Continuous Integration closure audit,
processing generation 5 stability, and zero open, unclassified, or deferred
extractor-review findings.

<!-- EXTRACTOR-COMPOSITION-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 115 — Extractor Composition Review Hardening

`115_EXTRACTOR_COMPOSITION_REVIEW_HARDENING.md` classifies stale review findings,
persists the typed processing manifest, adds explicit optional aircraft
enrichment and cache-disable policies, preserves per-request historical temporal
filtering, advances processing generation 6, and installs a permanent audit gate.

## Document 116: Aircraft Provider Review Hardening

`116_AIRCRAFT_PROVIDER_REVIEW_HARDENING.md` records atomic request coalescing, independent cancellation, bounded cache lifecycle, domain not-found semantics, lookup identity enforcement, and final aircraft provider review closure.

<!-- FEATURE-STORE-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 117 — Feature Store Review Hardening

`117_FEATURE_STORE_REVIEW_HARDENING.md` records semantic output fingerprinting, versioned persistence data transfer objects, strict Store implementation conformance, complete validation proof, bounded Memory Store capacity, permanent tests and Continuous Integration audit closure.


<!-- FLIGHT-FEATURES-SCHEMA-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 118 — Flight Features Schema Review Hardening

`118_FLIGHT_FEATURES_SCHEMA_REVIEW_HARDENING.md` records geographical schema-model alignment, the fifteen-field completeness denominator, centralized group-count ownership, version-aware schema lookup, processing generation seven, stale and rejected finding classifications, permanent tests and Continuous Integration audit closure.


<!-- TEMPORAL-BUILDER-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 119 — Temporal Builder Review Hardening

`119_TEMPORAL_BUILDER_REVIEW_HARDENING.md` records production segment-boundary temporal evidence fallback, centralized fractional-second duration policy, zero-duration metadata mismatch detection, strict context and cancellation contracts, exact evidence diagnostics, stale as-of finding classification, processing generation eight, permanent tests and Continuous Integration audit closure.


<!-- GEOGRAPHICAL-BUILDER-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 120 — Geographical Builder Review Hardening

`120_GEOGRAPHICAL_BUILDER_REVIEW_HARDENING.md` records chronological point-window filtering, order-independent point snapshots, disconnected segment-path semantics, circular longitude envelope validation, metadata-based fallback support counts, versioned Haversine and decimal-degree cell policies, compensated distance summation, processing generation nine, permanent tests and Continuous Integration closure.


<!-- OPERATIONAL-BUILDER-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 121 — Operational Builder Review Hardening

`121_OPERATIONAL_BUILDER_REVIEW_HARDENING.md` records explicit feature-only repeatable-read point hydration, nullable operational telemetry availability, trajectory-window filtering, deterministic ordering, strict heading and ground-altitude semantics, one-source altitude aggregation, explicit ground-state share denominators, compensated observation-weighted arithmetic, processing generation ten, permanent tests and Continuous Integration closure.
<!-- TRAJECTORY-BUILDER-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 122 — Trajectory Builder Review Hardening

`122_TRAJECTORY_BUILDER_REVIEW_HARDENING.md` records canonical point evidence, unique timestamp sampling, persisted point-count fallback, observation-supported coverage, disconnected path parts, explicit duration and distance policies, dynamic field availability, quality and Validator reconciliation, processing generation eleven, permanent tests, and Continuous Integration closure.

<!-- VALIDATOR-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 123 — Validator Review Hardening

`123_VALIDATOR_REVIEW_HARDENING.md` records strict integrity severity independent of feature availability, explainable partial and unavailable evidence, canonical unavailable payloads, current-evidence quality limitation reconstruction, observation-support ownership, dimensionless relative tolerance semantics, validator generation six, processing generation twelve, permanent regression tests, and Continuous Integration enforcement.


<!-- HISTORICAL-CONTRACT-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 124 — Historical Contract Review Hardening

`124_HISTORICAL_CONTRACT_REVIEW_HARDENING.md` records the single production metric catalog, exact count semantics over the heterogeneous float64 transport, metric-aware precision, confidence and availability reconciliation, comparison-to-summary integrity, complete schema registry coverage, lowercase region normalization, deliberately retained zero-event coverage semantics, permanent regression tests, and Continuous Integration enforcement.


<!-- HISTORICAL-WINDOW-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 125 — Historical Window Review Hardening

`125_HISTORICAL_WINDOW_REVIEW_HARDENING.md` records calendar-safe bucket generation, exact previous-window construction, plan canonicalization and validation, semantic fingerprint generation two, cancellation hardening, accepted review findings, and intentionally retained custom and optional-window contracts.


<!-- HISTORICAL-READ-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 126 — Historical Read Review Hardening

`126_HISTORICAL_READ_REVIEW_HARDENING.md` records repository-owned repeatable-read snapshot consistency, half-open temporal predicates, append-only flight and trajectory version history, event-time route selection with pre-limit latest-version deduplication, exact matched-row coverage, bounded route payload bytes, nullable provenance preservation, explicit numeric rounding, record validation, query-aligned PostgreSQL indexes, permanent tests, and Continuous Integration enforcement.

<!-- HISTORICAL-SERIES-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 127 — Historical Series Review Hardening

`127_HISTORICAL_SERIES_REVIEW_HARDENING.md` records bucket-level coverage evidence, fail-closed provenance timestamps and limitations, unique exclusion evidence, checked sample accumulation, focused builder responsibilities, accepted findings, rejected stale findings, permanent tests, Continuous Integration enforcement, exact engineering commits, GitHub Actions run `30305541816`, and formal closure with zero open, unclassified, or deferred findings.

<!-- HISTORICAL-ROUTE-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 128 — Historical Route Review Hardening

`128_HISTORICAL_ROUTE_REVIEW_HARDENING.md` records global-only route status ratios, complete Route Contract validation, persistence metadata reconciliation, snapshot-plan containment, exact incomplete global coverage, rejection of incomplete route-pair coverage, `StoredAt` fingerprint identity, coordinate-derived distance, scoped provenance, active directional route-pair semantics, compensated arithmetic, accepted findings, deliberately retained contracts, permanent regression tests, engineering commit `513fa1efc7f3b81b895cdc5f881e294d80362e2e`, GitHub Actions run `30334131538`, and formal closure with zero open, unclassified, or deferred findings.


<!-- HISTORICAL-COMPARISON-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 129 — Historical Comparison Review Hardening

`129_HISTORICAL_COMPARISON_REVIEW_HARDENING.md` records coverage-profile comparability, explicit two-period quality evidence, atomic provenance, both-period semantic fingerprinting, explicit scope equality, finite percentage arithmetic, temporal bucket-summary semantics, deliberately retained `float64` and undefined-percentage contracts, permanent regression tests, engineering commit `21734b85b9f50ae717dca031c798866161895989`, GitHub Actions run `30341011740`, and formal closure with zero open, unclassified, or deferred findings.


<!-- HISTORICAL-SIMILARITY-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 130 — Historical Similarity Review Hardening

`130_HISTORICAL_SIMILARITY_REVIEW_HARDENING.md` records explicit similarity-versus-confidence semantics, trajectory quality evidence, bounded sampling and input size, removal of the duplicate Rank API, canonical equal-timestamp ordering, exact prepared fingerprints, mathematical result validation, worst-endpoint scoring, exact relative difference, great-circle resampling, deliberately retained Go and floating-point contracts, permanent regression tests, engineering commit `6dbae4e6fe00295af0f7ba5303855736b76e8bde`, GitHub Actions run `30360637718`, and formal closure with zero open, unclassified, or deferred findings.


<!-- HISTORICAL-AGGREGATE-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 131 — Historical Aggregate Review Hardening

`131_HISTORICAL_AGGREGATE_REVIEW_HARDENING.md` records lowercase regional
persistence, the pre-existing full tuple cursor, complete JSON-versus-column
consistency, deterministic record identity, canonical payload idempotency,
writer-interface segregation, raw validation before storage canonicalization,
nil-context rejection, StoredAt causality, timestamp-mirror database guards,
migration 029, deliberately rejected mechanical findings, permanent regression
tests, engineering commit `18dde73b2d122d00476ea21accb256b33fc23527`,
GitHub Actions run `30374964285`, and formal closure with zero open,
unclassified, or deferred findings.


<!-- HISTORICAL-MATERIALIZATION-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 132 — Historical Materialization Review Hardening

`132_HISTORICAL_MATERIALIZATION_REVIEW_HARDENING.md` records independent
previous/current dataset limits, one-transaction two-period reads, exact snapshot
query and version validation, period-specific summaries, isolated builder
inputs, Historical Comparison provenance ownership, canonical persisted Outcome,
generated-time fingerprint identity, nil-context rejection, typed stage errors,
permanent regression tests, engineering commit
`2bbbd2439580536ffe17f8827c654c245d9b6b1e`, GitHub Actions run
`30384357559`, and formal closure with zero open, unclassified, or deferred
findings.


<!-- HISTORICAL-REPLAY-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 133 — Historical Replay Review Hardening

`133_HISTORICAL_REPLAY_REVIEW_HARDENING.md` records strict Materialization outcome
validation, self-contained complete, partial, and failed replay results,
structured failure evidence, deterministic replay fingerprinting, early global
request validation, bounded planning limits, cross-call shared-period continuity,
production completed-prefix reporting with non-zero failure exit, deliberately
rejected replay-wide transaction and checkpoint coupling, permanent regression
tests, engineering commit
`38b14fbb8649a2e7e875cd4ae7ed73b6a954a068`, GitHub Actions run
`30390451707`, and formal closure with zero open, unclassified, or deferred
findings.

<!-- PROJECTION-CONTRACT-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 134 — Projection Contract Review Hardening

`134_PROJECTION_CONTRACT_REVIEW_HARDENING.md` records exact horizon-grid
validation, explicit limited-status evidence, bounded confidence reasons,
weakest-evidence reconciliation, shared ordinal confidence vocabulary, strict
SHA-256 and ICAO formats, provenance chronology and uniqueness, typed result
validation, deliberately retained optional pointers and producer-owned physical
policies, permanent regression tests, engineering commit
`964556d0ca8a1ce9aa74c37c55961cdd006b3de8`, GitHub Actions run
`30396070318`, and formal closure with zero open, unclassified, or deferred
findings.
<!-- PROJECTION-HORIZON-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 135 — Projection Horizon Review Hardening

`135_PROJECTION_HORIZON_REVIEW_HARDENING.md` records exact fixed-step horizon
semantics, canonical plan validation and finalization, complete SHA-256 horizon
fingerprinting, reachable default HTTP duration, bounded point allocation,
normalized policy identity, typed nil-policy lifecycle errors, consumer validation
of alternative planners, deliberately retained idiomatic constructor and public-plan
contracts, permanent regression tests, engineering commits
`7249aa7625dd306bbd769dade6ce3262edca01ab` and
`d2bc87b07ea0eb6a0b9b25f0a1e3cb2cbc52cd1b`, GitHub Actions run `30402249129`,
and formal closure with zero open, unclassified, or deferred findings.

## Document 136 — Projection Baseline Review Hardening

`136_PROJECTION_BASELINE_REVIEW_HARDENING.md` records cutoff-safe quality recomputation,
completed segment and coverage-gap isolation, PostgreSQL cutoff alignment,
unavailable-result provenance, observation-age confidence, conservative physical
bounds, explicit altitude and eligibility policy evidence, deterministic rejection of
conflicting latest observations, explicit horizontal fallback, stationary limited
on-ground behavior, qualified review findings, engineering commits
`0f2c1b2c6f91f104b8e0880e85dc8144fed6a910`,
`af9c377193c21c048721e9cc28bf885d6ad276ec`, and
`560e4ed15cabbf0042110e00363a3a7c4d0c0d2e`, permanent audit commit
`51476c427f77b5a7375cd30b6f9a81d446c1c3f2`, GitHub Actions run `30408617024`, and
formal closure with zero open, unclassified, or deferred findings.

<!-- PROJECTION-NEIGHBORS-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 137 — Projection Neighbors Review Hardening

`137_PROJECTION_NEIGHBORS_REVIEW_HARDENING.md` records candidate eligibility
before the expensive evaluation budget, whole-input duplicate integrity, canonical
point and fingerprint ordering, systemic similarity failure propagation,
continuous continuation-gap enforcement, source-attested origin-destination route
scope, focused selector pipeline stages, explicit candidate-evaluation truncation
and qualified-selection limiting, permanent regression tests, engineering commits
`e2a4a7dc76e43942ca9deb0d8d5f83a09a42deff`,
`911a1b102c68af2746a13bfca48b008cf7225ff8`,
`3eee05fb44484aa6e389af66520aba23d4ae277e`, and
`353d19bc97f561e1897ece1967e7304c0e10b5fb`, permanent audit commit
`c409cc171507050625524af1a0b8b8a6f38b7a75`, GitHub Actions run `30452465009`, and formal closure with zero open,
unclassified, or deferred findings.

<!-- PROJECTION-PATTERN-CONFIDENCE-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 138 — Projection Pattern Confidence Review Hardening

`138_PROJECTION_PATTERN_CONFIDENCE_REVIEW_HARDENING.md` records semantic selected-neighbor and continuation fingerprinting,
strictly positive component
weights, similarity floor and dispersion evidence, freshness separation, five
canonical versioned components, pairwise continuation spread and divergence,
mandatory continuation-aware production interfaces, complete result reconstruction,
deliberately retained floating-point and compatibility contracts, engineering
commits `6e6ac17cfcfca688d57829adfe2468346db6db1a`,
`f73534feb275c5e109fa12fcfd9df5b69c56c03a`,
`5873ae911b40197ee45eea30e7558aa04af78064`, and
`e31fcb5bbbb76093305e8b2c137c793a85dc6795`, permanent audit commit
`cd8f114bfef698c51cfc6008ecd2ed01f9c1cc42`, GitHub Actions run
`30497703314`, and formal closure with zero open, unclassified, or deferred
findings.

<!-- PROJECTION-FRESHNESS-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 139 — Projection Freshness Review Hardening

`139_PROJECTION_FRESHNESS_REVIEW_HARDENING.md` records exact selected-neighbor
lineage, timestamp-derived selected-neighbor age evidence, overflow-safe mean-duration
calculation, ordered positive thresholds, strictly positive component weights,
semantic upstream-state fingerprinting, complete hard-violation reporting,
policy and upstream-state snapshots, component and decision reconstruction, evaluator-generated
production fixtures, deliberately retained floating-point and compatibility contracts,
engineering commits `0b47aa3231c93d573a6026651a4085d376a40583` and
`072d0eb349fcd0e42c1d3c0bcf54c51cefb08a19`, permanent audit commit
`619e24878a5025decf6fe21abddba537ce195560`, GitHub Actions run `30523502590`,
corrective weight-policy commit `e3e99758d6f654db12ccce32ec55ad1339fb518f`, GitHub Actions run
`30527541240`, and formal reclosure with zero open, unclassified, or deferred
findings.
<!-- PROJECTION-ROUTE-FREQUENCY-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 140 — Projection Route Frequency Review Hardening

`140_PROJECTION_ROUTE_FREQUENCY_REVIEW_HARDENING.md` records current-flight evidence isolation,
logical-flight deduplication, fixed full and recent exposure windows, distinct-flight
scoring, coherent policy targets, strictly positive thresholds and weights, complete
hard-violation reporting, weighted-score reconstruction, deterministic limitations,
semantic fingerprinting, production fixture correction, permanent regression tests,
and Backend Quality audit integration. Exact policy Continuous Integration run
`30544636679` and permanent-audit run `30548438062` are recorded, with formal closure
and zero open, unclassified, or deferred findings.

<!-- PROJECTION-CONTINUATION-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 141 — Projection Continuation Review Hardening

`141_PROJECTION_CONTINUATION_REVIEW_HARDENING.md` is the formally closed review
record for the critical authorization/projection evidence-identity defect,
compile-time Approved Evidence production contract, exact Pattern-to-Selection
fingerprint lineage, candidate-anchor binding, observed historical source provenance,
bounded interpolation and rate plausibility, conservative additive uncertainty,
effective weighted support, disagreement-aware confidence, zero-confidence status
semantics, near-antipodal rejection, fallback cause preservation, permanent strict
review audit, exact engineering-closure commit and Continuous Integration evidence,
and zero open, unclassified or deferred findings.

<!-- PROJECTION-ARRIVAL-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 142 — Projection Arrival Review Hardening

`142_PROJECTION_ARRIVAL_REVIEW_HARDENING.md` is the formally closed review record
for signed destination closing speed, bounded physical ground speed, preservation of
slow and receding samples, radial radius-entry uncertainty, complete Estimated Arrival
interval bounds, explicit nanosecond ceiling and overflow rejection, strict
current-trajectory identity, canonical used-sample fingerprint lineage, observed
current-endpoint provenance, confidence contribution reconstruction, duration-policy
coherence, focused regression tests, permanent strict review audit, exact engineering
closure commit and Continuous Integration evidence, and zero open, unclassified or
deferred findings.

<!-- PROJECTION-EVALUATION-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 143 — Projection Evaluation Review Hardening

`143_PROJECTION_EVALUATION_REVIEW_HARDENING.md` is the formally closed review record
for strict event-time and system-availability replay cutoffs, deterministic
equal-timestamp truth normalization, canonical projection and truth snapshot
fingerprints, altitude status lineage, physical interpolation limits, immutable
evaluation policy provenance, endpoint and lead-time metrics, bounded confidence
comparison, complete aggregation identity, unavailable-result accuracy isolation,
arrival prediction recall and airport accuracy, point micro and trajectory macro
averages, conventional median semantics, derived-metric recomputation,
GeneratedAt-independent aggregate input fingerprints, strict actual-arrival ICAO
validation, focused regression tests, permanent strict review audit, exact engineering
Continuous Integration evidence, and zero open, unclassified or deferred findings.

<!-- PROJECTION-PRODUCTION-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 144 — Projection Production Review Hardening

`144_PROJECTION_PRODUCTION_REVIEW_HARDENING.md` is the formally reclosed review
record for the completed single-plan, immutable-snapshot, route-binding,
evidence-graph, projector-postcondition, Estimated Arrival delta, fallback,
error-chain, and fingerprint hardening, together with the later Historical Projector
output-lineage correction. The final record includes the sealed production Historical
Projection adapter, typed approved-lineage receipt, independent continuation
fingerprint and canonical provenance reconstruction, exact selected-neighbor
provenance binding, malicious lineage-drift tests, permanent strict audit, exact
corrective engineering commit and Continuous Integration evidence, and zero open,
unclassified, or deferred findings.

<!-- PROJECTION-READ-REVIEW-HARDENING:DOCUMENT-INDEX -->

## Document 145 — Projection Read Review Hardening

`145_PROJECTION_READ_REVIEW_HARDENING.md` records the atomic repeatable-read snapshot,
strict context and as-of boundaries, snapshot and Composer postconditions, route-row payload metadata binding,
canonical default-duration handling, historical candidate backfill, record-level route-history lineage,
exact engineering commits and Continuous Integration evidence, permanent audit commit `e0557f6bc3115767ba124a9c94cbb008194c643b`,
GitHub Actions run `30651385019`, and formal closure with zero open, unclassified, or deferred findings.

<!-- PROJECTION-INTELLIGENCE-FINAL-RECONCILIATION:DOCUMENT-INDEX -->

## Document 146 — Projection Intelligence Final Cross-Module Reconciliation

`146_PROJECTION_INTELLIGENCE_FINAL_RECONCILIATION.md` records the aggregate reconciliation of all
twelve formally closed Projection Intelligence module reviews, the classification of the supplied
external static review declared at `a1689dc`, the retained evidence-based correctness requirements,
implementation commit `fb7fecd759a26c8d65d979ab8f541284ed82ed36`, GitHub Actions run `30658968264`,
the permanent cross-module audit, and formal closure with zero open confirmed cross-module findings.

<!-- BACKEND-CONTEXT-OWNERSHIP-AUDIT-CLOSURE:DOCUMENT-INDEX -->

## Document 147 — Backend Context Ownership Audit Closure

`147_BACKEND_CONTEXT_OWNERSHIP_AUDIT_CLOSURE.md` records the repository-wide
caller-context ownership review, the original twenty-four concrete replacements,
the correction of all runtime and verification-side findings, zero retained or
deferred replacements, abstract-syntax-tree enforcement for `context.Background()`
and `context.TODO()` parameter replacement, permanent Backend Quality integration,
focused regression evidence, exact grouped engineering commits and GitHub Actions
runs, and formal source closure with zero open, unclassified, or deferred findings.

<!-- FRONTEND-API-CLIENT-ABORT-TESTING-HARDENING:DOCUMENT-INDEX -->

## Document 148 — Frontend API Client Abort and Testing Hardening

`148_FRONTEND_API_CLIENT_ABORT_AND_TESTING_HARDENING.md` records the correction
of caller-cancellation versus request-timeout classification, the first dependency-free
frontend contract test harness, six API-client regression tests, Frontend Continuous
Integration enforcement, exact baseline and scope boundaries, and the requirement for
exact-commit Continuous Integration evidence before formal closure.

<!-- FRONTEND-AIRCRAFT-EXPLORER:DOCUMENT-INDEX -->

## Document 149 — Frontend Aircraft Explorer

`149_FRONTEND_AIRCRAFT_EXPLORER.md` records the searchable and sortable regional
aircraft index, shared map and detail-panel selection, bounded result rendering,
deterministic filtering and sorting semantics, five dependency-free model tests,
exact baseline and scope boundaries, and the requirement for exact-commit
Continuous Integration evidence before formal closure.

<!-- FRONTEND-TRAFFIC-WORKSPACE:DOCUMENT-INDEX -->

## Document 150 — Frontend Traffic Workspace

`150_FRONTEND_TRAFFIC_WORKSPACE.md` records the accessible Aircraft and Intelligence
workspace tabs, permanently visible map context, shared map and explorer selection,
deterministic ICAO24 normalization and panel transitions, four dependency-free state
model tests, exact baseline and scope boundaries, and the requirement for exact-commit
Continuous Integration evidence before formal closure.

<!-- FRONTEND-APPLICATION-SHELL:DOCUMENT-INDEX -->

## Document 151 — Frontend Application Shell

`151_FRONTEND_APPLICATION_SHELL.md` records the product-level application header,
hero, startup snapshot status semantics, stable page navigation, research-scope
communication, deterministic global styling, five status-model regression tests,
exact baseline and the requirement for exact-commit Continuous Integration evidence
before formal closure.

<!-- BACKEND-STARTUP-CONTEXT-HARDENING:DOCUMENT-INDEX -->

## Document 152 — Backend Startup Context Hardening

`152_BACKEND_STARTUP_CONTEXT_HARDENING.md` records caller-owned PostgreSQL startup
cancellation, the compatibility-preserving context-aware pool constructor, server
lifecycle propagation, focused cancellation and classification regression tests,
permanent context-ownership enforcement, exact baseline and the requirement for
exact-commit Continuous Integration evidence before formal closure.


<!-- RECRUITER-QUICKSTART:DOCUMENT-INDEX -->

## Document 153 — Recruiter Quickstart

`153_RECRUITER_QUICKSTART.md` records the reproducible reviewer startup path, the
local-only mutation authorization startup default, exact health, readiness and version
checks, frozen pnpm frontend setup, permanent quickstart verification, Backend
Continuous Integration reachability, exact baseline and the requirement for
exact-commit Continuous Integration evidence before formal closure.

<!-- FRONTEND-REGIONAL-TRAFFIC-BRIEF:DOCUMENT-INDEX -->

## Document 154 — Frontend Regional Traffic Brief

`154_FRONTEND_REGIONAL_TRAFFIC_BRIEF.md` records the deterministic current-snapshot
composition brief, airborne altitude bands, leading attributed airlines and origin
countries, explicit evidence boundaries, five dependency-free model tests, exact
baseline and the requirement for exact-commit Continuous Integration evidence before
formal closure.

<!-- FRONTEND-SHAREABLE-WORKSPACE-STATE:DOCUMENT-INDEX -->

## Document 155 — Frontend Shareable Workspace State

`155_FRONTEND_SHAREABLE_WORKSPACE_STATE.md` records URL-addressable region, aircraft and
workspace-panel state, browser Back and Forward restoration, canonical query serialization,
the explicit copy-link action, six dependency-free model tests, exact baseline and the
requirement for exact-commit Continuous Integration evidence before formal closure.

<!-- FRONTEND-LIVE-TRAFFIC-CONTROL:DOCUMENT-INDEX -->

## Document 156 — Frontend Live Traffic Control

`156_FRONTEND_LIVE_TRAFFIC_CONTROL.md` records explicit current, aging and stale
snapshot semantics, bounded automatic refresh choices, pause and resume controls,
countdown and retained-snapshot failure presentation, six dependency-free model tests,
exact baseline and the requirement for exact-commit Continuous Integration evidence
before formal closure.

<!-- FRONTEND-RESEARCH-SNAPSHOT-EXPORT:DOCUMENT-INDEX -->

## Document 157 — Frontend Research Snapshot Export

`157_FRONTEND_RESEARCH_SNAPSHOT_EXPORT.md` records deterministic CSV and GeoJSON
exports for the current regional traffic snapshot, fixed schemas, provenance metadata,
invalid-coordinate exclusion accounting, seven dependency-free model tests, exact
baseline and the requirement for exact-commit Continuous Integration evidence before
formal closure.

<!-- FRONTEND-TRAFFIC-DATA-QUALITY-LENS:DOCUMENT-INDEX -->

## Document 158 — Frontend Traffic Data Quality Lens

`158_FRONTEND_TRAFFIC_DATA_QUALITY_LENS.md` records browser-side structural checks for
identifiers, coordinates, motion, observation recency, airborne altitude and descriptive
attribution, deterministic issue ordering, seven dependency-free model tests, exact
baseline and the requirement for exact-commit Continuous Integration evidence before
formal closure.

<!-- FRONTEND-UNIFIED-AIRPORT-ANALYTICS-WORKSPACE:DOCUMENT-INDEX -->

## Document 159 — Frontend Unified Airport Analytics Workspace

`159_FRONTEND_UNIFIED_AIRPORT_ANALYTICS_WORKSPACE.md` records the production Airport
Intelligence ranking, search and deterministic sorting, digital airport passport,
completed-day history, trend and continuity evidence, merged limitations, runtime API
validation, eight dependency-free model tests, exact baseline and the requirement for
exact-commit Continuous Integration evidence before formal closure.

<!-- FRONTEND-HISTORICAL-ANALYTICS-COMPARISON:DOCUMENT-INDEX -->

## Document 160 — Frontend Historical Analytics and Comparison

`160_FRONTEND_HISTORICAL_ANALYTICS_COMPARISON.md` records the typed production
Historical Intelligence aggregate client, global, airport and route scope controls,
server-catalog metric filtering, bucket evidence visualization, previous-period and
persisted-record comparisons, eight dependency-free model tests, exact baseline and
the requirement for exact-commit Continuous Integration evidence before formal closure.

<!-- FRONTEND-PRODUCT-HARDENING:DOCUMENT-INDEX -->

## Document 161 — Frontend Product Hardening

`161_FRONTEND_PRODUCT_HARDENING.md` records keyboard skip links, mobile navigation,
forced-colors and coarse-pointer contracts, route and global error recovery, loading
and not-found surfaces, runtime connectivity announcements, bounded React Query retry
policy, dynamic research workspace delivery, fifteen dependency-free tests, exact
baseline and final Frontend Continuous Integration evidence for release SHA
`49e474e929dcca5b687464f0a47ce73fcd5a52a7`.

<!-- RELEASE-PORTFOLIO-CLOSURE:DOCUMENT-INDEX -->

## Document 162 — Release and Portfolio Closure

`162_RELEASE_AND_PORTFOLIO_CLOSURE.md` records the source-release definition,
independent source, exact-commit Continuous Integration and public-deployment states,
the exact release SHA, Backend and Frontend Continuous Integration run identifiers,
evidence policy, verified Neon, Render and Vercel deployment, exact-origin
Cross-Origin Resource Sharing, full production smoke and final portfolio MVP scope boundary.

## Document 163 — Production Deployment Runbook

`163_PRODUCTION_DEPLOYMENT_RUNBOOK.md` records the direct migration and pooled runtime
Neon connection policy, free-plan Render Blueprint, explicit direct-database migration,
verified public API and Next.js deployment, exact-origin Cross-Origin Resource Sharing,
full production smoke and rollback procedure.

## Document 164 — Recruiter Demo Script

`164_RECRUITER_DEMO_SCRIPT.md` provides a bounded seven-minute product and code walkthrough,
engineering decision prompts, likely reviewer questions and evidence-safe demo discipline.

## Document 165 — System Architecture and Decisions

`165_SYSTEM_ARCHITECTURE_AND_DECISIONS.md` records the modular-monolith topology, provider,
canonical data, analytical, HTTP and frontend boundaries, PostgreSQL and backend-owned
analytics decisions, reliability policies, evidence boundary and deliberately excluded
complexity.

<!-- BACKEND-OPERATIONS-EVIDENCE-CLOSURE:DOCUMENT-INDEX -->

## Document 166 — Backend Operations and Continuous Integration Evidence Closure

`166_BACKEND_OPERATIONS_AND_CI_EVIDENCE_CLOSURE.md` records exact product-release
Continuous Integration evidence, the free-plan Render Docker Blueprint, direct Neon
migration workflow, verified public API and Next.js deployment, exact-origin
Cross-Origin Resource Sharing, full production smoke, free-tier operational boundary and
the final evidence-only attestation rule.

---

### Document 167 — Backend Timeout Consistency Audit Closure

Path:

```text
docs/167_BACKEND_TIMEOUT_CONSISTENCY_AUDIT_CLOSURE.md
```

Purpose:

```text
Records the repository-wide timeout ownership audit, application request deadline
propagation, bounded provider HTTP execution, permanent timeout consistency audit gate,
regression evidence, configuration contract, and deliberate separation
between interactive HTTP budgets and administrative PostgreSQL operations.
```

---

### Document 168 — Backend Observability and Service-Level Objectives Closure

Path:

```text
docs/168_BACKEND_OBSERVABILITY_AND_SLO_CLOSURE.md
```

Purpose:

```text
Records the protected Prometheus-compatible observability foundation, bounded metric
labels, API and ingestion instrumentation, provider and fallback evidence, PostgreSQL and
reconciliation state collectors, service-level objectives, alert thresholds, permanent
observability audit gate, and deployment boundary.
```

## Document 169 — Release Truth and Deployment Revision Closure

`169_RELEASE_TRUTH_AND_DEPLOYMENT_REVISION_CLOSURE.md` separates historical deployment evidence, intended Render revision, observed API revision and current repository `HEAD`; removes the local-HEAD-equals-deployment assumption and installs permanent release truth gates.


## Document 170 — Dependency Maintenance Closure

- File: `170_DEPENDENCY_MAINTENANCE_CLOSURE.md`
- Status: `CLOSED`
- Purpose: closes the safe dependency update wave, groups related Dependabot updates, removes invalid label references, and defers TypeScript major migration behind an explicit compatibility stage.


## Document 171 — Dependabot Follow-up Reconciliation

- File: `171_DEPENDABOT_FOLLOW_UP_RECONCILIATION.md`
- Status: `CLOSED`
- Purpose: reconciles the regenerated Dependabot follow-up wave, applies safe patch updates, groups the Next.js toolchain, and defers ESLint 10, Node type definitions 26, MapLibre 6 and TypeScript 7 behind explicit migration stages.

## Document 172 — Repository Governance and Security Automation

- File: `172_REPOSITORY_GOVERNANCE_AND_SECURITY_AUTOMATION.md`
- Status: `PATCH PREPARED; SETTINGS PENDING EXACT-COMMIT CI`
- Purpose: establishes immutable Action pins, stable required CI gates, CodeQL, ownership and security policy files, reproducible GitHub settings automation, protected main-branch governance, and history-preserving stale-branch reconciliation.

<!-- CORE-FLIGHT-DATA-INGESTION-PRODUCTION-CLOSURE-V1:DOCUMENT-INDEX -->
## Document 173 — Core Flight Data Ingestion Production Closure

`173_CORE_FLIGHT_DATA_INGESTION_PRODUCTION_CLOSURE.md` records the stale production evidence, missing runtime root cause, bounded one-shot command, serialized free GitHub Actions schedule, secret boundary, end-to-end freshness gate, platform limitations and final runtime activation criteria.

<!-- OPENAPI-CONTRACT-FOUNDATION-V1:DOCUMENT-INDEX -->

## Document 175 — OpenAPI Contract Foundation

`175_OPENAPI_CONTRACT_FOUNDATION.md`

Defines the source-backed OpenAPI 3.1 contract for eight stable public GET operations,
typed success and error envelopes, route and DTO drift protection, dedicated Continuous
Integration verification, release-gate wiring, and the contract boundary required by the
next Playwright end-to-end increment.

<!-- PLAYWRIGHT-E2E-FOUNDATION-V1:DOCUMENT-INDEX -->

## Document 176 — Playwright End-to-End Testing Foundation

`176_PLAYWRIGHT_E2E_FOUNDATION.md`

Defines the isolated Playwright Chromium runtime, OpenAPI-aligned deterministic mock API,
semantic browser assertions, server-rendered snapshot, shareable region state, responsive
navigation, traffic failure recovery, evidence retention, dedicated Continuous Integration
workflow, and prohibition on targeting public deployments.

<!-- OPENAPI-CONTRACT-CLOSURE-INVENTORY-V1:DOCUMENT-INDEX -->

## Document 177 — OpenAPI Contract Closure Route Inventory

`177_OPENAPI_CONTRACT_CLOSURE_INVENTORY.md`

Records the source-backed inventory of 38 public operations and one internal metrics operation, the current 20-operation OpenAPI gap after the core read expansion, nested Fiber group and constant-backed path resolution, mutation and metrics authorization boundaries, permanent tests, and the dedicated Continuous Integration gate required before complete OpenAPI expansion.

<!-- OPENAPI-CORE-READ-SURFACE-V1:DOCUMENT-INDEX -->

## Document 178 — OpenAPI Core Read Surface

`178_OPENAPI_CORE_READ_SURFACE.md`

Closes the ten-operation core read slice for aircraft, flights, flight states, trajectories, route context, and the active-aircraft metric; expands the public OpenAPI contract from 8 to 18 operations; preserves nullable telemetry and typed error semantics; and keeps the Playwright mock surface exactly aligned.

<!-- OPENAPI-ADVANCED-INTELLIGENCE-READ-SURFACE-V1:DOCUMENT-INDEX -->

## Document 179 — OpenAPI Advanced Intelligence Read Surface

`179_OPENAPI_ADVANCED_INTELLIGENCE_READ_SURFACE.md`

Closes seventeen source-backed analytical GET operations for transponder evidence, current weather, analytical metrics, Airport Intelligence, Historical Intelligence, Projection Intelligence, Stability Intelligence, Weather Context, and Airspace Intelligence; expands OpenAPI coverage from 18 to 35 operations; and leaves only the three-operation Route Intelligence slice for final security-aware closure.

<!-- OPENAPI-ROUTE-INTELLIGENCE-CONTRACT-CLOSURE-V1:DOCUMENT-INDEX -->

## Document 180 — OpenAPI Route Intelligence Contract Closure

`180_OPENAPI_ROUTE_INTELLIGENCE_CONTRACT_CLOSURE.md`

Closes the final protected Route Intelligence POST and two materialized GET reads; defines the X-Internal-API-Key security scheme and authorization failures; expands OpenAPI coverage from 35 to all 38 production public operations; and permanently reduces missing and extra route inventory counts to zero.

<!-- PRODUCTION-OBSERVABILITY-CLOSURE-V1:DOCUMENT-INDEX -->
## Document 170 — Production Observability and Alerting Closure

Path:

```text
docs/170_PRODUCTION_OBSERVABILITY_AND_ALERTING_CLOSURE.md
```

Purpose:

```text
Records the protected production metrics path, Grafana Cloud stack namespace,
idempotent SLO dashboard and nine-rule provisioning, exact notification-policy
receiver, controlled email-delivery evidence, security boundaries, operational
limitations, and the formal production observability closure statement.
```

<!-- OPENAPI-DEVELOPER-EXPERIENCE:DOCUMENT-INDEX -->

## Document 181 — OpenAPI Developer Experience

`181_OPENAPI_DEVELOPER_EXPERIENCE.md` records the embedded same-binary API documentation surface, exact canonical specification endpoint, dependency-free browser explorer, protected-mutation browser boundary, deterministic TypeScript client generation, byte-level embedded specification drift enforcement, generated-client drift enforcement, TypeScript validation, OpenAPI workflow integration, release-gate integration, and the exact baseline for the developer-experience increment.

<!-- ZERO-COST-PRODUCTION-INGESTION-RELIABILITY:DOCUMENT-INDEX -->

## Document 182 — Zero-Cost Production Ingestion Reliability Closure

`182_ZERO_COST_PRODUCTION_INGESTION_RELIABILITY.md` records the closed
repository-owned Cloudflare primary scheduler and freshness watchdog, GitHub Actions
run-state deduplication and dispatch provenance, encrypted-token boundary, offset hourly
fallback, controlled live evidence, exact-revision runtime validation, operational
ownership, and the final production ingestion reliability closure statement.

<!-- CLOUDFLARE-INGESTION-LIVE-DEPLOYMENT-EVIDENCE:DOCUMENT-INDEX -->

## Document 183 — Cloudflare Ingestion Live Deployment and Closure Evidence

`183_CLOUDFLARE_INGESTION_LIVE_DEPLOYMENT_EVIDENCE.md` records the exact
Cloudflare Worker deployment identity, stable health route, encrypted GitHub
authorization boundary, watchdog recovery, real primary dispatch, recent-success
and active-run deduplication, bounded hourly fallback simulation, post-ingestion
freshness, exact-revision runtime validation, secret-safe logging, and the final
production ingestion reliability closure.

<!-- STAGE-13-FRONTEND-ANALYTICS-CLOSURE:DOCUMENT-INDEX -->

## Document 184 — Stage 13 Frontend Analytics Integration Completion

`184_STAGE_13_FRONTEND_ANALYTICS_INTEGRATION_COMPLETION.md`

Records formal completion of the Projection, Projection Map, Weather Context and
Stability and Explainability frontend slices, source-backed panel and query
wiring, observed-versus-estimated MapLibre source separation, research-only
scope boundaries, permanent Frontend CI and release-gate regression protection,
and the separate visual-redesign product phase.

<!-- CI-REQUIRED-CHECK-RECOVERY-HARDENING:DOCUMENT-INDEX -->

## Document 185 — CI Required-Check Recovery Hardening

`185_CI_REQUIRED_CHECK_RECOVERY_HARDENING.md` defines the fail-closed recovery
procedure for missing, cancelled or damaged required GitHub checks; exact
pull-request head-SHA binding; reuse of the existing full Backend CI
`workflow_dispatch`; read-only diagnosis; `cancel-in-progress` handling; the
prohibition on empty retrigger commits and reduced recovery workflows; and the
permanent CI/release regression contract.

<!-- PLAYWRIGHT-PRODUCT-COVERAGE-V1:DOCUMENT-INDEX -->

## Document 186 — Playwright Product Coverage Expansion

`186_PLAYWRIGHT_PRODUCT_COVERAGE.md` records the expansion from the original
four-scenario browser foundation to twenty deterministic Chromium product journeys
covering aircraft, airport, historical, Projection/Weather/Stability evidence,
CSV and GeoJSON exports, bounded failure/recovery paths, accessibility semantics,
desktop/mobile layout regression invariants, and screenshot evidence retained in
Playwright CI artifacts. Pixel-golden baselines remain intentionally coupled to
the later visual-redesign closure.

<!-- PRODUCTION-INGESTION-RESILIENCE-INCIDENT:CLOSURE:DOCUMENT-INDEX -->

## Document 191 — Production Ingestion Resilience Incident Closure

`191_PRODUCTION_INGESTION_RESILIENCE_INCIDENT_CLOSURE.md`

Records the August 2026 production ingestion incident caused by provider-level
`403 Unauthorized` responses combined with an unbounded external dispatch loop,
the fail-closed GitHub and Cloudflare containment, recent-failure circuit breaker,
multi-provider unauthorized fallback semantics, deployment evidence, remaining
provider-approval boundary, and controlled production-reactivation criteria.
