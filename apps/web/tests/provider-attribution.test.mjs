import test from 'node:test'
import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'

const pageSource = await readFile(new URL('../app/page.tsx', import.meta.url), 'utf8')

test('application keeps visible ADSB.lol ODbL attribution', () => {
  assert.match(pageSource, /https:\/\/adsb\.lol\//)
  assert.match(pageSource, /https:\/\/opendatacommons\.org\/licenses\/odbl\//)
  assert.match(pageSource, /Traffic data may include/)
  assert.match(pageSource, /ADSB\.lol/)
  assert.match(pageSource, /ODbL/)
})
