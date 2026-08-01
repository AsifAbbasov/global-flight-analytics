import type { TrafficAircraft } from '../../types/traffic'

export type RegionalAltitudeBandKey =
  | 'low'
  | 'medium'
  | 'high'
  | 'unknown'

export interface RegionalAltitudeBand {
  key: RegionalAltitudeBandKey
  label: string
  range: string
  count: number
  share: number
}

export interface RankedTrafficGroup {
  label: string
  count: number
  share: number
}

export interface RegionalTrafficBriefOptions {
  rankingLimit?: number
}

export interface RegionalTrafficBriefModel {
  totalCount: number
  airborneCount: number
  groundCount: number
  knownAltitudeCount: number
  altitudeCoverage: number
  altitudeBands: RegionalAltitudeBand[]
  topAirlines: RankedTrafficGroup[]
  topOriginCountries: RankedTrafficGroup[]
  unknownAirlineCount: number
  unknownOriginCountryCount: number
}

const defaultRankingLimit = 3
const maximumRankingLimit = 5

const altitudeBandDefinitions: ReadonlyArray<{
  key: RegionalAltitudeBandKey
  label: string
  range: string
}> = [
  { key: 'low', label: 'Low altitude', range: 'Below 3,000 m' },
  { key: 'medium', label: 'Cruise transition', range: '3,000–8,999 m' },
  { key: 'high', label: 'High altitude', range: '9,000 m and above' },
  { key: 'unknown', label: 'Altitude unavailable', range: 'No usable value' },
]

export function buildRegionalTrafficBrief(
  aircraft: TrafficAircraft[],
  options: RegionalTrafficBriefOptions = {}
): RegionalTrafficBriefModel {
  const totalCount = aircraft.length
  const airborneAircraft = aircraft.filter(item => !item.on_ground)
  const airborneCount = airborneAircraft.length
  const groundCount = totalCount - airborneCount
  const rankingLimit = normalizeRankingLimit(options.rankingLimit)

  const altitudeCounts: Record<RegionalAltitudeBandKey, number> = {
    low: 0,
    medium: 0,
    high: 0,
    unknown: 0,
  }

  let knownAltitudeCount = 0
  for (const item of airborneAircraft) {
    const altitude = usableAltitude(item.altitude_m)
    if (altitude === null) {
      altitudeCounts.unknown++
      continue
    }

    knownAltitudeCount++
    if (altitude < 3000) {
      altitudeCounts.low++
    } else if (altitude < 9000) {
      altitudeCounts.medium++
    } else {
      altitudeCounts.high++
    }
  }

  const airlineRanking = rankLabels(
    aircraft.map(item => item.airline),
    totalCount,
    rankingLimit
  )
  const countryRanking = rankLabels(
    aircraft.map(item => item.origin_country),
    totalCount,
    rankingLimit
  )

  return {
    totalCount,
    airborneCount,
    groundCount,
    knownAltitudeCount,
    altitudeCoverage: ratio(knownAltitudeCount, airborneCount),
    altitudeBands: altitudeBandDefinitions.map(definition => ({
      ...definition,
      count: altitudeCounts[definition.key],
      share: ratio(altitudeCounts[definition.key], airborneCount),
    })),
    topAirlines: airlineRanking.items,
    topOriginCountries: countryRanking.items,
    unknownAirlineCount: airlineRanking.unknownCount,
    unknownOriginCountryCount: countryRanking.unknownCount,
  }
}

function usableAltitude(value: number | null): number | null {
  if (value === null || !Number.isFinite(value) || value < 0) {
    return null
  }

  return value
}

function normalizeRankingLimit(value: number | undefined): number {
  if (value === undefined || !Number.isFinite(value)) {
    return defaultRankingLimit
  }

  return Math.min(
    maximumRankingLimit,
    Math.max(1, Math.trunc(value))
  )
}

function rankLabels(
  values: string[],
  denominator: number,
  limit: number
): { items: RankedTrafficGroup[]; unknownCount: number } {
  const counts = new Map<string, { label: string; count: number }>()
  let unknownCount = 0

  for (const value of values) {
    const label = normalizeLabel(value)
    if (label === '') {
      unknownCount++
      continue
    }

    const key = label.toLocaleLowerCase('en-US')
    const existing = counts.get(key)
    if (existing) {
      existing.count++
      if (compareLabels(label, existing.label) < 0) {
        existing.label = label
      }
    } else {
      counts.set(key, { label, count: 1 })
    }
  }

  const items = [...counts.values()]
    .sort((left, right) => {
      const countDifference = right.count - left.count
      return countDifference !== 0
        ? countDifference
        : compareLabels(left.label, right.label)
    })
    .slice(0, limit)
    .map(item => ({
      label: item.label,
      count: item.count,
      share: ratio(item.count, denominator),
    }))

  return { items, unknownCount }
}

function normalizeLabel(value: string): string {
  return value.trim().replace(/\s+/g, ' ')
}

function compareLabels(left: string, right: string): number {
  return left.localeCompare(right, 'en', {
    numeric: true,
    sensitivity: 'base',
  })
}

function ratio(numerator: number, denominator: number): number {
  return denominator > 0 ? numerator / denominator : 0
}
