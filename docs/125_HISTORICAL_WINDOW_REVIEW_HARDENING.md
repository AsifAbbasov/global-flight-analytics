# Historical Window Review Hardening

Status: implemented review remediation
Module: `apps/api/internal/historicalintelligence/historicalwindow`
Baseline: `fc254881fa446c7e80f94a959e2a9d5609874821`

## Confirmed defects

The review correctly identified that `time.Time.Sub` was unsafe for bucket-count and previous-window construction across windows longer than the `time.Duration` range. The planner now enforces the bucket limit during generation, advances only through validated calendar boundaries, and derives calendar previous windows from bucket steps rather than a saturated duration.

The review also correctly identified incomplete context cancellation, mutable derived plan evidence, execution-policy contamination of the semantic fingerprint, missing bucket and previous-window fingerprint evidence, unaligned public `NextBoundary` behavior, and negative duration helpers. These boundaries are now hardened.

## Deliberately rejected findings

`GranularityCustom` remains part of the internal Historical Intelligence contract. The production command intentionally exposes only hour, day, and week, but the reusable domain planner and schema support one custom bucket. Command exposure and domain capability are separate boundaries.

Nullable `EffectiveWindow` and `PreviousWindow` pointers remain intentional because a valid request may produce no complete bucket. Replacing semantic absence with zero-value windows would weaken the contract.

The nil-safe `BucketCountExceededError.Error` and `Unwrap` methods remain. Returning `nil` from `Unwrap` for a nil receiver follows the Go error-chain contract and is not equivalent to an unexplained domain null.

The long-function and switch-count observations were treated as maintainability signals rather than independent architecture failures. Refactoring was applied where it removed semantic duplication or enabled correctness, not to satisfy mechanical line-count rules.

## Resulting contract

- `Build` rejects a nil context and checks cancellation during every bucket iteration and before return.
- Bucket limits are enforced before every append.
- `NextBoundary` requires an aligned boundary.
- One granularity policy owns floor and shift behavior.
- Calendar previous windows use exact bucket shifts.
- Custom previous windows use checked second and nanosecond arithmetic instead of `time.Duration`.
- `CanonicalizePlan` reconstructs all derived evidence from request semantics.
- `ValidatePlan` verifies fingerprint, windows, buckets, keys, sequences, exclusions, UTC normalization, truncation, and execution-limit consistency.
- Traffic, airport, route, and generic series builders canonicalize plans before analytics and fingerprint generation.
- Fingerprint generation two excludes `MaximumBucketCount` while binding effective and previous windows, buckets, keys, sequences, exclusions, and truncation.
- Reversed bucket and exclusion intervals have zero duration and are rejected by canonical validation.

No PostgreSQL schema migration is required. Existing rows remain readable, but an exact replay against a pre-generation-two row may fail closed with `ErrResultConflict` because corrected semantic provenance must not silently overwrite immutable historical evidence. Any rebuild of pre-generation-two identities must therefore be an explicit governed reset rather than an automatic compatibility rewrite.

## Formal closure evidence

The Historical Window remediation is implemented by
`caa6a0ee7f1309801e1671e06d38535b28aa2437` (`fix: harden historical window planning`).
The original pull-request-triggered GitHub Actions run for that commit is not recoverable through the available repository evidence, so no historical exact-commit CI claim is reconstructed. Current exact-head regression evidence exists on PR #123 head `bbe20b1e2a9da873e0b5400aac136bd0b0c006c8`: Backend CI run `32835560290` completed successfully and Backend Quality executed `historicalwindowreviewaudit -strict` successfully.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
HISTORICAL_WINDOW_REVIEW_STATUS=CLOSED
```

## Canonical remediation history

Historical adversarial-review identities/comments and original exact-commit Continuous Integration evidence are unavailable; reconstruction is limited to repository source, tests, implementation commit history, permanent strict audit evidence, and later exact-head regression evidence. Severity labels are retrospective.

### GFA-DATA-223 — Historical window duration arithmetic could saturate across ranges longer than `time.Duration`

1. **Finding / symptom:** bucket-count and previous-window construction used `time.Time.Sub`, whose `time.Duration` result cannot represent every valid `time.Time` range.
2. **Root cause:** wall-clock/calendar interval planning was reduced to nanosecond duration arithmetic even when the requested interval could exceed the duration range.
3. **Failure scenario:** a very large valid request saturates duration arithmetic and derives an incorrect bucket count or previous window.
4. **Impact:** planner output, comparison periods and deterministic replay identity can be mathematically wrong for large windows.
5. **Severity rationale:** **P1 retrospective** because authoritative temporal boundaries can be silently miscomputed.
6. **Existing guarantees violated:** Historical Window must derive calendar boundaries from validated time semantics without overflow/saturation.
7. **Considered solutions:** cap all legal windows to `time.Duration`, use big integers, or build calendar windows through bounded bucket stepping and checked custom arithmetic.
8. **Chosen remediation:** calendar windows advance through exact granularity steps; previous calendar windows derive from bucket steps; custom windows use checked second/nanosecond arithmetic instead of `time.Duration`.
9. **Why this solution was selected:** it preserves the supported `time.Time` domain and existing granularity semantics without introducing arbitrary global range caps.
10. **Rejected alternatives:** relying on saturated `time.Time.Sub` and silently narrowing the domain to duration-representable windows.
11. **Trade-offs:** planning logic becomes more explicit and must maintain checked arithmetic helpers.
12. **Regression tests / protection:** long-range calendar/custom-window tests and `historicalwindowreviewaudit -strict`.
13. **Adversarial review findings:** custom granularity remains valid as an internal one-bucket contract and is not removed to avoid this arithmetic issue.
14. **Remediation iterations:** implemented in `caa6a0ee7f1309801e1671e06d38535b28aa2437`.
15. **Residual risks and limitations:** custom arithmetic remains bounded by the `time.Time` representable domain and explicit validation.
16. **Operational or deployment consequences:** no PostgreSQL migration; old immutable identities may conflict on exact replay after fingerprint generation changes and require governed reset/rematerialization.
17. **Exact evidence:** implementation commit `caa6a0ee7f1309801e1671e06d38535b28aa2437`; long-window tests; permanent historical-window audit; later exact-head PR #123 Backend CI run `32835560290`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** calendar-domain planning must not use `time.Duration` as an unchecked universal representation of elapsed wall-clock ranges.

### GFA-PERF-224 — Bucket limits were not enforced during generation before allocation growth

1. **Finding / symptom:** the maximum bucket count could be evaluated after planner work had already generated more buckets than the configured budget.
2. **Root cause:** resource policy was treated as a final validation condition rather than an invariant of the generation loop.
3. **Failure scenario:** an oversized request allocates/appends a large bucket sequence before returning a limit error.
4. **Impact:** avoidable CPU/memory growth and a weak denial-of-service boundary in a user-controlled planning path.
5. **Severity rationale:** **P2 retrospective** because the defect is bounded-work/resource integrity rather than direct fabrication of analytical values.
6. **Existing guarantees violated:** maximum bucket count must bound work before each incremental allocation.
7. **Considered solutions:** validate estimated count up front, keep post-generation validation, or enforce the limit before every append.
8. **Chosen remediation:** planner checks the normalized bucket budget before each append and fails immediately when the next bucket would exceed the limit.
9. **Why this solution was selected:** it remains correct even when calendar bucket count cannot be safely derived through a single duration division.
10. **Rejected alternatives:** allocate then reject and unsafe pre-count arithmetic based on `time.Duration`.
11. **Trade-offs:** one lightweight check is added to each generation iteration.
12. **Regression tests / protection:** boundary/exceeded-count tests and permanent review audit.
13. **Adversarial review findings:** `MaximumBucketCount` remains an operator/execution policy and is intentionally excluded from semantic fingerprint identity.
14. **Remediation iterations:** `caa6a0ee7f1309801e1671e06d38535b28aa2437`.
15. **Residual risks and limitations:** a request at the configured maximum still legitimately performs work proportional to that maximum.
16. **Operational or deployment consequences:** predictable memory/CPU budget for window planning.
17. **Exact evidence:** implementation commit, bucket-limit regression tests, `historicalwindowreviewaudit`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** collection-size policies must be enforced inside generation loops before allocation, not only after construction.

### GFA-OPS-225 — Historical Window could lose caller cancellation ownership

1. **Finding / symptom:** context cancellation was incomplete across bucket iteration and nil context did not have one strict entry contract.
2. **Root cause:** planning was treated as pure/cheap computation and lifecycle checks were not propagated through the full loop.
3. **Failure scenario:** a cancelled request continues generating a large bucket plan or a nil caller context creates work with no explicit lifecycle owner.
4. **Impact:** wasted CPU and inconsistent cancellation behavior across Historical Intelligence.
5. **Severity rationale:** **P2 retrospective** as an operational lifecycle correctness defect.
6. **Existing guarantees violated:** potentially large analytical work must have explicit caller-owned context and bounded cancellation responsiveness.
7. **Considered solutions:** substitute `context.Background`, check only on entry, check every iteration, or poll periodically.
8. **Chosen remediation:** reject nil context, check cancellation through each bucket iteration and once more before returning.
9. **Why this solution was selected:** planning iterations are bounded and the per-iteration context check gives simple deterministic responsiveness.
10. **Rejected alternatives:** implicit background lifetime and entry-only cancellation.
11. **Trade-offs:** minimal per-bucket context-check overhead.
12. **Regression tests / protection:** nil-context and mid-planning cancellation tests plus strict audit.
13. **Adversarial review findings:** nil-safe error `Unwrap` behavior remains idiomatic and unrelated to caller context ownership.
14. **Remediation iterations:** `caa6a0ee7f1309801e1671e06d38535b28aa2437`.
15. **Residual risks and limitations:** cancellation cannot interrupt an individual constant-time boundary calculation, only the loop boundary around it.
16. **Operational or deployment consequences:** cancelled materializations stop window planning predictably.
17. **Exact evidence:** implementation commit, context tests, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** all loops whose work scales with requested historical range must carry and test explicit context cancellation.

### GFA-DATA-226 — Mutable derived Historical Plan evidence could be trusted without canonical reconstruction

1. **Finding / symptom:** callers could mutate derived plan fields such as fingerprint, bucket key/sequence, previous window, exclusions or truncation after the original request had been planned.
2. **Root cause:** `Plan` combined request semantics with mutable derived evidence and downstream builders trusted the provided derived fields.
3. **Failure scenario:** a tampered plan reaches Traffic/Airport/Route/Series builders and changes analytics or fingerprint identity without changing the authoritative request semantics.
4. **Impact:** non-deterministic or forged Historical Result provenance and bucket identity.
5. **Severity rationale:** **P1 retrospective** because mutable derived metadata could influence durable analytical results.
6. **Existing guarantees violated:** derived evidence must be a deterministic function of canonical request semantics.
7. **Considered solutions:** make every nested plan field immutable through a new type system, trust callers, or canonicalize/revalidate at consumption boundaries.
8. **Chosen remediation:** `CanonicalizePlan` reconstructs derived evidence from request semantics; `ValidatePlan` verifies all derived windows, buckets, keys, sequences, exclusions, UTC normalization, truncation, fingerprint and limit consistency; builders canonicalize before use.
9. **Why this solution was selected:** it provides defense in depth without a broad breaking domain-type migration.
10. **Rejected alternatives:** silently accepting caller-mutated derived fields and replacing optional windows with zero-value sentinels.
11. **Trade-offs:** downstream builders perform deterministic plan reconstruction before analytics.
12. **Regression tests / protection:** mutable-plan tampering tests across generic Series, Traffic, Airport and Route builders; strict window audit.
13. **Adversarial review findings:** nullable Effective/Previous windows remain intentional semantic absence; canonicalization validates that absence rather than manufacturing zero windows.
14. **Remediation iterations:** `caa6a0ee7f1309801e1671e06d38535b28aa2437`.
15. **Residual risks and limitations:** caller-owned request semantics themselves still require normal validation and cannot be inferred if intentionally falsified.
16. **Operational or deployment consequences:** exact replays against pre-generation-two persisted fingerprints can fail closed instead of silently overwriting immutable evidence.
17. **Exact evidence:** implementation commit, plan canonicalization/integrity tests, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** any structure containing both authoritative inputs and derived identity evidence must be canonicalized from the inputs before persistence/fingerprinting.

### GFA-DATA-227 — Execution-only bucket limits contaminated semantic Historical plan identity

1. **Finding / symptom:** `MaximumBucketCount` could participate in the plan fingerprint even when two requests produced the same semantic windows/buckets.
2. **Root cause:** operator execution policy and analytical result identity were not separated.
3. **Failure scenario:** identical historical evidence planned under two different safety limits produces different fingerprints despite identical analytical semantics.
4. **Impact:** false identity divergence, unnecessary duplicate immutable records and replay conflicts.
5. **Severity rationale:** **P2 retrospective** because values remain correct but deterministic semantic identity is polluted by non-semantic execution configuration.
6. **Existing guarantees violated:** fingerprints must bind output-affecting semantics, not incidental execution budgets.
7. **Considered solutions:** include every config value, remove all policy from plan validation, or separate execution policy from semantic fingerprint while still validating limit consistency.
8. **Chosen remediation:** fingerprint generation two excludes `MaximumBucketCount`; plan validation still checks that the generated plan obeys the normalized limit.
9. **Why this solution was selected:** preserves operational safety without making resource configuration part of analytical identity.
10. **Rejected alternatives:** treating every runtime knob as semantic provenance.
11. **Trade-offs:** operators cannot infer the historical planning limit from the semantic fingerprint alone; it remains execution metadata.
12. **Regression tests / protection:** fingerprint equality tests across different safe limits and strict audit.
13. **Adversarial review findings:** this separation is the reason bucket limit remains a valid distinct operator control instead of being removed.
14. **Remediation iterations:** `caa6a0ee7f1309801e1671e06d38535b28aa2437`.
15. **Residual risks and limitations:** future configuration fields must be classified explicitly as semantic or execution-only.
16. **Operational or deployment consequences:** corrected generation-two identities may conflict with immutable pre-v2 rows on exact replay, requiring governed rematerialization.
17. **Exact evidence:** implementation commit and fingerprint policy tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** fingerprint reviews must classify every configuration field by whether changing it can change analytical output.

### GFA-DATA-228 — Historical plan fingerprint omitted output-affecting derived bucket/window evidence

1. **Finding / symptom:** fingerprint identity did not bind all effective/previous windows, bucket keys/sequences, exclusions and truncation that drive downstream analytics.
2. **Root cause:** fingerprinting focused on high-level request fields rather than the full canonical prepared plan.
3. **Failure scenario:** two materially different canonical plans collide under one fingerprint because a bucket boundary, exclusion or derived previous window differs.
4. **Impact:** immutable Historical Results can share an identity despite different input evidence.
5. **Severity rationale:** **P1 retrospective** because an incomplete fingerprint breaks durable idempotency/provenance guarantees.
6. **Existing guarantees violated:** every output-affecting canonical plan element must participate in deterministic input identity.
7. **Considered solutions:** fingerprint raw request only, serialize the mutable plan directly, or fingerprint the canonical reconstructed plan.
8. **Chosen remediation:** generation-two fingerprint binds effective and previous windows, buckets, keys, sequences, exclusions and truncation from canonical plan evidence.
9. **Why this solution was selected:** identity follows the exact prepared representation consumed by analytics.
10. **Rejected alternatives:** raw mutable-plan hashing and omission of derived output-affecting evidence.
11. **Trade-offs:** fingerprint generation is more explicit and must evolve deliberately when plan semantics change.
12. **Regression tests / protection:** mutation/collision and generation-two fingerprint tests plus strict audit.
13. **Adversarial review findings:** execution-only `MaximumBucketCount` is deliberately excluded while semantic derived evidence is deliberately included.
14. **Remediation iterations:** `caa6a0ee7f1309801e1671e06d38535b28aa2437`.
15. **Residual risks and limitations:** any newly introduced output-affecting plan field must be added to fingerprint generation.
16. **Operational or deployment consequences:** immutable identity changes are isolated by fingerprint generation two; no database migration.
17. **Exact evidence:** implementation commit, canonical fingerprint tests, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** fingerprint tests must fail when the canonical prepared plan gains semantic fields that are not mirrored in identity generation.

### GFA-CONTRACT-229 — Public boundary stepping accepted unaligned timestamps and duplicated granularity behavior

1. **Finding / symptom:** `NextBoundary` could advance an unaligned input and floor/shift behavior was implemented in separate logic paths.
2. **Root cause:** granularity semantics were expressed through duplicated switches instead of one policy owning both floor and shift behavior.
3. **Failure scenario:** a caller supplies `10:37` for hourly planning and receives an apparently valid next boundary that is not part of the canonical bucket sequence.
4. **Impact:** malformed bucket sequences, inconsistent planner behavior and hard-to-audit boundary identity.
5. **Severity rationale:** **P2 retrospective** because the defect corrupts boundary contract semantics but is typically caught before durable analytics.
6. **Existing guarantees violated:** boundary stepping must start from a canonical aligned boundary and use one granularity policy.
7. **Considered solutions:** silently floor inside `NextBoundary`, leave behavior caller-dependent, or reject unaligned boundaries and centralize floor/shift policy.
8. **Chosen remediation:** one `granularityPolicy` owns floor/shift behavior and `NextBoundary` rejects inputs that are not already aligned.
9. **Why this solution was selected:** explicit preconditions prevent accidental double-normalization and keep planner rules deterministic.
10. **Rejected alternatives:** silently changing caller-provided timestamps inside a function named as a pure next-step operation.
11. **Trade-offs:** callers must call the canonical floor/ceil planner APIs rather than relying on permissive stepping.
12. **Regression tests / protection:** hour/day/week alignment and invalid-sequence tests; strict audit.
13. **Adversarial review findings:** custom granularity has no generic next-step policy and remains intentionally constrained to its supported one-bucket semantics.
14. **Remediation iterations:** `caa6a0ee7f1309801e1671e06d38535b28aa2437`.
15. **Residual risks and limitations:** calendar policies remain UTC-bound by the Historical contract.
16. **Operational or deployment consequences:** stricter caller validation only.
17. **Exact evidence:** implementation commit and boundary-policy tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** calendar boundary helpers must state alignment preconditions and share one versioned granularity policy.

### GFA-DATA-230 — Reversed historical intervals could expose negative duration evidence

1. **Finding / symptom:** bucket or exclusion duration helpers could return negative values for reversed start/end intervals.
2. **Root cause:** duration calculation assumed validated ordering even when helpers could observe malformed mutable evidence before canonical validation.
3. **Failure scenario:** a reversed interval contributes a negative duration to diagnostics/fingerprint-adjacent logic before the invalid plan is rejected.
4. **Impact:** contradictory temporal metadata and unstable downstream arithmetic.
5. **Severity rationale:** **P2 retrospective** because canonical validation rejects the malformed plan, but helper-level negative evidence weakens defense in depth.
6. **Existing guarantees violated:** invalid/reversed intervals must never be represented as meaningful negative observed duration.
7. **Considered solutions:** allow signed duration, panic, or return canonical zero while validation rejects the interval structurally.
8. **Chosen remediation:** reversed bucket/exclusion intervals report zero duration and canonical validation rejects their ordering.
9. **Why this solution was selected:** helpers remain safe and non-negative while the domain validator remains responsible for rejecting invalid structure.
10. **Rejected alternatives:** treating negative duration as legitimate analytical evidence.
11. **Trade-offs:** callers inspecting duration alone cannot distinguish reversed from zero-length intervals; structural validation supplies that distinction.
12. **Regression tests / protection:** reversed interval duration and plan-validation tests.
13. **Adversarial review findings:** no new wrapper type was introduced because existing canonical validation already provides the structural error boundary.
14. **Remediation iterations:** `caa6a0ee7f1309801e1671e06d38535b28aa2437`.
15. **Residual risks and limitations:** consumers must not use helper duration as a substitute for full plan validation.
16. **Operational or deployment consequences:** no migration; safer diagnostic arithmetic.
17. **Exact evidence:** implementation commit, reversed-window/exclusion tests, strict audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** helper math over potentially malformed intervals must be non-dangerous, while canonical validators retain responsibility for structural rejection.
