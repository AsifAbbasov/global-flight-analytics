// FRONTEND_SHAREABLE_WORKSPACE_STATE_V1
// FRONTEND_HISTORICAL_ANALYTICS_COMPARISON_V1
// FRONTEND_PRODUCT_HARDENING_V1
// FRONTEND_MAP_FIRST_REDESIGN_V1
// FRONTEND_FLIGHT_TRACKER_REFERENCE_V2
// FRONTEND_RUNTIME_STABILIZATION_V2
// FRONTEND_AIRCRAFT_ONLY_SCOPE_V1
// FRONTEND_UNIFIED_RIGHT_SIDEBAR_EXCLUSIVITY_V1
'use client'

import dynamic from 'next/dynamic'
import { useEffect, useMemo, useState, type ReactNode } from 'react'

import { AnalyticsOverview } from '@/components/analytics/analytics-overview'
import type { MapToolPopoverID } from '@/components/traffic-dashboard'
import {
  type TrafficWorkspacePanel,
  type TrafficWorkspaceSelection,
} from '@/lib/traffic/traffic-workspace-model'
import {
  buildTrafficWorkspaceURL,
  parseTrafficWorkspaceSearch,
  type TrafficWorkspaceURLState,
} from '@/lib/traffic/workspace-url-state'
import type { Region } from '@/types/region'
import type { TrafficAircraft } from '@/types/traffic'

const HistoricalAnalyticsComparisonWorkspace = dynamic(
  () =>
    import(
      '@/components/analytics/historical-analytics-comparison-workspace'
    ).then(module => module.HistoricalAnalyticsComparisonWorkspace),
  {
    loading: () => <ResearchSectionLoading label='Historical Analytics' />,
  }
)

const TrafficDashboard = dynamic(
  () =>
    import('@/components/traffic-dashboard').then(
      module => module.TrafficDashboard
    ),
  {
    loading: () => <ResearchSectionLoading label='Live traffic workspace' />,
  }
)

interface RegionalTrafficExperienceProps {
  regions: Region[]
  initialTraffic: TrafficAircraft[]
  initialError: string | null
  regionsWarning: string | null
}

export function RegionalTrafficExperience({
  regions,
  initialTraffic,
  initialError,
  regionsWarning,
}: RegionalTrafficExperienceProps) {
  const availableRegionCodes = useMemo(
    () => regions.map(region => region.code),
    [regions]
  )
  const fallbackRegionCode = resolveInitialRegionCode(regions)
  const [selectedRegionCode, setSelectedRegionCode] = useState(
    fallbackRegionCode
  )
  const [selectedAircraftICAO24, setSelectedAircraftICAO24] = useState<
    string | null
  >(null)
  const [workspacePanel, setWorkspacePanel] =
    useState<TrafficWorkspacePanel>('aircraft')
  const [activeResearchPanel, setActiveResearchPanel] =
    useState<ResearchPanelID | null>(null)
  const [activeMapPopover, setActiveMapPopover] =
    useState<MapToolPopoverID | null>(null)

  const selectedRegion = useMemo(
    () =>
      regions.find(region => region.code === selectedRegionCode) ??
      regions[0],
    [regions, selectedRegionCode]
  )

  useEffect(() => {
    if (typeof window === 'undefined') {
      return
    }

    const restoreLocationState = () => {
      const nextState = parseTrafficWorkspaceSearch(
        window.location.search,
        availableRegionCodes,
        fallbackRegionCode
      )
      setSelectedRegionCode(nextState.regionCode)
      setSelectedAircraftICAO24(nextState.aircraftICAO24)
      setWorkspacePanel(nextState.panel)
      return nextState
    }

    const initialState = restoreLocationState()
    const canonicalURL = buildTrafficWorkspaceURL(
      window.location.pathname,
      window.location.search,
      window.location.hash,
      initialState
    )
    const currentURL = `${window.location.pathname}${window.location.search}${window.location.hash}`
    if (canonicalURL !== currentURL) {
      window.history.replaceState(window.history.state, '', canonicalURL)
    }

    const handlePopState = () => {
      restoreLocationState()
    }
    window.addEventListener('popstate', handlePopState)
    return () => {
      window.removeEventListener('popstate', handlePopState)
    }
  }, [availableRegionCodes, fallbackRegionCode])

  useEffect(() => {
    const syncResearchPanel = () => {
      const nextResearchPanel = resolveResearchPanelID(window.location.hash)
      setActiveResearchPanel(nextResearchPanel)
      if (nextResearchPanel !== null) {
        setActiveMapPopover(null)
      }
    }

    syncResearchPanel()
    window.addEventListener('hashchange', syncResearchPanel)

    return () => {
      window.removeEventListener('hashchange', syncResearchPanel)
    }
  }, [])

  const commitWorkspaceState = (nextState: TrafficWorkspaceURLState) => {
    setSelectedRegionCode(nextState.regionCode)
    setSelectedAircraftICAO24(nextState.aircraftICAO24)
    setWorkspacePanel(nextState.panel)

    if (typeof window === 'undefined') {
      return
    }

    const nextURL = buildTrafficWorkspaceURL(
      window.location.pathname,
      window.location.search,
      window.location.hash,
      nextState
    )
    const currentURL = `${window.location.pathname}${window.location.search}${window.location.hash}`
    if (nextURL !== currentURL) {
      window.history.pushState(window.history.state, '', nextURL)
    }
  }

  const changeRegion = (nextRegionCode: string) => {
    if (!regions.some(region => region.code === nextRegionCode)) {
      return
    }

    commitWorkspaceState({
      regionCode: nextRegionCode,
      aircraftICAO24: null,
      panel: 'aircraft',
    })
  }

  const selectAircraft = (selection: TrafficWorkspaceSelection) => {
    commitWorkspaceState({
      regionCode: selectedRegionCode,
      aircraftICAO24: selection.icao24,
      panel: selection.panel,
    })
  }

  const changeWorkspacePanel = (panel: TrafficWorkspacePanel) => {
    commitWorkspaceState({
      regionCode: selectedRegionCode,
      aircraftICAO24: selectedAircraftICAO24,
      panel,
    })
  }

  if (!selectedRegion) {
    return <p className='gfa-empty-state'>No traffic regions are available.</p>
  }

  const handleMapPopoverToggle = (
    popover: MapToolPopoverID,
    open: boolean
  ) => {
    if (!open) {
      setActiveMapPopover(current =>
        current === popover ? null : current
      )
      return
    }

    setActiveMapPopover(popover)
    setActiveResearchPanel(null)

    if (
      typeof window !== 'undefined' &&
      resolveResearchPanelID(window.location.hash) !== null
    ) {
      window.location.replace('#live-traffic')
    }
  }

  const closeMapPopovers = () => {
    setActiveMapPopover(null)
  }

  return (
    <>
      <div id='live-traffic' className='gfa-primary-live-slot'>
        <TrafficDashboard
          regions={regions}
          selectedRegion={selectedRegion}
          selectedAircraftICAO24={selectedAircraftICAO24}
          workspacePanel={workspacePanel}
          onSelectedRegionCodeChange={changeRegion}
          onWorkspaceSelectionChange={selectAircraft}
          onWorkspacePanelChange={changeWorkspacePanel}
          activeMapPopover={activeMapPopover}
          onMapPopoverToggle={handleMapPopoverToggle}
          onCloseMapPopovers={closeMapPopovers}
          initialTraffic={initialTraffic}
          initialError={initialError}
          regionsWarning={regionsWarning}
        />
      </div>

      <section className='gfa-analytics-deck' aria-label='Analytical workspaces'>
        <FloatingAnalyticsPanel
          id='overview'
          kicker='Regional intelligence'
          title='Analytics overview'
          active={activeResearchPanel === 'overview'}
        >
          <AnalyticsOverview selectedRegion={selectedRegion} />
        </FloatingAnalyticsPanel>

        <FloatingAnalyticsPanel
          id='historical-analytics'
          kicker='Historical intelligence'
          title='Historical workspace'
          active={activeResearchPanel === 'historical-analytics'}
        >
          <HistoricalAnalyticsComparisonWorkspace />
        </FloatingAnalyticsPanel>
      </section>
    </>
  )
}

function FloatingAnalyticsPanel({
  id,
  kicker,
  title,
  active,
  children,
}: {
  id: string
  kicker: string
  title: string
  active: boolean
  children: ReactNode
}) {
  return (
    <section id={id} className='gfa-floating-analytics-panel'>
      <header className='gfa-drawer-header'>
        <div>
          <p className='gfa-section-kicker'>{kicker}</p>
          <h2>{title}</h2>
        </div>
        <a href='#live-traffic' className='gfa-drawer-close' aria-label={`Close ${title}`}>
          ×
        </a>
      </header>
      <div className='gfa-floating-analytics-scroll'>
        {active ? children : null}
      </div>
    </section>
  )
}

function ResearchSectionLoading({ label }: { label: string }) {
  return (
    <section
      role='status'
      aria-live='polite'
      aria-busy='true'
      className='gfa-research-loading'
    >
      <span className='sr-only'>Loading {label}.</span>
      <div aria-hidden='true' className='animate-pulse'>
        <div className='h-3 w-40 rounded bg-zinc-800' />
        <div className='mt-4 h-8 max-w-xl rounded bg-zinc-800/80' />
        <div className='mt-5 grid gap-3 lg:grid-cols-3'>
          {Array.from({ length: 3 }, (_, index) => (
            <div
              key={index}
              className='h-28 rounded-lg border border-white/5 bg-black/20'
            />
          ))}
        </div>
      </div>
    </section>
  )
}

type ResearchPanelID =
  | 'overview'
  | 'historical-analytics'

function resolveResearchPanelID(
  hash: string
): ResearchPanelID | null {
  const normalizedHash = hash.trim().toLowerCase()

  switch (normalizedHash) {
    case '#overview':
      return 'overview'
    case '#historical-analytics':
      return 'historical-analytics'
    default:
      return null
  }
}

function resolveInitialRegionCode(regions: Region[]): string {
  if (regions.some(region => region.code === 'world')) {
    return 'world'
  }

  return regions[0]?.code ?? ''
}
