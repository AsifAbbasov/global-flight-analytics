# Historical Similarity Review Hardening

Status: implemented review remediation

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
