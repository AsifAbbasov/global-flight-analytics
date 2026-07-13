import type { AnalyticalMetric } from '@/types/analytics'

interface AnalyticalMetricCardProps {
  title: string
  description: string
  metric: AnalyticalMetric<number> | undefined
  isPending: boolean
  error: Error | null
  onRetry: () => void
  formatValue: (value: number) => string
  emptyMessage?: string
}

export function AnalyticalMetricCard({
  title,
  description,
  metric,
  isPending,
  error,
  onRetry,
  formatValue,
  emptyMessage = 'Waiting for the source metric.',
}: AnalyticalMetricCardProps) {
  return (
    <article
      className='flex min-h-64 flex-col rounded-xl border border-slate-800 bg-slate-900 p-5'
      aria-busy={isPending}
    >
      <div className='flex items-start justify-between gap-4'>
        <div>
          <h3 className='text-base font-semibold text-white'>{title}</h3>
          <p className='mt-1 text-sm leading-6 text-slate-400'>
            {description}
          </p>
        </div>

        {metric ? (
          <span className={statusBadgeClassName(metric.status)}>
            {formatStatus(metric.status)}
          </span>
        ) : null}
      </div>

      <div className='mt-6 flex flex-1 flex-col justify-center'>
        {isPending ? <MetricSkeleton /> : null}

        {!isPending && error ? (
          <MetricError error={error} onRetry={onRetry} />
        ) : null}

        {!isPending && !error && metric ? (
          <MetricContent metric={metric} formatValue={formatValue} />
        ) : null}

        {!isPending && !error && !metric ? (
          <p className='text-sm text-slate-400'>{emptyMessage}</p>
        ) : null}
      </div>
    </article>
  )
}

function MetricContent({
  metric,
  formatValue,
}: {
  metric: AnalyticalMetric<number>
  formatValue: (value: number) => string
}) {
  if (metric.status === 'denied') {
    return (
      <MetricStateMessage
        heading='Analytical scope denied'
        message={
          metric.eligibility?.reasons.join(', ') ||
          'The available observations are not eligible for this metric.'
        }
      />
    )
  }

  if (metric.status === 'failed') {
    return (
      <MetricStateMessage
        heading='Calculation failed'
        message={
          metric.failure?.message ||
          'The analytical calculation could not be completed.'
        }
      />
    )
  }

  if (!metric.has_value || metric.value === undefined) {
    return (
      <MetricStateMessage
        heading='Value unavailable'
        message='The API returned no publishable value.'
      />
    )
  }

  const primaryNotice =
    metric.limitations[0]?.message ?? metric.warnings[0]?.message

  return (
    <>
      <div className='flex items-end gap-3'>
        <p className='text-4xl font-semibold tracking-tight text-white'>
          {formatValue(metric.value)}
        </p>

        <span
          className={confidenceBadgeClassName(metric.confidence.level)}
          title={`Confidence score ${formatPercentage(
            metric.confidence.score
          )}`}
        >
          {metric.confidence.level} confidence
        </span>
      </div>

      <dl className='mt-5 grid grid-cols-2 gap-3 text-sm'>
        <MetricDetail
          label='Eligible inputs'
          value={`${metric.scope.allowed_count}/${metric.scope.input_count}`}
        />
        <MetricDetail
          label='Confidence'
          value={formatPercentage(metric.confidence.score)}
        />
      </dl>

      {primaryNotice ? (
        <p className='mt-4 border-l-2 border-amber-400/60 pl-3 text-sm leading-6 text-amber-100'>
          {primaryNotice}
        </p>
      ) : (
        <p className='mt-4 text-sm text-slate-400'>
          Calculated {formatTimestamp(metric.calculated_at)}.
        </p>
      )}
    </>
  )
}

function MetricError({
  error,
  onRetry,
}: {
  error: Error
  onRetry: () => void
}) {
  return (
    <div role='alert'>
      <p className='font-medium text-rose-200'>Request failed</p>
      <p className='mt-2 text-sm leading-6 text-slate-400'>
        {error.message}
      </p>
      <button
        type='button'
        onClick={onRetry}
        className='mt-4 rounded-md border border-rose-400/50 px-3 py-2 text-sm font-medium text-rose-100 transition hover:bg-rose-400/10'
      >
        Retry
      </button>
    </div>
  )
}

function MetricStateMessage({
  heading,
  message,
}: {
  heading: string
  message: string
}) {
  return (
    <div>
      <p className='font-medium text-amber-200'>{heading}</p>
      <p className='mt-2 text-sm leading-6 text-slate-400'>{message}</p>
    </div>
  )
}

function MetricSkeleton() {
  return (
    <div className='animate-pulse' aria-label='Loading analytical metric'>
      <div className='h-10 w-28 rounded bg-slate-800' />
      <div className='mt-5 h-4 w-full rounded bg-slate-800' />
      <div className='mt-3 h-4 w-2/3 rounded bg-slate-800' />
    </div>
  )
}

function MetricDetail({
  label,
  value,
}: {
  label: string
  value: string
}) {
  return (
    <div className='rounded-lg bg-slate-950/70 p-3'>
      <dt className='text-slate-500'>{label}</dt>
      <dd className='mt-1 font-medium text-slate-200'>{value}</dd>
    </div>
  )
}

function formatStatus(status: AnalyticalMetric<number>['status']): string {
  return status.charAt(0).toUpperCase() + status.slice(1)
}

function formatPercentage(value: number): string {
  return `${Math.round(value * 100)}%`
}

function formatTimestamp(value: string): string {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))
}

function statusBadgeClassName(
  status: AnalyticalMetric<number>['status']
): string {
  const baseClassName =
    'rounded-full border px-2.5 py-1 text-xs font-medium'

  switch (status) {
    case 'complete':
      return `${baseClassName} border-emerald-400/40 bg-emerald-400/10 text-emerald-200`
    case 'limited':
      return `${baseClassName} border-amber-400/40 bg-amber-400/10 text-amber-200`
    case 'denied':
      return `${baseClassName} border-orange-400/40 bg-orange-400/10 text-orange-200`
    case 'failed':
      return `${baseClassName} border-rose-400/40 bg-rose-400/10 text-rose-200`
  }
}

function confidenceBadgeClassName(
  level: AnalyticalMetric<number>['confidence']['level']
): string {
  const baseClassName =
    'mb-1 rounded-full px-2.5 py-1 text-xs font-medium'

  switch (level) {
    case 'high':
      return `${baseClassName} bg-emerald-400/10 text-emerald-200`
    case 'medium':
      return `${baseClassName} bg-sky-400/10 text-sky-200`
    case 'low':
      return `${baseClassName} bg-amber-400/10 text-amber-200`
    case 'none':
      return `${baseClassName} bg-slate-700 text-slate-300`
  }
}
