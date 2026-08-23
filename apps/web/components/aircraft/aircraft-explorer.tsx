'use client'

import { useMemo, useState, type ChangeEvent } from 'react'

import {
  buildAircraftExplorerModel,
  type AircraftExplorerMotionFilter,
  type AircraftExplorerSort,
} from '@/lib/traffic/aircraft-explorer-model'
import type { TrafficAircraft } from '@/types/traffic'

interface AircraftExplorerProps {
  aircraft: TrafficAircraft[]
  selectedAircraftICAO24: string | null
  isFetching: boolean
  onSelectAircraft: (icao24: string) => void
}

interface AircraftListDetail {
  key: string
  label: string
  value: string
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
  const [motion, setMotion] = useState<AircraftExplorerMotionFilter>('all')
  const [requireAltitudeEvidence, setRequireAltitudeEvidence] = useState(false)

  const model = useMemo(
    () =>
      buildAircraftExplorerModel(aircraft, {
        query,
        sort,
        motion,
        requireAltitudeEvidence,
        limit: visibleAircraftLimit,
      }),
    [aircraft, query, sort, motion, requireAltitudeEvidence]
  )

  const hasActiveFilters =
    query.trim().length > 0 || motion !== 'all' || requireAltitudeEvidence

  const resetFilters = () => {
    setQuery('')
    setMotion('all')
    setRequireAltitudeEvidence(false)
  }

  return (
    <aside
      className='rounded-xl border border-slate-700 bg-slate-950/95 p-3 sm:p-4'
      aria-labelledby='aircraft-explorer-title'
      aria-busy={isFetching}
    >
      <div className='flex flex-wrap items-start justify-between gap-3'>
        <div>
          <p className='text-[10px] font-semibold uppercase tracking-[0.18em] text-sky-300'>
            Live traffic index
          </p>
          <h3
            id='aircraft-explorer-title'
            className='mt-1.5 text-base font-semibold text-white'
          >
            Aircraft Explorer
          </h3>
          <p className='mt-1 text-xs leading-5 text-slate-400'>
            Only fields present in the current open-data traffic contract are
            rendered.
          </p>
        </div>
        {isFetching ? (
          <span className='text-xs text-sky-300'>Updating…</span>
        ) : null}
      </div>

      <div className='mt-3 grid gap-2 sm:grid-cols-[minmax(0,1fr)_160px] xl:grid-cols-1'>
        <div>
          <label
            htmlFor='aircraft-explorer-search'
            className='sr-only'
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
            placeholder='Search callsign, ICAO24, airline, model…'
            className='w-full rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-white outline-none transition placeholder:text-slate-600 focus:border-sky-400 focus:ring-2 focus:ring-sky-400/20'
          />
        </div>

        <div>
          <label htmlFor='aircraft-explorer-sort' className='sr-only'>
            Sort aircraft
          </label>
          <select
            id='aircraft-explorer-sort'
            value={sort}
            onChange={(event: ChangeEvent<HTMLSelectElement>) => {
              setSort(event.target.value as AircraftExplorerSort)
            }}
            className='w-full rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-white outline-none focus:border-sky-400 focus:ring-2 focus:ring-sky-400/20'
          >
            <option value='recent'>Latest observation</option>
            <option value='callsign'>Callsign</option>
            <option value='altitude-descending'>Highest altitude</option>
            <option value='speed-descending'>Highest speed</option>
          </select>
        </div>
      </div>

      <div className='mt-3 rounded-lg border border-slate-800 bg-slate-900/55 p-2.5'>
        <div className='flex flex-wrap items-center justify-between gap-2'>
          <div
            className='flex flex-wrap gap-1.5'
            role='group'
            aria-label='Aircraft motion filter'
          >
            {(
              [
                ['all', 'All'],
                ['airborne', 'Airborne'],
                ['ground', 'Ground'],
              ] as const
            ).map(([value, label]) => (
              <button
                key={value}
                type='button'
                onClick={() => setMotion(value)}
                aria-pressed={motion === value}
                className={`rounded-full border px-2.5 py-1 text-[11px] font-semibold transition ${
                  motion === value
                    ? 'border-amber-300/70 bg-amber-300/15 text-amber-100'
                    : 'border-slate-700 bg-slate-950/70 text-slate-400 hover:border-slate-500 hover:text-slate-200'
                }`}
              >
                {label}
              </button>
            ))}
          </div>

          <label className='flex cursor-pointer items-center gap-2 rounded-lg border border-slate-700 bg-slate-950/70 px-2.5 py-1.5 text-[11px] text-slate-300'>
            <input
              type='checkbox'
              checked={requireAltitudeEvidence}
              onChange={event => setRequireAltitudeEvidence(event.target.checked)}
              className='size-4 accent-amber-300'
            />
            Altitude only
          </label>
        </div>

        {hasActiveFilters ? (
          <div className='mt-2 flex items-center justify-between gap-3 border-t border-slate-800 pt-2'>
            <span className='text-[11px] text-slate-500'>Filters active</span>
            <button
              type='button'
              onClick={resetFilters}
              className='rounded-md border border-slate-700 px-2 py-1 text-[11px] font-medium text-slate-300 transition hover:border-slate-500 hover:bg-slate-800'
            >
              Reset
            </button>
          </div>
        ) : null}
      </div>

      <dl className='mt-3 grid grid-cols-2 gap-1.5 text-[11px] sm:grid-cols-4 xl:grid-cols-2'>
        <Summary label='Matched' value={`${model.matchedCount}/${model.totalCount}`} />
        <Summary label='Airborne' value={String(model.airborneCount)} />
        <Summary label='Ground' value={String(model.groundCount)} />
        <Summary label='Altitude gaps' value={String(model.unknownAltitudeCount)} />
      </dl>

      {model.items.length > 0 ? (
        <ul
          className='gfa-scrollbar mt-3 max-h-[calc(min(78vh,860px)-15rem)] min-h-64 space-y-1.5 overflow-y-auto pr-1'
          aria-label='Aircraft search results'
        >
          {model.items.map(item => {
            const normalizedICAO24 = item.icao24.trim().toLowerCase()
            const selected = normalizedICAO24 === selectedAircraftICAO24
            const identity = joinAircraftIdentity(item)
            const details = buildAircraftListDetails(item)

            return (
              <li key={normalizedICAO24}>
                <button
                  type='button'
                  onClick={() => onSelectAircraft(normalizedICAO24)}
                  aria-pressed={selected}
                  className={`w-full rounded-lg border px-3 py-2.5 text-left transition ${
                    selected
                      ? 'border-amber-300/70 bg-amber-300/10 shadow-lg shadow-amber-950/20'
                      : 'border-slate-800 bg-slate-900/70 hover:border-slate-600 hover:bg-slate-900'
                  }`}
                >
                  <div className='flex items-start justify-between gap-3'>
                    <div className='min-w-0'>
                      <p className='truncate text-sm font-semibold text-white'>
                        {item.callsign.trim() || item.icao24.toUpperCase()}
                      </p>
                      <p className='mt-0.5 font-mono text-[10px] uppercase tracking-wide text-slate-500'>
                        {item.icao24}
                      </p>
                    </div>
                    <span
                      className={`shrink-0 rounded-full px-2 py-1 text-[9px] font-semibold uppercase tracking-wide ${
                        item.on_ground
                          ? 'bg-amber-400/10 text-amber-200'
                          : 'bg-emerald-400/10 text-emerald-200'
                      }`}
                    >
                      {item.on_ground ? 'Ground' : 'Airborne'}
                    </span>
                  </div>

                  {identity ? (
                    <p className='mt-1.5 truncate text-[11px] text-slate-400'>
                      {identity}
                    </p>
                  ) : null}

                  {details.length > 0 ? (
                    <dl className='mt-2.5 grid grid-cols-2 gap-x-2 gap-y-2 text-[11px] sm:grid-cols-3 xl:grid-cols-2 2xl:grid-cols-3'>
                      {details.map(detail => (
                        <Detail
                          key={detail.key}
                          label={detail.label}
                          value={detail.value}
                        />
                      ))}
                    </dl>
                  ) : null}
                </button>
              </li>
            )
          })}
        </ul>
      ) : (
        <p className='mt-4 rounded-lg border border-dashed border-slate-700 p-4 text-sm leading-6 text-slate-400'>
          {aircraft.length === 0
            ? 'No aircraft are available in the current regional snapshot.'
            : 'No aircraft match the current search and filters.'}
        </p>
      )}

      {model.matchedCount > model.displayedCount ? (
        <p className='mt-3 text-xs leading-5 text-slate-500'>
          Showing the first {model.displayedCount} matching aircraft. Refine the
          search or filters to narrow the result set.
        </p>
      ) : null}
    </aside>
  )
}

function Summary({ label, value }: { label: string; value: string }) {
  return (
    <div className='rounded-md border border-slate-800 bg-slate-900/70 px-2.5 py-2'>
      <dt className='text-slate-500'>{label}</dt>
      <dd className='mt-0.5 font-semibold text-slate-200'>{value}</dd>
    </div>
  )
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className='text-[9px] uppercase tracking-wide text-slate-600'>
        {label}
      </dt>
      <dd className='mt-0.5 truncate font-semibold text-slate-300'>{value}</dd>
    </div>
  )
}

function joinAircraftIdentity(item: TrafficAircraft): string {
  return [item.aircraft_model, item.airline, item.origin_country]
    .map(value => value.trim())
    .filter(Boolean)
    .join(' · ')
}

function buildAircraftListDetails(item: TrafficAircraft): AircraftListDetail[] {
  const details: Array<AircraftListDetail | null> = [
    optionalDetail('altitude', 'Altitude', formatAltitude(item)),
    optionalDetail('speed', 'Speed', formatSpeed(item.velocity_mps)),
    optionalDetail('heading', 'Heading', formatHeading(item.heading_degrees)),
    optionalDetail('observed', 'Observed', formatObservedAt(item.observed_at)),
  ]

  return details.filter((detail): detail is AircraftListDetail => detail !== null)
}

function optionalDetail(
  key: string,
  label: string,
  value: string
): AircraftListDetail | null {
  return value ? { key, label, value } : null
}

function formatAltitude(item: TrafficAircraft): string {
  if (item.altitude_status === 'ground') return 'Ground'
  if (
    item.altitude_status !== 'observed' ||
    item.altitude_m === null ||
    !Number.isFinite(item.altitude_m)
  ) {
    return ''
  }

  return `${new Intl.NumberFormat(undefined, {
    maximumFractionDigits: 0,
  }).format(item.altitude_m)} m`
}

function formatSpeed(value: number): string {
  if (!Number.isFinite(value) || value < 0) return ''

  return `${new Intl.NumberFormat(undefined, {
    maximumFractionDigits: 0,
  }).format(value * 3.6)} km/h`
}

function formatHeading(value: number): string {
  if (!Number.isFinite(value)) return ''
  const normalized = ((value % 360) + 360) % 360
  return `${Math.round(normalized)}°`
}

function formatObservedAt(value: string): string {
  const timestamp = Date.parse(value)
  if (!Number.isFinite(timestamp)) return ''

  return new Intl.DateTimeFormat(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(new Date(timestamp))
}
