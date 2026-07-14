'use client'

import { useMemo, useState } from 'react'

import { AnalyticsOverview } from '@/components/analytics/analytics-overview'
import { TrafficDashboard } from '@/components/traffic-dashboard'
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
  const [selectedRegionCode, setSelectedRegionCode] = useState(
    resolveInitialRegionCode(regions)
  )

  const selectedRegion = useMemo(
    () =>
      regions.find(region => region.code === selectedRegionCode) ??
      regions[0],
    [regions, selectedRegionCode]
  )

  if (!selectedRegion) {
    return (
      <p className='mt-8 rounded-xl border border-rose-400/40 bg-rose-400/10 p-4 text-rose-100'>
        No traffic regions are available.
      </p>
    )
  }

  return (
    <>
      <AnalyticsOverview selectedRegion={selectedRegion} />

      <TrafficDashboard
        regions={regions}
        selectedRegionCode={selectedRegion.code}
        onSelectedRegionCodeChange={nextRegionCode => {
          if (regions.some(region => region.code === nextRegionCode)) {
            setSelectedRegionCode(nextRegionCode)
          }
        }}
        initialTraffic={initialTraffic}
        initialError={initialError}
        regionsWarning={regionsWarning}
      />
    </>
  )
}

function resolveInitialRegionCode(regions: Region[]): string {
  if (regions.some(region => region.code === 'world')) {
    return 'world'
  }

  return regions[0]?.code ?? ''
}
