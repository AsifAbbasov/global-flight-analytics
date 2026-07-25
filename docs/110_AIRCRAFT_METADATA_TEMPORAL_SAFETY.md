# Document 110 — Aircraft Metadata Temporal Safety

Status: IMPLEMENTED
Baseline commit: a4c563112abd459c90e23e33d191c5f059e5044f

## 1. Problem

The aircraft feature provider cached current metadata only by ICAO24. Historical
feature materialization could therefore consume registration, model, airline or
country values whose underlying records were updated after the requested feature
`AsOfTime`.

The repository does not own a complete versioned aircraft-history table, so it
must not invent historical values.

## 2. Conservative temporal contract

The PostgreSQL aircraft read model now returns one aggregate metadata update
timestamp. It is the greatest `updated_at` value across:

- aircraft;
- aircraft model;
- airline;
- country.

The extractor passes the feature request `AsOfTime` into the aircraft provider.
The provider applies temporal policy after cache lookup and after in-flight
request coalescing.

If aggregate metadata was updated after `AsOfTime`, the aircraft feature group
is returned as unavailable with:

```text
aircraft_metadata_newer_than_feature_as_of
```

No current value is presented as historical evidence.

## 3. Cache semantics

The cache remains keyed by normalized ICAO24 because it stores one current
metadata result. Temporal acceptance is evaluated independently for every
request. A recent request and a historical request may therefore share one
lookup without sharing one temporal decision.

The aggregate metadata timestamp is included in `AircraftFeatures`, and thus in
the extraction input fingerprint introduced by Document 109.

## 4. Generation boundary

Aircraft provider, extractor, extractor composition and feature processing
generations advance to version 3 where applicable. Earlier snapshots remain
readable through their stored processing versions.

## 5. Permanent evidence

```text
AIRCRAFT_METADATA_TEMPORAL_GATE=ENFORCED
AIRCRAFT_CACHE_AS_OF_ISOLATION=ENFORCED
FUTURE_AIRCRAFT_METADATA_LEAKAGE=CLOSED
HISTORICAL_AIRCRAFT_VALUES_NOT_INVENTED=ENFORCED
```
