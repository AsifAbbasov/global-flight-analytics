'use client'

import { AnalyticalMetricCard } from '@/components/analytics/metric-card'
import {
  useAnalyticalActiveAircraft,
  useAnalyticalCoverageScore,
  useAnalyticalDataFreshness,
} from '@/lib/queries/analytics'
import type {
  AnalyticalMetric,
  CoverageScoreMetricParameters,
  DataFreshnessMetricParameters,
} from '@/types/analytics'

const activeAircraftParameters = {
  windowMinutes: 15,
  limit: 1000,
} as const

const dataFreshnessMaximumAgeSeconds = 300

export function AnalyticsOverview() {
  const activeAircraftQuery = useAnalyticalActiveAircraft(
    activeAircraftParameters
  )

  const coverageParameters = buildCoverageParameters(
    activeAircraftQuery.data
  )
  const freshnessParameters = buildFreshnessParameters(
    activeAircraftQuery.data
  )

  const coverageQuery = useAnalyticalCoverageScore(coverageParameters)
  const freshnessQuery = useAnalyticalDataFreshness(freshnessParameters)

  return (
    <section className='mt-8' aria-labelledby='analytics-overview-title'>
      <div className='flex flex-wrap items-end justify-between gap-4'>
        <div>
          <h2
            id='analytics-overview-title'
            className='text-xl font-semibold text-white'
          >
            Live Analytics
          </h2>
          <p className='mt-2 max-w-3xl text-sm leading-6 text-slate-400'>
            Protected metrics expose calculation status, confidence,
            eligibility and limitations instead of presenting every number as
            equally reliable.
          </p>
        </div>

        <button
          type='button'
          onClick={() => {
            void Promise.all([
              activeAircraftQuery.refetch(),
              coverageParameters ? coverageQuery.refetch() : Promise.resolve(),
              freshnessParameters
                ? freshnessQuery.refetch()
                : Promise.resolve(),
            ])
          }}
          disabled={
            activeAircraftQuery.isFetching ||
            coverageQuery.isFetching ||
            freshnessQuery.isFetching
          }
          className='rounded-lg border border-slate-700 px-4 py-2 text-sm font-medium text-slate-200 transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60'
        >
          {activeAircraftQuery.isFetching ||
          coverageQuery.isFetching ||
          freshnessQuery.isFetching
            ? 'Refreshing…'
            : 'Refresh analytics'}
        </button>
      </div>

      <div className='mt-4 grid gap-4 lg:grid-cols-3'>
        <AnalyticalMetricCard
          title='Active Aircraft'
          description='Unique ICAO24 aircraft observed within the last fifteen minutes.'
          metric={activeAircraftQuery.data}
          isPending={activeAircraftQuery.isPending}
          error={activeAircraftQuery.error}
          onRetry={() => {
            void activeAircraftQuery.refetch()
          }}
          formatValue={formatInteger}
        />

        <AnalyticalMetricCard
          title='Eligibility Coverage'
          description='Share of unique trajectory contributors accepted by the traffic-metric policy.'
          metric={coverageQuery.data}
          isPending={
            activeAircraftQuery.isPending ||
            (coverageParameters !== null && coverageQuery.isPending)
          }
          error={activeAircraftQuery.error ?? coverageQuery.error}
          onRetry={() => {
            if (activeAircraftQuery.error) {
              void activeAircraftQuery.refetch()
              return
            }

            void coverageQuery.refetch()
          }}
          formatValue={formatRatio}
          emptyMessage='Eligibility coverage requires a usable Active Aircraft response.'
        />

        <AnalyticalMetricCard
          title='Observation Freshness'
          description='Freshness of the newest source observation against a five-minute maximum age.'
          metric={freshnessQuery.data}
          isPending={
            activeAircraftQuery.isPending ||
            (freshnessParameters !== null && freshnessQuery.isPending)
          }
          error={activeAircraftQuery.error ?? freshnessQuery.error}
          onRetry={() => {
            if (activeAircraftQuery.error) {
              void activeAircraftQuery.refetch()
              return
            }

            void freshnessQuery.refetch()
          }}
          formatValue={formatRatio}
          emptyMessage='Freshness requires a source observation timestamp.'
        />
      </div>
    </section>
  )
}

function buildCoverageParameters(
  metric: AnalyticalMetric<number> | undefined
): CoverageScoreMetricParameters | null {
  if (!metric || metric.scope.input_count <= 0) {
    return null
  }

  return {
    observedSamples: metric.scope.allowed_count,
    expectedSamples: metric.scope.input_count,
  }
}

function buildFreshnessParameters(
  metric: AnalyticalMetric<number> | undefined
): DataFreshnessMetricParameters | null {
  if (!metric) {
    return null
  }

  const observedAt = metric.sources.reduce<string | null>(
    (latestTimestamp, source) => {
      if (!source.observed_to) {
        return latestTimestamp
      }

      if (
        latestTimestamp === null ||
        Date.parse(source.observed_to) > Date.parse(latestTimestamp)
      ) {
        return source.observed_to
      }

      return latestTimestamp
    },
    null
  )

  if (observedAt === null) {
    return null
  }

  return {
    observedAt,
    maximumAgeSeconds: dataFreshnessMaximumAgeSeconds,
  }
}

function formatInteger(value: number): string {
  return new Intl.NumberFormat().format(value)
}

function formatRatio(value: number): string {
  return new Intl.NumberFormat(undefined, {
    style: 'percent',
    maximumFractionDigits: 1,
  }).format(value)
}
