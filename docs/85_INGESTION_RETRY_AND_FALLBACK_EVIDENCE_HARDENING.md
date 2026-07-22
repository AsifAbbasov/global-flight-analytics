# Document 85 — Ingestion Retry Scheduling and Fallback Evidence Hardening

Status: Implemented
Project: Global Flight Analytics
Scope: traffic ingestion retry timing, local denial semantics, fallback attempt evidence, and OpenSky polling ownership

## 1. Failure mode

The traffic ingestion daemon previously waited one fixed interval after every cycle. Provider-directed retry times were visible in errors but did not affect scheduling, repeated failures did not receive bounded backoff, and locally denied requests could still create failed ingestion-run rows even though no external HTTP operation occurred.

Fallback decisions also recorded only selected providers and one trigger reason. A recoverable primary failure followed by a non-recoverable secondary failure could terminate without a complete ordered attempt record.

## 2. Retry scheduling contract

The daemon now calculates the next delay from:

```text
normal interval
bounded exponential failure backoff
provider-directed RetryAt
```

The first failure waits the normal interval. Repeated consecutive failures double the delay up to `TRAFFIC_INGESTION_MAX_BACKOFF`. A later provider-directed retry time always wins over the local backoff.

A successful cycle resets the consecutive-failure count.

## 3. External-request evidence

Errors that represent a local policy decision expose:

```text
ExternalRequestAttempted() bool
RetryAtTime() time.Time
```

The contract is implemented by:

```text
ingestionorchestrator.AccessDeniedError
providerfallback.NoProviderAvailableError
opensky.PollingTooSoonError
```

Traffic ingestion does not create an `ingestion_runs` row when the entire provider chain was rejected locally before an HTTP attempt. Real provider requests that fail continue to create durable failed-run evidence.

## 4. Ordered fallback evidence

Every fallback decision now retains ordered attempt evidence:

```text
provider
outcome
reason
retry_at
error_class
request_attempted
```

Supported outcomes include success, denied, failed, and terminal failure. Mixed paths such as a primary server error followed by a secondary authorization error are recorded before the original terminal error is returned.

Provider Decision Collector copies attempt slices at both input and output boundaries so callers cannot mutate stored evidence.

## 5. OpenSky polling ownership

OpenSky now returns a typed polling-cooldown error with an exact retry time. A polling reservation can be released when request preparation fails before the HTTP transport is invoked.

The first unauthorized response body is explicitly closed before the authenticated retry starts.

## 6. Configuration

```text
TRAFFIC_INGESTION_MAX_BACKOFF=2m
```

The configured maximum must be at least the normal ingestion interval.

## 7. Verification

The permanent verification path includes:

```text
configuration tests
ingest daemon retry-policy tests
local-denial ingestion tests
fallback selector evidence tests
mixed fallback terminal-path tests
provider decision copy-boundary tests
OpenSky polling reservation tests
targeted race detector
full backend test suite
Go static analysis
working-tree diff validation
```

## 8. Completion boundary

This increment closes retry scheduling, false failed-run creation, incomplete fallback-chain evidence, OpenSky local polling retry metadata, polling-slot release before transport, and delayed first-401 body closure.

OurAirports publication reservation and commit semantics remain the final separate ingestion-layer increment.
