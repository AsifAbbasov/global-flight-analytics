# Document 102 — Analytical Core Review Closure

Status: Implemented Engineering Contract v1.0
Project: Global Flight Analytics
Closure baseline: `e48cb27655326fc6cc41d176a50120cdbf1ced6e`
Predecessor Continuous Integration evidence: Backend CI run `30128888158`

## 1. Purpose

This document completes classification and remediation of every finding from the
original Analytical Core Foundation review.

The closure policy follows the project Code Review Standard:

```text
fixed;
not applicable;
deliberately retained with evidence;
deliberately rejected with evidence;
deferred with owner, risk and revisit condition;
suggestion or nit that does not block merge.
```

No original finding is left implicit.

## 2. Closure totals

```text
Original findings: 19
Fixed: 14
Deliberately retained: 3
Rejected or non-blocking: 2
Not applicable: 0
Deferred: 0
Unclassified: 0
```

## 3. Original finding register

| Identifier | Original finding | Disposition | Primary evidence |
| --- | --- | --- | --- |
| AC-01 | Eligibility must precede aircraft-level deduplication. | FIXED | Document 97 and execution ordering tests. |
| AC-02 | Materially future observations must not contribute to metrics. | FIXED | Document 97 and trajectory eligibility tests. |
| AC-03 | Airport Activity was not owned by a concrete airport. | FIXED | Document 98 and airport-owned HTTP tests. |
| AC-04 | Traffic Density numerator and denominator used unrelated scopes. | FIXED | Document 98 and region-owned area tests. |
| AC-05 | Calculator and registry appeared to form a parallel runtime architecture. | DELIBERATELY_RETAINED | Document 100 classifies them as compatibility foundations outside server composition. |
| AC-06 | Metric execution exposed concrete dependencies and internal objects. | FIXED | Document 100, narrow executor interface and getter-removal tests. |
| AC-07 | Traffic Density permitted `NaN` or infinite area values. | FIXED | Document 97 and finite-number calculator tests. |
| AC-08 | Coverage Score and Data Freshness trusted caller snapshot parameters. | FIXED | Documents 99 and 101 plus server-owned production snapshot tests. |
| AC-09 | Analytical source provenance was incomplete or placeholder-based. | FIXED | Document 99, strict source validation and server query provenance. |
| AC-10 | `RecentRequest.Normalize` accepted a zero reference time. | FIXED | Document 100 and zero-time regression tests. |
| AC-11 | Parsed UUID identifiers were not canonicalized before deduplication. | FIXED | Document 100 and canonical UUID tests. |
| AC-12 | Public numeric precision had no explicit contract. | FIXED | Section 5 of this document and permanent source audit. |
| AC-13 | Long functions should fail a universal mechanical line threshold. | REJECTED_NON_BLOCKING | Section 8 and the existing code-review mechanical-rule policy. |
| AC-14 | `Value` plus `HasValue` was treated as redundant. | DELIBERATELY_RETAINED | Section 6 and status validation invariants. |
| AC-15 | Nullable analytical sections were treated as an unconditional design defect. | DELIBERATELY_RETAINED | Section 7 and status-discriminated response contracts. |
| AC-16 | Names containing `And` or `With` should be mechanically forbidden. | REJECTED_NON_BLOCKING | Section 8 and the existing code-review mechanical-rule policy. |
| AC-17 | Metric identifiers were inconsistent across packages and clients. | FIXED | Document 100 and canonical `traffic.*` identifiers. |
| AC-18 | Public analytical failures could expose raw operation error text. | FIXED | Document 99 and stable public failure regression tests. |
| AC-19 | Source names used a manual quadratic sorting helper. | FIXED | Standard-library `sort.Strings` and the final audit gate. |

## 4. Fixed correctness findings

Documents 97 through 101 form the implementation evidence chain:

```text
Document 97 — contributor ordering, future-observation exclusion and finite arithmetic;
Document 98 — airport ownership and one geographic Traffic Density scope;
Document 99 — strict provenance, placeholder rejection and safe public failures;
Document 100 — reference time, UUID canonicalization, Metric IDs and dependency boundaries;
Document 101 — server-owned Coverage Score and Data Freshness evidence.
```

The final closure increment additionally removes the manual quadratic source-name
sort and replaces it with the Go standard library.

## 5. Public numeric precision contract

Analytical API values are JSON numbers produced from their native Go numeric
values.

The backend:

```text
does not round metric values before HTTP publication;
does not convert metric values to formatted strings;
does not apply display precision to the response contract;
publishes `value` only when `has_value` is true.
```

The frontend may format numbers for a specific visual surface:

```text
ratios are displayed as percentages with one fractional digit;
density uses four significant display digits;
configured area uses whole-square-kilometre display formatting.
```

That formatting is presentation-only. It is not persisted, sent back to the
server or treated as the analytical value.

Consumers requiring deterministic comparisons must use the JSON number and the
metric identifier, not rendered text.

## 6. `Value` and `HasValue`

A numeric zero is a legitimate analytical result:

```text
zero active aircraft;
zero airport movements;
zero coverage;
zero freshness;
zero traffic density.
```

The zero value therefore cannot represent both a valid metric and absence.

`HasValue` is retained as the explicit presence discriminator:

```text
complete and limited results require `HasValue = true`;
denied and failed results require `HasValue = false`;
the HTTP response includes `value` only when `HasValue = true`.
```

`ValueOrZero` remains a compatibility helper on the internal generic value
object. The public response mapper does not use it and cannot convert absence
into a published zero.

## 7. Nullable result sections

Nullable result sections are status-discriminated evidence, not unstructured
optional state.

```text
eligibility may be absent when no trajectory capability decision exists;
failure is required only for failed results and forbidden otherwise;
data quality may be absent when no report can be built;
confidence report may be absent for denied or failed execution paths;
source observation bounds are both present or both absent.
```

Validation enforces the allowed combinations. The TypeScript contract mirrors
the same optional sections and retains `has_value` as the value discriminator.

Replacing all nullable fields with empty objects would erase the difference
between unavailable evidence and an evaluated empty result.

## 8. Rejected mechanical findings

### 8.1 Universal function-length threshold

A universal line-count failure rule is rejected.

Functions are decomposed when they mix responsibilities, hide resource
ownership, make policy untestable or create a demonstrated maintenance risk.
They are not split solely to satisfy an arbitrary line number.

The existing Code Review audit permanently rejects unsupported mechanical rules.

### 8.2 Mechanical `And` and `With` naming ban

The presence of `And` or `With` in an identifier is not itself a correctness,
cohesion or dependency defect.

Names remain subject to intent, ownership and readability review. A universal
substring ban is a non-blocking style preference and is rejected as a release
gate.

## 9. Compatibility package classification

`analytics/calculator` and `analytics/registry` remain compatibility foundation
packages.

They are:

```text
compiled and tested;
not stored by the runtime Executor;
not exposed through executor dependency getters;
not imported by the production server composition roots;
not required by the metric execution service interface.
```

A future breaking-version cleanup may remove their compatibility constructor
surface. That cleanup is not required to close the current production review.

## 10. Permanent verification

The repository adds:

```text
apps/api/tools/analyticalcorefinalaudit
```

The tool verifies:

```text
all nineteen finding classifications;
Documents 97 through 102 and public closure status;
eligibility-before-deduplication ordering;
future-observation and finite-number guards;
airport and region ownership;
server-owned quality evidence;
strict provenance and sanitized failures;
reference-time and UUID canonicalization;
canonical Metric IDs;
numeric presence and precision contracts;
compatibility-package runtime isolation;
frontend request and query ownership;
standard-library source sorting;
Backend and Frontend Continuous Integration reachability.
```

Backend Continuous Integration runs the tool in strict mode.

The Backend workflow path filter includes the analytical frontend API client,
React Query hooks and Analytics Overview. Frontend Continuous Integration
continues to run lint, TypeScript validation and a production build for every
`apps/web/**` change.

## 11. Required closure gates

Formal closure requires the closure commit to pass:

```text
Go formatting;
Analytical Core final audit unit tests;
complete backend tests;
analytical race tests;
Go vet;
project architecture and contract audit;
code review policy audit;
Stage 14 final audit;
strict Analytical Core final source audit;
Frontend ESLint;
Frontend TypeScript validation;
Frontend production build;
PostgreSQL 16 Integration;
Backend Container build and health smoke test.
```

The predecessor server-owned quality increment passed Backend CI run
`30128888158` on commit `e48cb27655326fc6cc41d176a50120cdbf1ced6e`. The closure commit must independently pass
the same required backend jobs and Frontend Continuous Integration.

## 12. Closure statement

When every gate in Section 11 passes on the same closure commit:

```text
Analytical Core Foundation review: CLOSED
ANALYTICAL_CORE_REVIEW_STATUS=CLOSED
Open release blockers: 0
Open required changes: 0
Deferred findings: 0
Unclassified original findings: 0
Release decision: ACCEPTABLE
```

Post-closure changes must preserve the permanent audit. A new confirmed blocker
requires a new document and an explicit status change; it must not silently
rewrite this evidence register.

## 13. Source-formatting resilience

The permanent source audit verifies semantic ownership rather than one exact
formatter layout.

Go permits a selector expression to be split after the period, and TypeScript
formatting may place a `URLSearchParams.set` argument on the following line.
The audit therefore removes formatting whitespace only for explicitly declared
compact source contracts.

The sampling interval is verified in
`analytical_production_snapshot.go`. The stale threshold is verified in its
actual public-handler owner, `analytical_metrics.go`. The frontend radius query
parameter is verified across Prettier-compatible line breaks.

Regression tests prove both directions:

```text
valid multiline Go and TypeScript contracts pass;
a genuinely absent compact contract fails.
```
