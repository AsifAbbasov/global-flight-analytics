// FRONTEND_TRAFFIC_DATA_QUALITY_LENS_V1
// FRONTEND_VISUAL_POLISH_V2
'use client'

import { useEffect, useState, type ChangeEvent } from 'react'

import { AircraftDetailPanel } from '@/components/aircraft/aircraft-detail-panel'
import { AircraftExplorer } from '@/components/aircraft/aircraft-explorer'
import { ProjectionIntelligencePanel } from '@/components/aircraft/projection-intelligence-panel'
import { RouteIntelligencePanel } from '@/components/aircraft/route-intelligence-panel'
import { StabilityIntelligencePanel } from '@/components/aircraft/stability-intelligence-panel'
import { WeatherContextPanel } from '@/components/aircraft/weather-context-panel'
import { TrafficGlobe } from '@/components/globe/traffic-globe'
import { MapEvidenceWorkspace } from '@/components/map/map-evidence-workspace'
import { LiveTrafficControl } from '@/components/traffic/live-traffic-control'
import { TrafficSnapshotExport } from '@/components/traffic/traffic-snapshot-export'
import { TrafficDataQualityLens } from '@/components/traffic/traffic-data-quality-lens'
import { RegionalTrafficBrief } from '@/components/traffic/regional-traffic-brief'
import { getRequestErrorMessage } from '@/lib/api/client'
import {
  defaultMapEvidenceVisibility,
  shouldRenderProjection,
  shouldRenderTrajectory,
  toggleProjectionVisibility,
  toggleTrajectoryVisibility,
} from '@/lib/map/map-evidence-controls'
import { useProjectionIntelligence } from '@/lib/queries/projection-intelligence'
import { useAircraftRouteContext } from '@/lib/queries/route-context'
import { useProcessedRouteIntelligence } from '@/lib/queries/route-intelligence'
import {
  buildStabilityAsOfTimes,
  useStabilityIntelligence,
} from '@/lib/queries/stability-intelligence'
import { useCurrentTraffic } from '@/lib/queries/traffic'
import {
  defaultLiveTrafficRefreshIntervalMilliseconds,
  type LiveTrafficRefreshIntervalMilliseconds,
} from '@/lib/traffic/live-traffic-status-model'
import { useLatestAircraftTrajectory } from '@/lib/queries/trajectory'
import { useWeatherContext } from '@/lib/queries/weather-context'
import {
  buildTrafficWorkspaceSelection,
  type TrafficWorkspacePanel,
  type TrafficWorkspaceSelection,
} from '@/lib/traffic/traffic-workspace-model'
import type { Region } from '@/types/region'
import type { TrafficAircraft } from '@/types/traffic'

interface TrafficDashboardProps {
  regions: Region[]
  selectedRegion: Region
  selectedAircraftICAO24: string | null
  workspacePanel: TrafficWorkspacePanel
  onSelectedRegionCodeChange: (regionCode: string) => void
  onWorkspaceSelectionChange: (selection: TrafficWorkspaceSelection) => void
  onWorkspacePanelChange: (panel: TrafficWorkspacePanel) => void
  initialTraffic: TrafficAircraft[]
  initialError: string | null
  regionsWarning: string | null
}

type ShareViewStatus = 'idle' | 'copied' | 'unavailable'

export function TrafficDashboard({
  regions,
  selectedRegion,
  selectedAircraftICAO24,
  workspacePanel,
  onSelectedRegionCodeChange,
  onWorkspaceSelectionChange,
  onWorkspacePanelChange,
  initialTraffic,
  initialError,
  regionsWarning,
}: TrafficDashboardProps) {
  const [shareViewStatus, setShareViewStatus] =
    useState<ShareViewStatus>('idle')
  const [autoRefreshEnabled, setAutoRefreshEnabled] = useState(true)
  const [refreshIntervalMilliseconds, setRefreshIntervalMilliseconds] =
    useState<LiveTrafficRefreshIntervalMilliseconds>(
      defaultLiveTrafficRefreshIntervalMilliseconds
    )
  const [mapEvidenceVisibility, setMapEvidenceVisibility] = useState(
    defaultMapEvidenceVisibility
  )
  const [statusNow, setStatusNow] = useState(() => Date.now())

  useEffect(() => {
    const intervalID = window.setInterval(() => {
      setStatusNow(Date.now())
    }, 1000)

    return () => {
      window.clearInterval(intervalID)
    }
  }, [])

  const initialData =
    selectedRegion.code === 'world' && initialError === null
      ? initialTraffic
      : undefined
  const trafficQuery = useCurrentTraffic(selectedRegion.code, {
    initialData,
    refreshIntervalMilliseconds: autoRefreshEnabled
      ? refreshIntervalMilliseconds
      : false,
  })
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
  const trajectoryFeatureCount = trajectoryQuery.data?.segments.length ?? 0
  const projectionPointCount = projectionQuery.data?.projection.points.length ?? 0
  const trajectoryVisible = shouldRenderTrajectory(
    mapEvidenceVisibility,
    trajectoryFeatureCount
  )
  const projectionVisible = shouldRenderProjection(
    mapEvidenceVisibility,
    projectionPointCount
  )

  const selectAircraft = (icao24: string | null) => {
    onWorkspaceSelectionChange(buildTrafficWorkspaceSelection(icao24))
  }

  const changeRegion = (regionCode: string) => {
    onSelectedRegionCodeChange(regionCode)
  }

  const copyCurrentViewLink = async () => {
    if (
      typeof window === 'undefined' ||
      typeof navigator === 'undefined' ||
      navigator.clipboard?.writeText === undefined
    ) {
      setShareViewStatus('unavailable')
      return
    }

    try {
      await navigator.clipboard.writeText(window.location.href)
      setShareViewStatus('copied')
    } catch {
      setShareViewStatus('unavailable')
    }
  }

  return (
    <>
      <section className='mt-3 rounded-xl border border-slate-800 bg-[#17191c] p-3 sm:p-4'>
        <div className='flex flex-wrap items-end justify-between gap-3'>
          <div className='min-w-56 flex-1 sm:max-w-sm'>
            <label
              className='text-[10px] font-semibold uppercase tracking-[0.16em] text-slate-500'
              htmlFor='traffic-region'
            >
              Traffic region
            </label>
            <select
              id='traffic-region'
              value={selectedRegion.code}
              onChange={(event: ChangeEvent<HTMLSelectElement>) => {
                changeRegion(event.target.value)
              }}
              className='mt-1.5 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm font-semibold text-white'
            >
              {regions.map(region => (
                <option key={region.code} value={region.code}>
                  {region.name}
                </option>
              ))}
            </select>
          </div>
          <div className='flex flex-wrap gap-2'>
            <button
              type='button'
              onClick={() => {
                void copyCurrentViewLink()
              }}
              className='rounded-lg border border-sky-400/35 bg-sky-400/5 px-3 py-2 text-xs font-semibold text-sky-100 transition hover:bg-sky-400/10'
            >
              {shareViewStatus === 'copied'
                ? 'Link copied'
                : shareViewStatus === 'unavailable'
                  ? 'Clipboard unavailable'
                  : 'Share view'}
            </button>
            <button
              type='button'
              onClick={() => {
                void trafficQuery.refetch()
              }}
              disabled={trafficQuery.isFetching}
              className='rounded-lg border border-slate-700 px-3 py-2 text-xs font-semibold text-slate-200 transition hover:bg-slate-900 disabled:opacity-60'
            >
              {trafficQuery.isFetching ? 'Refreshing…' : 'Refresh'}
            </button>
          </div>
        </div>

        <LiveTrafficControl
          regionName={selectedRegion.name}
          aircraftCount={traffic.length}
          selectedAircraftICAO24={selectedAircraftICAO24}
          dataUpdatedAt={trafficQuery.dataUpdatedAt}
          now={statusNow}
          isInitialLoading={isInitialLoading}
          isRefreshing={isRefreshing}
          errorMessage={errorMessage}
          regionsWarning={regionsWarning}
          autoRefreshEnabled={autoRefreshEnabled}
          refreshIntervalMilliseconds={refreshIntervalMilliseconds}
          onAutoRefreshEnabledChange={setAutoRefreshEnabled}
          onRefreshIntervalChange={setRefreshIntervalMilliseconds}
          onRetry={() => {
            void trafficQuery.refetch()
          }}
        />

        <TrafficSnapshotExport
          aircraft={traffic}
          regionCode={selectedRegion.code}
          regionName={selectedRegion.name}
          snapshotUpdatedAt={trafficQuery.dataUpdatedAt}
          selectedAircraftICAO24={selectedAircraftICAO24}
        />
      </section>

      <section
        className='mt-3 overflow-hidden rounded-xl border border-slate-700/80 bg-[#111315] shadow-2xl shadow-black/25'
        data-visual-polish='tracker-workspace-v2'
      >
        <div className='flex flex-wrap items-center gap-2 border-b border-white/10 bg-[#202225] px-3 py-2.5 sm:px-4'>
          <div className='mr-auto min-w-0'>
            <h2 className='truncate text-sm font-bold text-white sm:text-base'>
              Live Traffic · {selectedRegion.name}
            </h2>
            <p className='mt-0.5 text-[10px] uppercase tracking-[0.14em] text-slate-500'>
              Map-first open-data tracker
            </p>
          </div>

          <div
            className='flex flex-wrap items-center gap-1.5'
            role='group'
            aria-label='Map evidence visibility'
          >
            <button
              type='button'
              aria-pressed={mapEvidenceVisibility.trajectory}
              disabled={trajectoryFeatureCount === 0}
              onClick={() => {
                setMapEvidenceVisibility(current =>
                  toggleTrajectoryVisibility(current)
                )
              }}
              className={`rounded-full border px-2.5 py-1.5 text-[10px] font-bold uppercase tracking-[0.1em] transition disabled:cursor-not-allowed disabled:opacity-35 ${
                mapEvidenceVisibility.trajectory
                  ? 'border-sky-400/45 bg-sky-400/10 text-sky-100'
                  : 'border-slate-700 text-slate-400 hover:bg-slate-900'
              }`}
            >
              Trail
            </button>
            <button
              type='button'
              aria-pressed={mapEvidenceVisibility.projection}
              disabled={projectionPointCount === 0}
              onClick={() => {
                setMapEvidenceVisibility(current =>
                  toggleProjectionVisibility(current)
                )
              }}
              className={`rounded-full border px-2.5 py-1.5 text-[10px] font-bold uppercase tracking-[0.1em] transition disabled:cursor-not-allowed disabled:opacity-35 ${
                mapEvidenceVisibility.projection
                  ? 'border-violet-300/45 bg-violet-400/10 text-violet-100'
                  : 'border-slate-700 text-slate-400 hover:bg-slate-900'
              }`}
            >
              Projection
            </button>
            <span className='hidden px-1 text-[10px] text-slate-500 sm:inline' aria-live='polite'>
              {trajectoryVisible ? 'Trail on' : 'Trail off'} ·{' '}
              {projectionVisible ? 'Projection on' : 'Projection off'}
            </span>
          </div>
        </div>

        <div className='grid xl:grid-cols-[minmax(0,1fr)_420px]'>
          <div className='min-w-0 bg-slate-950' aria-busy={trafficQuery.isFetching}>
            <MapEvidenceWorkspace
              aircraft={traffic}
              region={selectedRegion}
              selectedAircraftICAO24={selectedAircraftICAO24}
              trajectory={trajectoryQuery.data}
              projection={projectionQuery.data?.projection}
              visibility={mapEvidenceVisibility}
              onSelectAircraft={selectAircraft}
            />
          </div>

          <aside
            className='gfa-tracker-rail min-w-0 border-t border-white/10 bg-[#17191c] p-2.5 xl:max-h-[min(78vh,860px)] xl:overflow-y-auto xl:border-l xl:border-t-0'
            aria-label='Traffic workspace'
          >
            <WorkspaceTabs
              activePanel={workspacePanel}
              selectedAircraftICAO24={selectedAircraftICAO24}
              onPanelChange={onWorkspacePanelChange}
            />

            <div
              id='traffic-workspace-panel'
              role='tabpanel'
              aria-labelledby={`traffic-workspace-${workspacePanel}-tab`}
              className='mt-2.5'
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
                  onOpenAircraft={() => onWorkspacePanelChange('aircraft')}
                />
              ) : (
                <div className='space-y-3'>
                  <SelectedAircraftContext
                    icao24={selectedAircraftICAO24}
                    onOpenAircraft={() => onWorkspacePanelChange('aircraft')}
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
          <p className='border-t border-white/10 px-4 py-3 text-sm text-slate-400'>
            No aircraft were returned for the selected region.
          </p>
        ) : null}
      </section>

      <div className='mt-6' aria-busy={trafficQuery.isFetching}>
        <TrafficGlobe aircraft={traffic} region={selectedRegion} />
      </div>

      <RegionalTrafficBrief
        aircraft={traffic}
        regionName={selectedRegion.name}
        isFetching={trafficQuery.isFetching}
      />

      <TrafficDataQualityLens
        aircraft={traffic}
        regionName={selectedRegion.name}
        snapshotUpdatedAt={trafficQuery.dataUpdatedAt}
        isFetching={trafficQuery.isFetching}
      />
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
      className='grid grid-cols-2 gap-1 rounded-lg border border-slate-800 bg-slate-950 p-1'
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
      <span className='block text-xs font-semibold'>{label}</span>
      <span className='mt-0.5 block truncate text-[10px] opacity-75'>
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
        projection, weather and stability evidence will open here only when the
        corresponding GFA evidence exists.
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
    <div className='flex flex-wrap items-center justify-between gap-2 rounded-lg border border-amber-300/35 bg-amber-300/5 p-2.5'>
      <div>
        <p className='text-[10px] font-semibold uppercase tracking-[0.16em] text-amber-300'>
          Selected aircraft
        </p>
        <p className='mt-0.5 font-mono text-sm font-bold uppercase text-white'>
          {icao24}
        </p>
      </div>
      <div className='flex flex-wrap gap-1.5'>
        <button
          type='button'
          onClick={onOpenAircraft}
          className='rounded-md border border-slate-700 px-2.5 py-1.5 text-[11px] font-medium text-slate-200 hover:bg-slate-900'
        >
          Aircraft list
        </button>
        <button
          type='button'
          onClick={onClear}
          className='rounded-md border border-rose-400/30 px-2.5 py-1.5 text-[11px] font-medium text-rose-100 hover:bg-rose-400/10'
        >
          Clear
        </button>
      </div>
    </div>
  )
}
