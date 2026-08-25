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

---

## Canonical remediation history

### GFA-SEC-118 — production Next.js baseline remained on 16.2.9 after the July 21, 2026 advisory set

1. **Finding / symptom.** Frontend production audit failed while the application and `eslint-config-next` remained pinned to Next.js 16.2.9, below the documented patched 16.2.11 baseline for the July 21, 2026 advisory set.
2. **Root cause.** Repository dependency-policy tests still encoded the previously accepted 16.2.9 baseline, so policy validation could pass even after the external security baseline had moved.
3. **Failure scenario.** A production build resolves the older 16.2.9 framework after relevant advisories are known; ordinary source tests pass because they verify repository policy rather than current production-audit safety.
4. **Impact.** The public frontend can ship a framework version that the repository's own production dependency audit no longer accepts, leaving known security exposure in a user-facing application surface.
5. **Severity rationale.** **P2 retrospective.** This is a confirmed production dependency-security failure requiring prompt remediation. The source document establishes an affected advisory set and failed moderate-or-higher audit, but does not preserve enough advisory-specific exploitability evidence to reconstruct a P1 classification honestly.
6. **Existing guarantees violated.** Production dependency audit must be green; repository policy must not declare an older vulnerable framework baseline safe; framework and matching lint configuration should move together to one supported patched release.
7. **Considered solutions.** Ignore the audit until a later framework upgrade; suppress individual advisories; patch only transitive packages; move Next.js and `eslint-config-next` to the documented 16.2.11 security baseline and strengthen policy/audit guards.
8. **Chosen remediation.** Pin `next` and `eslint-config-next` to 16.2.11, resolve the exact patched lockfile graph, raise `verify-frontend-dependency-security.mjs` minimum/pinned versions, and make the permanent Analytical Core source audit reject regression to 16.2.9.
9. **Why this solution was selected.** It repairs the framework at the direct dependency owner, keeps framework/lint versions aligned and converts the discovered external security requirement into executable repository policy.
10. **Rejected alternatives.** Audit suppression would hide rather than fix known exposure; partial transitive overrides cannot replace a patched direct framework release; deferral would leave the production audit intentionally failing.
11. **Trade-offs.** A framework patch release can change lockfile/transitive artifacts and therefore requires typecheck, lint and production build verification in addition to the security audit.
12. **Regression tests / protection.** Direct version pins, lockfile resolution, `verify-frontend-dependency-security.mjs`, `analyticalcorefinalaudit`, `pnpm audit --prod --audit-level moderate`, Frontend CI and Backend CI jointly protect the baseline.
13. **Adversarial review findings.** Passing repository policy before the production audit failed demonstrated that policy itself can become stale; the minimum safe version therefore must be encoded in more than one independent acceptance path and the real production audit remains authoritative for new advisories.
14. **Remediation iterations.** The post-Analytical-closure security commit upgraded framework and lint configuration together, updated dependency policy and taught the pre-existing Analytical Core audit to reject the old baseline without reopening analytical semantics.
15. **Residual risks and limitations.** Version pins protect only advisories known to the selected baseline. Future advisories can invalidate 16.2.11 and must be surfaced by recurring production dependency audit rather than assuming this version remains permanently safe.
16. **Operational or deployment consequences.** Frontend builds/deployments consume the updated Next.js dependency graph. No API or analytical formula contract changes are introduced.
17. **Exact evidence.** Historical remediation commit `48f274754fa0fbdbe4ed0a2b8f95985f38183629` (`fix: patch Next.js and PostCSS vulnerabilities`). Trigger evidence recorded by this document: Frontend CI run `30128888127` failed at `pnpm audit --prod --audit-level moderate`. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-SEC-118=CLOSED`.
19. **Prevention / future guard.** Production dependency audit must remain independent of static version-policy tests; a newly disclosed framework advisory that invalidates the pinned baseline requires a new security finding rather than silently rewriting this closure record.

### GFA-SEC-119 — PostCSS releases through 8.5.17 remained accepted after a high-severity path traversal advisory

1. **Finding / symptom.** The workspace override still accepted an older PostCSS security floor even though the documented high-severity path traversal advisory affected releases through 8.5.17.
2. **Root cause.** The existing PostCSS override represented an earlier Stage 14 security baseline and was not automatically advanced when a later advisory raised the minimum safe version.
3. **Failure scenario.** The lockfile resolves a PostCSS version at or below 8.5.17 through a transitive frontend dependency while repository policy continues to treat that graph as acceptable.
4. **Impact.** Build/runtime tooling in the production frontend dependency graph can retain a dependency covered by a high-severity path traversal advisory.
5. **Severity rationale.** **P1 retrospective.** The source document explicitly identifies a high-severity path traversal advisory affecting the accepted version range, making this a materially stronger security failure than ordinary dependency drift.
6. **Existing guarantees violated.** Known high-severity production dependency findings must be removed; transitive dependency overrides must encode the current minimum safe release; the lockfile and policy audit must agree on one patched graph.
7. **Considered solutions.** Rely on whichever PostCSS version transitive resolution chooses; suppress the advisory; bump only one importer; raise the workspace-wide PostCSS floor to 8.5.18 and verify the resolved graph.
8. **Chosen remediation.** Replace the older override with `postcss@<8.5.18: 8.5.18`, update the lockfile, encode `MINIMUM_SAFE_POSTCSS_VERSION`/`PINNED_POSTCSS_VERSION` as 8.5.18 and require the Analytical Core/frontend dependency audits to reject the former override.
9. **Why this solution was selected.** Workspace-level override ownership closes every affected transitive path consistently and prevents one package from retaining the vulnerable range.
10. **Rejected alternatives.** Uncontrolled transitive resolution is non-deterministic with respect to the security floor; suppressing the advisory removes evidence; importer-specific fixes can leave other workspace consumers exposed.
11. **Trade-offs.** Overriding a transitive dependency can expose compatibility issues, so the exact graph must pass frontend typecheck, lint and production build rather than assuming patch compatibility.
12. **Regression tests / protection.** Workspace override, exact lockfile graph, frontend dependency-security verifier, Analytical Core source audit and production `pnpm audit` all require the 8.5.18 floor.
13. **Adversarial review findings.** This is **not** the same finding as historical `GFA-SEC-041`: that earlier closure encoded a previous PostCSS security minimum. A later advisory invalidating the accepted range is a new post-closure security event and must retain separate evidence/history.
14. **Remediation iterations.** The same security commit that patched Next.js advanced the PostCSS workspace override and strengthened both dedicated dependency policy and cross-workflow source auditing.
15. **Residual risks and limitations.** A later PostCSS advisory can again raise the safe floor. Overrides should be revisited from current audit evidence, not treated as a permanent guarantee.
16. **Operational or deployment consequences.** Frontend dependency installation resolves PostCSS 8.5.18 for all versions below the security floor; no application API changes are intended.
17. **Exact evidence.** Historical remediation commit `48f274754fa0fbdbe4ed0a2b8f95985f38183629`; Document 103 identifies the affected range through 8.5.17 and the 8.5.18 minimum accepted release. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-SEC-119=CLOSED`.
19. **Prevention / future guard.** Treat each later advisory that invalidates an already-closed dependency baseline as a new finding with its own evidence; production audit and workspace security-floor tests must stay synchronized.