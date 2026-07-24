# Document 95 — Trusted Proxy and Build Metadata Closure

Status: Implemented Engineering Contract v1.0
Project: Global Flight Analytics
Baseline: `2573892ad7684f3d2646378e2021638a53173bc3`

## 1. Purpose

This increment resolves the two classified post-review Server debts recorded by
Document 94:

```text
trusted proxy client identity for rate limiting;
build-derived application version and provenance.
```

## 2. Trusted proxy client identity

The default remains fail-closed:

```text
no trusted proxy ranges configured
→ forwarded client headers ignored
→ direct transport peer used as rate-limit identity
```

Proxy-derived identity is enabled only when
`API_TRUSTED_PROXY_RANGES` contains explicit IP addresses or CIDR ranges.

Supported identity headers are:

```text
X-Forwarded-For
X-Real-IP
CF-Connecting-IP
```

`X-Forwarded-For` is the default when trusted ranges are configured and
`API_CLIENT_IP_HEADER` is omitted.

The resolver:

```text
rejects an identity header without trusted proxy ranges;
rejects 0.0.0.0/0 and ::/0 trust;
accepts at most 64 trusted ranges;
accepts at most 32 forwarded chain entries;
ignores the header when the transport peer is not trusted;
fails back to the transport peer when any chain item is malformed;
walks a valid trusted chain from right to left;
uses the first non-trusted hop as the client identity.
```

The same resolved identity is used by the rate limiter and the request logger.

The transport-peer lookup is an explicit function boundary. Production uses the
Fiber transport peer. Tests inject a deterministic peer resolver instead of
assuming that `httptest.Request.RemoteAddr` is transferred through Fiber's
in-memory test adapter.

No Render or hosting-platform proxy range is guessed in source code. Production
activation requires ranges confirmed for the selected deployment path. Until
then, direct peer identity remains safe even if multiple clients share one proxy
bucket.

## 3. Build-derived version endpoint

The server build owns three linker values:

```text
version
revision
built_at
```

Local development uses explicit conservative defaults:

```text
version = 1.0.0
revision = unknown
built_at = unknown
```

The Docker build accepts:

```text
APP_VERSION
VCS_REF
BUILD_DATE
```

Only the server binary receives those linker values. The runtime image also
publishes matching Open Container Initiative labels.

`GET /api/v1/version` now returns:

```json
{
  "success": true,
  "data": {
    "version": "ci-123.1",
    "revision": "<git commit>",
    "built_at": "2026-07-24T00:00:00Z"
  }
}
```

Invalid or empty build metadata fails closed to the documented development
defaults instead of exposing malformed provenance.

## 4. Verification

The permanent evidence includes:

```text
trusted and untrusted proxy resolution tests;
spoofed header rejection;
malformed chain fallback;
range and header configuration tests;
rate-limit bucket integration tests;
request-log identity test;
linker injection test;
version endpoint test;
Docker image label checks;
container endpoint metadata checks;
final PostgreSQL container process and target-database readiness verification;
full backend tests;
race detection;
static analysis;
architecture and code-review audits.
```

## 5. Container Continuous Integration database readiness

The container smoke test does not use `pg_isready` as sufficient evidence during
first-time PostgreSQL initialization. The official image temporarily starts a
bootstrap server, creates the requested database, stops that server, and then
executes the final PostgreSQL process.

The permanent readiness contract therefore requires both:

```text
/proc/1/comm == postgres
SELECT 1 succeeds in the requested database
```

This prevents migrations from starting in the handoff window between the
temporary bootstrap server and the final PostgreSQL process.

## 6. Closure statement

After all local gates and the four Backend Continuous Integration jobs pass on
the same commit:

```text
Trusted proxy code debt: CLOSED
Build-derived version debt: CLOSED
Deferred Server review code findings: 0
Unclassified Server review findings: 0
Release blockers: 0
```

Deployment still owns one environment-specific activation action: configure
verified proxy ranges before enabling proxy-derived identity. This is a
deployment input, not unimplemented application behavior.
