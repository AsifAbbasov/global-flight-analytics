'use client'

import { TrafficMap } from '@/components/map/traffic-map'
import {
  shouldRenderProjection,
  shouldRenderTrajectory,
  type MapEvidenceVisibility,
} from '@/lib/map/map-evidence-controls'
import type { ProjectionResult } from '@/types/projection-intelligence'
import type { Region } from '@/types/region'
import type { TrafficAircraft } from '@/types/traffic'
import type { AircraftTrajectory } from '@/types/trajectory'

interface MapEvidenceWorkspaceProps {
  aircraft: TrafficAircraft[]
  region: Region
  selectedAircraftICAO24: string | null
  trajectory: AircraftTrajectory | undefined
  projection: ProjectionResult | undefined
  visibility: MapEvidenceVisibility
  onSelectAircraft: (icao24: string) => void
}

export function MapEvidenceWorkspace({
  aircraft,
  region,
  selectedAircraftICAO24,
  trajectory,
  projection,
  visibility,
  onSelectAircraft,
}: MapEvidenceWorkspaceProps) {
  const trajectoryVisible = shouldRenderTrajectory(
    visibility,
    trajectory?.segments.length ?? 0
  )
  const projectionVisible = shouldRenderProjection(
    visibility,
    projection?.points.length ?? 0
  )

  return (
    <TrafficMap
      aircraft={aircraft}
      region={region}
      selectedAircraftICAO24={selectedAircraftICAO24}
      trajectory={trajectoryVisible ? trajectory : undefined}
      projection={projectionVisible ? projection : undefined}
      onSelectAircraft={onSelectAircraft}
    />
  )
}
