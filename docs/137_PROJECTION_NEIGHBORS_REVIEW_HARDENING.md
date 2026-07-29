# Projection Neighbors Review Hardening

Status: implemented pending permanent audit Continuous Integration

```text
AUTHORITATIVE_BASELINE_COMMIT=e13a117f969e2922d09a7804fe50005d01bc2ecf
CANDIDATE_INTEGRITY_COMMIT=e2a4a7dc76e43942ca9deb0d8d5f83a09a42deff
CONTINUATION_INTEGRITY_COMMIT=911a1b102c68af2746a13bfca48b008cf7225ff8
ROUTE_SCOPE_INTEGRITY_COMMIT=3eee05fb44484aa6e389af66520aba23d4ae277e
SELECTOR_PIPELINE_COMMIT=353d19bc97f561e1897ece1967e7304c0e10b5fb
PERMANENT_AUDIT_COMMIT=PENDING
PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=PENDING
PERMANENT_AUDIT_BACKEND_QUALITY_JOB=PENDING
PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=PENDING
PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=PENDING
PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=PENDING
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
PROJECTION_NEIGHBORS_ENGINEERING_IMPLEMENTATION=COMPLETE
PROJECTION_NEIGHBORS_ENGINEERING_DEBT=CLOSED
PROJECTION_NEIGHBORS_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=YES
PROJECTION_NEIGHBORS_REVIEW_STATUS=IMPLEMENTED_PENDING_PERMANENT_AUDIT_CI
```

## 1. Scope

This record covers the dedicated review and hardening of:

```text
apps/api/internal/projectionintelligence/projectionneighbors
```

It also covers the route-scope evidence path through:

```text
projectionread
projectionproduction
projectioncontinuation
```

The review is limited to the deterministic historical-neighbor selection contract.
It does not claim empirical prediction accuracy or operational aviation suitability.

## 2. Findings closed

### 2.1 Candidate integrity before expensive evaluation

The selector now performs source-independent eligibility and duplicate detection
before applying the expensive similarity budget. Duplicate identifiers are counted
across the whole input before truncation, so duplicate evidence cannot bypass the
guard by appearing outside the evaluation budget.

Eligible candidates are ordered by newest historical end time and then stable
trajectory identifier before `MaximumCandidateCount` is applied.

### 2.2 Canonical point and fingerprint ordering

Snapshot construction and fingerprint generation share canonical point ordering.
Equal timestamps are ordered by point identifier. Candidate fingerprints are
canonicalized independently of input order.

The selection fingerprint covers:

```text
selection contract version
similarity policy identity
route-scope fingerprint
as-of time
required continuation duration
maximum continuation gap
candidate and selection limits
similarity, distance and age policies
current trajectory snapshot
candidate trajectory snapshots
```

### 2.3 Similarity failure classification

Candidate-local non-comparability is represented as a deterministic rejection.
Systemic similarity-engine failure and malformed similarity evidence are returned
as typed errors and are not hidden as ordinary candidate rejection.

### 2.4 Continuous continuation evidence

The selector publishes structured `AnchorEvidence` and uses segmented linear
continuation search. A candidate continuation cannot cross an observation gap
larger than the configured maximum.

Unavailable duration and discontinuous continuation are distinct rejection cases.

### 2.5 Source-attested route scope

Historical candidates require explicit route-scope evidence. The PostgreSQL read
path constructs a uniform route attestation only after route-filtered candidate
loading. The read snapshot transports a defensive clone of that evidence.

Production composition validates the candidate route scope against the current
complete Route Intelligence result. Projection Continuation receives and forwards
the same route scope to its internal neighbor selector.

Cross-route candidates are rejected before anchor or similarity evaluation.

### 2.6 Selector pipeline decomposition

`Selector.Select` is a short coordinator for:

```text
prepareSelectionContext
evaluateCandidatePool
assembleSelectionResult
```

Request normalization, candidate preparation, expensive evaluation, deterministic
ranking, result assembly, limitation generation and result validation are separated
into focused files.

### 2.7 Explicit limiting semantics

Two independent conditions are now published:

```text
CandidateEvaluationTruncated
QualifiedSelectionLimited
```

`CandidateEvaluationTruncated` means the expensive evaluation budget prevented all
eligible candidates from being checked.

`QualifiedSelectionLimited` means more candidates qualified than could be returned
under `SelectionLimit`.

The deprecated `Truncated` field remains a compatibility alias for
`CandidateEvaluationTruncated` and is cross-field validated.

## 3. Deliberately retained contracts

The following were reviewed and deliberately retained:

```text
exact float64 comparison for deterministic ranking tie-breakers
idiomatic New(Config) (*Selector, error) constructor
public compatibility alias Result.Truncated
producer-owned similarity implementation behind the consumer-facing selector contract
```

No product requirement justified arbitrary coordinate or score quantization.

## 4. Permanent regression coverage

Permanent tests cover:

```text
eligibility before expensive budget
whole-input duplicate detection
canonical equal-timestamp ordering
systemic similarity failure propagation
similarity evidence validation
continuous continuation gaps
large linear anchor search
source-attested route scope
cross-route rejection before similarity
route-scope fingerprint identity
read, production and continuation propagation
candidate-evaluation truncation
qualified-selection limiting
cross-field result validation
```

## 5. Permanent audit gate

The source audit is:

```text
apps/api/tools/projectionneighborsreviewaudit
```

Backend Continuous Integration executes:

```text
go run ./tools/projectionneighborsreviewaudit -strict
```

The gate protects the hardened contracts, tests, documentation and workflow wiring.

## 6. Closure condition

Engineering implementation is complete and no confirmed finding remains open,
unclassified or deferred.

Formal closure is intentionally pending until the permanent audit commit completes
all Backend Continuous Integration jobs. The final evidence increment will replace
the pending commit, run and job identifiers and set:

```text
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
PROJECTION_NEIGHBORS_REVIEW_STATUS=CLOSED
```
