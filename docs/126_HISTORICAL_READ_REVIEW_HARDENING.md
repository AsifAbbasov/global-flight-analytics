# Historical Read Review Hardening

Status: closed

## Scope

This change hardens `apps/api/internal/historicalintelligence/historicalread` as the bounded PostgreSQL evidence boundary for Historical Intelligence.

## Corrected contracts

- All production datasets are read inside one read-only PostgreSQL `REPEATABLE READ` transaction.
- Flight and trajectory overlap predicates implement the half-open interval `[start, end)`.
- Mutable flight and trajectory rows are captured in append-only version tables and reconstructed at the analytical `AsOfTime`.
- Historical queries earlier than the version-history coverage boundary fail closed.
- Route membership is selected by trajectory event time rather than route calculation time.
- The latest admissible route version per trajectory is selected before the global row limit.
- Exact matched-row counts provide the coverage denominator; `limit + 1` remains only a bounded pagination sentinel.
- Route JSON has both a total byte budget and an explicit payload fingerprint.
- Route payload decoding is owned by the Historical Read boundary; downstream builders use decoded route results.
- Nullable identifiers retain explicit availability evidence instead of being erased by `COALESCE`.
- Numeric-to-float conversion uses explicit PostgreSQL rounding: twelve decimal places for quality scores and eight for coordinates.
- Alternative transaction executors are validated through the same record invariants as production reads.

## Compatibility decisions

`nil, error` constructor returns and nil-safe error unwrapping remain idiomatic Go and are not treated as defects. PostgreSQL-specific dependencies remain inside the PostgreSQL adapter. The old raw `RouteJSON` field is retained only for legacy test-fixture compatibility; production reads clear it after decoding, and downstream builders no longer parse persistence JSON.

## Database impact

Migration `028_harden_historical_read_snapshot.sql` creates temporal version history, capture triggers, the history coverage marker, and query-aligned indexes. Existing rows are backfilled as their currently known versions. Earlier overwritten states cannot be reconstructed retroactively, so an `AsOfTime` earlier than the migration coverage boundary is rejected.

## Formal closure evidence

The engineering remediation was completed by:

- Hardening commit: `b67546391984b4726e05d67a51471d401f921e29` (`fix: harden historical read integrity`).
- Final engineering commit: `98750a7eb5972cd770e6333f46cd0855eca8ad0e` (`test: align historical read fixtures with integrity`).

GitHub Actions run `30298888993` for commit `98750a7eb5972cd770e6333f46cd0855eca8ad0e` completed with all required jobs successful after the PostgreSQL infrastructure-only retry:

```text
Backend Quality=SUCCESS
Backend Race Safety=SUCCESS
PostgreSQL 16 Integration=SUCCESS
Backend Container=SUCCESS
```

The successful PostgreSQL job applied and verified the production migration catalog, ran the PostgreSQL correctness integration tests, and verified the PostgreSQL feature pipeline.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
HISTORICAL_READ_REVIEW_STATUS=CLOSED
```

## Canonical remediation history

Historical adversarial-review identities/comments are not reconstructed beyond the repository record. Unlike Documents 124–125, this review contains historical exact-commit Continuous Integration evidence: run `30298888993` on final engineering commit `98750a7eb5972cd770e6333f46cd0855eca8ad0e`. Severity labels are retrospective.

### GFA-DB-231 — Historical datasets were not guaranteed to come from one PostgreSQL snapshot

1. **Finding / symptom:** production Historical Read could query Flight State, Trajectory, Route and related evidence without one shared read snapshot.
2. **Root cause:** dataset queries were individually correct but transaction ownership was not the explicit aggregate contract.
3. **Failure scenario:** one dataset observes a row before a concurrent update while another observes the row after that update inside the same analytical request.
4. **Impact:** Historical Results can combine evidence that never coexisted at one database state.
5. **Severity rationale:** **P1 retrospective** because cross-dataset snapshot skew can create internally contradictory durable analytics.
6. **Existing guarantees violated:** one Historical snapshot request must represent one consistent database view.
7. **Considered solutions:** tolerate read-committed skew, manually coordinate timestamps, or run all production reads inside one read-only `REPEATABLE READ` transaction.
8. **Chosen remediation:** Historical Read owns one read-only PostgreSQL `REPEATABLE READ` transaction for all production datasets in a snapshot.
9. **Why this solution was selected:** PostgreSQL snapshot isolation directly enforces the required consistency boundary without distributed reconstruction logic.
10. **Rejected alternatives:** independent per-dataset transactions and application-level timestamp guesses.
11. **Trade-offs:** a snapshot transaction remains open for the bounded duration of all dataset reads.
12. **Regression tests / protection:** PostgreSQL snapshot integration tests, migration-backed fixtures, `historicalreadreviewaudit -strict`.
13. **Adversarial review findings:** PostgreSQL-specific transaction machinery remains correctly confined to the adapter boundary.
14. **Remediation iterations:** hardening `b67546391984b4726e05d67a51471d401f921e29`; final fixture/integrity alignment `98750a7eb5972cd770e6333f46cd0855eca8ad0e`.
15. **Residual risks and limitations:** repeatable-read guarantees database consistency, not correctness of upstream source timestamps or data already persisted incorrectly.
16. **Operational or deployment consequences:** bounded read-only transaction usage increases snapshot consistency with ordinary PostgreSQL MVCC cost.
17. **Exact evidence:** hardening/final commits above; GitHub Actions run `30298888993`; PostgreSQL 16 Integration SUCCESS; permanent historical-read audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every new dataset added to Historical Snapshot must be loaded through the same managed transaction and covered by integration tests.

### GFA-DATA-232 — Flight and trajectory overlap predicates did not enforce one half-open interval contract

1. **Finding / symptom:** overlap selection could disagree at exact bucket/window boundaries.
2. **Root cause:** temporal membership predicates were not uniformly defined as `[start, end)`.
3. **Failure scenario:** an event exactly at `end` is included in both adjacent windows or excluded inconsistently between Flight and Trajectory reads.
4. **Impact:** double counting, boundary gaps and non-deterministic adjacent-period comparison.
5. **Severity rationale:** **P1 retrospective** because a one-timestamp boundary error can directly change historical counts and route membership.
6. **Existing guarantees violated:** adjacent analytical windows must partition event time without overlap or holes.
7. **Considered solutions:** closed intervals, open intervals, per-query conventions, or one half-open convention.
8. **Chosen remediation:** production Flight and Trajectory overlap predicates implement `[start, end)` consistently.
9. **Why this solution was selected:** half-open windows compose cleanly for adjacent periods and match the planner contract.
10. **Rejected alternatives:** inclusive end boundaries that double-count shared edges.
11. **Trade-offs:** callers must understand that an event at the exact end belongs to the following window.
12. **Regression tests / protection:** exact-boundary PostgreSQL read tests and strict audit.
13. **Adversarial review findings:** this contract is event-time based and independent from storage/update timestamps.
14. **Remediation iterations:** `b67546391984b4726e05d67a51471d401f921e29`, finalized by `98750a7e...` fixture alignment.
15. **Residual risks and limitations:** timestamp precision remains limited by source/persistence precision.
16. **Operational or deployment consequences:** corrected boundary selection only; no API change.
17. **Exact evidence:** implementation commits, overlap predicate tests, CI run `30298888993`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** all new Historical time predicates must use the canonical half-open window helpers/tests.

### GFA-DB-233 — Mutable Flight and Trajectory rows could not be reconstructed honestly at historical `AsOfTime`

1. **Finding / symptom:** current mutable rows did not preserve the versions needed to answer an analytical request as of an earlier time.
2. **Root cause:** primary tables represented current state without append-only temporal history or a declared history coverage boundary.
3. **Failure scenario:** a row updated after the requested `AsOfTime` is read in its newer form and retroactively changes a historical result.
4. **Impact:** historical replay becomes system-time dependent and can rewrite past analytics.
5. **Severity rationale:** **P1 retrospective** because historical evidence can be fabricated from future mutations.
6. **Existing guarantees violated:** `AsOfTime` must bind the version of mutable evidence used in historical analytics.
7. **Considered solutions:** trust current rows, snapshot whole databases, append full raw copies in application code, or add temporal version tables/capture triggers with a coverage marker.
8. **Chosen remediation:** migration 028 adds append-only Flight/Trajectory version history, capture triggers, coverage marker and query-aligned indexes; reads reconstruct the admissible version at `AsOfTime`.
9. **Why this solution was selected:** it provides deterministic temporal reconstruction inside PostgreSQL without duplicating application persistence workflows.
10. **Rejected alternatives:** pretending backfilled current rows reconstruct overwritten history that was never captured.
11. **Trade-offs:** version history consumes additional storage and can only guarantee reconstruction from the migration coverage boundary forward.
12. **Regression tests / protection:** temporal version integration tests, coverage-boundary rejection tests, migration-catalog verification and strict audit.
13. **Adversarial review findings:** existing rows are backfilled only as currently known versions; the documentation explicitly refuses to invent pre-migration overwritten states.
14. **Remediation iterations:** migration/implementation in `b6754639...`; final fixture alignment `98750a7e...`.
15. **Residual risks and limitations:** an `AsOfTime` earlier than history coverage is unavailable by design and fails closed.
16. **Operational or deployment consequences:** migration `028_harden_historical_read_snapshot.sql`; additional version tables/triggers and storage growth.
17. **Exact evidence:** migration 028, hardening/final commits, PostgreSQL Integration in Actions run `30298888993`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** mutable evidence used by historical analytics must define temporal version ownership and a truthful reconstructable-history boundary before production use.

### GFA-DATA-234 — Historical route membership used route calculation time instead of trajectory event time

1. **Finding / symptom:** route evidence could be assigned to a Historical window according to when Route Intelligence calculated/stored the result rather than when the trajectory occurred.
2. **Root cause:** system/provenance time and domain event time were conflated in membership selection.
3. **Failure scenario:** a trajectory from one period is recalculated later and appears in the later period's route analytics.
4. **Impact:** routes move between historical windows based on processing latency rather than aviation events.
5. **Severity rationale:** **P1 retrospective** because period membership directly determines counts, ratios and comparisons.
6. **Existing guarantees violated:** Historical analytics are event-time based unless explicitly documented otherwise.
7. **Considered solutions:** route calculation time, storage time, trajectory event time, or dual membership models.
8. **Chosen remediation:** route membership is selected by trajectory event window while calculation/storage timestamps remain provenance/version-selection evidence.
9. **Why this solution was selected:** trajectory time represents the domain event being analyzed.
10. **Rejected alternatives:** using processing time as a proxy for flight occurrence.
11. **Trade-offs:** route version selection requires joining/reconciling route evidence to trajectory identity/time.
12. **Regression tests / protection:** event-time route membership tests and strict audit.
13. **Adversarial review findings:** later Route Builder hardening separately validates scoped route evidence and does not reopen membership ownership.
14. **Remediation iterations:** `b67546391984b4726e05d67a51471d401f921e29` plus final `98750a7e...`.
15. **Residual risks and limitations:** correctness depends on authoritative trajectory event timestamps.
16. **Operational or deployment consequences:** query semantics/indexes adjusted by migration 028.
17. **Exact evidence:** implementation commits, route event-time tests, Actions run `30298888993`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every persisted derived result must distinguish domain event time from calculation/storage provenance before Historical membership logic is added.

### GFA-DATA-235 — Global row limiting could occur before selecting the latest admissible route version per trajectory

1. **Finding / symptom:** the bounded route query could truncate candidate versions before deduplicating/selecting the correct version for each trajectory.
2. **Root cause:** result limiting and semantic latest-version selection were ordered incorrectly.
3. **Failure scenario:** multiple versions for early trajectories consume the global limit and exclude later trajectories even though only one admissible version per trajectory should count.
4. **Impact:** biased route evidence, false truncation and incorrect coverage/sample counts.
5. **Severity rationale:** **P1 retrospective** because bounded query ordering can systematically remove valid trajectories from analytics.
6. **Existing guarantees violated:** semantic dedup/version selection must precede global pagination/row limits.
7. **Considered solutions:** increase limit, deduplicate in application after fetch, or select latest admissible version per trajectory in SQL before limiting.
8. **Chosen remediation:** PostgreSQL selects the latest admissible route version per trajectory first, then applies the global row limit.
9. **Why this solution was selected:** it preserves bounded reads while making the bounded set semantically correct.
10. **Rejected alternatives:** arbitrary over-fetch multipliers and limit-before-dedup behavior.
11. **Trade-offs:** SQL is more explicit and depends on query-aligned indexes.
12. **Regression tests / protection:** multiple-version/limit-order tests, PostgreSQL integration, strict audit.
13. **Adversarial review findings:** stable version ordering remains tied to admissible analytical/provenance timestamps and record identity.
14. **Remediation iterations:** `b6754639...`; final fixture/invariant alignment `98750a7e...`.
15. **Residual risks and limitations:** the configured global limit can still make the final read incomplete; incompleteness is separately measured with exact counts.
16. **Operational or deployment consequences:** query/index behavior changed under migration 028.
17. **Exact evidence:** implementation commits, latest-version selection tests, Actions run `30298888993`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** SQL limits must be applied only after all semantic uniqueness/version-selection predicates that define one analytical record.

### GFA-DATA-236 — Incomplete Historical read coverage used a pagination sentinel instead of the exact matched-row denominator

1. **Finding / symptom:** `limit + 1` could be interpreted as evidence for total matched rows even though it proves only that more rows exist.
2. **Root cause:** pagination detection and analytical coverage denominator were conflated.
3. **Failure scenario:** 10,000 matching rows with a limit of 100 produce a denominator near 101 and materially overstate represented coverage.
4. **Impact:** bucket/series confidence and completeness can be grossly overstated under truncation.
5. **Severity rationale:** **P1 retrospective** because confidence/coverage is a trust metric and the denominator can be mathematically false.
6. **Existing guarantees violated:** incomplete-read coverage must use the exact size of the eligible evidence set.
7. **Considered solutions:** heuristic denominator, sentinel-only truncation flag, approximate count, or exact matched-row count inside the same snapshot.
8. **Chosen remediation:** exact matched-row counts provide the coverage denominator; `limit + 1` remains only the bounded pagination/truncation sentinel.
9. **Why this solution was selected:** it separates bounded payload selection from truthful completeness measurement.
10. **Rejected alternatives:** `selected/(selected+1)` or other limit-derived coverage estimates.
11. **Trade-offs:** exact counts require additional bounded count queries in the same transaction.
12. **Regression tests / protection:** high-match/low-limit coverage tests, snapshot consistency tests and strict audit.
13. **Adversarial review findings:** later Historical Route hardening reuses the exact denominator and rejects pair-specific incomplete coverage where no pair denominator exists.
14. **Remediation iterations:** `b6754639...`; final `98750a7e...`.
15. **Residual risks and limitations:** exact count reflects database-eligible rows, not semantic quality after every downstream builder filter; downstream limitations remain required.
16. **Operational or deployment consequences:** additional count work within the read transaction, bounded by indexed predicates.
17. **Exact evidence:** implementation commits, exact matched-count tests, PostgreSQL CI run `30298888993`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** truncation sentinels must never be reused as analytical coverage denominators.

### GFA-PERF-237 — Route payload reads lacked one bounded byte budget and deterministic payload identity

1. **Finding / symptom:** route JSON could consume unbounded aggregate bytes within a bounded row count and the raw payload lacked an explicit fingerprint at the read boundary.
2. **Root cause:** bounding considered row cardinality but not variable-size JSON evidence, while payload identity was delegated downstream.
3. **Failure scenario:** a small number of very large route payloads exhaust memory/latency budget or semantically changed JSON passes through without one read-boundary identity.
4. **Impact:** resource exhaustion risk and weaker provenance/idempotency for persisted route evidence.
5. **Severity rationale:** **P1 retrospective** because the combined defect affects both bounded production operation and identity of analytical source evidence.
6. **Existing guarantees violated:** every externalized persistence payload must be bounded by both cardinality and bytes and have deterministic integrity identity.
7. **Considered solutions:** row limit only, per-row limit only, downstream hashing, or total byte budget plus explicit payload fingerprint.
8. **Chosen remediation:** Historical Read enforces a total route JSON byte budget and records/validates an explicit route payload fingerprint.
9. **Why this solution was selected:** total bytes reflect real memory/transport cost and boundary-owned fingerprinting keeps source identity with source decoding.
10. **Rejected alternatives:** assuming row count bounds JSON size and hashing only after domain transformation.
11. **Trade-offs:** large but valid route datasets may be reported incomplete when byte budget is reached.
12. **Regression tests / protection:** route byte-limit/fingerprint tests and strict audit.
13. **Adversarial review findings:** downstream builders consume decoded Route Results rather than reinterpreting raw JSON.
14. **Remediation iterations:** `b6754639...`; final `98750a7e...`.
15. **Residual risks and limitations:** configured byte limit remains an operator trade-off between evidence volume and resource budget.
16. **Operational or deployment consequences:** bounded memory/JSON processing and explicit incomplete-read evidence under byte truncation.
17. **Exact evidence:** implementation commits, route payload budget/fingerprint tests, Actions run `30298888993`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** variable-size persisted payloads must declare both row/cardinality and total-byte limits plus deterministic payload identity.

### GFA-ARCH-238 — Persistence JSON decoding ownership leaked into downstream Historical builders

1. **Finding / symptom:** downstream analytics could parse raw `RouteJSON` instead of receiving a decoded persistence-boundary value.
2. **Root cause:** storage representation and domain evidence ownership were not cleanly separated.
3. **Failure scenario:** two builders decode the same persistence JSON with different compatibility/validation behavior and produce divergent analytics.
4. **Impact:** duplicated persistence coupling and potential decoder/validation drift.
5. **Severity rationale:** **P3 retrospective** because this is primarily an ownership/maintainability defect; strict validation defects are addressed separately.
6. **Existing guarantees violated:** PostgreSQL representation decoding belongs to the adapter/evidence boundary, not analytical builders.
7. **Considered solutions:** keep raw JSON everywhere, introduce shared downstream decoder, or decode once in Historical Read and pass typed route results.
8. **Chosen remediation:** Historical Read owns route payload decoding; production reads clear raw JSON after decoding and downstream builders consume typed decoded results.
9. **Why this solution was selected:** it gives one persistence compatibility boundary and reduces domain coupling.
10. **Rejected alternatives:** reintroducing persistence parsing in each analytical package.
11. **Trade-offs:** the read adapter carries more decoding responsibility and must remain synchronized with Route Contract versions.
12. **Regression tests / protection:** decoded-read fixtures, downstream no-raw-JSON audit assertions and strict review audit.
13. **Adversarial review findings:** raw `RouteJSON` remains only for legacy test-fixture compatibility, not as production analytical input.
14. **Remediation iterations:** `b6754639...`; final `98750a7e...`.
15. **Residual risks and limitations:** typed decoded results still require downstream domain validation appropriate to each builder, as later Route hardening enforces.
16. **Operational or deployment consequences:** no additional database migration beyond migration 028; less repeated JSON parsing downstream.
17. **Exact evidence:** implementation commits, compatibility tests, historical-read audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** storage-format decoding must have one adapter owner and production domain builders must not parse persistence JSON directly.

### GFA-DATA-239 — Nullable identifiers could be erased into apparently present fallback values

1. **Finding / symptom:** nullable identifiers could lose availability semantics through SQL `COALESCE` or equivalent zero/empty substitution.
2. **Root cause:** convenience scanning flattened absence and present-empty identity into one value state.
3. **Failure scenario:** a missing flight/trajectory/route identifier is read as an ordinary empty/fallback identifier and participates in deduplication or matching.
4. **Impact:** fabricated identity relationships and incorrect evidence grouping.
5. **Severity rationale:** **P1 retrospective** because absent identity can be treated as real identity and alter analytical membership.
6. **Existing guarantees violated:** nullable persistence fields must preserve explicit availability through the domain read boundary.
7. **Considered solutions:** COALESCE to empty, sentinel strings, pointer/availability fields, or typed nullable scanning with explicit availability.
8. **Chosen remediation:** nullable identifiers retain explicit availability evidence and are not erased by SQL coercion.
9. **Why this solution was selected:** it distinguishes missing evidence from legitimate zero/empty values without sentinel ambiguity.
10. **Rejected alternatives:** empty-string identity sentinels and implicit availability inference.
11. **Trade-offs:** read/domain records carry explicit availability state.
12. **Regression tests / protection:** nullable identifier PostgreSQL fixtures and identity validation tests.
13. **Adversarial review findings:** idiomatic `nil,error` constructor behavior is unrelated; this finding concerns persisted data availability, not Go return conventions.
14. **Remediation iterations:** `b6754639...`; final fixture alignment `98750a7e...`.
15. **Residual risks and limitations:** upstream writes must still enforce whether an identifier is legitimately optional for each record type.
16. **Operational or deployment consequences:** more faithful scan semantics; no additional migration beyond read hardening.
17. **Exact evidence:** implementation commits, nullable identifier tests, Actions run `30298888993`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** nullable persistence values must never be normalized into domain-present values without an explicit policy and availability evidence.

### GFA-DATA-240 — PostgreSQL numeric-to-float conversion had implicit rounding semantics

1. **Finding / symptom:** conversion of PostgreSQL numeric quality/coordinate values to float64 did not declare one deterministic rounding boundary.
2. **Root cause:** driver conversion behavior was relied on instead of a contract-owned PostgreSQL rounding policy.
3. **Failure scenario:** the same stored numeric value produces small representation differences across query/driver paths and changes fingerprints or validation comparisons.
4. **Impact:** reproducibility drift in Historical source evidence.
5. **Severity rationale:** **P2 retrospective** because the issue is deterministic conversion precision rather than gross value fabrication.
6. **Existing guarantees violated:** persistence-to-domain numerical conversion must be deterministic and versioned/documented.
7. **Considered solutions:** scan arbitrary-precision decimals through the domain, accept driver conversion, or round in SQL to explicit precision before float conversion.
8. **Chosen remediation:** SQL rounds quality scores to twelve decimal places and coordinates to eight before float64 conversion.
9. **Why this solution was selected:** it gives database-side deterministic precision while preserving existing non-monetary float64 contracts.
10. **Rejected alternatives:** project-wide decimal migration and undocumented driver-dependent conversion.
11. **Trade-offs:** precision beyond the declared decimal places is intentionally discarded at the read boundary.
12. **Regression tests / protection:** numeric conversion precision fixtures and strict audit.
13. **Adversarial review findings:** the chosen precision is a persistence conversion policy, not presentation rounding.
14. **Remediation iterations:** `b6754639...`; `98750a7e...`.
15. **Residual risks and limitations:** future fields with different precision needs require explicit separate policies.
16. **Operational or deployment consequences:** deterministic SQL expressions; negligible query cost.
17. **Exact evidence:** implementation commits, numeric conversion tests, PostgreSQL Integration run `30298888993`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every PostgreSQL numeric-to-binary64 boundary must state precision/rounding semantics in query code and tests.

### GFA-TEST-241 — Alternative Historical transaction executors could bypass production record invariants

1. **Finding / symptom:** injected/alternative transaction executors were not guaranteed to validate reconstructed records with the same invariants as normal production reads.
2. **Root cause:** testability/abstraction paths could bypass the canonical scanning/validation route.
3. **Failure scenario:** tests or alternate composition report success for malformed records that the production executor rejects, hiding contract drift.
4. **Impact:** false test confidence and divergence between supported executor paths.
5. **Severity rationale:** **P2 retrospective** because a verification boundary could fail to detect production-integrity regressions.
6. **Existing guarantees violated:** every supported executor path must produce records under the same historical read invariants.
7. **Considered solutions:** remove executor injection, trust test doubles, or route alternative executors through the same record validation functions.
8. **Chosen remediation:** alternative transaction executors are validated through the same record invariants as production reads.
9. **Why this solution was selected:** preserves testability/composition flexibility without weakening correctness.
10. **Rejected alternatives:** maintaining a looser fixture-only contract.
11. **Trade-offs:** alternative executor fixtures must provide fully valid current record metadata.
12. **Regression tests / protection:** executor parity fixtures, historical-read strict audit, full Backend Quality and PostgreSQL integration.
13. **Adversarial review findings:** the final compatibility commit `98750a7e...` specifically aligns historical read fixtures with the hardened integrity contract.
14. **Remediation iterations:** hardening `b6754639...`; fixture/compatibility closure `98750a7eb5972cd770e6333f46cd0855eca8ad0e`.
15. **Residual risks and limitations:** unsupported ad-hoc executors outside the package contract are not guaranteed.
16. **Operational or deployment consequences:** stronger test/alternate composition parity; no runtime feature change.
17. **Exact evidence:** final commit `98750a7e...`, GitHub Actions run `30298888993` with Backend Quality, Race Safety, PostgreSQL 16 Integration and Backend Container SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** any new executor seam must reuse canonical scan/record validation and be included in parity tests before production acceptance.
