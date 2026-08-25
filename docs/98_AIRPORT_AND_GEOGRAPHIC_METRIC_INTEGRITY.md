# Document 98 — Airport and Geographic Metric Integrity

Status: Implemented Engineering Increment v1.0
Project: Global Flight Analytics
Baseline: `c5fd1f32273af9215df9d83d1d40c227d3740646`

## 1. Purpose

This increment closes two semantic findings from the original Analytical Core
Foundation review:

```text
Airport Activity must be owned by a concrete airport and classified by the server;
Traffic Density must use one server-owned geographic scope for both contributors and area.
```

## 2. Airport Activity contract

The production endpoint now requires:

```text
airport_icao
optional radius_kilometers
optional window_minutes
optional limit
```

The server loads the airport from PostgreSQL, derives a bounded geographic
query, loads recent trajectories, applies eligibility, removes duplicate
eligible trajectories, and classifies movement from trajectory crossings of
the airport geofence.

The client no longer submits separate arrival and departure trajectory lists.
Unrelated and ambiguous trajectories are excluded with explicit limitations.

## 3. Traffic Density contract

Traffic Density now requires a configured region. The same region bounds own
both the contributor query and the calculated area.

The client-provided `area_square_kilometers` parameter is rejected.

## 4. Verification

The installer executes compile-only checks, targeted tests, race tests, the
complete backend test suite, Go vet, all existing architecture audits, static
contract checks, and whitespace validation.

## 5. Remaining Analytical Core review scope

```text
server-owned production Coverage Score and Data Freshness;
strict analytical provenance and safe public failure messages;
reference-time and UUID canonicalization;
obsolete analytical foundation classification;
metric identifier consolidation.
```

---

## Canonical remediation history

### GFA-DATA-102 / AC-03 — Airport Activity was not owned by a concrete airport

1. **Finding / symptom.** The public Airport Activity calculation accepted caller-separated arrival/departure trajectory collections instead of deriving activity for one concrete airport from server-owned evidence.
2. **Root cause.** The metric contract represented preclassified movement lists rather than the airport, geofence and retained trajectories from which movement classification should be derived.
3. **Failure scenario.** A caller supplies unrelated, incomplete or incorrectly classified trajectories and receives an Airport Activity value that appears to describe an airport even though the server never proved airport ownership.
4. **Impact.** Airport activity can be semantically wrong, non-reproducible from retained data and vulnerable to caller-controlled classification bias.
5. **Severity rationale.** **P1 retrospective.** The defect affects the meaning of a published aviation metric under the normal public request path.
6. **Existing guarantees violated.** Airport Intelligence metrics must be owned by a concrete airport; the server must own geographic selection and movement classification; ambiguous evidence must be excluded or disclosed rather than accepted as caller truth.
7. **Considered solutions.** Continue accepting preclassified lists with stricter validation; require only airport coordinates; require `airport_icao` and derive the complete query/classification server-side; move classification to the frontend.
8. **Chosen remediation.** The endpoint requires `airport_icao`, loads the airport from PostgreSQL, derives a bounded geofence query, loads recent trajectories, applies analytical eligibility and eligible deduplication, then classifies movement from trajectory/geofence crossings.
9. **Why this solution was selected.** It assigns one authoritative domain owner to airport identity, geography and movement classification while keeping the public request scoped to intent rather than fabricated evidence.
10. **Rejected alternatives.** Caller-provided classifications remain unverifiable; coordinates alone lose canonical airport identity; frontend classification duplicates domain policy and cannot own retained server evidence.
11. **Trade-offs.** Airport Activity now requires database-backed airport metadata and bounded trajectory reads; ambiguous trajectories are excluded and may reduce counts compared with permissive caller input.
12. **Regression tests / protection.** HTTP and service tests require airport-owned queries, server classification, exclusion of unrelated/ambiguous movement and the accepted radius/window boundaries. The Analytical Core final audit preserves the airport-owned contract.
13. **Adversarial review findings.** The airport identifier alone is insufficient if the server later accepts caller-provided movement sets again; geographic radius and query bounds must be derived into the same evidence path used for classification.
14. **Remediation iterations.** The public request was simplified from evidence submission to airport scope; movement classification and query construction moved into server-owned analytical composition.
15. **Residual risks and limitations.** Geofence crossing is an analytical approximation and does not claim operational arrival/departure truth. Sparse trajectories may remain ambiguous and are disclosed as limitations.
16. **Operational or deployment consequences.** Airport Activity depends on PostgreSQL airport metadata and retained trajectories; no new infrastructure is introduced.
17. **Exact evidence.** Historical implementation commit `0ae85ccbff7584a993030c0adcdee3290dd4b7bd` (`fix: enforce airport and geographic metric integrity`). Original review ID: `AC-03`. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-DATA-102=CLOSED`.
19. **Prevention / future guard.** Public domain metrics may accept scope and filters, but must not accept caller-preclassified evidence when the server has the canonical data needed to derive the classification.

### GFA-DATA-103 / AC-04 — Traffic Density numerator and denominator used unrelated geographic scopes

1. **Finding / symptom.** Traffic Density could calculate a trajectory-derived active-aircraft numerator while accepting an independent caller-provided `area_square_kilometers` denominator.
2. **Root cause.** Contributor selection and area calculation had separate ownership boundaries, so nothing proved they described the same physical region.
3. **Failure scenario.** The server counts aircraft from one trajectory query while a caller supplies the area of another region or an arbitrary small/large number, producing a mathematically valid but geographically meaningless density.
4. **Impact.** Published density values can be arbitrarily inflated or diluted and cannot be compared across requests as one consistent geographic metric.
5. **Severity rationale.** **P1 retrospective.** This is a direct semantic correctness defect in a primary analytical metric and can occur through the normal request contract without malformed data.
6. **Existing guarantees violated.** Traffic Density numerator and denominator must describe one server-owned region; public requests may select a configured region but may not independently define the denominator.
7. **Considered solutions.** Validate caller area against a broad range; echo caller area as a limitation; derive area from the same configured region used for the trajectory query; eliminate geographic density until polygon support exists.
8. **Chosen remediation.** Traffic Density requires a configured region; the server uses the same region bounds for contributor loading and area calculation, and rejects the legacy `area_square_kilometers` parameter.
9. **Why this solution was selected.** One region object provides a simple, reproducible ownership boundary without adding GIS infrastructure or trusting arbitrary client arithmetic.
10. **Rejected alternatives.** Range checks cannot prove geographic correspondence; warnings do not repair wrong arithmetic; removing the metric is unnecessary once the existing configured-region model can own both sides.
11. **Trade-offs.** Density is limited to configured server regions rather than arbitrary caller-defined areas. That constraint is preferable to a metric with unprovable geography.
12. **Regression tests / protection.** Tests require region-owned trajectory queries and area, reject client area input and verify the frontend sends region scope rather than `area_square_kilometers`. The final Analytical Core audit checks the same ownership rule.
13. **Adversarial review findings.** Merely deriving the denominator server-side is insufficient if the trajectory query is not constrained by the same region; both query bounds and area must originate from one normalized scope object.
14. **Remediation iterations.** Backend ownership moved first; later frontend reconciliation in Document 101 removed obsolete area formatting/request behavior and aligned React Query keys with the server contract.
15. **Residual risks and limitations.** Region area is based on the project's configured geographic representation and is not an airspace-capacity or ATC-sector metric.
16. **Operational or deployment consequences.** Deployments must define supported analytical regions; callers can no longer request density for arbitrary areas without adding a new server-owned region contract.
17. **Exact evidence.** Historical implementation commit `0ae85ccbff7584a993030c0adcdee3290dd4b7bd`. Original review ID: `AC-04`. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-DATA-103=CLOSED`.
19. **Prevention / future guard.** Any future density scope—polygon, country, airspace or custom region—must derive both contributors and physical area from the same canonical server-owned geometry and prove that correspondence in tests.