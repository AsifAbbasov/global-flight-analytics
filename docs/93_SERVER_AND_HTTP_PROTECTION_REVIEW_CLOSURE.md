# Document 93 — Server and HTTP Protection Blocker Closure

Status: Implemented Engineering Contract v1.1
Project: Global Flight Analytics
Baseline: `6b922cbd9df1bff3f880ad120dd883b37f658e53`

## 1. Scope

This increment closes the two accepted release blockers from the Server and
HTTP Protection review:

```text
1. state-changing Route Intelligence POST route required authentication;
2. process liveness was incorrectly used as production readiness while
   PostgreSQL and database-backed routes could be unavailable.
```

This document does not claim that every non-blocking observation from the
original Server review was closed by the same increment. Complete classification
and closure are recorded by Document 94.

## 2. Mutation endpoint protection

The state-changing Route Intelligence route remains protected by the internal
mutation authorization middleware before the request reaches the handler.

The established protection contract remains:

```text
POST /api/v1/trajectories/:id/route-intelligence
X-Internal-API-Key: <raw high-entropy operator key>
```

The backend stores only the configured SHA-256 digest and compares the presented
credential through the existing constant-time authorization boundary.

The Backend Container gate sends the same POST request without credentials and
requires `HTTP 401`. This proves that route registration cannot bypass the
middleware.

## 3. Liveness and readiness separation

The service exposes separate contracts:

```text
GET /api/v1/health
```

This is process liveness only. It proves that the Hypertext Transfer Protocol
process can answer requests.

```text
GET /api/v1/ready
```

This is production readiness. It succeeds only when PostgreSQL is configured
and responds to a bounded ping.

A nil database pool is converted to a nil readiness function before crossing
the handler boundary. This prevents the typed-nil interface failure mode.

Failure is fail-closed:

```text
HTTP 503
SERVICE_NOT_READY
```

Public responses do not expose PostgreSQL driver errors, network addresses,
connection strings, or credentials.

## 4. Container contract

The compiled container healthcheck targets `/api/v1/ready`.

The Backend Container Continuous Integration job:

```text
creates an isolated Docker network;
starts PostgreSQL 16;
waits for pg_isready;
applies the complete production migration catalog;
starts the API with every required database-backed configuration value;
waits for Docker readiness;
verifies /health;
verifies /ready;
verifies that the mutation POST route rejects a missing key with HTTP 401.
```

## 5. Blocker closure statement

The commit containing this increment passed complete Backend Continuous
Integration.

```text
Server review release blockers: CLOSED
Open release blockers: 0
Server review full closure: tracked by Document 94
Release decision: ACCEPTABLE
```
