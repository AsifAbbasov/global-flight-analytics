# Frontend Regional Traffic Brief

Status: Implementation prepared; exact-commit Continuous Integration closure pending
Project: Global Flight Analytics
Reviewed baseline: `c34264a122913272e3083a4f64397b0e8470c4f3`
Date: 2026-08-01

## 1. Purpose

This increment adds an explanatory layer between the live globe and the detailed map
workspace. The existing product already exposes protected analytical metrics and a
searchable aircraft index, but a reviewer still has to inspect many individual markers
to understand the composition of the current regional snapshot.

The Regional Traffic Brief converts the already-loaded traffic response into a compact,
deterministic summary without adding a backend request or claiming historical trends.

## 2. User-visible capability

For the selected region, the brief presents:

1. total, airborne and on-ground aircraft counts;
2. usable airborne-altitude coverage;
3. low, medium, high and unavailable airborne altitude bands;
4. the leading attributed airlines;
5. the leading provider-supplied aircraft origin countries;
6. explicit unattributed-record counts;
7. an updating state linked to the existing traffic query.

The brief remains visible above the map workspace and updates whenever the selected
region or current traffic snapshot changes.

## 3. Evidence boundary

All values are derived in the browser from the current API snapshot. The component does
not infer demand, congestion, safety, route popularity, historical growth or operational
conditions. Altitude-band shares use airborne aircraft as their denominator. Airline and
origin-country shares use the complete current snapshot and preserve unknown counts.

Altitude classification is explicit:

```text
low: below 3,000 metres
medium: 3,000 through 8,999 metres
high: 9,000 metres and above
unknown: null, negative or non-finite airborne altitude
```

Ground aircraft are excluded from the airborne altitude profile even when a provider
supplies a numeric altitude.

## 4. Deterministic model and tests

A pure TypeScript model owns classification, label normalization, ranking, shares and
bounded output. Rankings group case variants, collapse internal whitespace, sort by
count descending and use alphabetical tie-breaking. The ranking limit is clamped to a
maximum of five.

Five dependency-free Node tests cover empty snapshots, altitude boundaries, invalid
altitudes, deterministic label ranking and ranking-limit enforcement. Existing API
client, application status, aircraft explorer and traffic workspace tests remain intact.

## 5. Scope boundary

This increment does not add a chart library, backend endpoint, historical aggregation,
server pagination, telemetry, new package, lockfile change or operational claim.

Formal closure requires all twenty-five frontend contract tests, dependency policy,
ESLint, TypeScript validation, production build, installer rollback verification and
the exact post-commit Frontend Continuous Integration run to pass.
