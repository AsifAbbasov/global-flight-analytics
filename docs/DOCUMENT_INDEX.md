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
