#!/usr/bin/env bash
set -euo pipefail

REPOSITORY="${REPOSITORY:-AsifAbbasov/global-flight-analytics}"
EXPECTED_SHA="${EXPECTED_GOVERNANCE_SHA:-}"
API_VERSION="2026-03-10"
RULESET_NAME="Global Flight Analytics main protection"

if [ -z "$EXPECTED_SHA" ]; then
  printf '%s\n' 'EXPECTED_GOVERNANCE_SHA is required' >&2
  exit 1
fi

repository_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY")"
actions_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/actions/permissions")"
workflow_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/actions/permissions/workflow")"
selected_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/actions/permissions/selected-actions")"
rulesets_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/rulesets?per_page=100")"
rules_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/rules/branches/main")"
analyses_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/code-scanning/analyses?ref=refs/heads/main&per_page=100")"
dependabot_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/dependabot/alerts?state=open&per_page=100")"
secrets_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/secret-scanning/alerts?state=open&per_page=100")"

REPOSITORY_JSON="$repository_json" ACTIONS_JSON="$actions_json" WORKFLOW_JSON="$workflow_json" \
SELECTED_JSON="$selected_json" RULESETS_JSON="$rulesets_json" RULES_JSON="$rules_json" \
ANALYSES_JSON="$analyses_json" DEPENDABOT_JSON="$dependabot_json" SECRETS_JSON="$secrets_json" \
EXPECTED_SHA="$EXPECTED_SHA" RULESET_NAME="$RULESET_NAME" node <<'NODE'
const repository = JSON.parse(process.env.REPOSITORY_JSON);
const actions = JSON.parse(process.env.ACTIONS_JSON);
const workflow = JSON.parse(process.env.WORKFLOW_JSON);
const selected = JSON.parse(process.env.SELECTED_JSON);
const rulesets = JSON.parse(process.env.RULESETS_JSON);
const rules = JSON.parse(process.env.RULES_JSON);
const analyses = JSON.parse(process.env.ANALYSES_JSON);
JSON.parse(process.env.DEPENDABOT_JSON);
JSON.parse(process.env.SECRETS_JSON);

function assert(condition, message) {
  if (!condition) throw new Error(message);
}
assert(repository.default_branch === 'main', 'default branch is not main');
assert(repository.allow_merge_commit === false, 'merge commits are still allowed');
assert(repository.allow_rebase_merge === false, 'rebase merges are still allowed');
assert(repository.allow_squash_merge === true, 'squash merge is not allowed');
assert(repository.delete_branch_on_merge === true, 'delete branch on merge is disabled');
assert(repository.allow_update_branch === true, 'update branch is disabled');
assert(actions.enabled === true, 'Actions are disabled');
assert(actions.allowed_actions === 'selected', 'Actions are not restricted to selected');
assert(actions.sha_pinning_required === true, 'Action SHA pinning policy is disabled');
assert(workflow.default_workflow_permissions === 'read', 'workflow token is not read-only');
assert(workflow.can_approve_pull_request_reviews === false, 'workflow token can approve PRs');
assert(selected.github_owned_allowed === true, 'GitHub-owned actions are not allowed');
assert(selected.verified_allowed === false, 'all verified actions are unexpectedly allowed');
assert((selected.patterns_allowed ?? []).includes('pnpm/action-setup@*'), 'pnpm action allowlist missing');

const ruleset = rulesets.find(item => item.name === process.env.RULESET_NAME);
assert(ruleset?.enforcement === 'active', 'active main ruleset missing');
const types = new Set(rules.map(rule => rule.type));
for (const type of ['deletion', 'non_fast_forward', 'required_linear_history', 'pull_request', 'required_status_checks', 'code_scanning']) {
  assert(types.has(type), 'active main rule missing: ' + type);
}
for (const context of ['Backend CI Gate', 'Frontend CI Gate', 'CodeQL Security Gate']) {
  assert(JSON.stringify(rules).includes(context), 'required status context missing: ' + context);
}
assert(analyses.some(item => item.commit_sha === process.env.EXPECTED_SHA && item.tool?.name === 'CodeQL'), 'CodeQL analysis for expected SHA missing');
console.log('DEPENDABOT_ALERTS_API=PASS');
console.log('SECRET_SCANNING_ALERTS_API=PASS');
console.log('CODEQL_ANALYSIS=PASS');
console.log('MAIN_RULESET=PASS');
console.log('ACTIONS_POLICY=PASS');
console.log('REPOSITORY_GOVERNANCE_SETTINGS=PASS');
NODE
