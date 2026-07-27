# Historical Window Review Hardening

Status: implemented review remediation
Module: `apps/api/internal/historicalintelligence/historicalwindow`
Baseline: `fc254881fa446c7e80f94a959e2a9d5609874821`

## Confirmed defects

The review correctly identified that `time.Time.Sub` was unsafe for bucket-count and previous-window construction across windows longer than the `time.Duration` range. The planner now enforces the bucket limit during generation, advances only through validated calendar boundaries, and derives calendar previous windows from bucket steps rather than a saturated duration.

The review also correctly identified incomplete context cancellation, mutable derived plan evidence, execution-policy contamination of the semantic fingerprint, missing bucket and previous-window fingerprint evidence, unaligned public `NextBoundary` behavior, and negative duration helpers. These boundaries are now hardened.

## Deliberately rejected findings

`GranularityCustom` remains part of the internal Historical Intelligence contract. The production command intentionally exposes only hour, day, and week, but the reusable domain planner and schema support one custom bucket. Command exposure and domain capability are separate boundaries.

Nullable `EffectiveWindow` and `PreviousWindow` pointers remain intentional because a valid request may produce no complete bucket. Replacing semantic absence with zero-value windows would weaken the contract.

The nil-safe `BucketCountExceededError.Error` and `Unwrap` methods remain. Returning `nil` from `Unwrap` for a nil receiver follows the Go error-chain contract and is not equivalent to an unexplained domain null.

The long-function and switch-count observations were treated as maintainability signals rather than independent architecture failures. Refactoring was applied where it removed semantic duplication or enabled correctness, not to satisfy mechanical line-count rules.

## Resulting contract

- `Build` rejects a nil context and checks cancellation during every bucket iteration and before return.
- Bucket limits are enforced before every append.
- `NextBoundary` requires an aligned boundary.
- One granularity policy owns floor and shift behavior.
- Calendar previous windows use exact bucket shifts.
- Custom previous windows use checked second and nanosecond arithmetic instead of `time.Duration`.
- `CanonicalizePlan` reconstructs all derived evidence from request semantics.
- `ValidatePlan` verifies fingerprint, windows, buckets, keys, sequences, exclusions, UTC normalization, truncation, and execution-limit consistency.
- Traffic, airport, route, and generic series builders canonicalize plans before analytics and fingerprint generation.
- Fingerprint generation two excludes `MaximumBucketCount` while binding effective and previous windows, buckets, keys, sequences, exclusions, and truncation.
- Reversed bucket and exclusion intervals have zero duration and are rejected by canonical validation.

No PostgreSQL schema migration is required. Existing rows remain readable, but an exact replay against a pre-generation-two row may fail closed with `ErrResultConflict` because corrected semantic provenance must not silently overwrite immutable historical evidence. Any rebuild of pre-generation-two identities must therefore be an explicit governed reset rather than an automatic compatibility rewrite.
