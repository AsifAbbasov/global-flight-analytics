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
<!-- ANALYTICAL-CORE-REVIEW-CLOSURE:SERVER-OWNED-QUALITY -->

## 17. Post-closure resolution

Document 102 completes the final Analytical Core closure audit.

```text
ANALYTICAL_CORE_REVIEW_STATUS=CLOSED
Open server-owned quality findings: 0
```

The server-owned Coverage Score and Data Freshness contracts remain protected by
the permanent Analytical Core final audit.

---

## Canonical remediation history

### GFA-DATA-110 / AC-08 — production Coverage Score and Data Freshness trusted caller-owned snapshot evidence

1. **Finding / symptom.** Production quality-metric endpoints accepted caller-provided `observed_samples`, `expected_samples`, `observed_at` and `max_age_seconds`, so the caller supplied the evidence used to calculate Coverage Score and Data Freshness.
2. **Root cause.** Internal snapshot calculator value objects leaked through the public HTTP boundary and were treated as production data inputs rather than an internal computation seam.
3. **Failure scenario.** A caller submits arbitrary sample counts, timestamps or stale thresholds and receives a valid-looking metric with server presentation/provenance even though the server did not independently observe those values.
4. **Impact.** Quality metrics can be manipulated by request parameters and cannot be interpreted as measurements of retained GFA aviation data.
5. **Severity rationale.** **P1 retrospective.** This is a direct trust and analytical correctness failure on normal public endpoints, not merely a confidence-label issue.
6. **Existing guarantees violated.** Production analytics must derive evidence from server-owned retained data; public requests may define scope but not fabricate measurement snapshots; confidence/provenance must describe the real evidence path.
7. **Considered solutions.** Keep caller snapshots but mark low confidence; expose separate calculator/debug endpoints; derive production snapshots from retained trajectories; remove the metrics from production.
8. **Chosen remediation.** Public quality routes accept only `window_minutes` and `region`, reject former snapshot parameters, use a server-fixed bounded query limit, derive Coverage Score from covered observation intervals and Data Freshness from the latest usable retained observation, and use the server-owned stale threshold.
9. **Why this solution was selected.** It preserves useful deterministic formulas while moving evidence ownership to the service that owns retained trajectories.
10. **Rejected alternatives.** Low-confidence labeling still lets a caller manufacture the measured input; public calculator routes would need explicit non-production semantics; removing the metrics is unnecessary once server evidence is available.
11. **Trade-offs.** Results now depend on retained trajectory coverage and can legitimately be zero/limited when evidence is absent. Callers lose the ability to customize the stale threshold or truncate the evidence query.
12. **Regression tests / protection.** Tests cover covered-interval counting, duplicate suppression, EndTime fallback, empty evidence, server-fixed limit/stale threshold, server-derived inputs and rejection of every legacy snapshot parameter. Frontend tests/contracts use scope-only requests. The final Analytical Core audit protects server-owned quality evidence.
13. **Adversarial review findings.** A client-controlled `limit` would still let callers manipulate Coverage Score by truncating retained evidence, so the production limit must also be server-owned. Empty evidence must not invent an observation timestamp.
14. **Remediation iterations.** Document 99 first reduced confidence for caller snapshots as an interim honesty fix; Document 101 completes `AC-08` by removing caller snapshot ownership from production and reconciling frontend consumers/query keys.
15. **Residual risks and limitations.** Metrics reflect retained open-data observations, not guaranteed complete real-world traffic. Sampling interval and stale threshold are engineering policies documented by the data-quality layer.
16. **Operational or deployment consequences.** Quality metrics now require server access to retained trajectories and configured region scope; no new infrastructure is introduced.
17. **Exact evidence.** Historical implementation commit `e48cb27655326fc6cc41d176a50120cdbf1ced6e` (`fix: derive production quality metrics from server data`). Original review ID: `AC-08`. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-DATA-110=CLOSED`.
19. **Prevention / future guard.** Public production metrics may accept analytical scope, but measurement evidence and quality thresholds must remain server-owned unless an endpoint is explicitly classified as a calculator/simulation surface.

### GFA-CONTRACT-111 — rejected analytical parameters could produce a response and still continue query execution

1. **Finding / symptom.** The legacy-parameter rejection helper originally treated the response writer's returned `error` as the only signal that a rejection had occurred.
2. **Root cause.** Fiber can successfully serialize `400 Bad Request` and return `nil`; control-flow ownership conflated `response handled` with `response write failed`.
3. **Failure scenario.** A prohibited quality snapshot parameter is detected, a 400 response is written successfully, the helper returns `nil`, and handler logic continues into the server-owned analytical query path after the request was already rejected.
4. **Impact.** Rejected requests can perform unintended backend work and risk double-response/control-flow inconsistencies; tests that inspect only HTTP status may miss the extra query execution.
5. **Severity rationale.** **P2 retrospective.** The defect violates HTTP handler control flow and resource ownership, but the affected path is read-only analytical querying rather than mutation.
6. **Existing guarantees violated.** Once a handler has produced a terminal rejection response, no downstream analytical query may execute; response-write success and request-handled state are independent concepts.
7. **Considered solutions.** Return only `error`; return a sentinel non-nil error after every successful 400 write; return `(handled, error)`; panic/abort Fiber context.
8. **Chosen remediation.** The rejection helper returns two values: `handled` and `error`. `loadProductionQualityInput` exits whenever `handled` is true, including successful response writes with `error == nil`.
9. **Why this solution was selected.** It models the two independent states explicitly without abusing error values for normal control flow.
10. **Rejected alternatives.** Synthetic errors blur success/failure semantics; error-only handling recreates the bug; panic/abort is unnecessary for ordinary validation.
11. **Trade-offs.** Call sites carry one additional boolean, but terminal-response ownership becomes unambiguous.
12. **Regression tests / protection.** Each prohibited parameter is tested with an isolated handler and query stub; the stub must remain untouched after a rejected request.
13. **Adversarial review findings.** HTTP status alone is insufficient regression evidence because the same 400 can be returned even if downstream work incorrectly runs; tests must assert the repository/query seam was not invoked.
14. **Remediation iterations.** The bug was discovered while converting the production endpoints to server-owned evidence and was fixed before final Analytical Core closure.
15. **Residual risks and limitations.** Other handlers using unrelated response helpers must independently preserve terminal control flow; this finding owns the Analytical Core quality-parameter rejection seam.
16. **Operational or deployment consequences.** Rejected legacy requests perform less unnecessary backend work and terminate deterministically.
17. **Exact evidence.** Historical implementation commit `e48cb27655326fc6cc41d176a50120cdbf1ced6e`; Document 101 Section 13 records the failure mode and regression contract. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-CONTRACT-111=CLOSED`.
19. **Prevention / future guard.** Handler helpers that can emit terminal responses must expose handled-state independently from transport/write errors, and tests must assert no downstream operation after rejection.

### GFA-TEST-112 — temporary frontend verification used a non-hermetic external `node_modules` symlink

1. **Finding / symptom.** The detached temporary Git worktree could reuse frontend dependencies through an external `node_modules` symbolic link.
2. **Root cause.** Verification optimized dependency reuse without accounting for Next.js Turbopack's requirement that package resolution remain inside the temporary project root.
3. **Failure scenario.** The temporary production build resolves `node_modules` outside the worktree filesystem root and Turbopack rejects the layout, causing the preflight to fail for tooling topology rather than source correctness.
4. **Impact.** The remediation installer cannot provide reliable hermetic frontend verification and may block valid changes or encourage skipping the temporary build gate.
5. **Severity rationale.** **P2 retrospective.** This is a verification-integrity defect in a required release/remediation gate; it does not affect runtime production behavior directly.
6. **Existing guarantees violated.** Temporary worktree verification must reproduce the project inside one self-contained root and must not depend on path topology outside that root.
7. **Considered solutions.** Keep the symlink and disable Turbopack; copy all `node_modules`; perform an offline frozen-lockfile install inside the worktree; skip the temporary frontend build.
8. **Chosen remediation.** The installer runs `pnpm install --offline --frozen-lockfile` inside the worktree using the local pnpm store, then asserts `apps/web/node_modules` exists and is not a symlink before the build.
9. **Why this solution was selected.** It reuses cached package content without network access while preserving an actual workspace dependency tree under the temporary repository root.
10. **Rejected alternatives.** Disabling the real build path weakens evidence; copying dependencies is expensive and brittle; skipping the build defeats the cross-stack preflight.
11. **Trade-offs.** Temporary verification performs a package install step and depends on the local pnpm content-addressable store being populated.
12. **Regression tests / protection.** Installer assertions reject symlinked `node_modules`; temporary typecheck/lint/build run after the offline frozen install.
13. **Adversarial review findings.** A lockfile-preserving install is required so fixing hermeticity cannot silently update dependencies; `.next` output must be cleaned before builds to avoid stale artifact evidence.
14. **Remediation iterations.** External dependency linking was replaced with offline local-store installation; the same installer later added working-tree dependency self-repair as a separate guard.
15. **Residual risks and limitations.** Offline verification requires required packages to exist in the local pnpm store; a clean environment may need a prior dependency bootstrap outside this specific offline step.
16. **Operational or deployment consequences.** None for runtime. Developer/remediation verification becomes deterministic with respect to worktree path ownership.
17. **Exact evidence.** Historical implementation commit `e48cb27655326fc6cc41d176a50120cdbf1ced6e`; Document 101 Section 15. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-TEST-112=CLOSED`.
19. **Prevention / future guard.** Temporary frontend verification may reuse package stores but may not symlink dependency roots outside the worktree; frozen-lockfile and non-symlink assertions remain mandatory.

### GFA-TEST-113 — analytical remediation verification assumed the existing working-tree dependency install was healthy

1. **Finding / symptom.** The installer initially assumed an already-present frontend dependency tree was complete and usable before baseline/final typecheck, lint and build.
2. **Root cause.** Verification treated local installation state as trusted environmental evidence instead of reconstructing dependencies from the lockfile before critical gates.
3. **Failure scenario.** `node_modules` is partial, stale or damaged; verification fails for missing TypeScript/Next/ESLint binaries or, worse, runs against an inconsistent local tree rather than the lockfile-defined workspace.
4. **Impact.** Frontend acceptance becomes machine-state dependent and remediation rollback/final verification can be unreliable.
5. **Severity rationale.** **P2 retrospective.** Required cross-stack verification was not reproducible from repository state alone, weakening confidence in closure evidence.
6. **Existing guarantees violated.** Acceptance gates must derive dependencies from the committed lockfile, verify required executables and clean generated build state before judging source changes.
7. **Considered solutions.** Trust existing `node_modules`; fail with manual repair instructions; run an offline frozen workspace install before baseline and final checks; always perform a network install.
8. **Chosen remediation.** The installer performs `pnpm install --offline --frozen-lockfile`, verifies TypeScript/Next/ESLint executables, runs a baseline typecheck, repeats the offline install before final frontend gates, deletes `.next` before builds and attempts dependency repair during rollback.
9. **Why this solution was selected.** It makes verification lockfile-driven and repeatable without mutating dependency versions or requiring network availability.
10. **Rejected alternatives.** Trusting local state is non-hermetic; manual repair is not executable evidence; network installs add unnecessary availability and resolution variability.
11. **Trade-offs.** Verification does more local package-manager work and requires the pnpm store to contain locked packages.
12. **Regression tests / protection.** Installer executable checks, frozen-lockfile installs and clean builds are part of the documented acceptance workflow.
13. **Adversarial review findings.** Baseline and post-change verification both matter: repairing only after patch application cannot prove the starting tree was healthy; cleaning `.next` prevents stale build output from satisfying the gate.
14. **Remediation iterations.** This guard followed the temporary-worktree hermeticity repair and generalized dependency reconstruction to the ordinary working tree and rollback path.
15. **Residual risks and limitations.** The local pnpm store remains an environmental prerequisite for fully offline execution; CI provides an independent clean-environment evidence path.
16. **Operational or deployment consequences.** None for deployed services; developer acceptance scripts become less sensitive to prior workstation state.
17. **Exact evidence.** Historical implementation commit `e48cb27655326fc6cc41d176a50120cdbf1ced6e`; Document 101 Section 16. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-TEST-113=CLOSED`.
19. **Prevention / future guard.** Required local verification must reconstruct frontend dependencies from the frozen lockfile and clean generated artifacts rather than trusting existing installation/build state.