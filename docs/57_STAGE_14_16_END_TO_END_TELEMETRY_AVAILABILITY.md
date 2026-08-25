# Document 57 — Stage 14.16 End-to-End Telemetry Availability

Status: Remediation History v1.1
Project: Global Flight Analytics
Scope: preserve kinematic telemetry availability from provider to analytical reads

## 1. Correctness Problem

Projection Intelligence already rejected PostgreSQL rows where required kinematic columns were `NULL`. That boundary was necessary but not sufficient.

OpenSky mapping previously converted absent velocity, heading, and vertical rate values into numerical zero before persistence:

```text
provider NULL
optionalFloat64Value(nil)
Go zero
PostgreSQL zero
Projection IS NOT NULL
false complete telemetry
```

Once absence became zero, downstream code could not distinguish a real observed zero from an unavailable provider value.

## 2. Domain Contract

`flightstate.FlightState` carries:

```text
TelemetryAvailabilityKnown
VelocityAvailable
HeadingAvailable
VerticalRateAvailable
OnGroundAvailable
```

Values remain ordinary Go values for compatibility; availability is an independent fact. `TelemetryAvailabilityKnown=false` preserves legacy fixtures/producers that predate the contract, while production providers and PostgreSQL readers set it to true.

## 3. Provider Mapping

OpenSky finite optional mapping is:

```text
nil        -> 0, false
NaN        -> 0, false
Infinity   -> 0, false
zero       -> 0, true
finite     -> value, true
```

A fresh position can therefore be persisted even when movement telemetry is unavailable. Airplanes.live explicitly marks its mapped kinematics as available under its then-current response contract.

## 4. PostgreSQL Write Semantics

`SaveFlightStates` writes PostgreSQL `NULL` when availability is false and the numerical value when availability is true. Real zero values remain valid stored values. No migration was required because the existing telemetry columns already allow `NULL`.

## 5. PostgreSQL Read Semantics

General Flight State and reconciliation readers use nullable PostgreSQL types for movement telemetry and restore both value and availability. Latitude/longitude are not fabricated through zero fallbacks; rows without usable positions are excluded where required.

## 6. Traffic Read Boundary

The current traffic map contract remains non-nullable, so Traffic selects only rows with the display telemetry it needs rather than fabricating missing values.

## 7. Airspace Intelligence Boundary

Airspace calculations require complete motion telemetry. Missing velocity, heading, vertical rate or on-ground state is excluded from interaction/proximity/separation calculations rather than interpreted as a real zero.

## 8. Data Quality Validation

The traffic validator treats explicit unavailability as missing movement telemetry. Position-only observations receive a reduced completeness classification instead of a false complete state. Legacy states without explicit availability retain their compatibility behavior.

## 9. Final Correctness Audit Expansion

`backendfinalaudit` verifies FlightState availability fields, provider mapping, nullable PostgreSQL writes/reads, reconciliation, Traffic/Airspace/Projection eligibility and validator awareness. Dangerous numerical `COALESCE` expressions are protected against reintroduction in the relevant readers.

## 10. Acceptance Scenarios

Acceptance proves missing provider telemetry remains unavailable through persistence/readback, explicit zero remains available zero, and Traffic/Airspace/Projection/validation apply their documented eligibility semantics.

## 11. Verification

Acceptance requires focused provider/domain/repository/validator/Traffic/Airspace/Projection tests, expanded backendfinalaudit, strict projectaudit, race detector, complete Go build/vet/tests, frontend dependency/lint/type/build checks, backend container build and diff validation.

## 12. Canonical remediation history

### Finding / symptom

The Stage 14.13 Projection read fix preserved database `NULL`, but provider mapping could destroy absence **before** persistence by converting missing OpenSky kinematics to Go zero values. Once written as non-null zero, every downstream layer saw false evidence as observed telemetry.

### Root cause

Availability was implicit in pointer/null representation at the provider edge but the canonical `FlightState` value contract lacked independent availability metadata. Mapping therefore collapsed `(value, availability)` into only a numerical value.

### Failure scenario

```text
provider omits velocity/heading/vertical rate
↓
provider mapper converts nil to numeric zero
↓
repository persists zero as observed value
↓
readers correctly see non-NULL zero
↓
Traffic/Airspace/Projection/validator treat telemetry as genuinely observed
↓
analytics consume fabricated evidence even though read-side NULL protection is intact
```

### Impact

The defect could contaminate multiple analytical contexts, not just Projection: map traffic completeness, Airspace interaction/separation calculations, quality classification and downstream trajectory features could all receive false complete movement evidence.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P1 end-to-end data correctness** because source absence was irreversibly transformed into valid persisted evidence and propagated across several production analytical paths.

### Existing guarantees violated

```text
provider absence must survive canonicalization and persistence
observed zero must remain distinguishable from missing telemetry
canonical FlightState must carry enough information for downstream eligibility decisions
read-side correctness cannot compensate for evidence already corrupted at ingestion
```

### Considered solutions

1. Drop every observation that lacks any kinematic field.
2. Convert `FlightState` numerical fields to pointers everywhere.
3. Preserve existing value fields but add explicit availability metadata and persist `NULL` from that metadata.
4. Infer missingness from numerical zero downstream.

### Chosen remediation

Option 3: add an explicit availability contract to `FlightState`, map provider optional fields into `(value, available)`, write unavailable telemetry as PostgreSQL `NULL`, restore availability on reads, and make each consumer apply its own completeness requirements.

### Why this solution was selected

It preserves useful position-only observations, keeps compatibility with existing value-oriented algorithms/fixtures, distinguishes legitimate zero from absence, and lets different consumers define different telemetry eligibility without fabricating evidence.

### Rejected alternatives

Dropping partial observations was rejected because valid position evidence remains useful. A project-wide pointer migration was rejected as unnecessarily broad for the confirmed problem. Zero-as-sentinel inference was rejected because zero is legitimate telemetry.

### Trade-offs

```text
+ source absence remains observable end to end
+ legitimate zeros remain valid
+ partial observations can still be retained
+ each analytical context owns explicit eligibility
- FlightState carries additional availability state
- legacy producers require compatibility semantics through TelemetryAvailabilityKnown
- consumers must maintain their own required-field policy deliberately
```

### Regression tests / protection

Tests cover nil/NaN/infinity/zero provider mapping, PostgreSQL NULL versus zero writes and reads, reconciliation semantics, Traffic and Airspace row eligibility, validator completeness and Projection eligibility. `backendfinalaudit` protects the whole path.

### Adversarial review findings

The critical adversarial finding was that Stage 14.13 was locally correct yet globally insufficient: once a provider had already stored false zero evidence, `IS NOT NULL` predicates could not identify it. This forced the remediation boundary upstream to provider mapping and downstream through every production consumer.

Historical PR/reviewer evidence for the original July remediation is unavailable; reconstruction is limited to repository source, tests, commits, and stage documents.

### Remediation iterations

1. Document 53 / commit `1f30bae...` fixed nullable semantics at the Projection read boundary.
2. Adversarial review identified pre-persistence loss of availability.
3. Commit `9cfa9005baf9467ed94621602efd48e8b108bb44` (`fix: preserve telemetry availability end to end`) introduced the canonical availability contract and expanded production consumers/audits.
4. Later provider hardening work continued to refine optional telemetry semantics without changing the principle established here.

### Residual risks and limitations

`TelemetryAvailabilityKnown=false` is intentionally a legacy compatibility state and therefore cannot provide the same evidence precision as explicit production availability. Each new provider must map optional fields correctly; merely populating numeric values is insufficient. Consumer eligibility policies can also diverge legitimately and must remain documented.

### Operational or deployment consequences

No schema migration was required. Existing nullable columns are now used semantically. Operators should expect more PostgreSQL `NULL` values and fewer eligible observations in analytics when providers omit telemetry; this represents honest evidence preservation, not data loss to be “fixed” with defaults.

### Exact evidence

```text
implementation commit:
9cfa9005baf9467ed94621602efd48e8b108bb44

precursor read-boundary remediation:
1f30bae8bb8a9e4e27d634b44362dcf7547e54ff

permanent evidence:
provider mapping tests
FlightState availability tests
PostgreSQL write/read and reconciliation tests
Traffic/Airspace/Projection eligibility tests
validator tests
expanded backendfinalaudit
```

### Final canonical status

```text
FINDING_GFA_DATA_055_END_TO_END_TELEMETRY_AVAILABILITY=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/57_STAGE_14_16_END_TO_END_TELEMETRY_AVAILABILITY.md
IMPLEMENTATION_COMMIT=9cfa9005baf9467ed94621602efd48e8b108bb44
```

### Prevention / future guard

Every provider adapter must preserve availability separately from numeric value for optional evidence. Canonical models and persistence must never infer availability from zero, and new analytical consumers must explicitly state which fields they require. A read-side filter is not considered sufficient protection unless ingestion preserves missingness first.
