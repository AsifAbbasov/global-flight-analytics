// FRONTEND_TRAFFIC_DATA_QUALITY_LENS_V1
// FRONTEND_MAP_FIRST_REDESIGN_V1
// FRONTEND_FLIGHT_TRACKER_REFERENCE_V2
// FRONTEND_FLIGHT_TRACKER_REFERENCE_V3
// FRONTEND_FLIGHT_TRACKER_CLEAN_REBUILD_V1
// FRONTEND_RUNTIME_STABILIZATION_V2
// FRONTEND_AIRCRAFT_ONLY_SCOPE_V1
// FRONTEND_EXCLUSIVE_MAP_POPOVERS_V1
// FRONTEND_UNIFIED_RIGHT_SIDEBAR_EXCLUSIVITY_V1
'use client'

import { useEffect, useState, type ChangeEvent } from 'react'

import { AircraftDetailPanel } from '@/components/aircraft/aircraft-detail-panel'
import { AircraftExplorer } from '@/components/aircraft/aircraft-explorer'
import { ProjectionIntelligencePanel } from '@/components/aircraft/projection-intelligence-panel'
import { RouteIntelligencePanel } from '@/components/aircraft/route-intelligence-panel'
import { StabilityIntelligencePanel } from '@/components/aircraft/stability-intelligence-panel'
import { WeatherContextPanel } from '@/components/aircraft/weather-context-panel'
import { TrafficGlobe } from '@/components/globe/traffic-globe'
import { TrafficMap } from '@/components/map/traffic-map'
import { LiveTrafficControl } from '@/components/traffic/live-traffic-control'
import { RegionalTrafficBrief } from '@/components/traffic/regional-traffic-brief'
import { TrafficDataQualityLens } from '@/components/traffic/traffic-data-quality-lens'
import { TrafficSnapshotExport } from '@/components/traffic/traffic-snapshot-export'
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
  defaultLiveTrafficRefreshIntervalMilliseconds,
  type LiveTrafficRefreshIntervalMilliseconds,
} from '@/lib/traffic/live-traffic-status-model'
import {
  buildTrafficWorkspaceSelection,
  type TrafficWorkspacePanel,
  type TrafficWorkspaceSelection,
} from '@/lib/traffic/traffic-workspace-model'
import type { Region } from '@/types/region'
import type { TrafficAircraft } from '@/types/traffic'

export type MapToolPopoverID = 'live-data' | 'traffic-analysis'

interface TrafficDashboardProps {
  regions: Region[]
  selectedRegion: Region
  selectedAircraftICAO24: string | null
  workspacePanel: TrafficWorkspacePanel
  onSelectedRegionCodeChange: (regionCode: string) => void
  onWorkspaceSelectionChange: (selection: TrafficWorkspaceSelection) => void
  onWorkspacePanelChange: (panel: TrafficWorkspacePanel) => void
  activeMapPopover: MapToolPopoverID | null
  onMapPopoverToggle: (popover: MapToolPopoverID, open: boolean) => void
  onCloseMapPopovers: () => void
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
  activeMapPopover,
  onMapPopoverToggle,
  onCloseMapPopovers,
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
  const [statusNow, setStatusNow] = useState(0)

  useEffect(() => {
    const animationFrameID = window.requestAnimationFrame(() => {
      setStatusNow(Date.now())
    })
    const intervalID = window.setInterval(() => {
      setStatusNow(Date.now())
    }, 1000)

    return () => {
      window.cancelAnimationFrame(animationFrameID)
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
    <section className='gfa-tracker-page' aria-label='Live flight tracker'>
      <aside className='gfa-tracker-sidebar' aria-label='Traffic workspace'>
        <div className='gfa-sidebar-heading'>
          <div>
            <p>Flight tracker map</p>
            <strong>{selectedRegion.name}</strong>
          </div>
          <span className='gfa-sidebar-live'>
            <span aria-hidden='true' />
            LIVE
          </span>
        </div>

        <WorkspaceTabs
          activePanel={workspacePanel}
          selectedAircraftICAO24={selectedAircraftICAO24}
          onPanelChange={onWorkspacePanelChange}
        />

        <div
          id='traffic-workspace-panel'
          role='tabpanel'
          aria-labelledby={`traffic-workspace-${workspacePanel}-tab`}
          className='gfa-sidebar-scroll'
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
            <div className='space-y-4'>
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

      <div className='gfa-tracker-map-stage' aria-busy={trafficQuery.isFetching}>
        <TrafficMap
          aircraft={traffic}
          region={selectedRegion}
          selectedAircraftICAO24={selectedAircraftICAO24}
          trajectory={trajectoryQuery.data}
          projection={projectionQuery.data?.projection}
          onSelectAircraft={selectAircraft}
        />

        <div className='gfa-map-commandbar' aria-label='Map controls'>
          <label className='gfa-map-region' htmlFor='traffic-region'>
            <span aria-hidden='true'>⌖</span>
            <select
              id='traffic-region'
              value={selectedRegion.code}
              onChange={(event: ChangeEvent<HTMLSelectElement>) => {
                changeRegion(event.target.value)
              }}
              aria-label='Traffic region'
            >
              {regions.map(region => (
                <option key={region.code} value={region.code}>
                  {region.name}
                </option>
              ))}
            </select>
          </label>

          <button
            type='button'
            onClick={() => {
              void trafficQuery.refetch()
            }}
            disabled={trafficQuery.isFetching}
            className='gfa-map-command'
            aria-label='Refresh current traffic'
            title='Refresh traffic'
          >
            ↻
          </button>
          <button
            type='button'
            onClick={() => {
              void copyCurrentViewLink()
            }}
            className='gfa-map-command'
            aria-label='Share current map view'
            title={shareViewStatus === 'copied' ? 'Link copied' : 'Share map view'}
          >
            ↗
          </button>
        </div>

        <nav className='gfa-map-toolrail' aria-label='Map tools'>
          <a
            href='#overview'
            className='gfa-map-tool'
            aria-label='Open analytics overview'
            title='Analytics'
            onClick={onCloseMapPopovers}
          >
            <span aria-hidden='true'>▦</span>
          </a>
          <a
            href='#historical-analytics'
            className='gfa-map-tool'
            aria-label='Open Historical Analytics'
            title='History'
            onClick={onCloseMapPopovers}
          >
            <span aria-hidden='true'>◷</span>
          </a>

          <details
            className='gfa-map-tool-menu'
            name='gfa-map-tool-popovers'
            open={activeMapPopover === 'live-data'}
            onToggle={event => {
              onMapPopoverToggle('live-data', event.currentTarget.open)
            }}
          >
            <summary className='gfa-map-tool' aria-label='Open live data controls' title='Live data'>
              <span aria-hidden='true'>☷</span>
            </summary>
            <div className='gfa-map-popover'>
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
            </div>
          </details>

          <details
            className='gfa-map-tool-menu'
            name='gfa-map-tool-popovers'
            open={activeMapPopover === 'traffic-analysis'}
            onToggle={event => {
              onMapPopoverToggle('traffic-analysis', event.currentTarget.open)
            }}
          >
            <summary className='gfa-map-tool' aria-label='Open traffic analysis' title='Traffic analysis'>
              <span aria-hidden='true'>≋</span>
            </summary>
            <div className='gfa-map-popover gfa-map-analysis-popover'>
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
              <div className='gfa-globe-slot' aria-busy={trafficQuery.isFetching}>
                <TrafficGlobe aircraft={traffic} region={selectedRegion} />
              </div>
            </div>
          </details>
        </nav>

        <div className='gfa-map-statusbar' aria-live='polite'>
          <span className='gfa-map-live'>LIVE</span>
          <strong>{formatInteger(traffic.length)}</strong>
          <span>aircraft</span>
          <i aria-hidden='true' />
          <span>{selectedRegion.name}</span>
          {selectedAircraftICAO24 ? (
            <>
              <i aria-hidden='true' />
              <span className='font-mono uppercase'>{selectedAircraftICAO24}</span>
            </>
          ) : null}
        </div>

        {errorMessage ? (
          <div className='gfa-map-alert' role='alert'>
            <strong>Live data unavailable</strong>
            <span>{errorMessage}</span>
            <button
              type='button'
              onClick={() => {
                void trafficQuery.refetch()
              }}
            >
              Retry
            </button>
          </div>
        ) : null}

        {!trafficQuery.isFetching && !errorMessage && traffic.length === 0 ? (
          <div className='gfa-map-alert'>
            <strong>No live aircraft</strong>
            <span>No aircraft were returned for {selectedRegion.name}.</span>
          </div>
        ) : null}
      </div>
    </section>
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
      className='gfa-workspace-tabs'
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
      className={active ? 'gfa-workspace-tab is-active' : 'gfa-workspace-tab'}
    >
      <span>{label}</span>
      <small>{detail}</small>
    </button>
  )
}

function IntelligenceSelectionPrompt({
  onOpenAircraft,
}: {
  onOpenAircraft: () => void
}) {
  return (
    <div className='gfa-intelligence-prompt'>
      <span className='gfa-prompt-plane' aria-hidden='true'>
        <svg viewBox='0 0 24 24' focusable='false'>
          <path d='M12 2.25c.61 0 .98.45.98 1.05v5.58l6.62 3.59v1.67l-6.62-1.8v4.8l2 1.65v1.11L12 19l-2.98.9v-1.11l2-1.65v-4.8l-6.62 1.8v-1.67l6.62-3.59V3.3c0-.6.37-1.05.98-1.05Z' />
        </svg>
      </span>
      <p className='gfa-prompt-title'>Select an aircraft to open Intelligence</p>
      <p className='gfa-prompt-copy'>
        Use the Aircraft tab or select a marker on the map. Route, trajectory,
        projection, weather and stability evidence will open here.
      </p>
      <button type='button' onClick={onOpenAircraft} className='gfa-prompt-button'>
        Browse aircraft
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
    <div className='gfa-selected-aircraft-context'>
      <div>
        <p>Selected aircraft</p>
        <strong>{icao24.toUpperCase()}</strong>
      </div>
      <div className='gfa-selected-aircraft-actions'>
        <button type='button' onClick={onOpenAircraft}>
          Aircraft list
        </button>
        <button type='button' onClick={onClear} className='is-danger'>
          Clear selection
        </button>
      </div>
    </div>
  )
}

function formatInteger(value: number): string {
  return new Intl.NumberFormat().format(value)
}
