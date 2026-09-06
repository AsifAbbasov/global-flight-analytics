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
