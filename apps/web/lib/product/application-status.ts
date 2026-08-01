// FRONTEND_APPLICATION_SHELL_V1
export type InitialSnapshotAvailability =
  | 'ready'
  | 'degraded'
  | 'unavailable'

export interface ApplicationStatusInput {
  initialTrafficCount: number
  regionCount: number
  trafficUnavailable: boolean
  regionsUnavailable: boolean
}

export interface ApplicationInitialStatus {
  availability: InitialSnapshotAvailability
  label: string
  summary: string
  initialTrafficCount: number
  regionCount: number
}

export function buildApplicationInitialStatus(
  input: ApplicationStatusInput
): ApplicationInitialStatus {
  const initialTrafficCount = normalizeCount(input.initialTrafficCount)
  const regionCount = normalizeCount(input.regionCount)

  if (input.trafficUnavailable) {
    return {
      availability: 'unavailable',
      label: 'Initial API snapshot unavailable',
      summary:
        'The page opened without current aircraft data. The live workspace keeps its Retry path available.',
      initialTrafficCount,
      regionCount,
    }
  }

  if (input.regionsUnavailable) {
    return {
      availability: 'degraded',
      label: 'Region catalog degraded',
      summary:
        'The world snapshot is available, but the regional catalog could not be loaded during page startup.',
      initialTrafficCount,
      regionCount,
    }
  }

  return {
    availability: 'ready',
    label: 'Initial snapshot ready',
    summary:
      initialTrafficCount === 0
        ? 'The API returned a valid empty world snapshot. Live refresh remains available in the workspace.'
        : `The server loaded ${initialTrafficCount} aircraft before rendering the research workspace.`,
    initialTrafficCount,
    regionCount,
  }
}

function normalizeCount(value: number): number {
  if (!Number.isFinite(value) || value <= 0) return 0
  return Math.floor(value)
}
