import {
  normalizeTrafficAircraftICAO24,
  type TrafficWorkspacePanel,
} from './traffic-workspace-model'

export interface TrafficWorkspaceURLState {
  regionCode: string
  aircraftICAO24: string | null
  panel: TrafficWorkspacePanel
}

const regionParameter = 'region'
const aircraftParameter = 'aircraft'
const panelParameter = 'view'
const reservedParameters = new Set([
  regionParameter,
  aircraftParameter,
  panelParameter,
])

export function parseTrafficWorkspaceSearch(
  search: string,
  availableRegionCodes: readonly string[],
  fallbackRegionCode: string
): TrafficWorkspaceURLState {
  const parameters = new URLSearchParams(stripLeadingQuestionMark(search))
  const regionCode = resolveRegionCode(
    parameters.get(regionParameter),
    availableRegionCodes,
    fallbackRegionCode
  )
  const aircraftICAO24 = normalizeURLAircraftICAO24(
    parameters.get(aircraftParameter)
  )
  const panel = resolveWorkspacePanel(
    parameters.get(panelParameter),
    aircraftICAO24
  )

  return {
    regionCode,
    aircraftICAO24,
    panel,
  }
}

export function buildTrafficWorkspaceSearch(
  currentSearch: string,
  state: TrafficWorkspaceURLState
): string {
  const currentParameters = new URLSearchParams(
    stripLeadingQuestionMark(currentSearch)
  )
  const preservedEntries = Array.from(currentParameters.entries())
    .filter(([key]) => !reservedParameters.has(key))
    .sort(([leftKey, leftValue], [rightKey, rightValue]) => {
      const keyOrder = leftKey.localeCompare(rightKey)
      return keyOrder !== 0 ? keyOrder : leftValue.localeCompare(rightValue)
    })

  const parameters = new URLSearchParams()
  const regionCode = state.regionCode.trim()
  if (regionCode.length > 0) {
    parameters.set(regionParameter, regionCode)
  }

  const aircraftICAO24 = normalizeURLAircraftICAO24(state.aircraftICAO24)
  if (aircraftICAO24 !== null) {
    parameters.set(aircraftParameter, aircraftICAO24)
  }
  parameters.set(panelParameter, state.panel)

  for (const [key, value] of preservedEntries) {
    parameters.append(key, value)
  }

  const serialized = parameters.toString()
  return serialized.length > 0 ? `?${serialized}` : ''
}

export function buildTrafficWorkspaceURL(
  pathname: string,
  currentSearch: string,
  hash: string,
  state: TrafficWorkspaceURLState
): string {
  const normalizedPathname = pathname.trim().length > 0 ? pathname : '/'
  const normalizedHash = hash.length === 0 || hash.startsWith('#')
    ? hash
    : `#${hash}`

  return `${normalizedPathname}${buildTrafficWorkspaceSearch(
    currentSearch,
    state
  )}${normalizedHash}`
}

function resolveRegionCode(
  requestedRegionCode: string | null,
  availableRegionCodes: readonly string[],
  fallbackRegionCode: string
): string {
  const canonicalRegions = new Map<string, string>()
  for (const regionCode of availableRegionCodes) {
    const normalized = normalizeRegionCode(regionCode)
    if (normalized.length > 0 && !canonicalRegions.has(normalized)) {
      canonicalRegions.set(normalized, regionCode)
    }
  }

  const requested = normalizeRegionCode(requestedRegionCode)
  const fallback = normalizeRegionCode(fallbackRegionCode)

  return canonicalRegions.get(requested) ??
    canonicalRegions.get(fallback) ??
    availableRegionCodes[0] ??
    fallbackRegionCode.trim()
}

function resolveWorkspacePanel(
  requestedPanel: string | null,
  aircraftICAO24: string | null
): TrafficWorkspacePanel {
  if (requestedPanel === 'aircraft' || requestedPanel === 'intelligence') {
    return requestedPanel
  }

  return aircraftICAO24 === null ? 'aircraft' : 'intelligence'
}

function normalizeURLAircraftICAO24(
  value: string | null | undefined
): string | null {
  const normalized = normalizeTrafficAircraftICAO24(value)
  if (normalized === null || !/^[0-9a-f]{6}$/.test(normalized)) {
    return null
  }
  return normalized
}

function normalizeRegionCode(value: string | null | undefined): string {
  return value?.trim().toLowerCase() ?? ''
}

function stripLeadingQuestionMark(value: string): string {
  return value.startsWith('?') ? value.slice(1) : value
}
