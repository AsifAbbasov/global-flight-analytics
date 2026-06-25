export interface RegionBounds {
  min_latitude: number
  max_latitude: number
  min_longitude: number
  max_longitude: number
}

export interface Region {
  code: string
  name: string
  description: string
  bounds: RegionBounds
}
