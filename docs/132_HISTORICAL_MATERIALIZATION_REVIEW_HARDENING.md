# Historical Materialization Review Hardening

Status: closed

## Scope

This increment hardens
`apps/api/internal/historicalintelligence/historicalmaterialization` and adds the
minimal two-period Historical Read boundary required by that orchestrator.

## Accepted findings

- A single bounded read over the combined previous and current windows allowed
  chronologically earlier rows to consume the complete dataset limit and create
  false current-period declines.
- Materialization did not verify that repository output matched the requested
  query or the current Historical Read version.
- One combined read summary could not identify which period was truncated or
  which period had weaker represented evidence.
- Materialization returned the newly built comparison even when the writer
  returned a canonical persisted result.
- The materialization-specific fingerprint repair excluded `GeneratedAt` and
  duplicated provenance ownership already assigned to Historical Comparison.
- Nil context was replaced with `context.Background`, discarding caller
  cancellation ownership.
- Lower-layer errors did not identify the failed orchestration stage.
- The exported `DatasetLimitOr` helper duplicated already-normalized request
  state.
- Regression coverage did not protect independent limits, two-period atomicity,
  snapshot metadata, persisted outcome identity, stage errors, generated-time
  identity, default and maximum limits, or clone isolation.

## Corrected contracts

- `historicalread.PeriodRepository` reads previous and current queries inside one
  managed repeatable-read transaction and applies an independent dataset and
  route-payload limit to each period.
- `Config.Repository` retains its original source-compatible type. Construction
  fails closed unless the concrete repository also supports
  `historicalread.PeriodRepository`; no duplicate dependency field is exposed.
- The previous and current queries must be adjacent and share one analytical
  as-of time.
- Materialization validates each snapshot version, exact normalized query, and
  shared supported isolation label before invoking a builder.
- Outcome exposes separate previous and current read summaries, including loaded
  rows, matched rows, route payload bytes, and every limit signal. The old
  aggregate `ReadSummary` remains as a deprecated source-compatibility view and
  must not be used for period-sensitive quality decisions.
- Builders receive only their own period snapshot. Their fingerprints no longer
  include unrelated rows from the adjacent period.
- Historical Comparison remains the sole owner of both-period comparison
  provenance and fingerprinting. Its version-two fingerprint already binds
  `GeneratedAt`.
- After persistence, `Record.Result` is the canonical current result returned in
  Outcome. Persistence identity and contract metadata are checked before return.
- Materialization rejects nil context and preserves caller cancellation.
- Typed stage errors identify request validation, planning, read, snapshot
  contract, builder, comparison, persistence, and persistence-contract failures.
- `DatasetLimitOr` is removed and normalized limits are assigned directly.
- Materialization version advances to `historical-materialization-v2`.

## Findings already resolved before this review

- A normal `historicalread.NewPostgres` repository already opens one read-only
  PostgreSQL repeatable-read transaction for every Snapshot. Passing a pool in
  production does not mean the four dataset queries execute outside a snapshot.
- Historical Aggregate already exposes a narrow `Writer`; Materialization no
  longer depends on the complete aggregate Store.
- Historical Comparison version two already requires matching per-bucket
  coverage profiles and records both periods' status, confidence score, sample
  count, and previous-period limitations. Current-series confidence remaining
  contract-consistent is an explicit closed design decision.

## Qualified or rejected findings

- Metric classification and scope permission already come from the single
  `MetricSpecFor` catalog. The remaining three-family builder switch is a finite
  domain dispatch, not three independent registries and not an Open/Closed
  Principle defect.
- Returning `nil, error` from a Go constructor is idiomatic and is not a domain
  null payload.
- Function and file length are review signals rather than independent defects.
  Materialize is decomposed here because period planning, read validation, and
  persistence validation are distinct responsibilities.
- Cross-module scope normalization consolidation is a separate contract
  migration and is not mixed into this period-read integrity correction.
- The combined latest source timestamp is defined as the newest evidence across
  both periods. Period-specific results and read summaries retain the evidence
  needed to inspect each side separately.
- Historical metrics are not monetary calculations; no decimal-library
  migration belongs in this module.

```text
INDEPENDENT_PERIOD_LIMITS=ENFORCED
ATOMIC_TWO_PERIOD_READ=ENFORCED
SNAPSHOT_QUERY_AND_VERSION=VERIFIED
PERIOD_READ_SUMMARIES=EXPLICIT
PERIOD_BUILDER_INPUTS=ISOLATED
HISTORICAL_COMPARISON_PROVENANCE_OWNER=PRESERVED
PERSISTED_RESULT_IS_CANONICAL=YES
GENERATED_AT_FINGERPRINT_IDENTITY=BOUND
NIL_CONTEXT_REJECTED=YES
STAGE_ERRORS=EXPLICIT
DATASET_LIMIT_HELPER=REMOVED
HISTORICAL_MATERIALIZATION_ENGINEERING_REMEDIATION=IMPLEMENTED
```

## Permanent verification

`apps/api/tools/historicalmaterializationreviewaudit` protects the independent
period limits, one-transaction two-period read, snapshot query and version
validation, explicit period summaries, persisted canonical outcome, comparison
provenance ownership, generated-time identity, nil-context rejection, stage
errors, regression tests, and this review record in Backend Continuous
Integration.

## Formal closure evidence

The Historical Materialization engineering remediation was committed and
validated before this administrative closure:

```text
ENGINEERING_BASELINE_COMMIT=f723a54414a1ebb2c09aad89adead714a7446e3a
ENGINEERING_REMEDIATION_COMMIT=2bbbd2439580536ffe17f8827c654c245d9b6b1e
ENGINEERING_GITHUB_ACTIONS_RUN=30384357559
PostgreSQL 16 Integration=SUCCESS
PostgreSQL 16 Integration Job=90359781879
Backend Race Safety=SUCCESS
Backend Race Safety Job=90359781892
Backend Quality=SUCCESS
Backend Quality Job=90359781932
Backend Container=SUCCESS
Backend Container Job=90360092985
```

All accepted findings are implemented. Findings already resolved before this
review and qualified or rejected findings retain their documented rationale.
No Historical Materialization review item remains open, unclassified, or
deferred.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
HISTORICAL_MATERIALIZATION_ENGINEERING_DEBT=CLOSED
HISTORICAL_MATERIALIZATION_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
HISTORICAL_MATERIALIZATION_REVIEW_STATUS=CLOSED
```

## Canonical remediation history

The records below reconstruct finding-level ownership from repository evidence.
Severity labels are retrospective engineering classifications; they are not
represented as original reviewer labels where the historical record does not
prove that chronology. The implementation owner for this review is
`2bbbd2439580536ffe17f8827c654c245d9b6b1e`, with exact historical Backend CI
run `30384357559` completed successfully.

### GFA-DATA-288 — Combined two-period reads allowed cross-period limit starvation

1. **Finding / symptom:** previous and current periods shared one bounded read, so earlier rows could consume the entire dataset limit before current-period evidence was loaded.
2. **Root cause:** period isolation existed in analytical semantics but not in the materialization read budget.
3. **Failure scenario:** a dense previous period fills the shared limit; sparse or missing current rows then appear as a real current decline.
4. **Impact:** Historical Comparison can publish a materially false trend from truncation rather than aviation evidence.
5. **Severity rationale:** **P1 retrospective** because the defect can invert or fabricate a primary historical analytical conclusion.
6. **Existing guarantees violated:** independent period comparability, bounded-read honesty, snapshot-consistent Historical Intelligence.
7. **Considered solutions:** increase the shared limit; execute two unrelated reads; introduce a period-aware repository operation.
8. **Chosen remediation:** `historicalread.PeriodRepository` reads previous and current queries inside one managed repeatable-read transaction with independent dataset and route-payload limits.
9. **Why selected:** it fixes both budget independence and the atomic two-period observation boundary without a long-lived orchestration transaction.
10. **Rejected alternatives:** a larger shared limit only moves the starvation threshold; unrelated transactions permit cross-period snapshot drift.
11. **Trade-offs:** the repository exposes one additional narrow capability and executes both period reads under one managed snapshot.
12. **Regression tests / protection:** `TestPostgresRepositoryReadPeriodsUsesOneManagedSnapshotAndIndependentLimits` and the Historical Materialization review audit.
13. **Adversarial review findings:** a single-period repeatable-read guarantee was insufficient because Materialization compares two periods and therefore needs one shared snapshot boundary.
14. **Remediation iterations:** `2bbbd2439580536ffe17f8827c654c245d9b6b1e`.
15. **Residual risks / limitations:** independent limits still intentionally truncate oversized periods; the truncation is now period-specific and observable.
16. **Operational/deployment consequences:** no migration; production materialization requires a repository implementing `PeriodRepository`.
17. **Exact evidence:** review record, remediation commit, exact Backend CI run `30384357559` SUCCESS, `historicalmaterializationreviewaudit`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** comparison-oriented reads must budget and observe each comparison side independently while sharing the required snapshot boundary.

### GFA-DATA-289 — Materialization trusted unverified Historical Read snapshot identity

1. **Finding / symptom:** Materialization did not prove that repository output matched the requested normalized query or current Historical Read contract version.
2. **Root cause:** repository output was treated as trusted orchestration input instead of a versioned boundary value.
3. **Failure scenario:** a stale, buggy, or test repository returns a snapshot for another period/query/version and Materialization builds a valid-looking result from the wrong evidence.
4. **Impact:** provenance and analytical values can be internally consistent but refer to the wrong requested period.
5. **Severity rationale:** **P1 retrospective** because query/result identity can silently diverge.
6. **Existing guarantees violated:** exact query identity, versioned Historical Read contract, deterministic materialization provenance.
7. **Considered solutions:** trust production repository construction; validate only version; validate complete snapshot metadata before builders run.
8. **Chosen remediation:** verify snapshot version, exact normalized query, and shared supported isolation label for both periods.
9. **Why selected:** boundary validation fails closed before downstream builders can convert bad repository output into durable analytics.
10. **Rejected alternatives:** production-only trust leaves alternate executors and future regressions outside the contract.
11. **Trade-offs:** more explicit validation code and typed mismatch errors.
12. **Regression tests / protection:** `TestMaterializeRejectsSnapshotContractViolations` and permanent audit checks.
13. **Adversarial review findings:** validating the final Historical Result cannot reconstruct whether the source snapshot itself matched the requested repository query.
14. **Remediation iterations:** `2bbbd2439580536ffe17f8827c654c245d9b6b1e`.
15. **Residual risks / limitations:** validation proves declared snapshot identity, not external truthfulness of the underlying source data.
16. **Operational/deployment consequences:** malformed custom repositories fail before builder execution.
17. **Exact evidence:** remediation commit, Historical Materialization tests, run `30384357559`, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** versioned repository outputs consumed by orchestrators must be validated against the exact request that produced them.

### GFA-DATA-290 — Combined read summary obscured period-specific evidence quality

1. **Finding / symptom:** one aggregate read summary could not show which period was truncated or had weaker represented evidence.
2. **Root cause:** operational read evidence was collapsed after combining two analytically distinct periods.
3. **Failure scenario:** one side reaches a row or byte limit, but downstream quality logic sees only combined counts and cannot attribute the limitation correctly.
4. **Impact:** comparison quality and operator diagnostics can misstate which side is incomplete.
5. **Severity rationale:** **P1 retrospective** because period-specific incompleteness directly affects whether a comparison is trustworthy.
6. **Existing guarantees violated:** evidence provenance, explicit limitation reporting, two-period comparison transparency.
7. **Considered solutions:** add a period flag to aggregate fields; infer period from windows later; expose two summaries.
8. **Chosen remediation:** `PeriodReadSummaries` exposes separate previous/current loaded counts, matched counts, route payload bytes, and limit signals.
9. **Why selected:** explicit per-period evidence is deterministic and avoids reverse inference from aggregated counters.
10. **Rejected alternatives:** a single summary cannot faithfully encode asymmetric truncation.
11. **Trade-offs:** the old aggregate `ReadSummary` remains temporarily as a deprecated compatibility view.
12. **Regression tests / protection:** materialization independent-limit tests and review audit checks for explicit period summaries.
13. **Adversarial review findings:** retaining the compatibility summary is safe only if period-sensitive decisions are prohibited from depending on it.
14. **Remediation iterations:** `2bbbd2439580536ffe17f8827c654c245d9b6b1e`.
15. **Residual risks / limitations:** compatibility consumers can still read the deprecated aggregate summary, but the canonical period-sensitive path is explicit.
16. **Operational/deployment consequences:** outcome payload in Go gains period-specific evidence while source compatibility is retained.
17. **Exact evidence:** remediation commit, `PeriodReadSummaries`, regression tests, run `30384357559`, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** evidence for independently bounded analytical cohorts must remain independently observable through orchestration outputs.

### GFA-DATA-291 — Materialization returned a pre-persistence result instead of canonical persisted identity

1. **Finding / symptom:** after a successful write, Materialization returned the newly built comparison rather than the canonical result returned by the writer.
2. **Root cause:** orchestration treated persistence as a side effect rather than an identity-establishing boundary.
3. **Failure scenario:** a writer canonicalizes or validates persisted identity, but callers receive a different in-memory representation than the durable record.
4. **Impact:** downstream replay/reporting can diverge from the actual stored record.
5. **Severity rationale:** **P1 retrospective** because durable and returned analytical identity can disagree.
6. **Existing guarantees violated:** persisted-result authority, record/result identity, deterministic replayability.
7. **Considered solutions:** return the pre-write result; re-read after write; use the writer-returned canonical record.
8. **Chosen remediation:** validate persistence identity and return `Record.Result` as canonical `CurrentResult` in Outcome.
9. **Why selected:** the write boundary already returns the durable canonical representation, avoiding a redundant read.
10. **Rejected alternatives:** pre-write return ignores storage canonicalization; re-read adds unnecessary database work and a second boundary.
11. **Trade-offs:** writer contracts must return complete canonical records.
12. **Regression tests / protection:** `TestMaterializeUsesPersistedResultAsCanonicalOutcome`, persisted identity mismatch tests, permanent audit.
13. **Adversarial review findings:** successful persistence alone is insufficient; record key, result identity, and fingerprint must be reconciled before return.
14. **Remediation iterations:** `2bbbd2439580536ffe17f8827c654c245d9b6b1e`.
15. **Residual risks / limitations:** correctness still depends on the Aggregate Writer enforcing its own canonical record contract, which is separately audited.
16. **Operational/deployment consequences:** replay and command consumers now observe the same result identity that was durably stored.
17. **Exact evidence:** remediation commit, canonical-outcome tests, run `30384357559`, Historical Aggregate and Materialization audits.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** after durable writes, orchestration outputs must derive identity from the canonical persistence result rather than a pre-write copy.

### GFA-DATA-292 — Materialization duplicated comparison provenance and omitted GeneratedAt from repaired fingerprint identity

1. **Finding / symptom:** a materialization-specific fingerprint repair duplicated Historical Comparison provenance ownership and did not bind `GeneratedAt`.
2. **Root cause:** orchestration attempted to repair a domain identity owned by a lower comparison layer.
3. **Failure scenario:** identical source periods generated at different canonical generation times can collapse to the same repaired materialization fingerprint or drift from direct Comparison behavior.
4. **Impact:** provenance ownership and deterministic identity differ depending on which path produced the comparison.
5. **Severity rationale:** **P1 retrospective** because fingerprint identity and provenance are core auditability guarantees for historical results.
6. **Existing guarantees violated:** single-owner provenance, deterministic input identity, direct/materialized comparison equivalence.
7. **Considered solutions:** extend the materialization repair; add GeneratedAt only; remove the repair and rely on Historical Comparison v2.
8. **Chosen remediation:** preserve Historical Comparison as sole owner of both-period provenance/fingerprinting and remove the materialization-specific repair.
9. **Why selected:** Comparison v2 already binds `GeneratedAt` and complete two-period provenance; one owner prevents semantic drift.
10. **Rejected alternatives:** duplicated repairs require two formulas to remain permanently synchronized.
11. **Trade-offs:** Materialization intentionally delegates this identity rule instead of owning a local fingerprint implementation.
12. **Regression tests / protection:** `TestMaterializeBindsGeneratedAtIntoFingerprint`; audit forbids `materializationFingerprint` and `finalizeComparedResult`.
13. **Adversarial review findings:** orchestration-specific provenance repair was not an independent feature; it was duplicate ownership of an existing domain contract.
14. **Remediation iterations:** `2bbbd2439580536ffe17f8827c654c245d9b6b1e`.
15. **Residual risks / limitations:** changes to Historical Comparison fingerprint semantics require deliberate version review because Materialization now correctly inherits them.
16. **Operational/deployment consequences:** no migration; newly materialized fingerprints follow the canonical Comparison contract.
17. **Exact evidence:** remediation commit, generated-time regression test, run `30384357559`, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** deterministic provenance/fingerprint formulas must have one explicit domain owner; orchestrators compose them rather than reimplement them.

### GFA-OPS-293 — Materialization silently replaced nil caller context

1. **Finding / symptom:** nil context was converted to `context.Background()`.
2. **Root cause:** convenience fallback overrode caller ownership of cancellation and deadlines.
3. **Failure scenario:** an invalid caller loses cancellation semantics and starts repository/build/persistence work under an unrelated background lifetime.
4. **Impact:** unnecessary database/CPU work and misleading request lifecycle behavior.
5. **Severity rationale:** **P2 retrospective** because the issue affects operational control and resource ownership rather than the normal valid-context data path.
6. **Existing guarantees violated:** explicit context ownership, bounded work, cancellation propagation.
7. **Considered solutions:** keep background fallback; synthesize a timeout; reject nil context.
8. **Chosen remediation:** reject nil context with `ErrContextRequired` at request-validation stage.
9. **Why selected:** only the caller can define the correct lifecycle; the callee must not invent one.
10. **Rejected alternatives:** fallback contexts hide programming errors and cannot preserve caller cancellation.
11. **Trade-offs:** callers that previously passed nil must fix their contract usage.
12. **Regression tests / protection:** `TestMaterializeRejectsNilContext`; permanent audit forbids `context.Background()` in the materializer path.
13. **Adversarial review findings:** the same ownership rule must apply to command/replay layers; later Replay hardening enforces it there too.
14. **Remediation iterations:** `2bbbd2439580536ffe17f8827c654c245d9b6b1e`.
15. **Residual risks / limitations:** non-nil callers can still choose unbounded contexts intentionally; that policy belongs to the caller.
16. **Operational/deployment consequences:** invalid nil-context calls fail immediately before I/O.
17. **Exact evidence:** remediation commit, nil-context test, run `30384357559`, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** repository/orchestration APIs performing external or potentially long-running work must reject nil caller contexts.

### GFA-OPS-294 — Materialization errors did not identify the failed orchestration stage

1. **Finding / symptom:** lower-layer errors propagated without identifying whether planning, reading, building, comparison, persistence, or contract reconciliation failed.
2. **Root cause:** orchestration had no typed stage-error boundary.
3. **Failure scenario:** two different operational failures expose similar wrapped errors and operators/callers cannot reliably classify the failing step.
4. **Impact:** slower diagnosis, weaker automation, and ambiguous incident evidence.
5. **Severity rationale:** **P2 retrospective** because failures were not hidden, but their operational location was under-specified.
6. **Existing guarantees violated:** diagnosable orchestration, typed failure ownership, actionable production evidence.
7. **Considered solutions:** prefix strings; add logs only; introduce typed stages with wrapped causes.
8. **Chosen remediation:** `StageError` with explicit request-validation, planning, read, snapshot-contract, builder, comparison, persistence, and persistence-contract stages.
9. **Why selected:** typed errors preserve `errors.Is/As` behavior while adding stable orchestration classification.
10. **Rejected alternatives:** log-only or string-prefix schemes are not machine-checkable contracts.
11. **Trade-offs:** more exported stage vocabulary that must remain stable within the versioned materialization contract.
12. **Regression tests / protection:** `TestMaterializeWrapsRepositoryAndPersistenceStages` and audit checks for stage errors.
13. **Adversarial review findings:** stage labels must wrap the original cause instead of replacing it, otherwise classification would destroy causal evidence.
14. **Remediation iterations:** `2bbbd2439580536ffe17f8827c654c245d9b6b1e`.
15. **Residual risks / limitations:** stage classification identifies where failure occurred, not every provider/database root cause.
16. **Operational/deployment consequences:** callers and logs can classify failures without parsing prose.
17. **Exact evidence:** remediation commit, stage-error tests, run `30384357559`, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** multi-stage production orchestrators should expose stable typed stage ownership while preserving wrapped causes.

### GFA-MAINT-295 — Exported DatasetLimitOr duplicated already-normalized request state

1. **Finding / symptom:** `DatasetLimitOr` remained an exported helper even though request normalization had already selected the effective dataset limit.
2. **Root cause:** normalization policy and execution state were represented twice.
3. **Failure scenario:** future code changes one normalization path but another caller re-applies defaults through the helper, creating divergent effective limits.
4. **Impact:** maintainability drift and avoidable ambiguity around which value is authoritative.
5. **Severity rationale:** **P3 retrospective** because no independent production data-corruption scenario was proven once normalization was already correct.
6. **Existing guarantees violated:** single ownership of normalized configuration and minimal public API surface.
7. **Considered solutions:** retain/document helper; make it private; remove it and assign normalized values directly.
8. **Chosen remediation:** remove `DatasetLimitOr`; execution consumes normalized request state directly.
9. **Why selected:** one canonical normalization path is simpler and eliminates duplicated policy.
10. **Rejected alternatives:** retaining the helper preserves an unnecessary second policy surface.
11. **Trade-offs:** source callers of the unnecessary helper must use normalized request flow instead.
12. **Regression tests / protection:** default/maximum dataset-limit tests; audit forbids the helper.
13. **Adversarial review findings:** this is not classified as a P1/P2 data defect merely because it was removed in the same hardening commit.
14. **Remediation iterations:** `2bbbd2439580536ffe17f8827c654c245d9b6b1e`.
15. **Residual risks / limitations:** other modules can still duplicate normalization unless reviewed independently.
16. **Operational/deployment consequences:** none beyond source/API cleanup inside the internal module.
17. **Exact evidence:** remediation commit, dataset-limit regression tests, run `30384357559`, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** once configuration is normalized, downstream execution should consume the normalized value rather than re-default it.

### GFA-TEST-296 — Materialization regression coverage omitted critical two-period and persistence invariants

1. **Finding / symptom:** tests did not protect independent period limits, atomic two-period reads, snapshot metadata, persisted outcome identity, stage errors, generated-time identity, default/maximum limits, or clone isolation.
2. **Root cause:** tests covered happy-path orchestration more strongly than its cross-layer integrity boundaries.
3. **Failure scenario:** a refactor silently reintroduces combined limits, stale snapshot acceptance, pre-persistence return identity, or mutable clone leakage while ordinary result tests still pass.
4. **Impact:** previously corrected P1/P2 invariants could regress without CI detection.
5. **Severity rationale:** **P2 retrospective** because the gap weakens permanent prevention for multiple high-impact contracts.
6. **Existing guarantees violated:** regression protection, CI truth, durable remediation closure.
7. **Considered solutions:** rely on package tests indirectly; add targeted tests only; combine tests with a permanent source/contract audit.
8. **Chosen remediation:** add targeted regression tests and `historicalmaterializationreviewaudit` enforced by Backend CI.
9. **Why selected:** behavioral tests prove outcomes while the audit protects critical contract wiring and closure evidence.
10. **Rejected alternatives:** prose-only closure or incidental coverage does not prove the remediation remains reachable.
11. **Trade-offs:** CI has additional tests/audit work and source guards require deliberate updates when contracts evolve.
12. **Regression tests / protection:** the named Materialization tests plus strict review audit itself.
13. **Adversarial review findings:** audit checks are not substitutes for behavior tests; both layers are retained.
14. **Remediation iterations:** `2bbbd2439580536ffe17f8827c654c245d9b6b1e`.
15. **Residual risks / limitations:** no finite test suite proves every future orchestration failure mode.
16. **Operational/deployment consequences:** stronger merge gate; no runtime behavior added by the guard itself.
17. **Exact evidence:** remediation commit, `apps/api/tools/historicalmaterializationreviewaudit`, Backend CI run `30384357559` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** remediation closure for cross-layer orchestrators requires targeted behavioral tests plus a permanent CI-enforced contract audit.
