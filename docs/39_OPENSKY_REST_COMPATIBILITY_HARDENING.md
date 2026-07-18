# OpenSky REST Compatibility Hardening

Status: Implemented backend compatibility amendment

## Purpose

The official OpenSky REST API returns the aircraft category only when `extended=1` is requested. A base State Vector therefore contains seventeen fields, while the extended representation contains eighteen fields.

The production OpenSky provider now always requests `extended=1`, while the parser remains backward compatible with both representations.

## Enforced behavior

- Production regional State Vector requests include `extended=1`.
- Seventeen-field State Vectors are accepted.
- A missing category is represented as `AircraftCategoryNoInformation`.
- `CategoryAvailable` distinguishes an omitted category from a provider-reported category value of zero.
- Eighteen-field State Vectors preserve the provider category.
- Responses with fewer than seventeen fields remain invalid.
- The compatibility verifier performs no external network request.

## Scope boundary

`/states/own` is not enabled. It is useful only for receivers owned by the authenticated OpenSky account, while Global Flight Analytics has no project-owned receiver network.

The official Java binding is not added as a dependency. It is a Java wrapper over the REST API, while this project uses a Go backend and a domain-specific Go integration with provider budget, health, fallback, provenance, and source constraints.

## Verification

```bash
cd apps/api
go test ./internal/integrations/opensky ./cmd/verify-opensky-rest-compatibility
go vet ./internal/integrations/opensky ./cmd/verify-opensky-rest-compatibility
go run ./cmd/verify-opensky-rest-compatibility
go test ./...
```

<!-- OPEN-AVIATION-RESEARCH-EVIDENCE-V1-2:REST-METADATA -->

## Canonical Observation Metadata Preservation

REST compatibility now continues beyond parsing. Squawk code, Special Purpose Indicator, position source, aircraft category, and category availability are preserved in the canonical `FlightState` and PostgreSQL persistence boundary. Category availability remains distinct from an observed category value of zero.
