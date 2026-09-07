import type { FlightReplay } from '../../types/flight-replay'
import type { ProjectionIntelligenceResponse } from '../../types/projection-intelligence'

export const etaEvolutionMaximumSamples = 6

export interface ETAEvolutionInputPoint {
  asOfTime: string
  result?: ProjectionIntelligenceResponse
  error?: string
}

export interface ETAEvolutionPoint {
  asOfTime: string
  available: boolean
  projectionStatus: string | null
  strategy: string | null
  method: string | null
  arrivalStatus: string | null
  airportICAOCode: string | null
  earliestTime: string | null
  estimatedTime: string | null
  latestTime: string | null
  confidenceScore: number | null
  confidenceLevel: string | null
  arrivalWindowSeconds: number | null
  etaChangeSeconds: number | null
  inputFingerprint: string | null
  scopeGuard: string | null
  error: string | null
}

export interface ETAEvolutionSummary {
  sampledObservationCount: number
  availableETACount: number
  unavailableETACount: number
  firstAvailableETA: string | null
  latestAvailableETA: string | null
  netETAChangeSeconds: number | null
  largestAbsoluteAdjacentETAChangeSeconds: number | null
  points: ETAEvolutionPoint[]
}

export function selectETAEvolutionAsOfTimes(
  replay: FlightReplay | undefined,
  maximumSamples = etaEvolutionMaximumSamples
): string[] {
  if (!replay || replay.points.length === 0 || maximumSamples < 1) return []

  const timestamps = [...new Set(
    replay.points
      .map(point => point.observed_at.trim())
      .filter(value => value !== '' && Number.isFinite(Date.parse(value)))
  )].sort((left, right) => Date.parse(left) - Date.parse(right))

  if (timestamps.length <= maximumSamples) return timestamps
  if (maximumSamples === 1) return [timestamps[timestamps.length - 1]]

  const selectedIndexes = new Set<number>()
  for (let position = 0; position < maximumSamples; position += 1) {
    selectedIndexes.add(
      Math.round((position * (timestamps.length - 1)) / (maximumSamples - 1))
    )
  }

  return [...selectedIndexes]
    .sort((left, right) => left - right)
    .map(index => timestamps[index])
}

export function buildETAEvolutionSummary(
  inputs: ETAEvolutionInputPoint[]
): ETAEvolutionSummary {
  const ordered = inputs
    .filter(input => Number.isFinite(Date.parse(input.asOfTime)))
    .slice()
    .sort((left, right) => Date.parse(left.asOfTime) - Date.parse(right.asOfTime))

  const points: ETAEvolutionPoint[] = []
  let previousPoint: ETAEvolutionPoint | null = null

  for (const input of ordered) {
    const result = input.result
    const arrival = result?.projection.arrival
    const estimatedMillis = arrival ? Date.parse(arrival.estimated_time) : Number.NaN
    const earliestMillis = arrival ? Date.parse(arrival.earliest_time) : Number.NaN
    const latestMillis = arrival ? Date.parse(arrival.latest_time) : Number.NaN
    const validArrival =
      arrival !== undefined &&
      Number.isFinite(estimatedMillis) &&
      Number.isFinite(earliestMillis) &&
      Number.isFinite(latestMillis) &&
      latestMillis >= earliestMillis

    const point: ETAEvolutionPoint = {
      asOfTime: input.asOfTime,
      available: validArrival,
      projectionStatus: result?.projection.status ?? null,
      strategy: result?.strategy ?? null,
      method: result?.projection.method.name ?? null,
      arrivalStatus: result?.arrival_status ?? null,
      airportICAOCode: validArrival ? arrival!.airport_icao_code : null,
      earliestTime: validArrival ? arrival!.earliest_time : null,
      estimatedTime: validArrival ? arrival!.estimated_time : null,
      latestTime: validArrival ? arrival!.latest_time : null,
      confidenceScore: validArrival ? arrival!.confidence.score : null,
      confidenceLevel: validArrival ? arrival!.confidence.level : null,
      arrivalWindowSeconds: validArrival
        ? Math.max(0, (latestMillis - earliestMillis) / 1000)
        : null,
      etaChangeSeconds:
        validArrival && previousPoint?.available && previousPoint.estimatedTime !== null
          ? (estimatedMillis - Date.parse(previousPoint.estimatedTime)) / 1000
          : null,
      inputFingerprint: result?.input_fingerprint ?? null,
      scopeGuard: result?.projection.scope_guard ?? null,
      error: input.error?.trim() || null,
    }

    points.push(point)
    previousPoint = point
  }

  const availablePoints = points.filter(
    (point): point is ETAEvolutionPoint & { estimatedTime: string } =>
      point.available && point.estimatedTime !== null
  )
  const adjacentChanges = points
    .map(point => point.etaChangeSeconds)
    .filter((value): value is number => value !== null && Number.isFinite(value))

  const firstAvailable = availablePoints[0] ?? null
  const latestAvailable = availablePoints[availablePoints.length - 1] ?? null
  const netETAChangeSeconds =
    firstAvailable && latestAvailable
      ? (Date.parse(latestAvailable.estimatedTime) - Date.parse(firstAvailable.estimatedTime)) /
        1000
      : null

  return {
    sampledObservationCount: points.length,
    availableETACount: availablePoints.length,
    unavailableETACount: points.length - availablePoints.length,
    firstAvailableETA: firstAvailable?.estimatedTime ?? null,
    latestAvailableETA: latestAvailable?.estimatedTime ?? null,
    netETAChangeSeconds,
    largestAbsoluteAdjacentETAChangeSeconds:
      adjacentChanges.length === 0
        ? null
        : Math.max(...adjacentChanges.map(value => Math.abs(value))),
    points,
  }
}
