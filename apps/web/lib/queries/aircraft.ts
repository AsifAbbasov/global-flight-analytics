'use client'

import {
  useQuery,
  type UseQueryResult,
} from '@tanstack/react-query'

import { APIRequestError } from '@/lib/api/client'
import { getAircraftProfile } from '@/lib/api/aircraft'
import type { AircraftProfile } from '@/types/aircraft'

const aircraftQueryKeys = {
  all: ['aircraft'] as const,
  profile: (icao24: string | null) =>
    [
      ...aircraftQueryKeys.all,
      'profile',
      normalizeICAO24(icao24),
    ] as const,
}

export function useAircraftProfile(
  icao24: string | null
): UseQueryResult<AircraftProfile, Error> {
  const normalizedICAO24 = normalizeICAO24(icao24)

  return useQuery({
    queryKey: aircraftQueryKeys.profile(normalizedICAO24),
    queryFn: ({ signal }) => {
      if (normalizedICAO24 === null) {
        throw new APIRequestError(
          'Aircraft ICAO24 is not available.'
        )
      }

      return getAircraftProfile(normalizedICAO24, {
        signal,
      })
    },
    enabled: normalizedICAO24 !== null,
    staleTime: 5 * 60_000,
    retry: shouldRetryAircraftQuery,
  })
}

function shouldRetryAircraftQuery(
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
