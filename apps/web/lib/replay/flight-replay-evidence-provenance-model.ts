import type { FlightReplay } from '../../types/flight-replay'

export interface FlightReplayEvidenceSourceSummary {
  sourceName: string
  sampleCount: number
  sampleSharePercent: number
  firstObservedAt: string
  lastObservedAt: string
}

export interface FlightReplayEvidenceSourceTransition {
  fromIndex: number
  toIndex: number
  fromSourceName: string
  toSourceName: string
  startObservedAt: string
  endObservedAt: string
  elapsedSeconds: number | null
}

export interface FlightReplayEvidenceProvenanceProfile {
  sampleCount: number
  identifiedSourceCount: number
  unattributedSampleCount: number
  sources: FlightReplayEvidenceSourceSummary[]
  transitions: FlightReplayEvidenceSourceTransition[]
}

export function buildFlightReplayEvidenceProvenanceProfile(
  replay: FlightReplay | undefined
): FlightReplayEvidenceProvenanceProfile {
  if (!replay || replay.points.length === 0) {
    return {
      sampleCount: 0,
      identifiedSourceCount: 0,
      unattributedSampleCount: 0,
      sources: [],
      transitions: [],
    }
  }

  const sourceSummaries = new Map<
    string,
    {
      sampleCount: number
      firstObservedAt: string
      lastObservedAt: string
    }
  >()
  let unattributedSampleCount = 0

  for (const point of replay.points) {
    const sourceName = normalizeSourceName(point.source_name)
    if (sourceName === null) {
      unattributedSampleCount += 1
      continue
    }

    const existing = sourceSummaries.get(sourceName)
    if (existing) {
      existing.sampleCount += 1
      existing.lastObservedAt = point.observed_at
      continue
    }

    sourceSummaries.set(sourceName, {
      sampleCount: 1,
      firstObservedAt: point.observed_at,
      lastObservedAt: point.observed_at,
    })
  }

  const sources = [...sourceSummaries.entries()].map(
    ([sourceName, summary]): FlightReplayEvidenceSourceSummary => ({
      sourceName,
      sampleCount: summary.sampleCount,
      sampleSharePercent: Math.round(
        (summary.sampleCount / replay.points.length) * 100
      ),
      firstObservedAt: summary.firstObservedAt,
      lastObservedAt: summary.lastObservedAt,
    })
  )

  const transitions: FlightReplayEvidenceSourceTransition[] = []
  for (let toIndex = 1; toIndex < replay.points.length; toIndex++) {
    const fromIndex = toIndex - 1
    const from = replay.points[fromIndex]
    const to = replay.points[toIndex]
    if (!from || !to) continue

    const fromSourceName = normalizeSourceName(from.source_name)
    const toSourceName = normalizeSourceName(to.source_name)
    if (
      fromSourceName === null ||
      toSourceName === null ||
      fromSourceName === toSourceName
    ) {
      continue
    }

    const fromTimestampMS = Date.parse(from.observed_at)
    const toTimestampMS = Date.parse(to.observed_at)
    const elapsedSeconds =
      Number.isNaN(fromTimestampMS) || Number.isNaN(toTimestampMS)
        ? null
        : Math.max(0, (toTimestampMS - fromTimestampMS) / 1000)

    transitions.push({
      fromIndex,
      toIndex,
      fromSourceName,
      toSourceName,
      startObservedAt: from.observed_at,
      endObservedAt: to.observed_at,
      elapsedSeconds,
    })
  }

  return {
    sampleCount: replay.points.length,
    identifiedSourceCount: sources.length,
    unattributedSampleCount,
    sources,
    transitions,
  }
}

function normalizeSourceName(value: string): string | null {
  const normalized = value.trim()
  return normalized.length > 0 ? normalized : null
}
