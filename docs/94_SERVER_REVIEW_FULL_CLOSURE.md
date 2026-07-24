# Document 94 — Server Review Full Closure

Status: Implemented Engineering Contract v1.0
Project: Global Flight Analytics
Baseline: `1fc925c91117eebbb7c90c4bd6b3889548d55cb4`

## 1. Purpose

This document completes classification and remediation of every observation from
the original Server and HTTP Protection review.

The closure policy follows the project Code Review Standard:

```text
fixed;
not applicable;
deliberately rejected with evidence;
deferred with owner, risk and revisit condition;
suggestion or nit that does not block merge.
```

## 2. Fixed findings

### 2.1 Process lifecycle

The command lifecycle now has one controlled `run` path and one final process
exit point.

The implementation:

```text
waits for either listener failure or context cancellation;
returns listener failures through a buffered error channel;
uses ShutdownWithTimeout with a ten-second bound;
waits for the listener to stop after shutdown;
closes PostgreSQL through a deferred cleanup path;
contains executable lifecycle tests.
```

`os.Exit` remains only in `main`, after `run` has returned and its deferred
resource cleanup has completed.

### 2.2 Request logger final status

The request logger now invokes the configured Fiber error handler before reading
the response status.

It therefore logs the status actually returned to the client rather than a
provisional status that existed before centralized error handling.

Regression tests cover both an error response and a successful response.

### 2.3 Sensitive internal error logging

The global error handler no longer records arbitrary `err.Error()` text for
internal request failures.

The log contains stable metadata:

```text
request identifier;
method;
path;
final status;
Go error type.
```

The raw error message is excluded. A regression test injects a synthetic
credential-like value and proves that it does not reach the log.

### 2.4 Historical Intelligence read interface

The read-only route registration boundary now accepts
`historicalaggregatecontract.Reader` instead of the full read/write store.

The production PostgreSQL implementation still satisfies the reader contract,
while verification stubs no longer need unrelated write methods.

### 2.5 Readiness rate-limit exclusion

`/api/v1/ready` is treated as an infrastructure route and is excluded from the
application rate limiter together with `/health` and `/version`.

A regression test repeatedly calls readiness with a one-request application
limit and proves that the response remains the dependency status rather than
becoming `HTTP 429`.

## 3. Deliberately rejected findings

### 3.1 Mechanical function-length threshold

A universal forty-line or fifty-line failure rule is rejected.

The previous database composition root was already decomposed by Document 48.
The remaining server functions are split when they mix lifecycle, route
registration, persistence ownership or independently testable policy.

This increment specifically decomposes the process lifecycle because it carried
real cleanup and failure-propagation risk. Functions are not split solely to
reduce line count.

### 3.2 Duplicate repository and service objects

Core traffic routes and Route Intelligence intentionally own separate
composition graphs.

The duplicated objects are stateless adapters and services over the same
PostgreSQL pool. They do not own independent transactions, mutable caches,
credentials or configuration that can diverge.

Route Intelligence keeps a self-contained, versioned PostgreSQL composition so
the pipeline can be verified and materialized independently. Reusing object
identity would add cross-context coupling without changing persistence
correctness.

The proposal to introduce a shared service container is therefore deliberately
rejected until a concrete shared mutable policy exists.

### 3.3 Replacing the custom rate limiter solely because of size

The local fixed-window limiter remains appropriate for the approved
single-instance MVP.

The review did not provide representative load measurements proving that the
mutex or periodic map cleanup violates a latency or throughput objective.
Replacement based only on source length is rejected.

## 4. Deferred findings

### 4.1 Trusted proxy client identity

Risk:

```text
the default limiter key uses the direct connection address;
a deployment behind a reverse proxy may require a trusted proxy contract;
blindly trusting X-Forwarded-For would permit spoofing.
```

Owner:

```text
backend deployment hardening
```

Revisit condition:

```text
before public multi-user deployment behind Render or another reverse proxy;
before horizontal scaling;
or after measured evidence that the current connection identity groups
unrelated clients.
```

Required future evidence:

```text
documented hosting proxy ranges or a platform-supported trusted proxy mode;
spoofing regression tests;
deployment smoke evidence for the selected client identity header.
```

### 4.2 Build-derived version endpoint

The hardcoded version remains a suggestion, not a correctness defect.

Owner:

```text
release engineering
```

Revisit condition:

```text
before the first tagged public release or when build provenance is exposed in
deployment diagnostics.
```

## 5. Suggestions and nits

The following observations are classified as non-blocking nits:

```text
capitalization of isolated Go error strings;
the word And in a test name;
local formatting and naming preferences without a demonstrated failure mode.
```

They may be corrected when the affected code is otherwise changed.

## 6. Verification gates

Formal full closure requires one new commit to pass:

```text
Go formatting;
cmd/server lifecycle tests;
middleware request logger tests;
server error-log redaction test;
readiness rate-limit regression test;
complete backend tests;
targeted race tests including cmd/server;
Go vet;
project architecture and contract audit;
code review policy audit;
Stage 14 final audit;
PostgreSQL 16 Integration;
Backend Container.
```

## 7. Closure statement

When every gate in Section 6 passes on the same commit:

```text
Server and HTTP Protection review: CLOSED
Open release blockers: 0
Open required changes: 0
Unclassified original findings: 0
Deferred deployment findings: 2
Release decision: ACCEPTABLE
```

Deferred findings are classified debt with explicit owners and revisit
conditions. They do not make the review unclassified or block the current MVP.
