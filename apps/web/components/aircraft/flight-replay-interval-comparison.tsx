'use client'

import { useState } from 'react'

import {
  buildFlightReplayObservedIntervalComparison,
  type FlightReplayObservedIntervalComparison as IntervalComparison,
} from '@/lib/replay/flight-replay-interval-model'
import type { FlightReplay } from '@/types/flight-replay'

export function FlightReplayIntervalComparison({
  replay,
}: {
  replay: FlightReplay
}) {
  const firstPointID = replay.points[0]?.id ?? 'none'
  const lastPointID = replay.points[replay.points.length - 1]?.id ?? 'none'

  return (
    <FlightReplayIntervalComparisonState
      key={`${replay.trajectory_id}:${firstPointID}:${lastPointID}:${replay.points.length}`}
      replay={replay}
    />
  )
}

function FlightReplayIntervalComparisonState({ replay }: { replay: FlightReplay }) {
  const [firstSelectedIndex, setFirstSelectedIndex] = useState(0)
  const [secondSelectedIndex, setSecondSelectedIndex] = useState(
    Math.max(0, replay.points.length - 1)
  )
  const comparison = buildFlightReplayObservedIntervalComparison(
    replay,
    firstSelectedIndex,
    secondSelectedIndex
  )

  return (
    <div
      className='mt-4 rounded-lg border border-sky-400/15 bg-slate-950/45 px-3 py-3'
      aria-label='Observed interval comparison'
      data-flight-replay-interval-evidence='observed-endpoints-only'
      data-flight-replay-interval-interpolation='none'
      data-flight-replay-path-distance='not-claimed'
    >
      <div className='flex flex-wrap items-start justify-between gap-3'>
        <div>
          <p className='text-[10px] font-bold uppercase tracking-[0.12em] text-slate-500'>
            Observed interval comparison
          </p>
          <p className='mt-1 max-w-3xl text-[11px] leading-relaxed text-slate-500'>
            Compare two persisted observations. Endpoint deltas are evidence-derived; nothing is
            inferred about the aircraft path or intermediate positions.
          </p>
        </div>
        <span className='rounded-full border border-sky-400/25 bg-sky-400/10 px-2 py-0.5 text-[9px] font-bold uppercase tracking-[0.1em] text-sky-200'>
          Observed endpoints
        </span>
      </div>

      {replay.points.length < 2 ? (
        <p className='mt-3 text-xs text-slate-400'>
          At least two persisted observations are required for interval comparison.
        </p>
      ) : (
        <>
          <div className='mt-3 grid gap-2 md:grid-cols-2'>
            <ObservationSelect
              label='Observation A'
              replay={replay}
              value={firstSelectedIndex}
              onChange={setFirstSelectedIndex}
            />
            <ObservationSelect
              label='Observation B'
              replay={replay}
              value={secondSelectedIndex}
              onChange={setSecondSelectedIndex}
            />
          </div>

          <p className='mt-2 text-[10px] leading-relaxed text-slate-600'>
            A and B may be selected in either order. Results are normalized chronologically from
            the earlier persisted observation to the later one.
          </p>

          {comparison ? (
            <ComparisonResult comparison={comparison} />
          ) : (
            <p className='mt-3 rounded-md border border-amber-300/20 bg-amber-300/5 px-3 py-2 text-xs text-amber-100'>
              Select two different persisted observations to compare an observed interval.
            </p>
          )}
        </>
      )}
    </div>
  )
}

function ObservationSelect({
  label,
  replay,
  value,
  onChange,
}: {
  label: string
  replay: FlightReplay
  value: number
  onChange: (value: number) => void
}) {
  return (
    <label className='rounded-lg border border-white/10 bg-[#111315] px-2.5 py-2'>
      <span className='text-[9px] font-bold uppercase tracking-[0.12em] text-slate-600'>
        {label}
      </span>
      <select
        value={value}
        onChange={event => onChange(Number(event.target.value))}
        className='mt-1 w-full bg-transparent font-mono text-[11px] text-slate-300 outline-none'
        aria-label={`${label} persisted replay observation`}
      >
        {replay.points.map((point, index) => (
          <option key={point.id} value={index} className='bg-slate-950 text-slate-200'>
            #{index + 1} · {formatTimestamp(point.observed_at)}
          </option>
        ))}
      </select>
    </label>
  )
}

function ComparisonResult({ comparison }: { comparison: IntervalComparison }) {
  return (
    <div className='mt-3'>
      <div className='grid gap-2 sm:grid-cols-2 lg:grid-cols-4'>
        <ComparisonDatum label='Observed interval' value={formatDuration(comparison.elapsedSeconds)} />
        <ComparisonDatum
          label='Endpoint displacement'
          value={formatDistance(comparison.endpointGreatCircleDisplacementM)}
        />
        <ComparisonDatum label='Evidence samples' value={String(comparison.observedSampleCount)} />
        <ComparisonDatum
          label='Intermediate samples'
          value={String(comparison.intermediateSampleCount)}
        />
        <ComparisonDatum
          label='Largest internal gap'
          value={formatDuration(comparison.largestGapSeconds)}
        />
        <ComparisonDatum
          label='Altitude delta'
          value={formatSignedMeters(comparison.altitudeDeltaM)}
        />
        <ComparisonDatum
          label='Velocity delta'
          value={formatSigned(comparison.velocityDeltaMPS, 'm/s')}
        />
        <ComparisonDatum
          label='Vertical-rate delta'
          value={formatSigned(comparison.verticalRateDeltaMPS, 'm/s')}
        />
        <ComparisonDatum
          label='Heading change'
          value={`${comparison.headingChangeDegrees.toFixed(0)}°`}
        />
        <ComparisonDatum
          label='Flight-state transition'
          value={`${formatGroundState(comparison.startOnGround)} → ${formatGroundState(comparison.endOnGround)}`}
        />
        <ComparisonDatum label='Earlier observation' value={formatTimestamp(comparison.startObservedAt)} />
        <ComparisonDatum label='Later observation' value={formatTimestamp(comparison.endObservedAt)} />
      </div>

      <p className='mt-3 text-[11px] leading-relaxed text-slate-500'>
        Endpoint displacement is the great-circle distance between the two persisted coordinates,
        not travelled path distance. The selected interval may contain unobserved time, and no
        intermediate coordinate, route, phase or intent is synthesized.
      </p>
    </div>
  )
}

function ComparisonDatum({ label, value }: { label: string; value: string }) {
  return (
    <div className='rounded-lg border border-white/10 bg-[#111315] px-2.5 py-2'>
      <p className='text-[9px] font-bold uppercase tracking-[0.12em] text-slate-600'>{label}</p>
      <p className='mt-1 break-words font-mono text-[11px] text-slate-300'>{value}</p>
    </div>
  )
}

function formatTimestamp(value: string): string {
  const timestamp = Date.parse(value)
  if (Number.isNaN(timestamp)) return 'Unavailable'
  return new Date(timestamp).toISOString().replace('.000Z', 'Z')
}

function formatDuration(value: number): string {
  const totalSeconds = Math.max(0, Math.round(value))
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60

  if (hours > 0) return `${hours}h ${minutes}m ${seconds}s`
  if (minutes > 0) return `${minutes}m ${seconds}s`
  return `${seconds}s`
}

function formatDistance(valueM: number): string {
  if (!Number.isFinite(valueM)) return 'Unavailable'
  if (valueM >= 1000) return `${(valueM / 1000).toFixed(2)} km`
  return `${Math.round(valueM)} m`
}

function formatSignedMeters(value: number | null): string {
  if (value === null || !Number.isFinite(value)) return 'Unavailable'
  return `${value > 0 ? '+' : ''}${Math.round(value).toLocaleString()} m`
}

function formatSigned(value: number, unit: string): string {
  if (!Number.isFinite(value)) return 'Unavailable'
  return `${value > 0 ? '+' : ''}${value.toFixed(1)} ${unit}`
}

function formatGroundState(onGround: boolean): string {
  return onGround ? 'On ground' : 'Airborne'
}
