# Document 45 — Stage 14.5 Mutation Endpoint Protection

Status: Remediation History v1.1
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

## 8. Canonical finding record — GFA-SEC-039

### Finding / symptom

A production HTTP route using `POST` triggered Route Intelligence computation and PostgreSQL persistence without a repository-wide fail-closed rule requiring authorization on mutation/computation-triggering routes.

### Root cause

The public product was primarily read-only, and early HTTP composition treated route registration mostly as functional wiring. A dedicated security invariant distinguishing public reads from state-changing/computation-triggering operations had not yet been encoded structurally.

### Failure scenario

An unauthenticated external caller reaches `POST /api/v1/trajectories/:id/route-intelligence`, triggers non-trivial computation and persistence repeatedly, or a future developer adds another POST/PUT/PATCH/DELETE route without security middleware because nothing in architecture CI requires it.

### Impact

Unauthorized callers could consume backend/database resources and cause state-changing analytical materialization. The missing invariant also created a repeatable class of future authorization omissions.

### Severity rationale

**P1 retrospective.** This is an externally reachable authorization boundary on a route that triggers computation and persistence. The evidence does not claim private user data exposure, but unauthenticated mutation capability is high-impact enough to be treated as P1.

### Existing guarantees violated

- public anonymous access is limited to read-only open research data;
- state-changing/computation-triggering routes require trusted-caller authorization;
- mutation secrets must never be delivered to the browser;
- missing backend credentials must fail closed rather than silently disable protection.

### Considered solutions

1. leave mutation routes public because the data is public;
2. add user-account/session authentication to the whole product;
3. use a minimal internal high-entropy API key for trusted mutation callers and enforce middleware structurally on every mutation verb.

### Chosen remediation

Every POST/PUT/PATCH/DELETE route must place `mutationAuthorization` first. The backend stores only `API_MUTATION_KEY_SHA256`, hashes the presented `X-Internal-API-Key`, compares in constant time, fails startup/request handling closed when unavailable, and keeps all credential identifiers out of frontend source/CORS.

### Why this solution was selected

The product needs anonymous public reads but not full user identity/roles. A single internal credential matches the current trusted-automation/operator use case without introducing user auth infrastructure unrelated to the MVP.

### Rejected alternatives

Public mutations were rejected because openness of source data does not authorize arbitrary computation/persistence. Full user accounts/RBAC were rejected as unnecessary complexity for the current product. Frontend-held secrets were rejected because browser code cannot safely preserve server secrets.

### Trade-offs

Only one active digest is supported, so rotation requires coordinated backend/trusted-caller change and restart/redeploy. The design does not provide per-operator attribution. These limitations are accepted in exchange for a small fail-closed boundary.

### Regression tests / protection

`projectaudit -mode security -strict` scans all server production source and requires `mutationAuthorization` as first middleware for every mutation verb. It also scans frontend source for secret identifiers. Handler/config tests cover missing/invalid/equal-response behavior, no-store responses, and fail-closed configuration.

### Adversarial review findings

The review identified two easy but unsafe shortcuts and explicitly blocked them: exposing the raw/digest credential to Next.js, and treating computation-triggering POST as equivalent to a public GET merely because the underlying data is open.

### Remediation iterations

The Stage 14.1 architecture foundation first documented that mutation routes must be protected. Stage 14.5 implemented the concrete middleware, secret-storage, frontend-separation, and source-audit rules.

### Residual risks / limitations

A single shared key cannot identify individual operators, cannot express roles, and does not replace rate limiting or broader abuse controls. Compromise of the trusted raw key grants the mutation capability until rotated.

### Operational / deployment consequences

Database-backed production configuration must provide `API_MUTATION_KEY_SHA256`. Trusted automation must store the raw key separately and send `X-Internal-API-Key`. Rotation requires backend redeploy/restart and caller secret update.

### Exact evidence

Implementation commit: `50831ae06cb1a38c321ec8c7766bc1f28ddb5757` (`feat: protect mutation endpoints`). Historical PR/reviewer metadata is not fabricated where unavailable.

### Final canonical status

**CLOSED for the current internal mutation authorization boundary.**

### Prevention / future guard

Mutation authorization is verb-driven and structural: every new POST/PUT/PATCH/DELETE route must fail CI unless the approved authorization middleware is first. Public frontend source must remain free of internal credential names/values, and new privileged capabilities require a separately reviewed identity model rather than extending the shared key beyond its intended scope.
