// FRONTEND_REGIONAL_TRAFFIC_BRIEF_V1
'use client'

import { useMemo } from 'react'

import {
  buildRegionalTrafficBrief,
  type RankedTrafficGroup,
  type RegionalAltitudeBand,
  type RegionalAltitudeBandKey,
} from '@/lib/traffic/regional-traffic-brief-model'
import type { TrafficAircraft } from '@/types/traffic'

interface RegionalTrafficBriefProps {
  aircraft: TrafficAircraft[]
  regionName: string
  isFetching: boolean
}

export function RegionalTrafficBrief({
  aircraft,
  regionName,
  isFetching,
}: RegionalTrafficBriefProps) {
  const model = useMemo(
    () => buildRegionalTrafficBrief(aircraft),
    [aircraft]
  )

  return (
    <section
      className='mt-8 rounded-2xl border border-white/10 bg-slate-900/65 p-4 shadow-xl shadow-black/10 sm:p-6'
      aria-labelledby='regional-traffic-brief-title'
      aria-busy={isFetching}
    >
      <div className='flex flex-wrap items-start justify-between gap-4'>
        <div>
          <p className='text-xs font-semibold uppercase tracking-[0.2em] text-sky-300'>
            Current snapshot interpretation
          </p>
          <h2
            id='regional-traffic-brief-title'
            className='mt-2 text-xl font-semibold text-white'
          >
            Regional Traffic Brief — {regionName}
          </h2>
          <p className='mt-2 max-w-3xl text-sm leading-6 text-slate-400'>
            A deterministic browser-side summary of the current API snapshot.
            It describes observed composition only and does not infer historical
            trends or operational conditions.
          </p>
        </div>
        {isFetching ? (
          <span className='rounded-full border border-sky-400/30 bg-sky-400/10 px-3 py-1 text-xs font-medium text-sky-200'>
            Updating snapshot…
          </span>
        ) : null}
      </div>

      {model.totalCount === 0 ? (
        <p className='mt-5 rounded-xl border border-dashed border-slate-700 p-5 text-sm leading-6 text-slate-400'>
          No aircraft are available for a regional brief in the current snapshot.
        </p>
      ) : (
        <div className='mt-5 grid gap-4 xl:grid-cols-4'>
          <SnapshotComposition
            totalCount={model.totalCount}
            airborneCount={model.airborneCount}
            groundCount={model.groundCount}
            knownAltitudeCount={model.knownAltitudeCount}
            altitudeCoverage={model.altitudeCoverage}
          />
          <AltitudeProfile bands={model.altitudeBands} />
          <RankingCard
            title='Leading airlines'
            description='Attributed airline labels in the current snapshot.'
            items={model.topAirlines}
            unknownCount={model.unknownAirlineCount}
            emptyMessage='No airline attribution is available.'
          />
          <RankingCard
            title='Origin countries'
            description='Provider-supplied aircraft origin-country labels.'
            items={model.topOriginCountries}
            unknownCount={model.unknownOriginCountryCount}
            emptyMessage='No origin-country attribution is available.'
          />
        </div>
      )}
    </section>
  )
}

function SnapshotComposition({
  totalCount,
  airborneCount,
  groundCount,
  knownAltitudeCount,
  altitudeCoverage,
}: {
  totalCount: number
  airborneCount: number
  groundCount: number
  knownAltitudeCount: number
  altitudeCoverage: number
}) {
  return (
    <article className='rounded-xl border border-slate-800 bg-slate-950/65 p-4'>
      <CardHeading
        title='Snapshot composition'
        description='Observed aircraft status and altitude completeness.'
      />
      <dl className='mt-4 grid grid-cols-2 gap-3'>
        <Metric label='Total' value={formatInteger(totalCount)} />
        <Metric label='Airborne' value={formatInteger(airborneCount)} />
        <Metric label='On ground' value={formatInteger(groundCount)} />
        <Metric
          label='Known altitude'
          value={`${formatInteger(knownAltitudeCount)} · ${formatPercent(
            altitudeCoverage
          )}`}
        />
      </dl>
    </article>
  )
}

function AltitudeProfile({ bands }: { bands: RegionalAltitudeBand[] }) {
  return (
    <article className='rounded-xl border border-slate-800 bg-slate-950/65 p-4'>
      <CardHeading
        title='Airborne altitude profile'
        description='Shares use airborne aircraft as the denominator.'
      />
      <div className='mt-4 space-y-3'>
        {bands.map(band => (
          <div key={band.key}>
            <div className='flex items-end justify-between gap-3 text-xs'>
              <div>
                <p className='font-medium text-slate-200'>{band.label}</p>
                <p className='mt-0.5 text-[11px] text-slate-600'>{band.range}</p>
              </div>
              <p className='shrink-0 text-slate-400'>
                {formatInteger(band.count)} · {formatPercent(band.share)}
              </p>
            </div>
            <div
              className='mt-1.5 h-1.5 overflow-hidden rounded-full bg-slate-800'
              aria-hidden='true'
            >
              <div
                className={`h-full rounded-full ${altitudeBandClassName(
                  band.key
                )}`}
                style={{ width: `${Math.min(100, band.share * 100)}%` }}
              />
            </div>
          </div>
        ))}
      </div>
    </article>
  )
}

function RankingCard({
  title,
  description,
  items,
  unknownCount,
  emptyMessage,
}: {
  title: string
  description: string
  items: RankedTrafficGroup[]
  unknownCount: number
  emptyMessage: string
}) {
  return (
    <article className='rounded-xl border border-slate-800 bg-slate-950/65 p-4'>
      <CardHeading title={title} description={description} />
      {items.length > 0 ? (
        <ol className='mt-4 space-y-3'>
          {items.map((item, index) => (
            <li key={item.label} className='flex items-center gap-3'>
              <span className='flex h-7 w-7 shrink-0 items-center justify-center rounded-lg border border-slate-800 bg-slate-900 font-mono text-[11px] text-sky-300'>
                {String(index + 1).padStart(2, '0')}
              </span>
              <div className='min-w-0 flex-1'>
                <p className='truncate text-sm font-medium text-slate-200'>
                  {item.label}
                </p>
                <p className='mt-0.5 text-xs text-slate-500'>
                  {formatInteger(item.count)} aircraft · {formatPercent(item.share)}
                </p>
              </div>
            </li>
          ))}
        </ol>
      ) : (
        <p className='mt-4 text-sm leading-6 text-slate-500'>{emptyMessage}</p>
      )}
      {unknownCount > 0 ? (
        <p className='mt-4 border-t border-slate-800 pt-3 text-xs text-slate-600'>
          Unattributed records: {formatInteger(unknownCount)}
        </p>
      ) : null}
    </article>
  )
}

function CardHeading({
  title,
  description,
}: {
  title: string
  description: string
}) {
  return (
    <div>
      <h3 className='text-sm font-semibold text-white'>{title}</h3>
      <p className='mt-1 text-xs leading-5 text-slate-500'>{description}</p>
    </div>
  )
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className='rounded-lg border border-slate-800 bg-slate-900/60 p-3'>
      <dt className='text-[11px] uppercase tracking-wide text-slate-600'>
        {label}
      </dt>
      <dd className='mt-1 text-sm font-semibold text-slate-200'>{value}</dd>
    </div>
  )
}

function altitudeBandClassName(key: RegionalAltitudeBandKey): string {
  switch (key) {
    case 'low':
      return 'bg-emerald-300'
    case 'medium':
      return 'bg-sky-300'
    case 'high':
      return 'bg-violet-300'
    case 'unknown':
      return 'bg-slate-500'
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
