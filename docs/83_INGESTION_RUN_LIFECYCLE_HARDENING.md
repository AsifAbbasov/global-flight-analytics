# Document 83 — Ingestion Run Lifecycle Hardening

Status: Implemented Engineering Increment v1.0
Project: Global Flight Analytics

## Purpose

This increment prevents traffic ingestion history from remaining permanently in
`running` after cancellation, process termination, or a failed terminal status
write.

## Protected invariants

- Every created ingestion run must eventually become `success`, `failed`, or
  `partial`.
- Cancellation of the data-plane operation must not cancel the bounded terminal
  status update.
- Terminal status writes preserve caller context values while removing caller
  cancellation and deadlines.
- Startup recovery may finalize only runs older than the configured stale
  threshold.
- Concurrent recovery attempts may finalize a stale run only once.
- Existing terminal rows remain immutable.

## Runtime configuration

```text
TRAFFIC_INGESTION_TERMINAL_TIMEOUT=15s
TRAFFIC_INGESTION_STALE_RUN_AFTER=30m
```

`TRAFFIC_INGESTION_TERMINAL_TIMEOUT` bounds success, failure, and recovery writes.

`TRAFFIC_INGESTION_STALE_RUN_AFTER` must remain longer than the maximum expected
normal ingestion cycle duration. Runs older than this threshold are finalized as
failed during daemon startup with an explicit recovery reason.

## Implementation

- Traffic ingestion terminal transitions use `context.WithoutCancel` plus a new
  timeout.
- Provider-load failures are also recorded through the bounded terminal context.
- `IngestionRunRepository.RecoverStaleRunning` atomically updates only stale
  `running` rows.
- The ingestion daemon executes stale-run recovery before starting the first
  traffic cycle.
- Recovery count and thresholds are printed in startup evidence.

## Verification

The installer executes:

```text
go test ./internal/config
go test ./internal/services/traffic/ingestion
go test ./internal/repository/postgres
go test ./cmd/ingest
go test -race ./internal/services/traffic/ingestion ./cmd/ingest
go test ./...
go vet ./...
git diff --check
```

PostgreSQL integration tests run when `TEST_DATABASE_URL` is configured. They
verify stale versus fresh selection and concurrent recovery ownership.

## Scope boundary

This increment closes ingestion-run terminalization and stale-run recovery only.
Provider observer failure isolation, response-size limits, daemon retry scheduling,
publication lifecycle, and complete fallback attempt evidence remain separate
follow-up changes.

---

## Canonical remediation history

### GFA-OPS-063 — ingestion runs could remain permanently `running`

1. **Finding / symptom.** Cancellation, process termination, or failure of a terminal-status write could leave durable ingestion history permanently in `running`.
2. **Root cause.** Terminalization was coupled too closely to the cancelled data-plane context and there was no startup-owned stale-run recovery path.
3. **Failure scenario.** A provider call or downstream processing is cancelled, the request context is already done, and the final status update inherits that cancellation; alternatively the process exits after creating the run but before terminalization.
4. **Impact.** Operational evidence becomes permanently ambiguous, later health/recovery analysis cannot distinguish live work from abandoned work, and run-state invariants are no longer trustworthy.
5. **Severity rationale.** **P2 retrospective.** The defect damages durable operational evidence and recovery semantics but does not by itself fabricate aircraft observations or bypass an authorization boundary.
6. **Existing guarantees violated.** Every created ingestion run must eventually become terminal; terminal rows must remain immutable; cancellation of work must not erase final lifecycle evidence.
7. **Considered solutions.** Reuse the request context for terminal writes; use an unbounded background context; add a bounded independent terminal context; add only startup cleanup; combine bounded terminalization with stale-run recovery.
8. **Chosen remediation.** Use `context.WithoutCancel` with `TRAFFIC_INGESTION_TERMINAL_TIMEOUT` for bounded terminal writes and execute atomic stale-run recovery during ingestion-daemon startup.
9. **Why this solution was selected.** It preserves caller context values while decoupling cleanup from cancellation, remains time-bounded, and repairs both normal cancellation and crash/termination residue.
10. **Rejected alternatives.** Caller-context terminalization remains cancellation-sensitive; unbounded `context.Background()` can hang shutdown; startup cleanup alone leaves avoidable stale rows after ordinary cancellation.
11. **Trade-offs.** Final status writes may outlive the original request briefly, and stale recovery requires a configured age threshold that must remain longer than a normal ingestion cycle.
12. **Regression tests / protection.** Config, ingestion service, repository, command, race, full backend and PostgreSQL stale/fresh plus concurrent-recovery tests documented above.
13. **Adversarial review findings.** Recovery must not touch fresh `running` rows, concurrent recovery must have one winner, and an already terminal row must not be reopened or rewritten.
14. **Remediation iterations.** The final contract combines bounded terminal contexts with startup recovery; neither mechanism alone covers the full cancellation-plus-process-loss failure surface.
15. **Residual risks and limitations.** A badly configured stale threshold can delay recovery or classify unusually long legitimate work as stale; external database unavailability can still postpone terminal evidence until a later recovery attempt.
16. **Operational or deployment consequences.** Operators own `TRAFFIC_INGESTION_TERMINAL_TIMEOUT` and `TRAFFIC_INGESTION_STALE_RUN_AFTER`; startup logs expose recovery count and thresholds.
17. **Exact evidence.** Historical implementation commit `10eaeaff5f40ea7b0432da6a795b6d9a36ff9034` (`fix: harden ingestion run lifecycle`). Historical pull-request/reviewer evidence is unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-OPS-063=CLOSED`.
19. **Prevention / future guard.** Any future ingestion lifecycle implementation must keep terminalization bounded and independent from caller cancellation, retain stale-run recovery, and preserve concurrency/immutability tests.
