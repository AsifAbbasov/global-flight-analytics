# Stage 20 — Airspace Intelligence Frontend Integration

Status: implementation in progress; exact-head validation pending

```text
STAGE_20_AIRSPACE_INTELLIGENCE_FRONTEND=IN_PROGRESS
STAGE_20_BASE_MAIN=84989c72e4cc8fddb79e484f2c7de2e6e597ba0f
STAGE_20_BACKEND_ANALYTICS_RECOMPUTATION=NONE
STAGE_20_NEW_BACKEND_ENDPOINT=NONE
STAGE_20_NEW_DATABASE_DATA=NONE
STAGE_20_NEW_PROVIDER=NONE
STAGE_20_WORLD_ANALYTICS=DISABLED
STAGE_20_AS_OF_SOURCE=LATEST_VALID_TRAFFIC_OBSERVED_AT
STAGE_20_HEATMAP=DEFERRED_UNVERIFIED_GEOMETRY
STAGE_20_OPERATIONAL_AIRSPACE_CLAIMS=NONE
STAGE_20_ADDITIONAL_COST=0_RUB
```

## 1. Product need

The repository already contains a production-backed Airspace Intelligence service and the public read-only endpoint:

```text
GET /api/v1/airspace/regions/{code}/analytics
```

Stage 11 completed the backend foundation for bounded temporal occupancy, sector-complexity analytics, regional pressure, confidence, limitations, explanations and provenance. The generated OpenAPI client also already exposes the operation.

Stage 13 subsequently integrated Route, Projection, Weather and Stability Intelligence into the frontend, but Airspace Intelligence remained backend-only. The result was a material product gap: one of the strongest existing analytical modules could be inspected through API evidence but was not visible in the normal product experience.

Stage 20 closes that presentation gap. It does not create a new analytical engine.

## 2. Selected architecture

```text
existing regional traffic query
        ↓
TrafficAircraft[].observed_at
        ↓
latest valid observed timestamp
        ↓
existing Airspace Intelligence HTTP endpoint
        ↓
validated frontend transport parser
        ↓
TanStack Query
        ↓
regional Airspace Intelligence workspace
        ↓
research-only evidence panel
```

The existing region selector remains authoritative for product region state. The Airspace Intelligence workspace is region-level rather than aircraft-level because the backend result is a regional analytical aggregate.

The workspace reuses the existing `useCurrentTraffic(selectedRegion.code)` TanStack Query key. This does not create a second raw network path; React Query shares the current traffic cache across observers.

## 3. Analytical timestamp decision

The Airspace endpoint requires an `as_of_time`. Stage 20 deliberately does not use browser wall-clock time as analytical evidence.

Instead, the pure frontend model scans the already loaded regional traffic snapshot and selects the latest valid `TrafficAircraft.observed_at` timestamp. Invalid timestamps are ignored. If no valid observed timestamp exists, Airspace Intelligence remains unavailable rather than substituting a synthetic time.

```text
AS_OF_TIME_SOURCE=OBSERVED_TRAFFIC_TIMESTAMP
BROWSER_NOW_AS_EVIDENCE=NO
INVALID_TIMESTAMP_FALLBACK=NONE
```

This preserves the evidence boundary and avoids creating a future-time request because of browser/server clock skew.

## 4. Bounded-region decision

The frontend owns a synthetic `world` region for global traffic display. The production Airspace Intelligence service resolves region codes through its backend region resolver and expects bounded analytical regions.

Stage 20 therefore treats `world` as a display-only scope and does not send it to the Airspace endpoint.

The UI explicitly explains:

- World remains a frontend-wide traffic view;
- a bounded backend region must be selected for Airspace Intelligence;
- the absence of world analytics is an evidence/contract boundary rather than an application failure.

No hidden region substitution is allowed.

## 5. Frontend result surface

The Stage 20 panel exposes backend-owned fields including:

- current and unique aircraft counts;
- aircraft observation count;
- temporal coverage;
- peak and mean aircraft per bucket;
- occupied-cell count;
- unknown-altitude count;
- mean and peak complexity scores;
- highest complexity level;
- occupancy trend;
- Airspace Pressure Index and peak pressure;
- elevated, high and indeterminate risk-context counts;
- analytical confidence and confidence reasons;
- exact analysis window;
- limitations and explanations;
- source names;
- latest observed provenance timestamp;
- input fingerprint;
- backend scope guard.

The browser validates the fields it renders but does not recompute backend complexity, pressure, risk or confidence semantics.

## 6. Evidence and safety boundary

Stage 20 preserves the Stage 11 research-only contract.

```text
AIRSPACE_ANALYTICS=RESEARCH_ONLY
OFFICIAL_SECTOR_CLAIM=NONE
ATC_SUPPORT_CLAIM=NONE
CONTROLLER_WORKLOAD_CLAIM=NONE
REGULATORY_SEPARATION_CLAIM=NONE
COLLISION_PREDICTION_CLAIM=NONE
SAFETY_CRITICAL_GUIDANCE=NONE
CALIBRATED_OPERATIONAL_PROBABILITY=NONE
```

The frontend states that occupancy, complexity, pressure and risk-context counts are analytical research evidence. They are not official sectors, certified separation monitoring, collision prediction, controller workload measurement or air traffic control guidance.

Confidence remains evidence support for the analytical result and is not represented as operational certainty or a calibrated probability.

## 7. Why a heatmap is not included

The backend payload contains occupancy cells with latitude, longitude and altitude indices. That is not sufficient by itself to justify a frontend polygon or heatmap unless the frontend has a verified, versioned mapping from each index to exact geographic cell bounds.

Stage 20 therefore does not render an occupancy heatmap, polygon layer or sector geometry.

```text
GRID_INDICES_AVAILABLE=YES
FRONTEND_CELL_BOUNDARY_CONTRACT=NOT_VERIFIED
HEATMAP_RENDERING=DEFERRED
SYNTHETIC_CELL_GEOMETRY=PROHIBITED
```

A later map increment may proceed only after proving the geometry contract from existing backend policy/region evidence or adding an explicit server-owned geometry contract. Visual attractiveness is not evidence.

## 8. Files introduced or changed

Product implementation:

```text
apps/web/types/airspace-intelligence.ts
apps/web/lib/api/airspace-intelligence.ts
apps/web/lib/queries/airspace-intelligence.ts
apps/web/lib/airspace/airspace-intelligence-model.ts
apps/web/components/traffic/airspace-intelligence-panel.tsx
apps/web/components/analytics/airspace-intelligence-workspace.tsx
apps/web/components/regional-traffic-experience.tsx
```

Regression protection:

```text
apps/web/tests/airspace-intelligence-model.test.mjs
apps/web/tests/stage-20-airspace-intelligence-frontend.test.mjs
apps/web/tsconfig.test.json
apps/web/e2e/tests/advanced-intelligence.spec.mjs
```

Documentation:

```text
docs/206_STAGE_20_AIRSPACE_INTELLIGENCE_FRONTEND_INTEGRATION.md
```

## 9. Regression strategy

The pure production model is compiled through `apps/web/tsconfig.test.json` and executed by Node tests from `.test-dist`.

The model tests protect:

- latest-valid-observation timestamp selection;
- invalid timestamp rejection;
- no synthetic timestamp when observations are unavailable;
- bounded-region eligibility;
- explicit `world` rejection.

The source-contract tests protect:

- regional product integration;
- use of the existing Airspace endpoint;
- `requestAPIData` transport boundary;
- no component/query-level raw `fetch` or axios path;
- observed-time semantics;
- bounded `world` behavior;
- research-only copy;
- rendered backend-owned metric families;
- absence of an invented heatmap or direct grid-index rendering.

The existing advanced-intelligence Playwright journey is extended to require visible Airspace Intelligence evidence and the backend research-only separation/ATC limitation.

## 10. Infrastructure and cost impact

```text
New paid provider             = NO
New backend endpoint          = NO
New PostgreSQL table          = NO
New migration                 = NO
New ingestion path            = NO
New server                    = NO
New cache                     = NO
New persistent storage        = NO
New runtime dependency        = NO
New external aviation call    = NO
Frontend raw second data path = NO
Additional cost               = 0 RUB
```

The only new product request is to an existing GFA read-only backend endpoint for a capability already present in the production API surface.

## 11. Acceptance criteria

Stage 20 product implementation can be proposed for merge only when one exact Pull Request head demonstrates:

```text
Frontend CI          = SUCCESS
Backend CI           = SUCCESS
CodeQL               = SUCCESS
API Load Baseline    = SUCCESS
Playwright E2E       = SUCCESS
Vercel               = SUCCESS
```

The exact merge-candidate SHA must be revalidated after any remediation or documentation change. A successful earlier head cannot be reused as evidence for a later head.

## 12. Residual limitations

Even after product merge:

- Airspace Intelligence remains available only for bounded backend-resolved regions;
- global World traffic display does not imply world-scale Airspace Intelligence;
- analytical results depend on persisted open-data observations available for the requested window;
- gaps, missing altitude and provider limitations remain visible evidence constraints;
- region analytics do not create official airspace sectors;
- no continuous occupancy field is reconstructed between discrete evidence buckets;
- no grid heatmap is rendered until geographic cell bounds are independently verified;
- no operational separation, controller workload or collision-avoidance claim is introduced.

## 13. Current status

Implementation is present on the Stage 20 feature branch. Continuous Integration and browser validation on the final exact head remain pending at the time this engineering-history snapshot is written.

A later post-merge closure document must append the actual merge and post-merge evidence without rewriting this pre-merge history.
