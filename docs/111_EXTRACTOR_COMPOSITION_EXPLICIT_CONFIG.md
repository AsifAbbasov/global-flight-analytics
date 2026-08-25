# Document 111 — Extractor Composition Explicit Configuration

Status: IMPLEMENTED
Baseline commit: f574911a27b4bad10ddf137689b35286fdb485d3

## 1. Problem

`extractorcomposition.Config` previously exposed raw numeric fields where zero
meant both the Go zero value and an implicit request for package defaults.

This made incomplete production configuration look valid:

```go
extractorcomposition.Config{
    AircraftLookup: lookup,
}
```

The actual geographic precision and cache durations were only discovered during
construction. A caller could not distinguish omission from an intentional zero.

## 2. Explicit configuration contract

External callers now start with:

```go
extractorcomposition.DefaultConfig(lookup)
```

They may derive a modified value through immutable-style methods:

```go
config := extractorcomposition.DefaultConfig(lookup).
    WithGeographicCellPrecision(3).
    WithAircraftCacheDurations(20*time.Minute, 2*time.Minute).
    WithAircraftNotFoundPolicy("policy-v1", classifier).
    WithClock(now)
```

Configuration fields are private to the package. External packages cannot build
a partially initialized extractor composition through a struct literal.

## 3. Zero-value policy

`New` rejects zero geographic precision and zero positive or negative cache
duration. Defaults are installed only by `DefaultConfig`; they are no longer
inferred by `resolveProcessingIdentity`.

This increment does not change effective default values, feature output,
fingerprint content or processing generation. It changes only construction
safety.

## 4. Production wiring

The feature materializer now uses `DefaultConfig` and an explicit versioned
not-found policy. The PostgreSQL feature-pipeline verification command also
uses `DefaultConfig`. In-memory and PostgreSQL pipeline compositions inject
their shared clock through `WithClock`.

## 5. Permanent evidence

```text
EXTRACTOR_COMPOSITION_EXPLICIT_CONFIG=ENFORCED
ZERO_VALUE_CONFIG_SENTINELS=CLOSED
PRODUCTION_CONFIG_LITERAL_BYPASS=REJECTED
CONFIG_VALUE_DERIVATION_NON_MUTATING=ENFORCED
PROCESSING_GENERATION_UNCHANGED=ENFORCED
```

---

## Canonical remediation history

### GFA-CONTRACT-139 — zero-valued extractor composition fields ambiguously meant both omission and package defaults

1. **Finding / symptom.** Public `extractorcomposition.Config` numeric fields accepted zero as both the ordinary Go zero value and an implicit request for default geographic/cache settings.
2. **Root cause.** Default resolution occurred inside construction/processing-identity code while callers could freely build partially initialized public struct literals.
3. **Failure scenario.** Production code supplies only an aircraft lookup; configuration appears complete at compile time, but geographic precision/cache TTLs are silently injected later. A future caller cannot distinguish accidental omission from a deliberate zero/default request.
4. **Impact.** Effective processing policy is hidden at the call site, configuration review is unreliable, and future zero-valued semantics could silently change deterministic processing identity.
5. **Severity rationale.** **P2 retrospective.** This is a production configuration-contract ambiguity affecting deterministic processing policy, but it requires omission/misconfiguration rather than healthy explicit configuration.
6. **Existing guarantees violated.** Production composition should make semantic processing policy explicit and should not overload zero values as hidden sentinels when zero/omission have different meanings.
7. **Considered solutions.** Keep public fields and document zero defaults; use pointer/optional numeric fields; add an explicit `DefaultConfig` constructor with private fields and value-returning overrides.
8. **Chosen remediation.** Make configuration fields private, require callers to start from `DefaultConfig(lookup)`, expose immutable-style `With*` derivation methods, and make `New` reject zero precision/cache durations instead of inferring defaults.
9. **Why this solution was selected.** It keeps the configuration small and typed while making defaults visible at construction ownership and preventing partial public literals.
10. **Rejected alternatives.** Documentation-only sentinel rules remain unenforced; pointer-heavy config adds nil complexity for simple required values; retaining hidden resolution preserves ambiguity.
11. **Trade-offs.** Callers must use constructor/derivation methods rather than convenient struct literals; adding a new required policy needs a deliberate config API extension.
12. **Regression tests / protection.** Tests assert `DefaultConfig` effective defaults, rejection of zero explicit values, non-mutating `With*` derivation and production/verifier use of the default constructor.
13. **Adversarial review findings.** This change must not alter the effective default values or processing generation merely because the construction API becomes explicit; tests preserve fingerprint/output/version stability.
14. **Remediation iterations.** Document 109 first resolved effective zero defaults into processing identity; `ff84eefd…` then moved default ownership to `DefaultConfig` so raw omission could no longer masquerade as explicit policy. Document 115 later adds explicit enrichment/cache modes without reintroducing raw sentinel ambiguity.
15. **Residual risks and limitations.** Low-level builders may still define their own local zero/default semantics; production extractor composition remains responsible for rejecting ambiguity before invoking them.
16. **Operational or deployment consequences.** Production materializer and PostgreSQL verifier now construct extractor composition through `DefaultConfig`; effective runtime defaults and processing generation remain unchanged.
17. **Exact evidence.** Implementation commit `ff84eefdb8ab363e2bdd276f99e49df7235fb50f` (`fix: enforce explicit extractor composition config`). Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-CONTRACT-139=CLOSED`.
19. **Prevention / future guard.** New production composition options must distinguish required values, explicit disabled modes and defaults in the type/config API rather than overloading raw zero values.
