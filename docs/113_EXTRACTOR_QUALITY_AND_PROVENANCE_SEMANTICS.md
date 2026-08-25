# Document 113 — Extractor Quality and Provenance Semantics

Status: IMPLEMENTED
Baseline commit: e853f5931c78f6ed7b0fbcd0dd85a53cfbaa22f3

## 1. Required completeness and optional coverage

The schema already declared all six aircraft enrichment fields optional. The
quality model now follows that declaration instead of placing optional aircraft
fields in the denominator of core completeness.

`CompletenessScore` measures required feature fields only. The new
`OptionalCoverageScore` independently reports optional enrichment coverage.
Missing or partial optional aircraft metadata remains visible through evidence
status and limitations but no longer makes an otherwise complete required
feature set limited.

The current schema contract contains 46 required fields and 6 optional aircraft
fields. Aggregate evidence groups are prohibited from mixing required and
optional fields until the evidence model can report per-requirement counts.

## 2. Honest trajectory record provenance

`TrajectoryUpdatedAt` now contains only the real trajectory `UpdatedAt` value.
`TrajectoryCreatedAt` records the independent creation timestamp. When either
value is unknown it remains zero and is reported honestly; trajectory `EndTime`
is never substituted for a system record timestamp.

System record timestamps are not compared with historical `AsOfTime`, because
record materialization may legitimately happen after the aviation event window.
The validator instead checks creation/update ordering and reports entirely
missing record timestamps as a provenance limitation.

## 3. Aircraft metadata provenance

Production extractor composition now supplies an explicit aircraft metadata
source name and provider version. Every configured aircraft enrichment records:

- metadata source name;
- provider version;
- retrieval completion time.

The stable source name also participates in extractor composition processing
identity. Retrieval time is intentionally provenance, not deterministic input
identity, so replay fingerprints do not change merely because the same inputs
were processed at a different wall-clock time.

## 4. Validation semantics

Optional-only group absence and partial coverage no longer generate validation
warnings. Structural inconsistencies remain errors. Aircraft metadata that is
available or partial requires complete source, provider-version and retrieval
time provenance.

## 5. Processing generation

The extractor, extractor composition and feature processing pipeline advance to
version 5. The validator advances to version 2. Stored snapshots remain isolated
by explicit processing version.

```text
REQUIRED_COMPLETENESS_OPTIONAL_COVERAGE=SEPARATED
OPTIONAL_AIRCRAFT_ABSENCE_VALIDATION_PENALTY=REMOVED
TRAJECTORY_ENDTIME_PROVENANCE_FALLBACK=REMOVED
TRAJECTORY_CREATED_UPDATED_PROVENANCE=EXPLICIT
AIRCRAFT_METADATA_PROVENANCE=EXPLICIT
AIRCRAFT_METADATA_SOURCE_PROCESSING_IDENTITY=ENFORCED
EXTRACTOR_PROCESSING_GENERATION=v5
VALIDATOR_GENERATION=v2
```

---

## Canonical remediation history

### GFA-DATA-147 — optional aircraft enrichment incorrectly reduced required-feature completeness

1. **Finding / symptom.** Core `CompletenessScore` included six aircraft fields even though the versioned schema declared them optional.
2. **Root cause.** Quality aggregation treated every evidence field as one denominator rather than respecting schema requirement class.
3. **Failure scenario.** All required temporal/geographical/operational/trajectory features are complete but optional aircraft metadata is unavailable; the snapshot is scored incomplete/limited despite satisfying every required field.
4. **Impact.** Feature quality systematically understates valid snapshots and can misclassify analytical readiness because optional enrichment availability is confused with required completeness.
5. **Severity rationale.** **P1 retrospective.** Quality status is durable analytical trust evidence; incorrect denominator semantics can materially misclassify otherwise complete feature data.
6. **Existing guarantees violated.** Required completeness must derive from required schema fields only; optional enrichment coverage must remain observable without changing required-field truth.
7. **Considered solutions.** Keep one denominator; mark aircraft fields required; separate required completeness from optional coverage.
8. **Chosen remediation.** Compute `CompletenessScore` from required fields only, add `OptionalCoverageScore` for optional enrichment, and reject aggregate evidence groups that mix requirement classes until per-requirement counts are representable.
9. **Why this solution was selected.** It aligns quality semantics with the versioned schema rather than changing schema truth to match an implementation shortcut.
10. **Rejected alternatives.** Promoting optional aircraft fields to required would change product semantics; a single blended score obscures why evidence is missing.
11. **Trade-offs.** Consumers now have two quality dimensions to interpret; this is intentional because completeness and enrichment coverage answer different questions.
12. **Regression tests / protection.** Extractor/validator tests verify optional absence does not penalize required completeness and structural inconsistencies still fail; `featurequalityprovenanceaudit` protects requirement separation.
13. **Adversarial review findings.** Aggregate evidence groups cannot safely mix required and optional counts until the evidence model can expose requirement-level breakdown; silently splitting totals heuristically would fabricate semantics.
14. **Remediation iterations.** Strict count validity landed in Document 112; `3cdd9b15…` then corrected the denominator/validation meaning and advanced processing to v5 / validator v2.
15. **Residual risks and limitations.** The current schema requirement classification itself remains authoritative; future schema changes must update derived counts and quality tests.
16. **Operational or deployment consequences.** Previously limited snapshots may be classified differently under processing v5; old snapshots remain isolated by their stored processing version.
17. **Exact evidence.** Implementation commit `3cdd9b1532609c872343d00626ba44a9c9855609` (`fix: separate feature quality and provenance semantics`). Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-DATA-147=CLOSED`.
19. **Prevention / future guard.** Quality metrics must derive required/optional field counts from the versioned schema and may not silently blend requirement classes.

### GFA-DATA-148 — trajectory `EndTime` was fabricated as system record update provenance

1. **Finding / symptom.** When real trajectory `UpdatedAt`/`CreatedAt` evidence was missing, provenance could fall back to aviation-event `EndTime`.
2. **Root cause.** Event-time availability was used to fill a system-record provenance gap even though the two timestamps describe different facts.
3. **Failure scenario.** A trajectory record has unknown system update time; feature provenance publishes the flight's end time as `TrajectoryUpdatedAt`, making a nonexistent repository timestamp appear known.
4. **Impact.** Audit consumers cannot distinguish record lifecycle evidence from event chronology and may infer false freshness/history.
5. **Severity rationale.** **P1 retrospective.** This fabricates provenance in durable analytical evidence rather than merely omitting unavailable metadata.
6. **Existing guarantees violated.** Unknown provenance must remain unknown; aviation event time must not substitute for system materialization time.
7. **Considered solutions.** Keep fallback; use `CreatedAt` as update fallback; persist independent created/updated timestamps and leave either zero when unknown.
8. **Chosen remediation.** Add explicit `TrajectoryCreatedAt`, make `TrajectoryUpdatedAt` use only the real `UpdatedAt`, preserve zero for unavailable timestamps, and validate ordering/absence as provenance semantics rather than as event cutoffs.
9. **Why this solution was selected.** It reports only evidence the repository actually owns and keeps system provenance separate from historical aviation event-time policy.
10. **Rejected alternatives.** Cross-fallbacks continue fabricating timestamp meaning; comparing record timestamps to `AsOfTime` would incorrectly reject legitimate later persistence.
11. **Trade-offs.** Some snapshots explicitly expose missing record timestamps and associated limitations instead of always having a convenient timestamp.
12. **Regression tests / protection.** Tests cover created/updated independence, absence preservation, no EndTime fallback and validator provenance limitations; permanent quality/provenance audit enforces the contract.
13. **Adversarial review findings.** System record timestamps may legitimately occur after feature `AsOfTime`; only their internal lifecycle ordering is validated, not historical event cutoff.
14. **Remediation iterations.** Document 112 explicitly deferred this contract rather than conflating system and event time; `3cdd9b15…` closes it honestly.
15. **Residual risks and limitations.** Historical repository revisions still cannot be reconstructed where the source model never stored them.
16. **Operational or deployment consequences.** Consumers may see zero/unavailable provenance where older generation code supplied an event-time fallback; processing v5 isolates the changed semantics.
17. **Exact evidence.** Implementation commit `3cdd9b1532609c872343d00626ba44a9c9855609`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-DATA-148=CLOSED`.
19. **Prevention / future guard.** Provenance fields may only be populated from semantically matching source evidence; convenient cross-domain timestamp fallbacks are prohibited.

### GFA-DATA-149 — aircraft enrichment lacked durable source, provider-version and retrieval provenance

1. **Finding / symptom.** Aircraft metadata values could be present/partial in a feature snapshot without recording which metadata source/provider generation supplied them or when retrieval completed.
2. **Root cause.** Enrichment payload participated in features/fingerprints before its observational provenance was modeled explicitly.
3. **Failure scenario.** A registration/model/airline field is questioned later; the snapshot proves the value but cannot identify source/provider version or retrieval completion time used for that enrichment.
4. **Impact.** Aircraft enrichment is difficult to audit, reproduce or distinguish across provider changes, weakening trust in durable feature evidence.
5. **Severity rationale.** **P1 retrospective.** External enrichment affects durable analytical data; missing provenance prevents reliable attribution/reconstruction of accepted values.
6. **Existing guarantees violated.** Available/partial external enrichment must identify its source, provider contract and retrieval evidence.
7. **Considered solutions.** Use global source lists only; record source/version only; record source/version/retrieval time while keeping wall-clock retrieval time outside deterministic input identity.
8. **Chosen remediation.** Require production aircraft source name/provider version when an aircraft provider is configured; persist source, provider version and retrieval completion time; include stable source identity in composition processing identity but exclude retrieval wall-clock time from deterministic fingerprint.
9. **Why this solution was selected.** It gives durable audit provenance without making identical semantic input hash differently merely because it was retrieved later.
10. **Rejected alternatives.** Generic source lists do not identify enrichment provider contract; hashing retrieval time would destroy deterministic replay for unchanged inputs.
11. **Trade-offs.** Provider construction must supply stable provenance identifiers and snapshots carry additional provenance fields.
12. **Regression tests / protection.** Constructor tests require source/version for configured enrichment; validator requires complete provenance for available/partial aircraft evidence; fingerprint tests distinguish stable source identity from retrieval time; quality/provenance audit protects the boundary.
13. **Adversarial review findings.** Retrieval time is provenance but not semantic deterministic input; source/provider identity can affect interpretation and therefore participates in processing identity.
14. **Remediation iterations.** Aircraft output entered fingerprint identity in Document 109 and temporal update evidence in 110; `3cdd9b15…` completes source/version/retrieval attribution.
15. **Residual risks and limitations.** Provenance identifies the configured provider contract, not an immutable upstream dataset revision unless the provider exposes one.
16. **Operational or deployment consequences.** Processing advances to v5 and production composition supplies explicit aircraft metadata provenance identifiers.
17. **Exact evidence.** Implementation commit `3cdd9b1532609c872343d00626ba44a9c9855609`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-DATA-149=CLOSED`.
19. **Prevention / future guard.** New external enrichment groups must define source identity, provider/version identity, retrieval provenance and which of those dimensions belong in deterministic processing identity.
