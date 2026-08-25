# Projection Read Review Hardening

Status: closed
Module: `apps/api/internal/projectionintelligence/projectionread`
Review baseline: `87b853e5a74bc5b8e0cd9bcb3f1e8e13eec8df0e`
Permanent-audit baseline: `9dda4b102497028b59280143b86bf84564afb136`

---

## 1. Review result

The original external review contained both stale findings and confirmed production defects.
The stale findings described contracts that had already been replaced before this closure cycle:
`DataSource.LoadSnapshot` already owned one atomic snapshot boundary, the production PostgreSQL
source already used a read-only `REPEATABLE READ` transaction, required telemetry was already
filtered without NULL-to-zero coercion, and current trajectory and current flight evidence were
already excluded from route history.

The confirmed findings were corrected in two engineering increments. The read service now validates
snapshot and Composer postconditions, binds the returned projection to the requested trajectory,
loads and validates route JSON against persisted row metadata, rejects nil contexts, assigns
`GeneratedAt` only after snapshot acquisition, repairs stale `UpdatedAt`, and canonicalizes the
zero-duration sentinel before Production composition. The second increment backfills rejected
historical candidate identifiers and binds route-history fingerprints to the exact contributing
route records and their input fingerprints.

---

## 2. Contract-hardening evidence

```text
CONTRACT_HARDENING_COMMIT=4eeff2b9f5b5c17dd6b7ebe5d0be4a7bd836fb37
CONTRACT_HARDENING_GITHUB_ACTIONS_RUN=30638188394
CONTRACT_HARDENING_POSTGRESQL_16_INTEGRATION_JOB=91181076159
CONTRACT_HARDENING_BACKEND_RACE_SAFETY_JOB=91181076172
CONTRACT_HARDENING_BACKEND_QUALITY_JOB=91181076240
CONTRACT_HARDENING_BACKEND_CONTAINER_JOB=91181409362
ATOMIC_REPEATABLE_READ_SNAPSHOT=CI_CONFIRMED
NIL_CONTEXT_REJECTION=CI_CONFIRMED
SNAPSHOT_IDENTITY_AND_AS_OF_POSTCONDITIONS=CI_CONFIRMED
COMPOSER_OUTPUT_POSTCONDITIONS=CI_CONFIRMED
COMPOSER_OUTPUT_IDENTITY_BINDING=CI_CONFIRMED
ROUTE_ROW_PAYLOAD_BINDING=CI_CONFIRMED
ROUTE_CONTRACT_EARLY_VALIDATION=CI_CONFIRMED
GENERATED_AT_AFTER_SNAPSHOT_ACQUISITION=CI_CONFIRMED
UPDATED_AT_AS_OF_REPAIR=CI_CONFIRMED
DEFAULT_DURATION_BOUNDARY=CI_CONFIRMED
SOURCE_NAME_CANONICALIZATION=CI_CONFIRMED
```

---

## 3. Evidence-hardening evidence

```text
EVIDENCE_HARDENING_COMMIT=9dda4b102497028b59280143b86bf84564afb136
EVIDENCE_HARDENING_GITHUB_ACTIONS_RUN=30648605652
EVIDENCE_HARDENING_BACKEND_QUALITY_JOB=91216081636
EVIDENCE_HARDENING_POSTGRESQL_16_INTEGRATION_JOB=91216081729
EVIDENCE_HARDENING_BACKEND_RACE_SAFETY_JOB=91216081733
EVIDENCE_HARDENING_BACKEND_CONTAINER_JOB=91216395238
HISTORICAL_CANDIDATE_BACKFILL=CI_CONFIRMED
ACCEPTED_CANDIDATE_LIMIT_AFTER_HYDRATION=CI_CONFIRMED
ROUTE_HISTORY_RECORD_LINEAGE=CI_CONFIRMED
ROUTE_HISTORY_INPUT_FINGERPRINT_LINEAGE=CI_CONFIRMED
ROUTE_HISTORY_AGGREGATE_MIRROR_VALIDATION=CI_CONFIRMED
DETERMINISTIC_ROUTE_HISTORY_EVIDENCE_FINGERPRINT=CI_CONFIRMED
```

---

## 4. Review classification

The following review statements were rejected as stale or mechanical rather than production defects:

- four independent public reads without a snapshot;
- NULL telemetry silently converted to zero;
- current trajectory included in route history;
- `AllowOnGround bool` treated as a procedural flag argument;
- optional pointers treated as an automatic architecture violation;
- file length treated as proof of incorrect semantics.

The PostgreSQL source and composition root remain large. This is maintainability debt only; no
unresolved result-integrity defect was established, so decomposition does not block formal closure.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
ENGINEERING_IMPLEMENTATION=COMPLETE
ENGINEERING_DEBT=CLOSED
ADDITIONAL_PRODUCTION_CODE_FIXES_REQUIRED=NO
POSTGRES_SOURCE_DECOMPOSITION=NON_BLOCKING_MAINTAINABILITY_IMPROVEMENT
```

---

## 5. Permanent audit status

Permanent review enforcement is implemented in:

```text
apps/api/tools/projectionreadreviewaudit
```

Backend Quality registration is implemented in:

```text
.github/workflows/backend-ci.yml
```

Permanent audit commit `e0557f6bc3115767ba124a9c94cbb008194c643b` passed exact
push-triggered Backend Continuous Integration run `30651385019`. Backend Quality job
`91225208542` completed successfully and included the permanent `Run projection read review audit`
step. Backend Race Safety job `91225208660`, PostgreSQL 16 Integration job `91225208672`, and
Backend Container job `91225529938` also completed successfully.

```text
PERMANENT_AUDIT_COMMIT=e0557f6bc3115767ba124a9c94cbb008194c643b
PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30651385019
PERMANENT_AUDIT_BACKEND_QUALITY_JOB=91225208542
PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=91225208672
PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=91225208660
PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=91225529938
PERMANENT_REVIEW_AUDIT=CI_CONFIRMED
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_READ_FORMAL_CLOSURE=COMPLETE
PROJECTION_READ_REVIEW_STATUS=CLOSED
```


The formal-closure commit containing this record must pass Backend Quality, Backend Race Safety,
PostgreSQL 16 Integration, and Backend Container before the external final module-closure verdict
is issued.

## Canonical remediation history

The following twelve records reconcile the confirmed Projection Read defects. Stale review claims and the explicitly non-blocking source-decomposition observation remain outside the canonical finding count. Severity is retrospective.

### GFA-OPS-424 — Projection Read silently accepted a nil caller context

1. **Finding / symptom.** The public read boundary could substitute or continue with a nil context rather than requiring caller lifecycle ownership.
2. **Root cause.** Context validation was inconsistent with the hardened repository and Projection module conventions.
3. **Failure scenario.** A caller passes nil and loses cancellation/deadline ownership while expensive snapshot/composition work proceeds.
4. **Impact.** Request lifecycle and cancellation semantics become ambiguous.
5. **Severity rationale.** P2 retrospective because the defect affects operability and resource control but does not fabricate data by itself.
6. **Existing guarantees violated.** Explicit context ownership and fail-closed request lifecycle.
7. **Considered solutions.** Substitute background context; panic; reject nil with a typed error.
8. **Chosen remediation.** Reject nil context at the read service boundary.
9. **Why selected.** Callers, not the service, own request lifetime.
10. **Rejected alternatives.** Background substitution hides programming errors; panic is unnecessary.
11. **Trade-offs.** Legacy nil callers must be corrected.
12. **Regression tests / protection.** Nil-context regression and permanent read review audit.
13. **Adversarial review findings.** Atomic snapshot behavior was already present and is not duplicated as this finding.
14. **Remediation iterations.** Closed in contract-hardening commit.
15. **Residual risks / limitations.** Downstream dependencies must continue propagating the validated context.
16. **Operational/deployment consequences.** Invalid calls fail immediately instead of running detached work.
17. **Exact evidence.** `4eeff2b9f5b5c17dd6b7ebe5d0be4a7bd836fb37`, run `30638188394`; permanent audit `e0557f6bc3115767ba124a9c94cbb008194c643b`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** `projectionreadreviewaudit` protects nil-context rejection.

### GFA-DATA-425 — Snapshot identity and as-of postconditions were not independently enforced by Projection Read

1. **Finding / symptom.** A loaded snapshot could be structurally valid while belonging to another trajectory or analytical cutoff.
2. **Root cause.** Data-source success was trusted without read-service-specific ownership postconditions.
3. **Failure scenario.** A custom/incorrect source returns snapshot evidence for a different trajectory or `AsOfTime` and Production composes it.
4. **Impact.** Projection output can be derived from the wrong target or time boundary.
5. **Severity rationale.** P1 retrospective because snapshot ownership defines the complete production input.
6. **Existing guarantees violated.** Entity identity and temporal snapshot correctness.
7. **Considered solutions.** Trust `LoadSnapshot`; validate selected fields; require exact snapshot identity/as-of equality.
8. **Chosen remediation.** Validate snapshot identity and analytical cutoff immediately after load.
9. **Why selected.** It treats the data source as a trust boundary even when the production PostgreSQL implementation is correct.
10. **Rejected alternatives.** Concrete-source trust does not protect alternative implementations/tests.
11. **Trade-offs.** Custom sources have stricter postconditions.
12. **Regression tests / protection.** Snapshot identity/as-of drift regressions.
13. **Adversarial review findings.** The existing atomic `REPEATABLE READ` snapshot is closure evidence, not a new defect in this cycle.
14. **Remediation iterations.** Closed in contract-hardening increment.
15. **Residual risks / limitations.** Snapshot internal evidence still relies on its own canonical validators.
16. **Operational/deployment consequences.** Foreign snapshots fail closed before composition.
17. **Exact evidence.** `4eeff2b9f5b5c17dd6b7ebe5d0be4a7bd836fb37`, run `30638188394`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit protects snapshot postconditions.

### GFA-DATA-426 — Composer output was not fully bound back to the Projection Read request

1. **Finding / symptom.** A Composer dependency could return a valid result whose trajectory or request identity did not match the read request.
2. **Root cause.** Read orchestration validated the input snapshot more strongly than the returned production result.
3. **Failure scenario.** A substituted composer emits a result for another trajectory and Read returns it as the requested projection.
4. **Impact.** Public API can return semantically foreign output despite successful dependency calls.
5. **Severity rationale.** P1 retrospective because final result ownership is wrong.
6. **Existing guarantees violated.** Request/result identity and dependency postconditions.
7. **Considered solutions.** Trust composer; call only generic result validation; enforce request-specific output identity/postconditions.
8. **Chosen remediation.** Validate Composer output and bind it to the requested trajectory and read-service contract.
9. **Why selected.** Standalone result validity cannot prove ownership by this request.
10. **Rejected alternatives.** Trust-only composition leaves a boundary bypass.
11. **Trade-offs.** Custom composers must satisfy stricter output semantics.
12. **Regression tests / protection.** Composer output postcondition/identity tests and strict audit.
13. **Adversarial review findings.** Production review retains its own deeper evidence-lineage postconditions; Read adds the outer request binding.
14. **Remediation iterations.** Closed in contract-hardening increment.
15. **Residual risks / limitations.** Deep production lineage remains owned by Projection Production rather than duplicated here.
16. **Operational/deployment consequences.** Foreign output fails closed.
17. **Exact evidence.** `4eeff2b9f5b5c17dd6b7ebe5d0be4a7bd836fb37`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Read audit protects Composer postconditions and request binding.

### GFA-DATA-427 — Persisted Route row metadata and JSON payload could disagree before Projection use

1. **Finding / symptom.** Route JSON could be decoded without immediately proving agreement with denormalized persisted row metadata and the Route contract.
2. **Root cause.** Persistence representation and domain validation were separated too late in the read path.
3. **Failure scenario.** Row trajectory/identity/time fields disagree with the JSON payload yet the route reaches Production evidence assembly.
4. **Impact.** Projection may authorize against internally contradictory route evidence.
5. **Severity rationale.** P1 retrospective because Route evidence affects historical strategy and Estimated Arrival.
6. **Existing guarantees violated.** Persistence/domain mirror integrity and early validation.
7. **Considered solutions.** Trust JSON; trust row metadata; bind both and run Route Contract validation immediately.
8. **Chosen remediation.** Load route JSON, reconcile payload against persisted row mirrors, and validate the Route contract before composition.
9. **Why selected.** Neither persistence representation is allowed to silently override the other.
10. **Rejected alternatives.** One-sided trust retains contradiction risk.
11. **Trade-offs.** Corrupt legacy rows fail rather than being tolerated.
12. **Regression tests / protection.** Route row/payload mismatch and early-contract validation tests.
13. **Adversarial review findings.** This is distinct from Production's workflow-specific Route binding; Read protects persistence integrity first.
14. **Remediation iterations.** Closed in contract-hardening increment.
15. **Residual risks / limitations.** Database constraints cannot encode every nested JSON domain invariant, so application validation remains necessary.
16. **Operational/deployment consequences.** Contradictory route records are rejected.
17. **Exact evidence.** `4eeff2b9f5b5c17dd6b7ebe5d0be4a7bd836fb37`, run `30638188394`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit protects row/payload binding and early Route validation.

### GFA-DATA-428 — Projection `GeneratedAt` could precede completion of snapshot acquisition

1. **Finding / symptom.** Generation time could be assigned before the authoritative data snapshot was fully acquired.
2. **Root cause.** Result-generation metadata was initialized at request start instead of after input acquisition.
3. **Failure scenario.** Slow snapshot loading yields a projection whose `GeneratedAt` predates data it claims to have processed.
4. **Impact.** Provenance chronology is internally inconsistent.
5. **Severity rationale.** P1 retrospective because timestamp chronology is part of result provenance.
6. **Existing guarantees violated.** Temporal provenance and generation semantics.
7. **Considered solutions.** Request-start timestamp; snapshot-start timestamp; assign after successful snapshot acquisition.
8. **Chosen remediation.** Set `GeneratedAt` only after snapshot acquisition.
9. **Why selected.** The generation process cannot predate receipt of its complete input snapshot.
10. **Rejected alternatives.** Request-start time measures latency origin, not result generation provenance.
11. **Trade-offs.** `GeneratedAt` no longer doubles as request-received time.
12. **Regression tests / protection.** GeneratedAt chronology tests.
13. **Adversarial review findings.** If request timing is needed operationally, it should be separate observability metadata.
14. **Remediation iterations.** Closed in contract hardening.
15. **Residual risks / limitations.** Clock correctness still depends on the host clock.
16. **Operational/deployment consequences.** Result metadata becomes chronologically correct.
17. **Exact evidence.** `4eeff2b9f5b5c17dd6b7ebe5d0be4a7bd836fb37`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects assignment ordering.

### GFA-DATA-429 — Stale trajectory `UpdatedAt` could exceed the authorized projection as-of boundary

1. **Finding / symptom.** Snapshot trajectory metadata could carry an `UpdatedAt` value inconsistent with the historical `AsOfTime` used for Projection.
2. **Root cause.** Current persistence metadata leaked into an otherwise historical snapshot representation.
3. **Failure scenario.** Historical projection at T publishes trajectory update provenance from after T.
4. **Impact.** Replay provenance contains future metadata and can invalidate temporal interpretation.
5. **Severity rationale.** P1 retrospective because this is a future-evidence leak in a historical boundary.
6. **Existing guarantees violated.** As-of correctness and provenance chronology.
7. **Considered solutions.** Preserve database `UpdatedAt`; discard field; repair/canonicalize it to the snapshot boundary when stale.
8. **Chosen remediation.** Repair stale `UpdatedAt` so the snapshot representation does not claim post-as-of provenance.
9. **Why selected.** It keeps required metadata while enforcing the historical boundary.
10. **Rejected alternatives.** Preserving future value leaks knowledge; dropping required metadata loses provenance entirely.
11. **Trade-offs.** Repaired metadata reflects snapshot semantics rather than literal latest row-update time.
12. **Regression tests / protection.** UpdatedAt/as-of repair tests.
13. **Adversarial review findings.** The repair is explicit normalization, not silent fabrication of source observation time.
14. **Remediation iterations.** Closed in contract-hardening increment.
15. **Residual risks / limitations.** Exact historical row update chronology would require separately versioned persistence history.
16. **Operational/deployment consequences.** Historical snapshots expose cutoff-safe metadata.
17. **Exact evidence.** `4eeff2b9f5b5c17dd6b7ebe5d0be4a7bd836fb37`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Read audit protects the as-of repair marker and tests.

### GFA-CONTRACT-430 — Zero requested duration was not canonicalized consistently before Production composition

1. **Finding / symptom.** Zero duration sentinel semantics could be interpreted differently between the read boundary and downstream Projection policy.
2. **Root cause.** Default-duration normalization did not have one owner before composition.
3. **Failure scenario.** The same public zero-duration request produces different plan identity depending on downstream path/default handling.
4. **Impact.** Request identity and forecast horizon semantics can drift.
5. **Severity rationale.** P2 retrospective because this is a contract-boundary inconsistency rather than fabricated telemetry.
6. **Existing guarantees violated.** Canonical request normalization.
7. **Considered solutions.** Let Production own sentinel; reject zero; normalize at Read before Production.
8. **Chosen remediation.** Canonicalize zero-duration sentinel at Projection Read using the documented default boundary before composition.
9. **Why selected.** The public read request has one deterministic meaning before entering deeper orchestration.
10. **Rejected alternatives.** Multi-layer defaulting creates identity drift; rejecting zero breaks the documented sentinel contract.
11. **Trade-offs.** Read becomes the explicit owner of this public normalization.
12. **Regression tests / protection.** Default-duration boundary tests.
13. **Adversarial review findings.** This is not a generic criticism of optional/default fields; it is an observed cross-boundary sentinel mismatch.
14. **Remediation iterations.** Closed in contract hardening.
15. **Residual risks / limitations.** Any future default policy change must update request identity/versioning coherently.
16. **Operational/deployment consequences.** Zero-duration requests resolve consistently.
17. **Exact evidence.** `4eeff2b9f5b5c17dd6b7ebe5d0be4a7bd836fb37`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit protects default-duration normalization.

### GFA-DATA-431 — Projection Read source names were not consistently canonicalized

1. **Finding / symptom.** Equivalent source-name text could differ by formatting and produce inconsistent provenance or fingerprint evidence.
2. **Root cause.** Source text normalization was not enforced at the read boundary.
3. **Failure scenario.** Whitespace/case-equivalent source names are treated as distinct provenance identities.
4. **Impact.** Deterministic evidence identity and comparison are weakened.
5. **Severity rationale.** P2 retrospective because source provenance remains present but can fragment semantically equivalent evidence.
6. **Existing guarantees violated.** Canonical provenance identity.
7. **Considered solutions.** Preserve raw text; normalize only during hashing; canonicalize source name before it enters downstream evidence.
8. **Chosen remediation.** Canonicalize source names at Projection Read and use the canonical value downstream.
9. **Why selected.** One normalized representation avoids parallel display/fingerprint semantics.
10. **Rejected alternatives.** Hash-only normalization leaves payload provenance inconsistent.
11. **Trade-offs.** Raw formatting is not preserved as semantic identity.
12. **Regression tests / protection.** Source-name canonicalization tests.
13. **Adversarial review findings.** Canonicalization does not invent a source when none exists.
14. **Remediation iterations.** Closed in contract-hardening increment.
15. **Residual risks / limitations.** Provider renaming remains a real semantic change and should alter provenance.
16. **Operational/deployment consequences.** Cleaner deterministic provenance.
17. **Exact evidence.** `4eeff2b9f5b5c17dd6b7ebe5d0be4a7bd836fb37`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects canonical source handling.

### GFA-DATA-432 — Rejected historical candidate identifiers were lost from the Projection snapshot

1. **Finding / symptom.** Candidate hydration/validation could reject records without preserving their identifiers in the resulting evidence summary.
2. **Root cause.** The read path focused on accepted candidate payloads and did not backfill rejected-ID evidence.
3. **Failure scenario.** Two snapshots with different rejected historical records expose the same accepted candidate set and appear equivalent.
4. **Impact.** Coverage/limitation provenance hides part of the candidate evidence boundary.
5. **Severity rationale.** P1 retrospective because omitted rejected evidence affects reproducibility of neighbor availability.
6. **Existing guarantees violated.** Complete evidence accounting and snapshot identity.
7. **Considered solutions.** Ignore rejected IDs; publish only rejection count; retain exact rejected identifiers.
8. **Chosen remediation.** Backfill rejected historical candidate IDs into snapshot evidence.
9. **Why selected.** Exact identifiers make exclusion evidence reproducible without retaining invalid payloads.
10. **Rejected alternatives.** Counts alone cannot distinguish different rejected sets.
11. **Trade-offs.** Snapshot metadata grows with the bounded rejected set.
12. **Regression tests / protection.** Historical candidate backfill tests.
13. **Adversarial review findings.** Invalid payload content is still not admitted as a usable candidate.
14. **Remediation iterations.** Closed in evidence-hardening increment.
15. **Residual risks / limitations.** Identifier quality depends on the persistence record carrying a valid stable ID even when payload hydration fails.
16. **Operational/deployment consequences.** Better exclusion provenance only.
17. **Exact evidence.** `9dda4b102497028b59280143b86bf84564afb136`, run `30648605652`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit protects rejected-candidate backfill.

### GFA-DATA-433 — Historical candidate limit could be consumed before hydration determined which candidates were usable

1. **Finding / symptom.** A raw-record limit could be exhausted by rejected/unhydratable candidates before enough valid candidates were assembled.
2. **Root cause.** Bounding occurred before semantic hydration/acceptance rather than on the accepted candidate set.
3. **Failure scenario.** Early malformed candidates starve later valid candidates from Neighbor Selection input.
4. **Impact.** Available historical evidence depends on invalid record ordering and coverage is understated.
5. **Severity rationale.** P1 retrospective because candidate availability and downstream strategy selection can change.
6. **Existing guarantees violated.** Deterministic bounded evidence selection and fair accepted-candidate budgeting.
7. **Considered solutions.** Limit raw rows; overfetch arbitrary factor; apply accepted-candidate limit after hydration while keeping source query bounded.
8. **Chosen remediation.** Enforce the candidate acceptance limit after hydration/validation and preserve rejected evidence separately.
9. **Why selected.** Invalid records no longer consume the semantic candidate budget.
10. **Rejected alternatives.** Raw limiting makes results sensitive to bad-record placement; arbitrary overfetch factors are not semantic guarantees.
11. **Trade-offs.** Read path may inspect more bounded source rows to assemble the accepted set.
12. **Regression tests / protection.** Accepted-candidate-limit-after-hydration tests.
13. **Adversarial review findings.** This is distinct from the already-existing exclusion of current trajectory/current flight from history.
14. **Remediation iterations.** Closed in evidence-hardening increment.
15. **Residual risks / limitations.** Source-level hard bounds still cap total work and may limit coverage under extreme corruption.
16. **Operational/deployment consequences.** More stable candidate availability under mixed valid/invalid records.
17. **Exact evidence.** `9dda4b102497028b59280143b86bf84564afb136`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Audit protects post-hydration accepted budgeting.

### GFA-DATA-434 — Route-history fingerprint was not bound to the exact contributing route records and their input fingerprints

1. **Finding / symptom.** Route-history identity could summarize aggregates without proving the exact stored route records/input fingerprints that contributed.
2. **Root cause.** Aggregate history evidence did not carry one canonical record-level lineage representation.
3. **Failure scenario.** Different sets of route records produce equal counts/latest timestamps and collide in history identity.
4. **Impact.** Route Frequency can be authorized by an aggregate whose underlying evidence cannot be independently reconstructed.
5. **Severity rationale.** P1 retrospective because Route History is a production authorization input.
6. **Existing guarantees violated.** Exact contributing-evidence lineage and deterministic fingerprinting.
7. **Considered solutions.** Hash aggregates only; hash record IDs only; hash canonical contributing record identity plus each route input fingerprint.
8. **Chosen remediation.** Bind history to exact route records and their input fingerprints using a deterministic evidence fingerprint.
9. **Why selected.** It identifies the records and semantic route inputs, not only their aggregate outputs.
10. **Rejected alternatives.** Aggregate-only identity collides; record ID alone misses semantic payload changes.
11. **Trade-offs.** Fingerprint computation processes each bounded contributing route record.
12. **Regression tests / protection.** Route-history record/input fingerprint lineage and deterministic evidence fingerprint tests.
13. **Adversarial review findings.** Current target exclusion was already present and is not duplicated as this finding.
14. **Remediation iterations.** Closed in evidence-hardening increment.
15. **Residual risks / limitations.** Correctness depends on each route input fingerprint already being canonical.
16. **Operational/deployment consequences.** Route-history identities become more precise and may change from prior outputs.
17. **Exact evidence.** `9dda4b102497028b59280143b86bf84564afb136`, run `30648605652`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Read review audit protects record-level lineage and deterministic fingerprint semantics.

### GFA-DATA-435 — Route-history aggregate mirrors were not independently reconciled with contributing evidence

1. **Finding / symptom.** Published route-history counts/timestamps/summary fields could remain internally plausible without matching the exact contributing record set.
2. **Root cause.** Aggregate mirrors were trusted after construction rather than reconstructed during validation.
3. **Failure scenario.** A caller mutates a count/latest-time field while retaining a valid evidence fingerprint/list.
4. **Impact.** Downstream Route Frequency can consume contradictory support evidence.
5. **Severity rationale.** P1 retrospective because these aggregates directly affect production route-frequency decisions.
6. **Existing guarantees violated.** Derived-value integrity and cross-field consistency.
7. **Considered solutions.** Range checks; trust builder; recompute aggregate mirrors from canonical route-history evidence.
8. **Chosen remediation.** Validate aggregate mirrors against the exact contributing route records and deterministic evidence fingerprint.
9. **Why selected.** All deterministic derived evidence becomes independently checkable.
10. **Rejected alternatives.** Range checks cannot detect coordinated plausible drift.
11. **Trade-offs.** Validation performs bounded recomputation.
12. **Regression tests / protection.** Route-history aggregate mirror mutation tests.
13. **Adversarial review findings.** The evidence fingerprint and aggregate mirrors protect complementary aspects: identity and derived arithmetic.
14. **Remediation iterations.** Closed in evidence-hardening increment.
15. **Residual risks / limitations.** New aggregate fields must be added to reconstruction when introduced.
16. **Operational/deployment consequences.** Contradictory history summaries fail closed.
17. **Exact evidence.** `9dda4b102497028b59280143b86bf84564afb136`; permanent audit `e0557f6bc3115767ba124a9c94cbb008194c643b`, run `30651385019`.
18. **Final canonical status.** CLOSED.
19. **Prevention / future guard.** Permanent audit protects mirror validation and its regressions.
