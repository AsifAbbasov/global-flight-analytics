'use client'

import {
  useQuery,
  type UseQueryResult,
} from '@tanstack/react-query'

import { APIRequestError } from '@/lib/api/client'
import {
  getAnalyticalActiveAircraft,
  getAnalyticalAirportActivity,
  getAnalyticalCoverageScore,
  getAnalyticalDataFreshness,
  getAnalyticalTrafficDensity,
} from '@/lib/api/analytics'
import type {
  AirportActivityMetricParameters,
  AnalyticalMetric,
  CoverageScoreMetricParameters,
  DataFreshnessMetricParameters,
  RecentTrajectoryMetricParameters,
  TrafficDensityMetricParameters,
} from '@/types/analytics'

const analyticalMetricQueryKeys = {
  all: ['analytical-metrics'] as const,
  activeAircraft: (parameters: RecentTrajectoryMetricParameters) =>
    [
      ...analyticalMetricQueryKeys.all,
      'active-aircraft',
      parameters.windowMinutes ?? null,
      parameters.limit ?? null,
      normalizeRegionCode(parameters.regionCode),
    ] as const,
  trafficDensity: (parameters: TrafficDensityMetricParameters) =>
    [
      ...analyticalMetricQueryKeys.all,
      'traffic-density',
      parameters.windowMinutes ?? null,
      parameters.limit ?? null,
      normalizeRegionCode(parameters.regionCode),
      parameters.areaSquareKilometers ?? null,
    ] as const,
  airportActivity: (parameters: AirportActivityMetricParameters) =>
    [
      ...analyticalMetricQueryKeys.all,
      'airport-activity',
      normalizeQueryIdentifiers(parameters.arrivalTrajectoryIDs),
      normalizeQueryIdentifiers(parameters.departureTrajectoryIDs),
    ] as const,
  coverageScore: (parameters: CoverageScoreMetricParameters | null) =>
    [
      ...analyticalMetricQueryKeys.all,
      'coverage-score',
      parameters?.observedSamples ?? null,
      parameters?.expectedSamples ?? null,
    ] as const,
  dataFreshness: (parameters: DataFreshnessMetricParameters | null) =>
    [
      ...analyticalMetricQueryKeys.all,
      'data-freshness',
      parameters?.observedAt ?? null,
      parameters?.maximumAgeSeconds ?? null,
    ] as const,
}

export function useAnalyticalActiveAircraft(
  parameters: RecentTrajectoryMetricParameters = {}
): UseQueryResult<AnalyticalMetric<number>, Error> {
  return useQuery({
    queryKey: analyticalMetricQueryKeys.activeAircraft(parameters),
    queryFn: ({ signal }) =>
      getAnalyticalActiveAircraft(parameters, {
        signal,
      }),
    refetchInterval: 60_000,
    retry: shouldRetryAnalyticalQuery,
  })
}

export function useAnalyticalTrafficDensity(
  parameters: TrafficDensityMetricParameters
): UseQueryResult<AnalyticalMetric<number>, Error> {
  return useQuery({
    queryKey: analyticalMetricQueryKeys.trafficDensity(parameters),
    queryFn: ({ signal }) =>
      getAnalyticalTrafficDensity(parameters, {
        signal,
      }),
    refetchInterval: 60_000,
    retry: shouldRetryAnalyticalQuery,
  })
}

export function useAnalyticalAirportActivity(
  parameters: AirportActivityMetricParameters
): UseQueryResult<AnalyticalMetric<number>, Error> {
  return useQuery({
    queryKey: analyticalMetricQueryKeys.airportActivity(parameters),
    queryFn: ({ signal }) =>
      getAnalyticalAirportActivity(parameters, {
        signal,
      }),
    retry: shouldRetryAnalyticalQuery,
  })
}

export function useAnalyticalCoverageScore(
  parameters: CoverageScoreMetricParameters | null
): UseQueryResult<AnalyticalMetric<number>, Error> {
  return useQuery({
    queryKey: analyticalMetricQueryKeys.coverageScore(parameters),
    queryFn: ({ signal }) => {
      if (parameters === null) {
        throw new APIRequestError(
          'Coverage score parameters are not available.'
        )
      }

      return getAnalyticalCoverageScore(parameters, {
        signal,
      })
    },
    enabled: parameters !== null,
    retry: shouldRetryAnalyticalQuery,
  })
}

export function useAnalyticalDataFreshness(
  parameters: DataFreshnessMetricParameters | null
): UseQueryResult<AnalyticalMetric<number>, Error> {
  return useQuery({
    queryKey: analyticalMetricQueryKeys.dataFreshness(parameters),
    queryFn: ({ signal }) => {
      if (parameters === null) {
        throw new APIRequestError(
          'Data freshness parameters are not available.'
        )
      }

      return getAnalyticalDataFreshness(parameters, {
        signal,
      })
    },
    enabled: parameters !== null,
    refetchInterval: parameters === null ? false : 60_000,
    retry: shouldRetryAnalyticalQuery,
  })
}

function shouldRetryAnalyticalQuery(
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

function normalizeQueryIdentifiers(
  values: string[] | undefined
): string {
  if (!values) {
    return ''
  }

  return [...new Set(values.map(value => value.trim()).filter(Boolean))]
    .sort()
    .join(',')
}

function normalizeRegionCode(value: string | undefined): string {
  return value?.trim().toLowerCase() ?? ''
}
