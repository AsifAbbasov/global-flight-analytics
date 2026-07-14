'use client'

import 'maplibre-gl/dist/maplibre-gl.css'

import { useEffect, useRef } from 'react'
import maplibregl from 'maplibre-gl'

import { buildRegionView } from '@/lib/geo/region-view'
import type { Region } from '@/types/region'
import type { TrafficAircraft } from '@/types/traffic'
import type {
  AircraftTrajectory,
  TrajectorySegmentStatus,
} from '@/types/trajectory'

const trajectorySourceID = 'selected-aircraft-trajectory'
const trajectoryLayerIDs = {
  observed: 'selected-aircraft-trajectory-observed',
  interpolated: 'selected-aircraft-trajectory-interpolated',
  estimated: 'selected-aircraft-trajectory-estimated',
  invalid: 'selected-aircraft-trajectory-invalid',
} as const

interface TrafficMapProps {
  aircraft: TrafficAircraft[]
  region: Region
  selectedAircraftICAO24: string | null
  trajectory: AircraftTrajectory | undefined
  onSelectAircraft: (icao24: string) => void
}

interface AircraftMarkerRecord {
  marker: maplibregl.Marker
  popup: maplibregl.Popup
  root: HTMLButtonElement
  icon: HTMLSpanElement
  label: HTMLSpanElement
}

interface TrajectoryLineFeature {
  type: 'Feature'
  properties: {
    status: TrajectorySegmentStatus
    sequence_number: number
    quality_score: number
  }
  geometry: {
    type: 'LineString'
    coordinates: [[number, number], [number, number]]
  }
}

interface TrajectoryFeatureCollection {
  type: 'FeatureCollection'
  features: TrajectoryLineFeature[]
}

export function TrafficMap({
  aircraft,
  region,
  selectedAircraftICAO24,
  trajectory,
  onSelectAircraft,
}: TrafficMapProps) {
  const mapContainerRef = useRef<HTMLDivElement | null>(null)
  const mapRef = useRef<maplibregl.Map | null>(null)
  const markersRef = useRef<Map<string, AircraftMarkerRecord>>(
    new Map()
  )
  const onSelectAircraftRef = useRef(onSelectAircraft)

  useEffect(() => {
    onSelectAircraftRef.current = onSelectAircraft
  }, [onSelectAircraft])

  useEffect(() => {
    if (!mapContainerRef.current || mapRef.current) {
      return
    }

    const markers = markersRef.current

    const map = new maplibregl.Map({
      container: mapContainerRef.current,
      style: 'https://demotiles.maplibre.org/style.json',
      center: [0, 20],
      zoom: 0.8,
    })

    map.addControl(
      new maplibregl.NavigationControl(),
      'top-right'
    )
    mapRef.current = map

    return () => {
      for (const record of markers.values()) {
        record.marker.remove()
      }

      markers.clear()
      map.remove()
      mapRef.current = null
    }
  }, [])

  useEffect(() => {
    const map = mapRef.current
    const view = buildRegionView(region)

    if (!map || !view) {
      return
    }

    const focusSelectedRegion = () => {
      if (view.isWorld) {
        map.easeTo({
          center: [0, 20],
          zoom: 0.8,
          bearing: 0,
          pitch: 0,
          duration: 900,
        })
        return
      }

      map.fitBounds(
        [
          [view.bounds.west, view.bounds.south],
          [view.bounds.east, view.bounds.north],
        ],
        {
          padding: {
            top: 56,
            right: 56,
            bottom: 56,
            left: 56,
          },
          duration: 900,
          maxZoom: 7,
        }
      )
    }

    if (map.loaded()) {
      focusSelectedRegion()
      return
    }

    map.once('load', focusSelectedRegion)

    return () => {
      map.off('load', focusSelectedRegion)
    }
  }, [region])

  useEffect(() => {
    const map = mapRef.current

    if (!map) {
      return
    }

    const updateTrajectory = () => {
      ensureTrajectoryLayers(map)

      const featureCollection =
        buildTrajectoryFeatureCollection(trajectory)
      const source = map.getSource(
        trajectorySourceID
      ) as maplibregl.GeoJSONSource | undefined

      source?.setData(featureCollection)

      if (featureCollection.features.length > 0) {
        focusTrajectory(map, featureCollection)
      }
    }

    if (map.loaded()) {
      updateTrajectory()
      return
    }

    map.once('load', updateTrajectory)

    return () => {
      map.off('load', updateTrajectory)
    }
  }, [trajectory])

  useEffect(() => {
    const map = mapRef.current

    if (!map) {
      return
    }

    const normalizedSelectedICAO24 =
      selectedAircraftICAO24?.trim().toLowerCase() ?? null
    const nextAircraftKeys = new Set<string>()

    for (const item of aircraft) {
      if (!hasValidCoordinates(item)) {
        continue
      }

      const key = item.icao24.trim().toLowerCase()

      if (!key) {
        continue
      }

      nextAircraftKeys.add(key)

      const existingRecord = markersRef.current.get(key)
      const isSelected = key === normalizedSelectedICAO24

      if (existingRecord) {
        updateMarkerRecord(existingRecord, item, isSelected)
        continue
      }

      const nextRecord = createMarkerRecord(
        item,
        isSelected,
        icao24 => {
          onSelectAircraftRef.current(icao24)
        }
      )
      nextRecord.marker.addTo(map)
      markersRef.current.set(key, nextRecord)
    }

    for (const [key, record] of markersRef.current.entries()) {
      if (nextAircraftKeys.has(key)) {
        continue
      }

      record.marker.remove()
      markersRef.current.delete(key)
    }
  }, [aircraft, selectedAircraftICAO24])

  return (
    <div
      className='h-[600px] w-full overflow-hidden rounded-xl'
      ref={mapContainerRef}
      aria-label={`Current traffic map focused on ${region.name}`}
      data-region-code={region.code}
    />
  )
}

function ensureTrajectoryLayers(map: maplibregl.Map) {
  if (!map.getSource(trajectorySourceID)) {
    map.addSource(trajectorySourceID, {
      type: 'geojson',
      data: emptyTrajectoryFeatureCollection(),
    })
  }

  addTrajectoryLayer(
    map,
    trajectoryLayerIDs.observed,
    'observed',
    '#38bdf8',
    undefined,
    4,
    0.95
  )
  addTrajectoryLayer(
    map,
    trajectoryLayerIDs.interpolated,
    'interpolated',
    '#f59e0b',
    [2, 2],
    4,
    0.9
  )
  addTrajectoryLayer(
    map,
    trajectoryLayerIDs.estimated,
    'estimated',
    '#a78bfa',
    [1, 2],
    4,
    0.85
  )
  addTrajectoryLayer(
    map,
    trajectoryLayerIDs.invalid,
    'invalid',
    '#fb7185',
    [1, 1],
    5,
    0.8
  )
}

function addTrajectoryLayer(
  map: maplibregl.Map,
  layerID: string,
  status: TrajectorySegmentStatus,
  color: string,
  dashArray: number[] | undefined,
  width: number,
  opacity: number
) {
  if (map.getLayer(layerID)) {
    return
  }

  map.addLayer({
    id: layerID,
    type: 'line',
    source: trajectorySourceID,
    filter: ['==', ['get', 'status'], status],
    layout: {
      'line-cap': 'round',
      'line-join': 'round',
    },
    paint: {
      'line-color': color,
      'line-width': width,
      'line-opacity': opacity,
      ...(dashArray
        ? {
            'line-dasharray': dashArray,
          }
        : {}),
    },
  })
}

function buildTrajectoryFeatureCollection(
  trajectory: AircraftTrajectory | undefined
): TrajectoryFeatureCollection {
  if (!trajectory) {
    return emptyTrajectoryFeatureCollection()
  }

  const features = [...trajectory.segments]
    .sort(
      (left, right) =>
        left.sequence_number - right.sequence_number
    )
    .filter(segment =>
      hasValidSegmentCoordinates(
        segment.start_latitude,
        segment.start_longitude,
        segment.end_latitude,
        segment.end_longitude
      )
    )
    .map<TrajectoryLineFeature>(segment => ({
      type: 'Feature',
      properties: {
        status: segment.status,
        sequence_number: segment.sequence_number,
        quality_score: segment.quality_score,
      },
      geometry: {
        type: 'LineString',
        coordinates: [
          [
            segment.start_longitude,
            segment.start_latitude,
          ],
          [segment.end_longitude, segment.end_latitude],
        ],
      },
    }))

  return {
    type: 'FeatureCollection',
    features,
  }
}

function emptyTrajectoryFeatureCollection(): TrajectoryFeatureCollection {
  return {
    type: 'FeatureCollection',
    features: [],
  }
}

function focusTrajectory(
  map: maplibregl.Map,
  featureCollection: TrajectoryFeatureCollection
) {
  const bounds = new maplibregl.LngLatBounds()
  let coordinateCount = 0

  for (const feature of featureCollection.features) {
    for (const coordinate of feature.geometry.coordinates) {
      bounds.extend(coordinate)
      coordinateCount++
    }
  }

  if (coordinateCount === 0) {
    return
  }

  map.fitBounds(bounds, {
    padding: {
      top: 72,
      right: 72,
      bottom: 72,
      left: 72,
    },
    duration: 700,
    maxZoom: 9,
  })
}

function createMarkerRecord(
  item: TrafficAircraft,
  isSelected: boolean,
  onSelectAircraft: (icao24: string) => void
): AircraftMarkerRecord {
  const root = document.createElement('button')
  root.type = 'button'
  root.addEventListener('click', () => {
    onSelectAircraft(item.icao24)
  })
  root.setAttribute(
    'aria-label',
    `Open aircraft details for ${displayAircraftName(item)}`
  )

  const icon = document.createElement('span')
  icon.textContent = '✈'
  icon.style.display = 'inline-block'
  icon.style.fontSize = '18px'
  icon.style.lineHeight = '1'

  const label = document.createElement('span')

  root.append(icon, label)

  const popup = new maplibregl.Popup({
    closeButton: true,
    closeOnClick: true,
    maxWidth: '280px',
    offset: 28,
  })

  const marker = new maplibregl.Marker({
    element: root,
  })
    .setLngLat([item.longitude, item.latitude])
    .setPopup(popup)

  const record: AircraftMarkerRecord = {
    marker,
    popup,
    root,
    icon,
    label,
  }

  updateMarkerRecord(record, item, isSelected)

  return record
}

function updateMarkerRecord(
  record: AircraftMarkerRecord,
  item: TrafficAircraft,
  isSelected: boolean
) {
  const name = displayAircraftName(item)

  record.root.setAttribute(
    'aria-label',
    `Open aircraft details for ${name}`
  )
  record.root.setAttribute(
    'aria-pressed',
    isSelected ? 'true' : 'false'
  )
  record.root.className = isSelected
    ? 'flex items-center gap-2 rounded-full border border-amber-300 bg-amber-300 px-3 py-1 text-xs font-semibold text-slate-950 shadow-2xl ring-4 ring-amber-300/25'
    : 'flex items-center gap-2 rounded-full border border-sky-400/40 bg-slate-950/95 px-3 py-1 text-xs font-semibold text-white shadow-xl'
  record.icon.style.color = isSelected ? '#0f172a' : '#38bdf8'
  record.icon.style.transform =
    `rotate(${normalizeHeading(item.heading_degrees)}deg)`
  record.label.textContent = name
  record.marker.setLngLat([item.longitude, item.latitude])
  record.popup.setDOMContent(createPopupContent(item))
}

function createPopupContent(item: TrafficAircraft): HTMLElement {
  const container = document.createElement('div')
  container.style.width = '260px'
  container.style.maxWidth = '260px'
  container.style.padding = '14px'
  container.style.border =
    '1px solid rgba(56, 189, 248, 0.45)'
  container.style.borderRadius = '14px'
  container.style.background = 'rgba(15, 23, 42, 0.98)'
  container.style.color = '#e5e7eb'
  container.style.fontFamily = 'Arial, Helvetica, sans-serif'
  container.style.fontSize = '13px'
  container.style.lineHeight = '1.55'
  container.style.boxShadow =
    '0 18px 45px rgba(0, 0, 0, 0.55)'

  const title = document.createElement('div')
  title.textContent =
    item.callsign.trim() || 'Unknown callsign'
  title.style.fontSize = '16px'
  title.style.fontWeight = '700'
  title.style.color = '#38bdf8'

  const details = document.createElement('div')
  details.style.marginTop = '10px'
  details.style.display = 'grid'
  details.style.gap = '4px'

  appendDetail(details, 'ICAO24', item.icao24)
  appendDetail(details, 'Airline', item.airline || 'Unknown')
  appendDetail(
    details,
    'Aircraft',
    item.aircraft_model || 'Unknown'
  )
  appendDetail(details, 'Altitude', `${item.altitude_m} m`)
  appendDetail(details, 'Speed', `${item.velocity_mps} m/s`)
  appendDetail(
    details,
    'Heading',
    `${normalizeHeading(item.heading_degrees)}°`
  )
  appendDetail(
    details,
    'Status',
    item.on_ground ? 'On ground' : 'In air'
  )
  appendDetail(
    details,
    'Country',
    item.origin_country || 'Unknown'
  )

  const observedAt = document.createElement('div')
  observedAt.textContent = `Observed: ${formatObservedAt(
    item.observed_at
  )}`
  observedAt.style.marginTop = '10px'
  observedAt.style.borderTop =
    '1px solid rgba(148, 163, 184, 0.25)'
  observedAt.style.paddingTop = '8px'
  observedAt.style.color = '#94a3b8'

  container.append(title, details, observedAt)

  return container
}

function appendDetail(
  container: HTMLElement,
  label: string,
  value: string
) {
  const row = document.createElement('div')
  const labelElement = document.createElement('span')

  labelElement.textContent = `${label}: `
  labelElement.style.color = '#94a3b8'

  row.append(labelElement, document.createTextNode(value))
  container.appendChild(row)
}

function displayAircraftName(item: TrafficAircraft): string {
  return item.callsign.trim() || item.icao24
}

function hasValidCoordinates(item: TrafficAircraft): boolean {
  return (
    Number.isFinite(item.latitude) &&
    item.latitude >= -90 &&
    item.latitude <= 90 &&
    Number.isFinite(item.longitude) &&
    item.longitude >= -180 &&
    item.longitude <= 180
  )
}

function hasValidSegmentCoordinates(
  startLatitude: number,
  startLongitude: number,
  endLatitude: number,
  endLongitude: number
): boolean {
  return (
    Number.isFinite(startLatitude) &&
    startLatitude >= -90 &&
    startLatitude <= 90 &&
    Number.isFinite(startLongitude) &&
    startLongitude >= -180 &&
    startLongitude <= 180 &&
    Number.isFinite(endLatitude) &&
    endLatitude >= -90 &&
    endLatitude <= 90 &&
    Number.isFinite(endLongitude) &&
    endLongitude >= -180 &&
    endLongitude <= 180
  )
}

function normalizeHeading(headingDegrees: number): number {
  if (!Number.isFinite(headingDegrees)) {
    return 0
  }

  return ((headingDegrees % 360) + 360) % 360
}

function formatObservedAt(observedAt: string): string {
  const date = new Date(observedAt)

  if (Number.isNaN(date.getTime())) {
    return 'Unknown'
  }

  return date.toLocaleString()
}
