# Document 107 — Feature Processing Identity Test Fixture Alignment

Status: IMPLEMENTED
Baseline commit: f18d43689d53301db862bc10c0445c90dc6f277d

## Continuous Integration evidence

Backend Continuous Integration run `30157732731` applied the complete production
migration catalog successfully. The PostgreSQL correctness step then failed in:

```text
TestPostgresStorePreservesExactNanosecondsAndRejectsMirrorDrift
```

The exact PostgreSQL error was:

```text
column "processing_version" of relation "flight_feature_snapshots" does not exist
SQLSTATE 42703
```

## Root cause

The timestamp-consistency integration test creates an isolated PostgreSQL schema
and manually creates `flight_feature_snapshots`. Its fixture still represented
the pre-processing-identity schema:

- no `processing_version` column;
- uniqueness on trajectory, schema and as-of time only.

The production migration was correct, but it could not affect the isolated
schema created later by the test.

## Correction

The fixture DDL is now a named source contract and includes:

```sql
processing_version text NOT NULL
```

Its uniqueness contract now includes:

```text
trajectory_id
schema_version
processing_version
as_of_time_unix_nano
```

A unit test owns these fragments, and the permanent feature-processing-identity
audit verifies that the PostgreSQL integration fixture remains processing-aware.

## Closure condition

Final FP-02 closure still requires a green Backend Continuous Integration run on
the corrective commit, including the PostgreSQL feature-pipeline verifier.

```text
FEATURE_PROCESSING_IDENTITY_TEST_FIXTURE_ALIGNMENT=ENFORCED
```

---

## Canonical remediation history

### GFA-TEST-132 — isolated PostgreSQL feature-store fixture drifted behind processing-identity schema

1. **Finding / symptom.** After migration `026` and the production list-query correction, PostgreSQL correctness still failed because an isolated integration fixture manually created `flight_feature_snapshots` without `processing_version`.
2. **Root cause.** The test owned private DDL outside the production migration path and that DDL was not updated when processing version became part of the production schema/uniqueness contract.
3. **Failure scenario.** The timestamp-consistency integration test creates its isolated schema after production migrations complete; repository code references `processing_version`, but the manually created table does not have the column, producing SQLSTATE `42703`.
4. **Impact.** Backend CI cannot complete and the isolated fixture no longer verifies current repository behavior; if ignored, the test becomes misleading rather than protective.
5. **Severity rationale.** **P2 retrospective.** This is test/integration infrastructure drift rather than a production schema defect, but it blocks reliable PostgreSQL correctness evidence for a high-risk persistence change.
6. **Existing guarantees violated.** Full repository fixtures that model production tables must stay structurally compatible with the repository contracts they test.
7. **Considered solutions.** Disable the integration test; run production migrations inside every isolated fixture; update the named fixture DDL and audit its critical schema fragments.
8. **Chosen remediation.** Add `processing_version text NOT NULL`, extend fixture uniqueness to include processing version, name the DDL source contract and verify it with unit/permanent audit checks.
9. **Why this solution was selected.** The fixture intentionally isolates timestamp behavior, so keeping a small explicit table is valid as long as its repository-relevant columns and uniqueness semantics are guarded.
10. **Rejected alternatives.** Disabling the test loses nanosecond/mirror coverage; applying the entire production catalog inside every specialized isolated fixture would increase setup coupling and cost without improving the targeted test.
11. **Trade-offs.** Maintainers must update the explicit fixture when repository-required columns or identity constraints change.
12. **Regression tests / protection.** Unit tests own fixture DDL fragments; the processing-identity audit verifies the column and uniqueness identity; Backend PostgreSQL 16 Integration executes the fixture.
13. **Adversarial review findings.** Fixture-parity enforcement should target repository-relevant complete table fixtures, not mechanically require every minimal test-local schema to mirror all production columns.
14. **Remediation iterations.** `ab452c0c…` introduced the schema contract; `f18d4368…` fixed the production list query; Backend CI run `30157732731` then exposed this isolated fixture; `96751055…` aligned it.
15. **Residual risks and limitations.** Manually owned fixtures can drift again if future schema changes are not added to the permanent audit. Minimal purpose-built fixtures remain allowed where repository code does not require the omitted columns.
16. **Operational or deployment consequences.** None in production; CI regains valid PostgreSQL correctness coverage.
17. **Exact evidence.** Corrective commit `96751055657d75ee7800e40c8225ee114b0b52e4` (`fix: align feature snapshot postgres test fixture`); Backend CI failure run `30157732731` and SQLSTATE `42703` are recorded in this document. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-TEST-132=CLOSED`.
19. **Prevention / future guard.** Isolated fixtures that recreate production repository tables must expose named DDL contracts and be included in schema-semantic audits for every required identity column/constraint.
