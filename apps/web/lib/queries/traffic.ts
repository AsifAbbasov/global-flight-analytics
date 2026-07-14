'use client'

import {
  useQuery,
  type UseQueryResult,
} from '@tanstack/react-query'

import { APIRequestError } from '@/lib/api/client'
import { getCurrentTraffic } from '@/lib/api/traffic'
import type { TrafficAircraft } from '@/types/traffic'

const currentTrafficRefreshIntervalMilliseconds = 60_000

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
}

export function useCurrentTraffic(
  regionCode: string,
  options: UseCurrentTrafficOptions = {}
): UseQueryResult<TrafficAircraft[], Error> {
  const normalizedRegionCode = normalizeRegionCode(regionCode)

  return useQuery({
    queryKey: trafficQueryKeys.current(normalizedRegionCode),
    queryFn: ({ signal }) =>
      getCurrentTraffic(normalizedRegionCode, {
        signal,
      }),
    enabled: normalizedRegionCode.length > 0,
    initialData: options.initialData,
    refetchInterval: currentTrafficRefreshIntervalMilliseconds,
    refetchIntervalInBackground: false,
    retry: shouldRetryTrafficQuery,
  })
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
