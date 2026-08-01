// FRONTEND_HISTORICAL_ANALYTICS_COMPARISON_V1
'use client'

import { useMemo, useState, type ChangeEvent, type ReactNode } from 'react'

import { getRequestErrorMessage } from '@/lib/api/client'
import {
  buildHistoricalSeriesView,
  buildPeriodComparisonView,
  compareHistoricalRecords,
  defaultMetricForScope,
  historicalGranularities,
  mergeHistoricalLimitations,
  metricDefinition,
  metricsForScope,
  normalizeHistoricalSelection,
  selectionIsComplete,
  sortAggregateHistory,
} from '@/lib/analytics/historical-analytics-comparison-model'
import {
  useHistoricalIntelligenceHistory,
  useLatestHistoricalIntelligence,
} from '@/lib/queries/historical-intelligence'
import type {
  HistoricalGranularity,
  HistoricalIntelligenceAggregateRecord,
  HistoricalIntelligenceSelection,
  HistoricalMetricName,
  HistoricalScopeType,
} from '@/types/historical-intelligence'

const historyLimit = 20

export function HistoricalAnalyticsComparisonWorkspace() {
  const [scope, setScope] = useState<HistoricalScopeType>('global')
  const [metric, setMetric] = useState<HistoricalMetricName>('active_aircraft')
  const [granularity, setGranularity] =
    useState<HistoricalGranularity>('day')
  const [airportICAO, setAirportICAO] = useState('')
  const [originICAO, setOriginICAO] = useState('')
  const [destinationICAO, setDestinationICAO] = useState('')
  const [leftRecordID, setLeftRecordID] = useState<string | null>(null)
  const [rightRecordID, setRightRecordID] = useState<string | null>(null)

  const draftSelection: HistoricalIntelligenceSelection = {
    scope,
    metric,
    granularity,
    ...(scope === 'airport' ? { airportICAO } : {}),
    ...(scope === 'route' ? { originICAO, destinationICAO } : {}),
  }
  const normalizedSelection = normalizeHistoricalSelection(draftSelection)
  const querySelection = selectionIsComplete(normalizedSelection)
    ? normalizedSelection
    : null

  const latestQuery = useLatestHistoricalIntelligence(querySelection)
  const historyQuery = useHistoricalIntelligenceHistory(
    querySelection,
    historyLimit
  )

  const seriesView = useMemo(
    () => buildHistoricalSeriesView(latestQuery.data?.result.points ?? []),
    [latestQuery.data?.result.points]
  )
  const periodComparison = useMemo(
    () => buildPeriodComparisonView(latestQuery.data?.result),
    [latestQuery.data?.result]
  )
  const history = useMemo(
    () => sortAggregateHistory(historyQuery.data?.items ?? []),
    [historyQuery.data?.items]
  )
  const leftRecord = history.find(record => record.id === leftRecordID)
  const rightRecord = history.find(record => record.id === rightRecordID)
  const recordComparison = compareHistoricalRecords(leftRecord, rightRecord)
  const limitations = useMemo(
    () =>
      mergeHistoricalLimitations(
        latestQuery.data?.result.limitations,
        ...history.map(record => record.result.limitations)
      ),
    [latestQuery.data?.result.limitations, history]
  )
  const metricMeta = metricDefinition(metric)
  const isFetching = latestQuery.isFetching || historyQuery.isFetching

  const changeScope = (nextScope: HistoricalScopeType) => {
    setScope(nextScope)
    setMetric(defaultMetricForScope(nextScope))
    setLeftRecordID(null)
    setRightRecordID(null)
  }

  const refresh = () => {
    void latestQuery.refetch()
    void historyQuery.refetch()
  }

  return (
    <section
      aria-labelledby='historical-analytics-title'
      aria-busy={isFetching}
      className='mt-8 rounded-2xl border border-white/10 bg-slate-900/65 p-4 shadow-2xl shadow-black/15 sm:p-6'
    >
      <div className='flex flex-wrap items-start justify-between gap-5'>
        <div>
          <p className='text-xs font-semibold uppercase tracking-[0.24em] text-cyan-300'>
            Historical intelligence and comparison
          </p>
          <h2
            id='historical-analytics-title'
            className='mt-3 text-2xl font-semibold tracking-tight text-white sm:text-3xl'
          >
            Compare persisted analytical evidence
          </h2>
          <p className='mt-3 max-w-4xl text-sm leading-6 text-slate-400'>
            Read the production Historical Intelligence aggregate store across
            global, airport and route scopes. Inspect bucket quality, compare the
            current analytical window with its previous period and compare two
            persisted aggregate records without recomputing server-owned metrics.
          </p>
        </div>
        <button
          type='button'
          onClick={refresh}
          disabled={querySelection === null || isFetching}
          className='rounded-lg border border-cyan-400/35 bg-cyan-400/5 px-4 py-2 text-sm font-medium text-cyan-100 transition hover:bg-cyan-400/10 disabled:cursor-not-allowed disabled:opacity-50'
        >
          {isFetching ? 'Refreshing history…' : 'Refresh history'}
        </button>
      </div>

      <HistoricalSelectionPanel
        scope={scope}
        metric={metric}
        granularity={granularity}
        airportICAO={airportICAO}
        originICAO={originICAO}
        destinationICAO={destinationICAO}
        onScopeChange={changeScope}
        onMetricChange={setMetric}
        onGranularityChange={setGranularity}
        onAirportICAOChange={setAirportICAO}
        onOriginICAOChange={setOriginICAO}
        onDestinationICAOChange={setDestinationICAO}
      />

      {querySelection === null ? (
        <PanelMessage title='Complete the historical scope'>
          {scope === 'airport'
            ? 'Enter a four-character airport ICAO code.'
            : 'Enter four-character origin and destination ICAO codes.'}
        </PanelMessage>
      ) : latestQuery.isPending ? (
        <PanelMessage title='Loading latest aggregate'>
          Reading the latest persisted {metricMeta.label.toLowerCase()} series.
        </PanelMessage>
      ) : latestQuery.error ? (
        <ErrorPanel
          title='Latest aggregate unavailable'
          error={latestQuery.error}
          onRetry={() => void latestQuery.refetch()}
        />
      ) : latestQuery.data ? (
        <>
          <div className='mt-5 grid gap-4 xl:grid-cols-[minmax(0,1.65fr)_minmax(320px,0.8fr)]'>
            <SeriesPanel
              record={latestQuery.data}
              maximumAbsoluteValue={seriesView.maximumAbsoluteValue}
            />
            <EvidenceSummaryPanel
              record={latestQuery.data}
              availableCount={seriesView.availableCount}
              partialCount={seriesView.partialCount}
              unavailableCount={seriesView.unavailableCount}
              minimumValue={seriesView.minimumValue}
              maximumValue={seriesView.maximumValue}
            />
          </div>
          <PeriodComparisonPanel
            record={latestQuery.data}
            comparison={periodComparison}
          />
        </>
      ) : null}

      <StoredAggregateComparisonPanel
        records={history}
        isPending={historyQuery.isPending}
        error={historyQuery.error}
        hasMore={historyQuery.data?.has_more ?? false}
        leftRecordID={leftRecordID}
        rightRecordID={rightRecordID}
        comparison={recordComparison}
        onLeftRecordChange={setLeftRecordID}
        onRightRecordChange={setRightRecordID}
        onRetry={() => void historyQuery.refetch()}
      />

      <LimitationsRegister limitations={limitations} />

      <p className='mt-5 border-t border-slate-800 pt-4 text-xs leading-5 text-slate-600'>
        Historical Intelligence is built from bounded open-data observations and
        persisted analytical outputs. The interface does not reconstruct missing
        buckets, invent region-level metrics unsupported by the production catalog,
        or convert unavailable comparisons into zero change.
      </p>
    </section>
  )
}

function HistoricalSelectionPanel({
  scope,
  metric,
  granularity,
  airportICAO,
  originICAO,
  destinationICAO,
  onScopeChange,
  onMetricChange,
  onGranularityChange,
  onAirportICAOChange,
  onOriginICAOChange,
  onDestinationICAOChange,
}: {
  scope: HistoricalScopeType
  metric: HistoricalMetricName
  granularity: HistoricalGranularity
  airportICAO: string
  originICAO: string
  destinationICAO: string
  onScopeChange: (value: HistoricalScopeType) => void
  onMetricChange: (value: HistoricalMetricName) => void
  onGranularityChange: (value: HistoricalGranularity) => void
  onAirportICAOChange: (value: string) => void
  onOriginICAOChange: (value: string) => void
  onDestinationICAOChange: (value: string) => void
}) {
  const metrics = metricsForScope(scope)
  return (
    <div className='mt-6 grid gap-3 rounded-xl border border-slate-800 bg-slate-950/65 p-4 md:grid-cols-3 xl:grid-cols-6'>
      <SelectField
        label='Scope'
        value={scope}
        onChange={value => onScopeChange(value as HistoricalScopeType)}
        options={[
          ['global', 'Global'],
          ['airport', 'Airport'],
          ['route', 'Route pair'],
        ]}
      />
      <div className='md:col-span-2'>
        <SelectField
          label='Metric'
          value={metric}
          onChange={value => onMetricChange(value as HistoricalMetricName)}
          options={metrics.map(item => [item.name, item.label])}
        />
      </div>
      <SelectField
        label='Granularity'
        value={granularity}
        onChange={value => onGranularityChange(value as HistoricalGranularity)}
        options={historicalGranularities.map(value => [value, titleCase(value)])}
      />
      {scope === 'airport' ? (
        <TextField
          label='Airport ICAO'
          value={airportICAO}
          placeholder='UBBB'
          onChange={onAirportICAOChange}
        />
      ) : null}
      {scope === 'route' ? (
        <>
          <TextField
            label='Origin ICAO'
            value={originICAO}
            placeholder='UBBB'
            onChange={onOriginICAOChange}
          />
          <TextField
            label='Destination ICAO'
            value={destinationICAO}
            placeholder='LTFM'
            onChange={onDestinationICAOChange}
          />
        </>
      ) : null}
      {scope === 'global' ? (
        <div className='md:col-span-2 rounded-lg border border-slate-800 bg-slate-950 p-3 text-xs leading-5 text-slate-500'>
          Global scope requires no identifier. Region scope is intentionally absent
          because the current production metric catalog does not allow any metric
          for it.
        </div>
      ) : null}
    </div>
  )
}

function SeriesPanel({
  record,
  maximumAbsoluteValue,
}: {
  record: HistoricalIntelligenceAggregateRecord
  maximumAbsoluteValue: number
}) {
  const result = record.result
  const points = buildHistoricalSeriesView(result.points).points
  return (
    <article className='rounded-xl border border-slate-800 bg-slate-950/65 p-4'>
      <div className='flex flex-wrap items-start justify-between gap-3'>
        <div>
          <p className='text-[11px] font-semibold uppercase tracking-[0.16em] text-slate-500'>
            Latest persisted series
          </p>
          <h3 className='mt-1 text-lg font-semibold text-white'>
            {metricDefinition(result.metric.name).label}
          </h3>
          <p className='mt-1 text-xs leading-5 text-slate-500'>
            {scopeLabel(result)} · {titleCase(result.granularity)} buckets ·{' '}
            {formatTimestamp(result.window.start_time)} to{' '}
            {formatTimestamp(result.window.end_time)}
          </p>
        </div>
        <StatusBadge value={result.status} />
      </div>

      {points.length === 0 ? (
        <PanelMessage title='No published buckets'>
          The stored aggregate contains no time-series points.
        </PanelMessage>
      ) : (
        <div className='mt-5 overflow-x-auto'>
          <div
            className='flex min-w-[680px] items-end gap-2 border-b border-l border-slate-800 px-3 pb-3 pt-8'
            style={{ height: '280px' }}
            role='img'
            aria-label={`${metricDefinition(result.metric.name).label} historical bar chart`}
          >
            {points.map(point => {
              const height = Math.max(
                point.status === 'unavailable' ? 2 : 6,
                Math.round((Math.abs(point.value) / maximumAbsoluteValue) * 210)
              )
              return (
                <div key={point.key} className='group flex min-w-0 flex-1 flex-col items-center justify-end'>
                  <div className='mb-2 hidden rounded-md border border-slate-700 bg-slate-950 px-2 py-1 text-[10px] text-slate-300 group-hover:block'>
                    {formatValue(point.value, result.metric.name)} · {formatPercent(point.coverageRatio)} coverage
                  </div>
                  <div
                    className={`w-full min-w-2 max-w-10 rounded-t ${barClassName(point.status)}`}
                    style={{ height: `${height}px` }}
                  />
                  <span className='mt-2 max-w-14 truncate text-[9px] text-slate-600'>
                    {formatBucketLabel(point.startTime, result.granularity)}
                  </span>
                </div>
              )
            })}
          </div>
          <div className='mt-3 flex flex-wrap gap-4 text-[11px] text-slate-500'>
            <LegendDot className='bg-cyan-300' label='Complete' />
            <LegendDot className='bg-amber-300' label='Partial' />
            <LegendDot className='bg-slate-700' label='Unavailable' />
          </div>
        </div>
      )}
    </article>
  )
}

function EvidenceSummaryPanel({
  record,
  availableCount,
  partialCount,
  unavailableCount,
  minimumValue,
  maximumValue,
}: {
  record: HistoricalIntelligenceAggregateRecord
  availableCount: number
  partialCount: number
  unavailableCount: number
  minimumValue: number | null
  maximumValue: number | null
}) {
  const result = record.result
  return (
    <article className='rounded-xl border border-slate-800 bg-slate-950/65 p-4'>
      <h3 className='text-sm font-semibold text-white'>Evidence summary</h3>
      <MetricGrid
        rows={[
          ['Stored at', formatTimestamp(record.stored_at)],
          ['Confidence', `${formatPercent(result.confidence.score)} · ${titleCase(result.confidence.level)}`],
          ['Sample count', formatInteger(result.confidence.sample_count)],
          ['Summary total', formatValue(result.summary.total, result.metric.name)],
          ['Average', formatValue(result.summary.average, result.metric.name)],
          ['Median', formatValue(result.summary.median, result.metric.name)],
          ['Minimum', minimumValue === null ? 'Unavailable' : formatValue(minimumValue, result.metric.name)],
          ['Maximum', maximumValue === null ? 'Unavailable' : formatValue(maximumValue, result.metric.name)],
          ['Complete buckets', formatInteger(availableCount)],
          ['Partial buckets', formatInteger(partialCount)],
          ['Unavailable buckets', formatInteger(unavailableCount)],
          ['Latest source update', formatTimestamp(result.provenance.latest_source_updated_at)],
        ]}
      />
      <div className='mt-4 border-t border-slate-800 pt-3'>
        <p className='text-[11px] uppercase tracking-[0.14em] text-slate-600'>Sources</p>
        <p className='mt-1 text-xs leading-5 text-slate-500'>
          {result.provenance.source_names.join(', ') || 'No sources published'}
        </p>
      </div>
    </article>
  )
}

function PeriodComparisonPanel({
  record,
  comparison,
}: {
  record: HistoricalIntelligenceAggregateRecord
  comparison: ReturnType<typeof buildPeriodComparisonView>
}) {
  return (
    <article className='mt-4 rounded-xl border border-slate-800 bg-slate-950/65 p-4'>
      <div className='flex flex-wrap items-start justify-between gap-3'>
        <div>
          <h3 className='text-sm font-semibold text-white'>Previous-period comparison</h3>
          <p className='mt-1 text-xs leading-5 text-slate-500'>
            Server-owned comparison for the same metric, scope and window duration.
          </p>
        </div>
        <DirectionBadge direction={comparison.direction} />
      </div>
      {!comparison.available ? (
        <p className='mt-4 rounded-lg border border-dashed border-slate-700 p-4 text-sm text-slate-500'>
          No mathematically valid previous-period comparison was published.
        </p>
      ) : (
        <div className='mt-4 grid gap-3 sm:grid-cols-2 xl:grid-cols-5'>
          <MetricTile label='Previous' value={formatValue(comparison.previousValue ?? 0, record.result.metric.name)} />
          <MetricTile label='Current' value={formatValue(comparison.currentValue ?? 0, record.result.metric.name)} />
          <MetricTile label='Absolute change' value={formatSignedValue(comparison.absoluteChange, record.result.metric.name)} />
          <MetricTile label='Percentage change' value={comparison.percentageChange === null ? 'Undefined from zero' : formatSignedPercent(comparison.percentageChange)} />
          <MetricTile label='Previous window' value={comparison.previousWindowLabel ? formatWindowKey(comparison.previousWindowLabel) : 'Unavailable'} />
        </div>
      )}
    </article>
  )
}

function StoredAggregateComparisonPanel({
  records,
  isPending,
  error,
  hasMore,
  leftRecordID,
  rightRecordID,
  comparison,
  onLeftRecordChange,
  onRightRecordChange,
  onRetry,
}: {
  records: HistoricalIntelligenceAggregateRecord[]
  isPending: boolean
  error: Error | null
  hasMore: boolean
  leftRecordID: string | null
  rightRecordID: string | null
  comparison: ReturnType<typeof compareHistoricalRecords>
  onLeftRecordChange: (value: string | null) => void
  onRightRecordChange: (value: string | null) => void
  onRetry: () => void
}) {
  return (
    <article className='mt-4 rounded-xl border border-slate-800 bg-slate-950/65 p-4'>
      <div>
        <h3 className='text-sm font-semibold text-white'>Persisted aggregate comparison</h3>
        <p className='mt-1 text-xs leading-5 text-slate-500'>
          Compare the metric-level summary value of any two records returned by the
          bounded history endpoint. The first {historyLimit} records are loaded.
          {hasMore ? ' Older records exist beyond this page.' : ''}
        </p>
      </div>
      {isPending ? (
        <PanelMessage title='Loading aggregate history'>Reading stored versions.</PanelMessage>
      ) : error ? (
        <ErrorPanel title='Aggregate history unavailable' error={error} onRetry={onRetry} />
      ) : records.length === 0 ? (
        <PanelMessage title='No stored history'>No persisted versions match this selection.</PanelMessage>
      ) : (
        <>
          <div className='mt-4 grid gap-3 md:grid-cols-2'>
            <RecordSelect label='Earlier record' value={leftRecordID} records={records} onChange={onLeftRecordChange} />
            <RecordSelect label='Later record' value={rightRecordID} records={records} onChange={onRightRecordChange} />
          </div>
          <div className='mt-4 rounded-lg border border-slate-800 bg-slate-950 p-4'>
            {!comparison.comparable ? (
              <p className='text-sm text-slate-500'>{comparison.reason}</p>
            ) : (
              <div className='grid gap-3 sm:grid-cols-2 xl:grid-cols-5'>
                <MetricTile label='Earlier value' value={formatNumber(comparison.leftValue)} />
                <MetricTile label='Later value' value={formatNumber(comparison.rightValue)} />
                <MetricTile label='Absolute change' value={formatSignedNumber(comparison.absoluteChange)} />
                <MetricTile label='Percentage change' value={comparison.percentageChange === null ? 'Undefined from zero' : formatSignedPercent(comparison.percentageChange)} />
                <div className='rounded-lg border border-slate-800 p-3'>
                  <p className='text-[11px] uppercase tracking-[0.14em] text-slate-600'>Direction</p>
                  <div className='mt-2'><DirectionBadge direction={comparison.direction} /></div>
                </div>
              </div>
            )}
          </div>
        </>
      )}
    </article>
  )
}

function LimitationsRegister({ limitations }: { limitations: Array<{ code: string; message: string; scope: string }> }) {
  if (limitations.length === 0) return null
  return (
    <article className='mt-4 rounded-xl border border-amber-400/20 bg-amber-400/5 p-4'>
      <h3 className='text-sm font-semibold text-amber-100'>Published limitations</h3>
      <ul className='mt-3 grid gap-2 md:grid-cols-2'>
        {limitations.map(item => (
          <li key={`${item.scope}|${item.code}|${item.message}`} className='rounded-lg border border-amber-400/15 bg-slate-950/60 p-3'>
            <p className='font-mono text-[11px] text-amber-300'>{item.code}</p>
            <p className='mt-1 text-xs leading-5 text-slate-400'>{item.message}</p>
            <p className='mt-2 text-[10px] uppercase tracking-[0.14em] text-slate-600'>{item.scope}</p>
          </li>
        ))}
      </ul>
    </article>
  )
}

function RecordSelect({ label, value, records, onChange }: { label: string; value: string | null; records: HistoricalIntelligenceAggregateRecord[]; onChange: (value: string | null) => void }) {
  return (
    <label className='text-xs text-slate-500'>
      {label}
      <select value={value ?? ''} onChange={(event: ChangeEvent<HTMLSelectElement>) => onChange(event.target.value || null)} className='mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-slate-200'>
        <option value=''>Select stored aggregate</option>
        {records.map(record => (
          <option key={record.id} value={record.id}>
            {formatTimestamp(record.stored_at)} · {record.result.status} · {formatNumber(summaryComparisonValue(record))}
          </option>
        ))}
      </select>
    </label>
  )
}

function SelectField({ label, value, options, onChange }: { label: string; value: string; options: Array<readonly [string, string]>; onChange: (value: string) => void }) {
  return (
    <label className='text-xs text-slate-500'>
      {label}
      <select value={value} onChange={(event: ChangeEvent<HTMLSelectElement>) => onChange(event.target.value)} className='mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-slate-200'>
        {options.map(([optionValue, optionLabel]) => <option key={optionValue} value={optionValue}>{optionLabel}</option>)}
      </select>
    </label>
  )
}

function TextField({ label, value, placeholder, onChange }: { label: string; value: string; placeholder: string; onChange: (value: string) => void }) {
  return (
    <label className='text-xs text-slate-500'>
      {label}
      <input value={value} maxLength={4} placeholder={placeholder} onChange={(event: ChangeEvent<HTMLInputElement>) => onChange(event.target.value.toUpperCase())} className='mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 font-mono text-sm uppercase text-slate-200 placeholder:text-slate-700' />
    </label>
  )
}

function ErrorPanel({ title, error, onRetry }: { title: string; error: Error; onRetry: () => void }) {
  return (
    <div className='mt-4 rounded-lg border border-rose-400/25 bg-rose-400/5 p-4'>
      <p className='text-sm font-semibold text-rose-100'>{title}</p>
      <p className='mt-1 text-xs leading-5 text-rose-200/75'>{getRequestErrorMessage(error)}</p>
      <button type='button' onClick={onRetry} className='mt-3 rounded-md border border-rose-400/35 px-3 py-1.5 text-xs font-medium text-rose-100'>Retry</button>
    </div>
  )
}

function PanelMessage({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className='mt-5 rounded-xl border border-dashed border-slate-700 bg-slate-950/45 p-5'>
      <p className='text-sm font-semibold text-slate-300'>{title}</p>
      <p className='mt-1 text-xs leading-5 text-slate-500'>{children}</p>
    </div>
  )
}

function StatusBadge({ value }: { value: string }) {
  const className = value === 'complete' ? 'border-emerald-400/30 bg-emerald-400/10 text-emerald-200' : value === 'partial' ? 'border-amber-400/30 bg-amber-400/10 text-amber-200' : 'border-slate-600 bg-slate-800 text-slate-300'
  return <span className={`rounded-full border px-2.5 py-1 text-[11px] font-medium ${className}`}>{titleCase(value)}</span>
}

function DirectionBadge({ direction }: { direction: string }) {
  const label = direction === 'up' ? 'Increasing' : direction === 'down' ? 'Decreasing' : direction === 'flat' ? 'Stable' : 'Unavailable'
  const className = direction === 'up' ? 'border-emerald-400/30 bg-emerald-400/10 text-emerald-200' : direction === 'down' ? 'border-rose-400/30 bg-rose-400/10 text-rose-200' : direction === 'flat' ? 'border-sky-400/30 bg-sky-400/10 text-sky-200' : 'border-slate-700 bg-slate-900 text-slate-400'
  return <span className={`inline-flex rounded-full border px-2.5 py-1 text-[11px] font-medium ${className}`}>{label}</span>
}

function MetricGrid({ rows }: { rows: Array<[string, string]> }) {
  return <dl className='mt-4 grid gap-3 sm:grid-cols-2'>{rows.map(([label, value]) => <div key={label} className='rounded-lg border border-slate-800 p-3'><dt className='text-[11px] uppercase tracking-[0.14em] text-slate-600'>{label}</dt><dd className='mt-1 text-sm text-slate-300'>{value}</dd></div>)}</dl>
}
function MetricTile({ label, value }: { label: string; value: string }) {
  return <div className='rounded-lg border border-slate-800 bg-slate-950 p-3'><p className='text-[11px] uppercase tracking-[0.14em] text-slate-600'>{label}</p><p className='mt-2 text-sm font-semibold text-slate-200'>{value}</p></div>
}
function LegendDot({ className, label }: { className: string; label: string }) {
  return <span className='inline-flex items-center gap-2'><span className={`h-2.5 w-2.5 rounded-full ${className}`} />{label}</span>
}

function barClassName(status: string): string {
  return status === 'complete' ? 'bg-cyan-300/80' : status === 'partial' ? 'bg-amber-300/80' : 'bg-slate-700'
}
function scopeLabel(record: HistoricalIntelligenceResultLike): string {
  const scope = record.scope
  if (scope.type === 'global') return 'Global scope'
  if (scope.type === 'airport') return `Airport ${scope.airport_icao_code ?? 'unknown'}`
  return `${scope.origin_icao_code ?? 'unknown'} → ${scope.destination_icao_code ?? 'unknown'}`
}
type HistoricalIntelligenceResultLike = HistoricalIntelligenceAggregateRecord['result']
function summaryComparisonValue(record: HistoricalIntelligenceAggregateRecord): number {
  const result = record.result
  switch (result.metric.aggregation) {
    case 'count': case 'sum': return result.summary.total
    case 'minimum': return result.summary.minimum
    case 'maximum': return result.summary.maximum
    case 'median': return result.summary.median
    case 'average': case 'ratio': return result.summary.average
  }
}
function formatValue(value: number, metric: HistoricalMetricName): string {
  const definition = metricDefinition(metric)
  if (definition.valueKind === 'ratio') return formatPercent(value)
  if (definition.valueKind === 'count') return formatInteger(value)
  return `${formatNumber(value)} ${definition.unit}`
}
function formatSignedValue(value: number | null, metric: HistoricalMetricName): string {
  if (value === null) return 'Unavailable'
  const formatted = formatValue(Math.abs(value), metric)
  return `${value > 0 ? '+' : value < 0 ? '−' : ''}${formatted}`
}
function formatSignedNumber(value: number | null): string {
  if (value === null) return 'Unavailable'
  return `${value > 0 ? '+' : value < 0 ? '−' : ''}${formatNumber(Math.abs(value))}`
}
function formatSignedPercent(value: number): string {
  return `${value > 0 ? '+' : value < 0 ? '−' : ''}${formatNumber(Math.abs(value))}%`
}
function formatInteger(value: number): string { return new Intl.NumberFormat(undefined, { maximumFractionDigits: 0 }).format(value) }
function formatNumber(value: number | null): string { return value === null ? 'Unavailable' : new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 }).format(value) }
function formatPercent(value: number): string { return new Intl.NumberFormat(undefined, { style: 'percent', maximumFractionDigits: 1 }).format(value) }
function formatTimestamp(value: string): string { return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value)) }
function formatBucketLabel(value: string, granularity: HistoricalGranularity): string {
  const date = new Date(value)
  return granularity === 'hour'
    ? new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric', hour: '2-digit' }).format(date)
    : new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric' }).format(date)
}
function formatWindowKey(value: string): string {
  const [start, end] = value.split('|')
  return `${formatTimestamp(start)} → ${formatTimestamp(end)}`
}
function titleCase(value: string): string { return value.replace(/_/g, ' ').replace(/\b\w/g, letter => letter.toUpperCase()) }
