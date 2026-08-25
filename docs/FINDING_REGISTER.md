# Finding Register — Global Flight Analytics

Status: Canonical Finding Registry v1.20

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
| GFA-DATA-134 | Extractor fingerprint omitted effective composition policy and component identity | P1 retrospective | CLOSED | `109_EXTRACTOR_COMPOSITION_PROCESSING_IDENTITY.md` | `a4c563112abd459c90e23e33d191c5f059e5044f` |
| GFA-DATA-135 | Resolved aircraft metadata was absent from deterministic extraction input identity | P1 retrospective | CLOSED | `109_EXTRACTOR_COMPOSITION_PROCESSING_IDENTITY.md` | `a4c563112abd459c90e23e33d191c5f059e5044f` |
| GFA-CONTRACT-136 | Custom aircraft not-found classifier lacked stable policy identity | P2 retrospective | CLOSED | `109_EXTRACTOR_COMPOSITION_PROCESSING_IDENTITY.md` | `a4c563112abd459c90e23e33d191c5f059e5044f` |
| GFA-ARCH-137 | Typed-nil aircraft lookup could pass extractor-composition validation | P2 retrospective | CLOSED | `109_EXTRACTOR_COMPOSITION_PROCESSING_IDENTITY.md` | `a4c563112abd459c90e23e33d191c5f059e5044f` |
| GFA-DATA-138 | Current aircraft metadata could leak into historical feature materialization | P1 retrospective | CLOSED | `110_AIRCRAFT_METADATA_TEMPORAL_SAFETY.md` | `f574911a27b4bad10ddf137689b35286fdb485d3` |
| GFA-CONTRACT-139 | Zero-valued extractor composition fields ambiguously meant omission/defaults | P2 retrospective | CLOSED | `111_EXTRACTOR_COMPOSITION_EXPLICIT_CONFIG.md` | `ff84eefdb8ab363e2bdd276f99e49df7235fb50f` |
| GFA-DATA-140 | Nested trajectory evidence after `AsOfTime` could enter historical extraction | P1 retrospective | CLOSED | `112_EXTRACTOR_INPUT_CORRECTNESS_HARDENING.md` | `e853f5931c78f6ed7b0fbcd0dd85a53cfbaa22f3` |
| GFA-OPS-141 | Extractor silently replaced a nil caller context | P2 retrospective | CLOSED | `112_EXTRACTOR_INPUT_CORRECTNESS_HARDENING.md` | `e853f5931c78f6ed7b0fbcd0dd85a53cfbaa22f3` |
| GFA-ARCH-142 | Typed-nil extractor dependencies violated required/optional semantics | P2 retrospective | CLOSED | `112_EXTRACTOR_INPUT_CORRECTNESS_HARDENING.md` | `e853f5931c78f6ed7b0fbcd0dd85a53cfbaa22f3` |
| GFA-OPS-143 | Cancellation during aircraft enrichment was not rechecked before hashing/publication | P2 retrospective | CLOSED | `112_EXTRACTOR_INPUT_CORRECTNESS_HARDENING.md` | `e853f5931c78f6ed7b0fbcd0dd85a53cfbaa22f3` |
| GFA-DATA-144 | Invalid evidence counts could enter completeness calculation | P1 retrospective | CLOSED | `112_EXTRACTOR_INPUT_CORRECTNESS_HARDENING.md` | `e853f5931c78f6ed7b0fbcd0dd85a53cfbaa22f3` |
| GFA-DATA-145 | Non-finite or out-of-range trajectory quality could be masked into ordinary scores | P1 retrospective | CLOSED | `112_EXTRACTOR_INPUT_CORRECTNESS_HARDENING.md` | `e853f5931c78f6ed7b0fbcd0dd85a53cfbaa22f3` |
| GFA-DATA-146 | Semantically equivalent aircraft identity text produced different extraction fingerprints | P2 retrospective | CLOSED | `112_EXTRACTOR_INPUT_CORRECTNESS_HARDENING.md` | `e853f5931c78f6ed7b0fbcd0dd85a53cfbaa22f3` |
| GFA-DATA-147 | Optional aircraft enrichment incorrectly reduced required-feature completeness | P1 retrospective | CLOSED | `113_EXTRACTOR_QUALITY_AND_PROVENANCE_SEMANTICS.md` | `3cdd9b1532609c872343d00626ba44a9c9855609` |
| GFA-DATA-148 | Trajectory `EndTime` was fabricated as system record update provenance | P1 retrospective | CLOSED | `113_EXTRACTOR_QUALITY_AND_PROVENANCE_SEMANTICS.md` | `3cdd9b1532609c872343d00626ba44a9c9855609` |
| GFA-DATA-149 | Aircraft enrichment lacked durable source/provider/retrieval provenance | P1 retrospective | CLOSED | `113_EXTRACTOR_QUALITY_AND_PROVENANCE_SEMANTICS.md` | `3cdd9b1532609c872343d00626ba44a9c9855609` |
| GFA-MAINT-150 | ICAO24 normalization/validation ownership was duplicated | P3 retrospective | CLOSED | `114_EXTRACTOR_REVIEW_FINAL_CLOSURE.md` | `bcf7ff3e1a83024ee346c16638de0b389baf7e7a` |
| GFA-MAINT-151 | Defensive trajectory cloning ownership was duplicated in extractor | P3 retrospective | CLOSED | `114_EXTRACTOR_REVIEW_FINAL_CLOSURE.md` | `bcf7ff3e1a83024ee346c16638de0b389baf7e7a` |
| GFA-CONTRACT-152 | Aircraft feature field count was duplicated instead of schema-derived | P2 retrospective | CLOSED | `114_EXTRACTOR_REVIEW_FINAL_CLOSURE.md` | `bcf7ff3e1a83024ee346c16638de0b389baf7e7a` |
| GFA-DATA-153 | Deterministic fingerprint mirrors lacked structural drift protection | P1 retrospective | CLOSED | `114_EXTRACTOR_REVIEW_FINAL_CLOSURE.md` | `bcf7ff3e1a83024ee346c16638de0b389baf7e7a` |
| GFA-GOV-154 | Extractor review closure lacked one permanent CI-enforced reconciliation gate | P2 retrospective | CLOSED | `114_EXTRACTOR_REVIEW_FINAL_CLOSURE.md` | `bcf7ff3e1a83024ee346c16638de0b389baf7e7a` |
| GFA-DATA-155 | Typed extractor processing manifest was not persisted as feature provenance | P2 retrospective | CLOSED | `115_EXTRACTOR_COMPOSITION_REVIEW_HARDENING.md` | `28ff8414388d2e81db4e74b83a2fd0c23880d56a` |
| GFA-CONTRACT-156 | Optional aircraft enrichment had no explicit disabled configuration mode | P2 retrospective | CLOSED | `115_EXTRACTOR_COMPOSITION_REVIEW_HARDENING.md` | `28ff8414388d2e81db4e74b83a2fd0c23880d56a` |
| GFA-CONTRACT-157 | Aircraft caching had no explicit disabled mode | P2 retrospective | CLOSED | `115_EXTRACTOR_COMPOSITION_REVIEW_HARDENING.md` | `28ff8414388d2e81db4e74b83a2fd0c23880d56a` |
| GFA-DATA-158 | Enrichment/cache modes were absent from deterministic processing identity | P1 retrospective | CLOSED | `115_EXTRACTOR_COMPOSITION_REVIEW_HARDENING.md` | `28ff8414388d2e81db4e74b83a2fd0c23880d56a` |
| GFA-ARCH-159 | Extractor composition exposed concrete implementation instead of behavior | P3 retrospective | CLOSED | `115_EXTRACTOR_COMPOSITION_REVIEW_HARDENING.md` | `28ff8414388d2e81db4e74b83a2fd0c23880d56a` |
| GFA-MAINT-160 | Extractor composition construction mixed validation/component assembly responsibilities | P3 retrospective | CLOSED | `115_EXTRACTOR_COMPOSITION_REVIEW_HARDENING.md` | `28ff8414388d2e81db4e74b83a2fd0c23880d56a` |
| GFA-CONC-161 | Cache acquire and in-flight registration were not one atomic decision | P1 retrospective | CLOSED | `116_AIRCRAFT_PROVIDER_REVIEW_HARDENING.md` | `92691d993d7340112399a40bd9ecbc62ddb240ad` |
| GFA-CONC-162 | Shared lookup lifetime was coupled to one caller cancellation | P1 retrospective | CLOSED | `116_AIRCRAFT_PROVIDER_REVIEW_HARDENING.md` | `92691d993d7340112399a40bd9ecbc62ddb240ad` |
| GFA-PERF-163 | Aircraft metadata cache lacked bounded lifecycle guarantees | P2 retrospective | CLOSED | `116_AIRCRAFT_PROVIDER_REVIEW_HARDENING.md` | `92691d993d7340112399a40bd9ecbc62ddb240ad` |
| GFA-ARCH-164 | Default provider not-found policy depended on PostgreSQL errors | P2 retrospective | CLOSED | `116_AIRCRAFT_PROVIDER_REVIEW_HARDENING.md` | `92691d993d7340112399a40bd9ecbc62ddb240ad` |
| GFA-DATA-165 | Successful aircraft lookup did not require a valid matching identity | P1 retrospective | CLOSED | `116_AIRCRAFT_PROVIDER_REVIEW_HARDENING.md` | `92691d993d7340112399a40bd9ecbc62ddb240ad` |
| GFA-OPS-166 | Aircraft provider silently substituted nil request context | P2 retrospective | CLOSED | `116_AIRCRAFT_PROVIDER_REVIEW_HARDENING.md` | `92691d993d7340112399a40bd9ecbc62ddb240ad` |
| GFA-DATA-167 | Feature Store idempotency did not prove semantic output identity | P1 retrospective | CLOSED | `117_FEATURE_STORE_REVIEW_HARDENING.md` | `624fcf44d3260bd35ac32f67c0730689713198c0` |
| GFA-CONTRACT-168 | Feature Store serialized the Go domain model directly | P2 retrospective | CLOSED | `117_FEATURE_STORE_REVIEW_HARDENING.md` | `624fcf44d3260bd35ac32f67c0730689713198c0` |
| GFA-DATA-169 | Memory and PostgreSQL Stores enforced different write-validity contracts | P1 retrospective | CLOSED | `117_FEATURE_STORE_REVIEW_HARDENING.md` | `624fcf44d3260bd35ac32f67c0730689713198c0` |
| GFA-DATA-170 | Feature Store writes did not require complete durable validation proof | P1 retrospective | CLOSED | `117_FEATURE_STORE_REVIEW_HARDENING.md` | `624fcf44d3260bd35ac32f67c0730689713198c0` |
| GFA-PERF-171 | Memory Feature Store growth was unbounded | P2 retrospective | CLOSED | `117_FEATURE_STORE_REVIEW_HARDENING.md` | `624fcf44d3260bd35ac32f67c0730689713198c0` |
| GFA-OPS-172 | Feature Store accepted nil caller contexts | P2 retrospective | CLOSED | `117_FEATURE_STORE_REVIEW_HARDENING.md` | `624fcf44d3260bd35ac32f67c0730689713198c0` |
| GFA-DATA-173 | Versioned Flight Features schema omitted four produced geographical fields | P1 retrospective | CLOSED | `118_FLIGHT_FEATURES_SCHEMA_REVIEW_HARDENING.md` | `2eb2c49c11fc1894969ef62c0f9ea0e244a3103f` |
| GFA-DATA-174 | Production Temporal evidence could be empty when raw points were intentionally absent | P2 retrospective | CLOSED | `119_TEMPORAL_BUILDER_REVIEW_HARDENING.md` | `c0e3323328f81af8bf0b8841b1bf6756d3085d21` |
| GFA-DATA-175 | Zero duration overloaded valid zero and invalid/missing metadata | P2 retrospective | CLOSED | `119_TEMPORAL_BUILDER_REVIEW_HARDENING.md` | `c0e3323328f81af8bf0b8841b1bf6756d3085d21` |
| GFA-CONTRACT-176 | Fractional-second duration policy was duplicated and implicit | P2 retrospective | CLOSED | `119_TEMPORAL_BUILDER_REVIEW_HARDENING.md` | `c0e3323328f81af8bf0b8841b1bf6756d3085d21` |
| GFA-OPS-177 | Temporal Builder accepted nil context and lacked bounded cancellation checks | P2 retrospective | CLOSED | `119_TEMPORAL_BUILDER_REVIEW_HARDENING.md` | `c0e3323328f81af8bf0b8841b1bf6756d3085d21` |
| GFA-DATA-178 | Temporal limitations omitted exact rejected-evidence counts | P3 retrospective | CLOSED | `119_TEMPORAL_BUILDER_REVIEW_HARDENING.md` | `c0e3323328f81af8bf0b8841b1bf6756d3085d21` |
| GFA-DATA-179 | Materialized point-count metadata was not reconciled with actual point evidence | P2 retrospective | CLOSED | `119_TEMPORAL_BUILDER_REVIEW_HARDENING.md` | `c0e3323328f81af8bf0b8841b1bf6756d3085d21` |
| GFA-DATA-180 | Geographical point evidence lacked one deterministic temporal eligibility and ordering contract | P1 retrospective | CLOSED | `120_GEOGRAPHICAL_BUILDER_REVIEW_HARDENING.md` | `1bbfd0147092baf2615f5bb0838ca12768b54846` |
| GFA-DATA-181 | Segment fallback bridged unobserved discontinuities into path distance | P1 retrospective | CLOSED | `120_GEOGRAPHICAL_BUILDER_REVIEW_HARDENING.md` | `1bbfd0147092baf2615f5bb0838ca12768b54846` |
| GFA-DATA-182 | Segment fallback supporting-point count used endpoint-coordinate cardinality instead of authoritative evidence | P2 retrospective | CLOSED | `120_GEOGRAPHICAL_BUILDER_REVIEW_HARDENING.md` | `1bbfd0147092baf2615f5bb0838ca12768b54846` |
| GFA-DATA-183 | Circular longitude envelope semantics were conflated with chronological antimeridian crossing | P1 retrospective | CLOSED | `120_GEOGRAPHICAL_BUILDER_REVIEW_HARDENING.md` | `1bbfd0147092baf2615f5bb0838ca12768b54846` |
| GFA-CONTRACT-184 | Geographical numeric policies were implicit and accumulation was less stable than necessary | P2 retrospective | CLOSED | `120_GEOGRAPHICAL_BUILDER_REVIEW_HARDENING.md` | `1bbfd0147092baf2615f5bb0838ca12768b54846` |
| GFA-OPS-185 | Geographical Builder accepted nil context and long geometry passes lacked a complete cancellation/diagnostic contract | P2 retrospective | CLOSED | `120_GEOGRAPHICAL_BUILDER_REVIEW_HARDENING.md` | `1bbfd0147092baf2615f5bb0838ca12768b54846` |
| GFA-DB-186 | Production feature materialization did not hydrate operational point evidence | P1 retrospective | CLOSED | `121_OPERATIONAL_BUILDER_REVIEW_HARDENING.md` | `0b5ec52b503f1ef65c2ca5eeaba485e381710649` |
| GFA-DATA-187 | `FlightState` telemetry availability was lost when converted to `TrackPoint4D` | P1 retrospective | CLOSED | `121_OPERATIONAL_BUILDER_REVIEW_HARDENING.md` | `0b5ec52b503f1ef65c2ca5eeaba485e381710649` |
| GFA-DATA-188 | Operational points lacked trajectory-window filtering and deterministic ordering | P1 retrospective | CLOSED | `121_OPERATIONAL_BUILDER_REVIEW_HARDENING.md` | `0b5ec52b503f1ef65c2ca5eeaba485e381710649` |
| GFA-DATA-189 | Out-of-range finite headings were normalized into apparently valid evidence | P1 retrospective | CLOSED | `121_OPERATIONAL_BUILDER_REVIEW_HARDENING.md` | `0b5ec52b503f1ef65c2ca5eeaba485e381710649` |
| GFA-DATA-190 | Ground/airborne share denominator treated unavailable booleans as observed false | P1 retrospective | CLOSED | `121_OPERATIONAL_BUILDER_REVIEW_HARDENING.md` | `0b5ec52b503f1ef65c2ca5eeaba485e381710649` |
| GFA-DATA-191 | Ground altitude status could erase conflicting non-zero altitude evidence | P1 retrospective | CLOSED | `121_OPERATIONAL_BUILDER_REVIEW_HARDENING.md` | `0b5ec52b503f1ef65c2ca5eeaba485e381710649` |
| GFA-DATA-192 | Barometric and geometric altitude observations could be mixed into one mean | P1 retrospective | CLOSED | `121_OPERATIONAL_BUILDER_REVIEW_HARDENING.md` | `0b5ec52b503f1ef65c2ca5eeaba485e381710649` |
| GFA-DATA-193 | Heading change bridged unavailable or invalid observations | P1 retrospective | CLOSED | `121_OPERATIONAL_BUILDER_REVIEW_HARDENING.md` | `0b5ec52b503f1ef65c2ca5eeaba485e381710649` |
| GFA-DATA-194 | Operational `SupportingPointCount` counted raw records instead of usable contributing observations | P2 retrospective | CLOSED | `121_OPERATIONAL_BUILDER_REVIEW_HARDENING.md` | `0b5ec52b503f1ef65c2ca5eeaba485e381710649` |
| GFA-OPS-195 | Operational aggregation substituted nil contexts and lacked complete cancellation/finite-aggregation guards | P2 retrospective | CLOSED | `121_OPERATIONAL_BUILDER_REVIEW_HARDENING.md` | `0b5ec52b503f1ef65c2ca5eeaba485e381710649` |
| GFA-DATA-196 | Trajectory point-count and quality support had competing ownership rules | P1 retrospective | CLOSED | `122_TRAJECTORY_BUILDER_REVIEW_HARDENING.md` | `2872eb31e87500bdae1ae58fe2b75fb76c4b11d2` |
| GFA-DATA-197 | Sampling and path calculations did not share one canonical eligible point sequence | P1 retrospective | CLOSED | `122_TRAJECTORY_BUILDER_REVIEW_HARDENING.md` | `2872eb31e87500bdae1ae58fe2b75fb76c4b11d2` |
| GFA-DATA-198 | Duplicate observation timestamps created artificial zero sampling intervals | P2 retrospective | CLOSED | `122_TRAJECTORY_BUILDER_REVIEW_HARDENING.md` | `2872eb31e87500bdae1ae58fe2b75fb76c4b11d2` |
| GFA-DATA-199 | Trajectory path logic bridged declared discontinuities and could suppress better segment fallback | P1 retrospective | CLOSED | `122_TRAJECTORY_BUILDER_REVIEW_HARDENING.md` | `2872eb31e87500bdae1ae58fe2b75fb76c4b11d2` |
| GFA-DATA-200 | Coverage could claim complete observation without actual point or segment evidence | P1 retrospective | CLOSED | `122_TRAJECTORY_BUILDER_REVIEW_HARDENING.md` | `2872eb31e87500bdae1ae58fe2b75fb76c4b11d2` |
| GFA-DATA-201 | Coverage-gap clipping, overlap, zero-duration and duration-mirror semantics were inconsistent | P1 retrospective | CLOSED | `122_TRAJECTORY_BUILDER_REVIEW_HARDENING.md` | `2872eb31e87500bdae1ae58fe2b75fb76c4b11d2` |
| GFA-DATA-202 | Empty Trajectory evidence inflated field availability and allowed unsupported zero quality | P1 retrospective | CLOSED | `122_TRAJECTORY_BUILDER_REVIEW_HARDENING.md` | `2872eb31e87500bdae1ae58fe2b75fb76c4b11d2` |
| GFA-DATA-203 | Path-efficiency ratio was unconditionally clamped instead of distinguishing numerical noise from semantic contradiction | P1 retrospective | CLOSED | `122_TRAJECTORY_BUILDER_REVIEW_HARDENING.md` | `2872eb31e87500bdae1ae58fe2b75fb76c4b11d2` |
| GFA-OPS-204 | Trajectory Builder accepted nil context and large scans lacked bounded cancellation checks | P2 retrospective | CLOSED | `122_TRAJECTORY_BUILDER_REVIEW_HARDENING.md` | `2872eb31e87500bdae1ae58fe2b75fb76c4b11d2` |
| GFA-MAINT-205 | Unknown segment statuses produced unbounded duplicate limitation records | P3 retrospective | CLOSED | `122_TRAJECTORY_BUILDER_REVIEW_HARDENING.md` | `2872eb31e87500bdae1ae58fe2b75fb76c4b11d2` |
| GFA-DATA-206 | Partial availability downgraded mathematical integrity failures to warnings | P1 retrospective | CLOSED | `123_VALIDATOR_REVIEW_HARDENING.md` | `39549504bbeff1a6c272153bf3dcde469b766202` |
| GFA-DATA-207 | Partial or unavailable evidence could lack an explanatory domain limitation | P1 retrospective | CLOSED | `123_VALIDATOR_REVIEW_HARDENING.md` | `39549504bbeff1a6c272153bf3dcde469b766202` |
| GFA-DATA-208 | Unavailable required groups could retain stale or non-finite residual payload values | P1 retrospective | CLOSED | `123_VALIDATOR_REVIEW_HARDENING.md` | `39549504bbeff1a6c272153bf3dcde469b766202` |
| GFA-DATA-209 | Available observation-derived groups could claim zero supporting observations | P1 retrospective | CLOSED | `123_VALIDATOR_REVIEW_HARDENING.md` | `39549504bbeff1a6c272153bf3dcde469b766202` |
| GFA-DATA-210 | Ground/airborne share reconciliation ignored whether on-ground evidence was actually available | P2 retrospective | CLOSED | `123_VALIDATOR_REVIEW_HARDENING.md` | `39549504bbeff1a6c272153bf3dcde469b766202` |
| GFA-DATA-211 | Revalidation retained stale Validator and group-derived quality limitations | P1 retrospective | CLOSED | `123_VALIDATOR_REVIEW_HARDENING.md` | `39549504bbeff1a6c272153bf3dcde469b766202` |
| GFA-DATA-212 | Numeric tolerance was treated as an absolute quantity across incompatible units | P1 retrospective | CLOSED | `123_VALIDATOR_REVIEW_HARDENING.md` | `39549504bbeff1a6c272153bf3dcde469b766202` |
| GFA-OPS-213 | Validator silently accepted a nil caller context | P2 retrospective | CLOSED | `123_VALIDATOR_REVIEW_HARDENING.md` | `39549504bbeff1a6c272153bf3dcde469b766202` |
| GFA-CONTRACT-214 | Historical metric identity, aggregation and scope policy was fragmented across builders | P1 retrospective | CLOSED | `124_HISTORICAL_CONTRACT_REVIEW_HARDENING.md` | `fc254881fa446c7e80f94a959e2a9d5609874821` |
| GFA-DATA-215 | Count metrics could accept fractional, negative, non-finite or non-exact float64 values | P1 retrospective | CLOSED | `124_HISTORICAL_CONTRACT_REVIEW_HARDENING.md` | `fc254881fa446c7e80f94a959e2a9d5609874821` |
| GFA-DATA-216 | Historical numerical tolerance was not coherent across count, ratio and continuous values | P1 retrospective | CLOSED | `124_HISTORICAL_CONTRACT_REVIEW_HARDENING.md` | `fc254881fa446c7e80f94a959e2a9d5609874821` |
| GFA-DATA-217 | Unavailable or zero-coverage buckets could retain analytical value/confidence evidence | P1 retrospective | CLOSED | `124_HISTORICAL_CONTRACT_REVIEW_HARDENING.md` | `fc254881fa446c7e80f94a959e2a9d5609874821` |
| GFA-DATA-218 | Partial and unavailable Historical evidence could be structurally unexplained | P1 retrospective | CLOSED | `124_HISTORICAL_CONTRACT_REVIEW_HARDENING.md` | `fc254881fa446c7e80f94a959e2a9d5609874821` |
| GFA-DATA-219 | Comparison current values were not bound to the aggregation-selected current summary | P1 retrospective | CLOSED | `124_HISTORICAL_CONTRACT_REVIEW_HARDENING.md` | `fc254881fa446c7e80f94a959e2a9d5609874821` |
| GFA-DATA-220 | Confidence reason contributions were not reconciled to the declared confidence score | P1 retrospective | CLOSED | `124_HISTORICAL_CONTRACT_REVIEW_HARDENING.md` | `fc254881fa446c7e80f94a959e2a9d5609874821` |
| GFA-CONTRACT-221 | Historical schema registry omitted semantic fields already present in the result contract | P2 retrospective | CLOSED | `124_HISTORICAL_CONTRACT_REVIEW_HARDENING.md` | `fc254881fa446c7e80f94a959e2a9d5609874821` |
| GFA-CONTRACT-222 | Region scope normalization disagreed across Historical Contract and aggregate/builder boundaries | P2 retrospective | CLOSED | `124_HISTORICAL_CONTRACT_REVIEW_HARDENING.md` | `fc254881fa446c7e80f94a959e2a9d5609874821` |
| GFA-DATA-223 | Historical window duration arithmetic could saturate beyond the time.Duration range | P1 retrospective | CLOSED | `125_HISTORICAL_WINDOW_REVIEW_HARDENING.md` | `caa6a0ee7f1309801e1671e06d38535b28aa2437` |
| GFA-PERF-224 | Bucket limits were not enforced during generation before allocation growth | P2 retrospective | CLOSED | `125_HISTORICAL_WINDOW_REVIEW_HARDENING.md` | `caa6a0ee7f1309801e1671e06d38535b28aa2437` |
| GFA-OPS-225 | Historical Window could lose caller cancellation ownership | P2 retrospective | CLOSED | `125_HISTORICAL_WINDOW_REVIEW_HARDENING.md` | `caa6a0ee7f1309801e1671e06d38535b28aa2437` |
| GFA-DATA-226 | Mutable derived Historical Plan evidence could be trusted without canonical reconstruction | P1 retrospective | CLOSED | `125_HISTORICAL_WINDOW_REVIEW_HARDENING.md` | `caa6a0ee7f1309801e1671e06d38535b28aa2437` |
| GFA-DATA-227 | Execution-only bucket limits contaminated semantic Historical plan identity | P2 retrospective | CLOSED | `125_HISTORICAL_WINDOW_REVIEW_HARDENING.md` | `caa6a0ee7f1309801e1671e06d38535b28aa2437` |
| GFA-DATA-228 | Historical plan fingerprint omitted output-affecting derived bucket/window evidence | P1 retrospective | CLOSED | `125_HISTORICAL_WINDOW_REVIEW_HARDENING.md` | `caa6a0ee7f1309801e1671e06d38535b28aa2437` |
| GFA-CONTRACT-229 | Public boundary stepping accepted unaligned timestamps and duplicated granularity behavior | P2 retrospective | CLOSED | `125_HISTORICAL_WINDOW_REVIEW_HARDENING.md` | `caa6a0ee7f1309801e1671e06d38535b28aa2437` |
| GFA-DATA-230 | Reversed historical intervals could expose negative duration evidence | P2 retrospective | CLOSED | `125_HISTORICAL_WINDOW_REVIEW_HARDENING.md` | `caa6a0ee7f1309801e1671e06d38535b28aa2437` |
| GFA-DB-231 | Historical datasets were not guaranteed to come from one PostgreSQL snapshot | P1 retrospective | CLOSED | `126_HISTORICAL_READ_REVIEW_HARDENING.md` | `b67546391984b4726e05d67a51471d401f921e29`; final `98750a7eb5972cd770e6333f46cd0855eca8ad0e` |
| GFA-DATA-232 | Flight and trajectory overlap predicates did not enforce one half-open interval contract | P1 retrospective | CLOSED | `126_HISTORICAL_READ_REVIEW_HARDENING.md` | `b67546391984b4726e05d67a51471d401f921e29` |
| GFA-DB-233 | Mutable Flight and Trajectory rows could not be reconstructed honestly at historical AsOfTime | P1 retrospective | CLOSED | `126_HISTORICAL_READ_REVIEW_HARDENING.md` | `b67546391984b4726e05d67a51471d401f921e29`; migration 028 |
| GFA-DATA-234 | Historical route membership used route calculation time instead of trajectory event time | P1 retrospective | CLOSED | `126_HISTORICAL_READ_REVIEW_HARDENING.md` | `b67546391984b4726e05d67a51471d401f921e29` |
| GFA-DATA-235 | Global row limiting could occur before latest admissible route selection per trajectory | P1 retrospective | CLOSED | `126_HISTORICAL_READ_REVIEW_HARDENING.md` | `b67546391984b4726e05d67a51471d401f921e29` |
| GFA-DATA-236 | Incomplete Historical read coverage used a pagination sentinel instead of exact matched-row denominator | P1 retrospective | CLOSED | `126_HISTORICAL_READ_REVIEW_HARDENING.md` | `b67546391984b4726e05d67a51471d401f921e29` |
| GFA-PERF-237 | Route payload reads lacked one bounded byte budget and deterministic payload identity | P1 retrospective | CLOSED | `126_HISTORICAL_READ_REVIEW_HARDENING.md` | `b67546391984b4726e05d67a51471d401f921e29` |
| GFA-ARCH-238 | Persistence JSON decoding ownership leaked into downstream Historical builders | P3 retrospective | CLOSED | `126_HISTORICAL_READ_REVIEW_HARDENING.md` | `b67546391984b4726e05d67a51471d401f921e29` |
| GFA-DATA-239 | Nullable identifiers could be erased into apparently present fallback values | P1 retrospective | CLOSED | `126_HISTORICAL_READ_REVIEW_HARDENING.md` | `b67546391984b4726e05d67a51471d401f921e29` |
| GFA-DATA-240 | PostgreSQL numeric-to-float conversion had implicit rounding semantics | P2 retrospective | CLOSED | `126_HISTORICAL_READ_REVIEW_HARDENING.md` | `b67546391984b4726e05d67a51471d401f921e29` |
| GFA-TEST-241 | Alternative Historical transaction executors could bypass production record invariants | P2 retrospective | CLOSED | `126_HISTORICAL_READ_REVIEW_HARDENING.md` | `98750a7eb5972cd770e6333f46cd0855eca8ad0e` |
| GFA-DATA-242 | Global read coverage was copied into every Historical bucket | P1 retrospective | CLOSED | `127_HISTORICAL_SERIES_REVIEW_HARDENING.md` | `02bee5fd59d13927d0ffb995844c83d07327a2f9`; final `c863d03e5de711b78ab94027dbf951129665c110` |
| GFA-DATA-243 | Missing source and generation timestamps were synthesized from the analytical window | P1 retrospective | CLOSED | `127_HISTORICAL_SERIES_REVIEW_HARDENING.md` | `02bee5fd59d13927d0ffb995844c83d07327a2f9`; final `c863d03e5de711b78ab94027dbf951129665c110` |
| GFA-DATA-244 | Malformed Historical limitations were silently discarded | P1 retrospective | CLOSED | `127_HISTORICAL_SERIES_REVIEW_HARDENING.md` | `02bee5fd59d13927d0ffb995844c83d07327a2f9`; final `c863d03e5de711b78ab94027dbf951129665c110` |
| GFA-DATA-245 | Distinct window exclusions could collapse when they shared one reason | P2 retrospective | CLOSED | `127_HISTORICAL_SERIES_REVIEW_HARDENING.md` | `02bee5fd59d13927d0ffb995844c83d07327a2f9` |
| GFA-DATA-246 | Historical total sample accumulation could overflow integer arithmetic | P2 retrospective | CLOSED | `127_HISTORICAL_SERIES_REVIEW_HARDENING.md` | `02bee5fd59d13927d0ffb995844c83d07327a2f9` |
| GFA-MAINT-247 | Historical Series construction concentrated unrelated responsibilities in one build path | P3 retrospective | CLOSED | `127_HISTORICAL_SERIES_REVIEW_HARDENING.md` | `02bee5fd59d13927d0ffb995844c83d07327a2f9` |
| GFA-CONTRACT-248 | Route status-ratio metrics admitted an incompatible airport-pair scope | P1 retrospective | CLOSED | `128_HISTORICAL_ROUTE_REVIEW_HARDENING.md` | `513fa1efc7f3b81b895cdc5f881e294d80362e2e` |
| GFA-DATA-249 | Historical Route accepted compatibility-decoded payloads without full Route Contract validation | P1 retrospective | CLOSED | `128_HISTORICAL_ROUTE_REVIEW_HARDENING.md` | `513fa1efc7f3b81b895cdc5f881e294d80362e2e` |
| GFA-DATA-250 | Persisted Route metadata was not reconciled with the validated payload | P1 retrospective | CLOSED | `128_HISTORICAL_ROUTE_REVIEW_HARDENING.md` | `513fa1efc7f3b81b895cdc5f881e294d80362e2e` |
| GFA-DATA-251 | `StoredAt` affected Historical Route output but was absent from fingerprint identity | P1 retrospective | CLOSED | `128_HISTORICAL_ROUTE_REVIEW_HARDENING.md` | `513fa1efc7f3b81b895cdc5f881e294d80362e2e` |
| GFA-DATA-252 | Global route matched-count evidence could be reused as route-pair coverage | P1 retrospective | CLOSED | `128_HISTORICAL_ROUTE_REVIEW_HARDENING.md` | `513fa1efc7f3b81b895cdc5f881e294d80362e2e` |
| GFA-DATA-253 | Historical Route snapshot query evidence was not bound to the canonical plan | P1 retrospective | CLOSED | `128_HISTORICAL_ROUTE_REVIEW_HARDENING.md` | `513fa1efc7f3b81b895cdc5f881e294d80362e2e` |
| GFA-DATA-254 | Historical Route provenance was derived from unscoped rather than contributing evidence | P1 retrospective | CLOSED | `128_HISTORICAL_ROUTE_REVIEW_HARDENING.md` | `513fa1efc7f3b81b895cdc5f881e294d80362e2e` |
| GFA-DATA-255 | Historical Route trusted persisted distance instead of validated endpoint geometry | P1 retrospective | CLOSED | `128_HISTORICAL_ROUTE_REVIEW_HARDENING.md` | `513fa1efc7f3b81b895cdc5f881e294d80362e2e` |
| GFA-DATA-256 | Historical Route accumulation used ordinary floating-point summation | P2 retrospective | CLOSED | `128_HISTORICAL_ROUTE_REVIEW_HARDENING.md` | `513fa1efc7f3b81b895cdc5f881e294d80362e2e` |
| GFA-CONTRACT-257 | Historical Route metric semantics were insufficiently explicit | P2 retrospective | CLOSED | `128_HISTORICAL_ROUTE_REVIEW_HARDENING.md` | `513fa1efc7f3b81b895cdc5f881e294d80362e2e` |
| GFA-DATA-258 | Latest Route selection could substitute record identity for missing trajectory identity | P1 retrospective | CLOSED | `128_HISTORICAL_ROUTE_REVIEW_HARDENING.md` | `513fa1efc7f3b81b895cdc5f881e294d80362e2e` |
| GFA-MAINT-259 | Historical Route metric calculation responsibilities were concentrated in one switch | P3 retrospective | CLOSED | `128_HISTORICAL_ROUTE_REVIEW_HARDENING.md` | `513fa1efc7f3b81b895cdc5f881e294d80362e2e` |
| GFA-DATA-260 | Historical periods with incompatible coverage could be compared as ordinary growth | P1 retrospective | CLOSED | `129_HISTORICAL_COMPARISON_REVIEW_HARDENING.md` | `21734b85b9f50ae717dca031c798866161895989` |
| GFA-DATA-261 | Direct Historical Comparison output had incomplete two-period provenance and fingerprint identity | P1 retrospective | CLOSED | `129_HISTORICAL_COMPARISON_REVIEW_HARDENING.md` | `21734b85b9f50ae717dca031c798866161895989` |
| GFA-MAINT-262 | Historical Comparison `Attach` concentrated validation, arithmetic, selection, provenance and result construction | P3 retrospective | CLOSED | `129_HISTORICAL_COMPARISON_REVIEW_HARDENING.md` | `21734b85b9f50ae717dca031c798866161895989` |
| GFA-CONTRACT-263 | Historical scope identity depended on structural reflection equality | P2 retrospective | CLOSED | `129_HISTORICAL_COMPARISON_REVIEW_HARDENING.md` | `21734b85b9f50ae717dca031c798866161895989` |
| GFA-MAINT-264 | Historical Comparison exported an internal generic value-selection helper | P3 retrospective | CLOSED | `129_HISTORICAL_COMPARISON_REVIEW_HARDENING.md` | `21734b85b9f50ae717dca031c798866161895989` |
| GFA-DATA-265 | Non-finite percentage arithmetic was detected late and misclassified as invalid source evidence | P2 retrospective | CLOSED | `129_HISTORICAL_COMPARISON_REVIEW_HARDENING.md` | `21734b85b9f50ae717dca031c798866161895989` |
| GFA-TEST-266 | Historical Comparison regression coverage did not protect its critical invariants | P2 retrospective | CLOSED | `129_HISTORICAL_COMPARISON_REVIEW_HARDENING.md` | `21734b85b9f50ae717dca031c798866161895989` |
| GFA-DATA-267 | Similarity score was not separated from evidence confidence | P1 retrospective | CLOSED | `130_HISTORICAL_SIMILARITY_REVIEW_HARDENING.md` | `6dbae4e6fe00295af0f7ba5303855736b76e8bde` |
| GFA-DATA-268 | Historical Similarity ignored authoritative trajectory-quality evidence | P1 retrospective | CLOSED | `130_HISTORICAL_SIMILARITY_REVIEW_HARDENING.md` | `6dbae4e6fe00295af0f7ba5303855736b76e8bde` |
| GFA-PERF-269 | Historical Similarity accepted unbounded sample and input-point sizes | P2 retrospective | CLOSED | `130_HISTORICAL_SIMILARITY_REVIEW_HARDENING.md` | `6dbae4e6fe00295af0f7ba5303855736b76e8bde` |
| GFA-ARCH-270 | Public Similarity `Rank` duplicated the production candidate-selection workflow | P2 retrospective | CLOSED | `130_HISTORICAL_SIMILARITY_REVIEW_HARDENING.md` | `6dbae4e6fe00295af0f7ba5303855736b76e8bde` |
| GFA-DATA-271 | Similarity fingerprint did not bind the exact prepared scoring representation | P1 retrospective | CLOSED | `130_HISTORICAL_SIMILARITY_REVIEW_HARDENING.md` | `6dbae4e6fe00295af0f7ba5303855736b76e8bde` |
| GFA-DATA-272 | Equal-timestamp trajectory points depended on caller order | P1 retrospective | CLOSED | `130_HISTORICAL_SIMILARITY_REVIEW_HARDENING.md` | `6dbae4e6fe00295af0f7ba5303855736b76e8bde` |
| GFA-DATA-273 | Similarity result validation did not recompute analytical mathematics | P1 retrospective | CLOSED | `130_HISTORICAL_SIMILARITY_REVIEW_HARDENING.md` | `6dbae4e6fe00295af0f7ba5303855736b76e8bde` |
| GFA-DATA-274 | Endpoint similarity averaged endpoints instead of using the worse endpoint | P1 retrospective | CLOSED | `130_HISTORICAL_SIMILARITY_REVIEW_HARDENING.md` | `6dbae4e6fe00295af0f7ba5303855736b76e8bde` |
| GFA-DATA-275 | Relative-difference scoring used an undocumented one-kilometre floor | P2 retrospective | CLOSED | `130_HISTORICAL_SIMILARITY_REVIEW_HARDENING.md` | `6dbae4e6fe00295af0f7ba5303855736b76e8bde` |
| GFA-DATA-276 | Resampling linearly interpolated latitude/longitude instead of spherical geometry | P1 retrospective | CLOSED | `130_HISTORICAL_SIMILARITY_REVIEW_HARDENING.md` | `6dbae4e6fe00295af0f7ba5303855736b76e8bde` |
| GFA-MAINT-277 | Similarity preparation, scoring, quality, fingerprinting and validation were overly concentrated | P3 retrospective | CLOSED | `130_HISTORICAL_SIMILARITY_REVIEW_HARDENING.md` | `6dbae4e6fe00295af0f7ba5303855736b76e8bde` |
| GFA-REL-278 | `NewDefault` could panic for package-owned constant configuration | P2 retrospective | CLOSED | `130_HISTORICAL_SIMILARITY_REVIEW_HARDENING.md` | `6dbae4e6fe00295af0f7ba5303855736b76e8bde` |
| GFA-DB-279 | Historical Aggregate region persistence required uppercase while the canonical contract required lowercase | P1 retrospective | CLOSED | `131_HISTORICAL_AGGREGATE_REVIEW_HARDENING.md` | `18dde73b2d122d00476ea21accb256b33fc23527`; migration 029 |
| GFA-DATA-280 | Stored Historical JSON was not reconciled with denormalized row metadata | P1 retrospective | CLOSED | `131_HISTORICAL_AGGREGATE_REVIEW_HARDENING.md` | `18dde73b2d122d00476ea21accb256b33fc23527` |
| GFA-DATA-281 | Stored aggregate record identifiers were not recomputed from canonical result identity | P1 retrospective | CLOSED | `131_HISTORICAL_AGGREGATE_REVIEW_HARDENING.md` | `18dde73b2d122d00476ea21accb256b33fc23527` |
| GFA-DATA-282 | Aggregate idempotency accepted same fingerprint with different canonical payload | P1 retrospective | CLOSED | `131_HISTORICAL_AGGREGATE_REVIEW_HARDENING.md` | `18dde73b2d122d00476ea21accb256b33fc23527` |
| GFA-ARCH-283 | Historical Materialization depended on the full Aggregate Store although it only writes | P3 retrospective | CLOSED | `131_HISTORICAL_AGGREGATE_REVIEW_HARDENING.md` | `18dde73b2d122d00476ea21accb256b33fc23527` |
| GFA-DATA-284 | Aggregate Put canonicalized caller input before validating the original domain value | P1 retrospective | CLOSED | `131_HISTORICAL_AGGREGATE_REVIEW_HARDENING.md` | `18dde73b2d122d00476ea21accb256b33fc23527` |
| GFA-OPS-285 | Historical Aggregate silently substituted nil caller contexts | P2 retrospective | CLOSED | `131_HISTORICAL_AGGREGATE_REVIEW_HARDENING.md` | `18dde73b2d122d00476ea21accb256b33fc23527` |
| GFA-DATA-286 | Aggregate storage time could be zero or precede result generation | P1 retrospective | CLOSED | `131_HISTORICAL_AGGREGATE_REVIEW_HARDENING.md` | `18dde73b2d122d00476ea21accb256b33fc23527` |
| GFA-DB-287 | Historical Aggregate timestamp mirrors lacked database-level consistency constraints | P1 retrospective | CLOSED | `131_HISTORICAL_AGGREGATE_REVIEW_HARDENING.md` | `18dde73b2d122d00476ea21accb256b33fc23527`; migration 029 |
| GFA-DATA-288 | Combined two-period reads allowed cross-period limit starvation | P1 retrospective | CLOSED | `132_HISTORICAL_MATERIALIZATION_REVIEW_HARDENING.md` | `2bbbd2439580536ffe17f8827c654c245d9b6b1e` |
| GFA-DATA-289 | Materialization trusted unverified Historical Read snapshot identity | P1 retrospective | CLOSED | `132_HISTORICAL_MATERIALIZATION_REVIEW_HARDENING.md` | `2bbbd2439580536ffe17f8827c654c245d9b6b1e` |
| GFA-DATA-290 | Combined read summary obscured period-specific evidence quality | P1 retrospective | CLOSED | `132_HISTORICAL_MATERIALIZATION_REVIEW_HARDENING.md` | `2bbbd2439580536ffe17f8827c654c245d9b6b1e` |
| GFA-DATA-291 | Materialization returned a pre-persistence result instead of canonical persisted identity | P1 retrospective | CLOSED | `132_HISTORICAL_MATERIALIZATION_REVIEW_HARDENING.md` | `2bbbd2439580536ffe17f8827c654c245d9b6b1e` |
| GFA-DATA-292 | Materialization duplicated comparison provenance and omitted GeneratedAt from repaired fingerprint identity | P1 retrospective | CLOSED | `132_HISTORICAL_MATERIALIZATION_REVIEW_HARDENING.md` | `2bbbd2439580536ffe17f8827c654c245d9b6b1e` |
| GFA-OPS-293 | Materialization silently replaced nil caller context | P2 retrospective | CLOSED | `132_HISTORICAL_MATERIALIZATION_REVIEW_HARDENING.md` | `2bbbd2439580536ffe17f8827c654c245d9b6b1e` |
| GFA-OPS-294 | Materialization errors did not identify the failed orchestration stage | P2 retrospective | CLOSED | `132_HISTORICAL_MATERIALIZATION_REVIEW_HARDENING.md` | `2bbbd2439580536ffe17f8827c654c245d9b6b1e` |
| GFA-MAINT-295 | Exported DatasetLimitOr duplicated already-normalized request state | P3 retrospective | CLOSED | `132_HISTORICAL_MATERIALIZATION_REVIEW_HARDENING.md` | `2bbbd2439580536ffe17f8827c654c245d9b6b1e` |
| GFA-TEST-296 | Materialization regression coverage omitted critical two-period and persistence invariants | P2 retrospective | CLOSED | `132_HISTORICAL_MATERIALIZATION_REVIEW_HARDENING.md` | `2bbbd2439580536ffe17f8827c654c245d9b6b1e` |
| GFA-DATA-297 | Replay accepted unverified Materialization outcomes | P1 retrospective | CLOSED | `133_HISTORICAL_REPLAY_REVIEW_HARDENING.md` | `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068` |
| GFA-DATA-298 | Replay result status depended on a separate Go error instead of self-contained execution evidence | P1 retrospective | CLOSED | `133_HISTORICAL_REPLAY_REVIEW_HARDENING.md` | `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068` |
| GFA-OPS-299 | Production replay reporting discarded a successfully persisted completed prefix | P1 retrospective | CLOSED | `133_HISTORICAL_REPLAY_REVIEW_HARDENING.md` | `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068` |
| GFA-DATA-300 | Adjacent Replay Materializations could observe inconsistent shared-period evidence | P1 retrospective | CLOSED | `133_HISTORICAL_REPLAY_REVIEW_HARDENING.md` | `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068` |
| GFA-CONTRACT-301 | Global Replay request failures were validated after window execution began | P2 retrospective | CLOSED | `133_HISTORICAL_REPLAY_REVIEW_HARDENING.md` | `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068` |
| GFA-PERF-302 | Replay planning applied the smaller replay limit after larger allocation work | P2 retrospective | CLOSED | `133_HISTORICAL_REPLAY_REVIEW_HARDENING.md` | `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068` |
| GFA-OPS-303 | Replay and production command silently replaced nil caller context | P2 retrospective | CLOSED | `133_HISTORICAL_REPLAY_REVIEW_HARDENING.md` | `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068` |
| GFA-DATA-304 | Replay lacked canonical input identity and public result validation | P1 retrospective | CLOSED | `133_HISTORICAL_REPLAY_REVIEW_HARDENING.md` | `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068` |
| GFA-TEST-305 | Replay regression coverage omitted critical partial-progress and integrity invariants | P2 retrospective | CLOSED | `133_HISTORICAL_REPLAY_REVIEW_HARDENING.md` | `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068` |
| GFA-DATA-306 | Projection points were not bound to one exact horizon grid | P1 retrospective | CLOSED | `134_PROJECTION_CONTRACT_REVIEW_HARDENING.md` | `964556d0ca8a1ce9aa74c37c55961cdd006b3de8` |
| GFA-DATA-307 | Limited results could omit evidence explaining the limitation | P1 retrospective | CLOSED | `134_PROJECTION_CONTRACT_REVIEW_HARDENING.md` | `964556d0ca8a1ce9aa74c37c55961cdd006b3de8` |
| GFA-DATA-308 | Projection confidence was not reconciled with reasons and mandatory evidence | P1 retrospective | CLOSED | `134_PROJECTION_CONTRACT_REVIEW_HARDENING.md` | `964556d0ca8a1ce9aa74c37c55961cdd006b3de8` |
| GFA-CONTRACT-309 | Projection Contract duplicated the shared confidence vocabulary | P2 retrospective | CLOSED | `134_PROJECTION_CONTRACT_REVIEW_HARDENING.md` | `964556d0ca8a1ce9aa74c37c55961cdd006b3de8` |
| GFA-DATA-310 | Projection input fingerprints accepted arbitrary non-empty text | P2 retrospective | CLOSED | `134_PROJECTION_CONTRACT_REVIEW_HARDENING.md` | `964556d0ca8a1ce9aa74c37c55961cdd006b3de8` |
| GFA-CONTRACT-311 | ICAO24 identifiers lacked exact hexadecimal validation | P2 retrospective | CLOSED | `134_PROJECTION_CONTRACT_REVIEW_HARDENING.md` | `964556d0ca8a1ce9aa74c37c55961cdd006b3de8` |
| GFA-CONTRACT-312 | Estimated Arrival accepted non-ICAO airport location indicators | P2 retrospective | CLOSED | `134_PROJECTION_CONTRACT_REVIEW_HARDENING.md` | `964556d0ca8a1ce9aa74c37c55961cdd006b3de8` |
| GFA-DATA-313 | Projection provenance could omit source or observation-basis evidence | P1 retrospective | CLOSED | `134_PROJECTION_CONTRACT_REVIEW_HARDENING.md` | `964556d0ca8a1ce9aa74c37c55961cdd006b3de8` |
| GFA-DATA-314 | Projection provenance chronology allowed retrieval before observation | P1 retrospective | CLOSED | `134_PROJECTION_CONTRACT_REVIEW_HARDENING.md` | `964556d0ca8a1ce9aa74c37c55961cdd006b3de8` |
| GFA-DATA-315 | Duplicate Projection evidence and explanation identities were accepted | P2 retrospective | CLOSED | `134_PROJECTION_CONTRACT_REVIEW_HARDENING.md` | `964556d0ca8a1ce9aa74c37c55961cdd006b3de8` |
| GFA-CONTRACT-316 | Projection Result lacked a typed public validation boundary | P2 retrospective | CLOSED | `134_PROJECTION_CONTRACT_REVIEW_HARDENING.md` | `964556d0ca8a1ce9aa74c37c55961cdd006b3de8` |
| GFA-TEST-317 | Projection Contract regression tests omitted critical cross-field invariants | P2 retrospective | CLOSED | `134_PROJECTION_CONTRACT_REVIEW_HARDENING.md` | `964556d0ca8a1ce9aa74c37c55961cdd006b3de8` |
| GFA-TEST-318 | Projection fixtures encoded invalid aircraft/confidence/provenance evidence | P2 retrospective | CLOSED | `134_PROJECTION_CONTRACT_REVIEW_HARDENING.md` | `964556d0ca8a1ce9aa74c37c55961cdd006b3de8` |
| GFA-DATA-319 | Projection Horizon Step did not define one exact fixed grid | P1 retrospective | CLOSED | `135_PROJECTION_HORIZON_REVIEW_HARDENING.md` | `7249aa7625dd306bbd769dade6ce3262edca01ab`; audit `d2bc87b07ea0eb6a0b9b25f0a1e3cb2cbc52cd1b` |
| GFA-DATA-320 | Public Projection Horizon Plan lacked a canonical integrity boundary | P1 retrospective | CLOSED | `135_PROJECTION_HORIZON_REVIEW_HARDENING.md` | `7249aa7625dd306bbd769dade6ce3262edca01ab`; audit `d2bc87b07ea0eb6a0b9b25f0a1e3cb2cbc52cd1b` |
| GFA-REL-321 | Configured default Projection Horizon was unreachable through production HTTP | P2 retrospective | CLOSED | `135_PROJECTION_HORIZON_REVIEW_HARDENING.md` | `7249aa7625dd306bbd769dade6ce3262edca01ab`; audit `d2bc87b07ea0eb6a0b9b25f0a1e3cb2cbc52cd1b` |
| GFA-DATA-322 | Projection strategy fingerprints duplicated incomplete horizon identity | P1 retrospective | CLOSED | `135_PROJECTION_HORIZON_REVIEW_HARDENING.md` | `7249aa7625dd306bbd769dade6ce3262edca01ab`; audit `d2bc87b07ea0eb6a0b9b25f0a1e3cb2cbc52cd1b` |
| GFA-OPS-323 | Nil Projection Horizon policy returned an unrelated configuration error | P2 retrospective | CLOSED | `135_PROJECTION_HORIZON_REVIEW_HARDENING.md` | `7249aa7625dd306bbd769dade6ce3262edca01ab`; audit `d2bc87b07ea0eb6a0b9b25f0a1e3cb2cbc52cd1b` |
| GFA-CONTRACT-324 | Projection Horizon policy identity was not required to be canonical | P2 retrospective | CLOSED | `135_PROJECTION_HORIZON_REVIEW_HARDENING.md` | `7249aa7625dd306bbd769dade6ce3262edca01ab`; audit `d2bc87b07ea0eb6a0b9b25f0a1e3cb2cbc52cd1b` |
| GFA-PERF-325 | Projection Horizon configuration lacked a package-wide allocation bound | P2 retrospective | CLOSED | `135_PROJECTION_HORIZON_REVIEW_HARDENING.md` | `7249aa7625dd306bbd769dade6ce3262edca01ab`; audit `d2bc87b07ea0eb6a0b9b25f0a1e3cb2cbc52cd1b` |
| GFA-DATA-326 | Projection consumers trusted malformed alternative Horizon plans | P1 retrospective | CLOSED | `135_PROJECTION_HORIZON_REVIEW_HARDENING.md` | `7249aa7625dd306bbd769dade6ce3262edca01ab`; audit `d2bc87b07ea0eb6a0b9b25f0a1e3cb2cbc52cd1b` |
| GFA-TEST-327 | Projection Horizon remediation lacked permanent regression and CI enforcement | P2 retrospective | CLOSED | `135_PROJECTION_HORIZON_REVIEW_HARDENING.md` | `7249aa7625dd306bbd769dade6ce3262edca01ab`; audit `d2bc87b07ea0eb6a0b9b25f0a1e3cb2cbc52cd1b` |
| GFA-DATA-328 | Cutoff snapshots retained post-`as_of` aggregate quality evidence | P1 retrospective | CLOSED | `136_PROJECTION_BASELINE_REVIEW_HARDENING.md` | `0f2c1b2c6f91f104b8e0880e85dc8144fed6a910`; audit `51476c427f77b5a7375cd30b6f9a81d446c1c3f2` |
| GFA-DB-329 | PostgreSQL Projection hydration returned cutoff-unsafe trajectory quality | P1 retrospective | CLOSED | `136_PROJECTION_BASELINE_REVIEW_HARDENING.md` | `0f2c1b2c6f91f104b8e0880e85dc8144fed6a910`; audit `51476c427f77b5a7375cd30b6f9a81d446c1c3f2` |
| GFA-TEST-330 | Future-evidence regression could not detect aggregate quality leakage | P2 retrospective | CLOSED | `136_PROJECTION_BASELINE_REVIEW_HARDENING.md` | `0f2c1b2c6f91f104b8e0880e85dc8144fed6a910`; audit `51476c427f77b5a7375cd30b6f9a81d446c1c3f2` |
| GFA-DATA-331 | Unavailable Projection Baseline results lacked reproducible denial provenance | P1 retrospective | CLOSED | `136_PROJECTION_BASELINE_REVIEW_HARDENING.md` | `0f2c1b2c6f91f104b8e0880e85dc8144fed6a910`; collaboration `560e4ed15cabbf0042110e00363a3a7c4d0c0d2e`; audit `51476c427f77b5a7375cd30b6f9a81d446c1c3f2` |
| GFA-DATA-332 | Projection Baseline fingerprint omitted output-affecting evidence and policy identity | P1 retrospective | CLOSED | `136_PROJECTION_BASELINE_REVIEW_HARDENING.md` | `af9c377193c21c048721e9cc28bf885d6ad276ec`; collaboration `560e4ed15cabbf0042110e00363a3a7c4d0c0d2e`; audit `51476c427f77b5a7375cd30b6f9a81d446c1c3f2` |
| GFA-DATA-333 | Baseline confidence ignored latest-observation age | P1 retrospective | CLOSED | `136_PROJECTION_BASELINE_REVIEW_HARDENING.md` | `af9c377193c21c048721e9cc28bf885d6ad276ec`; audit `51476c427f77b5a7375cd30b6f9a81d446c1c3f2` |
| GFA-DATA-334 | Projection Baseline kinematics lacked conservative physical bounds | P1 retrospective | CLOSED | `136_PROJECTION_BASELINE_REVIEW_HARDENING.md` | `af9c377193c21c048721e9cc28bf885d6ad276ec`; audit `51476c427f77b5a7375cd30b6f9a81d446c1c3f2` |
| GFA-DATA-335 | Altitude and vertical-rate evidence lacked explicit reference semantics | P1 retrospective | CLOSED | `136_PROJECTION_BASELINE_REVIEW_HARDENING.md` | `af9c377193c21c048721e9cc28bf885d6ad276ec`; audit `51476c427f77b5a7375cd30b6f9a81d446c1c3f2` |
| GFA-DATA-336 | Conflicting latest observations at the same timestamp were selected lexically | P1 retrospective | CLOSED | `136_PROJECTION_BASELINE_REVIEW_HARDENING.md` | `560e4ed15cabbf0042110e00363a3a7c4d0c0d2e`; audit `51476c427f77b5a7375cd30b6f9a81d446c1c3f2` |
| GFA-DATA-337 | Allowed on-ground evidence reused the airborne propagation model | P1 retrospective | CLOSED | `136_PROJECTION_BASELINE_REVIEW_HARDENING.md` | `560e4ed15cabbf0042110e00363a3a7c4d0c0d2e`; audit `51476c427f77b5a7375cd30b6f9a81d446c1c3f2` |
| GFA-CONTRACT-338 | Default eligibility made the horizontal-only Projection branch unreachable | P2 retrospective | CLOSED | `136_PROJECTION_BASELINE_REVIEW_HARDENING.md` | `560e4ed15cabbf0042110e00363a3a7c4d0c0d2e`; audit `51476c427f77b5a7375cd30b6f9a81d446c1c3f2` |
| GFA-DATA-339 | Projection Baseline trusted malformed custom eligibility output | P1 retrospective | CLOSED | `136_PROJECTION_BASELINE_REVIEW_HARDENING.md` | `560e4ed15cabbf0042110e00363a3a7c4d0c0d2e`; audit `51476c427f77b5a7375cd30b6f9a81d446c1c3f2` |
| GFA-OPS-340 | Nil Projection Baseline returned an unrelated construction error | P2 retrospective | CLOSED | `136_PROJECTION_BASELINE_REVIEW_HARDENING.md` | `af9c377193c21c048721e9cc28bf885d6ad276ec`; audit `51476c427f77b5a7375cd30b6f9a81d446c1c3f2` |
| GFA-GOV-341 | Projection Baseline remediation lacked permanent regression and CI enforcement | P2 retrospective | CLOSED | `136_PROJECTION_BASELINE_REVIEW_HARDENING.md` | `51476c427f77b5a7375cd30b6f9a81d446c1c3f2` |
| GFA-DATA-342 | Candidate eligibility and duplicate checks occurred after the expensive evaluation budget | P1 retrospective | CLOSED | `137_PROJECTION_NEIGHBORS_REVIEW_HARDENING.md` | `e2a4a7dc76e43942ca9deb0d8d5f83a09a42deff`; audit `c409cc171507050625524af1a0b8b8a6f38b7a75` |
| GFA-DATA-343 | Candidate-budget truncation was input-order sensitive rather than recency deterministic | P2 retrospective | CLOSED | `137_PROJECTION_NEIGHBORS_REVIEW_HARDENING.md` | `e2a4a7dc76e43942ca9deb0d8d5f83a09a42deff`; audit `c409cc171507050625524af1a0b8b8a6f38b7a75` |
| GFA-CONTRACT-344 | Selection limit could exceed the maximum candidate evaluation budget | P2 retrospective | CLOSED | `137_PROJECTION_NEIGHBORS_REVIEW_HARDENING.md` | `e2a4a7dc76e43942ca9deb0d8d5f83a09a42deff`; audit `c409cc171507050625524af1a0b8b8a6f38b7a75` |
| GFA-DATA-345 | Snapshot and selection fingerprints were not fully canonical under equivalent input ordering | P1 retrospective | CLOSED | `137_PROJECTION_NEIGHBORS_REVIEW_HARDENING.md` | `e2a4a7dc76e43942ca9deb0d8d5f83a09a42deff`; audit `c409cc171507050625524af1a0b8b8a6f38b7a75` |
| GFA-OPS-346 | Systemic similarity failures were indistinguishable from ordinary candidate non-comparability | P1 retrospective | CLOSED | `137_PROJECTION_NEIGHBORS_REVIEW_HARDENING.md` | `e2a4a7dc76e43942ca9deb0d8d5f83a09a42deff`; pipeline `353d19bc97f561e1897ece1967e7304c0e10b5fb`; audit `c409cc171507050625524af1a0b8b8a6f38b7a75` |
| GFA-DATA-347 | Neighbor continuation could bridge unobserved temporal gaps | P1 retrospective | CLOSED | `137_PROJECTION_NEIGHBORS_REVIEW_HARDENING.md` | `911a1b102c68af2746a13bfca48b008cf7225ff8`; audit `c409cc171507050625524af1a0b8b8a6f38b7a75` |
| GFA-DATA-348 | Historical neighbors lacked source-attested route-scope ownership | P1 retrospective | CLOSED | `137_PROJECTION_NEIGHBORS_REVIEW_HARDENING.md` | `3eee05fb44484aa6e389af66520aba23d4ae277e`; audit `c409cc171507050625524af1a0b8b8a6f38b7a75` |
| GFA-MAINT-349 | Neighbor selection mixed preparation, evaluation, ranking, assembly, and validation in one coordinator | P3 retrospective | CLOSED | `137_PROJECTION_NEIGHBORS_REVIEW_HARDENING.md` | `353d19bc97f561e1897ece1967e7304c0e10b5fb`; audit `c409cc171507050625524af1a0b8b8a6f38b7a75` |
| GFA-CONTRACT-350 | One truncation flag conflated evaluation-budget truncation with qualified-result limiting | P2 retrospective | CLOSED | `137_PROJECTION_NEIGHBORS_REVIEW_HARDENING.md` | `353d19bc97f561e1897ece1967e7304c0e10b5fb`; audit `c409cc171507050625524af1a0b8b8a6f38b7a75` |
| GFA-DATA-351 | Pattern Confidence fingerprint omitted decision-relevant selected-neighbor evidence | P1 retrospective | CLOSED | `138_PROJECTION_PATTERN_CONFIDENCE_REVIEW_HARDENING.md` | `6e6ac17cfcfca688d57829adfe2468346db6db1a`; continuation `5873ae911b40197ee45eea30e7558aa04af78064`; audit `cd8f114bfef698c51cfc6008ecd2ed01f9c1cc42` |
| GFA-CONTRACT-352 | Pattern Confidence configuration admitted zero-information or internally incoherent policy states | P1 retrospective | CLOSED | `138_PROJECTION_PATTERN_CONFIDENCE_REVIEW_HARDENING.md` | `6e6ac17cfcfca688d57829adfe2468346db6db1a`; distribution `f73534feb275c5e109fa12fcfd9df5b69c56c03a`; continuation `5873ae911b40197ee45eea30e7558aa04af78064`; audit `cd8f114bfef698c51cfc6008ecd2ed01f9c1cc42` |
| GFA-DATA-353 | Mean similarity alone hid weak neighbors and unstable similarity distributions | P1 retrospective | CLOSED | `138_PROJECTION_PATTERN_CONFIDENCE_REVIEW_HARDENING.md` | `f73534feb275c5e109fa12fcfd9df5b69c56c03a`; audit `cd8f114bfef698c51cfc6008ecd2ed01f9c1cc42` |
| GFA-CONTRACT-354 | Pattern Confidence duplicated freshness ownership by scoring candidate age | P1 retrospective | CLOSED | `138_PROJECTION_PATTERN_CONFIDENCE_REVIEW_HARDENING.md` | `f73534feb275c5e109fa12fcfd9df5b69c56c03a`; audit `cd8f114bfef698c51cfc6008ecd2ed01f9c1cc42` |
| GFA-DATA-355 | Pattern Confidence did not verify that selected neighbors agreed on future continuation | P1 retrospective | CLOSED | `138_PROJECTION_PATTERN_CONFIDENCE_REVIEW_HARDENING.md` | `5873ae911b40197ee45eea30e7558aa04af78064`; audit `cd8f114bfef698c51cfc6008ecd2ed01f9c1cc42` |
| GFA-CONTRACT-356 | Production consumers could fall back to a continuation-unaware Pattern Confidence evaluator | P1 retrospective | CLOSED | `138_PROJECTION_PATTERN_CONFIDENCE_REVIEW_HARDENING.md` | optional `5873ae911b40197ee45eea30e7558aa04af78064`; mandatory `e31fcb5bbbb76093305e8b2c137c793a85dc6795`; audit `cd8f114bfef698c51cfc6008ecd2ed01f9c1cc42` |
| GFA-DATA-357 | Pattern Confidence validation did not independently reconstruct the published decision | P1 retrospective | CLOSED | `138_PROJECTION_PATTERN_CONFIDENCE_REVIEW_HARDENING.md` | `6e6ac17cfcfca688d57829adfe2468346db6db1a`; distribution `f73534feb275c5e109fa12fcfd9df5b69c56c03a`; continuation `5873ae911b40197ee45eea30e7558aa04af78064`; audit `cd8f114bfef698c51cfc6008ecd2ed01f9c1cc42` |

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
Documents 104–108 = Feature Pipeline review and closure chain enriched and merged through PR #120 (`9c5cb4541d85787f4d85b6e7c44c55932de97635`)
Documents 109–115 = Extractor processing/composition/correctness review chain enriched and merged through PR #121 (`b977e8929d637ee21ab6a938be3167b6deed0439`)
Documents 116–119 = Aircraft Provider / Feature Store / Flight Features schema / Temporal Builder review chain enriched and merged through PR #122 (`74b1b56858a62b633911d7047b7a12243b5acc04`)
Documents 120–123 = Geographical / Operational / Trajectory Builder / Validator review chain enriched and merged through PR #123 (`46cb38e022fd14094ce261dbeb243e25744fa8fb`)
Documents 124–126 = Historical Contract / Window / Read review chain enriched and merged through PR #124 (`849c2f4b0e99a35cf3c75ca39d2aaeefc35c41b1`)
Documents 127–129 = Historical Series / Route / Comparison review chain enriched and merged through PR #125 (`4f9b1d71177ad89fa43581ef6ed080ce5c1aa1ab`)
Documents 130–131 = Historical Similarity / Aggregate review chain enriched and merged through PR #126 (`feaaba300df6e4273083da2bf13dbc4346fb4425`)
Documents 132–133 = Historical Materialization / Replay review chain enriched and merged through PR #127 (`5847b8b30b8be8900361d95422859bfc5f70044f`)
Documents 134–135 = Projection Contract / Horizon review chain enriched and merged through PR #128 (`de9fcbcf43da759584c91b669bb42a70dfbb95ad`)
Documents 136–138 = Projection Baseline / Neighbors / Pattern Confidence review chain enriched to canonical standard in source
Canonical finding register covers 357 findings (001–357 with category prefixes)
Stage 14 retrospective finding extraction and canonical ownership reconciliation = CLOSED
Post-Stage-14 Ingestion / Provider finding extraction = CLOSED
Post-Stage-14 Server / HTTP finding extraction = CLOSED
Analytical Core original review reconciliation = CLOSED
Feature Pipeline review reconciliation = CLOSED AND MERGED
Extractor review/composition reconciliation = CLOSED AND MERGED
Provider/store/schema/temporal review reconciliation = CLOSED AND MERGED
Builder/Validator review reconciliation = CLOSED AND MERGED
Historical Contract/Window/Read review reconciliation = CLOSED AND MERGED
Historical Series/Route/Comparison review reconciliation = CLOSED AND MERGED
Historical Similarity/Aggregate review reconciliation = CLOSED AND MERGED
Historical Materialization/Replay review reconciliation = CLOSED AND MERGED
Projection Contract/Horizon review reconciliation = CLOSED AND MERGED
Projection Baseline/Neighbors/Pattern Confidence review reconciliation = CLOSED IN SOURCE
Documents 80–82 = closure/standard summary layer; no duplicate finding IDs created
Next post-Stage-14 audit range begins at Document 139 (Projection Freshness review)
README / DOCUMENT_INDEX navigation reconciliation = COMPLETE IN SOURCE; pull-request and merge evidence remain external GitHub history
```
