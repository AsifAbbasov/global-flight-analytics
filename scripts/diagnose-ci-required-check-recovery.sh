#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  scripts/diagnose-ci-required-check-recovery.sh OWNER/REPO PR_NUMBER EXPECTED_HEAD_SHA [WORKFLOW_FILE]

This command is READ-ONLY. It does not rerun, dispatch, commit, push, merge, or
change GitHub settings. It inspects the pull request head, exact-SHA workflow
runs, and exact-SHA check runs, then prints the safest next action.
EOF
}

if [ "$#" -lt 3 ] || [ "$#" -gt 4 ]; then
  usage >&2
  exit 2
fi

REPOSITORY="$1"
PR_NUMBER="$2"
EXPECTED_HEAD_SHA="$3"
WORKFLOW_FILE="${4:-backend-ci.yml}"

command -v gh >/dev/null 2>&1 || {
  printf '%s\n' "ERROR=gh CLI is required" >&2
  exit 1
}

gh auth status >/dev/null 2>&1 || {
  printf '%s\n' "ERROR=gh CLI authentication is required" >&2
  exit 1
}

case "$PR_NUMBER" in
  ''|*[!0-9]*)
    printf '%s\n' "ERROR=PR number must be numeric" >&2
    exit 2
    ;;
esac

if [[ ! "$EXPECTED_HEAD_SHA" =~ ^[0-9a-fA-F]{40}$ ]]; then
  printf '%s\n' "ERROR=expected head SHA must be exactly 40 hexadecimal characters" >&2
  exit 2
fi

ACTUAL_HEAD_SHA="$(gh api "repos/$REPOSITORY/pulls/$PR_NUMBER" --jq '.head.sha')"
HEAD_BRANCH="$(gh api "repos/$REPOSITORY/pulls/$PR_NUMBER" --jq '.head.ref')"

if [ "$ACTUAL_HEAD_SHA" != "$EXPECTED_HEAD_SHA" ]; then
  printf '%s\n' \
    "EXPECTED_HEAD_SHA=$EXPECTED_HEAD_SHA" \
    "ACTUAL_HEAD_SHA=$ACTUAL_HEAD_SHA" \
    "EXPECTED_HEAD_SHA_MATCH=FAIL" \
    "NEXT_SAFE_ACTION=STOP_AND_REVIEW_NEW_HEAD"
  exit 1
fi

printf '%s\n' \
  "REPOSITORY=$REPOSITORY" \
  "PR_NUMBER=$PR_NUMBER" \
  "HEAD_BRANCH=$HEAD_BRANCH" \
  "EXPECTED_HEAD_SHA=$EXPECTED_HEAD_SHA" \
  "EXPECTED_HEAD_SHA_MATCH=PASS" \
  "WORKFLOW_FILE=$WORKFLOW_FILE"

RUN_COUNT="$(
  gh run list --repo "$REPOSITORY" --workflow "$WORKFLOW_FILE" \
    --commit "$EXPECTED_HEAD_SHA" --limit 20 --json databaseId --jq 'length'
)"
ACTIVE_COUNT="$(
  gh run list --repo "$REPOSITORY" --workflow "$WORKFLOW_FILE" \
    --commit "$EXPECTED_HEAD_SHA" --limit 20 --json status \
    --jq '[.[] | select(.status == "queued" or .status == "in_progress" or .status == "waiting" or .status == "requested" or .status == "pending")] | length'
)"
FAILED_COUNT="$(
  gh run list --repo "$REPOSITORY" --workflow "$WORKFLOW_FILE" \
    --commit "$EXPECTED_HEAD_SHA" --limit 20 --json conclusion \
    --jq '[.[] | select(.conclusion == "failure" or .conclusion == "cancelled" or .conclusion == "timed_out" or .conclusion == "action_required" or .conclusion == "stale" or .conclusion == "startup_failure")] | length'
)"
SUCCESS_COUNT="$(
  gh run list --repo "$REPOSITORY" --workflow "$WORKFLOW_FILE" \
    --commit "$EXPECTED_HEAD_SHA" --limit 20 --json conclusion \
    --jq '[.[] | select(.conclusion == "success")] | length'
)"

printf '%s\n' \
  "EXACT_SHA_WORKFLOW_RUN_COUNT=$RUN_COUNT" \
  "EXACT_SHA_ACTIVE_RUN_COUNT=$ACTIVE_COUNT" \
  "EXACT_SHA_FAILED_RUN_COUNT=$FAILED_COUNT" \
  "EXACT_SHA_SUCCESS_RUN_COUNT=$SUCCESS_COUNT"

printf '%s\n' "EXACT_SHA_WORKFLOW_RUNS_BEGIN"
gh run list --repo "$REPOSITORY" --workflow "$WORKFLOW_FILE" \
  --commit "$EXPECTED_HEAD_SHA" --limit 20 \
  --json databaseId,event,status,conclusion,headSha,createdAt,url
printf '%s\n' "EXACT_SHA_WORKFLOW_RUNS_END"

printf '%s\n' "EXACT_SHA_CHECK_RUNS_BEGIN"
gh api -H "Accept: application/vnd.github+json" \
  "repos/$REPOSITORY/commits/$EXPECTED_HEAD_SHA/check-runs?per_page=100" \
  --paginate \
  --jq '.check_runs[] | [.name, .status, (.conclusion // ""), .details_url] | @tsv'
printf '%s\n' "EXACT_SHA_CHECK_RUNS_END"

if [ "$ACTIVE_COUNT" -gt 0 ]; then
  NEXT_SAFE_ACTION="WAIT_FOR_EXISTING_EXACT_SHA_RUN"
elif [ "$SUCCESS_COUNT" -gt 0 ]; then
  NEXT_SAFE_ACTION="VERIFY_REQUIRED_CHECK_GATE_FOR_EXACT_SHA"
elif [ "$FAILED_COUNT" -gt 0 ]; then
  NEXT_SAFE_ACTION="RERUN_EXISTING_EXACT_SHA_RUN_FIRST"
elif [ "$RUN_COUNT" -eq 0 ]; then
  NEXT_SAFE_ACTION="MANUAL_DISPATCH_EXISTING_WORKFLOW_FOR_EXACT_HEAD_BRANCH"
else
  NEXT_SAFE_ACTION="INSPECT_EXACT_SHA_RUN_STATE"
fi

printf '%s\n' \
  "NEXT_SAFE_ACTION=$NEXT_SAFE_ACTION" \
  "CI_REQUIRED_CHECK_RECOVERY_DIAGNOSTIC=PASS"
