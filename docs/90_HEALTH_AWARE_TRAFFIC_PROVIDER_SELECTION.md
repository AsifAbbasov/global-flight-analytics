# Health-Aware Traffic Provider Selection

Status: Implemented Engineering Contract v1.0

## Scope

This increment closes the health-aware primary and fallback selection
boundary for automatic traffic ingestion. The configured provider order
remains authoritative when health evidence is equal, unknown, or
unavailable. Stronger health evidence may change the attempt order before
an external request is made.

## Previous Failure Mode

Automatic mode always attempted Airplanes.live before OpenSky. Provider
health was collected and printed, but it was not used by the selection
path. A provider already classified as unavailable was therefore attempted
again before a provider with healthy evidence.

## Implemented Ordering Policy

Each cycle reads a fresh snapshot for every configured traffic provider.

The stable priority is:

1. healthy;
2. degraded or unknown;
3. unavailable.

Equal priorities preserve configuration order. The policy does not
permanently disable a provider and does not create a separate circuit
breaker. If the first health-preferred provider fails with a recoverable
error, the remaining configured provider is still attempted.

An unavailable configured primary receives a bounded recovery probe after two
minutes without an attempt. This prevents a previously failing primary from
being permanently starved by a healthy fallback. A successful recovery resets
its consecutive failure evidence through the existing health collector.

Snapshot lookup failure is fail-open. The affected provider receives an
unknown status and configured order is preserved unless another provider
has strictly stronger evidence.

## Decision Evidence

Fallback decisions now record:

- whether health evidence was evaluated;
- whether attempt order changed;
- the ordering reason;
- configured primary health status;
- selected provider health status.

`PrimaryProvider` continues to mean the configured primary provider, even
when a healthier secondary provider is attempted first. Selecting that
secondary provider is recorded as `fallback_selected`, not
`primary_selected`.

## Production Wiring

The traffic provider factory accepts an optional health snapshot source.
Existing isolated callers remain compatible and use unknown health
evidence. Production `cmd/ingest` passes the existing provider health
collector, so automatic mode is health-aware without introducing another
health store.

Provider health history remains process-local. The selection decision is
therefore operational evidence for the current daemon process, not a
cross-process global availability claim.

## Verification

The acceptance gate includes:

- healthy secondary preference over unavailable primary;
- configured order preservation for equal evidence;
- bounded configured-primary recovery probing;
- fail-open behavior when snapshots cannot be read;
- existing fallback compatibility tests;
- provider decision evidence tests;
- targeted race detector;
- complete backend tests;
- `go vet`;
- existing code review policy gates;
- clean Git diff validation.

## Remaining Review Boundary

After this increment, the remaining known Ingestion, Provider Adapters and
Orchestration review boundary is the explicit malformed-item policy for an
otherwise successful provider batch.
