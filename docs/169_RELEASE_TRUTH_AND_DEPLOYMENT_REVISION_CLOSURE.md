# Release Truth and Deployment Revision Closure

<!-- RELEASE-TRUTH-DEPLOYMENT-REVISION-V1 -->

Status: CLOSED
Date: 2026-08-03
Baseline: `09c9bc9890bbb2a0ed9751783e5b2518e3950270`

## Purpose

This closure removes the assumption that the current repository `HEAD` is automatically the
revision served by a mutable production alias.

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

That SHA remains immutable historical evidence for the recorded smoke test. The Render and
Vercel production aliases are mutable and may later serve another revision.

## Required verification workflow

1. Freeze and verify the source candidate.
2. Deploy or select the intended Render revision.
3. Copy its full SHA from Render deployment metadata into `DEPLOYED_API_REVISION`.
4. Pass that value explicitly as `EXPECTED_API_REVISION`.
5. Compare it with the API `/api/v1/version` response through the existing smoke scripts.
6. Record a new dated attestation when the production revision changes.

A local `git rev-parse HEAD` may be used only after independently confirming that the same
exact commit is the intended deployment. It is never deployment evidence by itself.

## Permanent gates

The release contract rejects:

- wording that presents the historical 2026-08-02 SHA as the perpetually current revision;
- production smoke examples that automatically derive `EXPECTED_API_REVISION` from local
  `HEAD`;
- omission of this historical-versus-current revision contract from release documentation.

## Formal closure

```text
HISTORICAL_DEPLOYMENT_EVIDENCE=PRESERVED
MUTABLE_ALIAS_CURRENT_SHA_CLAIM=REMOVED
LOCAL_HEAD_DEPLOYMENT_ASSUMPTION=REMOVED
EXPLICIT_DEPLOYMENT_REVISION_INPUT=REQUIRED
RELEASE_TRUTH_CONTRACT=PASS
```
