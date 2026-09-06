'use client'

import { useMemo } from 'react'

import { AirspaceIntelligencePanel } from '@/components/traffic/airspace-intelligence-panel'
import { resolveAirspaceAsOfTime } from '@/lib/airspace/airspace-intelligence-model'
import { useAirspaceRegionAnalytics } from '@/lib/queries/airspace-intelligence'
import { useCurrentTraffic } from '@/lib/queries/traffic'
import type { Region } from '@/types/region'

export function AirspaceIntelligenceWorkspace({
  selectedRegion,
}: {
  selectedRegion: Region
}) {
  const trafficQuery = useCurrentTraffic(selectedRegion.code)
  const airspaceAsOfTime = useMemo(
    () => resolveAirspaceAsOfTime(trafficQuery.data ?? []),
    [trafficQuery.data]
  )
  const airspaceQuery = useAirspaceRegionAnalytics(
    selectedRegion.code,
    airspaceAsOfTime
  )

  return (
    <AirspaceIntelligencePanel
      regionCode={selectedRegion.code}
      regionName={selectedRegion.name}
      asOfTime={airspaceAsOfTime}
      result={airspaceQuery.data}
      isPending={airspaceQuery.isPending}
      isFetching={airspaceQuery.isFetching}
      error={airspaceQuery.error}
      onRetry={() => {
        void airspaceQuery.refetch()
      }}
    />
  )
}
