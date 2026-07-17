'use client'

import { useQuery, type UseQueryResult } from '@tanstack/react-query'
import { APIRequestError } from '@/lib/api/client'
import { getStabilityIntelligence } from '@/lib/api/stability-intelligence'
import type { StabilityIntelligenceResponse } from '@/types/stability-intelligence'

export const defaultStabilityDurationSeconds = 300

const keys = {
  all: ['stability-intelligence'] as const,
  byRequest: (
    trajectoryID: string | null,
    asOfTimes: string[],
    durationSeconds: number
  ) =>
    [
      ...keys.all,
      trajectoryID,
      asOfTimes.join(','),
      durationSeconds,
    ] as const,
}

export function useStabilityIntelligence(
  trajectoryID: string | null,
  asOfTimes: string[],
  durationSeconds = defaultStabilityDurationSeconds
): UseQueryResult<StabilityIntelligenceResponse, Error> {
  const normalizedTrajectoryID = normalize(trajectoryID)
  const normalizedAsOfTimes = normalizeTimes(asOfTimes)
  const enabled =
    normalizedTrajectoryID !== null && normalizedAsOfTimes.length >= 2

  return useQuery({
    queryKey: keys.byRequest(
      normalizedTrajectoryID,
      normalizedAsOfTimes,
      durationSeconds
    ),
    queryFn: ({ signal }) => {
      if (normalizedTrajectoryID === null || normalizedAsOfTimes.length < 2) {
        throw new APIRequestError(
          'Stability Intelligence requires a trajectory and at least two analytical timestamps.'
        )
      }

      return getStabilityIntelligence({
        trajectoryID: normalizedTrajectoryID,
        asOfTimes: normalizedAsOfTimes,
        durationSeconds,
        signal,
      })
    },
    enabled,
    staleTime: 60_000,
    refetchInterval: false,
    retry: (failureCount, error) =>
      failureCount < 2 &&
      (!(error instanceof APIRequestError) ||
        error.status === null ||
        error.status >= 500),
  })
}

export function buildStabilityAsOfTimes(
  startTime: string | null,
  endTime: string | null
): string[] {
  const startMilliseconds = parseTime(startTime)
  const endMilliseconds = parseTime(endTime)

  if (
    startMilliseconds === null ||
    endMilliseconds === null ||
    endMilliseconds <= startMilliseconds
  ) {
    return []
  }

  const candidateMilliseconds = [
    Math.max(startMilliseconds, endMilliseconds - 60_000),
    Math.max(startMilliseconds, endMilliseconds - 30_000),
    endMilliseconds,
  ]
  const uniqueMilliseconds = [...new Set(candidateMilliseconds)].sort(
    (left, right) => left - right
  )

  if (uniqueMilliseconds.length < 2) return []
  return uniqueMilliseconds.map((value) => new Date(value).toISOString())
}

function normalize(value: string | null): string | null {
  const result = value?.trim() ?? ''
  return result === '' ? null : result
}

function normalizeTimes(values: string[]): string[] {
  return values.map((value) => value.trim()).filter((value) => value !== '')
}

function parseTime(value: string | null): number | null {
  const normalized = value?.trim() ?? ''
  if (normalized === '') return null
  const result = Date.parse(normalized)
  return Number.isNaN(result) ? null : result
}
