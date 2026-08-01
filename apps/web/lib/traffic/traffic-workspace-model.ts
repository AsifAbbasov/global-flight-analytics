export type TrafficWorkspacePanel = 'aircraft' | 'intelligence'

export interface TrafficWorkspaceSelection {
  icao24: string | null
  panel: TrafficWorkspacePanel
}

export function normalizeTrafficAircraftICAO24(
  value: string | null | undefined
): string | null {
  const normalized = value?.trim().toLowerCase() ?? ''
  return normalized.length > 0 ? normalized : null
}

export function buildTrafficWorkspaceSelection(
  value: string | null | undefined
): TrafficWorkspaceSelection {
  const icao24 = normalizeTrafficAircraftICAO24(value)

  return {
    icao24,
    panel: icao24 === null ? 'aircraft' : 'intelligence',
  }
}
