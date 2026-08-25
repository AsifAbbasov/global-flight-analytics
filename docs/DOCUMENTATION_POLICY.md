# Documentation Policy — Engineering Findings

Status: Documentation Governance v1.1

## Goal

Documentation must preserve engineering decisions, not only final outcomes.

A future engineer must understand not only what changed, but why the decision was made, which alternatives were rejected, what evidence closed the finding, and what guard makes the same class of defect harder to reintroduce.

## Canonical 19-field remediation structure

Every non-trivial engineering finding must record the following nineteen fields. A stage document may group several related findings, but each finding must remain distinguishable in the narrative and in `FINDING_REGISTER.md`.

### 1. Finding / Symptom

What was observed or which contract gap was confirmed.

### 2. Root Cause

Why the problem existed.

### 3. Failure Scenario

A concrete path showing how the defect could appear in production, CI, deployment, or maintenance work.

### 4. Impact

Effect on correctness, security, reliability, availability, performance, user experience, operational evidence, or maintainability.

### 5. Severity Rationale

Why the priority was selected. If an old finding did not historically carry a severity, any later classification must be labeled **retrospective** rather than presented as historical fact.

### 6. Existing Guarantees Violated

Which domain, persistence, architecture, security, lifecycle, deployment, or governance contract was broken.

### 7. Considered Solutions

Material options evaluated before or during remediation.

### 8. Chosen Remediation

The solution that was actually implemented.

### 9. Why This Solution Was Selected

Why the chosen remediation fits the failure mode and project constraints better than the alternatives.

### 10. Rejected Alternatives

Solutions intentionally not selected and the reason each was rejected.

### 11. Trade-offs

Benefits gained and costs, complexity, compatibility changes, or operational burden accepted.

### 12. Regression Tests / Protection

Tests, CI checks, source audits, runtime verification, database constraints, profiling gates, or other controls preventing recurrence.

### 13. Adversarial Review Findings

Additional scenarios found by stronger review, integration execution, failure injection, source audit, or later global review.

### 14. Remediation Iterations

How the solution changed across follow-up fixes or stronger regression guards. Do not collapse a multi-step remediation into a fictional one-shot solution.

### 15. Residual Risks / Limitations

Known boundaries that remain after the remediation and are not implied to be solved.

### 16. Operational / Deployment Consequences

Migration, rollout, compatibility, observability, resource, operator, or recovery implications. If none exist, state that explicitly.

### 17. Exact Evidence

Exact implementation commit(s), relevant PR/CI evidence when recoverable, permanent tests/audits, and any deployment/runtime evidence actually obtained.

### 18. Final Canonical Status

Authoritative state: `OPEN`, `IN_PROGRESS`, `CLOSED`, or `ACCEPTED_RISK`. Finding-level status must not be confused with broader stage/release status.

### 19. Prevention / Future Guard

The general rule, architecture constraint, test ownership, or review requirement that makes the defect class harder to reintroduce in future code.

## Evidence honesty rule

Historical evidence must never be invented to make an old remediation look more complete.

When a historical PR number, reviewer identity, review comment, exact CI run, or original severity cannot be recovered from repository evidence, the canonical document must say so explicitly, for example:

```text
Historical adversarial-review evidence unavailable; reconstruction is limited to repository source, tests, commits, documents, and recoverable CI evidence.
```

A later engineering reconstruction may describe failure modes, alternatives, or retrospective severity, but it must be labeled as reconstruction rather than historical fact.

## Finding versus stage status

A finding can be `CLOSED` while the larger stage remains `REOPENED` or incomplete. Conversely, a historically closed stage does not mean no later residual finding can ever be discovered.

`FINDING_REGISTER.md` owns finding-level status. Stage/release closure documents own their broader status boundaries.

## README rule

README documents high-level state, architecture, major milestones, and navigation only. It must not become the archive for detailed remediation history.

Detailed engineering history belongs in canonical stage/finding documents. README may summarize a finding or milestone and link to its canonical documentation, but the detailed reasoning, alternatives, trade-offs, review chronology, residual risk, and evidence remain outside README.

## Closure rule

A finding is not closed because code changed or a document says `CLOSED`.

Closure requires, at minimum:

```text
implemented remediation
+ regression protection
+ exact recoverable evidence
+ documentation alignment
+ canonical Finding Register status
```

Broader stage/release closure may require stronger integration, production-path, security, performance, frontend, container, or deployment evidence as defined by that stage's own authoritative closure document.
