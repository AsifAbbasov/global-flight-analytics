import type { ProjectionPoint } from '@/types/projection-intelligence'

const earthRadiusMeters = 6_371_008.8
const uncertaintyCircleSegments = 32

export interface ProjectionLineFeature {
  type: 'Feature'
  properties: { kind: 'projection-line'; point_count: number }
  geometry: { type: 'LineString'; coordinates: [number, number][] }
}

export interface ProjectionPointFeature {
  type: 'Feature'
  properties: { kind: 'projection-point'; sequence: number; confidence: number }
  geometry: { type: 'Point'; coordinates: [number, number] }
}

export interface ProjectionUncertaintyFeature {
  type: 'Feature'
  properties: { kind: 'uncertainty'; sequence: number; radius_m: number; confidence: number }
  geometry: { type: 'Polygon'; coordinates: [number, number][][] }
}

export type ProjectionMapFeature = ProjectionLineFeature | ProjectionPointFeature | ProjectionUncertaintyFeature
export interface ProjectionFeatureCollection { type: 'FeatureCollection'; features: ProjectionMapFeature[] }

export function buildProjectionFeatureCollection(points: ProjectionPoint[] | undefined): ProjectionFeatureCollection {
  const ordered = [...(points ?? [])]
    .filter(point => validCoordinate(point.position.latitude, point.position.longitude))
    .sort((left, right) => left.sequence - right.sequence)

  if (ordered.length === 0) return emptyProjectionFeatureCollection()

  const features: ProjectionMapFeature[] = ordered.map(point => ({
    type: 'Feature',
    properties: { kind: 'uncertainty', sequence: point.sequence, radius_m: point.uncertainty.horizontal_radius_m, confidence: point.confidence.score },
    geometry: { type: 'Polygon', coordinates: [geodesicCircle(point.position.longitude, point.position.latitude, point.uncertainty.horizontal_radius_m)] },
  }))

  if (ordered.length >= 2) {
    features.push({
      type: 'Feature',
      properties: { kind: 'projection-line', point_count: ordered.length },
      geometry: { type: 'LineString', coordinates: ordered.map(point => [point.position.longitude, point.position.latitude]) },
    })
  }

  features.push(...ordered.map(point => ({
    type: 'Feature' as const,
    properties: { kind: 'projection-point' as const, sequence: point.sequence, confidence: point.confidence.score },
    geometry: { type: 'Point' as const, coordinates: [point.position.longitude, point.position.latitude] as [number, number] },
  })))

  return { type: 'FeatureCollection', features }
}

export function emptyProjectionFeatureCollection(): ProjectionFeatureCollection {
  return { type: 'FeatureCollection', features: [] }
}

function geodesicCircle(longitude:number,latitude:number,radiusMeters:number):[number,number][] {
  const radius = Number.isFinite(radiusMeters) && radiusMeters > 0 ? radiusMeters : 0
  const angularDistance = radius / earthRadiusMeters
  const latitudeRadians = degreesToRadians(latitude)
  const longitudeRadians = degreesToRadians(longitude)
  const coordinates:[number,number][]=[]

  for(let index=0;index<=uncertaintyCircleSegments;index++){
    const bearing=(index/uncertaintyCircleSegments)*Math.PI*2
    const nextLatitude=Math.asin(Math.sin(latitudeRadians)*Math.cos(angularDistance)+Math.cos(latitudeRadians)*Math.sin(angularDistance)*Math.cos(bearing))
    const nextLongitude=longitudeRadians+Math.atan2(Math.sin(bearing)*Math.sin(angularDistance)*Math.cos(latitudeRadians),Math.cos(angularDistance)-Math.sin(latitudeRadians)*Math.sin(nextLatitude))
    coordinates.push([normalizeLongitude(radiansToDegrees(nextLongitude)),radiansToDegrees(nextLatitude)])
  }
  return coordinates
}

function validCoordinate(latitude:number,longitude:number):boolean{return Number.isFinite(latitude)&&latitude>=-90&&latitude<=90&&Number.isFinite(longitude)&&longitude>=-180&&longitude<=180}
function degreesToRadians(value:number):number{return value*Math.PI/180}
function radiansToDegrees(value:number):number{return value*180/Math.PI}
function normalizeLongitude(value:number):number{return ((value+540)%360)-180}
