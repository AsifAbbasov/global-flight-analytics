# Document 96 — Ingestion Race Coverage Closure

Status: Implemented Engineering Contract v1.0
Project: Global Flight Analytics
Baseline: `ae4d486d2341974a47173e2aedd78da530130cf6`

## Purpose

The original Ingestion, Provider Adapters and Orchestration review correctly
identified that permanent Backend Race Safety coverage omitted several
concurrency-sensitive ingestion packages.

The permanent race matrix now includes:

```text
cmd/ingest
internal/integrations/opensky
internal/orchestration/providerbudget
internal/orchestration/providerdecision
internal/orchestration/providerfallback
internal/orchestration/providerpolicy
internal/orchestration/providerresponse
internal/services/providerhealth
internal/services/traffic/application
internal/services/traffic/ingestion
```

These packages own provider budgets, retry evidence, health collection,
fallback decisions, ingestion lifecycle transitions and traffic persistence
orchestration. Ordinary tests do not replace race-detector evidence.

After Backend Quality, Backend Race Safety, PostgreSQL 16 Integration and
Backend Container pass on the same commit:

```text
Ingestion race coverage debt: CLOSED
Open accepted Ingestion review findings: 0
Unclassified Ingestion review findings: 0
Release blockers: 0
```

---

## Canonical remediation history

### GFA-GOV-098 — permanent race-detector coverage omitted concurrency-sensitive ingestion ownership

1. **Finding / symptom.** The permanent Backend Race Safety matrix did not include several ingestion/provider packages that owned mutable/concurrent provider budgets, decisions, health, retries and ingestion orchestration.
2. **Root cause.** Race coverage had grown around selected critical packages rather than a repository-owned inventory of concurrency-sensitive ingestion boundaries.
3. **Failure scenario.** A later change introduces a data race in provider budget state, fallback/decision collectors, provider health, ingestion lifecycle or composition code; unit tests pass because the package is not executed under `-race` in required CI.
4. **Impact.** Concurrency defects can reach merge despite a nominal race-safety gate, weakening confidence in precisely the orchestration paths that coordinate shared state and parallel work.
5. **Severity rationale.** **P2 retrospective.** This is a verification/governance gap rather than a confirmed production race, but the omitted code owns concurrency-sensitive production behavior and ordinary tests are insufficient evidence.
6. **Existing guarantees violated.** Required race CI must cover all critical concurrency owners identified by the review; a green ordinary test suite cannot substitute for race-detector execution.
7. **Considered solutions.** Rely on full ordinary tests; run `go test -race ./...` for the entire repository; add only the one package named by the review; maintain a targeted explicit critical ingestion race matrix.
8. **Chosen remediation.** Expand Backend Race Safety to `cmd/ingest`, OpenSky, provider budget/decision/fallback/policy/response, provider health, traffic application and traffic ingestion packages.
9. **Why this solution was selected.** It closes the known concurrency ownership surface while keeping CI cost bounded instead of mechanically running the whole repository under the race detector.
10. **Rejected alternatives.** Ordinary tests cannot detect races; one-package repair leaves sibling concurrency owners uncovered; all-repository race execution adds cost without evidence that every package needs it.
11. **Trade-offs.** The explicit matrix requires maintenance when concurrency ownership moves or new critical packages appear.
12. **Regression tests / protection.** Backend Race Safety itself is the permanent guard and is required alongside Backend Quality, PostgreSQL 16 Integration and Backend Container for the review closure boundary.
13. **Adversarial review findings.** The matrix must include composition (`cmd/ingest`) as well as lower-level mutable services because races can arise in wiring/lifecycle ownership; a package may not be removed merely because ordinary tests remain green.
14. **Remediation iterations.** Document 92 recorded the closure-evidence gap and later guard requirement; commit `1ddb65c5…` implemented the permanent expanded race matrix as a separate post-closure hardening step rather than rewriting the earlier review history.
15. **Residual risks and limitations.** Targeted race coverage cannot prove absence of races in packages outside the identified critical matrix; future concurrency ownership requires explicit reclassification.
16. **Operational or deployment consequences.** CI cost increases modestly; no runtime behavior or infrastructure is added.
17. **Exact evidence.** Historical implementation/guard commit `1ddb65c5e5471ce180314cc38a4b6d7baad80cd3` (`test: expand ingestion race coverage`). Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-GOV-098=CLOSED`.
19. **Prevention / future guard.** New shared mutable/concurrent ingestion ownership must update the explicit race matrix in the same change; removal requires evidence that concurrency ownership no longer exists.
