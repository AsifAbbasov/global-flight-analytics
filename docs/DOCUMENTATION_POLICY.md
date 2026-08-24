# Documentation Policy — Engineering Findings

Status: Documentation Governance v1.0

## Goal

Documentation must preserve engineering decisions, not only final outcomes.

A future engineer must understand not only what changed, but why the decision was made.

## Required remediation structure

Every remediation stage should include:

## Finding / Symptom

What was observed.

## Root Cause

Why the problem existed.

## Failure Scenario

How the problem could appear in production.

## Impact

Effect on correctness, security, reliability, availability, performance, or user experience.

## Severity Rationale

Why the priority was selected.

## Existing Guarantees Violated

Which architectural contract was broken.

## Considered Solutions

Options evaluated before implementation.

## Rejected Alternatives

Solutions intentionally not selected and the reason.

## Chosen Remediation

Implemented solution.

## Trade-offs

Benefits gained and costs accepted.

## Regression Protection

Tests, CI checks, or runtime verification preventing recurrence.

## Adversarial Review History

Additional scenarios discovered during review and how they affected the final design.

## Residual Risks

Known limitations that remain intentionally.

## Evidence

Commit, CI, tests, runtime verification, or deployment evidence.

## Canonical Status

Final authoritative state: OPEN, CLOSED, or ACCEPTED_RISK.

## README rule

README documents high-level state only. Detailed engineering history belongs in stage documents.

The canonical status must always come from the finding registry and linked stage document.
