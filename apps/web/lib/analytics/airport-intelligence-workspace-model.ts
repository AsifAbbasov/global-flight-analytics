// FRONTEND_UNIFIED_AIRPORT_ANALYTICS_WORKSPACE_V1

import type {
  AirportIntelligenceLimitation,
  AirportIntelligenceTrends,
  AirportRankedItem,
  AirportStatistics,
} from '../../types/airport-intelligence'

export type AirportRankingSort =
  | 'position'
  | 'activity'
  | 'confidence'
  | 'movements'
  | 'routes'

export interface AirportRankingViewOptions {
  search?: string
  sort?: AirportRankingSort
  limit?: number
}

export interface AirportRankingView {
  totalCount: number
  matchedCount: number
  visibleCount: number
  items: AirportRankedItem[]
  maximumActivityScore: number
  maximumMovements: number
}

export interface AirportHistoryPoint {
  key: string
  label: string
  windowStart: string
  windowEnd: string
  arrivals: number
  departures: number
  totalMovements: number
  movementsPerHour: number
  activeRoutes: number
  coverageScore: number
  freshnessScore: number
  movementShareOfPeak: number
}

export interface AirportHistorySeries {
  totalEntryCount: number
  visibleEntryCount: number
  peakMovements: number
  peakMovementsPerHour: number
  points: AirportHistoryPoint[]
}

export interface AirportTrendSummary {
  direction: 'increasing' | 'decreasing' | 'stable' | 'unknown'
  directionLabel: string
  movementDelta: number
  movementRateDelta: number
  movementRateDeltaPercent: number | null
  activeRoutesDelta: number
  coverageDelta: number
  freshnessDelta: number
  continuityScore: number
  comparedWindows: number
  hasGaps: boolean
}

export function normalizeAirportICAOCode(
  value: string | null | undefined
): string | null {
  const normalized = value?.trim().toUpperCase() ?? ''
  return /^[A-Z0-9]{4}$/.test(normalized) ? normalized : null
}

export function buildAirportRankingView(
  airports: AirportRankedItem[],
  options: AirportRankingViewOptions = {}
): AirportRankingView {
  const search = normalizeSearch(options.search)
  const sort = options.sort ?? 'position'
  const limit = normalizeLimit(options.limit, airports.length)
  const matched = airports.filter(item => matchesSearch(item, search))
  const ordered = [...matched].sort((left, right) => compareAirports(left, right, sort))
  const items = ordered.slice(0, limit)

  return {
    totalCount: airports.length,
    matchedCount: matched.length,
    visibleCount: items.length,
    items,
    maximumActivityScore: maximum(airports.map(item => item.activity_score)),
    maximumMovements: maximum(airports.map(item => item.total_movements)),
  }
}

export function buildAirportHistorySeries(
  entries: AirportStatistics[],
  limit = 14
): AirportHistorySeries {
  const normalizedLimit = normalizeLimit(limit, Math.max(entries.length, 1))
  const ordered = [...entries].sort(compareHistoryEntries)
  const visible = ordered.slice(Math.max(0, ordered.length - normalizedLimit))
  const peakMovements = maximum(visible.map(entry => entry.total_movements))
  const peakMovementsPerHour = maximum(
    visible.map(entry => entry.movements_per_hour)
  )

  return {
    totalEntryCount: entries.length,
    visibleEntryCount: visible.length,
    peakMovements,
    peakMovementsPerHour,
    points: visible.map(entry => ({
      key: `${entry.window_start}:${entry.window_end}`,
      label: formatHistoryLabel(entry.window_end),
      windowStart: entry.window_start,
      windowEnd: entry.window_end,
      arrivals: entry.arrivals,
      departures: entry.departures,
      totalMovements: entry.total_movements,
      movementsPerHour: entry.movements_per_hour,
      activeRoutes: entry.active_routes,
      coverageScore: entry.coverage_score,
      freshnessScore: entry.freshness_score,
      movementShareOfPeak: ratio(entry.total_movements, peakMovements),
    })),
  }
}

export function buildAirportTrendSummary(
  trends: AirportIntelligenceTrends
): AirportTrendSummary {
  const direction = normalizeDirection(trends.direction)
  return {
    direction,
    directionLabel: directionLabel(direction),
    movementDelta: trends.total_movements_change,
    movementRateDelta: trends.movements_per_hour_change,
    movementRateDeltaPercent:
      trends.movements_per_hour_change_percent_known
        ? trends.movements_per_hour_change_percent
        : null,
    activeRoutesDelta: trends.active_routes_change,
    coverageDelta: trends.coverage_score_change,
    freshnessDelta: trends.freshness_score_change,
    continuityScore: clampRatio(trends.continuity_score),
    comparedWindows: Math.max(0, Math.trunc(trends.compared_windows)),
    hasGaps: trends.gap_count > 0 || trends.gap_duration_seconds > 0,
  }
}

export function mergeAirportLimitations(
  ...groups: ReadonlyArray<readonly AirportIntelligenceLimitation[] | undefined>
): AirportIntelligenceLimitation[] {
  const unique = new Map<string, AirportIntelligenceLimitation>()
  for (const group of groups) {
    for (const item of group ?? []) {
      const code = item.code.trim()
      const message = item.message.trim()
      if (code === '' && message === '') continue
      const key = `${code.toLowerCase()}\u0000${message.toLowerCase()}`
      if (!unique.has(key)) {
        unique.set(key, { code, message })
      }
    }
  }

  return [...unique.values()].sort((left, right) => {
    const codeOrder = left.code.localeCompare(right.code)
    return codeOrder !== 0 ? codeOrder : left.message.localeCompare(right.message)
  })
}

function matchesSearch(item: AirportRankedItem, search: string): boolean {
  if (search === '') return true
  return [
    item.icao_code,
    item.iata_code,
    item.name,
    item.city,
    item.country,
  ].some(value => normalizeSearch(value).includes(search))
}

function compareAirports(
  left: AirportRankedItem,
  right: AirportRankedItem,
  sort: AirportRankingSort
): number {
  switch (sort) {
    case 'activity':
      return descending(left.activity_score, right.activity_score) || stableAirportOrder(left, right)
    case 'confidence':
      return descending(left.data_confidence, right.data_confidence) || stableAirportOrder(left, right)
    case 'movements':
      return descending(left.total_movements, right.total_movements) || stableAirportOrder(left, right)
    case 'routes':
      return descending(left.active_routes, right.active_routes) || stableAirportOrder(left, right)
    case 'position':
      return ascending(left.position, right.position) || stableAirportOrder(left, right)
  }
}

function stableAirportOrder(left: AirportRankedItem, right: AirportRankedItem): number {
  const icaoOrder = left.icao_code.localeCompare(right.icao_code)
  return icaoOrder !== 0 ? icaoOrder : left.name.localeCompare(right.name)
}

function compareHistoryEntries(left: AirportStatistics, right: AirportStatistics): number {
  const leftTime = parseTime(left.window_start)
  const rightTime = parseTime(right.window_start)
  const timeOrder = ascending(leftTime, rightTime)
  if (timeOrder !== 0) return timeOrder
  const endOrder = left.window_end.localeCompare(right.window_end)
  return endOrder !== 0 ? endOrder : left.icao_code.localeCompare(right.icao_code)
}

function normalizeDirection(
  value: string
): AirportTrendSummary['direction'] {
  const normalized = value.trim().toLowerCase()
  if (normalized === 'increasing' || normalized === 'up' || normalized === 'growth') {
    return 'increasing'
  }
  if (normalized === 'decreasing' || normalized === 'down' || normalized === 'decline') {
    return 'decreasing'
  }
  if (normalized === 'stable' || normalized === 'flat' || normalized === 'unchanged') {
    return 'stable'
  }
  return 'unknown'
}

function directionLabel(direction: AirportTrendSummary['direction']): string {
  switch (direction) {
    case 'increasing':
      return 'Activity increasing'
    case 'decreasing':
      return 'Activity decreasing'
    case 'stable':
      return 'Activity broadly stable'
    case 'unknown':
      return 'Direction unavailable'
  }
}

function normalizeSearch(value: string | undefined): string {
  return value?.trim().toLocaleLowerCase() ?? ''
}

function normalizeLimit(value: number | undefined, fallback: number): number {
  if (value === undefined || !Number.isFinite(value)) return Math.max(0, fallback)
  return Math.max(0, Math.min(200, Math.trunc(value)))
}

function parseTime(value: string): number {
  const parsed = Date.parse(value)
  return Number.isFinite(parsed) ? parsed : Number.POSITIVE_INFINITY
}

function formatHistoryLabel(value: string): string {
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return 'Invalid date'
  return parsed.toISOString().slice(5, 10)
}

function maximum(values: number[]): number {
  let result = 0
  for (const value of values) {
    if (Number.isFinite(value) && value > result) result = value
  }
  return result
}

function ratio(numerator: number, denominator: number): number {
  if (!Number.isFinite(numerator) || !Number.isFinite(denominator) || denominator <= 0) {
    return 0
  }
  return clampRatio(numerator / denominator)
}

function clampRatio(value: number): number {
  if (!Number.isFinite(value)) return 0
  return Math.min(1, Math.max(0, value))
}

function ascending(left: number, right: number): number {
  return left === right ? 0 : left < right ? -1 : 1
}

function descending(left: number, right: number): number {
  return ascending(right, left)
}
