# Dependency Maintenance Closure

<!-- DEPENDENCY-MAINTENANCE-CLOSURE-V1 -->

Status: CLOSED
Date: 2026-08-03
Baseline: `cabcaddd3467cbc37d9b8e335191fc278d138106`

## Scope

This closure consolidates the safe dependency maintenance wave instead of merging stale, isolated Dependabot pull requests.

## Applied dependency wave

- React and React DOM: `19.2.8`, upgraded together to preserve the peer contract;
- Three.js and its TypeScript declarations: `0.185.1`, upgraded together;
- Tailwind CSS and its PostCSS adapter: `4.3.3`, upgraded together;
- Fiber: `2.52.14`;
- GitHub Actions setup-node and setup-go: major version `7`.

## TypeScript boundary

TypeScript remains on major version 5. The proposed 5-to-7 jump is a separate compiler migration and is not mixed with routine dependency maintenance. Dependabot ignores automated TypeScript major version updates until an explicit migration stage supplies compatibility evidence. Security updates are not disabled by this version-update policy.

## Dependabot policy

Related React, Tailwind, Three.js, Fiber and setup-action updates are grouped. The existing Docker update ecosystem is preserved. Invalid label references are removed so Dependabot no longer reports missing-label configuration errors.

## Follow-up reconciliation

A second Dependabot cycle opened new pull requests from the updated policy. Safe patch updates for React types, TanStack Query and the Next.js toolchain are consolidated in Document 171. ESLint 10, Node type definitions 26 and MapLibre 6 remain explicit major migrations and are deferred by policy.

## Permanent verification

The stable `pnpm verify:release` entry point remains unchanged. Its shell implementation executes dependency maintenance tests and the dependency contract verifier before the existing release verification. Backend CI and Frontend CI also execute the dependency maintenance contract directly.

## Formal closure

```text
REACT_RUNTIME_GROUP=PASS
TAILWIND_TOOLCHAIN_GROUP=PASS
THREE_RUNTIME_GROUP=PASS
FIBER_PATCH_UPDATE=PASS
CI_SETUP_ACTIONS_V7=PASS
DEPENDENCY_MAINTENANCE_CI_ENFORCEMENT=PASS
TYPESCRIPT_MAJOR_MIGRATION=DEFERRED_BY_POLICY
DEPENDABOT_GROUPING=PASS
DEPENDABOT_DOCKER_ECOSYSTEM=PRESERVED
DEPENDABOT_LABEL_ERRORS=REMOVED
DEPENDABOT_FOLLOW_UP_RECONCILIATION=PASS
DEPENDENCY_MAINTENANCE_DEBT=CLOSED
```
