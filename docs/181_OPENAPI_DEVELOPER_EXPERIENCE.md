# OpenAPI Developer Experience

Status: implementation increment
Baseline: `b1f2b23a4f4bb72153d0993417e1888d538cfa75`
Scope: embedded API documentation, deterministic TypeScript client generation, drift prevention, and Continuous Integration enforcement

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

## Completion criteria

The increment is complete when:

1. the embedded specification is byte-identical to the canonical specification;
2. all 38 operations are represented in generated TypeScript metadata and types;
3. the API client typechecks under the pinned workspace TypeScript version;
4. the documentation application is compiled into and served by the Go API;
5. the browser mutation execution boundary remains disabled;
6. targeted tests and the full release verifier pass;
7. the exact change set is committed only after installation validation.
