// FRONTEND_TRAFFIC_DATA_QUALITY_LENS_V1
'use client'

import { useMemo, type ReactNode } from 'react'

import {
  buildTrafficDataQualityModel,
  type TrafficDataQualityIssue,
  type TrafficDataQualitySeverity,
  type TrafficIdentityCompleteness,
} from '@/lib/traffic/traffic-data-quality-model'
import type { TrafficAircraft } from '@/types/traffic'

interface TrafficDataQualityLensProps {
  aircraft: TrafficAircraft[]
  regionName: string
  snapshotUpdatedAt: number
  isFetching: boolean
}

export function TrafficDataQualityLens({
  aircraft,
  regionName,
  snapshotUpdatedAt,
  isFetching,
}: TrafficDataQualityLensProps) {
  const model = useMemo(
    () => buildTrafficDataQualityModel(aircraft, snapshotUpdatedAt),
    [aircraft, snapshotUpdatedAt]
  )

  return (
    <section
      aria-labelledby='traffic-data-quality-title'
      aria-busy={isFetching}
      className='mt-8 rounded-2xl border border-white/10 bg-slate-900/65 p-4 shadow-xl shadow-black/10 sm:p-6'
    >
      <div className='flex flex-wrap items-start justify-between gap-4'>
        <div>
          <p className='text-xs font-semibold uppercase tracking-[0.2em] text-amber-300'>
            Snapshot evidence quality
          </p>
          <h2
            id='traffic-data-quality-title'
            className='mt-2 text-xl font-semibold text-white'
          >
            Data Quality Lens — {regionName}
          </h2>
          <p className='mt-2 max-w-4xl text-sm leading-6 text-slate-400'>
            Browser-side structural checks for the current API response. No
            composite confidence score is invented: coordinates, timestamps,
            motion, altitude and attribution remain visible as separate evidence
            dimensions.
          </p>
        </div>
        {isFetching ? (
          <span className='rounded-full border border-sky-400/30 bg-sky-400/10 px-3 py-1 text-xs font-medium text-sky-200'>
            Re-evaluating snapshot…
          </span>
        ) : null}
      </div>

      {model.totalCount === 0 ? (
        <p className='mt-5 rounded-xl border border-dashed border-slate-700 p-5 text-sm leading-6 text-slate-400'>
          No aircraft records are available for browser-side quality checks.
        </p>
      ) : (
        <>
          <div className='mt-5 grid gap-4 xl:grid-cols-4'>
            <QualitySummaryCard
              totalCount={model.totalCount}
              coreUsableCount={model.coreUsableCount}
              coreUsableShare={model.coreUsableShare}
              uniqueAircraftCount={model.uniqueAircraftCount}
              validIdentifierCount={model.validIdentifierCount}
            />
            <StructuralEvidenceCard
              validCoordinateCount={model.validCoordinateCount}
              validCoordinateShare={model.validCoordinateShare}
              validMotionCount={model.validMotionCount}
              validMotionShare={model.validMotionShare}
              totalCount={model.totalCount}
            />
            <ObservationEvidenceCard
              referenceTimeISO={model.referenceTimeISO}
              totalCount={model.totalCount}
              validTimestampCount={model.validObservationTimestampCount}
              validTimestampShare={model.validObservationTimestampShare}
              recentObservationCount={model.recentObservationCount}
              recentObservationShare={model.recentObservationShare}
              staleObservationCount={model.staleObservationCount}
              futureObservationCount={model.futureObservationCount}
            />
            <AltitudeEvidenceCard
              airborneCount={model.airborneCount}
              usableCount={model.usableAirborneAltitudeCount}
              usableShare={model.usableAirborneAltitudeShare}
            />
          </div>

          <div className='mt-4 grid gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(0,1.35fr)]'>
            <IdentityCompletenessCard metrics={model.identityCompleteness} />
            <IssueRegister issues={model.issues} />
          </div>
        </>
      )}

      <p className='mt-5 border-t border-slate-800 pt-4 text-xs leading-5 text-slate-600'>
        This lens does not replace server-side data-quality metrics, provider
        health, ingestion diagnostics or operational safety assessment. Observation
        recency is measured against the successful browser response timestamp using
        a five-minute window and one-minute clock-skew tolerance.
      </p>
    </section>
  )
}

function QualitySummaryCard({
  totalCount,
  coreUsableCount,
  coreUsableShare,
  uniqueAircraftCount,
  validIdentifierCount,
}: {
  totalCount: number
  coreUsableCount: number
  coreUsableShare: number
  uniqueAircraftCount: number
  validIdentifierCount: number
}) {
  return (
    <QualityCard
      title='Core structural usability'
      description='Records with valid ICAO24, coordinates, timestamp, motion and required altitude semantics.'
    >
      <PrimaryMetric
        value={formatPercent(coreUsableShare)}
        detail={`${formatInteger(coreUsableCount)} of ${formatInteger(totalCount)} records`}
      />
      <MetricRows
        rows={[
          {
            label: 'Unique valid aircraft',
            value: formatInteger(uniqueAircraftCount),
          },
          {
            label: 'Valid ICAO24 records',
            value: `${formatInteger(validIdentifierCount)} / ${formatInteger(totalCount)}`,
          },
        ]}
      />
    </QualityCard>
  )
}

function StructuralEvidenceCard({
  validCoordinateCount,
  validCoordinateShare,
  validMotionCount,
  validMotionShare,
  totalCount,
}: {
  validCoordinateCount: number
  validCoordinateShare: number
  validMotionCount: number
  validMotionShare: number
  totalCount: number
}) {
  return (
    <QualityCard
      title='Map and motion evidence'
      description='Independent checks for usable geographic points and physical motion values.'
    >
      <ProgressMetric
        label='Valid coordinates'
        count={validCoordinateCount}
        denominator={totalCount}
        share={validCoordinateShare}
      />
      <ProgressMetric
        label='Valid speed and heading'
        count={validMotionCount}
        denominator={totalCount}
        share={validMotionShare}
      />
    </QualityCard>
  )
}

function ObservationEvidenceCard({
  referenceTimeISO,
  totalCount,
  validTimestampCount,
  validTimestampShare,
  recentObservationCount,
  recentObservationShare,
  staleObservationCount,
  futureObservationCount,
}: {
  referenceTimeISO: string | null
  totalCount: number
  validTimestampCount: number
  validTimestampShare: number
  recentObservationCount: number
  recentObservationShare: number | null
  staleObservationCount: number
  futureObservationCount: number
}) {
  return (
    <QualityCard
      title='Observation recency'
      description={
        referenceTimeISO === null
          ? 'Waiting for a successful browser response timestamp.'
          : `Compared with ${formatTimestamp(referenceTimeISO)}.`
      }
    >
      <ProgressMetric
        label='Parseable timestamps'
        count={validTimestampCount}
        denominator={totalCount}
        share={validTimestampShare}
      />
      <MetricRows
        rows={[
          {
            label: 'Within five minutes',
            value:
              recentObservationShare === null
                ? 'Unavailable'
                : `${formatInteger(recentObservationCount)} · ${formatPercent(
                    recentObservationShare
                  )}`,
          },
          {
            label: 'Older observations',
            value: formatInteger(staleObservationCount),
          },
          {
            label: 'Future-dated records',
            value: formatInteger(futureObservationCount),
          },
        ]}
      />
    </QualityCard>
  )
}

function AltitudeEvidenceCard({
  airborneCount,
  usableCount,
  usableShare,
}: {
  airborneCount: number
  usableCount: number
  usableShare: number
}) {
  return (
    <QualityCard
      title='Airborne altitude evidence'
      description='Ground records are excluded from the altitude denominator.'
    >
      <PrimaryMetric
        value={airborneCount === 0 ? 'No airborne records' : formatPercent(usableShare)}
        detail={
          airborneCount === 0
            ? 'No altitude completeness claim is required.'
            : `${formatInteger(usableCount)} of ${formatInteger(airborneCount)} airborne records`
        }
      />
      <MetricRows
        rows={[
          {
            label: 'Altitude unavailable',
            value: formatInteger(airborneCount - usableCount),
          },
        ]}
      />
    </QualityCard>
  )
}

function IdentityCompletenessCard({
  metrics,
}: {
  metrics: TrafficIdentityCompleteness[]
}) {
  return (
    <QualityCard
      title='Identity attribution completeness'
      description='Provider-supplied descriptive labels; absence does not invalidate positional evidence.'
    >
      <div className='space-y-3'>
        {metrics.map(metric => (
          <ProgressMetric
            key={metric.key}
            label={metric.label}
            count={metric.presentCount}
            denominator={metric.presentCount + metric.missingCount}
            share={metric.share}
          />
        ))}
      </div>
    </QualityCard>
  )
}

function IssueRegister({ issues }: { issues: TrafficDataQualityIssue[] }) {
  return (
    <QualityCard
      title='Detected issue register'
      description='Sorted by severity, affected records and stable issue key.'
    >
      {issues.length === 0 ? (
        <p className='rounded-lg border border-emerald-400/20 bg-emerald-400/5 p-4 text-sm leading-6 text-emerald-100'>
          No browser-detectable structural issues were found in this snapshot.
        </p>
      ) : (
        <ol className='grid gap-3 md:grid-cols-2'>
          {issues.map(issue => (
            <li
              key={issue.key}
              className='rounded-lg border border-slate-800 bg-slate-950/55 p-3'
            >
              <div className='flex items-start justify-between gap-3'>
                <div>
                  <p className='text-sm font-semibold text-slate-200'>
                    {issue.label}
                  </p>
                  <p className='mt-1 text-xs leading-5 text-slate-500'>
                    {issue.description}
                  </p>
                </div>
                <SeverityBadge severity={issue.severity} />
              </div>
              <p className='mt-3 text-xs text-slate-400'>
                {formatInteger(issue.count)} of {formatInteger(issue.denominator)} ·{' '}
                {formatPercent(issue.share)}
              </p>
            </li>
          ))}
        </ol>
      )}
    </QualityCard>
  )
}

function QualityCard({
  title,
  description,
  children,
}: {
  title: string
  description: string
  children: ReactNode
}) {
  return (
    <article className='rounded-xl border border-slate-800 bg-slate-950/65 p-4'>
      <h3 className='text-sm font-semibold text-white'>{title}</h3>
      <p className='mt-1 text-xs leading-5 text-slate-500'>{description}</p>
      <div className='mt-4'>{children}</div>
    </article>
  )
}

function PrimaryMetric({ value, detail }: { value: string; detail: string }) {
  return (
    <div>
      <p className='text-2xl font-semibold tracking-tight text-white'>{value}</p>
      <p className='mt-1 text-xs leading-5 text-slate-500'>{detail}</p>
    </div>
  )
}

function ProgressMetric({
  label,
  count,
  denominator,
  share,
}: {
  label: string
  count: number
  denominator: number
  share: number
}) {
  return (
    <div>
      <div className='flex items-center justify-between gap-3 text-xs'>
        <span className='text-slate-400'>{label}</span>
        <span className='text-slate-300'>
          {formatInteger(count)} / {formatInteger(denominator)} · {formatPercent(share)}
        </span>
      </div>
      <div className='mt-1.5 h-1.5 overflow-hidden rounded-full bg-slate-800'>
        <div
          aria-hidden='true'
          className='h-full rounded-full bg-amber-300'
          style={{ width: `${Math.min(100, share * 100)}%` }}
        />
      </div>
    </div>
  )
}

function MetricRows({
  rows,
}: {
  rows: Array<{ label: string; value: string }>
}) {
  return (
    <dl className='mt-4 space-y-2 border-t border-slate-800 pt-3 text-xs'>
      {rows.map(row => (
        <div key={row.label} className='flex items-start justify-between gap-3'>
          <dt className='text-slate-500'>{row.label}</dt>
          <dd className='text-right font-medium text-slate-300'>{row.value}</dd>
        </div>
      ))}
    </dl>
  )
}

function SeverityBadge({ severity }: { severity: TrafficDataQualitySeverity }) {
  return (
    <span
      className={`shrink-0 rounded-full border px-2 py-1 text-[10px] font-semibold uppercase tracking-wide ${severityClassName(
        severity
      )}`}
    >
      {severity}
    </span>
  )
}

function severityClassName(severity: TrafficDataQualitySeverity): string {
  switch (severity) {
    case 'critical':
      return 'border-rose-400/30 bg-rose-400/10 text-rose-200'
    case 'warning':
      return 'border-amber-400/30 bg-amber-400/10 text-amber-200'
    case 'information':
      return 'border-sky-400/30 bg-sky-400/10 text-sky-200'
  }
}

function formatInteger(value: number): string {
  return new Intl.NumberFormat().format(value)
}

function formatPercent(value: number): string {
  return new Intl.NumberFormat(undefined, {
    style: 'percent',
    maximumFractionDigits: 0,
  }).format(value)
}

function formatTimestamp(value: string): string {
  return new Intl.DateTimeFormat(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(new Date(value))
}
