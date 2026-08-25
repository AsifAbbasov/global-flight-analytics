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

---

## Canonical remediation history

### GFA-OPS-080 — production provider budgets were process-local

1. **Finding / symptom.** Fixed-window counters and provider-reported remaining state existed only in memory, so restart or a second process received a fresh budget view.
2. **Root cause.** Provider budget ownership was implemented as a process-scoped utility even though request limits apply across process lifetime and concurrent production workers.
3. **Failure scenario.** One process consumes provider budget and restarts, or two processes run concurrently; each believes more requests remain than the provider policy actually permits.
4. **Impact.** GFA can exceed intended/provider-observed request budgets, amplify rate limiting and jeopardize stable provider access.
5. **Severity rationale.** **P1 retrospective.** This is a production external-access correctness boundary with cross-process and restart consequences.
6. **Existing guarantees violated.** Request budget must survive restart and coordinate every production ingestion instance; acquisition must be atomic across all applicable limits.
7. **Considered solutions.** Keep one process only; persist periodic snapshots; use Redis; store counters/state transactionally in existing PostgreSQL.
8. **Chosen remediation.** Migration 024 adds durable fixed-window and provider-reported budget state; PostgreSQL row locking owns atomic acquisition, decrement, cooldown and probe leases.
9. **Why this solution was selected.** PostgreSQL already exists as shared durable infrastructure and satisfies cross-process coordination without adding a new runtime dependency.
10. **Rejected alternatives.** Single-process assumptions are not restart-safe; periodic snapshots race; Redis would add unnecessary infrastructure for the current modular-monolith/free-tier architecture.
11. **Trade-offs.** Every provider admission now depends on PostgreSQL and row-lock ordering, adding database round trips to request scheduling.
12. **Regression tests / protection.** Budget unit/delegation tests, deterministic lock-order validation, optional cross-instance PostgreSQL integration, race/full backend and migration catalog gates.
13. **Adversarial review findings.** Multiple applicable windows must lock in deterministic order; one denied window must prevent consumption from all windows; unknown/post-cooldown probing must allow only one durable lease owner.
14. **Remediation iterations.** The in-memory implementation was retained only for isolated tests/non-production callers while `cmd/ingest` was explicitly wired to `providerbudget.NewDurable`.
15. **Residual risks and limitations.** Durability enforces configured/observed policies but cannot discover undocumented provider quotas; stale provider headers remain bounded by the stored evidence available.
16. **Operational or deployment consequences.** Production ingestion requires PostgreSQL for budget admission and migration 024; restarting the daemon no longer resets provider quota state.
17. **Exact evidence.** Historical implementation commit `52a60d2b7136919e3a2ccf4850f6d542c6447461` (`fix: persist provider budget state`). Historical pull-request/reviewer evidence unavailable.
18. **Final canonical status.** `GFA-OPS-080=CLOSED`.
19. **Prevention / future guard.** Any production rate/budget policy shared across requests or instances must use durable atomic ownership; process-local budget stores remain test/non-production only.

### GFA-OPS-081 — exhausted provider-reported budget could deny without an actionable retry time

1. **Finding / symptom.** A provider-reported zero remaining budget could reject a request while returning an empty/zero `RetryAt`.
2. **Root cause.** The admission contract assumed provider retry headers would always be usable when remaining budget reached zero.
3. **Failure scenario.** Provider reports zero remaining but omits/invalidates the reset/retry header; orchestration receives denial with no future scheduling point.
4. **Impact.** Daemon retry logic can busy-loop or repeatedly reconsider an unavailable provider without a bounded recovery time.
5. **Severity rationale.** **P2 retrospective.** This is a retry/reliability defect; it does not by itself corrupt persisted aviation data.
6. **Existing guarantees violated.** A budget denial must always provide a non-zero future retry decision so the scheduler can remain bounded.
7. **Considered solutions.** Return zero retry; permanently disable provider; invent a provider quota/reset; apply an explicitly labeled engineering fallback probe interval.
8. **Chosen remediation.** When zero remaining has no usable provider retry header, schedule a one-minute engineering fallback before allowing a bounded future probe.
9. **Why this solution was selected.** It prevents tight loops without misrepresenting a local fallback as a provider-published quota.
10. **Rejected alternatives.** Zero retry is non-actionable; permanent disable can starve recovery; fabricating a provider reset time would violate evidence honesty.
11. **Trade-offs.** The one-minute interval is an engineering policy and may be shorter or longer than the provider's undocumented real reset time.
12. **Regression tests / protection.** Fake-store retry fallback tests and durable provider-budget integration verify non-zero retry behavior.
13. **Adversarial review findings.** The fallback must remain explicitly local policy evidence and must not overwrite a real later provider retry time.
14. **Remediation iterations.** The retry fallback was integrated with durable probe leasing so multiple processes do not all probe simultaneously after the local interval.
15. **Residual risks and limitations.** Without provider reset evidence, the exact optimal retry time remains unknowable.
16. **Operational or deployment consequences.** Monitoring may show an engineering retry fallback; consumers must not label it a provider quota promise.
17. **Exact evidence.** Historical implementation commit `52a60d2b7136919e3a2ccf4850f6d542c6447461`. Historical PR/reviewer evidence unavailable.
18. **Final canonical status.** `GFA-OPS-081=CLOSED`.
19. **Prevention / future guard.** Every denial path used by the daemon scheduler must produce an actionable bounded retry time or an explicitly documented local fallback.
