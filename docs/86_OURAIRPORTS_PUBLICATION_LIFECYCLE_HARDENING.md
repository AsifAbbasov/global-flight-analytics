# OurAirports Publication Lifecycle Hardening

Status: Implemented Engineering Contract v1.0
Project: Global Flight Analytics

## Purpose

This document records the durable publication lifecycle used by the production
OurAirports import command.

The lifecycle prevents a publication from being marked as processed before its
database import succeeds. It also prevents two import processes from reconciling
the same publication concurrently.

## Publication identity

Every downloaded OurAirports CSV response receives a deterministic identifier:

```text
sha256:<hexadecimal content digest>
```

The identifier is based on the complete bounded response body. HTTP validators
remain responsible for conditional retrieval, while the content digest is the
canonical import publication identity.

## Durable lifecycle

The PostgreSQL table `provider_publications` records one row for each provider
and publication identifier.

The supported lifecycle is:

```text
ReservePublication
→ execute idempotent airport reconciliation
→ CommitPublication
```

When reconciliation fails or is cancelled before commit:

```text
ReleasePublication
→ the same publication may be retried
```

A reservation contains a unique ownership token. Only the owner may commit or
release it.

## Lease and crash recovery

Reservations use a thirty-minute lease in the production command.

A second process receives an in-progress decision while the lease is active.
After the lease expires, another process may atomically reclaim the publication.
This allows recovery after process termination without permitting concurrent
active owners.

## Validator ordering

The source HTTP validator is persisted only after one of these outcomes:

1. the publication import completed and the publication was committed;
2. PostgreSQL already contained a committed record for the same publication.

The validator is not persisted after an import failure. Therefore a failed
publication cannot be hidden behind a later HTTP `304 Not Modified` response.

If validator persistence fails after publication commit, the next command run
may download the content again, observe that the publication is already
committed, skip database reconciliation, and repair the validator record.

## Concurrency and idempotency

The database reservation uses a transaction and row-level locking.

The contract guarantees:

- one active owner per provider and publication identifier;
- retry after explicit release;
- retry after lease expiry;
- idempotent commit by the same reservation token;
- rejection of stale or foreign reservation tokens;
- no duplicate airport reconciliation after a committed publication;
- safe retry when reconciliation completed but publication commit was not
  recorded, because airport reconciliation is implemented as an idempotent
  upsert.

## Verification

The implementation includes:

- process-local publication lifecycle tests;
- orchestration release, commit, duplicate, and coalescing tests;
- deterministic OurAirports publication identity tests;
- production import workflow tests;
- optional PostgreSQL integration tests for reserve, release, commit, lease
  reclaim, and concurrent reservation;
- race detector coverage;
- full backend tests;
- `go vet`;
- migration and existing policy gates.

## Scope boundary

The lifecycle is publication-driven and currently used by OurAirports.
It does not convert live traffic providers into publication sources and does
not replace their fixed-window or provider-reported budget policies.

---

## Canonical remediation history

### GFA-DB-070 — OurAirports publication processing lacked durable reservation/commit ownership

1. **Finding / symptom.** A publication could be treated as processed before database reconciliation succeeded, and multiple importer processes could work on the same publication concurrently.
2. **Root cause.** HTTP retrieval/validator state acted as de facto publication progress without a durable database-owned reservation and commit lifecycle.
3. **Failure scenario.** Process A downloads a publication and marks progress before reconciliation finishes; it crashes, or Process B sees the same source state and begins the same import concurrently.
4. **Impact.** Publications can be skipped after failed work, duplicate reconciliation can race, and import completion is no longer a trustworthy durable fact.
5. **Severity rationale.** **P1 retrospective.** This directly affects whether authoritative airport source data is applied exactly once/retry-safely.
6. **Existing guarantees violated.** Publication completion must mean successful reconciliation; only one active owner may process a publication; crashes must remain recoverable.
7. **Considered solutions.** Process-local mutex; rely on HTTP validators; insert a simple processed flag before work; database reservation with ownership token, lease, commit and release.
8. **Chosen remediation.** Persist deterministic content identity in `provider_publications`; use `ReservePublication → reconcile → CommitPublication`, with owner token, explicit release and lease reclaim.
9. **Why this solution was selected.** PostgreSQL is already the durable shared authority and can coordinate restarts and multiple processes without adding Redis or another service.
10. **Rejected alternatives.** Process-local locks do not coordinate instances or survive restart; validators identify HTTP freshness, not successful import; pre-work processed flags preserve the original loss mode.
11. **Trade-offs.** The workflow adds a durable lifecycle table, token/lease semantics and a lease duration that must exceed normal import time.
12. **Regression tests / protection.** Reserve/release/commit, duplicate/coalescing, lease reclaim, concurrent reservation, production command and race tests.
13. **Adversarial review findings.** Foreign/stale tokens must not commit or release; a live lease must prevent a second owner; lease expiry must allow crash recovery; already committed publications must not reconcile again.
14. **Remediation iterations.** Deterministic content digest became the canonical publication identity while HTTP validators were retained only for retrieval optimization.
15. **Residual risks and limitations.** A too-short lease can allow reclaim while a slow valid import is still running; a too-long lease delays crash recovery.
16. **Operational or deployment consequences.** The production command owns a thirty-minute reservation lease and PostgreSQL availability is required for publication admission.
17. **Exact evidence.** Historical implementation commit `db73719ec134da627128038f9be413f38cf4e0e6` (`fix: harden OurAirports publication lifecycle`). Historical pull-request/reviewer evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-DB-070=CLOSED`.
19. **Prevention / future guard.** Publication-style sources must use a durable reserve/commit/release lifecycle with explicit ownership and crash-recovery tests; retrieval validators alone cannot represent processing completion.

### GFA-DATA-071 — HTTP validator ordering could hide a failed publication behind `304 Not Modified`

1. **Finding / symptom.** Persisting source HTTP validators before successful import could cause a later `304 Not Modified` to suppress retry of a publication whose reconciliation had failed.
2. **Root cause.** Retrieval freshness evidence and processing-completion evidence were not ordered as separate state transitions.
3. **Failure scenario.** CSV download succeeds, validator is stored, import fails, and the next request returns 304 because the source did not change; the failed publication is never reconciled.
4. **Impact.** Current airport source data can be silently missing while the fetch layer reports no update required.
5. **Severity rationale.** **P1 retrospective.** The defect can cause durable source-data loss/omission without an obvious provider error.
6. **Existing guarantees violated.** A validator may optimize retrieval only after the publication is durably committed or already known committed.
7. **Considered solutions.** Save validator immediately after download; never persist validators; persist only after committed publication; persist before import but force redownload on importer failure.
8. **Chosen remediation.** Persist the validator only after publication commit or when PostgreSQL already confirms the same publication committed.
9. **Why this solution was selected.** It preserves conditional HTTP efficiency without allowing transport freshness to outrank database processing truth.
10. **Rejected alternatives.** Early persistence creates the loss condition; removing validators wastes bandwidth; custom failure invalidation is more fragile than commit-ordered persistence.
11. **Trade-offs.** If validator persistence fails after a successful commit, the next run may redownload the same content once, then skip reconciliation by committed content identity and repair the validator.
12. **Regression tests / protection.** Production workflow tests cover validator ordering, duplicate committed publication behavior and retry after failed import.
13. **Adversarial review findings.** Validator failure after commit must not roll back a completed import, and import failure must never advance validator state.
14. **Remediation iterations.** Content-digest publication identity decoupled idempotent processing from ETag/Last-Modified retrieval state.
15. **Residual risks and limitations.** Conditional retrieval still depends on provider validator correctness, but incorrect validators can no longer certify successful local processing.
16. **Operational or deployment consequences.** Occasional safe redownload is accepted after validator-write failure; database commit remains the canonical processing state.
17. **Exact evidence.** Historical implementation commit `db73719ec134da627128038f9be413f38cf4e0e6`. Historical PR/reviewer evidence unavailable.
18. **Final canonical status.** `GFA-DATA-071=CLOSED`.
19. **Prevention / future guard.** Any future conditional source importer must prove `download → successful durable processing → validator advance` ordering and test the failed-processing/next-304 scenario.
