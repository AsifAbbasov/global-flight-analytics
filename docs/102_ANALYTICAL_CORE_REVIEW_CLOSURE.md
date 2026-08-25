# Document 102 — Analytical Core Review Closure

Status: Implemented Engineering Contract v1.0
Project: Global Flight Analytics
Closure baseline: `e48cb27655326fc6cc41d176a50120cdbf1ced6e`
Predecessor Continuous Integration evidence: Backend CI run `30128888158`

## 1. Purpose

This document completes classification and remediation of every finding from the
original Analytical Core Foundation review.

The closure policy follows the project Code Review Standard:

```text
fixed;
not applicable;
deliberately retained with evidence;
deliberately rejected with evidence;
deferred with owner, risk and revisit condition;
suggestion or nit that does not block merge.
```

No original finding is left implicit.

## 2. Closure totals

```text
Original findings: 19
Fixed: 14
Deliberately retained: 3
Rejected or non-blocking: 2
Not applicable: 0
Deferred: 0
Unclassified: 0
```

## 3. Original finding register

| Identifier | Original finding | Disposition | Primary evidence |
| --- | --- | --- | --- |
| AC-01 | Eligibility must precede aircraft-level deduplication. | FIXED | Document 97 and execution ordering tests. |
| AC-02 | Materially future observations must not contribute to metrics. | FIXED | Document 97 and trajectory eligibility tests. |
| AC-03 | Airport Activity was not owned by a concrete airport. | FIXED | Document 98 and airport-owned HTTP tests. |
| AC-04 | Traffic Density numerator and denominator used unrelated scopes. | FIXED | Document 98 and region-owned area tests. |
| AC-05 | Calculator and registry appeared to form a parallel runtime architecture. | DELIBERATELY_RETAINED | Document 100 classifies them as compatibility foundations outside server composition. |
| AC-06 | Metric execution exposed concrete dependencies and internal objects. | FIXED | Document 100, narrow executor interface and getter-removal tests. |
| AC-07 | Traffic Density permitted `NaN` or infinite area values. | FIXED | Document 97 and finite-number calculator tests. |
| AC-08 | Coverage Score and Data Freshness trusted caller snapshot parameters. | FIXED | Documents 99 and 101 plus server-owned production snapshot tests. |
| AC-09 | Analytical source provenance was incomplete or placeholder-based. | FIXED | Document 99, strict source validation and server query provenance. |
| AC-10 | `RecentRequest.Normalize` accepted a zero reference time. | FIXED | Document 100 and zero-time regression tests. |
| AC-11 | Parsed UUID identifiers were not canonicalized before deduplication. | FIXED | Document 100 and canonical UUID tests. |
| AC-12 | Public numeric precision had no explicit contract. | FIXED | Section 5 of this document and permanent source audit. |
| AC-13 | Long functions should fail a universal mechanical line threshold. | REJECTED_NON_BLOCKING | Section 8 and the existing code-review mechanical-rule policy. |
| AC-14 | `Value` plus `HasValue` was treated as redundant. | DELIBERATELY_RETAINED | Section 6 and status validation invariants. |
| AC-15 | Nullable analytical sections were treated as an unconditional design defect. | DELIBERATELY_RETAINED | Section 7 and status-discriminated response contracts. |
| AC-16 | Names containing `And` or `With` should be mechanically forbidden. | REJECTED_NON_BLOCKING | Section 8 and the existing code-review mechanical-rule policy. |
| AC-17 | Metric identifiers were inconsistent across packages and clients. | FIXED | Document 100 and canonical `traffic.*` identifiers. |
| AC-18 | Public analytical failures could expose raw operation error text. | FIXED | Document 99 and stable public failure regression tests. |
| AC-19 | Source names used a manual quadratic sorting helper. | FIXED | Standard-library `sort.Strings` and the final audit gate. |

## 4. Fixed correctness findings

Documents 97 through 101 form the implementation evidence chain:

```text
Document 97 — contributor ordering, future-observation exclusion and finite arithmetic;
Document 98 — airport ownership and one geographic Traffic Density scope;
Document 99 — strict provenance, placeholder rejection and safe public failures;
Document 100 — reference time, UUID canonicalization, Metric IDs and dependency boundaries;
Document 101 — server-owned Coverage Score and Data Freshness evidence.
```

The final closure increment additionally removes the manual quadratic source-name
sort and replaces it with the Go standard library.

## 5. Public numeric precision contract

Analytical API values are JSON numbers produced from their native Go numeric
values.

The backend:

```text
does not round metric values before HTTP publication;
does not convert metric values to formatted strings;
does not apply display precision to the response contract;
publishes `value` only when `has_value` is true.
```

The frontend may format numbers for a specific visual surface:

```text
ratios are displayed as percentages with one fractional digit;
density uses four significant display digits;
configured area uses whole-square-kilometre display formatting.
```

That formatting is presentation-only. It is not persisted, sent back to the
server or treated as the analytical value.

Consumers requiring deterministic comparisons must use the JSON number and the
metric identifier, not rendered text.

## 6. `Value` and `HasValue`

A numeric zero is a legitimate analytical result:

```text
zero active aircraft;
zero airport movements;
zero coverage;
zero freshness;
zero traffic density.
```

The zero value therefore cannot represent both a valid metric and absence.

`HasValue` is retained as the explicit presence discriminator:

```text
complete and limited results require `HasValue = true`;
denied and failed results require `HasValue = false`;
the HTTP response includes `value` only when `HasValue = true`.
```

`ValueOrZero` remains a compatibility helper on the internal generic value
object. The public response mapper does not use it and cannot convert absence
into a published zero.

## 7. Nullable result sections

Nullable result sections are status-discriminated evidence, not unstructured
optional state.

```text
eligibility may be absent when no trajectory capability decision exists;
failure is required only for failed results and forbidden otherwise;
data quality may be absent when no report can be built;
confidence report may be absent for denied or failed execution paths;
source observation bounds are both present or both absent.
```

Validation enforces the allowed combinations. The TypeScript contract mirrors
the same optional sections and retains `has_value` as the value discriminator.

Replacing all nullable fields with empty objects would erase the difference
between unavailable evidence and an evaluated empty result.

## 8. Rejected mechanical findings

### 8.1 Universal function-length threshold

A universal line-count failure rule is rejected.

Functions are decomposed when they mix responsibilities, hide resource
ownership, make policy untestable or create a demonstrated maintenance risk.
They are not split solely to satisfy an arbitrary line number.

The existing Code Review audit permanently rejects unsupported mechanical rules.

### 8.2 Mechanical `And` and `With` naming ban

The presence of `And` or `With` in an identifier is not itself a correctness,
cohesion or dependency defect.

Names remain subject to intent, ownership and readability review. A universal
substring ban is a non-blocking style preference and is rejected as a release
gate.

## 9. Compatibility package classification

`analytics/calculator` and `analytics/registry` remain compatibility foundation
packages.

They are:

```text
compiled and tested;
not stored by the runtime Executor;
not exposed through executor dependency getters;
not imported by the production server composition roots;
not required by the metric execution service interface.
```

A future breaking-version cleanup may remove their compatibility constructor
surface. That cleanup is not required to close the current production review.

## 10. Permanent verification

The repository adds:

```text
apps/api/tools/analyticalcorefinalaudit
```

The tool verifies:

```text
all nineteen finding classifications;
Documents 97 through 102 and public closure status;
eligibility-before-deduplication ordering;
future-observation and finite-number guards;
airport and region ownership;
server-owned quality evidence;
strict provenance and sanitized failures;
reference-time and UUID canonicalization;
canonical Metric IDs;
numeric presence and precision contracts;
compatibility-package runtime isolation;
frontend request and query ownership;
standard-library source sorting;
Backend and Frontend Continuous Integration reachability.
```

Backend Continuous Integration runs the tool in strict mode.

The Backend workflow path filter includes the analytical frontend API client,
React Query hooks and Analytics Overview. Frontend Continuous Integration
continues to run lint, TypeScript validation and a production build for every
`apps/web/**` change.

## 11. Required closure gates

Formal closure requires the closure commit to pass:

```text
Go formatting;
Analytical Core final audit unit tests;
complete backend tests;
analytical race tests;
Go vet;
project architecture and contract audit;
code review policy audit;
Stage 14 final audit;
strict Analytical Core final source audit;
Frontend ESLint;
Frontend TypeScript validation;
Frontend production build;
PostgreSQL 16 Integration;
Backend Container build and health smoke test.
```

The predecessor server-owned quality increment passed Backend CI run
`30128888158` on commit `e48cb27655326fc6cc41d176a50120cdbf1ced6e`. The closure commit must independently pass
the same required backend jobs and Frontend Continuous Integration.

## 12. Closure statement

When every gate in Section 11 passes on the same closure commit:

```text
Analytical Core Foundation review: CLOSED
ANALYTICAL_CORE_REVIEW_STATUS=CLOSED
Open release blockers: 0
Open required changes: 0
Deferred findings: 0
Unclassified original findings: 0
Release decision: ACCEPTABLE
```

Post-closure changes must preserve the permanent audit. A new confirmed blocker
requires a new document and an explicit status change; it must not silently
rewrite this evidence register.

## 13. Source-formatting resilience

The permanent source audit verifies semantic ownership rather than one exact
formatter layout.

Go permits a selector expression to be split after the period, and TypeScript
formatting may place a `URLSearchParams.set` argument on the following line.
The audit therefore removes formatting whitespace only for explicitly declared
compact source contracts.

The sampling interval is verified in
`analytical_production_snapshot.go`. The stale threshold is verified in its
actual public-handler owner, `analytical_metrics.go`. The frontend radius query
parameter is verified across Prettier-compatible line breaks.

Regression tests prove both directions:

```text
valid multiline Go and TypeScript contracts pass;
a genuinely absent compact contract fails.
```
## 14. Post-closure frontend framework security baseline

The Analytical Core remains closed while its frontend publication surface moves
from Next.js 16.2.9 to the patched 16.2.11 security release.

This update does not change analytical formulas or public metric contracts. It
closes a dependency-security failure discovered by the independent Frontend
Continuous Integration evidence check.

---

## Canonical remediation history

### GFA-CONTRACT-114 / AC-12 — public analytical numeric precision had no explicit contract

1. **Finding / symptom.** The Analytical Core did not explicitly define whether backend metric values were rounded/formatted before publication or whether display precision was part of the API value contract.
2. **Root cause.** Numeric calculation and presentation formatting existed in different layers but the boundary between them was implicit rather than documented and permanently checked.
3. **Failure scenario.** A future backend or frontend change rounds a density/ratio before HTTP publication, serializes a number as formatted text, or treats rendered UI precision as the analytical value; consumers comparing raw metrics receive changed semantics without a declared contract change.
4. **Impact.** Numerical reproducibility and downstream comparison can drift even when formulas themselves remain unchanged.
5. **Severity rationale.** **P2 retrospective.** The gap is an API/analytical contract ambiguity with interoperability risk, not evidence that existing calculations were already numerically wrong.
6. **Existing guarantees violated.** The backend must publish native analytical numbers; presentation rounding must remain UI-only; absence and numeric zero must remain distinguishable through `has_value`.
7. **Considered solutions.** Standardize fixed decimal rounding in the backend; publish all values as strings; explicitly preserve native JSON numbers and leave formatting to presentation; leave precision unspecified.
8. **Chosen remediation.** Document and audit that backend values are native JSON numbers with no pre-publication rounding/string formatting, while frontend percentage/significant-digit formatting is presentation-only and never becomes stored/server input.
9. **Why this solution was selected.** It preserves calculation precision and lets each visual surface choose display formatting without changing analytical contract values.
10. **Rejected alternatives.** Backend display rounding loses information; strings weaken numeric APIs; unspecified precision permits silent semantic changes.
11. **Trade-offs.** Consumers that need deterministic textual rendering must format values themselves; they should compare numeric values plus Metric ID, not UI text.
12. **Regression tests / protection.** The Analytical Core final audit checks numeric presence/precision ownership and frontend publication contracts; status validation preserves `has_value` semantics.
13. **Adversarial review findings.** A valid metric value of zero must not be confused with absent value, so precision ownership cannot be separated from the explicit `HasValue` presence discriminator.
14. **Remediation iterations.** The final closure audit added the explicit backend/frontend precision contract after formula and evidence ownership were already stabilized in Documents 97–101.
15. **Residual risks and limitations.** IEEE-754/JSON number behavior still applies to consumers; the contract does not promise decimal arbitrary precision or financial-grade fixed-point semantics.
16. **Operational or deployment consequences.** None.
17. **Exact evidence.** Historical closure implementation commit `8aa8dfa9f0cb0f5eae94497939633f100a863ef8` (`audit: close analytical core review findings`). Original review ID: `AC-12`. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-CONTRACT-114=CLOSED`.
19. **Prevention / future guard.** Backend metric mappers may not introduce display rounding or formatted numeric strings without an explicit public-contract change and corresponding Analytical Core audit update.

### GFA-PERF-115 / AC-19 — source ordering used a manual quadratic sorting helper

1. **Finding / symptom.** Analytical source names were ordered with a handwritten nested-loop string sort instead of the standard library.
2. **Root cause.** Local deterministic ordering was implemented manually even though Go already provides a well-tested sorting primitive.
3. **Failure scenario.** Source cardinality grows or the helper is modified incorrectly; unnecessary quadratic work and bespoke ordering logic remain on a common publication path.
4. **Impact.** Avoidable algorithmic overhead and maintenance risk exist in provenance construction; the issue is bounded by expected small source counts.
5. **Severity rationale.** **P3 retrospective.** No production correctness failure was demonstrated; this is a simplification/performance-maintainability finding with low current scale risk.
6. **Existing guarantees violated.** Prefer standard-library algorithms over bespoke equivalents when semantics match; analytical publication ordering must remain deterministic with minimal custom logic.
7. **Considered solutions.** Keep the manual helper; optimize it manually; use `sort.Strings`; remove deterministic ordering.
8. **Chosen remediation.** Replace the custom nested-loop `sortStrings` helper with Go's `sort.Strings` and remove the redundant implementation.
9. **Why this solution was selected.** It is simpler, standard, tested and at least as performant while preserving lexical order exactly.
10. **Rejected alternatives.** Manual optimization has no benefit; retaining bespoke code adds maintenance surface; nondeterministic map iteration would weaken stable output/tests.
11. **Trade-offs.** Adds a standard-library import and no meaningful runtime cost.
12. **Regression tests / protection.** The final Analytical Core audit requires standard-library source sorting and forbids the obsolete helper contract.
13. **Adversarial review findings.** Determinism remains required; the remediation is replacement of the algorithm, not removal of sorting.
14. **Remediation iterations.** The helper was removed in the final closure commit after the provenance contract had stabilized.
15. **Residual risks and limitations.** Sorting complexity is negligible for current provider counts; future very large source sets should be profiled rather than speculatively optimized.
16. **Operational or deployment consequences.** None.
17. **Exact evidence.** Historical implementation commit `8aa8dfa9f0cb0f5eae94497939633f100a863ef8`. Original review ID: `AC-19`. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-PERF-115=CLOSED`.
19. **Prevention / future guard.** Do not introduce handwritten sorting/searching helpers where the Go standard library already provides the required deterministic behavior without measured justification.

### GFA-GOV-116 — Analytical Core closure lacked one permanent strict audit and complete CI path reachability

1. **Finding / symptom.** Before final closure, remediation evidence was spread across Documents 97–101 and tests, but there was no single strict source audit proving all 19 review dispositions and ensuring relevant analytical frontend changes trigger the backend closure audit.
2. **Root cause.** Incremental remediation gates validated individual findings but final review classification, cross-layer ownership and workflow path reachability were not yet encoded as one durable closure contract.
3. **Failure scenario.** A later change regresses a closed analytical invariant or modifies only the frontend analytical API/query/overview path while the strict backend analytical audit is not triggered; the repository can drift away from the claimed closed state.
4. **Impact.** Closure becomes documentary rather than executable, and cross-stack analytical contract regressions can evade the workflow that owns the final source audit.
5. **Severity rationale.** **P2 retrospective.** This is a governance/verification gap around a large correctness review; it does not itself change metric values but weakens durable closure evidence.
6. **Existing guarantees violated.** A closed review must have one permanent machine-checkable acceptance boundary; CI path filters must reach every source surface whose change can violate that boundary.
7. **Considered solutions.** Rely on ordinary unit tests and docs; add a manual checklist; create `analyticalcorefinalaudit` and wire it to Backend CI with relevant frontend paths; move all analytical checks exclusively to Frontend CI.
8. **Chosen remediation.** Add `apps/api/tools/analyticalcorefinalaudit`, encode every AC disposition and critical source contract, run it in strict Backend CI, and extend backend workflow path filters to analytical frontend components/API/query files while Frontend CI continues lint/type/build for `apps/web/**`.
9. **Why this solution was selected.** One executable audit binds closure claims, source ownership and cross-stack workflow reachability without duplicating full backend logic in frontend tooling.
10. **Rejected alternatives.** Documentation/manual review is non-executable; unit tests alone do not prove classification/reachability; frontend-only ownership cannot inspect backend analytical architecture.
11. **Trade-offs.** The audit adds maintenance work when intentional contracts move and must avoid becoming a brittle string-matching proxy for semantics.
12. **Regression tests / protection.** Strict audit unit tests, Backend CI invocation, path-filter coverage, Frontend CI type/lint/build and required closure jobs collectively protect the boundary.
13. **Adversarial review findings.** CI reachability is part of the contract: a perfect audit is ineffective if a relevant file can change without triggering it. The next finding separately addresses audit formatting brittleness.
14. **Remediation iterations.** Final closure introduced the audit and expanded workflow paths; Section 13 then hardened its matching rules against legal formatter line breaks.
15. **Residual risks and limitations.** Source audits complement rather than replace behavioral tests. New analytical surfaces must be added explicitly to audit rules and workflow triggers.
16. **Operational or deployment consequences.** Pull requests touching analytical frontend ownership can trigger backend closure verification in addition to normal frontend CI.
17. **Exact evidence.** Historical closure implementation commit `8aa8dfa9f0cb0f5eae94497939633f100a863ef8`; the commit adds the strict audit, Backend CI invocation and analytical frontend path filters. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-GOV-116=CLOSED`.
19. **Prevention / future guard.** New Analytical Core source surfaces must be added to both the permanent audit and the workflow path reachability set before they can participate in a closed production contract.

### GFA-TEST-117 — the permanent Analytical Core source audit was over-sensitive to legal formatter layout

1. **Finding / symptom.** Initial source-contract checks could depend on one exact textual layout even though Go and TypeScript formatters may legally split selector expressions or function arguments across lines.
2. **Root cause.** Some audit rules matched raw source strings rather than the minimal normalized representation needed to verify semantic ownership.
3. **Failure scenario.** `gofmt`/Prettier-equivalent source formatting changes line breaks without changing behavior; the strict closure audit reports failure. Conversely, broad whitespace stripping could accidentally hide a genuinely missing contract.
4. **Impact.** The permanent gate can produce false failures, become costly to maintain and incentivize weakening or bypassing the audit.
5. **Severity rationale.** **P2 retrospective.** A brittle required closure gate undermines reliability of repository acceptance even though runtime code remains correct.
6. **Existing guarantees violated.** Source audits should validate semantic ownership and required contracts, not arbitrary formatter layout; normalization must be narrow enough to still detect absent behavior.
7. **Considered solutions.** Match exact formatter output; remove source-level checks; parse full Go/TypeScript ASTs for every rule; use explicitly declared compact checks that normalize formatting whitespace only where necessary.
8. **Chosen remediation.** The audit distinguishes normal required source text from `RequiredCompact` contracts, removes formatting whitespace only for those checks, verifies thresholds in their actual owners and tests multiline-valid plus genuinely-absent cases.
9. **Why this solution was selected.** It reduces false failures without adding a large cross-language parser framework or weakening every source check.
10. **Rejected alternatives.** Exact-layout matching is brittle; deleting checks loses closure protection; full AST parsing across Go/TypeScript is disproportionate for the small set of formatting-sensitive contracts.
11. **Trade-offs.** Compact source matching remains a source-level heuristic and requires deliberate declaration for formatting-sensitive rules.
12. **Regression tests / protection.** Audit tests prove valid multiline Go/TypeScript contracts pass and missing compact contracts fail; sampling/stale/radius checks point to canonical source owners.
13. **Adversarial review findings.** Whitespace normalization must not be global; otherwise distinct tokens or missing semantics can collapse into false matches. Only declared compact rules receive this normalization.
14. **Remediation iterations.** The final audit was hardened in the same closure increment after valid formatter layouts exposed overly literal source assumptions.
15. **Residual risks and limitations.** Future complex semantic checks may deserve AST-based validation if compact matching becomes difficult to reason about.
16. **Operational or deployment consequences.** CI becomes resilient to ordinary formatter output while still failing real contract removal.
17. **Exact evidence.** Historical implementation commit `8aa8dfa9f0cb0f5eae94497939633f100a863ef8`; Document 102 Section 13 and audit regression tests. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-TEST-117=CLOSED`.
19. **Prevention / future guard.** New source-audit rules must prefer semantic/structural checks; formatting-normalized matching may be used only with an explicit compact contract and both positive multiline and negative-absence tests.

### Non-defect Analytical Core review dispositions

No canonical GFA finding IDs are created for the following observations because the final review evidence classifies them as retained/rejected contracts rather than engineering defects:

```text
AC-05 = DELIBERATELY_RETAINED — compatibility calculator/registry are not production runtime architecture;
AC-13 = REJECTED_NON_BLOCKING — universal function-length threshold;
AC-14 = DELIBERATELY_RETAINED — Value + HasValue is an intentional presence contract;
AC-15 = DELIBERATELY_RETAINED — nullable sections are status-discriminated evidence;
AC-16 = REJECTED_NON_BLOCKING — mechanical And/With naming ban.
```

This preserves the original 19-item review classification without inflating the canonical defect count.

### Post-closure frontend security ownership

Section 14 is preserved as a chronology pointer only. Commit `48f274754fa0fbdbe4ed0a2b8f95985f38183629` created `docs/103_NEXT_16_2_11_SECURITY_CLOSURE.md` as the dedicated owner of the later Next.js/PostCSS vulnerability remediation. That security defect will receive its canonical GFA finding ID with Document 103; it is not duplicated under Document 102.