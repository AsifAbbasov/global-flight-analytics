# 114. Extractor Review Final Closure

## Baseline

This closure increment is applied after commit
`3cdd9b1532609c872343d00626ba44a9c9855609`.

## Scope

The final increment removes the remaining contract duplication and installs a
permanent closure gate for the complete flight-feature extractor review.

## Centralized contracts

- ICAO24 canonicalization and validation belong to `domain/aircraft`.
- Defensive trajectory copying belongs to `domain/trajectory`.
- Aircraft feature field counts are derived from the current versioned feature
  schema rather than duplicated constants.
- Extractor, aircraft provider, validator, and fingerprint construction consume
  these shared contracts.

## Canonical fingerprint drift protection

Reflection-based tests compare every trajectory, point, segment, and coverage-gap
field against the corresponding canonical fingerprint structure in the same
order. A new, removed, renamed, or reordered evidence field therefore fails the
extractor test suite until the deterministic fingerprint contract is updated
intentionally.

Aircraft metadata retrieval time remains excluded from the deterministic
fingerprint. Source identity and provider version remain included.

## Processing identity

This increment preserves behavior and does not create a new processing
semantics generation. Extractor, composition, feature pipeline, and validator
versions remain at their established version 5 / version 2 contracts.

No PostgreSQL migration is required.

## Verification gates

The permanent `extractorreviewclosureaudit` checks centralized ownership
inside the extractor processing path, absence of the removed local helpers and
constants, fingerprint mirror tests, version stability, documentation status,
and Continuous Integration wiring.

```text
EXTRACTOR_ICAO24_CONTRACT=CENTRALIZED
EXTRACTOR_TRAJECTORY_CLONE_CONTRACT=CENTRALIZED
EXTRACTOR_AIRCRAFT_FIELD_COUNT=SCHEMA_DERIVED
EXTRACTOR_FINGERPRINT_MIRROR_GUARD=ENFORCED
EXTRACTOR_PROCESSING_GENERATION=v5
EXTRACTOR_REVIEW_STATUS=CLOSED
OPEN_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
```

---

## Canonical remediation history

### GFA-MAINT-150 — ICAO24 normalization and validation were duplicated across extractor/provider paths

1. **Finding / symptom.** Extractor and aircraft-provider code owned separate ICAO24 regex/trim/case normalization logic.
2. **Root cause.** A domain identity contract was reimplemented locally by multiple feature-processing packages.
3. **Failure scenario.** One copy changes accepted casing/validation rules while another does not, producing different canonical identity or rejection behavior for the same aircraft.
4. **Impact.** Domain identity semantics can drift across extraction, provider lookup and fingerprint construction.
5. **Severity rationale.** **P3 retrospective.** The observed issue is contract duplication/maintainability; no already-demonstrated production corruption required P1/P2 classification.
6. **Existing guarantees violated.** A core domain identity should have one canonical owner consumed by all processing paths.
7. **Considered solutions.** Keep synchronized helpers; move helper into feature package; centralize ICAO24 semantics in `domain/aircraft`.
8. **Chosen remediation.** Add canonical ICAO24 normalization/validation functions in `domain/aircraft` and migrate provider/extractor/fingerprint consumers.
9. **Why this solution was selected.** The domain package is the stable semantic owner and prevents feature-specific reimplementation.
10. **Rejected alternatives.** Duplicated helpers rely on manual synchronization; feature-local ownership is too narrow for a domain identity used elsewhere.
11. **Trade-offs.** Domain API gains explicit identity helpers and downstream packages depend on them intentionally.
12. **Regression tests / protection.** Domain tests cover valid/invalid/canonical forms; extractor closure audit rejects removed local helpers and verifies centralized ownership.
13. **Adversarial review findings.** Canonicalization and validity are separate operations: callers sometimes need normalized text even while determining validity, so the API preserves both concepts.
14. **Remediation iterations.** Document 112 first canonicalized fingerprint input locally; `bcf7ff3e…` moves the rule to its domain owner without changing processing generation.
15. **Residual risks and limitations.** Other future packages can still duplicate the rule unless closure/architecture review catches them.
16. **Operational or deployment consequences.** No runtime behavior or processing generation change was intended; this is contract ownership consolidation.
17. **Exact evidence.** Implementation commit `bcf7ff3e1a83024ee346c16638de0b389baf7e7a` (`refactor: close extractor review contracts`). Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-MAINT-150=CLOSED`.
19. **Prevention / future guard.** New ICAO24 consumers must use `domain/aircraft` canonicalization/validation instead of defining local regex/case rules.

### GFA-MAINT-151 — defensive trajectory cloning was duplicated inside extractor processing

1. **Finding / symptom.** The extractor maintained its own trajectory clone helper to isolate mutable point/segment/gap slices for builders.
2. **Root cause.** Defensive-copy semantics for the domain object were owned by a consumer instead of the trajectory type itself.
3. **Failure scenario.** The trajectory model gains or changes mutable evidence and the extractor-local clone is not updated, allowing one builder mutation to affect another builder or the caller snapshot.
4. **Impact.** Builder isolation can silently regress as the trajectory model evolves.
5. **Severity rationale.** **P3 retrospective.** This is maintainability/ownership debt with a plausible mutation failure but no documented current data corruption.
6. **Existing guarantees violated.** Defensive copying of a domain aggregate should be centrally owned by the aggregate so all consumers receive the same semantics.
7. **Considered solutions.** Keep extractor helper plus mirror tests; generic deep copy; implement `FlightTrajectory.Clone()` in the domain model.
8. **Chosen remediation.** Add domain-owned `FlightTrajectory.Clone()` preserving nil slices and migrate extractor builder calls.
9. **Why this solution was selected.** It places mutable-slice ownership with the type and provides reusable tested semantics without reflection/runtime serialization.
10. **Rejected alternatives.** Consumer-local copies drift; generic deep copy is slower/less explicit and unnecessary for value-only nested elements.
11. **Trade-offs.** Domain type exposes a cloning method and must evolve it when new mutable fields are added.
12. **Regression tests / protection.** Domain tests prove cloned evidence slices do not share storage and preserve nil-slice semantics; closure audit rejects the removed extractor helper.
13. **Adversarial review findings.** Nil slices are semantically distinguishable from allocated empty slices in some serialization paths, so clone must preserve nil behavior.
14. **Remediation iterations.** Builder isolation existed before; `bcf7ff3e…` centralizes the already-required defensive-copy contract without version change.
15. **Residual risks and limitations.** Future nested reference types require explicit clone expansion; value-only elements are currently safe.
16. **Operational or deployment consequences.** None beyond centralized code ownership; feature output remains unchanged.
17. **Exact evidence.** Implementation commit `bcf7ff3e1a83024ee346c16638de0b389baf7e7a`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-MAINT-151=CLOSED`.
19. **Prevention / future guard.** Mutable additions to `FlightTrajectory` must update and test the domain `Clone()` contract before feature consumers change.

### GFA-CONTRACT-152 — aircraft feature field count was duplicated instead of derived from the versioned schema

1. **Finding / symptom.** Aircraft provider/extractor code maintained hard-coded `6` field-count constants separate from the feature schema.
2. **Root cause.** Evidence counting duplicated schema cardinality rather than querying the schema's canonical group definition.
3. **Failure scenario.** The schema adds/removes an aircraft field but a local constant remains six; availability/completeness evidence uses the wrong denominator.
4. **Impact.** Quality/evidence status can drift from the declared versioned feature schema.
5. **Severity rationale.** **P2 retrospective.** Schema evolution could turn the duplicate constant into incorrect durable quality evidence even though the current value matched at closure time.
6. **Existing guarantees violated.** Quality field counts must derive from the authoritative versioned schema.
7. **Considered solutions.** Keep constant with equality tests; generate constants; use schema-owned `CurrentGroupFieldCount`.
8. **Chosen remediation.** Remove local aircraft count constants and derive counts from `flightfeatures.CurrentGroupFieldCount(FeatureGroupAircraft)`.
9. **Why this solution was selected.** It removes synchronization entirely instead of merely testing duplicated state.
10. **Rejected alternatives.** Equality tests still preserve two owners and require coordinated edits.
11. **Trade-offs.** Provider/extractor depend explicitly on schema metadata for evidence counts.
12. **Regression tests / protection.** Provider/extractor tests consume schema-derived counts; closure audit verifies removed constants and schema ownership.
13. **Adversarial review findings.** Schema-derived counts are correct only if schema requirement metadata itself remains versioned/reviewed; Document 113 already makes that layer authoritative.
14. **Remediation iterations.** Required-vs-optional semantics were corrected in Document 113; `bcf7ff3e…` removes the last duplicated aircraft cardinality constant.
15. **Residual risks and limitations.** Other feature groups must likewise avoid local cardinality constants.
16. **Operational or deployment consequences.** No migration/version bump; current count remains six while future schema changes propagate automatically.
17. **Exact evidence.** Implementation commit `bcf7ff3e1a83024ee346c16638de0b389baf7e7a`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-CONTRACT-152=CLOSED`.
19. **Prevention / future guard.** Evidence-count code must derive group cardinality/requirement classes from the versioned feature schema rather than duplicate numeric constants.

### GFA-DATA-153 — deterministic fingerprint mirrors had no structural guard against trajectory model drift

1. **Finding / symptom.** Canonical fingerprint structs manually mirrored trajectory, point, segment and coverage-gap fields without a test proving the mirror stayed structurally complete and ordered as domain models evolved.
2. **Root cause.** Deterministic hashing intentionally uses explicit canonical structs, but their field parity depended on manual reviewer memory.
3. **Failure scenario.** A new trajectory evidence field is added and affects feature semantics, but the fingerprint mirror is not updated; two materially different inputs hash identically and can replay the same durable record.
4. **Impact.** Silent deterministic identity collisions can occur after model evolution.
5. **Severity rationale.** **P1 retrospective.** Missing output-relevant evidence from a durable fingerprint can silently alias distinct analytical inputs.
6. **Existing guarantees violated.** Canonical deterministic fingerprint structures must evolve intentionally with every relevant domain evidence field.
7. **Considered solutions.** Hash the domain structs directly; manual code review; reflection-based structural mirror tests while retaining explicit canonical serialization.
8. **Chosen remediation.** Add reflection tests comparing every trajectory/point/segment/gap field and order with corresponding canonical fingerprint structs.
9. **Why this solution was selected.** Explicit canonical structs retain stable serialization control while reflection tests make drift a failing test rather than silent omission.
10. **Rejected alternatives.** Direct domain JSON hashing couples identity to tags/format/internal fields; manual review alone is not a permanent guard.
11. **Trade-offs.** Intentional domain changes now require explicit fingerprint-test reconciliation and a decision about processing generation.
12. **Regression tests / protection.** Mirror tests fail on new/removed/renamed/reordered evidence fields; `extractorreviewclosureaudit` requires those tests and CI wiring.
13. **Adversarial review findings.** Not every provenance field belongs in deterministic input identity—e.g. aircraft retrieval time remains intentionally excluded—so mirror enforcement applies to declared canonical evidence structures, not arbitrary all-fields hashing.
14. **Remediation iterations.** Documents 109/112 hardened fingerprint semantics; `bcf7ff3e…` installs permanent structural drift detection without changing the already-correct v5 fingerprint output.
15. **Residual risks and limitations.** A field can be structurally mirrored yet semantically normalized incorrectly; dedicated canonicalization tests remain necessary.
16. **Operational or deployment consequences.** Future domain model changes can fail extractor tests/CI until fingerprint semantics are explicitly reviewed.
17. **Exact evidence.** Implementation commit `bcf7ff3e1a83024ee346c16638de0b389baf7e7a`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-DATA-153=CLOSED`.
19. **Prevention / future guard.** Any domain evidence-field change must trigger fingerprint mirror reconciliation and an explicit processing-version decision before merge.

### GFA-GOV-154 — extractor review closure lacked one permanent CI-enforced reconciliation gate

1. **Finding / symptom.** Individual extractor fixes had targeted tests/audits, but final closure lacked one permanent gate proving centralized ownership, fingerprint drift guards, version stability and zero open/unclassified/deferred findings together.
2. **Root cause.** Remediation accumulated across several increments without a consolidated source-level closure contract.
3. **Failure scenario.** A later change reintroduces a removed local helper or bypasses a mirror guard while narrower tests stay green; documentation still claims extractor review closed.
4. **Impact.** Closure evidence can drift from current source and previously closed review debt can regress unnoticed.
5. **Severity rationale.** **P2 retrospective.** This is a governance/regression-protection gap over data-identity and provenance fixes rather than a new runtime defect itself.
6. **Existing guarantees violated.** Formal review closure must be reproducible by a permanent CI-reachable gate tied to the claimed contracts.
7. **Considered solutions.** Rely on documentation/manual review; expand generic architecture audit; add dedicated `extractorreviewclosureaudit` and CI step.
8. **Chosen remediation.** Add a strict extractor review closure audit covering centralized ownership, removed helpers/constants, fingerprint mirror tests, version stability, documentation status and CI reachability.
9. **Why this solution was selected.** A focused audit captures cross-file closure invariants that are awkward to express as one unit test and prevents documentation-only closure.
10. **Rejected alternatives.** Manual review is non-repeatable; overloading a generic audit weakens discoverability/ownership of extractor-specific contracts.
11. **Trade-offs.** Another permanent source audit must be maintained when intentional extractor architecture changes occur.
12. **Regression tests / protection.** Backend CI executes `extractorreviewclosureaudit -strict`; audit unit/source rules verify closed status and required guards.
13. **Adversarial review findings.** The audit must verify semantic ownership rather than fragile formatter-specific text wherever possible, following lessons from earlier Analytical Core audit hardening.
14. **Remediation iterations.** Added in `bcf7ff3e…`; Document 115 adds a separate composition-review audit for the later composition-specific increment rather than silently expanding historical closure scope.
15. **Residual risks and limitations.** Source audits complement—not replace—runtime/unit/integration tests and cannot prove production load or external source correctness.
16. **Operational or deployment consequences.** Extractor contract regressions become Backend CI merge blockers.
17. **Exact evidence.** Implementation commit `bcf7ff3e1a83024ee346c16638de0b389baf7e7a`; Backend workflow gained the strict closure audit. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-GOV-154=CLOSED`.
19. **Prevention / future guard.** Extractor review claims must remain executable in CI; intentional contract changes must update tests, audit and documentation in the same increment.
