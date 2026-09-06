'use client'

import { APIRequestError, getRequestErrorMessage } from '@/lib/api/client'
import type {
  AirspaceNotice,
  AirspaceRegionAnalyticsResponse,
} from '@/types/airspace-intelligence'

interface Props {
  regionCode: string
  regionName: string
  asOfTime: string | null
  result: AirspaceRegionAnalyticsResponse | undefined
  isPending: boolean
  isFetching: boolean
  error: Error | null
  onRetry: () => void
}

export function AirspaceIntelligencePanel({
  regionCode,
  regionName,
  asOfTime,
  result,
  isPending,
  isFetching,
  error,
  onRetry,
}: Props) {
  const world = regionCode.trim().toLowerCase() === 'world'
  const unavailable = error instanceof APIRequestError && error.status === 404

  return (
    <section
      className='mt-8 rounded-2xl border border-white/10 bg-slate-900/65 p-4 shadow-xl shadow-black/10 sm:p-6'
      aria-labelledby='airspace-intelligence-title'
      aria-busy={isFetching}
    >
      <div className='flex flex-wrap items-start justify-between gap-4'>
        <div>
          <p className='text-xs font-semibold uppercase tracking-[0.2em] text-violet-300'>
            Server-owned regional analytics
          </p>
          <h2
            id='airspace-intelligence-title'
            className='mt-2 text-xl font-semibold text-white'
          >
            Airspace Intelligence — {regionName}
          </h2>
          <p className='mt-2 max-w-3xl text-sm leading-6 text-slate-400'>
            Bounded research analytics over persisted regional observations.
            Occupancy, complexity and pressure are descriptive analytical
            evidence, not air traffic control guidance or certified separation
            monitoring.
          </p>
        </div>
        {isFetching ? (
          <span className='rounded-full border border-violet-400/30 bg-violet-400/10 px-3 py-1 text-xs font-medium text-violet-200'>
            Updating analysis…
          </span>
        ) : null}
      </div>

      {world ? (
        <Message>
          Select a bounded backend region to request Airspace Intelligence. The
          World option is a frontend-wide traffic view and is not treated as an
          analytical airspace region.
        </Message>
      ) : null}

      {!world && asOfTime === null && !error ? (
        <Message>
          Waiting for an observed regional traffic timestamp before requesting
          Airspace Intelligence.
        </Message>
      ) : null}

      {!world && asOfTime !== null && isPending && !error ? (
        <p className='mt-5 text-sm leading-6 text-slate-400'>
          Building the bounded regional analytical view from persisted evidence…
        </p>
      ) : null}

      {unavailable ? (
        <Message>
          Airspace Intelligence is unavailable for the selected region and
          evidence window.
        </Message>
      ) : null}

      {error && !unavailable ? (
        <div className='mt-5 rounded-xl border border-amber-400/30 bg-amber-400/10 p-4'>
          <p className='text-sm leading-6 text-amber-100'>
            {getRequestErrorMessage(error)}
          </p>
          <button
            type='button'
            onClick={onRetry}
            disabled={isFetching}
            className='mt-3 rounded-md border border-amber-300/40 px-3 py-1.5 text-sm font-medium text-amber-100 disabled:opacity-60'
          >
            Retry Airspace Intelligence
          </button>
        </div>
      ) : null}

      {result ? <Content result={result} /> : null}
    </section>
  )
}

function Content({ result }: { result: AirspaceRegionAnalyticsResponse }) {
  const metrics = result.metrics
  const occupancy = result.occupancy.metrics

  return (
    <>
      <div className='mt-5 grid gap-4 xl:grid-cols-4'>
        <MetricCard
          title='Regional occupancy'
          items={[
            ['Current aircraft', formatInteger(metrics.current_aircraft_count)],
            ['Unique aircraft', formatInteger(metrics.unique_aircraft_count)],
            ['Observations', formatInteger(metrics.aircraft_observation_count)],
            ['Temporal coverage', formatPercent(metrics.temporal_coverage)],
          ]}
        />
        <MetricCard
          title='Traffic envelope'
          items={[
            ['Peak per bucket', formatInteger(metrics.peak_aircraft_per_bucket)],
            ['Mean per bucket', formatNumber(metrics.mean_aircraft_per_bucket)],
            ['Occupied cells', formatInteger(metrics.occupied_cell_count)],
            ['Unknown altitude', formatInteger(metrics.unknown_altitude_count)],
          ]}
        />
        <MetricCard
          title='Complexity evidence'
          items={[
            ['Mean complexity', formatPercent(metrics.mean_complexity_score)],
            ['Peak complexity', formatPercent(metrics.peak_complexity_score)],
            ['Highest level', humanize(metrics.highest_complexity_level)],
            ['Occupancy trend', humanize(metrics.occupancy_trend)],
          ]}
        />
        <MetricCard
          title='Pressure and risk context'
          items={[
            ['Pressure index', formatPercent(metrics.airspace_pressure_index)],
            ['Peak pressure', formatPercent(metrics.peak_airspace_pressure_index)],
            ['Elevated / high', `${metrics.elevated_risk_count} / ${metrics.high_risk_count}`],
            ['Indeterminate', formatInteger(metrics.indeterminate_risk_count)],
          ]}
        />
      </div>

      <div className='mt-4 grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]'>
        <article className='rounded-xl border border-violet-400/20 bg-violet-400/5 p-4'>
          <div className='flex flex-wrap items-center justify-between gap-3'>
            <div>
              <h3 className='text-sm font-semibold text-violet-100'>
                Confidence and evidence window
              </h3>
              <p className='mt-1 text-xs leading-5 text-slate-500'>
                Confidence describes evidence support for this analytical result;
                it is not calibrated probability or operational certainty.
              </p>
            </div>
            <span className='rounded-full border border-violet-300/35 bg-violet-300/10 px-2.5 py-1 text-xs font-semibold uppercase tracking-wide text-violet-100'>
              {humanize(result.confidence.level)} ·{' '}
              {formatPercent(result.confidence.score)}
            </span>
          </div>
          <dl className='mt-4 grid grid-cols-2 gap-3 text-sm'>
            <Detail label='Window start' value={formatTimestamp(result.window_start)} />
            <Detail label='Window end' value={formatTimestamp(result.window_end)} />
            <Detail label='Buckets' value={`${occupancy.bucket_count} / ${occupancy.expected_bucket_count}`} />
            <Detail label='Latest observed' value={formatTimestamp(result.provenance.latest_observed_at)} />
          </dl>
          {result.confidence.reasons.length > 0 ? (
            <ul className='mt-4 space-y-2 border-t border-violet-300/10 pt-3 text-sm leading-5 text-slate-400'>
              {result.confidence.reasons.slice(0, 4).map(reason => (
                <li key={reason.code}>{reason.message}</li>
              ))}
            </ul>
          ) : null}
        </article>

        <article className='rounded-xl border border-slate-800 bg-slate-950/65 p-4'>
          <h3 className='text-sm font-semibold text-white'>Provenance and scope</h3>
          <dl className='mt-4 grid grid-cols-2 gap-3 text-sm'>
            <Detail label='Status' value={humanize(result.status)} />
            <Detail label='Region code' value={result.region_code.toUpperCase()} />
            <Detail
              label='Sources'
              value={result.provenance.source_names.join(', ') || 'Unknown'}
            />
            <Detail label='Generated' value={formatTimestamp(result.generated_at)} />
          </dl>
          <p className='mt-4 break-all font-mono text-[11px] leading-5 text-slate-500'>
            {result.scope_guard}
          </p>
          <p className='mt-2 break-all font-mono text-[11px] leading-5 text-slate-600'>
            {result.provenance.input_fingerprint}
          </p>
        </article>
      </div>

      <NoticeList title='Limitations' items={result.limitations} warning />
      <NoticeList title='Explanations' items={result.explanations} />

      <p className='mt-4 text-xs leading-5 text-slate-600'>
        Sector counts, pressure indices and risk-context counts are research-only
        analytical outputs derived by the existing backend. They do not represent
        official sectors, controller workload, regulatory separation minima,
        collision prediction or safety-critical guidance.
      </p>
    </>
  )
}

function MetricCard({
  title,
  items,
}: {
  title: string
  items: Array<[string, string]>
}) {
  return (
    <article className='rounded-xl border border-slate-800 bg-slate-950/65 p-4'>
      <h3 className='text-sm font-semibold text-white'>{title}</h3>
      <dl className='mt-4 grid grid-cols-2 gap-3'>
        {items.map(([label, value]) => (
          <Detail key={label} label={label} value={value} />
        ))}
      </dl>
    </article>
  )
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className='text-[11px] uppercase tracking-wide text-slate-600'>{label}</dt>
      <dd className='mt-1 break-words text-sm font-medium text-slate-200'>
        {value.trim() || 'Unknown'}
      </dd>
    </div>
  )
}

function Message({ children }: { children: string }) {
  return (
    <p className='mt-5 rounded-xl border border-dashed border-slate-700 p-5 text-sm leading-6 text-slate-400'>
      {children}
    </p>
  )
}

function NoticeList({
  title,
  items,
  warning = false,
}: {
  title: string
  items: AirspaceNotice[]
  warning?: boolean
}) {
  return (
    <div
      className={`mt-4 rounded-xl border p-4 ${
        warning
          ? 'border-amber-400/25 bg-amber-400/5'
          : 'border-slate-800 bg-slate-950/65'
      }`}
    >
      <h3
        className={`text-xs font-semibold uppercase tracking-wide ${
          warning ? 'text-amber-200' : 'text-slate-300'
        }`}
      >
        {title}
      </h3>
      {items.length > 0 ? (
        <ul
          className={`mt-3 space-y-2 text-sm leading-5 ${
            warning ? 'text-amber-100' : 'text-slate-400'
          }`}
        >
          {items.slice(0, 6).map(item => (
            <li key={item.code}>{item.message}</li>
          ))}
        </ul>
      ) : (
        <p className='mt-2 text-sm text-slate-500'>No {title.toLowerCase()} reported.</p>
      )}
    </div>
  )
}

function formatInteger(value: number): string {
  return new Intl.NumberFormat().format(value)
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat(undefined, { maximumFractionDigits: 1 }).format(value)
}

function formatPercent(value: number): string {
  return new Intl.NumberFormat(undefined, {
    style: 'percent',
    maximumFractionDigits: 0,
  }).format(value)
}

function formatTimestamp(value: string): string {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'medium',
    timeZone: 'UTC',
  }).format(new Date(value))
}

function humanize(value: string): string {
  return value.replaceAll('_', ' ')
}
