# Document 106 — Feature Processing Identity PostgreSQL List Fix

Status: IMPLEMENTED
Baseline commit: ab452c0cd039619e842c1991ec1bed10a42e5665

## Continuous Integration failure

Backend Continuous Integration run `30151121170` applied migration `026`
successfully. The PostgreSQL correctness step then failed before the feature
pipeline verifier.

The non-cursor PostgreSQL feature-snapshot list query accepted only three SQL
parameters:

```text
trajectory_id
schema_version
limit
```

The production caller already supplied four parameters:

```text
trajectory_id
schema_version
processing_version
limit
```

The SQL query therefore omitted the processing-version predicate and reused
placeholder `$3` for the limit.

## Correction

The non-cursor query now enforces:

```sql
AND processing_version = $3
LIMIT $4
```

A regression test owns the SQL placeholder layout and the production argument
order. The permanent feature-processing-identity audit also requires both
fragments.

## Closure condition

This correction is locally complete only after all tests, race tests, `go vet`
and architecture audits pass. Final FP-02 closure still requires a green
PostgreSQL 16 Integration job on the corrective commit.

```text
FEATURE_PROCESSING_IDENTITY_POSTGRES_LIST_FIX=ENFORCED
```

## Canonical UUID test contract

The PostgreSQL store normalizes trajectory UUID text before binding SQL
arguments. The regression test therefore compares the first argument with the
lowercase canonical form rather than the uppercase fixture spelling.

---

## Canonical remediation history

### GFA-DB-131 — processing-version migration left the non-cursor PostgreSQL list query on the old parameter contract

1. **Finding / symptom.** After processing version became part of snapshot identity, the non-cursor PostgreSQL list path still omitted `processing_version` from SQL while the production caller supplied it as an additional argument.
2. **Root cause.** The schema/store API migration updated caller argument construction but one SQL branch retained its pre-migration predicate and placeholder numbering.
3. **Failure scenario.** A processing-aware list call reaches the old SQL: the query filters only trajectory/schema, binds `$3` as limit, while the caller sends trajectory/schema/processing-version/limit.
4. **Impact.** The production list path can fail at PostgreSQL binding/execution and, absent the argument-count failure, would also lack processing-version isolation semantics.
5. **Severity rationale.** **P1 retrospective.** This was a production PostgreSQL correctness regression introduced during a durable identity migration and was detected by Backend CI before closure.
6. **Existing guarantees violated.** Every read path for processing-aware snapshots must filter by the same canonical identity dimensions and have SQL placeholders aligned with bound arguments.
7. **Considered solutions.** Revert processing-version filtering for list; alter caller arguments only; add the missing predicate and renumber limit.
8. **Chosen remediation.** Add `AND processing_version = $3`, move `LIMIT` to `$4`, and bind canonical UUID/schema/processing-version/limit in the same order.
9. **Why this solution was selected.** It restores semantic identity isolation and fixes the concrete SQL parameter contract without weakening the new schema model.
10. **Rejected alternatives.** Reverting caller/version filtering would reopen `FP-02`; dropping only the extra caller argument would make the query execute but return mixed processing generations.
11. **Trade-offs.** The SQL source contract becomes more explicit and requires tests to own placeholder order.
12. **Regression tests / protection.** A unit test asserts predicate/placeholder layout and production argument order; the permanent processing-identity audit requires both SQL fragments; PostgreSQL 16 Integration executes the production store path.
13. **Adversarial review findings.** Canonical trajectory UUID normalization must be asserted before SQL binding so test fixtures do not mistake input spelling for the actual database argument contract.
14. **Remediation iterations.** `ab452c0c…` introduced processing identity; CI run `30151121170` exposed this incomplete list branch; `f18d4368…` corrected it. The next CI run then exposed a separate isolated-fixture defect documented in 107.
15. **Residual risks and limitations.** Static SQL-fragment checks cannot prove all future dynamic query changes; PostgreSQL integration remains required.
16. **Operational or deployment consequences.** No new migration beyond `026`; the production list read now correctly selects one processing generation.
17. **Exact evidence.** Corrective commit `f18d43689d53301db862bc10c0445c90dc6f277d` (`fix: bind processing version in snapshot list query`); Backend CI failure run `30151121170` is recorded in this document. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-DB-131=CLOSED`.
19. **Prevention / future guard.** Identity migrations must enumerate every SQL read branch and test placeholder order plus bound-argument order, not only schema and write paths.
