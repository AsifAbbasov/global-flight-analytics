#!/usr/bin/env bash
set -euo pipefail

REPOSITORY="${REPOSITORY:-AsifAbbasov/global-flight-analytics}"
EXPECTED_SHA="${EXPECTED_GOVERNANCE_SHA:-}"
API_VERSION="2026-03-10"
MAIN_PROTECTION_RULESET_NAME="Global Flight Analytics main protection"
ARCHIVE_BRANCH="docs/open-aviation-metrics-positioning"
ARCHIVE_TAG="archive/open-aviation-metrics-positioning-2026-08-03"
MERGED_BRANCH="feature/active-aircraft-metric"

if [ -z "$EXPECTED_SHA" ]; then
  printf '%s\n' 'EXPECTED_GOVERNANCE_SHA is required' >&2
  exit 1
fi

test "$(git symbolic-ref --quiet --short HEAD)" = "main"
test -z "$(git status --porcelain)"
git fetch --prune origin
test "$(git rev-parse HEAD)" = "$EXPECTED_SHA"
test "$(git rev-parse origin/main)" = "$EXPECTED_SHA"
auth_status="$(gh auth status --hostname github.com 2>&1)"
printf '%s\n' "$auth_status" | grep -F 'admin:repo_hook' >/dev/null || {
  printf '%s\n' 'GitHub CLI token requires admin:repo_hook scope.' >&2
  printf '%s\n' 'Run: gh auth refresh -h github.com -s admin:repo_hook' >&2
  exit 1
}

node --test scripts/verify-repository-governance.test.mjs
node scripts/verify-repository-governance.mjs

runs_url="https://api.github.com/repos/$REPOSITORY/actions/runs?head_sha=$EXPECTED_SHA&per_page=100"
attempt=1
while [ "$attempt" -le 120 ]; do
  payload="$(curl --fail --silent --show-error --location "$runs_url")"
  result="$(RUNS_PAYLOAD="$payload" EXPECTED_SHA="$EXPECTED_SHA" node <<'NODE'
const data = JSON.parse(process.env.RUNS_PAYLOAD);
const required = ['Backend CI', 'Frontend CI', 'CodeQL'];
const selected = required.map(name => (data.workflow_runs ?? []).find(run =>
  run.name === name &&
  run.head_sha === process.env.EXPECTED_SHA &&
  run.event === 'push'
));
if (selected.some(run => !run || run.status !== 'completed')) {
  process.stdout.write('PENDING');
} else if (selected.every(run => run.conclusion === 'success')) {
  for (const run of selected) console.log(run.name + ' run_id=' + run.id + ' conclusion=success');
  console.log('SUCCESS');
} else {
  for (const run of selected) console.log(run.name + ' run_id=' + run.id + ' conclusion=' + run.conclusion);
  console.log('FAILURE');
}
NODE
)"
  case "$result" in
    *SUCCESS) printf '%s\n' "$result"; break ;;
    *FAILURE) printf '%s\n' "$result" >&2; exit 1 ;;
  esac
  sleep 5
  attempt=$((attempt + 1))
done
if [ "$attempt" -gt 120 ]; then
  printf '%s\n' 'GOVERNANCE_WORKFLOWS=TIMEOUT' >&2
  exit 1
fi

gh api --method PUT -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/vulnerability-alerts" >/dev/null
gh api --method PUT -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/automated-security-fixes" >/dev/null

repository_payload="$(mktemp)"
actions_payload="$(mktemp)"
selected_payload="$(mktemp)"
ruleset_payload="$(mktemp)"
trap 'rm -f "$repository_payload" "$actions_payload" "$selected_payload" "$ruleset_payload"' EXIT

cat >"$repository_payload" <<'JSON'
{
  "allow_merge_commit": false,
  "allow_squash_merge": true,
  "allow_rebase_merge": false,
  "allow_auto_merge": true,
  "delete_branch_on_merge": true,
  "allow_update_branch": true,
  "security_and_analysis": {
    "secret_scanning": { "status": "enabled" },
    "secret_scanning_push_protection": { "status": "enabled" }
  }
}
JSON
gh api --method PATCH -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY" --input "$repository_payload" >/dev/null

cat >"$actions_payload" <<'JSON'
{
  "enabled": true,
  "allowed_actions": "selected",
  "sha_pinning_required": true
}
JSON
gh api --method PUT -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/actions/permissions" --input "$actions_payload" >/dev/null

cat >"$selected_payload" <<'JSON'
{
  "github_owned_allowed": true,
  "verified_allowed": false,
  "patterns_allowed": ["pnpm/action-setup@*"]
}
JSON
gh api --method PUT -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/actions/permissions/selected-actions" --input "$selected_payload" >/dev/null

gh api --method PUT -H "X-GitHub-Api-Version: $API_VERSION" \
  "repos/$REPOSITORY/actions/permissions/workflow" \
  -f default_workflow_permissions=read \
  -F can_approve_pull_request_reviews=false >/dev/null

repository_id="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY" --jq '.id')"
cat >"$ruleset_payload" <<JSON
{
  "name": "$MAIN_PROTECTION_RULESET_NAME",
  "target": "branch",
  "enforcement": "active",
  "conditions": {
    "ref_name": {
      "include": ["~DEFAULT_BRANCH"],
      "exclude": []
    }
  },
  "rules": [
    { "type": "deletion" },
    { "type": "non_fast_forward" },
    { "type": "required_linear_history" },
    {
      "type": "pull_request",
      "parameters": {
        "allowed_merge_methods": ["squash"],
        "dismiss_stale_reviews_on_push": true,
        "require_code_owner_review": false,
        "require_last_push_approval": false,
        "required_approving_review_count": 0,
        "required_review_thread_resolution": true
      }
    },
    {
      "type": "required_status_checks",
      "parameters": {
        "do_not_enforce_on_create": false,
        "strict_required_status_checks_policy": true,
        "required_status_checks": [
          { "context": "Backend CI Gate" },
          { "context": "Frontend CI Gate" },
          { "context": "CodeQL Security Gate" }
        ]
      }
    },
    {
      "type": "code_scanning",
      "parameters": {
        "code_scanning_tools": [
          {
            "tool": "CodeQL",
            "alerts_threshold": "errors",
            "security_alerts_threshold": "high_or_higher"
          }
        ]
      }
    }
  ]
}
JSON

existing_ruleset_id="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/rulesets?per_page=100" --jq ".[] | select(.name == \"$MAIN_PROTECTION_RULESET_NAME\") | .id" | head -n 1)"
if [ -n "$existing_ruleset_id" ]; then
  gh api --method PUT -H "X-GitHub-Api-Version: $API_VERSION" \
    "repos/$REPOSITORY/rulesets/$existing_ruleset_id" --input "$ruleset_payload" >/dev/null
else
  gh api --method POST -H "X-GitHub-Api-Version: $API_VERSION" \
    "repos/$REPOSITORY/rulesets" --input "$ruleset_payload" >/dev/null
fi

git fetch --prune origin
if git show-ref --verify --quiet "refs/remotes/origin/$ARCHIVE_BRANCH"; then
  archive_head="$(git rev-parse "origin/$ARCHIVE_BRANCH")"
  branch_only="$(git rev-list --left-right --count "origin/main...origin/$ARCHIVE_BRANCH" | awk '{print $2}')"
  test "$branch_only" -gt 0
  if git ls-remote --exit-code --tags origin "refs/tags/$ARCHIVE_TAG" >/dev/null 2>&1; then
    test "$(git ls-remote --tags origin "refs/tags/$ARCHIVE_TAG^{}" "refs/tags/$ARCHIVE_TAG" | tail -n 1 | awk '{print $1}')" = "$archive_head"
  else
    git tag -a "$ARCHIVE_TAG" "$archive_head" -m "Archive abandoned Open Aviation Metrics positioning branch"
    git push origin "refs/tags/$ARCHIVE_TAG"
  fi
  git push origin --delete "$ARCHIVE_BRANCH"
fi

git fetch --prune origin
if git show-ref --verify --quiet "refs/remotes/origin/$MERGED_BRANCH"; then
  branch_only="$(git rev-list --left-right --count "origin/main...origin/$MERGED_BRANCH" | awk '{print $2}')"
  test "$branch_only" = "0"
  git push origin --delete "$MERGED_BRANCH"
fi

EXPECTED_GOVERNANCE_SHA="$EXPECTED_SHA" bash scripts/verify-repository-governance-settings.sh
printf '%s\n' 'REPOSITORY_GOVERNANCE_CONFIGURATION=PASS'
