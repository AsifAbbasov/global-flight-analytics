import type {
  FlightReplay,
  FlightReplayPoint,
} from '../../types/flight-replay'

export const flightReplaySpeeds = [1, 2, 4] as const
export type FlightReplaySpeed = (typeof flightReplaySpeeds)[number]

export interface FlightReplayFrame {
  cursorIndex: number
  point: FlightReplayPoint | null
  trailPoints: FlightReplayPoint[]
  progress: number
}

export interface FlightReplayAdvance {
  cursorIndex: number
  completed: boolean
}

export interface FlightReplayGap {
  fromIndex: number
  toIndex: number
  startObservedAt: string
  endObservedAt: string
  durationSeconds: number
}

export interface FlightReplayGapSummary {
  gaps: FlightReplayGap[]
  previousGapSeconds: number | null
  nextGapSeconds: number | null
  largestGapSeconds: number
  totalObservedSpanSeconds: number
}

export interface FlightReplayAnalyticsSummary {
  sampleCount: number
  observedSpanSeconds: number
  medianGapSeconds: number | null
  largestGapSeconds: number
  altitudeCoveragePercent: number
  minObservedAltitudeM: number | null
  maxObservedAltitudeM: number | null
  peakVelocityMPS: number | null
  maxClimbRateMPS: number | null
  steepestDescentRateMPS: number | null
  airborneSampleCount: number
  onGroundSampleCount: number
}

export function clampFlightReplayCursorIndex(
  cursorIndex: number,
  pointCount: number
): number {
  if (!Number.isFinite(cursorIndex) || !Number.isFinite(pointCount)) {
    return 0
  }

  const normalizedPointCount = Math.max(0, Math.trunc(pointCount))
  if (normalizedPointCount === 0) return 0

  return Math.min(
    Math.max(0, Math.trunc(cursorIndex)),
    normalizedPointCount - 1
  )
}

export function buildFlightReplayFrame(
  replay: FlightReplay | undefined,
  cursorIndex: number
): FlightReplayFrame {
  if (!replay || replay.points.length === 0) {
    return {
      cursorIndex: 0,
      point: null,
      trailPoints: [],
      progress: 0,
    }
  }

  const currentIndex = clampFlightReplayCursorIndex(
    cursorIndex,
    replay.points.length
  )
  const point = replay.points[currentIndex] ?? null

  return {
    cursorIndex: currentIndex,
    point,
    trailPoints: replay.points.slice(0, currentIndex + 1),
    progress:
      replay.points.length === 1
        ? 1
        : currentIndex / (replay.points.length - 1),
  }
}

export function buildFlightReplayGaps(
  replay: FlightReplay | undefined
): FlightReplayGap[] {
  if (!replay || replay.points.length <= 1) return []

  const gaps: FlightReplayGap[] = []
  for (let toIndex = 1; toIndex < replay.points.length; toIndex++) {
    const fromIndex = toIndex - 1
    const from = replay.points[fromIndex]
    const to = replay.points[toIndex]
    if (!from || !to) continue

    const durationSeconds = Math.max(
      0,
      (Date.parse(to.observed_at) - Date.parse(from.observed_at)) / 1000
    )

    gaps.push({
      fromIndex,
      toIndex,
      startObservedAt: from.observed_at,
      endObservedAt: to.observed_at,
      durationSeconds,
    })
  }

  return gaps
}

export function buildFlightReplayGapSummary(
  replay: FlightReplay | undefined,
  cursorIndex: number
): FlightReplayGapSummary {
  const gaps = buildFlightReplayGaps(replay)
  if (!replay || replay.points.length === 0) {
    return {
      gaps,
      previousGapSeconds: null,
      nextGapSeconds: null,
      largestGapSeconds: 0,
      totalObservedSpanSeconds: 0,
    }
  }

  const currentIndex = clampFlightReplayCursorIndex(
    cursorIndex,
    replay.points.length
  )
  const first = replay.points[0]
  const last = replay.points[replay.points.length - 1]
  const totalObservedSpanSeconds =
    first && last
      ? Math.max(
          0,
          (Date.parse(last.observed_at) - Date.parse(first.observed_at)) / 1000
        )
      : 0

  return {
    gaps,
    previousGapSeconds:
      currentIndex > 0 ? (gaps[currentIndex - 1]?.durationSeconds ?? null) : null,
    nextGapSeconds:
      currentIndex < replay.points.length - 1
        ? (gaps[currentIndex]?.durationSeconds ?? null)
        : null,
    largestGapSeconds: gaps.reduce(
      (largest, gap) => Math.max(largest, gap.durationSeconds),
      0
    ),
    totalObservedSpanSeconds,
  }
}

export function buildFlightReplayAnalyticsSummary(
  replay: FlightReplay | undefined
): FlightReplayAnalyticsSummary {
  if (!replay || replay.points.length === 0) {
    return {
      sampleCount: 0,
      observedSpanSeconds: 0,
      medianGapSeconds: null,
      largestGapSeconds: 0,
      altitudeCoveragePercent: 0,
      minObservedAltitudeM: null,
      maxObservedAltitudeM: null,
      peakVelocityMPS: null,
      maxClimbRateMPS: null,
      steepestDescentRateMPS: null,
      airborneSampleCount: 0,
      onGroundSampleCount: 0,
    }
  }

  const gaps = buildFlightReplayGaps(replay)
  const gapDurations = gaps.map(gap => gap.durationSeconds)
  const altitudes = replay.points
    .map(observedAltitudeMeters)
    .filter((value): value is number => value !== null && Number.isFinite(value))
  const velocities = replay.points
    .map(point => point.velocity_mps)
    .filter(Number.isFinite)
  const climbRates = replay.points
    .map(point => point.vertical_rate_mps)
    .filter(value => Number.isFinite(value) && value > 0)
  const descentRates = replay.points
    .map(point => point.vertical_rate_mps)
    .filter(value => Number.isFinite(value) && value < 0)
  const first = replay.points[0]
  const last = replay.points[replay.points.length - 1]
  const observedSpanSeconds =
    first && last
      ? Math.max(
          0,
          (Date.parse(last.observed_at) - Date.parse(first.observed_at)) / 1000
        )
      : 0

  return {
    sampleCount: replay.points.length,
    observedSpanSeconds,
    medianGapSeconds: median(gapDurations),
    largestGapSeconds: gapDurations.reduce(
      (largest, duration) => Math.max(largest, duration),
      0
    ),
    altitudeCoveragePercent: Math.round(
      (altitudes.length / replay.points.length) * 100
    ),
    minObservedAltitudeM:
      altitudes.length > 0 ? Math.min(...altitudes) : null,
    maxObservedAltitudeM:
      altitudes.length > 0 ? Math.max(...altitudes) : null,
    peakVelocityMPS:
      velocities.length > 0 ? Math.max(...velocities) : null,
    maxClimbRateMPS:
      climbRates.length > 0 ? Math.max(...climbRates) : null,
    steepestDescentRateMPS:
      descentRates.length > 0 ? Math.min(...descentRates) : null,
    airborneSampleCount: replay.points.filter(point => !point.on_ground).length,
    onGroundSampleCount: replay.points.filter(point => point.on_ground).length,
  }
}

export function advanceFlightReplayCursor(
  cursorIndex: number,
  pointCount: number
): FlightReplayAdvance {
  const normalizedPointCount = Math.max(0, Math.trunc(pointCount))
  if (normalizedPointCount === 0) {
    return { cursorIndex: 0, completed: true }
  }

  const currentIndex = clampFlightReplayCursorIndex(
    cursorIndex,
    normalizedPointCount
  )
  if (currentIndex >= normalizedPointCount - 1) {
    return { cursorIndex: currentIndex, completed: true }
  }

  const nextIndex = currentIndex + 1
  return {
    cursorIndex: nextIndex,
    completed: nextIndex >= normalizedPointCount - 1,
  }
}

function observedAltitudeMeters(point: FlightReplayPoint): number | null {
  if (
    point.barometric_altitude_m !== null &&
    (point.barometric_altitude_status === 'observed' ||
      point.barometric_altitude_status === 'ground')
  ) {
    return point.barometric_altitude_m
  }

  if (
    point.geometric_altitude_m !== null &&
    (point.geometric_altitude_status === 'observed' ||
      point.geometric_altitude_status === 'ground')
  ) {
    return point.geometric_altitude_m
  }

  return null
}

function median(values: number[]): number | null {
  if (values.length === 0) return null

  const sorted = [...values].sort((left, right) => left - right)
  const middle = Math.floor(sorted.length / 2)
  if (sorted.length % 2 === 1) {
    return sorted[middle] ?? null
  }

  const lower = sorted[middle - 1]
  const upper = sorted[middle]
  if (lower === undefined || upper === undefined) return null
  return (lower + upper) / 2
}
