# Document 55 — Stage 14.15 Weather Composition Boundary

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: separate Weather HTTP registration from provider and application wiring

## 1. Problem

The Weather route composition previously lived inside one function:

```text
registerWeatherRoute
```

That function performed every responsibility in the production chain:

```text
create provider budget manager
create provider response controller
create provider response observer
create ingestion orchestrator
create Open-Meteo integration client
create orchestrated weather provider client
create PostgreSQL weather repository
create weather application service
create HTTP handler
register the HTTP route
```

The code was functional, but its boundary was misleading. A function named
`registerWeatherRoute` also owned provider governance, external integration,
orchestration, persistence, application service construction, and HTTP
registration.

This made focused testing harder and allowed future provider changes to modify
the route registration boundary directly.

## 2. Production Decision

The same-package server composition is split by responsibility without adding
new runtime layers or exported abstractions.

The resulting files are:

```text
weather_route.go
weather_composition.go
weather_provider_composition.go
weather_application_composition.go
weather_route_registration.go
```

## 3. Responsibility Boundaries

### 3.1 Route Coordinator

`weather_route.go` performs only:

```text
request dependency composition
register the composed handler
return composition or registration errors
```

It does not import or construct provider, repository, service, or handler
implementations.

### 3.2 Provider Composition

`weather_provider_composition.go` owns:

```text
provider budget manager
provider response controller
provider response observer
request coalescing orchestrator
Open-Meteo integration client
orchestrated weather provider client
```

Existing error wrapping is preserved:

```text
initialize provider budget manager
initialize provider response controller
initialize provider response observer
initialize ingestion orchestrator
initialize open-meteo client
initialize orchestrated weather client
```

### 3.3 Application Composition

`weather_application_composition.go` owns:

```text
PostgreSQL weather repository
weather application service
weather HTTP handler
```

It does not register routes or construct external providers.

### 3.4 Route Registration

`weather_route_registration.go` owns only:

```text
CurrentWeatherPath
router and handler validation
Fiber GET route registration
```

The route remains:

```text
GET /api/v1/weather/current
```

## 4. Preserved Runtime Behavior

This increment does not change:

```text
Open-Meteo base URL
Open-Meteo timeout behavior
provider policy
provider budget accounting
provider response observation
request coalescing
weather request key
coordinate validation
PostgreSQL weather persistence
HTTP query parameters
HTTP response contract
error codes
database schema
migrations
frontend behavior
```

The dependency graph contains the same concrete implementations in the same
order as before.

## 5. Why Same-Package Decomposition

The components remain inside the `server` package because this is composition
code, not a new domain.

Creating additional exported packages or interfaces only to reduce file length
would add artificial architecture. The split instead uses file-level
responsibility boundaries while keeping compile-time visibility narrow.

## 6. Testing Strategy

Automated tests verify:

```text
invalid Open-Meteo timeout preserves the existing wrapped error
a complete provider-to-handler graph can be constructed without network access
route registration preserves GET /weather/current
latitude and longitude reach the same handler service contract
nil router and handler dependencies are rejected
the route coordinator contains no provider or application constructors
provider composition contains no persistence or route registration
application composition contains no provider or route registration
route registration contains no provider, repository, or service construction
```

The successful composition test intentionally permits a nil PostgreSQL pool
because the existing repository constructor also permits it. Runtime access
still returns the repository's existing pool-required error. This preserves
the previous construction semantics rather than silently changing startup
behavior.

## 7. Acceptance

The increment is accepted only after:

```text
focused Weather server tests
Weather provider and service tests
Weather repository tests
Weather HTTP handler tests
architecture regression tests
race detector
strict project architecture audit
complete Go build
go vet
complete Go test suite
frontend dependency security verification
frontend production dependency audit
ESLint
TypeScript validation
Next.js production build
backend Docker image build
git diff check
```
