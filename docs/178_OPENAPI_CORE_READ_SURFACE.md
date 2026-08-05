# Document 178 — OpenAPI Core Read Surface

Status: IMPLEMENTED
Project: Global Flight Analytics
Scope: ten source-backed core public GET operations

## Purpose

This increment expanded the original eight-operation OpenAPI foundation with ten core reads:

```text
GET /api/v1/metrics/active-aircraft
GET /api/v1/aircraft/{icao24}/trajectory
GET /api/v1/aircraft/{icao24}/route-context
GET /api/v1/trajectories/{id}
GET /api/v1/flights/{flightID}/states
GET /api/v1/aircraft/{icao24}/latest-state
GET /api/v1/flights
GET /api/v1/flights/{id}
GET /api/v1/aircraft
GET /api/v1/aircraft/{icao24}
```

## Preserved contracts

The slice preserves aircraft and flight DTOs, nullable altitude values with explicit status,
trajectory segments and coverage gaps, inferred route-context confidence and limitations, and
bounded active-aircraft metric windows.

## Verification

The original increment established:

```text
OPENAPI_CONTRACT_PATHS=18
OPENAPI_CORE_READ_OPERATIONS=10
OPENAPI_DOCUMENTED_OPERATIONS=18
OPENAPI_MISSING_OPERATIONS=20
```

Those ten operations remain unchanged and are now part of the complete thirty-eight-operation
public contract. Their Playwright fixtures continue to provide deterministic trajectory,
nullable-altitude, route-context, and metric envelopes.

## Current relationship to closure

Document 179 adds seventeen advanced intelligence reads without modifying this slice's route,
parameter, or DTO semantics. Document 180 subsequently closes the three Route Intelligence
operations without changing this core slice.
