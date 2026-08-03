# Repository Governance and Security Automation

<!-- REPOSITORY-GOVERNANCE-SECURITY-AUTOMATION-V1 -->

Status: Repository patch prepared; GitHub settings activation pending exact-commit CI
Date: 2026-08-03
Baseline: `72fc2be736ec4de386a055d826592fa832dcc0fc`

## Audit finding

The public repository had no active ruleset or legacy branch protection. Direct pushes,
force pushes and branch deletion were not blocked by GitHub. Actions accepted mutable
major tags, all third-party actions were allowed, CodeQL had no analysis, secret scanning
was disabled, Dependabot alerts were disabled, CODEOWNERS was absent and two stale remote
branches remained.

## Repository patch

The repository patch establishes:

- full immutable SHA pins for every external GitHub Action with readable release comments;
- Backend CI and Frontend CI on every pull request targeting the default branch;
- stable Backend CI Gate and Frontend CI Gate contexts;
- CodeQL security-extended analysis for Go and JavaScript/TypeScript;
- a stable CodeQL Security Gate context;
- CODEOWNERS and SECURITY policy files;
- permanent governance contract tests in the release gate;
- migrated recruiter, release, and backend-operations verifiers that recognize unconditional pull request CI while preserving explicit push visibility paths;
- idempotent GitHub settings configuration and verification scripts.

## GitHub settings phase

Settings activation occurs only after the patch commit has passed Backend CI, Frontend CI
and CodeQL on the exact commit SHA. The configuration then restricts Actions, enables
security alerts and secret protection, activates the main branch ruleset and reconciles
stale branches.

The solo-owner repository requires pull requests but does not require an approving review
from another person. Review conversations must be resolved. The ruleset allows squash merge
only, requires the three stable gates, enforces CodeQL high-severity security results,
prevents force pushes and deletion, and requires linear history.

## Branch reconciliation

The fully merged `feature/active-aircraft-metric` branch has no unique commits and is
deleted after verification.

The `docs/open-aviation-metrics-positioning` branch contains two unique commits for an
abandoned Open Aviation Metrics API product pivot. Its head is preserved as the annotated
tag `archive/open-aviation-metrics-positioning-2026-08-03` before the active branch is
deleted.

## Installer release quality contract

Every package that changes this repository must be validated as a complete transaction.
A package must not be released after only one discovered assertion or one-line failure is
fixed. Before a package may be provided, its installer must:

- execute positive transformation and validation fixtures;
- execute negative fixtures that prove malformed input and stale contracts are rejected;
- execute a rollback fixture that deliberately fails after mutation and proves restoration;
- prove that the Go source tree is formatted before the patch;
- run `gofmt` after the patch and prove the Go source digest is unchanged;
- run the complete repository release validation in a detached worktree;
- verify the exact changed-file manifest;
- roll the real repository back to the exact baseline after any patch,
  formatting, validation or manifest failure.

## Closure sequence

```text
INSTALLER_POSITIVE_FIXTURES=PASS
INSTALLER_NEGATIVE_FIXTURES=PASS
INSTALLER_ROLLBACK_SELF_TEST=PASS
GOFMT_BEFORE_AND_AFTER=PASS
FULL_REPOSITORY_VALIDATION=PASS
EXACT_BASELINE_GITHUB_ACTIONS_VALIDATION=REQUIRED_BEFORE_PACKAGE_RELEASE
REPOSITORY_GOVERNANCE_PATCH=PASS
BACKEND_CI_GATE=PASS
FRONTEND_CI_GATE=PASS
CODEQL_SECURITY_GATE=PASS
DEPENDABOT_ALERTS=ENABLED
SECRET_SCANNING=ENABLED
SECRET_PUSH_PROTECTION=ENABLED
ACTIONS_SHA_PINNING=ENFORCED
MAIN_RULESET=ACTIVE
STALE_BRANCH_HISTORY=PRESERVED
REPOSITORY_GOVERNANCE_STAGE=READY_FOR_SETTINGS
```
