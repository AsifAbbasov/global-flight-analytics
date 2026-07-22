# Code Review Standard

Status: Authoritative Engineering Policy v1.0
Project: Global Flight Analytics
Scope: Pull requests, audits, review findings, remediation decisions, and closure evidence

## 1. Purpose

This document defines how engineering findings must be written, classified, discussed, fixed, rejected, and verified.

The purpose of review is to reduce concrete engineering risk. Review must not convert style preferences, isolated vocabulary, or arbitrary size thresholds into architectural defects.

This standard is authoritative when another document, audit prompt, reviewer checklist, or automated tool uses SOLID, KISS, YAGNI, DRY, Law of Demeter, Occam's Razor, BDUF, APO, naming, function length, nullability, or structure as review criteria.

## 2. Finding Severity

Every finding must use exactly one of these classifications.

### Blocker

A blocker proves a credible risk of one or more of the following:

- incorrect or contradictory persisted state;
- security or authorization failure;
- data loss, silent corruption, or fabricated provenance;
- unsafe migration or concurrency behavior;
- broken public contract without an approved compatibility plan;
- production failure that cannot be safely contained;
- a required verification gate that fails.

A blocker must be fixed before merge unless the change is explicitly abandoned.

### Required change

A required change identifies a material maintainability or correctness risk that is not an immediate production blocker, such as:

- mixed responsibilities that make a safety rule difficult to verify;
- hidden mode switching through flags or sentinel values;
- duplicated logic that can diverge across production paths;
- missing boundary validation or insufficient regression evidence;
- unclear ownership that creates a realistic future defect path.

A required change must be fixed before merge or rejected with written technical evidence and reviewer agreement.

### Suggestion

A suggestion improves clarity, consistency, local design, or future maintainability without proving a material current risk.

Suggestions do not block merge. They may be accepted, deferred, or rejected without classifying the change as unsafe.

### Nit

A nit concerns formatting, wording, local naming preference, or another minor issue that does not affect correctness or maintainability in a meaningful way.

Nits never block merge.

## 3. Mandatory Finding Contract

Every Blocker or Required change finding must contain:

1. **Location** — file, symbol, query, migration, or contract boundary;
2. **Evidence** — the exact behavior observed in code, tests, plans, logs, or runtime output;
3. **Risk** — the concrete failure mode, not only a principle name;
4. **Severity** — why the finding is Blocker or Required change;
5. **Required change** — the smallest acceptable correction or closure condition;
6. **Verification** — the test, audit, plan, or runtime evidence that proves closure.

A finding that only says “violates SOLID”, “too long”, “contains And”, “uses nil”, or “looks overengineered” is incomplete and cannot be classified above Suggestion.

## 4. Evidence and Scope

Every audit or review summary must record:

- the exact commit or diff that was reviewed;
- the files or modules included in scope;
- which commands and runtime checks were executed;
- which checks were not executed;
- whether conclusions are static evidence, executable evidence, integration evidence, or inference;
- the date when time-sensitive dependencies, vulnerabilities, or external contracts were verified.

A score or production-readiness conclusion is valid only for the reviewed commit and declared scope. It must not be presented as a permanent property of the repository.

## 5. Diagnostic Principles

SOLID, KISS, YAGNI, DRY, Law of Demeter, Occam's Razor, BDUF, and APO are diagnostic lenses, not standalone evidence.

A reviewer may use a principle to explain a demonstrated failure mode. The principle name must not replace evidence.

Examples:

- a non-atomic migration is a migration-integrity and concurrency defect; calling it BDUF adds no proof;
- duplicated migration parsers are a DRY concern only because their accepted languages diverge;
- a service boundary is not unnecessary merely because it delegates; the reviewer must show that it adds no policy, validation, normalization, isolation, or testability;
- an abstraction is not overengineering merely because it is unfamiliar; the reviewer must show unnecessary indirection or cost without present value.

## 6. Function Length

Function length is a review signal, not a verdict.

There is no project-wide fifty-line failure threshold.

A long function requires deeper review when it mixes responsibilities such as transaction ownership, validation, mapping, persistence, retry policy, or response construction. A function must be decomposed when the separation improves correctness, independent testing, ownership clarity, or change safety.

A function must not be split solely to reduce line count when the extraction creates artificial indirection, weak names, fragmented control flow, or helpers with no independent responsibility.

## 7. Naming

Words such as `And` and `With` are not globally forbidden.

Names are reviewed for:

- domain intent;
- unambiguous behavior;
- stable abstraction level;
- consistency with the public contract;
- absence of implementation leakage;
- reasonable discoverability.

A specific name may be rejected when it is overloaded or unclear. The finding must explain the ambiguity. The presence of one word is not sufficient evidence.

## 8. Nullability and Optional Values

`nil` is not globally forbidden.

Nullability is acceptable when it represents a real optional state and remains inside an appropriate boundary, including PostgreSQL adapter values, optional timestamps, optional external evidence, and test doubles.

Nullability becomes a material finding when:

- absence is converted into a valid zero value;
- the caller cannot distinguish missing, unavailable, invalid, and observed values;
- a nil dependency can reach runtime unexpectedly;
- nil changes behavior as an undocumented mode;
- an adapter representation leaks into the domain or public contract without explicit semantics.

The review must evaluate meaning and boundary ownership, not syntax alone.

## 9. Flags and Sentinel Values

Boolean flags and sentinel values are not automatically defects.

They require a Required change when they hide materially different operations, SQL statements, authorization paths, transaction behavior, or domain outcomes behind one ambiguous contract.

They may remain when the states are small, local, explicit, validated, and easier to understand than additional types or methods.

## 10. Database and Analytical Review

Database and analytical findings must prioritize:

1. semantic integrity;
2. transactional and concurrency safety;
3. provenance and availability semantics;
4. precision and deterministic conversion policy;
5. query correctness;
6. measured performance evidence;
7. maintainability.

Indexes must not be required only from visual inspection. Performance findings that request an index must include representative `EXPLAIN (ANALYZE, BUFFERS)` evidence or explicitly state that profiling remains pending.

## 11. Rejection and Deferral

A finding is closed only as one of:

- **fixed** — production behavior changed and verification proves closure;
- **not applicable** — the referenced implementation or contract no longer exists;
- **deliberately rejected** — evidence shows that the proposed rule is mechanical, harmful, or not relevant to the project;
- **deferred** — accepted debt with an owner, rationale, risk, and revisit condition.

“Won't fix” without evidence is not a closure classification.

## 12. Review Comment Template

Use this structure for Blocker and Required change findings:

```text
Classification: Blocker | Required change
Location: <file and symbol>
Evidence: <observed behavior>
Risk: <concrete failure mode>
Required change: <smallest acceptable correction>
Verification: <test, audit, query plan, or runtime proof>
```

Suggestions and Nits may be shorter, but must still describe the intended improvement clearly.

## 13. Merge Decision

A pull request may merge when:

- all Blockers are fixed;
- all Required changes are fixed or deliberately rejected with evidence;
- required tests and audits pass;
- unexecuted checks and residual risks are disclosed;
- documentation is updated when the architecture or public contract changes.

Suggestions and Nits do not block merge.

## 14. Final Rule

Review must be strict about demonstrated risk and conservative about unsupported labels.

The project standard is:

```text
Evidence before severity.
Failure mode before principle name.
Responsibility before line count.
Semantics before syntax.
Verification before closure.
```
