'use client'

import { useEffect, useMemo, useState } from 'react'

import { FlightReplayControl } from '@/components/aircraft/flight-replay-control'
import { TrafficMap } from '@/components/map/traffic-map'
import {
  shouldRenderProjection,
  shouldRenderTrajectory,
  type MapEvidenceVisibility,
} from '@/lib/map/map-evidence-controls'
import { useTrajectoryFlightReplay } from '@/lib/queries/flight-replay'
import {
  advanceFlightReplayCursor,
  buildFlightReplayFrame,
  type FlightReplaySpeed,
} from '@/lib/replay/flight-replay-model'
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

interface ReplayMapWorkspaceProps {
  aircraft: TrafficAircraft[]
  region: Region
  selectedAircraftICAO24: string | null
  trajectory: AircraftTrajectory | undefined
  projection: ProjectionResult | undefined
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
  const replayIdentity = trajectory
    ? `${trajectory.id}:${trajectory.updated_at}`
    : `selection:${selectedAircraftICAO24 ?? 'none'}`

  return (
    <ReplayMapWorkspace
      key={replayIdentity}
      aircraft={aircraft}
      region={region}
      selectedAircraftICAO24={selectedAircraftICAO24}
      trajectory={trajectoryVisible ? trajectory : undefined}
      projection={projectionVisible ? projection : undefined}
      onSelectAircraft={onSelectAircraft}
    />
  )
}

function ReplayMapWorkspace({
  aircraft,
  region,
  selectedAircraftICAO24,
  trajectory,
  projection,
  onSelectAircraft,
}: ReplayMapWorkspaceProps) {
  const replayQuery = useTrajectoryFlightReplay(trajectory)
  const [replayCursorIndex, setReplayCursorIndex] = useState(0)
  const [replayPlaying, setReplayPlaying] = useState(false)
  const [replaySpeed, setReplaySpeed] = useState<FlightReplaySpeed>(1)
  const replay = replayQuery.data
  const replayFrame = useMemo(
    () => buildFlightReplayFrame(replay, replayCursorIndex),
    [replay, replayCursorIndex]
  )

  useEffect(() => {
    if (!replayPlaying || !replay || replay.points.length <= 1) return

    const intervalID = window.setInterval(() => {
      setReplayCursorIndex(current => {
        const next = advanceFlightReplayCursor(current, replay.points.length)
        if (next.completed) setReplayPlaying(false)
        return next.cursorIndex
      })
    }, 1000 / replaySpeed)

    return () => window.clearInterval(intervalID)
  }, [replay, replayPlaying, replaySpeed])

  const setPlaying = (nextPlaying: boolean) => {
    if (!nextPlaying) {
      setReplayPlaying(false)
      return
    }

    if (!replay || replay.points.length <= 1) {
      setReplayPlaying(false)
      return
    }
    if (replayCursorIndex >= replay.points.length - 1) {
      setReplayCursorIndex(0)
    }
    setReplayPlaying(true)
  }

  const replayUnavailableReason =
    selectedAircraftICAO24 === null
      ? null
      : trajectory && trajectory.flight_id.trim() === ''
        ? 'Replay is unavailable because this trajectory has no durable flight identifier.'
        : null

  return (
    <div>
      <TrafficMap
        aircraft={aircraft}
        region={region}
        selectedAircraftICAO24={selectedAircraftICAO24}
        trajectory={trajectory}
        projection={projection}
        replayPoint={replayFrame.point ?? undefined}
        replayTrail={replayFrame.trailPoints}
        onSelectAircraft={onSelectAircraft}
      />

      {selectedAircraftICAO24 !== null ? (
        <FlightReplayControl
          replay={replay}
          currentPoint={replayFrame.point}
          cursorIndex={replayFrame.cursorIndex}
          isPlaying={replayPlaying}
          speed={replaySpeed}
          isPending={replayQuery.isPending}
          isFetching={replayQuery.isFetching}
          error={replayQuery.error}
          unavailableReason={replayUnavailableReason}
          onCursorIndexChange={setReplayCursorIndex}
          onPlayingChange={setPlaying}
          onSpeedChange={setReplaySpeed}
          onRetry={() => {
            void replayQuery.refetch()
          }}
        />
      ) : null}
    </div>
  )
}
