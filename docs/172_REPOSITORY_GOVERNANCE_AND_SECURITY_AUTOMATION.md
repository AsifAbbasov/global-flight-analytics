# Repository Governance and Security Automation

<!-- REPOSITORY-GOVERNANCE-SECURITY-AUTOMATION-V1 -->

Status: CANONICAL RECONCILIATION COMPLETE — four findings CLOSED; external security-settings verification remains IN_PROGRESS  
Date: 2026-08-03  
Reviewed baseline: `72fc2be736ec4de386a055d826592fa832dcc0fc`  
Repository hardening commit: `5c1c0862581842a78c323f5581c1425641b2b363`  
Exact historical Backend CI: `30799569901` — SUCCESS  
Exact historical Frontend CI: `30799569378` — SUCCESS  
Exact historical CodeQL: `30799569096` — SUCCESS

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

The repository hardening commit passed Backend CI, Frontend CI and CodeQL on the exact
commit SHA. Current repository evidence independently confirms that ruleset `20286514`,
`Global Flight Analytics main protection`, is active for the default branch and enforces:

- branch deletion protection;
- non-fast-forward protection;
- required linear history;
- pull-request-only integration with review-thread resolution;
- squash-only merge;
- strict required `Backend CI Gate`, `Frontend CI Gate`, and `CodeQL Security Gate` checks;
- CodeQL code-scanning enforcement for high-or-higher security alerts.

Current repository metadata also confirms merge commits and rebase merges are disabled,
squash merge is enabled, merged branches are deleted automatically, and branch updates are
allowed.

The source-owned settings verifier additionally checks selected GitHub Actions policy,
required SHA pinning, read-only default workflow permissions, Dependabot alerts API access,
and Secret Scanning alerts API access. Those external settings endpoints are not readable
through the connector used for this retrospective reconciliation. Their current state is
therefore not represented as independently re-verified here.

## Branch reconciliation

The fully merged `feature/active-aircraft-metric` branch had no unique commits and was a
cleanup candidate after verification.

The `docs/open-aviation-metrics-positioning` branch contained two unique commits for an
abandoned Open Aviation Metrics API product pivot. The documented preservation policy was
to retain its head as annotated tag
`archive/open-aviation-metrics-positioning-2026-08-03` before deleting the active branch.

Branch cleanup is closure evidence and repository hygiene. It is not assigned a separate
canonical finding because the material governance failure modes are owned by the protection,
CI, security-automation, and ownership findings below.

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

## Current reconciliation state

```text
REPOSITORY_GOVERNANCE_PATCH=PASS
BACKEND_CI=PASS
FRONTEND_CI=PASS
CODEQL_WORKFLOW=PASS
SOURCE_ACTION_SHA_PINNING=PASS
UNCONDITIONAL_PULL_REQUEST_CI=PASS
BACKEND_CI_GATE=PASS
FRONTEND_CI_GATE=PASS
CODEQL_SECURITY_GATE=PASS
MAIN_RULESET=ACTIVE
MAIN_RULESET_ID=20286514
MAIN_RULESET_REQUIRED_CHECKS=PASS
MAIN_RULESET_CODE_SCANNING=PASS
CODEOWNERS=PASS
SECURITY_POLICY=PASS
SELECTED_ACTIONS_EXTERNAL_POLICY=CURRENT_STATE_UNVERIFIED
DEPENDABOT_ALERTS_EXTERNAL_SETTING=CURRENT_STATE_UNVERIFIED
SECRET_SCANNING_EXTERNAL_SETTING=CURRENT_STATE_UNVERIFIED
SECRET_PUSH_PROTECTION_EXTERNAL_SETTING=CURRENT_STATE_UNVERIFIED
STALE_BRANCH_RECONCILIATION=CLOSURE_EVIDENCE_NOT_SEPARATE_FINDING
```

---

## Canonical remediation record — GFA-GOV-442

### 1. Finding / symptom

The default branch had no active repository ruleset or legacy branch protection, so direct
pushes, force pushes, and branch deletion were not blocked by GitHub.

### 2. Root cause

Repository governance existed only as workflow/source conventions. No GitHub-hosted branch
rule made pull-request integration, immutable history, required checks, or deletion
protection mandatory for `main`.

### 3. Failure scenario

A privileged push could bypass pull-request CI entirely, rewrite or delete protected history,
or land a revision that had not passed the repository's required quality/security gates.

### 4. Impact

The repository could not treat `main` as an independently enforced release-quality boundary.
A governance bypass could invalidate CI evidence, release provenance, and history integrity.

### 5. Severity rationale

**P1 retrospective.** The defect allowed complete bypass of the repository's intended
integration and history-integrity controls. No historical severity label is claimed; the
classification is reconstructed from the documented failure mode.

### 6. Existing guarantees violated

- default-branch changes must pass through pull requests;
- required CI/security gates must be enforced independently of contributor behavior;
- protected history must reject force pushes and deletion;
- merge strategy must preserve linear, auditable history.

### 7. Considered solutions

- rely on contributor discipline and workflow conventions;
- use legacy branch protection only;
- install a repository ruleset covering the default branch with required checks and history
  protections;
- require an approving second reviewer even though the repository is solo-owned.

### 8. Chosen remediation

Create and activate the `Global Flight Analytics main protection` ruleset for the default
branch, require pull requests and resolution of review conversations, allow squash merge
only, require linear history, block deletion/non-fast-forward updates, and require the three
stable CI/security gate contexts.

### 9. Why this solution was selected

A repository ruleset makes the acceptance boundary server-enforced and reviewable instead
of depending on local process. Zero mandatory external approvals preserves the solo-owner
workflow while still requiring the PR/gate path.

### 10. Rejected alternatives

- contributor discipline was rejected because it is not an enforcement mechanism;
- unprotected direct pushes were rejected because they bypass evidence;
- mandatory approval from another person was rejected because it would deadlock the
  documented solo-owner repository model.

### 11. Trade-offs

Every default-branch change must satisfy the protected PR path, so emergency changes carry
more ceremony. The ruleset intentionally protects process integrity rather than guaranteeing
that every test catches every defect.

### 12. Regression tests / protection

`scripts/verify-repository-governance-settings.sh` verifies the active ruleset, required
rule types and required check contexts. Repository governance contract tests are part of the
release verification path.

### 13. Adversarial review findings

The review distinguished server-side enforcement from source-controlled workflow intent and
kept the solo-owner approval count at zero while requiring PRs, thread resolution, strict
checks, and linear history.

### 14. Remediation iterations

The source patch landed in `5c1c0862581842a78c323f5581c1425641b2b363` after baseline
`72fc2be736ec4de386a055d826592fa832dcc0fc`. The external ruleset is recorded as active with
ID `20286514`, created on 2026-08-03.

### 15. Residual risks and limitations

Repository administrators and hosting-platform behavior remain outside source control. A
ruleset cannot prove the semantic correctness of a change; it only enforces the required
acceptance path.

### 16. Operational or deployment consequences

Default-branch changes require pull requests and the configured status checks. Squash is the
only allowed merge method; force pushes and deletion are blocked.

### 17. Exact evidence

- baseline: `72fc2be736ec4de386a055d826592fa832dcc0fc`;
- repository hardening: `5c1c0862581842a78c323f5581c1425641b2b363`;
- Backend CI `30799569901` — SUCCESS;
- Frontend CI `30799569378` — SUCCESS;
- CodeQL `30799569096` — SUCCESS;
- active ruleset ID `20286514`, name `Global Flight Analytics main protection`;
- current ruleset requires `Backend CI Gate`, `Frontend CI Gate`, and `CodeQL Security Gate`.

### 18. Final canonical status

**CLOSED.**

### 19. Prevention / future guard

The active GitHub ruleset and `verify-repository-governance-settings.sh` provide independent
hosting-side and source-owned protection against regression of the default-branch governance
contract.

---

## Canonical remediation record — GFA-SEC-443

### 1. Finding / symptom

External GitHub Actions in repository workflows were referenced through mutable major-version
tags rather than immutable commit SHAs.

### 2. Root cause

Workflow maintenance optimized for readable major-version references such as `@v7` without
cryptographically fixing the exact action implementation consumed by CI.

### 3. Failure scenario

An upstream action tag could move to different code without any repository commit. A future
CI run could therefore execute third-party workflow code that had not been reviewed as part
of this repository's history.

### 4. Impact

Mutable action resolution weakened CI supply-chain reproducibility and could expose
repository tokens, build inputs, or release evidence to an upstream action compromise or
unexpected tag movement.

### 5. Severity rationale

**P1 retrospective.** CI executes external code with repository-scoped capabilities, so an
unreviewed implementation change at a mutable tag is a material supply-chain risk. No
historical severity label is claimed.

### 6. Existing guarantees violated

- external CI code must be reviewable from repository history;
- workflow execution must be reproducible against an exact action revision;
- release evidence must not change because a remote mutable tag moved.

### 7. Considered solutions

- keep major tags for maintenance convenience;
- pin only high-risk actions;
- pin every external action to a full immutable SHA while retaining a readable version
  comment;
- vendor action implementations into the repository.

### 8. Chosen remediation

Replace external action major tags with full commit SHAs and keep human-readable version
comments beside each pin. Governance verification rejects regression to mutable refs.

### 9. Why this solution was selected

Full SHA pins preserve ordinary GitHub Actions ergonomics while making the executed external
code revision explicit in repository history.

### 10. Rejected alternatives

- mutable major tags were rejected because they are not immutable evidence;
- selective pinning was rejected because it leaves inconsistent supply-chain policy;
- vendoring was rejected as unnecessary maintenance overhead for this repository.

### 11. Trade-offs

Action upgrades now require an explicit repository change. Version comments must remain
aligned with the pinned SHA for readability.

### 12. Regression tests / protection

Repository governance tests inspect workflow references. The settings verifier also contains
a contract for GitHub's selected-actions/SHA-pinning policy when those external endpoints are
available to the executing owner environment.

### 13. Adversarial review findings

The review treated workflow actions as executable supply-chain dependencies rather than
mere YAML syntax and required the exact external implementation to be reviewable.

### 14. Remediation iterations

Immutable pins were introduced in
`5c1c0862581842a78c323f5581c1425641b2b363` together with the permanent governance tests.

### 15. Residual risks and limitations

A pinned action can still contain a vulnerability; pinning prevents silent upstream
replacement but does not replace dependency review. Current external selected-actions policy
cannot be independently read through the reconciliation connector and is owned by
`GFA-SEC-445`'s residual external-settings boundary.

### 16. Operational or deployment consequences

Routine action upgrades require explicit SHA updates and normal CI verification.

### 17. Exact evidence

- baseline: `72fc2be736ec4de386a055d826592fa832dcc0fc`;
- repository hardening: `5c1c0862581842a78c323f5581c1425641b2b363`;
- Backend CI `30799569901` — SUCCESS;
- Frontend CI `30799569378` — SUCCESS;
- CodeQL `30799569096` — SUCCESS;
- hardening diff replaces mutable action refs with full immutable SHAs.

### 18. Final canonical status

**CLOSED.**

### 19. Prevention / future guard

Permanent governance contract verification rejects mutable external action references; action
updates remain explicit reviewed commits.

---

## Canonical remediation record — GFA-GOV-444

### 1. Finding / symptom

Backend and Frontend CI pull-request workflows used path filters, so some default-branch
changes could avoid one or both stable required CI contexts.

### 2. Root cause

Workflow optimization attempted to skip apparently unrelated jobs based on changed paths,
but the repository later used stable status contexts as governance requirements across the
whole default-branch acceptance boundary.

### 3. Failure scenario

A pull request could modify governance, shared configuration, release evidence, or another
cross-cutting file outside a workflow's path list and never produce a required backend or
frontend gate result.

### 4. Impact

The repository could not safely require stable CI contexts at the hosting layer because
workflow reachability depended on file selection. Changes could bypass intended validation
or leave required-check state inconsistent.

### 5. Severity rationale

**P1 retrospective.** The defect affected the integrity of the repository-wide acceptance
gate by making required validation conditional on path classification. No historical
severity label is claimed.

### 6. Existing guarantees violated

- every pull request targeting the default branch must produce the stable required gate
  contexts;
- cross-cutting changes must not be able to bypass release verification through path-list
  omissions;
- hosting-side required checks must have deterministic workflow reachability.

### 7. Considered solutions

- continually expand path lists;
- require only one aggregate workflow;
- run Backend and Frontend CI unconditionally for default-branch pull requests while keeping
  selective push visibility where useful;
- remove hosting-side required checks.

### 8. Chosen remediation

Remove pull-request path filters from Backend and Frontend CI, add stable Backend/Frontend
gate jobs, and update repository verifiers to recognize unconditional PR execution while
preserving explicit push paths.

### 9. Why this solution was selected

Unconditional pull-request reachability makes required status contexts deterministic and
eliminates correctness dependence on maintaining exhaustive path lists.

### 10. Rejected alternatives

- ever-growing path lists were rejected because omissions are inevitable in a cross-cutting
  monorepo;
- removing required checks was rejected because it weakens governance;
- one aggregate workflow was unnecessary because stable domain-specific gates already exist.

### 11. Trade-offs

Documentation-only or narrowly scoped pull requests may run more CI than strictly necessary,
trading compute time for deterministic repository protection.

### 12. Regression tests / protection

Governance source tests assert unconditional PR triggers and stable gate jobs. The active
ruleset requires the resulting `Backend CI Gate` and `Frontend CI Gate` contexts.

### 13. Adversarial review findings

The review recognized that path-filter optimization conflicts with strict required-check
semantics when cross-cutting files can affect both application halves.

### 14. Remediation iterations

The unconditional PR triggers and stable gate jobs landed in
`5c1c0862581842a78c323f5581c1425641b2b363`.

### 15. Residual risks and limitations

Unconditional execution increases CI consumption and still depends on each workflow's own
internal test coverage being meaningful.

### 16. Operational or deployment consequences

Every pull request to `main` now receives deterministic backend and frontend gate contexts,
which are required by the active ruleset.

### 17. Exact evidence

- baseline: `72fc2be736ec4de386a055d826592fa832dcc0fc`;
- hardening: `5c1c0862581842a78c323f5581c1425641b2b363`;
- Backend CI `30799569901` — SUCCESS;
- Frontend CI `30799569378` — SUCCESS;
- current ruleset ID `20286514` requires `Backend CI Gate` and `Frontend CI Gate`.

### 18. Final canonical status

**CLOSED.**

### 19. Prevention / future guard

The active required-check ruleset plus governance source tests prevent silent reintroduction
of PR path-filter gaps without failing repository governance verification.

---

## Canonical remediation record — GFA-SEC-445

### 1. Finding / symptom

Repository security automation and security settings were not comprehensively enforced:
CodeQL had no analysis, the audit recorded Secret Scanning and Dependabot alerts as disabled,
and GitHub Actions policy allowed broader third-party execution than the intended hardened
boundary.

### 2. Root cause

Security controls were treated as optional repository-host settings rather than one
source-described, verifiable governance contract with stable security-analysis gates and an
idempotent settings configuration path.

### 3. Failure scenario

Security-relevant code changes could merge without CodeQL analysis, vulnerable dependencies
or committed secrets could lack the intended GitHub-hosted detection surfaces, and Actions
execution policy could remain broader than the repository's reviewed external-action set.

### 4. Impact

The repository had weaker prevention/detection against code vulnerabilities, dependency
risk, credential exposure, and CI supply-chain drift than its production/release claims
required.

### 5. Severity rationale

**P1 retrospective.** Multiple repository-level security detection/enforcement controls were
absent or disabled. The current reconciliation can prove the CodeQL/ruleset portion but not
independently re-read every external security-settings endpoint, so the canonical status
remains IN_PROGRESS rather than overstating closure.

### 6. Existing guarantees violated

- security analysis must be part of the protected default-branch acceptance path;
- repository security settings must be reproducibly configured and independently verifiable;
- workflow permissions and allowed external Actions must follow least privilege;
- credential/dependency detection must not exist only as undocumented owner settings.

### 7. Considered solutions

- leave security settings as manual account knowledge;
- add CodeQL only;
- add source-owned CodeQL plus idempotent GitHub settings configure/verify scripts covering
  Actions permissions, Dependabot, Secret Scanning, and ruleset state;
- require a paid external security platform.

### 8. Chosen remediation

Add CodeQL security-extended analysis for Go and JavaScript/TypeScript, add a stable
`CodeQL Security Gate`, add governance configuration and settings-verification scripts, and
make the active default-branch ruleset require the CodeQL gate and high-or-higher CodeQL
security policy.

The source-owned verifier also checks selected Actions policy, SHA pinning, read-only default
workflow permissions, Dependabot alerts API access, and Secret Scanning alerts API access.

### 9. Why this solution was selected

It keeps the security boundary reproducible in repository source while using GitHub-native
controls available to the project instead of introducing paid infrastructure.

### 10. Rejected alternatives

- manual undocumented settings were rejected because they cannot be audited from source;
- CodeQL without a required gate was rejected because analysis could become advisory only;
- paid external security tooling was unnecessary for the documented project constraints.

### 11. Trade-offs

Some controls live in GitHub account/repository settings rather than Git history. Exact
current verification therefore requires an authenticated owner context capable of reading
those settings endpoints.

### 12. Regression tests / protection

The CodeQL workflow and security gate are source-controlled. The active ruleset requires
`CodeQL Security Gate` and code-scanning enforcement. `scripts/verify-repository-governance-settings.sh`
checks the remaining GitHub-hosted settings when run with an authenticated owner token.

### 13. Adversarial review findings

The reconciliation deliberately refuses to infer current Dependabot, Secret Scanning,
secret push-protection, or selected-Actions settings from source intent. Successful source
CI and an active ruleset prove only the controls they actually expose.

### 14. Remediation iterations

The source security/governance patch landed in
`5c1c0862581842a78c323f5581c1425641b2b363`; the active ruleset was created later on
2026-08-03. Current reconciliation independently re-read the ruleset but could not read the
sensitive settings endpoint families through the available connector.

### 15. Residual risks and limitations

Current state of selected Actions policy, default workflow-token policy, Dependabot alerts,
Secret Scanning, and secret push protection is **not independently verified in this
reconciliation**. Their configuration may in fact remain correct, but evidence honesty
forbids converting source intent into a current PASS claim.

### 16. Operational or deployment consequences

Before representing full repository security-settings closure as current fact, an owner must
run the settings verifier against the intended exact revision using authenticated GitHub CLI
access and retain the resulting evidence outside secrets-bearing logs.

### 17. Exact evidence

- baseline: `72fc2be736ec4de386a055d826592fa832dcc0fc`;
- source security/governance patch: `5c1c0862581842a78c323f5581c1425641b2b363`;
- Backend CI `30799569901` — SUCCESS;
- Frontend CI `30799569378` — SUCCESS;
- CodeQL `30799569096` — SUCCESS;
- active ruleset ID `20286514` requires `CodeQL Security Gate` and high-or-higher CodeQL
  code-scanning policy;
- current `scripts/verify-repository-governance-settings.sh` checks Actions permissions,
  selected Actions, CodeQL analysis, Dependabot alerts API, Secret Scanning alerts API, and
  the active ruleset;
- direct current reads for the sensitive security-settings endpoint families are unavailable
  through the reconciliation connector.

### 18. Final canonical status

**IN_PROGRESS.** CodeQL workflow/gate and ruleset enforcement are proven; full current
external security-settings verification remains open.

### 19. Prevention / future guard

Run `scripts/verify-repository-governance-settings.sh` with `EXPECTED_GOVERNANCE_SHA` set to
the intended exact revision whenever repository security settings are changed or when full
current closure is asserted.

---

## Canonical remediation record — GFA-GOV-446

### 1. Finding / symptom

The repository had no CODEOWNERS file and no SECURITY reporting policy.

### 2. Root cause

Ownership and vulnerability-reporting expectations existed only implicitly through the solo
maintainer workflow and were not represented as stable repository contracts.

### 3. Failure scenario

A contributor or security reporter could not discover the canonical code owner or safe
vulnerability-reporting path from repository metadata, increasing the chance of unowned
changes or public disclosure of sensitive exploit details.

### 4. Impact

Repository governance and security-response expectations were ambiguous and not reviewable
as part of source history.

### 5. Severity rationale

**P2 retrospective.** The defect weakened governance and coordinated vulnerability reporting
but repository evidence does not establish a direct exploit or production outage caused by
it.

### 6. Existing guarantees violated

- code ownership must be explicit and discoverable;
- vulnerability reports must have a documented non-public handling path;
- governance expectations should be reviewable in repository source.

### 7. Considered solutions

- rely on the repository owner profile;
- document ownership only in README;
- add `.github/CODEOWNERS` and `.github/SECURITY.md`;
- require mandatory code-owner approval despite the solo-owner model.

### 8. Chosen remediation

Add a repository ownership baseline in CODEOWNERS and a SECURITY policy defining supported
version, private reporting expectations, required report context, and project-specific
security scope.

### 9. Why this solution was selected

GitHub-native ownership and security-policy files are discoverable by contributors and
security tooling without imposing a second-human approval dependency on the solo owner.

### 10. Rejected alternatives

- implicit ownership was rejected because it is not a repository contract;
- README-only documentation was rejected because GitHub has dedicated ownership/security
  surfaces;
- mandatory second-person approval was rejected as incompatible with the solo-owner model.

### 11. Trade-offs

CODEOWNERS records responsibility but the active ruleset intentionally does not require a
separate code-owner approval. The SECURITY policy describes reporting process rather than
providing a security guarantee.

### 12. Regression tests / protection

Governance contract tests verify the presence and expected repository governance/security
files as part of release validation.

### 13. Adversarial review findings

The review separated discoverable ownership/reporting policy from mandatory approval. This
preserves explicit governance without creating an impossible approval requirement.

### 14. Remediation iterations

CODEOWNERS and SECURITY policy files were introduced in
`5c1c0862581842a78c323f5581c1425641b2b363`.

### 15. Residual risks and limitations

A documented reporting path still depends on the owner responding appropriately. CODEOWNERS
does not replace required CI/security checks.

### 16. Operational or deployment consequences

Contributors and security reporters now have repository-native ownership and disclosure
guidance; no runtime deployment behavior changes.

### 17. Exact evidence

- baseline: `72fc2be736ec4de386a055d826592fa832dcc0fc`;
- hardening: `5c1c0862581842a78c323f5581c1425641b2b363`;
- Backend CI `30799569901` — SUCCESS;
- Frontend CI `30799569378` — SUCCESS;
- CodeQL `30799569096` — SUCCESS;
- hardening diff adds `.github/CODEOWNERS` and `.github/SECURITY.md`.

### 18. Final canonical status

**CLOSED.**

### 19. Prevention / future guard

Permanent governance tests retain CODEOWNERS and SECURITY policy as source-controlled
repository contracts.
