# Frontend Research Snapshot Export

Status: Implementation prepared; exact-commit Continuous Integration closure pending  
Project: Global Flight Analytics  
Reviewed baseline: `fc7d0cb307b9c1a3c326908df4c1dcf2755042b9`  
Date: 2026-08-01

## 1. Purpose

This increment adds a browser-owned research export path for the current regional
traffic snapshot. The product can now move observed records from the visual workspace
into analysis tools without adding a backend endpoint, server-side export queue or a
new dependency.

## 2. Export formats

CSV uses a fixed schema containing snapshot provenance followed by every typed traffic
field. Records are sorted deterministically by normalized ICAO24, observation time and
callsign. Commas, quotes and line breaks are escaped according to standard CSV rules.

GeoJSON uses a FeatureCollection with Point geometries in longitude-latitude order.
Top-level metadata records the schema version, region, successful snapshot time,
export-generation time, selected aircraft context, source-record count, exported
feature count, invalid-coordinate exclusion count and the evidence boundary.

## 3. Coordinate policy

CSV retains all current snapshot records, including records whose coordinate fields
cannot be represented as a valid geographic point. GeoJSON accepts only finite
coordinates within latitude `[-90, 90]` and longitude `[-180, 180]`. Rejected point
records are never silently lost: their count is published in metadata and in the
browser-side export model result.

## 4. Evidence boundary

The export represents one current API response. It is not a historical series,
provider-health report, operational navigation product, authoritative flight-status
feed or proof of traffic growth. Browser export generation does not modify the source
records or increase their analytical confidence.

## 5. Verification

Dependency-free tests verify:

1. fixed CSV schema and deterministic record ordering;
2. CSV escaping for commas, quotes and line breaks;
3. safe serialization of null and non-finite optional numbers;
4. GeoJSON coordinate order and provenance metadata;
5. invalid-coordinate exclusion accounting;
6. deterministic filenames based on the successful snapshot timestamp;
7. explicit rejection of invalid generation timestamps.

Full frontend tests, ESLint, TypeScript, production build and dependency-security
policies remain required gates.

## 6. Scope boundary

This increment does not add backend export endpoints, compressed archives, historical
querying, authentication, cloud storage, analytics telemetry, spreadsheet generation,
new packages or lockfile changes.
