'use client'

import {
  buildFlightReplayEvidenceProvenanceProfile,
  type FlightReplayEvidenceSourceSummary,
  type FlightReplayEvidenceSourceTransition,
} from '@/lib/replay/flight-replay-evidence-provenance-model'
import type { FlightReplay } from '@/types/flight-replay'

interface FlightReplayEvidenceProvenanceProps {
  replay: FlightReplay
  onSelectObservation?: (cursorIndex: number) => void
}

export function FlightReplayEvidenceProvenance({
  replay,
  onSelectObservation,
}: FlightReplayEvidenceProvenanceProps) {
  const profile = buildFlightReplayEvidenceProvenanceProfile(replay)

  return (
    <div
      className='mt-3 rounded-lg border border-white/10 bg-slate-950/50 px-3 py-2.5'
      aria-label='Replay evidence provenance profile'
      data-flight-replay-provenance='persisted-source-labels-only'
      data-flight-replay-provider-ranking='none'
      data-flight-replay-provider-accuracy-claim='none'
    >
      <div className='flex flex-wrap items-start justify-between gap-2'>
        <div>
          <p className='text-[10px] font-bold uppercase tracking-[0.12em] text-slate-500'>
            Replay evidence provenance
          </p>
          <p className='mt-1 max-w-3xl text-[11px] leading-relaxed text-slate-500'>
            Describes the persisted source label attached to each replay observation. It does not
            rank provider accuracy, infer a better source, or reconstruct the exact switch instant
            inside an unobserved interval.
          </p>
        </div>
        <span className='rounded-full border border-sky-400/25 bg-sky-400/10 px-2 py-0.5 text-[9px] font-bold uppercase tracking-[0.1em] text-sky-200'>
          Observed provenance
        </span>
      </div>

      <div className='mt-2 grid gap-2 sm:grid-cols-2 lg:grid-cols-4'>
        <ReplayProvenanceDatum label='Samples' value={String(profile.sampleCount)} />
        <ReplayProvenanceDatum
          label='Identified sources'
          value={String(profile.identifiedSourceCount)}
        />
        <ReplayProvenanceDatum
          label='Unattributed samples'
          value={String(profile.unattributedSampleCount)}
        />
        <ReplayProvenanceDatum
          label='Observed source transitions'
          value={String(profile.transitions.length)}
        />
      </div>

      <div className='mt-3 grid gap-3 lg:grid-cols-2'>
        <section aria-label='Replay evidence sources'>
          <p className='text-[9px] font-bold uppercase tracking-[0.12em] text-slate-600'>
            Source composition
          </p>
          {profile.sources.length === 0 ? (
            <p className='mt-2 text-xs text-slate-400'>
              No identified source label is available for these persisted observations.
            </p>
          ) : (
            <div className='mt-2 space-y-2'>
              {profile.sources.map(source => (
                <ReplaySourceSummary key={source.sourceName} source={source} />
              ))}
            </div>
          )}
        </section>

        <section aria-label='Replay evidence source transitions'>
          <p className='text-[9px] font-bold uppercase tracking-[0.12em] text-slate-600'>
            Adjacent source-label transitions
          </p>
          {profile.transitions.length === 0 ? (
            <p className='mt-2 text-xs text-slate-400'>
              No adjacent persisted observations carry different identified source labels.
            </p>
          ) : (
            <div className='mt-2 max-h-48 space-y-2 overflow-y-auto pr-1'>
              {profile.transitions.map(transition => (
                <ReplaySourceTransition
                  key={`${transition.fromIndex}-${transition.toIndex}-${transition.startObservedAt}`}
                  transition={transition}
                  onSelectObservation={onSelectObservation}
                />
              ))}
            </div>
          )}
        </section>
      </div>

      <p className='mt-3 text-[11px] leading-relaxed text-slate-500'>
        Percentages are shares of persisted samples, not shares of elapsed time. A source-label
        transition means only that two adjacent persisted observations name different sources; the
        exact provider switch time between those observations is unknown.
      </p>
    </div>
  )
}

function ReplaySourceSummary({
  source,
}: {
  source: FlightReplayEvidenceSourceSummary
}) {
  return (
    <div className='rounded-md border border-white/10 bg-slate-950/70 px-2.5 py-2'>
      <div className='flex flex-wrap items-center justify-between gap-2'>
        <span className='font-mono text-xs text-slate-200'>{source.sourceName}</span>
        <span className='font-mono text-[10px] text-slate-500'>
          {source.sampleCount} samples · {source.sampleSharePercent}%
        </span>
      </div>
      <p className='mt-1 text-[10px] text-slate-600'>
        {formatTimestamp(source.firstObservedAt)} → {formatTimestamp(source.lastObservedAt)}
      </p>
    </div>
  )
}

function ReplaySourceTransition({
  transition,
  onSelectObservation,
}: {
  transition: FlightReplayEvidenceSourceTransition
  onSelectObservation?: (cursorIndex: number) => void
}) {
  return (
    <div className='rounded-md border border-white/10 bg-slate-950/70 px-2.5 py-2'>
      <div className='flex flex-wrap items-center justify-between gap-2'>
        <span className='font-mono text-[11px] text-slate-300'>
          {transition.fromSourceName} → {transition.toSourceName}
        </span>
        {onSelectObservation ? (
          <button
            type='button'
            onClick={() => onSelectObservation(transition.toIndex)}
            className='rounded border border-slate-700 px-2 py-1 text-[10px] font-semibold text-slate-300 hover:bg-slate-900'
          >
            Jump to later observation
          </button>
        ) : null}
      </div>
      <p className='mt-1 font-mono text-[10px] text-slate-600'>
        {formatTimestamp(transition.startObservedAt)} → {formatTimestamp(transition.endObservedAt)}
      </p>
      <p className='mt-1 text-[10px] text-slate-600'>
        Adjacent observation interval:{' '}
        <span className='font-mono text-slate-400'>
          {transition.elapsedSeconds === null
            ? 'Unavailable'
            : formatDuration(transition.elapsedSeconds)}
        </span>
      </p>
    </div>
  )
}

function ReplayProvenanceDatum({ label, value }: { label: string; value: string }) {
  return (
    <div className='rounded-lg border border-white/10 bg-slate-950/60 px-2.5 py-2'>
      <p className='text-[9px] font-bold uppercase tracking-[0.12em] text-slate-600'>{label}</p>
      <p className='mt-1 font-mono text-[11px] text-slate-300'>{value}</p>
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
