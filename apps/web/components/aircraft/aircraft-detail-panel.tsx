'use client'

import { getRequestErrorMessage } from '@/lib/api/client'
import { useAircraftProfile } from '@/lib/queries/aircraft'
import type { TrafficAircraft } from '@/types/traffic'

interface AircraftDetailPanelProps {
  selectedICAO24: string | null
  aircraft: TrafficAircraft | undefined
  onClose: () => void
}

export function AircraftDetailPanel({
  selectedICAO24,
  aircraft,
  onClose,
}: AircraftDetailPanelProps) {
  const profileQuery = useAircraftProfile(selectedICAO24)

  if (selectedICAO24 === null) {
    return (
      <aside className='flex min-h-[360px] items-center justify-center rounded-xl border border-dashed border-slate-700 bg-slate-950/70 p-6 text-center'>
        <div>
          <p className='text-sm font-semibold text-slate-200'>
            No aircraft selected
          </p>
          <p className='mt-2 text-sm leading-6 text-slate-400'>
            Select an aircraft marker on the map to inspect its live
            observation and registered profile.
          </p>
        </div>
      </aside>
    )
  }

  const displayName =
    aircraft?.callsign.trim() ||
    profileQuery.data?.registration.trim() ||
    selectedICAO24.toUpperCase()

  return (
    <aside
      className='rounded-xl border border-slate-700 bg-slate-950/95 p-5'
      aria-labelledby='aircraft-detail-title'
    >
      <div className='flex items-start justify-between gap-4'>
        <div>
          <p className='text-xs font-semibold uppercase tracking-[0.18em] text-sky-300'>
            Aircraft detail
          </p>
          <h3
            id='aircraft-detail-title'
            className='mt-2 text-xl font-semibold text-white'
          >
            {displayName}
          </h3>
          <p className='mt-1 font-mono text-xs uppercase text-slate-400'>
            ICAO24 {selectedICAO24}
          </p>
        </div>

        <button
          type='button'
          onClick={onClose}
          className='rounded-md border border-slate-700 px-3 py-1.5 text-sm font-medium text-slate-300 transition hover:bg-slate-800'
          aria-label='Close aircraft details'
        >
          Close
        </button>
      </div>

      <section className='mt-5' aria-labelledby='live-observation-title'>
        <h4
          id='live-observation-title'
          className='text-sm font-semibold text-slate-200'
        >
          Live observation
        </h4>

        {aircraft ? (
          <dl className='mt-3 grid grid-cols-2 gap-x-4 gap-y-3 text-sm'>
            <Detail label='Callsign' value={aircraft.callsign} />
            <Detail
              label='Status'
              value={aircraft.on_ground ? 'On ground' : 'In air'}
            />
            <Detail
              label='Altitude'
              value={formatAltitude(aircraft.altitude_m)}
            />
            <Detail
              label='Speed'
              value={formatSpeed(aircraft.velocity_mps)}
            />
            <Detail
              label='Heading'
              value={formatHeading(aircraft.heading_degrees)}
            />
            <Detail
              label='Origin country'
              value={aircraft.origin_country}
            />
            <Detail
              label='Latitude'
              value={formatCoordinate(aircraft.latitude)}
            />
            <Detail
              label='Longitude'
              value={formatCoordinate(aircraft.longitude)}
            />
            <div className='col-span-2'>
              <Detail
                label='Observed'
                value={formatTimestamp(aircraft.observed_at)}
              />
            </div>
          </dl>
        ) : (
          <p className='mt-3 rounded-lg border border-amber-400/30 bg-amber-400/10 p-3 text-sm leading-6 text-amber-100'>
            This aircraft is no longer present in the latest regional
            traffic response. Its registered profile remains available
            below.
          </p>
        )}
      </section>

      <section
        className='mt-6 border-t border-slate-800 pt-5'
        aria-labelledby='registered-profile-title'
      >
        <div className='flex items-center justify-between gap-3'>
          <h4
            id='registered-profile-title'
            className='text-sm font-semibold text-slate-200'
          >
            Registered profile
          </h4>

          {profileQuery.isFetching ? (
            <span className='text-xs text-sky-300'>Loading…</span>
          ) : null}
        </div>

        {profileQuery.data ? (
          <dl className='mt-3 grid grid-cols-2 gap-x-4 gap-y-3 text-sm'>
            <Detail
              label='Registration'
              value={profileQuery.data.registration}
            />
            <Detail
              label='Type'
              value={profileQuery.data.aircraft_type}
            />
            <Detail
              label='Manufacturer'
              value={profileQuery.data.manufacturer}
            />
            <Detail label='Model' value={profileQuery.data.model} />
            <Detail
              label='Airline'
              value={profileQuery.data.airline}
            />
            <Detail
              label='Country'
              value={profileQuery.data.country}
            />
          </dl>
        ) : null}

        {profileQuery.isPending ? (
          <p className='mt-3 text-sm text-slate-400'>
            Loading the registered aircraft profile…
          </p>
        ) : null}

        {profileQuery.error ? (
          <div className='mt-3 rounded-lg border border-amber-400/30 bg-amber-400/10 p-3'>
            <p className='text-sm leading-6 text-amber-100'>
              {getRequestErrorMessage(profileQuery.error)}
            </p>
            <button
              type='button'
              onClick={() => {
                void profileQuery.refetch()
              }}
              disabled={profileQuery.isFetching}
              className='mt-3 rounded-md border border-amber-300/40 px-3 py-1.5 text-sm font-medium text-amber-100 transition hover:bg-amber-300/10 disabled:cursor-not-allowed disabled:opacity-60'
            >
              Retry profile
            </button>
          </div>
        ) : null}
      </section>
    </aside>
  )
}

interface DetailProps {
  label: string
  value: string
}

function Detail({ label, value }: DetailProps) {
  return (
    <div>
      <dt className='text-xs uppercase tracking-wide text-slate-500'>
        {label}
      </dt>
      <dd className='mt-1 break-words text-slate-200'>
        {value.trim() || 'Unknown'}
      </dd>
    </div>
  )
}

function formatAltitude(value: number): string {
  if (!Number.isFinite(value)) {
    return 'Unknown'
  }

  return `${new Intl.NumberFormat().format(Math.round(value))} m`
}

function formatSpeed(value: number): string {
  if (!Number.isFinite(value)) {
    return 'Unknown'
  }

  const kilometersPerHour = value * 3.6

  return `${new Intl.NumberFormat(undefined, {
    maximumFractionDigits: 0,
  }).format(kilometersPerHour)} km/h`
}

function formatHeading(value: number): string {
  if (!Number.isFinite(value)) {
    return 'Unknown'
  }

  const normalized = ((value % 360) + 360) % 360
  return `${Math.round(normalized)}°`
}

function formatCoordinate(value: number): string {
  if (!Number.isFinite(value)) {
    return 'Unknown'
  }

  return value.toFixed(5)
}

function formatTimestamp(value: string): string {
  const timestamp = new Date(value)

  if (Number.isNaN(timestamp.getTime())) {
    return 'Unknown'
  }

  return timestamp.toLocaleString()
}
