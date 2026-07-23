# Exact Deduplication and Airplanes.live Telemetry Hardening

Status: Implemented Engineering Contract v1.0
Project: Global Flight Analytics

## Purpose

This increment closes the confirmed correctness gaps in exact in-memory observation deduplication and Airplanes.live telemetry mapping.

## Exact observation equality

`RemoveExactDuplicates` now compares the complete canonical observation payload:

```text
aircraft identity and callsign
observation time
position
both altitude values and statuses
velocity, heading and vertical rate values and availability
on-ground value and availability
telemetry availability knowledge
country, squawk and Special Purpose Indicator
position source
aircraft category and availability
provider source
```

Internal persistence identifiers are deliberately excluded:

```text
Flight State identifier
Flight identifier
Aircraft identifier
Ingestion Run identifier
```

Those identifiers may differ during replay without changing the source observation itself.

## Nullable telemetry

Airplanes.live numeric telemetry now retains the distinction between:

```text
observed zero
missing field
explicit null
invalid value
```

Missing, null and invalid telemetry does not set the corresponding canonical availability flag.

## Time safety

Provider response milliseconds and per-aircraft `seen` seconds are checked before conversion to `int64` or `time.Duration`.

Invalid response time produces an unknown zero timestamp. Invalid or overflowing `seen` evidence preserves the valid provider response time instead of performing an unsafe subtraction.

## Provider construction

`NewProvider(nil)` refuses to construct a usable provider. Direct use of a nil provider receiver returns `ErrClientRequired` rather than dereferencing a nil transport.

## Verification

The increment requires:

```text
targeted deduplicator and Airplanes.live tests
targeted race detector tests
full backend tests
go vet
existing code review policy gates
git diff check
```

## Remaining boundary

This increment does not change distributed provider rate limiting or fallback health policy. Those are separate architecture decisions and remain classified independently.
