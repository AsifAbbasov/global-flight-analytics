# Extractor Composition Review Hardening

## Scope

This increment re-evaluates the external review of
`internal/features/extractorcomposition` against the current repository state at
baseline `bcf7ff3e1a83024ee346c16638de0b389baf7e7a`.

## Findings already closed before this increment

- Geographic precision already participated in deterministic fingerprint identity.
- Component versions already participated in deterministic fingerprint identity.
- Aircraft not-found policy version already participated in identity.
- `NewExtractor` had already been removed.
- Typed-nil aircraft lookup rejection was already enforced.
- PostgreSQL feature materialization already reached the composition in production code.
- Aircraft cache hits already applied the historical `AsOfTime` gate per request.

## Changes in this increment

- The complete typed processing manifest is persisted in feature provenance.
- Optional aircraft enrichment can be disabled explicitly without a fake lookup.
- Aircraft caching can be disabled explicitly while preserving request coalescing.
- Enrichment mode and cache mode participate in deterministic processing identity.
- Composition exposes the extractor through its behavioral interface.
- Construction is decomposed into validation and component-specific helpers.
- Provider generation advances to version 3.
- Extractor and composition generations advance to version 6.
- Feature processing generation advances to version 6.

## Review disagreements

The cache key remains ICAO24-only by design. Cached current metadata is filtered
against every request's `AsOfTime` after retrieval, so adding `AsOfTime` to the
cache key would duplicate identical source records without improving temporal
safety. A true historical dataset revision or effective-validity interval cannot
be invented by composition when the source repository does not provide one.

The private value-returning composition configuration remains intentionally flat.
It is small, cannot be initialized with public raw literals, and already separates
construction policy from the pure extractor. Additional nested configuration
layers would add ceremony without strengthening an invariant.

The low-level geographic builder retains zero as its default sentinel, while
production composition rejects zero after `DefaultConfig` has resolved an
explicit precision. Production processing therefore never confuses an omitted
precision with an intentional precision.

## Closure evidence

```text
STALE_REVIEW_FINDINGS=CLASSIFIED
PROCESSING_MANIFEST_PERSISTENCE=ENFORCED
OPTIONAL_AIRCRAFT_ENRICHMENT=EXPLICIT
AIRCRAFT_CACHE_DISABLE_MODE=EXPLICIT
CACHE_AS_OF_POLICY=PER_REQUEST
PRODUCTION_REACHABILITY=ENFORCED
TYPED_NIL_LOOKUP_REJECTION=ENFORCED
EXTRACTOR_PROCESSING_GENERATION=v6
AIRCRAFT_PROVIDER_GENERATION=v3
DATABASE_MIGRATION=NOT_REQUIRED
```

No PostgreSQL migration is required because feature snapshots already persist the
versioned feature payload as JSON and processing identity isolation already uses
the existing `processing_version` column.
