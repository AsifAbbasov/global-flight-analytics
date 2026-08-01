// FRONTEND_TRAFFIC_WORKSPACE_V1
'use client'

import { useState, type ChangeEvent } from 'react'

import { AircraftDetailPanel } from '@/components/aircraft/aircraft-detail-panel'
import { AircraftExplorer } from '@/components/aircraft/aircraft-explorer'
import { ProjectionIntelligencePanel } from '@/components/aircraft/projection-intelligence-panel'
import { RouteIntelligencePanel } from '@/components/aircraft/route-intelligence-panel'
import { StabilityIntelligencePanel } from '@/components/aircraft/stability-intelligence-panel'
import { WeatherContextPanel } from '@/components/aircraft/weather-context-panel'
import { TrafficGlobe } from '@/components/globe/traffic-globe'
import { TrafficMap } from '@/components/map/traffic-map'
import { RegionalTrafficBrief } from '@/components/traffic/regional-traffic-brief'
import { getRequestErrorMessage } from '@/lib/api/client'
import { useProjectionIntelligence } from '@/lib/queries/projection-intelligence'
import { useAircraftRouteContext } from '@/lib/queries/route-context'
import { useProcessedRouteIntelligence } from '@/lib/queries/route-intelligence'
import {
  buildStabilityAsOfTimes,
  useStabilityIntelligence,
} from '@/lib/queries/stability-intelligence'
import { useCurrentTraffic } from '@/lib/queries/traffic'
import { useLatestAircraftTrajectory } from '@/lib/queries/trajectory'
import { useWeatherContext } from '@/lib/queries/weather-context'
import {
  buildTrafficWorkspaceSelection,
  type TrafficWorkspacePanel,
} from '@/lib/traffic/traffic-workspace-model'
import type { Region } from '@/types/region'
import type { TrafficAircraft } from '@/types/traffic'

interface TrafficDashboardProps {
  regions: Region[]
  selectedRegion: Region
  onSelectedRegionCodeChange: (regionCode: string) => void
  initialTraffic: TrafficAircraft[]
  initialError: string | null
  regionsWarning: string | null
}

export function TrafficDashboard({
  regions,
  selectedRegion,
  onSelectedRegionCodeChange,
  initialTraffic,
  initialError,
  regionsWarning,
}: TrafficDashboardProps) {
  const [selectedAircraftICAO24, setSelectedAircraftICAO24] = useState<
    string | null
  >(null)
  const [workspacePanel, setWorkspacePanel] =
    useState<TrafficWorkspacePanel>('aircraft')

  const initialData =
    selectedRegion.code === 'world' && initialError === null
      ? initialTraffic
      : undefined
  const trafficQuery = useCurrentTraffic(selectedRegion.code, { initialData })
  const routeContextQuery = useAircraftRouteContext(selectedAircraftICAO24)
  const trajectoryQuery = useLatestAircraftTrajectory(selectedAircraftICAO24)
  const routeIntelligenceTrajectoryID =
    routeContextQuery.data?.trajectory_id ?? null
  const routeIntelligenceQuery = useProcessedRouteIntelligence(
    routeIntelligenceTrajectoryID
  )
  const projectionTrajectoryID =
    trajectoryQuery.data?.id ?? routeIntelligenceTrajectoryID
  const projectionAsOfTime = trajectoryQuery.data?.end_time ?? null
  const projectionQuery = useProjectionIntelligence(
    projectionTrajectoryID,
    projectionAsOfTime
  )
  const weatherContextQuery = useWeatherContext(
    projectionTrajectoryID,
    projectionAsOfTime
  )
  const stabilityAsOfTimes = buildStabilityAsOfTimes(
    trajectoryQuery.data?.start_time ?? null,
    trajectoryQuery.data?.end_time ?? null
  )
  const stabilityQuery = useStabilityIntelligence(
    projectionTrajectoryID,
    stabilityAsOfTimes
  )

  const traffic = trafficQuery.data ?? []
  const selectedAircraft =
    selectedAircraftICAO24 === null
      ? undefined
      : traffic.find(
          (item: TrafficAircraft) =>
            item.icao24.trim().toLowerCase() === selectedAircraftICAO24
        )
  const isInitialLoading = trafficQuery.isPending
  const isRefreshing = trafficQuery.isFetching && !trafficQuery.isPending
  const errorMessage = trafficQuery.error
    ? getRequestErrorMessage(trafficQuery.error)
    : trafficQuery.isPending
      ? initialError
      : null

  const selectAircraft = (icao24: string | null) => {
    const selection = buildTrafficWorkspaceSelection(icao24)
    setSelectedAircraftICAO24(selection.icao24)
    setWorkspacePanel(selection.panel)
  }

  const changeRegion = (regionCode: string) => {
    selectAircraft(null)
    onSelectedRegionCodeChange(regionCode)
  }

  return (
    <>
      <section className='mt-6 rounded-xl border border-slate-800 bg-slate-900 p-4'>
        <div className='flex flex-wrap items-end justify-between gap-4'>
          <div className='min-w-64 flex-1'>
            <label
              className='block text-sm font-medium text-slate-300'
              htmlFor='traffic-region'
            >
              Region
            </label>
            <select
              id='traffic-region'
              value={selectedRegion.code}
              onChange={(event: ChangeEvent<HTMLSelectElement>) => {
                changeRegion(event.target.value)
              }}
              className='mt-2 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-white'
            >
              {regions.map(region => (
                <option key={region.code} value={region.code}>
                  {region.name}
                </option>
              ))}
            </select>
          </div>
          <button
            type='button'
            onClick={() => {
              void trafficQuery.refetch()
            }}
            disabled={trafficQuery.isFetching}
            className='rounded-lg border border-slate-700 px-4 py-2 text-sm font-medium text-slate-200 disabled:opacity-60'
          >
            {trafficQuery.isFetching ? 'Refreshing traffic…' : 'Refresh traffic'}
          </button>
        </div>

        <div
          aria-live='polite'
          className='mt-3 flex flex-wrap items-center gap-3 text-sm'
        >
          <span className='text-slate-300'>Aircraft: {traffic.length}</span>
          <span className='text-slate-500'>View: {selectedRegion.name}</span>
          {selectedAircraftICAO24 ? (
            <span className='text-sky-300'>
              Selected: {selectedAircraftICAO24.toUpperCase()}
            </span>
          ) : null}
          {trafficQuery.dataUpdatedAt > 0 ? (
            <span className='text-slate-500'>
              Updated {formatTimestamp(trafficQuery.dataUpdatedAt)}
            </span>
          ) : null}
          {regionsWarning ? (
            <span className='text-amber-300'>{regionsWarning}</span>
          ) : null}
          {isInitialLoading ? (
            <span className='text-sky-300'>Loading current traffic…</span>
          ) : null}
          {isRefreshing ? (
            <span className='text-sky-300'>Updating current traffic…</span>
          ) : null}
          {errorMessage ? (
            <>
              <span className='text-amber-300'>{errorMessage}</span>
              <button
                type='button'
                onClick={() => {
                  void trafficQuery.refetch()
                }}
                disabled={trafficQuery.isFetching}
                className='rounded-md border border-amber-400/50 px-3 py-1 font-medium text-amber-200 disabled:opacity-60'
              >
                Retry
              </button>
            </>
          ) : null}
        </div>
      </section>

      <div className='mt-4' aria-busy={trafficQuery.isFetching}>
        <TrafficGlobe aircraft={traffic} region={selectedRegion} />
      </div>

      <RegionalTrafficBrief
        aircraft={traffic}
        regionName={selectedRegion.name}
        isFetching={trafficQuery.isFetching}
      />

      <section className='mt-8 rounded-xl border border-slate-800 bg-slate-900 p-4 sm:p-6'>
        <h2 className='text-xl font-semibold'>
          Current Traffic — {selectedRegion.name}
        </h2>
        <p className='mt-2 text-sm leading-6 text-slate-400'>
          The map remains visible while the workspace switches between aircraft
          discovery and the selected aircraft intelligence record.
        </p>

        <div className='mt-4 grid gap-4 xl:grid-cols-[minmax(0,1fr)_480px]'>
          <div aria-busy={trafficQuery.isFetching}>
            <TrafficMap
              aircraft={traffic}
              region={selectedRegion}
              selectedAircraftICAO24={selectedAircraftICAO24}
              trajectory={trajectoryQuery.data}
              projection={projectionQuery.data?.projection}
              onSelectAircraft={selectAircraft}
            />
          </div>

          <aside
            className='min-w-0 rounded-xl border border-slate-700 bg-slate-950/60 p-3'
            aria-label='Traffic workspace'
          >
            <WorkspaceTabs
              activePanel={workspacePanel}
              selectedAircraftICAO24={selectedAircraftICAO24}
              onPanelChange={setWorkspacePanel}
            />

            <div
              id='traffic-workspace-panel'
              role='tabpanel'
              aria-labelledby={`traffic-workspace-${workspacePanel}-tab`}
              className='mt-3'
            >
              {workspacePanel === 'aircraft' ? (
                <AircraftExplorer
                  aircraft={traffic}
                  selectedAircraftICAO24={selectedAircraftICAO24}
                  onSelectAircraft={selectAircraft}
                  isFetching={trafficQuery.isFetching}
                />
              ) : selectedAircraftICAO24 === null ? (
                <IntelligenceSelectionPrompt
                  onOpenAircraft={() => setWorkspacePanel('aircraft')}
                />
              ) : (
                <div className='space-y-4'>
                  <SelectedAircraftContext
                    icao24={selectedAircraftICAO24}
                    onOpenAircraft={() => setWorkspacePanel('aircraft')}
                    onClear={() => selectAircraft(null)}
                  />
                  <AircraftDetailPanel
                    selectedICAO24={selectedAircraftICAO24}
                    aircraft={selectedAircraft}
                    routeContext={routeContextQuery.data}
                    routeContextIsPending={routeContextQuery.isPending}
                    routeContextIsFetching={routeContextQuery.isFetching}
                    routeContextError={routeContextQuery.error}
                    onRetryRouteContext={() => {
                      void routeContextQuery.refetch()
                    }}
                    trajectory={trajectoryQuery.data}
                    trajectoryIsPending={trajectoryQuery.isPending}
                    trajectoryIsFetching={trajectoryQuery.isFetching}
                    trajectoryError={trajectoryQuery.error}
                    onRetryTrajectory={() => {
                      void trajectoryQuery.refetch()
                    }}
                    onClose={() => selectAircraft(null)}
                  />
                  <RouteIntelligencePanel
                    selectedICAO24={selectedAircraftICAO24}
                    trajectoryID={routeIntelligenceTrajectoryID}
                    record={routeIntelligenceQuery.data}
                    isPending={routeIntelligenceQuery.isPending}
                    isFetching={routeIntelligenceQuery.isFetching}
                    error={routeIntelligenceQuery.error}
                    onRetry={() => {
                      void routeIntelligenceQuery.refetch()
                    }}
                  />
                  <ProjectionIntelligencePanel
                    selectedICAO24={selectedAircraftICAO24}
                    trajectoryID={projectionTrajectoryID}
                    result={projectionQuery.data}
                    isPending={projectionQuery.isPending}
                    isFetching={projectionQuery.isFetching}
                    error={projectionQuery.error}
                    onRetry={() => {
                      void projectionQuery.refetch()
                    }}
                  />
                  <WeatherContextPanel
                    selectedICAO24={selectedAircraftICAO24}
                    trajectoryID={projectionTrajectoryID}
                    result={weatherContextQuery.data}
                    isPending={weatherContextQuery.isPending}
                    isFetching={weatherContextQuery.isFetching}
                    error={weatherContextQuery.error}
                    onRetry={() => {
                      void weatherContextQuery.refetch()
                    }}
                  />
                  <StabilityIntelligencePanel
                    selectedICAO24={selectedAircraftICAO24}
                    trajectoryID={projectionTrajectoryID}
                    asOfTimes={stabilityAsOfTimes}
                    result={stabilityQuery.data}
                    isPending={stabilityQuery.isPending}
                    isFetching={stabilityQuery.isFetching}
                    error={stabilityQuery.error}
                    onRetry={() => {
                      void stabilityQuery.refetch()
                    }}
                  />
                </div>
              )}
            </div>
          </aside>
        </div>

        {!trafficQuery.isFetching && !errorMessage && traffic.length === 0 ? (
          <p className='mt-4 text-sm text-slate-400'>
            No aircraft were returned for the selected region.
          </p>
        ) : null}
      </section>
    </>
  )
}

function WorkspaceTabs({
  activePanel,
  selectedAircraftICAO24,
  onPanelChange,
}: {
  activePanel: TrafficWorkspacePanel
  selectedAircraftICAO24: string | null
  onPanelChange: (panel: TrafficWorkspacePanel) => void
}) {
  return (
    <div
      role='tablist'
      aria-label='Traffic workspace view'
      className='grid grid-cols-2 gap-2 rounded-lg border border-slate-800 bg-slate-950 p-1'
    >
      <WorkspaceTab
        id='traffic-workspace-aircraft-tab'
        active={activePanel === 'aircraft'}
        label='Aircraft'
        detail='Search and select'
        onClick={() => onPanelChange('aircraft')}
      />
      <WorkspaceTab
        id='traffic-workspace-intelligence-tab'
        active={activePanel === 'intelligence'}
        label='Intelligence'
        detail={
          selectedAircraftICAO24 === null
            ? 'No selection'
            : selectedAircraftICAO24.toUpperCase()
        }
        onClick={() => onPanelChange('intelligence')}
      />
    </div>
  )
}

function WorkspaceTab({
  id,
  active,
  label,
  detail,
  onClick,
}: {
  id: string
  active: boolean
  label: string
  detail: string
  onClick: () => void
}) {
  return (
    <button
      id={id}
      type='button'
      role='tab'
      aria-selected={active}
      aria-controls='traffic-workspace-panel'
      tabIndex={active ? 0 : -1}
      onClick={onClick}
      className={`rounded-md px-3 py-2 text-left transition ${
        active
          ? 'bg-sky-400/10 text-sky-100 ring-1 ring-sky-400/40'
          : 'text-slate-400 hover:bg-slate-900 hover:text-slate-200'
      }`}
    >
      <span className='block text-sm font-semibold'>{label}</span>
      <span className='mt-0.5 block truncate text-[11px] opacity-75'>
        {detail}
      </span>
    </button>
  )
}

function IntelligenceSelectionPrompt({
  onOpenAircraft,
}: {
  onOpenAircraft: () => void
}) {
  return (
    <div className='rounded-xl border border-dashed border-slate-700 bg-slate-950/60 p-5'>
      <p className='text-sm font-semibold text-slate-200'>
        Select an aircraft to open Intelligence
      </p>
      <p className='mt-2 text-sm leading-6 text-slate-400'>
        Use the Aircraft tab or select a marker on the map. Route, trajectory,
        projection, weather and stability evidence will open here.
      </p>
      <button
        type='button'
        onClick={onOpenAircraft}
        className='mt-4 rounded-lg border border-sky-400/40 px-3 py-2 text-sm font-medium text-sky-100 transition hover:bg-sky-400/10'
      >
        Open Aircraft Explorer
      </button>
    </div>
  )
}

function SelectedAircraftContext({
  icao24,
  onOpenAircraft,
  onClear,
}: {
  icao24: string
  onOpenAircraft: () => void
  onClear: () => void
}) {
  return (
    <div className='flex flex-wrap items-center justify-between gap-3 rounded-xl border border-sky-400/30 bg-sky-400/5 p-3'>
      <div>
        <p className='text-[11px] font-semibold uppercase tracking-[0.16em] text-sky-300'>
          Selected aircraft
        </p>
        <p className='mt-1 font-mono text-sm uppercase text-white'>{icao24}</p>
      </div>
      <div className='flex flex-wrap gap-2'>
        <button
          type='button'
          onClick={onOpenAircraft}
          className='rounded-md border border-slate-700 px-3 py-1.5 text-xs font-medium text-slate-200 hover:bg-slate-900'
        >
          Aircraft list
        </button>
        <button
          type='button'
          onClick={onClear}
          className='rounded-md border border-rose-400/30 px-3 py-1.5 text-xs font-medium text-rose-100 hover:bg-rose-400/10'
        >
          Clear selection
        </button>
      </div>
    </div>
  )
}

function formatTimestamp(value: number): string {
  return new Intl.DateTimeFormat(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  }).format(new Date(value))
}
