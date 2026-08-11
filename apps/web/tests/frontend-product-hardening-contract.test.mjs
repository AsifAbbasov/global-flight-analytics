// FRONTEND_PRODUCT_HARDENING_V1

import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'

function source(relativePath) {
  return readFileSync(new URL(`../${relativePath}`, import.meta.url), 'utf8')
}

test('application shell publishes keyboard skip links and a focusable main landmark', () => {
  const shell = source('components/product/application-shell.tsx')
  assert.match(shell, /href='#main-content'/)
  assert.match(shell, /href='#live-traffic'/)
  assert.match(shell, /<main id='main-content' tabIndex=\{-1\}/)
  assert.match(shell, /id='top'/)
})

test('application shell exposes aircraft-first desktop and mobile research navigation', () => {
  const shell = source('components/product/application-shell.tsx')
  assert.match(shell, /<details className='gfa-mobile-menu'>/)
  assert.match(shell, /<summary>Navigate<\/summary>/)
  assert.match(shell, /aria-label='Mobile primary navigation'/)
  assert.match(shell, /className='gfa-mobile-navigation'/)
  for (const target of [
    '#overview',
    '#historical-analytics',
    '#live-traffic',
    '#research-scope',
  ]) {
    assert.match(shell, new RegExp(target))
  }
  assert.doesNotMatch(shell, /#airport-intelligence/)
  assert.doesNotMatch(shell, /Open airport intelligence/)
})

test('global styles preserve focus reduced motion forced colors and coarse pointer targets', () => {
  const styles = source('app/globals.css')
  assert.match(styles, /\.skip-link/)
  assert.match(styles, /prefers-reduced-motion: reduce/)
  assert.match(styles, /forced-colors: active/)
  assert.match(styles, /pointer: coarse/)
  assert.match(styles, /min-height: 44px/)
})

test('segment and global error boundaries expose explicit recovery actions', () => {
  const segmentError = source('app/error.tsx')
  const globalError = source('app/global-error.tsx')
  assert.match(segmentError, /role='alert'/)
  assert.match(segmentError, /onClick=\{reset\}/)
  assert.match(segmentError, /Return to home/)
  assert.match(globalError, /role='alert'/)
  assert.match(globalError, /onClick=\{reset\}/)
  assert.match(globalError, /<html lang='en'>/)
})

test('loading and not-found routes remain accessible and actionable', () => {
  const loading = source('app/loading.tsx')
  const notFound = source('app/not-found.tsx')
  assert.match(loading, /role='status'/)
  assert.match(loading, /aria-busy='true'/)
  assert.match(loading, /Loading aviation research workspace/)
  assert.match(notFound, /id='not-found-title'/)
  assert.match(notFound, /href='\/'/)
})

test('query provider uses bounded retry reconnect and garbage collection policy', () => {
  const provider = source('providers/query-provider.tsx')
  assert.match(provider, /shouldRetryFrontendQuery/)
  assert.match(provider, /frontendRetryDelayMilliseconds/)
  assert.match(provider, /refetchOnReconnect: true/)
  assert.match(provider, /gcTime: 5 \* 60_000/)
  assert.match(provider, /networkMode: 'online'/)
})

test('remaining heavy research workspaces are split behind deterministic loading fallbacks', () => {
  const experience = source('components/regional-traffic-experience.tsx')
  assert.match(experience, /import dynamic from 'next\/dynamic'/)
  assert.match(experience, /HistoricalAnalyticsComparisonWorkspace = dynamic/)
  assert.match(experience, /TrafficDashboard = dynamic/)
  assert.match(experience, /ResearchSectionLoading/)
  assert.doesNotMatch(experience, /UnifiedAirportAnalyticsWorkspace = dynamic/)
})

test('root layout installs the runtime connectivity boundary and viewport policy', () => {
  const layout = source('app/layout.tsx')
  assert.match(layout, /RuntimeResilienceBoundary/)
  assert.match(layout, /export const viewport: Viewport/)
  assert.match(layout, /themeColor: '#111315'/)
  assert.match(layout, /<RuntimeResilienceBoundary>/)
})

test('aircraft-only frontend scope does not mount airport intelligence', () => {
  const shell = source('components/product/application-shell.tsx')
  const experience = source('components/regional-traffic-experience.tsx')
  const dashboard = source('components/traffic-dashboard.tsx')
  assert.match(shell, /Search aircraft or ICAO24/)
  assert.doesNotMatch(shell, /airport-intelligence/)
  assert.doesNotMatch(experience, /UnifiedAirportAnalyticsWorkspace/)
  assert.doesNotMatch(experience, /id='airport-intelligence'/)
  assert.doesNotMatch(dashboard, /airport-intelligence/)
  assert.doesNotMatch(dashboard, /Open Airport Intelligence/)
})

test('live traffic hydration uses deterministic initial timestamps and a post-mount clock', () => {
  const dashboard = source('components/traffic-dashboard.tsx')
  const trafficQuery = source('lib/queries/traffic.ts')
  const control = source('components/traffic/live-traffic-control.tsx')
  assert.match(dashboard, /useState\(0\)/)
  assert.match(dashboard, /window\.requestAnimationFrame\(\(\) =>/)
  assert.match(dashboard, /setStatusNow\(Date\.now\(\)\)/)
  assert.doesNotMatch(dashboard, /useState\(\(\) => Date\.now\(\)\)/)
  assert.match(trafficQuery, /initialDataUpdatedAt/)
  assert.match(trafficQuery, /Date\.parse\(item\.observed_at\)/)
  assert.match(control, /timeZone: 'UTC'/)
})

test('hidden analytical drawers do not mount their query workspaces until targeted', () => {
  const experience = source('components/regional-traffic-experience.tsx')
  assert.match(experience, /activeResearchPanel/)
  assert.match(experience, /resolveResearchPanelID/)
  assert.match(experience, /window\.addEventListener\('hashchange'/)
  assert.match(experience, /\{active \? children : null\}/)
})

test('map tool popovers are mutually exclusive through one native details group', () => {
  const dashboard = source('components/traffic-dashboard.tsx')
  assert.match(dashboard, /name='gfa-map-tool-popovers'/)
  assert.match(dashboard, /open=\{activeMapPopover === 'live-data'\}/)
  assert.match(
    dashboard,
    /open=\{activeMapPopover === 'traffic-analysis'\}/
  )
  assert.match(dashboard, /onMapPopoverToggle\('live-data'/)
  assert.match(dashboard, /onMapPopoverToggle\('traffic-analysis'/)
})

test('all right-side surfaces are mutually exclusive across map popovers and analytics drawers', () => {
  const dashboard = source('components/traffic-dashboard.tsx')
  const experience = source('components/regional-traffic-experience.tsx')
  assert.match(
    dashboard,
    /FRONTEND_UNIFIED_RIGHT_SIDEBAR_EXCLUSIVITY_V1/
  )
  assert.match(
    experience,
    /FRONTEND_UNIFIED_RIGHT_SIDEBAR_EXCLUSIVITY_V1/
  )
  assert.match(experience, /useState<MapToolPopoverID \| null>\(null\)/)
  assert.match(experience, /nextResearchPanel !== null/)
  assert.match(experience, /setActiveMapPopover\(null\)/)
  assert.match(experience, /setActiveResearchPanel\(null\)/)
  assert.match(experience, /window\.location\.replace\('#live-traffic'\)/)
  assert.match(dashboard, /onClick=\{onCloseMapPopovers\}/)
})

test('MapLibre tracker map owns the full map stage and observes container resize', () => {
  const map = source('components/map/traffic-map.tsx')
  const styles = source('app/globals.css')
  assert.match(map, /className='gfa-tracker-map-shell'/)
  assert.match(map, /new ResizeObserver\(resizeMap\)/)
  assert.match(map, /map\.resize\(\)/)
  assert.match(styles, /\.gfa-tracker-map-shell/)
  assert.match(styles, /height: 100%/)
})
