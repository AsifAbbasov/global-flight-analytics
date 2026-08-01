// FRONTEND_SHAREABLE_WORKSPACE_STATE_V1
'use client'

import { useEffect, useMemo, useState } from 'react'

import { AnalyticsOverview } from '@/components/analytics/analytics-overview'
import { TrafficDashboard } from '@/components/traffic-dashboard'
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
    return (
      <p className='mt-8 rounded-xl border border-rose-400/40 bg-rose-400/10 p-4 text-rose-100'>
        No traffic regions are available.
      </p>
    )
  }

  return (
    <>
      <div id='overview' className='scroll-mt-28'>
        <AnalyticsOverview selectedRegion={selectedRegion} />
      </div>

      <div id='live-traffic' className='scroll-mt-28'>
        <TrafficDashboard
          regions={regions}
          selectedRegion={selectedRegion}
          selectedAircraftICAO24={selectedAircraftICAO24}
          workspacePanel={workspacePanel}
          onSelectedRegionCodeChange={changeRegion}
          onWorkspaceSelectionChange={selectAircraft}
          onWorkspacePanelChange={changeWorkspacePanel}
          initialTraffic={initialTraffic}
          initialError={initialError}
          regionsWarning={regionsWarning}
        />
      </div>
    </>
  )
}

function resolveInitialRegionCode(regions: Region[]): string {
  if (regions.some(region => region.code === 'world')) {
    return 'world'
  }

  return regions[0]?.code ?? ''
}
