# Frontend Live Traffic Control

Status: Implementation prepared; exact-commit Continuous Integration closure pending  
Project: Global Flight Analytics  
Reviewed baseline: `8120101937da487757d9c00c84b7d988c21db760`  
Date: 2026-08-01

## 1. Purpose

This increment makes the live traffic refresh policy visible and controllable. The
existing traffic query already refreshed every sixty seconds, but the interface exposed
only a manual refresh button and the timestamp of the latest response. Users could not
see whether a snapshot was current, aging or stale, could not pause automatic requests,
and could not choose a bounded refresh interval.

## 2. Refresh policy

The published automatic refresh choices are thirty seconds, one minute and two minutes.
One minute remains the default. The user may pause and resume automatic traffic requests
without clearing the latest successful snapshot. Manual refresh remains available while
automatic refresh is paused.

The preference is intentionally local interface state. It is not serialized into the
shareable workspace URL because request cadence is an operator preference rather than
part of the analytical evidence being shared.

## 3. Freshness semantics

The browser derives freshness from the latest successful React Query `dataUpdatedAt`
timestamp and the selected refresh interval:

```text
age < one interval            → current
one interval ≤ age < two      → aging
age ≥ two intervals           → stale
no successful timestamp       → waiting
```

A failed refresh does not erase the latest successful response. The control reports a
degraded state and explicitly says that the previous snapshot is retained. A first-load
failure is classified as unavailable.

## 4. User-visible evidence

The control displays:

1. current regional aircraft count and selected aircraft context;
2. absolute and relative time of the latest successful update;
3. interval progress toward the next automatic request;
4. selected refresh cadence, countdown, due state or paused state;
5. loading, refreshing, degraded, unavailable, current, aging and stale status;
6. regional catalog warnings and traffic errors with a retry action.

The browser clock is presentation evidence only. It does not replace server-owned
observation timestamps or analytical freshness metrics.

## 5. Query integration

`useCurrentTraffic` now accepts a validated refresh interval or `false`. Published
interval values are normalized by the pure live-traffic status model. React Query keeps
its existing retry policy and continues to avoid background-tab polling.

## 6. Regression evidence

Dependency-free tests verify:

1. only published intervals are accepted;
2. absent timestamps remain stable;
3. current snapshots expose bounded countdown and progress;
4. aging and stale boundaries are explicit;
5. paused refresh removes the countdown without rewriting freshness;
6. future timestamps cannot produce negative age.

Full frontend tests, ESLint, TypeScript validation, dependency security policy and the
Next.js production build remain mandatory gates.

## 7. Scope boundary

This increment does not change backend ingestion cadence, add WebSockets, introduce
server-sent events, poll hidden browser tabs, persist refresh preferences, add a routing
parameter, change analytical freshness contracts or add a dependency.
