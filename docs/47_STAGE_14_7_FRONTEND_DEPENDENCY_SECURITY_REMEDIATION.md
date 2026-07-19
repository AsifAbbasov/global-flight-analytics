# Document 47 — Stage 14.7 Frontend Dependency Security Remediation

Status: Implementation Baseline v1.0
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
