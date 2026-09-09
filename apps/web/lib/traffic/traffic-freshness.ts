import type { TrafficAircraft } from '../../types/traffic'

export const defaultRecentObservationWindowMilliseconds = 5 * 60 * 1000
export const defaultFutureClockSkewToleranceMilliseconds = 60 * 1000

export type TrafficPositionFreshnessStatus =
  | 'fresh'
  | 'older'
  | 'future'
  | 'invalid'
  | 'reference-unavailable'

export interface TrafficFreshnessOptions {
  recentObservationWindowMilliseconds?: number
  futureClockSkewToleranceMilliseconds?: number
}

export interface TrafficFreshnessEvidence {
  positionObservedAt: string
  messageObservedAt: string | null
  positionStatus: TrafficPositionFreshnessStatus
  positionAgeMilliseconds: number | null
  messageAgeMilliseconds: number | null
}

export function resolvePositionObservationTimestamp(
  aircraft: TrafficAircraft
): string {
  const explicit = aircraft.position_observed_at?.trim()
  return explicit || aircraft.observed_at.trim()
}

export function buildTrafficFreshnessEvidence(
  aircraft: TrafficAircraft,
  referenceTimeMilliseconds: number,
  options: TrafficFreshnessOptions = {}
): TrafficFreshnessEvidence {
  const positionObservedAt = resolvePositionObservationTimestamp(aircraft)
  const messageObservedAt = cleanOptionalTimestamp(aircraft.message_observed_at)
  const recentWindow = normalizePositiveDuration(
    options.recentObservationWindowMilliseconds,
    defaultRecentObservationWindowMilliseconds
  )
  const futureTolerance = normalizeNonNegativeDuration(
    options.futureClockSkewToleranceMilliseconds,
    defaultFutureClockSkewToleranceMilliseconds
  )
  const referenceTime = normalizeReferenceTime(referenceTimeMilliseconds)
  const positionTime = parseTimestamp(positionObservedAt)
  const messageTime = parseTimestamp(messageObservedAt)

  let positionStatus: TrafficPositionFreshnessStatus = 'invalid'
  let positionAgeMilliseconds: number | null = null

  if (positionTime !== null) {
    if (referenceTime === null) {
      positionStatus = 'reference-unavailable'
    } else {
      positionAgeMilliseconds = referenceTime - positionTime
      if (positionAgeMilliseconds < -futureTolerance) {
        positionStatus = 'future'
      } else if (positionAgeMilliseconds <= recentWindow) {
        positionStatus = 'fresh'
      } else {
        positionStatus = 'older'
      }
    }
  }

  const messageAgeMilliseconds =
    referenceTime !== null && messageTime !== null
      ? referenceTime - messageTime
      : null

  return {
    positionObservedAt,
    messageObservedAt,
    positionStatus,
    positionAgeMilliseconds,
    messageAgeMilliseconds,
  }
}

export function formatTrafficEvidenceAge(
  ageMilliseconds: number | null
): string | null {
  if (ageMilliseconds === null || !Number.isFinite(ageMilliseconds)) {
    return null
  }
  if (ageMilliseconds < 0) {
    const aheadSeconds = Math.max(1, Math.round(Math.abs(ageMilliseconds) / 1000))
    return `${aheadSeconds}s ahead of snapshot`
  }
  const seconds = Math.round(ageMilliseconds / 1000)
  if (seconds < 1) return '<1s old'
  if (seconds < 60) return `${seconds}s old`
  const minutes = Math.round(seconds / 60)
  if (minutes < 60) return `${minutes}m old`
  const hours = Math.round(minutes / 60)
  return `${hours}h old`
}

function cleanOptionalTimestamp(value: string | null | undefined): string | null {
  const normalized = value?.trim() ?? ''
  return normalized.length > 0 ? normalized : null
}

function parseTimestamp(value: string | null): number | null {
  if (!value) return null
  const parsed = Date.parse(value)
  return Number.isFinite(parsed) ? parsed : null
}

function normalizeReferenceTime(value: number): number | null {
  return Number.isFinite(value) && value > 0 ? value : null
}

function normalizePositiveDuration(value: number | undefined, fallback: number): number {
  return value !== undefined && Number.isFinite(value) && value > 0 ? value : fallback
}

function normalizeNonNegativeDuration(value: number | undefined, fallback: number): number {
  return value !== undefined && Number.isFinite(value) && value >= 0 ? value : fallback
}
