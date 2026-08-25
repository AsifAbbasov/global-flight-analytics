# Document 122 — Trajectory Builder Review Hardening

Status: Review closure candidate
Baseline commit: `0b5ec52b503f1ef65c2ca5eeaba485e381710649`
Scope: `apps/api/internal/features/trajectorybuilder` and the minimum quality and validation contracts required for consistent output

## 1. Review classification

The original static review was based on stale commit `a1689dc`. Its claim that production feature materialization never loads point records is no longer current: Operational Builder hardening introduced `FeatureTrajectoryReader`, which hydrates flight-state points only for feature materialization inside the existing repeatable-read snapshot.

The following findings remained confirmed on the current baseline:

- trajectory count semantics preferred loaded slice length without a persisted-metadata fallback;
- sampling and path used different point order and eligibility rules;
- duplicate timestamps contributed zero sampling intervals;
- point evidence was not restricted to the trajectory window;
- point and segment paths bridged declared discontinuities;
- one point blocked better segment fallback;
- empty evidence claimed eleven base fields plus a zero quality score;
- coverage could report one without point or segment observation evidence;
- zero duration was handled before excluding outside-window gaps;
- zero gap-duration metadata escaped mismatch reporting;
- fractional duration policy was undocumented;
- path ratios were unconditionally clamped;
- context cancellation was absent from several large scans;
- unknown segment statuses produced one limitation per record.

## 2. Canonical evidence contract

Trajectory Builder generation two creates one canonical evidence view before calculating any field:

1. validate the trajectory start and end window;
2. reject points with missing timestamps;
3. reject points outside the authoritative trajectory window;
4. sort by observation time and deterministic point identity fields;
5. collapse duplicate timestamps to one deterministic observation;
6. share that sequence between point count, sampling and point-path calculations;
7. sort segments and coverage gaps without mutating caller-owned slices.

`AsOfTime` remains owned by Extractor snapshot validation. Builders continue to enforce their own trajectory-window membership and do not duplicate the global historical cutoff contract.

## 3. Point-count and quality reconciliation

When point records are materialized, `trajectory.point_count` is the number of unique temporally eligible canonical observations. Metadata disagreement is disclosed.

When point records are intentionally unavailable but persisted `PointCount` is positive, the persisted value is retained as a fallback count and supporting-point count. Sampling and point-path fields remain unavailable rather than being fabricated.

Extractor initial quality now derives supporting-point count exclusively from built group evidence. It no longer independently reintroduces raw trajectory metadata after group builders have canonicalized evidence. Validator and Extractor therefore reconcile against the same values.

## 4. Sampling policy

The policy is:

```text
unique-chronological-observation-instants-kahan-v1
```

Duplicate timestamps are collapsed before interval calculation. Sampling mean uses compensated summation and maximum gap uses the same unique chronological sequence.

## 5. Coverage policy

The policy is:

```text
non-invalid-segment-evidence-plus-clipped-gap-union-v1
```

Positive-duration coverage requires materialized non-invalid segment evidence. A zero-duration instant requires one canonical point. Persisted point-count metadata or absence of gaps alone is insufficient. Valid gaps are clipped to the trajectory window and merged without double counting. Zero-duration windows are evaluated only after outside-window gaps are excluded.

Gap metadata uses the shared `truncate_fractional_seconds_toward_zero` policy. Subsecond gaps still contribute their full nanosecond-resolution duration to the coverage ratio; only the integer metadata comparison is truncated.

## 6. Path policy

The policy is:

```text
continuous-parts-no-gap-or-segment-discontinuity-bridging-v1
```

Coverage gaps split point paths. Invalid point coordinates also split continuity. Segment fallback joins only exactly contiguous segment endpoints into one path part. Coordinate mismatches, invalid segments, missing temporal evidence and declared gaps split parts, so distance between disconnected segments is never called observed movement.

Path efficiency sums direct distance and observed distance within each continuous part. Ratio values are clamped only inside an explicit numerical tolerance. Larger violations become unavailable evidence with a limitation.

Both Trajectory and Geographical builders use the documented mean-Earth spherical Haversine model. They retain separate implementations because their evidence semantics differ: Geographical Builder owns envelopes and geographical movement, while Trajectory Builder owns segmented path efficiency.

## 7. Completeness semantics

Field availability is counted explicitly from zero. Empty input is unavailable. Zero segment counts may support zero status counts when an authoritative trajectory envelope exists, but segment shares are undefined for a zero denominator and are not counted as available fields. A zero quality score is accepted only when trajectory evidence exists.

## 8. Validation alignment

Validator generation five skips the geographical-to-trajectory path ratio comparison when Trajectory Builder explicitly reports:

- coverage-gap path segmentation;
- segment fallback;
- insufficient path evidence;
- zero or non-finite path distance;
- an out-of-range ratio.

This avoids treating two intentionally different evidence models as contradictory.

## 9. Version boundary

```text
trajectory-feature-builder-v2
flight-feature-validator-v5
flight-feature-processing-pipeline-v11
flight-features-v1
```

No PostgreSQL migration is required.

## 10. Closure markers

```text
TRAJECTORY_CANONICAL_EVIDENCE=ENFORCED
TRAJECTORY_POINT_WINDOW_FILTERING=ENFORCED
TRAJECTORY_POINT_CHRONOLOGICAL_ORDER=ENFORCED
TRAJECTORY_DUPLICATE_TIMESTAMP_POLICY=UNIQUE_DETERMINISTIC
TRAJECTORY_POINT_COUNT_FALLBACK=PERSISTED_METADATA_WHEN_UNMATERIALIZED
TRAJECTORY_QUALITY_SUPPORT_RECONCILIATION=GROUP_EVIDENCE_OWNED
TRAJECTORY_SAMPLING_POLICY=UNIQUE_CHRONOLOGICAL_KAHAN_V1
TRAJECTORY_COVERAGE_REQUIRES_OBSERVATION_EVIDENCE=ENFORCED
TRAJECTORY_COVERAGE_GAP_UNION=CLIPPED_AND_DEDUPLICATED
TRAJECTORY_GAP_DURATION_POLICY=TRUNCATE_FRACTIONAL_SECONDS_TOWARD_ZERO
TRAJECTORY_PATH_GAP_BRIDGING=CLOSED
TRAJECTORY_PATH_SEGMENT_FALLBACK=INDEPENDENT_PARTS
TRAJECTORY_PATH_RATIO_TOLERANCE=ENFORCED
TRAJECTORY_EMPTY_COMPLETENESS=NOT_INFLATED
TRAJECTORY_NIL_CONTEXT=REJECTED
TRAJECTORY_LOOP_CANCELLATION=ENFORCED
TRAJECTORY_BUILDER_GENERATION=v2
TRAJECTORY_VALIDATOR_GENERATION=v5
TRAJECTORY_BUILDER_PROCESSING_GENERATION=v11
DATABASE_MIGRATION=NOT_REQUIRED
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
TRAJECTORY_BUILDER_REVIEW_STATUS=CLOSED
```

## 11. Canonical remediation history

Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, and Continuous Integration evidence. Historical exact-commit CI evidence for the original implementation is not reconstructed here; current permanent audits provide repository-state regression evidence. Severity labels are retrospective.

### GFA-DATA-196 — Trajectory point-count and quality support had competing ownership rules
1. **Finding / symptom:** loaded slice length dominated point count while persisted metadata had no honest fallback, and Extractor could independently reintroduce raw point metadata into quality support.
2. **Root cause:** count/support semantics were split between Builder and Extractor instead of owned by canonical group evidence.
3. **Failure scenario:** unmaterialized production points report zero count/support while persisted count is positive, or quality support disagrees with the builder's filtered evidence.
4. **Impact:** inconsistent evidence accounting and completeness.
5. **Severity rationale:** **P1 retrospective** because durable quality/support evidence could contradict the feature values it describes.
6. **Existing guarantees violated:** one canonical evidence owner and test/production parity.
7. **Considered solutions:** always trust slice length, always trust metadata, or use canonical materialized count with explicit persisted fallback.
8. **Chosen remediation:** unique eligible points own count when materialized; persisted `PointCount` is fallback only when records are absent; Extractor quality uses group support only.
9. **Why selected:** it distinguishes real detail from bounded-storage metadata without fabricating sampling/path evidence.
10. **Rejected alternatives:** manufacturing points from metadata and dual Builder/Extractor support calculations.
11. **Trade-offs:** fallback can supply count/support but not sampling/path fields.
12. **Regression tests / protection:** materialized/unmaterialized count, metadata mismatch, quality reconciliation tests and audit.
13. **Adversarial review findings:** stale claim that production never hydrates points was not re-opened after Document 121.
14. **Remediation iterations:** `2872eb31e87500bdae1ae58fe2b75fb76c4b11d2`.
15. **Residual risks and limitations:** persisted count can be stale; disagreement is disclosed rather than silently resolved.
16. **Operational or deployment consequences:** no migration; processing advances to v11.
17. **Exact evidence:** implementation commit, point-count/quality tests, `trajectorybuilderreviewaudit`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** support/count evidence must have one canonical owner and explicit fallback rules.

### GFA-DATA-197 — Sampling and path calculations did not share one canonical eligible point sequence
1. **Finding / symptom:** sampling and path could use different ordering/eligibility, and point evidence was not uniformly restricted to the trajectory window.
2. **Root cause:** each metric independently interpreted raw point slices.
3. **Failure scenario:** the same snapshot uses one point order for sampling and another for path, or includes outside-window evidence in one metric only.
4. **Impact:** internally contradictory and permutation-sensitive Trajectory features.
5. **Severity rationale:** **P1 retrospective** because durable metrics could disagree about which observations constitute the trajectory.
6. **Existing guarantees violated:** one authoritative temporal evidence view and deterministic replay.
7. **Considered solutions:** duplicate per-metric sorting/filtering or canonicalize once.
8. **Chosen remediation:** one canonical evidence view validates window, filters timestamps, sorts deterministically, and is shared by count/sampling/path.
9. **Why selected:** prevents future drift among metrics.
10. **Rejected alternatives:** duplicating global `AsOfTime` logic; Extractor retains that ownership.
11. **Trade-offs:** raw input order is intentionally discarded.
12. **Regression tests / protection:** permutation/window/canonical evidence tests.
13. **Adversarial review findings:** point slices are sorted through copied evidence, preserving caller ownership.
14. **Remediation iterations:** `2872eb31...`.
15. **Residual risks and limitations:** canonicalization cannot repair incorrect source timestamps.
16. **Operational or deployment consequences:** corrected semantics under v11.
17. **Exact evidence:** implementation commit and canonical evidence tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** all point-derived Trajectory metrics must consume the canonical evidence object.

### GFA-DATA-198 — Duplicate observation timestamps created artificial zero sampling intervals
1. **Finding / symptom:** multiple observations at the same timestamp contributed zero-length intervals.
2. **Root cause:** sampling operated on records rather than unique observation instants.
3. **Failure scenario:** duplicate timestamps lower mean sampling interval and alter maximum/mean relationship without new temporal evidence.
4. **Impact:** biased sampling metrics.
5. **Severity rationale:** **P2 retrospective** because it distorts a derived metric but does not create future/out-of-window evidence.
6. **Existing guarantees violated:** sampling describes elapsed time between distinct observation instants.
7. **Considered solutions:** retain zero intervals, arbitrarily drop records, or collapse duplicate instants deterministically.
8. **Chosen remediation:** one deterministic representative per timestamp; sampling uses unique chronological instants with compensated mean.
9. **Why selected:** preserves temporal meaning and reproducibility.
10. **Rejected alternatives:** allowing zero intervals to encode duplicate record density.
11. **Trade-offs:** within-timestamp multiplicity is intentionally not a sampling-frequency signal.
12. **Regression tests / protection:** duplicate-timestamp and permutation tests.
13. **Adversarial review findings:** point identity tie-breakers make representative choice deterministic.
14. **Remediation iterations:** `2872eb31...`.
15. **Residual risks and limitations:** timestamp resolution of the source bounds temporal precision.
16. **Operational or deployment consequences:** none beyond v11 output isolation.
17. **Exact evidence:** implementation commit, sampling policy tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** interval metrics must define duplicate-time semantics explicitly.

### GFA-DATA-199 — Trajectory path logic bridged declared discontinuities and could suppress better segment fallback
1. **Finding / symptom:** point/segment paths could cross declared gaps; a lone point could prevent segment fallback even when segments contained stronger path evidence.
2. **Root cause:** path construction treated evidence availability as one flattened sequence and used point-count presence as a fallback gate.
3. **Failure scenario:** distance is added over a coverage gap or useful segment parts are ignored because one insufficient point exists.
4. **Impact:** fabricated or unnecessarily unavailable path efficiency.
5. **Severity rationale:** **P1 retrospective** because observed movement could be invented across unobserved gaps.
6. **Existing guarantees violated:** path distance only within continuous evidenced parts.
7. **Considered solutions:** always prefer points, always prefer segments, bridge gaps, or build independent continuous path parts with evidence-quality fallback.
8. **Chosen remediation:** coverage gaps/invalid coordinates/discontinuous segments split path parts; segment fallback remains available when point evidence is insufficient.
9. **Why selected:** preserves maximum honest evidence without manufacturing continuity.
10. **Rejected alternatives:** interpolation and global flattening.
11. **Trade-offs:** path can consist of multiple independent parts.
12. **Regression tests / protection:** gap-bridging, one-point fallback, segment continuity tests.
13. **Adversarial review findings:** separate Haversine implementations remain because Geographical and Trajectory evidence semantics differ.
14. **Remediation iterations:** `2872eb31...`.
15. **Residual risks and limitations:** summary segments cannot reproduce raw track detail.
16. **Operational or deployment consequences:** processing v11 isolates corrected values.
17. **Exact evidence:** implementation commit and path fallback/discontinuity tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every path metric must model discontinuities as first-class boundaries.

### GFA-DATA-200 — Coverage could claim complete observation without actual point or segment evidence
1. **Finding / symptom:** absence of declared gaps or positive persisted point metadata could allow coverage ratio `1` without materialized observation evidence.
2. **Root cause:** “no known gap” was treated as equivalent to “observed coverage.”
3. **Failure scenario:** a positive-duration trajectory with no usable segment/point evidence is reported fully covered.
4. **Impact:** fabricated data coverage and overstated trust.
5. **Severity rationale:** **P1 retrospective** because missing evidence could be published as complete evidence.
6. **Existing guarantees violated:** coverage must be observation-supported.
7. **Considered solutions:** infer full coverage from metadata, infer from no gaps, or require materialized observation support.
8. **Chosen remediation:** positive-duration coverage requires non-invalid segment evidence; zero-duration instant requires a canonical point.
9. **Why selected:** fail-closed coverage semantics separate absence of known gaps from proof of observation.
10. **Rejected alternatives:** using persisted count alone as coverage proof.
11. **Trade-offs:** more snapshots can report unavailable/limited coverage when detail is not materialized.
12. **Regression tests / protection:** no-evidence coverage and zero-duration coverage tests.
13. **Adversarial review findings:** bounded-storage fallback remains valid for count, not for fabricated coverage.
14. **Remediation iterations:** `2872eb31...`.
15. **Residual risks and limitations:** segment-based coverage is summary evidence, not raw point continuity.
16. **Operational or deployment consequences:** corrected v11 coverage semantics.
17. **Exact evidence:** implementation commit and coverage evidence tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** “unknown gap” and “observed coverage” must never be interchangeable states.

### GFA-DATA-201 — Coverage-gap clipping, overlap, zero-duration and duration-mirror semantics were inconsistent
1. **Finding / symptom:** outside-window gaps could affect zero-duration handling, overlapping gaps could double-count, zero duration metadata could hide mismatch, and fractional-second mirror policy was undocumented.
2. **Root cause:** gap normalization and metadata validation were performed in fragmented order.
3. **Failure scenario:** coverage ratio or mismatch diagnostics change because gaps are un-clipped/overlapped or integer metadata is interpreted differently from nanosecond duration.
4. **Impact:** incorrect coverage and inconsistent gap provenance.
5. **Severity rationale:** **P1 retrospective** because coverage is a trust metric and could be numerically wrong.
6. **Existing guarantees violated:** gaps must be evaluated only inside the authoritative window and unioned without double counting.
7. **Considered solutions:** trust persisted duration, sum raw gaps, round seconds, or normalize/clip/merge then compare using the shared duration policy.
8. **Chosen remediation:** clip to window, merge overlaps, exclude outside-window gaps before zero-window decisions, compare metadata with shared truncate-toward-zero policy while ratio uses full duration precision.
9. **Why selected:** separates high-resolution analytical duration from integer mirror validation honestly.
10. **Rejected alternatives:** raw summation and treating zero metadata as missing.
11. **Trade-offs:** metadata comparison intentionally has integer-second precision.
12. **Regression tests / protection:** overlap/clipping/subsecond/zero-duration mismatch tests.
13. **Adversarial review findings:** fractional policy reuses the shared Temporal duration contract rather than creating another rounding rule.
14. **Remediation iterations:** `2872eb31...`.
15. **Residual risks and limitations:** integer persisted mirror cannot encode subsecond duration exactly by design.
16. **Operational or deployment consequences:** no migration; output change isolated by v11.
17. **Exact evidence:** implementation commit and gap-union/duration tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** coverage-gap pipelines must normalize window/overlap before aggregation or metadata comparison.

### GFA-DATA-202 — Empty Trajectory evidence inflated field availability and allowed unsupported zero quality
1. **Finding / symptom:** empty evidence started with eleven fields considered available and could publish a zero quality score without any supporting trajectory evidence.
2. **Root cause:** availability count was initialized optimistically and zero was accepted as a value regardless of evidence presence.
3. **Failure scenario:** an empty trajectory looks partially/fully populated with a legitimate quality measurement.
4. **Impact:** fabricated completeness and quality evidence.
5. **Severity rationale:** **P1 retrospective** because absence could masquerade as available analytical data.
6. **Existing guarantees violated:** field availability must be earned from evidence; quality requires support.
7. **Considered solutions:** preserve zero defaults, special-case empty aggregate, or count availability explicitly from zero.
8. **Chosen remediation:** availability starts at zero; fields increment only when supported; quality zero is valid only when trajectory evidence exists.
9. **Why selected:** eliminates zero-value ambiguity at the group level.
10. **Rejected alternatives:** interpreting Go zero values as implicit available data.
11. **Trade-offs:** fixtures relying on optimistic defaults must declare evidence explicitly.
12. **Regression tests / protection:** empty-evidence/completeness/zero-quality tests.
13. **Adversarial review findings:** Validator v5 aligns cross-builder relationship checks with explicit Trajectory limitations.
14. **Remediation iterations:** `2872eb31...`.
15. **Residual risks and limitations:** availability remains group/field-contract based rather than per-value wrapper based by design.
16. **Operational or deployment consequences:** processing v11 isolates corrected completeness.
17. **Exact evidence:** implementation commit and empty-input tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** feature groups must initialize availability from no evidence, never from struct zero values.

### GFA-DATA-203 — Path-efficiency ratio was unconditionally clamped instead of distinguishing numerical noise from semantic contradiction
1. **Finding / symptom:** any out-of-range ratio could be forced into `[0,1]`.
2. **Root cause:** numerical stabilization and data-integrity validation were conflated.
3. **Failure scenario:** a materially impossible direct/path relationship is hidden as exactly zero or one.
4. **Impact:** contradictory geometry is published as plausible evidence.
5. **Severity rationale:** **P1 retrospective** because mathematical contradiction could be silently masked.
6. **Existing guarantees violated:** fail-closed integrity for impossible analytical relationships.
7. **Considered solutions:** always clamp, never clamp, or clamp only inside explicit floating-point tolerance.
8. **Chosen remediation:** tolerance-bounded clamping only; larger violations become unavailable evidence with a limitation.
9. **Why selected:** handles binary64 noise without hiding real contradictions.
10. **Rejected alternatives:** unconditional normalization.
11. **Trade-offs:** some pathological inputs now lose the metric rather than receiving a sanitized value.
12. **Regression tests / protection:** near-boundary/tolerance/out-of-range tests and Validator alignment.
13. **Adversarial review findings:** Geographical/Trajectory relationship validation skips comparison where explicit path limitations make models intentionally incomparable.
14. **Remediation iterations:** `2872eb31...`.
15. **Residual risks and limitations:** tolerance policy must remain dimensionally appropriate and is guarded later by Validator review.
16. **Operational or deployment consequences:** corrected v11 output semantics.
17. **Exact evidence:** implementation commit and ratio tolerance tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** clamping is allowed only for documented numerical tolerance, never as a general integrity repair.

### GFA-OPS-204 — Trajectory Builder accepted nil context and large scans lacked bounded cancellation checks
1. **Finding / symptom:** several evidence passes could continue after caller cancellation and nil context was not rejected consistently.
2. **Root cause:** lifecycle handling was incomplete across pure-computation helper passes.
3. **Failure scenario:** cancelled materialization scans large points/segments/gaps unnecessarily.
4. **Impact:** CPU waste and inconsistent request lifecycle behavior.
5. **Severity rationale:** **P2 retrospective** as an operational correctness defect.
6. **Existing guarantees violated:** explicit context ownership and bounded work.
7. **Considered solutions:** ignore context, entry-only check, or periodic checks across all potentially large loops.
8. **Chosen remediation:** reject nil and propagate cancellation through canonicalization/sampling/coverage/path passes.
9. **Why selected:** consistent with other hardened feature builders with bounded overhead.
10. **Rejected alternatives:** implicit `context.Background()` and post-hoc cancellation only.
11. **Trade-offs:** small polling overhead.
12. **Regression tests / protection:** nil/mid-loop cancellation tests and audit.
13. **Adversarial review findings:** caller-owned slices remain unmodified by sorting copied evidence.
14. **Remediation iterations:** `2872eb31...`.
15. **Residual risks and limitations:** cancellation latency follows polling cadence.
16. **Operational or deployment consequences:** more responsive cancellation; no migration.
17. **Exact evidence:** implementation commit, context tests, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** new large evidence passes must carry context through their full call chain.

### GFA-MAINT-205 — Unknown segment statuses produced unbounded duplicate limitation records
1. **Finding / symptom:** one identical unknown-status limitation could be emitted per segment record.
2. **Root cause:** diagnostics were appended at record granularity rather than aggregated by reason/count.
3. **Failure scenario:** a malformed batch with many unknown statuses bloats feature limitations and durable JSON without adding information.
4. **Impact:** noisy diagnostics and avoidable payload/memory growth.
5. **Severity rationale:** **P3 retrospective** because analytical values remain fail-closed; the defect is diagnostic scalability.
6. **Existing guarantees violated:** bounded, useful explainability evidence.
7. **Considered solutions:** keep per-record messages, silently ignore, or aggregate identical reasons with counts.
8. **Chosen remediation:** aggregate unknown-status evidence into bounded typed limitations with counts.
9. **Why selected:** preserves explainability without linear duplicate messages.
10. **Rejected alternatives:** hiding malformed status evidence entirely.
11. **Trade-offs:** limitation no longer identifies every individual segment ID.
12. **Regression tests / protection:** many-unknown-status diagnostic tests and audit.
13. **Adversarial review findings:** exact per-record diagnostic expansion was not justified for a group-level feature payload.
14. **Remediation iterations:** `2872eb31e87500bdae1ae58fe2b75fb76c4b11d2`.
15. **Residual risks and limitations:** detailed forensic IDs remain upstream/repository responsibility.
16. **Operational or deployment consequences:** smaller bounded validation payloads under malformed data.
17. **Exact evidence:** implementation commit, limitation aggregation tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** repeated group-level data-quality conditions should use counted/aggregated diagnostics unless per-record identity is contractually required.
