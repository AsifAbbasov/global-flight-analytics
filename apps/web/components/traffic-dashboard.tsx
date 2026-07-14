'use client'

import { TrafficGlobe } from '@/components/globe/traffic-globe'
import { TrafficMap } from '@/components/map/traffic-map'
import { getRequestErrorMessage } from '@/lib/api/client'
import { useCurrentTraffic } from '@/lib/queries/traffic'
import type { Region } from '@/types/region'
import type { TrafficAircraft } from '@/types/traffic'

interface TrafficDashboardProps {
  regions: Region[]
  selectedRegion: Region
  onSelectedRegionCodeChange: (regionCode: string) => void
  initialTraffic: TrafficAircraft[]
  initialError: string | null
  regionsWarning: string | null
}

export function TrafficDashboard({
  regions,
  selectedRegion,
  onSelectedRegionCodeChange,
  initialTraffic,
  initialError,
  regionsWarning,
}: TrafficDashboardProps) {
  const initialData =
    selectedRegion.code === 'world' && initialError === null
      ? initialTraffic
      : undefined

  const trafficQuery = useCurrentTraffic(selectedRegion.code, {
    initialData,
  })

  const traffic = trafficQuery.data ?? []
  const isInitialLoading = trafficQuery.isPending
  const isRefreshing =
    trafficQuery.isFetching && !trafficQuery.isPending
  const errorMessage = trafficQuery.error
    ? getRequestErrorMessage(trafficQuery.error)
    : trafficQuery.isPending
      ? initialError
      : null

  return (
    <>
      <section className='mt-6 rounded-xl border border-slate-800 bg-slate-900 p-4'>
        <div className='flex flex-wrap items-end justify-between gap-4'>
          <div className='min-w-64 flex-1'>
            <label
              className='block text-sm font-medium text-slate-300'
              htmlFor='traffic-region'
            >
              Region
            </label>

            <select
              id='traffic-region'
              value={selectedRegion.code}
              onChange={event => {
                onSelectedRegionCodeChange(event.target.value)
              }}
              className='mt-2 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-white'
            >
              {regions.map(region => (
                <option key={region.code} value={region.code}>
                  {region.name}
                </option>
              ))}
            </select>
          </div>

          <button
            type='button'
            onClick={() => {
              void trafficQuery.refetch()
            }}
            disabled={trafficQuery.isFetching}
            className='rounded-lg border border-slate-700 px-4 py-2 text-sm font-medium text-slate-200 transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:opacity-60'
          >
            {trafficQuery.isFetching
              ? 'Refreshing traffic…'
              : 'Refresh traffic'}
          </button>
        </div>

        <div
          aria-live='polite'
          className='mt-3 flex flex-wrap items-center gap-3 text-sm'
        >
          <span className='text-slate-300'>
            Aircraft: {traffic.length}
          </span>

          <span className='text-slate-500'>
            View: {selectedRegion.name}
          </span>

          {trafficQuery.dataUpdatedAt > 0 ? (
            <span className='text-slate-500'>
              Updated {formatTimestamp(trafficQuery.dataUpdatedAt)}
            </span>
          ) : null}

          {regionsWarning ? (
            <span className='text-amber-300'>{regionsWarning}</span>
          ) : null}

          {isInitialLoading ? (
            <span className='text-sky-300'>
              Loading current traffic…
            </span>
          ) : null}

          {isRefreshing ? (
            <span className='text-sky-300'>
              Updating current traffic…
            </span>
          ) : null}

          {errorMessage ? (
            <>
              <span className='text-amber-300'>{errorMessage}</span>
              <button
                type='button'
                onClick={() => {
                  void trafficQuery.refetch()
                }}
                disabled={trafficQuery.isFetching}
                className='rounded-md border border-amber-400/50 px-3 py-1 font-medium text-amber-200 transition hover:bg-amber-400/10 disabled:cursor-not-allowed disabled:opacity-60'
              >
                Retry
              </button>
            </>
          ) : null}
        </div>
      </section>

      <div className='mt-4' aria-busy={trafficQuery.isFetching}>
        <TrafficGlobe
          aircraft={traffic}
          region={selectedRegion}
        />
      </div>

      <section className='mt-8 rounded-xl border border-slate-800 bg-slate-900 p-4 sm:p-6'>
        <h2 className='text-xl font-semibold'>
          Current Traffic — {selectedRegion.name}
        </h2>

        <div
          className='mt-4'
          aria-busy={trafficQuery.isFetching}
        >
          <TrafficMap
            aircraft={traffic}
            region={selectedRegion}
          />
        </div>

        {!trafficQuery.isFetching &&
        !errorMessage &&
        traffic.length === 0 ? (
          <p className='mt-4 text-sm text-slate-400'>
            No aircraft were returned for the selected region.
          </p>
        ) : null}
      </section>
    </>
  )
}

function formatTimestamp(value: number): string {
  return new Intl.DateTimeFormat(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(new Date(value))
}
