# Frontend Traffic Data Quality Lens

Status: Implementation prepared; exact-commit Continuous Integration closure pending  
Project: Global Flight Analytics  
Reviewed baseline: `aacaf25fec5fbac20a0391d825e1b48d060aaa56`  
Date: 2026-08-01

## 1. Purpose

This increment adds a browser-side data-quality lens for the current regional traffic
snapshot. It turns fields already present in the typed traffic response into explicit,
separate structural checks without inventing a composite confidence score or changing
backend analytical contracts.

## 2. Structural evidence dimensions

The lens publishes independent counts and shares for:

1. valid six-character hexadecimal ICAO24 identifiers;
2. unique normalized aircraft identifiers and duplicate records;
3. latitude and longitude values inside valid geographic bounds;
4. non-negative velocity and heading inside the `[0, 360)` range;
5. parseable provider observation timestamps;
6. observations inside a five-minute browser recency window;
7. future-dated observations beyond a one-minute clock-skew tolerance;
8. usable observed altitude among airborne aircraft only;
9. callsign, aircraft-model, airline and origin-country attribution completeness.

A record is structurally usable only when its identifier, coordinates, timestamp,
motion values and required altitude semantics are usable. Optional descriptive
attribution is reported separately and does not invalidate positional evidence.

## 3. Deterministic issue register

Detected issues are sorted by severity, affected-record count and stable issue key.
Critical issues cover invalid identifiers, coordinates, timestamps and motion values.
Warnings cover stale or future-dated observations, duplicate identifiers and missing
airborne altitude. Missing descriptive labels remain informational.

Every issue preserves its own denominator. Airborne altitude gaps use airborne aircraft,
while optional attribution and structural checks use the complete snapshot.

## 4. Evidence boundary

The browser response timestamp is only a reference for client-side recency presentation.
This lens does not replace server-side data-quality metrics, source health, ingestion
health, provider timestamps, historical analysis, authoritative flight status or
operational safety assessment.

No aggregate quality grade is generated. Each dimension remains inspectable so that a
polished interface does not hide which evidence is complete and which evidence is not.

## 5. Verification

Dependency-free regression tests verify:

1. stable empty-snapshot semantics;
2. complete recent record behavior;
3. independent coordinate, motion and timestamp failures;
4. normalized duplicate ICAO24 detection;
5. airborne-only altitude denominators;
6. trimmed identity-attribution completeness;
7. exact recency boundaries and deterministic issue ordering.

Full frontend tests, ESLint, TypeScript, production build and dependency-security
policies remain required gates.

## 6. Scope boundary

This increment does not add backend endpoints, database writes, provider-health claims,
historical persistence, telemetry, user scoring, safety guidance, new packages or
lockfile changes.
