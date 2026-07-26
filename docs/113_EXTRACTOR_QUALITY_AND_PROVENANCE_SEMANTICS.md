# Document 113 — Extractor Quality and Provenance Semantics

Status: IMPLEMENTED
Baseline commit: e853f5931c78f6ed7b0fbcd0dd85a53cfbaa22f3

## 1. Required completeness and optional coverage

The schema already declared all six aircraft enrichment fields optional. The
quality model now follows that declaration instead of placing optional aircraft
fields in the denominator of core completeness.

`CompletenessScore` measures required feature fields only. The new
`OptionalCoverageScore` independently reports optional enrichment coverage.
Missing or partial optional aircraft metadata remains visible through evidence
status and limitations but no longer makes an otherwise complete required
feature set limited.

The current schema contract contains 46 required fields and 6 optional aircraft
fields. Aggregate evidence groups are prohibited from mixing required and
optional fields until the evidence model can report per-requirement counts.

## 2. Honest trajectory record provenance

`TrajectoryUpdatedAt` now contains only the real trajectory `UpdatedAt` value.
`TrajectoryCreatedAt` records the independent creation timestamp. When either
value is unknown it remains zero and is reported honestly; trajectory `EndTime`
is never substituted for a system record timestamp.

System record timestamps are not compared with historical `AsOfTime`, because
record materialization may legitimately happen after the aviation event window.
The validator instead checks creation/update ordering and reports entirely
missing record timestamps as a provenance limitation.

## 3. Aircraft metadata provenance

Production extractor composition now supplies an explicit aircraft metadata
source name and provider version. Every configured aircraft enrichment records:

- metadata source name;
- provider version;
- retrieval completion time.

The stable source name also participates in extractor composition processing
identity. Retrieval time is intentionally provenance, not deterministic input
identity, so replay fingerprints do not change merely because the same inputs
were processed at a different wall-clock time.

## 4. Validation semantics

Optional-only group absence and partial coverage no longer generate validation
warnings. Structural inconsistencies remain errors. Aircraft metadata that is
available or partial requires complete source, provider-version and retrieval
time provenance.

## 5. Processing generation

The extractor, extractor composition and feature processing pipeline advance to
version 5. The validator advances to version 2. Stored snapshots remain isolated
by explicit processing version.

```text
REQUIRED_COMPLETENESS_OPTIONAL_COVERAGE=SEPARATED
OPTIONAL_AIRCRAFT_ABSENCE_VALIDATION_PENALTY=REMOVED
TRAJECTORY_ENDTIME_PROVENANCE_FALLBACK=REMOVED
TRAJECTORY_CREATED_UPDATED_PROVENANCE=EXPLICIT
AIRCRAFT_METADATA_PROVENANCE=EXPLICIT
AIRCRAFT_METADATA_SOURCE_PROCESSING_IDENTITY=ENFORCED
EXTRACTOR_PROCESSING_GENERATION=v5
VALIDATOR_GENERATION=v2
```
