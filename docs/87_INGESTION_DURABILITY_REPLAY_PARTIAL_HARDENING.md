# Document 87 — Ingestion Durability, Replay and Partial Status Hardening

Status: Implemented Engineering Contract v1.0
Project: Global Flight Analytics

## Purpose

This increment closes three related correctness findings from the original
Ingestion, Provider Adapters and Orchestration review:

```text
durable ingestion evidence before provider transport
replay-safe Flight State persistence
explicit partial completion after durable observation writes
```

It also repairs the migration catalog version collision introduced when the
OurAirports publication lifecycle was committed as a second migration 019.

## Durable pre-request ingestion run

`LoadAndProcessByPoint` now commits the `running` ingestion row before invoking
any provider method. Therefore a process crash, container termination or panic
inside provider transport leaves a durable row that startup stale-run recovery
can finalize.

The row initially records the provider chain source. After fallback selection,
the source is updated while the run is still `running` so the final evidence
records the provider that actually supplied the observations.

When orchestration proves that a local budget or polling denial did not execute
an external request, the provisional `running` row is deleted. This preserves
the previous rule that local denial must not become a false failed provider run.
Deletion is allowed only for a still-running row with no linked Flight States.

## Partial terminal status

The repository port and PostgreSQL repository now expose `MarkPartial`.

The ingestion service uses:

```text
StoredFlightStateCount > 0 and later processing error → partial
StoredFlightStateCount = 0 and processing error       → failed
```

This aligns the run status with the existing independent durability-unit
contract: source observations may commit before quality or trajectory
derivations fail and enter reconciliation.

## Replay-safe Flight State identity

Migration 023 installs the provider observation identity:

```text
(source_name, icao24, observed_at)
```

The migration fails closed when historical duplicates already exist rather than
silently deleting observations or cascading deletion into quality evidence.

The production insert uses:

```sql
ON CONFLICT (source_name, icao24, observed_at)
DO NOTHING
```

`FlightStateRepository.SaveFlightStatesCounted` returns the number of rows
actually inserted. The application service detects this optional capability and
propagates the real count into `ProcessAndStoreResult` and the ingestion run.
The legacy repository interface remains available for non-PostgreSQL tests and
adapters.

## Migration catalog repair

The canonical sequence is now:

```text
019_data_quality_parent_integrity.sql
020_stage14_correctness_hardening.sql
021_trajectory_query_profiles.sql
022_provider_publication_lifecycle.sql
023_ingestion_durability_replay_partial.sql
```

A permanent production-catalog regression test rejects duplicate versions,
sequence gaps and future accidental renaming of these ownership boundaries.

## Verification

The increment includes:

- ordering tests proving durable run creation precedes provider invocation;
- fail-fast tests proving provider transport does not start when run creation
  fails;
- provisional-run deletion tests for local access denial;
- partial-versus-failed terminal status tests;
- selected fallback source update coverage;
- counted replay persistence tests;
- PostgreSQL active-run lifecycle integration coverage;
- replay-safe SQL contract checks;
- migration catalog uniqueness and sequence checks;
- race detector coverage;
- full backend tests;
- `go vet`;
- code review policy gates;
- rollback-safe installation.

## Remaining review boundary

This document does not claim complete Ingestion review closure. Remaining
separate work includes exact in-memory deduplication semantics, Airplanes.live
nullable telemetry and conversion bounds, duration overflow protection,
malformed-record batch policy, provider constructor validation, and explicit
classification of multi-instance rate limiting and health-aware fallback.

---

## Canonical remediation history

### GFA-OPS-072 — provider transport could start before durable ingestion-run evidence existed

1. **Finding / symptom.** External provider work could begin before the `running` ingestion row was durably committed.
2. **Root cause.** Ingestion evidence creation was ordered after or too near the provider operation instead of being a prerequisite to transport.
3. **Failure scenario.** The process crashes, is killed, or panics inside provider transport before the run record reaches PostgreSQL.
4. **Impact.** A real external request can occur with no durable ingestion-run evidence for startup recovery or later audit.
5. **Severity rationale.** **P1 retrospective.** The system could lose durable provenance for real production source activity.
6. **Existing guarantees violated.** External data acquisition must have durable run ownership before transport; startup recovery requires a persisted running row.
7. **Considered solutions.** Create run after provider success; create and commit before transport; hold the run insert in one long transaction around the provider call; use only logs.
8. **Chosen remediation.** Commit the provisional `running` run before invoking any provider method, then update selected source and terminal status as later lifecycle steps.
9. **Why this solution was selected.** It survives process loss without holding a database transaction open across network I/O.
10. **Rejected alternatives.** Post-success creation loses crash evidence; a long transaction across HTTP increases contention and failure coupling; logs are not durable relational lifecycle evidence.
11. **Trade-offs.** Local admission denial can leave a provisional row unless explicitly deleted, requiring the guarded no-linked-state deletion path.
12. **Regression tests / protection.** Ordering tests, fail-fast create-run tests, provisional local-denial deletion tests and active-run PostgreSQL integration.
13. **Adversarial review findings.** Provider transport must not start if run creation fails; provisional deletion is permitted only while the row remains `running` and has no linked Flight States.
14. **Remediation iterations.** The earlier local-denial semantics from Document 85 were preserved by adding provisional-run cleanup rather than reverting pre-request durability.
15. **Residual risks and limitations.** A database outage prevents the provider request entirely by design; stale recovery remains required for crashes after durable creation.
16. **Operational or deployment consequences.** PostgreSQL availability becomes an explicit admission prerequisite for traffic provider calls.
17. **Exact evidence.** Historical implementation commit `302158e4c9cbfb8532ee03147f6dcd31603b72fa` (`fix: harden ingestion durability and replay safety`). Historical pull-request/reviewer evidence unavailable.
18. **Final canonical status.** `GFA-OPS-072=CLOSED`.
19. **Prevention / future guard.** External ingestion commands must persist durable run ownership before transport and tests must assert call ordering plus fail-fast behavior.

### GFA-DATA-073 — Flight State replay had no provider-observation uniqueness contract

1. **Finding / symptom.** Replaying the same provider observations could insert duplicate Flight State rows.
2. **Root cause.** Persistence identity did not encode the source observation key `(source_name, icao24, observed_at)` and inserts lacked conflict handling for replay.
3. **Failure scenario.** A retry/replay fetches the same observation set and writes it again under new internal identifiers.
4. **Impact.** Historical counts, trajectories, quality evidence and analytics can be inflated by duplicate source observations.
5. **Severity rationale.** **P1 retrospective.** This directly corrupts analytical input cardinality and downstream evidence.
6. **Existing guarantees violated.** One provider observation identity must map to at most one persisted Flight State regardless of replay/internal IDs.
7. **Considered solutions.** Application-only dedup; delete duplicates after insert; database unique identity with `ON CONFLICT DO NOTHING`; use internal UUID as replay identity.
8. **Chosen remediation.** Migration 023 adds uniqueness on `(source_name, icao24, observed_at)` and production inserts use conflict-do-nothing; actual inserted count is returned and propagated.
9. **Why this solution was selected.** PostgreSQL is the final concurrency-safe owner and the key represents source evidence rather than replay-generated internal identifiers.
10. **Rejected alternatives.** In-memory dedup is process-local; cleanup after insertion can cascade or distort evidence; internal UUIDs change across replay.
11. **Trade-offs.** The migration fails closed if historical duplicates already exist, requiring explicit operator resolution instead of automatic deletion.
12. **Regression tests / protection.** Counted replay persistence tests, SQL contract checks and migration-catalog tests.
13. **Adversarial review findings.** Migration must not silently delete historical duplicates because linked quality evidence may exist; repeated insert must report zero newly inserted rows rather than count attempted rows.
14. **Remediation iterations.** The repository gained `SaveFlightStatesCounted` while the legacy interface was retained for isolated non-PostgreSQL callers.
15. **Residual risks and limitations.** The chosen key assumes one provider does not intentionally publish materially distinct observations for the same aircraft at the exact same observation timestamp.
16. **Operational or deployment consequences.** Historical duplicate preflight can block migration 023; operators must resolve such data explicitly.
17. **Exact evidence.** Historical implementation commit `302158e4c9cbfb8532ee03147f6dcd31603b72fa`. Historical PR/reviewer evidence unavailable.
18. **Final canonical status.** `GFA-DATA-073=CLOSED`.
19. **Prevention / future guard.** Replayable source tables must define source-evidence identity at the database boundary and expose actual inserted counts rather than attempted counts.

### GFA-DATA-074 — durable observation writes followed by downstream failure were classified only as failed

1. **Finding / symptom.** A run could durably store Flight States and then fail in quality/trajectory processing without a terminal status that represented partial success.
2. **Root cause.** Run lifecycle had success/failed semantics but no status derived from the durability-unit boundary after source observations committed.
3. **Failure scenario.** Flight States commit successfully; a later derivation fails and enters reconciliation; the run is marked failed as if no durable source evidence survived.
4. **Impact.** Operators and analytics cannot distinguish total ingestion failure from partial processing failure with preserved raw observations.
5. **Severity rationale.** **P1 retrospective.** The durable run status misrepresents what source data actually exists and can distort recovery/monitoring decisions.
6. **Existing guarantees violated.** Terminal run evidence must correspond to committed durability units; preserved observations must not be reported as total failure.
7. **Considered solutions.** Keep binary success/failed; roll back source observations with every downstream failure; add explicit partial status tied to stored count.
8. **Chosen remediation.** Add `MarkPartial`; use `StoredFlightStateCount > 0 && later error → partial`, otherwise failed when no states persisted.
9. **Why this solution was selected.** It matches the architecture where source observations and derived quality/trajectory work are intentionally separate durability units.
10. **Rejected alternatives.** Binary status lies about durable evidence; global rollback couples independent stages and sacrifices recoverable source data.
11. **Trade-offs.** Consumers must understand a third terminal state and reconciliation workflows must handle partial runs explicitly.
12. **Regression tests / protection.** Partial-versus-failed service tests, selected-source update coverage and ingestion-run repository integration.
13. **Adversarial review findings.** Attempted insert count is insufficient; status must depend on actual stored row count after replay conflict handling.
14. **Remediation iterations.** `SaveFlightStatesCounted` and partial status were coupled so replay no-ops do not falsely imply durable new observations.
15. **Residual risks and limitations.** Partial status records that some evidence survived, not which downstream derivations remain incomplete; reconciliation owns that detail.
16. **Operational or deployment consequences.** Monitoring and run consumers must treat `partial` as terminal but reconciliation-relevant rather than equivalent to success or failed.
17. **Exact evidence.** Historical implementation commit `302158e4c9cbfb8532ee03147f6dcd31603b72fa`. Historical PR/reviewer evidence unavailable.
18. **Final canonical status.** `GFA-DATA-074=CLOSED`.
19. **Prevention / future guard.** Any new independently durable ingestion stage must preserve terminal status semantics that distinguish no durable output from partial durable output.

### GFA-DB-075 — a new duplicate migration version reintroduced catalog ambiguity

1. **Finding / symptom.** The OurAirports publication lifecycle was introduced as a second migration version `019`, colliding with the existing Data Quality parent-integrity migration.
2. **Root cause.** Post-Stage-14 migration creation bypassed or preceded enforcement of a sufficiently broad canonical production-catalog regression guard.
3. **Failure scenario.** Two different SQL files claim version 019, making ordering/history identity ambiguous for the production migrator.
4. **Impact.** Clean or incremental migration behavior becomes non-deterministic/blocked and migration history loses trustworthy version ownership.
5. **Severity rationale.** **P1 retrospective.** This is a recurrence of a class already proven capable of invalidating stage closure and clean database bootstrap.
6. **Existing guarantees violated.** Migration versions must be unique, ordered and stable across the production catalog.
7. **Considered solutions.** Accept duplicate version based on filename order; edit old applied migration; renumber the new publication migration and install sequence regression tests.
8. **Chosen remediation.** Canonicalize publication lifecycle as `022`, introduce ingestion durability as `023`, and enforce unique versions, sequence and expected ownership through production-catalog tests.
9. **Why this solution was selected.** It preserves historical applied migration ownership while repairing the newly introduced collision without rewriting older canonical files.
10. **Rejected alternatives.** Filename tie-breaking leaves two histories for one version; rewriting already canonical historical migration identity creates deployment divergence.
11. **Trade-offs.** Explicit expected migration ownership makes intentional future renumbering a reviewed contract change instead of a transparent filesystem edit.
12. **Regression tests / protection.** Production migration catalog uniqueness, sequence and expected canonical filename checks through version 023.
13. **Adversarial review findings.** The recurrence demonstrates that a prior one-time catalog cleanup is insufficient; every new migration must pass the production-catalog path.
14. **Remediation iterations.** This is a post-closure recurrence of the duplicate-version class previously captured by `GFA-DB-013`, but it is registered separately because it occurred later and required new remediation evidence.
15. **Residual risks and limitations.** A catalog test only protects files included in its production discovery path; bypassing that path in future tooling would reintroduce risk.
16. **Operational or deployment consequences.** Migration 022/023 ownership becomes fixed; clean bootstrap and upgrade validation must run before merge.
17. **Exact evidence.** Historical remediation commit `302158e4c9cbfb8532ee03147f6dcd31603b72fa`; the document explicitly records the second-019 collision and repaired sequence. Historical PR/reviewer evidence unavailable.
18. **Final canonical status.** `GFA-DB-075=CLOSED`.
19. **Prevention / future guard.** New migrations must pass production-catalog uniqueness/sequence tests in CI; duplicate version recurrence is treated as a release blocker, not a documentation typo.
