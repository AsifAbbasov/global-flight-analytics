# Exact Deduplication and Airplanes.live Telemetry Hardening

Status: Implemented Engineering Contract v1.0
Project: Global Flight Analytics

## Purpose

This increment closes the confirmed correctness gaps in exact in-memory observation deduplication and Airplanes.live telemetry mapping.

## Exact observation equality

`RemoveExactDuplicates` now compares the complete canonical observation payload:

```text
aircraft identity and callsign
observation time
position
both altitude values and statuses
velocity, heading and vertical rate values and availability
on-ground value and availability
telemetry availability knowledge
country, squawk and Special Purpose Indicator
position source
aircraft category and availability
provider source
```

Internal persistence identifiers are deliberately excluded:

```text
Flight State identifier
Flight identifier
Aircraft identifier
Ingestion Run identifier
```

Those identifiers may differ during replay without changing the source observation itself.

## Nullable telemetry

Airplanes.live numeric telemetry now retains the distinction between:

```text
observed zero
missing field
explicit null
invalid value
```

Missing, null and invalid telemetry does not set the corresponding canonical availability flag.

## Time safety

Provider response milliseconds and per-aircraft `seen` seconds are checked before conversion to `int64` or `time.Duration`.

Invalid response time produces an unknown zero timestamp. Invalid or overflowing `seen` evidence preserves the valid provider response time instead of performing an unsafe subtraction.

## Provider construction

`NewProvider(nil)` refuses to construct a usable provider. Direct use of a nil provider receiver returns `ErrClientRequired` rather than dereferencing a nil transport.

## Verification

The increment requires:

```text
targeted deduplicator and Airplanes.live tests
targeted race detector tests
full backend tests
go vet
existing code review policy gates
git diff check
```

## Remaining boundary

This increment does not change distributed provider rate limiting or fallback health policy. Those are separate architecture decisions and remain classified independently.

---

## Canonical remediation history

### GFA-DATA-076 — exact in-memory deduplication compared an incomplete observation identity

1. **Finding / symptom.** `RemoveExactDuplicates` could treat observations as identical without comparing the complete canonical source payload.
2. **Root cause.** Equality was implemented over a partial subset of Flight State fields rather than the semantic source-observation contract.
3. **Failure scenario.** Two observations share identity/time/position but differ in altitude status, availability, motion telemetry, squawk/category/provenance or another omitted field and one is removed as an exact duplicate.
4. **Impact.** Real source evidence can be lost before persistence and downstream analytics may see a simplified/fabricated history.
5. **Severity rationale.** **P1 retrospective.** The defect can delete materially distinct source observations on the ingestion data path.
6. **Existing guarantees violated.** Exact deduplication may remove only observations whose complete canonical source evidence is equal; internal persistence IDs are not source identity.
7. **Considered solutions.** Compare a small stable key; serialize entire structs including internal IDs; maintain an explicit semantic equality function over canonical source fields.
8. **Chosen remediation.** Compare the complete canonical observation payload while deliberately excluding Flight State, Flight, Aircraft and Ingestion Run persistence identifiers.
9. **Why this solution was selected.** It preserves every source-semantic distinction while allowing replay-generated internal identifiers to differ.
10. **Rejected alternatives.** Partial keys lose evidence; raw struct equality would couple deduplication to internal persistence identity and incidental fields.
11. **Trade-offs.** Equality code must evolve whenever a new canonical observation field is introduced.
12. **Regression tests / protection.** Targeted deduplicator tests cover complete equality and differences in canonical payload; race/full backend gates provide broader protection.
13. **Adversarial review findings.** Legitimate zero values and availability flags must participate independently; source/provenance differences must prevent deduplication even when coordinates/time match.
14. **Remediation iterations.** The equality contract was explicitly documented field-by-field so future schema additions have an obvious audit surface.
15. **Residual risks and limitations.** Exact equality is not near-duplicate/noise reduction; semantically close but non-identical observations intentionally remain separate.
16. **Operational or deployment consequences.** Slightly more observations may survive deduplication, preserving correctness at the cost of modest additional downstream work.
17. **Exact evidence.** Historical implementation commit `eef7fdc056ebef71f95cfd17ce986dcf429f6c62` (`fix: harden exact deduplication and provider telemetry`). Historical pull-request/reviewer evidence unavailable.
18. **Final canonical status.** `GFA-DATA-076=CLOSED`.
19. **Prevention / future guard.** Adding a canonical Flight State field requires updating semantic exact-equality tests; persistence-only identifiers remain explicitly excluded.

### GFA-DATA-077 — Airplanes.live missing/null/invalid telemetry could collapse into observed zero

1. **Finding / symptom.** Numeric telemetry mapping did not reliably preserve the distinction between observed zero, absent field, JSON null and invalid value.
2. **Root cause.** Provider wire decoding used value-oriented numeric fields where presence/validity was part of the domain evidence.
3. **Failure scenario.** Missing heading, vertical rate or another optional numeric metric is decoded as zero and marked available, making absence look like a real measurement.
4. **Impact.** Traffic, trajectory and analytical consumers can reason from fabricated telemetry availability and values.
5. **Severity rationale.** **P1 retrospective.** This directly changes source evidence semantics and can contaminate analytics.
6. **Existing guarantees violated.** Missing evidence must remain missing; legitimate zero must remain distinguishable from unavailable or invalid telemetry.
7. **Considered solutions.** Treat zero as missing; keep value fields and infer presence from provider conventions; use presence-aware nullable decoding and explicit canonical availability flags.
8. **Chosen remediation.** Decode optional Airplanes.live numeric telemetry with presence/validity awareness and set canonical availability only for valid observed values.
9. **Why this solution was selected.** It preserves valid zero without inventing values for missing/null/invalid source fields.
10. **Rejected alternatives.** Zero-as-missing destroys legitimate zero evidence; inference from numeric value alone cannot distinguish missing from observed zero.
11. **Trade-offs.** Wire models and mapping logic become more explicit and verbose because availability is first-class evidence.
12. **Regression tests / protection.** Provider mapping tests cover zero, missing, explicit null and invalid telemetry separately.
13. **Adversarial review findings.** Every optional metric must test all four states; a valid zero must set availability true while null/missing/invalid must not.
14. **Remediation iterations.** Availability semantics were aligned with the earlier end-to-end nullable telemetry policy rather than introducing provider-specific zero conventions.
15. **Residual risks and limitations.** Provider documentation/behavior can add new sentinel conventions requiring explicit adapter updates.
16. **Operational or deployment consequences.** Some previously displayed/stored zeros become unavailable evidence, improving truthfulness but potentially reducing apparent completeness.
17. **Exact evidence.** Historical implementation commit `eef7fdc056ebef71f95cfd17ce986dcf429f6c62`. Historical PR/reviewer evidence unavailable.
18. **Final canonical status.** `GFA-DATA-077=CLOSED`.
19. **Prevention / future guard.** Provider adapters must use presence-aware decoding for optional telemetry and maintain observed-zero versus missing tests.

### GFA-DATA-078 — provider time conversion could overflow or perform unsafe subtraction

1. **Finding / symptom.** Provider response milliseconds and per-aircraft `seen` seconds could overflow conversion to `int64`/`time.Duration` or produce unsafe observation-time subtraction.
2. **Root cause.** External numeric time values were converted before validating representable bounds.
3. **Failure scenario.** Malformed/extreme response time or `seen` value overflows duration arithmetic and creates an invalid timestamp.
4. **Impact.** Observation ordering/freshness evidence can become nonsensical, and arithmetic behavior can diverge from provider intent.
5. **Severity rationale.** **P2 retrospective.** It is a correctness boundary on untrusted provider input, but valid timestamps remain recoverable through conservative fallback behavior.
6. **Existing guarantees violated.** External time values must be range-checked before integer/duration conversion; invalid age evidence must not corrupt an otherwise valid provider response timestamp.
7. **Considered solutions.** Blind casts; clamp arbitrary extremes; reject the whole batch; validate bounds and preserve a safe known timestamp where possible.
8. **Chosen remediation.** Check representable ranges before conversion; invalid response time becomes unknown zero time, while invalid/overflowing `seen` preserves the valid provider response time instead of subtracting.
9. **Why this solution was selected.** It fails conservatively without allowing malformed age metadata to destroy valid higher-level response time evidence.
10. **Rejected alternatives.** Blind casts overflow; arbitrary clamping fabricates precision; rejecting all data for bad optional age evidence is unnecessarily destructive.
11. **Trade-offs.** Invalid `seen` loses per-aircraft age precision and uses the response timestamp as the conservative observation-time evidence.
12. **Regression tests / protection.** Provider time conversion boundary and overflow tests are part of targeted Airplanes.live coverage.
13. **Adversarial review findings.** Bounds must be checked before multiplication/subtraction; invalid response-level time and invalid per-item age have different fallback semantics.
14. **Remediation iterations.** Time validation was separated into response-time and age-time ownership rather than using one generic conversion fallback.
15. **Residual risks and limitations.** A representable but semantically wrong provider timestamp cannot be detected by overflow checks alone.
16. **Operational or deployment consequences.** Malformed time evidence is surfaced as unknown/conservative time rather than process-level arithmetic corruption.
17. **Exact evidence.** Historical implementation commit `eef7fdc056ebef71f95cfd17ce986dcf429f6c62`. Historical PR/reviewer evidence unavailable.
18. **Final canonical status.** `GFA-DATA-078=CLOSED`.
19. **Prevention / future guard.** All external timestamp/duration conversions must validate bounds before multiplication, cast or subtraction and include overflow regression cases.

### GFA-REL-079 — Airplanes.live provider could be constructed or called without a client

1. **Finding / symptom.** A nil HTTP client/provider path could reach dereference behavior instead of a controlled domain/integration error.
2. **Root cause.** Constructor and receiver lifecycle contracts did not fail closed on missing transport dependency.
3. **Failure scenario.** Misconfigured composition or direct nil receiver use reaches provider methods and panics/dereferences instead of returning a typed failure.
4. **Impact.** Configuration errors can become process failures and are harder to diagnose/recover from.
5. **Severity rationale.** **P2 retrospective.** The defect is a reliability/constructor-contract issue; it requires misconfiguration or invalid direct use rather than valid provider data.
6. **Existing guarantees violated.** Required provider dependencies must be validated at construction/use boundaries and failures must be controlled.
7. **Considered solutions.** Assume composition always passes a client; panic in constructor; substitute a default client; reject nil and return `ErrClientRequired` on direct nil use.
8. **Chosen remediation.** `NewProvider(nil)` refuses usable construction and nil provider/client use returns `ErrClientRequired`.
9. **Why this solution was selected.** It keeps dependency requirements explicit without hidden default networking behavior or panics.
10. **Rejected alternatives.** Implicit default clients hide configuration; panic-based contracts are unsuitable for normal dependency validation.
11. **Trade-offs.** Callers must handle a constructor/use error path explicitly.
12. **Regression tests / protection.** Provider constructor and nil-client tests in targeted adapter coverage.
13. **Adversarial review findings.** Both constructor misuse and direct nil-receiver method calls must remain controlled because tests/adapters can bypass the preferred composition path.
14. **Remediation iterations.** The final contract protects both normal construction and defensive direct use.
15. **Residual risks and limitations.** A non-nil but incorrectly configured client can still fail at transport time, which is a separate typed network error path.
16. **Operational or deployment consequences.** Misconfiguration fails deterministically before/at provider use instead of causing an uncontrolled panic.
17. **Exact evidence.** Historical implementation commit `eef7fdc056ebef71f95cfd17ce986dcf429f6c62`. Historical PR/reviewer evidence unavailable.
18. **Final canonical status.** `GFA-REL-079=CLOSED`.
19. **Prevention / future guard.** Provider constructors must validate required dependencies and tests must include nil/typed-nil failure behavior where applicable.
