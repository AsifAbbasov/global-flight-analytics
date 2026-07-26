# Document 112 — Extractor Input Correctness Hardening

Status: IMPLEMENTED
Baseline commit: ff84eefdb8ab363e2bdd276f99e49df7235fb50f

## 1. Scope

This increment closes the critical input-correctness findings from the
`internal/features/extractor` review without conflating event time with system
provenance time.

The extractor now rejects any trajectory snapshot containing:

- a point observed after `Request.AsOfTime`;
- a segment start or end after `Request.AsOfTime`;
- a coverage-gap start or end after `Request.AsOfTime`.

`CreatedAt` and `UpdatedAt` remain system provenance timestamps and are not
incorrectly treated as aviation-event cutoffs.

## 2. Dependency and cancellation contract

Nil contexts are rejected explicitly. Required builder interfaces reject typed
nil values. A typed-nil optional aircraft provider is treated as unavailable.
Cancellation is checked again immediately after aircraft enrichment before
canonical serialization and hashing.

## 3. Quality and mathematics integrity

Initial quality construction rejects negative field counts, available counts
greater than total counts, negative supporting-point counts, non-finite input
quality and input quality outside the inclusive range `[0, 1]`. Invalid values
are no longer converted into ordinary zero or one scores.

## 4. Canonical fingerprint semantics

ICAO24 and callsign values are canonicalized consistently before hashing.
Semantically equivalent casing and surrounding whitespace therefore do not
create different snapshot identities.

## 5. Processing generation

Nested evidence acceptance and fingerprint canonicalization change processing
semantics. The extractor, extractor composition and feature processing pipeline
therefore advance to generation `v4`. Existing stored snapshots remain readable
through their explicit processing versions.

## 6. Deferred review scope

This increment intentionally does not claim closure for:

- explicit aircraft source and retrieval provenance;
- replacement of the `TrajectoryUpdatedAt` fallback;
- separation of required completeness from optional coverage;
- centralization of duplicate ICAO24 and clone helpers.

Those items require a separate contract change rather than being hidden inside
this correctness patch.

```text
EXTRACTOR_NESTED_TEMPORAL_GUARD=ENFORCED
EXTRACTOR_NIL_CONTEXT=REJECTED
EXTRACTOR_TYPED_NIL_DEPENDENCIES=REJECTED
EXTRACTOR_POST_PROVIDER_CANCELLATION=ENFORCED
EXTRACTOR_SEMANTIC_FINGERPRINT_NORMALIZATION=ENFORCED
EXTRACTOR_INVALID_EVIDENCE_COUNTS=REJECTED
EXTRACTOR_INVALID_MATH_MASKING=CLOSED
EXTRACTOR_PROCESSING_GENERATION=v4
```
