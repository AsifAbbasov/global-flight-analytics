# Document 195 — Frontend Product Closure

Status: CLOSED — exact-head CI verified and PR #98 merged; post-closure Visual Polish V2 recorded separately
Date: 2026-08-23
Scope: `FRONTEND_PRODUCT_CLOSURE`

## Purpose

This document records the final closure boundary for the August 2026 frontend product redesign.
It does not claim feature or visual parity with Flightradar24 and it does not convert
unsupported aviation data into placeholders. The product remains an independent open-data
research interface with its own design system and evidence semantics.

## Closed frontend increments

The original closed scope includes the repository work merged through PRs #90–#98:

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

At the time of the original PR #98 closure, pixel-golden baselines were deliberately left
as a separate optional improvement because the repository did not own reviewed
`toHaveScreenshot` golden files for the redesigned product. That historical decision did
not reopen the structural product closure proven by PR #98.

## Post-closure Visual Polish V2

After the original closure, PR #102 applied a separate zero-budget visual-polish increment
using the supplied Flightradar24 material only as a UX/information-hierarchy reference.
That increment promoted the two-dimensional live map to the primary tracker workspace,
compacted the aircraft/intelligence rail and popup hierarchy, prioritized supported live
telemetry, strengthened selected-aircraft state, and removed unsupported per-aircraft
placeholder values.

PR #102 exact head `f1d2ba16cec215b217955fbcd049dd8b728562c4` completed all five repository
merge-evidence workflows successfully, including a clean 20/20 Chromium run with zero
flaky retries, before expected-head squash merge
`ce7ee7ab0b95655fff7a1b546e277a4e1c0b842f`.

The V2 evidence and final comparison boundary are recorded in
[`196_FRONTEND_VISUAL_POLISH_V2_CLOSURE.md`](196_FRONTEND_VISUAL_POLISH_V2_CLOSURE.md).

The post-V2 test-strategy decision is:

```text
STRUCTURAL_VISUAL_REGRESSION=CLOSED
RETAINED_SCREENSHOT_EVIDENCE=CLOSED
PIXEL_GOLDEN_VISUAL_REGRESSION=NOT_ADOPTED_NONBLOCKING
FRONTEND_VISUAL_POLISH_V2=CLOSED
```

A full live-map pixel golden would either couple CI to externally rendered OpenFreeMap/OSM
pixels or mask the map surface itself. The repository therefore keeps deterministic
semantic/layout regression plus retained screenshots instead of treating a noisy golden as
a frontend completion requirement.

## Production boundary

Frontend source closure is independent from production-provider recovery. Since this
frontend closure was recorded, PR #100 merged the repository-side ADSB.lol provider-guidance
compliance hardening after the operator response was received. That later work does not
change the frontend closure decision and does not constitute live production-recovery
evidence.

The following external/runtime items remain open and must not be implied closed by this
document:

```text
ADSBLOL_PRODUCTION_RESPONSE=RECEIVED
ADSBLOL_COMPLIANCE_HARDENING=MERGED
PRODUCTION_PROVIDER_RECOVERY=OPEN_RUNTIME_VALIDATION
PRODUCTION_INGESTION=INTENTIONALLY_OFFLINE
FREE_TIER_INFRASTRUCTURE_RECOVERY=OPEN_RUNTIME_VALIDATION
FINAL_EXACT_PRODUCTION_VALIDATION=OPEN
V1_RELEASE_CLOSURE=OPEN
```

## Documentation index note

Document 196 is the post-closure visual-polish record. The documentation register must list
Documents 194, 195 and 196 before the former index-governance debt can be considered closed.

## Closure decision

The exact PR head containing the strengthened original visual-regression contract completed
every required repository merge-evidence workflow successfully before merge. PR #98 was
then squash-merged using an expected-head guard. The later PR #102 visual-polish increment
was independently exact-head verified and expected-head squash-merged without reopening the
original closure.

Therefore:

```text
FRONTEND_PRODUCT_SOURCE_IMPLEMENTATION=COMPLETE
FRONTEND_VISUAL_AND_INTERACTION_REDESIGN=IMPLEMENTED
FRONTEND_PRODUCT_CLOSURE=CLOSED
FRONTEND_VISUAL_POLISH_V2=CLOSED
STRUCTURAL_VISUAL_REGRESSION=CLOSED
RETAINED_SCREENSHOT_EVIDENCE=CLOSED
PIXEL_GOLDEN_VISUAL_REGRESSION=NOT_ADOPTED_NONBLOCKING
ADSBLOL_PRODUCTION_RESPONSE=RECEIVED
ADSBLOL_COMPLIANCE_HARDENING=MERGED
PRODUCTION_PROVIDER_RECOVERY=OPEN_RUNTIME_VALIDATION
V1_RELEASE=OPEN
```
