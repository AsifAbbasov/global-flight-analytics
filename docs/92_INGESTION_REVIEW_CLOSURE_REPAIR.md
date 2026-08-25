# Document 92 — Ingestion Review Closure Repair

Status: Implemented Engineering Contract v1.0
Project: Global Flight Analytics
Baseline: `b7bf2b762290e55a45fa8d40641435248d1aeddf`

## 1. Scope

This increment closes the remaining verified findings from the Ingestion,
Provider Adapters and Orchestration review without changing the approved
modular-monolith architecture or adding runtime infrastructure.

The repaired boundaries are:

```text
Retry-After duration overflow protection
OpenSky OAuth expires_in duration overflow protection
Open-Meteo missing metric preservation
PostgreSQL NULL persistence for unavailable weather metrics
typed atomic fail-whole OurAirports publication parsing
PostgreSQL isolated fixture migration alignment
accurate review closure wording
```

## 2. Bounded external duration conversion

Both numeric `Retry-After` parsing paths reject values that cannot be represented
as `time.Duration` before multiplying by `time.Second`.

OpenSky OAuth token responses use the same fail-closed rule for `expires_in`.
A missing or non-positive lifetime still receives the existing 1,800-second
engineering default. A positive value above the representable duration limit is
classified as `ErrTokenResponseInvalid`.

## 3. Open-Meteo metric availability

Open-Meteo current-weather values decode through nullable pointers. The adapter
therefore distinguishes:

```text
JSON 0    → available metric with value zero
JSON null → unavailable metric
missing   → unavailable metric
```

`weather.CurrentSnapshot` carries explicit metric availability while preserving
legacy in-process callers: snapshots without explicit availability retain the
previous all-metrics-available interpretation.

The current weather Hypertext Transfer Protocol response uses JSON `null` for
unavailable metrics. It does not publish invented zero observations.

## 4. PostgreSQL weather persistence

Migration 025 formalizes nullable weather metric columns and records the meaning
of `NULL` through column comments.

`WeatherRepository` validates only available metrics and sends unavailable
metrics to PostgreSQL as `NULL`. Existing complete snapshots keep their previous
validation and persistence behavior.

Weather Context production selection already requires every weather metric to be
non-NULL. Incomplete snapshots therefore remain stored as truthful source
evidence but are not promoted into complete Weather Context analysis.

## 5. OurAirports atomic publication policy

OurAirports remains an atomic fail-whole publication import.

A malformed CSV row or a row that violates the airport parsing contract returns
the typed `AtomicPublicationError`, matches
`ErrAtomicPublicationRejected`, and returns no partial airport collection.

This is a deliberate publication policy, not an accidental parser side effect.

## 6. PostgreSQL fixture alignment

The Flight State altitude fixture now applies both migrations required by the
current repository insert contract:

```text
006_flight_state_altitude_semantics.sql
023_ingestion_durability_replay_partial.sql
```

The provider publication fixture now uses the canonical renamed migration:

```text
022_provider_publication_lifecycle.sql
```

These changes repair isolated integration schemas. They do not modify production
migration history.

## 7. Formal closure gates

The review may be marked closed only when a new commit containing this increment
passes all of the following:

```text
Go formatting
complete backend tests
Go vet
project architecture and contract audit
code review policy audit
Stage 14 final audit
critical race tests
PostgreSQL 16 migration apply and replay
PostgreSQL correctness integration tests
backend container verification
```

Local tests without `TEST_DATABASE_URL` do not independently prove the PostgreSQL
gate. A successful GitHub Backend Continuous Integration run on the new commit is
the final closure evidence.

## 8. Closure statement

When every gate in Section 7 passes on the same new commit:

```text
Open technical findings: 0
Unclassified findings: 0
Ingestion review: CLOSED
Release decision: ACCEPTABLE
```

Until then, the correct status remains:

```text
Ingestion review: CONDITIONALLY ACCEPTABLE
Formal closure: PENDING CONTINUOUS INTEGRATION
```

## 9. Post-closure permanent race coverage

Document 96 expands the permanent Backend Race Safety matrix across OpenSky,
provider budget, selection, response, health, traffic application, traffic
ingestion and the ingestion composition root. This closes the remaining
Continuous Integration coverage gap identified by the original review.

---

## Canonical remediation history

### GFA-DATA-084 — external retry/token durations could overflow Go duration arithmetic

1. **Finding / symptom.** Numeric `Retry-After` values and OpenSky OAuth `expires_in` values could be multiplied by `time.Second` without first proving representable `time.Duration` bounds.
2. **Root cause.** External integer duration evidence crossed into Go duration arithmetic before overflow validation.
3. **Failure scenario.** A malformed or extreme positive duration is accepted from an HTTP header or OAuth JSON response and overflows during conversion, producing a wrapped/invalid scheduling or expiry interval.
4. **Impact.** Retry scheduling or token lifetime can become incorrect, potentially causing premature requests, unexpectedly long waits, or invalid authentication timing.
5. **Severity rationale.** **P2 retrospective.** The defect is on an untrusted external timing boundary and affects reliability/access behavior, but it does not directly fabricate persisted aviation observations.
6. **Existing guarantees violated.** External duration values must be validated before multiplication/conversion; positive out-of-range OAuth lifetime must fail closed rather than wrap.
7. **Considered solutions.** Blind conversion; clamp to maximum; accept provider values as trusted; explicit pre-multiplication bound checks with typed invalid-token classification.
8. **Chosen remediation.** Validate numeric seconds against the maximum representable `time.Duration` before multiplying; preserve the existing 1,800-second engineering default only for missing/non-positive OAuth lifetime, while out-of-range positive lifetime returns `ErrTokenResponseInvalid`.
9. **Why this solution was selected.** It distinguishes absent/defaultable evidence from malformed positive evidence and prevents arithmetic overflow without inventing a provider value.
10. **Rejected alternatives.** Clamping fabricates duration semantics; trusting external values leaves the overflow path; treating every invalid value as the default would hide malformed provider evidence.
11. **Trade-offs.** Extremely large provider values now fail explicitly even if a caller might otherwise tolerate them; this is intentional fail-closed behavior.
12. **Regression tests / protection.** Retry-After and OAuth expiry boundary tests, complete backend tests, `go vet`, Stage 14/source audits and same-commit CI closure gate.
13. **Adversarial review findings.** Bounds must be checked before multiplication; missing/non-positive OAuth expiry and positive overflow are semantically different cases and must not share one fallback.
14. **Remediation iterations.** The same conversion rule was applied to both Retry-After parsing surfaces and OAuth expiry so timing safety is consistent across provider access paths.
15. **Residual risks and limitations.** A representable but semantically incorrect remote duration cannot be detected without stronger provider-side evidence.
16. **Operational or deployment consequences.** Malformed duration evidence becomes a controlled typed failure rather than silent scheduler/token corruption.
17. **Exact evidence.** Historical engineering closure commit `6b922cbd9df1bff3f880ad120dd883b37f658e53` (`fix: close ingestion review findings`). Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-DATA-084=CLOSED`.
19. **Prevention / future guard.** Every external seconds-to-duration conversion must prove representable bounds before arithmetic and include overflow regression cases.

### GFA-DATA-085 — Open-Meteo missing metrics could be represented as observed zero across adapter, HTTP and persistence boundaries

1. **Finding / symptom.** Missing or `null` Open-Meteo current-weather metrics were not guaranteed to remain unavailable through domain mapping, HTTP output and PostgreSQL persistence.
2. **Root cause.** Weather metrics used value-only fields while metric availability was part of the source evidence; database columns and transport DTOs did not consistently own nullability semantics.
3. **Failure scenario.** A weather metric is absent/null upstream, decodes to a zero value, is persisted/displayed as a real zero observation, and may be consumed as complete Weather Context evidence.
4. **Impact.** Missing weather evidence can become fabricated measurements and contaminate weather-aware analytics/confidence.
5. **Severity rationale.** **P1 retrospective.** This is direct evidence fabrication across a source-to-database-to-API analytical path.
6. **Existing guarantees violated.** Missing evidence remains missing; legitimate zero remains distinct from unavailable; incomplete weather snapshots must not be promoted into complete Weather Context analysis.
7. **Considered solutions.** Treat zero as missing; drop incomplete snapshots; keep value-only fields and infer availability later; introduce explicit availability plus nullable HTTP/PostgreSQL representation.
8. **Chosen remediation.** Decode Open-Meteo metrics through nullable pointers; add `CurrentMetricAvailability`; emit JSON `null` for unavailable values; migration 025 permits/defines PostgreSQL `NULL`; repository validation applies only to available metrics; complete Weather Context selection continues to require non-NULL metrics.
9. **Why this solution was selected.** It preserves source truth end to end while retaining incomplete snapshots as useful provenance instead of deleting them.
10. **Rejected alternatives.** Zero-as-missing loses legitimate zeros; dropping incomplete records discards evidence; delayed inference cannot reliably distinguish missing from observed zero after information is lost.
11. **Trade-offs.** Weather contracts become more explicit and consumers must handle nullable metrics; apparent completeness may decrease because formerly fabricated zero values become unavailable.
12. **Regression tests / protection.** Adapter null/missing/zero tests, HTTP null serialization, repository NULL persistence/integration, migration 025 checks and same-commit Backend CI.
13. **Adversarial review findings.** JSON zero must remain available; JSON null and missing must remain unavailable; stored incomplete snapshots are valid provenance but cannot satisfy complete Weather Context selection.
14. **Remediation iterations.** The repair spans adapter, domain availability, DTO, schema and repository so no intermediate boundary can collapse absence back into zero.
15. **Residual risks and limitations.** Provider-specific sentinel values that are syntactically numeric but semantically unavailable require explicit future adapter handling.
16. **Operational or deployment consequences.** API clients and analytics must tolerate `null` weather values; migration 025 is required for truthful persistence.
17. **Exact evidence.** Historical engineering closure commit `6b922cbd9df1bff3f880ad120dd883b37f658e53`. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-DATA-085=CLOSED`.
19. **Prevention / future guard.** Optional provider metrics require end-to-end availability tests covering adapter → domain → persistence → HTTP; zero may never be used as an absence sentinel.

### GFA-CONTRACT-086 — OurAirports fail-whole publication behavior was not an explicit typed contract

1. **Finding / symptom.** OurAirports parsing effectively failed a publication on malformed rows, but that behavior was not represented as a deliberate typed publication policy.
2. **Root cause.** Parser failure behavior and ingestion policy were conflated, leaving callers unable to distinguish an intentional atomic-publication rejection from incidental parsing errors.
3. **Failure scenario.** A malformed CSV row occurs and a caller cannot safely determine whether partial airport rows are publishable or whether the entire publication must be rejected.
4. **Impact.** Future refactors could accidentally publish a partial reference-data snapshot or inconsistently handle malformed publications.
5. **Severity rationale.** **P2 retrospective.** The existing fail-whole behavior limited immediate data corruption, but the contract ambiguity threatened correctness of authoritative airport reference imports.
6. **Existing guarantees violated.** OurAirports publication import is atomic; malformed publication evidence must return no partial airport collection and must be machine-classifiable.
7. **Considered solutions.** Keep implicit parser errors; accept valid rows from malformed publication; add a configurable mixed-row policy; formalize fail-whole behavior with a typed error.
8. **Chosen remediation.** Define `AtomicPublicationError` matching `ErrAtomicPublicationRejected` and guarantee no partial airport collection on malformed/contract-invalid rows.
9. **Why this solution was selected.** OurAirports is publication-style reference data where a coherent snapshot is preferable to silently mixing accepted and rejected rows; typed policy prevents accidental behavior drift.
10. **Rejected alternatives.** Partial acceptance changes reference-data semantics and would require explicit rejected-row provenance that the approved design did not own; implicit errors remain ambiguous.
11. **Trade-offs.** One malformed row rejects the whole publication, reducing availability in exchange for coherent reference-data snapshots.
12. **Regression tests / protection.** Typed atomic-publication parsing tests plus production publication lifecycle and review closure gates.
13. **Adversarial review findings.** The policy must return zero partial airport output on rejection and remain distinct from the mixed-item live-traffic policy in Document 91.
14. **Remediation iterations.** The review explicitly separated live telemetry mixed-batch semantics from OurAirports publication atomicity instead of forcing one generic malformed-record rule.
15. **Residual risks and limitations.** Provider-wide malformed publications remain unavailable until corrected upstream or parser contract is deliberately updated.
16. **Operational or deployment consequences.** Operators see a typed publication rejection and the durable reservation can be released/retried rather than committing partial reference data.
17. **Exact evidence.** Historical engineering closure commit `6b922cbd9df1bff3f880ad120dd883b37f658e53`. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-CONTRACT-086=CLOSED`.
19. **Prevention / future guard.** Reference-data publication parsers must document atomic versus partial policy explicitly and expose a typed rejection contract with partial-output regression tests.

### GFA-TEST-087 — isolated PostgreSQL fixtures drifted behind current repository migration dependencies

1. **Finding / symptom.** Isolated Flight State/provider-publication PostgreSQL fixtures did not apply the complete/current migrations required by production repository SQL, including migration 023 and the canonical renamed publication migration 022.
2. **Root cause.** Test-local schemas selected historical migration subsets manually and were not automatically aligned when repository contracts/migration ownership changed.
3. **Failure scenario.** Integration tests fail with missing columns/constraints or validate an obsolete schema while the production migration catalog is correct.
4. **Impact.** CI can report false regressions or, worse, provide false confidence from tests that do not exercise the current repository contract.
5. **Severity rationale.** **P2 retrospective.** The defect primarily affects verification trust rather than production schema directly, but unreliable PostgreSQL evidence can invalidate closure decisions.
6. **Existing guarantees violated.** Integration fixtures that claim repository parity must include every migration required by the repository path under test and use canonical migration filenames.
7. **Considered solutions.** Hand-copy schema SQL; run entire production catalog for every isolated test; update required migration subset explicitly; loosen repository integration tests.
8. **Chosen remediation.** Align the Flight State fixture with migrations 006+023 and provider-publication fixture with canonical migration 022 while retaining isolated focused schemas.
9. **Why this solution was selected.** It restores exact dependencies without forcing every focused test to pay the full production-catalog setup cost.
10. **Rejected alternatives.** Hand-copied schema duplicates ownership; full catalog for every focused fixture adds unnecessary cost; loosening tests hides real contract drift.
11. **Trade-offs.** Focused fixtures still require maintenance when their repository dependencies change; the production-catalog integration remains the broader safety net.
12. **Regression tests / protection.** PostgreSQL 16 migration apply/replay, repository correctness integrations and canonical migration catalog tests.
13. **Adversarial review findings.** Not every minimal test-local table should be forced to mirror the full production schema; parity requirements apply only to fixtures claiming complete repository-path ownership.
14. **Remediation iterations.** This follows the earlier Stage 14 fixture-parity overreach lesson (`GFA-GOV-062`): fix concrete stale complete fixtures without imposing one global full-schema rule.
15. **Residual risks and limitations.** Future repository SQL changes can create new focused-fixture dependencies; source tests must continue to identify them explicitly.
16. **Operational or deployment consequences.** No production schema change from fixture alignment itself; CI evidence becomes representative of current repository behavior.
17. **Exact evidence.** Historical engineering closure commit `6b922cbd9df1bff3f880ad120dd883b37f658e53`. Historical adversarial-review/PR evidence unavailable.
18. **Final canonical status.** `GFA-TEST-087=CLOSED`.
19. **Prevention / future guard.** PostgreSQL fixture changes must be reviewed against repository SQL dependencies and canonical migration filenames; production-catalog integration remains mandatory alongside focused fixtures.

### GFA-GOV-088 — ingestion review closure could not be claimed from local/source evidence alone

1. **Finding / symptom.** The remaining ingestion fixes involved PostgreSQL, race-sensitive orchestration and container/runtime paths whose closure could not be proven by local tests without the exact same-commit GitHub Backend CI result.
2. **Root cause.** Review completion wording needed an explicit separation between implemented source changes and independently executed closure evidence.
3. **Failure scenario.** Documentation says `CLOSED` after local tests while PostgreSQL integration, critical race tests or backend container verification have not run on the exact candidate commit.
4. **Impact.** Governance can overstate review completion and make later failures look like new regressions when closure was never fully evidenced.
5. **Severity rationale.** **P2 retrospective.** This is a release/review evidence integrity problem, not a runtime defect, but it materially affects whether high-risk ingestion changes are safe to declare closed.
6. **Existing guarantees violated.** Closure requires implementation evidence plus exact-commit regression/CI evidence; local tests without `TEST_DATABASE_URL` do not prove PostgreSQL behavior.
7. **Considered solutions.** Declare closure when source is committed; rely on local tests; require only unit tests; use conditional status until the complete Backend CI gate succeeds on the same commit.
8. **Chosen remediation.** Document explicit `CONDITIONALLY ACCEPTABLE / PENDING CONTINUOUS INTEGRATION` state and permit `CLOSED` only after formatting, full tests, vet, audits, race, PostgreSQL 16 integration and container verification pass on the same commit.
9. **Why this solution was selected.** It keeps source readiness and verification readiness distinct and prevents documentation from outrunning executable evidence.
10. **Rejected alternatives.** Immediate closure is unsupported; unit-only closure misses the exact high-risk paths being remediated; local PostgreSQL absence cannot be treated as a pass.
11. **Trade-offs.** Formal closure may lag code completion until CI executes, but that delay is evidence discipline rather than technical debt.
12. **Regression tests / protection.** The Backend CI closure matrix listed in Section 7; subsequent Document 96 expands permanent race coverage across the full ingestion/provider ownership surface.
13. **Adversarial review findings.** A green result on a different SHA cannot close the candidate; skipped PostgreSQL integration cannot be silently equated to success; later missing race coverage must be tracked separately rather than rewriting historical closure.
14. **Remediation iterations.** Commit `6b922cbd9df1bff3f880ad120dd883b37f658e53` supplied the closure repair; later commit `1ddb65c5e5471ce180314cc38a4b6d7baad80cd3` expanded permanent ingestion race coverage, documented in Document 96.
15. **Residual risks and limitations.** Historical CI run identifiers/reviewer evidence are not fully reconstructed here; the repository preserves source and later permanent gates, while this retroactive record does not invent missing review events.
16. **Operational or deployment consequences.** No runtime change from governance wording; merge/closure decisions require the complete Backend CI evidence boundary.
17. **Exact evidence.** Engineering closure commit `6b922cbd9df1bff3f880ad120dd883b37f658e53`; post-closure race-coverage guard `1ddb65c5e5471ce180314cc38a4b6d7baad80cd3`. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-GOV-088=CLOSED` for the review-closure evidence contract. The distinct permanent race-coverage gap is owned by Document 96 and should be registered separately during the 93–102 audit.
19. **Prevention / future guard.** Review documents must distinguish source-ready, conditionally acceptable and exact-commit CI-closed states; required integration jobs may never be inferred from local success.
