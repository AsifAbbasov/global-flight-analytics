# Projection Route Frequency Review Hardening

Status: closed

```text
EVIDENCE_ISOLATION_COMMIT=c6fff15f8d0c770197db40a69d54f8856044d8d2
EVIDENCE_ISOLATION_GITHUB_ACTIONS_RUN=30534243693
EVIDENCE_ISOLATION_POSTGRESQL_16_INTEGRATION_JOB=90843560950
EVIDENCE_ISOLATION_BACKEND_QUALITY_JOB=90843560991
EVIDENCE_ISOLATION_BACKEND_RACE_SAFETY_JOB=90843560999
EVIDENCE_ISOLATION_BACKEND_CONTAINER_JOB=90843853362
POLICY_DECISION_INTEGRITY_COMMIT=ee7c79bc8213dc030ce0d98f13d1065c9bb96275
POLICY_DECISION_INTEGRITY_GITHUB_ACTIONS_RUN=30544636679
POLICY_DECISION_INTEGRITY_POSTGRESQL_16_INTEGRATION_JOB=90877578926
POLICY_DECISION_INTEGRITY_BACKEND_RACE_SAFETY_JOB=90877578928
POLICY_DECISION_INTEGRITY_BACKEND_QUALITY_JOB=90877579007
POLICY_DECISION_INTEGRITY_BACKEND_CONTAINER_JOB=90877915808
PERMANENT_AUDIT_COMMIT=6f039b33c96cdb67370158b0eda5d0fc87593de5
PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30548438062
PERMANENT_AUDIT_BACKEND_QUALITY_JOB=90890525039
PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=90890525126
PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=90890525150
PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=90890829745
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
PROJECTION_ROUTE_FREQUENCY_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_ROUTE_FREQUENCY_ENGINEERING_DEBT=CLOSED
PROJECTION_ROUTE_FREQUENCY_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_ROUTE_FREQUENCY_REVIEW_STATUS=CLOSED
```

## 1. Scope

This record covers the dedicated code review and hardening of:

```text
apps/api/internal/projectionintelligence/projectionroutefrequency
```

It also covers the production evidence boundary through:

```text
apps/api/internal/projectionintelligence/projectionread
apps/api/internal/projectionintelligence/projectionproduction
apps/api/cmd/verify-postgres-projection-intelligence-historical-http-api
```

The review is limited to historical route evidence isolation, logical-flight
deduplication, exposure-window integrity, scoring, decision semantics, validation,
fingerprinting, production fixture correctness, and permanent regression enforcement.

It does not claim empirical route-prediction accuracy, operational aviation
suitability, authoritative schedule knowledge, or that historical route frequency
proves future aircraft intent.

## 2. Confirmed production findings corrected

### 2.1 Current trajectory and current flight leakage

Route-history evidence now excludes both:

```text
current trajectory_id
current flight_id
```

The current prediction target therefore cannot act as historical support for itself.
The correction applies before observation count, recent support, distinct UTC-day
support, and latest-observation age are calculated.

### 2.2 Logical-flight evidence deduplication

Multiple trajectories belonging to one logical flight no longer inflate route
frequency. The PostgreSQL query builds a stable evidence identity:

```text
flight_id when available
trajectory_id only when flight_id is unavailable
```

The latest route result is selected per evidence identity before aggregate counts are
computed.

### 2.3 Distinct-flight ownership

The evaluator uses `DistinctFlightCount`, not raw trajectory count, for:

```text
observation support score
minimum historical support guard
```

`ObservationCount` remains available as evidence metadata, but it does not silently
override the logical-flight contract.

### 2.4 Full and recent exposure windows

The domain configuration now publishes:

```text
HistoryWindow
RecentWindow
```

The evaluator requires exact boundaries:

```text
WindowStart       = AsOfTime - HistoryWindow
WindowEnd         = AsOfTime
RecentWindowStart = AsOfTime - RecentWindow
```

Production policy fixes the windows at:

```text
HistoryWindow = 180 days
RecentWindow  = 30 days
```

The Projection Read policy requires exact equality between data-source windows and the
Route Frequency evaluator windows.

### 2.5 Configuration integrity

Configuration validation rejects:

```text
TargetDistinctDayCount greater than TargetObservationCount
TargetRecentObservationCount greater than TargetObservationCount
RecentWindow greater than HistoryWindow
zero route-confidence threshold
zero usable-score threshold
zero complete-score threshold
zero or negative component weights
non-finite values
component weights that do not sum to one
```

These checks prevent unreachable targets and nominally allowed zero-information
decisions.

### 2.6 Complete blocking evidence

The evaluator accumulates every applicable hard violation rather than returning only
the first matching reason. A blocked result can simultaneously explain:

```text
route confidence below minimum
distinct-flight support below minimum
distinct-day support below minimum
recent support below minimum
latest historical observation too old
aggregate score below minimum
```

Limitations are normalized, deduplicated, and deterministically ordered.

### 2.7 Weighted-score reconstruction

`Result.Validate()` independently reconstructs:

```text
sum(Component.Score multiplied by Component.Weight)
```

The reconstructed score must match the published aggregate score within the explicit
comparison tolerance. Component weights must be strictly positive and sum to one.

This validation exposed and corrected a stale production fixture:

```text
old published score: 0.85
actual weighted score: 0.94
```

### 2.8 Semantic fingerprint

The Route Frequency fingerprint binds:

```text
contract and fingerprint versions
Route Intelligence schema version and status
trajectory and flight identifiers
resolved route availability and canonical route key
same-airport state
route confidence
Route Intelligence provenance fingerprint
route-history fingerprint
full and recent window boundaries
historical support counts and latest observation
all policy thresholds
all component weights
```

Decision-relevant mutations therefore change the fingerprint.

## 3. Exact evidence already closed

The evidence-isolation commit completed all four Backend Continuous Integration jobs:

```text
COMMIT=c6fff15f8d0c770197db40a69d54f8856044d8d2
GITHUB_ACTIONS_RUN=30534243693
POSTGRESQL_16_INTEGRATION_JOB=90843560950
BACKEND_QUALITY_JOB=90843560991
BACKEND_RACE_SAFETY_JOB=90843560999
BACKEND_CONTAINER_JOB=90843853362
```

This exact evidence closes current-target leakage and multiple-trajectory inflation at
the committed baseline.

## 4. Policy and decision integrity evidence closed

The policy and decision integrity implementation completed all four exact Backend
Continuous Integration jobs:

```text
POLICY_DECISION_INTEGRITY_COMMIT=ee7c79bc8213dc030ce0d98f13d1065c9bb96275
POLICY_DECISION_INTEGRITY_GITHUB_ACTIONS_RUN=30544636679
POLICY_DECISION_INTEGRITY_POSTGRESQL_16_INTEGRATION_JOB=90877578926
POLICY_DECISION_INTEGRITY_BACKEND_RACE_SAFETY_JOB=90877578928
POLICY_DECISION_INTEGRITY_BACKEND_QUALITY_JOB=90877579007
POLICY_DECISION_INTEGRITY_BACKEND_CONTAINER_JOB=90877915808
```

This exact evidence closes exposure-window, policy-target, decision-reporting,
weighted-score, and semantic-fingerprint integrity at the committed baseline.

## 5. Deliberately retained and rejected recommendations

The following contracts are deliberately retained:

```text
fixed five-component versioned schema
float64 normalized scores with finite checks and explicit tolerance
idiomatic New(Config) (*Evaluator, error)
Usable compatibility field with cross-field validation
count-based support inside a strictly fixed exposure window
small local fingerprint and normalization helpers
```

The following mechanical recommendations are rejected:

```text
mandatory basis-point conversion for non-monetary scores
mandatory dynamic component registry
constructor criticism based only on nil plus error
blanket removal of every compatibility field
function-size criticism without a correctness or maintenance defect
renaming tests merely because a name contains the word And
```

A route-frequency rate per day is not required while the full exposure window is
strictly fixed, validated, and fingerprinted.

## 6. Permanent regression coverage

Permanent tests cover:

```text
incoherent target relationships
zero score thresholds
zero component weights
full-window mismatch
recent-window mismatch
all simultaneous blocking reasons
weighted-score mutation
current trajectory exclusion
current flight exclusion
logical-flight deduplication
canonical evidence identity
UTC distinct-day calculation
recent-window boundary
production policy window alignment
production fixture weighted-score consistency
```

## 7. Permanent audit gate

The source audit introduced by the current increment is:

```text
apps/api/tools/projectionroutefrequencyreviewaudit
```

Backend Continuous Integration executes:

```text
go run ./tools/projectionroutefrequencyreviewaudit -strict
```

The gate protects the production implementation, SQL evidence boundary, configuration,
fingerprint, regression tests, production fixture, Stage 9 markers, this authoritative
review record, Documentation Index registration, and workflow wiring.

The corresponding Backend Quality step is:

```text
Run projection route frequency review audit
```

The permanent-audit increment completed all four exact Backend Continuous Integration
jobs:

```text
PERMANENT_AUDIT_COMMIT=6f039b33c96cdb67370158b0eda5d0fc87593de5
PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30548438062
PERMANENT_AUDIT_BACKEND_QUALITY_JOB=90890525039
PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=90890525126
PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=90890525150
PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=90890829745
```

## 8. Formal closure

All confirmed findings are corrected and permanently protected. Exact Continuous
Integration is recorded for evidence isolation, policy and decision integrity, and the
permanent audit. No confirmed production-code defect is deferred, unclassified, or
left open.

The commit that introduces this formal-closure record must itself complete the same
four Backend Continuous Integration jobs before the external
`FORMAL_MODULE_CLOSURE=PASS` verdict is declared. That post-commit verification does
not require another source or documentation change when all four jobs succeed.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
PROJECTION_ROUTE_FREQUENCY_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_ROUTE_FREQUENCY_ENGINEERING_DEBT=CLOSED
PROJECTION_ROUTE_FREQUENCY_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_ROUTE_FREQUENCY_REVIEW_STATUS=CLOSED
```
