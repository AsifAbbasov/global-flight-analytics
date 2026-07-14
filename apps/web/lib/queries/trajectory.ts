'use client'

import {
  useQuery,
  type UseQueryResult,
} from '@tanstack/react-query'

import { APIRequestError } from '@/lib/api/client'
import { getLatestAircraftTrajectory } from '@/lib/api/trajectory'
import type { AircraftTrajectory } from '@/types/trajectory'

const trajectoryRefreshIntervalMilliseconds = 60_000

const trajectoryQueryKeys = {
  all: ['trajectories'] as const,
  latestByAircraft: (icao24: string | null) =>
    [
      ...trajectoryQueryKeys.all,
      'latest-by-aircraft',
      normalizeICAO24(icao24),
    ] as const,
}

export function useLatestAircraftTrajectory(
  icao24: string | null
): UseQueryResult<AircraftTrajectory, Error> {
  const normalizedICAO24 = normalizeICAO24(icao24)

  return useQuery({
    queryKey:
      trajectoryQueryKeys.latestByAircraft(normalizedICAO24),
    queryFn: ({ signal }) => {
      if (normalizedICAO24 === null) {
        throw new APIRequestError(
          'Aircraft ICAO24 is not available.'
        )
      }

      return getLatestAircraftTrajectory(normalizedICAO24, {
        signal,
      })
    },
    enabled: normalizedICAO24 !== null,
    staleTime: 30_000,
    refetchInterval:
      normalizedICAO24 === null
        ? false
        : trajectoryRefreshIntervalMilliseconds,
    refetchIntervalInBackground: false,
    retry: shouldRetryTrajectoryQuery,
  })
}

function shouldRetryTrajectoryQuery(
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
