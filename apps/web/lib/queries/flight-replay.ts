'use client'

import { useQuery, type UseQueryResult } from '@tanstack/react-query'

import { APIRequestError } from '@/lib/api/client'
import { getTrajectoryFlightReplay } from '@/lib/api/flight-replay'
import type { FlightReplay } from '@/types/flight-replay'
import type { AircraftTrajectory } from '@/types/trajectory'

const flightReplayQueryKeys = {
  all: ['flight-replay'] as const,
  byTrajectory: (trajectory: AircraftTrajectory | undefined) =>
    [
      ...flightReplayQueryKeys.all,
      trajectory?.id ?? null,
      trajectory?.updated_at ?? null,
      trajectory?.flight_id ?? null,
    ] as const,
}

export function useTrajectoryFlightReplay(
  trajectory: AircraftTrajectory | undefined
): UseQueryResult<FlightReplay, Error> {
  const flightID = trajectory?.flight_id.trim() ?? ''

  return useQuery({
    queryKey: flightReplayQueryKeys.byTrajectory(trajectory),
    queryFn: ({ signal }) => {
      if (!trajectory || flightID === '') {
        throw new APIRequestError(
          'Historical replay requires a trajectory with a durable flight identifier.'
        )
      }

      return getTrajectoryFlightReplay(trajectory, { signal })
    },
    enabled: trajectory !== undefined && flightID !== '',
    staleTime: 60_000,
    retry: shouldRetryFlightReplayQuery,
  })
}

function shouldRetryFlightReplayQuery(
  failureCount: number,
  error: Error
): boolean {
  if (failureCount >= 2) return false
  if (error instanceof APIRequestError) {
    return error.status === null || error.status >= 500
  }
  return true
}
