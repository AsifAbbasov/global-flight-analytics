# Document 186 — Playwright Product Coverage Expansion

Status: IMPLEMENTED
Project: Global Flight Analytics
Scope: deterministic Chromium product journeys on the real production Next.js build

## 1. Purpose

This increment expands the original Playwright foundation from four shell-oriented
scenarios into twenty deterministic Chromium browser scenarios covering the
implemented product surface.

The browser still runs against the production Next.js build and the local mock API.
No Vercel deployment, Render deployment, production credential or external aviation
provider is used.

## 2. Product Journey Coverage

The suite now covers:

1. server-rendered startup state;
2. canonical region and workspace URL state;
3. mobile primary navigation;
4. traffic outage and retry;
5. Aircraft Explorer selection;
6. selected-aircraft deep-link restoration and clearing;
7. Airport Intelligence ranking, passport, history and trends;
8. Historical Intelligence global evidence plus bounded airport/route input states;
9. Projection Intelligence evidence;
10. Weather Context evidence;
11. Stability and Explainability evidence;
12. deterministic CSV export content;
13. deterministic GeoJSON export content and provenance;
14. region-catalog fallback and reload recovery;
15. aircraft profile, route-context and trajectory recovery;
16. Airport Intelligence recovery;
17. Historical Intelligence recovery;
18. advanced Projection/Weather/Stability recovery;
19. keyboard, landmark and mobile accessibility coverage;
20. desktop/mobile visual-layout regression evidence.

## 3. Deterministic Failure Surface

The mock API publishes seven private test scenarios:

```text
healthy
traffic-error
regions-error
aircraft-error
airport-error
historical-error
intelligence-error
```

Failure scenarios remain test-only and are controlled by `/__e2e/scenario`.
They do not alter the public OpenAPI path surface.

The Stability Intelligence mock is request-reflective: it returns exactly the
requested ordered `as_of_times` and synthesizes matching projection, forecast-version
and transition lineage. This mirrors the production client contract, which rejects
stability history whose timestamps do not match the requested analytical history.

## 4. Accessibility Boundary

Browser coverage verifies semantic roles and names, skip-link focusability and
keyboard activation, desktop/mobile navigation, analytical tabs, progress-bar labels
and the absence of horizontal page overflow on the mobile research workspace. The
focus assertion is intentionally deterministic across macOS and Linux, where native
Tab-to-link behavior can differ with operating-system keyboard-navigation settings.

Tests continue to prohibit `data-testid` selectors and direct `page.locator(...)`
selectors so product semantics remain the preferred browser contract.

## 5. Visual Regression Boundary

Before the separate visual redesign phase, the suite protects visual structure rather
than freezing the current pre-redesign pixels.

The visual regression tests therefore:

- pin desktop and mobile viewport sizes;
- assert analytical section ordering and minimum usable desktop width;
- reject horizontal page overflow;
- capture full-page desktop and selected-aircraft mobile screenshots;
- attach those screenshots to Playwright evidence artifacts on every run.

Pixel-golden `toHaveScreenshot` baselines are intentionally deferred until the visual
redesign is complete. Creating golden pixels now would fossilize an interface that is
already scheduled for redesign and would immediately make the baselines obsolete.

## 6. Permanent Verification

Canonical commands:

```bash
pnpm run test:playwright-e2e-contract
pnpm run verify:playwright-e2e
pnpm run run:playwright-e2e
pnpm run verify:release
```

Expected markers:

```text
PLAYWRIGHT_E2E_VERSION=1.62.0
PLAYWRIGHT_E2E_OPENAPI_PATHS=38
PLAYWRIGHT_E2E_MOCK_SCENARIOS=7
PLAYWRIGHT_E2E_BROWSER_SCENARIOS=20
PLAYWRIGHT_E2E_PRODUCT_COVERAGE=PASS
PLAYWRIGHT_E2E_VISUAL_LAYOUT_REGRESSION=PASS
PLAYWRIGHT_E2E_MOCK_API=PASS
PLAYWRIGHT_E2E_CONTRACT=PASS
PLAYWRIGHT_E2E=PASS
```

## 7. Completion Boundary

This closes the current functional Playwright product-journey debt.

The later Frontend Visual and Interaction Redesign must add final pixel-golden
screenshot baselines after the redesigned interface is stable. That is a visual
design release gate, not missing functional browser coverage in this increment.
