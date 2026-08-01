// FRONTEND_UNIFIED_AIRPORT_ANALYTICS_WORKSPACE_V1
'use client'

import { useQuery, type UseQueryResult } from '@tanstack/react-query'

import {
  getAirportIntelligenceHistory,
  getAirportIntelligenceOverview,
  getAirportIntelligenceRanking,
  getAirportIntelligenceTrends,
} from '@/lib/api/airport-intelligence'
import { APIRequestError } from '@/lib/api/client'
import type {
  AirportIntelligenceHistory,
  AirportIntelligenceOverview,
  AirportIntelligenceRanking,
  AirportIntelligenceTrends,
} from '@/types/airport-intelligence'

const staleTimeMilliseconds = 60_000

const keys = {
  all: ['airport-intelligence'] as const,
  ranking: (days: number, limit: number) =>
    [...keys.all, 'ranking', days, limit] as const,
  overview: (icaoCode: string | null, days: number) =>
    [...keys.all, 'overview', icaoCode, days] as const,
  history: (icaoCode: string | null, days: number) =>
    [...keys.all, 'history', icaoCode, days] as const,
  trends: (icaoCode: string | null, days: number) =>
    [...keys.all, 'trends', icaoCode, days] as const,
}

export function useAirportIntelligenceRanking(
  days: number,
  limit: number
): UseQueryResult<AirportIntelligenceRanking, Error> {
  return useQuery({
    queryKey: keys.ranking(days, limit),
    queryFn: ({ signal }: { signal: AbortSignal }) =>
      getAirportIntelligenceRanking({ days, limit, signal }),
    staleTime: staleTimeMilliseconds,
    retry: shouldRetry,
  })
}

export function useAirportIntelligenceOverview(
  icaoCode: string | null,
  days: number
): UseQueryResult<AirportIntelligenceOverview, Error> {
  const normalizedICAO = normalizeICAOCode(icaoCode)
  return useQuery({
    queryKey: keys.overview(normalizedICAO, days),
    queryFn: ({ signal }: { signal: AbortSignal }) => {
      if (normalizedICAO === null) {
        throw new APIRequestError('Airport ICAO code is unavailable.')
      }
      return getAirportIntelligenceOverview(normalizedICAO, { days, signal })
    },
    enabled: normalizedICAO !== null,
    staleTime: staleTimeMilliseconds,
    retry: shouldRetry,
  })
}

export function useAirportIntelligenceHistory(
  icaoCode: string | null,
  days: number
): UseQueryResult<AirportIntelligenceHistory, Error> {
  const normalizedICAO = normalizeICAOCode(icaoCode)
  return useQuery({
    queryKey: keys.history(normalizedICAO, days),
    queryFn: ({ signal }: { signal: AbortSignal }) => {
      if (normalizedICAO === null) {
        throw new APIRequestError('Airport ICAO code is unavailable.')
      }
      return getAirportIntelligenceHistory(normalizedICAO, { days, signal })
    },
    enabled: normalizedICAO !== null,
    staleTime: staleTimeMilliseconds,
    retry: shouldRetry,
  })
}

export function useAirportIntelligenceTrends(
  icaoCode: string | null,
  days: number
): UseQueryResult<AirportIntelligenceTrends, Error> {
  const normalizedICAO = normalizeICAOCode(icaoCode)
  return useQuery({
    queryKey: keys.trends(normalizedICAO, days),
    queryFn: ({ signal }: { signal: AbortSignal }) => {
      if (normalizedICAO === null) {
        throw new APIRequestError('Airport ICAO code is unavailable.')
      }
      return getAirportIntelligenceTrends(normalizedICAO, { days, signal })
    },
    enabled: normalizedICAO !== null,
    staleTime: staleTimeMilliseconds,
    retry: shouldRetry,
  })
}

function normalizeICAOCode(value: string | null): string | null {
  const normalized = value?.trim().toUpperCase() ?? ''
  return /^[A-Z0-9]{4}$/.test(normalized) ? normalized : null
}

function shouldRetry(failureCount: number, error: Error): boolean {
  if (failureCount >= 2) return false
  if (error instanceof APIRequestError) {
    return error.status === null || error.status >= 500
  }
  return true
}
