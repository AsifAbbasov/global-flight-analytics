'use client'

import { getRequestErrorMessage } from '@/lib/api/client'
import {
  flightReplaySpeeds,
  type FlightReplaySpeed,
} from '@/lib/replay/flight-replay-model'
import type {
  FlightReplay,
  FlightReplayPoint,
} from '@/types/flight-replay'

interface FlightReplayControlProps {
  replay: FlightReplay | undefined
  currentPoint: FlightReplayPoint | null
  cursorIndex: number
  isPlaying: boolean
  speed: FlightReplaySpeed
  isPending: boolean
  isFetching: boolean
  error: Error | null
  unavailableReason: string | null
  onCursorIndexChange: (cursorIndex: number) => void
  onPlayingChange: (isPlaying: boolean) => void
  onSpeedChange: (speed: FlightReplaySpeed) => void
  onRetry: () => void
}

export function FlightReplayControl({
  replay,
  currentPoint,
  cursorIndex,
  isPlaying,
  speed,
  isPending,
  isFetching,
  error,
  unavailableReason,
  onCursorIndexChange,
  onPlayingChange,
  onSpeedChange,
  onRetry,
}: FlightReplayControlProps) {
  const pointCount = replay?.points.length ?? 0
  const hasReplay = pointCount > 0 && currentPoint !== null
  const observedAt = currentPoint
    ? formatObservedAt(currentPoint.observed_at)
    : null
  const altitude = currentPoint ? formatObservedAltitude(currentPoint) : null

  return (
    <section
      className='border-t border-white/10 bg-[#151719] px-3 py-3 sm:px-4'
      aria-label='Historical flight replay'
      data-flight-replay-evidence='observed-only'
      data-flight-replay-interpolation='none'
    >
      <div className='flex flex-wrap items-start justify-between gap-3'>
        <div>
          <h3 className='text-sm font-bold text-white'>Historical flight replay</h3>
          <div className='mt-1 flex flex-wrap gap-1.5'>
            <span className='rounded-full border border-emerald-400/30 bg-emerald-400/10 px-2 py-0.5 text-[10px] font-bold uppercase tracking-[0.1em] text-emerald-200'>
              Observed only
            </span>
            <span className='rounded-full border border-slate-600 bg-slate-900 px-2 py-0.5 text-[10px] font-bold uppercase tracking-[0.1em] text-slate-300'>
              No interpolation
            </span>
          </div>
        </div>
        {replay ? (
          <p className='font-mono text-[10px] uppercase tracking-[0.08em] text-slate-500'>
            {replay.icao24.toUpperCase()} · {replay.points.length} samples
          </p>
        ) : null}
      </div>

      {unavailableReason ? (
        <p className='mt-3 text-xs text-slate-400'>{unavailableReason}</p>
      ) : error ? (
        <div className='mt-3 flex flex-wrap items-center gap-2'>
          <p className='text-xs text-rose-300'>{getRequestErrorMessage(error)}</p>
          <button
            type='button'
            onClick={onRetry}
            className='rounded-md border border-slate-600 px-2 py-1 text-xs font-semibold text-slate-200 hover:bg-slate-900'
          >
            Retry
          </button>
        </div>
      ) : isPending && !replay ? (
        <p className='mt-3 text-xs text-slate-400'>Loading persisted observations…</p>
      ) : !hasReplay ? (
        <p className='mt-3 text-xs text-slate-400'>
          No persisted observed states exist inside this trajectory window.
        </p>
      ) : (
        <div className='mt-3 space-y-3'>
          <div className='flex flex-wrap items-center gap-2'>
            <button
              type='button'
              onClick={() => onPlayingChange(!isPlaying)}
              className='min-w-20 rounded-lg border border-sky-400/35 bg-sky-400/10 px-3 py-2 text-xs font-bold text-sky-100 hover:bg-sky-400/15'
            >
              {isPlaying ? 'Pause' : 'Play'}
            </button>
            <div
              className='flex items-center gap-1'
              role='group'
              aria-label='Historical replay speed'
            >
              {flightReplaySpeeds.map(candidate => (
                <button
                  key={candidate}
                  type='button'
                  aria-pressed={speed === candidate}
                  onClick={() => onSpeedChange(candidate)}
                  className={`rounded-md border px-2.5 py-2 text-xs font-bold ${
                    speed === candidate
                      ? 'border-amber-300/50 bg-amber-300/10 text-amber-100'
                      : 'border-slate-700 text-slate-400 hover:bg-slate-900'
                  }`}
                >
                  {candidate}x
                </button>
              ))}
            </div>
            <span className='ml-auto text-xs text-slate-400'>
              Sample {cursorIndex + 1} / {pointCount}
              {isFetching ? ' · refreshing' : ''}
            </span>
          </div>

          <input
            type='range'
            min={0}
            max={Math.max(0, pointCount - 1)}
            step={1}
            value={cursorIndex}
            onChange={event => {
              onPlayingChange(false)
              onCursorIndexChange(Number(event.target.value))
            }}
            aria-label='Historical replay position'
            className='w-full accent-sky-400'
          />

          <div className='grid gap-2 text-xs sm:grid-cols-3'>
            <ReplayDatum label='Observed at' value={observedAt ?? 'Unavailable'} />
            <ReplayDatum label='Altitude' value={altitude ?? 'Unavailable'} />
            <ReplayDatum
              label='Source'
              value={currentPoint.source_name.trim() || 'Unspecified'}
            />
          </div>
        </div>
      )}

      <p className='mt-3 text-[11px] leading-relaxed text-slate-500'>
        Playback advances only across persisted observed flight states. Gaps are not filled,
        and positions between observations are not synthesized.
      </p>
    </section>
  )
}

function ReplayDatum({ label, value }: { label: string; value: string }) {
  return (
    <div className='rounded-lg border border-white/10 bg-slate-950/60 px-2.5 py-2'>
      <p className='text-[9px] font-bold uppercase tracking-[0.12em] text-slate-600'>
        {label}
      </p>
      <p className='mt-1 truncate font-mono text-[11px] text-slate-300'>{value}</p>
    </div>
  )
}

function formatObservedAt(value: string): string | null {
  const timestamp = Date.parse(value)
  if (Number.isNaN(timestamp)) return null
  return new Date(timestamp).toISOString().replace('.000Z', 'Z')
}

function formatObservedAltitude(point: FlightReplayPoint): string | null {
  if (
    point.barometric_altitude_m !== null &&
    (point.barometric_altitude_status === 'observed' ||
      point.barometric_altitude_status === 'ground')
  ) {
    return `${Math.round(point.barometric_altitude_m).toLocaleString()} m · barometric`
  }

  if (
    point.geometric_altitude_m !== null &&
    (point.geometric_altitude_status === 'observed' ||
      point.geometric_altitude_status === 'ground')
  ) {
    return `${Math.round(point.geometric_altitude_m).toLocaleString()} m · geometric`
  }

  return null
}
