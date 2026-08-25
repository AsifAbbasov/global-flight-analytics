# OpenAPI Developer Experience

Status: CANONICAL RECONCILIATION COMPLETE  
Baseline: `b1f2b23a4f4bb72153d0993417e1888d538cfa75`  
Initial Developer Experience merge: `64400024cddad5e856d4e72433457b4bd2eacb11`  
Conditional-request remediation PR: #54  
Remediation head: `837a97b8696333fcbe8c7a50d76b3f76521a149b`  
Remediation merge: `855f82bf97cf0db47d1a3918f75ea70f7f2b06fe`  
Scope: embedded API documentation, deterministic TypeScript client generation, drift prevention, Continuous Integration enforcement, and production conditional-request correctness

## Purpose

The source-backed OpenAPI contract is complete at 38 public operations. This increment turns that contract into a usable developer surface without changing the production API operation inventory.

The implementation provides:

- an embedded documentation application at `/api/docs`;
- the exact OpenAPI 3.1 contract at `/api/docs/openapi.json`;
- same-binary HTML, JavaScript, CSS, and OpenAPI assets with no CDN dependency;
- a deterministic TypeScript client package generated from `openapi/openapi.json`;
- byte-level embedded-spec drift detection;
- generated-client drift detection;
- generated-client runtime tests for request construction, protected authentication, typed errors, and base-URL guards;
- Go, Node, TypeScript, workflow, release, and documentation verification.

The original Developer Experience implementation is feature/contract evidence and does not receive a synthetic finding ID. The later production-discovered weak-ETag behavior is a separate remediation finding owned by `GFA-CONTRACT-450` below.

## Route boundary

The developer routes are registered directly on the Fiber application:

```text
GET /api/docs
GET /api/docs/
GET /api/docs/openapi.json
GET /api/docs/assets/app.js
GET /api/docs/assets/app.css
```

They are intentionally outside `/api/v1`.

Therefore:

```text
SOURCE_PUBLIC_OPERATIONS=38
OPENAPI_DOCUMENTED_OPERATIONS=38
```

remain unchanged. The documentation transport is not counted as a production domain API operation and is not recursively documented inside the public OpenAPI contract.

## Embedded runtime contract

`apps/api/internal/http/apidocs` embeds:

- the exact canonical OpenAPI JSON;
- the documentation HTML shell;
- the application JavaScript;
- the stylesheet.

The assets are compiled into the Go API binary. Runtime containers do not need a repository checkout, Node.js, pnpm, a static file volume, or an external documentation CDN.

The origin OpenAPI JSON and static assets expose a strong SHA-256 ETag and support `If-None-Match`. Production intermediaries may legally weaken that validator to `W/"<sha256>"`; conditional requests therefore use weak entity-tag comparison while retaining exact opaque SHA-256 identity. The HTML shell uses `Cache-Control: no-store` while immutable build assets use bounded public caching.

## Security boundary

The repository-wide security middleware defaults to a deny-all Content Security Policy. The documentation handler replaces that header only for its own responses with a narrow policy:

```text
default-src 'none'
base-uri 'none'
frame-ancestors 'none'
script-src 'self'
style-src 'self'
connect-src 'self'
img-src 'self' data:
form-action 'none'
```

Additional rules:

- the documentation surface is marked `X-Robots-Tag: noindex, nofollow`;
- the browser explorer executes only same-origin public GET operations;
- the protected Route Intelligence POST is visible but disabled;
- the browser does not request, persist, or read an internal API key;
- the browser does not use local storage, session storage, or cookies for credentials;
- mutation curl output uses only the placeholder `<trusted-server-key>`.

## TypeScript client

The workspace package is:

```text
packages/api-client
```

The package exports:

- `GlobalFlightAnalyticsClient`;
- `APIError`;
- `OperationId`;
- `OperationParameters`;
- `OperationResponses`;
- generated operation definitions;
- generated component schema types.

The runtime client:

- accepts an explicit absolute HTTP or HTTPS base URL;
- rejects credentials embedded in the base URL;
- supports caller-owned cancellation through `AbortSignal`;
- builds path, query, and header parameters from generated metadata;
- parses typed JSON success responses;
- preserves HTTP status, API error code, request ID, and details on failure;
- attaches `X-Internal-API-Key` only when the selected generated operation is protected;
- never persists credentials.

## Generation contract

Generation command:

```text
pnpm run generate:openapi-client
```

Drift check:

```text
node scripts/generate-openapi-client.mjs --check
```

The generator supports the JSON Schema forms used by the repository contract:

- local schema references;
- objects with required and optional properties;
- arrays;
- enums and constants;
- `oneOf`, `anyOf`, and `allOf`;
- nullable type arrays;
- additional properties;
- operation path, query, header, and JSON body parameters;
- typed first-success JSON responses;
- protected-operation metadata.

The generated file includes the exact SHA-256 of `openapi/openapi.json`. Manual edits fail the drift check.

## Production proxy conditional-request hardening

Live production validation observed that the Render delivery path transforms the origin validator:

```text
"<sha256>"
```

into the semantically equivalent weak response validator:

```text
W/"<sha256>"
```

A browser or HTTP client correctly reuses the received value in `If-None-Match`. The original handler compared the request header to the origin strong validator byte-for-byte, so the proxy-weakened validator returned `200` instead of `304`.

The hardened handler now:

- applies the weak comparison required for `If-None-Match` on `GET`;
- accepts both strong and proxy-weakened forms for the same opaque SHA-256 tag;
- accepts comma-separated validator lists;
- accepts the `*` wildcard;
- preserves `200` for a non-matching validator;
- keeps the origin-generated ETag deterministic and SHA-256-backed;
- changes no OpenAPI operation, schema, authentication, or caching duration.

Regression tests cover strong, weak, listed, wildcard, and mismatched validators.

## Verification

Permanent checks enforce:

```text
OPENAPI_DEVELOPER_DOCS_ROUTE=/api/docs
OPENAPI_DEVELOPER_SPEC_ROUTE=/api/docs/openapi.json
OPENAPI_GENERATED_CLIENT_OPERATIONS=38
OPENAPI_BROWSER_MUTATION_EXECUTION=DISABLED
OPENAPI_EMBEDDED_SPEC_DRIFT=PASS
OPENAPI_GENERATED_CLIENT_DRIFT=PASS
OPENAPI_DEVELOPER_EXPERIENCE=PASS
```

PR #54 exact-head CI on `837a97b8696333fcbe8c7a50d76b3f76521a149b`:

```text
Backend CI        31059068320 = SUCCESS
Frontend CI       31059067951 = SUCCESS
CodeQL            31059068195 = SUCCESS
OpenAPI Contract  31059068374 = SUCCESS
API Load Baseline 31059068194 = SUCCESS
```

The OpenAPI workflow runs:

- source route inventory verification;
- OpenAPI contract verification;
- developer-experience regression tests;
- embedded and generated drift verification;
- deterministic generator check;
- frozen pnpm installation;
- generated API client TypeScript checking;
- generated API client runtime tests.

The root release verifier runs the same developer-experience tests, verifier, generator drift check, and client typecheck before frontend and backend release gates.

## Deliberate exclusions

This increment does not:

- add operations to the public `/api/v1` contract;
- expose `/internal/metrics`;
- make the protected mutation executable from a browser;
- place the internal API key in frontend environment variables;
- add Swagger UI, Scalar, Redoc, or another CDN/runtime dependency;
- generate and commit compiled JavaScript output;
- publish the API client to a public package registry.

---

## Canonical remediation record — GFA-CONTRACT-450

### 1. Finding / symptom

Production conditional GET requests for embedded OpenAPI assets could return HTTP `200` instead of `304 Not Modified` when the client reused a semantically equivalent weak ETag produced by the Render delivery path.

### 2. Root cause

The origin emitted a strong SHA-256 ETag, while the production intermediary could legally rewrite it as a weak validator. The handler compared `If-None-Match` and the origin ETag byte-for-byte instead of using weak entity-tag comparison semantics appropriate to GET conditional requests.

### 3. Failure scenario

The origin emits `"<sha256>"`; Render exposes `W/"<sha256>"`; a browser sends that received value in `If-None-Match`; the application compares the two strings literally and returns the full asset with `200` although the opaque validator identifies the same representation.

### 4. Impact

Conditional caching was semantically incorrect in production, causing unnecessary response bodies and making live validation disagree with the intended HTTP cache contract. OpenAPI content, authentication, and data integrity were not corrupted.

### 5. Severity rationale

**P2 retrospective.** This was a real production HTTP contract defect with reproducible proxy-dependent behavior, but it did not expose credentials, mutate data, corrupt the OpenAPI specification, or disable the API.

### 6. Existing guarantees violated

- `If-None-Match` on GET must honor weak comparison semantics;
- a delivery intermediary must not make an otherwise equivalent validator unusable;
- production validation must observe `304` for a matching conditional request;
- mismatched validators must continue to return `200`.

### 7. Considered solutions

- force the proxy to preserve the origin strong ETag;
- disable ETag/conditional caching;
- compare only exact strong strings;
- implement standards-aligned weak ETag comparison, including validator lists and wildcard handling.

### 8. Chosen remediation

Introduce explicit conditional-request matching that strips the weak prefix for comparison, accepts strong and weak forms of the same opaque tag, supports comma-separated `If-None-Match` candidates and `*`, and preserves ordinary `200` behavior for non-matches.

### 9. Why this solution was selected

The application controls its HTTP semantics but not every production intermediary. Weak comparison makes the handler robust to legal proxy transformations without changing the deterministic origin SHA-256 identity or cache duration.

### 10. Rejected alternatives

- relying on Render to preserve strong validators was rejected because the observed production path did not do so;
- disabling caching was rejected because the representation already has stable deterministic identity;
- exact string comparison was the defective behavior and could not support a legal weak validator.

### 11. Trade-offs

The matching helper must correctly parse the limited ETag syntax it supports. Weak equivalence intentionally treats strong and weak forms of the same opaque value as matching for GET cache validation; it does not make weak tags suitable for strong precondition semantics such as `If-Match`.

### 12. Regression tests / protection

`apps/api/internal/http/apidocs` tests cover strong, weak, comma-separated, wildcard, and mismatched validators. OpenAPI Developer Experience verification and release validation include the embedded documentation surface.

### 13. Adversarial review findings

The review distinguished deterministic content identity from validator strength. The fix preserves the exact SHA-256 opaque value while accepting proxy weakening only in the conditional GET comparison path.

### 14. Remediation iterations

The Developer Experience feature merged through PR #52 as `64400024cddad5e856d4e72433457b4bd2eacb11`. Live production validation then exposed the weak-validator behavior. PR #54 fixed it with head `837a97b8696333fcbe8c7a50d76b3f76521a149b` and merge `855f82bf97cf0db47d1a3918f75ea70f7f2b06fe`.

### 15. Residual risks and limitations

Future proxy/cache layers can still introduce other HTTP transformations. The current guard is specific to GET `If-None-Match` semantics for the embedded API documentation assets and does not claim support for unrelated conditional-write semantics.

### 16. Operational or deployment consequences

Clients and intermediary caches can reuse either strong or proxy-weakened equivalent validators and receive `304` without retransferring unchanged documentation assets. No API operation, schema, security requirement, or cache lifetime changed.

### 17. Exact evidence

- Developer Experience baseline: `b1f2b23a4f4bb72153d0993417e1888d538cfa75`;
- original feature merge: `64400024cddad5e856d4e72433457b4bd2eacb11`;
- PR #54 base: `c106c003cdc77f948ee34ad895a9fbbbd616cea8`;
- PR #54 head: `837a97b8696333fcbe8c7a50d76b3f76521a149b`;
- PR #54 merge: `855f82bf97cf0db47d1a3918f75ea70f7f2b06fe`;
- production evidence in PR #54: Render returned `W/"<sha256>"`; reuse previously returned `200` instead of `304`;
- Backend CI `31059068320` — SUCCESS;
- Frontend CI `31059067951` — SUCCESS;
- CodeQL `31059068195` — SUCCESS;
- OpenAPI Contract `31059068374` — SUCCESS;
- API Load Baseline `31059068194` — SUCCESS.

### 18. Final canonical status

**CLOSED.**

### 19. Prevention / future guard

Retain the weak/list/wildcard/mismatch regression tests and keep production OpenAPI conditional-GET validation in the release/runtime evidence path so intermediary behavior cannot silently reintroduce literal-string ETag matching.
