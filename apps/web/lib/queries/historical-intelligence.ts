// FRONTEND_HISTORICAL_ANALYTICS_COMPARISON_V1
'use client'

import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { APIRequestError } from '@/lib/api/client'
import {
  getHistoricalIntelligenceHistory,
  getLatestHistoricalIntelligence,
} from '@/lib/api/historical-intelligence'
import type {
  HistoricalIntelligenceAggregateHistory,
  HistoricalIntelligenceAggregateRecord,
  HistoricalIntelligenceSelection,
} from '@/types/historical-intelligence'

const keys = {
  all: ['historical-intelligence'] as const,
  latest: (selection: HistoricalIntelligenceSelection | null) => [
    ...keys.all,
    'latest',
    selectionKey(selection),
  ] as const,
  history: (selection: HistoricalIntelligenceSelection | null, limit: number) => [
    ...keys.all,
    'history',
    selectionKey(selection),
    limit,
  ] as const,
}

export function useLatestHistoricalIntelligence(
  selection: HistoricalIntelligenceSelection | null
): UseQueryResult<HistoricalIntelligenceAggregateRecord, Error> {
  return useQuery({
    queryKey: keys.latest(selection),
    queryFn: ({ signal }: { signal: AbortSignal }) => {
      if (selection === null) throw new APIRequestError('Historical selection is incomplete.')
      return getLatestHistoricalIntelligence(selection, { signal })
    },
    enabled: selection !== null,
    staleTime: 60_000,
    retry: retryHistoricalRequest,
  })
}

export function useHistoricalIntelligenceHistory(
  selection: HistoricalIntelligenceSelection | null,
  limit = 20
): UseQueryResult<HistoricalIntelligenceAggregateHistory, Error> {
  return useQuery({
    queryKey: keys.history(selection, limit),
    queryFn: ({ signal }: { signal: AbortSignal }) => {
      if (selection === null) throw new APIRequestError('Historical selection is incomplete.')
      return getHistoricalIntelligenceHistory(selection, { signal, limit })
    },
    enabled: selection !== null,
    staleTime: 60_000,
    retry: retryHistoricalRequest,
  })
}

function selectionKey(selection: HistoricalIntelligenceSelection | null): string {
  if (selection === null) return 'disabled'
  return [
    selection.scope,
    selection.metric,
    selection.granularity,
    selection.airportICAO ?? '',
    selection.originICAO ?? '',
    selection.destinationICAO ?? '',
  ].join('|')
}

function retryHistoricalRequest(count: number, error: Error): boolean {
  if (count >= 2) return false
  return !(error instanceof APIRequestError) || error.status === null || error.status >= 500
}
