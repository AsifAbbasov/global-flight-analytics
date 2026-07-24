# Document 98 — Airport and Geographic Metric Integrity

Status: Implemented Engineering Increment v1.0
Project: Global Flight Analytics
Baseline: `c5fd1f32273af9215df9d83d1d40c227d3740646`

## 1. Purpose

This increment closes two semantic findings from the original Analytical Core
Foundation review:

```text
Airport Activity must be owned by a concrete airport and classified by the server;
Traffic Density must use one server-owned geographic scope for both contributors and area.
```

## 2. Airport Activity contract

The production endpoint now requires:

```text
airport_icao
optional radius_kilometers
optional window_minutes
optional limit
```

The server loads the airport from PostgreSQL, derives a bounded geographic
query, loads recent trajectories, applies eligibility, removes duplicate
eligible trajectories, and classifies movement from trajectory crossings of
the airport geofence.

The client no longer submits separate arrival and departure trajectory lists.
Unrelated and ambiguous trajectories are excluded with explicit limitations.

## 3. Traffic Density contract

Traffic Density now requires a configured region. The same region bounds own
both the contributor query and the calculated area.

The client-provided `area_square_kilometers` parameter is rejected.

## 4. Verification

The installer executes compile-only checks, targeted tests, race tests, the
complete backend test suite, Go vet, all existing architecture audits, static
contract checks, and whitespace validation.

## 5. Remaining Analytical Core review scope

```text
server-owned production Coverage Score and Data Freshness;
strict analytical provenance and safe public failure messages;
reference-time and UUID canonicalization;
obsolete analytical foundation classification;
metric identifier consolidation.
```
