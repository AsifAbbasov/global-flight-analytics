# Historical Similarity Review Hardening

Status: closed

## Scope

This increment hardens `apps/api/internal/historicalintelligence/historicalsimilarity` as the deterministic, bounded comparison boundary for historical trajectory shape similarity.

## Accepted findings

- Similarity score and similarity level described route-shape closeness but no separate confidence contract expressed whether the two trajectory inputs were trustworthy enough to support that conclusion.
- Trajectory `QualityScore`, segment quality and status, coverage gaps, excluded points, and observation cadence were ignored.
- `SampleCount` had no upper bound and a single comparison could accept an unbounded trajectory point slice.
- The public `Rank` method duplicated the production-owned `projectionneighbors` selection workflow, silently discarded candidate errors, re-prepared the reference for every candidate, and exposed unbounded candidate processing.
- Fingerprinting rounded coordinates and configuration, sorted semantic sequence records globally, and used raw input instead of the prepared representation that actually drives scoring.
- Equal timestamps preserved caller order and could make path geometry depend on incidental input sequence.
- Result validation accepted unknown component names and did not verify component observations, formulas, weighted score, mean-versus-maximum distance, or confidence mathematics.
- Endpoint scoring averaged the two endpoints instead of using the worse endpoint.
- Relative difference used an undocumented one-kilometre floor.
- Latitude and longitude were interpolated linearly instead of following the same spherical great-circle model used for distance.
- `Compare`, preparation, resampling, fingerprinting, quality assessment, and validation were concentrated in oversized files and functions.
- `NewDefault` used a panic path for a package-owned constant configuration.

## Corrected contracts

- `Result.Score` and `Result.Level` are explicitly similarity-only fields. `Result.Confidence` is separate and uses the weaker reference or candidate evidence score.
- Evidence quality binds declared trajectory quality, point-weighted segment quality adjusted by segment status, coverage continuity, observation cadence regularity, and usable-point retention.
- Missing source identity is reported as a limitation; provider reliability is not fabricated because the trajectory domain exposes a source name but no provider-quality metric.
- `SampleCount` is bounded by `MaximumSampleCount`, and input trajectories are bounded by `MaximumInputPointCount`.
- Historical Similarity exposes only `Compare`. Candidate ranking, truncation, rejection reasons, and result limits remain owned by the existing production `projectionneighbors` selector.
- Equal timestamps are canonicalized by timestamp, coordinates, and point identifier.
- Fingerprint generation two consumes the prepared canonical points, resampled points, evidence-quality values, limitations, and exact `math.Float64bits` configuration values without global record sorting.
- Result validation requires the four version-two components in canonical order and recomputes observed values, component scores, policy weights, weighted similarity, confidence, and evidence-quality formulas.
- Endpoint score uses `max(start distance, end distance)`.
- Relative difference is exact: zero versus non-zero is fully different; zero versus zero is equal.
- Resampling uses spherical great-circle interpolation with a deterministic near-antipodal fallback and explicit limitation.
- Compensated accumulation protects path length, sample distance, and weighted-score sums.
- The implementation is decomposed into configuration, preparation, geodesy, quality, scoring, fingerprint, notices, and validation files.

## Qualified or rejected findings

- A fixed four-component model is not by itself an Open/Closed Principle violation. It is a versioned analytical policy whose component set must change deliberately with a new contract version. The remediation centralizes construction and validation so the set cannot drift across unrelated functions.
- Returning `nil` with an error from `New` is idiomatic Go and is not a domain `null` value. It remains unchanged. The unnecessary panic in `NewDefault` is removed.
- Moving every duplicated geodesic helper across Projection, Weather, and Historical Intelligence into one shared package is a separate cross-cutting migration. This increment corrects Historical Similarity geodesy without coupling unrelated modules or changing their established contracts.
- Similarity and confidence values retain full finite `float64` precision. This is coordinate and trigonometric analytics, not financial arithmetic. Presentation rounding is outside domain identity, while fingerprints use exact floating-point bits.

```text
SIMILARITY_CONFIDENCE_SEPARATED=YES
TRAJECTORY_QUALITY_EVIDENCE=BOUND
SAMPLE_COUNT_MAXIMUM=ENFORCED
PUBLIC_RANK_API=REMOVED
FINGERPRINT_PREPARED_EXACT=ENFORCED
EQUAL_TIMESTAMP_CANONICALIZATION=ENFORCED
RESULT_MATHEMATICAL_VALIDATION=ENFORCED
ENDPOINT_SCORE_USES_WORST_ENDPOINT=YES
RELATIVE_DIFFERENCE_ZERO_SCALE=EXACT
GREAT_CIRCLE_RESAMPLING=ENFORCED
HISTORICAL_SIMILARITY_ENGINEERING_REMEDIATION=IMPLEMENTED
```

## Permanent verification

`apps/api/tools/historicalsimilarityreviewaudit` protects the version-two similarity-versus-confidence boundary, bounded inputs, removal of the duplicate Rank API, canonical preparation, exact fingerprint identity, trajectory quality evidence, mathematical result validation, worst-endpoint scoring, exact relative difference, great-circle resampling, regression tests, and this review record in Backend Continuous Integration.

## Formal closure evidence

The Historical Similarity engineering remediation was committed and validated
before this administrative closure:

```text
ENGINEERING_BASELINE_COMMIT=2d61a3fa3be100312708d2fae0e5d1ae43f419f5
ENGINEERING_REMEDIATION_COMMIT=6dbae4e6fe00295af0f7ba5303855736b76e8bde
ENGINEERING_GITHUB_ACTIONS_RUN=30360637718
PostgreSQL 16 Integration=SUCCESS
PostgreSQL 16 Integration Job=90279277488
Backend Race Safety=SUCCESS
Backend Race Safety Job=90279277503
Backend Quality=SUCCESS
Backend Quality Job=90279277576
Backend Container=SUCCESS
Backend Container Job=90279633063
```

All accepted findings are implemented. Qualified or rejected findings retain
their documented rationale, and no Historical Similarity review item remains
open, unclassified, or deferred.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
HISTORICAL_SIMILARITY_ENGINEERING_DEBT=CLOSED
HISTORICAL_SIMILARITY_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
HISTORICAL_SIMILARITY_REVIEW_STATUS=CLOSED
```

## Canonical remediation history

Historical reviewer identities and review-comment chronology are not reconstructed. The records below are reconstructed from the review document, remediation commit `6dbae4e6fe00295af0f7ba5303855736b76e8bde`, regression tests, permanent strict audit, and exact Backend CI run `30360637718`. Severity labels are retrospective.

### GFA-DATA-267 — Similarity score was not separated from evidence confidence
1. **Finding / symptom:** similarity score/level described shape closeness without a separate statement of input trustworthiness.
2. **Root cause:** geometric likeness and evidence quality were collapsed into one interpretation.
3. **Failure scenario:** two poorly observed trajectories receive a high shape-similarity score that consumers mistake for a high-confidence conclusion.
4. **Impact:** analytical trust is overstated even when geometry happens to align.
5. **Severity rationale:** **P1 retrospective** because the result could make a strong analytical claim without sufficient evidence confidence.
6. **Existing guarantees violated:** analytical score and confidence must represent different properties.
7. **Considered solutions:** fold quality into score, expose only limitations, or add an independent evidence-confidence contract.
8. **Chosen remediation:** `Score`/`Level` remain similarity-only; `Confidence` separately uses the weaker trajectory evidence score.
9. **Why selected:** preserves interpretable geometry while exposing trust explicitly.
10. **Rejected alternatives:** silently penalizing similarity with quality, which would make the score semantically ambiguous.
11. **Trade-offs:** consumers must read both similarity and confidence.
12. **Regression tests / protection:** `TestCompareSeparatesSimilarityFromEvidenceConfidence` and strict audit.
13. **Adversarial review findings:** full finite `float64` precision remains intentional and unrelated.
14. **Remediation iterations:** `6dbae4e6fe00295af0f7ba5303855736b76e8bde`.
15. **Residual risks and limitations:** confidence can only use quality evidence exposed by the trajectory domain.
16. **Operational or deployment consequences:** result contract advances to version two; no migration stated by this review.
17. **Exact evidence:** remediation commit, run `30360637718` SUCCESS, `historicalsimilarityreviewaudit`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** new similarity components must not conflate closeness with source/evidence trust.

### GFA-DATA-268 — Historical Similarity ignored authoritative trajectory-quality evidence
1. **Finding / symptom:** trajectory quality score, segment quality/status, coverage gaps, excluded points and observation cadence did not influence evidence confidence.
2. **Root cause:** preparation focused on geometric coordinates and duration rather than quality provenance.
3. **Failure scenario:** sparse, gap-heavy or low-quality trajectories are treated as equally trustworthy as complete trajectories.
4. **Impact:** confidence does not reflect material evidence limitations.
5. **Severity rationale:** **P1 retrospective** because confidence could be materially overstated.
6. **Existing guarantees violated:** confidence must be source- and quality-aware.
7. **Considered solutions:** trust `QualityScore` alone, use provider reliability, or compose available trajectory/segment/coverage/cadence/retention evidence.
8. **Chosen remediation:** evidence quality binds declared quality, segment quality/status, coverage continuity, cadence regularity and usable-point retention.
9. **Why selected:** uses available domain evidence without fabricating provider reliability.
10. **Rejected alternatives:** invented provider-quality metric and source-name-as-quality proxy.
11. **Trade-offs:** confidence calculation is more complex and versioned.
12. **Regression tests / protection:** quality/gap/retention/confidence tests and strict audit quality fragments.
13. **Adversarial review findings:** missing source identity is a limitation, not a fabricated reliability score.
14. **Remediation iterations:** `6dbae4e6fe00295af0f7ba5303855736b76e8bde`.
15. **Residual risks and limitations:** unavailable upstream quality dimensions cannot be reconstructed here.
16. **Operational or deployment consequences:** low-quality evidence yields lower confidence/limitations rather than altered geometry score.
17. **Exact evidence:** remediation commit, quality tests, exact CI run, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** trajectory-derived confidence changes require explicit coverage of every quality input used or intentionally unavailable.

### GFA-PERF-269 — Historical Similarity accepted unbounded sample and input-point sizes
1. **Finding / symptom:** `SampleCount` had no maximum and a comparison could consume an unbounded point slice.
2. **Root cause:** configuration validated only minimums and assumed upstream bounded inputs.
3. **Failure scenario:** oversized trajectories or sampling configuration cause excessive CPU/memory work in preparation, resampling and distance loops.
4. **Impact:** resource exhaustion and unpredictable latency.
5. **Severity rationale:** **P2 retrospective** because this is a production resource-boundary defect rather than evidence corruption by itself.
6. **Existing guarantees violated:** research analytics must remain explicitly bounded under free-tier constraints.
7. **Considered solutions:** rely on callers, truncate silently, or enforce explicit maxima.
8. **Chosen remediation:** enforce `MaximumSampleCount` and `MaximumInputPointCount`.
9. **Why selected:** fail-closed limits are deterministic and visible.
10. **Rejected alternatives:** silent truncation without evidence disclosure and unbounded caller ownership.
11. **Trade-offs:** oversized comparisons are rejected and require upstream selection/downsampling.
12. **Regression tests / protection:** excessive sample/input tests and strict audit.
13. **Adversarial review findings:** candidate-set ranking limits remain owned by Projection Neighbors rather than this engine.
14. **Remediation iterations:** `6dbae4e6fe00295af0f7ba5303855736b76e8bde`.
15. **Residual risks and limitations:** accepted maxima still consume non-zero CPU proportional to configured bounds.
16. **Operational or deployment consequences:** bounded comparison cost.
17. **Exact evidence:** remediation commit, bound tests, run `30360637718`, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every new collection/sampling dimension in Similarity requires an explicit maximum before production exposure.

### GFA-ARCH-270 — Public Similarity `Rank` duplicated the production candidate-selection workflow
1. **Finding / symptom:** `Rank` reimplemented candidate iteration/ranking, silently skipped errors, re-prepared the reference and exposed unbounded candidate processing.
2. **Root cause:** ranking responsibility was duplicated inside Similarity despite existing production ownership in `projectionneighbors`.
3. **Failure scenario:** direct callers receive a different candidate/error/truncation policy from production Projection selection.
4. **Impact:** behavioral drift, hidden candidate failures and avoidable repeated work.
5. **Severity rationale:** **P2 retrospective** because the duplicate path could produce inconsistent analytical selection and operational cost.
6. **Existing guarantees violated:** one production owner for candidate ranking and rejection semantics.
7. **Considered solutions:** harden both rankers, delegate from Similarity, or remove the duplicate public API.
8. **Chosen remediation:** Historical Similarity exposes only `Compare`; ranking/truncation/rejections remain in `projectionneighbors`.
9. **Why selected:** removes competing ownership rather than synchronizing two policies.
10. **Rejected alternatives:** maintaining a second public ranking contract.
11. **Trade-offs:** callers needing ranking must use the production selector.
12. **Regression tests / protection:** strict audit forbids `func (engine *Engine) Rank` and silent `continue` path.
13. **Adversarial review findings:** removing a duplicate API is remediation/prevention; no shared generic ranking abstraction was required.
14. **Remediation iterations:** `6dbae4e6fe00295af0f7ba5303855736b76e8bde`.
15. **Residual risks and limitations:** Projection Neighbors remains responsible for its own bounded candidate policy.
16. **Operational or deployment consequences:** one ranking workflow and less repeated preparation.
17. **Exact evidence:** commit diff, permanent audit, exact CI run.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** selection/ranking APIs must have one domain owner; Similarity remains pairwise comparison only.

### GFA-DATA-271 — Similarity fingerprint did not bind the exact prepared scoring representation
1. **Finding / symptom:** fingerprinting rounded coordinates/configuration, globally sorted sequence records and hashed raw input instead of the prepared representation used for scoring.
2. **Root cause:** fingerprint identity was implemented independently of canonical preparation/scoring.
3. **Failure scenario:** two scoring-distinct prepared inputs collide after rounding/sorting, or scoring-equivalent raw inputs receive unrelated identity.
4. **Impact:** non-reproducible immutable comparison identity.
5. **Severity rationale:** **P1 retrospective** because input fingerprint could fail to identify the actual analytical computation.
6. **Existing guarantees violated:** fingerprint must bind exact output-affecting canonical evidence and policy.
7. **Considered solutions:** increase decimal precision, hash raw structures, or hash exact prepared representation and config bits.
8. **Chosen remediation:** generation-two fingerprint binds prepared canonical/resampled points, quality/limitations and `math.Float64bits` configuration without global semantic reordering.
9. **Why selected:** identity follows exactly what drives the score.
10. **Rejected alternatives:** formatted decimal hashing and globally sorting ordered trajectory evidence.
11. **Trade-offs:** fingerprint generation is intentionally incompatible with generation one.
12. **Regression tests / protection:** exact-float-bit and prepared-fingerprint tests; strict audit forbids `%.12f`, global record sorting and raw trajectory fingerprint input.
13. **Adversarial review findings:** presentation rounding remains outside domain identity.
14. **Remediation iterations:** `6dbae4e6fe00295af0f7ba5303855736b76e8bde`.
15. **Residual risks and limitations:** future preparation-policy changes require fingerprint/version review.
16. **Operational or deployment consequences:** corrected deterministic identity for replay/cache/persistence consumers.
17. **Exact evidence:** remediation commit, fingerprint tests, run `30360637718`, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** fingerprint code must consume canonical prepared evidence, not a parallel approximation of inputs.

### GFA-DATA-272 — Equal-timestamp trajectory points depended on caller order
1. **Finding / symptom:** observations sharing a timestamp retained incidental input ordering.
2. **Root cause:** chronological sorting had no deterministic tie-breakers for equal timestamps.
3. **Failure scenario:** permuting equivalent input slices changes path geometry, resampling or fingerprint.
4. **Impact:** non-deterministic similarity scores and identity.
5. **Severity rationale:** **P1 retrospective** because durable analytical output could depend on collection order rather than evidence.
6. **Existing guarantees violated:** deterministic replay for equivalent evidence sets.
7. **Considered solutions:** preserve caller order, deduplicate timestamps arbitrarily, or define semantic tie-breakers.
8. **Chosen remediation:** canonicalize by timestamp, coordinates and point identifier.
9. **Why selected:** deterministic and preserves all eligible observations.
10. **Rejected alternatives:** input-order dependence and arbitrary dropping of same-time points.
11. **Trade-offs:** coordinate/ID ordering becomes part of canonical preparation policy.
12. **Regression tests / protection:** `TestCompareCanonicalizesEqualTimestamps` and strict audit marker.
13. **Adversarial review findings:** sequence records are not globally sorted after preparation; chronological semantics remain primary.
14. **Remediation iterations:** `6dbae4e6fe00295af0f7ba5303855736b76e8bde`.
15. **Residual risks and limitations:** identical timestamp/coordinate/ID duplicates remain semantically indistinguishable.
16. **Operational or deployment consequences:** deterministic repeated comparison.
17. **Exact evidence:** commit, equal-time tests, exact CI run, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every ordered evidence collection requires documented deterministic tie-breakers.

### GFA-DATA-273 — Similarity result validation did not recompute analytical mathematics
1. **Finding / symptom:** validation accepted unknown components and did not verify observed values, formulas, weights, weighted score, distance relationships or confidence/quality mathematics.
2. **Root cause:** validator checked shape/ranges more than derivation integrity.
3. **Failure scenario:** a tampered or buggy result remains structurally valid while its components no longer reconcile to the declared score/confidence.
4. **Impact:** corrupted analytical output can cross the trust gate.
5. **Severity rationale:** **P1 retrospective** because the validator could certify mathematically inconsistent results.
6. **Existing guarantees violated:** persisted/returned analytical results must be recomputable from their evidence and versioned policy.
7. **Considered solutions:** range-only validation, component checksum, or full deterministic mathematical reconciliation.
8. **Chosen remediation:** require four canonical v2 components and recompute observations, scores, weights, weighted similarity, confidence and evidence-quality relationships.
9. **Why selected:** validation becomes independent proof of the result contract.
10. **Rejected alternatives:** trusting the builder because it created the object.
11. **Trade-offs:** more validation computation and tighter coupling to versioned policy.
12. **Regression tests / protection:** unknown-component, weighted-score, confidence and distance-relation tamper tests plus audit.
13. **Adversarial review findings:** fixed four-component policy is deliberate versioned policy, not an OCP defect.
14. **Remediation iterations:** `6dbae4e6fe00295af0f7ba5303855736b76e8bde`.
15. **Residual risks and limitations:** new component versions require coordinated builder/validator changes.
16. **Operational or deployment consequences:** malformed results fail closed.
17. **Exact evidence:** remediation commit, validator tests, CI run, strict audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** any new output field derived from scoring must have a corresponding independent validation equation.

### GFA-DATA-274 — Endpoint similarity averaged endpoints instead of using the worse endpoint
1. **Finding / symptom:** endpoint component used the average of start/end distances.
2. **Root cause:** scoring allowed one excellent endpoint to compensate for one materially poor endpoint.
3. **Failure scenario:** trajectories with matching starts but divergent destinations still receive a strong endpoint component.
4. **Impact:** route-shape similarity is overstated for operationally different paths.
5. **Severity rationale:** **P1 retrospective** because a core similarity component implemented the wrong conservative policy.
6. **Existing guarantees violated:** endpoint compatibility should reflect the weakest endpoint agreement.
7. **Considered solutions:** mean, sum, separate components, or maximum/worst endpoint distance.
8. **Chosen remediation:** endpoint observed value is `max(start distance, end distance)`.
9. **Why selected:** prevents one endpoint from masking the other.
10. **Rejected alternatives:** arithmetic averaging and schema expansion into separate endpoint metrics.
11. **Trade-offs:** endpoint score is intentionally more conservative.
12. **Regression tests / protection:** `TestEndpointComponentUsesWorstEndpoint` and audit requirement `math.Max`.
13. **Adversarial review findings:** component weighting remains versioned policy.
14. **Remediation iterations:** `6dbae4e6fe00295af0f7ba5303855736b76e8bde`.
15. **Residual risks and limitations:** endpoint distance alone does not describe intermediate path shape; geometry component remains separate.
16. **Operational or deployment consequences:** corrected v2 scores.
17. **Exact evidence:** commit, endpoint test, run `30360637718`, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** conservative component semantics must be encoded in named tests, not inferred from generic aggregation helpers.

### GFA-DATA-275 — Relative-difference scoring used an undocumented one-kilometre floor
1. **Finding / symptom:** relative difference normalized small/zero values against an arbitrary one-kilometre scale.
2. **Root cause:** division-by-zero avoidance was embedded as an undocumented domain policy.
3. **Failure scenario:** zero versus non-zero or very small path values appear artificially similar because the denominator floor dominates.
4. **Impact:** path-length/duration similarity can be biased around zero scale.
5. **Severity rationale:** **P2 retrospective** because the scoring formula is wrong in a bounded edge region rather than universally corrupted.
6. **Existing guarantees violated:** versioned analytical formulas must have explicit zero-scale semantics.
7. **Considered solutions:** fixed floor, epsilon floor, special-case zero, or undefined metric.
8. **Chosen remediation:** exact policy: zero/zero equal; zero/non-zero fully different; otherwise ordinary relative difference.
9. **Why selected:** mathematically explicit without hidden physical-unit constants.
10. **Rejected alternatives:** undocumented kilometre/epsilon floors.
11. **Trade-offs:** discontinuity at exact zero is deliberate and testable.
12. **Regression tests / protection:** `TestRelativeDifferenceUsesExactZeroScalePolicy` and strict audit.
13. **Adversarial review findings:** domain values retain full finite float precision.
14. **Remediation iterations:** `6dbae4e6fe00295af0f7ba5303855736b76e8bde`.
15. **Residual risks and limitations:** very small non-zero values remain sensitive by mathematical definition.
16. **Operational or deployment consequences:** deterministic edge-case scores.
17. **Exact evidence:** remediation commit, zero-scale tests, exact CI run.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every normalization formula must document and test its exact zero-denominator policy.

### GFA-DATA-276 — Resampling linearly interpolated latitude/longitude instead of spherical geometry
1. **Finding / symptom:** intermediate coordinates were created by linear latitude/longitude interpolation while distance used spherical great-circle geometry.
2. **Root cause:** resampling and distance computations used inconsistent geographic models.
3. **Failure scenario:** dateline, polar or long routes produce samples that do not lie on the intended great-circle path and distort geometry score.
4. **Impact:** geographically incorrect similarity calculations.
5. **Severity rationale:** **P1 retrospective** because valid real-world trajectories can receive wrong geometry evidence.
6. **Existing guarantees violated:** one coherent geodesic model across resampling and distance.
7. **Considered solutions:** retain linear interpolation, project coordinates, or use spherical great-circle interpolation.
8. **Chosen remediation:** great-circle interpolation with deterministic near-antipodal fallback and explicit limitation.
9. **Why selected:** matches the spherical Haversine distance model and handles global routes.
10. **Rejected alternatives:** cross-module geodesy refactor as unnecessary scope expansion for this fix.
11. **Trade-offs:** more trigonometric computation and an explicit fallback for numerically ambiguous near-antipodal cases.
12. **Regression tests / protection:** dateline/polar route tests, geodesy audit fragments and compensated accumulation guards.
13. **Adversarial review findings:** duplicated geodesic helpers across unrelated modules remain a separate migration decision.
14. **Remediation iterations:** `6dbae4e6fe00295af0f7ba5303855736b76e8bde`.
15. **Residual risks and limitations:** spherical Earth is an intentional approximation, not ellipsoidal geodesy.
16. **Operational or deployment consequences:** globally consistent resampling semantics.
17. **Exact evidence:** remediation commit, geodesy tests, run `30360637718`, audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** interpolation and distance policies for one metric must use a compatible geometric model.

### GFA-MAINT-277 — Similarity preparation, scoring, quality, fingerprinting and validation were overly concentrated
1. **Finding / symptom:** large files/functions mixed comparison orchestration with configuration, preparation, geodesy, quality, scoring, fingerprint and validation.
2. **Root cause:** initial implementation evolved in one module path without clear policy boundaries.
3. **Failure scenario:** a formula change risks touching unrelated identity or validation code and makes review/regression mapping difficult.
4. **Impact:** elevated maintenance and regression risk.
5. **Severity rationale:** **P3 retrospective** because this is structural debt after concrete correctness defects are separately owned.
6. **Existing guarantees violated:** focused, reviewable ownership of analytical policies.
7. **Considered solutions:** retain monolith, generic framework extraction, or decompose by domain responsibility.
8. **Chosen remediation:** split configuration, preparation, geodesy, quality, scoring, fingerprint, notices and validation.
9. **Why selected:** makes each policy independently testable without cross-module abstraction.
10. **Rejected alternatives:** mechanical line-count/OCP refactoring and broad shared geodesy migration.
11. **Trade-offs:** more files and explicit internal boundaries.
12. **Regression tests / protection:** permanent strict audit checks the decomposed files and core tests.
13. **Adversarial review findings:** fixed four-component policy remains deliberate, not an extensibility defect.
14. **Remediation iterations:** `6dbae4e6fe00295af0f7ba5303855736b76e8bde`.
15. **Residual risks and limitations:** coordination is still required when a versioned policy spans scoring and validation.
16. **Operational or deployment consequences:** none; maintainability hardening.
17. **Exact evidence:** commit structure, audit, CI run `30360637718`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** new Similarity behavior should extend its owning policy file and matching validator/tests, not reconcentrate `Compare`.

### GFA-REL-278 — `NewDefault` could panic for package-owned constant configuration
1. **Finding / symptom:** default engine construction called validating constructor and panicked if the package’s own constant config was invalid.
2. **Root cause:** an impossible-by-contract internal configuration failure was represented as a runtime panic path.
3. **Failure scenario:** accidental default-policy drift turns ordinary service construction into process termination.
4. **Impact:** avoidable availability failure and harder recovery.
5. **Severity rationale:** **P2 retrospective** because a package-owned default could crash production initialization.
6. **Existing guarantees violated:** default constructors for internally controlled valid policy should not introduce unnecessary panic paths.
7. **Considered solutions:** retain panic, return `(*Engine,error)`, or construct directly while tests/audit guarantee the package default.
8. **Chosen remediation:** `NewDefault` directly returns an engine with `DefaultConfig`; default validity is protected by tests/audit.
9. **Why selected:** keeps source-compatible convenience API without runtime panic.
10. **Rejected alternatives:** changing idiomatic `New(config)` nil/error behavior, which was not a defect.
11. **Trade-offs:** package tests/audit become the guard for default constant validity.
12. **Regression tests / protection:** configuration tests and strict audit forbidding `panic(` in `config.go`.
13. **Adversarial review findings:** `New` returning nil with error remains idiomatic and intentionally retained.
14. **Remediation iterations:** `6dbae4e6fe00295af0f7ba5303855736b76e8bde`.
15. **Residual risks and limitations:** deliberate source changes can still invalidate defaults, but CI catches them before merge.
16. **Operational or deployment consequences:** default construction no longer has an avoidable panic failure mode.
17. **Exact evidence:** remediation commit, configuration audit, run `30360637718` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** package-owned default configuration must be validated by tests/CI rather than defended with runtime panic.
