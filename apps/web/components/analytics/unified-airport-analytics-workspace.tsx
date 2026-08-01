// FRONTEND_UNIFIED_AIRPORT_ANALYTICS_WORKSPACE_V1
'use client'

import {
  useMemo,
  useState,
  type ChangeEvent,
  type ReactNode,
} from 'react'

import { getRequestErrorMessage } from '@/lib/api/client'
import {
  buildAirportHistorySeries,
  buildAirportRankingView,
  buildAirportTrendSummary,
  mergeAirportLimitations,
  type AirportRankingSort,
} from '@/lib/analytics/airport-intelligence-workspace-model'
import {
  useAirportIntelligenceHistory,
  useAirportIntelligenceOverview,
  useAirportIntelligenceRanking,
  useAirportIntelligenceTrends,
} from '@/lib/queries/airport-intelligence'
import type {
  AirportIntelligenceLimitation,
  AirportIntelligenceOverview,
  AirportIntelligenceTrends,
  AirportRankedItem,
} from '@/types/airport-intelligence'

type AirportWindowDays = 7 | 30 | 90
type AirportProfilePanel = 'overview' | 'history' | 'trends'

const windowChoices: AirportWindowDays[] = [7, 30, 90]
const rankingLimit = 100

export function UnifiedAirportAnalyticsWorkspace() {
  const [days, setDays] = useState<AirportWindowDays>(30)
  const [search, setSearch] = useState('')
  const [sort, setSort] = useState<AirportRankingSort>('position')
  const [selectedICAOCode, setSelectedICAOCode] = useState<string | null>(null)
  const [profilePanel, setProfilePanel] =
    useState<AirportProfilePanel>('overview')

  const rankingQuery = useAirportIntelligenceRanking(days, rankingLimit)
  const overviewQuery = useAirportIntelligenceOverview(selectedICAOCode, days)
  const historyQuery = useAirportIntelligenceHistory(selectedICAOCode, days)
  const trendsQuery = useAirportIntelligenceTrends(selectedICAOCode, days)

  const rankingView = useMemo(
    () =>
      buildAirportRankingView(rankingQuery.data?.airports ?? [], {
        search,
        sort,
        limit: rankingLimit,
      }),
    [rankingQuery.data?.airports, search, sort]
  )

  const limitations = useMemo(
    () =>
      mergeAirportLimitations(
        rankingQuery.data?.limitations,
        overviewQuery.data?.limitations,
        historyQuery.data?.limitations,
        trendsQuery.data?.limitations
      ),
    [
      rankingQuery.data?.limitations,
      overviewQuery.data?.limitations,
      historyQuery.data?.limitations,
      trendsQuery.data?.limitations,
    ]
  )

  const selectAirport = (icaoCode: string) => {
    setSelectedICAOCode(icaoCode)
    setProfilePanel('overview')
  }

  const refreshWorkspace = () => {
    void rankingQuery.refetch()
    if (selectedICAOCode !== null) {
      void overviewQuery.refetch()
      void historyQuery.refetch()
      void trendsQuery.refetch()
    }
  }

  const isFetching =
    rankingQuery.isFetching ||
    overviewQuery.isFetching ||
    historyQuery.isFetching ||
    trendsQuery.isFetching

  return (
    <section
      aria-labelledby='airport-intelligence-workspace-title'
      aria-busy={isFetching}
      className='mt-8 rounded-2xl border border-white/10 bg-slate-900/65 p-4 shadow-2xl shadow-black/15 sm:p-6'
    >
      <div className='flex flex-wrap items-start justify-between gap-5'>
        <div>
          <p className='text-xs font-semibold uppercase tracking-[0.24em] text-violet-300'>
            Unified analytics workspace
          </p>
          <h2
            id='airport-intelligence-workspace-title'
            className='mt-3 text-2xl font-semibold tracking-tight text-white sm:text-3xl'
          >
            Airport Intelligence
          </h2>
          <p className='mt-3 max-w-4xl text-sm leading-6 text-slate-400'>
            Explore the production airport ranking, open a digital passport and
            inspect completed-day history and trend evidence without leaving the
            main research workspace. Ranking is global because the current API
            does not publish a region filter for Airport Intelligence.
          </p>
        </div>

        <div className='flex flex-wrap items-end gap-2'>
          <label className='text-xs text-slate-500'>
            Completed-day window
            <select
              value={days}
              onChange={(event: ChangeEvent<HTMLSelectElement>) => {
                setDays(Number(event.target.value) as AirportWindowDays)
              }}
              className='mt-1 block rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-slate-200'
            >
              {windowChoices.map(value => (
                <option key={value} value={value}>
                  {value} days
                </option>
              ))}
            </select>
          </label>
          <button
            type='button'
            onClick={refreshWorkspace}
            disabled={isFetching}
            className='rounded-lg border border-violet-400/35 bg-violet-400/5 px-4 py-2 text-sm font-medium text-violet-100 transition hover:bg-violet-400/10 disabled:opacity-60'
          >
            {isFetching ? 'Refreshing analytics…' : 'Refresh analytics'}
          </button>
        </div>
      </div>

      <div className='mt-6 grid gap-5 xl:grid-cols-[minmax(360px,0.82fr)_minmax(0,1.55fr)]'>
        <AirportRankingPanel
          days={days}
          search={search}
          sort={sort}
          selectedICAOCode={selectedICAOCode}
          ranking={rankingQuery.data}
          rankingView={rankingView}
          isPending={rankingQuery.isPending}
          isFetching={rankingQuery.isFetching}
          error={rankingQuery.error}
          onSearchChange={setSearch}
          onSortChange={setSort}
          onSelectAirport={selectAirport}
          onRetry={() => {
            void rankingQuery.refetch()
          }}
        />

        <AirportProfilePanelView
          days={days}
          selectedICAOCode={selectedICAOCode}
          activePanel={profilePanel}
          overview={overviewQuery.data}
          history={historyQuery.data}
          trends={trendsQuery.data}
          overviewPending={overviewQuery.isPending}
          historyPending={historyQuery.isPending}
          trendsPending={trendsQuery.isPending}
          overviewError={overviewQuery.error}
          historyError={historyQuery.error}
          trendsError={trendsQuery.error}
          onPanelChange={setProfilePanel}
          onClear={() => setSelectedICAOCode(null)}
          onRetryOverview={() => {
            void overviewQuery.refetch()
          }}
          onRetryHistory={() => {
            void historyQuery.refetch()
          }}
          onRetryTrends={() => {
            void trendsQuery.refetch()
          }}
        />
      </div>

      <LimitationsRegister limitations={limitations} />

      <p className='mt-5 border-t border-slate-800 pt-4 text-xs leading-5 text-slate-600'>
        Airport Intelligence uses completed-day analytical windows. It is a
        research summary, not an operational airport status, capacity declaration,
        timetable or safety assessment. Empty and insufficient-history responses
        remain visible instead of being converted into synthetic values.
      </p>
    </section>
  )
}

function AirportRankingPanel({
  days,
  search,
  sort,
  selectedICAOCode,
  ranking,
  rankingView,
  isPending,
  isFetching,
  error,
  onSearchChange,
  onSortChange,
  onSelectAirport,
  onRetry,
}: {
  days: AirportWindowDays
  search: string
  sort: AirportRankingSort
  selectedICAOCode: string | null
  ranking: ReturnType<typeof useAirportIntelligenceRanking>['data']
  rankingView: ReturnType<typeof buildAirportRankingView>
  isPending: boolean
  isFetching: boolean
  error: Error | null
  onSearchChange: (value: string) => void
  onSortChange: (value: AirportRankingSort) => void
  onSelectAirport: (icaoCode: string) => void
  onRetry: () => void
}) {
  return (
    <article className='min-w-0 rounded-xl border border-slate-800 bg-slate-950/65 p-4'>
      <div className='flex items-start justify-between gap-3'>
        <div>
          <p className='text-[11px] font-semibold uppercase tracking-[0.16em] text-slate-500'>
            Global ranking
          </p>
          <h3 className='mt-1 text-lg font-semibold text-white'>
            Airport discovery
          </h3>
          <p className='mt-1 text-xs leading-5 text-slate-500'>
            Completed {days}-day window · up to {rankingLimit} airports.
          </p>
        </div>
        {isFetching && !isPending ? (
          <span className='rounded-full border border-violet-400/30 bg-violet-400/10 px-2.5 py-1 text-[11px] text-violet-200'>
            Updating
          </span>
        ) : null}
      </div>

      <div className='mt-4 grid gap-2 sm:grid-cols-[minmax(0,1fr)_150px]'>
        <label className='text-xs text-slate-500'>
          Search airport
          <input
            type='search'
            value={search}
            onChange={(event: ChangeEvent<HTMLInputElement>) =>
              onSearchChange(event.target.value)
            }
            placeholder='ICAO, IATA, name, city, country'
            className='mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-slate-200 placeholder:text-slate-700'
          />
        </label>
        <label className='text-xs text-slate-500'>
          Sort by
          <select
            value={sort}
            onChange={(event: ChangeEvent<HTMLSelectElement>) =>
              onSortChange(event.target.value as AirportRankingSort)
            }
            className='mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-slate-200'
          >
            <option value='position'>Published rank</option>
            <option value='activity'>Activity score</option>
            <option value='confidence'>Data confidence</option>
            <option value='movements'>Movements</option>
            <option value='routes'>Active routes</option>
          </select>
        </label>
      </div>

      {isPending ? (
        <PanelMessage title='Loading airport ranking'>
          Reading the completed-day ranking from the production API.
        </PanelMessage>
      ) : error ? (
        <ErrorPanel error={error} onRetry={onRetry} />
      ) : rankingView.items.length === 0 ? (
        <PanelMessage title='No matching airports'>
          {ranking?.airports.length === 0
            ? 'The API returned no eligible airports for this completed-day window.'
            : 'No airport matches the current search text.'}
        </PanelMessage>
      ) : (
        <>
          <p className='mt-4 text-xs text-slate-600'>
            Showing {formatInteger(rankingView.visibleCount)} of{' '}
            {formatInteger(rankingView.matchedCount)} matching airports ·{' '}
            {formatInteger(rankingView.totalCount)} published.
          </p>
          <ol className='mt-3 max-h-[690px] space-y-2 overflow-y-auto pr-1'>
            {rankingView.items.map(item => (
              <AirportRankingRow
                key={item.icao_code}
                item={item}
                active={selectedICAOCode === item.icao_code}
                onSelect={() => onSelectAirport(item.icao_code)}
              />
            ))}
          </ol>
        </>
      )}
    </article>
  )
}

function AirportRankingRow({
  item,
  active,
  onSelect,
}: {
  item: AirportRankedItem
  active: boolean
  onSelect: () => void
}) {
  return (
    <li>
      <button
        type='button'
        onClick={onSelect}
        aria-pressed={active}
        className={`w-full rounded-lg border p-3 text-left transition ${
          active
            ? 'border-violet-400/45 bg-violet-400/10'
            : 'border-slate-800 bg-slate-950/70 hover:border-slate-700 hover:bg-slate-900'
        }`}
      >
        <div className='flex items-start gap-3'>
          <span className='flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border border-slate-700 bg-slate-900 font-mono text-xs text-violet-200'>
            {item.position}
          </span>
          <div className='min-w-0 flex-1'>
            <div className='flex flex-wrap items-center gap-x-2 gap-y-1'>
              <span className='font-mono text-sm font-semibold text-white'>
                {item.icao_code}
              </span>
              {item.iata_code.trim() !== '' ? (
                <span className='text-xs text-sky-300'>{item.iata_code}</span>
              ) : null}
              <span className='text-xs text-slate-600'>
                {formatPercent(item.data_confidence)} confidence
              </span>
            </div>
            <p className='mt-1 truncate text-sm text-slate-300'>{item.name}</p>
            <p className='mt-1 truncate text-xs text-slate-600'>
              {[item.city, item.country].filter(Boolean).join(' · ') || 'Location unavailable'}
            </p>
            <div className='mt-2 grid grid-cols-3 gap-2 text-[11px] text-slate-500'>
              <MetricFragment label='Movements' value={formatInteger(item.total_movements)} />
              <MetricFragment label='Routes' value={formatInteger(item.active_routes)} />
              <MetricFragment label='Rate' value={`${formatDecimal(item.movements_per_hour)}/h`} />
            </div>
          </div>
        </div>
      </button>
    </li>
  )
}

function AirportProfilePanelView({
  days,
  selectedICAOCode,
  activePanel,
  overview,
  history,
  trends,
  overviewPending,
  historyPending,
  trendsPending,
  overviewError,
  historyError,
  trendsError,
  onPanelChange,
  onClear,
  onRetryOverview,
  onRetryHistory,
  onRetryTrends,
}: {
  days: AirportWindowDays
  selectedICAOCode: string | null
  activePanel: AirportProfilePanel
  overview: ReturnType<typeof useAirportIntelligenceOverview>['data']
  history: ReturnType<typeof useAirportIntelligenceHistory>['data']
  trends: ReturnType<typeof useAirportIntelligenceTrends>['data']
  overviewPending: boolean
  historyPending: boolean
  trendsPending: boolean
  overviewError: Error | null
  historyError: Error | null
  trendsError: Error | null
  onPanelChange: (panel: AirportProfilePanel) => void
  onClear: () => void
  onRetryOverview: () => void
  onRetryHistory: () => void
  onRetryTrends: () => void
}) {
  if (selectedICAOCode === null) {
    return (
      <article className='flex min-h-[520px] items-center justify-center rounded-xl border border-dashed border-slate-700 bg-slate-950/45 p-8 text-center'>
        <div className='max-w-md'>
          <p className='text-xs font-semibold uppercase tracking-[0.18em] text-violet-300'>
            Airport profile
          </p>
          <h3 className='mt-3 text-2xl font-semibold text-white'>
            Select an airport from the ranking
          </h3>
          <p className='mt-3 text-sm leading-6 text-slate-400'>
            The workspace will load its passport, completed-day operating statistics,
            historical windows, trend direction, continuity and explicit limitations.
          </p>
        </div>
      </article>
    )
  }

  return (
    <article className='min-w-0 rounded-xl border border-slate-800 bg-slate-950/65 p-4 sm:p-5'>
      <div className='flex flex-wrap items-start justify-between gap-4'>
        <div>
          <p className='text-[11px] font-semibold uppercase tracking-[0.16em] text-violet-300'>
            Selected airport
          </p>
          <h3 className='mt-1 font-mono text-xl font-semibold text-white'>
            {selectedICAOCode}
          </h3>
          <p className='mt-1 text-xs text-slate-500'>
            Completed {days}-day analytical window
          </p>
        </div>
        <button
          type='button'
          onClick={onClear}
          className='rounded-lg border border-slate-700 px-3 py-2 text-xs font-medium text-slate-300 transition hover:bg-slate-900'
        >
          Clear airport
        </button>
      </div>

      <div
        role='tablist'
        aria-label='Airport Intelligence profile sections'
        className='mt-5 grid grid-cols-3 gap-2 rounded-lg border border-slate-800 bg-slate-950 p-1'
      >
        <ProfileTab
          panel='overview'
          label='Overview'
          active={activePanel === 'overview'}
          onSelect={onPanelChange}
        />
        <ProfileTab
          panel='history'
          label='History'
          active={activePanel === 'history'}
          onSelect={onPanelChange}
        />
        <ProfileTab
          panel='trends'
          label='Trends'
          active={activePanel === 'trends'}
          onSelect={onPanelChange}
        />
      </div>

      <div
        id='airport-profile-panel'
        role='tabpanel'
        aria-labelledby={`airport-profile-${activePanel}-tab`}
        className='mt-4'
      >
        {activePanel === 'overview' ? (
          overviewPending ? (
            <PanelMessage title='Loading airport passport'>
              Reading airport identity, operations, ranking and evidence quality.
            </PanelMessage>
          ) : overviewError ? (
            <ErrorPanel error={overviewError} onRetry={onRetryOverview} />
          ) : overview ? (
            <AirportOverviewContent overview={overview} />
          ) : null
        ) : activePanel === 'history' ? (
          historyPending ? (
            <PanelMessage title='Loading historical windows'>
              Reading completed-day Airport Intelligence observations.
            </PanelMessage>
          ) : historyError ? (
            <ErrorPanel error={historyError} onRetry={onRetryHistory} />
          ) : history ? (
            <AirportHistoryContent entries={history.entries} />
          ) : null
        ) : trendsPending ? (
          <PanelMessage title='Loading trend evidence'>
            Comparing eligible historical windows and continuity evidence.
          </PanelMessage>
        ) : trendsError ? (
          <ErrorPanel
            error={trendsError}
            onRetry={onRetryTrends}
            note='Trend calculation may be unavailable when fewer than two observed daily windows exist.'
          />
        ) : trends ? (
          <AirportTrendsContent trends={trends} />
        ) : null}
      </div>
    </article>
  )
}

function ProfileTab({
  panel,
  label,
  active,
  onSelect,
}: {
  panel: AirportProfilePanel
  label: string
  active: boolean
  onSelect: (panel: AirportProfilePanel) => void
}) {
  return (
    <button
      id={`airport-profile-${panel}-tab`}
      type='button'
      role='tab'
      aria-selected={active}
      aria-controls='airport-profile-panel'
      tabIndex={active ? 0 : -1}
      onClick={() => onSelect(panel)}
      className={`rounded-md px-3 py-2 text-sm font-medium transition ${
        active
          ? 'bg-violet-400/10 text-violet-100 ring-1 ring-violet-400/40'
          : 'text-slate-500 hover:bg-slate-900 hover:text-slate-200'
      }`}
    >
      {label}
    </button>
  )
}

function AirportOverviewContent({ overview }: { overview: AirportIntelligenceOverview }) {
  const { passport, statistics, ranking } = overview
  return (
    <div className='space-y-4'>
      <section className='rounded-xl border border-slate-800 bg-slate-950/60 p-4'>
        <div className='flex flex-wrap items-start justify-between gap-4'>
          <div>
            <p className='text-xs uppercase tracking-[0.16em] text-slate-600'>
              Digital passport
            </p>
            <h4 className='mt-2 text-xl font-semibold text-white'>
              {passport.identity.name}
            </h4>
            <p className='mt-1 text-sm text-slate-400'>
              {passport.identity.icao_code}
              {passport.identity.iata_code ? ` · ${passport.identity.iata_code}` : ''}
              {' · '}
              {[passport.location.city, passport.location.country]
                .filter(Boolean)
                .join(', ') || 'Location unavailable'}
            </p>
          </div>
          <span className='rounded-full border border-violet-400/30 bg-violet-400/10 px-3 py-1 text-xs text-violet-100'>
            Rank {ranking.position} / {ranking.total_airports}
          </span>
        </div>
        {passport.description.trim() !== '' ? (
          <p className='mt-4 text-sm leading-6 text-slate-400'>{passport.description}</p>
        ) : null}
        <div className='mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-4'>
          <MetricCard label='Latitude' value={formatCoordinate(passport.location.latitude)} />
          <MetricCard label='Longitude' value={formatCoordinate(passport.location.longitude)} />
          <MetricCard
            label='Elevation'
            value={
              passport.location.elevation_m === null
                ? passport.location.elevation_status
                : `${formatInteger(Math.round(passport.location.elevation_m))} m`
            }
          />
          <MetricCard label='Timezone' value={passport.location.timezone || 'Unavailable'} />
        </div>
      </section>

      <div className='grid gap-4 md:grid-cols-2'>
        <EvidenceCard title='Operating activity' description='Completed-day statistics for the selected analytical window.'>
          <MetricGrid
            items={[
              ['Total movements', formatInteger(statistics.total_movements)],
              ['Arrivals', formatInteger(statistics.arrivals)],
              ['Departures', formatInteger(statistics.departures)],
              ['Movements per hour', formatDecimal(statistics.movements_per_hour)],
              ['Active routes', formatInteger(statistics.active_routes)],
              ['Active aircraft', formatInteger(statistics.active_aircraft)],
            ]}
          />
        </EvidenceCard>
        <EvidenceCard title='Evidence quality' description='Coverage, freshness and published ranking confidence remain separate.'>
          <ProgressRow label='Coverage' value={statistics.coverage_score} />
          <ProgressRow label='Freshness' value={statistics.freshness_score} />
          <ProgressRow label='Ranking confidence' value={ranking.data_confidence} />
          <p className='mt-3 text-xs text-slate-600'>
            {formatInteger(statistics.observed_samples)} observed of{' '}
            {formatInteger(statistics.expected_samples)} expected samples.
          </p>
        </EvidenceCard>
      </div>

      <EvidenceCard title='Ranking composition' description='Published normalized components, not an independently invented browser score.'>
        <div className='grid gap-3 sm:grid-cols-2 lg:grid-cols-4'>
          <ScoreTile label='Activity score' value={ranking.activity_score} />
          <ScoreTile label='Movements component' value={ranking.movements_component} />
          <ScoreTile label='Routes component' value={ranking.routes_component} />
          <ScoreTile label='Observation component' value={ranking.observations_component} />
        </div>
      </EvidenceCard>
    </div>
  )
}

function AirportHistoryContent({ entries }: { entries: AirportIntelligenceOverview['statistics'][] }) {
  const series = useMemo(() => buildAirportHistorySeries(entries, 14), [entries])
  if (series.points.length === 0) {
    return (
      <PanelMessage title='No historical windows'>
        No completed-day entries were available for this airport and window.
      </PanelMessage>
    )
  }

  return (
    <div className='space-y-4'>
      <EvidenceCard
        title='Movement history'
        description={`Latest ${series.visibleEntryCount} of ${series.totalEntryCount} completed-day windows. Bars use the visible peak as their denominator.`}
      >
        <div className='flex h-64 items-end gap-2 overflow-x-auto border-b border-slate-800 pb-3'>
          {series.points.map(point => (
            <div key={point.key} className='flex h-full min-w-10 flex-1 flex-col justify-end'>
              <div className='flex flex-1 items-end justify-center'>
                <div
                  title={`${point.totalMovements} movements`}
                  className='w-full max-w-10 rounded-t bg-gradient-to-t from-violet-500/70 to-sky-300/80'
                  style={{
                    height: `${Math.max(3, point.movementShareOfPeak * 100)}%`,
                  }}
                />
              </div>
              <p className='mt-2 text-center font-mono text-[10px] text-slate-600'>
                {point.label}
              </p>
            </div>
          ))}
        </div>
        <div className='mt-4 grid gap-3 sm:grid-cols-3'>
          <MetricCard label='Visible peak' value={formatInteger(series.peakMovements)} />
          <MetricCard
            label='Peak hourly rate'
            value={`${formatDecimal(series.peakMovementsPerHour)}/h`}
          />
          <MetricCard label='Visible windows' value={formatInteger(series.visibleEntryCount)} />
        </div>
      </EvidenceCard>

      <div className='overflow-x-auto rounded-xl border border-slate-800'>
        <table className='min-w-full divide-y divide-slate-800 text-left text-xs'>
          <thead className='bg-slate-950 text-slate-500'>
            <tr>
              <TableHeading>Window end</TableHeading>
              <TableHeading>Movements</TableHeading>
              <TableHeading>Arrivals</TableHeading>
              <TableHeading>Departures</TableHeading>
              <TableHeading>Routes</TableHeading>
              <TableHeading>Coverage</TableHeading>
            </tr>
          </thead>
          <tbody className='divide-y divide-slate-800 bg-slate-950/40 text-slate-300'>
            {[...series.points].reverse().map(point => (
              <tr key={point.key}>
                <TableCell>{formatDate(point.windowEnd)}</TableCell>
                <TableCell>{formatInteger(point.totalMovements)}</TableCell>
                <TableCell>{formatInteger(point.arrivals)}</TableCell>
                <TableCell>{formatInteger(point.departures)}</TableCell>
                <TableCell>{formatInteger(point.activeRoutes)}</TableCell>
                <TableCell>{formatPercent(point.coverageScore)}</TableCell>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}

function AirportTrendsContent({ trends }: { trends: AirportIntelligenceTrends }) {
  const summary = useMemo(() => buildAirportTrendSummary(trends), [trends])
  return (
    <div className='space-y-4'>
      <section className='rounded-xl border border-slate-800 bg-slate-950/60 p-4'>
        <div className='flex flex-wrap items-start justify-between gap-4'>
          <div>
            <p className='text-xs uppercase tracking-[0.16em] text-slate-600'>
              Published trend direction
            </p>
            <h4 className='mt-2 text-xl font-semibold text-white'>
              {summary.directionLabel}
            </h4>
            <p className='mt-2 text-sm leading-6 text-slate-400'>
              Based on {formatInteger(summary.comparedWindows)} eligible completed-day
              windows. The browser presents the API result without reclassifying its
              underlying trend model.
            </p>
          </div>
          <span className={`rounded-full border px-3 py-1 text-xs ${trendBadgeClassName(summary.direction)}`}>
            {summary.direction}
          </span>
        </div>
      </section>

      <div className='grid gap-3 sm:grid-cols-2 lg:grid-cols-4'>
        <DeltaCard label='Movement change' value={formatSignedInteger(summary.movementDelta)} />
        <DeltaCard
          label='Hourly-rate change'
          value={formatSignedDecimal(summary.movementRateDelta)}
          detail={
            summary.movementRateDeltaPercent === null
              ? 'Percent change unavailable'
              : `${formatSignedDecimal(summary.movementRateDeltaPercent)}%`
          }
        />
        <DeltaCard label='Active-route change' value={formatSignedInteger(summary.activeRoutesDelta)} />
        <DeltaCard label='Continuity' value={formatPercent(summary.continuityScore)} detail={summary.hasGaps ? 'Observed gaps are present' : 'No reported gaps'} />
      </div>

      <div className='grid gap-4 md:grid-cols-3'>
        <TrendPointCard title='Baseline' point={trends.baseline} />
        <TrendPointCard title='Current' point={trends.current} />
        <TrendPointCard title='Peak' point={trends.peak} />
      </div>

      <EvidenceCard title='Evidence movement' description='Changes in data coverage and freshness are shown independently from traffic activity.'>
        <MetricGrid
          items={[
            ['Coverage change', formatSignedPercentPoints(summary.coverageDelta)],
            ['Freshness change', formatSignedPercentPoints(summary.freshnessDelta)],
            ['Reported gaps', formatInteger(trends.gap_count)],
            ['Gap duration', formatDuration(trends.gap_duration_seconds)],
          ]}
        />
      </EvidenceCard>
    </div>
  )
}

function TrendPointCard({
  title,
  point,
}: {
  title: string
  point: AirportIntelligenceTrends['current']
}) {
  return (
    <EvidenceCard title={title} description={`${formatDate(point.window_start)} → ${formatDate(point.window_end)}`}>
      <MetricGrid
        items={[
          ['Movements', formatInteger(point.total_movements)],
          ['Rate', `${formatDecimal(point.movements_per_hour)}/h`],
          ['Routes', formatInteger(point.active_routes)],
          ['Coverage', formatPercent(point.coverage_score)],
        ]}
      />
    </EvidenceCard>
  )
}

function LimitationsRegister({
  limitations,
}: {
  limitations: AirportIntelligenceLimitation[]
}) {
  if (limitations.length === 0) return null
  return (
    <section className='mt-5 rounded-xl border border-amber-400/20 bg-amber-400/5 p-4'>
      <h3 className='text-sm font-semibold text-amber-100'>
        Published limitations
      </h3>
      <p className='mt-1 text-xs leading-5 text-amber-100/60'>
        Combined from ranking, overview, history and trend responses with duplicate
        messages removed deterministically.
      </p>
      <ul className='mt-3 grid gap-2 lg:grid-cols-2'>
        {limitations.map(item => (
          <li
            key={`${item.code}:${item.message}`}
            className='rounded-lg border border-amber-400/15 bg-slate-950/45 p-3'
          >
            <p className='font-mono text-[11px] text-amber-300'>{item.code}</p>
            <p className='mt-1 text-xs leading-5 text-slate-400'>{item.message}</p>
          </li>
        ))}
      </ul>
    </section>
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
      <h4 className='text-sm font-semibold text-white'>{title}</h4>
      <p className='mt-1 text-xs leading-5 text-slate-500'>{description}</p>
      <div className='mt-4'>{children}</div>
    </section>
  )
}

function MetricGrid({ items }: { items: Array<[string, string]> }) {
  return (
    <dl className='grid grid-cols-2 gap-3'>
      {items.map(([label, value]) => (
        <div key={label} className='rounded-lg border border-slate-800 bg-slate-900/45 p-3'>
          <dt className='text-[10px] uppercase tracking-[0.12em] text-slate-600'>
            {label}
          </dt>
          <dd className='mt-1 text-sm font-semibold text-slate-200'>{value}</dd>
        </div>
      ))}
    </dl>
  )
}

function MetricCard({ label, value }: { label: string; value: string }) {
  return (
    <div className='rounded-lg border border-slate-800 bg-slate-900/45 p-3'>
      <p className='text-[10px] uppercase tracking-[0.12em] text-slate-600'>{label}</p>
      <p className='mt-1 truncate text-sm font-semibold text-slate-200' title={value}>
        {value}
      </p>
    </div>
  )
}

function ScoreTile({ label, value }: { label: string; value: number }) {
  return (
    <div className='rounded-lg border border-slate-800 bg-slate-900/45 p-3'>
      <div className='flex items-end justify-between gap-2'>
        <p className='text-xs text-slate-500'>{label}</p>
        <p className='text-sm font-semibold text-slate-200'>{formatPercent(value)}</p>
      </div>
      <div className='mt-2 h-1.5 overflow-hidden rounded-full bg-slate-800'>
        <div className='h-full rounded-full bg-violet-300' style={{ width: `${clampRatio(value) * 100}%` }} />
      </div>
    </div>
  )
}

function ProgressRow({ label, value }: { label: string; value: number }) {
  return (
    <div className='mt-3 first:mt-0'>
      <div className='flex items-center justify-between gap-3 text-xs'>
        <span className='text-slate-500'>{label}</span>
        <span className='text-slate-300'>{formatPercent(value)}</span>
      </div>
      <div className='mt-1.5 h-1.5 overflow-hidden rounded-full bg-slate-800'>
        <div className='h-full rounded-full bg-sky-300' style={{ width: `${clampRatio(value) * 100}%` }} />
      </div>
    </div>
  )
}

function DeltaCard({
  label,
  value,
  detail,
}: {
  label: string
  value: string
  detail?: string
}) {
  return (
    <div className='rounded-xl border border-slate-800 bg-slate-950/60 p-4'>
      <p className='text-[10px] uppercase tracking-[0.14em] text-slate-600'>{label}</p>
      <p className='mt-2 text-xl font-semibold text-white'>{value}</p>
      {detail ? <p className='mt-1 text-xs text-slate-500'>{detail}</p> : null}
    </div>
  )
}

function MetricFragment({ label, value }: { label: string; value: string }) {
  return (
    <span>
      <span className='block text-slate-700'>{label}</span>
      <span className='mt-0.5 block text-slate-400'>{value}</span>
    </span>
  )
}

function TableHeading({ children }: { children: ReactNode }) {
  return <th className='whitespace-nowrap px-3 py-2 font-medium'>{children}</th>
}

function TableCell({ children }: { children: ReactNode }) {
  return <td className='whitespace-nowrap px-3 py-2'>{children}</td>
}

function PanelMessage({
  title,
  children,
}: {
  title: string
  children: ReactNode
}) {
  return (
    <div className='mt-4 rounded-xl border border-dashed border-slate-700 bg-slate-950/40 p-6 text-center'>
      <p className='text-sm font-semibold text-slate-200'>{title}</p>
      <p className='mx-auto mt-2 max-w-md text-xs leading-5 text-slate-500'>{children}</p>
    </div>
  )
}

function ErrorPanel({
  error,
  onRetry,
  note,
}: {
  error: Error
  onRetry: () => void
  note?: string
}) {
  return (
    <div className='mt-4 rounded-xl border border-rose-400/25 bg-rose-400/5 p-5'>
      <p className='text-sm font-semibold text-rose-100'>Analytics request failed</p>
      <p className='mt-2 text-xs leading-5 text-rose-100/70'>
        {getRequestErrorMessage(error)}
      </p>
      {note ? <p className='mt-2 text-xs leading-5 text-slate-500'>{note}</p> : null}
      <button
        type='button'
        onClick={onRetry}
        className='mt-4 rounded-lg border border-rose-400/35 px-3 py-2 text-xs font-medium text-rose-100 transition hover:bg-rose-400/10'
      >
        Retry request
      </button>
    </div>
  )
}

function trendBadgeClassName(
  direction: ReturnType<typeof buildAirportTrendSummary>['direction']
): string {
  switch (direction) {
    case 'increasing':
      return 'border-emerald-400/30 bg-emerald-400/10 text-emerald-100'
    case 'decreasing':
      return 'border-rose-400/30 bg-rose-400/10 text-rose-100'
    case 'stable':
      return 'border-sky-400/30 bg-sky-400/10 text-sky-100'
    case 'unknown':
      return 'border-slate-700 bg-slate-900 text-slate-400'
  }
}

function formatInteger(value: number): string {
  return new Intl.NumberFormat().format(value)
}

function formatDecimal(value: number): string {
  return new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 }).format(value)
}

function formatPercent(value: number): string {
  return new Intl.NumberFormat(undefined, {
    style: 'percent',
    maximumFractionDigits: 0,
  }).format(clampRatio(value))
}

function formatSignedInteger(value: number): string {
  const normalized = Math.trunc(value)
  return `${normalized > 0 ? '+' : ''}${formatInteger(normalized)}`
}

function formatSignedDecimal(value: number): string {
  return `${value > 0 ? '+' : ''}${formatDecimal(value)}`
}

function formatSignedPercentPoints(value: number): string {
  return `${value > 0 ? '+' : ''}${formatDecimal(value * 100)} pp`
}

function formatCoordinate(value: number): string {
  return new Intl.NumberFormat(undefined, { maximumFractionDigits: 5 }).format(value)
}

function formatDate(value: string): string {
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return 'Invalid date'
  return new Intl.DateTimeFormat(undefined, {
    year: 'numeric',
    month: 'short',
    day: '2-digit',
  }).format(parsed)
}

function formatDuration(seconds: number): string {
  if (!Number.isFinite(seconds) || seconds <= 0) return '0 seconds'
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
}

function clampRatio(value: number): number {
  if (!Number.isFinite(value)) return 0
  return Math.min(1, Math.max(0, value))
}
