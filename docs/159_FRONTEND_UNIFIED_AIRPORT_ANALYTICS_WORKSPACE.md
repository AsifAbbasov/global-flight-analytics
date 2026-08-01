# Frontend Unified Airport Analytics Workspace

## Status

Implemented against exact baseline commit:

`dac83d14a6f54590089f69675cf6e24083bf2bc5`

Formal closure still requires exact-commit Continuous Integration evidence after the
increment is committed and pushed.

## Product objective

This increment converts the existing production Airport Intelligence API into one
coherent frontend research workflow instead of splitting ranking, airport passport,
history and trends into separate small increments.

The workspace is reachable from the application shell through the stable
`#airport-intelligence` anchor and is rendered between the regional overview and the
live traffic workspace.

## Included capabilities

- global production airport ranking;
- seven, thirty and ninety completed-day analytical windows;
- search by ICAO, IATA, airport name, city or country;
- deterministic sorting by published rank, activity, confidence, movements or routes;
- explicit airport selection and clearing;
- digital airport passport with location and elevation semantics;
- operating activity and evidence-quality metrics;
- ranking-component presentation without inventing a browser-side composite score;
- completed-day movement history with a bounded fourteen-window chart and table;
- trend direction, changes, peak and baseline comparison;
- continuity and gap evidence;
- merged and de-duplicated API limitations;
- independent loading, empty, insufficient-history and error states;
- explicit retry actions;
- a shared refresh action for ranking and selected-airport records.

## Production API contracts used

- `GET /api/v1/airports/intelligence/ranking`
- `GET /api/v1/airports/:icao/intelligence/overview`
- `GET /api/v1/airports/:icao/intelligence/history`
- `GET /api/v1/airports/:icao/intelligence/trends`

The frontend sends the published `days` and `limit` query parameters and validates
all returned structures at the API boundary before rendering them.

## Evidence boundaries

The ranking is presented as global because the current Airport Intelligence API does
not publish a region-filter parameter. The frontend does not infer regional membership
from city, country or coordinates.

Completed-day windows are historical analytical summaries. They are not live airport
capacity, operational status, timetable, safety guidance or authoritative movement
counts.

The browser does not recalculate the backend ranking model. Activity, confidence,
coverage, freshness and component values remain independently visible.

## Deterministic model

`airport-intelligence-workspace-model.ts` owns pure presentation logic:

- ICAO normalization;
- search matching;
- stable ranking sorts;
- bounded visible result counts;
- chronological history selection;
- visible-peak chart normalization;
- trend direction normalization;
- continuity and gap interpretation;
- limitation de-duplication and ordering.

## Regression coverage

Eight dependency-free tests cover:

1. ICAO normalization;
2. search across identity and location fields;
3. deterministic metric sorting;
4. ranking display limits and counts;
5. chronological history and visible-peak normalization;
6. empty history behavior;
7. trend direction, unavailable percentages and gaps;
8. deterministic limitation merging.

## Files

Modified:

- `apps/web/components/product/application-shell.tsx`
- `apps/web/components/regional-traffic-experience.tsx`
- `apps/web/tsconfig.test.json`
- `docs/DOCUMENT_INDEX.md`

Added:

- `apps/web/components/analytics/unified-airport-analytics-workspace.tsx`
- `apps/web/lib/api/airport-intelligence.ts`
- `apps/web/lib/queries/airport-intelligence.ts`
- `apps/web/lib/analytics/airport-intelligence-workspace-model.ts`
- `apps/web/types/airport-intelligence.ts`
- `apps/web/tests/airport-intelligence-workspace-model.test.mjs`
- `docs/159_FRONTEND_UNIFIED_AIRPORT_ANALYTICS_WORKSPACE.md`

## Verification contract

Installation is accepted only after:

- exact clean baseline validation;
- Git 2.15 compatibility scan;
- forced rollback verification;
- temporary worktree application;
- exact changed-file manifest verification;
- dependency policy checks;
- all frontend tests;
- ESLint;
- TypeScript type checking;
- production Next.js build;
- `git diff --check`;
- the same checks on the real repository.
