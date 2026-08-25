# Geographical Builder Review Hardening

## Scope

This increment closes the static review of `apps/api/internal/features/geographicalbuilder` against baseline commit `c0e3323328f81af8bf0b8841b1bf6756d3085d21`.

## Confirmed findings

### Chronological point evidence

Production point coordinates are now eligible only when their observation timestamps are present and fall inside the authoritative trajectory start/end window. Eligible points are ordered deterministically by observation timestamp, point identity and original index before endpoint, path and antimeridian calculations.

`AsOfTime` remains owned by Extractor. Extractor rejects nested point, segment and coverage-gap event times after the historical cutoff before any feature builder runs. Geographical Builder therefore does not duplicate the cross-builder historical request contract.

### Segment fallback path semantics

When fewer than two usable point observations exist, non-invalid segment endpoints may provide fallback geometry. The fallback now maintains two separate structures:

- an ordered coordinate envelope for endpoints, bounds, displacement and geographic cells;
- observed segment edges for path distance and antimeridian path crossing.

Distances between disconnected segments are never added to `ObservedPathDistanceKM`. Segment discontinuities are counted and surfaced as limitations. Supporting point count comes from authoritative trajectory or segment metadata rather than the number of endpoint coordinates.

### Circular longitude envelope

`MinimumLongitude` and `MaximumLongitude` remain wire-compatible field names, but they are explicitly documented as western and eastern circular envelope bounds. A western bound greater than the eastern bound is valid and means the envelope wraps the antimeridian. This envelope property is independent from whether the chronological observed path crosses the antimeridian.

Validator now derives the declared circular span from the two envelope bounds and no longer treats wrapped bounds as evidence of a path crossing.

### Numeric policy

Distance calculations remain based on the mean-Earth spherical Haversine model with radius `6371.0088` kilometres. This deliberate analytical approximation is versioned as `mean-earth-sphere-haversine-v1`.

Observed path accumulation uses Kahan compensated summation. Outputs remain unrounded binary64 analytical values.

Geographic cells remain decimal-degree buckets using `math.Round`, whose half values round away from zero. They are not equal-area physical cells. The policy is versioned as `decimal-degree-round-half-away-from-zero-v1`.

### Context and diagnostics

Nil contexts are rejected. Long coordinate collection and geometry passes observe cancellation. Invalid-status segments, invalid coordinates, missing timestamps, out-of-window evidence and segment discontinuities produce exact-count limitations.

## Stale or rejected findings

### Schema mismatch

The four geographic bounds are already registered in schema version one. `GeographicCellPrecision` is a processing-configuration mirror whose authoritative value is stored in Processing Identity, so it is intentionally excluded from the fifteen analytical geographical fields.

### Builder-level AsOf request

A separate `BuildRequest` carrying `AsOfTime` was not introduced. Extractor already owns and enforces the historical cutoff for every nested evidence type. Repeating that policy in each builder would create divergent cross-builder logic.

### Precision zero as a real cell precision

Zero remains an explicit configuration sentinel selecting the default precision of two. Effective precision is intentionally constrained to one through six because Validator and Processing Identity require a positive analytical precision. The previous error text was corrected to describe this contract accurately.

### Field renaming

`MinimumLongitude` and `MaximumLongitude` were not renamed because they are persisted and serialized fields. Their circular west/east semantics are now documented and validated without a breaking payload migration.

### Shared clone abstraction

The local clone remains a short, type-specific ownership boundary. A common generic helper would not remove the need to identify each mutable nested field and would increase coupling between feature groups.

## Versioning

- Geographical Builder: `geographical-feature-builder-v3`
- Validator: `flight-feature-validator-v4`
- Processing Pipeline: `flight-feature-processing-pipeline-v9`
- Schema: `flight-features-v1`

No PostgreSQL migration is required. Processing version participates in immutable snapshot identity and isolates generation-nine outputs from generation-eight records.

## Closure markers

```text
GEOGRAPHICAL_POINT_WINDOW_FILTERING=ENFORCED
GEOGRAPHICAL_POINT_CHRONOLOGICAL_ORDER=ENFORCED
GEOGRAPHICAL_SEGMENT_GAP_BRIDGING=CLOSED
GEOGRAPHICAL_SEGMENT_SUPPORT_COUNT=METADATA_BASED
GEOGRAPHICAL_CIRCULAR_ENVELOPE_VALIDATION=ENFORCED
GEOGRAPHICAL_PATH_CROSSING_INDEPENDENCE=ENFORCED
GEOGRAPHICAL_KAHAN_DISTANCE_SUMMATION=ENFORCED
GEOGRAPHICAL_DISTANCE_MODEL=MEAN_EARTH_SPHERE_HAVERSINE_V1
GEOGRAPHICAL_CELL_POLICY=DECIMAL_DEGREE_ROUND_HALF_AWAY_FROM_ZERO_V1
GEOGRAPHICAL_BUILDER_NIL_CONTEXT=REJECTED
GEOGRAPHICAL_BUILDER_LOOP_CANCELLATION=ENFORCED
GEOGRAPHICAL_BUILDER_GENERATION=v3
GEOGRAPHICAL_VALIDATOR_GENERATION=v4
GEOGRAPHICAL_BUILDER_PROCESSING_GENERATION=v9
DATABASE_MIGRATION=NOT_REQUIRED
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
GEOGRAPHICAL_BUILDER_REVIEW_STATUS=CLOSED
```

## Canonical remediation history

Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, and Continuous Integration evidence. Severity labels are retrospective.

### GFA-DATA-180 — Geographical point evidence lacked one deterministic temporal eligibility and ordering contract
1. **Finding / symptom:** point coordinates could participate without a complete timestamp/window gate and without deterministic chronological ordering.
2. **Root cause:** input slice order and partial coordinate validity were implicitly trusted by downstream geometry calculations.
3. **Failure scenario:** equivalent point sets in different orders, or points outside the trajectory window, produce different endpoints, path distance, cells, or antimeridian evidence.
4. **Impact:** non-reproducible and temporally invalid geographical features.
5. **Severity rationale:** **P1 retrospective** because durable analytical geometry could change from ordering rather than evidence.
6. **Existing guarantees violated:** deterministic replay and authoritative trajectory-window evidence.
7. **Considered solutions:** trust upstream ordering, sort only for path distance, or canonicalize all eligible point evidence once.
8. **Chosen remediation:** require timestamps inside the trajectory window and order by observation time, point identity, then original index.
9. **Why selected:** one canonical sequence protects every downstream geographical calculation.
10. **Rejected alternatives:** duplicating `AsOfTime` policy in the builder; Extractor remains owner of the historical cutoff.
11. **Trade-offs:** malformed/missing-timestamp points are excluded and reported instead of guessed.
12. **Regression tests / protection:** permutation, window-filter, missing-timestamp tests and `geographicalbuilderreviewaudit`.
13. **Adversarial review findings:** schema mismatch and builder-level `AsOfTime` findings were stale/rejected, not reopened.
14. **Remediation iterations:** closed in `1bbfd0147092baf2615f5bb0838ca12768b54846`.
15. **Residual risks and limitations:** source timestamp correctness remains an upstream evidence responsibility.
16. **Operational or deployment consequences:** no migration; processing advances to generation nine.
17. **Exact evidence:** `1bbfd0147092baf2615f5bb0838ca12768b54846`, targeted builder tests, permanent CI audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** all new point-derived geometry must consume the same canonical eligible sequence.

### GFA-DATA-181 — Segment fallback bridged unobserved discontinuities into path distance
1. **Finding / symptom:** endpoint coordinates from separate segments could be treated as one continuous observed path.
2. **Root cause:** envelope coordinates and observed path edges were represented by the same flattened coordinate list.
3. **Failure scenario:** disconnected segments create synthetic distance across a gap that was never observed.
4. **Impact:** inflated `ObservedPathDistanceKM`, distorted displacement relationships, and false movement claims.
5. **Severity rationale:** **P1 retrospective** because the metric could manufacture observed movement.
6. **Existing guarantees violated:** observed path must include only evidenced continuous movement.
7. **Considered solutions:** disable segment fallback, join every endpoint, or separate envelope geometry from observed edges.
8. **Chosen remediation:** maintain a coordinate envelope separately from explicit segment path edges; never add distance across discontinuities.
9. **Why selected:** preserves honest fallback geometry without fabricating path continuity.
10. **Rejected alternatives:** persisting/synthesizing additional points solely to bridge gaps.
11. **Trade-offs:** fallback path may contain multiple independent parts rather than one continuous line.
12. **Regression tests / protection:** disconnected-segment and fallback-distance tests plus strict audit.
13. **Adversarial review findings:** field renaming was rejected because semantics can be fixed without a breaking payload migration.
14. **Remediation iterations:** `1bbfd0147092baf2615f5bb0838ca12768b54846`.
15. **Residual risks and limitations:** segment endpoints describe available summary evidence, not raw flight tracks.
16. **Operational or deployment consequences:** no schema/database migration; corrected output isolated by processing v9.
17. **Exact evidence:** implementation commit and geographical fallback/path regression tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** path metrics must be constructed from explicit observed edges/parts, never from a generic coordinate envelope.

### GFA-DATA-182 — Segment fallback supporting-point count used endpoint-coordinate cardinality instead of authoritative evidence
1. **Finding / symptom:** fallback support could be inferred from the number of generated endpoint coordinates.
2. **Root cause:** a geometry representation was reused as an evidence-count representation.
3. **Failure scenario:** repeated/shared endpoints or segment topology change supporting-count semantics without a real change in observed evidence.
4. **Impact:** misleading completeness/support evidence.
5. **Severity rationale:** **P2 retrospective** because it corrupts evidence accounting rather than coordinate values themselves.
6. **Existing guarantees violated:** supporting counts must come from authoritative trajectory/segment metadata.
7. **Considered solutions:** coordinate count, deduplicated coordinate count, or authoritative metadata.
8. **Chosen remediation:** use trajectory/segment evidence metadata and report unavailable support explicitly when it cannot be established.
9. **Why selected:** evidence ownership remains separate from geometry implementation details.
10. **Rejected alternatives:** manufacturing support from endpoint list shape.
11. **Trade-offs:** some fallbacks remain available with an explicit support limitation instead of a guessed count.
12. **Regression tests / protection:** metadata-based support-count tests and audit.
13. **Adversarial review findings:** generic clone abstraction was rejected as unrelated to this invariant.
14. **Remediation iterations:** `1bbfd0147092baf2615f5bb0838ca12768b54846`.
15. **Residual risks and limitations:** upstream metadata can still be inconsistent; mismatches are surfaced separately.
16. **Operational or deployment consequences:** none beyond corrected evidence output.
17. **Exact evidence:** implementation commit and fallback support-count tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** quantitative evidence support must not be derived from rendering/geometry container length.

### GFA-DATA-183 — Circular longitude envelope semantics were conflated with chronological antimeridian crossing
1. **Finding / symptom:** wrapped west/east bounds could be interpreted as proof that the observed chronological path crossed the antimeridian.
2. **Root cause:** circular envelope geometry and path-edge topology were not treated as independent contracts.
3. **Failure scenario:** a valid wrapped envelope triggers a false path-crossing conclusion, or validator rejects a correct wrapped envelope.
4. **Impact:** incorrect geographical flags and validation failures near the antimeridian.
5. **Severity rationale:** **P1 retrospective** because valid real-world geography can be semantically misclassified.
6. **Existing guarantees violated:** envelope span and path crossing must describe different evidence properties.
7. **Considered solutions:** prohibit west > east, rename persisted fields, or formalize circular bounds and validate crossing from path edges.
8. **Chosen remediation:** west > east is valid circular wrapping; `CircularLongitudeSpanDegrees` validates span while crossing is derived independently from chronological edges.
9. **Why selected:** fixes semantics without a breaking persisted-field rename.
10. **Rejected alternatives:** field rename/payload migration and treating every wrapped envelope as a crossing.
11. **Trade-offs:** consumers must understand the documented circular west/east meaning of legacy field names.
12. **Regression tests / protection:** wrapped-envelope, crossing-independence, validator reconciliation tests.
13. **Adversarial review findings:** field renaming explicitly rejected for compatibility reasons.
14. **Remediation iterations:** `1bbfd0147092baf2615f5bb0838ca12768b54846`.
15. **Residual risks and limitations:** external consumers ignoring documented circular semantics can still misread the wire fields.
16. **Operational or deployment consequences:** no migration; validator moves to v4 with processing v9.
17. **Exact evidence:** commit, `CircularLongitudeSpanDegrees` tests, geographical/validator audits.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** envelope and path-topology properties require separate named policies and tests.

### GFA-CONTRACT-184 — Geographical numeric policies were implicit and accumulation was less stable than necessary
1. **Finding / symptom:** distance model/cell quantization semantics were not versioned contracts and path accumulation used ordinary summation.
2. **Root cause:** numerical implementation details were not promoted to reproducibility metadata.
3. **Failure scenario:** a future refactor changes Earth radius, rounding, or summation and silently changes immutable feature output.
4. **Impact:** analytical drift and weaker reproducibility across generations.
5. **Severity rationale:** **P2 retrospective** because deterministic analytical semantics were under-specified, though no evidence showed catastrophic current error.
6. **Existing guarantees violated:** versioned formulas for durable analytical features.
7. **Considered solutions:** leave implementation implicit, round output, use alternative geodesy, or version current Haversine/cell policies and use compensated summation.
8. **Chosen remediation:** version mean-Earth Haversine and decimal-degree half-away-from-zero cell policies; use Kahan path accumulation.
9. **Why selected:** preserves intended current model while making future changes explicit generation decisions.
10. **Rejected alternatives:** pretending decimal-degree buckets are equal-area physical cells or silently changing the distance model.
11. **Trade-offs:** model remains an approximation by design and outputs remain binary64 rather than decimal-rounded.
12. **Regression tests / protection:** policy constants, numeric tests, audit.
13. **Adversarial review findings:** precision zero remains a documented config sentinel; effective production precision remains positive.
14. **Remediation iterations:** `1bbfd0147092baf2615f5bb0838ca12768b54846`.
15. **Residual risks and limitations:** spherical Haversine is not ellipsoidal geodesy; limitation is deliberate and versioned.
16. **Operational or deployment consequences:** no migration; processing v9 isolates corrected semantics.
17. **Exact evidence:** implementation commit, policy constants/tests, Backend Quality audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** numerical model/rounding changes require a named policy version and processing-generation review.

### GFA-OPS-185 — Geographical Builder accepted nil context and long geometry passes lacked a complete cancellation/diagnostic contract
1. **Finding / symptom:** nil context was replaced implicitly and large collection/geometry passes could continue without bounded cancellation checks; rejection diagnostics lacked exact accounting.
2. **Root cause:** lifecycle and evidence-diagnostic concerns were treated as incidental to pure geometry.
3. **Failure scenario:** cancelled materialization burns CPU over a large evidence set, or discarded evidence cannot be audited quantitatively.
4. **Impact:** operational waste and weaker explainability.
5. **Severity rationale:** **P2 retrospective** for lifecycle ownership; diagnostic precision is lower-risk but closed in the same scan-control remediation.
6. **Existing guarantees violated:** explicit context ownership, bounded cancellation responsiveness, explainable evidence rejection.
7. **Considered solutions:** background fallback, only entry cancellation checks, or strict context plus periodic checks and exact-count limitations.
8. **Chosen remediation:** reject nil context, poll cancellation through collection/geometry passes, and report exact rejected categories/counts.
9. **Why selected:** consistent with the hardened builder/repository context contract and low overhead.
10. **Rejected alternatives:** implicit background lifetime and a new global limitation registry.
11. **Trade-offs:** small polling/accounting overhead.
12. **Regression tests / protection:** nil-context, cancellation, limitation-count tests and strict audit.
13. **Adversarial review findings:** shared clone and closed limitation-registry recommendations remained rejected non-defects.
14. **Remediation iterations:** `1bbfd0147092baf2615f5bb0838ca12768b54846`.
15. **Residual risks and limitations:** cancellation latency is bounded by polling cadence.
16. **Operational or deployment consequences:** more responsive cancelled requests; no infrastructure changes.
17. **Exact evidence:** implementation commit, context/cancellation tests, `geographicalbuilderreviewaudit`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** all potentially large builder passes must state context and rejection-accounting semantics explicitly.
