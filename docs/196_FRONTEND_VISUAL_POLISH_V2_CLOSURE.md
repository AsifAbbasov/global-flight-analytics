# Document 196 — Frontend Visual Polish V2 Closure

Status: CLOSED — exact-head CI verified and PR #102 merged
Date: 2026-08-23
Scope: `FRONTEND_VISUAL_POLISH_V2`

## Purpose

This document records the post-closure visual-polish increment performed after the original
Frontend Product Closure. The supplied Flightradar24 material was used only as a UX and
information-hierarchy reference. No proprietary Flightradar24 code, assets, imagery,
branding, commercial data, or paid service dependency was copied into GFA.

The governing product rule is unchanged:

```text
SUPPORTED_OPEN_OR_PROJECT_DATA   render when present and valid
INFERRED_DATA                    label as inference with confidence/limitations
UNSUPPORTED_OR_UNAVAILABLE_DATA  omit
FABRICATED_PLACEHOLDERS          prohibited
NEW_PAID_FRONTEND_DEPENDENCY     prohibited
```

## Visual changes closed by PR #102

The merged increment:

- promotes the live two-dimensional MapLibre tracker ahead of the analytical globe and
  secondary research cards;
- enlarges the live map into a viewport-aware primary workspace;
- introduces a narrower, scrollable aircraft/intelligence rail beside the desktop map;
- keeps the mobile map ahead of status, export, globe, and secondary analytical surfaces;
- strengthens selected-aircraft map chrome and marker emphasis;
- replaces the loose aircraft popup with a compact telemetry hierarchy;
- prioritizes supported altitude, speed, heading, and state evidence in the selected-aircraft
  panel;
- compacts Aircraft Explorer search, filtering, and result density;
- preserves Trail and Projection as separately controlled evidence layers;
- renders origin/destination as **Probable route** only when GFA inference evidence exists,
  together with confidence and limitations;
- preserves route-confidence and track-quality accessibility semantics;
- removes per-aircraft `Unknown` / `Unavailable` filler rows where evidence is absent;
- keeps keyboard, reduced-motion, forced-colors, reduced-transparency, coarse-pointer, and
  no-horizontal-overflow contracts intact.

## Zero-budget / data-availability boundary

The visual reference exposed many useful Flightradar24 capabilities that GFA cannot
truthfully reproduce from the current zero-cost/open-data contracts. They were therefore
not added merely to increase visual density.

Not introduced by this phase:

- authoritative airline schedules or filed flight plans;
- authoritative ETA, ATD, STA, gate, terminal, or baggage data;
- squawk when not present in the GFA contract;
- vertical speed, IAS, TAS, or Mach when not present in the GFA contract;
- FIR/UIR commercial airspace layers;
- aircraft MSN, age, commercial photo database, or livery library;
- proprietary satellite/hybrid map layers;
- premium weather products;
- proprietary worldwide receiver coverage;
- paid map, aviation, image, or tracking APIs.

The omission is intentional product truth, not an incomplete UI placeholder policy.

## Exact-head verification evidence

PR #102 final exact head:

```text
PR=102
PR_HEAD=f1d2ba16cec215b217955fbcd049dd8b728562c4
FRONTEND_CI_RUN=32646381902
FRONTEND_CI=SUCCESS
BACKEND_CI_RUN=32646381901
BACKEND_CI=SUCCESS
CODEQL_RUN=32646381961
CODEQL=SUCCESS
API_LOAD_BASELINE_RUN=32646382037
API_LOAD_BASELINE=SUCCESS
PLAYWRIGHT_E2E_RUN=32646381896
PLAYWRIGHT_E2E=SUCCESS
PLAYWRIGHT_BROWSER_SCENARIOS=20
PLAYWRIGHT_BROWSER_RESULT=20_PASSED
PLAYWRIGHT_FLAKY_RETRIES=0
```

The final Chromium run completed all twenty browser journeys on the exact PR head. The
run included desktop and mobile structural visual-regression scenarios and uploaded the
retained screenshot artifact `playwright-e2e-32646381896-1`.

PR #102 was then squash-merged using an expected-head guard:

```text
MERGE_SHA=ce7ee7ab0b95655fff7a1b546e277a4e1c0b842f
```

The exact-head CI evidence above is evidence for `f1d2ba16...`; it is not automatically
relabelled as CI evidence for the squash merge SHA.

## Post-merge frontend deployment status

GitHub recorded Vercel status `success` for merge SHA
`ce7ee7ab0b95655fff7a1b546e277a4e1c0b842f` after the squash merge. This establishes that
the merged frontend revision passed the connected Vercel deployment status. It does **not**
close the separate Render/Neon/provider production-recovery track and does not replace the
final exact-production release smoke required for `v1.0.0`.

## Pixel-golden decision

The repository continues to use semantic browser assertions, deterministic layout
invariants, no-overflow checks, and retained desktop/mobile screenshot evidence instead of
committing `toHaveScreenshot` pixel-golden images for the full live map workspace.

A full-map golden would either couple CI to externally rendered OpenFreeMap/OSM tile and
font pixels or require masking the map surface that the test is intended to evaluate. That
trade-off would add maintenance noise without materially increasing confidence over the
current structural and screenshot-evidence suite.

Therefore this phase closes with:

```text
STRUCTURAL_VISUAL_REGRESSION=CLOSED
RETAINED_SCREENSHOT_EVIDENCE=CLOSED
PIXEL_GOLDEN_VISUAL_REGRESSION=NOT_ADOPTED_NONBLOCKING
```

This is a deliberate test-strategy decision, not an open frontend implementation defect.

## Runtime and release boundary

Nothing in PR #102 enabled production ingestion, changed provider policy, changed backend or
database behavior, or supplied the still-missing post-reset Neon runtime evidence.

```text
PRODUCTION_INGESTION=INTENTIONALLY_OFFLINE
PRODUCTION_PROVIDER_RECOVERY=OPEN_RUNTIME_VALIDATION
FREE_TIER_INFRASTRUCTURE_RECOVERY=OPEN_RUNTIME_VALIDATION
FINAL_EXACT_PRODUCTION_VALIDATION=OPEN
V1_RELEASE=OPEN
```

## Closure statement

```text
FRONTEND_PRODUCT_CLOSURE=CLOSED
FRONTEND_VISUAL_POLISH_V2=CLOSED
ZERO_BUDGET_DATA_BOUNDARY=ENFORCED
UNSUPPORTED_FR24_FIELDS=NOT_RENDERED
STRUCTURAL_VISUAL_REGRESSION=CLOSED
RETAINED_SCREENSHOT_EVIDENCE=CLOSED
PIXEL_GOLDEN_VISUAL_REGRESSION=NOT_ADOPTED_NONBLOCKING
PRODUCTION_PROVIDER_RECOVERY=OPEN_RUNTIME_VALIDATION
V1_RELEASE=OPEN
```
