import type { FlightReplay } from '../../types/flight-replay'
import {
  buildFlightReplayAnalyticsSummary,
  buildFlightReplayGaps,
} from './flight-replay-model'

export interface FlightReplayEvidenceQualityProfile {
  sampleCount: number
  intervalCount: number
  observedSpanSeconds: number
  meanGapSeconds: number | null
  medianGapSeconds: number | null
  p90GapSeconds: number | null
  largestGapSeconds: number | null
  largestGapSharePercent: number | null
  observationDensityPerHour: number | null
  altitudeCoveragePercent: number
  largestGapStartObservedAt: string | null
  largestGapEndObservedAt: string | null
}

export function buildFlightReplayEvidenceQualityProfile(
  replay: FlightReplay | undefined
): FlightReplayEvidenceQualityProfile {
  const analytics = buildFlightReplayAnalyticsSummary(replay)
  const gaps = buildFlightReplayGaps(replay)

  if (!replay || replay.points.length === 0) {
    return {
      sampleCount: 0,
      intervalCount: 0,
      observedSpanSeconds: 0,
      meanGapSeconds: null,
      medianGapSeconds: null,
      p90GapSeconds: null,
      largestGapSeconds: null,
      largestGapSharePercent: null,
      observationDensityPerHour: null,
      altitudeCoveragePercent: 0,
      largestGapStartObservedAt: null,
      largestGapEndObservedAt: null,
    }
  }

  if (gaps.length === 0) {
    return {
      sampleCount: analytics.sampleCount,
      intervalCount: 0,
      observedSpanSeconds: analytics.observedSpanSeconds,
      meanGapSeconds: null,
      medianGapSeconds: analytics.medianGapSeconds,
      p90GapSeconds: null,
      largestGapSeconds: null,
      largestGapSharePercent: null,
      observationDensityPerHour: null,
      altitudeCoveragePercent: analytics.altitudeCoveragePercent,
      largestGapStartObservedAt: null,
      largestGapEndObservedAt: null,
    }
  }

  const gapDurations = gaps.map(gap => gap.durationSeconds)
  const sortedGapDurations = [...gapDurations].sort((left, right) => left - right)
  const totalGapSeconds = gapDurations.reduce((sum, duration) => sum + duration, 0)
  const p90Index = Math.max(0, Math.ceil(sortedGapDurations.length * 0.9) - 1)
  const largestGap = gaps.reduce((largest, gap) =>
    gap.durationSeconds > largest.durationSeconds ? gap : largest
  )
  const observedSpanSeconds = analytics.observedSpanSeconds

  return {
    sampleCount: analytics.sampleCount,
    intervalCount: gaps.length,
    observedSpanSeconds,
    meanGapSeconds: totalGapSeconds / gaps.length,
    medianGapSeconds: analytics.medianGapSeconds,
    p90GapSeconds: sortedGapDurations[p90Index] ?? null,
    largestGapSeconds: largestGap.durationSeconds,
    largestGapSharePercent:
      observedSpanSeconds > 0
        ? Math.round((largestGap.durationSeconds / observedSpanSeconds) * 100)
        : null,
    observationDensityPerHour:
      observedSpanSeconds > 0
        ? (analytics.sampleCount * 3600) / observedSpanSeconds
        : null,
    altitudeCoveragePercent: analytics.altitudeCoveragePercent,
    largestGapStartObservedAt: largestGap.startObservedAt,
    largestGapEndObservedAt: largestGap.endObservedAt,
  }
}