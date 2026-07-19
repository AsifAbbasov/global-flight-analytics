# Document 49 — Stage 14.9 HTTP Query and Contract Boundary Hardening

Status: Implementation Baseline v1.0
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
