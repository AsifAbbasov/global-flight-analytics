# Projection Read Review Hardening

Status: engineering complete; formal closure pending permanent-audit exact Continuous Integration
Module: `apps/api/internal/projectionintelligence/projectionread`
Review baseline: `87b853e5a74bc5b8e0cd9bcb3f1e8e13eec8df0e`
Permanent-audit baseline: `9dda4b102497028b59280143b86bf84564afb136`

---

## 1. Review result

The original external review contained both stale findings and confirmed production defects.
The stale findings described contracts that had already been replaced before this closure cycle:
`DataSource.LoadSnapshot` already owned one atomic snapshot boundary, the production PostgreSQL
source already used a read-only `REPEATABLE READ` transaction, required telemetry was already
filtered without NULL-to-zero coercion, and current trajectory and current flight evidence were
already excluded from route history.

The confirmed findings were corrected in two engineering increments. The read service now validates
snapshot and Composer postconditions, binds the returned projection to the requested trajectory,
loads and validates route JSON against persisted row metadata, rejects nil contexts, assigns
`GeneratedAt` only after snapshot acquisition, repairs stale `UpdatedAt`, and canonicalizes the
zero-duration sentinel before Production composition. The second increment backfills rejected
historical candidate identifiers and binds route-history fingerprints to the exact contributing
route records and their input fingerprints.

---

## 2. Contract-hardening evidence

```text
CONTRACT_HARDENING_COMMIT=4eeff2b9f5b5c17dd6b7ebe5d0be4a7bd836fb37
CONTRACT_HARDENING_GITHUB_ACTIONS_RUN=30638188394
CONTRACT_HARDENING_POSTGRESQL_16_INTEGRATION_JOB=91181076159
CONTRACT_HARDENING_BACKEND_RACE_SAFETY_JOB=91181076172
CONTRACT_HARDENING_BACKEND_QUALITY_JOB=91181076240
CONTRACT_HARDENING_BACKEND_CONTAINER_JOB=91181409362
ATOMIC_REPEATABLE_READ_SNAPSHOT=CI_CONFIRMED
NIL_CONTEXT_REJECTION=CI_CONFIRMED
SNAPSHOT_IDENTITY_AND_AS_OF_POSTCONDITIONS=CI_CONFIRMED
COMPOSER_OUTPUT_POSTCONDITIONS=CI_CONFIRMED
COMPOSER_OUTPUT_IDENTITY_BINDING=CI_CONFIRMED
ROUTE_ROW_PAYLOAD_BINDING=CI_CONFIRMED
ROUTE_CONTRACT_EARLY_VALIDATION=CI_CONFIRMED
GENERATED_AT_AFTER_SNAPSHOT_ACQUISITION=CI_CONFIRMED
UPDATED_AT_AS_OF_REPAIR=CI_CONFIRMED
DEFAULT_DURATION_BOUNDARY=CI_CONFIRMED
SOURCE_NAME_CANONICALIZATION=CI_CONFIRMED
```

---

## 3. Evidence-hardening evidence

```text
EVIDENCE_HARDENING_COMMIT=9dda4b102497028b59280143b86bf84564afb136
EVIDENCE_HARDENING_GITHUB_ACTIONS_RUN=30648605652
EVIDENCE_HARDENING_BACKEND_QUALITY_JOB=91216081636
EVIDENCE_HARDENING_POSTGRESQL_16_INTEGRATION_JOB=91216081729
EVIDENCE_HARDENING_BACKEND_RACE_SAFETY_JOB=91216081733
EVIDENCE_HARDENING_BACKEND_CONTAINER_JOB=91216395238
HISTORICAL_CANDIDATE_BACKFILL=CI_CONFIRMED
ACCEPTED_CANDIDATE_LIMIT_AFTER_HYDRATION=CI_CONFIRMED
ROUTE_HISTORY_RECORD_LINEAGE=CI_CONFIRMED
ROUTE_HISTORY_INPUT_FINGERPRINT_LINEAGE=CI_CONFIRMED
ROUTE_HISTORY_AGGREGATE_MIRROR_VALIDATION=CI_CONFIRMED
DETERMINISTIC_ROUTE_HISTORY_EVIDENCE_FINGERPRINT=CI_CONFIRMED
```

---

## 4. Review classification

The following review statements were rejected as stale or mechanical rather than production defects:

- four independent public reads without a snapshot;
- NULL telemetry silently converted to zero;
- current trajectory included in route history;
- `AllowOnGround bool` treated as a procedural flag argument;
- optional pointers treated as an automatic architecture violation;
- file length treated as proof of incorrect semantics.

The PostgreSQL source and composition root remain large. This is maintainability debt only; no
unresolved result-integrity defect was established, so decomposition does not block formal closure.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
ENGINEERING_IMPLEMENTATION=COMPLETE
ENGINEERING_DEBT=CLOSED
ADDITIONAL_PRODUCTION_CODE_FIXES_REQUIRED=NO
POSTGRES_SOURCE_DECOMPOSITION=NON_BLOCKING_MAINTAINABILITY_IMPROVEMENT
```

---

## 5. Permanent audit status

Permanent review enforcement is implemented in:

```text
apps/api/tools/projectionreadreviewaudit
```

Backend Quality registration is implemented in:

```text
.github/workflows/backend-ci.yml
```

The permanent-audit commit and its exact push-triggered Backend Continuous Integration run are not
known until this enforcement increment is committed and executed. They must be recorded in a final
formal-closure increment.

```text
PERMANENT_AUDIT_COMMIT=PENDING
PERMANENT_AUDIT_GITHUB_ACTIONS_RUN=PENDING
PERMANENT_AUDIT_BACKEND_QUALITY_JOB=PENDING
PERMANENT_AUDIT_POSTGRESQL_16_INTEGRATION_JOB=PENDING
PERMANENT_AUDIT_BACKEND_RACE_SAFETY_JOB=PENDING
PERMANENT_AUDIT_BACKEND_CONTAINER_JOB=PENDING
PERMANENT_REVIEW_AUDIT=IMPLEMENTED_PENDING_EXACT_CI
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=YES
PROJECTION_READ_FORMAL_CLOSURE=OPEN_PENDING_EXACT_CI
PROJECTION_READ_REVIEW_STATUS=OPEN_PENDING_EXACT_CI_AND_FORMAL_CLOSURE
```
