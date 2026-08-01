import assert from 'node:assert/strict'
import test from 'node:test'

const modelModuleURL = new URL(
  '../.test-dist/lib/traffic/workspace-url-state.js',
  import.meta.url
)
const importedModelModule = await import(modelModuleURL.href)
const modelModule = importedModelModule.default ?? importedModelModule
const {
  buildTrafficWorkspaceSearch,
  buildTrafficWorkspaceURL,
  parseTrafficWorkspaceSearch,
} = modelModule

const regionCodes = ['world', 'az', 'tr']

test('valid deep links restore region, aircraft and inferred intelligence view', () => {
  const state = parseTrafficWorkspaceSearch(
    '?region=AZ&aircraft=ABC123',
    regionCodes,
    'world'
  )

  assert.deepEqual(state, {
    regionCode: 'az',
    aircraftICAO24: 'abc123',
    panel: 'intelligence',
  })
})

test('invalid region and aircraft parameters fall back safely', () => {
  const state = parseTrafficWorkspaceSearch(
    '?region=missing&aircraft=not-an-icao24&view=intelligence',
    regionCodes,
    'world'
  )

  assert.deepEqual(state, {
    regionCode: 'world',
    aircraftICAO24: null,
    panel: 'intelligence',
  })
})

test('explicit aircraft view remains valid while a selection is retained', () => {
  const state = parseTrafficWorkspaceSearch(
    '?region=tr&aircraft=4b1805&view=aircraft',
    regionCodes,
    'world'
  )

  assert.deepEqual(state, {
    regionCode: 'tr',
    aircraftICAO24: '4b1805',
    panel: 'aircraft',
  })
})

test('serialization is canonical and preserves unrelated parameters', () => {
  const search = buildTrafficWorkspaceSearch(
    '?z=last&utm_source=portfolio&region=world&utm_source=review',
    {
      regionCode: 'az',
      aircraftICAO24: 'ABC123',
      panel: 'intelligence',
    }
  )

  assert.equal(
    search,
    '?region=az&aircraft=abc123&view=intelligence&utm_source=portfolio&utm_source=review&z=last'
  )
})

test('clearing a selection removes the aircraft parameter', () => {
  const search = buildTrafficWorkspaceSearch(
    '?region=az&aircraft=abc123&view=intelligence',
    {
      regionCode: 'az',
      aircraftICAO24: null,
      panel: 'aircraft',
    }
  )

  assert.equal(search, '?region=az&view=aircraft')
})

test('complete URLs preserve pathname and hash fragments', () => {
  const url = buildTrafficWorkspaceURL(
    '/research',
    '?campaign=demo',
    'live-traffic',
    {
      regionCode: 'tr',
      aircraftICAO24: '4b1805',
      panel: 'intelligence',
    }
  )

  assert.equal(
    url,
    '/research?region=tr&aircraft=4b1805&view=intelligence&campaign=demo#live-traffic'
  )
})
