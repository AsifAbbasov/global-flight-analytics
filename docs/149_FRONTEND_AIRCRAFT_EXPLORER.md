# Frontend Aircraft Explorer

Status: Implementation prepared; exact-commit Continuous Integration closure pending
Project: Global Flight Analytics
Reviewed baseline: `ee2bb13e60c29ae9ecdcb7736d4fe39561e3b28d`
Date: 2026-08-01

## 1. Purpose

This increment converts the regional traffic snapshot from a map-only discovery
surface into a searchable and sortable aircraft index. The explorer is linked to
the existing map selection and aircraft intelligence panels, so selecting a row
opens the same route, trajectory, projection, weather and stability evidence.

## 2. User-visible capability

The Aircraft Explorer provides:

- search across callsign, ICAO24, aircraft model, airline and origin country;
- sorting by latest observation, callsign, altitude or speed;
- matched, airborne, on-ground and unknown-altitude summaries;
- a bounded list of one hundred visible results;
- explicit selection state shared with the map and detail panels;
- altitude, speed and observation-time context for every listed aircraft;
- empty states for unavailable traffic and unmatched searches.

## 3. Deterministic model boundary

Filtering, sorting, limiting and summary calculations live in a pure TypeScript
model. The model uses deterministic ICAO24 tie-breaking and keeps null altitude or
invalid timestamps after usable values. Five dependency-free Node tests cover the
search contract, altitude sorting, speed sorting, timestamp ordering and summary
semantics.

The existing frontend test command now runs every `*.test.mjs` file, and the test
TypeScript configuration compiles both the API client and the aircraft explorer
model into the isolated `.test-dist` directory.

## 4. Scope boundary

This increment deliberately does not add:

- a new backend endpoint;
- pagination owned by the server;
- browser end-to-end automation;
- a new package or lockfile change;
- virtualization, Redis or a search service.

The current traffic snapshot is already bounded by the backend and analytical
query limits. Server-owned pagination should be introduced only when measured
snapshot sizes justify it.

## 5. Closure requirements

Formal closure requires the existing six API-client tests, the five aircraft
explorer model tests, ESLint, TypeScript validation, production frontend build,
dependency policy gates, installer rollback verification and the exact post-commit
Frontend Continuous Integration run to pass.
