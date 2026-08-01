export const liveTrafficRefreshIntervals = [
  30_000,
  60_000,
  120_000,
] as const

export type LiveTrafficRefreshIntervalMilliseconds =
  (typeof liveTrafficRefreshIntervals)[number]

export type LiveTrafficFreshness =
  | 'waiting'
  | 'current'
  | 'aging'
  | 'stale'

export interface LiveTrafficStatusInput {
  now: number
  dataUpdatedAt: number
  refreshIntervalMilliseconds?: number
  autoRefreshEnabled: boolean
}

export interface LiveTrafficStatusModel {
  refreshIntervalMilliseconds: LiveTrafficRefreshIntervalMilliseconds
  freshness: LiveTrafficFreshness
  ageMilliseconds: number | null
  nextRefreshInMilliseconds: number | null
  intervalProgress: number
  refreshDue: boolean
}

export const defaultLiveTrafficRefreshIntervalMilliseconds = 60_000

export function normalizeLiveTrafficRefreshInterval(
  value: number | undefined
): LiveTrafficRefreshIntervalMilliseconds {
  for (const interval of liveTrafficRefreshIntervals) {
    if (value === interval) {
      return interval
    }
  }

  return defaultLiveTrafficRefreshIntervalMilliseconds
}

export function buildLiveTrafficStatus(
  input: LiveTrafficStatusInput
): LiveTrafficStatusModel {
  const refreshIntervalMilliseconds =
    normalizeLiveTrafficRefreshInterval(
      input.refreshIntervalMilliseconds
    )
  const dataUpdatedAt = normalizeTimestamp(input.dataUpdatedAt)

  if (dataUpdatedAt === null) {
    return {
      refreshIntervalMilliseconds,
      freshness: 'waiting',
      ageMilliseconds: null,
      nextRefreshInMilliseconds: null,
      intervalProgress: 0,
      refreshDue: false,
    }
  }

  const normalizedNow = normalizeTimestamp(input.now) ?? dataUpdatedAt
  const ageMilliseconds = Math.max(0, normalizedNow - dataUpdatedAt)
  const freshness = resolveFreshness(
    ageMilliseconds,
    refreshIntervalMilliseconds
  )
  const refreshDue = ageMilliseconds >= refreshIntervalMilliseconds

  return {
    refreshIntervalMilliseconds,
    freshness,
    ageMilliseconds,
    nextRefreshInMilliseconds: input.autoRefreshEnabled
      ? Math.max(0, refreshIntervalMilliseconds - ageMilliseconds)
      : null,
    intervalProgress: Math.min(
      1,
      ageMilliseconds / refreshIntervalMilliseconds
    ),
    refreshDue,
  }
}

function resolveFreshness(
  ageMilliseconds: number,
  refreshIntervalMilliseconds: number
): LiveTrafficFreshness {
  if (ageMilliseconds < refreshIntervalMilliseconds) {
    return 'current'
  }
  if (ageMilliseconds < refreshIntervalMilliseconds * 2) {
    return 'aging'
  }

  return 'stale'
}

function normalizeTimestamp(value: number): number | null {
  if (!Number.isFinite(value) || value <= 0) {
    return null
  }

  return value
}
