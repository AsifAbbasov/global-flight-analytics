# Document 109 — Extractor Composition Processing Identity

Status: IMPLEMENTED
Baseline commit: abd038c10d1d382843dbaefb8b506efeff5fdeda

## 1. Problem

The durable snapshot key already contains `processing_version`, but the extractor
input fingerprint previously represented only the trajectory. Changing
geographic cell precision, component versions, aircraft cache policy, the
not-found classifier or resolved aircraft metadata could therefore reuse the
same idempotent record incorrectly.

## 2. Contract

The production extractor composition now creates a deterministic processing
identity from:

- every extractor component version;
- effective geographic cell precision;
- effective positive and negative aircraft cache durations;
- an explicit aircraft not-found policy version.

The identity is included in the extractor fingerprint input. Resolved aircraft
features are also included, so mutable lookup results cannot silently replay an
older record.

Custom not-found classifiers require an explicit stable policy version.
Typed-nil aircraft lookup dependencies are rejected. The unused `NewExtractor`
shortcut is removed so production callers retain the complete composition and
its identity evidence.

## 3. Generation boundary

The extractor, extractor composition and processing pipeline generations first
advanced to version 2 in this increment. Later temporal-safety corrections use
version 3. Version 1 and version 2 snapshots remain readable through their
explicit stored processing versions and are not overwritten.

## 4. Permanent evidence

`featureprocessingidentityaudit` requires the identity model, effective policy
inputs, output-sensitive fingerprint, regression tests and absence of
`NewExtractor`.

```text
EXTRACTOR_COMPOSITION_PROCESSING_IDENTITY=ENFORCED
GEOGRAPHIC_PRECISION_FINGERPRINT_COLLISION=CLOSED
AIRCRAFT_METADATA_FINGERPRINT_INPUT=ENFORCED
CUSTOM_NOT_FOUND_POLICY_IDENTITY=ENFORCED
TYPED_NIL_AIRCRAFT_LOOKUP_REJECTED=ENFORCED
```

---

## Canonical remediation history

### GFA-DATA-134 — extractor fingerprint omitted effective composition policy and component identity

1. **Finding / symptom.** The extractor fingerprint represented trajectory evidence but not the effective component versions, geographic precision or aircraft cache policy used to produce features.
2. **Root cause.** Deterministic input identity was owned at the extractor level while composition-specific processing policy remained outside the hashed contract.
3. **Failure scenario.** The same trajectory is reprocessed after geographic precision, component generation or cache policy changes; the old fingerprint still matches and an idempotent durable record can be reused despite different processing semantics.
4. **Impact.** Semantically different feature outputs can collapse onto one replay identity, invalidating reproducibility and processing-version isolation.
5. **Severity rationale.** **P1 retrospective.** This is a durable analytical identity defect capable of silently reusing output produced under a different processing contract.
6. **Existing guarantees violated.** Every deterministic processing input that can materially change durable feature output must participate in processing identity.
7. **Considered solutions.** Rely only on pipeline version; persist policy as diagnostics only; hash the effective composition identity together with extraction input.
8. **Chosen remediation.** Build a deterministic `ProcessingIdentity` from component versions, effective geographic precision, cache durations and not-found policy version, hash it, and include that identity in the extraction fingerprint.
9. **Why this solution was selected.** It binds actual effective configuration rather than only a coarse release version and reuses the existing idempotent fingerprint boundary.
10. **Rejected alternatives.** Diagnostics-only policy cannot prevent collisions; a single pipeline version would require broad version bumps for every composition policy change and still hide effective configuration.
11. **Trade-offs.** Any semantic composition-policy change now requires stable serializable identity and intentional generation/version handling.
12. **Regression tests / protection.** Processing-identity tests prove effective defaults and policy changes alter fingerprints; `featureprocessingidentityaudit` requires the identity model and fingerprint wiring.
13. **Adversarial review findings.** Identity must use resolved/effective configuration rather than raw zero/default sentinels, otherwise two equivalent configurations could hash differently or omitted defaults could remain invisible.
14. **Remediation iterations.** Initial composition identity landed in `a4c56311…`; Document 111 later removes ambiguous zero-value defaults while preserving the same effective identity semantics; Document 115 persists the typed manifest and extends identity with explicit enrichment/cache modes.
15. **Residual risks and limitations.** Correctness still depends on maintainers including every future semantic composition dimension in the typed identity.
16. **Operational or deployment consequences.** Processing generation advanced to v2 so old snapshots remain isolated/readable instead of being silently reused.
17. **Exact evidence.** Implementation commit `a4c563112abd459c90e23e33d191c5f059e5044f` (`fix: bind extractor composition processing identity`). Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-DATA-134=CLOSED`.
19. **Prevention / future guard.** New extractor composition options must declare whether they affect deterministic output and, if they do, be added to processing identity plus fingerprint-change tests.

### GFA-DATA-135 — resolved aircraft metadata was absent from deterministic extraction input identity

1. **Finding / symptom.** Mutable aircraft lookup output could change while the trajectory-only fingerprint remained unchanged.
2. **Root cause.** Aircraft enrichment was treated as derived output rather than an external input to deterministic feature materialization.
3. **Failure scenario.** Registration/model/airline/country metadata changes for the same ICAO24; replay sees the same trajectory fingerprint and returns an older stored feature record instead of reflecting the newly resolved enrichment.
4. **Impact.** Durable feature snapshots can silently preserve stale enrichment while appearing idempotently current for the same extraction request.
5. **Severity rationale.** **P1 retrospective.** Mutable external enrichment could alter analytical output without altering durable input identity.
6. **Existing guarantees violated.** All output-sensitive resolved inputs must participate in deterministic replay identity.
7. **Considered solutions.** Ignore mutable enrichment for fingerprinting; use only aircraft ICAO24; include the resolved normalized aircraft feature payload in fingerprint input.
8. **Chosen remediation.** Include resolved `AircraftFeatures` in the canonical extraction fingerprint payload.
9. **Why this solution was selected.** It binds the actual enrichment consumed by extraction without inventing a separate aircraft revision system the repository does not own.
10. **Rejected alternatives.** ICAO24 alone identifies the aircraft, not the mutable metadata values; ignoring enrichment preserves stale replay collisions.
11. **Trade-offs.** Legitimate metadata changes intentionally create a new fingerprint/durable processing result.
12. **Regression tests / protection.** Processing-identity tests verify aircraft feature changes alter extraction fingerprints; the permanent identity audit requires output-sensitive fingerprinting.
13. **Adversarial review findings.** Retrieval wall-clock time must not be included merely because metadata is fetched again; only deterministic semantic input should drive replay identity. Later Document 113 explicitly keeps retrieval time as provenance, not identity.
14. **Remediation iterations.** Aircraft feature payload entered the fingerprint in `a4c56311…`; later provenance/source identity refinements in Document 113 and typed manifest persistence in Document 115 preserve the distinction between semantic identity and observational provenance.
15. **Residual risks and limitations.** If the upstream aircraft repository changes semantically without changing the normalized fields consumed by features, the fingerprint appropriately remains stable.
16. **Operational or deployment consequences.** Metadata updates can create distinct feature snapshots under the new processing generation rather than replaying older enrichment.
17. **Exact evidence.** Implementation commit `a4c563112abd459c90e23e33d191c5f059e5044f`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-DATA-135=CLOSED`.
19. **Prevention / future guard.** Any new mutable enrichment that affects feature output must either be included in deterministic input identity or explicitly documented as non-deterministic/non-idempotent.

### GFA-CONTRACT-136 — custom aircraft not-found classifier lacked stable policy identity

1. **Finding / symptom.** Callers could supply custom `IsAircraftNotFound` behavior without a stable version describing that classifier in processing identity.
2. **Root cause.** Function values are executable policy but cannot be deterministically serialized or compared as semantic identity.
3. **Failure scenario.** A custom classifier changes which lookup errors are interpreted as unavailable aircraft metadata, but processing identity remains unchanged because the function itself is not representable in the fingerprint.
4. **Impact.** The same durable identity can represent different enrichment availability behavior across releases/configurations.
5. **Severity rationale.** **P2 retrospective.** A custom classifier is required, but when present it changes processing semantics without stable identity.
6. **Existing guarantees violated.** Non-serializable policy functions that affect deterministic output require an explicit stable version token.
7. **Considered solutions.** Hash function pointers; ban custom classifiers; require a caller-owned policy version alongside custom behavior.
8. **Chosen remediation.** Require `AircraftNotFoundPolicyVersion` for custom classifiers and include the normalized version in `ProcessingIdentity`; provide a stable default version for the built-in policy.
9. **Why this solution was selected.** A semantic version token is deterministic, portable and reviewable, unlike runtime function addresses.
10. **Rejected alternatives.** Function pointers are process/build dependent; banning customization would remove a legitimate adapter boundary.
11. **Trade-offs.** Callers modifying custom classifier semantics must also update the policy version deliberately.
12. **Regression tests / protection.** Tests reject custom classifiers without a version and verify policy-version changes alter fingerprint identity; permanent identity audit enforces the contract.
13. **Adversarial review findings.** A version token is meaningful only if callers advance it when behavior changes; automated code cannot infer semantic equivalence of arbitrary functions.
14. **Remediation iterations.** Added in `a4c56311…`; Document 111 later makes the version/default path explicit through `DefaultConfig` and `WithAircraftNotFoundPolicy`.
15. **Residual risks and limitations.** Human discipline is still required to advance a custom policy version when implementation semantics change.
16. **Operational or deployment consequences.** Misconfigured custom classifiers now fail construction instead of silently sharing old processing identity.
17. **Exact evidence.** Implementation commit `a4c563112abd459c90e23e33d191c5f059e5044f`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-CONTRACT-136=CLOSED`.
19. **Prevention / future guard.** Every callback/function policy that can alter deterministic persisted output must have a stable explicit identity token or be excluded from customizable production configuration.

### GFA-ARCH-137 — typed-nil aircraft lookup could pass extractor-composition dependency validation

1. **Finding / symptom.** An interface containing a nil concrete aircraft lookup pointer could compare non-nil and pass composition construction checks.
2. **Root cause.** Required-dependency validation used interface nil semantics rather than concrete nil-aware validation.
3. **Failure scenario.** Composition accepts a typed-nil lookup and the aircraft provider later invokes it, causing panic or implementation-specific nil behavior during feature extraction.
4. **Impact.** Invalid production composition can survive startup validation and fail only when enrichment executes.
5. **Severity rationale.** **P2 retrospective.** Misconfiguration is required, but it bypasses intended fail-fast construction and can cause runtime failure.
6. **Existing guarantees violated.** Required composition dependencies must be concretely usable, not merely non-nil interface values.
7. **Considered solutions.** Document caller responsibility; recover from panic; detect typed nil at composition construction.
8. **Chosen remediation.** Add nil-capable reflection-based `dependencyMissing` validation and reject typed-nil aircraft lookup dependencies.
9. **Why this solution was selected.** It closes the Go interface typed-nil edge at the single construction boundary with deterministic typed failure.
10. **Rejected alternatives.** Documentation cannot enforce safety; panic recovery detects the error too late and obscures construction ownership.
11. **Trade-offs.** Small reflection logic is retained in cold-path validation.
12. **Regression tests / protection.** Composition tests exercise typed-nil lookup rejection; `featureprocessingidentityaudit` requires the guard.
13. **Adversarial review findings.** Reflection must call `IsNil` only for nil-capable kinds; the helper explicitly switches on those kinds.
14. **Remediation iterations.** Closed in `a4c56311…`; later explicit configuration and optional enrichment modes preserve typed-nil handling instead of reintroducing fake dependencies.
15. **Residual risks and limitations.** A non-nil lookup can still be behaviorally unhealthy; this guard addresses construction validity, not external availability.
16. **Operational or deployment consequences.** Invalid lookup composition fails immediately during construction.
17. **Exact evidence.** Implementation commit `a4c563112abd459c90e23e33d191c5f059e5044f`. Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-ARCH-137=CLOSED`.
19. **Prevention / future guard.** Required interface dependencies in composition constructors must use typed-nil-aware validation when pointer implementations are expected.

### Non-finding cleanup

Removal of the unused `NewExtractor` shortcut is recorded as prevention evidence for keeping callers on the complete composition/identity boundary. It is not assigned a separate defect ID because the repository evidence does not demonstrate an independent production failure mode beyond the identity findings above.
