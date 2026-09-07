'use client'

import type { ReactNode } from 'react'

import { getRequestErrorMessage } from '@/lib/api/client'
import { useETAEvolution } from '@/lib/queries/eta-evolution'
import { useTrajectoryFlightReplay } from '@/lib/queries/flight-replay'
import type { AircraftTrajectory } from '@/types/trajectory'

interface Props {
  selectedICAO24: string | null
  trajectory: AircraftTrajectory | undefined
}

export function ETAEvolutionPanel({ selectedICAO24, trajectory }: Props) {
  const replayQuery = useTrajectoryFlightReplay(trajectory)
  const evolution = useETAEvolution(trajectory?.id ?? null, replayQuery.data)

  if (selectedICAO24 === null) return null

  return (
    <aside
      className='rounded-xl border border-slate-700 bg-slate-950/95 p-5'
      aria-labelledby='eta-evolution-title'
      data-eta-evolution-evidence='historically-recomputed-from-persisted-observations'
      data-eta-evolution-persisted-forecast-history='none'
      data-eta-evolution-interpolation='none'
      data-eta-evolution-cause-inference='none'
    >
      <div className='flex flex-wrap items-start justify-between gap-4'>
        <div>
          <p className='text-xs font-semibold uppercase tracking-[0.18em] text-fuchsia-300'>
            Historical forecast recomputation
          </p>
          <h3 id='eta-evolution-title' className='mt-2 text-lg font-semibold text-white'>
            Estimated Arrival Evolution
          </h3>
          <p className='mt-1 max-w-3xl text-xs leading-5 text-slate-400'>
            Recomputes the current Projection Intelligence model at a bounded set of real persisted
            observation timestamps. These are not immutable forecast outputs stored at those past
            moments.
          </p>
        </div>
        {evolution.isFetching ? (
          <span className='text-xs text-sky-300'>Recomputing…</span>
        ) : null}
      </div>

      <p className='mt-3 rounded-lg border border-amber-400/20 bg-amber-400/5 p-3 text-xs leading-5 text-amber-100/80'>
        No ETA is interpolated between samples, and a change in estimated arrival does not establish
        a weather, delay, ATC, airport-congestion, or operational cause.
      </p>

      {!trajectory ? (
        <Message>Waiting for a persisted trajectory before building ETA evolution.</Message>
      ) : replayQuery.isPending ? (
        <Message>Loading persisted replay observations for historical timestamps…</Message>
      ) : replayQuery.error ? (
        <ErrorMessage
          message={getRequestErrorMessage(replayQuery.error)}
          onRetry={() => {
            void replayQuery.refetch()
          }}
        />
      ) : evolution.asOfTimes.length === 0 ? (
        <Message>No valid persisted observation timestamps are available for ETA evolution.</Message>
      ) : evolution.isPending && evolution.summary.availableETACount === 0 ? (
        <Message>
          Recomputing Projection Intelligence at {evolution.asOfTimes.length} persisted observation
          timestamps…
        </Message>
      ) : (
        <EvolutionContent evolution={evolution} />
      )}
    </aside>
  )
}

function EvolutionContent({
  evolution,
}: {
  evolution: ReturnType<typeof useETAEvolution>
}) {
  const { summary } = evolution
  return (
    <>
      <div className='mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-4'>
        <Metric label='Persisted samples used' value={String(summary.sampledObservationCount)} />
        <Metric label='ETA available' value={String(summary.availableETACount)} />
        <Metric label='ETA unavailable' value={String(summary.unavailableETACount)} />
        <Metric
          label='Net ETA change'
          value={formatSignedDuration(summary.netETAChangeSeconds)}
        />
      </div>

      <div className='mt-4 space-y-2' aria-label='ETA evolution recomputed points'>
        {summary.points.map((point, index) => (
          <div
            key={`${point.asOfTime}:${index}`}
            className='rounded-lg border border-slate-800 bg-slate-900/60 p-3'
          >
            <div className='flex flex-wrap items-start justify-between gap-3'>
              <div>
                <p className='font-mono text-[11px] text-slate-400'>
                  {formatTimestamp(point.asOfTime)}
                </p>
                <p className='mt-1 text-xs text-slate-500'>persisted observation timestamp</p>
              </div>
              <span
                className={`rounded-full border px-2 py-1 text-[10px] font-semibold uppercase tracking-wide ${
                  point.available
                    ? 'border-fuchsia-400/30 bg-fuchsia-400/10 text-fuchsia-200'
                    : 'border-slate-700 bg-slate-900 text-slate-400'
                }`}
              >
                {point.available ? 'ETA available' : 'ETA unavailable'}
              </span>
            </div>

            {point.available && point.estimatedTime !== null ? (
              <dl className='mt-3 grid grid-cols-2 gap-x-4 gap-y-3 text-xs lg:grid-cols-4'>
                <Datum label='Airport' value={point.airportICAOCode ?? 'Unavailable'} />
                <Datum label='Estimated arrival' value={formatTimestamp(point.estimatedTime)} />
                <Datum
                  label='Change from prior sample'
                  value={formatSignedDuration(point.etaChangeSeconds)}
                />
                <Datum
                  label='Arrival window'
                  value={formatDuration(point.arrivalWindowSeconds)}
                />
                <Datum
                  label='Confidence'
                  value={formatConfidence(point.confidenceLevel, point.confidenceScore)}
                />
                <Datum label='Method' value={point.method ?? 'Unavailable'} />
                <Datum label='Strategy' value={point.strategy ?? 'Unavailable'} />
                <Datum label='Scope' value={point.scopeGuard ?? 'Unavailable'} />
              </dl>
            ) : (
              <p className='mt-3 text-xs leading-5 text-slate-400'>
                Projection/arrival evidence was unavailable for this persisted observation sample.
                No previous ETA is carried forward.
              </p>
            )}
          </div>
        ))}
      </div>

      {evolution.hasErrors ? (
        <div className='mt-4 rounded-lg border border-amber-400/25 bg-amber-400/5 p-3'>
          <p className='text-xs leading-5 text-amber-100'>
            One or more historical recomputation requests failed. Failed points remain unavailable;
            they are not replaced with nearby ETA values.
          </p>
          <button
            type='button'
            onClick={() => {
              void evolution.refetch()
            }}
            disabled={evolution.isFetching}
            className='mt-2 rounded-md border border-amber-300/40 px-3 py-1.5 text-xs font-medium text-amber-100 disabled:opacity-60'
          >
            Retry ETA evolution
          </button>
        </div>
      ) : null}

      <p className='mt-4 text-[11px] leading-5 text-slate-500'>
        Largest absolute adjacent ETA change:{' '}
        <span className='font-mono'>
          {formatDuration(summary.largestAbsoluteAdjacentETAChangeSeconds)}
        </span>
        . Historical points are recomputed with the current production Projection Intelligence
        implementation against persisted evidence available at each sampled as-of time.
      </p>
    </>
  )
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className='rounded-lg border border-slate-800 bg-slate-900/70 p-3'>
      <p className='text-[10px] uppercase tracking-[0.12em] text-slate-500'>{label}</p>
      <p className='mt-2 font-mono text-sm font-semibold text-slate-100'>{value}</p>
    </div>
  )
}

function Datum({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className='uppercase tracking-wide text-slate-600'>{label}</dt>
      <dd className='mt-1 break-words font-mono text-slate-300'>{value}</dd>
    </div>
  )
}

function Message({ children }: { children: ReactNode }) {
  return (
    <p className='mt-4 rounded-lg border border-slate-800 bg-slate-900/60 p-3 text-sm leading-6 text-slate-400'>
      {children}
    </p>
  )
}

function ErrorMessage({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <div className='mt-4 rounded-lg border border-amber-400/25 bg-amber-400/5 p-3'>
      <p className='text-sm leading-6 text-amber-100'>{message}</p>
      <button
        type='button'
        onClick={onRetry}
        className='mt-2 rounded-md border border-amber-300/40 px-3 py-1.5 text-xs font-medium text-amber-100'
      >
        Retry replay evidence
      </button>
    </div>
  )
}

function formatTimestamp(value: string): string {
  const parsed = Date.parse(value)
  return Number.isFinite(parsed)
    ? new Date(parsed).toISOString().replace('.000Z', 'Z')
    : 'Unavailable'
}

function formatDuration(value: number | null): string {
  if (value === null || !Number.isFinite(value)) return 'Unavailable'
  const seconds = Math.max(0, Math.round(value))
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const remainingSeconds = seconds % 60
  if (hours > 0) return `${hours}h ${minutes}m`
  if (minutes > 0) return `${minutes}m ${remainingSeconds}s`
  return `${remainingSeconds}s`
}

function formatSignedDuration(value: number | null): string {
  if (value === null || !Number.isFinite(value)) return 'Unavailable'
  const sign = value > 0 ? '+' : value < 0 ? '−' : ''
  return `${sign}${formatDuration(Math.abs(value))}`
}

function formatConfidence(level: string | null, score: number | null): string {
  if (level === null || score === null || !Number.isFinite(score)) return 'Unavailable'
  return `${level} · ${Math.round(Math.min(1, Math.max(0, score)) * 100)}%`
}
