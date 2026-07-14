import type { RegionBounds } from '@/types/region'

const meanEarthRadiusKilometers = 6371.0088

export function calculateRegionAreaSquareKilometers(
  bounds: RegionBounds
): number | null {
  if (!hasValidBounds(bounds)) {
    return null
  }

  const minimumLatitudeRadians = degreesToRadians(bounds.min_latitude)
  const maximumLatitudeRadians = degreesToRadians(bounds.max_latitude)
  const longitudeWidthRadians = degreesToRadians(
    bounds.max_longitude - bounds.min_longitude
  )

  const area =
    meanEarthRadiusKilometers *
    meanEarthRadiusKilometers *
    Math.abs(
      Math.sin(maximumLatitudeRadians) -
        Math.sin(minimumLatitudeRadians)
    ) *
    longitudeWidthRadians

  return Number.isFinite(area) && area > 0 ? area : null
}

function hasValidBounds(bounds: RegionBounds): boolean {
  const values = [
    bounds.min_latitude,
    bounds.max_latitude,
    bounds.min_longitude,
    bounds.max_longitude,
  ]

  return (
    values.every(Number.isFinite) &&
    bounds.min_latitude >= -90 &&
    bounds.max_latitude <= 90 &&
    bounds.min_longitude >= -180 &&
    bounds.max_longitude <= 180 &&
    bounds.min_latitude < bounds.max_latitude &&
    bounds.min_longitude < bounds.max_longitude
  )
}

function degreesToRadians(value: number): number {
  return (value * Math.PI) / 180
}
