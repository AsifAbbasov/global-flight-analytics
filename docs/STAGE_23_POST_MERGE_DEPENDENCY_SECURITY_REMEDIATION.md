# Stage 23 Post-Merge Dependency Security Remediation

## Status

`IN_PROGRESS`

This document records a post-merge security gate discovered after PR #171 was squash-merged into canonical `main`.

Canonical Stage 23 merge commit:

```text
9a995d68ebe9a252414ebbffdefb7c47e8001a3d
```

Stage 23 must not be reported as closed until the dependency remediation described here is merged and the resulting canonical `main` passes the required post-merge validation matrix.

## Trigger

The post-merge Frontend CI run on `9a995d68ebe9a252414ebbffdefb7c47e8001a3d` failed at the production dependency audit step. Product compilation had not established a Stage 23 regression; the release gate stopped because newly available package advisories made the previously accepted dependency set no longer acceptable.

The blocking dependency evidence was:

- `next 16.2.12` — two critical advisories; patched release boundary `>=16.3.3`;
- `maplibre-gl 5.24.0` — critical sanitizer/XSS advisory; patched release boundary `>=6.4.1`;
- `sharp 0.35.3` — high-severity advisory; patched release boundary `>=0.35.4`;
- `baseline-browser-mapping` below `2.11.0` — moderate advisory in the resolved graph.

The repository's dynamic `pnpm audit --prod --audit-level moderate` gate therefore behaved correctly and prevented a false Stage 23 closure.

## Consolidated remediation

The isolated branch `security/frontend-dependency-hotfix` consolidates the dependency repair rather than merging three individually incomplete Dependabot PRs.

The generated candidate uses:

```text
next=16.3.4
eslint-config-next=16.3.4
sharp=0.35.4
maplibre-gl=^6.4.1
minimum_safe_next=16.3.3
minimum_safe_sharp=0.35.4
```

The workspace Sharp override is updated to:

```text
sharp@<0.35.4 -> 0.35.4
```

The lockfile is regenerated from the consolidated patched dependency set. The generation workflow also runs the repository dependency-security contract tests and `pnpm audit --prod --audit-level moderate` before it is allowed to commit the generated candidate.

## Compatibility remediation evidence

Validation of the consolidated dependency candidate exposed two additional compatibility/governance defects that were not visible while the normal package-age gate prevented installation:

- MapLibre GL v6 no longer satisfies the previous TypeScript default-import assumption in `traffic-map.tsx`; the consumer now uses the supported namespace import;
- the dependency-maintenance regression test and verifier still pinned the pre-remediation Next.js and `eslint-config-next` target `16.2.12`; both governance contracts now require the hotfix target `16.3.4`.

A one-time validation workflow applied those corrective changes and, with a runner-local `minimumReleaseAge=0` override only for candidate validation, successfully executed the dependency-maintenance tests and verifier, frontend dependency-security tests and verifier, production dependency audit, ESLint, TypeScript validation, frontend tests, and production frontend build before the corrective source commit was created. This evidence proves the compatibility correction itself, but it is deliberately **not** treated as final release evidence because the repository minimum-release-age policy was temporarily bypassed only inside that disposable validation run.

The final exact PR head must therefore repeat the normal repository checks without that override before merge readiness can be claimed.

## Release-age policy boundary

A separate Dependabot MapLibre validation exposed `@maplibre/maplibre-gl-style-spec@26.4.2` as younger than the normal package minimum-release-age window.

The repository minimum-release-age security policy is **not** being relaxed or removed.

The one-time lockfile generator used a generation-only `minimumReleaseAge=0` override so that the consolidated candidate could be produced and audited. This is not release evidence and is not transferable to the PR or to canonical `main`.

Final PR validation must run under the repository's normal supply-chain policy. If the resolved MapLibre transitive dependency is still too young, the PR remains blocked until the normal release-age gate accepts it. No permanent allowlist or policy reduction is introduced for this hotfix.

## Scope

This remediation changes only the frontend dependency/security boundary required to restore the post-merge release gate.

It does not add:

- Redis or Valkey;
- Python;
- a new service;
- a database migration or table;
- paid infrastructure;
- a new provider;
- a Vercel bypass;
- a weakened audit threshold;
- a weakened minimum-release-age policy.

## Closure gate

The remediation is not merge-ready until the final exact PR head proves all required repository checks and Vercel on the same SHA.

After an authorized merge, canonical `main` must be checked again. Stage 23 may be reported `CLOSED` only when its required post-merge evidence is successful.

```text
STAGE_23_MERGE=COMPLETE
STAGE_23_MAIN_SHA=9a995d68ebe9a252414ebbffdefb7c47e8001a3d
STAGE_23_POST_MERGE_DEPENDENCY_GATE=REMEDIATION_IN_PROGRESS
DEPENDENCY_HOTFIX_MERGED=NO
STAGE_23_CLOSED=NO
```
