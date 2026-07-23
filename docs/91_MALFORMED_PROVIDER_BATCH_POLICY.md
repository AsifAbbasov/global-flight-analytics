# Malformed Provider Batch Policy

Status: Implemented Engineering Contract v1.0

## Scope

This increment defines one explicit item-level policy for successful
Airplanes.live and OpenSky transport responses.

## Policy

A provider batch is accounted as received items, accepted items, malformed
rejected items, and unavailable or stale rejected items. Every received
item belongs to exactly one accepted or rejected category.

A mixed batch remains usable. Valid items continue through normalization,
validation, persistence and trajectory construction. The ingestion run is
marked partial, and provider rejection evidence is retained.

An empty provider batch is a valid empty result.

A non-empty batch with zero accepted items returns the typed
`providerbatch.ErrAllItemsRejected` error. Automatic traffic mode treats
this as a recoverable invalid-response failure and may try the next
provider.

## Airplanes.live Boundary

An item is malformed when its required aircraft identity is empty, its
latitude or longitude is not finite, or its coordinates are outside valid
geographic ranges. Invalid optional telemetry remains represented through
existing availability and altitude-status semantics.

An invalid response timestamp makes every non-empty item malformed because
no trustworthy observation time can be produced.

## OpenSky Boundary

State vectors that fail validity evaluation are malformed. State vectors
with unavailable or stale provider positions are rejected as unusable.
Neither category terminates a mixed batch.

## Evidence Propagation

Provider batch evidence travels through provider orchestration, automatic
fallback and traffic ingestion. Ingestion-run records use raw provider
received counts. Provider health combines provider-level rejections with
processing invalid and duplicate counts.

## Verification

The acceptance gate includes provider accounting invariants, mixed and
fully rejected batches for both providers, partial-run policy tests,
fallback invalid-response classification, race tests, complete backend
tests, `go vet`, policy gates and clean diff validation.

## Review Closure

This closes the final previously identified Ingestion, Provider Adapters
and Orchestration review boundary. Review closure remains conditional on
every installer acceptance gate passing in the repository baseline.
