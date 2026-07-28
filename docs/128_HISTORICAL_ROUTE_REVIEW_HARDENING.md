# Historical Route Review Hardening

Status: closed

## Scope

This increment hardens `apps/api/internal/historicalintelligence/historicalroute` as the validated historical analytics boundary over persisted Route Intelligence results.

## Accepted findings

- A complete airport-pair scope cannot classify valid partial or unavailable route results because those statuses deliberately lack a complete pair identity.
- Historical Route used the compatibility-oriented partial decoder instead of the complete Route Intelligence contract validator.
- Persistence metadata was not reconciled with the decoded payload before analytics.
- `StoredAt` affected source freshness but was absent from the result fingerprint.
- A bounded global route denominator cannot be reused as route-pair coverage evidence.
- Snapshot query evidence was not checked against the canonical analytical plan.
- Scoped source freshness and source names were not derived from the scoped validated evidence.
- Persisted route distance was trusted instead of being recomputed from validated endpoint coordinates.
- Confidence and distance accumulation lacked compensated floating-point summation.
- Active route semantics, route-confidence semantics, and zero-denominator evidence were under-documented.
- Latest-record selection accepted a record identifier as a substitute for missing trajectory identity.
- Metric calculation responsibilities were concentrated in one large switch.

## Corrected contracts

- Route status ratios are global metrics only. A route-pair scope is rejected by the central metric catalog and the Historical Route builder.
- Every selected payload passes `routecontract.Validate` before it can contribute to a metric.
- Persisted trajectory identity, status, confidence level, input fingerprint, as-of time, event window, validation warning count, storage time, and payload fingerprint are reconciled with the payload.
- Invalid JSON, unsupported schemas, invalid endpoint cardinality, endpoint-role errors, invalid airports, future evidence, provenance failures, and metadata mismatches fail closed with typed causes.
- A production snapshot query window must contain the canonical plan effective window and use the same analytical `AsOfTime`. A larger previous-plus-current materialization window remains valid.
- Incomplete global reads require the exact `RouteMatchedCount`. Incomplete route-pair reads are rejected because no pair-specific denominator exists.
- Fingerprint generation two binds every output-affecting selected record, including `StoredAt`, and binds limit state plus exact denominator evidence when the read is incomplete.
- Complete route distance is recomputed with a documented haversine great-circle policy over validated endpoint coordinates.
- Provenance source names are the sorted union of the route dataset and actual Route Contract sources. Latest source time is calculated only from scoped validated evidence.
- `active_routes` is explicitly a count of unique directional complete route pairs and uses the unit `route_pairs`. Sample count remains the number of validated route results supporting the distinct count.
- Route confidence is the compensated unweighted mean of validated result-level confidence. Evidence quantity and quality are already represented inside each Route Contract confidence score and are not counted twice.
- Status-ratio buckets with no observations retain value `0` together with sample count `0` and complete bucket coverage, so zero numerator and absent denominator remain distinguishable.
- Latest selection requires real trajectory identity and uses `AsOfTime`, `StoredAt`, then stable record identifier ordering.
- Metric calculation uses a calculator registry and focused evidence, selection, scope, coverage, fingerprint, provenance, and limitation helpers.

```text
ROUTE_PAIR_STATUS_RATIOS=GLOBAL_ONLY
ROUTE_CONTRACT_VALIDATION=ENFORCED
PERSISTENCE_METADATA_RECONCILIATION=ENFORCED
ROUTE_SCOPE_INCOMPLETE_COVERAGE=REJECTED
SNAPSHOT_PLAN_WINDOW_CONTAINMENT=ENFORCED
ROUTE_FINGERPRINT_STORED_AT=BOUND
ROUTE_DISTANCE_COORDINATE_RECOMPUTATION=ENFORCED
ROUTE_PROVENANCE_SCOPED=ENFORCED
ACTIVE_ROUTES_SEMANTICS=UNIQUE_DIRECTIONAL_ROUTE_PAIRS
HISTORICAL_ROUTE_ENGINEERING_REMEDIATION=IMPLEMENTED
```

## Rejected, stale, or deliberately retained findings

- The claim that `RouteLimitReached` is absent from the fingerprint was stale. The existing builder already bound both row-limit and byte-limit flags. The new fingerprint retains them and adds `StoredAt`.
- Production Historical Read already exposes exact matched-row counts. The legacy `selected/(selected+1)` fallback is not used by the hardened Historical Route builder; incomplete reads without an exact count are rejected.
- A complete zero-event bucket remains a valid Historical Contract state. Historical Route does not invent source freshness when no scoped route evidence exists; it fails closed with `ErrRouteSourceEvidenceUnavailable` instead.
- The heterogeneous `float64` transport remains intentional. The central metric catalog enforces exact integral count values through the exact IEEE-754 integer boundary and validates ratios on `[0,1]` with a fixed absolute tolerance.
- Confidence is not reweighted by evidence count. Route Contract confidence already incorporates evidence quantity, weight, and quality; weighting it again would double-count the same evidence.
- Generic summary `Total` remains descriptive. Ratio comparisons are bound by the metric catalog to `Summary.Average`, not to the temporal sum of ratios.
- Raw persistence JSON decoding already belongs to Historical Read rather than Historical Route. The strict validator is added at that boundary instead of reintroducing JSON parsing into the domain builder.
- Direct navigation through the adjacent Route Contract value object is not treated as a Law of Demeter defect.
- Similar helper shapes in Historical Airport and Historical Route do not by themselves justify a shared abstraction because their identities, filters, and coverage semantics differ.

## Permanent verification

`apps/api/tools/historicalroutereviewaudit` enforces the version-two builder boundary, global-only status-ratio scopes, strict Route Contract validation, metadata reconciliation, exact incomplete coverage, `StoredAt` fingerprint identity, scoped provenance, compensated arithmetic, focused calculators, regression tests, and this engineering remediation record in Backend Continuous Integration.

## Formal closure evidence

The Historical Route engineering remediation was committed and validated before
this administrative closure:

```text
ENGINEERING_BASELINE_COMMIT=9741c4fce04e2b2c06ee0236cf13b5c384f38ffd
ENGINEERING_REMEDIATION_COMMIT=513fa1efc7f3b81b895cdc5f881e294d80362e2e
ENGINEERING_GITHUB_ACTIONS_RUN=30334131538
Backend Quality=SUCCESS
Backend Quality Job=90195300495
PostgreSQL 16 Integration=SUCCESS
PostgreSQL 16 Integration Job=90195300516
Backend Race Safety=SUCCESS
Backend Race Safety Job=90195300546
Backend Container=SUCCESS
Backend Container Job=90195525282
```

The accepted findings are fully implemented, stale or rejected findings retain
their documented disposition, and no Historical Route review item remains open,
unclassified, or deferred.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
HISTORICAL_ROUTE_ENGINEERING_DEBT=CLOSED
HISTORICAL_ROUTE_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
HISTORICAL_ROUTE_REVIEW_STATUS=CLOSED
```
