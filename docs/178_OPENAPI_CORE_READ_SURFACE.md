# Document 178 — OpenAPI Core Read Surface

Status: IMPLEMENTED
Project: Global Flight Analytics
Baseline commit: `8a873c6c60add160b7f355f2a90bde6771990156`
Scope: ten source-backed public GET operations

## 1. Decision

The first OpenAPI closure slice expands one coherent production boundary instead of adding
unrelated endpoints one at a time. The selected slice owns the core entities that other
analytical surfaces depend on:

```text
aircraft
flights
flight states
trajectories
route context
active-aircraft metric
```

This moves the public contract from eight to eighteen operations while keeping every new
operation read-only.

## 2. Added operations

```text
GET /api/v1/metrics/active-aircraft
GET /api/v1/aircraft
GET /api/v1/aircraft/{icao24}
GET /api/v1/flights
GET /api/v1/flights/{id}
GET /api/v1/flights/{flightID}/states
GET /api/v1/aircraft/{icao24}/latest-state
GET /api/v1/aircraft/{icao24}/trajectory
GET /api/v1/trajectories/{id}
GET /api/v1/aircraft/{icao24}/route-context
```

## 3. Source evidence

Route registration is owned by:

```text
apps/api/internal/server/core_database_routes.go
```

Request and response behavior is derived from:

```text
apps/api/internal/http/handlers/metrics.go
apps/api/internal/http/handlers/aircraft.go
apps/api/internal/http/handlers/flights.go
apps/api/internal/http/handlers/flightstates.go
apps/api/internal/http/handlers/trajectories.go
apps/api/internal/http/handlers/route_context.go
```

Canonical transport fields are derived from:

```text
apps/api/internal/http/dto/metrics.go
apps/api/internal/http/dto/aircraft.go
apps/api/internal/http/dto/flight.go
apps/api/internal/http/dto/flightstate.go
apps/api/internal/http/dto/trajectory.go
apps/api/internal/http/dto/route_context.go
```

## 4. Important semantic boundaries

### Active-aircraft metric

- `window_minutes` defaults to 15;
- accepted values are 1 through 180;
- `region` is optional;
- scope, observation window, confidence, sources, and limitations remain explicit.

### Flight-state altitude

Barometric and geometric altitude values are nullable. Their companion status fields remain
required and distinguish `observed`, `ground`, `unknown`, `unavailable`, and `invalid`.
Missing or invalid altitude is never converted to zero.

### Trajectory

The contract preserves identity basis, split reason, time range, counts, quality score,
segments, coverage gaps, source name, and persistence timestamps. A trajectory response is
not reduced to map geometry only.

### Route context

Origin and destination candidates are optional because open observations may be insufficient.
Confidence, reasons, limitations, and generation time remain explicit. The API does not assert
a filed flight plan.

## 5. Error contracts

The new operations preserve source-backed error classes:

- `400` for invalid ICAO24, trajectory identifier, or metric window;
- `404` for absent aircraft, flight, flight state, trajectory, region, or route context;
- `429` for the public rate-limit boundary;
- `500` for unexpected backend failures;
- `503` where trajectory or route-context services are unavailable.

All errors use the existing typed error envelope and `X-Request-ID` header.

## 6. Playwright alignment

`apps/web/e2e/mock-api.mjs` mirrors all eighteen documented paths. The added fixtures are
deterministic and preserve the same success envelopes, nullable altitude semantics,
trajectory structure, metric scope, and route-context limitations as the OpenAPI contract.

The browser scenario count does not increase in this increment. Product-level selected-aircraft
and trajectory scenarios remain separate from transport contract closure.

## 7. Verification

Required markers:

```text
OPENAPI_CONTRACT_PATHS=18
OPENAPI_CORE_READ_OPERATIONS=10
OPENAPI_CONTRACT_SCHEMAS=PASS
OPENAPI_CONTRACT_ROUTE_DRIFT=PASS
OPENAPI_CONTRACT=PASS
OPENAPI_DOCUMENTED_OPERATIONS=18
OPENAPI_MISSING_OPERATIONS=20
OPENAPI_EXTRA_OPERATIONS=0
PLAYWRIGHT_E2E_OPENAPI_PATHS=18
PLAYWRIGHT_E2E_MOCK_API=PASS
PLAYWRIGHT_E2E_CONTRACT=PASS
FULL_RELEASE_VALIDATION=PASS
```

## 8. Remaining work

Twenty public operations remain outside the OpenAPI contract. They should be closed in
bounded slices rather than one giant specification edit:

1. transponder evidence and route intelligence;
2. weather and analytical metrics;
3. airport intelligence;
4. historical intelligence;
5. projection, stability, weather context, and airspace analytics;
6. protected mutation security scheme and final complete-surface gate.
