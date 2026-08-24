# Finding Register — Global Flight Analytics

Status: Canonical Finding Registry v1.2

## Purpose

This document is the single index of engineering findings, remediation stages, and final status.

The registry prevents documentation drift between README files, stage documents, commits, and implementation state.

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
12. Regression tests
13. Adversarial review findings
14. Remediation iterations
15. Residual risks and limitations
16. Operational or deployment consequences
17. Exact evidence
18. Final canonical status

## Status values

- OPEN
- IN_PROGRESS
- CLOSED
- ACCEPTED_RISK

## Rule

A finding is not considered closed only because code changed. Closure requires implementation evidence, regression protection, and documentation alignment.

Detailed technical history belongs in stage documents. This registry is the navigation and status authority.

## Evidence honesty rule

Retroactive records must distinguish historical evidence from later reconstruction.

If an original severity, pull-request number, review comment, or reviewer iteration cannot be recovered from repository evidence, the registry or canonical stage document must say so explicitly. A later reconstruction may classify impact or alternatives, but it must not fabricate historical review events.

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

## Canonical-status interpretation

A finding status is narrower than a stage or release status.

For example, `GFA-DB-001=CLOSED` means the migration atomicity defect is closed. It does not imply that every Stage 14 PostgreSQL debt, production validation item, or V1 release condition was closed at the same historical moment. Broader stage and release status remains governed by the corresponding later closure documents and current release evidence.

A retrospective severity classification is an engineering reconstruction from the documented failure mode and impact. It is not represented as a historical severity label when the original remediation evidence did not record one.

## Retroactive enrichment progress

```text
Documents 58–69 = enriched to the canonical remediation-history standard
Documents 71–79 and post-closure remediation evidence = audit/enrichment pending
Document 70 and Document 78 are closure/audit documents and are reviewed for status/evidence consistency rather than automatically treated as one finding each
```
