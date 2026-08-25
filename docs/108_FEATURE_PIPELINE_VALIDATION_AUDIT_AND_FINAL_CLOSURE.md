# Document 108 — Feature Pipeline Validation Audit and Final Closure

Status: IMPLEMENTED
Baseline commit: 96751055657d75ee7800e40c8225ee114b0b52e4

## 1. Remaining review debt

The original Feature Pipeline review correctly observed that the transient
validation report was not part of the stored snapshot. Reading a snapshot could
recover only the copied validation status, not:

- validator version;
- validation time;
- issue counts;
- issue paths, groups, severities, codes and messages.

Two composition observations also lacked final dispositions:

- infrastructure composition remains in the internal feature-pipeline package;
- composition handles expose constructed components.

## 2. Durable validation report

`flightfeatures.FlightFeatures` now owns `ValidationReport`. The validator report
types are aliases of this durable model, so the pipeline does not translate
between two independently evolving report schemas.

A complete report contains:

- audit state;
- validator version;
- validation status;
- error and warning counts;
- complete issues;
- validation time.

The pipeline validates and normalizes the report, attaches it to the validated
features before storage, and replaces the successful result report with the
report returned inside the stored record. An idempotent replay therefore returns
the original durable report rather than a newer transient validation timestamp.

```text
FEATURE_PIPELINE_VALIDATION_AUDIT_TRAIL=CLOSED
```

## 3. Legacy compatibility

Migration `027_flight_feature_validation_audit.sql` backfills snapshots created
before this contract with:

```text
AuditState=legacy_unavailable
```

Legacy rows preserve their known validation status but deliberately leave
validator version, validation time and issues unavailable. The migration does
not invent historical evidence.

New reports use:

```text
AuditState=complete
```

PostgreSQL constraints require:

- a `ValidationReport` JSON object;
- an accepted audit state;
- validation report status equal to the indexed `validation_status` mirror.

## 4. Composition dispositions

### FP-07 — composition placement

```text
FP-07_COMPOSITION_PLACEMENT=DELIBERATELY_RETAINED_NON_BLOCKING
```

Core `pipeline.go` has no PostgreSQL dependency. Infrastructure wiring is
isolated in `postgres_composition.go`, where it owns mutually exclusive
Pool/Executor construction, canonical versions and one reusable composition for
the production materializer and transactional verifier. Moving identical wiring
into commands would duplicate invariants without resolving a correctness fault.

### FP-09 — composition handle

```text
FP-09_COMPOSITION_HANDLE=DELIBERATELY_RETAINED_NON_BLOCKING
```

The composition is an internal package rather than an external public API. Its
pipeline and store handles are required by the operational materializer and
transactional verifier. Additional constructed handles provide diagnostic
construction evidence and remain intentionally internal.

## 5. Permanent evidence

The Feature Pipeline contract audit now requires:

- durable report types and defensive cloning;
- stored-report normalization and validation;
- pipeline attachment before storage;
- replay replacement from the stored record;
- migration `027` constraints;
- PostgreSQL verifier marker;
- final composition dispositions;
- zero unclassified and deferred findings.

The PostgreSQL verifier checks that a complete report survives Put, replay and
Get, then emits:

```text
Validation audit trail: PASS
```

## 6. Final review status

```text
FEATURE_PIPELINE_RELEASE_BLOCKERS=CLOSED
FEATURE_PIPELINE_VALIDATION_AUDIT_TRAIL=CLOSED
FEATURE_PIPELINE_REVIEW_STATUS=CLOSED
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
```

Final closure still requires a green Backend Continuous Integration run on the
corrective commit, including migration application, PostgreSQL correctness tests,
the PostgreSQL feature-pipeline verifier, race tests and the permanent contract
audit.

---

## Canonical remediation history

### GFA-DATA-133 — complete Feature Pipeline validation evidence was transient and disappeared after persistence

1. **Finding / symptom.** A stored feature snapshot retained only validation status; validator version, validation time, counts and issue details existed only in the transient pipeline result.
2. **Root cause.** Validation was modeled as a processing-time report rather than durable evidence owned by the feature snapshot that the report authorized.
3. **Failure scenario.** A feature snapshot is persisted, later reloaded or replayed, and consumers can recover its accepted status but cannot determine which validator version made the decision, when validation occurred or which warnings/issues justified a limited result.
4. **Impact.** Durable analytical features lose the audit evidence needed to reproduce, investigate and trust their validation decision; replay could return a newer transient report unrelated to the originally stored record.
5. **Severity rationale.** **P1 retrospective.** Validation is the admission boundary for durable analytical features; losing its complete evidence after persistence materially weakens data provenance and can misrepresent replay history.
6. **Existing guarantees violated.** Durable snapshots must preserve the evidence that authorized them, and idempotent replay must return the stored decision rather than newly generated audit metadata.
7. **Considered solutions.** Keep only status; store report in a separate audit table; embed the complete normalized report in `FlightFeatures`; reconstruct missing legacy reports.
8. **Chosen remediation.** Make `flightfeatures.FlightFeatures` own `ValidationReport`, alias validator report types to the durable model, normalize/validate it before storage, persist it in JSON, enforce status consistency, and return the stored report on successful/replayed results.
9. **Why this solution was selected.** The validation decision is intrinsic to the stored feature payload; embedding the report keeps identity, replay and persistence atomic without a second lifecycle/table.
10. **Rejected alternatives.** Status-only persistence remains insufficient; a separate audit table introduces synchronization/orphan risks; inventing historical validator/time/issues would fabricate evidence.
11. **Trade-offs.** Snapshot payloads become larger and migrations/validators must maintain the durable report schema. Legacy rows carry explicit unavailable audit state rather than complete detail.
12. **Regression tests / protection.** Pipeline tests prove persistence and idempotent replay return the original stored report; memory/PostgreSQL stores validate report consistency; migration `027` enforces JSON presence/status mirror; PostgreSQL verifier emits `Validation audit trail: PASS`; permanent contract audit protects the boundary.
13. **Adversarial review findings.** Legacy compatibility must preserve known status without synthesizing unknown history, hence `AuditState=legacy_unavailable`; complete reports must be defensively cloned and normalized on both write and read paths.
14. **Remediation iterations.** Transient report integrity was hardened in `312afe2b…`; processing identity/CI regressions were closed through Documents 105–107; the remaining durable-audit debt was closed in `abd038c1…` with migration `027` and final Feature Pipeline dispositions.
15. **Residual risks and limitations.** Rows classified `legacy_unavailable` intentionally lack historical validator/time/issue detail; the repository cannot reconstruct evidence that was never persisted.
16. **Operational or deployment consequences.** Migration `027_flight_feature_validation_audit.sql` changes durable feature snapshot JSON/constraints; new snapshots carry complete audit state while historical rows are explicitly backfilled as unavailable.
17. **Exact evidence.** Implementation commit `abd038c10d1d382843dbaefb8b506efeff5fdeda` (`fix: persist feature validation audit trail`). The predecessor baseline `96751055657d75ee7800e40c8225ee114b0b52e4` already contained the processing-identity fixture fix. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, migrations, commits, documents and CI evidence.
18. **Final canonical status.** `GFA-DATA-133=CLOSED`.
19. **Prevention / future guard.** Any validation or quality decision that authorizes durable analytical data must define whether its evidence is durable; if yes, the persisted record, replay path, migrations and database constraints must preserve it without fabrication.

### Non-finding closure dispositions

`FP-07` and `FP-09` remain deliberately retained non-blocking for the reasons in
Section 4. The old integration/BDUF observation is stale because the production
materializer composes the pipeline. Mechanical naming/line-count/nil-return
observations remain review preferences without demonstrated failure modes. None
of these dispositions receives a canonical defect ID.
