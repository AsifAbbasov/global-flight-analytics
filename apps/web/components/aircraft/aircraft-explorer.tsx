'use client'

import { useMemo, useState, type ChangeEvent } from 'react'

import {
  buildAircraftExplorerModel,
  type AircraftExplorerSort,
} from '@/lib/traffic/aircraft-explorer-model'
import type { TrafficAircraft } from '@/types/traffic'

interface AircraftExplorerProps {
  aircraft: TrafficAircraft[]
  selectedAircraftICAO24: string | null
  isFetching: boolean
  onSelectAircraft: (icao24: string) => void
}

const visibleAircraftLimit = 100

export function AircraftExplorer({
  aircraft,
  selectedAircraftICAO24,
  isFetching,
  onSelectAircraft,
}: AircraftExplorerProps) {
  const [query, setQuery] = useState('')
  const [sort, setSort] = useState<AircraftExplorerSort>('recent')

  const model = useMemo(
    () =>
      buildAircraftExplorerModel(aircraft, {
        query,
        sort,
        limit: visibleAircraftLimit,
      }),
    [aircraft, query, sort]
  )

  return (
    <aside
      className='rounded-xl border border-slate-700 bg-slate-950/95 p-4'
      aria-labelledby='aircraft-explorer-title'
      aria-busy={isFetching}
    >
      <div className='flex flex-wrap items-start justify-between gap-3'>
        <div>
          <p className='text-xs font-semibold uppercase tracking-[0.18em] text-sky-300'>
            Regional traffic index
          </p>
          <h3
            id='aircraft-explorer-title'
            className='mt-2 text-lg font-semibold text-white'
          >
            Aircraft Explorer
          </h3>
          <p className='mt-1 text-xs leading-5 text-slate-400'>
            Search the active regional snapshot and select an aircraft to open
            its full intelligence record.
          </p>
        </div>
        {isFetching ? (
          <span className='text-xs text-sky-300'>Updating…</span>
        ) : null}
      </div>

      <div className='mt-4 grid gap-3 sm:grid-cols-[minmax(0,1fr)_180px] xl:grid-cols-1'>
        <div>
          <label
            htmlFor='aircraft-explorer-search'
            className='text-xs font-medium text-slate-300'
          >
            Search aircraft
          </label>
          <input
            id='aircraft-explorer-search'
            type='search'
            value={query}
            onChange={(event: ChangeEvent<HTMLInputElement>) => {
              setQuery(event.target.value)
            }}
            placeholder='Callsign, ICAO24, airline, model…'
            className='mt-1.5 w-full rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-white outline-none transition placeholder:text-slate-600 focus:border-sky-400 focus:ring-2 focus:ring-sky-400/20'
          />
        </div>

        <div>
          <label
            htmlFor='aircraft-explorer-sort'
            className='text-xs font-medium text-slate-300'
          >
            Sort by
          </label>
          <select
            id='aircraft-explorer-sort'
            value={sort}
            onChange={(event: ChangeEvent<HTMLSelectElement>) => {
              setSort(event.target.value as AircraftExplorerSort)
            }}
            className='mt-1.5 w-full rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-white outline-none focus:border-sky-400 focus:ring-2 focus:ring-sky-400/20'
          >
            <option value='recent'>Latest observation</option>
            <option value='callsign'>Callsign</option>
            <option value='altitude-descending'>Highest altitude</option>
            <option value='speed-descending'>Highest speed</option>
          </select>
        </div>
      </div>

      <dl className='mt-4 grid grid-cols-2 gap-2 text-xs sm:grid-cols-4 xl:grid-cols-2'>
        <Summary label='Matched' value={`${model.matchedCount}/${model.totalCount}`} />
        <Summary label='Airborne' value={String(model.airborneCount)} />
        <Summary label='On ground' value={String(model.groundCount)} />
        <Summary label='Altitude unknown' value={String(model.unknownAltitudeCount)} />
      </dl>

      {model.items.length > 0 ? (
        <ul
          className='mt-4 max-h-[34rem] space-y-2 overflow-y-auto pr-1'
          aria-label='Aircraft search results'
        >
          {model.items.map(item => {
            const normalizedICAO24 = item.icao24.trim().toLowerCase()
            const selected = normalizedICAO24 === selectedAircraftICAO24

            return (
              <li key={normalizedICAO24}>
                <button
                  type='button'
                  onClick={() => onSelectAircraft(normalizedICAO24)}
                  aria-pressed={selected}
                  className={`w-full rounded-lg border p-3 text-left transition ${
                    selected
                      ? 'border-sky-400 bg-sky-400/10'
                      : 'border-slate-800 bg-slate-900/70 hover:border-slate-600 hover:bg-slate-900'
                  }`}
                >
                  <div className='flex items-start justify-between gap-3'>
                    <div className='min-w-0'>
                      <p className='truncate text-sm font-semibold text-white'>
                        {item.callsign.trim() || item.icao24.toUpperCase()}
                      </p>
                      <p className='mt-0.5 font-mono text-[11px] uppercase tracking-wide text-slate-500'>
                        {item.icao24}
                      </p>
                    </div>
                    <span
                      className={`shrink-0 rounded-full px-2 py-1 text-[10px] font-semibold uppercase tracking-wide ${
                        item.on_ground
                          ? 'bg-amber-400/10 text-amber-200'
                          : 'bg-emerald-400/10 text-emerald-200'
                      }`}
                    >
                      {item.on_ground ? 'Ground' : 'Airborne'}
                    </span>
                  </div>

                  <p className='mt-2 truncate text-xs text-slate-400'>
                    {joinAircraftIdentity(item)}
                  </p>

                  <dl className='mt-3 grid grid-cols-3 gap-2 text-xs'>
                    <Detail label='Altitude' value={formatAltitude(item)} />
                    <Detail label='Speed' value={formatSpeed(item.velocity_mps)} />
                    <Detail label='Observed' value={formatObservedAt(item.observed_at)} />
                  </dl>
                </button>
              </li>
            )
          })}
        </ul>
      ) : (
        <p className='mt-4 rounded-lg border border-dashed border-slate-700 p-4 text-sm leading-6 text-slate-400'>
          {aircraft.length === 0
            ? 'No aircraft are available in the current regional snapshot.'
            : 'No aircraft match the current search.'}
        </p>
      )}

      {model.matchedCount > model.displayedCount ? (
        <p className='mt-3 text-xs leading-5 text-slate-500'>
          Showing the first {model.displayedCount} matching aircraft. Refine the
          search to narrow the result set.
        </p>
      ) : null}
    </aside>
  )
}

function Summary({ label, value }: { label: string; value: string }) {
  return (
    <div className='rounded-lg border border-slate-800 bg-slate-900/70 p-2.5'>
      <dt className='text-slate-500'>{label}</dt>
      <dd className='mt-1 font-semibold text-slate-200'>{value}</dd>
    </div>
  )
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className='text-[10px] uppercase tracking-wide text-slate-600'>
        {label}
      </dt>
      <dd className='mt-1 truncate text-slate-300'>{value}</dd>
    </div>
  )
}

function joinAircraftIdentity(item: TrafficAircraft): string {
  return [item.aircraft_model, item.airline, item.origin_country]
    .map(value => value.trim())
    .filter(Boolean)
    .join(' · ') || 'Aircraft identity unavailable'
}

function formatAltitude(item: TrafficAircraft): string {
  if (item.on_ground) return 'Ground'
  if (item.altitude_m === null) return formatStatus(item.altitude_status)

  return `${new Intl.NumberFormat(undefined, {
    maximumFractionDigits: 0,
  }).format(item.altitude_m)} m`
}

function formatSpeed(value: number): string {
  return `${new Intl.NumberFormat(undefined, {
    maximumFractionDigits: 0,
  }).format(value * 3.6)} km/h`
}

function formatObservedAt(value: string): string {
  const timestamp = Date.parse(value)
  if (!Number.isFinite(timestamp)) return 'Unknown'

  return new Intl.DateTimeFormat(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(new Date(timestamp))
}

function formatStatus(value: string): string {
  return value.replaceAll('_', ' ')
}
