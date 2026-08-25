# Document 110 — Aircraft Metadata Temporal Safety

Status: IMPLEMENTED
Baseline commit: a4c563112abd459c90e23e33d191c5f059e5044f

## 1. Problem

The aircraft feature provider cached current metadata only by ICAO24. Historical
feature materialization could therefore consume registration, model, airline or
country values whose underlying records were updated after the requested feature
`AsOfTime`.

The repository does not own a complete versioned aircraft-history table, so it
must not invent historical values.

## 2. Conservative temporal contract

The PostgreSQL aircraft read model now returns one aggregate metadata update
timestamp. It is the greatest `updated_at` value across:

- aircraft;
- aircraft model;
- airline;
- country.

The extractor passes the feature request `AsOfTime` into the aircraft provider.
The provider applies temporal policy after cache lookup and after in-flight
request coalescing.

If aggregate metadata was updated after `AsOfTime`, the aircraft feature group
is returned as unavailable with:

```text
aircraft_metadata_newer_than_feature_as_of
```

No current value is presented as historical evidence.

## 3. Cache semantics

The cache remains keyed by normalized ICAO24 because it stores one current
metadata result. Temporal acceptance is evaluated independently for every
request. A recent request and a historical request may therefore share one
lookup without sharing one temporal decision.

The aggregate metadata timestamp is included in `AircraftFeatures`, and thus in
the extraction input fingerprint introduced by Document 109.

## 4. Generation boundary

Aircraft provider, extractor, extractor composition and feature processing
generations advance to version 3 where applicable. Earlier snapshots remain
readable through their stored processing versions.

## 5. Permanent evidence

```text
AIRCRAFT_METADATA_TEMPORAL_GATE=ENFORCED
AIRCRAFT_CACHE_AS_OF_ISOLATION=ENFORCED
FUTURE_AIRCRAFT_METADATA_LEAKAGE=CLOSED
HISTORICAL_AIRCRAFT_VALUES_NOT_INVENTED=ENFORCED
```

---

## Canonical remediation history

### GFA-DATA-138 — current aircraft metadata could leak into historical feature materialization

1. **Finding / symptom.** Historical feature extraction could consume current registration/model/airline/country metadata even when those records were updated after the requested feature `AsOfTime`.
2. **Root cause.** The aircraft provider cached/read current metadata by ICAO24 without carrying a repository-owned metadata update boundary into the per-request historical acceptance decision.
3. **Failure scenario.** Metadata changes on July 25; a feature snapshot is materialized for July 20; cached/current values are accepted and presented as if they were valid historical evidence on July 20.
4. **Impact.** Historical analytics can contain future knowledge and misrepresent aircraft attributes that the repository cannot prove existed at the requested time.
5. **Severity rationale.** **P1 retrospective.** This is temporal data leakage across a historical analytics boundary and can make persisted historical features factually false.
6. **Existing guarantees violated.** Historical materialization must not use evidence known to be newer than its as-of boundary, and the system must not invent historical versions it does not own.
7. **Considered solutions.** Put `AsOfTime` in the cache key; maintain a full aircraft-history table; ignore metadata update time; reuse current cache but apply a conservative per-request temporal gate.
8. **Chosen remediation.** Return aggregate `MetadataUpdatedAt` from PostgreSQL, pass `AsOfTime` to the provider, and after cache/coalescing mark metadata unavailable whenever the aggregate update timestamp is later than the request boundary.
9. **Why this solution was selected.** It prevents future leakage using evidence the repository actually owns while avoiding fake historical reconstruction or duplicating identical current records by request time.
10. **Rejected alternatives.** `AsOfTime` cache keys duplicate current metadata without creating historical truth; inventing historical revisions is unsupported; ignoring timestamps preserves leakage.
11. **Trade-offs.** Historical requests may lose aircraft enrichment even when some individual fields were unchanged, because the aggregate timestamp is intentionally conservative.
12. **Regression tests / protection.** Provider tests prove temporal policy runs after cache lookup, recent and historical requests can share one lookup but receive different acceptance, and unavailable evidence preserves the update timestamp/reason.
13. **Adversarial review findings.** Temporal gating must run after cache hits and in-flight coalescing for every requester; otherwise the first request's temporal decision could contaminate later requests.
14. **Remediation iterations.** Composition identity was first bound in Document 109; `f574911a…` added aggregate update evidence and per-request temporal policy, advancing processing generation to v3. Document 115 deliberately retains ICAO24-only cache keys because this per-request gate already provides the required isolation.
15. **Residual risks and limitations.** The aggregate timestamp proves only that current metadata is not newer than the request boundary; it does not provide complete effective-validity history for aircraft attributes.
16. **Operational or deployment consequences.** No new history table is introduced. Repository aircraft reads include aggregate metadata update time, and historical extraction may return aircraft enrichment unavailable rather than fabricate older values.
17. **Exact evidence.** Implementation commit `f574911a27b4bad10ddf137689b35286fdb485d3` (`fix: enforce aircraft metadata temporal safety`). Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, PRs, and CI evidence.
18. **Final canonical status.** `GFA-DATA-138=CLOSED`.
19. **Prevention / future guard.** Any current-state enrichment used in historical materialization must expose a temporal validity/update boundary and be evaluated independently for each request; absent such evidence, historical use must fail closed.
