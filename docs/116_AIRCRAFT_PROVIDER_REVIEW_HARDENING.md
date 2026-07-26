# Aircraft Provider Review Hardening

## Scope

This increment closes the current, reproducible findings from the static review of
`internal/features/aircraftprovider`. The source review was created against commit
`bb9f351`; stale findings already closed by later extractor and composition work are
classified rather than reimplemented.

## Implemented corrections

1. Cache lookup and in-flight registration now happen under one mutex acquisition.
2. A shared lookup uses a cancellation-independent context with a bounded timeout.
   Each caller waits with its own context and may cancel independently.
3. The in-memory cache is bounded and expired entries are pruned on every acquire
   and store path. Capacity pressure evicts the entry with the earliest expiry.
4. The default not-found classifier recognizes the domain `aircraft.ErrNotFound`
   contract instead of depending on PostgreSQL errors.
5. A successful lookup must return a present, valid, matching ICAO24 identity.
6. Nil request contexts are rejected instead of silently replaced.
7. Aircraft provider processing advances to provider generation 4. Composition
   processing identity changes through the provider version manifest.

## Historical policy classification

The cache remains keyed by canonical ICAO24. This is deliberate: temporal policy is
applied separately for every request after shared lookup or cache retrieval. Metadata
newer than `AsOfTime` is returned as unavailable evidence. A true historical metadata
lookup requires effective validity intervals or dataset revisions from the data source;
this package does not invent unavailable history or multiply identical cache entries by
request timestamp.

## Deliberately rejected review suggestions

- `AsOfTime` is not added to the cache key without a versioned historical source.
- Idiomatic Go constructors returning `nil, error` are retained.
- Nil-safe `Unwrap` behavior is retained.
- The provider is not split into artificial public classes; cache and coordination
  responsibilities are separated through private methods and explicit invariants.

## Closure markers

```text
AIRCRAFT_PROVIDER_ACQUIRE=ATOMIC
SHARED_LOOKUP_CANCELLATION=ISOLATED
SHARED_LOOKUP_TIMEOUT=BOUNDED
AIRCRAFT_CACHE_CAPACITY=BOUNDED
AIRCRAFT_CACHE_EXPIRY_SWEEP=ENFORCED
AIRCRAFT_DOMAIN_NOT_FOUND=DEFAULT
AIRCRAFT_LOOKUP_IDENTITY=REQUIRED
AIRCRAFT_TEMPORAL_GATE=RETAINED
STALE_REVIEW_FINDINGS=CLASSIFIED
AIRCRAFT_PROVIDER_GENERATION=v4
AIRCRAFT_PROVIDER_REVIEW_STATUS=CLOSED
OPEN_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
```
