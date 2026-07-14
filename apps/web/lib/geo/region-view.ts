import type { Region } from '@/types/region'

export interface RegionView {
  code: string
  name: string
  center: {
    latitude: number
    longitude: number
  }
  bounds: {
    south: number
    west: number
    north: number
    east: number
  }
  latitudeSpan: number
  longitudeSpan: number
  isWorld: boolean
  globeCameraDistance: number
}

export function buildRegionView(region: Region): RegionView | null {
  const {
    min_latitude: south,
    min_longitude: west,
    max_latitude: north,
    max_longitude: east,
  } = region.bounds

  const values = [south, west, north, east]
  if (!values.every(Number.isFinite)) {
    return null
  }

  if (
    south < -90 ||
    north > 90 ||
    west < -180 ||
    east > 180 ||
    south >= north ||
    west >= east
  ) {
    return null
  }

  const latitudeSpan = north - south
  const longitudeSpan = east - west
  const maximumSpan = Math.max(latitudeSpan, longitudeSpan)
  const isWorld =
    region.code.trim().toLowerCase() === 'world' ||
    (latitudeSpan >= 179 && longitudeSpan >= 359)

  return {
    code: region.code,
    name: region.name,
    center: {
      latitude: south + latitudeSpan / 2,
      longitude: west + longitudeSpan / 2,
    },
    bounds: {
      south,
      west,
      north,
      east,
    },
    latitudeSpan,
    longitudeSpan,
    isWorld,
    globeCameraDistance: resolveGlobeCameraDistance(
      maximumSpan,
      isWorld
    ),
  }
}

function resolveGlobeCameraDistance(
  maximumSpan: number,
  isWorld: boolean
): number {
  if (isWorld || maximumSpan >= 120) {
    return 4
  }

  if (maximumSpan >= 60) {
    return 3.45
  }

  if (maximumSpan >= 30) {
    return 3
  }

  if (maximumSpan >= 15) {
    return 2.65
  }

  return 2.35
}
