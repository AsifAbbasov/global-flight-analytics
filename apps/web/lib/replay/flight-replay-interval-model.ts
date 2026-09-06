import type { FlightReplay } from '../../types/flight-replay'
import {
  buildFlightReplayGaps,
  buildFlightReplayObservedChange,
  clampFlightReplayCursorIndex,
} from './flight-replay-model'

export interface FlightReplayObservedIntervalComparison {
  startIndex: number
  endIndex: number
  startPointID: string
  endPointID: string
  startObservedAt: string
  endObservedAt: string
  elapsedSeconds: number
  observedSampleCount: number
  intermediateSampleCount: number
  largestGapSeconds: number
  endpointGreatCircleDisplacementM: number
  altitudeDeltaM: number | null
  velocityDeltaMPS: number
  verticalRateDeltaMPS: number
  headingChangeDegrees: number
  startOnGround: boolean
  endOnGround: boolean
}

export function buildFlightReplayObservedIntervalComparison(
  replay: FlightReplay | undefined,
  firstSelectedIndex: number,
  secondSelectedIndex: number
): FlightReplayObservedIntervalComparison | null {
  if (!replay || replay.points.length < 2) return null

  const firstIndex = clampFlightReplayCursorIndex(
    firstSelectedIndex,
    replay.points.length
  )
  const secondIndex = clampFlightReplayCursorIndex(
    secondSelectedIndex,
    replay.points.length
  )
  if (firstIndex === secondIndex) return null

  const startIndex = Math.min(firstIndex, secondIndex)
  const endIndex = Math.max(firstIndex, secondIndex)
  const start = replay.points[startIndex]
  const end = replay.points[endIndex]
  if (!start || !end) return null

  const endpointReplay: FlightReplay = {
    ...replay,
    points: [start, end],
  }
  const endpointChange = buildFlightReplayObservedChange(endpointReplay, 1)
  if (!endpointChange) return null

  const intervalGaps = buildFlightReplayGaps(replay).filter(
    gap => gap.fromIndex >= startIndex && gap.toIndex <= endIndex
  )

  return {
    startIndex,
    endIndex,
    startPointID: start.id,
    endPointID: end.id,
    startObservedAt: start.observed_at,
    endObservedAt: end.observed_at,
    elapsedSeconds: endpointChange.elapsedSeconds,
    observedSampleCount: endIndex - startIndex + 1,
    intermediateSampleCount: Math.max(0, endIndex - startIndex - 1),
    largestGapSeconds: intervalGaps.reduce(
      (largest, gap) => Math.max(largest, gap.durationSeconds),
      0
    ),
    endpointGreatCircleDisplacementM:
      endpointChange.greatCircleDisplacementM,
    altitudeDeltaM: endpointChange.altitudeDeltaM,
    velocityDeltaMPS: endpointChange.velocityDeltaMPS,
    verticalRateDeltaMPS: end.vertical_rate_mps - start.vertical_rate_mps,
    headingChangeDegrees: endpointChange.headingChangeDegrees,
    startOnGround: endpointChange.previousOnGround,
    endOnGround: endpointChange.currentOnGround,
  }
}
