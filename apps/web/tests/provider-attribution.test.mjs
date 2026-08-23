import test from 'node:test'
import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'

const pageSource = await readFile(new URL('../app/page.tsx', import.meta.url), 'utf8')

test('application keeps visible ADSB.lol ODbL attribution', () => {
  assert.ok(pageSource.includes("href='https://adsb.lol/'"))
  assert.ok(
    pageSource.includes("href='https://opendatacommons.org/licenses/odbl/'")
  )
  assert.ok(pageSource.includes('Traffic data may include'))
  assert.ok(pageSource.includes('ADSB.lol'))
  assert.ok(pageSource.includes('ODbL'))
})
