# Document 48 — Stage 14.8 Server Composition Root Decomposition

Status: Remediation History v1.1
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

## 8. Canonical finding record — GFA-MAINT-042

### Finding / symptom

`database_routes.go` combined production dependency construction, bounded-context runtime wiring, handler construction, HTTP route registration, mutation middleware placement, and error wrapping inside one >400-line composition/registration surface.

### Root cause

As bounded contexts were added, their wiring accumulated in the original central registration file. The code remained functional, but composition and HTTP topology were not separated into stable responsibilities.

### Failure scenario

A developer changes one bounded context's repository/service construction and must edit the same coordinator that owns unrelated routes and security middleware. The expanded diff/review surface makes accidental route removal, duplicate registration, middleware reordering, or cross-context dependency coupling easier.

### Impact

The finding primarily affects maintainability, reviewability, and architecture inspection. Because mutation middleware placement shares the file, an unsafe structural edit could also become security-significant, but no historical route/security regression is asserted.

### Severity rationale

**P3 retrospective.** The historical document explicitly frames this as structural decomposition with preserved behavior, not a confirmed production correctness incident.

### Existing guarantees violated

- the composition root should make dependency selection explicit and inspectable;
- route files should own HTTP topology, not infrastructure construction;
- composition files should not register HTTP verbs;
- coordinator changes for one bounded context should not require editing unrelated context internals;
- security middleware ordering must remain independently testable.

### Considered solutions

1. keep the monolithic file and add comments/regions;
2. introduce a dependency-injection framework/service locator;
3. keep explicit same-package composition but split coordinator, context composition, and route registration by responsibility.

### Chosen remediation

`internal/server` remains the composition root, while bounded-context composition and route registration move into dedicated same-package files. `database_routes.go` becomes a narrow ordered coordinator.

### Why selected

The problem was source ownership, not lack of runtime indirection. Same-package decomposition reduces review surface while preserving explicit concrete adapter selection and compile-time wiring.

### Rejected alternatives

Comments/regions do not reduce coupled change surface. Reflection DI, service locators, global containers, generated wiring, and runtime plugin discovery were rejected because they add abstraction/operational complexity without solving a need for dynamic dependency selection.

### Trade-offs

The server package contains more files and developers navigate between composition and route owners. This is accepted because each file has a narrower responsibility and the dependency graph remains explicit.

### Regression tests / protection

Route topology tests protect methods/paths/counts/security-middleware order. Architecture tests prevent coordinators/composition files from registering verbs and route files from constructing PostgreSQL/domain infrastructure.

### Adversarial review findings

The remediation explicitly preserves the mutation authorization middleware as the first handler and avoids rewriting Weather/Metrics/Airport/Historical contexts that already had dedicated boundaries. This limits refactor blast radius and prevents decomposition from becoming an unrelated architecture migration.

### Remediation iterations

Stage 14.8 decomposed the core database composition. Stage 14.15 later applied the same evidence-based responsibility split to Weather composition, demonstrating the pattern was reused only where a concrete cohesion issue existed.

### Residual risks / limitations

File-level separation cannot prevent all conceptual coupling. The server package can grow again as new bounded contexts are added, and architecture tests must evolve with real ownership boundaries.

### Operational / deployment consequences

None. HTTP routes, DTOs, schema, provider behavior, and deployment topology are unchanged.

### Exact evidence

Implementation commit: `e9e9e658958db3ddced2f74d06ab50d0b8034853` (`refactor: decompose server composition root`). Historical PR/reviewer metadata is not asserted where unavailable.

### Final canonical status

**CLOSED.**

### Prevention / future guard

Keep runtime dependency selection explicit inside the server composition root, but split wiring and route topology when one coordinator begins owning independent concerns. Do not introduce DI frameworks merely to reduce file length; require a concrete dependency-selection problem first.
