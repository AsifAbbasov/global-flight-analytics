import type { TrafficAircraft } from '../../types/traffic'

export function resolveAirspaceAsOfTime(
  aircraft: TrafficAircraft[]
): string | null {
  let latestMilliseconds = Number.NEGATIVE_INFINITY
  let latestTimestamp: string | null = null

  for (const item of aircraft) {
    const timestamp = item.observed_at.trim()
    const milliseconds = Date.parse(timestamp)
    if (!Number.isFinite(milliseconds)) {
      continue
    }
    if (milliseconds > latestMilliseconds) {
      latestMilliseconds = milliseconds
      latestTimestamp = new Date(milliseconds).toISOString()
    }
  }

  return latestTimestamp
}

export function canRequestAirspaceRegionAnalytics(
  regionCode: string,
  asOfTime: string | null
): boolean {
  const normalizedRegionCode = regionCode.trim().toLowerCase()
  return (
    normalizedRegionCode !== '' &&
    normalizedRegionCode !== 'world' &&
    asOfTime !== null
  )
}
