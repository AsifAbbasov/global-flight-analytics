# Document 49 — Stage 14.9 HTTP Query and Contract Boundary Hardening

Status: Remediation History v1.1
Project: Global Flight Analytics
Scope: Historical Intelligence HTTP query intent and PostgreSQL dependency boundary

## 1. Confirmed Problems

The Historical Intelligence HTTP handler used one parser with a boolean mode
argument:

```text
parseHistoricalIntelligenceQuery(values, false)
parseHistoricalIntelligenceQuery(values, true)
```

The boolean changed whether history pagination was parsed.

The HTTP handler and DTO layer also imported the concrete
`historicalaggregate` package. That package contains the PostgreSQL adapter.

The handler additionally knew about:

```text
pgx.ErrNoRows
ErrPostgresPoolRequired
ErrPostgresExecutorRequired
```

Those are infrastructure details and must not participate in HTTP behavior.

## 2. Query Intent Resolution

The boolean mode is replaced by explicit entry points:

```text
parseHistoricalLatestQuery
parseHistoricalHistoryQuery
```

Shared metric, scope, and granularity parsing remains in:

```text
parseHistoricalQueryBase
```

Only the history parser reads `limit` and `before_window_end`.

The latest parser cannot accidentally activate pagination behavior.

## 3. Pure Store Contract

A new package contains the database-independent persistence contract:

```text
internal/historicalintelligence/historicalaggregatecontract
```

It owns:

```text
ResultKey
Record
ListQuery
Page
Store
semantic store errors
ValidationError
list limit policy
```

It depends only on the standard library and `historicalcontract`.

It does not import:

```text
pgx
pgxpool
Fiber
repository implementations
external integrations
```

## 4. Compatibility

The existing `historicalaggregate` implementation package exposes type aliases
to the pure contract.

Therefore existing materializers, verification commands, tests, and
PostgreSQL code keep their source-compatible public types.

The PostgreSQL-only declarations are moved from `contracts.go` into
`postgres_contracts.go`.

## 5. HTTP Error Boundary

The HTTP handler recognizes only semantic contract errors:

```text
result not found
invalid scope
invalid list limit
context deadline
context cancellation
```

The PostgreSQL adapter is responsible for converting `pgx.ErrNoRows` into the
semantic `ErrResultNotFound`.

PostgreSQL construction errors are handled during server composition, not in
request processing.

## 6. Regression Gates

Automated tests ensure:

```text
HTTP handler and DTO do not import historicalaggregate implementation
HTTP handler and DTO do not import pgx
pure aggregate contract has no infrastructure imports
historical query parsers have no boolean mode parameters
latest query remains pagination-free
history query validates pagination
contract clone operations preserve ownership boundaries
```

## 7. Intentionally Rejected Changes

The increment does not move the complete PostgreSQL store implementation to a
new package.

The compatibility alias boundary already prevents high-level HTTP modules from
depending on that implementation, while avoiding a broad migration of
materializers and verification commands in the same increment.

The increment does not rename every function containing `With` or `And`.
Naming changes are accepted only when they clarify responsibility.

The increment does not replace domain-state booleans or optional floating point
values without evidence that additional states are required.

## 8. Acceptance

The increment is accepted only after:

```text
focused contract and handler tests
targeted race tests
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

## 9. Canonical finding decomposition

```text
GFA-MAINT-043  boolean mode argument obscuring Historical query intent
GFA-ARCH-044   HTTP/DTO dependency on PostgreSQL implementation and pgx errors
```

## 10. GFA-MAINT-043 — Boolean mode argument obscuring Historical query intent

### Finding / symptom

One parser accepted a boolean whose value changed the meaning of the operation by enabling/disabling history pagination:

```text
parseHistoricalIntelligenceQuery(values, false)
parseHistoricalIntelligenceQuery(values, true)
```

Call sites did not communicate whether they were parsing a latest-result query or a history-list query without inspecting the parser implementation.

### Root cause

Two HTTP operations shared most query fields and were initially represented as one parser with a mode flag instead of distinct semantic entry points.

### Failure scenario

A new call site passes the wrong boolean or the shared parser gains pagination behavior that unintentionally leaks into the latest endpoint. The code compiles because `true/false` is type-correct, while the operation intent is wrong.

### Impact

The main risk is maintainability and future query-contract regression: hidden mode semantics make review harder and allow accidental activation of pagination-only rules on the wrong endpoint.

### Severity rationale

**P3 retrospective.** The historical evidence does not assert a production pagination bug from the flag; it identifies an ambiguous control-flow contract that made such regressions easier.

### Existing guarantees violated

- public operation intent should be explicit at the call site;
- latest-result queries must not accidentally parse list pagination;
- parser APIs should make invalid operation combinations difficult to express.

### Considered solutions

1. keep the boolean and improve comments/naming;
2. introduce a general parser-options struct with mode fields;
3. expose explicit latest/history entry points sharing a private base parser.

### Chosen remediation

`parseHistoricalLatestQuery` and `parseHistoricalHistoryQuery` represent the two operations explicitly. Shared metric/scope/granularity parsing remains in `parseHistoricalQueryBase`; only history parsing owns pagination.

### Why selected

The solution removes the ambiguous control flag without duplicating common parsing or introducing a generalized options abstraction for only two known operations.

### Rejected alternatives

Comments do not prevent reversed booleans. A broad options struct would preserve indirect mode selection and add more abstraction than the problem requires.

### Trade-offs

There are more named parser functions, but each has one obvious semantic purpose. Shared logic remains centralized.

### Regression tests / protection

Tests forbid boolean mode parameters, require the latest parser to remain pagination-free, and verify history pagination validation.

### Adversarial review findings

The remediation deliberately avoids a repository-wide campaign against all booleans or functions containing `With`/`And`; only a boolean that actually encoded operation mode was changed.

### Remediation iterations

This was resolved in one behavior-preserving Stage 14.9 refactor after the Stage 14 audit identified hidden query intent.

### Residual risks / limitations

Future Historical query operations may require a third explicit parser or a better shared request model. The current design should not be stretched into a growing collection of mode flags.

### Operational / deployment consequences

None. HTTP methods/paths and valid query behavior remain unchanged.

### Exact evidence

Implementation commit: `2842d09fc2eb0dcc746a28dd126611fba0f2d1a8` (`refactor: harden historical contract boundary`). Historical PR/reviewer metadata is not invented when unavailable.

### Final canonical status

**CLOSED.**

### Prevention / future guard

Do not encode materially different HTTP operation semantics as anonymous booleans. Prefer explicit operation entry points or typed request modes when the distinction affects validation, pagination, security, or persistence behavior.

## 11. GFA-ARCH-044 — HTTP/DTO dependency on PostgreSQL implementation and pgx errors

### Finding / symptom

Historical Intelligence HTTP/DTO code imported the concrete `historicalaggregate` implementation package, which contained PostgreSQL code, and the handler directly recognized `pgx.ErrNoRows`, `ErrPostgresPoolRequired`, and `ErrPostgresExecutorRequired`.

### Root cause

The persistence contract and PostgreSQL implementation originally lived in the same package. High-level HTTP code therefore depended on implementation-level types/errors because no database-independent store contract boundary existed.

### Failure scenario

A PostgreSQL adapter error, constructor concern, or implementation refactor forces changes in HTTP handlers/DTOs. A future alternate store becomes harder to introduce, and request behavior can accidentally branch on low-level driver details rather than semantic domain/store outcomes.

### Impact

The dependency inversion is reversed: transport code becomes coupled to PostgreSQL/pgx. This increases change blast radius and risks leaking infrastructure semantics into public HTTP error behavior.

### Severity rationale

**P2 retrospective.** This is an architecture/contract-boundary defect on a production HTTP path. No incorrect public response is asserted, but direct driver dependency can make error behavior and implementation changes unsafe.

### Existing guarantees violated

- HTTP/DTO layers should depend on semantic contracts rather than database adapters;
- `pgx.ErrNoRows` must be translated at the PostgreSQL boundary;
- database construction failures belong to server composition, not request handling;
- pure store contracts must not import pgx/Fiber/repository implementations.

### Considered solutions

1. keep implementation imports and wrap driver errors in handlers;
2. move the entire PostgreSQL store to a new package in one large migration;
3. extract a pure `historicalaggregatecontract` package and keep compatibility aliases in the implementation package.

### Chosen remediation

The new pure contract package owns store types, semantic errors, validation, and list limits. PostgreSQL-only declarations remain in the implementation package, which aliases the pure types for source compatibility. The adapter translates `pgx.ErrNoRows` to `ErrResultNotFound`; composition handles construction errors.

### Why selected

This corrects dependency direction immediately while avoiding a broad migration of materializers, verification commands, and PostgreSQL code in the same patch.

### Rejected alternatives

Handler-level driver wrapping preserves the wrong dependency. Moving the whole store was rejected as unnecessarily large when compatibility aliases could establish the desired high-level boundary with lower migration risk.

### Trade-offs

The repository temporarily maintains an implementation package with compatibility aliases to the pure contract. This indirection is intentional to reduce refactor blast radius and can be revisited only if it creates measured maintenance cost.

### Regression tests / protection

Architecture tests forbid `historicalaggregate` implementation and `pgx` imports from HTTP handler/DTO layers, require the pure contract to remain infrastructure-free, and verify semantic error translation.

### Adversarial review findings

The stage explicitly rejected unrelated large-scale cleanup: moving every implementation file, renaming functions mechanically, or replacing unrelated booleans/optional floats without evidence. This keeps the boundary fix narrow.

### Remediation iterations

Stage 14.9 establishes the pure contract/alias bridge. Stage 14.14 later strengthens Historical pagination semantics on top of this contract rather than reintroducing PostgreSQL knowledge into HTTP.

### Residual risks / limitations

The implementation package still exists and source compatibility aliases create a transitional dependency surface for lower-level callers. The key invariant is that high-level HTTP/DTO code cannot depend on it.

### Operational / deployment consequences

None. Public HTTP contracts and PostgreSQL schema remain unchanged; server composition continues to choose the PostgreSQL implementation.

### Exact evidence

Implementation commit: `2842d09fc2eb0dcc746a28dd126611fba0f2d1a8`.

### Final canonical status

**CLOSED for the Historical HTTP/store dependency boundary.**

### Prevention / future guard

Transport/application layers must consume semantic store/service contracts. Database adapters own translation from driver-specific errors/types, while construction errors terminate composition/startup rather than becoming request-time branches.
