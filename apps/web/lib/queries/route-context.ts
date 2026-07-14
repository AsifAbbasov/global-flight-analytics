'use client'

import {
  useQuery,
  type UseQueryResult,
} from '@tanstack/react-query'

import { APIRequestError } from '@/lib/api/client'
import { getAircraftRouteContext } from '@/lib/api/route-context'
import type { AircraftRouteContext } from '@/types/route-context'

const routeContextRefreshIntervalMilliseconds = 60_000

const routeContextQueryKeys = {
  all: ['route-context'] as const,
  byAircraft: (icao24: string | null) =>
    [
      ...routeContextQueryKeys.all,
      'by-aircraft',
      normalizeICAO24(icao24),
    ] as const,
}

export function useAircraftRouteContext(
  icao24: string | null
): UseQueryResult<AircraftRouteContext, Error> {
  const normalizedICAO24 = normalizeICAO24(icao24)

  return useQuery({
    queryKey:
      routeContextQueryKeys.byAircraft(normalizedICAO24),
    queryFn: ({ signal }) => {
      if (normalizedICAO24 === null) {
        throw new APIRequestError(
          'Aircraft ICAO24 is not available.'
        )
      }

      return getAircraftRouteContext(normalizedICAO24, {
        signal,
      })
    },
    enabled: normalizedICAO24 !== null,
    staleTime: 30_000,
    refetchInterval:
      normalizedICAO24 === null
        ? false
        : routeContextRefreshIntervalMilliseconds,
    refetchIntervalInBackground: false,
    retry: shouldRetryRouteContextQuery,
  })
}

function shouldRetryRouteContextQuery(
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

function normalizeICAO24(
  value: string | null
): string | null {
  const normalized = value?.trim().toLowerCase() ?? ''

  return normalized === '' ? null : normalized
}
