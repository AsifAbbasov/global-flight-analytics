'use client'

import { useQueries } from '@tanstack/react-query'

import { APIRequestError } from '@/lib/api/client'
import { getProjectionIntelligence } from '@/lib/api/projection-intelligence'
import {
  buildETAEvolutionSummary,
  selectETAEvolutionAsOfTimes,
  type ETAEvolutionInputPoint,
} from '@/lib/projection/eta-evolution-model'
import { defaultProjectionDurationSeconds } from '@/lib/queries/projection-intelligence'
import type { FlightReplay } from '@/types/flight-replay'

export function useETAEvolution(
  trajectoryID: string | null,
  replay: FlightReplay | undefined,
  durationSeconds = defaultProjectionDurationSeconds
) {
  const normalizedTrajectoryID = trajectoryID?.trim() || null
  const asOfTimes = selectETAEvolutionAsOfTimes(replay)

  const queries = useQueries({
    queries: asOfTimes.map(asOfTime => ({
      queryKey: [
        'eta-evolution',
        normalizedTrajectoryID,
        asOfTime,
        durationSeconds,
      ] as const,
      queryFn: ({ signal }: { signal: AbortSignal }) => {
        if (normalizedTrajectoryID === null) {
          throw new APIRequestError(
            'ETA Evolution requires a persisted trajectory identifier.'
          )
        }
        return getProjectionIntelligence({
          trajectoryID: normalizedTrajectoryID,
          asOfTime,
          durationSeconds,
          signal,
        })
      },
      enabled: normalizedTrajectoryID !== null,
      staleTime: Number.POSITIVE_INFINITY,
      refetchInterval: false as const,
      refetchOnWindowFocus: false,
      retry: (failureCount: number, error: Error) =>
        failureCount < 2 &&
        (!(error instanceof APIRequestError) ||
          error.status === null ||
          error.status >= 500),
    })),
  })

  const inputs: ETAEvolutionInputPoint[] = asOfTimes.map((asOfTime, index) => {
    const query = queries[index]
    return {
      asOfTime,
      ...(query?.data ? { result: query.data } : {}),
      ...(query?.error ? { error: query.error.message } : {}),
    }
  })

  return {
    asOfTimes,
    summary: buildETAEvolutionSummary(inputs),
    isPending: queries.some(query => query.isPending),
    isFetching: queries.some(query => query.isFetching),
    hasErrors: queries.some(query => query.error !== null),
    refetch: () => Promise.all(queries.map(query => query.refetch())),
  }
}
