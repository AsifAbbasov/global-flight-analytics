# Historical Route Review Hardening

Status: closed

## Scope

This increment hardens `apps/api/internal/historicalintelligence/historicalroute` as the validated historical analytics boundary over persisted Route Intelligence results.

## Accepted findings

- A complete airport-pair scope cannot classify valid partial or unavailable route results because those statuses deliberately lack a complete pair identity.
- Historical Route used the compatibility-oriented partial decoder instead of the complete Route Intelligence contract validator.
- Persistence metadata was not reconciled with the decoded payload before analytics.
- `StoredAt` affected source freshness but was absent from the result fingerprint.
- A bounded global route denominator cannot be reused as route-pair coverage evidence.
- Snapshot query evidence was not checked against the canonical analytical plan.
- Scoped source freshness and source names were not derived from the scoped validated evidence.
- Persisted route distance was trusted instead of being recomputed from validated endpoint coordinates.
- Confidence and distance accumulation lacked compensated floating-point summation.
- Active route semantics, route-confidence semantics, and zero-denominator evidence were under-documented.
- Latest-record selection accepted a record identifier as a substitute for missing trajectory identity.
- Metric calculation responsibilities were concentrated in one large switch.

## Corrected contracts

- Route status ratios are global metrics only. A route-pair scope is rejected by the central metric catalog and the Historical Route builder.
- Every selected payload passes `routecontract.Validate` before it can contribute to a metric.
- Persisted trajectory identity, status, confidence level, input fingerprint, as-of time, event window, validation warning count, storage time, and payload fingerprint are reconciled with the payload.
- Invalid JSON, unsupported schemas, invalid endpoint cardinality, endpoint-role errors, invalid airports, future evidence, provenance failures, and metadata mismatches fail closed with typed causes.
- A production snapshot query window must contain the canonical plan effective window and use the same analytical `AsOfTime`. A larger previous-plus-current materialization window remains valid.
- Incomplete global reads require the exact `RouteMatchedCount`. Incomplete route-pair reads are rejected because no pair-specific denominator exists.
- Fingerprint generation two binds every output-affecting selected record, including `StoredAt`, and binds limit state plus exact denominator evidence when the read is incomplete.
- Complete route distance is recomputed with a documented haversine great-circle policy over validated endpoint coordinates.
- Provenance source names are the sorted union of the route dataset and actual Route Contract sources. Latest source time is calculated only from scoped validated evidence.
- `active_routes` is explicitly a count of unique directional complete route pairs and uses the unit `route_pairs`. Sample count remains the number of validated route results supporting the distinct count.
- Route confidence is the compensated unweighted mean of validated result-level confidence. Evidence quantity and quality are already represented inside each Route Contract confidence score and are not counted twice.
- Status-ratio buckets with no observations retain value `0` together with sample count `0` and complete bucket coverage, so zero numerator and absent denominator remain distinguishable.
- Latest selection requires real trajectory identity and uses `AsOfTime`, `StoredAt`, then stable record identifier ordering.
- Metric calculation uses a calculator registry and focused evidence, selection, scope, coverage, fingerprint, provenance, and limitation helpers.

```text
ROUTE_PAIR_STATUS_RATIOS=GLOBAL_ONLY
ROUTE_CONTRACT_VALIDATION=ENFORCED
PERSISTENCE_METADATA_RECONCILIATION=ENFORCED
ROUTE_SCOPE_INCOMPLETE_COVERAGE=REJECTED
SNAPSHOT_PLAN_WINDOW_CONTAINMENT=ENFORCED
ROUTE_FINGERPRINT_STORED_AT=BOUND
ROUTE_DISTANCE_COORDINATE_RECOMPUTATION=ENFORCED
ROUTE_PROVENANCE_SCOPED=ENFORCED
ACTIVE_ROUTES_SEMANTICS=UNIQUE_DIRECTIONAL_ROUTE_PAIRS
HISTORICAL_ROUTE_ENGINEERING_REMEDIATION=IMPLEMENTED
```

## Rejected, stale, or deliberately retained findings

- The claim that `RouteLimitReached` is absent from the fingerprint was stale. The existing builder already bound both row-limit and byte-limit flags. The new fingerprint retains them and adds `StoredAt`.
- Production Historical Read already exposes exact matched-row counts. The legacy `selected/(selected+1)` fallback is not used by the hardened Historical Route builder; incomplete reads without an exact count are rejected.
- A complete zero-event bucket remains a valid Historical Contract state. Historical Route does not invent source freshness when no scoped route evidence exists; it fails closed with `ErrRouteSourceEvidenceUnavailable` instead.
- The heterogeneous `float64` transport remains intentional. The central metric catalog enforces exact integral count values through the exact IEEE-754 integer boundary and validates ratios on `[0,1]` with a fixed absolute tolerance.
- Confidence is not reweighted by evidence count. Route Contract confidence already incorporates evidence quantity, weight, and quality; weighting it again would double-count the same evidence.
- Generic summary `Total` remains descriptive. Ratio comparisons are bound by the metric catalog to `Summary.Average`, not to the temporal sum of ratios.
- Raw persistence JSON decoding already belongs to Historical Read rather than Historical Route. The strict validator is added at that boundary instead of reintroducing JSON parsing into the domain builder.
- Direct navigation through the adjacent Route Contract value object is not treated as a Law of Demeter defect.
- Similar helper shapes in Historical Airport and Historical Route do not by themselves justify a shared abstraction because their identities, filters, and coverage semantics differ.

## Permanent verification

`apps/api/tools/historicalroutereviewaudit` enforces the version-two builder boundary, global-only status-ratio scopes, strict Route Contract validation, metadata reconciliation, exact incomplete coverage, `StoredAt` fingerprint identity, scoped provenance, compensated arithmetic, focused calculators, regression tests, and this engineering remediation record in Backend Continuous Integration.

## Formal closure evidence

The Historical Route engineering remediation was committed and validated before
this administrative closure:

```text
ENGINEERING_BASELINE_COMMIT=9741c4fce04e2b2c06ee0236cf13b5c384f38ffd
ENGINEERING_REMEDIATION_COMMIT=513fa1efc7f3b81b895cdc5f881e294d80362e2e
ENGINEERING_GITHUB_ACTIONS_RUN=30334131538
Backend Quality=SUCCESS
Backend Quality Job=90195300495
PostgreSQL 16 Integration=SUCCESS
PostgreSQL 16 Integration Job=90195300516
Backend Race Safety=SUCCESS
Backend Race Safety Job=90195300546
Backend Container=SUCCESS
Backend Container Job=90195525282
```

The accepted findings are fully implemented, stale or rejected findings retain
their documented disposition, and no Historical Route review item remains open,
unclassified, or deferred.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
HISTORICAL_ROUTE_ENGINEERING_DEBT=CLOSED
HISTORICAL_ROUTE_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
HISTORICAL_ROUTE_REVIEW_STATUS=CLOSED
```

## Canonical remediation history

Historical reviewer identities and review-comment chronology are not reconstructed. The records below are reconstructed from repository source/tests, the accepted/rejected review dispositions, remediation commit `513fa1efc7f3b81b895cdc5f881e294d80362e2e`, permanent audit, and exact Backend CI run `30334131538`. Severity labels are retrospective.

### GFA-CONTRACT-248 — Route status-ratio metrics admitted an incompatible airport-pair scope
1. **Finding / symptom:** route status ratios could be requested for a complete airport-pair scope even though partial/unavailable route results intentionally lack complete pair identity.
2. **Root cause:** scope permission was broader than the evidence identity required by the metric.
3. **Failure scenario:** incomplete route evidence cannot be assigned honestly to a requested pair, so ratio classification either drops evidence or fabricates pair membership.
4. **Impact:** scoped status ratios can become analytically misleading.
5. **Severity rationale:** **P1 retrospective** because the metric could claim pair-level evidence that the source contract does not provide.
6. **Existing guarantees violated:** scope eligibility must be supported by complete evidence identity.
7. **Considered solutions:** infer pair identity, exclude incomplete statuses, add pair-specific incomplete identity, or make status ratios global-only.
8. **Chosen remediation:** central metric catalog and Historical Route reject route-pair scope for status-ratio metrics.
9. **Why selected:** matches the actual Route Contract without inventing missing identity.
10. **Rejected alternatives:** inferred pair assignment and silent incomplete-record exclusion.
11. **Trade-offs:** pair-scoped status ratios are unavailable until a future source contract supplies honest pair identity for all statuses.
12. **Regression tests / protection:** scope-permission and global-only ratio tests plus `historicalroutereviewaudit`.
13. **Adversarial review findings:** complete zero-event global buckets remain valid and are not a scope defect.
14. **Remediation iterations:** `513fa1efc7f3b81b895cdc5f881e294d80362e2e`.
15. **Residual risks and limitations:** product consumers must use global status ratios or complete-pair metrics with valid identity.
16. **Operational or deployment consequences:** invalid pair requests fail closed; no migration.
17. **Exact evidence:** remediation commit, run `30334131538` SUCCESS, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every new metric catalog entry must prove that each allowed scope can be reconstructed from every contributing evidence status.

### GFA-DATA-249 — Historical Route accepted compatibility-decoded payloads without full Route Contract validation
1. **Finding / symptom:** selected route payloads could pass a partial compatibility decoder instead of the complete domain validator.
2. **Root cause:** read compatibility and analytical trust validation were conflated.
3. **Failure scenario:** structurally decodable but semantically invalid endpoints, airports, provenance, future evidence or schema content contributes to a historical metric.
4. **Impact:** invalid Route Intelligence evidence can enter durable Historical analytics.
5. **Severity rationale:** **P1 retrospective** because the historical builder consumed untrusted domain payloads.
6. **Existing guarantees violated:** every analytical input must pass the authoritative source-domain contract.
7. **Considered solutions:** expand the partial decoder, validate selected payloads in Historical Route, or trust persistence.
8. **Chosen remediation:** every selected payload passes `routecontract.Validate`; typed causes fail closed.
9. **Why selected:** preserves decoding ownership in Historical Read while enforcing domain semantics before analytics.
10. **Rejected alternatives:** reintroducing raw JSON parsing into the builder and trusting stored JSON by construction.
11. **Trade-offs:** legacy malformed payloads can no longer contribute.
12. **Regression tests / protection:** unsupported schema, endpoint/cardinality/airport/provenance/future-evidence validation tests and strict audit.
13. **Adversarial review findings:** raw JSON decoding remains correctly owned by Historical Read.
14. **Remediation iterations:** `513fa1efc7f3b81b895cdc5f881e294d80362e2e`.
15. **Residual risks and limitations:** source-domain validator evolution must remain version-aware.
16. **Operational or deployment consequences:** invalid historical route rows fail closed at analytics time.
17. **Exact evidence:** remediation commit, route validation tests, run `30334131538` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** compatibility decoding must never be treated as equivalent to authoritative domain validation.

### GFA-DATA-250 — Persisted Route metadata was not reconciled with the validated payload
1. **Finding / symptom:** denormalized persistence identity/status/provenance fields could disagree with decoded Route Contract data without being rejected.
2. **Root cause:** analytics trusted JSON and row metadata as independent compatible views.
3. **Failure scenario:** stale or corrupted row columns select/filter one identity while payload semantics describe another.
4. **Impact:** historical selection, confidence, scope and provenance can be based on contradictory evidence.
5. **Severity rationale:** **P1 retrospective** because persistence contradiction could alter which records contribute.
6. **Existing guarantees violated:** denormalized persistence mirrors must reconcile with canonical payload identity.
7. **Considered solutions:** trust JSON only, trust row columns only, or reconcile every output-affecting field.
8. **Chosen remediation:** reconcile trajectory identity, status, confidence, input fingerprint, as-of/event window, warning count, storage time and payload fingerprint.
9. **Why selected:** detects corruption/drift at the persistence-to-domain trust boundary.
10. **Rejected alternatives:** precedence rules that silently choose one contradictory representation.
11. **Trade-offs:** legacy inconsistent rows become unreadable for analytics until repaired/rematerialized.
12. **Regression tests / protection:** metadata mismatch tests and permanent route audit.
13. **Adversarial review findings:** the fix does not move PostgreSQL concerns into the domain contract; reconciliation remains at the adapter/builder boundary.
14. **Remediation iterations:** `513fa1efc7f3b81b895cdc5f881e294d80362e2e`.
15. **Residual risks and limitations:** fields not mirrored in persistence cannot be cross-checked at this layer.
16. **Operational or deployment consequences:** contradictory rows fail closed.
17. **Exact evidence:** remediation commit, metadata tests, run `30334131538` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every new denormalized Route field used for selection or provenance requires mirror reconciliation coverage.

### GFA-DATA-251 — `StoredAt` affected Historical Route output but was absent from fingerprint identity
1. **Finding / symptom:** storage time influenced latest selection/source freshness while generation-one fingerprint did not bind it.
2. **Root cause:** fingerprint identity omitted one output-affecting persistence field.
3. **Failure scenario:** two semantically different selected-record sets can share the same input fingerprint even though `StoredAt` changes latest selection or provenance time.
4. **Impact:** replay/idempotency identity does not fully describe the produced result.
5. **Severity rationale:** **P1 retrospective** because immutable analytical identity could collide across different evidence.
6. **Existing guarantees violated:** fingerprints must bind every output-affecting input.
7. **Considered solutions:** remove `StoredAt` from semantics, bind it only to provenance, or include it in the fingerprint.
8. **Chosen remediation:** fingerprint generation two binds `StoredAt` for every selected record.
9. **Why selected:** preserves current selection/freshness semantics and restores deterministic identity.
10. **Rejected alternatives:** keeping an output-affecting field outside identity.
11. **Trade-offs:** old and corrected fingerprints are intentionally different generations.
12. **Regression tests / protection:** `StoredAt` mutation/fingerprint tests and strict audit.
13. **Adversarial review findings:** row-limit and byte-limit signals were already fingerprinted; that stale finding was not duplicated.
14. **Remediation iterations:** `513fa1efc7f3b81b895cdc5f881e294d80362e2e`.
15. **Residual risks and limitations:** any future selection/provenance input must be added deliberately to fingerprint policy.
16. **Operational or deployment consequences:** corrected identity may conflict with pre-generation-two immutable rows and must be governed rather than silently overwritten.
17. **Exact evidence:** remediation commit, fingerprint tests, run `30334131538` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** fingerprint review is mandatory when any selected-record field begins affecting output.

### GFA-DATA-252 — Global route matched-count evidence could be reused as route-pair coverage
1. **Finding / symptom:** a bounded global denominator could be treated as evidence for an airport-pair subset.
2. **Root cause:** read completeness was modeled at dataset level while the builder supported narrower filters.
3. **Failure scenario:** global `RouteMatchedCount` makes one route pair look well or poorly covered even though no pair-specific denominator exists.
4. **Impact:** false pair-scoped coverage and confidence.
5. **Severity rationale:** **P1 retrospective** because a completeness metric could be mathematically valid but scoped to the wrong population.
6. **Existing guarantees violated:** coverage denominator must describe exactly the selected analytical scope.
7. **Considered solutions:** derive pair denominator from selected rows, estimate from global count, add pair-specific query counts, or reject incomplete pair reads.
8. **Chosen remediation:** incomplete global reads require exact `RouteMatchedCount`; incomplete pair reads fail closed without pair-specific denominator.
9. **Why selected:** avoids statistical inference not supported by the read contract.
10. **Rejected alternatives:** legacy selected/(selected+1) approximation and global-denominator reuse.
11. **Trade-offs:** incomplete pair analytics are unavailable until exact pair counts exist.
12. **Regression tests / protection:** global/pair incomplete coverage tests and route audit.
13. **Adversarial review findings:** Production Historical Read already provides exact global matched counts; the old fallback is not used.
14. **Remediation iterations:** `513fa1efc7f3b81b895cdc5f881e294d80362e2e`.
15. **Residual risks and limitations:** exact pair coverage remains intentionally unsupported for bounded incomplete reads.
16. **Operational or deployment consequences:** stricter fail-closed behavior for pair requests.
17. **Exact evidence:** remediation commit, coverage tests, run `30334131538` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** denominator evidence must carry the same scope identity as the metric it qualifies.

### GFA-DATA-253 — Historical Route snapshot query evidence was not bound to the canonical plan
1. **Finding / symptom:** builder input could come from a snapshot whose query window/as-of semantics did not prove compatibility with the requested plan.
2. **Root cause:** snapshot version was trusted without full query-to-plan reconciliation.
3. **Failure scenario:** rows from the wrong analytical window or `AsOfTime` contribute to a result carrying another plan identity.
4. **Impact:** time-travel and replay correctness can be violated.
5. **Severity rationale:** **P1 retrospective** because historical evidence can be assigned to the wrong temporal request.
6. **Existing guarantees violated:** source snapshot must contain the plan effective window and share the same analytical as-of time.
7. **Considered solutions:** require exact query equality, trust the caller, or allow a containing materialization window with exact as-of match.
8. **Chosen remediation:** production snapshot query must contain the canonical effective window and use identical `AsOfTime`; previous-plus-current supersets remain valid.
9. **Why selected:** supports atomic materialization reads without weakening temporal identity.
10. **Rejected alternatives:** exact-window-only restriction and caller trust.
11. **Trade-offs:** query validation logic is explicit and must evolve with materialization contracts.
12. **Regression tests / protection:** window containment/as-of mismatch tests and strict audit.
13. **Adversarial review findings:** larger previous-plus-current reads are deliberately valid and not over-fetch defects by themselves.
14. **Remediation iterations:** `513fa1efc7f3b81b895cdc5f881e294d80362e2e`.
15. **Residual risks and limitations:** correctness depends on Historical Read truthfully reporting normalized query evidence.
16. **Operational or deployment consequences:** mismatched snapshots are rejected before metric calculation.
17. **Exact evidence:** remediation commit, plan reconciliation tests, run `30334131538` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every historical builder must validate snapshot query identity against its canonical plan before consuming rows.

### GFA-DATA-254 — Historical Route provenance was derived from unscoped rather than contributing evidence
1. **Finding / symptom:** source names/latest source time could reflect records outside the final scoped validated set.
2. **Root cause:** provenance aggregation occurred before or independently of filtering/validation.
3. **Failure scenario:** a route-pair/global metric claims a source or freshness timestamp contributed by a row that did not support the metric.
4. **Impact:** provenance overstates evidence and can distort freshness.
5. **Severity rationale:** **P1 retrospective** because provenance is a trust claim about actual contributors.
6. **Existing guarantees violated:** provenance must be computed from the exact evidence set supporting output.
7. **Considered solutions:** dataset-wide provenance, selected-row provenance, or route-contract-only provenance.
8. **Chosen remediation:** source names are the sorted union of dataset and actual contributing Route Contract sources; latest source time is scoped to validated contributing evidence.
9. **Why selected:** preserves dataset origin while preventing unrelated records from influencing freshness.
10. **Rejected alternatives:** global snapshot freshness and invented freshness when no scoped evidence exists.
11. **Trade-offs:** zero-evidence scopes fail with explicit source-evidence unavailability instead of a convenient timestamp.
12. **Regression tests / protection:** scoped provenance/freshness and zero-evidence tests plus audit.
13. **Adversarial review findings:** complete zero-event buckets remain valid only where source coverage is independently evidenced.
14. **Remediation iterations:** `513fa1efc7f3b81b895cdc5f881e294d80362e2e`.
15. **Residual risks and limitations:** source naming quality still depends on Route Contract provenance correctness.
16. **Operational or deployment consequences:** provenance failures are explicit.
17. **Exact evidence:** remediation commit, provenance tests, run `30334131538` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** provenance aggregation must happen after final eligibility/scope decisions.

### GFA-DATA-255 — Historical Route trusted persisted distance instead of validated endpoint geometry
1. **Finding / symptom:** stored route distance could be consumed without recomputing it from validated origin/destination coordinates.
2. **Root cause:** a derived persistence field was treated as authoritative analytical evidence.
3. **Failure scenario:** stale/corrupt distance disagrees with valid endpoints and biases route-distance metrics.
4. **Impact:** incorrect distance analytics despite trustworthy coordinate evidence.
5. **Severity rationale:** **P1 retrospective** because a persisted derived value could override canonical geometric evidence.
6. **Existing guarantees violated:** derived analytics should be reconstructed from validated source evidence when deterministic and cheap.
7. **Considered solutions:** trust persisted distance, reconcile with tolerance, or recompute from endpoints under one documented policy.
8. **Chosen remediation:** recompute complete route distance using documented haversine great-circle policy over validated endpoints.
9. **Why selected:** removes stale derived-state dependency and makes formula reproducible.
10. **Rejected alternatives:** silent persisted-value precedence.
11. **Trade-offs:** small computation cost per selected route.
12. **Regression tests / protection:** endpoint/distance recomputation tests and strict audit.
13. **Adversarial review findings:** cross-module geodesy consolidation is not required to close this local correctness defect.
14. **Remediation iterations:** `513fa1efc7f3b81b895cdc5f881e294d80362e2e`.
15. **Residual risks and limitations:** Haversine remains a spherical approximation by documented policy.
16. **Operational or deployment consequences:** no migration; stored distance ceases to be trusted input for this metric.
17. **Exact evidence:** remediation commit, distance tests, run `30334131538` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** persisted derived values require reconciliation or recomputation before use at analytical trust boundaries.

### GFA-DATA-256 — Historical Route accumulation used ordinary floating-point summation
1. **Finding / symptom:** confidence and distance totals accumulated with ordinary addition.
2. **Root cause:** numerical stability policy was not explicit for repeated continuous-value aggregation.
3. **Failure scenario:** long series with heterogeneous magnitudes accumulates avoidable binary64 rounding drift.
4. **Impact:** small non-deterministic-looking numerical differences can propagate into averages/fingerprints/validation boundaries.
5. **Severity rationale:** **P2 retrospective** because error is bounded but avoidable in durable analytical calculations.
6. **Existing guarantees violated:** stable reproducible numerical aggregation.
7. **Considered solutions:** ordinary sum, output rounding, decimal arithmetic, or compensated summation.
8. **Chosen remediation:** compensated summation for confidence and distance accumulation.
9. **Why selected:** improves stability with minimal cost and no schema change.
10. **Rejected alternatives:** domain rounding and decimal-library migration for non-monetary analytics.
11. **Trade-offs:** slightly more arithmetic/state per accumulator.
12. **Regression tests / protection:** arithmetic/mean tests and permanent route audit.
13. **Adversarial review findings:** confidence is deliberately not reweighted by evidence count to avoid double counting Route Contract evidence.
14. **Remediation iterations:** `513fa1efc7f3b81b895cdc5f881e294d80362e2e`.
15. **Residual risks and limitations:** binary64 model error remains; compensation addresses accumulation, not source measurement uncertainty.
16. **Operational or deployment consequences:** none beyond corrected calculations.
17. **Exact evidence:** remediation commit, numerical tests, run `30334131538` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** repeated continuous-value sums in durable analytics should use the project’s compensated-aggregation policy.

### GFA-CONTRACT-257 — Historical Route metric semantics were insufficiently explicit
1. **Finding / symptom:** `active_routes`, route confidence, and zero-denominator status-ratio behavior could be interpreted differently by maintainers/consumers.
2. **Root cause:** metric formulas existed without one canonical semantic statement for distinct-count identity, weighting, units and empty-denominator evidence.
3. **Failure scenario:** a refactor counts route results instead of directional route pairs, reweights confidence by sample count, or treats zero observations as unavailable despite complete coverage.
4. **Impact:** semantic drift without a type/schema change.
5. **Severity rationale:** **P2 retrospective** because ambiguity could change durable analytical meaning even when code remains mathematically valid.
6. **Existing guarantees violated:** versioned metrics require explicit formula and evidence semantics.
7. **Considered solutions:** infer from implementation, add new metrics, or document/enforce current intended definitions.
8. **Chosen remediation:** define `active_routes` as unique directional complete pairs with unit `route_pairs`; confidence as compensated unweighted mean of validated result confidence; zero-observation status ratios retain zero value/sample count with complete coverage.
9. **Why selected:** preserves established product schema while eliminating interpretive ambiguity.
10. **Rejected alternatives:** evidence-count reweighting, which would double-count evidence already represented by Route Contract confidence.
11. **Trade-offs:** semantics are deliberately specific and require a new version for future formula changes.
12. **Regression tests / protection:** active-route distinct-count, confidence and zero-denominator tests plus audit.
13. **Adversarial review findings:** generic summary `Total` remains descriptive; ratio comparisons use catalog-selected `Average`.
14. **Remediation iterations:** `513fa1efc7f3b81b895cdc5f881e294d80362e2e`.
15. **Residual risks and limitations:** consumers must respect directional pair identity; reverse routes are distinct.
16. **Operational or deployment consequences:** none; semantic contract clarification and enforcement.
17. **Exact evidence:** remediation commit, semantic tests/audit, run `30334131538` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every historical metric must state identity, denominator, weighting, unit and empty-evidence behavior in one versioned contract.

### GFA-DATA-258 — Latest Route selection could substitute record identity for missing trajectory identity
1. **Finding / symptom:** a stable record identifier could stand in for absent trajectory identity during latest-record grouping.
2. **Root cause:** selection attempted to remain permissive when the actual entity key was unavailable.
3. **Failure scenario:** multiple records for one trajectory are treated as unrelated, allowing stale/duplicate versions to contribute simultaneously.
4. **Impact:** duplicated or incorrect route metrics.
5. **Severity rationale:** **P1 retrospective** because entity-version selection can fail silently and alter counts/means.
6. **Existing guarantees violated:** latest-version selection requires the real domain entity identity.
7. **Considered solutions:** fallback to record ID, hash payload, or reject records without trajectory identity.
8. **Chosen remediation:** require real trajectory identity; order versions by `AsOfTime`, `StoredAt`, then stable record ID.
9. **Why selected:** grouping follows the domain entity while record ID remains only a deterministic tie-breaker.
10. **Rejected alternatives:** record-ID substitution and payload-derived pseudo identity.
11. **Trade-offs:** rows missing trajectory identity are no longer analytically usable.
12. **Regression tests / protection:** missing-trajectory and latest-selection ordering tests plus audit.
13. **Adversarial review findings:** route calculation time is not used as entity identity or event membership.
14. **Remediation iterations:** `513fa1efc7f3b81b895cdc5f881e294d80362e2e`.
15. **Residual risks and limitations:** correctness depends on persistence preserving canonical trajectory identity.
16. **Operational or deployment consequences:** malformed legacy rows fail closed.
17. **Exact evidence:** remediation commit, selection tests, run `30334131538` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** stable storage record IDs may break ties but must never replace missing domain entity keys.

### GFA-MAINT-259 — Historical Route metric calculation responsibilities were concentrated in one switch
1. **Finding / symptom:** one large dispatch path owned metric selection plus evidence-specific calculations.
2. **Root cause:** new metrics accumulated as branches in orchestration rather than focused calculators.
3. **Failure scenario:** adding or changing one metric unintentionally alters another metric’s scope, denominator or provenance flow.
4. **Impact:** elevated maintenance and regression risk.
5. **Severity rationale:** **P3 retrospective** because this is structural debt after correctness invariants are separately addressed.
6. **Existing guarantees violated:** focused analytical-policy ownership and testability.
7. **Considered solutions:** retain switch, create generic reflection dispatch, or introduce a typed calculator registry with focused helpers.
8. **Chosen remediation:** calculator registry plus focused evidence, selection, scope, coverage, fingerprint, provenance and limitation helpers.
9. **Why selected:** reduces coupling while keeping domain policy explicit.
10. **Rejected alternatives:** cross-module helper abstraction based only on similar code shape.
11. **Trade-offs:** more internal types/files and registry wiring.
12. **Regression tests / protection:** calculator dispatch tests and `historicalroutereviewaudit`.
13. **Adversarial review findings:** a finite enumeration is not automatically an Open/Closed Principle defect; refactoring is justified here by concentrated domain responsibility.
14. **Remediation iterations:** `513fa1efc7f3b81b895cdc5f881e294d80362e2e`.
15. **Residual risks and limitations:** registry additions still require deliberate metric-catalog coordination.
16. **Operational or deployment consequences:** none; internal maintainability hardening.
17. **Exact evidence:** remediation commit, permanent audit, run `30334131538` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** new Historical Route metrics should add focused calculator policy rather than expand orchestration conditionals.
