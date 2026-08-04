import assert from 'node:assert/strict'
import fs from 'node:fs'
import { execFileSync } from 'node:child_process'
import test from 'node:test'

const workflow = fs.readFileSync('.github/workflows/production-smoke.yml', 'utf8')

test('scheduled production smoke contract passes', () => {
  const output = execFileSync(
    process.execPath,
    ['scripts/verify-scheduled-production-smoke.mjs'],
    { encoding: 'utf8' },
  )
  assert.match(output, /SCHEDULED_PRODUCTION_SMOKE_CONTRACT=PASS/)
})

test('scheduled smoke requires explicit deployment truth', () => {
  assert.match(
    workflow,
    /EXPECTED_API_REVISION: \$\{\{ inputs\.expected_api_revision \|\| vars\.PRODUCTION_API_REVISION \}\}/,
  )
  assert.doesNotMatch(workflow, /\$\{\{ github\.sha \}\}/)
  assert.doesNotMatch(workflow, /git\s+rev-parse\s+HEAD/)
  assert.match(workflow, /\^\[0-9a-f\]\{40\}\$/)
})

test('scheduled smoke exercises the complete public release path', () => {
  assert.match(workflow, /FRONTEND_URL: https:\/\/global-flight-analytics-web\.vercel\.app/)
  assert.match(workflow, /API_BASE_URL: https:\/\/global-flight-analytics-api\.onrender\.com/)
  assert.match(workflow, /bash scripts\/smoke-production-release\.sh/)
  assert.match(workflow, /SCHEDULED_PRODUCTION_SMOKE=PASS/)
})
