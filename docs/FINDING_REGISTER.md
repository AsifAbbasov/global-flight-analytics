# Finding Register — Global Flight Analytics

Status: Canonical Finding Registry v1.11

## Purpose

This document is the single index of engineering findings, remediation stages, and final finding-level status.

The registry prevents documentation drift between README files, stage documents, commits, tests, CI evidence, and implementation state.

## Mandatory finding record

Every non-trivial engineering finding must contain:

1. Finding / symptom
2. Root cause
3. Failure scenario
4. Impact
5. Severity rationale
6. Existing guarantees violated
7. Considered solutions
8. Chosen remediation
9. Why this solution was selected
10. Rejected alternatives
11. Trade-offs
12. Regression tests / protection
13. Adversarial review findings
14. Remediation iterations
15. Residual risks and limitations
16. Operational or deployment consequences
17. Exact evidence
18. Final canonical status
19. Prevention / future guard

`docs/DOCUMENTATION_POLICY.md` is the normative description of these fields.

## Status values

- OPEN
- IN_PROGRESS
- CLOSED
- ACCEPTED_RISK

## Rule

A finding is not considered closed only because code changed. Closure requires implementation evidence, regression protection, documentation alignment, and a canonical registry status.

Detailed technical history belongs in stage/finding documents. This registry is the navigation and finding-status authority.

## Evidence honesty rule

Retroactive records must distinguish historical evidence from later reconstruction.

If an original severity, pull-request number, review comment, reviewer identity, or exact historical CI run cannot be recovered from repository evidence, the registry or canonical stage document must say so explicitly. A later reconstruction may classify impact or alternatives, but it must not fabricate historical review events.

## Registered findings

| ID | Finding | Severity | Status | Canonical document | Implementation evidence |
|---|---|---|---|---|---|
| GFA-DB-001 | PostgreSQL migration atomicity | P1 retrospective | CLOSED | `58_STAGE_14_17_POSTGRES_MIGRATION_ATOMICITY.md` | `07c0907eb4b739ca2ba12259600df537254a1075` |
| GFA-DB-002 | Unsafe migration baseline | P1 retrospective | CLOSED | `59_STAGE_14_18_POSTGRES_BASELINE_REMOVAL.md` | `3dafcf8ad08a8a4b270456cc2a023e8f4d0ffd53` |
| GFA-DB-003 | Data Quality parent integrity | P1 retrospective | CLOSED | `60_STAGE_14_19_DATA_QUALITY_PARENT_INTEGRITY.md` | `0d3d1d37a65423ca6263df0816360eabf3c66235` |
| GFA-DB-004 | FlightTrajectory read snapshot consistency | P1 retrospective | CLOSED | `61_STAGE_14_20_TRAJECTORY_READ_SNAPSHOT_CONSISTENCY.md` | `fcc601db509d8fb71d2f2db273548fec3832d3bd` |
| GFA-DB-005 | Ingestion Run terminal integrity | P1 retrospective | CLOSED | `62_STAGE_14_21_INGESTION_RUN_TERMINAL_INTEGRITY.md` | `b3603311d86f23c66bc945c8a61471142ccbec63` |
| GFA-DB-006 | FlightTrajectory relational integrity | P1 retrospective | CLOSED | `63_STAGE_14_22_TRAJECTORY_RELATIONAL_INTEGRITY.md` | `5bb4a6aab7b16bc13e8477ca31f11eaa27e808ff` |
| GFA-DB-007 | Canonical migration filename contract | P2 retrospective | CLOSED | `64_STAGE_14_23_CANONICAL_MIGRATION_FILENAME_CONTRACT.md` | `4c41b8588e9119f59c090c976cef494c55683e18` |
| GFA-DB-008 | Explicit altitude integer persistence policy | P2 retrospective | CLOSED | `65_STAGE_14_24_EXPLICIT_ALTITUDE_INTEGER_POLICY.md` | `467d6bf5f6e66febbb83944664735ce26e7054c3` |
| GFA-DATA-009 | Traffic altitude status semantics | P1 retrospective | CLOSED | `66_STAGE_14_25_TRAFFIC_ALTITUDE_STATUS_SEMANTICS.md` | `18545100dda0e9852927dc4a93dafcc25394b1e1` |
| GFA-DATA-010 | Airport elevation semantics | P2 retrospective | CLOSED | `67_STAGE_14_26_AIRPORT_ELEVATION_SEMANTICS.md` | `75247fd242b293de65fa85f164fd594c64343b9a` |
| GFA-DB-011 | Flight Feature timestamp mirror consistency | P1 retrospective | CLOSED | `68_STAGE_14_27_FLIGHT_FEATURE_TIMESTAMP_CONSISTENCY.md` | `d76c526601a35ad7964fd9e93513396b0b4e4d6b` |
| GFA-MAINT-012 | PostgreSQL Trajectory Repository cohesion | P3 retrospective | CLOSED | `69_STAGE_14_28_POSTGRES_TRAJECTORY_REPOSITORY_DECOMPOSITION.md` | `24aa9a41abf4b5048e207c72d6aa4f93ab86319a` |
| GFA-DB-013 | Production migration catalog duplicate-version integrity | P1 retrospective | CLOSED | `71_STAGE_14_29_MIGRATION_CATALOG_INTEGRITY.md` | `4ef16aaa53e5b749e841a4b3226516c65da1bd06` |
| GFA-DB-014 | Ingestion Run terminal evidence invariants | P1 retrospective | CLOSED | `72_STAGE_14_30_POSTGRES_CORRECTNESS_HARDENING.md` | `04137ad17690ca197f1aea74b434057f2157dd7d` |
| GFA-DB-015 | Route Result timestamp mirror consistency | P1 retrospective | CLOSED | `72_STAGE_14_30_POSTGRES_CORRECTNESS_HARDENING.md` | `04137ad17690ca197f1aea74b434057f2157dd7d` |
| GFA-DB-016 | Historical Aggregate timestamp mirror consistency | P1 retrospective | CLOSED | `72_STAGE_14_30_POSTGRES_CORRECTNESS_HARDENING.md` | `04137ad17690ca197f1aea74b434057f2157dd7d` |
| GFA-DB-017 | Cancelled-context rollback independence | P2 retrospective | CLOSED | `72_STAGE_14_30_POSTGRES_CORRECTNESS_HARDENING.md` | `04137ad17690ca197f1aea74b434057f2157dd7d` |
| GFA-MAINT-018 | Airport Import and Flight State write coordinator cohesion | P3 retrospective | CLOSED | `73_STAGE_14_31_POSTGRES_WRITE_REPOSITORY_DECOMPOSITION.md` | `520779faef05b88fdeba4d9d244feb09f569010c` |
| GFA-PERF-019 | Bounded deterministic Airport catalog pagination | P2 retrospective | CLOSED | `74_STAGE_14_32_AIRPORT_KEYSET_PAGINATION.md` | `06f87ba4adfa3202cfe4f68232712a97e6812630` |
| GFA-MAINT-020 | Canonical Airport row-scanning ownership | P3 retrospective | CLOSED | `74_STAGE_14_32_AIRPORT_KEYSET_PAGINATION.md` | `06f87ba4adfa3202cfe4f68232712a97e6812630` |
| GFA-DB-021 | Repository caller-context substitution | P2 retrospective | CLOSED | `75_STAGE_14_33_EXPLICIT_REPOSITORY_CONTEXT_AND_TRAJECTORY_WRITE_MODE.md` | `211c774bb4820b6607bdbb6bd4e9cf17f1bc697b` |
| GFA-DB-022 | Implicit Trajectory write-mode sentinel | P2 retrospective | CLOSED | `75_STAGE_14_33_EXPLICIT_REPOSITORY_CONTEXT_AND_TRAJECTORY_WRITE_MODE.md` | `211c774bb4820b6607bdbb6bd4e9cf17f1bc697b` |
| GFA-DB-023 | Hard-coded migration-repair sequence and duplicated checksum evidence | P2 retrospective | CLOSED | `76_STAGE_14_34_POSTGRESQL_CONTRACT_CONSOLIDATION.md` | `e850eeb0b29c9a83fb1e1f8ee2215fe80828e969` |
| GFA-DB-024 | Nullable repository argument typed-nil and validation ambiguity | P2 retrospective | CLOSED | `76_STAGE_14_34_POSTGRESQL_CONTRACT_CONSOLIDATION.md` | `e850eeb0b29c9a83fb1e1f8ee2215fe80828e969` |
| GFA-DATA-025 | Fabricated source provenance through `unknown` fallback | P1 retrospective | CLOSED | `76_STAGE_14_34_POSTGRESQL_CONTRACT_CONSOLIDATION.md` | `e850eeb0b29c9a83fb1e1f8ee2215fe80828e969` |
| GFA-PERF-026 | UUID-column text casts in membership queries | P2 retrospective | CLOSED | `76_STAGE_14_34_POSTGRESQL_CONTRACT_CONSOLIDATION.md` | `e850eeb0b29c9a83fb1e1f8ee2215fe80828e969` |
| GFA-MAINT-027 | Duplicated Trajectory read SQL and row-mapping ownership | P3 retrospective | CLOSED | `77_STAGE_14_35_TRAJECTORY_QUERY_CONSOLIDATION_AND_PROFILING.md` | `f414f6638f8ba5fbe61321e55a21ff3ac91a4986` |
| GFA-DB-028 | Trajectory read caller-context substitution | P2 retrospective | CLOSED | `77_STAGE_14_35_TRAJECTORY_QUERY_CONSOLIDATION_AND_PROFILING.md` | `f414f6638f8ba5fbe61321e55a21ff3ac91a4986` |
| GFA-PERF-029 | Trajectory query/index ordering mismatch and missing plan evidence | P2 retrospective | CLOSED | `77_STAGE_14_35_TRAJECTORY_QUERY_CONSOLIDATION_AND_PROFILING.md` | `f414f6638f8ba5fbe61321e55a21ff3ac91a4986` |
| GFA-DB-030 | Residual post-closure migrator caller-context substitution | P2 retrospective | CLOSED | `79_POST_CLOSURE_MIGRATOR_CONTEXT_HARDENING.md` | `1c4a7bb992056e6b2c1d1394424643f913d31b00`; guard `384f526474282a8ae63250fa36d8182eb342f772` |
| GFA-ARCH-031 | Duplicated cross-domain confidence vocabulary | P3 retrospective | CLOSED | `41_STAGE_14_1_ARCHITECTURE_CONSOLIDATION_FOUNDATION.md` | `fc6c3dbafa302d061653587163457d72f08c7a77` |
| GFA-CONTRACT-032 | Go/TypeScript/runtime Trajectory contract drift | P2 retrospective | CLOSED | `41_STAGE_14_1_ARCHITECTURE_CONSOLIDATION_FOUNDATION.md` | `fc6c3dbafa302d061653587163457d72f08c7a77` |
| GFA-REL-033 | Production reachability evidence gap | P2 retrospective | CLOSED | `41_STAGE_14_1_ARCHITECTURE_CONSOLIDATION_FOUNDATION.md` | `fc6c3dbafa302d061653587163457d72f08c7a77` |
| GFA-OPS-034 | Frontend package-manager execution instability | P3 retrospective | CLOSED | `41_STAGE_14_1_ARCHITECTURE_CONSOLIDATION_FOUNDATION.md` | Stage 14.1 follow-up; foundation `fc6c3dba...` |
| GFA-REL-035 | Unowned non-runtime / confirmed dead analytical package lifecycle | P2 retrospective | CLOSED | `42_STAGE_14_2_DEAD_CODE_CLASSIFICATION_AND_REMOVAL.md` | `8bcc73ad1281d468fc17dc9f0628d54f79d7e2b0` |
| GFA-REL-036 | Airport Intelligence implemented but production-unreachable | P2 retrospective | CLOSED | `43_STAGE_14_3_AIRPORT_INTELLIGENCE_PRODUCTION_INTEGRATION.md` | `bb9f3510fd9fead1a80edb688c1ab125b8fbdb1b` |
| GFA-REL-037 | Feature Pipeline implemented without operational materialization root | P2 retrospective | CLOSED | `44_STAGE_14_4_FEATURE_MATERIALIZATION_AND_PROFILER_REMOVAL.md` | `a1689dc71baa9b2c2b4d66febb30b86436b893c1` |
| GFA-MAINT-038 | Isolated `datasetprofiler` facade | P3 retrospective | CLOSED | `44_STAGE_14_4_FEATURE_MATERIALIZATION_AND_PROFILER_REMOVAL.md` | `a1689dc71baa9b2c2b4d66febb30b86436b893c1` |
| GFA-SEC-039 | Unprotected mutation/computation-triggering HTTP boundary | P1 retrospective | CLOSED | `45_STAGE_14_5_MUTATION_ENDPOINT_PROTECTION.md` | `50831ae06cb1a38c321ec8c7766bc1f28ddb5757`; preserved by Server container evidence `1fc925c91117eebbb7c90c4bd6b3889548d55cb4` |
| GFA-GOV-040 | Projection benchmark and calibration governance gap | P2 retrospective | CLOSED | `46_STAGE_14_6_FORMULA_BENCHMARK_AND_CALIBRATION_GATE.md` | `f817bad2d6d12fe1619bb5c3bba3238d94d4c620` |
| GFA-SEC-041 | Vulnerable production PostCSS resolution and permissive audit threshold | P2 retrospective | CLOSED | `47_STAGE_14_7_FRONTEND_DEPENDENCY_SECURITY_REMEDIATION.md` | `4c2e5f5d534721a0c6a0a168d5f196deb590e212` |
| GFA-MAINT-042 | Server composition-root responsibility concentration | P3 retrospective | CLOSED | `48_STAGE_14_8_SERVER_COMPOSITION_ROOT_DECOMPOSITION.md` | `e9e9e658958db3ddced2f74d06ab50d0b8034853` |
| GFA-MAINT-043 | Boolean mode argument obscuring Historical query intent | P3 retrospective | CLOSED | `49_STAGE_14_9_HTTP_QUERY_AND_CONTRACT_BOUNDARY_HARDENING.md` | `2842d09fc2eb0dcc746a28dd126611fba0f2d1a8` |
| GFA-ARCH-044 | Historical HTTP/DTO dependency on PostgreSQL implementation and pgx errors | P2 retrospective | CLOSED | `49_STAGE_14_9_HTTP_QUERY_AND_CONTRACT_BOUNDARY_HARDENING.md` | `2842d09fc2eb0dcc746a28dd126611fba0f2d1a8` |
| GFA-REL-045 | Transponder Evidence implemented but production-unreachable / claim boundary unresolved | P2 retrospective | CLOSED | `50_STAGE_14_10_TRANSPONDER_EVIDENCE_PRODUCTION_INTEGRATION.md` | `19b187d848e993d13a72b0c3c4f212db8c7577fb` |
| GFA-MAINT-046 | Historical Intelligence validation responsibility concentration | P3 retrospective | CLOSED | `51_STAGE_14_11_TARGETED_LARGE_MODULE_HARDENING.md` | `d1fc34b6f25b5d7e8c18ac287709241e42617000` |
| GFA-MAINT-047 | Route Intelligence validation responsibility concentration | P3 retrospective | CLOSED | `51_STAGE_14_11_TARGETED_LARGE_MODULE_HARDENING.md` | `d1fc34b6f25b5d7e8c18ac287709241e42617000` |
| GFA-MAINT-048 | Historical-neighbor Projection orchestration concentration | P3 retrospective | CLOSED | `51_STAGE_14_11_TARGETED_LARGE_MODULE_HARDENING.md` | `d1fc34b6f25b5d7e8c18ac287709241e42617000` |
| GFA-MAINT-049 | Estimated-arrival orchestration and computation concentration | P3 retrospective | CLOSED | `51_STAGE_14_11_TARGETED_LARGE_MODULE_HARDENING.md` | `d1fc34b6f25b5d7e8c18ac287709241e42617000` |
| GFA-DB-050 | Projection input reads lacked one PostgreSQL snapshot | P1 retrospective | CLOSED | `52_STAGE_14_12_PROJECTION_READ_SNAPSHOT_CONSISTENCY.md` | `4f5920a25e6a5ba8e5a3f5db82fee8e7a90a5649` |
| GFA-DATA-051 | Nullable telemetry fabricated into valid zero/false values on Projection reads | P1 retrospective | CLOSED | `53_STAGE_14_13_NULLABLE_TELEMETRY_INTEGRITY.md` | `1f30bae8bb8a9e4e27d634b44362dcf7547e54ff` |
| GFA-DATA-052 | Historical pagination cursor omitted stable-order terms and could skip records | P1 retrospective | CLOSED | `54_STAGE_14_14_COMPOSITE_HISTORICAL_PAGINATION_CURSOR.md` | `6a78070499ec0cbe9f905fa94d4b0995d41f2a40` |
| GFA-MAINT-053 | Weather composition responsibility concentration | P3 retrospective | CLOSED | `55_STAGE_14_15_WEATHER_COMPOSITION_BOUNDARY.md` | `cd5e3540cd4f849f606c50433f4e033548b59002` |
| GFA-GOV-054 | High-risk backend remediations lacked one permanent consolidated correctness gate | P2 retrospective | CLOSED | `56_BACKEND_FINAL_CORRECTNESS_AUDIT.md` | `483815bdd60251e16960ec480cadd3bb93ee7f28` |
| GFA-DATA-055 | Provider telemetry availability was lost before persistence | P1 retrospective | CLOSED | `57_STAGE_14_16_END_TO_END_TELEMETRY_AVAILABILITY.md` | `9cfa9005baf9467ed94621602efd48e8b108bb44` |
| GFA-GOV-056 | Missing unified cross-stack Stage 14 acceptance gate | P2 retrospective | CLOSED | `70_STAGE_14_FINAL_COMPLETION_AUDIT.md` | `eb37e03c6793314e446cdb048ae9584e38f2567c` |
| GFA-GOV-057 | Flight Feature timestamp PostgreSQL CI coverage gap | P2 retrospective | CLOSED | `70_STAGE_14_FINAL_COMPLETION_AUDIT.md` | `eb37e03c6793314e446cdb048ae9584e38f2567c` |
| GFA-SEC-058 | Reachable Go standard-library vulnerabilities and toolchain drift | P1 retrospective | CLOSED | `70_STAGE_14_FINAL_COMPLETION_AUDIT.md` | `eb37e03c6793314e446cdb048ae9584e38f2567c` |
| GFA-DB-059 | Migration 018 ambiguous aggregate blocked clean PostgreSQL migration | P1 retrospective | CLOSED | `70_STAGE_14_FINAL_COMPLETION_AUDIT.md` | `eb37e03c6793314e446cdb048ae9584e38f2567c` |
| GFA-TEST-060 | Full FlightStateRepository integration fixture parity drift | P2 retrospective | CLOSED | `70_STAGE_14_FINAL_COMPLETION_AUDIT.md` | `eb37e03c6793314e446cdb048ae9584e38f2567c` |
| GFA-TEST-061 | Terminal ingestion-run fixture omitted `finished_at` | P2 retrospective | CLOSED | `70_STAGE_14_FINAL_COMPLETION_AUDIT.md` | `eb37e03c6793314e446cdb048ae9584e38f2567c` |
| GFA-GOV-062 | Over-broad PostgreSQL fixture-parity audit rule | P3 retrospective | CLOSED | `70_STAGE_14_FINAL_COMPLETION_AUDIT.md` | `eb37e03c6793314e446cdb048ae9584e38f2567c` |
| GFA-OPS-063 | Ingestion runs could remain permanently `running` | P2 retrospective | CLOSED | `83_INGESTION_RUN_LIFECYCLE_HARDENING.md` | `10eaeaff5f40ea7b0432da6a795b6d9a36ff9034` |
| GFA-REL-064 | Auxiliary provider observation could replace the primary data-plane result | P1 retrospective | CLOSED | `84_PROVIDER_HTTP_RESILIENCE_HARDENING.md` | `57a67488e1717f1109eab3a850e09d4525ca444d` |
| GFA-REL-065 | Provider response bodies were not explicitly bounded | P1 retrospective | CLOSED | `84_PROVIDER_HTTP_RESILIENCE_HARDENING.md` | `57a67488e1717f1109eab3a850e09d4525ca444d` |
| GFA-OPS-066 | Provider retry evidence did not control ingestion scheduling | P2 retrospective | CLOSED | `85_INGESTION_RETRY_AND_FALLBACK_EVIDENCE_HARDENING.md` | `bd291eaa758a30329abb10ffb15542c70d05e82e` |
| GFA-DATA-067 | Local policy denial could create false failed provider-run evidence | P2 retrospective | CLOSED | `85_INGESTION_RETRY_AND_FALLBACK_EVIDENCE_HARDENING.md` | `bd291eaa758a30329abb10ffb15542c70d05e82e` |
| GFA-DATA-068 | Fallback history did not preserve the complete ordered attempt chain | P2 retrospective | CLOSED | `85_INGESTION_RETRY_AND_FALLBACK_EVIDENCE_HARDENING.md` | `bd291eaa758a30329abb10ffb15542c70d05e82e` |
| GFA-REL-069 | OpenSky polling reservation and unauthorized-response lifecycle were incomplete | P2 retrospective | CLOSED | `85_INGESTION_RETRY_AND_FALLBACK_EVIDENCE_HARDENING.md` | `bd291eaa758a30329abb10ffb15542c70d05e82e` |
| GFA-DB-070 | OurAirports publication processing lacked durable reservation/commit ownership | P1 retrospective | CLOSED | `86_OURAIRPORTS_PUBLICATION_LIFECYCLE_HARDENING.md` | `db73719ec134da627128038f9be413f38cf4e0e6` |
| GFA-DATA-071 | HTTP validator ordering could hide a failed publication behind `304 Not Modified` | P1 retrospective | CLOSED | `86_OURAIRPORTS_PUBLICATION_LIFECYCLE_HARDENING.md` | `db73719ec134da627128038f9be413f38cf4e0e6` |
| GFA-OPS-072 | Provider transport could start before durable ingestion-run evidence existed | P1 retrospective | CLOSED | `87_INGESTION_DURABILITY_REPLAY_PARTIAL_HARDENING.md` | `302158e4c9cbfb8532ee03147f6dcd31603b72fa` |
| GFA-DATA-073 | Flight State replay had no provider-observation uniqueness contract | P1 retrospective | CLOSED | `87_INGESTION_DURABILITY_REPLAY_PARTIAL_HARDENING.md` | `302158e4c9cbfb8532ee03147f6dcd31603b72fa` |
| GFA-DATA-074 | Durable observation writes followed by downstream failure were classified only as failed | P1 retrospective | CLOSED | `87_INGESTION_DURABILITY_REPLAY_PARTIAL_HARDENING.md` | `302158e4c9cbfb8532ee03147f6dcd31603b72fa` |
| GFA-DB-075 | Post-closure duplicate migration version recurrence | P1 retrospective | CLOSED | `87_INGESTION_DURABILITY_REPLAY_PARTIAL_HARDENING.md` | `302158e4c9cbfb8532ee03147f6dcd31603b72fa` |
| GFA-DATA-076 | Exact in-memory deduplication compared an incomplete observation identity | P1 retrospective | CLOSED | `88_EXACT_DEDUPLICATION_AND_AIRPLANESLIVE_TELEMETRY_HARDENING.md` | `eef7fdc056ebef71f95cfd17ce986dcf429f6c62` |
| GFA-DATA-077 | Airplanes.live missing/null/invalid telemetry could collapse into observed zero | P1 retrospective | CLOSED | `88_EXACT_DEDUPLICATION_AND_AIRPLANESLIVE_TELEMETRY_HARDENING.md` | `eef7fdc056ebef71f95cfd17ce986dcf429f6c62` |
| GFA-DATA-078 | Provider time conversion could overflow or perform unsafe subtraction | P2 retrospective | CLOSED | `88_EXACT_DEDUPLICATION_AND_AIRPLANESLIVE_TELEMETRY_HARDENING.md` | `eef7fdc056ebef71f95cfd17ce986dcf429f6c62` |
| GFA-REL-079 | Airplanes.live provider could be constructed or called without a client | P2 retrospective | CLOSED | `88_EXACT_DEDUPLICATION_AND_AIRPLANESLIVE_TELEMETRY_HARDENING.md` | `eef7fdc056ebef71f95cfd17ce986dcf429f6c62` |
| GFA-OPS-080 | Production provider budgets were process-local | P1 retrospective | CLOSED | `89_PROVIDER_BUDGET_DURABILITY_AND_RETRY_SCHEDULING.md` | `52a60d2b7136919e3a2ccf4850f6d542c6447461` |
| GFA-OPS-081 | Exhausted provider-reported budget could deny without an actionable retry time | P2 retrospective | CLOSED | `89_PROVIDER_BUDGET_DURABILITY_AND_RETRY_SCHEDULING.md` | `52a60d2b7136919e3a2ccf4850f6d542c6447461` |
| GFA-OPS-082 | Provider health evidence was collected but ignored by automatic selection | P2 retrospective | CLOSED | `90_HEALTH_AWARE_TRAFFIC_PROVIDER_SELECTION.md` | `a9896ade17f6a36b80a5cef6abb8ffd9a5687cc1` |
| GFA-DATA-083 | Successful provider batches had no explicit malformed-item accounting policy | P1 retrospective | CLOSED | `91_MALFORMED_PROVIDER_BATCH_POLICY.md` | `b7bf2b762290e55a45fa8d40641435248d1aeddf`; closure extension `6b922cbd9df1bff3f880ad120dd883b37f658e53` |
| GFA-DATA-084 | External retry/token durations could overflow Go duration arithmetic | P2 retrospective | CLOSED | `92_INGESTION_REVIEW_CLOSURE_REPAIR.md` | `6b922cbd9df1bff3f880ad120dd883b37f658e53` |
| GFA-DATA-085 | Open-Meteo missing metrics could be represented as observed zero end to end | P1 retrospective | CLOSED | `92_INGESTION_REVIEW_CLOSURE_REPAIR.md` | `6b922cbd9df1bff3f880ad120dd883b37f658e53` |
| GFA-CONTRACT-086 | OurAirports fail-whole publication behavior was not an explicit typed contract | P2 retrospective | CLOSED | `92_INGESTION_REVIEW_CLOSURE_REPAIR.md` | `6b922cbd9df1bff3f880ad120dd883b37f658e53` |
| GFA-TEST-087 | Isolated PostgreSQL fixtures drifted behind current repository migration dependencies | P2 retrospective | CLOSED | `92_INGESTION_REVIEW_CLOSURE_REPAIR.md` | `6b922cbd9df1bff3f880ad120dd883b37f658e53` |
| GFA-GOV-088 | Ingestion review closure could not be claimed from local/source evidence alone | P2 retrospective | CLOSED | `92_INGESTION_REVIEW_CLOSURE_REPAIR.md` | `6b922cbd9df1bff3f880ad120dd883b37f658e53`; later race guard `1ddb65c5e5471ce180314cc38a4b6d7baad80cd3` |
| GFA-OPS-089 | Process liveness was used as production readiness while PostgreSQL could be unavailable | P1 retrospective | CLOSED | `93_SERVER_AND_HTTP_PROTECTION_REVIEW_CLOSURE.md` | `1fc925c91117eebbb7c90c4bd6b3889548d55cb4` |
| GFA-OPS-090 | Server process lifecycle lacked one controlled shutdown/error-propagation path | P2 retrospective | CLOSED | `94_SERVER_REVIEW_FULL_CLOSURE.md` | `2573892ad7684f3d2646378e2021638a53173bc3` |
| GFA-OBS-091 | Request logging could capture a provisional status before centralized error handling | P2 retrospective | CLOSED | `94_SERVER_REVIEW_FULL_CLOSURE.md` | `2573892ad7684f3d2646378e2021638a53173bc3` |
| GFA-SEC-092 | Raw internal error messages could be written to server logs | P1 retrospective | CLOSED | `94_SERVER_REVIEW_FULL_CLOSURE.md` | `2573892ad7684f3d2646378e2021638a53173bc3` |
| GFA-ARCH-093 | Read-only Historical HTTP registration depended on a read/write store contract | P3 retrospective | CLOSED | `94_SERVER_REVIEW_FULL_CLOSURE.md` | `2573892ad7684f3d2646378e2021638a53173bc3` |
| GFA-OPS-094 | Application rate limiting could throttle the readiness probe | P1 retrospective | CLOSED | `94_SERVER_REVIEW_FULL_CLOSURE.md` | `2573892ad7684f3d2646378e2021638a53173bc3` |
| GFA-SEC-095 | Client identity behind reverse proxies lacked an explicit trust contract | P1 retrospective | CLOSED | `95_TRUSTED_PROXY_AND_BUILD_METADATA_CLOSURE.md` | `cfb079b6f881b03b517f92b06210c3fdc9968893` |
| GFA-GOV-096 | Application version endpoint lacked build-derived revision provenance | P2 retrospective | CLOSED | `95_TRUSTED_PROXY_AND_BUILD_METADATA_CLOSURE.md` | `cfb079b6f881b03b517f92b06210c3fdc9968893` |
| GFA-TEST-097 | Container PostgreSQL readiness could pass during bootstrap handoff | P2 retrospective | CLOSED | `95_TRUSTED_PROXY_AND_BUILD_METADATA_CLOSURE.md` | `ae4d486d2341974a47173e2aedd78da530130cf6` |
| GFA-GOV-098 | Permanent race-detector coverage omitted concurrency-sensitive ingestion ownership | P2 retrospective | CLOSED | `96_INGESTION_RACE_COVERAGE_CLOSURE.md` | `1ddb65c5e5471ce180314cc38a4b6d7baad80cd3` |
| GFA-DATA-099 | Eligibility ran after aircraft-level deduplication | P1 retrospective | CLOSED | `97_ANALYTICAL_CONTRIBUTOR_SEMANTICS_HARDENING.md` | `c5fd1f32273af9215df9d83d1d40c227d3740646` |
| GFA-DATA-100 | Materially future observations could contribute to analytics | P1 retrospective | CLOSED | `97_ANALYTICAL_CONTRIBUTOR_SEMANTICS_HARDENING.md` | `c5fd1f32273af9215df9d83d1d40c227d3740646` |
| GFA-DATA-101 | Traffic Density accepted non-finite arithmetic inputs | P2 retrospective | CLOSED | `97_ANALYTICAL_CONTRIBUTOR_SEMANTICS_HARDENING.md` | `c5fd1f32273af9215df9d83d1d40c227d3740646` |
| GFA-DATA-102 | Airport Activity was not owned by a concrete airport | P1 retrospective | CLOSED | `98_AIRPORT_AND_GEOGRAPHIC_METRIC_INTEGRITY.md` | `0ae85ccbff7584a993030c0adcdee3290dd4b7bd` |
| GFA-DATA-103 | Traffic Density numerator and denominator used unrelated geographic scopes | P1 retrospective | CLOSED | `98_AIRPORT_AND_GEOGRAPHIC_METRIC_INTEGRITY.md` | `0ae85ccbff7584a993030c0adcdee3290dd4b7bd` |
| GFA-DATA-104 | Published analytical provenance was incomplete or placeholder-based | P1 retrospective | CLOSED | `99_PROVENANCE_AND_ANALYTICAL_TRUST_HARDENING.md` | `a31fd8ce3fb6f42a9c90a5153f902c37e7b0f111` |
| GFA-SEC-105 | Public analytical failures could expose raw operation error text | P1 retrospective | CLOSED | `99_PROVENANCE_AND_ANALYTICAL_TRUST_HARDENING.md` | `a31fd8ce3fb6f42a9c90a5153f902c37e7b0f111` |
| GFA-DATA-106 | Analytical recent queries accepted a zero reference time | P2 retrospective | CLOSED | `100_QUERY_AND_ARCHITECTURE_CONSOLIDATION.md` | `b8ccbf590ef3b9ffc221d72e0274e1d78da6c650` |
| GFA-DATA-107 | Accepted UUID identifiers were not canonicalized before deduplication | P2 retrospective | CLOSED | `100_QUERY_AND_ARCHITECTURE_CONSOLIDATION.md` | `b8ccbf590ef3b9ffc221d72e0274e1d78da6c650` |
| GFA-CONTRACT-108 | Analytical Metric IDs were inconsistent across packages and clients | P2 retrospective | CLOSED | `100_QUERY_AND_ARCHITECTURE_CONSOLIDATION.md` | `b8ccbf590ef3b9ffc221d72e0274e1d78da6c650` |
| GFA-ARCH-109 | Metric execution exposed concrete dependencies and retained legacy calculator state | P3 retrospective | CLOSED | `100_QUERY_AND_ARCHITECTURE_CONSOLIDATION.md` | `b8ccbf590ef3b9ffc221d72e0274e1d78da6c650` |
| GFA-DATA-110 | Production Coverage Score and Data Freshness trusted caller-owned snapshot evidence | P1 retrospective | CLOSED | `101_SERVER_OWNED_QUALITY_METRICS.md` | `e48cb27655326fc6cc41d176a50120cdbf1ced6e` |
| GFA-CONTRACT-111 | Rejected analytical parameters could produce a response and still continue query execution | P2 retrospective | CLOSED | `101_SERVER_OWNED_QUALITY_METRICS.md` | `e48cb27655326fc6cc41d176a50120cdbf1ced6e` |
| GFA-TEST-112 | Temporary frontend verification used a non-hermetic external `node_modules` symlink | P2 retrospective | CLOSED | `101_SERVER_OWNED_QUALITY_METRICS.md` | `e48cb27655326fc6cc41d176a50120cdbf1ced6e` |
| GFA-TEST-113 | Analytical remediation verification assumed existing working-tree dependencies were healthy | P2 retrospective | CLOSED | `101_SERVER_OWNED_QUALITY_METRICS.md` | `e48cb27655326fc6cc41d176a50120cdbf1ced6e` |
| GFA-CONTRACT-114 | Public analytical numeric precision had no explicit contract | P2 retrospective | CLOSED | `102_ANALYTICAL_CORE_REVIEW_CLOSURE.md` | `8aa8dfa9f0cb0f5eae94497939633f100a863ef8` |
| GFA-PERF-115 | Source ordering used a manual quadratic sorting helper | P3 retrospective | CLOSED | `102_ANALYTICAL_CORE_REVIEW_CLOSURE.md` | `8aa8dfa9f0cb0f5eae94497939633f100a863ef8` |
| GFA-GOV-116 | Analytical Core closure lacked one permanent strict audit and complete CI path reachability | P2 retrospective | CLOSED | `102_ANALYTICAL_CORE_REVIEW_CLOSURE.md` | `8aa8dfa9f0cb0f5eae94497939633f100a863ef8` |
| GFA-TEST-117 | Permanent Analytical Core source audit was over-sensitive to legal formatter layout | P2 retrospective | CLOSED | `102_ANALYTICAL_CORE_REVIEW_CLOSURE.md` | `8aa8dfa9f0cb0f5eae94497939633f100a863ef8` |
| GFA-SEC-118 | Production Next.js baseline remained on 16.2.9 after the July 21, 2026 advisory set | P2 retrospective | CLOSED | `103_NEXT_16_2_11_SECURITY_CLOSURE.md` | `48f274754fa0fbdbe4ed0a2b8f95985f38183629` |
| GFA-SEC-119 | PostCSS releases through 8.5.17 remained accepted after a high-severity path traversal advisory | P1 retrospective | CLOSED | `103_NEXT_16_2_11_SECURITY_CLOSURE.md` | `48f274754fa0fbdbe4ed0a2b8f95985f38183629` |
| GFA-DATA-120 | Feature Pipeline successful result exposed two independent `FlightFeatures` sources | P2 retrospective | CLOSED | `104_FEATURE_PIPELINE_REVIEW_TRIAGE_AND_CONTRACT_INTEGRITY.md` | `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1` |
| GFA-DATA-121 | Validation status mirrors were not an enforced cross-object invariant | P2 retrospective | CLOSED | `104_FEATURE_PIPELINE_REVIEW_TRIAGE_AND_CONTRACT_INTEGRITY.md` | `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1`; durable extension `abd038c10d1d382843dbaefb8b506efeff5fdeda` |
| GFA-DATA-122 | Incomplete or contradictory validation reports could be accepted | P1 retrospective | CLOSED | `104_FEATURE_PIPELINE_REVIEW_TRIAGE_AND_CONTRACT_INTEGRITY.md` | `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1`; durable extension `abd038c10d1d382843dbaefb8b506efeff5fdeda` |
| GFA-ARCH-123 | Core Feature Pipeline depended on the full feature-store contract | P3 retrospective | CLOSED | `104_FEATURE_PIPELINE_REVIEW_TRIAGE_AND_CONTRACT_INTEGRITY.md` | `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1` |
| GFA-DB-124 | PostgreSQL feature composition allowed ambiguous Pool and Executor ownership | P2 retrospective | CLOSED | `104_FEATURE_PIPELINE_REVIEW_TRIAGE_AND_CONTRACT_INTEGRITY.md` | `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1` |
| GFA-OPS-125 | Feature processing silently substituted a nil caller context | P2 retrospective | CLOSED | `104_FEATURE_PIPELINE_REVIEW_TRIAGE_AND_CONTRACT_INTEGRITY.md` | `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1` |
| GFA-ARCH-126 | Typed-nil Feature Pipeline dependencies passed construction checks | P2 retrospective | CLOSED | `104_FEATURE_PIPELINE_REVIEW_TRIAGE_AND_CONTRACT_INTEGRITY.md` | `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1` |
| GFA-GOV-127 | PostgreSQL Feature Pipeline verifier was absent from required CI | P2 retrospective | CLOSED | `104_FEATURE_PIPELINE_REVIEW_TRIAGE_AND_CONTRACT_INTEGRITY.md` | `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1` |
| GFA-DATA-128 | Feature composition version was absent from the processing version manifest | P2 retrospective | CLOSED | `104_FEATURE_PIPELINE_REVIEW_TRIAGE_AND_CONTRACT_INTEGRITY.md` | `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1` |
| GFA-TEST-129 | Materializer/verifier consumers still validated the obsolete duplicate Result feature value | P2 retrospective | CLOSED | `104_FEATURE_PIPELINE_REVIEW_TRIAGE_AND_CONTRACT_INTEGRITY.md` | `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1` |
| GFA-DATA-130 | Durable feature snapshot identity omitted processing version | P1 retrospective | CLOSED | `105_FEATURE_SNAPSHOT_PROCESSING_IDENTITY.md` | `ab452c0cd039619e842c1991ec1bed10a42e5665`; corrections `f18d43689d53301db862bc10c0445c90dc6f277d`, `96751055657d75ee7800e40c8225ee114b0b52e4` |
| GFA-DB-131 | Processing-aware PostgreSQL list query retained the old predicate/placeholder contract | P1 retrospective | CLOSED | `106_FEATURE_PROCESSING_IDENTITY_POSTGRES_LIST_FIX.md` | `f18d43689d53301db862bc10c0445c90dc6f277d` |
| GFA-TEST-132 | Isolated PostgreSQL feature-store fixture omitted processing identity | P2 retrospective | CLOSED | `107_FEATURE_PROCESSING_IDENTITY_TEST_FIXTURE_ALIGNMENT.md` | `96751055657d75ee7800e40c8225ee114b0b52e4` |
| GFA-DATA-133 | Complete Feature Pipeline validation evidence was transient and disappeared after persistence | P1 retrospective | CLOSED | `108_FEATURE_PIPELINE_VALIDATION_AUDIT_AND_FINAL_CLOSURE.md` | `abd038c10d1d382843dbaefb8b506efeff5fdeda` |

## Stage-level closure evidence

Document 78 is deliberately not registered as one synthetic finding. It is the Stage 14.36 closure/audit document that explains the independent final acceptance boundary after the individual Stage 14 findings were remediated.

Canonical Stage 14 closure evidence:

```text
committed technical baseline before final closure:
f414f6638f8ba5fbe61321e55a21ff3ac91a4986

final closure commit:
202c00cabb352b50a6d3a2a6002ad3401c1ad23e

stage-level status:
STAGE_14_OVERALL_STATUS=CLOSED
```

The earlier invalid closure/reopening chronology remains preserved in Documents 70 and 71.

## Canonical-status interpretation

A finding status is narrower than a stage or release status.

For example, `GFA-DB-001=CLOSED` means the migration atomicity defect is closed. It does not imply that every Stage 14 PostgreSQL debt, production validation item, or V1 release condition was closed at the same historical moment. Broader stage and release status remains governed by the corresponding later closure documents and current release evidence.

A retrospective severity classification is an engineering reconstruction from the documented failure mode and impact. It is not represented as a historical severity label when the original remediation evidence did not record one.

A later post-closure finding does not automatically rewrite an earlier stage-level closure. The later finding must be registered and remediated on its own evidence boundary, while the closure document remains responsible only for the scope it explicitly claimed.

## Retroactive enrichment progress

```text
Documents 41–77 = enriched to canonical remediation-history standard
Document 70 = unique audit/closure blockers reconciled without duplicating later finding owners
Document 78 = enriched as stage-level closure-governance evidence by design, not a synthetic finding
Document 79 = enriched as post-closure remediation history
Documents 83–92 = post-Stage-14 Ingestion / Provider remediation chain enriched and merged
Documents 93–96 = post-Stage-14 Server / HTTP remediation chain enriched and merged through PR #117 (`ac478bfb0ab5796890e30ba99f5dae0a4a09589a`)
Documents 97–102 = Analytical Core remediation and closure chain enriched and merged through PR #118 (`48b91fd87289c54c8492f90aa3c47ec0de61d4d6`)
Document 103 = post-Analytical frontend security remediation enriched and merged through PR #119 (`3104b191e76d7257bcf965fb68537241405e5845`)
Documents 104–108 = Feature Pipeline review and closure chain enriched to canonical standard in source
Canonical finding register covers 133 findings (001–133 with category prefixes)
Stage 14 retrospective finding extraction and canonical ownership reconciliation = CLOSED
Post-Stage-14 Ingestion / Provider finding extraction = CLOSED
Post-Stage-14 Server / HTTP finding extraction = CLOSED
Analytical Core original review reconciliation = CLOSED; 14 FIXED findings mapped to canonical GFA owners, 3 deliberately retained and 2 rejected non-blocking observations preserved as non-defect dispositions
Document 103 security extraction = CLOSED AND MERGED
Feature Pipeline review reconciliation = CLOSED IN SOURCE; accepted FP findings mapped to canonical owners, FP-07/FP-09 retained non-blocking, stale/mechanical observations preserved as non-defect dispositions; exact-head pull-request CI/merge evidence pending
Documents 80–82 = closure/standard summary layer; no duplicate finding IDs created
Next post-Stage-14 audit range = Documents 109–115 (Extractor composition/correctness review chain)
README / DOCUMENT_INDEX navigation reconciliation = COMPLETE IN SOURCE; pull-request and merge evidence remain external GitHub history
```