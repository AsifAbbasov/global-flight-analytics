# Finding Register — Global Flight Analytics

Status: Canonical Finding Registry v1.1

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

## Canonical-status interpretation

A finding status is narrower than a stage or release status.

For example, `GFA-DB-001=CLOSED` means the migration atomicity defect is closed. It does not imply that every Stage 14 PostgreSQL debt, production validation item, or V1 release condition was closed at the same historical moment. Broader stage and release status remains governed by the corresponding later closure documents and current release evidence.

## Retroactive enrichment progress

```text
Documents 58–61 = enriched to the canonical remediation-history standard
Remaining historical remediation documents = audit/enrichment pending
```
