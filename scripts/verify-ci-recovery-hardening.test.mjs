import test from 'node:test'
import assert from 'node:assert/strict'

import { verifyCIRecoveryHardening } from './verify-ci-recovery-hardening.mjs'

function fixture(overrides = {}) {
  return {
    backendCI: `
workflow_dispatch:
concurrency:
  cancel-in-progress: true
- name: Verify required-check recovery hardening
  run: |
    node --test scripts/verify-ci-recovery-hardening.test.mjs
    node scripts/verify-ci-recovery-hardening.mjs
    bash -n scripts/diagnose-ci-required-check-recovery.sh
`,
    diagnostic: `
EXPECTED_HEAD_SHA_MATCH=FAIL
STOP_AND_REVIEW_NEW_HEAD
WAIT_FOR_EXISTING_EXACT_SHA_RUN
RERUN_EXISTING_EXACT_SHA_RUN_FIRST
MANUAL_DISPATCH_EXISTING_WORKFLOW_FOR_EXACT_HEAD_BRANCH
gh api "repos/$REPOSITORY/commits/$EXPECTED_HEAD_SHA/check-runs?per_page=100"
gh run list --commit "$EXPECTED_HEAD_SHA"
if [[ ! "$EXPECTED_HEAD_SHA" =~ ^[0-9a-fA-F]{40}$ ]]; then exit 2; fi
if [ "$ACTIVE_COUNT" -gt 0 ]; then
  NEXT_SAFE_ACTION="WAIT_FOR_EXISTING_EXACT_SHA_RUN"
elif [ "$SUCCESS_COUNT" -gt 0 ]; then
  NEXT_SAFE_ACTION="VERIFY_REQUIRED_CHECK_GATE_FOR_EXACT_SHA"
elif [ "$FAILED_COUNT" -gt 0 ]; then
  NEXT_SAFE_ACTION="RERUN_EXISTING_EXACT_SHA_RUN_FIRST"
elif [ "$RUN_COUNT" -eq 0 ]; then
  NEXT_SAFE_ACTION="MANUAL_DISPATCH_EXISTING_WORKFLOW_FOR_EXACT_HEAD_BRANCH"
fi
`,
    runbook: `
Every recovery decision is bound to the exact pull-request \`head.sha\`.
A green run for an older SHA never validates a newer head.
Empty commits created only to retrigger checks are prohibited.
Do not create an empty commit.
A separate reduced "recovery CI" workflow is prohibited.
gh run rerun "$RUN_ID" --failed --repo "$REPOSITORY"
gh workflow run backend-ci.yml
test "$HEAD_SHA" = "$EXPECTED_HEAD_SHA"
Administrative branch-protection bypass is not part of the normal recovery path.
`,
    documentIndex: `
## Document 185 — CI Required-Check Recovery Hardening
185_CI_REQUIRED_CHECK_RECOVERY_HARDENING.md
`,
    packageJSON: {
      scripts: {
        'test:ci-recovery-hardening':
          'node --test scripts/verify-ci-recovery-hardening.test.mjs',
        'verify:ci-recovery-hardening':
          'node scripts/verify-ci-recovery-hardening.mjs',
      },
    },
    releaseVerifier: `
pnpm run test:ci-recovery-hardening
pnpm run verify:ci-recovery-hardening
`,
    workflowFilenames: ['backend-ci.yml', 'frontend-ci.yml'],
    ...overrides,
  }
}

test('accepts the hardened recovery contract', () => {
  const result = verifyCIRecoveryHardening(fixture())
  assert.equal(result.exactShaPolicy, true)
  assert.equal(result.readOnlyDiagnostic, true)
  assert.equal(result.emptyCommitProhibition, true)
  assert.equal(result.singleWorkflowPolicy, true)
})

test('rejects Backend CI without workflow_dispatch', () => {
  const value = fixture({
    backendCI: fixture().backendCI.replace('workflow_dispatch:', ''),
  })
  assert.throws(
    () => verifyCIRecoveryHardening(value),
    /manually dispatchable/
  )
})

test('rejects diagnostic commands that mutate workflow state', () => {
  const value = fixture({
    diagnostic: `${fixture().diagnostic}\ngh run rerun 123`,
  })
  assert.throws(
    () => verifyCIRecoveryHardening(value),
    /must not contain mutating command/
  )
})

test('rejects diagnostic SHA validation weaker than exact 40 hex characters', () => {
  const value = fixture({
    diagnostic: fixture().diagnostic.replace(
      '^[0-9a-fA-F]{40}$',
      '^[0-9a-fA-F]+$'
    ),
  })
  assert.throws(
    () => verifyCIRecoveryHardening(value),
    /exact 40-character hexadecimal SHA/
  )
})

test('rejects stale failure precedence over an existing exact-SHA success', () => {
  const value = fixture({
    diagnostic: fixture().diagnostic.replace(
      `elif [ "$SUCCESS_COUNT" -gt 0 ]; then
  NEXT_SAFE_ACTION="VERIFY_REQUIRED_CHECK_GATE_FOR_EXACT_SHA"
elif [ "$FAILED_COUNT" -gt 0 ]; then
  NEXT_SAFE_ACTION="RERUN_EXISTING_EXACT_SHA_RUN_FIRST"`,
      `elif [ "$FAILED_COUNT" -gt 0 ]; then
  NEXT_SAFE_ACTION="RERUN_EXISTING_EXACT_SHA_RUN_FIRST"
elif [ "$SUCCESS_COUNT" -gt 0 ]; then
  NEXT_SAFE_ACTION="VERIFY_REQUIRED_CHECK_GATE_FOR_EXACT_SHA"`
    ),
  })
  assert.throws(
    () => verifyCIRecoveryHardening(value),
    /prefer an existing exact-SHA success/
  )
})

test('rejects a runbook that loses exact-SHA binding', () => {
  const value = fixture({
    runbook: fixture().runbook.replace(
      'Every recovery decision is bound to the exact pull-request `head.sha`.',
      ''
    ),
  })
  assert.throws(
    () => verifyCIRecoveryHardening(value),
    /exact pull-request head binding/
  )
})

test('rejects a runbook that permits empty retrigger commits', () => {
  const value = fixture({
    runbook: fixture().runbook.replace(
      'Empty commits created only to retrigger checks are prohibited.',
      ''
    ),
  })
  assert.throws(
    () => verifyCIRecoveryHardening(value),
    /empty retrigger commits/
  )
})

test('rejects a separate reduced recovery workflow', () => {
  const value = fixture({
    workflowFilenames: [
      'backend-ci.yml',
      'backend-ci-recovery.yml',
    ],
  })
  assert.throws(
    () => verifyCIRecoveryHardening(value),
    /Separate recovery workflows are prohibited/
  )
})

test('rejects release verification that does not enforce the contract', () => {
  const value = fixture({
    releaseVerifier: 'echo release',
  })
  assert.throws(
    () => verifyCIRecoveryHardening(value),
    /Full release verification/
  )
})

test('rejects an unregistered recovery runbook', () => {
  const value = fixture({
    documentIndex: '## Document 184',
  })
  assert.throws(
    () => verifyCIRecoveryHardening(value),
    /Documentation Index/
  )
})
