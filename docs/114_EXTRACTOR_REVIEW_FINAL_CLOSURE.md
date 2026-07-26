# 114. Extractor Review Final Closure

## Baseline

This closure increment is applied after commit
`3cdd9b1532609c872343d00626ba44a9c9855609`.

## Scope

The final increment removes the remaining contract duplication and installs a
permanent closure gate for the complete flight-feature extractor review.

## Centralized contracts

- ICAO24 canonicalization and validation belong to `domain/aircraft`.
- Defensive trajectory copying belongs to `domain/trajectory`.
- Aircraft feature field counts are derived from the current versioned feature
  schema rather than duplicated constants.
- Extractor, aircraft provider, validator, and fingerprint construction consume
  these shared contracts.

## Canonical fingerprint drift protection

Reflection-based tests compare every trajectory, point, segment, and coverage-gap
field against the corresponding canonical fingerprint structure in the same
order. A new, removed, renamed, or reordered evidence field therefore fails the
extractor test suite until the deterministic fingerprint contract is updated
intentionally.

Aircraft metadata retrieval time remains excluded from the deterministic
fingerprint. Source identity and provider version remain included.

## Processing identity

This increment preserves behavior and does not create a new processing
semantics generation. Extractor, composition, feature pipeline, and validator
versions remain at their established version 5 / version 2 contracts.

No PostgreSQL migration is required.

## Verification gates

The permanent `extractorreviewclosureaudit` checks centralized ownership
inside the extractor processing path, absence of the removed local helpers and
constants, fingerprint mirror tests, version stability, documentation status,
and Continuous Integration wiring.

```text
EXTRACTOR_ICAO24_CONTRACT=CENTRALIZED
EXTRACTOR_TRAJECTORY_CLONE_CONTRACT=CENTRALIZED
EXTRACTOR_AIRCRAFT_FIELD_COUNT=SCHEMA_DERIVED
EXTRACTOR_FINGERPRINT_MIRROR_GUARD=ENFORCED
EXTRACTOR_PROCESSING_GENERATION=v5
EXTRACTOR_REVIEW_STATUS=CLOSED
OPEN_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
```
