# Document 97 — Analytical Contributor Semantics Hardening

Status: Implemented Engineering Increment v1.0
Project: Global Flight Analytics
Baseline: `1ddb65c5e5471ce180314cc38a4b6d7baad80cd3`

## 1. Purpose

This increment closes three correctness findings from the original Analytical
Core Foundation review:

```text
eligibility must run before aircraft-level deduplication;
observations materially in the future must not contribute to analytics;
traffic-density arithmetic must reject non-finite area values.
```

## 2. Contributor ordering contract

For Active Aircraft and Traffic Density the production order is now:

```text
evaluate every trajectory for capability eligibility;
retain eligible contributors;
deduplicate eligible contributors by aircraft identity;
calculate the metric;
publish duplicate and exclusion evidence.
```

An ineligible newer trajectory can no longer replace an eligible older
trajectory for the same aircraft.

Denied contributors remain visible in the scope summary. Duplicate removal is
performed only among eligible contributors.

## 3. Future observation contract

Every analytical capability uses a bounded clock-skew tolerance:

```text
default maximum future observation skew: 30 seconds
```

A trajectory whose end time exceeds the evaluation time plus the configured
tolerance receives:

```text
future_observation
```

and is excluded from the affected capability.

A timestamp exactly at the tolerance boundary remains eligible.

## 4. Traffic Density finite-number contract

Traffic Density now rejects:

```text
negative active-aircraft counts;
zero or negative area;
NaN area;
positive or negative infinite area.
```

The domain calculator owns these checks independently of HTTP parsing.

## 5. Verification

The installer executes targeted analytical tests, targeted race tests, complete
backend tests, Go static analysis, architecture audits and whitespace checks.

## 6. Remaining Analytical Core review scope

This increment does not close the full review. Remaining accepted work includes:

```text
airport-owned Airport Activity classification;
geographically bound Traffic Density requests;
server-owned production Coverage Score and Data Freshness;
strict analytical provenance and safe public failures;
reference-time and UUID normalization;
obsolete analytical foundation classification;
metric identifier consolidation.
```
