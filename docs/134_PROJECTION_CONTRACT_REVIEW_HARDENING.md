# Projection Contract Review Hardening

Status: closed

## Scope

This increment hardens:

```text
apps/api/internal/projectionintelligence/projectioncontract
apps/api/internal/domain/confidence
```

It does not rewrite Projection Intelligence producers or introduce new prediction
methods, machine-learning calibration, operational aviation claims, database
migrations, or public HTTP response changes.

## Accepted findings

- Horizon duration and point ordering did not prove that every point occupied the
  exact slot defined by `AsOfTime + n*Step`.
- `limited` required points but did not require evidence explaining why the result
  was not `complete`.
- Positive confidence did not require reasons, reason contributions were only
  checked for finiteness, duplicate reason codes were accepted, and the result
  confidence could exceed mandatory point or Estimated Arrival confidence.
- Projection Contract duplicated the shared ordinal confidence vocabulary.
- Usable input fingerprints were checked only for non-empty text.
- ICAO24 was not checked as a six-character hexadecimal identifier.
- Estimated Arrival accepted digits in a four-character airport location
  indicator.
- Observed, openly sourced, derived, and estimated provenance did not consistently
  require source and observation-basis evidence.
- Retrieval time could precede observation time.
- Duplicate provenance inputs, limitations, explanations, and confidence reasons
  were accepted.
- The aggregate had no public `Result.Validate` method or typed validation error.
- Regression tests did not protect these cross-field contracts.
- Several Projection Intelligence test fixtures used non-hexadecimal ICAO24
  placeholders or incomplete confidence and provenance evidence.

## Corrected contracts

- Projection Contract advances to `projection-intelligence-contract-v2`; the
  serialized payload schema remains `projection-intelligence-v1` because no field
  layout is changed.
- Validation advances to
  `projection-intelligence-contract-validation-v2` and returns deterministic
  issue ordering.
- Horizon duration must be exactly divisible by `Step`.
- Projection points must form a zero-based, contiguous prefix of the exact horizon
  grid. A `complete` result must populate every grid slot and end exactly at
  `Horizon.EndTime`.
- A `limited` result may still cover the full effective horizon, because horizon
  truncation or unavailable altitude can limit evidence without creating a
  temporal gap. It must, however, carry a limitation beyond generic method
  assumptions and the research-only guard.
- Positive confidence requires at least one normalized reason. Contributions must
  be finite and between negative one and one, and reason codes must be unique.
- Overall confidence cannot exceed the weakest mandatory projection point or
  Estimated Arrival confidence.
- `projectioncontract.ConfidenceLevel` becomes an alias of the shared
  `domain/confidence.Level`. The shared value object now exposes `IsKnown` so
  source compatibility is retained.
- Present fingerprints must use `sha256:` followed by sixty-four lowercase
  hexadecimal characters.
- Present ICAO24 values must contain exactly six hexadecimal characters.
- Estimated Arrival airport identifiers must be normalized four-letter ICAO
  location indicators.
- Observed, openly sourced, derived, and estimated inputs require normalized source
  names and observation or analytical-basis times.
- Present retrieval time must not precede observation time or exceed
  `GeneratedAt`.
- Provenance input names must be non-empty, trimmed, and unique. They remain
  producer-owned identifiers and may carry qualified suffixes such as
  `historical_neighbor:<trajectory-id>`. Confidence reason codes, explanation
  codes, and limitation scope/code pairs must be normalized and unique.
- `Result.Validate` returns a typed error containing cloned validation issues.
- Existing producer tests now use valid hexadecimal ICAO24 fixtures and complete
  fallback confidence and provenance evidence.
- Legacy non-increasing point-time violations continue to emit
  `point_time_invalid` in addition to the more specific grid issue.

## Qualified or rejected findings

- `limited` is not defined as an incomplete time series. The production kinematic
  baseline legitimately returns a fully populated effective horizon while marking
  the result limited when altitude is unavailable or the originally requested
  horizon was truncated. The corrected contract therefore requires explicit
  limiting evidence instead of forbidding full temporal coverage.
- A universal mapping such as `0.01 = low` is not introduced. Projection methods
  own explicit medium and high thresholds through validated versioned
  configuration. The contract owns the ordinal vocabulary and zero-versus-positive
  consistency, not one hidden global score policy.
- Confidence remains `float64`. Replacing it with fixed-point storage would be a
  payload and producer migration, not a validator correction. Determinism is
  protected through finite range checks, bounded contributions, evidence bounds,
  and producer fingerprints.
- Public structures are retained. This internal package is a versioned data
  contract used by multiple producers and transport adapters. Hiding every field
  behind constructors would cause a broad migration while still not preventing
  mutation after copying. Producers already validate before returning; the new
  `Result.Validate` method makes the boundary explicit.
- Pointer fields for altitude, vertical uncertainty, and Estimated Arrival are
  retained as idiomatic presence semantics. Zero altitude is a valid value, so a
  pointer distinguishes absence without inventing sentinel values.
- The package `Version` and payload `SchemaVersionV1` remain separate: one identifies
  the implementation contract generation and one identifies serialized payload
  compatibility.
- `ValidationSeverityWarning` remains reserved public validation vocabulary. No
  warning is invented merely to justify the symbol.
- The contract does not impose arbitrary maximum altitude, speed, uncertainty, or
  Estimated Arrival duration. Those limits depend on the selected prediction
  method and are already versioned in producer configuration.
- Cross-point uncertainty and confidence are not forced to be monotonic. Historical
  neighbor continuation can legitimately obtain different support and dispersion
  at later grid slots. The generic contract instead enforces exact grid identity
  and bounds overall confidence by the weakest mandatory evidence.
- Arrival destination equality cannot be reconstructed from `Result` alone because
  the route contract identity is not persisted in this payload. The Estimated
  Arrival producer already validates the route destination before attaching the
  estimate. Persisting route identity would require a separate schema revision.
- The input fingerprint is not recomputed from `Provenance.Inputs`: producer
  fingerprints intentionally bind richer method configuration and source state
  than the summarized provenance list exposes. This increment enforces its
  cryptographic format without weakening producer-owned fingerprint semantics.

```text
PROJECTION_GRID_ALIGNMENT=ENFORCED
LIMITED_STATUS_EVIDENCE=REQUIRED
POSITIVE_CONFIDENCE_REASONS=REQUIRED
CONFIDENCE_CONTRIBUTIONS=BOUNDED
RESULT_CONFIDENCE_EVIDENCE_BOUND=ENFORCED
SHARED_CONFIDENCE_VOCABULARY=USED
PROJECTION_FINGERPRINT_FORMAT=SHA256
PROVENANCE_CHRONOLOGY=ENFORCED
PROVENANCE_DUPLICATES=REJECTED
ICAO24_FORMAT=HEX_24_BIT
AIRPORT_ICAO_FORMAT=FOUR_LETTERS
DOMAIN_OPTIONAL_POINTERS=RETAINED
PUBLIC_RESULT_STRUCT=RETAINED_WITH_VALIDATE
FIXED_POINT_CONFIDENCE_SCORE=REJECTED_FOR_NOW
PHYSICAL_POLICY_LIMITS=PRODUCER_OWNED
PROJECTION_CONTRACT_ENGINEERING_REMEDIATION=IMPLEMENTED
```

## Permanent verification

`apps/api/tools/projectioncontractreviewaudit` protects the shared confidence
vocabulary, contract and validation versions, exact horizon grid, limited-status
evidence, confidence reason integrity, weakest-evidence reconciliation, SHA-256
fingerprints, ICAO identifiers, provenance chronology and uniqueness,
`Result.Validate`, regression tests, this review record, and Backend Continuous
Integration enforcement.

## Formal closure evidence

The Projection Contract engineering remediation was committed and validated before
this administrative closure:

```text
ENGINEERING_BASELINE_COMMIT=405b141c431b7cd3e8b8150e88ac238924992e15
ENGINEERING_REMEDIATION_COMMIT=964556d0ca8a1ce9aa74c37c55961cdd006b3de8
ENGINEERING_GITHUB_ACTIONS_RUN=30396070318
Backend Quality=SUCCESS
Backend Quality Job=90399157528
PostgreSQL 16 Integration=SUCCESS
PostgreSQL 16 Integration Job=90399157430
Backend Race Safety=SUCCESS
Backend Race Safety Job=90399157564
Backend Container=SUCCESS
Backend Container Job=90399476002
```

All accepted findings are implemented. Qualified and rejected findings retain the
rationale recorded in this review. No Projection Contract review item remains
open, unclassified, or deferred.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
PROJECTION_CONTRACT_ENGINEERING_DEBT=CLOSED
PROJECTION_CONTRACT_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_CONTRACT_REVIEW_STATUS=CLOSED
```

## Canonical remediation history

The records below reconstruct finding-level ownership from repository evidence.
Severity labels are retrospective engineering classifications; no historical reviewer
severity or identity is invented. The implementation owner is
`964556d0ca8a1ce9aa74c37c55961cdd006b3de8`, and exact historical Backend CI run
`30396070318` completed successfully.

### GFA-DATA-306 — Projection points were not bound to one exact horizon grid

1. **Finding / symptom:** horizon duration and point ordering did not prove that every projection point occupied the exact `AsOfTime + n*Step` slot.
2. **Root cause:** validation checked chronology and broad horizon bounds without reconstructing the canonical fixed-step grid.
3. **Failure scenario:** a result contains monotonically increasing but off-grid timestamps or a duration not divisible by `Step`, yet still appears structurally valid.
4. **Impact:** downstream comparisons, Estimated Arrival attachment, and replay can consume projections whose temporal identity differs from the declared horizon.
5. **Severity rationale:** **P1 retrospective** because point-time identity is a primary analytical contract; off-grid values can materially change forecast interpretation.
6. **Existing guarantees violated:** deterministic horizon identity, point ordering, reproducible projection timing.
7. **Considered solutions:** require only monotonic timestamps; round points to the nearest step; reconstruct and verify the exact grid.
8. **Chosen remediation:** require horizon duration divisibility and a zero-based contiguous prefix of `AsOfTime + (index+1)*Step`; complete results must end at `Horizon.EndTime`.
9. **Why selected:** validation proves producer output rather than mutating it and makes temporal identity deterministic.
10. **Rejected alternatives:** rounding would silently rewrite producer evidence; monotonic-only checks permit semantically different grids.
11. **Trade-offs:** producers and fixtures must emit exact canonical times.
12. **Regression tests / protection:** `TestValidateRejectsNonDivisibleHorizonGrid`, `TestValidateRejectsProjectionPointOffGrid`, permanent Projection Contract audit.
13. **Adversarial review findings:** a limited result may legitimately contain a full effective horizon, so grid integrity is separated from limited-status evidence semantics.
14. **Remediation iterations:** `964556d0ca8a1ce9aa74c37c55961cdd006b3de8`.
15. **Residual risks / limitations:** the generic contract validates temporal structure, not whether a producer selected an appropriate horizon policy.
16. **Operational/deployment consequences:** malformed producer output now fails contract validation; no schema or database migration.
17. **Exact evidence:** remediation commit, exact Backend CI run `30396070318`, grid regression tests, `projectioncontractreviewaudit`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** any projection producer must be validated against the canonical horizon grid before publication or persistence.

### GFA-DATA-307 — Limited results could omit evidence explaining the limitation

1. **Finding / symptom:** `limited` required points but did not require evidence explaining why the result was not `complete`.
2. **Root cause:** status validation checked structural presence but did not reconcile status with limitation evidence.
3. **Failure scenario:** a fully populated or partially populated result is labeled limited with only generic research/method boilerplate and no domain-specific reason.
4. **Impact:** operators and consumers cannot distinguish truncation, missing altitude, weak evidence, or another limitation cause.
5. **Severity rationale:** **P1 retrospective** because status semantics directly communicate analytical trust and can otherwise become unverifiable.
6. **Existing guarantees violated:** explicit limitation provenance, status/evidence reconciliation, research-only honesty.
7. **Considered solutions:** require temporal incompleteness; require one arbitrary limitation; require non-generic limiting evidence.
8. **Chosen remediation:** permit full effective-horizon coverage but require a limitation beyond generic method assumptions/research-only guard.
9. **Why selected:** it matches legitimate production cases such as truncated requested horizon or unavailable altitude without redefining `limited` as incomplete time coverage.
10. **Rejected alternatives:** forbidding full temporal coverage would reject valid limited projections.
11. **Trade-offs:** producers must publish an explicit domain limitation whenever status is limited.
12. **Regression tests / protection:** `TestValidateLimitedStatusRequiresExplicitEvidence`; permanent audit.
13. **Adversarial review findings:** limitation evidence is semantic, not inferred from point count alone.
14. **Remediation iterations:** `964556d0ca8a1ce9aa74c37c55961cdd006b3de8`.
15. **Residual risks / limitations:** the validator proves an explanation exists and is normalized, not that the underlying producer diagnosis is scientifically optimal.
16. **Operational/deployment consequences:** incomplete limiting evidence now fails before result publication.
17. **Exact evidence:** remediation commit, regression test, run `30396070318`, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** analytical status values that reduce trust must carry explicit machine-readable domain evidence.

### GFA-DATA-308 — Projection confidence was not reconciled with reasons and mandatory evidence

1. **Finding / symptom:** positive confidence could lack reasons, contributions were merely finite, duplicate reason codes were accepted, and overall confidence could exceed mandatory point or Estimated Arrival confidence.
2. **Root cause:** confidence fields were validated independently rather than as one cross-field evidence contract.
3. **Failure scenario:** a result advertises high confidence despite weaker point/arrival evidence, duplicated reasons, or no positive supporting reason at all.
4. **Impact:** trust signals can overstate the strength and diversity of supporting evidence.
5. **Severity rationale:** **P1 retrospective** because confidence is a primary published analytical trust signal.
6. **Existing guarantees violated:** evidence-bounded confidence, explanatory confidence provenance, deterministic reason identity.
7. **Considered solutions:** add global score bands; validate only numeric range; reconcile positive confidence with normalized unique reasons and weakest mandatory evidence.
8. **Chosen remediation:** positive confidence requires reasons; contributions must be finite within `[-1,+1]`; reason codes must be unique; overall confidence may not exceed the weakest mandatory point or Estimated Arrival evidence.
9. **Why selected:** it constrains generic contract integrity without inventing one universal producer calibration policy.
10. **Rejected alternatives:** a global low/medium/high numeric mapping would conflict with producer-owned versioned thresholds.
11. **Trade-offs:** producers must emit complete confidence evidence and cannot smooth away a weaker mandatory component at the contract boundary.
12. **Regression tests / protection:** positive-reason, duplicate-reason, contribution-bound, and weakest-evidence tests protected by the audit.
13. **Adversarial review findings:** `float64` remains valid for these non-monetary scores; the defect was semantic reconciliation, not numeric storage type.
14. **Remediation iterations:** `964556d0ca8a1ce9aa74c37c55961cdd006b3de8`.
15. **Residual risks / limitations:** the contract does not determine whether a producer's configured confidence formula is empirically calibrated.
16. **Operational/deployment consequences:** invalid confidence claims fail contract validation; payload layout unchanged.
17. **Exact evidence:** remediation commit, run `30396070318`, confidence regression tests, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** published confidence must remain explainable, deduplicated, bounded, and no stronger than required supporting evidence.

### GFA-CONTRACT-309 — Projection Contract duplicated the shared confidence vocabulary

1. **Finding / symptom:** Projection Contract defined a second local ordinal confidence type instead of using the domain confidence vocabulary.
2. **Root cause:** module-local evolution duplicated an already shared cross-domain value object.
3. **Failure scenario:** one confidence vocabulary adds or changes known-value behavior while Projection accepts a divergent set or validation rule.
4. **Impact:** contract drift across analytical domains and transport adapters.
5. **Severity rationale:** **P2 retrospective** because the primary risk is contract divergence rather than immediate numerical corruption.
6. **Existing guarantees violated:** shared semantic vocabulary, single ownership of confidence levels, cross-module compatibility.
7. **Considered solutions:** keep synchronized enums; convert at boundaries; alias the shared domain type.
8. **Chosen remediation:** `projectioncontract.ConfidenceLevel` aliases `domain/confidence.Level`; shared type exposes `IsKnown`.
9. **Why selected:** it preserves source compatibility while eliminating duplicate semantic ownership.
10. **Rejected alternatives:** conversion layers preserve two vocabularies and future drift risk.
11. **Trade-offs:** Projection Contract now explicitly depends on the shared domain confidence value object.
12. **Regression tests / protection:** permanent audit forbids a local `type ConfidenceLevel string` and checks shared aliases.
13. **Adversarial review findings:** producer-specific score thresholds remain producer-owned and are not collapsed into the shared ordinal vocabulary.
14. **Remediation iterations:** `964556d0ca8a1ce9aa74c37c55961cdd006b3de8`.
15. **Residual risks / limitations:** numeric score semantics can still differ by method, intentionally.
16. **Operational/deployment consequences:** no payload change; source compatibility retained through aliasing.
17. **Exact evidence:** remediation commit, shared confidence code, run `30396070318`, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** cross-domain ordinal vocabularies should have one semantic owner unless a deliberate versioned distinction is documented.

### GFA-DATA-310 — Projection input fingerprints accepted arbitrary non-empty text

1. **Finding / symptom:** usable input fingerprints were validated only as non-empty strings.
2. **Root cause:** provenance validation treated fingerprint presence as sufficient without enforcing canonical cryptographic identity format.
3. **Failure scenario:** malformed, truncated, uppercase, or arbitrary strings are published as if they were deterministic SHA-256 provenance identities.
4. **Impact:** auditability and cross-run identity comparison become unreliable.
5. **Severity rationale:** **P2 retrospective** because malformed identity weakens reproducibility, while producer-owned fingerprint contents remain separately protected.
6. **Existing guarantees violated:** canonical provenance identity, deterministic comparison, hash-format contract.
7. **Considered solutions:** recompute from summarized provenance; accept any string; enforce canonical SHA-256 format only.
8. **Chosen remediation:** present fingerprints must match `sha256:` plus sixty-four lowercase hexadecimal characters.
9. **Why selected:** it enforces cryptographic identity shape without pretending the summarized provenance list contains all producer fingerprint inputs.
10. **Rejected alternatives:** recomputation from `Provenance.Inputs` would omit richer producer configuration/source state.
11. **Trade-offs:** legacy malformed fingerprints are rejected rather than normalized.
12. **Regression tests / protection:** malformed-fingerprint test and audit `fingerprintPattern` requirement.
13. **Adversarial review findings:** format validation is intentionally distinct from producer-specific fingerprint formula validation.
14. **Remediation iterations:** `964556d0ca8a1ce9aa74c37c55961cdd006b3de8`.
15. **Residual risks / limitations:** a syntactically valid SHA-256 string can still be semantically wrong if a producer hashes the wrong inputs; producer audits own that risk.
16. **Operational/deployment consequences:** malformed projection provenance fails validation.
17. **Exact evidence:** remediation commit, malformed fingerprint regression test, run `30396070318`, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** shared contracts should enforce canonical fingerprint encoding while leaving richer hash-input ownership to versioned producers.

### GFA-CONTRACT-311 — ICAO24 identifiers lacked exact hexadecimal validation

1. **Finding / symptom:** ICAO24 values were not required to be exactly six hexadecimal characters.
2. **Root cause:** validation treated identifier text as loosely formatted provenance rather than a typed 24-bit transponder identifier.
3. **Failure scenario:** placeholders or malformed aircraft identities pass through Projection results and fingerprints as if they were real ICAO24 values.
4. **Impact:** aircraft identity and provenance can be ambiguous or invalid.
5. **Severity rationale:** **P2 retrospective** because identity correctness is important but the defect does not by itself alter the projection formula.
6. **Existing guarantees violated:** canonical aircraft identity, provenance validity, cross-module identifier consistency.
7. **Considered solutions:** trim-only validation; normalize arbitrary text; enforce six hexadecimal characters.
8. **Chosen remediation:** present ICAO24 must match exactly six hexadecimal characters.
9. **Why selected:** it matches the domain identity without inventing case-sensitive semantic differences.
10. **Rejected alternatives:** loose placeholders are unsuitable for canonical analytical evidence.
11. **Trade-offs:** invalid fixtures and producers must be corrected instead of silently accepted.
12. **Regression tests / protection:** `TestValidateRejectsInvalidICAO24`, fixture corrections, permanent audit.
13. **Adversarial review findings:** cross-package identifier-helper consolidation is not required for this contract correction.
14. **Remediation iterations:** `964556d0ca8a1ce9aa74c37c55961cdd006b3de8`.
15. **Residual risks / limitations:** syntactic validity does not prove an ICAO24 was assigned to the claimed physical aircraft at the source time.
16. **Operational/deployment consequences:** malformed identities fail contract validation.
17. **Exact evidence:** remediation commit, run `30396070318`, ICAO24 tests, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** externally meaningful identifiers must be validated against their canonical grammar at publication boundaries.

### GFA-CONTRACT-312 — Estimated Arrival accepted non-ICAO airport location indicators

1. **Finding / symptom:** four-character airport identifiers could contain digits rather than four uppercase letters.
2. **Root cause:** validation enforced length but not the canonical ICAO location-indicator grammar.
3. **Failure scenario:** an Estimated Arrival result carries a syntactically four-character but invalid airport identifier and is accepted by the generic contract.
4. **Impact:** arrival identity can become non-portable or inconsistent with route/airport domains.
5. **Severity rationale:** **P2 retrospective** because destination identity can be invalid even if estimated time arithmetic is otherwise correct.
6. **Existing guarantees violated:** normalized airport identity, cross-domain contract compatibility.
7. **Considered solutions:** length-only check; alphanumeric check; four-uppercase-letter ICAO grammar.
8. **Chosen remediation:** present airport ICAO values must match `[A-Z]{4}`.
9. **Why selected:** it preserves the documented airport identity contract without adding route fields to this payload.
10. **Rejected alternatives:** accepting digits broadens the identifier domain beyond the intended ICAO location indicator.
11. **Trade-offs:** legacy placeholder fixtures must use valid-looking ICAO identifiers.
12. **Regression tests / protection:** airport ICAO validation tests and permanent audit.
13. **Adversarial review findings:** destination equality with the route cannot be reconstructed from this payload and remains producer-owned until a schema revision carries route identity.
14. **Remediation iterations:** `964556d0ca8a1ce9aa74c37c55961cdd006b3de8`.
15. **Residual risks / limitations:** grammar validation does not prove the airport exists in the current catalog.
16. **Operational/deployment consequences:** invalid airport location indicators fail result validation.
17. **Exact evidence:** remediation commit, run `30396070318`, airport identifier tests, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** shared analytical payloads must enforce the canonical grammar of identifiers they publish.

### GFA-DATA-313 — Projection provenance could omit source or observation-basis evidence

1. **Finding / symptom:** observed, openly sourced, derived, and estimated inputs did not consistently require source names and observation/analytical-basis times.
2. **Root cause:** provenance entries shared one permissive shape without enforcing semantics appropriate to evidence-bearing input classes.
3. **Failure scenario:** a projection claims an observed or derived input with no source identity or temporal basis, making the contribution impossible to reproduce or audit.
4. **Impact:** provenance can look structurally complete while lacking the evidence needed to understand when and from what it was derived.
5. **Severity rationale:** **P1 retrospective** because temporal/source provenance is core to historical and research analytical trust.
6. **Existing guarantees violated:** source attribution, temporal correctness, reproducible analytical evidence.
7. **Considered solutions:** require source/time for every input indiscriminately; keep optional; enforce by evidence class.
8. **Chosen remediation:** observed/openly sourced/derived/estimated inputs require normalized source identity and observation or analytical-basis time as applicable.
9. **Why selected:** it strengthens evidence-bearing classes without fabricating meaningless fields for inputs that legitimately do not use them.
10. **Rejected alternatives:** universal mandatory fields can invent false provenance; permissive optional fields leave evidence unverifiable.
11. **Trade-offs:** producers must classify and populate provenance consistently.
12. **Regression tests / protection:** incomplete observed-input tests and permanent provenance audit rules.
13. **Adversarial review findings:** producer-qualified input names remain allowed because uniqueness, not one global naming vocabulary, is the contract concern.
14. **Remediation iterations:** `964556d0ca8a1ce9aa74c37c55961cdd006b3de8`.
15. **Residual risks / limitations:** source strings identify declared origins but do not independently authenticate external providers.
16. **Operational/deployment consequences:** incomplete evidence provenance fails before publication.
17. **Exact evidence:** remediation commit, provenance regression tests, run `30396070318`, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** evidence classes must carry the minimum source and temporal basis required to reproduce their analytical contribution.

### GFA-DATA-314 — Projection provenance chronology allowed retrieval before observation

1. **Finding / symptom:** retrieval time could precede the observation time it purportedly retrieved.
2. **Root cause:** provenance timestamps were range-checked individually but not reconciled causally.
3. **Failure scenario:** an input declares `RetrievedAt < ObservedAt`, or retrieval after result generation, yet remains accepted as valid evidence.
4. **Impact:** temporal provenance becomes causally impossible and historical reconstruction can be misleading.
5. **Severity rationale:** **P1 retrospective** because impossible time ordering directly violates temporal correctness.
6. **Existing guarantees violated:** causal provenance, `GeneratedAt` boundary, historical evidence integrity.
7. **Considered solutions:** tolerate clock skew; compare only with `GeneratedAt`; enforce exact observation <= retrieval <= generation ordering.
8. **Chosen remediation:** present retrieval time must not precede observation time or exceed `GeneratedAt`.
9. **Why selected:** no repository evidence justified an arbitrary tolerance; the fields represent causal analytical events.
10. **Rejected alternatives:** silent tolerance would weaken deterministic validation and mask malformed producer evidence.
11. **Trade-offs:** producers must normalize timestamp semantics before constructing provenance.
12. **Regression tests / protection:** provenance chronology tests and permanent audit.
13. **Adversarial review findings:** this rule concerns declared observation/retrieval semantics and does not conflate processing timestamps with event time where domains distinguish them.
14. **Remediation iterations:** `964556d0ca8a1ce9aa74c37c55961cdd006b3de8`.
15. **Residual risks / limitations:** source clock accuracy remains external; the contract only rejects internally impossible declared order.
16. **Operational/deployment consequences:** causally invalid provenance is rejected.
17. **Exact evidence:** remediation commit, run `30396070318`, chronology validation/audit requirements.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** provenance contracts must explicitly reconcile causal timestamp order, not validate timestamps in isolation.

### GFA-DATA-315 — Duplicate Projection evidence and explanation identities were accepted

1. **Finding / symptom:** duplicate provenance inputs, limitation scope/code pairs, explanation codes, and confidence reason codes could coexist in one result.
2. **Root cause:** validation checked individual item shape without enforcing collection-level semantic uniqueness.
3. **Failure scenario:** the same evidence/reason is repeated and appears to provide multiple independent justifications or creates order-dependent consumers.
4. **Impact:** evidence weighting and human interpretation can be overstated or ambiguous.
5. **Severity rationale:** **P2 retrospective** because duplication distorts explanatory/provenance integrity but does not automatically change producer numerical output.
6. **Existing guarantees violated:** normalized evidence identity, deterministic explanation sets, non-duplicated confidence rationale.
7. **Considered solutions:** silently deduplicate; preserve duplicates; reject duplicates at validation.
8. **Chosen remediation:** require unique normalized input names, confidence reason codes, explanation codes, and limitation scope/code pairs.
9. **Why selected:** fail-closed rejection preserves producer intent and prevents the contract from guessing which duplicate to keep.
10. **Rejected alternatives:** silent deduplication can erase materially different malformed records under the same identity.
11. **Trade-offs:** producers must canonicalize collections before publication.
12. **Regression tests / protection:** duplicate provenance/reason/limitation/explanation tests and audit checks.
13. **Adversarial review findings:** qualified producer input names such as `historical_neighbor:<id>` remain valid; uniqueness is semantic within the result, not global across all producer families.
14. **Remediation iterations:** `964556d0ca8a1ce9aa74c37c55961cdd006b3de8`.
15. **Residual risks / limitations:** distinct codes can still describe overlapping concepts; taxonomy governance remains producer/domain-owned.
16. **Operational/deployment consequences:** duplicate semantic identities fail validation.
17. **Exact evidence:** remediation commit, run `30396070318`, duplicate tests, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** explanatory and provenance collections must enforce canonical identity uniqueness at contract boundaries.

### GFA-CONTRACT-316 — Projection Result lacked a typed public validation boundary

1. **Finding / symptom:** validation reports existed, but `Result` had no public `Validate()` method returning a typed error.
2. **Root cause:** validation was exposed as a separate utility/report path rather than an object-level boundary usable by arbitrary consumers.
3. **Failure scenario:** a consumer receives or constructs a `Result` and either skips validation or must know package-internal report conventions to fail correctly.
4. **Impact:** collaborator and transport boundaries can inconsistently enforce the same contract.
5. **Severity rationale:** **P2 retrospective** because the underlying validator existed, but safe consumption lacked one explicit idiomatic adapter.
6. **Existing guarantees violated:** easy fail-closed consumption, typed validation errors, boundary consistency.
7. **Considered solutions:** constructors/private fields; leave report-only API; add `Result.Validate()` wrapping cloned issues in a typed error.
8. **Chosen remediation:** `Result.Validate()` returns `ResultValidationError` with cloned validation issues.
9. **Why selected:** it provides a stable fail-closed boundary without a broad migration of public structs and fixtures.
10. **Rejected alternatives:** hiding all fields would not prevent post-copy mutation and would create unrelated producer migration work.
11. **Trade-offs:** both report-oriented and error-oriented validation APIs coexist intentionally.
12. **Regression tests / protection:** `TestResultValidateReturnsTypedError`; permanent audit checks method/error presence.
13. **Adversarial review findings:** public mutable structs are retained by design; validation before publication/consumption is therefore a required boundary.
14. **Remediation iterations:** `964556d0ca8a1ce9aa74c37c55961cdd006b3de8`.
15. **Residual risks / limitations:** callers can still intentionally skip `Validate`; producer/consumer audits protect production paths.
16. **Operational/deployment consequences:** consumers can use standard `errors.As` semantics for invalid projection results.
17. **Exact evidence:** remediation commit, typed-error test, run `30396070318`, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** mutable versioned aggregate contracts should expose a direct typed validation boundary for consumers.

### GFA-TEST-317 — Projection Contract regression tests omitted critical cross-field invariants

1. **Finding / symptom:** tests did not permanently protect grid alignment, limited evidence, confidence reconciliation, identifier grammar, fingerprint format, provenance chronology/uniqueness, and typed validation behavior.
2. **Root cause:** existing tests focused on field-level/happy-path validation rather than cross-field contract relationships.
3. **Failure scenario:** a refactor removes one reconciliation check while ordinary producer tests still pass.
4. **Impact:** multiple high-impact Projection Contract remediations could regress without CI detection.
5. **Severity rationale:** **P2 retrospective** because the gap weakens durable prevention for P1/P2 contract defects.
6. **Existing guarantees violated:** CI truth, remediation closure, regression protection.
7. **Considered solutions:** rely on producer tests; add contract tests; combine targeted tests with permanent strict source/contract audit.
8. **Chosen remediation:** add targeted review-hardening tests and `projectioncontractreviewaudit` wired into Backend CI.
9. **Why selected:** behavioral tests prove outcomes while the audit protects critical wiring, versions, and closure evidence.
10. **Rejected alternatives:** incidental producer coverage does not prove generic contract behavior remains enforced.
11. **Trade-offs:** CI maintenance increases when the contract deliberately evolves.
12. **Regression tests / protection:** the review-hardening test suite plus strict audit is itself the guard.
13. **Adversarial review findings:** source audit is not treated as a replacement for behavioral tests; both are required.
14. **Remediation iterations:** `964556d0ca8a1ce9aa74c37c55961cdd006b3de8`.
15. **Residual risks / limitations:** no finite suite proves every future producer interaction.
16. **Operational/deployment consequences:** stronger merge gate; no runtime behavior added by the guard itself.
17. **Exact evidence:** remediation commit, `apps/api/tools/projectioncontractreviewaudit`, exact run `30396070318` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** cross-field analytical contracts require dedicated behavioral regression coverage plus permanent CI enforcement.

### GFA-TEST-318 — Projection fixtures encoded invalid aircraft/confidence/provenance evidence

1. **Finding / symptom:** several Projection Intelligence fixtures used non-hex ICAO24 placeholders or incomplete confidence/provenance evidence.
2. **Root cause:** fixtures reflected previously permissive contracts and were not representative of valid canonical production evidence.
3. **Failure scenario:** strengthened validation appears to break for the wrong reason, or tests keep teaching future code that malformed identities/evidence are acceptable.
4. **Impact:** test suites can mask contract drift and produce misleading failures during hardening.
5. **Severity rationale:** **P2 retrospective** because invalid fixtures undermine verification of production contracts even though they are not runtime records themselves.
6. **Existing guarantees violated:** fixture parity, test realism, trustworthy regression evidence.
7. **Considered solutions:** weaken validation for fixtures; special-case test values; correct fixtures to satisfy the production contract.
8. **Chosen remediation:** update producer fixtures to valid hexadecimal ICAO24 values and complete fallback confidence/provenance evidence.
9. **Why selected:** tests should conform to production contracts rather than creating a second test-only grammar.
10. **Rejected alternatives:** validator exceptions for fixtures would make CI evidence less trustworthy.
11. **Trade-offs:** fixture maintenance must follow intentional contract revisions.
12. **Regression tests / protection:** corrected producer tests execute under the permanent Projection Contract audit and full Backend CI.
13. **Adversarial review findings:** fixture corrections are registered separately from missing regression coverage because stale invalid data and absent invariant tests are distinct failure modes.
14. **Remediation iterations:** `964556d0ca8a1ce9aa74c37c55961cdd006b3de8`.
15. **Residual risks / limitations:** other unrelated test packages may still need independent fixture review as contracts evolve.
16. **Operational/deployment consequences:** none at runtime; CI evidence becomes contract-faithful.
17. **Exact evidence:** remediation commit, producer test changes, run `30396070318`, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** test fixtures used as evidence must satisfy the same canonical identifier and provenance contracts as production results.
