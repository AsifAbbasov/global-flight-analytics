# Validator Review Hardening

## Scope

This increment hardens `apps/api/internal/features/validator` as the production trust gate for Flight Feature snapshots at baseline `2872eb31e87500bdae1ae58fe2b75fb76c4b11d2`.

## Confirmed findings

The prior severity policy downgraded relationship and numerical integrity failures in partial groups to warnings. Since the Feature Pipeline persists both valid and limited snapshots, non-finite values and contradictory relationships could reach the Feature Store. Quality limitations were also merged with earlier validator output, allowing stale validation findings to survive a corrected revalidation.

## Rejected or stale findings

The current materialization path no longer rejects ordinary trajectories merely because `TrajectoryUpdatedAt` is later than `AsOfTime`; that validator rule is absent. The geographical schema already contains fifteen authoritative analytical fields, while `GeographicCellPrecision` is an intentional processing-configuration mirror outside schema counts. Longitude span reconciliation is already implemented through the circular longitude envelope policy. Returning a nil pointer with a non-nil constructor error remains idiomatic Go and is not a domain-integrity violation.

## Implemented policy

- Mathematical integrity is always error-severity, regardless of `available` or `partial` status.
- Evidence incompleteness remains warning-severity only when it is normalized and explained by a domain limitation.
- Partial and unavailable evidence without an explanation is invalid.
- Unavailable required groups must expose canonical zero-value payloads so stale or non-finite values cannot be hidden.
- Available geographical, operational, and trajectory groups require observation support.
- Operational ground and airborne shares are reconciled only when on-ground fields are claimed available by the typed limitation contract.
- Quality limitations are rebuilt from current group evidence on every validation pass before current validator issues are merged.
- Validator-owned stale limitations and stale group-derived aggregate limitations are not retained after source evidence is corrected.
- Numeric tolerance is dimensionless and relative; it is never added directly to values expressed in degrees, kilometres, metres, seconds, or ratios.
- Nil validation contexts are rejected.

## Version boundary

```text
Validator: flight-feature-validator-v5 -> flight-feature-validator-v6
Processing Pipeline: flight-feature-processing-pipeline-v11 -> flight-feature-processing-pipeline-v12
Schema: flight-features-v1 unchanged
PostgreSQL migration: not required
```

## Permanent evidence

The increment adds regression tests for partial mathematical corruption, stale limitation removal, unavailable residual payloads, zero-support available operational evidence, missing partial explanations, longitude envelope mismatch, and nil contexts. `tools/validatorreviewaudit` is part of Backend Quality in Continuous Integration.

## Review classification

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
VALIDATOR_REVIEW_STATUS=PENDING_EXACT_COMMIT_CI
```

## Canonical remediation history

Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, and Continuous Integration evidence. Historical exact-commit CI evidence for the original implementation is not reconstructed here; current permanent audits provide repository-state regression evidence. Severity labels are retrospective.

### GFA-DATA-206 — Partial availability downgraded mathematical integrity failures to warnings
1. **Finding / symptom:** non-finite values and contradictory numerical/relationship checks could be warning-severity merely because a group was `partial`.
2. **Root cause:** validation severity was coupled to evidence availability rather than to the nature of the integrity violation.
3. **Failure scenario:** a limited snapshot contains an impossible relationship or non-finite value, remains persistable, and reaches Feature Store.
4. **Impact:** mathematically invalid analytical evidence can be durably published.
5. **Severity rationale:** **P1 retrospective** because Validator is the production trust gate and the Store accepts limited snapshots.
6. **Existing guarantees violated:** mathematical integrity must fail closed independently of completeness.
7. **Considered solutions:** reject all partial groups, preserve warning downgrade, or separate evidence incompleteness from mathematical integrity severity.
8. **Chosen remediation:** mathematical integrity is always error-severity; explainable incompleteness may remain warning-severity.
9. **Why selected:** preserves useful limited snapshots without permitting corrupted mathematics.
10. **Rejected alternatives:** blanket rejection of all partial evidence and blanket warning downgrade.
11. **Trade-offs:** previously tolerated partial snapshots with real contradictions now fail validation.
12. **Regression tests / protection:** partial-corruption/non-finite/relationship tests and `validatorreviewaudit`.
13. **Adversarial review findings:** stale `UpdatedAt > AsOfTime`, schema-count, and circular-longitude findings were not reintroduced.
14. **Remediation iterations:** closed in `39549504bbeff1a6c272153bf3dcde469b766202`.
15. **Residual risks and limitations:** Validator correctness still depends on complete domain-specific relationship coverage.
16. **Operational or deployment consequences:** processing advances to v12; no migration.
17. **Exact evidence:** implementation commit, validator severity tests, permanent CI audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** validation severity must be classified by invariant type, not by feature availability status.

### GFA-DATA-207 — Partial or unavailable evidence could lack an explanatory domain limitation
1. **Finding / symptom:** a group could claim `partial`/`unavailable` without explaining why evidence was missing.
2. **Root cause:** status/count consistency was checked more strongly than evidence explainability.
3. **Failure scenario:** missing fields are accepted as limited evidence even though no provider/builder limitation justifies the absence.
4. **Impact:** silent data loss becomes indistinguishable from an understood limitation.
5. **Severity rationale:** **P1 retrospective** because unaccounted missing evidence can be persisted as a legitimate limited snapshot.
6. **Existing guarantees violated:** limited/unavailable data must be explainable and auditable.
7. **Considered solutions:** trust status, synthesize generic explanation, or require a non-validator domain limitation.
8. **Chosen remediation:** partial/unavailable evidence without an explanation is invalid.
9. **Why selected:** forces evidence-producing components to own their limitations rather than letting Validator invent provenance.
10. **Rejected alternatives:** generic fabricated limitation and permissive status-only validation.
11. **Trade-offs:** builders/providers must supply meaningful limitation metadata.
12. **Regression tests / protection:** missing-explanation tests and audit.
13. **Adversarial review findings:** limitation codes remain open-world evidence identifiers; no central closed registry was required.
14. **Remediation iterations:** `39549504bbeff1a6c272153bf3dcde469b766202`.
15. **Residual risks and limitations:** explanation presence does not prove the explanation is semantically perfect; source-specific tests remain necessary.
16. **Operational or deployment consequences:** malformed limited snapshots fail closed under v12.
17. **Exact evidence:** implementation commit and missing-limitation tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every non-available evidence state must carry a concrete domain-owned limitation.

### GFA-DATA-208 — Unavailable required groups could retain stale or non-finite residual payload values
1. **Finding / symptom:** a group marked unavailable could still carry non-zero/non-finite feature values.
2. **Root cause:** availability status and payload canonicalization were validated independently.
3. **Failure scenario:** consumers ignore status or stale values influence hashing/debugging while the group claims no evidence.
4. **Impact:** contradictory durable payload and hidden stale data.
5. **Severity rationale:** **P1 retrospective** because unavailable payloads can contain misleading analytical values.
6. **Existing guarantees violated:** unavailable required groups have canonical zero payloads.
7. **Considered solutions:** ignore values when unavailable, sanitize them in Validator, or reject noncanonical unavailable payloads.
8. **Chosen remediation:** Validator rejects unavailable required groups unless payload values are canonical zero values.
9. **Why selected:** fail-closed trust boundary avoids mutating source evidence during validation.
10. **Rejected alternatives:** silently zeroing values in Validator.
11. **Trade-offs:** upstream builders must construct canonical unavailable payloads correctly.
12. **Regression tests / protection:** residual/non-finite unavailable payload tests and audit.
13. **Adversarial review findings:** generic per-field Metric wrappers remain unnecessary; group evidence plus strict canonical payload is sufficient.
14. **Remediation iterations:** `39549504...`.
15. **Residual risks and limitations:** optional groups follow their explicit schema/evidence contract separately.
16. **Operational or deployment consequences:** invalid stale payloads are rejected rather than stored.
17. **Exact evidence:** implementation commit and unavailable-payload tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** availability transitions require tests for both evidence metadata and residual payload state.

### GFA-DATA-209 — Available observation-derived groups could claim zero supporting observations
1. **Finding / symptom:** Geographical, Operational, or Trajectory groups could be `available` while `SupportingPointCount == 0`.
2. **Root cause:** field-count availability was not reconciled with observation support.
3. **Failure scenario:** calculated-looking feature values are marked fully available without any supporting observation evidence.
4. **Impact:** unsupported analytical claims.
5. **Severity rationale:** **P1 retrospective** because available observation-derived output without observations is fabricated evidence.
6. **Existing guarantees violated:** available observational features require explicit support.
7. **Considered solutions:** allow zero for all groups, infer support from field count, or require positive support for observation-derived groups.
8. **Chosen remediation:** available Geographical/Operational/Trajectory groups require `SupportingPointCount > 0`.
9. **Why selected:** ties availability to real evidence while leaving non-observation groups under their own contracts.
10. **Rejected alternatives:** implicit support inference.
11. **Trade-offs:** zero-support fixtures and malformed builders fail earlier.
12. **Regression tests / protection:** zero-support available-group tests and audit.
13. **Adversarial review findings:** support ownership introduced in prior builders is now enforced at the trust gate.
14. **Remediation iterations:** `39549504...`.
15. **Residual risks and limitations:** support count does not encode per-field support distribution.
16. **Operational or deployment consequences:** stricter v12 persistence eligibility.
17. **Exact evidence:** implementation commit and support tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every new observation-derived group must declare its support invariant in Validator.

### GFA-DATA-210 — Ground/airborne share reconciliation ignored whether on-ground evidence was actually available
1. **Finding / symptom:** Validator could enforce share relationships even when typed Operational limitations explicitly said on-ground measurements were unavailable.
2. **Root cause:** cross-field validation did not consume the same availability contract as Operational Builder.
3. **Failure scenario:** an honestly limited Operational group is rejected for missing/zero ground shares that it never claimed to observe, or invalid claimed shares escape targeted reconciliation.
4. **Impact:** false validation failures or inconsistent ground-state integrity.
5. **Severity rationale:** **P2 retrospective** because the trust gate and builder could disagree about a typed availability contract.
6. **Existing guarantees violated:** Validator must reconcile only fields claimed available by source evidence.
7. **Considered solutions:** always reconcile shares, never reconcile partial groups, or make reconciliation conditional on typed on-ground limitations.
8. **Chosen remediation:** reconcile ground/airborne shares only when on-ground fields are claimed available.
9. **Why selected:** aligns validation with the explicit Operational availability contract.
10. **Rejected alternatives:** status-only heuristics.
11. **Trade-offs:** Validator depends on typed limitation semantics for this conditional relationship.
12. **Regression tests / protection:** on-ground unavailable/available share tests and audit.
13. **Adversarial review findings:** typed limitation contract was preferred over copying complete upstream DataQuality.
14. **Remediation iterations:** `39549504...`.
15. **Residual risks and limitations:** limitation-code changes must remain coordinated with Validator tests.
16. **Operational or deployment consequences:** fewer false failures and stronger claimed-share checks.
17. **Exact evidence:** implementation commit and Operational share validation tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** conditional cross-field validation must key off canonical typed availability evidence.

### GFA-DATA-211 — Revalidation retained stale Validator and group-derived quality limitations
1. **Finding / symptom:** corrected evidence could still carry limitations/issues produced by an earlier validation pass.
2. **Root cause:** quality limitations were merged incrementally instead of rebuilt from current source evidence.
3. **Failure scenario:** a defect is fixed, snapshot revalidated, but stale warning/error limitations remain and continue to mark output limited.
4. **Impact:** validation result is history-dependent rather than a pure function of current evidence/policy.
5. **Severity rationale:** **P1 retrospective** because durable validation proof could misrepresent current snapshot integrity.
6. **Existing guarantees violated:** deterministic revalidation and current-state trust evidence.
7. **Considered solutions:** preserve all history, selectively remove validator codes, or rebuild current group limitations then merge current validator findings.
8. **Chosen remediation:** reconstruct quality limitations from current group evidence; strip stale validator-owned/group-derived aggregate limitations before current issues are added.
9. **Why selected:** validation becomes repeatable and current-state based while preserving genuine source limitations.
10. **Rejected alternatives:** append-only validation history inside the snapshot payload.
11. **Trade-offs:** historical validation events are not retained in current quality limitations; audit history belongs elsewhere.
12. **Regression tests / protection:** stale limitation removal/revalidation tests and audit.
13. **Adversarial review findings:** durable Store validation report remains the proof of the current completed validation pass.
14. **Remediation iterations:** `39549504...`.
15. **Residual risks and limitations:** external audit/event history requires separate operational logging if needed.
16. **Operational or deployment consequences:** revalidated snapshots no longer inherit obsolete limitation state.
17. **Exact evidence:** implementation commit and stale-limitation tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** validators must be idempotent pure functions of current payload plus explicit policy, not prior validator output.

### GFA-DATA-212 — Numeric tolerance was treated as an absolute quantity across incompatible units
1. **Finding / symptom:** the same tolerance could be added directly to degrees, kilometres, metres, seconds, and ratios.
2. **Root cause:** dimensionless numerical comparison tolerance was implemented as an absolute unit-bearing offset.
3. **Failure scenario:** one tolerance value is far too permissive for one metric and too strict for another, hiding or inventing relationship failures.
4. **Impact:** unit-dependent false passes/failures in the trust gate.
5. **Severity rationale:** **P1 retrospective** because integrity relationships could be accepted/rejected using dimensionally invalid math.
6. **Existing guarantees violated:** numerical comparisons must have dimensionally coherent tolerance semantics.
7. **Considered solutions:** per-unit absolute tolerances, zero tolerance, or dimensionless relative comparison.
8. **Chosen remediation:** `NumericTolerance` is explicitly dimensionless; relationship helpers use relative approximate equality.
9. **Why selected:** one policy can apply across units without pretending the tolerance carries physical dimensions.
10. **Rejected alternatives:** direct absolute addition to unit-bearing values.
11. **Trade-offs:** relative tolerance behavior near zero requires explicit tests and comparison helper semantics.
12. **Regression tests / protection:** unit-scale/relationship/tolerance tests and audit.
13. **Adversarial review findings:** path-ratio tolerance from Trajectory remains a separate local numerical-bound policy where appropriate.
14. **Remediation iterations:** `39549504...`.
15. **Residual risks and limitations:** future relationships may require domain-specific absolute floors in addition to relative tolerance; those must be explicit.
16. **Operational or deployment consequences:** validator v6 / processing v12 isolate corrected comparisons.
17. **Exact evidence:** implementation commit, relative-tolerance helpers/tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every tolerance must state whether it is relative, dimensionless, or expressed in a named physical unit.

### GFA-OPS-213 — Validator silently accepted a nil caller context
1. **Finding / symptom:** validation could proceed without an explicit caller-owned context.
2. **Root cause:** lifecycle ownership was not enforced at the trust-gate entry point.
3. **Failure scenario:** a caller bug invokes validation with nil and loses consistent cancellation/deadline semantics.
4. **Impact:** lifecycle inconsistency and potential unbounded validation work on large snapshots.
5. **Severity rationale:** **P2 retrospective** as an operational/lifecycle correctness defect.
6. **Existing guarantees violated:** explicit context ownership across production processing stages.
7. **Considered solutions:** background fallback, panic, or typed error.
8. **Chosen remediation:** nil validation contexts return `ErrContextRequired`.
9. **Why selected:** fail-fast without panics and consistent with provider/store/builder contracts.
10. **Rejected alternatives:** implicit `context.Background()`.
11. **Trade-offs:** invalid callers must be corrected explicitly.
12. **Regression tests / protection:** nil-context tests and `validatorreviewaudit`.
13. **Adversarial review findings:** idiomatic constructors returning `nil,error` remain unrelated and were correctly rejected as a finding.
14. **Remediation iterations:** `39549504bbeff1a6c272153bf3dcde469b766202`.
15. **Residual risks and limitations:** caller still owns appropriate deadline selection.
16. **Operational or deployment consequences:** consistent lifecycle semantics; no migration.
17. **Exact evidence:** implementation commit, nil-context tests, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** all trust-gate entry points must reject nil contexts and test cancellation ownership explicitly.
