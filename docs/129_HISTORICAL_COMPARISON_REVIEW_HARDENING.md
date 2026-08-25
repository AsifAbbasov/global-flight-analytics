# Historical Comparison Review Hardening

Status: closed

## Scope

This increment hardens `apps/api/internal/historicalintelligence/historicalcomparison` as the atomic domain boundary for adjacent-period Historical Intelligence comparisons.

## Accepted findings

- The previous period could be partial while the current period was complete, producing a mathematically valid but domain-misleading percentage from incomparable coverage.
- Direct `Attach` returned a derived result whose builder version, fingerprint, source union, and latest source time described only the current period until the materializer performed a later optional repair.
- `Attach` mixed source validation, compatibility checks, value selection, arithmetic, result construction, provenance, and final validation.
- Scope identity depended on `reflect.DeepEqual`, so future non-semantic fields could silently alter comparison behavior.
- The package exported the generic `Values` helper without an external use case.
- Percentage overflow was detected only by final contract validation and was misclassified as an invalid current source result.
- Regression coverage did not prove mismatch classification, coverage comparability, all aggregation selectors, downward and flat direction, direct provenance, arithmetic overflow, nested comparison rejection, or previous-period limitations.

## Corrected contracts

- Current and previous series must have identical top-level status, point count, bucket duration, bucket status, and per-bucket coverage ratio within the Historical Contract ratio tolerance.
- Complete-versus-partial and differently partial series fail closed with `ErrCoverageMismatch`; a coverage difference cannot masquerade as metric growth.
- Matched partial periods remain comparable. The current-series confidence stays contract-consistent, while a comparison-scoped quality limitation records both statuses, confidence scores, sample counts, previous-period limitations, and the matched-partial constraint.
- `Attach` now constructs a standalone provenance atomically. Builder identity includes Historical Comparison plus both source builders; semantic fingerprinting binds both complete source results; sources are merged and sorted; latest source time is the later period evidence time.
- Nested comparison attachment is rejected so a derived result cannot be treated as a raw source period.
- Scope comparison uses the explicit domain method `Scope.Equal`.
- Value selection uses a focused aggregation selector registry and the helper type is package-private.
- Percentage calculation preserves full finite `float64` domain precision, treats a zero previous value as explicitly undefined, and rejects non-finite arithmetic with `ErrComparisonArithmeticInvalid`.
- Final derived-result validation uses `ErrComparisonResultInvalid`, keeping source validity and comparison construction failures distinct.

## Deliberately retained contracts

- The module does not process money. Historical metrics are exact counts within the IEEE-754 exact-integer boundary, bounded ratios, or continuous aviation measurements. Replacing the project-wide `float64` transport with a decimal library would be a breaking cross-contract migration without a domain benefit here.
- Percentage values are not rounded in the domain. The finite full-precision value is stored; presentation layers may apply display rounding without changing analytical identity.
- `PercentageChange == nil` remains the explicit contract representation for an undefined percentage when the previous value is zero. Replacing it with a new object would break persistence and API schemas without improving the mathematical state.
- `Summary.Average` is the unweighted temporal mean of represented bucket values. `Summary.Median` is the temporal median of represented bucket values. Neither field claims to be an observation-weighted mean or a raw-observation median.
- The nil receiver behavior of `ResultValidationError.Unwrap` remains idiomatic: a nil error wrapper has no underlying error.
- A finite enumeration switch is not inherently an Open/Closed Principle violation. The comparison module nevertheless now uses a selector registry to keep the orchestration path focused.

```text
COMPARISON_COVERAGE_PROFILE_MATCH=ENFORCED
PREVIOUS_PERIOD_QUALITY=BOUND
COMPARISON_PROVENANCE=ATOMIC
COMPARISON_FINGERPRINT_BOTH_PERIODS=BOUND
COMPARISON_ARITHMETIC_FINITE=ENFORCED
SCOPE_EQUALITY=EXPLICIT
PERCENTAGE_ZERO_BASE=UNDEFINED_OPTIONAL
TEMPORAL_SUMMARY_SEMANTICS=DOCUMENTED
HISTORICAL_COMPARISON_ENGINEERING_REMEDIATION=IMPLEMENTED
```

## Permanent verification

`apps/api/tools/historicalcomparisonreviewaudit` protects the version-two comparison boundary, explicit scope equality, coverage-profile compatibility, explicit two-period quality evidence, atomic provenance, both-period fingerprint identity, finite percentage arithmetic, package-private helper surface, regression tests, and this review record in Backend Continuous Integration.

## Formal closure evidence

The Historical Comparison engineering remediation was committed and validated
before this administrative closure:

```text
ENGINEERING_BASELINE_COMMIT=d60af19d87fbbb234bab72fb4389a8d503ae06b9
ENGINEERING_REMEDIATION_COMMIT=21734b85b9f50ae717dca031c798866161895989
ENGINEERING_GITHUB_ACTIONS_RUN=30341011740
Backend Quality=SUCCESS
Backend Quality Job=90216363225
PostgreSQL 16 Integration=SUCCESS
PostgreSQL 16 Integration Job=90216363216
Backend Race Safety=SUCCESS
Backend Race Safety Job=90216363189
Backend Container=SUCCESS
Backend Container Job=90216611574
```

All accepted findings are implemented. Deliberately retained contracts preserve
their documented rationale, and no Historical Comparison review item remains
open, unclassified, or deferred.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
HISTORICAL_COMPARISON_ENGINEERING_DEBT=CLOSED
HISTORICAL_COMPARISON_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
HISTORICAL_COMPARISON_REVIEW_STATUS=CLOSED
```

## Canonical remediation history

Historical reviewer identities and review-comment chronology are not reconstructed. The records below are reconstructed from repository source/tests, accepted/retained review dispositions, remediation commit `21734b85b9f50ae717dca031c798866161895989`, permanent audit, and exact Backend CI run `30341011740`. Severity labels are retrospective.

### GFA-DATA-260 — Historical periods with incompatible coverage could be compared as ordinary growth
1. **Finding / symptom:** a complete current period could be compared with a partial previous period, or two differently partial periods could yield an ordinary percentage change.
2. **Root cause:** source validity was checked without proving that the compared temporal evidence profiles were equivalent.
3. **Failure scenario:** missing observations in one period appear as real growth/decline in another.
4. **Impact:** mathematically finite but domain-misleading trend output.
5. **Severity rationale:** **P1 retrospective** because comparison could convert evidence loss into a business/traffic signal.
6. **Existing guarantees violated:** adjacent-period comparison requires like-for-like temporal coverage.
7. **Considered solutions:** compare any valid summaries, scale by confidence, reject every partial period, or require matching coverage profiles.
8. **Chosen remediation:** current and previous series must match top-level status, point count, bucket duration/status and per-bucket coverage ratio within contract tolerance; mismatches return `ErrCoverageMismatch`.
9. **Why selected:** preserves honestly matched partial comparisons while blocking incomparable evidence.
10. **Rejected alternatives:** confidence-only adjustment and blanket rejection of all partial evidence.
11. **Trade-offs:** fewer comparisons are available when coverage differs, even if arithmetic could be computed.
12. **Regression tests / protection:** complete-vs-partial, differently-partial, bucket/status/coverage mismatch tests and `historicalcomparisonreviewaudit`.
13. **Adversarial review findings:** matched partial periods remain valid when a comparison-scoped limitation records the constraint.
14. **Remediation iterations:** `21734b85b9f50ae717dca031c798866161895989`.
15. **Residual risks and limitations:** matching coverage does not guarantee identical source quality; confidence/limitations carry remaining quality evidence.
16. **Operational or deployment consequences:** incompatible comparisons fail closed before persistence.
17. **Exact evidence:** remediation commit, comparison coverage tests, GitHub Actions run `30341011740` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** derived period-over-period metrics must prove temporal evidence comparability independently of arithmetic validity.

### GFA-DATA-261 — Direct Historical Comparison output had incomplete two-period provenance and fingerprint identity
1. **Finding / symptom:** `Attach` initially returned builder identity, fingerprint, sources and latest source time based mainly on the current period; a later materializer repair could patch them.
2. **Root cause:** provenance ownership was split between the comparison domain operation and an optional orchestration follow-up.
3. **Failure scenario:** direct callers receive a derived result whose identity does not bind the previous period even though the percentage depends on it.
4. **Impact:** non-atomic provenance, possible fingerprint collisions and caller-dependent result truth.
5. **Severity rationale:** **P1 retrospective** because immutable derived identity could omit one complete source operand.
6. **Existing guarantees violated:** a derived result must be self-contained and bind every source that affects its value.
7. **Considered solutions:** require all callers to run a repair helper, keep materializer ownership, or make `Attach` construct complete provenance atomically.
8. **Chosen remediation:** `Attach` owns version-two comparison identity: both source builders/results are fingerprinted, sources merged/sorted, latest source time is the later source evidence time.
9. **Why selected:** every caller receives the same valid object without optional repair.
10. **Rejected alternatives:** post-construction patching and current-period-only identity.
11. **Trade-offs:** Comparison must understand and validate more provenance fields at construction.
12. **Regression tests / protection:** direct-call provenance, both-period fingerprint, source union/latest-time and nested-comparison tests plus audit.
13. **Adversarial review findings:** Historical Materialization remains an orchestrator but no longer owns Comparison provenance repair.
14. **Remediation iterations:** `21734b85b9f50ae717dca031c798866161895989`.
15. **Residual risks and limitations:** future provenance fields that affect identity require versioned comparison updates.
16. **Operational or deployment consequences:** direct and materialized comparison results now share one canonical identity.
17. **Exact evidence:** remediation commit, provenance/fingerprint tests, run `30341011740` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** derived-domain constructors must return fully valid provenance; orchestration layers must not be required to repair domain identity.

### GFA-MAINT-262 — Historical Comparison `Attach` concentrated validation, arithmetic, selection, provenance and result construction
1. **Finding / symptom:** one function mixed source validation, compatibility checks, aggregation selection, percentage arithmetic, result creation, provenance and final validation.
2. **Root cause:** the initial implementation grew around one public operation without separating policy responsibilities.
3. **Failure scenario:** modifying one aggregation or validation rule risks changing unrelated comparison or provenance behavior.
4. **Impact:** elevated regression risk at a derived analytical trust boundary.
5. **Severity rationale:** **P3 retrospective** because correctness defects are separately identified; this is structural maintainability debt.
6. **Existing guarantees violated:** focused policy ownership and independently testable analytical steps.
7. **Considered solutions:** keep the monolith, generic reflection-based pipeline, or focused internal helpers/selector registry.
8. **Chosen remediation:** split compatibility, selector, arithmetic, result/provenance and validation responsibilities; use focused aggregation selector registry.
9. **Why selected:** reduces coupling without broad cross-domain framework code.
10. **Rejected alternatives:** refactoring solely to satisfy function-length metrics and treating every finite switch as an architecture defect.
11. **Trade-offs:** more internal helpers and explicit sequencing.
12. **Regression tests / protection:** selector coverage, provenance, arithmetic and validation tests plus permanent audit.
13. **Adversarial review findings:** a finite enumeration remains acceptable when it represents a versioned policy; registry use is justified by focused orchestration here.
14. **Remediation iterations:** `21734b85b9f50ae717dca031c798866161895989`.
15. **Residual risks and limitations:** Comparison remains one domain package and still intentionally coordinates its internal policies.
16. **Operational or deployment consequences:** none; internal maintainability hardening.
17. **Exact evidence:** remediation commit, permanent audit, run `30341011740` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** new comparison policies should extend focused helpers rather than add unrelated responsibilities to `Attach`.

### GFA-CONTRACT-263 — Historical scope identity depended on structural reflection equality
1. **Finding / symptom:** scope compatibility used `reflect.DeepEqual` on the complete struct.
2. **Root cause:** semantic equality was not owned explicitly by the domain type.
3. **Failure scenario:** adding a non-semantic/cache/display field to `Scope` unexpectedly makes historically equivalent scopes compare unequal, or a representation-only difference changes comparison behavior.
4. **Impact:** hidden contract drift from ordinary struct evolution.
5. **Severity rationale:** **P2 retrospective** because current values could be valid while future maintenance silently changes compatibility semantics.
6. **Existing guarantees violated:** domain identity must depend only on declared semantic fields.
7. **Considered solutions:** reflection, manual comparison inside Comparison, serialized equality, or domain-owned `Scope.Equal`.
8. **Chosen remediation:** use explicit `Scope.Equal` method owned by Historical Contract.
9. **Why selected:** equality evolves deliberately with the scope contract.
10. **Rejected alternatives:** reflection and generic serialization equality.
11. **Trade-offs:** semantic field additions require explicit equality-method review.
12. **Regression tests / protection:** scope equality/mismatch tests and comparison audit.
13. **Adversarial review findings:** direct navigation through adjacent domain values remains acceptable; equality ownership is the real issue.
14. **Remediation iterations:** `21734b85b9f50ae717dca031c798866161895989`.
15. **Residual risks and limitations:** correctness depends on future scope fields being classified correctly as semantic or non-semantic.
16. **Operational or deployment consequences:** stable comparison compatibility across representation-only changes.
17. **Exact evidence:** remediation commit, scope tests, run `30341011740` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** identity-bearing domain structs must expose explicit semantic equality instead of relying on reflection.

### GFA-MAINT-264 — Historical Comparison exported an internal generic value-selection helper
1. **Finding / symptom:** the package exposed `Values` despite no external contract requiring it.
2. **Root cause:** an implementation helper leaked into the public package surface during early construction.
3. **Failure scenario:** outside callers couple to a generic helper, making future aggregation-policy changes harder and creating a second quasi-public API.
4. **Impact:** unnecessary compatibility burden and policy ownership ambiguity.
5. **Severity rationale:** **P3 retrospective** because no incorrect analytical output is independently demonstrated.
6. **Existing guarantees violated:** minimal public surface and clear domain operation ownership.
7. **Considered solutions:** retain/export and document, move to shared utility, or make the helper package-private.
8. **Chosen remediation:** value-selection helper becomes private; public behavior stays through Comparison operations.
9. **Why selected:** removes unsupported API surface without affecting consumers.
10. **Rejected alternatives:** shared generic abstraction without a cross-package use case.
11. **Trade-offs:** tests requiring the helper exercise behavior through package internals/public operations instead.
12. **Regression tests / protection:** strict audit verifies package-private helper surface and selector behavior.
13. **Adversarial review findings:** percentage/summary types remain public where they are actual domain contract fields.
14. **Remediation iterations:** `21734b85b9f50ae717dca031c798866161895989`.
15. **Residual risks and limitations:** future consumers needing raw value selection should define a real use case before API expansion.
16. **Operational or deployment consequences:** none.
17. **Exact evidence:** remediation commit, source-surface audit, run `30341011740` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** internal helpers must remain unexported until an external behavior contract exists.

### GFA-DATA-265 — Non-finite percentage arithmetic was detected late and misclassified as invalid source evidence
1. **Finding / symptom:** overflow/non-finite comparison arithmetic could reach final result validation and surface as an error against the current source result.
2. **Root cause:** percentage calculation had no local finite-result guard or dedicated error classification.
3. **Failure scenario:** valid finite source periods produce overflow in `(current-previous)/previous`, but diagnostics blame source validity rather than derived arithmetic.
4. **Impact:** fail-closed behavior existed late, but the failure cause and trust boundary were wrong, complicating debugging and policy handling.
5. **Severity rationale:** **P2 retrospective** because invalid output was ultimately rejected, but error semantics and local arithmetic invariants were incomplete.
6. **Existing guarantees violated:** derived arithmetic must fail at its own boundary with a typed cause.
7. **Considered solutions:** rely on final validator, clamp/round percentages, decimal arithmetic, or validate finite arithmetic directly.
8. **Chosen remediation:** preserve full finite `float64` precision, treat zero previous value as undefined optional percentage, and return `ErrComparisonArithmeticInvalid` for non-finite results.
9. **Why selected:** corrects classification without changing the non-monetary numeric model.
10. **Rejected alternatives:** clamping, domain rounding and decimal-library migration.
11. **Trade-offs:** extreme finite source values can now fail comparison explicitly instead of producing a percentage.
12. **Regression tests / protection:** overflow, zero-base, upward/downward/flat percentage tests and audit.
13. **Adversarial review findings:** `PercentageChange == nil` for zero previous value remains intentional and is not a defect.
14. **Remediation iterations:** `21734b85b9f50ae717dca031c798866161895989`.
15. **Residual risks and limitations:** binary64 precision remains appropriate for these aviation metrics and is not presentation-rounded in the domain.
16. **Operational or deployment consequences:** clearer typed errors for derived arithmetic failures.
17. **Exact evidence:** remediation commit, arithmetic tests, run `30341011740` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every derived numeric operation must validate finiteness and classify failures before constructing a domain result.

### GFA-TEST-266 — Historical Comparison regression coverage did not protect its critical invariants
1. **Finding / symptom:** tests did not prove coverage-mismatch classification, all aggregation selectors, downward/flat direction, direct provenance, arithmetic overflow, nested comparison rejection or previous-period limitation handling.
2. **Root cause:** tests focused on happy-path comparison rather than adversarial contract boundaries.
3. **Failure scenario:** a refactor reopens a fixed comparison defect while generic success tests remain green.
4. **Impact:** high-value remediation lacked durable regression evidence.
5. **Severity rationale:** **P2 retrospective** because this was a verification gap around a production analytical trust boundary.
6. **Existing guarantees violated:** closed correctness findings require permanent targeted regression protection.
7. **Considered solutions:** rely on final contract tests, manual review, or add focused adversarial tests plus strict CI audit.
8. **Chosen remediation:** add focused mismatch/selector/direction/provenance/arithmetic/nesting/limitation tests and permanent `historicalcomparisonreviewaudit`.
9. **Why selected:** converts the remediation into executable repository policy.
10. **Rejected alternatives:** documentation-only closure and broad tests that do not exercise the exact failure modes.
11. **Trade-offs:** larger test/audit surface to maintain alongside deliberate version changes.
12. **Regression tests / protection:** the finding is closed by the targeted tests themselves and the permanent Backend CI audit.
13. **Adversarial review findings:** retained numeric/optional contracts are tested as intentional behavior rather than reclassified as defects.
14. **Remediation iterations:** `21734b85b9f50ae717dca031c798866161895989`.
15. **Residual risks and limitations:** tests cannot prove future unmodeled relationships; new contract fields require new adversarial cases.
16. **Operational or deployment consequences:** no runtime change beyond the already-remediated behavior; stronger CI gate.
17. **Exact evidence:** remediation commit, `historicalcomparisonreviewaudit`, exact Backend CI run `30341011740` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every future Comparison remediation must add a regression case matching its concrete historical failure scenario.
