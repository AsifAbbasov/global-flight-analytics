# Document 101 — Server-Owned Quality Metrics

Status: Implemented Engineering Increment v1.0
Project: Global Flight Analytics
Baseline: `b8ccbf590ef3b9ffc221d72e0274e1d78da6c650`

## 1. Purpose

This increment removes caller-owned snapshot evidence from the production
Coverage Score and Data Freshness HTTP routes.

The public routes now derive their values exclusively from retained trajectory
data loaded by the server.

## 2. Public request boundary

The production routes accept analytical scope only:

```text
window_minutes
region
```

The following former snapshot parameters are explicitly rejected:

```text
observed_samples
expected_samples
observed_at
max_age_seconds
limit
```

`limit` is fixed to the server-owned bounded maximum of five thousand rows so a
caller cannot manipulate sampling coverage by truncating the evidence query.

## 3. Coverage Score evidence

Coverage Score uses the complete normalized query window.

The server:

```text
loads retained trajectories;
collects point observation times inside the query window;
uses trajectory EndTime only when a trajectory has no usable retained points;
maps observations into ten-second expected intervals;
counts unique covered intervals;
divides covered intervals by total expected intervals.
```

The ten-second expected interval is the existing
`dataqualityintegration.DefaultExpectedObservationInterval`.

Raw duplicate points inside one interval do not inflate Coverage Score.

## 4. Data Freshness evidence

Data Freshness uses the latest usable retained observation from the same
server-owned query window.

The maximum age is fixed to the existing
`dataqualityintegration.DefaultStaleAfter`, currently five minutes.

No client timestamp or client maximum-age threshold participates in production
calculation.

## 5. Empty evidence

An empty server query does not fabricate an observation timestamp.

```text
Coverage Score = 0
Data Freshness = 0
latest observation time = unavailable
confidence = low
limitation = no_trajectory_observations
```

## 6. Provenance

Production quality results include:

```text
provider trajectory sources;
server_trajectory_query derived source;
normalized observed-from and observed-to window;
server retrieval time;
existing open-data limitations;
server_owned_production_snapshot limitation;
data-quality report when construction is possible.
```

## 7. Internal calculator seam

`CoverageScoreRequest` and `DataFreshnessRequest` continue to use the internal
snapshot value object as a computation contract.

That type is not populated from public snapshot query parameters in production.
The HTTP handlers are the production ownership boundary.

## 8. Confidence

The old caller-supplied confidence factor is removed.

Server-owned evidence receives high calculation confidence when usable
observations exist. Absence of usable observations receives low confidence and
an explicit limitation.

Confidence describes trust in the calculation and evidence path; it does not
change a low metric value into a high value.

## 9. Permanent regression coverage

The increment adds tests for:

```text
covered interval counting;
duplicate observation suppression;
EndTime fallback;
empty evidence preservation;
server-fixed result limit;
server-derived Coverage Score inputs;
server-derived Data Freshness timestamp;
server-fixed stale threshold;
legacy snapshot parameter rejection;
server-owned confidence behavior;
zero-observation Data Freshness publication.
```

## 10. Verification

Before changing `main`, the installer validates every patch anchor, scans the
frontend for obsolete production parameter usage, applies the patch in a
detached temporary Git worktree, and runs:

```text
complete backend compilation;
targeted quality-metric tests;
complete backend tests.
```

The working tree then runs targeted compilation, targeted tests, race tests,
the complete backend suite, Go vet, architecture audits, static ownership
contracts, documentation checks, and whitespace validation.

## 11. Remaining Analytical Core review scope

Only the final Analytical Core closure audit remains.

That audit must classify every original finding as fixed, rejected,
not applicable, or explicitly deferred, and bind the decision to committed
tests, documents, and successful Continuous Integration evidence.

## 12. Frontend contract reconciliation

The Next.js analytical API client is updated in the same increment.

```text
Traffic Density requires regionCode and never sends area_square_kilometers.
Airport Activity sends airport_icao and optional radius_kilometers.
Coverage Score sends only window_minutes and region.
Data Freshness sends only window_minutes and region.
All frontend Metric IDs use the canonical traffic.* namespace.
```

Generated `.next` output and unrelated domain fields named `observed_at` are not
used as source-contract evidence. Verification is restricted to the committed
analytical API client and analytical parameter types.

Frontend typecheck, lint, and production build run both in the detached
temporary worktree and after the patch is applied to the working tree.

## 13. Rejected-response control flow

Fiber response writers may successfully serialize a `400 Bad Request` response
and return a nil transport error. A nil return does not mean the request should
continue through the analytical query path.

The parameter-rejection helper therefore returns two independent values:

```text
handled — whether an HTTP response has already been produced;
error — whether writing that response failed.
```

`loadProductionQualityInput` returns immediately whenever `handled` is true,
including the normal case where the response was written successfully and the
transport error is nil.

Each prohibited query parameter is tested with an isolated handler and query
stub. The query stub must remain untouched after every rejected request.

## 14. Frontend consumers

The application-level analytical consumers now use the same production scope as
the HTTP client.

`AnalyticsOverview` sends one server-owned quality scope:

```text
windowMinutes = fifteen minutes;
regionCode = the selected configured region.
```

Coverage Score is no longer derived from Active Aircraft scope counts.
Data Freshness is no longer derived from response source timestamps.

React Query keys now represent only the actual public request contract:

```text
Traffic Density: window, limit, region;
Airport Activity: window, limit, region, airport ICAO, radius;
Coverage Score: window, region;
Data Freshness: window, region.
```

There are no nullable quality-metric parameter states, no client-side snapshot
builders, and no dependency on a successful Active Aircraft response before
requesting server-owned quality metrics.

## 15. Hermetic temporary frontend verification

The temporary Git worktree must remain inside one filesystem root from the
perspective of Next.js Turbopack.

An external `node_modules` symbolic link is prohibited because Turbopack rejects
package resolution paths that escape the temporary project root.

The installer now prepares the worktree through:

```text
pnpm install --offline --frozen-lockfile
```

This operation uses the local pnpm content-addressable store, does not modify
the lockfile, and creates workspace package links inside the temporary
repository.

Before the temporary production build, the installer asserts that
`apps/web/node_modules` exists and is not a symbolic link.

The obsolete `formatPositiveFiniteNumber` helper is also removed after Traffic
Density ownership moves completely to configured server regions.

## 16. Working-tree dependency self-repair

The installer no longer assumes that an existing frontend dependency tree is healthy.
Before patch application it performs an offline frozen-lockfile workspace install,
verifies the TypeScript, Next.js and ESLint executables, and runs a baseline typecheck.

After all backend gates it repeats the offline workspace install before the final
frontend typecheck, lint and clean production build. Both builds delete `.next` first.

Rollback restores the Git baseline, removes the increment files and incomplete `.next`,
and attempts another offline dependency repair.
