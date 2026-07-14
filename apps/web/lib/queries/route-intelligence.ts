'use client'

import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { APIRequestError } from '@/lib/api/client'
import { processRouteIntelligence } from '@/lib/api/route-intelligence'
import type { RouteIntelligenceRecord } from '@/types/route-intelligence'

const refreshIntervalMilliseconds = 60_000
const keys = {
  all: ['route-intelligence'] as const,
  byTrajectory: (id: string | null) => [...keys.all, 'processed', normalize(id)] as const,
}

export function useProcessedRouteIntelligence(
  trajectoryID: string | null
): UseQueryResult<RouteIntelligenceRecord, Error> {
  const id = normalize(trajectoryID)
  return useQuery({
    queryKey: keys.byTrajectory(id),
    queryFn: ({ signal }: { signal: AbortSignal }) => {
      if (id === null) throw new APIRequestError('Trajectory identifier is not available.')
      return processRouteIntelligence(id, { signal })
    },
    enabled: id !== null,
    staleTime: 30_000,
    refetchInterval: id === null ? false : refreshIntervalMilliseconds,
    refetchIntervalInBackground: false,
    retry: (count: number, error: Error) => {
      if (count >= 2) return false
      return !(error instanceof APIRequestError) || error.status === null || error.status >= 500
    },
  })
}
function normalize(value: string | null): string | null { const v=value?.trim().toLowerCase()??''; return v===''?null:v }
