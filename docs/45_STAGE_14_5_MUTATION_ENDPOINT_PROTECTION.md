# Document 45 — Stage 14.5 Mutation Endpoint Protection

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: fail-closed authorization for state-changing and computation-triggering HTTP routes

## 1. Security Boundary

Public read-only routes remain unauthenticated:

```text
GET
HEAD
OPTIONS
```

Every route registered through:

```text
POST
PUT
PATCH
DELETE
```

must use the internal mutation authorization middleware as its first route
middleware.

The current protected route is:

```text
POST /api/v1/trajectories/:id/route-intelligence
```

This route triggers calculation and PostgreSQL persistence and is therefore
not treated as a public read operation.

## 2. Credential Storage

The server does not store the raw internal API key.

Deployment configuration contains only:

```text
API_MUTATION_KEY_SHA256
```

This value is the 64-character hexadecimal SHA-256 digest of a high-entropy
key held by an internal operator or trusted automation client.

Example local generation:

```bash
KEY="$(openssl rand -hex 32)"
DIGEST="$(printf '%s' "$KEY" | shasum -a 256 | awk '{print $1}')"
```

Store `DIGEST` as `API_MUTATION_KEY_SHA256`.

Store `KEY` only in the trusted caller's secret storage.

Never place either value in source control or frontend environment variables.

## 3. Request Contract

Trusted callers send the raw key through:

```text
X-Internal-API-Key
```

The server hashes the presented key and compares the digest using a
constant-time comparison.

Missing and invalid keys return the same response:

```text
HTTP 401
MUTATION_AUTHENTICATION_REQUIRED
```

A database-backed server without configured mutation credentials fails during
configuration loading.

A directly composed test or diagnostic server with an unconfigured
authorizer returns:

```text
HTTP 503
MUTATION_AUTHENTICATION_UNAVAILABLE
```

The response is marked:

```text
Cache-Control: no-store
```

## 4. Frontend Separation

The public Next.js application must not contain:

```text
X-Internal-API-Key
API_MUTATION_KEY_SHA256
```

The mutation credential header is intentionally not included in the public
CORS allowlist.

The browser frontend therefore remains a read-only client.

## 5. Architecture Gate

`projectaudit -mode security -strict` scans every production Go source file in
`internal/server`.

Every `Post`, `Put`, `Patch`, or `Delete` route must have:

```text
mutationAuthorization
```

as its first route middleware.

The audit also scans frontend source files and fails if mutation credential
identifiers appear there.

This gate runs as part of `projectaudit -mode all -strict`.

## 6. Rotation

Credential rotation requires:

```text
1. Generate a new high-entropy raw key.
2. Compute its SHA-256 digest.
3. Replace API_MUTATION_KEY_SHA256 in backend secret configuration.
4. Restart or redeploy the backend.
5. Replace the raw key in the trusted caller.
6. Revoke the old raw key.
```

Only one active digest is supported in the current minimal infrastructure.

## 7. Limitations

This internal key is not user authentication.

It does not provide:

```text
user accounts
roles
per-user authorization
session management
audit identity for multiple administrators
```

Those capabilities are unnecessary for the current read-only public product
and must not be simulated with frontend secrets.
