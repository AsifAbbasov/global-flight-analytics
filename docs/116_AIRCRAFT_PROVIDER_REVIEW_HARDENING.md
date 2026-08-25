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

## Canonical remediation history

Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, and Continuous Integration evidence. The severity labels below are retrospective classifications.

### GFA-CONC-161 — Cache acquire and in-flight registration were not one atomic decision

1. **Finding / symptom:** cache lookup and creation of the shared in-flight lookup were separate coordination steps.
2. **Root cause:** one logical acquire operation was split across synchronization boundaries.
3. **Failure scenario:** concurrent callers for the same ICAO24 could both miss cache/coalescing state and start redundant upstream lookups.
4. **Impact:** duplicated provider work, weaker request-coalescing guarantees, and avoidable pressure on a free external-data path.
5. **Severity rationale:** **P1 retrospective** because the defect was a concurrency race in a shared provider boundary and could multiply external calls under load.
6. **Existing guarantees violated:** one active lookup per canonical aircraft identity and deterministic shared-request coordination.
7. **Considered solutions:** a separate singleflight dependency, a second mutex-protected registry, or one atomic acquire critical section.
8. **Chosen remediation:** cache inspection and in-flight registration occur under one mutex acquisition.
9. **Why this solution was selected:** it preserves the package's existing bounded cache/coalescing design without another concurrency subsystem.
10. **Rejected alternatives:** introducing a new public coordination abstraction or external singleflight layer without a demonstrated need.
11. **Trade-offs:** the mutex protects a slightly wider section, but no network I/O occurs while it is held.
12. **Regression tests / protection:** concurrency tests and `aircraftproviderreviewaudit` protect atomic acquire behavior.
13. **Adversarial review findings:** later closure review also required cancellation isolation and cache bounds; those are separate findings below.
14. **Remediation iterations:** closed in implementation commit `92691d993d7340112399a40bd9ecbc62ddb240ad`.
15. **Residual risks and limitations:** process-local coalescing does not deduplicate lookups across multiple application replicas.
16. **Operational or deployment consequences:** no database migration or deployment topology change; upstream request amplification is reduced.
17. **Exact evidence:** commit `92691d993d7340112399a40bd9ecbc62ddb240ad`, provider concurrency tests, and permanent Backend Quality review audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** any new cache/coalescing path must preserve one atomic acquire decision and remain covered by race tests.

### GFA-CONC-162 — Shared lookup lifetime was coupled to one caller's cancellation

1. **Finding / symptom:** a coalesced lookup could inherit a caller-owned cancellation lifetime.
2. **Root cause:** shared provider work and individual request lifetime were not modeled as separate ownership domains.
3. **Failure scenario:** the caller that happened to lead the lookup could cancel and abort work still needed by other waiting callers.
4. **Impact:** unrelated requests could fail spuriously and shared lookup behavior depended on caller timing.
5. **Severity rationale:** **P1 retrospective** because cancellation leakage crossed request boundaries in concurrent production code.
6. **Existing guarantees violated:** independent caller cancellation and shared-work ownership.
7. **Considered solutions:** ignore cancellation, let the leader own the lookup, or give shared work an independent bounded context.
8. **Chosen remediation:** shared lookup uses an independent context with a finite timeout; each waiter observes its own context separately.
9. **Why this solution was selected:** it isolates caller lifetimes while keeping provider work bounded.
10. **Rejected alternatives:** unbounded `context.Background()` work and leader-owned cancellation.
11. **Trade-offs:** upstream work can continue briefly after all callers leave, but only until the bounded provider timeout.
12. **Regression tests / protection:** cancellation/coalescing tests plus race detector coverage.
13. **Adversarial review findings:** timeout ownership was evaluated together with cancellation isolation and retained as an explicit bound.
14. **Remediation iterations:** implementation commit `92691d993d7340112399a40bd9ecbc62ddb240ad`.
15. **Residual risks and limitations:** timeout value is still an operational policy and may need tuning from measured provider latency.
16. **Operational or deployment consequences:** fewer cross-request cancellation failures; no infrastructure change.
17. **Exact evidence:** `92691d993d7340112399a40bd9ecbc62ddb240ad`, provider cancellation tests, Backend Race Safety.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** shared asynchronous work must never directly inherit a single request's cancellation authority.

### GFA-PERF-163 — Aircraft metadata cache had no bounded lifecycle guarantee

1. **Finding / symptom:** cache growth and expiry cleanup were not governed by an explicit capacity-and-sweep contract.
2. **Root cause:** time-to-live behavior existed without a complete memory-bound invariant.
3. **Failure scenario:** a long-running process observing many aircraft identities could retain unnecessary entries and grow memory use.
4. **Impact:** avoidable memory pressure and degraded long-lived process predictability.
5. **Severity rationale:** **P2 retrospective** because it is a bounded-resource reliability problem rather than analytical corruption.
6. **Existing guarantees violated:** bounded in-memory infrastructure for the free-tier modular-monolith deployment model.
7. **Considered solutions:** unbounded map plus periodic goroutine, Least Recently Used cache dependency, or capacity-bounded map with opportunistic expiry pruning.
8. **Chosen remediation:** explicit capacity, expiry pruning on acquire/store, and earliest-expiry eviction under pressure.
9. **Why this solution was selected:** simple deterministic behavior without a background maintenance goroutine or external dependency.
10. **Rejected alternatives:** Redis/external cache and speculative cache frameworks.
11. **Trade-offs:** pruning work occurs on foreground cache operations; eviction policy optimizes expiry rather than recency.
12. **Regression tests / protection:** capacity, expiry, and cache behavior tests; permanent provider audit.
13. **Adversarial review findings:** cache remains ICAO24-keyed by design because temporal acceptance happens after retrieval.
14. **Remediation iterations:** `92691d993d7340112399a40bd9ecbc62ddb240ad`.
15. **Residual risks and limitations:** the cache is per-process and provides no cross-replica sharing.
16. **Operational or deployment consequences:** deterministic memory ceiling for provider cache state.
17. **Exact evidence:** implementation commit `92691d993d7340112399a40bd9ecbc62ddb240ad` and provider cache regression tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every new process-local cache must declare capacity, expiry, and cleanup semantics.

### GFA-ARCH-164 — Default provider not-found policy depended on PostgreSQL-specific error semantics

1. **Finding / symptom:** default not-found classification was coupled to storage-layer PostgreSQL errors instead of the aircraft-domain contract.
2. **Root cause:** provider composition leaked repository implementation details upward.
3. **Failure scenario:** replacing or adapting the repository could change not-found behavior despite returning the domain's canonical absence error.
4. **Impact:** brittle portability and inconsistent negative-cache/evidence behavior.
5. **Severity rationale:** **P2 retrospective** because incorrect absence classification can alter provider evidence and caching, but the defect is primarily a boundary contract violation.
6. **Existing guarantees violated:** domain errors own domain absence semantics; provider must not require a concrete database implementation.
7. **Considered solutions:** retain pgx checks, inject every classifier manually, or use `aircraft.ErrNotFound` as the default domain contract.
8. **Chosen remediation:** default classifier recognizes the aircraft-domain not-found error.
9. **Why this solution was selected:** keeps storage details inside repository adapters and makes provider behavior implementation-agnostic.
10. **Rejected alternatives:** provider-level pgx coupling and mandatory custom classifiers for the normal path.
11. **Trade-offs:** repository adapters must correctly translate storage absence to the domain error.
12. **Regression tests / protection:** provider not-found tests and review audit.
13. **Adversarial review findings:** custom not-found policies remain versioned in processing identity where they are intentionally supplied.
14. **Remediation iterations:** `92691d993d7340112399a40bd9ecbc62ddb240ad`.
15. **Residual risks and limitations:** custom external adapters can still misclassify errors if they violate the domain contract.
16. **Operational or deployment consequences:** no deployment change; improves repository replaceability.
17. **Exact evidence:** implementation commit `92691d993d7340112399a40bd9ecbc62ddb240ad` and domain-not-found regression tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** provider defaults may depend on domain contracts, not concrete database-driver errors.

### GFA-DATA-165 — Successful aircraft lookup did not require a valid matching identity

1. **Finding / symptom:** a successful lookup result could be accepted without proving that returned ICAO24 identity was present, valid, and equal to the requested canonical identity.
2. **Root cause:** successful transport/repository execution was treated as sufficient evidence of semantic identity.
3. **Failure scenario:** a provider adapter returns an empty, malformed, or mismatched aircraft row and feature enrichment attaches metadata to the wrong aircraft.
4. **Impact:** silent provenance and analytical feature corruption.
5. **Severity rationale:** **P1 retrospective** because this can associate real metadata with the wrong durable flight-feature identity.
6. **Existing guarantees violated:** canonical aircraft identity and input-fingerprint correctness.
7. **Considered solutions:** trust repository lookup key, accept empty returned identity, or validate the returned canonical identity explicitly.
8. **Chosen remediation:** successful lookups require a present valid ICAO24 matching the requested canonical value.
9. **Why this solution was selected:** fail-closed semantic verification at the provider boundary prevents adapter mistakes from becoming durable features.
10. **Rejected alternatives:** silently overwriting returned identity with the request value or accepting missing identity as implied.
11. **Trade-offs:** stricter adapters may expose previously hidden data-quality defects as errors.
12. **Regression tests / protection:** missing/invalid/mismatched identity tests and permanent provider review audit.
13. **Adversarial review findings:** historical cache-key policy was retained because identity validation and per-request temporal filtering already protect the relevant invariants.
14. **Remediation iterations:** `92691d993d7340112399a40bd9ecbc62ddb240ad`.
15. **Residual risks and limitations:** identity equality does not prove the upstream metadata itself is otherwise correct; source provenance remains required.
16. **Operational or deployment consequences:** malformed source rows fail rather than silently enriching snapshots.
17. **Exact evidence:** commit `92691d993d7340112399a40bd9ecbc62ddb240ad`, identity mismatch tests, feature-processing audits.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every enrichment adapter must validate returned semantic identity before publishing evidence.

### GFA-OPS-166 — Nil provider request contexts were silently substituted

1. **Finding / symptom:** provider calls accepted nil context and replaced caller intent with a background context.
2. **Root cause:** convenience fallback obscured an invalid lifecycle contract.
3. **Failure scenario:** a caller accidentally passes nil and the provider performs work that cannot be cancelled or traced through the intended request lifetime.
4. **Impact:** lifecycle leaks and inconsistent cancellation semantics.
5. **Severity rationale:** **P2 retrospective** because this is a request-lifecycle correctness defect, not direct analytical corruption.
6. **Existing guarantees violated:** explicit context ownership at repository/provider boundaries.
7. **Considered solutions:** keep background fallback, panic, or return a typed required-context error.
8. **Chosen remediation:** nil contexts are rejected explicitly.
9. **Why this solution was selected:** fail-fast behavior preserves caller ownership without panics.
10. **Rejected alternatives:** implicit `context.Background()` substitution.
11. **Trade-offs:** callers that previously relied on undocumented fallback must pass a real context.
12. **Regression tests / protection:** nil-context provider tests and strict review audit.
13. **Adversarial review findings:** independent shared-work context remains allowed only inside the explicitly bounded coalescing implementation.
14. **Remediation iterations:** `92691d993d7340112399a40bd9ecbc62ddb240ad`.
15. **Residual risks and limitations:** context misuse outside the provider remains the responsibility of upstream components.
16. **Operational or deployment consequences:** clearer cancellation and timeout behavior.
17. **Exact evidence:** implementation commit `92691d993d7340112399a40bd9ecbc62ddb240ad` and provider context tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** production provider/repository entry points must reject nil caller contexts unless a cleanup-specific independent context is explicitly documented.
