# Backend Startup Context Hardening

Status: Implementation prepared; exact-commit Continuous Integration closure pending
Project: Global Flight Analytics
Reviewed baseline: `cd99532aa40fe7a61fb19580455ac2d1ba5f650c`
Date: 2026-08-01

## 1. Purpose

This increment closes the remaining PostgreSQL startup ownership gap identified while
reviewing external recommendations. The existing connection timeout was already
bounded, but the API server database path created that timeout from a new background
context. A shutdown signal received during startup therefore could not immediately
cancel the in-flight PostgreSQL initialization.

## 2. Context ownership

The database package now exposes `NewPostgresPoolContext`. It validates a non-nil
caller context, derives the configured connection timeout from that caller, creates the
pool and performs the startup ping within the same bounded context.

The legacy `NewPostgresPool` function remains as a compatibility wrapper for existing
commands and verification tools. This avoids an unrelated repository-wide migration.
The API server uses the context-aware function and passes its signal-owned lifecycle
context through `run`, `openServerDatabase`, pool creation and ping.

## 3. Cancellation semantics

The ownership chain is now:

```text
SIGINT or SIGTERM
→ server lifecycle context
→ PostgreSQL connection timeout child
→ pgxpool creation
→ startup ping
```

The first terminating condition wins. Caller cancellation remains observable through
`errors.Is(err, context.Canceled)`, while the server retains its existing
`SERVER_DATABASE_CONNECTION_FAILED` classification.

## 4. Regression evidence

Focused tests verify:

1. context-aware pool creation rejects a nil caller context;
2. non-positive timeouts retain their explicit classification;
3. an already cancelled caller context is preserved without waiting for the timeout;
4. database-optional server startup remains supported;
5. server startup wraps PostgreSQL cancellation as a database connection failure while
   preserving the original cancellation cause.

The permanent Backend Context Ownership audit, full Go tests, Go vet, project audit
and code review policy audit remain required gates.

## 5. Scope boundary

This increment does not replace pgxpool, change connection-string semantics, alter
migrations, add retries, introduce a circuit breaker, change the server shutdown
contract or migrate every existing command to the context-aware function.
