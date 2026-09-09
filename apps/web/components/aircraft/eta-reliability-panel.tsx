'use client'

import { APIRequestError, getRequestErrorMessage } from '@/lib/api/client'
import { useETAReliability } from '@/lib/queries/eta-reliability'
import type { ProjectionIntelligenceResponse } from '@/types/projection-intelligence'

interface Props {
  trajectoryID: string | null
  projection: ProjectionIntelligenceResponse
}

export function ETAReliabilityPanel({ trajectoryID, projection }: Props) {
  const arrival = projection.projection.arrival
  const query = useETAReliability(
    trajectoryID,
    projection.projection.horizon.as_of_time,
    projection.projection.horizon.duration_seconds,
    arrival !== undefined
  )

  if (arrival === undefined) return null

  const unavailable = query.error instanceof APIRequestError && (query.error.status === 404 || query.error.status === 422)

  return (
    <section className='mt-3 rounded-lg border border-sky-400/25 bg-sky-400/5 p-3' aria-labelledby='eta-reliability-title'>
      <div className='flex flex-wrap items-start justify-between gap-3'>
        <div>
          <p className='text-xs font-semibold uppercase tracking-wide text-sky-200'>Historical ETA Reliability</p>
          <h4 id='eta-reliability-title' className='mt-1 text-sm font-semibold text-white'>How reliable have comparable ETA estimates been?</h4>
        </div>
        {query.data ? <Status value={query.data.status} /> : null}
      </div>

      {query.isPending && !query.error ? (
        <p className='mt-3 text-sm leading-5 text-slate-400'>Comparing this ETA with bounded historical route evidence…</p>
      ) : null}

      {unavailable ? (
        <div className='mt-3 rounded-md border border-slate-700 bg-slate-950/55 p-3'>
          <p className='text-sm leading-5 text-slate-300'>Historical ETA reliability is unavailable for the current route, lead time or persisted evidence.</p>
          <p className='mt-2 text-xs leading-5 text-slate-500'>No synthetic reliability score is shown when comparable history is insufficient.</p>
        </div>
      ) : null}

      {query.error && !unavailable ? (
        <div className='mt-3 rounded-md border border-amber-400/25 bg-amber-400/5 p-3'>
          <p className='text-sm leading-5 text-amber-100'>{getRequestErrorMessage(query.error)}</p>
          <button type='button' onClick={() => void query.refetch()} disabled={query.isFetching} className='mt-3 rounded-md border border-amber-300/40 px-3 py-1.5 text-sm font-medium text-amber-100 disabled:opacity-60'>Retry reliability</button>
        </div>
      ) : null}

      {query.data ? <ReliabilityContent data={query.data} /> : null}
    </section>
  )
}

function ReliabilityContent({ data }: { data: NonNullable<ReturnType<typeof useETAReliability>['data']> }) {
  const metrics = data.metrics
  return (
    <>
      <dl className='mt-3 grid grid-cols-2 gap-x-4 gap-y-3 text-sm'>
        <Detail label='Comparable route' value={`${data.route.origin_icao_code} → ${data.route.destination_icao_code}`} />
        <Detail label='Historical samples' value={`${data.eligible_sample_count} eligible / ${data.candidate_count} checked`} />
        <Detail label='Comparable lead' value={`${formatMinutes(data.target_lead_seconds)} ± ${formatMinutes(data.lead_tolerance_seconds)}`} />
        <Detail label='Method' value={data.method.name} />
      </dl>

      {metrics ? (
        <div className='mt-3 grid gap-2 sm:grid-cols-2'>
          <Metric label='Median ETA error' value={formatDuration(metrics.median_absolute_error_seconds)} />
          <Metric label='80% error threshold' value={`≤ ${formatDuration(metrics.p80_absolute_error_seconds)}`} />
          <Metric label='Within ±5 min' value={formatPercent(metrics.within_five_minutes_ratio)} />
          <Metric label='Within ±10 min' value={formatPercent(metrics.within_ten_minutes_ratio)} />
          <Metric label='ETA window covered endpoint' value={formatPercent(metrics.interval_coverage_ratio)} wide />
        </div>
      ) : (
        <p className='mt-3 rounded-md border border-dashed border-slate-700 p-3 text-sm leading-5 text-slate-400'>Only {data.eligible_sample_count} comparable historical samples were eligible. Reliability metrics remain hidden until the backend evidence threshold is met.</p>
      )}

      <div className='mt-3 rounded-md border border-slate-700/80 bg-slate-950/55 p-3'>
        <p className='text-xs font-semibold uppercase tracking-wide text-slate-400'>Evidence boundary</p>
        <p className='mt-1 text-xs leading-5 text-slate-500'>Historical arrival truth uses the last persisted trajectory observation within {data.endpoint_radius_km.toFixed(0)} km of the destination. It is an observed endpoint proxy, not an official touchdown, gate or schedule timestamp.</p>
      </div>

      {data.limitations.length ? (
        <ul className='mt-3 space-y-1.5 text-xs leading-5 text-slate-500'>
          {data.limitations.slice(0, 4).map(item => <li key={`${item.code}:${item.message}`}>{item.message}</li>)}
        </ul>
      ) : null}
      <p className='mt-3 break-all font-mono text-[10px] leading-4 text-slate-600'>{data.evidence_class} · {data.input_fingerprint}</p>
    </>
  )
}

function Status({ value }: { value: string }) {
  return <span className='rounded-full border border-sky-400/35 bg-sky-400/10 px-2.5 py-1 text-xs font-semibold uppercase tracking-wide text-sky-200'>{value}</span>
}

function Detail({ label, value }: { label: string; value: string }) {
  return <div><dt className='text-xs uppercase tracking-wide text-slate-500'>{label}</dt><dd className='mt-1 break-words text-slate-200'>{value}</dd></div>
}

function Metric({ label, value, wide = false }: { label: string; value: string; wide?: boolean }) {
  return <div className={`rounded-md border border-slate-800 bg-slate-950/60 p-2.5 ${wide ? 'sm:col-span-2' : ''}`}><p className='text-[11px] uppercase tracking-wide text-slate-500'>{label}</p><p className='mt-1 text-base font-semibold text-white'>{value}</p></div>
}

function formatDuration(seconds: number): string {
  const rounded = Math.max(0, Math.round(seconds))
  const minutes = Math.floor(rounded / 60)
  const remaining = rounded % 60
  return minutes === 0 ? `${remaining}s` : remaining === 0 ? `${minutes}m` : `${minutes}m ${remaining}s`
}

function formatMinutes(seconds: number): string {
  const minutes = Math.round(seconds / 60)
  return `${minutes} min`
}

function formatPercent(value: number): string {
  return `${Math.round(value * 100)}%`
}
