# Document 36 — Free Data Source and Evidence Boundaries

Status: Architecture Baseline v1.1
Project: Global Flight Analytics
Scope: Immutable source constraints, OpenSky integration boundaries, temporal validity, attribution, and analytical claim limits

---

## 1. Non-Negotiable Project Constraints

Global Flight Analytics is permanently evaluated through three hard constraints:

```text
1. Only free data sources may be used.
2. The project owns no aircraft-surveillance collection infrastructure.
3. The project has no satellite-surveillance or commercial aviation-data access.
```

Consequences:

```text
no project-owned ADS-B receiver network
no project-owned ground stations
no first-party global surveillance feed
no satellite oceanic coverage
no licensed airline operations feed
no official airport operations feed
no air traffic control instruction feed
no certified aviation weather feed
```

These are architectural boundaries, not temporary deployment problems.

---

## 2. Evidence Vocabulary

Every external value must be classified as one of:

```text
observed  — reported by an identified external source
            and never described as project-owned observation

derived   — calculated from observed data by a declared algorithm

estimated — uncertain provider or project inference that must remain labelled

unknown   — insufficient evidence; no substitute explanation may be invented

blocked   — a claim that exceeds the available evidence or project scope
```

Observed does not mean complete, continuous, official, first-party, or newly received at the snapshot timestamp.

---

## 3. OpenSky Role

OpenSky Network is an optional free external community-surveillance provider for research and non-commercial use.

Permitted use:

```text
bounded regional state-vector observations
provider source and category semantics
bounded provider flight history
estimated departure and arrival airport context
experimental track context as secondary evidence
rate-limit and freshness evidence
```

Mandatory labels:

```text
external community observation
coverage may be incomplete
provider attribution required
non-commercial research use only
estimated airport context where applicable
experimental provider track where applicable
```

Forbidden claims:

```text
project-owned receiver observation
continuous global tracking
guaranteed oceanic tracking
satellite tracking
commercial flight operations data
official schedules, gates, delays, or delay causes
pilot intent
air traffic control instructions
certified separation monitoring
safety-critical decision support
```

---

## 4. Attribution and Usage Terms

Any public web page, article, public presentation, or other publication using OpenSky data must preserve the provider attribution requirement.

Required citation baseline:

```text
Matthias Schäfer, Martin Strohmeier, Vincent Lenders,
Ivan Martinovic and Matthias Wilhelm.
"Bringing Up OpenSky: A Large-scale ADS-B Sensor Network for Research".
In Proceedings of the 13th IEEE/ACM International Symposium on
Information Processing in Sensor Networks, pages 83-94, April 2014.
```

Provider identity:

```text
The OpenSky Network
https://opensky-network.org
```

The free research feed must not be presented as a licensed commercial real-time aviation-data product. Commercial schedules, gates, delays, official delay causes, and other information that cannot be derived from ADS-B observations remain unavailable.

---

## 5. Authentication, Polling, and Deployment Policy

The REST integration uses OAuth2 Client Credentials when free account credentials are configured. Anonymous access remains an explicit reduced-capability fallback.

```text
authenticated minimum state polling interval: 5 seconds
anonymous minimum state polling interval: 10 seconds
anonymous historical state request: blocked
authenticated historical state lookback: maximum 1 hour
```

Regional bounding boxes are mandatory for normal ingestion. Global requests are prohibited by project policy even when the provider endpoint technically permits them.

Rate-limit headers are evidence and must be retained:

```text
X-Rate-Limit-Remaining
X-Rate-Limit-Retry-After-Seconds
```

Provider access from large cloud-hosting IP ranges is not guaranteed. OpenSky must therefore remain an optional provider behind provider health, budget, fallback, and availability controls. The project must not depend on OpenSky as its only production ingestion source.

---

## 6. State Vector Temporal Semantics

An OpenSky State Vector is a server-side summary whose fields may have different source timestamps. It is not one simultaneous sensor packet.

The response snapshot timestamp is rounded to a whole second while position, speed, callsign, and other fields may have been updated at different moments.

Provider validity boundary:

```text
maximum last-known position reuse window: 15 seconds
position older than 15 seconds: unavailable or blocked from observed-position analytics
missing position: must remain missing
missing speed: must remain missing
no interpolation may be published as an observed OpenSky position
```

A position within the fifteen-second provider window may still be a reused last-known position. It is usable as provider-valid external evidence, but its age must be retained and disclosed.

The adapter must preserve:

```text
snapshot time
position time
last contact time
position age
last-contact age
position validity status
position usability decision
limitations
```

---

## 7. State Vector Contract

The OpenSky adapter preserves all eighteen documented State Vector fields, including:

```text
ICAO24
callsign
origin country
time of last position
last contact
longitude
latitude
barometric altitude
on-ground flag
velocity
true track
vertical rate
sensor identifiers when available
geometric altitude
squawk
special-purpose indicator
position source
aircraft category
```

Position source and aircraft category are provenance and classification evidence. They are not proof of coverage quality, receiver ownership, operational identity, or simultaneous measurement.

---

## 8. Historical and Track Boundaries

OpenSky flight and airport records are bounded provider products. Estimated airport fields must never be displayed as official airport records.

The OpenSky track endpoint is experimental. It may be used for comparison and diagnostics, but it must not replace:

```text
Canonical FlightState
Track Builder
Coverage Gap Detector
TrajectorySegment
FlightTrajectory
Trajectory Quality
```

The project-owned historical aggregate store remains the primary analytical history.

---

## 9. Executable Enforcement

The executable policy is implemented in:

```text
apps/api/internal/analytics/sourceconstraints
```

The OpenSky contract foundation is implemented in:

```text
apps/api/internal/integrations/opensky
```

Runtime-enforced additions include:

```text
provider attribution obligation
non-commercial usage boundary
cloud-hosting availability warning
fifteen-second provider position-validity boundary
per-field timestamp preservation
blocking of stale or missing position as observed evidence
```

Verification command:

```text
go run ./cmd/verify-source-constraints
```

The command must show allowed, limited, and blocked decisions for every declared capability.

---

## 10. Integration Boundary

This document and the executable policy do not automatically make OpenSky the active production ingestion provider.

Production activation requires a separate verified wiring increment covering:

```text
provider policy registration
provider budget configuration
fallback ordering
canonical FlightState mapping
persistence of new provenance and temporal-validity fields
migration review
runtime PostgreSQL verification
cloud-hosting connectivity verification
HTTP and frontend attribution disclosure verification
```

Until then, the existing production ingestion path remains unchanged.
