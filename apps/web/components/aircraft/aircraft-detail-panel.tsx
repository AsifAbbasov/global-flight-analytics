'use client'

import {
  APIRequestError,
  getRequestErrorMessage,
} from '@/lib/api/client'
import { formatTrafficAltitude } from '@/lib/traffic/altitude'
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
      <aside className='flex min-h-[360px] items-center justify-center rounded-xl border border-dashed border-slate-700 bg-slate-950/70 p-6 text-center'>
        <div>
          <p className='text-sm font-semibold text-slate-200'>
            No aircraft selected
          </p>
          <p className='mt-2 text-sm leading-6 text-slate-400'>
            Select an aircraft marker on the map to inspect its live
            observation, probable route and airport context, registered
            profile and latest trajectory.
          </p>
        </div>
      </aside>
    )
  }

  const displayName =
    aircraft?.callsign.trim() ||
    profileQuery.data?.registration.trim() ||
    selectedICAO24.toUpperCase()

  return (
    <aside
      className='rounded-xl border border-slate-700 bg-slate-950/95 p-5'
      aria-labelledby='aircraft-detail-title'
    >
      <div className='flex items-start justify-between gap-4'>
        <div>
          <p className='text-xs font-semibold uppercase tracking-[0.18em] text-sky-300'>
            Aircraft detail
          </p>
          <h3
            id='aircraft-detail-title'
            className='mt-2 text-xl font-semibold text-white'
          >
            {displayName}
          </h3>
          <p className='mt-1 font-mono text-xs uppercase text-slate-400'>
            ICAO24 {selectedICAO24}
          </p>
        </div>

        <button
          type='button'
          onClick={onClose}
          className='rounded-md border border-slate-700 px-3 py-1.5 text-sm font-medium text-slate-300 transition hover:bg-slate-800'
          aria-label='Close aircraft details'
        >
          Close
        </button>
      </div>

      <section className='mt-5' aria-labelledby='live-observation-title'>
        <h4
          id='live-observation-title'
          className='text-sm font-semibold text-slate-200'
        >
          Live observation
        </h4>

        {aircraft ? (
          <dl className='mt-3 grid grid-cols-2 gap-x-4 gap-y-3 text-sm'>
            <Detail label='Callsign' value={aircraft.callsign} />
            <Detail
              label='Status'
              value={aircraft.on_ground ? 'On ground' : 'In air'}
            />
            <Detail
              label='Altitude'
              value={formatTrafficAltitude(aircraft)}
            />
            <Detail
              label='Speed'
              value={formatSpeed(aircraft.velocity_mps)}
            />
            <Detail
              label='Heading'
              value={formatHeading(aircraft.heading_degrees)}
            />
            <Detail
              label='Origin country'
              value={aircraft.origin_country}
            />
            <Detail
              label='Latitude'
              value={formatCoordinate(aircraft.latitude)}
            />
            <Detail
              label='Longitude'
              value={formatCoordinate(aircraft.longitude)}
            />
            <div className='col-span-2'>
              <Detail
                label='Observed'
                value={formatTimestamp(aircraft.observed_at)}
              />
            </div>
          </dl>
        ) : (
          <p className='mt-3 rounded-lg border border-amber-400/30 bg-amber-400/10 p-3 text-sm leading-6 text-amber-100'>
            This aircraft is no longer present in the latest regional
            traffic response. Cached profile, route context and
            trajectory information remain available below.
          </p>
        )}
      </section>

      <RouteContextSection
        routeContext={routeContext}
        isPending={routeContextIsPending}
        isFetching={routeContextIsFetching}
        error={routeContextError}
        onRetry={onRetryRouteContext}
      />

      <TrajectorySection
        trajectory={trajectory}
        isPending={trajectoryIsPending}
        isFetching={trajectoryIsFetching}
        error={trajectoryError}
        onRetry={onRetryTrajectory}
      />

      <section
        className='mt-6 border-t border-slate-800 pt-5'
        aria-labelledby='registered-profile-title'
      >
        <div className='flex items-center justify-between gap-3'>
          <h4
            id='registered-profile-title'
            className='text-sm font-semibold text-slate-200'
          >
            Registered profile
          </h4>

          {profileQuery.isFetching ? (
            <span className='text-xs text-sky-300'>Loading…</span>
          ) : null}
        </div>

        {profileQuery.data ? (
          <dl className='mt-3 grid grid-cols-2 gap-x-4 gap-y-3 text-sm'>
            <Detail
              label='Registration'
              value={profileQuery.data.registration}
            />
            <Detail
              label='Type'
              value={profileQuery.data.aircraft_type}
            />
            <Detail
              label='Manufacturer'
              value={profileQuery.data.manufacturer}
            />
            <Detail label='Model' value={profileQuery.data.model} />
            <Detail
              label='Airline'
              value={profileQuery.data.airline}
            />
            <Detail
              label='Country'
              value={profileQuery.data.country}
            />
          </dl>
        ) : null}

        {profileQuery.isPending ? (
          <p className='mt-3 text-sm text-slate-400'>
            Loading the registered aircraft profile…
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
    </aside>
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
  const isNotFound =
    error instanceof APIRequestError && error.status === 404

  return (
    <section
      className='mt-6 border-t border-slate-800 pt-5'
      aria-labelledby='route-context-title'
    >
      <div className='flex items-center justify-between gap-3'>
        <h4
          id='route-context-title'
          className='text-sm font-semibold text-slate-200'
        >
          Probable route and airport context
        </h4>

        {isFetching ? (
          <span className='text-xs text-sky-300'>Updating…</span>
        ) : null}
      </div>

      {routeContext ? (
        <>
          <div className='mt-3 rounded-lg border border-slate-800 bg-slate-900/70 p-3'>
            <div className='flex items-center justify-between gap-3'>
              <span className='text-xs uppercase tracking-wide text-slate-500'>
                Route confidence
              </span>
              <ConfidenceBadge
                level={routeContext.confidence.level}
                score={routeContext.confidence.score}
              />
            </div>

            <div
              className='mt-2 h-2 overflow-hidden rounded-full bg-slate-800'
              role='progressbar'
              aria-label='Probable route confidence score'
              aria-valuemin={0}
              aria-valuemax={100}
              aria-valuenow={Math.round(
                routeContext.confidence.score * 100
              )}
            >
              <div
                className='h-full rounded-full bg-emerald-400'
                style={{
                  width: `${routeContext.confidence.score * 100}%`,
                }}
              />
            </div>

            <p className='mt-3 text-xs leading-5 text-slate-400'>
              Generated {formatTimestamp(routeContext.generated_at)} from
              persisted trajectory {routeContext.trajectory_id}.
            </p>
          </div>

          <div className='mt-3 grid gap-3'>
            <AirportCandidateCard
              label='Probable origin'
              candidate={routeContext.origin}
            />
            <AirportCandidateCard
              label='Probable destination'
              candidate={routeContext.destination}
            />
          </div>

          <div className='mt-4 rounded-lg border border-amber-400/25 bg-amber-400/5 p-3'>
            <h5 className='text-xs font-semibold uppercase tracking-wide text-amber-200'>
              Route limitations
            </h5>

            {routeContext.limitations.length > 0 ? (
              <ul className='mt-2 space-y-2 text-sm leading-5 text-amber-100'>
                {routeContext.limitations.map(limitation => (
                  <li key={limitation.code}>
                    {limitation.message}
                  </li>
                ))}
              </ul>
            ) : (
              <p className='mt-2 text-sm leading-5 text-slate-300'>
                No route-context limitations were reported.
              </p>
            )}
          </div>
        </>
      ) : null}

      {isPending && !error ? (
        <p className='mt-3 text-sm text-slate-400'>
          Inferring probable airport candidates from the persisted
          trajectory endpoints…
        </p>
      ) : null}

      {isNotFound ? (
        <p className='mt-3 rounded-lg border border-slate-700 bg-slate-900/70 p-3 text-sm leading-6 text-slate-300'>
          Route context is unavailable because no persisted trajectory
          exists for this aircraft yet.
        </p>
      ) : null}

      {error && !isNotFound ? (
        <div className='mt-3 rounded-lg border border-amber-400/30 bg-amber-400/10 p-3'>
          <p className='text-sm leading-6 text-amber-100'>
            {getRequestErrorMessage(error)}
          </p>
          <button
            type='button'
            onClick={onRetry}
            disabled={isFetching}
            className='mt-3 rounded-md border border-amber-300/40 px-3 py-1.5 text-sm font-medium text-amber-100 transition hover:bg-amber-300/10 disabled:cursor-not-allowed disabled:opacity-60'
          >
            Retry route context
          </button>
        </div>
      ) : null}
    </section>
  )
}

function AirportCandidateCard({
  label,
  candidate,
}: {
  label: string
  candidate: RouteContextAirportCandidate | undefined
}) {
  if (!candidate) {
    return (
      <div className='rounded-lg border border-dashed border-slate-700 bg-slate-900/40 p-3'>
        <p className='text-xs uppercase tracking-wide text-slate-500'>
          {label}
        </p>
        <p className='mt-2 text-sm leading-5 text-slate-400'>
          No airport candidate is close enough to the corresponding
          trajectory endpoint.
        </p>
      </div>
    )
  }

  const airportCode = [
    candidate.airport.icao_code,
    candidate.airport.iata_code,
  ]
    .filter(Boolean)
    .join(' / ')

  return (
    <div className='rounded-lg border border-slate-800 bg-slate-900/60 p-3'>
      <div className='flex items-start justify-between gap-3'>
        <div>
          <p className='text-xs uppercase tracking-wide text-slate-500'>
            {label}
          </p>
          <p className='mt-1 text-sm font-semibold text-white'>
            {candidate.airport.name}
          </p>
          <p className='mt-1 font-mono text-xs text-sky-300'>
            {airportCode}
          </p>
        </div>

        <ConfidenceBadge
          level={candidate.confidence.level}
          score={candidate.confidence.score}
        />
      </div>

      <dl className='mt-3 grid grid-cols-2 gap-x-4 gap-y-3 text-sm'>
        <Detail
          label='Location'
          value={[
            candidate.airport.city,
            candidate.airport.country,
          ]
            .filter(Boolean)
            .join(', ')}
        />
        <Detail
          label='Endpoint distance'
          value={formatDistance(candidate.distance_km)}
        />
        <Detail
          label='Elevation'
          value={formatAltitude(candidate.airport.elevation_m)}
        />
        <Detail
          label='Timezone'
          value={candidate.airport.timezone}
        />
      </dl>

      {candidate.confidence.reasons.length > 0 ? (
        <ul className='mt-3 space-y-1 text-xs leading-5 text-slate-400'>
          {candidate.confidence.reasons.map(reason => (
            <li key={reason.code}>{reason.message}</li>
          ))}
        </ul>
      ) : null}
    </div>
  )
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
      className={`rounded-full border px-2.5 py-1 text-xs font-semibold uppercase tracking-wide ${className}`}
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
  const isNotFound =
    error instanceof APIRequestError && error.status === 404

  return (
    <section
      className='mt-6 border-t border-slate-800 pt-5'
      aria-labelledby='trajectory-quality-title'
    >
      <div className='flex items-center justify-between gap-3'>
        <h4
          id='trajectory-quality-title'
          className='text-sm font-semibold text-slate-200'
        >
          Latest trajectory
        </h4>

        {isFetching ? (
          <span className='text-xs text-sky-300'>Updating…</span>
        ) : null}
      </div>

      {trajectory ? (
        <>
          <div className='mt-3 rounded-lg border border-slate-800 bg-slate-900/70 p-3'>
            <div className='flex items-center justify-between gap-3'>
              <span className='text-xs uppercase tracking-wide text-slate-500'>
                Track quality
              </span>
              <span className='text-sm font-semibold text-white'>
                {formatQualityScore(trajectory.quality_score)}
              </span>
            </div>

            {normalizeQualityScore(trajectory.quality_score) !== null ? (
              <div
                className='mt-2 h-2 overflow-hidden rounded-full bg-slate-800'
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
                    width: `${
                      normalizeQualityScore(
                        trajectory.quality_score
                      )! * 100
                    }%`,
                  }}
                />
              </div>
            ) : null}

            <dl className='mt-3 grid grid-cols-2 gap-x-4 gap-y-3 text-sm'>
              <Detail
                label='Segments'
                value={String(trajectory.segment_count)}
              />
              <Detail
                label='Points'
                value={String(trajectory.point_count)}
              />
              <Detail
                label='Coverage gaps'
                value={String(trajectory.coverage_gap_count)}
              />
              <Detail
                label='Duration'
                value={formatDuration(
                  trajectory.duration_seconds
                )}
              />
              <Detail
                label='Started'
                value={formatTimestamp(trajectory.start_time)}
              />
              <Detail
                label='Updated'
                value={formatTimestamp(trajectory.updated_at)}
              />
              <div className='col-span-2'>
                <Detail
                  label='Source'
                  value={trajectory.source_name}
                />
              </div>
            </dl>
          </div>

          <SegmentStatusSummary trajectory={trajectory} />
          <TrajectoryLimitations trajectory={trajectory} />
        </>
      ) : null}

      {isPending && !error ? (
        <p className='mt-3 text-sm text-slate-400'>
          Loading the latest trajectory and quality evidence…
        </p>
      ) : null}

      {isNotFound ? (
        <p className='mt-3 rounded-lg border border-slate-700 bg-slate-900/70 p-3 text-sm leading-6 text-slate-300'>
          No trajectory has been built for this aircraft yet. The map
          will show a route line when recent points form a persisted
          trajectory.
        </p>
      ) : null}

      {error && !isNotFound ? (
        <div className='mt-3 rounded-lg border border-amber-400/30 bg-amber-400/10 p-3'>
          <p className='text-sm leading-6 text-amber-100'>
            {getRequestErrorMessage(error)}
          </p>
          <button
            type='button'
            onClick={onRetry}
            disabled={isFetching}
            className='mt-3 rounded-md border border-amber-300/40 px-3 py-1.5 text-sm font-medium text-amber-100 transition hover:bg-amber-300/10 disabled:cursor-not-allowed disabled:opacity-60'
          >
            Retry trajectory
          </button>
        </div>
      ) : null}
    </section>
  )
}

function SegmentStatusSummary({
  trajectory,
}: {
  trajectory: AircraftTrajectory
}) {
  const counts = countSegmentStatuses(trajectory)

  return (
    <div className='mt-3 grid grid-cols-2 gap-2 text-xs'>
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
          className='flex items-center justify-between rounded-md border border-slate-800 bg-slate-900/50 px-2.5 py-2'
        >
          <span className='text-slate-400'>{label}</span>
          <span className='font-semibold text-slate-200'>
            {counts[status]}
          </span>
        </div>
      ))}
    </div>
  )
}

function TrajectoryLimitations({
  trajectory,
}: {
  trajectory: AircraftTrajectory
}) {
  const counts = countSegmentStatuses(trajectory)
  const notices: string[] = []

  if (trajectory.segments.length === 0) {
    notices.push(
      'The trajectory contains no drawable segment geometry.'
    )
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
      `${trajectory.coverage_gap_count} coverage gap${trajectory.coverage_gap_count === 1 ? '' : 's'} interrupt the observed track. Gaps are explained below and are not drawn as continuous geometry.`
    )
  }

  return (
    <div className='mt-4 rounded-lg border border-amber-400/25 bg-amber-400/5 p-3'>
      <h5 className='text-xs font-semibold uppercase tracking-wide text-amber-200'>
        Data limitations
      </h5>

      {notices.length > 0 ? (
        <ul className='mt-2 space-y-2 text-sm leading-5 text-amber-100'>
          {notices.map(notice => (
            <li key={notice}>{notice}</li>
          ))}
        </ul>
      ) : (
        <p className='mt-2 text-sm leading-5 text-slate-300'>
          No coverage gaps or non-observed segment statuses are
          reported for the latest trajectory.
        </p>
      )}

      {trajectory.coverage_gaps.length > 0 ? (
        <div className='mt-3 space-y-2'>
          {trajectory.coverage_gaps.slice(0, 5).map(gap => (
            <CoverageGapItem key={gap.id} gap={gap} />
          ))}

          {trajectory.coverage_gaps.length > 5 ? (
            <p className='text-xs text-amber-200/80'>
              {trajectory.coverage_gaps.length - 5} additional gaps
              are omitted from this compact panel.
            </p>
          ) : null}
        </div>
      ) : null}
    </div>
  )
}

function CoverageGapItem({ gap }: { gap: CoverageGap }) {
  return (
    <div className='rounded-md border border-amber-300/20 bg-slate-950/50 p-2.5'>
      <div className='flex items-center justify-between gap-3'>
        <span className='text-xs font-semibold text-amber-100'>
          {formatGapReason(gap.reason)}
        </span>
        <span className='text-xs text-slate-400'>
          {formatDuration(gap.duration_seconds)}
        </span>
      </div>
      <p className='mt-1 text-xs leading-5 text-slate-400'>
        Distance: {formatDistance(gap.distance_km)}. Filled by:{' '}
        {gap.filled_by.trim() || 'nothing'}.
      </p>
    </div>
  )
}

interface DetailProps {
  label: string
  value: string
}

function Detail({ label, value }: DetailProps) {
  return (
    <div>
      <dt className='text-xs uppercase tracking-wide text-slate-500'>
        {label}
      </dt>
      <dd className='mt-1 break-words text-slate-200'>
        {value.trim() || 'Unknown'}
      </dd>
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
  if (!Number.isFinite(value) || value < 0 || value > 1) {
    return null
  }

  return value
}

function formatQualityScore(value: number): string {
  const normalized = normalizeQualityScore(value)

  if (normalized !== null) {
    return new Intl.NumberFormat(undefined, {
      style: 'percent',
      maximumFractionDigits: 1,
    }).format(normalized)
  }

  if (!Number.isFinite(value)) {
    return 'Unknown'
  }

  return new Intl.NumberFormat(undefined, {
    maximumFractionDigits: 3,
  }).format(value)
}

function formatGapReason(
  reason: CoverageGap['reason']
): string {
  if (reason === 'time_gap') {
    return 'Time gap'
  }

  if (reason === 'movement_jump') {
    return 'Movement jump'
  }

  return 'Unknown gap'
}

function formatAltitude(value: number): string {
  if (!Number.isFinite(value)) {
    return 'Unknown'
  }

  return `${new Intl.NumberFormat().format(Math.round(value))} m`
}

function formatSpeed(value: number): string {
  if (!Number.isFinite(value)) {
    return 'Unknown'
  }

  const kilometersPerHour = value * 3.6

  return `${new Intl.NumberFormat(undefined, {
    maximumFractionDigits: 0,
  }).format(kilometersPerHour)} km/h`
}

function formatHeading(value: number): string {
  if (!Number.isFinite(value)) {
    return 'Unknown'
  }

  const normalized = ((value % 360) + 360) % 360
  return `${Math.round(normalized)}°`
}

function formatCoordinate(value: number): string {
  if (!Number.isFinite(value)) {
    return 'Unknown'
  }

  return value.toFixed(5)
}

function formatRatio(value: number): string {
  if (!Number.isFinite(value)) {
    return 'Unknown'
  }

  return new Intl.NumberFormat(undefined, {
    style: 'percent',
    maximumFractionDigits: 1,
  }).format(value)
}

function formatDistance(value: number): string {
  if (!Number.isFinite(value)) {
    return 'Unknown distance'
  }

  return `${new Intl.NumberFormat(undefined, {
    maximumFractionDigits: 1,
  }).format(value)} km`
}

function formatDuration(value: number): string {
  if (!Number.isFinite(value) || value < 0) {
    return 'Unknown'
  }

  const totalSeconds = Math.round(value)
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60

  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }

  if (minutes > 0) {
    return `${minutes}m ${seconds}s`
  }

  return `${seconds}s`
}

function formatTimestamp(value: string): string {
  const timestamp = new Date(value)

  if (Number.isNaN(timestamp.getTime())) {
    return 'Unknown'
  }

  return timestamp.toLocaleString()
}
