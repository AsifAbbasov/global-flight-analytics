'use client'

import {
  useQuery,
  type UseQueryResult,
} from '@tanstack/react-query'

import { APIRequestError } from '@/lib/api/client'
import { getCurrentTraffic } from '@/lib/api/traffic'
import { normalizeLiveTrafficRefreshInterval } from '@/lib/traffic/live-traffic-status-model'
import type { TrafficAircraft } from '@/types/traffic'

const trafficQueryKeys = {
  all: ['traffic'] as const,
  current: (regionCode: string) =>
    [
      ...trafficQueryKeys.all,
      'current',
      normalizeRegionCode(regionCode),
    ] as const,
}

interface UseCurrentTrafficOptions {
  initialData?: TrafficAircraft[]
  refreshIntervalMilliseconds?: number | false
}

export function useCurrentTraffic(
  regionCode: string,
  options: UseCurrentTrafficOptions = {}
): UseQueryResult<TrafficAircraft[], Error> {
  const normalizedRegionCode = normalizeRegionCode(regionCode)
  const refreshIntervalMilliseconds =
    options.refreshIntervalMilliseconds === false
      ? false
      : normalizeLiveTrafficRefreshInterval(
          options.refreshIntervalMilliseconds
        )
  const initialDataUpdatedAt =
    options.initialData === undefined
      ? undefined
      : resolveInitialTrafficUpdatedAt(options.initialData)

  return useQuery({
    queryKey: trafficQueryKeys.current(normalizedRegionCode),
    queryFn: ({ signal }) =>
      getCurrentTraffic(normalizedRegionCode, {
        signal,
      }),
    enabled: normalizedRegionCode.length > 0,
    initialData: options.initialData,
    initialDataUpdatedAt,
    refetchInterval: refreshIntervalMilliseconds,
    refetchIntervalInBackground: false,
    retry: shouldRetryTrafficQuery,
  })
}

function resolveInitialTrafficUpdatedAt(
  aircraft: TrafficAircraft[]
): number {
  let latestTimestamp = 0

  for (const item of aircraft) {
    const timestamp = Date.parse(item.observed_at)
    if (Number.isFinite(timestamp) && timestamp > latestTimestamp) {
      latestTimestamp = timestamp
    }
  }

  return latestTimestamp
}

function shouldRetryTrafficQuery(
  failureCount: number,
  error: Error
): boolean {
  if (failureCount >= 2) {
    return false
  }

  if (error instanceof APIRequestError) {
    return error.status === null || error.status >= 500
  }

  return true
}

function normalizeRegionCode(value: string): string {
  return value.trim().toLowerCase()
}
