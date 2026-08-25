# Document 55 — Stage 14.15 Weather Composition Boundary

Status: Remediation History v1.1
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

Existing error wrapping is preserved.

### 3.3 Application Composition

`weather_application_composition.go` owns PostgreSQL weather repository, weather application service, and weather HTTP handler. It does not register routes or construct external providers.

### 3.4 Route Registration

`weather_route_registration.go` owns only `CurrentWeatherPath`, router and handler validation, and Fiber GET route registration. The route remains `GET /api/v1/weather/current`.

## 4. Preserved Runtime Behavior

This increment does not change Open-Meteo URL/timeout behavior, provider policy, budget accounting, response observation, request coalescing, request keys, coordinate validation, PostgreSQL persistence, HTTP contracts, database schema, migrations, or frontend behavior.

## 5. Why Same-Package Decomposition

The components remain inside the `server` package because this is composition code, not a new domain. Creating exported packages or interfaces solely to reduce file length would add artificial architecture.

## 6. Testing Strategy

Automated tests verify composition error preservation, network-free graph construction, route preservation, dependency propagation, nil dependency rejection, and static responsibility boundaries between coordinator/provider/application/registration files.

## 7. Acceptance

Acceptance requires focused Weather tests, architecture regression tests, race detector, strict project audit, complete Go build/vet/tests, frontend dependency checks/lint/typecheck/build, backend container build, and diff checks.

## 8. Canonical remediation history

### Finding / symptom

`registerWeatherRoute` concentrated provider governance, external integration, persistence, application construction, handler creation and HTTP registration in one function despite its route-registration name.

### Root cause

Production wiring grew incrementally inside the original route function without a responsibility boundary between provider composition, application composition and HTTP registration.

### Failure scenario

```text
provider or persistence change is needed
↓
engineer edits route-registration function
↓
unrelated provider/application/HTTP responsibilities share one review surface
↓
future changes can accidentally alter route behavior while modifying infrastructure wiring
```

### Impact

The code was functionally correct, so this was primarily a maintainability, reviewability and regression-isolation problem. It increased coupling and made focused tests harder rather than producing a known incorrect response.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P3 maintainability** because behavior was already correct and the remediation reduced future change risk rather than repairing a current data/security failure.

### Existing guarantees violated

```text
composition root responsibilities should remain inspectable
HTTP registration should not own provider/persistence construction
bounded-context wiring changes should have narrow review surfaces
simple architecture should be preferred over artificial abstraction layers
```

### Considered solutions

1. Keep the function unchanged.
2. Introduce a dependency-injection framework or exported composition packages.
3. Split responsibility-specific files/functions inside the existing `server` package.

### Chosen remediation

Option 3: same-package decomposition into route coordinator, provider composition, application composition and route registration boundaries.

### Why this solution was selected

It improves cohesion and testability while preserving explicit concrete wiring and avoiding new runtime abstractions.

### Rejected alternatives

Keeping the monolithic function was rejected because the confirmed cohesion problem would remain. DI frameworks/service locators/new exported packages were rejected as disproportionate complexity for a single composition root.

### Trade-offs

```text
+ narrower change/review surfaces
+ focused responsibility tests
+ explicit dependency graph remains visible
- more source files in internal/server
- file boundaries require architecture checks to prevent gradual recombination
```

### Regression tests / protection

Architecture tests prohibit provider/application constructors in route registration, route registration in provider/application files, and dependency construction in the coordinator. Behavioral tests preserve timeout errors and the GET route contract.

### Adversarial review findings

Review deliberately rejected a mechanical line-count solution and retained same-package composition. It also preserved the existing nil-Pool construction behavior rather than silently broadening this structural refactor into startup-semantics changes.

Historical PR/reviewer evidence for the original remediation is unavailable; reconstruction is limited to repository source, tests, commits, and this stage document.

### Remediation iterations

The responsibility split landed in `cd5e3540cd4f849f606c50433f4e033548b59002` (`refactor: separate weather composition boundaries`). Later backend-final auditing made these boundaries permanent source-level invariants.

### Residual risks and limitations

Same-package decomposition does not prevent all future coupling. Constructors may still grow within their proper owner files, and startup semantics such as nil-Pool acceptance remain separate concerns.

### Operational or deployment consequences

None by design: HTTP method/path, provider behavior, persistence, schema and deployment configuration remain unchanged. The change affects source ownership and regression protection only.

### Exact evidence

```text
implementation commit:
cd5e3540cd4f849f606c50433f4e033548b59002

permanent evidence:
focused Weather server/provider/service/repository/HTTP tests
architecture boundary tests
backend final correctness audit
```

### Final canonical status

```text
FINDING_GFA_MAINT_053_WEATHER_COMPOSITION_CONCENTRATION=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/55_STAGE_14_15_WEATHER_COMPOSITION_BOUNDARY.md
IMPLEMENTATION_COMMIT=cd5e3540cd4f849f606c50433f4e033548b59002
```

### Prevention / future guard

Composition changes must preserve separation between provider governance, application/persistence construction and HTTP registration. New abstraction layers require a concrete need; file-count or function-length pressure alone is not sufficient justification.
