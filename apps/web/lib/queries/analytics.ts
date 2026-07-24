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
    ] as const,
  airportActivity: (parameters: AirportActivityMetricParameters) =>
    [
      ...analyticalMetricQueryKeys.all,
      'airport-activity',
      parameters.windowMinutes ?? null,
      parameters.limit ?? null,
      normalizeRegionCode(parameters.regionCode),
      parameters.airportICAO.trim().toUpperCase(),
      parameters.radiusKilometers ?? null,
    ] as const,
  coverageScore: (parameters: CoverageScoreMetricParameters) =>
    [
      ...analyticalMetricQueryKeys.all,
      'coverage-score',
      parameters.windowMinutes ?? null,
      normalizeRegionCode(parameters.regionCode),
    ] as const,
  dataFreshness: (parameters: DataFreshnessMetricParameters) =>
    [
      ...analyticalMetricQueryKeys.all,
      'data-freshness',
      parameters.windowMinutes ?? null,
      normalizeRegionCode(parameters.regionCode),
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
  parameters: CoverageScoreMetricParameters = {}
): UseQueryResult<AnalyticalMetric<number>, Error> {
  return useQuery({
    queryKey: analyticalMetricQueryKeys.coverageScore(parameters),
    queryFn: ({ signal }) =>
      getAnalyticalCoverageScore(parameters, {
        signal,
      }),
    refetchInterval: 60_000,
    retry: shouldRetryAnalyticalQuery,
  })
}

export function useAnalyticalDataFreshness(
  parameters: DataFreshnessMetricParameters = {}
): UseQueryResult<AnalyticalMetric<number>, Error> {
  return useQuery({
    queryKey: analyticalMetricQueryKeys.dataFreshness(parameters),
    queryFn: ({ signal }) =>
      getAnalyticalDataFreshness(parameters, {
        signal,
      }),
    refetchInterval: 60_000,
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

function normalizeRegionCode(value: string | undefined): string {
  return value?.trim().toLowerCase() ?? ''
}
