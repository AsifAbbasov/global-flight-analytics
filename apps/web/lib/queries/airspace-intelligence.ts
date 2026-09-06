'use client'

import { useQuery, type UseQueryResult } from '@tanstack/react-query'

import { APIRequestError } from '@/lib/api/client'
import { getAirspaceRegionAnalytics } from '@/lib/api/airspace-intelligence'
import { canRequestAirspaceRegionAnalytics } from '@/lib/airspace/airspace-intelligence-model'
import type { AirspaceRegionAnalyticsResponse } from '@/types/airspace-intelligence'

export const defaultAirspaceWindowSeconds = 300

const keys = {
  all: ['airspace-intelligence'] as const,
  byRequest: (regionCode: string, asOfTime: string | null, windowSeconds: number) =>
    [...keys.all, regionCode, asOfTime, windowSeconds] as const,
}

export function useAirspaceRegionAnalytics(
  regionCode: string,
  asOfTime: string | null,
  windowSeconds = defaultAirspaceWindowSeconds
): UseQueryResult<AirspaceRegionAnalyticsResponse, Error> {
  const normalizedRegionCode = regionCode.trim().toLowerCase()
  const enabled = canRequestAirspaceRegionAnalytics(
    normalizedRegionCode,
    asOfTime
  )

  return useQuery({
    queryKey: keys.byRequest(normalizedRegionCode, asOfTime, windowSeconds),
    queryFn: ({ signal }) => {
      if (!enabled || asOfTime === null) {
        throw new APIRequestError(
          'Airspace Intelligence requires a bounded region and an observed traffic timestamp.'
        )
      }
      return getAirspaceRegionAnalytics({
        regionCode: normalizedRegionCode,
        asOfTime,
        windowSeconds,
        signal,
      })
    },
    enabled,
    staleTime: 30_000,
    refetchInterval: enabled ? 60_000 : false,
    refetchIntervalInBackground: false,
    retry: (count, error) =>
      count < 2 &&
      (!(error instanceof APIRequestError) ||
        error.status === null ||
        error.status >= 500),
  })
}
