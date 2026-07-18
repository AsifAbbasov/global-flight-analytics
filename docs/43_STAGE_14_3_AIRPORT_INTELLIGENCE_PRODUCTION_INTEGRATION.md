# Document 43 — Stage 14.3 Airport Intelligence Production Integration

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: PostgreSQL composition, read-only HTTP integration, and runtime reachability of the complete Airport Intelligence domain

## 1. Purpose

Stage 14.3 converts the previously isolated Airport Intelligence domain foundation into a production read path.

The domain packages are retained because they implement distinct validated responsibilities:

```text
passport
statistics
ranking
overview
history
trends
```

They are now composed by `airportintelligence/airportproduction`.

## 2. Production Data Boundary

The production reader uses existing project tables:

```text
airports
airport_profiles
airport_statistics
route_statistics
route_predictions
```

No official airport operations feed is claimed.

Daily airport observations are built only from completed Coordinated Universal Time dates present in `airport_statistics`. Active-route context is derived from `route_statistics`. Active-aircraft context is derived from available `route_predictions`.

## 3. Read-Only HTTP Routes

```text
GET /api/v1/airports/intelligence/ranking
GET /api/v1/airports/:icao/intelligence/overview
GET /api/v1/airports/:icao/intelligence/history
GET /api/v1/airports/:icao/intelligence/trends
```

Supported query parameters are `days`, `as_of_time`, and ranking-only `limit`.

## 4. Window Semantics

The current partial Coordinated Universal Time day is excluded. This prevents partial-day values from being ranked against complete daily windows.

## 5. Statistics Semantics

For a multi-day window:

```text
arrivals = sum of daily arrivals
departures = sum of daily departures
active aircraft = peak available daily active-aircraft count
active routes = peak available daily active-route count
observed samples = number of dates with airport statistics
expected samples = requested completed-day count
```

## 6. Ranking Semantics

Airport Activity Score remains relative to airports with observations in the same requested window. It is not an absolute worldwide airport classification.

## 7. Security Boundary

All Stage 14.3 routes are read-only and expose open research data. User authentication is not required for these routes. Administrative materialization and mutation routes must be protected before deployment.

## 8. Runtime Completion Gate

Stage 14.3 is complete only when strict architecture audit reports every Airport Intelligence package as reachable from `cmd/server`.

## 9. Known Limitations

```text
No official airport operations feed
No guaranteed complete route coverage
No guaranteed complete active-aircraft coverage
No current partial-day ranking
No causal explanations
No forecasting
No universal calibration of project-derived ranking weights
```
