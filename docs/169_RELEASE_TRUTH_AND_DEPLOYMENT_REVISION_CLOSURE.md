# Release Truth and Deployment Revision Closure

<!-- RELEASE-TRUTH-DEPLOYMENT-REVISION-V1 -->

Status: CLOSED — canonical remediation history reconciled  
Date: 2026-08-03  
Baseline: `09c9bc9890bbb2a0ed9751783e5b2518e3950270`  
Implementation and permanent release-truth guard: `cabcaddd3467cbc37d9b8e335191fc278d138106`  
Exact historical Backend CI: run `30766520751` — SUCCESS  
Exact historical Frontend CI: run `30766520758` — SUCCESS

## Purpose

This closure removes the assumption that the current repository `HEAD` is automatically the revision served by a mutable production alias.

The defect is canonicalized as `GFA-GOV-441`. The historical deployment SHA remains valid evidence for the dated verification event; the defect was treating that evidence, mutable aliases, intended deployment revision, observed runtime revision, and current repository state as if they were one perpetually identical fact.

## Independent revision facts

```text
SOURCE_CANDIDATE_SHA
HISTORICALLY_VERIFIED_DEPLOYMENT_SHA
INTENDED_RENDER_DEPLOYMENT_SHA
OBSERVED_API_VERSION_SHA
CURRENT_REPOSITORY_HEAD
```

These values may coincide, but no contract assumes that they do.

## Historical evidence

The production verification performed on 2026-08-02 recorded application revision:

```text
6bca02a8ed1487195b165ae9ced3ca687a373666
```

That SHA remains immutable historical evidence for the recorded smoke test. The Render and Vercel production aliases are mutable and may later serve another revision.

## Required verification workflow

1. Freeze and verify the source candidate.
2. Deploy or select the intended Render revision.
3. Copy its full SHA from Render deployment metadata into `DEPLOYED_API_REVISION`.
4. Pass that value explicitly as `EXPECTED_API_REVISION`.
5. Compare it with the API `/api/v1/version` response through the existing smoke scripts.
6. Record a new dated attestation when the production revision changes.

A local `git rev-parse HEAD` may be used only after independently confirming that the same exact commit is the intended deployment. It is never deployment evidence by itself.

## Permanent gates

The release contract rejects:

- wording that presents the historical 2026-08-02 SHA as the perpetually current revision;
- production smoke examples that automatically derive `EXPECTED_API_REVISION` from local `HEAD`;
- omission of this historical-versus-current revision contract from release documentation.

## Formal closure

```text
HISTORICAL_DEPLOYMENT_EVIDENCE=PRESERVED
MUTABLE_ALIAS_CURRENT_SHA_CLAIM=REMOVED
LOCAL_HEAD_DEPLOYMENT_ASSUMPTION=REMOVED
EXPLICIT_DEPLOYMENT_REVISION_INPUT=REQUIRED
RELEASE_TRUTH_CONTRACT=PASS
```

Exact historical closure evidence for `cabcaddd3467cbc37d9b8e335191fc278d138106` includes Backend CI run `30766520751` and Frontend CI run `30766520758`, both successful. Backend CI included Backend Quality job `91546112769`, Backend Race Safety `91546112791`, PostgreSQL 16 Integration `91546112732`, and Backend Container `91546251064`, all successful.

---

## Canonical remediation record — GFA-GOV-441

### 1. Finding / symptom

Release documentation and smoke examples could treat the current local repository `HEAD` as if it were automatically the revision served by mutable production aliases.

### 2. Root cause

Several release surfaces conflated independent revision facts: source candidate, historically verified deployment revision, intended Render deployment revision, observed API revision, and current repository `HEAD`. Smoke examples derived `EXPECTED_API_REVISION` from `git rev-parse HEAD` without independently proving that the same commit was the deployed revision.

### 3. Failure scenario

The repository advances after a successful deployment while Render or Vercel continues serving an older revision. An operator runs the documented production smoke from the newer local checkout; the expected revision is derived from local `HEAD`, so the check is no longer anchored to the deployment selected in the cloud platform. Alternatively, static wording can falsely present an old historically verified SHA as the revision currently served by a mutable alias.

### 4. Impact

Release evidence can become stale or internally contradictory, causing incorrect claims about what code is deployed and weakening the trustworthiness of production verification, rollback decisions, and portfolio/reviewer evidence.

### 5. Severity rationale

**P2 retrospective.** This is a release-governance and operational evidence-integrity defect. It can invalidate deployment assertions but does not itself modify runtime data or establish a P1 service/data-corruption event. No historical severity label was recorded.

### 6. Existing guarantees violated

- deployment evidence must refer to the exact revision actually selected and observed in production;
- historical smoke evidence must remain immutable without being presented as perpetual current state;
- local source state, intended deployment state, and observed runtime state must remain separately attestable.

### 7. Considered solutions

- keep deriving expected revision from local `HEAD`;
- pin one historical SHA as the permanent production revision in docs;
- infer revision from mutable deployment URLs;
- require an explicit deployment revision from provider metadata and independently compare it with the runtime version endpoint.

### 8. Chosen remediation

Separate all revision facts explicitly, require `DEPLOYED_API_REVISION`/`EXPECTED_API_REVISION` to be supplied from the intended Render deployment metadata, compare that value with `/api/v1/version`, preserve the 2026-08-02 SHA only as dated historical evidence, and install release-contract guards against local-HEAD and perpetual-current-revision wording.

### 9. Why this solution was selected

It makes deployment verification source-independent and evidence-based: repository history, provider deployment metadata, and observed runtime version can agree when they should, but none is silently substituted for another.

### 10. Rejected alternatives

- automatic `git rev-parse HEAD` was rejected because repository state can advance independently of deployment;
- a permanently pinned historical SHA was rejected because Render/Vercel aliases are mutable;
- URL reachability alone was rejected because a reachable alias does not prove which revision it serves.

### 11. Trade-offs

Operators must obtain and pass the intended deployment SHA explicitly. This adds one manual evidence step, but it removes an unverifiable assumption and makes later deployment attestations auditable.

### 12. Regression tests / protection

The release contract verifies the presence of the release-truth closure, rejects wording that presents historical deployment evidence as perpetual current state, and rejects smoke examples that automatically derive `EXPECTED_API_REVISION` from local `HEAD`.

### 13. Adversarial review findings

The remediation preserves the previously verified `6bca02a8ed1487195b165ae9ced3ca687a373666` event instead of rewriting history, while explicitly refusing to claim that mutable production aliases still serve it. This separates evidence freshness from historical evidence validity.

### 14. Remediation iterations

The release-truth contract, documentation corrections, smoke-example changes, and permanent release verification guards landed in `cabcaddd3467cbc37d9b8e335191fc278d138106`.

### 15. Residual risks and limitations

Provider metadata can still be copied incorrectly by an operator; the runtime `/api/v1/version` comparison is therefore required as an independent check. A successful revision match also does not prove that every external dependency or dataset is identical to an earlier deployment.

### 16. Operational or deployment consequences

Production verification must record the intended cloud deployment SHA explicitly and create a new dated attestation when the deployed revision changes. Local `HEAD` may be reused only after independently proving equality with the intended deployment revision.

### 17. Exact evidence

- baseline: `09c9bc9890bbb2a0ed9751783e5b2518e3950270`;
- historical production verification SHA: `6bca02a8ed1487195b165ae9ced3ca687a373666`;
- remediation/guard: `cabcaddd3467cbc37d9b8e335191fc278d138106`;
- Backend CI: run `30766520751` — SUCCESS;
- Frontend CI: run `30766520758` — SUCCESS;
- Backend Quality: job `91546112769` — SUCCESS;
- Backend Race Safety: job `91546112791` — SUCCESS;
- PostgreSQL 16 Integration: job `91546112732` — SUCCESS;
- Backend Container: job `91546251064` — SUCCESS.

### 18. Final canonical status

**CLOSED.**

### 19. Prevention / future guard

Release verification must keep source candidate, intended deployment, observed runtime, historical deployment evidence, and current repository HEAD as independent facts. The release contract remains the permanent guard against reintroducing local-HEAD-equals-deployment assumptions.
