'use client'

import { AnalyticalMetricCard } from '@/components/analytics/metric-card'
import { calculateRegionAreaSquareKilometers } from '@/lib/geo/region-area'
import {
  useAnalyticalActiveAircraft,
  useAnalyticalCoverageScore,
  useAnalyticalDataFreshness,
  useAnalyticalTrafficDensity,
} from '@/lib/queries/analytics'
import type { Region } from '@/types/region'

const analyticalWindowMinutes = 15
const analyticalResultLimit = 1000

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
  const productionQualityParameters = {
    windowMinutes: analyticalWindowMinutes,
    regionCode: selectedRegion.code,
  }

  const activeAircraftQuery =
    useAnalyticalActiveAircraft(recentParameters)
  const trafficDensityQuery = useAnalyticalTrafficDensity(
    recentParameters
  )
  const coverageQuery = useAnalyticalCoverageScore(
    productionQualityParameters
  )
  const freshnessQuery = useAnalyticalDataFreshness(
    productionQualityParameters
  )

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
              coverageQuery.refetch(),
              freshnessQuery.refetch(),
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
          title='Observation Coverage'
          description='Share of ten-second observation intervals covered by retained regional trajectory evidence.'
          metric={coverageQuery.data}
          isPending={coverageQuery.isPending}
          error={coverageQuery.error}
          onRetry={() => {
            void coverageQuery.refetch()
          }}
          formatValue={formatRatio}
          emptyMessage='No retained regional trajectory observations were available for coverage calculation.'
        />

        <AnalyticalMetricCard
          title='Observation Freshness'
          description='Freshness of the newest retained regional observation against the server-owned five-minute stale threshold.'
          metric={freshnessQuery.data}
          isPending={freshnessQuery.isPending}
          error={freshnessQuery.error}
          onRetry={() => {
            void freshnessQuery.refetch()
          }}
          formatValue={formatRatio}
          emptyMessage='No retained regional observation timestamp was available.'
        />
      </div>
    </section>
  )
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
