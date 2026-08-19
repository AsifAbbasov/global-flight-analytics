// FRONTEND_LIVE_TRAFFIC_CONTROL_V1
// FRONTEND_RUNTIME_STABILIZATION_V2
'use client'

import type { ChangeEvent, ReactNode } from 'react'

import {
  buildLiveTrafficStatus,
  liveTrafficRefreshIntervals,
  type LiveTrafficFreshness,
  type LiveTrafficRefreshIntervalMilliseconds,
} from '@/lib/traffic/live-traffic-status-model'

interface LiveTrafficControlProps {
  regionName: string
  aircraftCount: number
  selectedAircraftICAO24: string | null
  dataUpdatedAt: number
  now: number
  isInitialLoading: boolean
  isRefreshing: boolean
  errorMessage: string | null
  regionsWarning: string | null
  autoRefreshEnabled: boolean
  refreshIntervalMilliseconds: LiveTrafficRefreshIntervalMilliseconds
  onAutoRefreshEnabledChange: (enabled: boolean) => void
  onRefreshIntervalChange: (
    interval: LiveTrafficRefreshIntervalMilliseconds
  ) => void
  onRetry: () => void
}

export function LiveTrafficControl({
  regionName,
  aircraftCount,
  selectedAircraftICAO24,
  dataUpdatedAt,
  now,
  isInitialLoading,
  isRefreshing,
  errorMessage,
  regionsWarning,
  autoRefreshEnabled,
  refreshIntervalMilliseconds,
  onAutoRefreshEnabledChange,
  onRefreshIntervalChange,
  onRetry,
}: LiveTrafficControlProps) {
  const status = buildLiveTrafficStatus({
    now,
    dataUpdatedAt,
    refreshIntervalMilliseconds,
    autoRefreshEnabled,
  })
  const presentation = resolveStatusPresentation({
    freshness: status.freshness,
    isInitialLoading,
    isRefreshing,
    hasError: errorMessage !== null,
    hasSnapshot: dataUpdatedAt > 0,
  })

  return (
    <section
      aria-label='Live traffic data controls'
      className='mt-4 rounded-xl border border-slate-800 bg-slate-950/55 p-4'
    >
      <div className='flex flex-wrap items-start justify-between gap-4'>
        <div aria-live='polite' className='min-w-0'>
          <div className='flex items-center gap-2'>
            <span
              aria-hidden='true'
              className={`h-2.5 w-2.5 rounded-full ${statusDotClassName(
                presentation.kind
              )}`}
            />
            <p className='text-[11px] font-semibold uppercase tracking-[0.16em] text-slate-500'>
              Live data status
            </p>
          </div>
          <p className='mt-2 text-sm font-semibold text-slate-100'>
            {presentation.label}
          </p>
          <p className='mt-1 max-w-2xl text-xs leading-5 text-slate-500'>
            {presentation.summary}
          </p>
        </div>

        <div className='flex flex-wrap items-end gap-2'>
          <label className='text-xs text-slate-500'>
            Refresh interval
            <select
              value={refreshIntervalMilliseconds}
              onChange={(event: ChangeEvent<HTMLSelectElement>) => {
                onRefreshIntervalChange(
                  Number(event.target.value) as LiveTrafficRefreshIntervalMilliseconds
                )
              }}
              className='mt-1 block rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-slate-200'
            >
              {liveTrafficRefreshIntervals.map(interval => (
                <option key={interval} value={interval}>
                  {formatIntervalLabel(interval)}
                </option>
              ))}
            </select>
          </label>
          <button
            type='button'
            onClick={() => {
              onAutoRefreshEnabledChange(!autoRefreshEnabled)
            }}
            className='rounded-lg border border-slate-700 px-3 py-2 text-sm font-medium text-slate-200 transition hover:bg-slate-900'
          >
            {autoRefreshEnabled ? 'Pause auto-refresh' : 'Resume auto-refresh'}
          </button>
        </div>
      </div>

      <div className='mt-4 grid gap-3 md:grid-cols-3'>
        <StatusMetric
          label='Current snapshot'
          value={`${formatInteger(aircraftCount)} aircraft`}
          detail={
            selectedAircraftICAO24 === null
              ? regionName
              : `${regionName} · ${selectedAircraftICAO24.toUpperCase()} selected`
          }
        />
        <StatusMetric
          label='Last successful update'
          value={
            dataUpdatedAt > 0
              ? formatAbsoluteTimestamp(dataUpdatedAt)
              : 'Not available'
          }
          detail={
            status.ageMilliseconds === null
              ? 'Waiting for the first successful response.'
              : `${formatDuration(status.ageMilliseconds)} ago`
          }
        >
          {status.ageMilliseconds !== null ? (
            <div
              aria-hidden='true'
              className='mt-2 h-1.5 overflow-hidden rounded-full bg-slate-800'
            >
              <div
                className={`h-full rounded-full ${progressClassName(
                  status.freshness
                )}`}
                style={{ width: `${status.intervalProgress * 100}%` }}
              />
            </div>
          ) : null}
        </StatusMetric>
        <StatusMetric
          label='Refresh policy'
          value={
            autoRefreshEnabled
              ? `Every ${formatIntervalLabel(
                  status.refreshIntervalMilliseconds
                )}`
              : 'Automatic refresh paused'
          }
          detail={refreshPolicyDetail({
            autoRefreshEnabled,
            isRefreshing,
            nextRefreshInMilliseconds: status.nextRefreshInMilliseconds,
            refreshDue: status.refreshDue,
          })}
        />
      </div>

      {regionsWarning || errorMessage ? (
        <div className='mt-4 flex flex-wrap items-center gap-3 border-t border-slate-800 pt-3 text-sm'>
          {regionsWarning ? (
            <span className='text-amber-300'>{regionsWarning}</span>
          ) : null}
          {errorMessage ? (
            <>
              <span className='text-rose-200'>{errorMessage}</span>
              <button
                type='button'
                onClick={onRetry}
                disabled={isRefreshing}
                className='rounded-md border border-rose-400/40 px-3 py-1.5 text-xs font-medium text-rose-100 disabled:opacity-60'
              >
                Retry traffic request
              </button>
            </>
          ) : null}
        </div>
      ) : null}
    </section>
  )
}

function StatusMetric({
  label,
  value,
  detail,
  children,
}: {
  label: string
  value: string
  detail: string
  children?: ReactNode
}) {
  return (
    <article className='rounded-lg border border-slate-800 bg-slate-950/70 p-3'>
      <p className='text-[11px] uppercase tracking-[0.14em] text-slate-600'>
        {label}
      </p>
      <p className='mt-1.5 text-sm font-semibold text-slate-200'>{value}</p>
      <p className='mt-1 text-xs leading-5 text-slate-500'>{detail}</p>
      {children}
    </article>
  )
}

interface StatusPresentation {
  kind:
    | 'loading'
    | 'refreshing'
    | 'current'
    | 'aging'
    | 'stale'
    | 'degraded'
    | 'unavailable'
    | 'waiting'
  label: string
  summary: string
}

function resolveStatusPresentation({
  freshness,
  isInitialLoading,
  isRefreshing,
  hasError,
  hasSnapshot,
}: {
  freshness: LiveTrafficFreshness
  isInitialLoading: boolean
  isRefreshing: boolean
  hasError: boolean
  hasSnapshot: boolean
}): StatusPresentation {
  if (isInitialLoading) {
    return {
      kind: 'loading',
      label: 'Loading current traffic',
      summary: 'The first regional snapshot request is still in progress.',
    }
  }
  if (isRefreshing) {
    return {
      kind: 'refreshing',
      label: 'Refreshing traffic snapshot',
      summary: 'The previous successful snapshot remains visible during refresh.',
    }
  }
  if (hasError && hasSnapshot) {
    return {
      kind: 'degraded',
      label: 'Refresh failed · previous snapshot retained',
      summary: 'The interface is showing the latest successful response instead of clearing useful evidence.',
    }
  }
  if (hasError) {
    return {
      kind: 'unavailable',
      label: 'Traffic snapshot unavailable',
      summary: 'No successful regional traffic response is currently available.',
    }
  }

  switch (freshness) {
    case 'current':
      return {
        kind: 'current',
        label: 'Snapshot current',
        summary: 'The latest response is inside the selected automatic refresh window.',
      }
    case 'aging':
      return {
        kind: 'aging',
        label: 'Snapshot aging',
        summary: 'A refresh interval has elapsed since the latest successful response.',
      }
    case 'stale':
      return {
        kind: 'stale',
        label: 'Snapshot stale',
        summary: 'More than two refresh intervals have elapsed without a successful response.',
      }
    case 'waiting':
      return {
        kind: 'waiting',
        label: 'Waiting for traffic data',
        summary: 'No successful timestamp is available yet.',
      }
  }
}

function statusDotClassName(kind: StatusPresentation['kind']): string {
  switch (kind) {
    case 'current':
      return 'bg-emerald-300 shadow-[0_0_14px_rgba(110,231,183,0.75)]'
    case 'loading':
    case 'refreshing':
      return 'bg-sky-300 shadow-[0_0_14px_rgba(125,211,252,0.7)]'
    case 'aging':
    case 'degraded':
      return 'bg-amber-300 shadow-[0_0_14px_rgba(252,211,77,0.65)]'
    case 'stale':
    case 'unavailable':
      return 'bg-rose-300 shadow-[0_0_14px_rgba(253,164,175,0.65)]'
    case 'waiting':
      return 'bg-slate-500'
  }
}

function progressClassName(freshness: LiveTrafficFreshness): string {
  switch (freshness) {
    case 'current':
      return 'bg-emerald-300'
    case 'aging':
      return 'bg-amber-300'
    case 'stale':
      return 'bg-rose-300'
    case 'waiting':
      return 'bg-slate-600'
  }
}

function refreshPolicyDetail({
  autoRefreshEnabled,
  isRefreshing,
  nextRefreshInMilliseconds,
  refreshDue,
}: {
  autoRefreshEnabled: boolean
  isRefreshing: boolean
  nextRefreshInMilliseconds: number | null
  refreshDue: boolean
}): string {
  if (!autoRefreshEnabled) {
    return 'No automatic traffic requests will run until resumed.'
  }
  if (isRefreshing) {
    return 'A traffic request is currently in progress.'
  }
  if (refreshDue || nextRefreshInMilliseconds === 0) {
    return 'The next automatic refresh is due.'
  }
  if (nextRefreshInMilliseconds === null) {
    return 'Waiting for the first successful response.'
  }

  return `Next automatic refresh in ${formatDuration(
    nextRefreshInMilliseconds
  )}.`
}

function formatIntervalLabel(value: number): string {
  if (value < 60_000) {
    return `${Math.round(value / 1000)} seconds`
  }

  const minutes = value / 60_000
  return minutes === 1 ? '1 minute' : `${minutes} minutes`
}

function formatDuration(value: number): string {
  const totalSeconds = Math.max(0, Math.ceil(value / 1000))
  if (totalSeconds < 60) {
    return `${totalSeconds} seconds`
  }

  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60
  if (seconds === 0) {
    return minutes === 1 ? '1 minute' : `${minutes} minutes`
  }

  return `${minutes}m ${seconds}s`
}

function formatAbsoluteTimestamp(value: number): string {
  return `${new Intl.DateTimeFormat('en-GB', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
    timeZone: 'UTC',
  }).format(new Date(value))} UTC`
}
function formatInteger(value: number): string {
  return new Intl.NumberFormat().format(value)
}
