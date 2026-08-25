# Document 47 — Stage 14.7 Frontend Dependency Security Remediation

Status: Remediation History v1.1
Project: Global Flight Analytics
Scope: removal and regression prevention for the PostCSS moderate vulnerability

## 1. Root Cause

The frontend dependency graph contained two PostCSS versions:

```text
8.4.31
8.5.15
```

Tailwind CSS already resolved the safe version.

Next.js 16.2.9 declared PostCSS 8.4.31 as a transitive production dependency.
Versions below 8.5.10 are affected by CVE-2026-41305 and
GHSA-qx2v-qp2m-jg93.

## 2. Resolution

The root pnpm workspace configuration contains the targeted rule:

```yaml
overrides:
  'postcss@<8.5.10': 8.5.15
```

This does not replace Next.js, React, Tailwind CSS, or the package manager.

It only redirects vulnerable PostCSS resolutions to the already present safe
version.

The override is stored in `pnpm-workspace.yaml` because the project uses
pnpm 11.

## 3. Lockfile Policy

The committed lockfile must satisfy all conditions:

```text
no PostCSS version below 8.5.10
PostCSS 8.5.15 is present
Next.js 16.2.9 resolves PostCSS 8.5.15
the targeted workspace override is present
```

The repository script:

```text
pnpm run verify:web-dependencies
```

checks these conditions without network access.

Its unit tests run through:

```text
pnpm run test:web-dependency-policy
```

## 4. Live Advisory Audit

Frontend continuous integration runs:

```text
pnpm audit --prod --audit-level moderate
```

The previous `high` threshold allowed moderate findings to pass.

The new threshold blocks moderate, high, and critical production dependency
findings.

## 5. Compatibility Verification

The remediation is accepted only after:

```text
frozen lockfile installation
dependency policy tests
lockfile security verification
production dependency audit
ESLint
TypeScript validation
Next.js production build
backend architecture audit
complete Go build, vet, and tests
production Docker build
```

## 6. Boundaries

The project does not run `pnpm audit fix --force`.

That command may introduce unrelated framework version changes.

The project does not downgrade Next.js.

The project does not suppress or ignore GHSA-qx2v-qp2m-jg93.

The selected version is an explicit, reviewable, reproducible lockfile
resolution.

## 7. Canonical finding record — GFA-SEC-041

### Finding / symptom

The committed production frontend dependency graph included vulnerable PostCSS `8.4.31` through Next.js while a safe PostCSS `8.5.15` already existed elsewhere in the same lockfile. CI used `pnpm audit --prod --audit-level high`, so a moderate production advisory could pass the security gate.

### Root cause

Transitive dependency resolution and the CI severity threshold were governed separately. The framework's transitive constraint retained an older PostCSS while the repository security gate intentionally ignored moderate advisories.

### Failure scenario

A vulnerable moderate-severity transitive production dependency remains in the lockfile indefinitely because builds/tests pass and the audit threshold returns success. The application ships the affected package even though a compatible patched version is available.

### Impact

The production frontend dependency graph contained a known vulnerable version, and the automated gate was configured so the same severity class could recur without blocking merge.

### Severity rationale

**P2 retrospective.** The historical advisory was classified moderate by the source evidence, so the remediation does not inflate it to P1. It is still a real production dependency vulnerability and a CI policy gap.

### Existing guarantees violated

- known production dependency vulnerabilities at the accepted blocking threshold must fail CI;
- the committed lockfile must resolve the patched dependency deterministically;
- remediation must be targeted and reproducible rather than relying on mutable local installs;
- security gates must protect the same severity class that triggered remediation.

### Considered solutions

1. ignore/suppress the moderate advisory;
2. run `pnpm audit fix --force` or upgrade/downgrade framework packages broadly;
3. apply a targeted pnpm override to the already compatible patched PostCSS and lower the production audit threshold to moderate.

### Chosen remediation

The workspace pins vulnerable PostCSS resolutions `<8.5.10` to `8.5.15`, verifies the lockfile offline, and runs production `pnpm audit` at moderate severity in CI.

### Why selected

The safe version was already compatible/present. A targeted override removes the vulnerable transitive node without unrelated framework churn, while the threshold change prevents recurrence of the same class of ignored finding.

### Rejected alternatives

Ignoring/suppressing was rejected because the project had a clean patched resolution. `audit fix --force` was rejected because it can introduce unrelated major/framework changes. Downgrading Next.js was rejected because the vulnerability could be removed without changing the product framework version.

### Trade-offs

A workspace override assumes compatibility between Next.js and the patched PostCSS version and therefore requires frozen-lockfile, lint, typecheck, build, and cross-stack verification. The project must remove/reassess the override if upstream Next.js dependency constraints change.

### Regression tests / protection

`verify:web-dependencies` fails if any PostCSS below 8.5.10 appears, requires 8.5.15 and the targeted override, and verifies Next.js resolves the patched version. Unit tests cover the policy. Live CI runs `pnpm audit --prod --audit-level moderate`.

### Adversarial review findings

The remediation explicitly rejected the tempting broad `--force` repair and recognized that a code fix alone was incomplete while CI still allowed moderate advisories to pass. The gate was therefore tightened at the same time as the lockfile.

### Remediation iterations

The historical fix combined dependency-resolution correction and CI threshold correction in one increment. Later dependency-maintenance work continued to preserve the production dependency policy rather than treating this advisory as a one-off exception.

### Residual risks / limitations

Package-manager audits depend on advisory availability and ecosystem metadata; they cannot prove absence of unknown vulnerabilities. Overrides can also become unnecessary or incompatible after upstream dependency changes and require periodic review.

### Operational / deployment consequences

No application API/schema change. CI becomes stricter and may block future moderate production advisories. Frozen-lockfile installs produce the patched resolution reproducibly.

### Exact evidence

Implementation commit: `4c2e5f5d534721a0c6a0a168d5f196deb590e212` (`fix: remediate frontend dependency vulnerability`). The canonical historical document records CVE-2026-41305 / GHSA-qx2v-qp2m-jg93 and the version boundary. Historical PR/reviewer metadata is not asserted without recoverable evidence.

### Final canonical status

**CLOSED for the recorded PostCSS advisory and moderate-threshold policy gap.**

### Prevention / future guard

Production dependency policy must block the lowest severity class the project has decided requires remediation. Targeted overrides are preferred over broad forced upgrades when compatibility can be proven, and every override must be lockfile-tested, build-verified, and periodically re-evaluated against upstream dependency changes.
