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
