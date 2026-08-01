import assert from 'node:assert/strict'
import test from 'node:test'

const moduleURL = new URL(
  '../.test-dist/lib/traffic/traffic-workspace-model.js',
  import.meta.url
)
const importedModule = await import(moduleURL.href)
const workspaceModule = importedModule.default ?? importedModule

const {
  buildTrafficWorkspaceSelection,
  normalizeTrafficAircraftICAO24,
} = workspaceModule

test('ICAO24 normalization trims and lowercases a valid identifier', () => {
  assert.equal(normalizeTrafficAircraftICAO24('  ABC123  '), 'abc123')
})

test('empty aircraft identifiers normalize to null', () => {
  assert.equal(normalizeTrafficAircraftICAO24('   '), null)
  assert.equal(normalizeTrafficAircraftICAO24(null), null)
  assert.equal(normalizeTrafficAircraftICAO24(undefined), null)
})

test('selecting an aircraft opens the intelligence workspace', () => {
  assert.deepEqual(buildTrafficWorkspaceSelection(' A1B2C3 '), {
    icao24: 'a1b2c3',
    panel: 'intelligence',
  })
})

test('clearing selection returns to the aircraft workspace', () => {
  assert.deepEqual(buildTrafficWorkspaceSelection(null), {
    icao24: null,
    panel: 'aircraft',
  })
})
