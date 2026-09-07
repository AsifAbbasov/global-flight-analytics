'use client'

import { useMemo, type ReactNode } from 'react'

import { buildAirportCongestionSummary } from '@/lib/analytics/airport-intelligence-workspace-model'
import type { AirportCongestionIntelligence } from '@/types/airport-intelligence'

export function AirportCongestionContent({
  congestion,
}: {
  congestion: AirportCongestionIntelligence
}) {
  const summary = useMemo(
    () => buildAirportCongestionSummary(congestion),
    [congestion]
  )

  return (
    <div
      className='space-y-4'
      data-airport-congestion-scope='relative-observed-activity-only'
      data-airport-capacity-model='none'
      data-airport-delay-inference='none'
    >
      <section className='rounded-xl border border-slate-800 bg-slate-950/60 p-4'>
        <div className='flex flex-wrap items-start justify-between gap-4'>
          <div className='max-w-2xl'>
            <p className='text-xs uppercase tracking-[0.16em] text-slate-600'>
              Relative observed activity
            </p>
            <h4 className='mt-2 text-xl font-semibold text-white'>
              Airport congestion intelligence
            </h4>
            <p className='mt-2 text-sm leading-6 text-slate-400'>
              {congestion.explanation}
            </p>
          </div>
          <span className='rounded-full border border-violet-400/30 bg-violet-400/10 px-3 py-1 text-xs text-violet-100'>
            {summary.status === 'available' ? 'Observed proxy available' : 'Proxy unavailable'}
          </span>
        </div>

        <div className='mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-4'>
          <MetricCard
            label='Observed activity score'
            value={summary.score === null ? 'Unavailable' : formatPercent(summary.score)}
          />
          <MetricCard
            label='Current / baseline'
            value={formatRatio(summary.currentToBaselineRatio)}
          />
          <MetricCard
            label='Current / prior peak'
            value={formatRatio(summary.currentToPriorPeakRatio)}
          />
          <MetricCard
            label='Current movement rate'
            value={`${formatDecimal(congestion.current.movements_per_hour)}/h`}
          />
        </div>

        <p className='mt-4 rounded-lg border border-amber-400/20 bg-amber-400/5 p-3 text-xs leading-5 text-amber-100/80'>
          This score compares persisted completed-day activity with this airport&apos;s own prior observed activity. It is not airport capacity, runway occupancy, queue length, slot pressure, delay evidence or an official operational congestion measure.
        </p>
      </section>

      <div className='grid gap-4 md:grid-cols-2'>
        <EvidenceCard
          title='Historical reference'
          description='The backend preserves the baseline and prior observed peak separately from the bounded score.'
        >
          <MetricRow
            label='Baseline windows'
            value={formatInteger(congestion.baseline_window_count)}
          />
          <MetricRow
            label='Median prior movements/hour'
            value={formatDecimal(congestion.baseline_median_movements_per_hour)}
          />
          <MetricRow
            label='Prior observed peak/hour'
            value={formatDecimal(congestion.prior_peak_movements_per_hour)}
          />
          <MetricRow
            label='Above prior observed peak'
            value={summary.exceedsPriorObservedActivityPeak ? 'Yes' : 'No'}
          />
        </EvidenceCard>

        <EvidenceCard
          title='Evidence support'
          description='Missing completed-day windows reduce support instead of being silently filled.'
        >
          <ProgressRow label='Window coverage' value={summary.evidenceCoverage} />
          <ProgressRow label='Conservative support' value={summary.evidenceSupport} />
          <MetricRow
            label='Observed / expected windows'
            value={`${formatInteger(summary.observedWindowCount)} / ${formatInteger(summary.expectedWindowCount)}`}
          />
          <MetricRow
            label='Gaps / trailing gaps'
            value={`${formatInteger(summary.gapWindowCount)} / ${formatInteger(summary.trailingGapWindowCount)}`}
          />
          <MetricRow
            label='Latest expected day observed'
            value={summary.currentWindowIsLatestExpected ? 'Yes' : 'No'}
          />
        </EvidenceCard>
      </div>

      <section className='rounded-xl border border-slate-800 bg-slate-950/50 p-4'>
        <p className='text-xs font-semibold uppercase tracking-[0.14em] text-slate-500'>
          Published score semantics
        </p>
        <p className='mt-2 text-sm leading-6 text-slate-400'>
          {congestion.score_semantics}
        </p>
        <p className='mt-3 font-mono text-[11px] leading-5 text-slate-600'>
          scope_guard={congestion.scope_guard}
        </p>
      </section>
    </div>
  )
}

function EvidenceCard({
  title,
  description,
  children,
}: {
  title: string
  description: string
  children: ReactNode
}) {
  return (
    <section className='rounded-xl border border-slate-800 bg-slate-950/60 p-4'>
      <h5 className='text-sm font-semibold text-slate-200'>{title}</h5>
      <p className='mt-1 text-xs leading-5 text-slate-600'>{description}</p>
      <div className='mt-4 space-y-3'>{children}</div>
    </section>
  )
}

function MetricCard({ label, value }: { label: string; value: string }) {
  return (
    <div className='rounded-lg border border-slate-800 bg-slate-950 p-3'>
      <p className='text-[11px] uppercase tracking-[0.12em] text-slate-600'>{label}</p>
      <p className='mt-2 text-lg font-semibold text-slate-100'>{value}</p>
    </div>
  )
}

function MetricRow({ label, value }: { label: string; value: string }) {
  return (
    <div className='flex items-center justify-between gap-4 text-xs'>
      <span className='text-slate-500'>{label}</span>
      <span className='text-right font-mono text-slate-300'>{value}</span>
    </div>
  )
}

function ProgressRow({ label, value }: { label: string; value: number }) {
  const normalized = Math.min(1, Math.max(0, Number.isFinite(value) ? value : 0))
  return (
    <div>
      <div className='flex items-center justify-between gap-3 text-xs'>
        <span className='text-slate-500'>{label}</span>
        <span className='font-mono text-slate-300'>{formatPercent(normalized)}</span>
      </div>
      <div className='mt-1 h-1.5 overflow-hidden rounded-full bg-slate-800'>
        <div
          className='h-full rounded-full bg-violet-300/70'
          style={{ width: `${normalized * 100}%` }}
        />
      </div>
    </div>
  )
}

function formatPercent(value: number): string {
  return `${Math.round(value * 100)}%`
}

function formatRatio(value: number | null): string {
  return value === null ? 'Unavailable' : `${formatDecimal(value)}×`
}

function formatDecimal(value: number): string {
  return Number.isFinite(value) ? value.toFixed(2) : 'Unavailable'
}

function formatInteger(value: number): string {
  return Number.isFinite(value) ? Math.max(0, Math.trunc(value)).toLocaleString('en-US') : '0'
}
