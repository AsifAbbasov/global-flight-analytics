# Document 103 — Next.js 16.2.11 Security Closure

Status: IMPLEMENTED
Baseline commit: 8aa8dfa9f0cb0f5eae94497939633f100a863ef8
Scope: frontend framework, PostCSS security, and cross-workflow evidence

## 1. Trigger

Frontend Continuous Integration run 30128888127 failed at:

```text
pnpm audit --prod --audit-level moderate
```

Dependency policy tests passed before the audit, proving that the
repository policy still encoded the previous Next.js 16.2.9 baseline.

## 2. Security update

The frontend now pins:

```text
next = 16.2.11
eslint-config-next = 16.2.11
```

Next.js 16.2.11 is the patched release for the July 21, 2026 advisory
set affecting the 16.2 line below 16.2.11.

PostCSS 8.5.18 is the minimum accepted release after the high-severity
path traversal advisory affecting versions through 8.5.17.

## 3. Permanent controls

```text
apps/web/package.json pins the patched framework and lint configuration;
pnpm-workspace.yaml upgrades every PostCSS release below 8.5.18;
pnpm-lock.yaml resolves the exact patched dependency graph;
verify-frontend-dependency-security.mjs rejects older framework pins;
analyticalcorefinalaudit rejects regression to 16.2.9;
pnpm audit blocks moderate or more severe production findings;
Backend Continuous Integration and Frontend Continuous Integration run
on the same security commit.
```

## 4. Completion contract

```text
NEXT_SECURITY_BASELINE=16.2.11
POSTCSS_SECURITY_BASELINE=8.5.18
FRONTEND_DEPENDENCY_POLICY=PASS
FRONTEND_PRODUCTION_AUDIT=PASS
FRONTEND_TYPECHECK=PASS
FRONTEND_LINT=PASS
FRONTEND_BUILD=PASS
ANALYTICAL_CORE_FINAL_SOURCE_AUDIT=PASS
```
