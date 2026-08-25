# Document 112 — Extractor Input Correctness Hardening

Status: IMPLEMENTED
Baseline commit: ff84eefdb8ab363e2bdd276f99e49df7235fb50f

## 1. Scope

This increment closes the critical input-correctness findings from the
`internal/features/extractor` review without conflating event time with system
provenance time.

The extractor now rejects any trajectory snapshot containing:

- a point observed after `Request.AsOfTime`;
- a segment start or end after `Request.AsOfTime`;
- a coverage-gap start or end after `Request.AsOfTime`.

`CreatedAt` and `UpdatedAt` remain system provenance timestamps and are not
incorrectly treated as aviation-event cutoffs.

## 2. Dependency and cancellation contract

Nil contexts are rejected explicitly. Required builder interfaces reject typed
nil values. A typed-nil optional aircraft provider is treated as unavailable.
Cancellation is checked again immediately after aircraft enrichment before
canonical serialization and hashing.

## 3. Quality and mathematics integrity

Initial quality construction rejects negative field counts, available counts
greater than total counts, negative supporting-point counts, non-finite input
quality and input quality outside the inclusive range `[0, 1]`. Invalid values
are no longer converted into ordinary zero or one scores.

## 4. Canonical fingerprint semantics

ICAO24 and callsign values are canonicalized consistently before hashing.
Semantically equivalent casing and surrounding whitespace therefore do not
create different snapshot identities.

## 5. Processing generation

Nested evidence acceptance and fingerprint canonicalization change processing
semantics. The extractor, extractor composition and feature processing pipeline
therefore advance to generation `v4`. Existing stored snapshots remain readable
through their explicit processing versions.

## 6. Deferred review scope

This increment intentionally does not claim closure for:

- explicit aircraft source and retrieval provenance;
- replacement of the `TrajectoryUpdatedAt` fallback;
- separation of required completeness from optional coverage;
- centralization of duplicate ICAO24 and clone helpers.

Those items require a separate contract change rather than being hidden inside
this correctness patch.

```text
EXTRACTOR_NESTED_TEMPORAL_GUARD=ENFORCED
EXTRACTOR_NIL_CONTEXT=REJECTED
EXTRACTOR_TYPED_NIL_DEPENDENCIES=REJECTED
EXTRACTOR_POST_PROVIDER_CANCELLATION=ENFORCED
EXTRACTOR_SEMANTIC_FINGERPRINT_NORMALIZATION=ENFORCED
EXTRACTOR_INVALID_EVIDENCE_COUNTS=REJECTED
EXTRACTOR_INVALID_MATH_MASKING=CLOSED
EXTRACTOR_PROCESSING_GENERATION=v4
```

---

## Canonical remediation history

### GFA-DATA-140 — nested trajectory evidence after `AsOfTime` could enter historical feature extraction

1. **Finding / symptom.** Top-level trajectory end-time validation did not prove that every point, segment bound and coverage-gap bound was at or before `Request.AsOfTime`.
2. **Root cause.** Historical admission validated only the aggregate trajectory window and assumed nested evidence was temporally consistent with it.
3. **Failure scenario.** A trajectory reports an acceptable `EndTime` but contains a point or nested interval extending one nanosecond beyond the requested historical boundary; that future evidence contributes to extracted features.
4. **Impact.** Historical feature snapshots can leak future aviation observations even when their top-level window appears valid.
5. **Severity rationale.** **P1 retrospective.** This is direct temporal leakage into historical analytical data.
6. **Existing guarantees violated.** Every aviation-event timestamp contributing to a historical snapshot must be bounded by the request as-of time; system provenance timestamps are a separate contract.
7. **Considered solutions.** Trust trajectory invariants; truncate nested evidence; reject snapshots containing out-of-bound evidence.
8. **Chosen remediation.** Validate point observations, segment start/end and coverage-gap start/end against `AsOfTime` and return typed errors for any future nested evidence.
9. **Why this solution was selected.** Rejecting preserves evidence integrity without inventing partial segments/gaps or silently mutating caller snapshots.
10. **Rejected alternatives.** Trust-only handling leaves the leak open; truncation changes domain evidence and can fabricate segment semantics.
11. **Trade-offs.** A malformed/inconsistent trajectory must be repaired upstream before historical feature extraction can proceed.
12. **Regression tests / protection.** Tests cover point, segment-start, segment-end and coverage-gap future evidence; `extractorcorrectnessaudit` requires the nested temporal guard.
13. **Adversarial review findings.** `CreatedAt` and `UpdatedAt` are system provenance and must not be compared as aviation event cutoffs; doing so would reject legitimate later materialization.
14. **Remediation iterations.** Closed in `e853f593…`; later provenance work in Document 113 explicitly preserves the event-time/system-time distinction.
15. **Residual risks and limitations.** The guard validates timestamps present in the trajectory model; it cannot validate external evidence not represented by those fields.
16. **Operational or deployment consequences.** Invalid historical extraction requests fail fast instead of persisting leakage; processing generation advances to v4.
17. **Exact evidence.** Implementation commit `e853f5931c78f6ed7b0fbcd0dd85a53cfbaa22f3` (`fix: enforce extractor input correctness`). Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-DATA-140=CLOSED`.
19. **Prevention / future guard.** New nested aviation-event timestamps added to trajectory evidence must be incorporated into the as-of validation and fingerprint mirror tests before release.

### GFA-OPS-141 — extractor silently replaced a nil caller context

1. **Finding / symptom.** `Extract` accepted nil context and substituted `context.Background()`.
2. **Root cause.** Nil was treated as convenience fallback rather than a caller lifecycle contract violation.
3. **Failure scenario.** A job/request accidentally loses its context; extraction continues detached from cancellation, deadlines or shutdown.
4. **Impact.** Feature work can outlive its owner and continue builder/provider processing unexpectedly.
5. **Severity rationale.** **P2 retrospective.** Caller misuse is required, but silent substitution breaks cancellation ownership and reliability.
6. **Existing guarantees violated.** Long-running processing contexts must be explicitly caller-owned; background contexts are reserved for documented bounded cleanup.
7. **Considered solutions.** Preserve Background fallback; synthesize a timeout; reject nil.
8. **Chosen remediation.** Return `ErrContextRequired` for nil input.
9. **Why this solution was selected.** It exposes the programming error at the boundary and retains caller deadlines/cancellation.
10. **Rejected alternatives.** Internal fallback cannot reconstruct caller ownership or deadline intent.
11. **Trade-offs.** Tests/callers must pass a concrete context explicitly.
12. **Regression tests / protection.** `TestExtractorRejectsNilContext` and permanent correctness audit protect the rule.
13. **Adversarial review findings.** Explicit `context.Background()` remains valid when deliberately owned by the caller; the extractor must not infer intent.
14. **Remediation iterations.** Closed in `e853f593…`; no later increment reintroduces context substitution.
15. **Residual risks and limitations.** Callers can still choose an unbounded context intentionally.
16. **Operational or deployment consequences.** Accidental nil-context extraction fails immediately.
17. **Exact evidence.** Implementation commit `e853f5931c78f6ed7b0fbcd0dd85a53cfbaa22f3`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-OPS-141=CLOSED`.
19. **Prevention / future guard.** Extractor/builder entry points must reject nil caller contexts unless explicitly documented as cleanup-only functions.

### GFA-ARCH-142 — typed-nil extractor dependencies were not modeled according to required/optional semantics

1. **Finding / symptom.** Required builder interfaces could contain typed-nil concrete pointers, while an optional typed-nil aircraft provider could appear present.
2. **Root cause.** Constructor checks relied on interface nil comparison rather than concrete nil-aware dependency semantics.
3. **Failure scenario.** A typed-nil required builder survives construction and panics on use, or a typed-nil optional provider is invoked instead of being treated as unavailable enrichment.
4. **Impact.** Construction validation can be bypassed, producing runtime failure or incorrect optional-dependency behavior.
5. **Severity rationale.** **P2 retrospective.** Misconfiguration is required, but invalid dependencies can survive startup and break production extraction.
6. **Existing guarantees violated.** Required dependencies must be concretely non-nil; optional missing dependencies must map to explicit unavailable behavior.
7. **Considered solutions.** Document caller responsibility; validate typed nils; remove optional interface support.
8. **Chosen remediation.** Use nil-capable `dependencyMissing` checks for builders/providers; reject typed-nil required builders and normalize typed-nil optional aircraft provider to absent.
9. **Why this solution was selected.** It preserves the intended required/optional API while closing Go interface typed-nil ambiguity at construction.
10. **Rejected alternatives.** Documentation cannot enforce runtime safety; removing optional enrichment would reduce supported processing modes.
11. **Trade-offs.** Small reflection logic remains in constructor validation.
12. **Regression tests / protection.** Tests cover typed-nil required builder rejection and optional provider normalization; correctness audit enforces the dependency contract.
13. **Adversarial review findings.** Required and optional typed nils must not share one outcome: required values error, optional enrichment becomes unavailable.
14. **Remediation iterations.** Extractor-level handling added in `e853f593…`, complementing composition-level lookup typed-nil handling from Document 109.
15. **Residual risks and limitations.** Non-nil dependencies can still fail behaviorally; availability/functional errors remain separate contracts.
16. **Operational or deployment consequences.** Invalid required composition fails at construction; absent optional enrichment is represented without panic.
17. **Exact evidence.** Implementation commit `e853f5931c78f6ed7b0fbcd0dd85a53cfbaa22f3`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-ARCH-142=CLOSED`.
19. **Prevention / future guard.** New interface dependencies must declare required/optional semantics and include typed-nil regression tests.

### GFA-OPS-143 — cancellation occurring during aircraft enrichment was not rechecked before hashing/publication

1. **Finding / symptom.** The extractor checked cancellation before work, but after the aircraft provider returned it could continue canonical serialization/fingerprinting even if the provider had caused or observed cancellation.
2. **Root cause.** Cancellation checks bracketed the overall operation incompletely; the external enrichment boundary was not followed by a second ownership check.
3. **Failure scenario.** Context is cancelled while aircraft enrichment completes; extractor continues into hashing/assembly and may return a successful feature payload after cancellation.
4. **Impact.** Callers can receive work/result completion after they cancelled the operation, weakening deterministic shutdown and resource ownership.
5. **Severity rationale.** **P2 retrospective.** This is lifecycle correctness at an external/possibly blocking boundary, not data corruption under normal uncancelled execution.
6. **Existing guarantees violated.** Cancellation must be honored before expensive post-provider work and before successful publication.
7. **Considered solutions.** Rely on provider to return cancellation; check only at method start; recheck immediately after provider enrichment.
8. **Chosen remediation.** Call `ctx.Err()` immediately after aircraft enrichment and return cancellation before fingerprint serialization/assembly.
9. **Why this solution was selected.** It protects the boundary even when a provider returns a value while the context becomes cancelled concurrently.
10. **Rejected alternatives.** Provider-only enforcement couples extractor correctness to every provider implementation and leaves a race between return and post-processing.
11. **Trade-offs.** One additional cheap context check is performed per extraction.
12. **Regression tests / protection.** A test provider cancels the context during `Provide`; extractor must return `context.Canceled`; correctness audit protects the check.
13. **Adversarial review findings.** Cancellation is a lifecycle signal, not a reason to persist a partial feature result; the method returns no successful feature snapshot after the check fails.
14. **Remediation iterations.** Closed in `e853f593…` with no later semantic reversal.
15. **Residual risks and limitations.** Cancellation can still arrive after the final check; ordinary Go cancellation is cooperative and cannot make every instruction atomic.
16. **Operational or deployment consequences.** Cancelled materialization stops earlier and avoids unnecessary hashing/publication work.
17. **Exact evidence.** Implementation commit `e853f5931c78f6ed7b0fbcd0dd85a53cfbaa22f3`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-OPS-143=CLOSED`.
19. **Prevention / future guard.** New external/blocking enrichment stages must define post-stage cancellation checkpoints before deterministic persistence/publication work.

### GFA-DATA-144 — invalid evidence counts could enter completeness calculation

1. **Finding / symptom.** Initial quality construction could receive negative field counts, available counts greater than totals or negative supporting-point counts.
2. **Root cause.** Evidence aggregation assumed builder-produced counts were valid and computed scores without independently enforcing count invariants.
3. **Failure scenario.** A builder returns `available=2,total=1` or a negative count; extractor converts impossible evidence into an ordinary numeric completeness/support value.
4. **Impact.** Feature quality can encode mathematically impossible evidence while appearing valid to downstream consumers.
5. **Severity rationale.** **P1 retrospective.** Quality scores are trust/admission evidence for analytical features; impossible counts must not be normalized into valid-looking output.
6. **Existing guarantees violated.** Evidence counts must be non-negative, available must not exceed total, and supporting-point counts must be non-negative before score construction.
7. **Considered solutions.** Clamp invalid counts; trust builders; reject invalid evidence explicitly.
8. **Chosen remediation.** Make `buildInitialQuality` return errors for invalid field/supporting counts before calculating quality.
9. **Why this solution was selected.** Fail-closed handling preserves evidence truth; clamping would fabricate plausible values from invalid input.
10. **Rejected alternatives.** Clamping/masking hides upstream contract failures; builder-only validation leaves aggregation vulnerable to new/custom builders.
11. **Trade-offs.** Builder bugs now fail extraction rather than producing degraded output.
12. **Regression tests / protection.** Tests cover available>total and related invalid count cases; correctness audit requires typed count validation.
13. **Adversarial review findings.** Count integrity must be validated before summation/division to avoid both invalid ratios and overflow-like reasoning errors from negative values.
14. **Remediation iterations.** Closed in `e853f593…`; Document 113 later refines which required/optional counts belong in each score without weakening count invariants.
15. **Residual risks and limitations.** Count validity does not prove that builders counted the semantically correct fields; schema-derived counts/requirement semantics provide additional protection later.
16. **Operational or deployment consequences.** Invalid builder evidence becomes an extraction error and cannot be persisted as valid quality metadata.
17. **Exact evidence.** Implementation commit `e853f5931c78f6ed7b0fbcd0dd85a53cfbaa22f3`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-DATA-144=CLOSED`.
19. **Prevention / future guard.** New evidence groups/builders must preserve count invariants and aggregation must continue validating them independently.

### GFA-DATA-145 — non-finite or out-of-range trajectory quality could be masked into ordinary scores

1. **Finding / symptom.** `NaN`, infinities or input quality outside `[0,1]` could reach initial quality construction and be converted/masked into apparently normal values.
2. **Root cause.** Numeric normalization prioritized producing a bounded score instead of rejecting invalid upstream mathematical evidence.
3. **Failure scenario.** Trajectory quality is `NaN`, `+Inf`, negative or greater than one; downstream feature quality receives zero/one or another valid-looking value rather than an error.
4. **Impact.** Corrupt/non-finite analytical evidence becomes indistinguishable from legitimate low/high quality.
5. **Severity rationale.** **P1 retrospective.** Invalid floating-point evidence can silently become trusted feature-quality data.
6. **Existing guarantees violated.** Input quality must be finite and within the documented inclusive probability/score range before use.
7. **Considered solutions.** Clamp values; map non-finite to zero; reject invalid numeric evidence.
8. **Chosen remediation.** Return `ErrInvalidInputQualityScore` for non-finite or out-of-range values.
9. **Why this solution was selected.** Rejection preserves the distinction between genuine low quality and invalid mathematics.
10. **Rejected alternatives.** Clamping and zero substitution manufacture valid evidence from invalid state.
11. **Trade-offs.** Upstream quality bugs now block feature materialization until corrected.
12. **Regression tests / protection.** Tests cover `NaN`, both infinities and range validation; correctness audit owns the finite/range contract.
13. **Adversarial review findings.** Checks must test finiteness explicitly; ordinary comparisons do not reject `NaN` reliably because comparisons with NaN are false.
14. **Remediation iterations.** Closed in `e853f593…`; later quality semantics change score composition but retain strict numeric admission.
15. **Residual risks and limitations.** Finite in-range values can still be semantically wrong if upstream calculation is wrong; this finding addresses representational validity.
16. **Operational or deployment consequences.** Invalid input quality fails extraction instead of being persisted.
17. **Exact evidence.** Implementation commit `e853f5931c78f6ed7b0fbcd0dd85a53cfbaa22f3`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-DATA-145=CLOSED`.
19. **Prevention / future guard.** All new floating-point quality/confidence inputs must validate finiteness and domain range before normalization or aggregation.

### GFA-DATA-146 — semantically equivalent aircraft identity text produced different extraction fingerprints

1. **Finding / symptom.** ICAO24/callsign casing and surrounding whitespace were not canonicalized consistently before fingerprint serialization.
2. **Root cause.** Runtime/domain consumers normalized identity fields differently from the canonical fingerprint structure.
3. **Failure scenario.** `abc123`, `ABC123` and `  ABC123 ` describe the same aircraft but produce different fingerprint bytes and therefore different durable snapshot identities.
4. **Impact.** Equivalent extraction inputs can create duplicate snapshots and break deterministic replay/idempotency.
5. **Severity rationale.** **P2 retrospective.** The data remains interpretable, but semantic identity fragmentation creates duplicate durable processing records.
6. **Existing guarantees violated.** Canonical fingerprinting must normalize identity fields according to domain semantics before hashing.
7. **Considered solutions.** Preserve raw caller text; normalize only at HTTP/repository boundary; normalize canonical fingerprint fields.
8. **Chosen remediation.** Canonicalize ICAO24 and callsign casing/whitespace before serialization/hash.
9. **Why this solution was selected.** Deterministic identity belongs at the fingerprint boundary and must not depend on which upstream caller already normalized text.
10. **Rejected alternatives.** Raw hashing makes idempotency sensitive to presentation differences; relying solely on upstream normalization is fragile across call sites.
11. **Trade-offs.** Fingerprint semantics change intentionally, requiring processing generation v4 to isolate previous snapshots.
12. **Regression tests / protection.** Tests assert equivalent spelling produces identical fingerprint; correctness audit and later reflection mirror guards protect canonical structure.
13. **Adversarial review findings.** Canonicalization must preserve semantic distinctions; only domain-equivalent casing/outer whitespace is normalized.
14. **Remediation iterations.** Fixed in `e853f593…`; Document 114 later centralizes ICAO24 canonicalization in `domain/aircraft` and installs structural fingerprint mirror tests.
15. **Residual risks and limitations.** Callsign normalization follows current semantic policy; future domain changes require an intentional processing-generation review.
16. **Operational or deployment consequences.** Equivalent identity spellings converge on one fingerprint under generation v4+.
17. **Exact evidence.** Implementation commit `e853f5931c78f6ed7b0fbcd0dd85a53cfbaa22f3`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-DATA-146=CLOSED`.
19. **Prevention / future guard.** Canonical fingerprint fields representing domain identities must consume centralized normalization semantics and carry regression tests for equivalent representations.
