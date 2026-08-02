# Dependabot Follow-up Reconciliation

<!-- DEPENDABOT-FOLLOW-UP-RECONCILIATION-V1 -->

Status: CLOSED
Date: 2026-08-03
Baseline: `54e0efbe98b000e9f12f787ececc425479dbda50`

## Trigger

After the first dependency maintenance commit, Dependabot regenerated pull requests from the corrected policy. This document reconciles that follow-up cycle before the stage is considered complete.

## Safe follow-up wave

- React type declarations: `@types/react 19.2.18` and `@types/react-dom 19.2.4`;
- TanStack React Query: `5.101.4`;
- Next.js and eslint-config-next: `16.2.12`, upgraded together as one toolchain boundary.

## Explicit major-migration boundary

The following proposals are not routine maintenance and are deferred until dedicated compatibility stages provide migration evidence:

- ESLint 9 to 10;
- Node type definitions 20 to 26;
- MapLibre GL 5 to 6;
- TypeScript 5 to 7.

Dependabot ignores these automated major version updates. Security updates remain enabled.

## Pull request reconciliation

The stale first-wave pull requests 14, 15, 16, 17, 18, 20, 21 and 22 were closed without merge after their safe changes were consolidated in commit `54e0efbe98b000e9f12f787ececc425479dbda50`. Follow-up pull requests 24 through 30 are reconciled by the safe update wave or the explicit major-migration policy.

## Formal closure

```text
REACT_TYPES_PATCH_WAVE=PASS
TANSTACK_QUERY_PATCH_WAVE=PASS
NEXT_TOOLCHAIN_PATCH_WAVE=PASS
NEXT_TOOLCHAIN_GROUP=PASS
ESLINT_10_MIGRATION=DEFERRED_BY_POLICY
NODE_TYPES_26_MIGRATION=DEFERRED_BY_POLICY
MAPLIBRE_6_MIGRATION=DEFERRED_BY_POLICY
TYPESCRIPT_7_MIGRATION=DEFERRED_BY_POLICY
DEPENDABOT_FOLLOW_UP_RECONCILIATION=PASS
DEPENDENCY_MAINTENANCE_DEBT=CLOSED
```
