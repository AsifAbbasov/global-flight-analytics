import assert from 'node:assert/strict'
import fs from 'node:fs'
import test from 'node:test'

function read(path) {
  return fs.readFileSync(path, 'utf8')
}

const closure = read('../../docs/213_VERSION_2_RECONCILIATION_POST_MERGE_CLOSURE.md')
const index = read('../../docs/DOCUMENT_INDEX.md')
const findings = read('../../docs/FINDING_REGISTER.md')
const reconciliation = read('../../docs/212_VERSION_2_RECONCILIATION_AUDIT.md')

test('Version 2 reconciliation closure preserves exact merge identity and independent post-merge evidence', () => {
  assert.match(closure, /VERSION_2_RECONCILIATION_PR=169/)
  assert.match(
    closure,
    /VERSION_2_RECONCILIATION_APPROVED_HEAD=ded9b4e27c51fb2792d2f9e510c8a1380ddc6b32/,
  )
  assert.match(
    closure,
    /VERSION_2_RECONCILIATION_MAIN_SHA=32d23d30e976097e4dd6f662b3e2890f70b7f33c/,
  )
  assert.match(closure, /VERSION_2_PREMERGE_EXACT_HEAD_CI=6_OF_6_PASS/)
  assert.match(closure, /VERSION_2_POST_MERGE_GITHUB_CI=5_OF_5_PASS/)
  assert.match(closure, /VERSION_2_POST_MERGE_VERCEL=PASS/)
  assert.match(closure, /VERSION_2_POST_MERGE_API_LOAD=NOT_TRIGGERED_BY_PUSH_POLICY/)
})

test('closure protects all exact pre-merge and post-merge workflow evidence', () => {
  for (const run of [
    '34111853003',
    '34111853094',
    '34111852994',
    '34111853101',
    '34111853123',
    '34111853108',
    '34112888003',
    '34112888088',
    '34112888023',
    '34112888002',
    '34112888055',
  ]) {
    assert.match(closure, new RegExp(run))
  }
  assert.match(closure, /A3ejfZXTT37S8pxNtHEnZacx6EnB/)
  assert.match(closure, /8nNZm6JqM7hS6j3eRsWvyhRC7Y5f/)
})

test('closure preserves rejected intermediate validation rather than laundering failed heads', () => {
  assert.match(closure, /4488c6268215921a6796ed5410262b51407d693b/)
  assert.match(closure, /34111059709/)
  assert.match(closure, /4780aa5a4025def743bcb2707f56ce356e586c3b/)
  assert.match(closure, /34111566742/)
})

test('closure preserves bounded Version 2 scope instead of inventing deferred research', () => {
  assert.match(closure, /VERSION_2_ORIGINAL_SCOPE_ITEMS=18/)
  assert.match(closure, /VERSION_2_IMPLEMENTED_CAPABILITIES=14/)
  assert.match(closure, /VERSION_2_BOUNDED_UNCALIBRATED_CAPABILITIES=1/)
  assert.match(closure, /VERSION_2_DEFERRED_RESEARCH_CAPABILITIES=3/)
  assert.match(closure, /VERSION_2_DISCRETE_FRECHET=DEFERRED_RESEARCH/)
  assert.match(closure, /VERSION_2_TRAJECTORY_SPATIAL_INDEX=DEFERRED_RESEARCH/)
  assert.match(closure, /VERSION_2_WEATHER_GRID=DEFERRED_RESEARCH/)
  assert.match(closure, /VERSION_2_NEW_BACKEND_ENDPOINT=NONE/)
  assert.match(closure, /VERSION_2_NEW_DATABASE_TABLE=NONE/)
  assert.match(closure, /VERSION_2_NEW_MIGRATION=NONE/)
  assert.match(closure, /VERSION_2_ADDITIONAL_COST=0_RUB/)
})

test('Document 212 remains immutable candidate history while Document 213 owns post-merge closure', () => {
  assert.match(reconciliation, /VERSION_2_RECONCILIATION=CANDIDATE/)
  assert.match(
    reconciliation,
    /VERSION_2_CANONICAL_CLOSURE=PENDING_EXACT_HEAD_CI_MERGE_AND_POST_MERGE_VALIDATION/,
  )
  assert.doesNotMatch(reconciliation, /VERSION_2_RECONCILIATION=CLOSED/)
  assert.match(index, /Document 213/)
  assert.match(index, /213_VERSION_2_RECONCILIATION_POST_MERGE_CLOSURE\.md/)
})

test('Finding Register closes only GFA-GOV-457 and leaves GFA-SEC-445 independent', () => {
  assert.match(findings, /GFA-GOV-457[^\n]*CLOSED/)
  assert.match(findings, /GFA-SEC-445[^\n]*IN_PROGRESS/)
  assert.match(findings, /Document 213[^\n]*Version 2 reconciliation/i)
  assert.match(findings, /one finding \(`GFA-SEC-445`\) remains IN_PROGRESS/)
})

test('closure candidate cannot claim canonical closure before its own merge and final main validation', () => {
  assert.match(closure, /VERSION_2_RECONCILIATION_POST_MERGE=CLOSURE_CANDIDATE/)
  assert.match(closure, /GFA_GOV_457=CLOSURE_CANDIDATE/)
  assert.match(closure, /VERSION_2_RECONCILIATION_IMPLEMENTATION=MERGED/)
  assert.match(closure, /VERSION_2_RECONCILIATION_DOCUMENTATION=CLOSURE_CANDIDATE/)
  assert.match(closure, /VERSION_2_RECONCILIATION=CLOSED/)
  assert.match(closure, /VERSION_2_RECONCILIATION_CI_VERIFIED=YES/)
})
