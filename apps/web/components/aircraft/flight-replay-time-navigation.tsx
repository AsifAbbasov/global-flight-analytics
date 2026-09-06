'use client'

import {
  buildFlightReplayTimeNavigation,
  type FlightReplayTimeFrame,
} from '@/lib/replay/flight-replay-model'
import type { FlightReplay } from '@/types/flight-replay'

interface FlightReplayTimeNavigationProps {
  replay: FlightReplay | undefined
  frame: FlightReplayTimeFrame
  isPlaying: boolean
  onCursorSecondsChange: (cursorSeconds: number) => void
  onPlayingChange: (isPlaying: boolean) => void
}

export function FlightReplayTimeNavigation({
  replay,
  frame,
  isPlaying,
  onCursorSecondsChange,
  onPlayingChange,
}: FlightReplayTimeNavigationProps) {
  if (!replay || replay.points.length === 0 || frame.point === null) return null

  const navigation = buildFlightReplayTimeNavigation(replay, frame.cursorSeconds)
  const canNavigateTime = frame.totalObservedSpanSeconds > 0
  const cursorLabel = formatTimestamp(frame.cursorObservedAt)
  const heldObservationLabel = formatTimestamp(frame.point.observed_at)

  const jump = (cursorSeconds: number | null) => {
    if (cursorSeconds === null) return
    onPlayingChange(false)
    onCursorSecondsChange(cursorSeconds)
  }

  return (
    <section
      className='border-t border-white/10 bg-[#111315] px-3 py-3 sm:px-4'
      aria-label='Historical replay time navigation'
      data-flight-replay-time-cursor='elapsed-time'
      data-flight-replay-position-policy='last-persisted-observation'
      data-flight-replay-interpolation='none'
    >
      <div className='flex flex-wrap items-start justify-between gap-3'>
        <div>
          <h3 className='text-sm font-bold text-white'>Observed time navigation</h3>
          <p className='mt-1 max-w-3xl text-[11px] leading-relaxed text-slate-500'>
            The cursor moves through real elapsed observation time. Aircraft position changes only
            when another persisted observation is reached; no position is synthesized inside a gap.
          </p>
        </div>
        <div className='text-right'>
          <p className='font-mono text-xs text-slate-300'>{cursorLabel}</p>
          <p className='mt-0.5 text-[10px] uppercase tracking-[0.1em] text-slate-600'>
            cursor time
          </p>
        </div>
      </div>

      <div className='mt-3 grid gap-2 sm:grid-cols-3'>
        <ReplayTimeDatum label='Elapsed from start' value={formatDuration(frame.cursorSeconds)} />
        <ReplayTimeDatum
          label='Observed span'
          value={formatDuration(frame.totalObservedSpanSeconds)}
        />
        <ReplayTimeDatum
          label='Held observation'
          value={`#${frame.cursorIndex + 1} · ${heldObservationLabel}`}
        />
      </div>

      <div
        className={`mt-3 rounded-lg border px-3 py-2 text-xs ${
          frame.exactObservationAtCursor
            ? 'border-emerald-400/25 bg-emerald-400/5 text-emerald-200'
            : 'border-amber-300/25 bg-amber-300/5 text-amber-100'
        }`}
        aria-live='polite'
      >
        {frame.exactObservationAtCursor ? (
          <p>Exact persisted observation at cursor.</p>
        ) : (
          <p>
            No observation at cursor. Holding the last persisted position from{' '}
            <span className='font-mono'>{heldObservationLabel}</span> ·{' '}
            <span className='font-mono'>{formatDuration(frame.secondsSinceObserved)}</span> since
            observed.
          </p>
        )}
      </div>

      <input
        type='range'
        min={0}
        max={Math.max(0, frame.totalObservedSpanSeconds)}
        step={0.25}
        value={frame.cursorSeconds}
        disabled={!canNavigateTime}
        onChange={event => {
          onPlayingChange(false)
          onCursorSecondsChange(Number(event.target.value))
        }}
        aria-label='Historical replay time cursor'
        className='mt-3 w-full accent-amber-300 disabled:opacity-50'
      />

      <div className='mt-1 flex justify-between gap-2 font-mono text-[10px] text-slate-600'>
        <span>{formatTimestamp(replay.points[0]?.observed_at ?? null)}</span>
        <span>{formatTimestamp(replay.points[replay.points.length - 1]?.observed_at ?? null)}</span>
      </div>

      <div
        className='mt-3 flex flex-wrap gap-2'
        role='group'
        aria-label='Historical replay time jumps'
      >
        <ReplayJumpButton
          label='Start'
          disabled={!canNavigateTime || frame.cursorSeconds <= 0}
          onClick={() => jump(navigation.startCursorSeconds)}
        />
        <ReplayJumpButton
          label='Previous observation'
          disabled={navigation.previousObservationCursorSeconds === null}
          onClick={() => jump(navigation.previousObservationCursorSeconds)}
        />
        <ReplayJumpButton
          label='Next observation'
          disabled={navigation.nextObservationCursorSeconds === null}
          onClick={() => jump(navigation.nextObservationCursorSeconds)}
        />
        <ReplayJumpButton
          label={`Largest gap${navigation.largestGapSeconds > 0 ? ` · ${formatDuration(navigation.largestGapSeconds)}` : ''}`}
          disabled={navigation.largestGapCursorSeconds === null}
          onClick={() => jump(navigation.largestGapCursorSeconds)}
        />
        <ReplayJumpButton
          label='End'
          disabled={!canNavigateTime || frame.cursorSeconds >= navigation.endCursorSeconds}
          onClick={() => jump(navigation.endCursorSeconds)}
        />
      </div>

      <p className='mt-3 text-[11px] leading-relaxed text-slate-500'>
        Time can advance through an unobserved interval, but map position remains anchored to the
        latest persisted observation until the next observed sample timestamp is reached.
        {isPlaying ? ' Playback is currently advancing elapsed time.' : ''}
      </p>
    </section>
  )
}

function ReplayJumpButton({
  label,
  disabled,
  onClick,
}: {
  label: string
  disabled: boolean
  onClick: () => void
}) {
  return (
    <button
      type='button'
      disabled={disabled}
      onClick={onClick}
      className='rounded-md border border-slate-700 px-2.5 py-1.5 text-[11px] font-semibold text-slate-300 hover:bg-slate-900 disabled:cursor-not-allowed disabled:opacity-35'
    >
      {label}
    </button>
  )
}

function ReplayTimeDatum({ label, value }: { label: string; value: string }) {
  return (
    <div className='rounded-lg border border-white/10 bg-slate-950/60 px-2.5 py-2'>
      <p className='text-[9px] font-bold uppercase tracking-[0.12em] text-slate-600'>{label}</p>
      <p className='mt-1 truncate font-mono text-[11px] text-slate-300'>{value}</p>
    </div>
  )
}

function formatTimestamp(value: string | null): string {
  if (!value) return 'Unavailable'
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
