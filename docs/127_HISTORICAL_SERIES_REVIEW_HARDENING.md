# Historical Series Review Hardening

Status: closed

## Scope

This change hardens `apps/api/internal/historicalintelligence/historicalseries` as the generic time-series assembly boundary for Historical Intelligence.

## Accepted findings

- A single global coverage ratio was incorrectly copied into every bucket.
- Missing source and generation timestamps were replaced with synthetic window timestamps.
- Malformed limitations were silently deleted.
- Distinct window exclusions with the same reason could collapse into one limitation.
- Total sample accumulation was not protected from integer overflow.
- The main build function mixed validation, provenance, point construction, status derivation, confidence, and final contract validation.

## Corrected contracts

- Every `BucketValue` carries explicit `CoverageEvidence`.
- Dataset consumers bind complete or incomplete read evidence before series construction.
- Incomplete reads produce a conservative per-bucket lower bound and mark buckets without represented evidence unavailable.
- Series status is derived from bucket statuses rather than one floating-point state surrogate.
- Series confidence is the mean temporal coverage across planned buckets; sample count remains evidence volume rather than an artificial confidence multiplier.
- `LatestSourceUpdatedAt` and `GeneratedAt` are required and are never synthesized from the analytical window.
- Limitations fail closed when malformed or duplicated.
- Window exclusion limitations include unique codes and exact excluded intervals.
- Total sample count uses checked integer addition.
- `Build` delegates validation, point construction, status derivation, confidence, and final validation to focused helpers.
- Historical Traffic, Airport, and Route builders use the new coverage binding contract.
- A permanent strict audit is part of Backend Continuous Integration.

## Rejected or already-closed findings

- Zero coverage already produced an unavailable series before this increment.
- Mutable `historicalwindow.Plan` evidence was already canonicalized and validated by the Historical Window hardening.
- Generic `float64` value transport remains intentional: the central metric catalog rejects fractional count values and values beyond the exact integer boundary.
- A complete, fully observed empty bucket may legitimately have confidence `1` with sample count `0`; sample count is not source completeness.
- Nil-safe error formatting remains idiomatic Go and is not treated as a domain null state.
- Direct access to the canonical `historicalwindow.Plan` value object is an intentional adjacent-domain contract, not a Law of Demeter violation.

## Formal closure evidence

The Historical Series engineering remediation was completed by:

- Baseline commit: `0cc0dc7be96860120d69c8dd53158b36ac72d9c6`.
- Hardening commit: `02bee5fd59d13927d0ffb995844c83d07327a2f9` (`fix: harden historical series integrity`).
- Final engineering compatibility commit: `c863d03e5de711b78ab94027dbf951129665c110` (`fix: align historical contract audit with series hardening`).

GitHub Actions run `30305541816` for commit `c863d03e5de711b78ab94027dbf951129665c110` completed with all required jobs successful:

```text
Backend Quality=SUCCESS
Backend Race Safety=SUCCESS
PostgreSQL 16 Integration=SUCCESS
Backend Container=SUCCESS
```

Backend Quality executed the Historical Contract, Historical Window, Historical Read, and Historical Series strict audits successfully. The repository-level tests, Go vet analysis, race-safety checks, PostgreSQL integration checks, and backend container build were successful.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
HISTORICAL_SERIES_ENGINEERING_DEBT=CLOSED
HISTORICAL_SERIES_ADDITIONAL_CODE_FIXES_REQUIRED=NO
FORMAL_CLOSURE_DOCUMENTATION_REQUIRED=NO
HISTORICAL_SERIES_REVIEW_STATUS=CLOSED
```

## Canonical remediation history

Historical reviewer identities and review-comment chronology are not reconstructed. The records below are reconstructed from the review document, repository implementation, permanent tests/audit, commits, and exact GitHub Actions run `30305541816`. Severity labels are retrospective.

### GFA-DATA-242 — Global read coverage was copied into every Historical bucket
1. **Finding / symptom:** one dataset-level coverage ratio was applied to all planned buckets.
2. **Root cause:** the series assembler treated read completeness as a scalar property instead of time-local evidence.
3. **Failure scenario:** one truncated interval makes a bucket with no represented evidence appear partially covered, or a well-represented bucket inherits an unrelated lower ratio.
4. **Impact:** bucket status, confidence, and downstream comparison semantics can misrepresent actual temporal evidence.
5. **Severity rationale:** **P1 retrospective** because durable historical analytics could claim evidence in the wrong time bucket.
6. **Existing guarantees violated:** bucket-local evidence ownership and deterministic time-window semantics.
7. **Considered solutions:** keep one global ratio, infer coverage from sample count, or bind explicit coverage evidence per bucket.
8. **Chosen remediation:** each `BucketValue` carries `CoverageEvidence`; incomplete reads produce conservative per-bucket lower bounds and unrepresented buckets become unavailable.
9. **Why selected:** coverage follows the evidence interval rather than an aggregate transport flag.
10. **Rejected alternatives:** sample count as a completeness surrogate; empty but fully observed buckets are legitimate.
11. **Trade-offs:** callers must bind coverage explicitly before series construction.
12. **Regression tests / protection:** per-bucket coverage, incomplete-read, zero-event and unavailable-bucket tests plus `historicalseriesreviewaudit`.
13. **Adversarial review findings:** zero coverage already mapped to unavailable before this increment and was not duplicated as a new defect.
14. **Remediation iterations:** hardening `02bee5fd59d13927d0ffb995844c83d07327a2f9`; compatibility completion `c863d03e5de711b78ab94027dbf951129665c110`.
15. **Residual risks and limitations:** incomplete source reads can provide only conservative lower-bound coverage when exact bucket-local matched counts do not exist.
16. **Operational or deployment consequences:** no migration; corrected semantics are enforced at series assembly.
17. **Exact evidence:** commits above; GitHub Actions run `30305541816` SUCCESS; permanent Historical Series audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** no time-series builder may project one aggregate coverage scalar onto independent buckets without explicit evidence semantics.

### GFA-DATA-243 — Missing source and generation timestamps were synthesized from the analytical window
1. **Finding / symptom:** absent provenance timestamps could be replaced by window-derived times.
2. **Root cause:** required provenance was treated as a convenience default rather than authoritative source evidence.
3. **Failure scenario:** a result with unknown source freshness appears to have been updated/generated at a plausible analytical boundary.
4. **Impact:** fabricated provenance weakens freshness, replay, and audit claims.
5. **Severity rationale:** **P1 retrospective** because missing evidence could be published as concrete provenance.
6. **Existing guarantees violated:** provenance timestamps must describe real source/generation events, never inferred analytical windows.
7. **Considered solutions:** window fallback, zero timestamp, optional provenance, or fail closed when required timestamps are absent.
8. **Chosen remediation:** require `LatestSourceUpdatedAt` and `GeneratedAt`; never synthesize either from the analytical window.
9. **Why selected:** preserves the distinction between event time and processing/source evidence time.
10. **Rejected alternatives:** plausible timestamp invention and silent optionality.
11. **Trade-offs:** callers with incomplete provenance now fail instead of receiving a synthetic result.
12. **Regression tests / protection:** missing-source-time, missing-generated-time and provenance validation tests plus strict audit.
13. **Adversarial review findings:** window canonicalization belongs to Historical Window and is not a provenance substitute.
14. **Remediation iterations:** `02bee5fd59d13927d0ffb995844c83d07327a2f9`, finalized by `c863d03e5de711b78ab94027dbf951129665c110`.
15. **Residual risks and limitations:** correctness still depends on upstream source timestamps being truthful.
16. **Operational or deployment consequences:** malformed provenance fails earlier; no database migration.
17. **Exact evidence:** implementation commits, provenance tests, run `30305541816` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** analytical window boundaries must never be reused as missing source or processing provenance.

### GFA-DATA-244 — Malformed Historical limitations were silently discarded
1. **Finding / symptom:** invalid limitation records could be deleted during normalization instead of invalidating the result.
2. **Root cause:** limitation cleanup favored permissive output over evidence integrity.
3. **Failure scenario:** an incomplete bucket arrives with malformed explanatory evidence; cleanup removes it and the resulting series appears less constrained than the source state.
4. **Impact:** missing or corrupted limitation evidence can be hidden from consumers.
5. **Severity rationale:** **P1 retrospective** because explainability is part of the Historical trust contract for partial/unavailable data.
6. **Existing guarantees violated:** malformed evidence must fail closed rather than disappear.
7. **Considered solutions:** silently drop, repair generic text, retain malformed values, or reject malformed/duplicate limitations.
8. **Chosen remediation:** limitation validation is strict; malformed or duplicate limitations make construction fail.
9. **Why selected:** preserves evidence honesty and avoids invented repairs.
10. **Rejected alternatives:** best-effort cleanup and generic replacement limitations.
11. **Trade-offs:** previously tolerated malformed inputs now require caller correction.
12. **Regression tests / protection:** malformed/duplicate limitation tests and permanent strict audit.
13. **Adversarial review findings:** nil-safe error formatting remains idiomatic and unrelated.
14. **Remediation iterations:** `02bee5fd59d13927d0ffb995844c83d07327a2f9` with final compatibility `c863d03e5de711b78ab94027dbf951129665c110`.
15. **Residual risks and limitations:** semantic quality of otherwise structurally valid limitation text remains domain-owner responsibility.
16. **Operational or deployment consequences:** invalid analytical evidence is rejected earlier.
17. **Exact evidence:** commits, limitation tests, run `30305541816` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** evidence normalization may canonicalize representation but must not erase invalid limitation state.

### GFA-DATA-245 — Distinct window exclusions could collapse when they shared one reason
1. **Finding / symptom:** exclusions with the same reason could deduplicate into one limitation even when intervals differed.
2. **Root cause:** limitation identity was keyed too coarsely by reason rather than by the excluded evidence interval.
3. **Failure scenario:** two separate excluded windows produce one surviving limitation, hiding the actual temporal shape of missing evidence.
4. **Impact:** explainability and replay diagnostics lose exact exclusion coverage.
5. **Severity rationale:** **P2 retrospective** because analytical values remain bounded but supporting limitation evidence is incomplete.
6. **Existing guarantees violated:** distinct excluded intervals require distinct deterministic evidence identity.
7. **Considered solutions:** reason-only deduplication, free-form text differences, or unique codes carrying exact interval boundaries.
8. **Chosen remediation:** window-exclusion limitations use unique codes and exact start/end evidence.
9. **Why selected:** deterministic identity follows the actual omitted interval.
10. **Rejected alternatives:** reason-only collapse and text-only disambiguation.
11. **Trade-offs:** more limitation records may be emitted for genuinely distinct exclusions.
12. **Regression tests / protection:** repeated-reason/different-window tests and strict audit.
13. **Adversarial review findings:** no central closed-world limitation registry was required.
14. **Remediation iterations:** `02bee5fd59d13927d0ffb995844c83d07327a2f9`.
15. **Residual risks and limitations:** very large exclusion sets remain bounded by upstream planning/read policies rather than this record format.
16. **Operational or deployment consequences:** more precise diagnostics, no migration.
17. **Exact evidence:** hardening commit, exclusion tests, run `30305541816` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** deduplication keys for evidence limitations must include every field that changes the omitted evidence claim.

### GFA-DATA-246 — Historical total sample accumulation could overflow integer arithmetic
1. **Finding / symptom:** total sample count summed bucket counts without checked addition.
2. **Root cause:** bounded individual counts were assumed to make aggregate integer arithmetic safe.
3. **Failure scenario:** a sufficiently large valid series wraps the aggregate count and publishes a smaller or negative support total.
4. **Impact:** corrupted evidence-volume metadata and downstream validation behavior.
5. **Severity rationale:** **P2 retrospective** because the primary metric value can remain correct while support evidence becomes invalid.
6. **Existing guarantees violated:** counts must remain non-negative, exact, and representable.
7. **Considered solutions:** wider transport, floating-point count, saturating arithmetic, or checked integer addition.
8. **Chosen remediation:** checked integer addition fails closed on overflow.
9. **Why selected:** preserves exact count semantics without changing public schemas.
10. **Rejected alternatives:** saturation and float count accumulation because both hide loss of exactness.
11. **Trade-offs:** pathological oversized series are rejected rather than partially represented.
12. **Regression tests / protection:** overflow-boundary tests and Historical Contract count validation.
13. **Adversarial review findings:** generic `float64` metric transport remains intentional for heterogeneous metric values; sample counts remain integers.
14. **Remediation iterations:** `02bee5fd59d13927d0ffb995844c83d07327a2f9`.
15. **Residual risks and limitations:** upstream planners still own practical series-size limits.
16. **Operational or deployment consequences:** deterministic fail-closed behavior on impossible aggregate support volume.
17. **Exact evidence:** implementation commit, overflow tests, run `30305541816` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every accumulated count must use checked arithmetic even when per-record values are bounded.

### GFA-MAINT-247 — Historical Series construction concentrated unrelated responsibilities in one build path
1. **Finding / symptom:** `Build` combined input validation, point creation, provenance, status, confidence and final contract validation.
2. **Root cause:** the initial implementation grew orchestration and domain calculations together.
3. **Failure scenario:** a change to one responsibility alters another invariant or makes targeted regression protection difficult.
4. **Impact:** higher change risk at a trust-sensitive analytical boundary.
5. **Severity rationale:** **P3 retrospective** because this is maintainability debt with an indirect correctness risk, not evidence of an independently wrong value.
6. **Existing guarantees violated:** focused ownership and auditable analytical policy boundaries.
7. **Considered solutions:** leave monolithic flow, create generic abstraction framework, or extract focused domain helpers.
8. **Chosen remediation:** `Build` delegates validation, point construction, status derivation, confidence, provenance and final validation to focused helpers.
9. **Why selected:** reduces semantic coupling without introducing cross-module abstractions.
10. **Rejected alternatives:** mechanical line-count refactoring and generic framework extraction.
11. **Trade-offs:** more internal functions/files and explicit call sequencing.
12. **Regression tests / protection:** existing series tests plus `historicalseriesreviewaudit` enforce the decomposed contract.
13. **Adversarial review findings:** direct use of canonical `historicalwindow.Plan` remains an intentional adjacent-domain contract.
14. **Remediation iterations:** `02bee5fd59d13927d0ffb995844c83d07327a2f9`, compatibility completion `c863d03e5de711b78ab94027dbf951129665c110`.
15. **Residual risks and limitations:** finite orchestration still exists in `Build`; decomposition does not eliminate all future complexity.
16. **Operational or deployment consequences:** none; internal maintainability hardening only.
17. **Exact evidence:** implementation commits, permanent audit, GitHub Actions run `30305541816` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** future series features should extend focused policy helpers rather than rebuilding validation/provenance logic inside the orchestration function.
