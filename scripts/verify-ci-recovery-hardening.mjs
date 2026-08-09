#!/usr/bin/env node

import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const repositoryRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  '..'
)

function assert(condition, message) {
  if (!condition) {
    throw new Error(message)
  }
}

function read(relativePath, root = repositoryRoot) {
  return fs.readFileSync(path.join(root, relativePath), 'utf8')
}

export function verifyCIRecoveryHardening({
  backendCI,
  diagnostic,
  runbook,
  documentIndex,
  packageJSON,
  releaseVerifier,
  workflowFilenames,
}) {
  assert(
    backendCI.includes('workflow_dispatch:'),
    'Backend CI must remain manually dispatchable'
  )
  assert(
    backendCI.includes('cancel-in-progress: true'),
    'Backend CI must keep explicit cancel-in-progress semantics'
  )
  assert(
    backendCI.includes('Verify required-check recovery hardening'),
    'Backend CI must enforce the required-check recovery contract'
  )
  assert(
    backendCI.includes('node --test scripts/verify-ci-recovery-hardening.test.mjs') &&
      backendCI.includes('node scripts/verify-ci-recovery-hardening.mjs') &&
      backendCI.includes('bash -n scripts/diagnose-ci-required-check-recovery.sh'),
    'Backend CI recovery step must run tests, verifier and shell syntax validation'
  )

  assert(
    runbook.includes('Every recovery decision is bound to the exact pull-request `head.sha`.') &&
      runbook.includes('A green run for an older SHA never validates a newer head.'),
    'Runbook must require exact pull-request head binding'
  )
  assert(
    runbook.includes('Empty commits created only to retrigger checks are prohibited.') &&
      runbook.includes('Do not create an empty commit.'),
    'Runbook must prohibit empty retrigger commits'
  )
  assert(
    runbook.includes('A separate reduced "recovery CI" workflow is prohibited.'),
    'Runbook must prohibit a second reduced recovery workflow'
  )
  assert(
    runbook.includes('gh run rerun "$RUN_ID" --failed --repo "$REPOSITORY"') &&
      runbook.includes('gh workflow run backend-ci.yml') &&
      runbook.includes('test "$HEAD_SHA" = "$EXPECTED_HEAD_SHA"'),
    'Runbook must define rerun, exact-SHA check and existing-workflow dispatch'
  )
  assert(
    runbook.includes('Administrative branch-protection bypass is not part of the normal recovery path.'),
    'Runbook must keep normal recovery fail-closed'
  )

  const forbiddenDiagnosticMutators = [
    'gh run rerun',
    'gh workflow run',
    'git commit',
    'git push',
    'gh pr merge',
    '--admin',
  ]
  for (const token of forbiddenDiagnosticMutators) {
    assert(
      !diagnostic.includes(token),
      `Read-only diagnostic must not contain mutating command: ${token}`
    )
  }

  assert(
    diagnostic.includes('EXPECTED_HEAD_SHA_MATCH=FAIL') &&
      diagnostic.includes('STOP_AND_REVIEW_NEW_HEAD') &&
      diagnostic.includes('WAIT_FOR_EXISTING_EXACT_SHA_RUN') &&
      diagnostic.includes('RERUN_EXISTING_EXACT_SHA_RUN_FIRST') &&
      diagnostic.includes('MANUAL_DISPATCH_EXISTING_WORKFLOW_FOR_EXACT_HEAD_BRANCH'),
    'Diagnostic must classify exact-SHA recovery states'
  )
  assert(
    diagnostic.includes('/check-runs?per_page=100') &&
      diagnostic.includes('--commit "$EXPECTED_HEAD_SHA"'),
    'Diagnostic must inspect exact-SHA workflow and check runs'
  )
  assert(
    diagnostic.includes('^[0-9a-fA-F]{40}$'),
    'Diagnostic must require an exact 40-character hexadecimal SHA'
  )

  const successDecision = diagnostic.indexOf(
    'elif [ "$SUCCESS_COUNT" -gt 0 ]'
  )
  const failedDecision = diagnostic.indexOf(
    'elif [ "$FAILED_COUNT" -gt 0 ]'
  )
  assert(
    successDecision >= 0 &&
      failedDecision >= 0 &&
      successDecision < failedDecision,
    'Diagnostic must prefer an existing exact-SHA success over stale failed runs'
  )

  const recoveryWorkflowFiles = workflowFilenames.filter(
    (name) => /recovery/i.test(name)
  )
  assert(
    recoveryWorkflowFiles.length === 0,
    `Separate recovery workflows are prohibited: ${recoveryWorkflowFiles.join(', ')}`
  )

  assert(
    packageJSON.scripts?.['test:ci-recovery-hardening'] ===
      'node --test scripts/verify-ci-recovery-hardening.test.mjs',
    'package.json must expose CI recovery regression tests'
  )
  assert(
    packageJSON.scripts?.['verify:ci-recovery-hardening'] ===
      'node scripts/verify-ci-recovery-hardening.mjs',
    'package.json must expose CI recovery verification'
  )
  assert(
    releaseVerifier.includes('pnpm run test:ci-recovery-hardening') &&
      releaseVerifier.includes('pnpm run verify:ci-recovery-hardening'),
    'Full release verification must enforce CI recovery hardening'
  )
  assert(
    documentIndex.includes('## Document 185 — CI Required-Check Recovery Hardening') &&
      documentIndex.includes('185_CI_REQUIRED_CHECK_RECOVERY_HARDENING.md'),
    'Documentation Index must register Document 185'
  )

  return {
    exactShaPolicy: true,
    readOnlyDiagnostic: true,
    emptyCommitProhibition: true,
    singleWorkflowPolicy: true,
  }
}

export function loadRepositorySources(root = repositoryRoot) {
  const workflowDirectory = path.join(root, '.github', 'workflows')
  return {
    backendCI: read('.github/workflows/backend-ci.yml', root),
    diagnostic: read('scripts/diagnose-ci-required-check-recovery.sh', root),
    runbook: read('docs/185_CI_REQUIRED_CHECK_RECOVERY_HARDENING.md', root),
    documentIndex: read('docs/DOCUMENT_INDEX.md', root),
    packageJSON: JSON.parse(read('package.json', root)),
    releaseVerifier: read('scripts/verify-release.sh', root),
    workflowFilenames: fs.readdirSync(workflowDirectory),
  }
}

function main() {
  verifyCIRecoveryHardening(loadRepositorySources())
  process.stdout.write(
    [
      'CI_RECOVERY_EXACT_SHA_POLICY=PASS',
      'CI_RECOVERY_READ_ONLY_DIAGNOSTIC=PASS',
      'CI_RECOVERY_EMPTY_COMMIT_PROHIBITION=PASS',
      'CI_RECOVERY_SINGLE_WORKFLOW_POLICY=PASS',
      'CI_REQUIRED_CHECK_RECOVERY_HARDENING=PASS',
    ].join('\n') + '\n'
  )
}

const currentFile = fileURLToPath(import.meta.url)
if (process.argv[1] && path.resolve(process.argv[1]) === currentFile) {
  main()
}
