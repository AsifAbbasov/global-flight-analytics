# Document 105 — Feature Snapshot Processing Identity

Status: IMPLEMENTED
Baseline commit: 312afe2b9ddcc05da0c2068e50c05e0741a7a1c1

## Problem

The previous durable snapshot identity used only trajectory, feature schema and
as-of time. Different processing contracts could collide in one snapshot slot.

## Contract

Every snapshot now carries processing version in provenance, the in-memory key,
record identifier, PostgreSQL row, uniqueness boundary and read filters.

Existing rows are backfilled as:

```text
flight-feature-processing-legacy-v1
```

New production materialization uses:

```text
flight-feature-processing-pipeline-v1
```

Blank versions at old internal call sites normalize to the current version for
source compatibility. Explicit legacy and future versions remain readable.

```text
FP-02_PROCESSING_IDENTITY_STATUS=CLOSED
FEATURE_SNAPSHOT_PROCESSING_IDENTITY=ENFORCED
```

## Legacy compatibility

The migration updates both the relational processing-version column and the
embedded JSON provenance. Existing identifiers are accepted through the
documented legacy identifier algorithm; new identifiers include processing
version.

## Continuous verification

The PostgreSQL verifier stores two records with the same trajectory, schema,
as-of time and input fingerprint but distinct processing versions.

## Migration catalog ownership

Migration `026_flight_feature_processing_identity.sql` is registered in the
canonical production migration catalog. The catalog regression test requires
exactly twenty-six contiguous migrations and verifies the canonical filename.

## Semantic contract audit

The feature-pipeline contract audit parses the Go syntax tree. It validates the
types of `Config.Writer` and `Config.ProcessingVersion` without depending on
gofmt whitespace alignment.

---

## Canonical remediation history

### GFA-DATA-130 — `FP-02`: durable feature snapshot identity omitted processing version

1. **Finding / symptom.** Snapshot identity was based on trajectory, feature schema and as-of time but did not distinguish the processing contract that produced the feature payload.
2. **Root cause.** Processing-version provenance existed conceptually outside the durable key, record identifier, indexes, PostgreSQL uniqueness and read filters.
3. **Failure scenario.** The same trajectory/schema/as-of input is materialized under two processing algorithms or generations; both target one durable identity and one can collide with or overwrite the other.
4. **Impact.** Distinct analytical feature generations can become indistinguishable, making replay, comparison and audit results incorrect.
5. **Severity rationale.** **P1 retrospective.** This is a durable data-identity defect: two semantically different outputs could occupy the same persistence identity.
6. **Existing guarantees violated.** Snapshot identity must include every processing dimension that can materially change the stored feature result.
7. **Considered solutions.** Treat processing version as metadata only; create a separate history table; extend the canonical snapshot key/ID/uniqueness/read contract with processing version.
8. **Chosen remediation.** Add processing version to feature provenance, pipeline configuration, `SnapshotKey`, record-ID generation, memory indexes, PostgreSQL schema/uniqueness, list/latest filters and verification.
9. **Why this solution was selected.** It makes processing generation part of the existing canonical identity instead of creating a parallel persistence model.
10. **Rejected alternatives.** Metadata-only provenance cannot prevent collisions; a second history table would duplicate store semantics and complicate reads without stronger guarantees.
11. **Trade-offs.** Existing rows need explicit legacy classification and callers/readers must select or normalize a processing version.
12. **Regression tests / protection.** Unit tests cover key/ID normalization; migration `026` owns the column/uniqueness change; the PostgreSQL verifier stores two otherwise-identical snapshots with different processing versions; the permanent processing-identity audit protects code and schema ownership.
13. **Adversarial review findings.** Legacy compatibility must not rewrite old rows as if they were produced by the current algorithm; blank legacy call sites may normalize for source compatibility, while persisted historical rows retain the explicit legacy marker.
14. **Remediation iterations.** Core identity landed in `ab452c0c…`; Backend CI then exposed a missing processing predicate/placeholder in the non-cursor PostgreSQL list query (Document 106, `f18d4368…`) and a stale isolated PostgreSQL fixture (Document 107, `96751055…`). Final closure therefore depends on the full three-commit chain, not the first migration alone.
15. **Residual risks and limitations.** Processing version distinguishes declared processing contracts; it cannot detect an implementation change if maintainers incorrectly fail to advance the version when required.
16. **Operational or deployment consequences.** Migration `026_flight_feature_processing_identity.sql` changes the production feature-snapshot schema and uniqueness boundary; legacy rows are explicitly backfilled and new materialization writes the current processing version.
17. **Exact evidence.** Primary implementation commit `ab452c0cd039619e842c1991ec1bed10a42e5665` (`fix: enforce feature snapshot processing identity`); corrective commits `f18d43689d53301db862bc10c0445c90dc6f277d` and `96751055657d75ee7800e40c8225ee114b0b52e4`. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, migrations, commits, documents and CI evidence.
18. **Final canonical status.** `GFA-DATA-130=CLOSED`.
19. **Prevention / future guard.** Any processing contract capable of changing durable features must be represented in processing identity and tested by storing two snapshots that differ only by that contract dimension.
