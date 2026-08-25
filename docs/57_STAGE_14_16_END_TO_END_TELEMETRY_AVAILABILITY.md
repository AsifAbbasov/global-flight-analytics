# Document 57 — Stage 14.16 End-to-End Telemetry Availability

Status: Remediation History v1.1
Project: Global Flight Analytics
Scope: preserve kinematic telemetry availability from provider to analytical reads

## 1. Correctness Problem

Projection Intelligence already rejected PostgreSQL rows where required
kinematic columns were `NULL`.

That boundary was necessary but not sufficient.

OpenSky mapping previously converted absent velocity, heading, and vertical
rate values into numerical zero before persistence:

```text
provider NULL
optionalFloat64Value(nil)
Go zero
PostgreSQL zero
Projection IS NOT NULL
false complete telemetry
```

Once absence became zero, downstream code could not distinguish:

```text
a real observed zero
an unavailable provider value
```

## 2. Domain Contract

`flightstate.FlightState` now carries:

```text
TelemetryAvailabilityKnown
VelocityAvailable
HeadingAvailable
VerticalRateAvailable
OnGroundAvailable
```

The values remain ordinary Go values for compatibility. Availability is an
independent fact.

Examples:

```text
VelocityMPS=0, VelocityAvailable=true
means the provider observed zero velocity

VelocityMPS=0, VelocityAvailable=false
means velocity was unavailable
```

`TelemetryAvailabilityKnown=false` preserves compatibility with existing
legacy fixtures and internal producers that predate the availability contract.

Production providers and PostgreSQL readers set it to `true`.

## 3. Provider Mapping

OpenSky uses a finite optional-number mapper:

```text
nil        -> 0, false
NaN        -> 0, false
Infinity   -> 0, false
zero       -> 0, true
finite     -> value, true
```

A fresh OpenSky position remains usable even when kinematics are missing.
The state is persisted as a position-only observation rather than discarded.

Airplanes.live currently exposes these mapped fields through non-nullable
provider response values. Its mapper explicitly marks the mapped kinematics
as available so existing runtime behavior remains unchanged.

## 4. PostgreSQL Write Semantics

`SaveFlightStates` now writes:

```text
PostgreSQL NULL when availability=false
the numerical value when availability=true
```

Real zero values remain valid database values.

No migration is required because the existing flight-state telemetry columns
already accept `NULL`.

## 5. PostgreSQL Read Semantics

General Flight State and reconciliation readers now use PostgreSQL nullable
types for:

```text
velocity_mps
heading_degrees
vertical_rate_mps
on_ground
```

They restore both the value and the availability flag.

Latitude and longitude are no longer fabricated with zero fallbacks. Readers
exclude historical rows without a usable position.

## 6. Traffic Read Boundary

The current traffic contract remains non-nullable for map rendering.

Therefore Traffic does not fabricate fallback values. It selects only rows
where the required display telemetry exists:

```text
latitude
longitude
velocity
heading
on_ground
```

This preserves the current HTTP contract while preventing unavailable
kinematics from appearing as zero.

## 7. Airspace Intelligence Boundary

Airspace calculations require complete motion telemetry.

The PostgreSQL observation reader now selects only observations where:

```text
velocity
heading
vertical rate
on_ground
```

are present.

Missing telemetry is excluded from proximity, interaction, and separation
calculations instead of becoming a real numerical zero.

## 8. Data Quality Validation

The traffic validator now treats explicit unavailability as missing movement
telemetry.

Missing velocity, heading, vertical rate, or on-ground state produces a
position-only completeness classification rather than a false complete
observation.

Legacy states without an explicit availability contract retain their previous
validation behavior.

## 9. Final Correctness Audit Expansion

`backendfinalaudit` now verifies the complete chain:

```text
FlightState availability fields and methods
OpenSky finite optional mapping
Airplanes.live explicit availability
PostgreSQL nullable writes
Flight State nullable reads
Reconciliation nullable reads
Traffic complete-row selection
Airspace complete-row selection
Projection complete-row selection
validator availability awareness
```

The audit fails if dangerous numerical `COALESCE` expressions return in the
protected production readers.

## 10. Acceptance Scenarios

The increment must prove:

```text
OpenSky nil velocity -> PostgreSQL NULL semantics
OpenSky nil heading -> PostgreSQL NULL semantics
OpenSky nil vertical rate -> PostgreSQL NULL semantics

OpenSky velocity zero -> available zero
OpenSky heading zero -> available zero
OpenSky vertical rate zero -> available zero

PostgreSQL NULL -> availability=false
PostgreSQL zero -> availability=true

Traffic excludes incomplete display telemetry
Airspace excludes incomplete analytical telemetry
validator reports explicit unavailability as missing
Projection continues to require complete kinematics
```

## 11. Verification

Acceptance requires:

```text
focused provider, domain, repository, validator, Traffic, Airspace, and
Projection tests
expanded backendfinalaudit
strict projectaudit
race detector
complete Go build
go vet
complete Go tests
frontend dependency security verification
frontend lint
frontend TypeScript validation
frontend production build
backend Docker image build
git diff check
```

After this stage, the previously identified nullable telemetry correctness risk
is closed end-to-end for the production provider, persistence, and analytical
read path.

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
