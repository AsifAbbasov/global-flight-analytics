export interface MapEvidenceVisibility {
  trajectory: boolean
  projection: boolean
}

export const defaultMapEvidenceVisibility: MapEvidenceVisibility = {
  trajectory: true,
  projection: true,
}

export function toggleTrajectoryVisibility(
  visibility: MapEvidenceVisibility
): MapEvidenceVisibility {
  return {
    ...visibility,
    trajectory: !visibility.trajectory,
  }
}

export function toggleProjectionVisibility(
  visibility: MapEvidenceVisibility
): MapEvidenceVisibility {
  return {
    ...visibility,
    projection: !visibility.projection,
  }
}

export function shouldRenderTrajectory(
  visibility: MapEvidenceVisibility,
  featureCount: number
): boolean {
  return visibility.trajectory && Number.isSafeInteger(featureCount) && featureCount > 0
}

export function shouldRenderProjection(
  visibility: MapEvidenceVisibility,
  pointCount: number
): boolean {
  return visibility.projection && Number.isSafeInteger(pointCount) && pointCount > 0
}
