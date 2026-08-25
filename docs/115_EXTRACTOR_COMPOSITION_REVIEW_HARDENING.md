# Extractor Composition Review Hardening

## Scope

This increment re-evaluates the external review of
`internal/features/extractorcomposition` against the current repository state at
baseline `bcf7ff3e1a83024ee346c16638de0b389baf7e7a`.

## Findings already closed before this increment

- Geographic precision already participated in deterministic fingerprint identity.
- Component versions already participated in deterministic fingerprint identity.
- Aircraft not-found policy version already participated in identity.
- `NewExtractor` had already been removed.
- Typed-nil aircraft lookup rejection was already enforced.
- PostgreSQL feature materialization already reached the composition in production code.
- Aircraft cache hits already applied the historical `AsOfTime` gate per request.

## Changes in this increment

- The complete typed processing manifest is persisted in feature provenance.
- Optional aircraft enrichment can be disabled explicitly without a fake lookup.
- Aircraft caching can be disabled explicitly while preserving request coalescing.
- Enrichment mode and cache mode participate in deterministic processing identity.
- Composition exposes the extractor through its behavioral interface.
- Construction is decomposed into validation and component-specific helpers.
- Provider generation advances to version 3.
- Extractor and composition generations advance to version 6.
- Feature processing generation advances to version 6.

## Review disagreements

The cache key remains ICAO24-only by design. Cached current metadata is filtered
against every request's `AsOfTime` after retrieval, so adding `AsOfTime` to the
cache key would duplicate identical source records without improving temporal
safety. A true historical dataset revision or effective-validity interval cannot
be invented by composition when the source repository does not provide one.

The private value-returning composition configuration remains intentionally flat.
It is small, cannot be initialized with public raw literals, and already separates
construction policy from the pure extractor. Additional nested configuration
layers would add ceremony without strengthening an invariant.

The low-level geographic builder retains zero as its default sentinel, while
production composition rejects zero after `DefaultConfig` has resolved an
explicit precision. Production processing therefore never confuses an omitted
precision with an intentional precision.

## Closure evidence

```text
STALE_REVIEW_FINDINGS=CLASSIFIED
PROCESSING_MANIFEST_PERSISTENCE=ENFORCED
OPTIONAL_AIRCRAFT_ENRICHMENT=EXPLICIT
AIRCRAFT_CACHE_DISABLE_MODE=EXPLICIT
CACHE_AS_OF_POLICY=PER_REQUEST
PRODUCTION_REACHABILITY=ENFORCED
TYPED_NIL_LOOKUP_REJECTION=ENFORCED
EXTRACTOR_PROCESSING_GENERATION=v6
AIRCRAFT_PROVIDER_GENERATION=v3
DATABASE_MIGRATION=NOT_REQUIRED
```

No PostgreSQL migration is required because feature snapshots already persist the
versioned feature payload as JSON and processing identity isolation already uses
the existing `processing_version` column.

---

## Canonical remediation history

### GFA-DATA-155 — typed extractor processing manifest participated in identity but was not persisted as feature provenance

1. **Finding / symptom.** Composition policy was hashed into a fingerprint but the complete typed manifest was not stored with the resulting feature payload.
2. **Root cause.** Deterministic identity and human/machine-readable provenance evolved as separate concerns; only the hash crossed into persisted features.
3. **Failure scenario.** A stored snapshot identifies its processing generation/fingerprint but an investigator cannot recover the effective component versions, geographic precision, enrichment mode, cache mode or policy values represented by that hash.
4. **Impact.** Reproducibility and auditability depend on source-code reconstruction rather than durable snapshot evidence.
5. **Severity rationale.** **P2 retrospective.** Identity collision protection already existed, but absence of the typed manifest weakened durable provenance and operational explainability.
6. **Existing guarantees violated.** Durable analytical output should preserve both deterministic identity and the interpretable processing contract that produced it.
7. **Considered solutions.** Store only the hash; duplicate selected config fields; persist the complete typed `ProcessingIdentity` plus its fingerprint.
8. **Chosen remediation.** Add typed `ProcessingIdentity` and `ProcessingIdentityFingerprint` to feature provenance and pass the resolved composition manifest through extractor construction into durable snapshots.
9. **Why this solution was selected.** It makes the already-authoritative processing contract inspectable without creating a parallel loosely typed metadata representation.
10. **Rejected alternatives.** Hash-only evidence is opaque; selected duplicated fields can drift behind the actual identity model.
11. **Trade-offs.** Feature JSON grows and typed manifest changes require processing-generation/version review.
12. **Regression tests / protection.** Extractor composition review tests/audit require manifest persistence and fingerprint consistency; processing generation advances to v6.
13. **Adversarial review findings.** The persisted manifest and identity hash must originate from the same resolved object; independently rebuilding either at persistence time could create mismatched provenance.
14. **Remediation iterations.** Document 109 created deterministic composition identity; `28ff8414…` makes that identity durable provenance and extends it with explicit enrichment/cache modes.
15. **Residual risks and limitations.** The manifest describes repository processing policy, not immutable versions of external datasets unless such revisions are supplied separately.
16. **Operational or deployment consequences.** New v6 feature snapshots carry typed processing provenance; no PostgreSQL migration is needed because versioned feature payload JSON already persists provenance.
17. **Exact evidence.** Implementation commit `28ff8414388d2e81db4e74b83a2fd0c23880d56a` (`fix: harden extractor composition identity`). Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-DATA-155=CLOSED`.
19. **Prevention / future guard.** Deterministic processing manifests must be persisted with durable analytical outputs whenever their hash/version is used for replay identity.

### GFA-CONTRACT-156 — optional aircraft enrichment had no explicit disabled configuration mode

1. **Finding / symptom.** Composition required an aircraft lookup even when a deployment intentionally wanted no aircraft enrichment, encouraging fake/no-op lookup implementations.
2. **Root cause.** Optional runtime capability was modeled as a required dependency plus implicit behavior rather than an explicit policy mode.
3. **Failure scenario.** A caller supplies a fake lookup solely to satisfy construction; processing identity/provenance can imply enrichment infrastructure exists while the fake implementation actually disables it.
4. **Impact.** Configuration intent becomes opaque and optional enrichment behavior is difficult to audit or fingerprint correctly.
5. **Severity rationale.** **P2 retrospective.** This is a production configuration/identity ambiguity affecting feature semantics, though it requires deployments that intentionally disable enrichment.
6. **Existing guarantees violated.** Optional processing capabilities must have explicit typed modes and must not require fake dependencies to express absence.
7. **Considered solutions.** Keep fake/no-op lookup; allow nil lookup to imply disabled; add explicit `AircraftEnrichmentModeEnabled/Disabled` configuration.
8. **Chosen remediation.** Add typed enrichment mode, allow disabled composition without a lookup, validate enabled mode requires a real lookup, and propagate the mode into processing identity.
9. **Why this solution was selected.** It makes intent explicit and deterministic while preserving strict dependency validation for enabled enrichment.
10. **Rejected alternatives.** Fake lookup hides policy in implementation; nil-as-mode recreates zero/nil sentinel ambiguity addressed in Document 111.
11. **Trade-offs.** Configuration gains one explicit mode and validation branch.
12. **Regression tests / protection.** Composition tests cover enabled/disabled validation and unavailable aircraft output; composition review audit requires explicit mode ownership.
13. **Adversarial review findings.** Disabled enrichment must also disable cache configuration consistently; otherwise irrelevant cache settings could still perturb identity.
14. **Remediation iterations.** Earlier configuration required lookup unconditionally; `28ff8414…` introduces explicit disabled enrichment and includes the mode in v6 identity.
15. **Residual risks and limitations.** Disabling enrichment intentionally reduces optional coverage; required completeness remains unaffected by Document 113 semantics.
16. **Operational or deployment consequences.** Deployments can run feature extraction without aircraft lookup using an explicit supported mode rather than a fake provider.
17. **Exact evidence.** Implementation commit `28ff8414388d2e81db4e74b83a2fd0c23880d56a`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-CONTRACT-156=CLOSED`.
19. **Prevention / future guard.** Optional processing capabilities must use explicit typed enabled/disabled policy and define corresponding dependency requirements.

### GFA-CONTRACT-157 — aircraft caching had no explicit disabled mode

1. **Finding / symptom.** Cache TTL configuration could tune caching but did not express a first-class disabled policy while preserving provider request coalescing.
2. **Root cause.** Cache presence was inferred from TTL/default mechanics instead of an independent behavioral mode.
3. **Failure scenario.** A deployment wants every request to re-read source metadata but cannot state that intent without abusing TTL values or custom provider behavior.
4. **Impact.** Cache policy is ambiguous, hard to audit and cannot participate cleanly in deterministic processing identity.
5. **Severity rationale.** **P2 retrospective.** Cache policy can affect which mutable enrichment value is observed, so an implicit/unsupported disable path is a processing-contract gap.
6. **Existing guarantees violated.** Output-sensitive caching policy must be explicit, validated and represented in processing identity.
7. **Considered solutions.** TTL zero means disabled; negative TTL means disabled; typed `CacheModeEnabled/Disabled` independent of TTLs.
8. **Chosen remediation.** Add typed cache mode; enabled mode uses validated TTLs, disabled mode forces TTLs to zero and bypasses cache read/write while retaining in-flight request coalescing.
9. **Why this solution was selected.** It avoids overloaded numeric sentinels and separates caching from duplicate concurrent request suppression.
10. **Rejected alternatives.** TTL sentinels reintroduce the ambiguity removed in Document 111 and conflate two distinct provider behaviors.
11. **Trade-offs.** Disabled mode increases upstream lookup frequency by design.
12. **Regression tests / protection.** Provider tests require two sequential source calls with cache disabled, reject unknown modes, and composition identity tests cover cache mode.
13. **Adversarial review findings.** Request coalescing is not caching: concurrent identical requests may still share one in-flight lookup even when durable/time-based cache is disabled.
14. **Remediation iterations.** `28ff8414…` introduces provider v3 cache mode and integrates it into extractor composition v6.
15. **Residual risks and limitations.** Cache disabling does not create historical data revisions; temporal safety continues to use the per-request gate from Document 110.
16. **Operational or deployment consequences.** Operators can explicitly trade source request volume for cache-free reads without custom providers.
17. **Exact evidence.** Implementation commit `28ff8414388d2e81db4e74b83a2fd0c23880d56a`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-CONTRACT-157=CLOSED`.
19. **Prevention / future guard.** Cache behavior must be controlled by typed policy rather than magic TTL values, with identity/provenance updated when cache semantics can alter observed output.

### GFA-DATA-158 — enrichment and cache modes were absent from deterministic processing identity

1. **Finding / symptom.** Before v6, processing identity captured cache durations/policy versions but did not model the later first-class decisions to disable enrichment or caching.
2. **Root cause.** New behavioral modes did not exist in the original identity model and therefore required an explicit identity extension when introduced.
3. **Failure scenario.** The same trajectory is processed once with enrichment/cache enabled and once disabled; if modes are not identity dimensions, distinct output behavior can target one durable idempotent identity.
4. **Impact.** Snapshots produced under materially different processing policies can collide/replay incorrectly.
5. **Severity rationale.** **P1 retrospective.** Omitted output-sensitive policy from durable deterministic identity can silently alias distinct analytical results.
6. **Existing guarantees violated.** Every processing mode capable of changing feature payload/source observation must participate in deterministic identity.
7. **Considered solutions.** Bump a coarse version only; infer modes from lookup/TTL values; add typed modes directly to `ProcessingIdentity`.
8. **Chosen remediation.** Include aircraft enrichment mode and cache mode in the typed processing manifest/fingerprint and advance extractor/composition/pipeline generations to v6.
9. **Why this solution was selected.** Explicit modes are stable semantic identity dimensions and avoid inference from nullable dependencies or magic numeric values.
10. **Rejected alternatives.** Coarse version alone loses inspectable policy; inferred identity recreates configuration ambiguity.
11. **Trade-offs.** Mode changes intentionally create distinct processing identities/snapshots.
12. **Regression tests / protection.** Tests verify enrichment/cache mode changes alter processing fingerprint; composition review audit requires the fields and version generation.
13. **Adversarial review findings.** Disabled enrichment should normalize unrelated lookup/cache policy so irrelevant configuration does not create meaningless distinct identities.
14. **Remediation iterations.** Document 109 established composition identity; Document 111 made defaults explicit; `28ff8414…` extends the identity for newly explicit runtime modes and persists the manifest.
15. **Residual risks and limitations.** Future processing modes still require manual classification as identity-relevant or operational-only.
16. **Operational or deployment consequences.** v6 snapshots are isolated from earlier generations and from each other when enrichment/cache policy differs.
17. **Exact evidence.** Implementation commit `28ff8414388d2e81db4e74b83a2fd0c23880d56a`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-DATA-158=CLOSED`.
19. **Prevention / future guard.** Any new explicit mode must include a deterministic-identity review and tests proving whether mode changes should or should not alter the processing fingerprint.

### GFA-ARCH-159 — extractor composition exposed a concrete extractor instead of its behavioral contract

1. **Finding / symptom.** Composition exposed the concrete `*extractor.Extractor` where downstream composition only required extractor behavior.
2. **Root cause.** Construction result leaked implementation type rather than the consumer-facing interface already defining feature extraction.
3. **Failure scenario.** Consumers become coupled to concrete methods/fields/construction details, making substitution/testing or internal refactoring unnecessarily broad.
4. **Impact.** Architectural coupling expands the change surface of extractor composition.
5. **Severity rationale.** **P3 retrospective.** This is maintainability/dependency-direction debt without a demonstrated independent production correctness failure.
6. **Existing guarantees violated.** Composition boundaries should expose the narrow behavior consumers need rather than concrete implementation types.
7. **Considered solutions.** Keep concrete pointer; add a wrapper; expose the existing behavioral extractor interface.
8. **Chosen remediation.** Expose the extractor through its behavioral interface while retaining concrete construction internally.
9. **Why this solution was selected.** It narrows dependency direction with minimal abstraction and no behavior change beyond the surrounding v6 semantic work.
10. **Rejected alternatives.** A new wrapper duplicates the existing interface; concrete exposure preserves unnecessary coupling.
11. **Trade-offs.** Consumers cannot depend on implementation-specific methods without explicitly crossing the abstraction boundary.
12. **Regression tests / protection.** Composition/audit source contracts verify behavioral exposure and production reachability.
13. **Adversarial review findings.** The interface must be owned by the consumer behavior and remain small; abstraction solely for abstraction's sake is not a goal.
14. **Remediation iterations.** Closed in `28ff8414…` alongside explicit composition-mode hardening.
15. **Residual risks and limitations.** Internal construction still necessarily knows the concrete implementation; only outward exposure is narrowed.
16. **Operational or deployment consequences.** None; dependency shape changes at compile time.
17. **Exact evidence.** Implementation commit `28ff8414388d2e81db4e74b83a2fd0c23880d56a`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-ARCH-159=CLOSED`.
19. **Prevention / future guard.** Composition results should expose narrow behavioral interfaces unless a concrete type is required by a demonstrated consumer contract.

### GFA-MAINT-160 — extractor composition construction mixed validation and component-specific assembly responsibilities

1. **Finding / symptom.** `New` accumulated policy validation, geographical-builder construction, aircraft-provider construction and extractor assembly in one coordinator.
2. **Root cause.** Successive configuration/policy additions were appended to the same construction function.
3. **Failure scenario.** Further modes or validation rules make construction harder to review and increase the risk of applying a policy in the wrong component/order.
4. **Impact.** Maintenance/review complexity rises around a high-value processing identity boundary.
5. **Severity rationale.** **P3 retrospective.** This is targeted cohesion debt; runtime behavior was already correct after prior fixes.
6. **Existing guarantees violated.** Composition root may coordinate components but component-specific validation/construction should remain locally reviewable as complexity grows.
7. **Considered solutions.** Leave `New` intact; introduce a generic DI framework; extract focused validation/geographical/aircraft construction helpers.
8. **Chosen remediation.** Decompose into `validateConfig` and component-specific construction helpers while preserving one explicit composition root.
9. **Why this solution was selected.** Focused helpers improve reviewability without introducing containers/frameworks or fragmenting ownership across packages.
10. **Rejected alternatives.** A DI framework is disproportionate; leaving the growing function unchanged preserves responsibility concentration.
11. **Trade-offs.** More small private functions exist, but all remain within one composition package and preserve linear orchestration.
12. **Regression tests / protection.** Existing constructor behavior tests plus composition review audit verify policy validation, component construction and version/identity output.
13. **Adversarial review findings.** Decomposition must follow responsibility/failure boundaries, not arbitrary line-count thresholds; the review explicitly rejects ceremony-only nested configuration and mechanical size rules.
14. **Remediation iterations.** Prior increments added identity, temporal policy and explicit defaults; `28ff8414…` performs targeted decomposition once real responsibility concentration existed.
15. **Residual risks and limitations.** Future substantial components may still require further extraction, but no speculative framework is introduced now.
16. **Operational or deployment consequences.** None; construction semantics are preserved while v6 changes come from explicit mode/identity work.
17. **Exact evidence.** Implementation commit `28ff8414388d2e81db4e74b83a2fd0c23880d56a`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-MAINT-160=CLOSED`.
19. **Prevention / future guard.** Composition roots should be decomposed only when demonstrated responsibilities/failure boundaries accumulate, not to satisfy arbitrary function-length/style rules.

### Rejected / retained review observations

The following receive no canonical defect IDs because repository evidence supports the existing dispositions rather than a production failure:

- ICAO24-only aircraft cache key is deliberately retained; temporal acceptance is evaluated per request after cache lookup, and the repository has no historical dataset revision to key by.
- The private value-returning configuration remains intentionally flat; extra nested config layers do not strengthen an invariant.
- The low-level geographic builder may retain its local zero sentinel because production composition resolves explicit defaults and rejects zero before invoking it.

These dispositions are part of the canonical review history and must not be reintroduced as blockers without a new concrete failure mode.
