'use client'

import { useEffect, useMemo, useState } from 'react'

import { FlightReplayControl } from '@/components/aircraft/flight-replay-control'
import { FlightReplayTimeNavigation } from '@/components/aircraft/flight-replay-time-navigation'
import { TrafficMap } from '@/components/map/traffic-map'
import {
  shouldRenderProjection,
  shouldRenderTrajectory,
  type MapEvidenceVisibility,
} from '@/lib/map/map-evidence-controls'
import { useTrajectoryFlightReplay } from '@/lib/queries/flight-replay'
import {
  advanceFlightReplayTimeCursor,
  buildFlightReplayTimeFrame,
  flightReplayObservationCursorSeconds,
  resolveFlightReplayTimeCursorFromSearch,
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
  mapTrajectory: AircraftTrajectory | undefined
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
      trajectory={trajectory}
      mapTrajectory={trajectoryVisible ? trajectory : undefined}
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
  mapTrajectory,
  projection,
  onSelectAircraft,
}: ReplayMapWorkspaceProps) {
  const replayQuery = useTrajectoryFlightReplay(trajectory)
  const [replayCursorSeconds, setReplayCursorSeconds] = useState<number | null>(null)
  const [replayPlaying, setReplayPlaying] = useState(false)
  const [replaySpeed, setReplaySpeed] = useState<FlightReplaySpeed>(1)
  const replay = replayQuery.data
  const resolvedReplayCursorSeconds = useMemo(() => {
    if (replayCursorSeconds !== null) return replayCursorSeconds
    if (typeof window === 'undefined') return 0
    return resolveFlightReplayTimeCursorFromSearch(replay, window.location.search) ?? 0
  }, [replay, replayCursorSeconds])
  const replayFrame = useMemo(
    () => buildFlightReplayTimeFrame(replay, resolvedReplayCursorSeconds),
    [replay, resolvedReplayCursorSeconds]
  )

  useEffect(() => {
    if (
      !replayPlaying ||
      !replay ||
      replay.points.length <= 1 ||
      replayFrame.totalObservedSpanSeconds <= 0
    ) {
      return
    }

    const tickSeconds = 0.25
    const intervalID = window.setInterval(() => {
      setReplayCursorSeconds(current => {
        const next = advanceFlightReplayTimeCursor(
          current ?? replayFrame.cursorSeconds,
          replayFrame.totalObservedSpanSeconds,
          tickSeconds * replaySpeed
        )
        if (next.completed) setReplayPlaying(false)
        return next.cursorSeconds
      })
    }, tickSeconds * 1000)

    return () => window.clearInterval(intervalID)
  }, [
    replay,
    replayFrame.cursorSeconds,
    replayFrame.totalObservedSpanSeconds,
    replayPlaying,
    replaySpeed,
  ])

  const setPlaying = (nextPlaying: boolean) => {
    if (!nextPlaying) {
      setReplayPlaying(false)
      return
    }

    if (
      !replay ||
      replay.points.length <= 1 ||
      replayFrame.totalObservedSpanSeconds <= 0
    ) {
      setReplayPlaying(false)
      return
    }
    if (replayFrame.cursorSeconds >= replayFrame.totalObservedSpanSeconds) {
      setReplayCursorSeconds(0)
    }
    setReplayPlaying(true)
  }

  const setObservationCursorIndex = (cursorIndex: number) => {
    setReplayCursorSeconds(
      flightReplayObservationCursorSeconds(replay, cursorIndex)
    )
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
        trajectory={mapTrajectory}
        projection={projection}
        replayPoint={replayFrame.point ?? undefined}
        replayTrail={replayFrame.trailPoints}
        onSelectAircraft={onSelectAircraft}
      />

      {selectedAircraftICAO24 !== null ? (
        <>
          <FlightReplayTimeNavigation
            replay={replay}
            frame={replayFrame}
            isPlaying={replayPlaying}
            onCursorSecondsChange={setReplayCursorSeconds}
            onPlayingChange={setPlaying}
          />
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
            onCursorIndexChange={setObservationCursorIndex}
            onPlayingChange={setPlaying}
            onSpeedChange={setReplaySpeed}
            onRetry={() => {
              void replayQuery.refetch()
            }}
          />
        </>
      ) : null}
    </div>
  )
}
