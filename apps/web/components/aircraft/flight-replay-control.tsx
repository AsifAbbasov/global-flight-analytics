'use client'

import { getRequestErrorMessage } from '@/lib/api/client'
import {
  buildFlightReplayAnalyticsSummary,
  buildFlightReplayGapSummary,
  flightReplaySpeeds,
  type FlightReplayAnalyticsSummary,
  type FlightReplayGapSummary,
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
  const canPlay = pointCount > 1
  const observedAt = currentPoint
    ? formatObservedAt(currentPoint.observed_at)
    : null
  const altitude = currentPoint ? formatObservedAltitude(currentPoint) : null
  const gapSummary = buildFlightReplayGapSummary(replay, cursorIndex)
  const analyticsSummary = buildFlightReplayAnalyticsSummary(replay)

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
              disabled={!canPlay}
              onClick={() => onPlayingChange(!isPlaying)}
              className='min-w-20 rounded-lg border border-sky-400/35 bg-sky-400/10 px-3 py-2 text-xs font-bold text-sky-100 hover:bg-sky-400/15 disabled:cursor-not-allowed disabled:opacity-40'
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
                  disabled={!canPlay}
                  onClick={() => onSpeedChange(candidate)}
                  className={`rounded-md border px-2.5 py-2 text-xs font-bold disabled:cursor-not-allowed disabled:opacity-40 ${
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
            disabled={!canPlay}
            onChange={event => {
              onPlayingChange(false)
              onCursorIndexChange(Number(event.target.value))
            }}
            aria-label='Historical replay position'
            className='w-full accent-sky-400 disabled:opacity-50'
          />

          {replay ? (
            <ReplayEvidenceTimeline
              replay={replay}
              cursorIndex={cursorIndex}
              gapSummary={gapSummary}
              onSelectSample={nextIndex => {
                onPlayingChange(false)
                onCursorIndexChange(nextIndex)
              }}
            />
          ) : null}

          <ReplayAnalyticsSummary summary={analyticsSummary} />

          <div className='grid gap-2 text-xs sm:grid-cols-3 xl:grid-cols-5'>
            <ReplayDatum label='Observed at' value={observedAt ?? 'Unavailable'} />
            <ReplayDatum label='Position' value={formatPosition(currentPoint)} />
            <ReplayDatum label='Altitude' value={altitude ?? 'Unavailable'} />
            <ReplayDatum label='Velocity' value={formatVelocity(currentPoint.velocity_mps)} />
            <ReplayDatum label='Heading' value={formatHeading(currentPoint.heading_degrees)} />
            <ReplayDatum
              label='Vertical rate'
              value={formatVerticalRate(currentPoint.vertical_rate_mps)}
            />
            <ReplayDatum
              label='Flight state'
              value={currentPoint.on_ground ? 'On ground' : 'Airborne'}
            />
            <ReplayDatum
              label='Aircraft origin country'
              value={currentPoint.origin_country.trim() || 'Unspecified'}
            />
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

function ReplayEvidenceTimeline({
  replay,
  cursorIndex,
  gapSummary,
  onSelectSample,
}: {
  replay: FlightReplay
  cursorIndex: number
  gapSummary: FlightReplayGapSummary
  onSelectSample: (cursorIndex: number) => void
}) {
  if (replay.points.length <= 1) {
    return (
      <div
        className='rounded-lg border border-white/10 bg-slate-950/50 px-3 py-2'
        aria-label='Historical evidence gap timeline'
      >
        <p className='text-[10px] font-bold uppercase tracking-[0.12em] text-slate-500'>
          Evidence gaps
        </p>
        <p className='mt-1 text-xs text-slate-400'>
          Only one persisted observation is available, so no inter-sample interval can be measured.
        </p>
      </div>
    )
  }

  return (
    <div
      className='rounded-lg border border-white/10 bg-slate-950/50 px-3 py-2.5'
      aria-label='Historical evidence gap timeline'
    >
      <div className='flex items-center justify-between gap-2'>
        <p className='text-[10px] font-bold uppercase tracking-[0.12em] text-slate-500'>
          Evidence gaps
        </p>
        <p className='font-mono text-[10px] text-slate-500'>
          largest {formatDuration(gapSummary.largestGapSeconds)} · span{' '}
          {formatDuration(gapSummary.totalObservedSpanSeconds)}
        </p>
      </div>

      <div className='mt-3 flex min-h-6 items-center' role='group' aria-label='Observed replay samples'>
        <ReplaySampleButton
          sampleIndex={0}
          observedAt={replay.points[0]?.observed_at ?? ''}
          selected={cursorIndex === 0}
          onSelectSample={onSelectSample}
        />
        {gapSummary.gaps.map(gap => {
          const observedAt = replay.points[gap.toIndex]?.observed_at ?? gap.endObservedAt
          return (
            <div
              key={`${gap.fromIndex}-${gap.toIndex}-${gap.startObservedAt}`}
              className='flex min-w-0 items-center'
              style={{ flexBasis: 0, flexGrow: Math.max(1, gap.durationSeconds) }}
              title={`No persisted intermediate position for ${formatDuration(gap.durationSeconds)}`}
            >
              <span className='w-full border-t border-dashed border-amber-300/45' />
              <ReplaySampleButton
                sampleIndex={gap.toIndex}
                observedAt={observedAt}
                selected={cursorIndex === gap.toIndex}
                onSelectSample={onSelectSample}
              />
            </div>
          )
        })}
      </div>

      <div className='mt-2 grid gap-1 text-[11px] text-slate-500 sm:grid-cols-2'>
        <p>
          Previous unobserved interval:{' '}
          <span className='font-mono text-slate-300'>
            {gapSummary.previousGapSeconds === null
              ? 'none'
              : formatDuration(gapSummary.previousGapSeconds)}
          </span>
        </p>
        <p>
          Next unobserved interval:{' '}
          <span className='font-mono text-slate-300'>
            {gapSummary.nextGapSeconds === null
              ? 'none'
              : formatDuration(gapSummary.nextGapSeconds)}
          </span>
        </p>
      </div>
    </div>
  )
}

function ReplayAnalyticsSummary({
  summary,
}: {
  summary: FlightReplayAnalyticsSummary
}) {
  return (
    <div
      className='rounded-lg border border-white/10 bg-slate-950/50 px-3 py-2.5'
      aria-label='Observed replay analytics'
      data-flight-replay-analytics='observed-samples-only'
    >
      <div className='flex flex-wrap items-start justify-between gap-2'>
        <div>
          <p className='text-[10px] font-bold uppercase tracking-[0.12em] text-slate-500'>
            Observed replay analytics
          </p>
          <p className='mt-1 text-[11px] leading-relaxed text-slate-500'>
            Aggregates use persisted samples only. No values are inferred between observations.
          </p>
        </div>
        <span className='rounded-full border border-emerald-400/25 bg-emerald-400/10 px-2 py-0.5 text-[9px] font-bold uppercase tracking-[0.1em] text-emerald-200'>
          Evidence-derived
        </span>
      </div>

      <div className='mt-2 grid gap-2 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-5'>
        <ReplayDatum label='Samples' value={String(summary.sampleCount)} />
        <ReplayDatum label='Observed span' value={formatDuration(summary.observedSpanSeconds)} />
        <ReplayDatum
          label='Median sample gap'
          value={summary.medianGapSeconds === null ? 'Unavailable' : formatDuration(summary.medianGapSeconds)}
        />
        <ReplayDatum
          label='Largest sample gap'
          value={summary.sampleCount <= 1 ? 'Unavailable' : formatDuration(summary.largestGapSeconds)}
        />
        <ReplayDatum
          label='Altitude coverage'
          value={`${summary.altitudeCoveragePercent}%`}
        />
        <ReplayDatum
          label='Observed altitude range'
          value={formatAltitudeRange(summary.minObservedAltitudeM, summary.maxObservedAltitudeM)}
        />
        <ReplayDatum
          label='Peak observed velocity'
          value={summary.peakVelocityMPS === null ? 'Unavailable' : formatVelocity(summary.peakVelocityMPS)}
        />
        <ReplayDatum
          label='Max observed climb'
          value={summary.maxClimbRateMPS === null ? 'None observed' : formatVerticalRate(summary.maxClimbRateMPS)}
        />
        <ReplayDatum
          label='Steepest observed descent'
          value={summary.steepestDescentRateMPS === null ? 'None observed' : formatVerticalRate(summary.steepestDescentRateMPS)}
        />
        <ReplayDatum
          label='Airborne / ground samples'
          value={`${summary.airborneSampleCount} / ${summary.onGroundSampleCount}`}
        />
      </div>
    </div>
  )
}

function ReplaySampleButton({
  sampleIndex,
  observedAt,
  selected,
  onSelectSample,
}: {
  sampleIndex: number
  observedAt: string
  selected: boolean
  onSelectSample: (cursorIndex: number) => void
}) {
  return (
    <button
      type='button'
      aria-label={`Replay sample ${sampleIndex + 1} observed ${formatObservedAt(observedAt) ?? observedAt}`}
      aria-pressed={selected}
      onClick={() => onSelectSample(sampleIndex)}
      className={`h-3 w-3 shrink-0 rounded-full border ${
        selected
          ? 'border-sky-200 bg-sky-300 ring-2 ring-sky-300/25'
          : 'border-slate-500 bg-slate-800 hover:border-slate-300'
      }`}
    />
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

function formatPosition(point: FlightReplayPoint): string {
  return `${point.latitude.toFixed(4)}, ${point.longitude.toFixed(4)}`
}

function formatVelocity(value: number): string {
  return `${value.toFixed(1)} m/s · ${Math.round(value * 3.6)} km/h`
}

function formatHeading(value: number): string {
  return `${Math.round(value)}°`
}

function formatVerticalRate(value: number): string {
  const sign = value > 0 ? '+' : ''
  return `${sign}${value.toFixed(1)} m/s`
}

function formatAltitudeRange(minimum: number | null, maximum: number | null): string {
  if (minimum === null || maximum === null) return 'Unavailable'
  return `${Math.round(minimum).toLocaleString()}–${Math.round(maximum).toLocaleString()} m`
}

function formatDuration(value: number): string {
  const totalSeconds = Math.max(0, Math.round(value))
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60

  if (hours > 0) {
    return `${hours}h ${minutes}m ${seconds}s`
  }
  if (minutes > 0) {
    return `${minutes}m ${seconds}s`
  }
  return `${seconds}s`
}
