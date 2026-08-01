# Frontend Traffic Workspace

Status: Implementation prepared; exact-commit Continuous Integration closure pending
Project: Global Flight Analytics
Reviewed baseline: `12da70d42d1279d074e681afcba14a62991bdf08`
Date: 2026-08-01

## 1. Purpose

This increment replaces the continuously stacked aircraft explorer and intelligence
panels with an explicit traffic workspace. The map remains visible while the right
column switches between aircraft discovery and the selected aircraft intelligence
record.

## 2. User-visible behavior

The workspace provides two accessible tabs:

- `Aircraft` for search, sorting, regional summaries and selection;
- `Intelligence` for aircraft detail, route, projection, weather and stability
  evidence.

Selecting an aircraft from either the map or Aircraft Explorer normalizes the ICAO24
identifier, preserves one shared selection and opens Intelligence automatically.
Returning to the Aircraft tab does not discard the selection. Clearing the selection
or changing region returns the workspace to Aircraft and prevents stale cross-region
intelligence from remaining visible.

## 3. Information architecture correction

Before this increment, the right column rendered Aircraft Explorer and every
intelligence panel as one uninterrupted vertical stack. That made discovery and
analysis compete for the same surface and forced users to scroll through unrelated
content.

The new workspace keeps the geographic context stable and exposes one task at a
time without removing any existing analytical capability.

## 4. Deterministic state boundary

A pure TypeScript model owns ICAO24 normalization and the panel transition caused by
selection or clearing. Four dependency-free Node tests cover normalization, empty
identifiers, selection-to-intelligence transition and clearing-to-aircraft transition.

The frontend test configuration compiles the new state model into `.test-dist`, and
the existing wildcard test command runs these tests together with the API-client and
Aircraft Explorer suites.

## 5. Scope boundary

This increment deliberately does not add:

- a routing library or new browser URL contract;
- a global state manager;
- a new backend endpoint;
- a new package or lockfile change;
- browser end-to-end automation.

URL-addressable workspace state and browser automation remain separate increments
that require an explicit product contract.

## 6. Closure requirements

Formal closure requires all fifteen frontend tests, ESLint, TypeScript validation,
production frontend build, dependency policy gates, installer rollback verification
and the exact post-commit Frontend Continuous Integration run to pass.
