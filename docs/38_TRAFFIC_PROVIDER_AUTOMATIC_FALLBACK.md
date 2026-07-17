# Traffic Provider Automatic Fallback

Status: Production Integration Baseline v1.0
Project: Global Flight Analytics

## Purpose

This document closes the production traffic-provider integration sequence by
adding an optional automatic fallback mode between the two free external
observation providers currently supported by the backend.

The mode is enabled explicitly:

```text
TRAFFIC_PROVIDER=auto
```

The ordered provider chain is:

```text
airplanes.live
↓ only after a recoverable failure or access denial
OpenSky Network
```

The platform does not query both providers concurrently for the same ingestion
cycle. The secondary provider is called only after the primary provider cannot
serve the request under the bounded fallback policy.

## Supported provider modes

```text
TRAFFIC_PROVIDER=airplanes.live
TRAFFIC_PROVIDER=opensky
TRAFFIC_PROVIDER=auto
```

The default remains `airplanes.live`. Automatic fallback is opt-in so that a
deployment does not begin using OpenSky without an explicit operational choice.

## Recoverable fallback triggers

The automatic chain may continue to OpenSky after:

- an access denial from the provider budget controller;
- a provider rate-limit response;
- a local OpenSky polling cooldown;
- a provider server failure;
- a network or transport timeout while the operation context remains usable.

These conditions mean that the current provider is temporarily unable to
satisfy the bounded request. They do not mean that its data is incorrect.

## Non-recoverable failures

The chain deliberately stops and returns the error after:

- authentication or authorization failure;
- invalid client configuration;
- invalid request parameters;
- provider contract or decoding failure;
- operation context cancellation;
- evidence that cannot be mapped without violating the canonical contract.

Automatic fallback must not hide configuration problems, malformed provider
responses, or analytical contract violations.

## Actual-source provenance

A successful ingestion run records the provider that actually supplied the
accepted state collection.

Examples:

```text
Primary success:
ingestion_runs.source_name = airplanes.live

Primary unavailable, OpenSky selected:
ingestion_runs.source_name = opensky
```

The selected source is also propagated to every canonical `FlightState` and to
provider observation-health evidence. The chain name is never stored as if it
were a data source.

## Empty successful responses

A provider may legitimately return an empty regional state collection. The
source-aware provider contract carries the selected source separately from the
state slice, so an empty successful OpenSky response is still attributed to
OpenSky.

## Budget and request coalescing

Each provider remains behind its own existing orchestration controls:

```text
Provider policy
↓
Provider budget
↓
Request coalescing
↓
Provider HTTP client
```

The automatic provider sits above the individually orchestrated providers. It
does not bypass request budgets and does not duplicate simultaneous calls.

## Decision evidence

The existing provider decision collector records:

- primary selected;
- fallback selected;
- no provider available;
- primary trigger reason;
- selected provider;
- considered providers;
- retry time when known.

This evidence remains process-local and resets when the ingestion process
restarts.

## Scope boundaries

Automatic fallback does not create:

- project-owned receiver infrastructure;
- satellite coverage;
- global continuous coverage;
- commercial flight status;
- official airport schedules;
- official delay reasons;
- air traffic control intent or instruction data.

Both providers remain external, free, best-effort observation sources. Every
analytical result remains subject to data quality, confidence, explainability,
and source-constraint enforcement.

## Runtime verification

The installation increment verifies:

```text
go test ./internal/config
go test ./internal/services/traffic/ingestion
go test ./internal/orchestration/providerfallback
go test ./internal/orchestration/providerdecision
go test ./cmd/ingest
go vet for the same packages
go test ./...
git diff --check
```

A live external-provider request is not required by the installer and is not
claimed as verified until runtime credentials, network access, and PostgreSQL
are configured.
