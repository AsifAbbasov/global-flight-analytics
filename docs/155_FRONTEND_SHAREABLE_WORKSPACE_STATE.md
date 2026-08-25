# Frontend Shareable Workspace State

Status: CLOSED — feature implementation and Continuous Integration evidence reconciled  
Project: Global Flight Analytics  
Reviewed baseline: `70716701d6e676b49670aa4f32b4608d52f58bd6`  
Implementation commit: `8120101937da487757d9c00c84b7d988c21db760`  
Frontend CI: `30710813766` — SUCCESS  
Backend CI: `30710813764` — SUCCESS  
Date: 2026-08-01
Canonical reconciliation: 2026-08-25

## 1. Purpose

This increment makes the regional traffic workspace addressable through the browser URL.
Previously, region choice, aircraft selection and the Aircraft or Intelligence panel were
component-local state. Refreshing the page discarded them and a reviewer could not send
a link that reopened the same analytical context.

## 2. URL contract

The workspace owns three query parameters:

```text
region=<registered region code>
aircraft=<six-character ICAO24 identifier>
view=aircraft|intelligence
```

Unknown region codes fall back to the configured initial region. Invalid aircraft values
are removed. When an aircraft is supplied without an explicit view, Intelligence is the
default. Other query parameters are preserved and sorted deterministically.

## 3. Browser navigation

User-driven region, selection and panel changes use `history.pushState`. Browser Back and
Forward events restore the complete workspace state through a `popstate` listener. Initial
page hydration canonicalizes malformed or incomplete parameters with `replaceState` so it
does not create a fake navigation entry.

Changing region clears aircraft selection and returns to the Aircraft panel. Selecting an
aircraft opens Intelligence. A user may then return to the Aircraft panel while retaining
the selected identifier, and that state remains representable in the URL.

## 4. Share action

The traffic command bar exposes a Copy view link action. It copies the current canonical
browser URL and reports success or clipboard unavailability without changing analytical
state.

## 5. Regression evidence

Six dependency-free tests verify valid deep links, invalid input fallback, selected-aircraft
panel retention, canonical serialization, selection clearing and pathname/hash preservation.
Existing API client, application status, Aircraft Explorer, Traffic Workspace and Regional
Traffic Brief tests remain unchanged.

## 6. Scope boundary

This increment does not add a routing dependency, backend endpoint, server session, user
account, persisted preference, analytics tracker or cross-device synchronization. URL state
contains only public workspace identifiers and never stores credentials or mutation keys.

The historical closure requirements were all frontend tests, dependency policy, ESLint,
TypeScript validation, production build, rollback verification and exact-commit Frontend
Continuous Integration evidence.

## 7. Historical closure evidence

The exact implementation owner is:

```text
8120101937da487757d9c00c84b7d988c21db760
feat: add shareable traffic workspace state
```

GitHub Actions evidence for that exact commit is:

```text
Frontend CI 30710813766 = SUCCESS
Backend CI  30710813764 = SUCCESS
```

The frontend run completed dependency policy, ESLint, TypeScript validation, the
workspace-state contract tests and the production build. Exact-commit CI therefore
exists for the implementation, so the original pending status is historical drift rather
than an open present-day item.

## 8. Canonical classification

This document is **frontend feature implementation / closure evidence**, not a
remediation finding record.

Component-local state being non-shareable was the prior product capability boundary.
The source evidence does not establish data corruption, contract violation or reliability
failure requiring a synthetic `GFA-*` remediation ID merely because URL-addressable state
was added later.

```text
Canonical finding ID: none by design
Classification: frontend feature implementation / closure evidence
Historical implementation: CLOSED
Exact-commit Frontend CI: CLOSED
Open remediation findings owned by this document: 0
```

## 9. Residual boundaries and prevention

URL state remains intentionally limited to public workspace identifiers; credentials,
mutation keys and request-cadence preferences must not enter the share contract.
Regression ownership remains with URL-state tests, frontend contract tests, Frontend CI
and later browser navigation/Playwright coverage.
