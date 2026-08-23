'use client'

import {
  APIRequestError,
  getRequestErrorMessage,
} from '@/lib/api/client'
import {
  buildAircraftIntelligenceDisplay,
  type AircraftIntelligenceField,
} from '@/lib/aircraft/aircraft-intelligence-display'
import { useAircraftProfile } from '@/lib/queries/aircraft'
import type {
  AircraftRouteContext,
  RouteContextAirportCandidate,
} from '@/types/route-context'
import type { TrafficAircraft } from '@/types/traffic'
import type {
  AircraftTrajectory,
  CoverageGap,
  TrajectorySegmentStatus,
} from '@/types/trajectory'

interface AircraftDetailPanelProps {
  selectedICAO24: string | null
  aircraft: TrafficAircraft | undefined
  routeContext: AircraftRouteContext | undefined
  routeContextIsPending: boolean
  routeContextIsFetching: boolean
  routeContextError: Error | null
  onRetryRouteContext: () => void
  trajectory: AircraftTrajectory | undefined
  trajectoryIsPending: boolean
  trajectoryIsFetching: boolean
  trajectoryError: Error | null
  onRetryTrajectory: () => void
  onClose: () => void
}

const primaryTelemetryKeys = new Set(['altitude', 'speed', 'heading', 'status'])

export function AircraftDetailPanel({
  selectedICAO24,
  aircraft,
  routeContext,
  routeContextIsPending,
  routeContextIsFetching,
  routeContextError,
  onRetryRouteContext,
  trajectory,
  trajectoryIsPending,
  trajectoryIsFetching,
  trajectoryError,
  onRetryTrajectory,
  onClose,
}: AircraftDetailPanelProps) {
  const profileQuery = useAircraftProfile(selectedICAO24)

  if (selectedICAO24 === null) {
    return (
      <aside className='flex min-h-[300px] items-center justify-center rounded-xl border border-dashed border-slate-700 bg-slate-950/70 p-6 text-center'>
        <div>
          <p className='text-sm font-semibold text-slate-200'>
            No aircraft selected
          </p>
          <p className='mt-2 text-sm leading-6 text-slate-400'>
            Select a marker to inspect only the live observation, profile and
            analytical evidence GFA can actually support.
          </p>
        </div>
      </aside>
    )
  }

  const display = buildAircraftIntelligenceDisplay(
    selectedICAO24,
    aircraft,
    profileQuery.data
  )
  const primaryTelemetry = display.observedFields.filter(field =>
    primaryTelemetryKeys.has(field.key)
  )
  const secondaryObservation = display.observedFields.filter(
    field => !primaryTelemetryKeys.has(field.key)
  )

  return (
    <aside
      className='overflow-hidden rounded-xl border border-slate-700 bg-[#17191c] shadow-xl shadow-black/20'
      aria-labelledby='aircraft-detail-title'
    >
      <header className='border-b border-white/10 bg-[#202328] px-3.5 py-3.5'>
        <div className='flex items-start justify-between gap-3'>
          <div className='min-w-0'>
            <div className='flex flex-wrap items-center gap-1.5'>
              <span className='rounded bg-amber-300 px-1.5 py-0.5 text-[9px] font-black uppercase tracking-[0.14em] text-slate-950'>
                Aircraft
              </span>
              {aircraft ? (
                <span className='rounded border border-emerald-300/25 bg-emerald-300/10 px-1.5 py-0.5 text-[9px] font-semibold uppercase tracking-[0.12em] text-emerald-200'>
                  Live
                </span>
              ) : null}
            </div>
            <h3
              id='aircraft-detail-title'
              className='mt-2 truncate text-xl font-bold tracking-tight text-white'
            >
              {display.title}
            </h3>
            <p className='mt-0.5 font-mono text-[10px] uppercase tracking-[0.14em] text-slate-400'>
              {display.subtitle}
            </p>
          </div>

          <button
            type='button'
            onClick={onClose}
            className='min-h-9 rounded-md border border-white/10 bg-white/5 px-2.5 text-xs font-semibold text-slate-200 transition hover:bg-white/10'
            aria-label='Close aircraft details'
          >
            Close
          </button>
        </div>
      </header>

      <div className='p-3.5'>
        {primaryTelemetry.length > 0 ? (
          <PrimaryTelemetry fields={primaryTelemetry} />
        ) : null}

        <RouteContextSection
          routeContext={routeContext}
          isPending={routeContextIsPending}
          isFetching={routeContextIsFetching}
          error={routeContextError}
          onRetry={onRetryRouteContext}
        />

        {secondaryObservation.length > 0 ? (
          <section className='mt-4 border-t border-white/10 pt-4' aria-labelledby='observation-detail-title'>
            <SectionHeading
              id='observation-detail-title'
              label='Observation details'
              evidence='Observed'
            />
            <EvidenceFieldGrid fields={secondaryObservation} />
          </section>
        ) : aircraft ? null : (
          <p className='mt-4 rounded-lg border border-amber-400/30 bg-amber-400/10 p-3 text-sm leading-6 text-amber-100'>
            This aircraft is no longer present in the latest traffic response.
            Persisted profile and trajectory evidence can still remain below.
          </p>
        )}

        <section
          className='mt-4 border-t border-white/10 pt-4'
          aria-labelledby='registered-profile-title'
        >
          <div className='flex items-center justify-between gap-3'>
            <SectionHeading
              id='registered-profile-title'
              label='Registered profile'
              evidence='Profile'
            />
            {profileQuery.isFetching ? (
              <span className='text-xs text-sky-300'>Loading…</span>
            ) : null}
          </div>

          {display.profileFields.length > 0 ? (
            <EvidenceFieldGrid fields={display.profileFields} />
          ) : null}

          {profileQuery.isPending ? (
            <p className='mt-3 text-sm text-slate-400'>
              Loading registered profile evidence…
            </p>
          ) : null}

          {profileQuery.error ? (
            <div className='mt-3 rounded-lg border border-amber-400/30 bg-amber-400/10 p-3'>
              <p className='text-sm leading-6 text-amber-100'>
                {getRequestErrorMessage(profileQuery.error)}
              </p>
              <button
                type='button'
                onClick={() => {
                  void profileQuery.refetch()
                }}
                disabled={profileQuery.isFetching}
                className='mt-3 rounded-md border border-amber-300/40 px-3 py-1.5 text-sm font-medium text-amber-100 transition hover:bg-amber-300/10 disabled:cursor-not-allowed disabled:opacity-60'
              >
                Retry profile
              </button>
            </div>
          ) : null}
        </section>

        <TrajectorySection
          trajectory={trajectory}
          isPending={trajectoryIsPending}
          isFetching={trajectoryIsFetching}
          error={trajectoryError}
          onRetry={onRetryTrajectory}
        />
      </div>
    </aside>
  )
}

function PrimaryTelemetry({ fields }: { fields: AircraftIntelligenceField[] }) {
  return (
    <section aria-labelledby='primary-telemetry-title'>
      <div className='flex items-center justify-between gap-2'>
        <h4
          id='primary-telemetry-title'
          className='text-[10px] font-bold uppercase tracking-[0.16em] text-slate-500'
        >
          Live telemetry
        </h4>
        <span className='text-[9px] font-semibold uppercase tracking-[0.12em] text-emerald-300'>
          Observed only
        </span>
      </div>
      <dl className='mt-2 grid grid-cols-2 gap-px overflow-hidden rounded-lg border border-white/10 bg-white/10'>
        {fields.map(field => (
          <div key={field.key} className='min-w-0 bg-[#202328] p-2.5'>
            <dt className='text-[9px] font-bold uppercase tracking-[0.12em] text-slate-500'>
              {field.label}
            </dt>
            <dd className='mt-1 truncate text-sm font-bold text-white'>
              {field.value}
            </dd>
          </div>
        ))}
      </dl>
    </section>
  )
}

function SectionHeading({
  id,
  label,
  evidence,
}: {
  id: string
  label: string
  evidence: string
}) {
  return (
    <div className='flex flex-wrap items-center justify-between gap-2'>
      <h4 id={id} className='text-xs font-semibold text-slate-100'>
        {label}
      </h4>
      <span className='rounded-full border border-white/10 bg-white/5 px-2 py-0.5 text-[9px] font-semibold uppercase tracking-[0.12em] text-slate-400'>
        {evidence}
      </span>
    </div>
  )
}

function EvidenceFieldGrid({ fields }: { fields: AircraftIntelligenceField[] }) {
  return (
    <dl className='mt-2 grid grid-cols-2 gap-px overflow-hidden rounded-lg border border-white/10 bg-white/10'>
      {fields.map(field => (
        <div key={field.key} className='min-w-0 bg-[#202328] p-2.5'>
          <dt className='text-[9px] font-semibold uppercase tracking-[0.12em] text-slate-500'>
            {field.label}
          </dt>
          <dd className='mt-1 break-words text-xs font-semibold text-slate-100'>
            {field.value}
          </dd>
        </div>
      ))}
    </dl>
  )
}

interface RouteContextSectionProps {
  routeContext: AircraftRouteContext | undefined
  isPending: boolean
  isFetching: boolean
  error: Error | null
  onRetry: () => void
}

function RouteContextSection({
  routeContext,
  isPending,
  isFetching,
  error,
  onRetry,
}: RouteContextSectionProps) {
  const isNotFound = error instanceof APIRequestError && error.status === 404

  return (
    <section
      className='mt-4 border-t border-white/10 pt-4'
      aria-labelledby='route-context-title'
    >
      <div className='flex items-center justify-between gap-3'>
        <div>
          <h4
            id='route-context-title'
            aria-label='Probable route and airport context'
            className='text-xs font-semibold text-slate-100'
          >
            Probable route
          </h4>
          <p className='mt-0.5 text-[10px] leading-4 text-slate-500'>
            Inferred from persisted trajectory evidence, not an airline schedule.
          </p>
        </div>
        {isFetching ? <span className='text-xs text-sky-300'>Updating…</span> : null}
      </div>

      {routeContext ? (
        <>
          <div className='mt-2.5 rounded-lg border border-slate-700 bg-slate-950/75 p-3'>
            <div className='grid grid-cols-[1fr_auto_1fr] items-center gap-2'>
              <RouteEndpoint label='Origin' candidate={routeContext.origin} align='left' />
              <div className='flex items-center gap-1 text-slate-600' aria-hidden='true'>
                <span className='h-px w-4 bg-slate-600' />
                <span className='text-sm'>✈</span>
                <span className='h-px w-4 bg-slate-600' />
              </div>
              <RouteEndpoint
                label='Destination'
                candidate={routeContext.destination}
                align='right'
              />
            </div>
            <div className='mt-3 flex items-center justify-between gap-3 border-t border-white/10 pt-2.5'>
              <span className='text-[9px] font-semibold uppercase tracking-[0.12em] text-slate-500'>
                Confidence
              </span>
              <ConfidenceBadge
                level={routeContext.confidence.level}
                score={routeContext.confidence.score}
              />
            </div>
          </div>

          {routeContext.limitations.length > 0 ? (
            <div className='mt-2 rounded-lg border border-amber-400/25 bg-amber-400/5 p-2.5'>
              <p className='text-[9px] font-bold uppercase tracking-[0.12em] text-amber-200'>
                Route limitations
              </p>
              <ul className='mt-1.5 space-y-1 text-xs leading-5 text-amber-100'>
                {routeContext.limitations.map(limitation => (
                  <li key={limitation.code}>{limitation.message}</li>
                ))}
              </ul>
            </div>
          ) : null}
        </>
      ) : null}

      {isPending && !error ? (
        <p className='mt-2.5 text-xs text-slate-400'>
          Inferring airport candidates from persisted trajectory endpoints…
        </p>
      ) : null}

      {isNotFound ? (
        <p className='mt-2.5 rounded-lg border border-slate-700 bg-slate-900/70 p-2.5 text-xs leading-5 text-slate-300'>
          No persisted trajectory exists yet, so GFA does not display a probable
          origin or destination.
        </p>
      ) : null}

      {error && !isNotFound ? (
        <div className='mt-2.5 rounded-lg border border-amber-400/30 bg-amber-400/10 p-3'>
          <p className='text-xs leading-5 text-amber-100'>
            {getRequestErrorMessage(error)}
          </p>
          <button
            type='button'
            onClick={onRetry}
            disabled={isFetching}
            className='mt-2 rounded-md border border-amber-300/40 px-2.5 py-1.5 text-xs font-medium text-amber-100 transition hover:bg-amber-300/10 disabled:opacity-60'
          >
            Retry route context
          </button>
        </div>
      ) : null}
    </section>
  )
}

function RouteEndpoint({
  label,
  candidate,
  align,
}: {
  label: string
  candidate: RouteContextAirportCandidate | undefined
  align: 'left' | 'right'
}) {
  const code = candidate ? airportCode(candidate) : ''
  const location = candidate
    ? [candidate.airport.city, candidate.airport.country]
        .map(value => value.trim())
        .filter(Boolean)
        .join(', ')
    : ''

  return (
    <div className={align === 'right' ? 'min-w-0 text-right' : 'min-w-0'}>
      <p className='text-[9px] font-bold uppercase tracking-[0.12em] text-slate-600'>
        {label}
      </p>
      {candidate ? (
        <>
          <p className='mt-1 truncate font-mono text-base font-black text-white'>
            {code || candidate.airport.name}
          </p>
          {code ? (
            <p className='mt-0.5 truncate text-[10px] text-slate-400'>
              {candidate.airport.name}
            </p>
          ) : null}
          {location ? (
            <p className='mt-0.5 truncate text-[9px] text-slate-600'>{location}</p>
          ) : null}
          <div className={`mt-1.5 flex ${align === 'right' ? 'justify-end' : ''}`}>
            <ConfidenceBadge
              level={candidate.confidence.level}
              score={candidate.confidence.score}
            />
          </div>
        </>
      ) : (
        <p className='mt-1 text-[10px] leading-4 text-slate-500'>
          No supported candidate
        </p>
      )}
    </div>
  )
}

function airportCode(candidate: RouteContextAirportCandidate): string {
  return [candidate.airport.iata_code, candidate.airport.icao_code]
    .map(value => value.trim())
    .filter(Boolean)
    .join(' / ')
}

function ConfidenceBadge({
  level,
  score,
}: {
  level: AircraftRouteContext['confidence']['level']
  score: number
}) {
  const className =
    level === 'high'
      ? 'border-emerald-400/40 bg-emerald-400/10 text-emerald-200'
      : level === 'medium'
        ? 'border-sky-400/40 bg-sky-400/10 text-sky-200'
        : level === 'low'
          ? 'border-amber-400/40 bg-amber-400/10 text-amber-200'
          : 'border-slate-600 bg-slate-800 text-slate-300'

  return (
    <span
      className={`rounded-full border px-2 py-0.5 text-[9px] font-semibold uppercase tracking-wide ${className}`}
    >
      {level} · {formatRatio(score)}
    </span>
  )
}

interface TrajectorySectionProps {
  trajectory: AircraftTrajectory | undefined
  isPending: boolean
  isFetching: boolean
  error: Error | null
  onRetry: () => void
}

function TrajectorySection({
  trajectory,
  isPending,
  isFetching,
  error,
  onRetry,
}: TrajectorySectionProps) {
  const isNotFound = error instanceof APIRequestError && error.status === 404

  return (
    <section
      className='mt-4 border-t border-slate-800 pt-4'
      aria-labelledby='trajectory-quality-title'
    >
      <div className='flex items-center justify-between gap-3'>
        <h4 id='trajectory-quality-title' className='text-xs font-semibold text-slate-200'>
          Latest trajectory
        </h4>
        {isFetching ? <span className='text-xs text-sky-300'>Updating…</span> : null}
      </div>

      {trajectory ? (
        <>
          <div className='mt-2.5 rounded-lg border border-slate-800 bg-slate-900/70 p-2.5'>
            <div className='flex items-center justify-between gap-3'>
              <span className='text-[9px] font-bold uppercase tracking-[0.12em] text-slate-500'>
                Track quality
              </span>
              <span className='text-xs font-semibold text-white'>
                {formatQualityScore(trajectory.quality_score)}
              </span>
            </div>

            {normalizeQualityScore(trajectory.quality_score) !== null ? (
              <div
                className='mt-2 h-1.5 overflow-hidden rounded-full bg-slate-800'
                role='progressbar'
                aria-label='Track quality score'
                aria-valuemin={0}
                aria-valuemax={100}
                aria-valuenow={Math.round(
                  normalizeQualityScore(trajectory.quality_score)! * 100
                )}
              >
                <div
                  className='h-full rounded-full bg-sky-400'
                  style={{
                    width: `${normalizeQualityScore(trajectory.quality_score)! * 100}%`,
                  }}
                />
              </div>
            ) : null}

            <dl className='mt-2.5 grid grid-cols-2 gap-x-3 gap-y-2 text-xs'>
              <Metric label='Segments' value={String(trajectory.segment_count)} />
              <Metric label='Points' value={String(trajectory.point_count)} />
              <Metric label='Coverage gaps' value={String(trajectory.coverage_gap_count)} />
              <Metric label='Duration' value={formatDuration(trajectory.duration_seconds)} />
              <Metric label='Started' value={formatTimestamp(trajectory.start_time)} />
              <Metric label='Updated' value={formatTimestamp(trajectory.updated_at)} />
              <Metric label='Source' value={trajectory.source_name.trim()} className='col-span-2' />
            </dl>
          </div>

          <SegmentStatusSummary trajectory={trajectory} />
          <TrajectoryLimitations trajectory={trajectory} />
        </>
      ) : null}

      {isPending && !error ? (
        <p className='mt-2.5 text-xs text-slate-400'>
          Loading trajectory quality evidence…
        </p>
      ) : null}

      {isNotFound ? (
        <p className='mt-2.5 rounded-lg border border-slate-700 bg-slate-900/70 p-2.5 text-xs leading-5 text-slate-300'>
          No trajectory has been built for this aircraft yet. GFA does not draw
          a historical trail until persisted evidence exists.
        </p>
      ) : null}

      {error && !isNotFound ? (
        <div className='mt-2.5 rounded-lg border border-amber-400/30 bg-amber-400/10 p-3'>
          <p className='text-xs leading-5 text-amber-100'>
            {getRequestErrorMessage(error)}
          </p>
          <button
            type='button'
            onClick={onRetry}
            disabled={isFetching}
            className='mt-2 rounded-md border border-amber-300/40 px-2.5 py-1.5 text-xs font-medium text-amber-100 transition hover:bg-amber-300/10 disabled:opacity-60'
          >
            Retry trajectory
          </button>
        </div>
      ) : null}
    </section>
  )
}

function Metric({
  label,
  value,
  className = '',
}: {
  label: string
  value: string
  className?: string
}) {
  if (!value.trim()) return null

  return (
    <div className={className}>
      <dt className='text-[9px] uppercase tracking-wide text-slate-600'>{label}</dt>
      <dd className='mt-0.5 break-words text-[11px] font-semibold text-slate-300'>
        {value}
      </dd>
    </div>
  )
}

function SegmentStatusSummary({ trajectory }: { trajectory: AircraftTrajectory }) {
  const counts = countSegmentStatuses(trajectory)

  return (
    <div className='mt-2 grid grid-cols-2 gap-1.5 text-[10px]'>
      {(
        [
          ['observed', 'Observed'],
          ['interpolated', 'Interpolated'],
          ['estimated', 'Estimated'],
          ['invalid', 'Invalid'],
        ] as const
      ).map(([status, label]) => (
        <div
          key={status}
          className='flex items-center justify-between rounded-md border border-slate-800 bg-slate-900/50 px-2 py-1.5'
        >
          <span className='text-slate-500'>{label}</span>
          <span className='font-semibold text-slate-200'>{counts[status]}</span>
        </div>
      ))}
    </div>
  )
}

function TrajectoryLimitations({ trajectory }: { trajectory: AircraftTrajectory }) {
  const counts = countSegmentStatuses(trajectory)
  const notices: string[] = []

  if (trajectory.segments.length === 0) {
    notices.push('The trajectory contains no drawable segment geometry.')
  }
  if (counts.interpolated > 0) {
    notices.push(
      `${counts.interpolated} segment${counts.interpolated === 1 ? '' : 's'} use interpolation rather than direct observation.`
    )
  }
  if (counts.estimated > 0) {
    notices.push(
      `${counts.estimated} segment${counts.estimated === 1 ? '' : 's'} are estimated and should not be treated as measured positions.`
    )
  }
  if (counts.invalid > 0) {
    notices.push(
      `${counts.invalid} segment${counts.invalid === 1 ? '' : 's'} are marked invalid by the trajectory pipeline.`
    )
  }
  if (trajectory.coverage_gap_count > 0) {
    notices.push(
      `${trajectory.coverage_gap_count} coverage gap${trajectory.coverage_gap_count === 1 ? '' : 's'} interrupt the observed track.`
    )
  }

  if (notices.length === 0 && trajectory.coverage_gaps.length === 0) return null

  return (
    <div className='mt-2.5 rounded-lg border border-amber-400/25 bg-amber-400/5 p-2.5'>
      <h5 className='text-[9px] font-bold uppercase tracking-wide text-amber-200'>
        Data limitations
      </h5>
      {notices.length > 0 ? (
        <ul className='mt-1.5 space-y-1 text-xs leading-5 text-amber-100'>
          {notices.map(notice => <li key={notice}>{notice}</li>)}
        </ul>
      ) : null}
      {trajectory.coverage_gaps.length > 0 ? (
        <div className='mt-2 space-y-1.5'>
          {trajectory.coverage_gaps.slice(0, 5).map(gap => (
            <CoverageGapItem key={gap.id} gap={gap} />
          ))}
          {trajectory.coverage_gaps.length > 5 ? (
            <p className='text-[10px] text-amber-200/80'>
              {trajectory.coverage_gaps.length - 5} additional gaps omitted from
              this compact panel.
            </p>
          ) : null}
        </div>
      ) : null}
    </div>
  )
}

function CoverageGapItem({ gap }: { gap: CoverageGap }) {
  const distance = formatDistance(gap.distance_km)
  const filledBy = gap.filled_by.trim()

  return (
    <div className='rounded-md border border-amber-300/20 bg-slate-950/50 p-2'>
      <div className='flex items-center justify-between gap-3'>
        <span className='text-[10px] font-semibold text-amber-100'>
          {formatGapReason(gap.reason)}
        </span>
        <span className='text-[10px] text-slate-400'>
          {formatDuration(gap.duration_seconds)}
        </span>
      </div>
      {distance || filledBy ? (
        <p className='mt-1 text-[10px] leading-4 text-slate-400'>
          {[distance ? `Distance ${distance}` : '', filledBy ? `Filled by ${filledBy}` : '']
            .filter(Boolean)
            .join(' · ')}
        </p>
      ) : null}
    </div>
  )
}

function countSegmentStatuses(
  trajectory: AircraftTrajectory
): Record<TrajectorySegmentStatus, number> {
  const counts: Record<TrajectorySegmentStatus, number> = {
    observed: 0,
    interpolated: 0,
    estimated: 0,
    invalid: 0,
  }

  for (const segment of trajectory.segments) {
    counts[segment.status]++
  }

  return counts
}

function normalizeQualityScore(value: number): number | null {
  if (!Number.isFinite(value) || value < 0 || value > 1) return null
  return value
}

function formatQualityScore(value: number): string {
  const normalized = normalizeQualityScore(value)
  if (normalized === null) return ''

  return new Intl.NumberFormat(undefined, {
    style: 'percent',
    maximumFractionDigits: 1,
  }).format(normalized)
}

function formatGapReason(reason: CoverageGap['reason']): string {
  if (reason === 'time_gap') return 'Time gap'
  if (reason === 'movement_jump') return 'Movement jump'
  return 'Coverage gap'
}

function formatRatio(value: number): string {
  if (!Number.isFinite(value)) return 'n/a'
  return new Intl.NumberFormat(undefined, {
    style: 'percent',
    maximumFractionDigits: 1,
  }).format(value)
}

function formatDistance(value: number): string {
  if (!Number.isFinite(value)) return ''
  return `${new Intl.NumberFormat(undefined, {
    maximumFractionDigits: 1,
  }).format(value)} km`
}

function formatDuration(value: number): string {
  if (!Number.isFinite(value) || value < 0) return ''

  const totalSeconds = Math.round(value)
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60

  if (hours > 0) return `${hours}h ${minutes}m`
  if (minutes > 0) return `${minutes}m ${seconds}s`
  return `${seconds}s`
}

function formatTimestamp(value: string): string {
  const timestamp = new Date(value)
  if (Number.isNaN(timestamp.getTime())) return ''
  return timestamp.toLocaleString()
}
