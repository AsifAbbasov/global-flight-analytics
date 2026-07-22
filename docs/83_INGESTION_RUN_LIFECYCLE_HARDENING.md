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
