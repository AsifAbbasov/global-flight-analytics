#!/usr/bin/env node
import fs from 'node:fs'

function fail(message) {
  console.error(`SCHEDULED_PRODUCTION_SMOKE_CONTRACT=FAIL reason=${message}`)
  process.exit(1)
}

function read(path) {
  return fs.readFileSync(path, 'utf8')
}

const workflow = read('.github/workflows/production-smoke.yml')
const packageJson = JSON.parse(read('package.json'))
const release = read('scripts/verify-release.sh')
const frontendCI = read('.github/workflows/frontend-ci.yml')
const runbook = read('docs/163_PRODUCTION_DEPLOYMENT_RUNBOOK.md')

for (const literal of [
  "cron: '17 5 * * *'",
  'workflow_dispatch:',
  'contents: read',
  'group: production-smoke',
  'cancel-in-progress: false',
  'EXPECTED_API_REVISION: ${{ inputs.expected_api_revision || vars.PRODUCTION_API_REVISION }}',
  'bash scripts/smoke-production-release.sh',
  'SCHEDULED_PRODUCTION_SMOKE=PASS',
]) {
  if (!workflow.includes(literal)) fail(`workflow is missing ${literal}`)
}

if (workflow.includes('${{ github.sha }}')) fail('workflow derives deployed revision from github.sha')
if (/git\s+rev-parse\s+HEAD/.test(workflow)) fail('workflow derives deployed revision from local HEAD')
if (!/\^\[0-9a-f\]\{40\}\$/.test(workflow)) fail('workflow does not validate one full lowercase SHA')

if (packageJson.scripts?.['test:scheduled-production-smoke'] !== 'node --test scripts/verify-scheduled-production-smoke.test.mjs') {
  fail('package test entry point is missing')
}
if (packageJson.scripts?.['verify:scheduled-production-smoke'] !== 'node scripts/verify-scheduled-production-smoke.mjs') {
  fail('package verification entry point is missing')
}
for (const command of [
  'pnpm run test:scheduled-production-smoke',
  'pnpm run verify:scheduled-production-smoke',
]) {
  if (!release.includes(command)) fail(`release verification is missing ${command}`)
  if (!frontendCI.includes(command)) fail(`Frontend CI is missing ${command}`)
}
for (const watchedPath of [
  ".github/workflows/production-smoke.yml",
  "scripts/verify-scheduled-production-smoke.mjs",
  "scripts/verify-scheduled-production-smoke.test.mjs",
]) {
  if (!frontendCI.includes(`'${watchedPath}'`)) fail(`Frontend CI does not watch ${watchedPath}`)
}
for (const runbookLiteral of [
  'PRODUCTION_API_REVISION',
  'SCHEDULED_PRODUCTION_SMOKE=PASS',
]) {
  if (!runbook.includes(runbookLiteral)) fail(`runbook is missing ${runbookLiteral}`)
}
if (!/does\s+not\s+infer\s+the\s+deployed\s+revision\s+from\s+local\s+`HEAD`\s+or\s+`github\.sha`/s.test(runbook)) {
  fail('runbook does not preserve the local HEAD and github.sha deployment-truth boundary')
}

console.log('SCHEDULED_PRODUCTION_SMOKE_CONTRACT=PASS')
