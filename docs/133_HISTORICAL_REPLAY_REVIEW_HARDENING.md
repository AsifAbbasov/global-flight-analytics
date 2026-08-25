# Historical Replay Review Hardening

Status: closed

## Scope

This increment hardens
`apps/api/internal/historicalintelligence/historicalreplay` and the replay branch
of `apps/api/cmd/materialize-historical-intelligence`.

## Accepted findings

- Replay accepted a successful Materializer call without validating the returned
  Materialization version, plans, period summaries, historical results, record
  key, record identifier, persisted fingerprint, storage time, or comparison.
- Replay results did not independently preserve complete, partial, or failed
  execution state after separation from the returned Go error.
- The production command discarded a successfully persisted completed prefix
  whenever a later replay window failed.
- Separate per-window Materialization transactions could observe a shared period
  differently between adjacent calls even though the replay used one analytical
  as-of time.
- Global metric, scope, dataset-limit, generation-time, and replay-limit failures
  were not rejected before entering the window loop.
- Replay planning could use the larger Historical Window allocation limit and
  only afterwards apply the smaller replay-window limit.
- Nil context was replaced with a new background context, losing caller
  cancellation ownership.
- Replay had no canonical input fingerprint or public result validation contract.
- Regression coverage did not protect malformed Materialization outcomes,
  self-contained partial progress, continuity, early validation, bounded
  planning, clone isolation, result tampering, or production partial reporting.

## Corrected contracts

- Historical Replay advances to `historical-replay-v2`.
- Every successful Materializer call is validated before its record can enter the
  replay result. Validation covers the Materialization version, canonical
  one-bucket current and previous plans, read-summary windows, dataset limits,
  snapshot isolation label, all three Historical Results, required comparison,
  record identity, record payload, fingerprints, and storage causality.
- The replay `Result` records `complete`, `partial`, or `failed` status, planned
  and completed window counts, structured failure information, generated,
  started, and completed times, and a deterministic replay input fingerprint.
- `Result.Validate` reconstructs the replay fingerprint and verifies the plan,
  limits, status/count relationship, completed-prefix ordering, record contracts,
  failure location, and adjacent-period continuity.
- Global replay request validation occurs before Materialization. Metric and
  scope permission use the production `MetricSpecFor` catalog; dataset and both
  count limits are normalized once; generation time must be at or after the
  analytical as-of time and not after replay execution starts.
- Planning uses the lower of the normalized Historical Window bucket limit and
  replay window limit. A replay-window limit is reported as
  `WindowCountExceededError`; a lower bucket-allocation limit remains the
  Historical Window error. Every one-window Materialization request receives
  `MaximumBucketCount=1`.
- Adjacent Materializations must agree on the raw input fingerprint for their
  shared period. A mismatch stops the replay with an explicit partial result
  instead of silently persisting an internally inconsistent chain.
- The command operation now returns a non-empty partial report together with the
  replay error. The executable writes that JSON report to standard output and
  still returns a non-zero exit code and the failure on standard error.
- Replay and command operation reject nil context.

## Findings already resolved before this review

- Historical Materialization version two already removed the combined bounded
  read. Every Materialization call reads previous and current periods with
  independent limits inside one managed PostgreSQL repeatable-read transaction.
- Historical Contract version two already validates coverage, comparison-current
  value, confidence evidence, provenance, finite values, and metric-specific
  count, ratio, and continuous-number semantics.
- Historical Aggregate already enforces deterministic identity, canonical
  payload idempotency, full row-versus-JSON consistency, and storage-time
  causality.
- `WindowError.Unwrap` already returned its wrapped error. No correction was
  required.

## Qualified or rejected findings

- A replay-wide PostgreSQL transaction is deliberately not introduced. A long
  transaction across as many as ten thousand durable writes conflicts with the
  completed-prefix recovery model, increases database retention and lock costs,
  and would make partial persistence ambiguous. Shared-period fingerprint
  continuity detects the actual cross-call consistency risk without changing
  commit semantics.
- `MaximumBucketCount` and `MaximumWindowCount` remain separate because they are
  different operator controls: one bounds Historical Window planning and one
  bounds replay work. Their interaction is now deterministic and allocation is
  bounded by the lower value.
- A checkpoint or resume token is a future operational feature, not a missing
  integrity invariant. Aggregate writes are already idempotent, and a repeated
  replay remains safe. Resume semantics require a separate versioned API and are
  not mixed into this review correction.
- No arbitrary future-time tolerance is added. The causal rule is exact:
  `GeneratedAt` cannot precede `AsOfTime` and cannot exceed replay `StartedAt`.
- Returning `nil, error` from a Go constructor remains idiomatic and is not a
  domain null result.
- Function length is a review signal rather than an independent defect. Runner
  responsibilities are decomposed into request normalization, planning,
  Materialization outcome validation, fingerprinting, result validation, and
  execution control.
- Historical metric `float64` transport and comparison confidence semantics are
  retained closed Historical Contract and Historical Comparison decisions; they
  are not Replay defects.

```text
MATERIALIZATION_OUTCOME_VALIDATION=ENFORCED
REPLAY_RESULT_STATUS=EXPLICIT
PARTIAL_PROGRESS_REPORTING=PRESERVED
OVERLAPPING_PERIOD_CONTINUITY=VERIFIED
REPLAY_REQUEST_VALIDATION=EARLY
REPLAY_PLANNING_LIMITS=BOUNDED
NIL_CONTEXT_REJECTED=YES
REPLAY_INPUT_FINGERPRINT=BOUND
REPLAY_WIDE_TRANSACTION=REJECTED_BY_DESIGN
CHECKPOINT_RESUME_TOKEN=SEPARATE_PRODUCT_FEATURE
HISTORICAL_REPLAY_ENGINEERING_REMEDIATION=IMPLEMENTED
```

## Permanent verification

`apps/api/tools/historicalreplayreviewaudit` protects the replay version,
self-contained status and failure model, early request validation, bounded plan,
strict Materialization outcome validation, cross-call shared-period continuity,
canonical replay fingerprint, production completed-prefix reporting, nil-context
rejection, regression tests, and this review record in Backend Continuous
Integration.

## Formal closure evidence

The Historical Replay engineering remediation was committed and validated before
this administrative closure:

```text
ENGINEERING_BASELINE_COMMIT=d73c27b5e54108c7d2b9a009cb157496f7c67bde
ENGINEERING_REMEDIATION_COMMIT=38b14fbb8649a2e7e875cd4ae7ed73b6a954a068
ENGINEERING_GITHUB_ACTIONS_RUN=30390451707
Backend Quality=SUCCESS
Backend Quality Job=90380396908
PostgreSQL 16 Integration=SUCCESS
PostgreSQL 16 Integration Job=90380396909
Backend Race Safety=SUCCESS
Backend Race Safety Job=90380396961
Backend Container=SUCCESS
Backend Container Job=90380713650
```

All accepted findings are implemented. Findings already resolved before this
review and qualified or rejected findings retain their documented rationale.
No Historical Replay review item remains open, unclassified, or deferred.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
HISTORICAL_REPLAY_ENGINEERING_DEBT=CLOSED
HISTORICAL_REPLAY_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
HISTORICAL_REPLAY_REVIEW_STATUS=CLOSED
```

## Canonical remediation history

The records below reconstruct finding-level ownership from repository evidence.
Severity labels are retrospective engineering classifications rather than claimed
historical reviewer labels. The implementation owner is
`38b14fbb8649a2e7e875cd4ae7ed73b6a954a068`; exact historical Backend CI run
`30390451707` completed successfully.

### GFA-DATA-297 — Replay accepted unverified Materialization outcomes

1. **Finding / symptom:** a successful Materializer call could enter Replay without validating Materialization version, current/previous plans, read summaries, Historical Results, record key/identifier, persisted fingerprint, storage time, or comparison.
2. **Root cause:** Replay equated a nil lower-layer error with a trustworthy versioned outcome.
3. **Failure scenario:** a malformed or stale Materializer returns success with internally inconsistent outcome fields and Replay persists/returns the record as valid replay evidence.
4. **Impact:** durable replay chains can contain records that violate Materialization, Historical Contract, or Aggregate identity rules.
5. **Severity rationale:** **P1 retrospective** because unverified lower-layer output could become accepted historical replay evidence.
6. **Existing guarantees violated:** versioned boundary validation, durable record identity, comparison integrity, storage causality.
7. **Considered solutions:** trust Materializer; validate only final record; validate the complete outcome contract before acceptance.
8. **Chosen remediation:** `validateOutcome` verifies Materialization version, canonical one-bucket plans, period summaries, limits/isolation, all Historical Results, comparison, record identity/payload/fingerprints, and storage causality.
9. **Why selected:** Replay is a separate trust boundary and must reject malformed successful outputs before they enter its result chain.
10. **Rejected alternatives:** final-record-only validation cannot prove plan/read-summary or previous/current Materialization structure.
11. **Trade-offs:** Replay intentionally duplicates validation of boundary relationships that are observable only at the orchestration level.
12. **Regression tests / protection:** `TestRunRejectsInvalidMaterializationOutcome`; `historicalreplayreviewaudit` checks the outcome-validation boundary.
13. **Adversarial review findings:** already-hardened Materialization does not remove the need to validate injected/custom implementations at the Replay interface.
14. **Remediation iterations:** `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068`.
15. **Residual risks / limitations:** validation proves declared inter-object consistency, not external truth of upstream aviation data.
16. **Operational/deployment consequences:** malformed successful outcomes now stop Replay with explicit failure evidence.
17. **Exact evidence:** remediation commit, outcome-validation tests, exact Backend CI run `30390451707` SUCCESS, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every versioned orchestration boundary must validate complete returned identity/evidence before promoting it into a durable higher-level workflow.

### GFA-DATA-298 — Replay result status depended on a separate Go error instead of self-contained execution evidence

1. **Finding / symptom:** Replay results did not independently preserve whether execution was complete, partial, or failed once separated from the returned Go error.
2. **Root cause:** execution status lived in control flow rather than the replay result contract.
3. **Failure scenario:** a persisted or transported result is inspected without the original error and appears ambiguous about whether all planned windows completed.
4. **Impact:** operators and downstream consumers can misinterpret partial historical replay as complete.
5. **Severity rationale:** **P1 retrospective** because loss of failure/partial state changes the meaning of the replay evidence itself.
6. **Existing guarantees violated:** self-describing durable evidence, partial-progress honesty, deterministic result validation.
7. **Considered solutions:** rely on error pairing; add a boolean failure flag; define complete/partial/failed status with counts and structured failure.
8. **Chosen remediation:** Replay v2 records status, planned/completed counts, failure details, generated/started/completed times, and validates their relationships.
9. **Why selected:** the result remains interpretable after serialization or separation from transient control-flow errors.
10. **Rejected alternatives:** one boolean cannot distinguish complete, partial, and failed-with-zero-prefix semantics safely.
11. **Trade-offs:** a larger versioned result contract and stricter validation.
12. **Regression tests / protection:** complete, partial, failed, clone and tampering tests; permanent audit.
13. **Adversarial review findings:** status must reconcile with counts, prefix ordering and failure location, not merely exist as an enum field.
14. **Remediation iterations:** `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068`.
15. **Residual risks / limitations:** the result records the first terminal replay failure, not an exhaustive list of hypothetical later failures.
16. **Operational/deployment consequences:** serialized replay reports become self-contained and machine-classifiable.
17. **Exact evidence:** Replay v2 contracts/validation, remediation commit, run `30390451707`, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** batch/replay results must encode completion state and progress independently of transient language-level errors.

### GFA-OPS-299 — Production replay reporting discarded a successfully persisted completed prefix

1. **Finding / symptom:** when a later replay window failed, the production command returned the error but discarded already persisted successful windows from its report.
2. **Root cause:** command execution treated error and report as mutually exclusive outcomes.
3. **Failure scenario:** windows 1–N persist successfully, window N+1 fails, and the operator receives no structured evidence of the durable prefix.
4. **Impact:** recovery decisions become ambiguous and can cause unnecessary replay work or incorrect assumptions about what was stored.
5. **Severity rationale:** **P1 retrospective** because operational reporting could contradict durable persistence state after partial success.
6. **Existing guarantees violated:** completed-prefix recovery model, operator observability, persistence/report consistency.
7. **Considered solutions:** roll back all replay writes; return only error; return a partial report together with the error.
8. **Chosen remediation:** command operation preserves a non-empty replay report on failure; stdout receives JSON prefix/status while stderr and exit code still expose failure.
9. **Why selected:** Aggregate writes are intentionally durable/idempotent per window, so reporting must reflect committed progress rather than pretend atomic rollback.
10. **Rejected alternatives:** replay-wide rollback conflicts with durable-prefix recovery and long-running operational constraints.
11. **Trade-offs:** callers must handle simultaneous useful report data and a non-nil error.
12. **Regression tests / protection:** `TestCommandOperationPreservesPartialReplayReport`, `TestWriteCommandOutcomeEmitsPrefixBeforeFailure`, permanent audit.
13. **Adversarial review findings:** non-zero exit status is retained; preserving partial output must not convert failure into apparent success.
14. **Remediation iterations:** `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068`.
15. **Residual risks / limitations:** no checkpoint/resume token is introduced; reruns rely on separately hardened idempotent Aggregate writes.
16. **Operational/deployment consequences:** automation can identify the completed durable prefix while still treating the command as failed.
17. **Exact evidence:** remediation commit, production command tests, run `30390451707`, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** long-running workflows with durable incremental commits must report committed progress even when the overall operation terminates with an error.

### GFA-DATA-300 — Adjacent Replay Materializations could observe inconsistent shared-period evidence

1. **Finding / symptom:** each replay window used its own Materialization transaction, so the period shared by adjacent comparisons could differ across calls despite one analytical `AsOfTime`.
2. **Root cause:** per-call snapshot correctness did not establish continuity across independently committed replay windows.
3. **Failure scenario:** source state changes between calls; period T is the current side of one comparison and the previous side of the next with different raw evidence.
4. **Impact:** the replay chain becomes internally inconsistent even though each individual Materialization is locally valid.
5. **Severity rationale:** **P1 retrospective** because adjacent historical comparisons can encode contradictory versions of the same analytical period.
6. **Existing guarantees violated:** replay-chain continuity, shared-period identity, deterministic analytical history.
7. **Considered solutions:** one replay-wide PostgreSQL transaction; persist source snapshots; compare shared-period raw input fingerprints across calls.
8. **Chosen remediation:** adjacent Materializations must agree on the raw input fingerprint for the shared period; mismatch stops replay with explicit partial status.
9. **Why selected:** it detects the actual cross-call integrity risk without a transaction spanning potentially thousands of durable writes.
10. **Rejected alternatives:** replay-wide transaction would undermine completed-prefix commit semantics and increase retention/lock costs.
11. **Trade-offs:** Replay can terminate partially when underlying evidence changes during a long run.
12. **Regression tests / protection:** `TestRunRejectsOverlappingReadContinuityMismatch`; result validation and permanent audit.
13. **Adversarial review findings:** same `AsOfTime` alone is not evidence identity; continuity requires a deterministic fingerprint of the shared raw period.
14. **Remediation iterations:** `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068`.
15. **Residual risks / limitations:** continuity detects mismatches but deliberately does not freeze the entire database for replay duration.
16. **Operational/deployment consequences:** long replays may fail safely with a preserved prefix if shared source evidence changes.
17. **Exact evidence:** remediation commit, continuity test, run `30390451707`, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** multi-transaction replay chains that reuse analytical periods must verify deterministic shared-period identity between adjacent operations.

### GFA-CONTRACT-301 — Global Replay request failures were validated after window execution began

1. **Finding / symptom:** invalid metric, scope, dataset limit, generation time, or replay limits were not all rejected before entering the replay loop.
2. **Root cause:** request normalization/validation was distributed between planning and per-window execution.
3. **Failure scenario:** Replay performs planning or even lower-layer work before discovering a global request defect that could have been rejected once at entry.
4. **Impact:** unnecessary work, inconsistent error location, and risk of partial side effects before recognizing invalid global intent.
5. **Severity rationale:** **P2 retrospective** because the main risk is contract/lifecycle correctness and avoidable work; the corrected path fails before durable window processing.
6. **Existing guarantees violated:** fail-fast request validation, single normalization ownership, predictable orchestration errors.
7. **Considered solutions:** keep lazy validation; validate only structural times; normalize all global controls before planning/materialization.
8. **Chosen remediation:** validate metric/scope through `MetricSpecFor`, normalize dataset and count limits once, and enforce generation-time causality before Materialization.
9. **Why selected:** global invalidity should be decided once before any repeated work begins.
10. **Rejected alternatives:** per-window validation repeats policy and permits later failure after avoidable processing.
11. **Trade-offs:** Replay request code owns a more explicit normalization phase.
12. **Regression tests / protection:** `TestRunValidatesGlobalRequestBeforeReplay`; request audit requirements.
13. **Adversarial review findings:** this does not duplicate the Historical Contract catalog; it consumes `MetricSpecFor` as the single production policy owner.
14. **Remediation iterations:** `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068`.
15. **Residual risks / limitations:** per-window failures remain possible for data/persistence conditions that cannot be known globally.
16. **Operational/deployment consequences:** invalid replay requests fail before materialization writes begin.
17. **Exact evidence:** remediation commit, request tests, run `30390451707`, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** global batch/replay parameters must be normalized and rejected before iterative side-effecting work begins.

### GFA-PERF-302 — Replay planning applied the smaller replay limit after larger allocation work

1. **Finding / symptom:** Historical Window planning could allocate using a larger bucket limit before Replay applied the smaller `MaximumWindowCount`.
2. **Root cause:** two valid operator limits were enforced at different stages rather than combined at allocation boundary.
3. **Failure scenario:** a request with a very large bucket allowance but a small replay limit allocates/plans more windows than Replay is permitted to execute.
4. **Impact:** avoidable memory/CPU work and weaker bounded-work guarantees.
5. **Severity rationale:** **P2 retrospective** because the defect can defeat intended resource bounds without corrupting valid analytical values.
6. **Existing guarantees violated:** bounded planning, free-tier resource discipline, deterministic operator controls.
7. **Considered solutions:** merge both limits into one field; truncate after planning; plan using the lower normalized limit.
8. **Chosen remediation:** `PlanningMaximumBucketCount` uses the lower applicable limit; each one-window Materialization receives `MaximumBucketCount=1`.
9. **Why selected:** it preserves distinct semantics of bucket and replay-window controls while bounding allocation before growth.
10. **Rejected alternatives:** merging controls erases different operator meanings; post-plan truncation retains excessive allocation.
11. **Trade-offs:** planning logic must classify which limit caused rejection so Historical Window and Replay errors remain distinct.
12. **Regression tests / protection:** `TestRunUsesBoundedPlanningLimits`; permanent audit checks planning limits.
13. **Adversarial review findings:** separate limits are not themselves a defect; only their previous enforcement order was.
14. **Remediation iterations:** `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068`.
15. **Residual risks / limitations:** work within the configured maximum remains potentially substantial by operator choice.
16. **Operational/deployment consequences:** replay planning respects the tighter operator bound before allocation.
17. **Exact evidence:** remediation commit, bounded-planning tests, run `30390451707`, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** when multiple independent limits constrain the same allocation, allocation must use the tightest effective bound while preserving error ownership.

### GFA-OPS-303 — Replay and production command silently replaced nil caller context

1. **Finding / symptom:** nil context was replaced with a background context in Replay/command execution.
2. **Root cause:** fallback lifecycle ownership was invented inside a long-running orchestration layer.
3. **Failure scenario:** an invalid caller loses cancellation/deadline semantics and starts a potentially large replay under an unbounded background lifetime.
4. **Impact:** runaway work, database/provider resource consumption, and misleading cancellation behavior.
5. **Severity rationale:** **P2 retrospective** because it is an operational lifecycle defect on invalid API use, not a normal-context data corruption path.
6. **Existing guarantees violated:** explicit context ownership, bounded work, cancellation propagation.
7. **Considered solutions:** background fallback; internal timeout; reject nil.
8. **Chosen remediation:** Replay and command operation reject nil context; cancellation remains caller-owned.
9. **Why selected:** only the caller can define the correct lifetime for a potentially long replay.
10. **Rejected alternatives:** arbitrary internal lifetime substitution hides programming errors and can outlive the initiating request/job.
11. **Trade-offs:** callers must supply a valid context explicitly.
12. **Regression tests / protection:** `TestRunRejectsNilAndCanceledContext`, `TestCommandOperationRejectsNilContext`, audit forbids background substitution.
13. **Adversarial review findings:** canceled non-nil contexts are also protected, including partial cancellation after completed windows.
14. **Remediation iterations:** `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068`.
15. **Residual risks / limitations:** callers may intentionally provide long/no-deadline contexts; that remains their policy responsibility.
16. **Operational/deployment consequences:** nil-context misuse fails before replay execution.
17. **Exact evidence:** remediation commit, context regression tests, run `30390451707`, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** long-running orchestration entrypoints must never synthesize caller lifecycle from nil context.

### GFA-DATA-304 — Replay lacked canonical input identity and public result validation

1. **Finding / symptom:** Replay had no canonical input fingerprint and no public contract that could revalidate a Replay result after construction/transport.
2. **Root cause:** determinism and integrity were enforced procedurally inside execution rather than encoded in the result boundary.
3. **Failure scenario:** a result is mutated, reconstructed, or serialized with changed plan/limits/windows/failure data and no canonical validator can prove tampering or input identity drift.
4. **Impact:** replay provenance and result integrity cannot be independently verified.
5. **Severity rationale:** **P1 retrospective** because deterministic identity and tamper detection are core historical auditability guarantees.
6. **Existing guarantees violated:** deterministic replay identity, clone/transport integrity, self-validating versioned contracts.
7. **Considered solutions:** fingerprint only request fields; rely on per-window record fingerprints; define replay input fingerprint plus `Result.Validate`.
8. **Chosen remediation:** Replay v2 binds deterministic input fingerprint and `Result.Validate` reconstructs fingerprint, plans, limits, statuses/counts, prefix order, record contracts, failure location, and continuity.
9. **Why selected:** Replay has orchestration-level identity that no single materialized record can represent.
10. **Rejected alternatives:** per-record validation cannot prove replay plan/status/failure semantics or global input identity.
11. **Trade-offs:** future contract changes that affect identity require deliberate replay fingerprint/version evolution.
12. **Regression tests / protection:** deterministic/clone tests, `TestResultValidateRejectsTampering`, permanent audit.
13. **Adversarial review findings:** fingerprint must be reconstructed from canonical normalized inputs rather than trusted as opaque caller text.
14. **Remediation iterations:** `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068`.
15. **Residual risks / limitations:** the replay fingerprint proves canonical request identity, while raw data identity remains represented through validated window/materialization fingerprints.
16. **Operational/deployment consequences:** stored or transported Replay results can be independently validated.
17. **Exact evidence:** Replay v2 contracts/validation, remediation commit, run `30390451707`, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** durable analytical orchestration results need canonical input identity and a public validator that can detect post-construction drift.

### GFA-TEST-305 — Replay regression coverage omitted critical partial-progress and integrity invariants

1. **Finding / symptom:** regression tests did not protect malformed Materialization outcomes, self-contained partial progress, continuity, early validation, bounded planning, clone isolation, result tampering, or production partial reporting.
2. **Root cause:** tests emphasized successful window iteration rather than failure and cross-window contract boundaries.
3. **Failure scenario:** later refactors reintroduce loss of partial reports, continuity mismatches, oversized planning, or tamper acceptance without breaking basic replay success tests.
4. **Impact:** multiple P1/P2 Replay remediations could regress silently.
5. **Severity rationale:** **P2 retrospective** because the missing guard weakens durable closure of high-impact replay contracts.
6. **Existing guarantees violated:** CI truth, regression protection, durable remediation evidence.
7. **Considered solutions:** rely on lower-layer tests; add only a few replay tests; install comprehensive targeted tests plus permanent strict audit.
8. **Chosen remediation:** add complete/partial/failed, malformed-outcome, continuity, early-validation, bounded-planning, context, deterministic clone, tamper and command partial-report tests plus `historicalreplayreviewaudit` in Backend CI.
9. **Why selected:** Replay-specific invariants span multiple packages and require a dedicated permanent acceptance boundary.
10. **Rejected alternatives:** lower-layer correctness does not prove higher-level prefix/status/continuity semantics.
11. **Trade-offs:** additional CI runtime and maintenance when Replay contracts deliberately evolve.
12. **Regression tests / protection:** named Replay and command tests plus the strict review audit.
13. **Adversarial review findings:** source audits and behavioral tests are complementary; neither is treated as sufficient alone.
14. **Remediation iterations:** `38b14fbb8649a2e7e875cd4ae7ed73b6a954a068`.
15. **Residual risks / limitations:** test coverage cannot enumerate every future provider/database failure, but contract-level failure classes are permanently exercised.
16. **Operational/deployment consequences:** stronger merge-time assurance; no runtime feature added by the guard itself.
17. **Exact evidence:** remediation commit, `apps/api/tools/historicalreplayreviewaudit`, exact Backend CI run `30390451707` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** replay/batch remediation closure requires explicit tests for partial success, cross-step continuity, tamper resistance, and production reporting behavior.
