# Document 195 — Frontend Product Closure

Status: CLOSED — exact-head CI verified and PR #98 merged
Date: 2026-08-23
Scope: `FRONTEND_PRODUCT_CLOSURE`

## Purpose

This document records the final closure boundary for the August 2026 frontend product redesign.
It does not claim feature or visual parity with Flightradar24 and it does not convert
unsupported aviation data into placeholders. The product remains an independent open-data
research interface with its own design system and evidence semantics.

## Closed frontend increments

The closed scope includes the repository work merged through PRs #90–#98:

- map-first application shell;
- semantic Playwright contract repair;
- zero-cost OpenFreeMap aviation basemap and MapLibre chrome;
- aircraft marker visual states, selection and labels;
- evidence-aware aircraft intelligence display model and dense detail layout;
- search plus evidence-backed motion and altitude filters;
- independent persisted-trajectory and projected-path visibility controls;
- responsive map viewport, coarse-pointer targets, reduced-motion, forced-colors and
  reduced-transparency support;
- strengthened desktop/mobile structural visual invariants and retained screenshot evidence;
- deterministic fixes for GeoJSON export hydration and dynamic visual-surface measurement.

## Data truth boundary

Frontend presentation remains constrained by actual GFA contracts.

```text
OBSERVED_OR_PROFILE_DATA       render when present and valid
INFERRED_ROUTE_DATA            render only with inference/confidence semantics
PROJECTED_PATH                 render as estimated evidence, separate from observations
UNAVAILABLE_FR24_ONLY_FIELDS   omit
FABRICATED_PLACEHOLDERS        prohibited
```

The redesign therefore does not add unsupported squawk, vertical speed, Mach, true or
indicated airspeed, authoritative ETA, gate, terminal, baggage or satellite-receiver
metadata when the current GFA contracts do not provide those values.

## Map and cost boundary

```text
MAP_ENGINE=MapLibre GL JS
BASEMAP_PROVIDER=OpenFreeMap
MAP_API_KEY_REQUIRED=NO
MAP_BILLING_ACCOUNT_REQUIRED=NO
NEW_PAID_FRONTEND_DEPENDENCY=NO
CURRENT_FRONTEND_MAP_COST=$0
```

No proprietary Flightradar24 code, imagery, aircraft assets or commercial map assets are
part of this closure.

## Browser and visual regression evidence

Playwright provides deterministic Chromium product coverage and retained screenshot
evidence for desktop and mobile selected-aircraft workspaces. The visual regression suite
checks semantic readiness, section order, wide-screen layout width, horizontal-overflow
absence, map-evidence control presence, and the responsive mobile map-height contract.

PR #98 exact-head evidence:

```text
PR=98
PR_HEAD=b43b6a87eb6224efc2e3c899effe21655d58b996
MERGE_SHA=c215e7ce5466577f149dc5669c5e0311daf6a56d
FRONTEND_CI=SUCCESS
BACKEND_CI=SUCCESS
CODEQL=SUCCESS
API_LOAD_BASELINE=SUCCESS
PLAYWRIGHT_E2E=SUCCESS
PLAYWRIGHT_BROWSER_SCENARIOS=20
PLAYWRIGHT_FLAKY_TESTS=0
```

The final Chromium run completed 20/20 scenarios successfully without a flaky retry. CSV,
GeoJSON, desktop visual-regression and mobile visual-regression scenarios all passed on the
same exact PR head.

This is intentionally not represented as pixel-golden regression.

```text
STRUCTURAL_VISUAL_REGRESSION=IMPLEMENTED
SCREENSHOT_EVIDENCE=IMPLEMENTED
PIXEL_GOLDEN_VISUAL_REGRESSION=OPEN
FRONTEND_PRODUCT_CLOSURE=CLOSED
```

Pixel-golden baselines remain a separate optional future improvement because the repository
does not currently own reviewed `toHaveScreenshot` golden files for the redesigned product.
Their absence does not reopen the structural product closure proven by PR #98.

## Production boundary

Frontend source closure is independent from production-provider recovery. The following
external/runtime items remain open and must not be implied closed by this document:

```text
PRODUCTION_PROVIDER_RECOVERY=OPEN_EXTERNAL
ADSBLOL_PRODUCTION_RESPONSE=PENDING
PRODUCTION_INGESTION=INTENTIONALLY_OFFLINE
FREE_TIER_INFRASTRUCTURE_RECOVERY=OPEN_RUNTIME_VALIDATION
FINAL_EXACT_PRODUCTION_VALIDATION=OPEN
V1_RELEASE_CLOSURE=OPEN
```

## Documentation index note

`docs/DOCUMENT_INDEX.md` currently stops at Document 193 and therefore still needs a
separate index reconciliation for Documents 194 and 195. This document does not hide that
existing documentation debt.

## Closure decision

The exact PR head containing the strengthened visual-regression contract completed every
required repository merge-evidence workflow successfully before merge. PR #98 was then
squash-merged using an expected-head guard, producing merge SHA
`c215e7ce5466577f149dc5669c5e0311daf6a56d`.

Therefore:

```text
FRONTEND_PRODUCT_SOURCE_IMPLEMENTATION=COMPLETE
FRONTEND_VISUAL_AND_INTERACTION_REDESIGN=IMPLEMENTED
FRONTEND_PRODUCT_CLOSURE=CLOSED
PIXEL_GOLDEN_VISUAL_REGRESSION=OPEN
DOCUMENT_INDEX_194_195=OPEN_GOVERNANCE_DEBT
V1_RELEASE=OPEN
```
