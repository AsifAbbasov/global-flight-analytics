# Document 72 — Stage 14.30 PostgreSQL Correctness Hardening

Status: Implemented current-scope baseline
Project: Global Flight Analytics
Scope: close Ingestion Run evidence invariants, Route and Historical timestamp mirror drift, and cancelled-context rollback risk

## 1. Correctness scope

This increment closes three independently confirmed PostgreSQL correctness gaps:

```text
Ingestion Run processed counts and error evidence were not tied to status
Route Results did not compare timestamptz mirrors with exact Unix nanoseconds
Historical Aggregate Results did not compare four timestamptz mirrors with exact Unix nanoseconds
Airport Import, Flight State and Trajectory writes rolled back with the caller context
```

It does not claim that the remaining Stage 14 maintainability and Clean Code register is complete.

## 2. Ingestion Run evidence contract

Repository finalization now validates before issuing SQL:

```text
all counters are non-negative
records_inserted + records_updated <= records_received
success has no error message
failed and partial require a non-empty normalized error message
only terminal statuses can enter the completion validator
```

Migration `020_stage14_correctness_hardening.sql` applies the same rules as PostgreSQL check constraints. Direct SQL and future repository implementations therefore cannot bypass the evidence contract.

## 3. Exact timestamp ownership

Unix nanoseconds remain the canonical exact representation. PostgreSQL `timestamptz` columns remain operator-readable mirrors with microsecond precision.

Every Route Result read now selects and validates:

```text
as_of_time ↔ as_of_time_unix_nano
stored_at ↔ stored_at_unix_nano
```

Every Historical Aggregate read now selects and validates:

```text
window_start ↔ window_start_unix_nano
window_end ↔ window_end_unix_nano
as_of_time ↔ as_of_time_unix_nano
stored_at ↔ stored_at_unix_nano
```

A difference below one microsecond is accepted as PostgreSQL precision loss. A difference of one microsecond or more fails closed as corrupt persisted evidence.

Migration 020 also installs database constraints for all six mirror pairs.

## 4. Independent rollback context

Repository transaction rollback now uses one shared helper with a fresh bounded context derived from `context.Background()`.

The helper owns rollback for:

```text
AirportRepository.UpsertImported
FlightStateRepository.SaveFlightStates
TrajectoryRepository.saveTrajectory
```

Caller cancellation continues to stop normal work and commit. It no longer prevents the deferred rollback attempt from reaching PostgreSQL.

## 5. Production catalog and integration evidence

The permanent Stage 14 PostgreSQL gate:

```text
applies the complete production migration catalog 001–020
runs the migrator a second time to prove no pending migrations remain
runs repository, Feature Store, Route Store and Historical Aggregate package tests
runs an isolated-schema integration test through the real migrator
proves PostgreSQL rejects invalid counters, invalid error semantics and timestamp drift
```

GitHub Actions now inspects the catalog before application and verifies the absence of pending migrations only after application. A fresh continuous-integration database is no longer incorrectly rejected merely because its migrations are initially pending.

## 6. Regression protection

Permanent tests protect:

```text
repository fail-fast Ingestion Run validation
PostgreSQL Ingestion Run check violations
Route mirror selection and scan validation
Historical mirror selection and scan validation
sub-microsecond precision tolerance
one-microsecond corruption rejection
independent bounded rollback context
production migration 020 ownership
continuous-integration execution of all affected packages
```

## 7. Completion boundary

This increment closes the three correctness groups named in this document.

Stage 14 remains reopened for the separate confirmed backlog, including large PostgreSQL method decomposition, query-contract cleanup, pagination, nil-context policy, nullable helper semantics, migration-repair generalization, repeated SQL and Scan contours, and evidence-backed trajectory index profiling.

## 8. Canonical finding decomposition

This historical increment contains four independently tracked findings. The original document described them together because they were delivered in one PostgreSQL hardening increment; the Finding Register tracks them separately so their severity, residual risk, and future regression boundaries remain explicit.

```text
GFA-DB-014  Ingestion Run terminal evidence invariants
GFA-DB-015  Route Result timestamp mirror consistency
GFA-DB-016  Historical Aggregate timestamp mirror consistency
GFA-DB-017  Cancelled-context rollback independence
```

## 9. GFA-DB-014 — Ingestion Run terminal evidence invariants

### Finding / symptom

A terminal ingestion run could satisfy the existing lifecycle status transition while still carrying internally impossible evidence. Processed counters were not constrained against records received, and terminal status was not consistently coupled to error-message semantics.

### Root cause

The earlier terminal-integrity work correctly made terminal rows immutable, but lifecycle immutability and evidence validity are different guarantees. The repository and schema protected *when* a run could transition without fully protecting *what evidence* a terminal row was allowed to claim.

### Failure scenario

A writer could persist a successful terminal run with an error message, or persist inserted and updated counts whose sum exceeded the received count. Such a row would be immutable and therefore durable, but its immutable evidence would be contradictory.

### Impact

Ingestion reliability metrics, operator diagnostics, reconciliation, and later analytics could treat impossible counters or contradictory status/error evidence as factual historical input.

### Severity rationale

**P1 retrospective.** This is persisted operational-evidence integrity. Once admitted as a terminal immutable row, contradictory evidence can influence downstream reasoning and cannot be corrected through the normal lifecycle path. The original remediation did not record a historical severity label; this classification is retrospective.

### Existing guarantees violated

- terminal ingestion evidence must be internally coherent;
- successful ingestion must not claim failure evidence;
- failed/partial ingestion must retain an explanation;
- processed counts must not exceed observed input.

### Considered solutions

1. validate only in application code;
2. add only PostgreSQL constraints;
3. validate at both repository and PostgreSQL boundaries.

### Chosen remediation and why

The repository rejects impossible completion requests before SQL, while migration 020 installs the same invariants as database constraints. The dual boundary gives callers useful typed failures while preserving PostgreSQL as the final authority for direct SQL and future writers.

### Rejected alternatives

Application-only validation was rejected because direct SQL and future repository implementations could bypass it. Database-only validation was rejected because it would defer predictable domain failures to opaque SQL errors and lose fail-fast diagnostics.

### Trade-offs

The contract is stricter and requires tests/fixtures to construct realistic terminal evidence. That additional fixture discipline is intentional because permissive fixtures would conceal the same class of production inconsistency.

### Regression tests

Repository validation, PostgreSQL integration constraints, production migration ownership, and Stage 14 integration gates protect the invariant.

### Adversarial review and remediation iterations

Document 62 first made terminal completion one-way and immutable. A later Stage 14 review identified that immutable evidence could still be impossible. Stage 14.30 therefore extended the lifecycle contract rather than replacing Document 62. This is a genuine second remediation iteration, not a retroactively invented reviewer event.

### Residual risk / limitations

The database can enforce arithmetic and status/error invariants, but it cannot independently prove that upstream providers reported truthful record counts. Semantic truth of source observations remains an ingestion/provenance responsibility.

### Operational / deployment consequences

Migration 020 must be applied before relying on direct-database enforcement. Legacy rows violating the new constraints require explicit inspection rather than silent normalization.

### Exact evidence and final status

Implementation commit: `04137ad17690ca197f1aea74b434057f2157dd7d` (`fix: harden postgres correctness invariants`). Historical PR/reviewer metadata is not reconstructed where repository evidence does not establish it. **Canonical finding status: CLOSED.**

## 10. GFA-DB-015 — Route Result timestamp mirror consistency

### Finding / symptom

Route Result reads trusted exact Unix-nanosecond columns without verifying that the operator-readable PostgreSQL `timestamptz` mirrors still represented the same instant.

### Root cause

The storage model intentionally kept two representations with different precision, but the read path treated the exact integer as sufficient and did not fail when the mirror drifted beyond PostgreSQL's expected microsecond rounding envelope.

### Failure scenario

A direct repair, future writer, or partial historical migration could change `as_of_time` or `stored_at` while leaving its Unix-nanosecond counterpart unchanged. Reads would return the exact integer silently even though the persisted row contained contradictory time evidence.

### Impact

Operators, SQL diagnostics, and application consumers could reason from different timestamps for the same Route Result. Corrupt temporal evidence could survive unnoticed and contaminate ordering, audit interpretation, and historical debugging.

### Severity rationale

**P1 retrospective.** Route-result time is persisted analytical evidence, and silent disagreement between two representations weakens deterministic temporal identity. The original remediation did not record a severity label.

### Existing guarantees violated

- one logical instant must not have contradictory persisted representations;
- exact Unix nanoseconds remain canonical;
- PostgreSQL mirrors are compatibility mirrors, not independent facts.

### Considered solutions

1. remove `timestamptz` mirrors;
2. trust Unix nanoseconds and ignore mirror drift;
3. compare both representations on every read and enforce compatible drift in PostgreSQL.

### Chosen remediation and why

Both representations are selected and validated. Sub-microsecond differences caused by PostgreSQL precision are accepted; a difference of one microsecond or more fails closed. Migration 020 adds database checks for the same relationship.

### Rejected alternatives

Removing mirrors was rejected as an unnecessary destructive schema change that would reduce operator readability. Ignoring mirrors was rejected because contradictory persisted evidence would remain invisible.

### Trade-offs

Reads scan additional columns and perform a small consistency check. The cost is deliberately accepted in exchange for fail-closed evidence integrity.

### Regression tests

Route Store tests protect mirror selection, sub-microsecond tolerance, one-microsecond rejection, and production migration enforcement.

### Adversarial review and remediation iterations

Document 68 established the same policy first for Flight Feature snapshots. Stage 14.30 identified Route Results as an uncovered parallel persistence surface and extended the proven invariant instead of creating a competing timestamp policy.

### Residual risk / limitations

The check proves internal representation consistency, not correctness of the upstream event time itself.

### Operational / deployment consequences

Existing inconsistent rows would fail reads after deployment and must be repaired from authoritative evidence; they are not silently normalized.

### Exact evidence and final status

Implementation commit: `04137ad17690ca197f1aea74b434057f2157dd7d`. Historical PR/reviewer metadata unavailable from confirmed repository evidence. **Canonical finding status: CLOSED.**

## 11. GFA-DB-016 — Historical Aggregate timestamp mirror consistency

### Finding / symptom

Historical Aggregate Results persisted four logical instants in both `timestamptz` and exact Unix-nanosecond form without verifying their agreement on reads.

### Root cause

The dual-representation design had been implemented consistently on writes but not defended across every read path. Historical Aggregate storage therefore had more mirror pairs than the Feature Store while lacking the same fail-closed scanner contract.

### Failure scenario

One mirror column could be edited, repaired incorrectly, or written by a future path without updating the exact integer. The aggregate could then expose contradictory window start/end, as-of, or stored-at evidence.

### Impact

Window boundaries and materialization identity are central to historical analytics. Drift could cause operators and application logic to disagree about what interval an aggregate represents.

### Severity rationale

**P1 retrospective.** This is analytical evidence integrity across window identity and storage time. A silent mismatch can undermine historical comparison and reproducibility.

### Existing guarantees violated

- historical window identity must be deterministic;
- all persisted representations of each logical instant must agree within known storage precision;
- corrupt evidence must fail closed.

### Considered solutions, chosen remediation, rejected alternatives, and trade-offs

The same alternatives and precision rationale as GFA-DB-015 apply, but across four mirror pairs. The chosen solution validates all pairs on reads and enforces them through migration 020. Removing mirrors or tolerating arbitrary divergence were rejected for the same operator-readability and integrity reasons. The trade-off is additional scan/check work proportional only to four timestamp pairs per aggregate row.

### Regression tests

Historical Aggregate integration tests exercise all mirror pairs, expected PostgreSQL precision loss, corruption rejection, and production migration constraints.

### Adversarial review and remediation iterations

This finding is the Historical counterpart discovered while generalizing the Flight Feature timestamp invariant. The remediation deliberately reuses one temporal-integrity policy across persistence subsystems.

### Residual risk / limitations

Internal consistency does not prove that the chosen historical window was analytically appropriate; window-definition correctness belongs to the historical analytics contract.

### Operational / deployment consequences

Corrupt legacy rows fail closed and require evidence-backed repair.

### Exact evidence and final status

Implementation commit: `04137ad17690ca197f1aea74b434057f2157dd7d`. Historical PR/reviewer metadata not asserted without recoverable evidence. **Canonical finding status: CLOSED.**

## 12. GFA-DB-017 — Cancelled-context rollback independence

### Finding / symptom

Several repository write transactions deferred rollback using the caller context. If the caller context was already cancelled, the cleanup attempt could itself be prevented from reaching PostgreSQL.

### Root cause

Normal database work and transaction cleanup shared one context-ownership policy even though their lifecycle requirements differ. Caller cancellation should stop forward progress, but cleanup must still get a bounded independent opportunity to release transaction state.

### Failure scenario

A request is cancelled after a transaction begins but before commit. Deferred rollback inherits the cancelled context and cannot execute normally. The connection may remain in transaction until driver/pool cleanup, increasing lock/resource retention and making failure behavior dependent on lower-level connection handling.

### Impact

The primary risk is resource and lock retention, connection-pool pressure, and less deterministic cleanup under failure. The remediation does not claim that a cancelled rollback would commit data.

### Severity rationale

**P2 retrospective.** The defect affects transactional cleanup reliability and production resilience but is not evidence of silent committed-data corruption. The original remediation did not record a historical severity.

### Existing guarantees violated

- caller cancellation stops normal work;
- started transactions receive a deterministic bounded cleanup attempt;
- cleanup ownership must be explicit and separate from caller intent.

### Considered solutions

1. reuse caller context;
2. use unbounded `context.Background()` for rollback;
3. use a shared bounded cleanup context derived independently from the caller.

### Chosen remediation and why

A shared rollback helper uses a fresh bounded context from `context.Background()`. This preserves cancellation semantics for normal work while giving rollback a finite independent cleanup window.

### Rejected alternatives

Reusing the caller context reproduced the defect. Unbounded background cleanup was rejected because cleanup must not hang indefinitely or create hidden lifecycle ownership.

### Trade-offs

Cleanup may continue briefly after the initiating request has ended. That is intentional and bounded. The helper introduces a narrowly sanctioned use of `context.Background()` that must not spread into normal database operations.

### Regression tests

Architecture/source tests protect the shared rollback helper and its bounded independent context. Later context-hardening stages explicitly distinguish this cleanup exception from forbidden caller-context substitution.

### Adversarial review and remediation iterations

Stage 14.33 and the later post-closure migrator review both revisited context ownership. They preserved this rollback exception while removing background-context substitution from normal work, demonstrating that the exception survived adversarial review for a specific lifecycle reason.

### Residual risk / limitations

If PostgreSQL or the network remains unavailable beyond the cleanup timeout, rollback still cannot be guaranteed by the application. Connection destruction/pool recovery remains the final lower-level safety path.

### Operational / deployment consequences

No schema migration is needed for the helper itself. Operators benefit from more deterministic lock/connection cleanup during cancellation storms or shutdown races.

### Exact evidence and final status

Implementation commit: `04137ad17690ca197f1aea74b434057f2157dd7d`. Later context-policy evidence is recorded in Documents 75 and 79. **Canonical finding status: CLOSED.**

## 13. Prevention / future guard

The common prevention rule for all four findings is that persisted evidence and cleanup semantics must be enforced at the narrowest authoritative boundary rather than inferred from happy-path callers.

Future PostgreSQL work must therefore:

- encode durable evidence invariants in database constraints when PostgreSQL can express them;
- provide fail-fast application validation when it improves diagnostics;
- verify every duplicate representation of a logical value rather than silently choosing one;
- keep normal caller-owned contexts distinct from explicitly bounded cleanup contexts;
- extend an existing invariant owner instead of creating subsystem-specific variants.

## 14. Evidence honesty note

The technical source, tests, migration ownership, and implementation commit are recoverable. Historical adversarial-review comments or a specific historical PR number are not asserted where current repository evidence does not prove them. The review chronology above is limited to relationships visible in the committed Stage 14 documents and commits.
