# CI Required-Check Recovery Hardening

Status: ACTIVE RUNBOOK
Scope: GitHub Actions required-check recovery for pull requests
Primary workflow: `.github/workflows/backend-ci.yml`

## Purpose

This runbook defines a fail-closed recovery path for a required GitHub check that is
missing, cancelled, damaged, or otherwise not usable for the current pull-request
head. Recovery must not weaken branch protection, create empty retrigger commits,
or introduce a second reduced recovery workflow.

The repository already keeps the full Backend CI workflow manually dispatchable and
uses `cancel-in-progress: true`. Recovery therefore reuses the existing full workflow.

## Non-negotiable invariants

1. Every recovery decision is bound to the exact pull-request `head.sha`.
2. A green run for an older SHA never validates a newer head.
3. A queued or in-progress exact-SHA run is not duplicated.
4. A failed or cancelled exact-SHA run is inspected before any fresh dispatch.
5. Empty commits created only to retrigger checks are prohibited.
6. A separate reduced "recovery CI" workflow is prohibited.
7. Branch protection is not bypassed during normal recovery.
8. Manual `workflow_dispatch` uses the existing full `backend-ci.yml`.
9. Recovery evidence records run ID, event, head SHA, status, conclusion and URL.
10. A legitimate code change that changes the head SHA requires verification again.

## Read-only diagnosis

Run:

```bash
scripts/diagnose-ci-required-check-recovery.sh \
  OWNER/REPO \
  PR_NUMBER \
  EXPECTED_HEAD_SHA \
  backend-ci.yml
```

The diagnostic command is intentionally read-only. It verifies the current PR head,
lists workflow runs for the exact SHA, lists exact-SHA check runs and prints one
recommended next action.

If `EXPECTED_HEAD_SHA_MATCH=FAIL`, stop. Do not rerun or dispatch anything until the
new head has been reviewed.

## Recovery decision tree

### 1. Exact-SHA run is queued or in progress

Do not create another run.

```text
NEXT_SAFE_ACTION=WAIT_FOR_EXISTING_EXACT_SHA_RUN
```

### 2. Exact-SHA run failed or was cancelled

Inspect the first real failure. If retrying the same run is appropriate, retry only
the failed jobs first:

```bash
gh run rerun "$RUN_ID" --failed --repo "$REPOSITORY"
```

If the run is structurally damaged and retrying failed jobs cannot recover it, rerun
the full existing run:

```bash
gh run rerun "$RUN_ID" --repo "$REPOSITORY"
```

Do not repeatedly rerun the same damaged run without new evidence.

### 3. No workflow run exists for the exact PR head

Do not create an empty commit.

Confirm the PR head again, then dispatch the existing full Backend CI workflow on the
actual PR head branch:

```bash
HEAD_SHA="$(gh api "repos/$REPOSITORY/pulls/$PR_NUMBER" --jq '.head.sha')"
HEAD_BRANCH="$(gh api "repos/$REPOSITORY/pulls/$PR_NUMBER" --jq '.head.ref')"
test "$HEAD_SHA" = "$EXPECTED_HEAD_SHA"

gh workflow run backend-ci.yml \
  --repo "$REPOSITORY" \
  --ref "$HEAD_BRANCH"
```

After dispatch, verify that the run still targets the expected SHA:

```bash
gh run list \
  --repo "$REPOSITORY" \
  --workflow backend-ci.yml \
  --branch "$HEAD_BRANCH" \
  --event workflow_dispatch \
  --limit 10 \
  --json databaseId,headSha,status,conclusion,url
```

A manual run for the wrong SHA is not recovery evidence.

### 4. A legitimate corrective code change is required

Make the real code change, review it, and push it normally. The new commit becomes a
new head and must receive a new verification suite.

Never create a commit whose only purpose is to alter the SHA and retrigger Actions.

### 5. Exact-SHA Backend CI is green

Verify the PR's required checks still correspond to the exact head before merging.
Branch protection remains authoritative.

## `cancel-in-progress` boundary

The Backend CI concurrency policy intentionally uses `cancel-in-progress: true`.
A newer run for the same workflow/ref may cancel an older run. A cancelled run is not
automatically a product failure, but it is also not valid closure evidence.

Before retrying a cancelled run:

- confirm the PR head did not move;
- confirm whether a newer exact-SHA run replaced it;
- avoid duplicate dispatch while another exact-SHA run is active.

## Administrative exception policy

Administrative branch-protection bypass is not part of the normal recovery path.
If GitHub itself prevents required-check delivery during a verified platform incident,
an owner may decide on an exceptional procedure only with explicit approval and
separate evidence. The default repository procedure remains fail-closed.

## Evidence to retain

For any non-trivial recovery, record:

```text
PR_NUMBER=<number>
EXPECTED_HEAD_SHA=<40-char SHA>
WORKFLOW_RUN_ID=<id>
WORKFLOW_EVENT=<pull_request|workflow_dispatch|...>
WORKFLOW_STATUS=<status>
WORKFLOW_CONCLUSION=<conclusion>
WORKFLOW_URL=<url>
REQUIRED_CHECK_EXACT_SHA=PASS
```

Do not commit raw operational logs solely for this purpose. GitHub Actions execution
history and artifacts are preferred evidence surfaces.

## Acceptance

```text
CI_RECOVERY_EXACT_SHA_POLICY=PASS
CI_RECOVERY_READ_ONLY_DIAGNOSTIC=PASS
CI_RECOVERY_EMPTY_COMMIT_PROHIBITION=PASS
CI_RECOVERY_SINGLE_WORKFLOW_POLICY=PASS
CI_REQUIRED_CHECK_RECOVERY_HARDENING=PASS
```
