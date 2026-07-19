# Document 48 — Stage 14.8 Server Composition Root Decomposition

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: structural decomposition of database-backed server composition without HTTP behavior changes

## 1. Problem

The previous `database_routes.go` combined:

```text
PostgreSQL repository construction
domain service construction
intelligence runtime construction
HTTP handler construction
route registration
mutation middleware placement
error wrapping
```

The file contained more than four hundred lines. Its main registration function
had to change whenever unrelated bounded contexts gained or changed routes.

That structure increased the review surface and made the production dependency
graph difficult to inspect.

## 2. Decision

`internal/server` remains the composition root.

The project does not introduce:

```text
reflection-based dependency injection
a service locator
a global dependency container
framework-generated wiring
runtime plugin discovery
```

Concrete PostgreSQL adapters are still selected explicitly inside the server
composition root.

The change separates detailed composition from HTTP route registration.

## 3. Resulting Structure

```text
database_routes.go
core_database_composition.go
core_database_routes.go
route_intelligence_database_composition.go
route_intelligence_database_routes.go
projection_database_composition.go
projection_database_routes.go
airspace_database_composition.go
airspace_database_routes.go
```

`database_routes.go` now describes only the ordered bounded-context startup map.

Composition files create repositories, services, readers, and handlers.

Route files register HTTP methods and paths against already composed runtimes.

## 4. Preserved Behavior

The increment does not change:

```text
HTTP methods
HTTP paths
request DTO contracts
response DTO contracts
PostgreSQL schema
SQL migrations
analytical formulas
provider behavior
frontend API clients
```

The protected mutation route remains:

```text
POST /api/v1/trajectories/:id/route-intelligence
```

Its mutation authorization handler remains the first handler in the route
chain.

## 5. Regression Gates

Route topology tests verify all eighteen core and Route Intelligence routes.

They verify:

```text
method and path preservation
absence of duplicate routes
handler count preservation
mutation authorization ordering
```

Architecture tests verify:

```text
the coordinator does not import bounded-context implementations
the coordinator does not register HTTP verbs directly
composition files do not register HTTP verbs
route files do not construct PostgreSQL or domain infrastructure
registerDatabaseRoutes remains narrow
```

## 6. Intentional Boundaries

Existing bounded contexts that already own dedicated registration functions
remain in their existing files:

```text
Weather
Analytical Metrics
Airport Intelligence
Historical Intelligence
```

They are coordinated through the new route-group map but are not rewritten in
this increment.

This avoids combining structural decomposition with unrelated behavior
changes.

## 7. Acceptance

The increment is accepted only after:

```text
focused server tests
targeted race detector
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
