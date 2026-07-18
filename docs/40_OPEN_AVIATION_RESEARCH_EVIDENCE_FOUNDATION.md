# Document 40 — Open Aviation Research Evidence Foundation

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: canonical observation metadata, bounded transponder evidence, external research dataset governance, and reproducible offline benchmark contracts

## 1. Purpose

This foundation strengthens the analytical core without turning public research datasets into hidden production dependencies.

The implementation preserves observation metadata that providers already expose:

```text
squawk code
Special Purpose Indicator
position source
aircraft category
aircraft category availability
```

It also creates executable governance for selected OpenSky scientific datasets.

## 2. Canonical Evidence Boundary

A special transponder code is an observed external field.

Allowed claim:

```text
Observed transponder code 7700.
```

Blocked claim:

```text
Confirmed aircraft emergency.
```

The analytical result cannot infer incident cause, pilot intent, air traffic control instructions, unlawful interference, radio failure, or operational response from a transponder code alone.

## 3. Selected Research Datasets

Adopted for bounded offline research:

```text
OpenSky emergency reference dataset
OpenSky climbing aircraft dataset
March 2026 one-day Trino snapshot, allowlisted tables only
weekly Monday State Vector samples
OpenSky and EUROCONTROL PRC take-off weight dataset
```

Deferred:

```text
raw physical-layer data
LocaRDS
COVID-19 dataset
mixed-source aircraft metadata database
GICB capabilities dataset
```

Blocked:

```text
OpenSky ADS-C dataset
readsb_adsc_sv
```

ADS-C remains blocked because it is satellite-derived and can expose navigation intent and oceanic surveillance evidence outside the fixed project boundary.

## 4. Manifest Gate

No dataset can be imported without a manifest containing:

```text
dataset identifier and version
selected files
file formats
file sizes
SHA-256 checksums
total byte limit
record limit
region filter when required
selected Trino tables
licence review confirmation
attribution confirmation
offline-only confirmation
production-dependency prohibition
```

The manifest gate rejects:

```text
unselected datasets
deferred datasets
blocked datasets
blocked tables
files without checksums
unbounded bytes or records
missing region filters
missing licence review
missing attribution
production dependencies
```

## 5. Benchmark Contracts

The architecture defines bounded plans for:

```text
transponder evidence retention
climb prediction
OpenSky schema compatibility
external historical replay
take-off weight estimation
```

These plans define metrics and applicability guards but do not download data or train models.

## 6. Persistence

Canonical flight states persist provider observation metadata in PostgreSQL.

The database preserves the distinction between:

```text
aircraft category unavailable
```

and:

```text
aircraft category observed with value zero
```

The special-code index supports bounded analytical reads without creating an operational alert system.

## 7. Non-Goals

This foundation does not provide:

```text
confirmed emergency detection
safety-critical alerts
official incident information
satellite surveillance
pilot intent
air traffic control instructions
live measured take-off weight
automatic external dataset downloads
machine-learning model training
production reliance on static research datasets
```
