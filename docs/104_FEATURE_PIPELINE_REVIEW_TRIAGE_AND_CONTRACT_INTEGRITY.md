# Document 104 — Feature Pipeline Review Triage and Contract Integrity

Status: IMPLEMENTED
Baseline commit: 48f274754fa0fbdbe4ed0a2b8f95985f38183629

## 1. Review baseline correction

The supplied review was performed against commit `bb9f351`.

The current repository no longer classifies
`internal/features/featurepipeline` as an unintegrated package. The production
command `cmd/materialize-flight-features` composes and executes the PostgreSQL
feature pipeline. Therefore the old BDUF/YAGNI finding is rejected as stale.

## 2. Findings fixed in this increment

```text
FP-03 Result no longer owns a second FlightFeatures value.
FP-04 Validation status copies are treated as an enforced invariant.
FP-05 Incomplete or contradictory validation reports are rejected, while identical codes remain valid on distinct issue paths.
FP-06 Pipeline depends on a narrow FeatureWriter interface.
FP-08 PostgreSQL Pool and Executor are mutually exclusive.
FP-10 Nil context is rejected instead of replaced.
FP-11 Typed-nil dependencies are rejected.
FP-12 PostgreSQL verifier is executed by Continuous Integration.
FP-13 PostgreSQL composition version is part of the version manifest.
FP-14 Materializer tests and the PostgreSQL verifier consume stored features through the new Result contract.
```

## 3. Deliberately rejected mechanical observations

The following are not correctness rules:

```text
a constructor returning nil together with an error;
the word With in a constructor name;
a fixed forty-line or fifty-line function threshold;
nil returned by Unwrap when the receiver itself is nil.
```

They can be discussed as style preferences, but they must not be recorded as
production blockers without a concrete failure mode.

## 4. Processing identity resolution

The former schema-level blocker was closed by Documents 105 through 107.
Snapshot keys, record identifiers, memory-store indexes, PostgreSQL uniqueness,
reads, migration compatibility and the PostgreSQL verifier now own processing
version.

```text
FP-02_PROCESSING_IDENTITY_STATUS=CLOSED
```

## 5. Durable validation audit resolution

A complete validation report is now part of the durable FlightFeatures payload.
It survives memory and PostgreSQL reads with validator version, validation time,
status, counts and issues intact. Idempotent replay returns the report attached
to the stored record rather than a newer transient report.

Existing rows are explicitly backfilled as `legacy_unavailable`; the system does
not invent historical validator versions, validation times or issues.

```text
FEATURE_PIPELINE_VALIDATION_AUDIT_TRAIL=CLOSED
```

## 6. Composition boundary dispositions

The two composition observations are now formally classified.

`FP-07` is deliberately retained as non-blocking. The package is internal, core
orchestration remains free of PostgreSQL imports, and the isolated
`postgres_composition.go` file owns the canonical construction invariants and
version manifest used by the production materializer and verifier. Moving the
same wiring into command packages would duplicate those invariants without
removing a correctness failure.

`FP-09` is deliberately retained as non-blocking. The internal composition
handle intentionally exposes pipeline and store access required by the
transactional verifier and operational materializer. Validator and extractor
handles remain diagnostic construction evidence inside an internal package, not
a public external application programming interface.

```text
FP-07_COMPOSITION_PLACEMENT=DELIBERATELY_RETAINED_NON_BLOCKING
FP-09_COMPOSITION_HANDLE=DELIBERATELY_RETAINED_NON_BLOCKING
```

## 7. Final review status

Every accepted correctness finding is implemented. Mechanical observations are
rejected with rationale, and both composition observations are deliberately
retained with explicit non-blocking dispositions.

```text
FEATURE_PIPELINE_RELEASE_BLOCKERS=CLOSED
FEATURE_PIPELINE_REVIEW_STATUS=CLOSED
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
```

---

## Canonical remediation history

The canonical owners below correspond only to accepted defects from the Feature
Pipeline review. `FP-07` and `FP-09` remain deliberately retained non-blocking
composition decisions, and the stale BDUF/YAGNI observation plus the mechanical
style observations in Section 3 do not receive synthetic GFA finding IDs.

### GFA-DATA-120 — `FP-03`: successful pipeline result had two independent `FlightFeatures` sources

1. **Finding / symptom.** `featurepipeline.Result` carried both a top-level `Features` value and `Record.Features`, allowing one successful result to expose two copies of the same logical feature payload.
2. **Root cause.** The result contract accumulated an intermediate validated value alongside the normalized value returned by durable storage.
3. **Failure scenario.** Storage normalization changes a field such as ICAO24, but a caller reads the stale top-level copy while another caller reads `Record.Features`.
4. **Impact.** Consumers can publish or verify different feature values for the same successful pipeline execution.
5. **Severity rationale.** **P2 retrospective.** The failure requires a normalization difference, but once present it is a real data-contract inconsistency across successful consumers.
6. **Existing guarantees violated.** A successful pipeline execution must have one canonical stored source of truth.
7. **Considered solutions.** Keep both values with equality assertions; remove `Record`; remove the duplicate top-level value and derive features from the stored record.
8. **Chosen remediation.** Remove the stored top-level `Features` field and expose `Result.Features()` as a defensive clone of `Record.Features`.
9. **Why this solution was selected.** It makes the durable normalized record authoritative without removing the existing ergonomic features accessor.
10. **Rejected alternatives.** Runtime equality assertions preserve duplicate ownership; removing `Record` would discard persistence identity and audit evidence.
11. **Trade-offs.** Callers must use the accessor or record instead of a struct field; this is intentional ownership tightening.
12. **Regression tests / protection.** Tests verify that storage normalization is visible through `Result.Features()` and that materializer/verifier consumers use the stored contract.
13. **Adversarial review findings.** Defensive cloning remains required so callers cannot mutate the record-owned payload through the accessor.
14. **Remediation iterations.** Closed in the contract-hardening increment; later Documents 105–108 extend the stored record with processing identity and durable validation evidence without reintroducing a second feature value.
15. **Residual risks and limitations.** Internal consumers can still misuse `Record.Features` directly, but both access paths now point to the same stored value.
16. **Operational or deployment consequences.** No schema or API deployment change; materializer output now follows the durable normalized record consistently.
17. **Exact evidence.** Implementation commit `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1` (`fix: harden feature pipeline contracts`). Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits, documents and CI evidence.
18. **Final canonical status.** `GFA-DATA-120=CLOSED`.
19. **Prevention / future guard.** New pipeline result fields may not duplicate durable payload ownership without an explicit independent semantic contract.

### GFA-DATA-121 — `FP-04`: copied validation status was not an enforced cross-object invariant

1. **Finding / symptom.** Validation status existed in multiple feature/report locations without one enforced consistency rule.
2. **Root cause.** The pipeline treated copied validation state as convenience data rather than a contract that must agree before persistence/publication.
3. **Failure scenario.** A validator returns a report status that disagrees with the validated feature quality status and the contradictory pair proceeds downstream.
4. **Impact.** Stored or returned feature evidence can claim mutually inconsistent validation outcomes.
5. **Severity rationale.** **P2 retrospective.** The inconsistency is correctness-relevant but requires a contradictory validator/result path rather than ordinary healthy input.
6. **Existing guarantees violated.** Validation status mirrors must describe one decision and therefore agree exactly.
7. **Considered solutions.** Delete one status; silently overwrite one copy; reject contradictory state and preserve the intentional mirror.
8. **Chosen remediation.** Treat status copies as an invariant and validate consistency before accepting the result.
9. **Why this solution was selected.** It retains useful indexed/embedded status evidence while preventing silent disagreement.
10. **Rejected alternatives.** Silent overwrite hides upstream corruption; deleting all mirrors would weaken efficient persistence/query contracts later formalized by migration constraints.
11. **Trade-offs.** More explicit validation code and tests are required whenever report/status models evolve.
12. **Regression tests / protection.** Pipeline and validator tests cover contradictory status rejection; later durable-report storage validation enforces the same mirror.
13. **Adversarial review findings.** The invariant must be checked at both transient pipeline and durable read/write boundaries because persisted data can outlive process validation.
14. **Remediation iterations.** Initial invariant added in `312afe2b…`; Document 108 extends it to the durable validation report and PostgreSQL constraint surface.
15. **Residual risks and limitations.** Legacy rows cannot reconstruct missing historical report detail and are explicitly classified `legacy_unavailable`.
16. **Operational or deployment consequences.** Contradictory validation output now fails closed instead of being persisted.
17. **Exact evidence.** Implementation commit `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1`; durable extension `abd038c10d1d382843dbaefb8b506efeff5fdeda`. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-DATA-121=CLOSED`.
19. **Prevention / future guard.** Any new validation-status mirror must be normalized and equality-checked at persistence and decode boundaries.

### GFA-DATA-122 — `FP-05`: incomplete or contradictory validation reports could be accepted

1. **Finding / symptom.** The pipeline could receive structurally incomplete or internally contradictory validation report evidence.
2. **Root cause.** Report completeness/consistency validation was weaker than the semantic contract later required from accepted valid/limited feature output.
3. **Failure scenario.** A validator reports an accepted status while omitting required validator/time evidence or returns inconsistent counts/issues; the pipeline treats it as accepted.
4. **Impact.** Invalid audit evidence can authorize durable analytical features, undermining trust in downstream feature quality.
5. **Severity rationale.** **P1 retrospective.** Accepted validation is an admission boundary for durable analytical data; malformed evidence could make invalid feature state appear trustworthy.
6. **Existing guarantees violated.** Accepted validation evidence must be complete, internally consistent and path-aware.
7. **Considered solutions.** Trust validators; normalize missing fields; reject malformed reports at the pipeline boundary.
8. **Chosen remediation.** Validate complete report semantics and reject incomplete/contradictory reports while allowing the same issue code on distinct issue paths.
9. **Why this solution was selected.** Fail-closed validation protects durable feature admission without over-constraining legitimate repeated issue codes.
10. **Rejected alternatives.** Fabricating missing evidence creates false provenance; globally deduplicating issue codes loses distinct field-level findings.
11. **Trade-offs.** Custom validators must satisfy the complete report contract explicitly.
12. **Regression tests / protection.** Tests cover incomplete report rejection, contradiction rejection and path-sensitive issue handling; Feature Pipeline contract audit preserves the rule.
13. **Adversarial review findings.** Issue identity cannot be only the code because the same rule can legitimately fail at multiple paths.
14. **Remediation iterations.** Transient contract fixed in `312afe2b…`; durable report preservation in Document 108 prevents accepted evidence from disappearing after storage.
15. **Residual risks and limitations.** Semantic truth of third-party/source data still depends on extractor evidence; this finding protects validation-report integrity, not source accuracy itself.
16. **Operational or deployment consequences.** Malformed validator output now fails the pipeline rather than producing a durable snapshot.
17. **Exact evidence.** Implementation commit `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1`; later durable evidence commit `abd038c10d1d382843dbaefb8b506efeff5fdeda`. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-DATA-122=CLOSED`.
19. **Prevention / future guard.** New validator/report fields require explicit normalization, validation and regression cases before they can authorize persistence.

### GFA-ARCH-123 — `FP-06`: core pipeline depended on the full feature store contract

1. **Finding / symptom.** Core processing required a broad `featurestore.Store` even though the pipeline only needed to persist accepted features.
2. **Root cause.** The composition dependency exposed read/list behavior that the processing algorithm did not own.
3. **Failure scenario.** Store API changes or test doubles must implement unrelated read behavior just to satisfy the pipeline constructor.
4. **Impact.** Unnecessary coupling increases change surface and obscures the pipeline's actual dependency.
5. **Severity rationale.** **P3 retrospective.** This is architecture/maintainability debt; no independent production data corruption path was required to justify remediation.
6. **Existing guarantees violated.** Core orchestration should depend on the narrow behavior it owns.
7. **Considered solutions.** Keep `Store`; create a local writer interface; move all persistence orchestration outside the pipeline.
8. **Chosen remediation.** Introduce narrow `FeatureWriter` with only `Put` and make `Config.Writer` depend on it.
9. **Why this solution was selected.** It improves dependency direction without relocating stable orchestration or duplicating persistence policy.
10. **Rejected alternatives.** Full store retention preserves needless coupling; pushing all writes outward would fragment the successful processing transaction contract.
11. **Trade-offs.** One small local interface is added and composition adapters must satisfy it.
12. **Regression tests / protection.** AST-based contract audit verifies `Config.Writer` ownership independently of formatting.
13. **Adversarial review findings.** The interface belongs with the consumer (pipeline), not the PostgreSQL implementation.
14. **Remediation iterations.** Closed directly in `312afe2b…`; later processing-version changes extend `Config` without widening the writer again.
15. **Residual risks and limitations.** The internal PostgreSQL composition handle intentionally retains wider store access for verifier/materializer use (`FP-09`, deliberately retained).
16. **Operational or deployment consequences.** None; runtime composition continues to use the same feature store implementation.
17. **Exact evidence.** Implementation commit `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1`. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-ARCH-123=CLOSED`.
19. **Prevention / future guard.** Pipeline dependencies should be reviewed from the consumer's required behavior before accepting broader infrastructure interfaces.

### GFA-DB-124 — `FP-08`: PostgreSQL composition accepted ambiguous Pool and Executor ownership

1. **Finding / symptom.** PostgreSQL feature-pipeline construction could be supplied both a pool and an executor without an explicit exclusivity rule.
2. **Root cause.** Two database execution sources were modeled as optional configuration fields but their mutually exclusive ownership was implicit.
3. **Failure scenario.** A caller supplies both sources and composition silently chooses one, making transaction/session ownership dependent on implementation detail.
4. **Impact.** Database execution can occur on an unintended connection/transaction boundary and make verification behavior ambiguous.
5. **Severity rationale.** **P2 retrospective.** Misconfiguration is required, but the resulting database ownership ambiguity can invalidate transactional assumptions.
6. **Existing guarantees violated.** PostgreSQL composition must have exactly one explicit execution source.
7. **Considered solutions.** Prefer executor when both exist; prefer pool; reject both/none; replace fields with a tagged union.
8. **Chosen remediation.** Require exactly one of Pool or Executor and return typed construction errors for missing or ambiguous sources.
9. **Why this solution was selected.** It makes the existing API fail closed without introducing a larger abstraction solely for two construction modes.
10. **Rejected alternatives.** Silent precedence preserves hidden behavior; a new hierarchy/tagged union adds complexity without additional runtime guarantees.
11. **Trade-offs.** Existing misconfigured callers fail at construction rather than receiving implicit precedence.
12. **Regression tests / protection.** Composition tests cover pool-only, executor-only, neither and both; permanent contract audit protects exclusivity.
13. **Adversarial review findings.** Typed-nil database dependencies must be treated as missing, not valid interface values; this is covered with the broader typed-nil finding.
14. **Remediation iterations.** Closed in `312afe2b…`; the same construction boundary later carries processing-version and verifier evidence.
15. **Residual risks and limitations.** Correct transaction semantics still depend on the supplied executor's caller-owned lifecycle.
16. **Operational or deployment consequences.** Invalid PostgreSQL composition now fails during startup/tool construction rather than later database work.
17. **Exact evidence.** Implementation commit `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1`. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-DB-124=CLOSED`.
19. **Prevention / future guard.** Multiple infrastructure ownership modes require explicit exclusivity and typed construction failure semantics.

### GFA-OPS-125 — `FP-10`: nil processing context was silently replaced

1. **Finding / symptom.** Feature processing accepted a nil context instead of requiring caller-owned cancellation/deadline semantics.
2. **Root cause.** Nil was treated as shorthand for an internally created background context.
3. **Failure scenario.** A caller accidentally drops the request/job context and feature extraction, validation or storage continues after cancellation or shutdown.
4. **Impact.** Work can outlive its owner, consume provider/database resources and weaken shutdown predictability.
5. **Severity rationale.** **P2 retrospective.** This is lifecycle/reliability correctness; it requires a caller contract violation but silently converts it into unbounded ownership.
6. **Existing guarantees violated.** Production processing contexts are caller-owned; only bounded cleanup paths may create independent contexts.
7. **Considered solutions.** Keep Background fallback; derive a timeout internally; reject nil.
8. **Chosen remediation.** Return `ErrContextRequired` when `Process` receives nil.
9. **Why this solution was selected.** It exposes the programming error at the boundary and preserves caller cancellation semantics.
10. **Rejected alternatives.** Internal background/timeout contexts cannot reconstruct the caller's cancellation ownership.
11. **Trade-offs.** Tests and callers must pass an explicit context.
12. **Regression tests / protection.** `TestPipelineRejectsNilContext` and permanent contract checks cover the behavior.
13. **Adversarial review findings.** Independent contexts remain appropriate for tightly bounded cleanup only; this processing path is not cleanup.
14. **Remediation iterations.** Closed in `312afe2b…` with no later semantic reversal.
15. **Residual risks and limitations.** A caller may still deliberately pass `context.Background()`; explicit caller ownership is observable, intent is not.
16. **Operational or deployment consequences.** Accidental nil-context calls fail immediately rather than launching detached work.
17. **Exact evidence.** Implementation commit `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1`. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-OPS-125=CLOSED`.
19. **Prevention / future guard.** New processing entry points must reject nil contexts unless they are explicitly documented bounded-cleanup functions.

### GFA-ARCH-126 — `FP-11`: typed-nil dependencies passed interface presence checks

1. **Finding / symptom.** Interface-typed extractor, validator or writer dependencies could contain a nil pointer while the interface itself compared non-nil.
2. **Root cause.** Constructor presence checks relied on interface nil comparison alone.
3. **Failure scenario.** Composition accepts a typed-nil dependency and later invokes a method on it, producing a panic or implementation-specific nil behavior.
4. **Impact.** Invalid composition can survive startup validation and fail during feature processing.
5. **Severity rationale.** **P2 retrospective.** Construction misconfiguration is required, but failure can occur in production runtime and bypass intended typed errors.
6. **Existing guarantees violated.** Required feature-pipeline dependencies must be concretely non-nil at construction.
7. **Considered solutions.** Document caller responsibility; recover from panics; detect nil-capable reflected values during construction.
8. **Chosen remediation.** Add `dependencyMissing` handling for interface-wrapped nil-capable values and reject them with the existing typed required-dependency errors.
9. **Why this solution was selected.** It closes Go's typed-nil interface edge case at the single construction boundary.
10. **Rejected alternatives.** Documentation does not enforce safety; panic recovery is later and less precise.
11. **Trade-offs.** Small reflection use is isolated to construction validation rather than hot-path processing.
12. **Regression tests / protection.** Constructor tests pass typed-nil extractor, validator and writer values and require typed errors.
13. **Adversarial review findings.** Reflection must only call `IsNil` on nil-capable kinds; the helper explicitly switches on those kinds.
14. **Remediation iterations.** Closed in `312afe2b…`; PostgreSQL source exclusivity uses the same fail-fast construction philosophy.
15. **Residual risks and limitations.** Dependencies can still be logically unusable despite being non-nil; behavioral health remains their own contract.
16. **Operational or deployment consequences.** Invalid composition fails at construction rather than mid-processing.
17. **Exact evidence.** Implementation commit `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1`. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-ARCH-126=CLOSED`.
19. **Prevention / future guard.** Constructors accepting interface dependencies must account for typed nils when nil concrete values are possible.

### GFA-GOV-127 — `FP-12`: PostgreSQL feature-pipeline verifier was not part of CI

1. **Finding / symptom.** A transactional PostgreSQL verifier existed but was not a required Continuous Integration execution path.
2. **Root cause.** Repository correctness tests and migration checks did not invoke the production feature-pipeline composition/verifier end to end.
3. **Failure scenario.** A change breaks real PostgreSQL feature materialization while unit tests and unrelated PostgreSQL tests remain green.
4. **Impact.** Merge evidence can claim feature-pipeline correctness without exercising its production database path.
5. **Severity rationale.** **P2 retrospective.** This is a governance/evidence gap that can allow production integration regressions through otherwise green CI.
6. **Existing guarantees violated.** Runtime-reachable persistence paths require executable production-like verification in CI.
7. **Considered solutions.** Rely on unit tests; add mocks; run the existing PostgreSQL verifier in the integration job.
8. **Chosen remediation.** Add `go run ./cmd/verify-postgres-feature-pipeline` to PostgreSQL 16 Integration with production migration/database environment.
9. **Why this solution was selected.** It reuses the actual verifier and database stack rather than constructing another parallel test harness.
10. **Rejected alternatives.** Mock-only evidence cannot prove migrations, SQL, transaction behavior or decoder contracts.
11. **Trade-offs.** CI performs additional PostgreSQL work and may expose fixture/schema drift that unit tests miss — intentionally.
12. **Regression tests / protection.** Backend workflow permanently executes the verifier; later Documents 106–108 demonstrate that this gate detected real regressions.
13. **Adversarial review findings.** CI reachability is meaningful only if the workflow executes after production migrations on the same database service.
14. **Remediation iterations.** Gate added in `312afe2b…`; processing-identity and durable-validation audits were subsequently added around it.
15. **Residual risks and limitations.** The verifier is deterministic integration evidence, not a production load or Neon availability test.
16. **Operational or deployment consequences.** Feature-pipeline database regressions become merge blockers in Backend CI.
17. **Exact evidence.** Implementation commit `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1`; later CI-discovered failures are documented in 106 and 107. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-GOV-127=CLOSED`.
19. **Prevention / future guard.** New production feature persistence semantics must extend the existing verifier and keep its workflow reachability intact.

### GFA-DATA-128 — `FP-13`: PostgreSQL composition version was absent from the version manifest

1. **Finding / symptom.** Feature-pipeline component versions did not include a distinct composition version identifying the concrete wiring contract.
2. **Root cause.** Version evidence covered pipeline/extractor/validator/store components but not the assembly semantics that selected and connected them.
3. **Failure scenario.** Two materially different composition contracts can report the same component-version manifest, weakening reproducibility and audit interpretation.
4. **Impact.** Operators and stored evidence cannot distinguish wiring revisions from component revisions alone.
5. **Severity rationale.** **P2 retrospective.** This is provenance/reproducibility debt affecting trustworthy processing identity, not immediate arithmetic correctness.
6. **Existing guarantees violated.** Materialized feature provenance should identify all processing contracts that can materially alter output or persistence behavior.
7. **Considered solutions.** Reuse pipeline version; omit composition identity; add an explicit composition version to the manifest.
8. **Chosen remediation.** Add stable in-memory/PostgreSQL composition version ownership and expose it through `CurrentVersions()`.
9. **Why this solution was selected.** It separates orchestration identity from algorithm identity while retaining the existing manifest structure.
10. **Rejected alternatives.** Reusing pipeline version conflates two independently evolving contracts; omission leaves an audit blind spot.
11. **Trade-offs.** Composition changes that affect the contract require deliberate version maintenance.
12. **Regression tests / protection.** Version-manifest tests and contract audit require the composition version.
13. **Adversarial review findings.** A version string only has value if it is stable, canonical and updated for semantic—not cosmetic—composition changes.
14. **Remediation iterations.** Composition version added in `312afe2b…`; Document 105 then adds processing version to durable snapshot identity.
15. **Residual risks and limitations.** The manifest does not encode every build SHA; repository/build provenance remains a separate layer.
16. **Operational or deployment consequences.** Diagnostic/version output gains explicit composition identity.
17. **Exact evidence.** Implementation commit `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1`. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-DATA-128=CLOSED`.
19. **Prevention / future guard.** New material processing layers must declare whether they are part of algorithm, schema, processing or composition identity.

### GFA-TEST-129 — `FP-14`: materializer/verifier consumers did not consistently use the stored-result contract

1. **Finding / symptom.** Operational/test consumers still read the former top-level `Result.Features` rather than the record-owned stored features.
2. **Root cause.** Consumer code preserved the pre-`FP-03` result shape after canonical ownership moved to `Record.Features`.
3. **Failure scenario.** Core result ownership is fixed but materializer/verifier tests continue validating an obsolete transient value, leaving production normalization unproved.
4. **Impact.** Regression evidence can pass while failing to assert the actual durable value consumed by production.
5. **Severity rationale.** **P2 retrospective.** This is a test/operational contract gap that could mask durable normalization defects.
6. **Existing guarantees violated.** Acceptance evidence must exercise the same canonical successful value contract as production consumers.
7. **Considered solutions.** Keep compatibility field for tests; modify consumers to use `Result.Features()`/record; test both indefinitely.
8. **Chosen remediation.** Update materializer tests and PostgreSQL verifier to consume the stored-result contract.
9. **Why this solution was selected.** It aligns evidence with the one source of truth established by `FP-03`.
10. **Rejected alternatives.** Keeping duplicate compatibility evidence would preserve the ambiguity the remediation removed.
11. **Trade-offs.** Tests are coupled intentionally to the canonical stored-result API rather than the removed field.
12. **Regression tests / protection.** Materializer and PostgreSQL verifier checks use `Result.Features()` and compare durable records.
13. **Adversarial review findings.** A regression test that only compiles the new API is insufficient; it must prove normalized stored values are what consumers observe.
14. **Remediation iterations.** Consumer migration completed in `312afe2b…`; later verifier additions for processing identity and validation audit build on the same record-owned contract.
15. **Residual risks and limitations.** Presentation/reporting code outside feature materialization may transform values after this boundary; those transformations require their own contracts.
16. **Operational or deployment consequences.** Materializer diagnostic output reflects normalized durable features.
17. **Exact evidence.** Implementation commit `312afe2b9ddcc05da0c2068e50c05e0741a7a1c1`. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-TEST-129=CLOSED`.
19. **Prevention / future guard.** When canonical ownership moves, production verifiers and tests must be migrated in the same increment rather than validating compatibility copies.
