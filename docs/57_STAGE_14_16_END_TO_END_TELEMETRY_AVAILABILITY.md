# Document 57 — Stage 14.16 End-to-End Telemetry Availability

Status: Implementation Baseline v1.0
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
