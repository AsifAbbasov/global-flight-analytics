'use client'

import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { APIRequestError } from '@/lib/api/client'
import { getETAReliability } from '@/lib/api/eta-reliability'
import type { ETAReliabilityResponse } from '@/types/eta-reliability'

const keys = {
  all: ['eta-reliability'] as const,
  byRequest: (id: string | null, asOf: string | null, duration: number) => [...keys.all, id, asOf, duration] as const,
}

export function useETAReliability(
  trajectoryID: string | null,
  asOfTime: string | null,
  durationSeconds: number,
  enabled: boolean
): UseQueryResult<ETAReliabilityResponse, Error> {
  const id = normalize(trajectoryID)
  const asOf = normalize(asOfTime)
  return useQuery({
    queryKey: keys.byRequest(id, asOf, durationSeconds),
    queryFn: ({ signal }) => {
      if (id === null || asOf === null) {
        throw new APIRequestError('ETA Reliability requires a trajectory and an analytical timestamp.')
      }
      return getETAReliability({ trajectoryID: id, asOfTime: asOf, durationSeconds, signal })
    },
    enabled: enabled && id !== null && asOf !== null,
    staleTime: 5 * 60_000,
    refetchInterval: false,
    refetchOnWindowFocus: false,
    retry: (count, error) => count < 1 && (!(error instanceof APIRequestError) || error.status === null || error.status >= 500),
  })
}

function normalize(value: string | null): string | null {
  const result = value?.trim() ?? ''
  return result === '' ? null : result
}
