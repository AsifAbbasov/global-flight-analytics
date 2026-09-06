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
