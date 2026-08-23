// FRONTEND_VISUAL_POLISH_V2

import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

function source(relativePath) {
  return readFileSync(new URL(`../${relativePath}`, import.meta.url), 'utf8')
}

test('live 2D map is promoted ahead of secondary globe analytics', () => {
  const dashboard = source('components/traffic-dashboard.tsx')

  assert.match(dashboard, /FRONTEND_VISUAL_POLISH_V2/)
  assert.match(dashboard, /data-visual-polish='tracker-workspace-v2'/)
  assert.match(
    dashboard,
    /data-visual-polish='tracker-workspace-v2'[\s\S]*?<MapEvidenceWorkspace[\s\S]*?<TrafficGlobe/
  )
  assert.match(dashboard, /xl:grid-cols-\[minmax\(0,1fr\)_420px\]/)
})

test('tracker map uses a viewport-aware surface and explicit selected-aircraft chrome', () => {
  const map = source('components/map/traffic-map.tsx')

  assert.match(map, /h-\[min\(78vh,860px\)\] min-h-\[560px\]/)
  assert.match(map, /Selected · \{selectedLabel\}/)
  assert.match(map, /\{aircraft\.length\.toLocaleString\(\)\} aircraft/)
})

test('map popup renders only supported values instead of Unknown placeholders', () => {
  const map = source('components/map/traffic-map.tsx')

  assert.match(map, /appendOptionalDetail/)
  assert.match(map, /formatObservedAltitude/)
  assert.match(map, /formatObservedSpeed/)
  assert.match(map, /formatObservedHeading/)
  assert.doesNotMatch(map, /Unknown callsign/)
  assert.doesNotMatch(map, /item\.airline \|\| 'Unknown'/)
  assert.doesNotMatch(map, /item\.aircraft_model \|\| 'Unknown'/)
  assert.doesNotMatch(map, /item\.origin_country \|\| 'Unknown'/)
})

test('aircraft explorer omits unavailable per-aircraft metrics and identity placeholders', () => {
  const explorer = source('components/aircraft/aircraft-explorer.tsx')

  assert.match(explorer, /buildAircraftListDetails/)
  assert.match(explorer, /\{identity \? \(/)
  assert.match(explorer, /\{details\.length > 0 \? \(/)
  assert.doesNotMatch(explorer, /Aircraft identity unavailable/)
  assert.doesNotMatch(explorer, /return 'Unavailable'/)
})

test('aircraft intelligence prioritizes real altitude speed heading and state telemetry', () => {
  const display = source('lib/aircraft/aircraft-intelligence-display.ts')

  assert.match(
    display,
    /field\('altitude'[\s\S]*?field\('speed'[\s\S]*?field\('heading'[\s\S]*?field\('status'/
  )
  assert.match(display, /value \* 3\.6/)
})

test('visual polish keeps accessible map chrome and adds compact tracker popup styling', () => {
  const styles = source('app/globals.css')

  assert.match(styles, /FRONTEND_VISUAL_POLISH_V2/)
  assert.match(styles, /\.gfa-aircraft-popup__telemetry/)
  assert.match(styles, /\.gfa-tracker-rail/)
  assert.match(styles, /\.gfa-scrollbar/)
  assert.match(styles, /prefers-reduced-transparency: reduce/)
  assert.match(styles, /forced-colors: active/)
})
