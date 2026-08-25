# Document 118 — Flight Features Schema Review Hardening

Status: IMPLEMENTED
Baseline commit: `624fcf44d3260bd35ac32f67c0730689713198c0`

## Review basis

The supplied review was produced against the older commit `a1689dc`. This
increment re-evaluates every finding against the current baseline after the
extractor-composition, aircraft-provider and feature-store hardening work.

## Confirmed defect: geographical schema drift

The data model and builder already produced minimum and maximum latitude and
longitude, while the versioned schema omitted those four analytical fields.
They are now registered as required geographical definitions. The geographical
required field count therefore changes from eleven to fifteen.

`GeographicCellPrecision` is deliberately not a schema feature. It is the
processing-configuration mirror used to derive geographic-cell aggregation,
and its authoritative value is already persisted in
`Provenance.ProcessingIdentity.GeographicCellPrecision`.

```text
FLIGHT_FEATURES_SCHEMA_MODEL_ALIGNMENT=CLOSED
GEOGRAPHICAL_COMPLETENESS_DENOMINATOR=15
GEOGRAPHIC_CELL_PRECISION_CLASSIFICATION=PROCESSING_CONFIGURATION
```

## Processing identity

Because required completeness semantics change, this is not shipped under the
previous processing generation. The corrective generation is:

```text
FLIGHT_FEATURES_PROCESSING_GENERATION=v7
GEOGRAPHICAL_BUILDER_GENERATION=v2
VALIDATOR_GENERATION=v3
```

The snapshot key already includes processing version, so generation-seven
results cannot collide with generation-six snapshots.

## Schema evolution

The registry now exposes explicit version lookup, supported-version discovery,
version-specific definition lookup and compatibility classification. Version
one remains the current schema because the persisted data shape already
contained the four bounds; this change repairs registry metadata rather than
introducing a new payload shape.

```text
SCHEMA_VERSION_LOOKUP=ENFORCED
SCHEMA_SUPPORTED_VERSION_DISCOVERY=ENFORCED
SCHEMA_COMPATIBILITY_CLASSIFICATION=ENFORCED
```

## Central field-count ownership

Canonical group counts now live in `flightfeatures`. Builder contracts alias
those constants, while executable tests compare them against definitions
actually registered by the schema. Validator completeness remains derived from
the registry.

```text
FEATURE_GROUP_FIELD_COUNT_CONTRACT=CENTRALIZED
SCHEMA_EXACT_GROUP_COUNTS=TESTED
```

## Findings already closed before this increment

The current model already persists processing version, processing identity and
its fingerprint, component versions, geographic precision, aircraft metadata
source, provider version and retrieval time. Validator version is stored in the
durable validation report. Semantic output fingerprint is owned by the Feature
Store record and versioned payload, avoiding a self-referential hash inside the
features object. PostgreSQL serialization no longer directly marshals the Go
domain model.

```text
PROVENANCE_REPRODUCIBILITY_FINDING=STALE
OUTPUT_FINGERPRINT_FINDING=STALE
DIRECT_MODEL_SERIALIZATION_FINDING=STALE
```

## Deliberately retained contracts

`FlightFeatures` remains a draft and validated data-transfer aggregate with
public fields. The validator produces the durable complete audit proof and the
Feature Store rejects writes without that proof. Wrapping the same value in an
internal `ValidatedFlightFeatures` type would not prove that validation ran and
would duplicate the established trust boundary.

Unavailable group values continue to use zero values together with mandatory
`GroupEvidence`. A generic per-field `Metric[T]` would duplicate availability
state across every field and conflict with the group-level evidence contract.

Go string-backed enumerations cannot be made closed. Central `IsValid` methods
now expose the supported value sets, while validator and persistence boundaries
continue to reject unknown values.

Feature limitation codes remain open-world evidence identifiers. They are not
validation issues: structured severity, group and field path already belong to
`ValidationIssue`. A closed central limitation registry would force every new
provider or builder limitation to modify the central domain package.

```text
VALIDATED_WRAPPER_RECOMMENDATION=REJECTED_NON_BLOCKING
PER_FIELD_METRIC_WRAPPER=REJECTED_NON_BLOCKING
STRING_ENUM_CLOSURE=LANGUAGE_NOT_APPLICABLE
LIMITATION_REGISTRY=DELIBERATELY_OPEN_NON_BLOCKING
```

## Permanent verification

Continuous Integration now runs `flightfeaturesreviewaudit` in addition to the
existing feature, processing-identity, provenance, extractor, provider and
store audits. Targeted tests verify exact geographical definitions, exact group
counts, version lookup, compatibility and enumeration validity.

```text
FLIGHT_FEATURES_SCHEMA_MODEL_ALIGNMENT=CLOSED
GEOGRAPHICAL_COMPLETENESS_DENOMINATOR=15
FLIGHT_FEATURES_PROCESSING_GENERATION=v7
FLIGHT_FEATURES_REVIEW_STATUS=CLOSED
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
```

## Canonical remediation history

Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, and Continuous Integration evidence. Severity is retrospective.

### GFA-DATA-173 — Versioned Flight Features schema omitted four produced geographical fields

1. **Finding / symptom:** the geographical builder and data model produced minimum/maximum latitude and longitude, but schema version one did not register those four analytical fields.
2. **Root cause:** versioned schema metadata drifted behind the actual model/builder contract.
3. **Failure scenario:** completeness and schema-driven validation count only eleven geographical fields while the runtime analytical aggregate actually owns fifteen required fields.
4. **Impact:** quality/completeness evidence can be overstated or interpreted against the wrong denominator, and schema introspection misrepresents the durable model.
5. **Severity rationale:** **P1 retrospective** because schema-driven completeness is analytical evidence used to judge whether a snapshot is trustworthy.
6. **Existing guarantees violated:** schema-model alignment, exact group-count ownership, and processing-generation isolation when semantics change.
7. **Considered solutions:** remove the four model fields, treat them as optional/unregistered, introduce a new payload schema version, or repair version-one registry metadata and advance processing semantics.
8. **Chosen remediation:** register all four bounds as required geographical definitions, centralize exact group counts, add explicit schema-version lookup/compatibility APIs, and advance processing to generation seven.
9. **Why this solution was selected:** the persisted payload already contained the fields, so the defect was registry metadata rather than a new wire shape; processing version correctly isolates the changed completeness semantics.
10. **Rejected alternatives:** inventing a schema-v2 payload migration, treating geographic precision as a feature field, and introducing generic per-field wrappers unrelated to the confirmed drift.
11. **Trade-offs:** version one now more accurately describes data already stored, while consumers relying on the old incorrect denominator must adopt generation-seven semantics.
12. **Regression tests / protection:** exact definition/count tests, schema version lookup/compatibility tests, enum validity tests, and `flightfeaturesreviewaudit` in Backend Quality.
13. **Adversarial review findings:** stale provenance/output-fingerprint/direct-serialization findings were not re-opened; wrapper, per-field metric, closed-enum and limitation-registry proposals were explicitly rejected as non-blocking.
14. **Remediation iterations:** implemented in `2eb2c49c11fc1894969ef62c0f9ea0e244a3103f` after Feature Store hardening `624fcf44d3260bd35ac32f67c0730689713198c0`.
15. **Residual risks and limitations:** schema metadata still requires intentional maintenance whenever future feature fields are introduced; the permanent audit guards the current contract.
16. **Operational or deployment consequences:** no PostgreSQL migration; processing generation changes from six to seven so new outputs cannot collide with earlier snapshots.
17. **Exact evidence:** commit `2eb2c49c11fc1894969ef62c0f9ea0e244a3103f`, schema/count regression tests, `flightfeaturesreviewaudit`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every feature-field addition/removal must update schema definitions, canonical group counts, processing identity when semantics change, and executable exact-count tests in the same change.
