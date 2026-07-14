'use client'

import { AnalyticalMetricCard } from '@/components/analytics/metric-card'
import { calculateRegionAreaSquareKilometers } from '@/lib/geo/region-area'
import {
  useAnalyticalActiveAircraft,
  useAnalyticalCoverageScore,
  useAnalyticalDataFreshness,
  useAnalyticalTrafficDensity,
} from '@/lib/queries/analytics'
import type {
  AnalyticalMetric,
  CoverageScoreMetricParameters,
  DataFreshnessMetricParameters,
} from '@/types/analytics'
import type { Region } from '@/types/region'

const analyticalWindowMinutes = 15
const analyticalResultLimit = 1000
const dataFreshnessMaximumAgeSeconds = 300

interface AnalyticsOverviewProps {
  selectedRegion: Region
}

export function AnalyticsOverview({
  selectedRegion,
}: AnalyticsOverviewProps) {
  const recentParameters = {
    windowMinutes: analyticalWindowMinutes,
    limit: analyticalResultLimit,
    regionCode: selectedRegion.code,
  }

  const activeAircraftQuery =
    useAnalyticalActiveAircraft(recentParameters)
  const trafficDensityQuery = useAnalyticalTrafficDensity(
    recentParameters
  )

  const coverageParameters = buildCoverageParameters(
    activeAircraftQuery.data
  )
  const freshnessParameters = buildFreshnessParameters(
    activeAircraftQuery.data
  )

  const coverageQuery = useAnalyticalCoverageScore(coverageParameters)
  const freshnessQuery = useAnalyticalDataFreshness(freshnessParameters)

  const regionArea = calculateRegionAreaSquareKilometers(
    selectedRegion.bounds
  )

  const analyticsAreFetching =
    activeAircraftQuery.isFetching ||
    trafficDensityQuery.isFetching ||
    coverageQuery.isFetching ||
    freshnessQuery.isFetching

  return (
    <section className='mt-8' aria-labelledby='analytics-overview-title'>
      <div className='flex flex-wrap items-end justify-between gap-4'>
        <div>
          <h2
            id='analytics-overview-title'
            className='text-xl font-semibold text-white'
          >
            Live Analytics — {selectedRegion.name}
          </h2>
          <p className='mt-2 max-w-3xl text-sm leading-6 text-slate-400'>
            Protected metrics, traffic map and globe now share one regional
            scope. Confidence, eligibility and limitations remain visible for
            every published value.
          </p>
          <p className='mt-1 text-xs text-slate-500'>
            Configured rectangular area:{' '}
            {regionArea === null
              ? 'unavailable'
              : formatArea(regionArea)}
          </p>
        </div>

        <button
          type='button'
          onClick={() => {
            void Promise.all([
              activeAircraftQuery.refetch(),
              trafficDensityQuery.refetch(),
              coverageParameters ? coverageQuery.refetch() : Promise.resolve(),
              freshnessParameters
                ? freshnessQuery.refetch()
                : Promise.resolve(),
            ])
          }}
          disabled={analyticsAreFetching}
          className='rounded-lg border border-slate-700 px-4 py-2 text-sm font-medium text-slate-200 transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60'
        >
          {analyticsAreFetching ? 'Refreshing…' : 'Refresh analytics'}
        </button>
      </div>

      <div className='mt-4 grid gap-4 md:grid-cols-2 xl:grid-cols-4'>
        <AnalyticalMetricCard
          title='Active Aircraft'
          description={`Unique ICAO24 aircraft observed in ${selectedRegion.name} during the last fifteen minutes.`}
          metric={activeAircraftQuery.data}
          isPending={activeAircraftQuery.isPending}
          error={activeAircraftQuery.error}
          onRetry={() => {
            void activeAircraftQuery.refetch()
          }}
          formatValue={formatInteger}
        />

        <AnalyticalMetricCard
          title='Traffic Density'
          description={`Eligible aircraft per square kilometre across the configured ${selectedRegion.name} bounds.`}
          metric={trafficDensityQuery.data}
          isPending={trafficDensityQuery.isPending}
          error={trafficDensityQuery.error}
          onRetry={() => {
            void trafficDensityQuery.refetch()
          }}
          formatValue={formatDensity}
        />

        <AnalyticalMetricCard
          title='Eligibility Coverage'
          description='Share of regional trajectory contributors accepted by the traffic-metric policy.'
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
          description='Freshness of the newest regional source observation against a five-minute maximum age.'
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
          emptyMessage='Freshness requires a regional source observation timestamp.'
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

function formatDensity(value: number): string {
  if (value === 0) {
    return '0 / km²'
  }

  return `${new Intl.NumberFormat(undefined, {
    maximumSignificantDigits: 4,
  }).format(value)} / km²`
}

function formatArea(value: number): string {
  return `${new Intl.NumberFormat(undefined, {
    maximumFractionDigits: 0,
  }).format(value)} km²`
}
