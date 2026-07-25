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
