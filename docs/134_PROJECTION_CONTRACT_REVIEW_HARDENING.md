# Projection Contract Review Hardening

Status: implemented review remediation

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
