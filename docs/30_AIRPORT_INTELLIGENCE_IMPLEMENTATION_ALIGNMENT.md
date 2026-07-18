# Document 30 — Airport Intelligence Implementation Alignment

Status: Implementation Alignment v1.0  
Project: Global Flight Analytics  
Scope: Implemented Airport Intelligence domain foundation, analytical contracts, limitations, and next integration steps

---

## 1. Purpose

This document records the Airport Intelligence domain implementation completed after the analytical research audit and the correction of the airport ranking contract.

It does not replace Documents 08, 22, 23, 24, 25, 27, or 28. It records how the current Go implementation aligns with them and which capabilities remain intentionally outside the completed increment.

---

## 2. Implementation Status

The following packages are implemented under:

```text
apps/api/internal/airportintelligence
```

Implemented packages:

```text
passport
statistics
ranking
overview
history
trends
```

Implemented capabilities:

```text
Airport Passport domain model and Builder
Airport Passport Service
Airport Statistics Calculator
Airport Activity Score
Airport Data Confidence
Airport Ranking
Airport Overview Assembler
Airport History Builder
Airport Trends Analyzer
```

The completed code is domain-focused. It is intentionally independent from HTTP, PostgreSQL, JSON transport models, and external provider response formats.

---

## 3. Research-to-Code Classification

The current implementation uses the decision labels defined by Document 28.

### 3.1 SOURCE-BACKED

The following requirements are directly aligned with approved project documentation:

```text
Airport Activity Score uses the range 0–100.
Airport Activity Score represents statistical activity.
Airport Activity Score considers movements, active routes, observations, and traffic intensity.
Airport profile data and analytical data remain separate concerns.
Weak data quality must not be presented as strong analytical certainty.
Historical gaps must remain visible instead of being silently hidden.
```

### 3.2 PROJECT-DERIVED

The current baseline formulas are project-derived and deterministic:

```text
Airport Activity Score weights:
25% total movements
25% active routes
25% observed samples
25% movements per hour
```

```text
Airport Data Confidence weights:
50% coverage score
50% freshness score
```

These weights are explicit product policies. They are not presented as universal scientific constants. They may be recalibrated later through historical replay and measured product data.

### 3.3 EXPERIMENTAL

Any future adaptive weighting, learned ranking, anomaly detection, forecasting, or causal explanation must remain experimental until a measurement plan, historical replay method, baseline comparison, and limitations contract exist.

---

## 4. Airport Statistics Contract

Airport Statistics describes one airport over one explicit time window.

The current model includes:

```text
ICAO code
window start
window end
arrivals
departures
total movements
arrival share
departure share
movements per hour
active aircraft
active routes
observed samples
expected samples
coverage score
freshness score
latest observation time
generation time
```

Core rules:

```text
ICAO codes are normalized.
Times are normalized to Coordinated Universal Time.
Counters must be non-negative.
Expected samples must be positive.
Coverage Score remains within 0–1.
Freshness Score remains within 0–1.
The latest observation must belong to the valid analytical time range.
The result must be deterministic.
The input must not be mutated.
```

---

## 5. Airport Activity Score Contract

Airport Activity Score answers:

```text
How active is this airport relative to the current comparison set?
```

Range:

```text
0–100
```

Components:

```text
Total Movements Score
Active Routes Score
Observed Samples Score
Traffic Intensity Score
```

Default policy:

```text
Activity Score =
0.25 × normalized total movements
+ 0.25 × normalized active routes
+ 0.25 × normalized observed samples
+ 0.25 × normalized movements per hour
```

Normalization is relative to the maximum valid value in the current comparable airport set.

Consequences:

```text
The score is suitable for ranking airports inside the same analytical comparison window.
The score is not an absolute worldwide airport classification.
A score from one unrelated comparison set must not be compared directly with a score from another set without a shared normalization policy.
```

---

## 6. Airport Data Confidence Contract

Airport Data Confidence answers:

```text
How trustworthy is the data used to describe the airport activity?
```

Range:

```text
0–100
```

Default policy:

```text
Data Confidence =
0.50 × Coverage Score
+ 0.50 × Freshness Score
```

The result is converted from the 0–1 quality scale to 0–100.

Important separation:

```text
Activity Score measures observed operational activity.
Data Confidence measures trust in the supporting observations.
Data Confidence must not inflate or suppress Activity Score.
```

This separation prevents a quiet airport with excellent data from appearing highly active and prevents an active airport with weak coverage from appearing operationally inactive.

---

## 7. Airport Ranking Contract

Airport Ranking:

```text
requires comparable statistics windows
rejects duplicate ICAO codes
normalizes ICAO codes
rejects invalid numerical values
sorts deterministically
preserves input values
returns positions beginning with 1
publishes Activity Score and Data Confidence separately
```

Deterministic tie resolution uses analytical values and normalized ICAO code instead of unstable collection order.

---

## 8. Airport Overview Contract

Airport Overview composes already calculated domain results. It does not duplicate analytical calculations.

It combines:

```text
Airport Passport
Airport Statistics
optional Airport Ranking information
Activity Score
Data Confidence
```

The assembler verifies identity and time consistency between child objects.

---

## 9. Airport History Contract

Airport History builds an immutable historical sequence from Airport Statistics.

Rules:

```text
Statistics are sorted by time.
Duplicate windows are rejected.
Overlapping windows are rejected.
Honest gaps are allowed and preserved.
Different airports cannot be mixed.
Invalid counters and scores are rejected.
Input data is not mutated.
```

A historical gap is not automatically interpolated and is not silently treated as observed activity.

---

## 10. Airport Trends Contract

Airport Trends is descriptive historical analytics.

It is not forecasting, anomaly detection, causal analysis, or machine learning.

The analyzer reports:

```text
increasing activity
decreasing activity
stable activity
absolute movement change
movements-per-hour change
active-routes change
coverage change
freshness change
peak window
gap count
gap duration
observed duration
continuity score
```

Comparison rules:

```text
Only windows with equal duration are compared.
Trend direction uses movements per hour.
Percentage change is unavailable when the initial value is zero.
Historical gaps remain explicit.
The analyzer does not invent reasons for activity changes.
```

Continuity Score baseline:

```text
Continuity Score =
observed window duration / total historical period duration
```

This is a PROJECT-DERIVED baseline and must be recalibrated only after measured historical use.

---

## 11. Validation and Verification

The implemented Airport Intelligence domain has been verified with:

```text
targeted unit tests
invalid input tests
boundary tests
normalization tests
deterministic ranking tests
input non-mutation tests
static analysis through go vet
race detector execution
full backend go test ./...
git diff validation
```

The successful domain verification does not prove production performance, database correctness, HTTP contract correctness, or real-world analytical calibration.

---

## 12. Intentional Boundaries

The completed increment intentionally does not include:

```text
PostgreSQL repositories for Airport Intelligence
HTTP handlers and transport projections
OpenStreetMap infrastructure enrichment
Wikidata and Wikipedia enrichment
complete runway and terminal models
transport connections
airport facilities
popular route persistence
historical replay user interface
traffic forecasting
machine learning
anomaly detection
causal explanations
```

These omissions are deliberate. The current increment establishes domain contracts before infrastructure integration.

---

## 13. Known Limitations

The current Passport is a domain foundation, not the complete digital passport described by Document 08.

The current Activity Score is relative to a comparison set.

The current Data Confidence baseline uses coverage and freshness only. Historical continuity is exposed by Airport Trends and is not silently counted twice.

The current History and Trends packages analyze supplied statistics. They do not yet prove that the underlying statistics were produced from correctly separated individual flights.

Stable flight identity and conservative flight splitting remain mandatory architectural debts before advanced route, historical, and airport analytics are treated as fully reliable.

---

## 14. Next Required Work

After this document and the current domain packages are committed, the next sequence is:

```text
1. Stable Flight Identity
2. Flight Splitter
3. Conservative Track Splitting
4. Analytics Permission Flags
5. Unified Analytical Result
6. Confidence Report
7. Source Limitation Guard
8. Basic Origin Detection
9. Basic Destination Detection
10. Route Confidence
11. Basic Flight Phase Detection
12. Airport Intelligence PostgreSQL integration
13. Airport Intelligence HTTP integration
```

---

## 15. Completion Statement

The Airport Intelligence domain foundation is complete when:

```text
all packages are committed to Git
all targeted tests pass
race detector passes
full backend tests pass
documentation is committed
```

This completion statement applies only to the domain foundation.

It does not mean that the full Airport Intelligence product module, persistence layer, HTTP application layer, frontend integration, or research-grade calibration is complete.

<!-- STAGE-14-3-AIRPORT-INTELLIGENCE-PRODUCTION:DOCUMENT-30 -->

## Production Integration Amendment

Stage 14.3 completes the PostgreSQL and HTTP integration steps previously listed as pending. The implementation remains open-data research software and does not claim official airport operations, complete route coverage, or calibrated universal ranking weights. See Document 43.
