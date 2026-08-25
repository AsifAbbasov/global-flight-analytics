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

---

## Canonical remediation history

### GFA-SEC-095 — client identity behind reverse proxies lacked an explicit trust contract

1. **Finding / symptom.** The default rate-limit identity used the direct transport peer, which can group unrelated users behind one reverse proxy; blindly switching to forwarded headers would instead permit spoofing.
2. **Root cause.** Proxy-derived client identity had no configuration contract tying accepted forwarding headers to explicitly trusted transport peers/ranges.
3. **Failure scenario.** In a public proxied deployment, either all users share the proxy's limiter bucket, or an unsafe implementation trusts attacker-controlled `X-Forwarded-For`/similar values and lets clients choose their own identity.
4. **Impact.** The first mode harms availability/fairness; the second can bypass per-client rate-limit/logging identity and falsify request attribution.
5. **Severity rationale.** **P1 retrospective.** The deferred issue becomes a security boundary once proxy-derived identity is activated; a wrong trust model can undermine rate limiting and audit attribution.
6. **Existing guarantees violated.** Forwarded identity may be trusted only from verified proxy peers; default configuration must fail closed; rate limiter and logger must use the same resolved identity.
7. **Considered solutions.** Always trust `X-Forwarded-For`; always use direct peer; hardcode Render ranges; accept a configurable header without peer validation; explicit trusted IP/CIDR ranges plus bounded chain parsing.
8. **Chosen remediation.** Add `API_TRUSTED_PROXY_RANGES` and `API_CLIENT_IP_HEADER`; ignore forwarded headers without trusted ranges; reject universal trust ranges; bound ranges/chain length; walk trusted chains right-to-left and fall back to direct peer on malformed/untrusted input.
9. **Why this solution was selected.** It enables correct proxied identity only when deployment supplies verified trust inputs and avoids guessing hosting-provider topology in source code.
10. **Rejected alternatives.** Blind forwarded-header trust is spoofable; hardcoded provider ranges become stale/non-portable; direct-peer-only remains safe but unusable for fair per-client identity behind shared proxies.
11. **Trade-offs.** Deployment must supply/maintain trusted proxy ranges; until then, users behind one proxy may intentionally share a bucket rather than weaken security.
12. **Regression tests / protection.** Trusted/untrusted peer tests, spoofed-header rejection, malformed-chain fallback, range/header config limits, rate-limit bucket integration and request-log identity tests.
13. **Adversarial review findings.** `0.0.0.0/0` and `::/0` trust must be rejected; malformed chain items fail back to direct peer; chain length and trusted-range count are bounded; one resolver feeds both logging and limiting.
14. **Remediation iterations.** The original Document 94 finding was deliberately deferred until a deployable trust model existed; Document 95 closes code debt without inventing hosting ranges.
15. **Residual risks and limitations.** Correct production behavior still depends on operator-supplied ranges matching the actual reverse-proxy path; that environment input is not derivable safely from source alone.
16. **Operational or deployment consequences.** Proxy-derived identity remains disabled until trusted ranges are configured; deployment runbooks must verify those ranges before activation.
17. **Exact evidence.** Historical implementation commit `cfb079b6f881b03b517f92b06210c3fdc9968893` (`fix: close trusted proxy and build metadata debt`). Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-SEC-095=CLOSED` for application code; verified proxy ranges remain an environment-specific activation input.
19. **Prevention / future guard.** Forwarded client identity may never be enabled by header name alone; trust requires verified peer/range configuration and spoofing/chain regression tests.

### GFA-GOV-096 — application version endpoint lacked build-derived revision provenance

1. **Finding / symptom.** Application version information was hardcoded and did not reliably identify the exact build revision or build time serving a deployment.
2. **Root cause.** Version diagnostics were source constants rather than build-owned metadata injected into the server binary/container.
3. **Failure scenario.** Operators query `/version` during an incident and cannot distinguish which commit/image is actually running, or a stale hardcoded version implies provenance that the binary does not prove.
4. **Impact.** Release/deployment diagnosis and evidence binding are weaker; mutable deployment aliases cannot be tied confidently to source revision.
5. **Severity rationale.** **P2 retrospective.** This is release/observability provenance debt rather than application-data corruption, but exact revision identity is important for production evidence and rollback.
6. **Existing guarantees violated.** Deployment diagnostics must distinguish version, VCS revision and build time; invalid/absent build metadata must not masquerade as exact provenance.
7. **Considered solutions.** Keep hardcoded version; read Git at runtime; fetch deployment metadata externally; inject linker values during build and mirror them in OCI labels.
8. **Chosen remediation.** Define linker-owned `version`, `revision`, `built_at`; Docker accepts `APP_VERSION`, `VCS_REF`, `BUILD_DATE`; `/version` exposes the validated values and image labels mirror them, with conservative `unknown` defaults for local development.
9. **Why this solution was selected.** Build-time injection is deterministic, works in scratch/container runtime without `.git`, and makes binary plus image carry the same provenance.
10. **Rejected alternatives.** Runtime Git is unavailable/unreliable in production images; hardcoded values drift; external-only metadata cannot prove what the binary itself reports.
11. **Trade-offs.** Build systems must supply correct metadata; local development intentionally reports conservative unknown revision/build time.
12. **Regression tests / protection.** Linker injection, version endpoint, Docker image label and container endpoint metadata checks.
13. **Adversarial review findings.** Empty/malformed build metadata must fall back to documented development defaults rather than expose partial false provenance; only the server binary needs these linker values.
14. **Remediation iterations.** This item was a non-blocking/deferred release-engineering concern in Document 94 and was closed before public tagged/revision-aware deployment evidence relied on it.
15. **Residual risks and limitations.** Build provenance is only as trustworthy as the CI/build inputs; later release-truth documents separately distinguish repository HEAD from observed deployment revision.
16. **Operational or deployment consequences.** Container builds should supply exact revision/build timestamp; `/api/v1/version` becomes a release smoke/evidence endpoint.
17. **Exact evidence.** Historical implementation commit `cfb079b6f881b03b517f92b06210c3fdc9968893`. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-GOV-096=CLOSED`.
19. **Prevention / future guard.** Production builds must keep binary metadata and OCI labels aligned; deployment verification must compare expected revision to the served `/version` result rather than assuming mutable aliases.

### GFA-TEST-097 — container PostgreSQL readiness could pass during bootstrap handoff rather than final server readiness

1. **Finding / symptom.** `pg_isready` alone could succeed while the official PostgreSQL image was still in its temporary bootstrap-server phase, before the final process owned the requested database runtime.
2. **Root cause.** Container CI treated one protocol-level readiness signal as sufficient without accounting for the image's initialization lifecycle and process handoff.
3. **Failure scenario.** The smoke workflow observes `pg_isready` from the bootstrap server, starts migrations, and races the bootstrap shutdown/final PostgreSQL startup transition.
4. **Impact.** CI/container verification can become flaky or falsely attribute migration/database failures to application changes.
5. **Severity rationale.** **P2 retrospective.** The defect affects release evidence trust and deterministic container validation rather than production schema semantics directly.
6. **Existing guarantees violated.** Database readiness for migration tests must prove the final PostgreSQL process and successful access to the target database.
7. **Considered solutions.** Add sleep; retry `pg_isready`; inspect only PID 1; require both final process identity and target-database `SELECT 1`.
8. **Chosen remediation.** Require `/proc/1/comm == postgres` and a successful `SELECT 1` in the requested database before migrations/health smoke proceed.
9. **Why this solution was selected.** It verifies both lifecycle ownership and actual target-database usability without arbitrary timing delays.
10. **Rejected alternatives.** Sleeps are timing-dependent; repeated `pg_isready` can still hit bootstrap; process identity alone does not prove target database creation/connectivity.
11. **Trade-offs.** Container CI performs an additional process/database check but becomes deterministic around first-time initialization.
12. **Regression tests / protection.** Backend Container script permanently checks final process plus target-database readiness before applying migrations and starting API smoke.
13. **Adversarial review findings.** Both conditions are required; either one alone leaves a handoff or database-selection gap.
14. **Remediation iterations.** Follow-up commit `ae4d486d2341974a47173e2aedd78da530130cf6` strengthened the initial trusted-proxy/build metadata closure after the bootstrap handoff behavior was observed.
15. **Residual risks and limitations.** Docker/runtime infrastructure failures outside PostgreSQL lifecycle can still fail CI; this guard specifically removes false readiness during image initialization.
16. **Operational or deployment consequences.** CI waits for real final PostgreSQL readiness rather than adding a fixed delay; no application runtime behavior changes.
17. **Exact evidence.** Historical follow-up commit `ae4d486d2341974a47173e2aedd78da530130cf6` (`fix: wait for final postgres container readiness`), following `cfb079b6f881b03b517f92b06210c3fdc9968893`. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-TEST-097=CLOSED`.
19. **Prevention / future guard.** Containerized dependency smoke tests must validate the dependency's final runtime state and target resource, not rely on a transient bootstrap readiness probe alone.
