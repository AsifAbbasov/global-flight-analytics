# Finding Register — Global Flight Analytics

Status: Canonical Finding Registry v1.6

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
| GFA-SEC-039 | Unprotected mutation/computation-triggering HTTP boundary | P1 retrospective | CLOSED | `45_STAGE_14_5_MUTATION_ENDPOINT_PROTECTION.md` | `50831ae06cb1a38c321ec8c7766bc1f28ddb5757` |
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
Canonical finding register covers 62 findings (001–062 with category prefixes)
Stage 14 retrospective finding extraction and canonical ownership reconciliation = CLOSED
README / DOCUMENT_INDEX navigation reconciliation = final remaining documentation-governance step in this PR
```
