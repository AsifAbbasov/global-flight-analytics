# Projection Freshness Review Hardening

Status: closed

```text
LINEAGE_AGE_INTEGRITY_COMMIT=0b47aa3231c93d573a6026651a4085d376a40583
LINEAGE_AGE_INTEGRITY_GITHUB_ACTIONS_RUN=30502639621
POLICY_DECISION_INTEGRITY_COMMIT=072d0eb349fcd0e42c1d3c0bcf54c51cefb08a19
POLICY_DECISION_INTEGRITY_GITHUB_ACTIONS_RUN=30503845277
PERMANENT_AUDIT_COMMIT=619e24878a5025decf6fe21abddba537ce195560
PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30523502590
PERMANENT_AUDIT_BACKEND_QUALITY_JOB=90809046060
PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=90809046046
PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=90809046013
PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=90809225151
WEIGHT_POLICY_CORRECTION_COMMIT=e3e99758d6f654db12ccce32ec55ad1339fb518f
WEIGHT_POLICY_CORRECTION_GITHUB_ACTIONS_RUN=30527541240
WEIGHT_POLICY_CORRECTION_BACKEND_QUALITY_JOB=90821894564
WEIGHT_POLICY_CORRECTION_BACKEND_RACE_SAFETY_JOB=90821894465
WEIGHT_POLICY_CORRECTION_POSTGRESQL_16_INTEGRATION_JOB=90821894536
WEIGHT_POLICY_CORRECTION_BACKEND_CONTAINER_JOB=90822090361
WEIGHT_POLICY_CONSISTENCY=CLOSED
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
PROJECTION_FRESHNESS_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_FRESHNESS_ENGINEERING_DEBT=CLOSED
PROJECTION_FRESHNESS_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_FRESHNESS_REVIEW_STATUS=CLOSED
```

## 1. Scope

This record covers the dedicated review and hardening of:

```text
apps/api/internal/projectionintelligence/projectionfreshness
```

It also covers the lineage and production-fixture boundaries through:

```text
projectionneighbors
projectionpatternconfidence
projectionproduction
projectioncontinuation
projectionread
```

The review is limited to deterministic selected-neighbor age evidence, policy,
lineage, scoring, decision semantics, validation, fingerprinting, and production
contract integrity. It does not claim empirical forecast accuracy, operational
aviation suitability, calibrated probability semantics, or that historical
freshness alone proves future trajectory quality.

## 2. Findings closed

### 2.1 Pattern usability and exact lineage

Freshness now blocks historical continuation when Pattern Confidence is not usable.
The evaluator requires all of the following to agree before freshness is evaluated:

```text
Pattern Confidence source-selection fingerprint
Neighbor Selection input fingerprint
Pattern Confidence selection status
Neighbor Selection status
sorted selected trajectory identifiers
```

A Pattern Confidence result from another selection, another selection state, or
another selected-neighbor catalog cannot be reused silently.

### 2.2 Timestamp-derived age evidence

Freshness derives every candidate age from:

```text
Neighbor Selection AsOfTime - CandidateEndTime
```

The upstream `CandidateAge` field is validated by Neighbor Selection, but Freshness
does not treat the duplicated duration as its source of truth. This preserves one
canonical time calculation at the decision boundary.

The arithmetic mean uses quotient-and-remainder accumulation rather than summing all
nanoseconds first, preventing signed 64-bit overflow for valid large durations.

### 2.3 Configuration contract hardening

Configuration validation now requires:

```text
zero < maximum newest age <= maximum mean age <= maximum oldest age
zero < recent-neighbor age limit <= maximum oldest age
zero < minimum recent-neighbor count <= target recent-neighbor count
zero < minimum usable score <= complete score minimum <= one
finite positive component weights
component-weight total equal to one
```

Zero score thresholds and incoherent age ordering are rejected before an evaluator
is created.

### 2.3.1 Strictly positive component-weight correction

The original closure document stated that every component weight is finite and
strictly positive, while the implementation rejected only negative weights. That
allowed a zero weight to disable one of the four canonical Freshness components while
keeping the remaining weights normalized to one.

The correction changes both configuration and result validation to reject every
non-finite or non-positive component weight. Regression tests cover zero values for
newest age, mean age, oldest age, and recent support, including a coordinated Result
mutation whose policy, component catalog, weighted score, decision, usability, and
limitations otherwise remain internally consistent.

The permanent audit requires the `<= 0` guards and forbids the former `< 0`
contracts. Corrective commit `e3e99758d6f654db12ccce32ec55ad1339fb518f` completed exact Continuous
Integration run `30527541240`, so the code, tests, permanent audit, and formal
weight-policy contract are now aligned.

### 2.4 Semantic freshness fingerprint

The input fingerprint binds:

```text
Freshness fingerprint version
Neighbor Selection input fingerprint and status
Pattern Confidence input fingerprint
Pattern Confidence source-selection fingerprint
Pattern Confidence selection status
Pattern Confidence status and usability
normalized AsOfTime
all normalized policy thresholds and weights
sorted selected trajectory identifiers
candidate end timestamps
timestamp-derived candidate ages
```

Upstream status or usability changes therefore produce a different Freshness input
identity even when selected trajectory identifiers remain unchanged.

### 2.5 Complete hard-violation reporting

The policy evaluator accumulates every simultaneous blocking reason rather than
returning only the first matching switch branch. The result can report all applicable
violations, including:

```text
historical neighbors unavailable
Pattern Confidence unusable
newest selected neighbor too old
mean selected-neighbor age too old
oldest selected neighbor too old
insufficient recent-neighbor support
Freshness score below the usable minimum
```

Limitations are normalized, sorted, and unique.

### 2.6 Policy and upstream-state snapshots

Every Result publishes the normalized policy that created it and the upstream state
that authorized its calculation:

```text
age thresholds
recent-support thresholds
score thresholds
component weights
Neighbor Selection status
Pattern Confidence status
Pattern Confidence usability
source Neighbor Selection fingerprint
source Pattern Confidence fingerprint
```

The snapshot makes the result independently inspectable without relying on mutable
runtime evaluator configuration.

### 2.7 Component and decision reconstruction

`Result.Validate()` reconstructs and verifies:

```text
policy validity
upstream status and usability consistency
neighbor counts and ordered age aggregates
canonical four-component catalog
component scores from aggregate evidence
component weights from the policy snapshot
weighted total score
Decision
Usable
exact required limitations
sorted unique trajectory identifiers
sorted unique limitations
source lineage fingerprints
```

A caller cannot make a coordinated mutation to component scores, total score,
policy, decision, usability, or limitations and still pass validation.

### 2.8 Production fixture contract integrity

Production tests no longer hand-assemble nominal Freshness results. Allowed, limited,
and blocked fixtures are generated through the real evaluator with valid Neighbor
Selection and Pattern Confidence inputs.

A dedicated fixture-contract test validates policy snapshots, source fingerprints,
and the final Freshness result so future contract changes cannot leave stale test
objects that bypass production semantics.

## 3. Deliberately retained and rejected recommendations

The following contracts were reviewed and deliberately retained:

```text
four fixed components as a versioned domain schema
float64 arithmetic with explicit finite checks and comparison tolerance
idiomatic New(Config) (*Evaluator, error) constructor
Usable compatibility fields with cross-field validation
blocked domain result for unusable upstream pattern evidence
small local numeric and normalization helpers
```

The following recommendations were rejected as unsupported or outdated mechanical
findings:

```text
mandatory integer basis points
mandatory dynamic component registry
constructor criticism based only on nil plus error
claim that candidate age is not validated upstream
claim that Freshness remains duplicated inside Pattern Confidence
blanket removal of every compatibility usability field
blanket rejection of small local helpers
```

The fixed component catalog is intentional because `Result.Validate()` reconstructs
a stable, versioned Freshness schema. Floating-point values remain appropriate for
normalized ratios when all inputs are finite and equality-sensitive checks use an
explicit comparison tolerance.

## 4. Permanent regression coverage

Permanent tests cover:

```text
unusable Pattern Confidence blocking
source-selection fingerprint mismatch
selection-status lineage mismatch
upstream status fingerprint mutation
timestamp-derived age consistency
signed 64-bit mean-duration overflow avoidance
ordered age-threshold validation
positive score-threshold validation
all simultaneous hard violations
weighted-score mutation rejection
policy snapshot mutation rejection
coordinated component-and-score mutation rejection
decision-and-limitation mutation rejection
upstream status and usability mutation rejection
source Pattern Confidence fingerprint mutation
valid limited-decision reconstruction
production fixture evaluator generation
production fixture contract drift
```

## 5. Permanent audit gate

The source audit is:

```text
apps/api/tools/projectionfreshnessreviewaudit
```

Backend Continuous Integration executes:

```text
go run ./tools/projectionfreshnessreviewaudit -strict
```

The gate protects the hardened implementation, regression tests, production fixture
contracts, workflow wiring, Stage 9 closure markers, this authoritative record, and
the Documentation Index entry.

## 6. Closure evidence

The permanent audit commit completed every Backend Continuous Integration job:

```text
PERMANENT_AUDIT_COMMIT=619e24878a5025decf6fe21abddba537ce195560
PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=30523502590
PERMANENT_AUDIT_BACKEND_QUALITY_JOB=90809046060
PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=90809046046
PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=90809046013
PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=90809225151
```

The `Backend Quality` job executed and passed the dedicated step:

```text
Run projection freshness review audit
```

The strictly positive component-weight correction completed every required Backend
Continuous Integration job:

```text
WEIGHT_POLICY_CORRECTION_COMMIT=e3e99758d6f654db12ccce32ec55ad1339fb518f
WEIGHT_POLICY_CORRECTION_GITHUB_ACTIONS_RUN=30527541240
WEIGHT_POLICY_CORRECTION_BACKEND_QUALITY_JOB=90821894564
WEIGHT_POLICY_CORRECTION_BACKEND_RACE_SAFETY_JOB=90821894465
WEIGHT_POLICY_CORRECTION_POSTGRESQL_16_INTEGRATION_JOB=90821894536
WEIGHT_POLICY_CORRECTION_BACKEND_CONTAINER_JOB=90822090361
```

The corrective `Backend Quality` job executed and passed the dedicated step:

```text
Run projection freshness review audit
```

Engineering implementation and formal closure documentation are complete. No
confirmed finding remains open, unclassified, or deferred. The permanent audit gate
remains mandatory in Backend Continuous Integration.
