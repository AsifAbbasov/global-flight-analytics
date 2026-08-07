import assert from 'node:assert/strict'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import test from 'node:test'

import {
  repositoryFiles,
  validateRepository,
} from './verify-stage-13-frontend-analytics-closure.mjs'

const fixtureFiles = Object.values(repositoryFiles)

function copyRepositoryFixture() {
  const temporaryRoot = fs.mkdtempSync(
    path.join(os.tmpdir(), 'gfa-stage13-frontend-closure-')
  )

  for (const relativePath of fixtureFiles) {
    const source = path.join(process.cwd(), relativePath)
    const target = path.join(temporaryRoot, relativePath)
    fs.mkdirSync(path.dirname(target), { recursive: true })
    fs.copyFileSync(source, target)
  }

  return temporaryRoot
}

function replaceInFixture(root, relativePath, from, to) {
  const target = path.join(root, relativePath)
  const source = fs.readFileSync(target, 'utf8')
  assert(source.includes(from), `fixture is missing expected token: ${from}`)
  fs.writeFileSync(target, source.replace(from, to))
}

test('repository verifier accepts the current checkout', () => {
  assert.deepEqual(validateRepository(process.cwd()), [])
})

test('verifier rejects a Stage 13 status regression', () => {
  const root = copyRepositoryFixture()
  replaceInFixture(
    root,
    repositoryFiles.implementationSequence,
    'Status: COMPLETED on 2026-08-07.',
    'Status: IN PROGRESS from 2026-07-17.'
  )

  assert(
    validateRepository(root).some(error =>
      error.includes('must remain completed')
    )
  )
})

test('verifier rejects removal of a rendered analytical panel', () => {
  const root = copyRepositoryFixture()
  replaceInFixture(
    root,
    repositoryFiles.dashboard,
    '<WeatherContextPanel',
    '<RemovedWeatherContextPanel'
  )

  assert(
    validateRepository(root).some(error =>
      error.includes('<WeatherContextPanel')
    )
  )
})

test('verifier rejects collapsed observed and projected map sources', () => {
  const root = copyRepositoryFixture()
  replaceInFixture(
    root,
    repositoryFiles.trafficMap,
    "const projectionSourceID = 'selected-aircraft-projection'",
    "const projectionSourceID = 'selected-aircraft-trajectory'"
  )

  assert(
    validateRepository(root).some(error =>
      error.includes('sources must remain distinct')
    )
  )
})

test('verifier rejects removal of the README closure disclosure', () => {
  const root = copyRepositoryFixture()
  replaceInFixture(
    root,
    repositoryFiles.readme,
    '<!-- STAGE-13-FRONTEND-ANALYTICS-CLOSURE-V1 -->',
    '<!-- REMOVED-STAGE-13-CLOSURE -->'
  )

  assert(
    validateRepository(root).some(error =>
      error.includes('README is missing the Stage 13 closure marker')
    )
  )
})

test('verifier rejects missing Frontend CI reachability', () => {
  const root = copyRepositoryFixture()
  replaceInFixture(
    root,
    repositoryFiles.frontendWorkflow,
    'pnpm run verify:stage13-frontend-analytics-closure',
    'printf "stage13 verifier removed\\n"'
  )

  assert(
    validateRepository(root).some(error =>
      error.includes('Frontend CI is missing')
    )
  )
})
