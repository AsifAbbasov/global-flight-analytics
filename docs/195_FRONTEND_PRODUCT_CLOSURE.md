# Document 195 — Frontend Product Closure

Status: CANDIDATE — requires exact-head CI before merge
Date: 2026-08-23
Scope: `FRONTEND_PRODUCT_CLOSURE`

## Purpose

This document records the closure boundary for the August 2026 frontend product redesign.
It does not claim feature or visual parity with Flightradar24 and it does not convert
unsupported aviation data into placeholders. The product remains an independent open-data
research interface with its own design system and evidence semantics.

## Closed frontend increments

The closure candidate includes the repository work merged through PRs #90–#97:

- map-first application shell;
- semantic Playwright contract repair;
- zero-cost OpenFreeMap aviation basemap and MapLibre chrome;
- aircraft marker visual states, selection and labels;
- evidence-aware aircraft intelligence display model and dense detail layout;
- search plus evidence-backed motion and altitude filters;
- independent persisted-trajectory and projected-path visibility controls;
- responsive map viewport, coarse-pointer targets, reduced-motion, forced-colors and
  reduced-transparency support.

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

This is intentionally not represented as pixel-golden regression.

```text
STRUCTURAL_VISUAL_REGRESSION=IMPLEMENTED
SCREENSHOT_EVIDENCE=IMPLEMENTED
PIXEL_GOLDEN_VISUAL_REGRESSION=OPEN
```

Pixel-golden baselines remain a separate future improvement because the repository does
not currently own reviewed `toHaveScreenshot` golden files for the redesigned product.

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

## Closure criteria

`FRONTEND_PRODUCT_CLOSURE` may be marked closed only after the exact PR head containing
this document and the strengthened visual-regression contract completes all of the
repository's frontend merge evidence checks successfully:

- Frontend CI;
- Playwright E2E;
- Backend CI;
- CodeQL;
- API Load Baseline.

Until that exact-head evidence is green, this document remains a closure candidate rather
than final closure evidence.
