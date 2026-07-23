# Provider Budget Durability and Retry Scheduling

Status: Implemented Engineering Contract v1.0

## Scope

This increment closes the process-local provider budget finding for the
traffic ingestion runtime. It covers fixed-window request counters,
provider-reported remaining budget, retry scheduling, restart recovery,
multi-process coordination, and production wiring.

## Previous Failure Mode

The provider budget manager stored fixed-window counters and OpenSky
reported state only in process memory. A daemon restart or second process
therefore received a fresh budget view. In addition, an exhausted
provider-reported budget could deny a request without a non-zero `RetryAt`.

## Implemented Contract

PostgreSQL migration 024 introduces two durable state owners:

- `provider_budget_fixed_windows` stores one bounded atomic counter per provider and policy limit; the row is reset
  in place when the canonical window changes;
- `provider_budget_reported_states` stores remaining-known state,
  remaining requests, retry time, observation time, and update time.

Fixed-window acquisition locks every applicable window in deterministic
policy order. The request is consumed only when every limit permits it.
Denial returns the latest applicable window end as `RetryAt`.

Provider-reported acquisition locks the provider row. Remaining budget is
decremented atomically, cooldown survives restart, and a durable probe lease
permits exactly one unknown or post-cooldown request before fresh headers are
observed. A second process receives the same non-zero retry time.

## Missing Provider Retry Header

When the provider reports zero remaining requests without a usable retry
header, the manager schedules a one-minute engineering fallback. This value
is not represented as a provider-published quota. It exists only to prevent
busy-loop denial with an empty retry time and to allow a bounded future probe.

## Production Wiring

`cmd/ingest` creates `postgres.ProviderBudgetStore` from the daemon
PostgreSQL pool and terminal timeout, then constructs the provider budget
manager through `providerbudget.NewDurable`.

The existing in-memory constructor remains available for isolated unit
tests and non-production callers. Production traffic ingestion no longer
uses it.

## Verification

The acceptance gate includes:

- provider budget unit tests;
- fake-store delegation and retry fallback tests;
- PostgreSQL constructor and deterministic lock-order validation;
- optional isolated PostgreSQL cross-instance integration;
- migration catalog regression through version 024;
- targeted race detector;
- complete backend tests;
- `go vet`;
- existing code review policy gates;
- clean Git diff validation.

## Remaining Review Boundaries

This increment does not close the complete Ingestion, Provider Adapters and
Orchestration review. The remaining boundaries are:

- health-aware primary and fallback selection;
- explicit malformed-item policy for otherwise successful provider batches.
