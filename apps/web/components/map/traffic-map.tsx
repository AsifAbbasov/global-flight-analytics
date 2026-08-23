'use client'

import 'maplibre-gl/dist/maplibre-gl.css'

import { useEffect, useRef } from 'react'
import maplibregl from 'maplibre-gl'

import { buildRegionView } from '@/lib/geo/region-view'
import {
  buildProjectionFeatureCollection,
  emptyProjectionFeatureCollection,
  type ProjectionFeatureCollection,
} from '@/lib/geo/projection-geometry'
import { aviationBasemap, aviationMapView } from '@/lib/map/aviation-basemap'
import {
  buildAircraftVisualState,
  normalizeAircraftHeading,
  type AircraftVisualState,
} from '@/lib/map/aircraft-visual'
import { formatTrafficAltitude } from '@/lib/traffic/altitude'
import type { ProjectionResult } from '@/types/projection-intelligence'
import type { Region } from '@/types/region'
import type { TrafficAircraft } from '@/types/traffic'
import type {
  AircraftTrajectory,
  TrajectorySegmentStatus,
} from '@/types/trajectory'

const trajectorySourceID = 'selected-aircraft-trajectory'
const projectionSourceID = 'selected-aircraft-projection'
const projectionLayerIDs = {
  uncertainty: 'selected-aircraft-projection-uncertainty',
  line: 'selected-aircraft-projection-line',
  points: 'selected-aircraft-projection-points',
} as const
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
  projection: ProjectionResult | undefined
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
  projection,
  onSelectAircraft,
}: TrafficMapProps) {
  const mapContainerRef = useRef<HTMLDivElement | null>(null)
  const mapRef = useRef<maplibregl.Map | null>(null)
  const markersRef = useRef<Map<string, AircraftMarkerRecord>>(new Map())
  const onSelectAircraftRef = useRef(onSelectAircraft)

  useEffect(() => {
    onSelectAircraftRef.current = onSelectAircraft
  }, [onSelectAircraft])

  useEffect(() => {
    if (!mapContainerRef.current || mapRef.current) return

    const markers = markersRef.current
    const map = new maplibregl.Map({
      container: mapContainerRef.current,
      style: aviationBasemap.styleURL,
      center: aviationMapView.worldCenter,
      zoom: aviationMapView.worldZoom,
    })

    map.addControl(
      new maplibregl.NavigationControl({ showCompass: true }),
      'top-right'
    )
    map.addControl(new maplibregl.FullscreenControl(), 'top-right')
    mapRef.current = map

    return () => {
      for (const record of markers.values()) record.marker.remove()
      markers.clear()
      map.remove()
      mapRef.current = null
    }
  }, [])

  useEffect(() => {
    const map = mapRef.current
    const view = buildRegionView(region)
    if (!map || !view) return

    const focusSelectedRegion = () => {
      if (view.isWorld) {
        map.easeTo({
          center: aviationMapView.worldCenter,
          zoom: aviationMapView.worldZoom,
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
          padding: { top: 56, right: 56, bottom: 56, left: 56 },
          duration: 900,
          maxZoom: aviationMapView.maxRegionalZoom,
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
    if (!map) return

    const updateTrajectory = () => {
      ensureTrajectoryLayers(map)
      const featureCollection = buildTrajectoryFeatureCollection(trajectory)
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
    if (!map) return

    const updateProjection = () => {
      ensureProjectionLayers(map)
      const featureCollection = buildProjectionFeatureCollection(
        projection?.points
      )
      const source = map.getSource(
        projectionSourceID
      ) as maplibregl.GeoJSONSource | undefined
      source?.setData(featureCollection)
      if (featureCollection.features.length > 0) {
        focusProjection(map, featureCollection)
      }
    }

    if (map.loaded()) {
      updateProjection()
      return
    }

    map.once('load', updateProjection)
    return () => {
      map.off('load', updateProjection)
    }
  }, [projection])

  useEffect(() => {
    const map = mapRef.current
    if (!map) return

    const nextAircraftKeys = new Set<string>()

    for (const item of aircraft) {
      const visualState = buildAircraftVisualState(item, selectedAircraftICAO24)
      if (!visualState) continue

      nextAircraftKeys.add(visualState.key)
      const existingRecord = markersRef.current.get(visualState.key)

      if (existingRecord) {
        updateMarkerRecord(existingRecord, item, visualState)
        continue
      }

      const nextRecord = createMarkerRecord(item, visualState, icao24 => {
        onSelectAircraftRef.current(icao24)
      })
      nextRecord.marker.addTo(map)
      markersRef.current.set(visualState.key, nextRecord)
    }

    for (const [key, record] of markersRef.current.entries()) {
      if (nextAircraftKeys.has(key)) continue
      record.marker.remove()
      markersRef.current.delete(key)
    }
  }, [aircraft, selectedAircraftICAO24])

  const selectedLabel = selectedAircraftICAO24?.trim().toUpperCase() ?? null

  return (
    <div className='relative'>
      <div
        className='h-[min(78vh,860px)] min-h-[560px] w-full overflow-hidden rounded-xl border border-slate-700/70 bg-slate-950 shadow-2xl'
        ref={mapContainerRef}
        aria-label={`Current traffic map focused on ${region.name}`}
        data-region-code={region.code}
        data-basemap-provider={aviationBasemap.provider}
      />
      <div className='pointer-events-none absolute left-3 top-3 flex max-w-[calc(100%-5.5rem)] flex-wrap gap-2'>
        <div className='rounded-full border border-slate-600/70 bg-slate-950/90 px-3 py-1.5 text-[11px] font-semibold uppercase tracking-[0.14em] text-slate-300 shadow-xl backdrop-blur-md'>
          {region.name} · {aviationBasemap.provider}
        </div>
        <div className='rounded-full border border-slate-600/70 bg-slate-950/90 px-3 py-1.5 text-[11px] font-semibold text-slate-300 shadow-xl backdrop-blur-md'>
          {aircraft.length.toLocaleString()} aircraft
        </div>
        {selectedLabel ? (
          <div className='rounded-full border border-amber-300/50 bg-amber-300/15 px-3 py-1.5 font-mono text-[11px] font-bold uppercase tracking-[0.1em] text-amber-100 shadow-xl backdrop-blur-md'>
            Selected · {selectedLabel}
          </div>
        ) : null}
      </div>
      {projection?.points.length ? (
        <div className='pointer-events-none absolute bottom-3 left-3 rounded-lg border border-violet-300/30 bg-slate-950/90 px-3 py-2 text-xs text-slate-200 shadow-xl backdrop-blur-md'>
          <p>
            <span className='mr-2 inline-block h-0.5 w-5 border-t-2 border-dashed border-violet-300 align-middle' />
            Projected path
          </p>
          <p className='mt-1'>
            <span className='mr-2 inline-block h-3 w-5 rounded border border-violet-300/50 bg-violet-400/20 align-middle' />
            Horizontal uncertainty
          </p>
        </div>
      ) : null}
    </div>
  )
}

function ensureProjectionLayers(map: maplibregl.Map) {
  if (!map.getSource(projectionSourceID)) {
    map.addSource(projectionSourceID, {
      type: 'geojson',
      data: emptyProjectionFeatureCollection(),
    })
  }

  if (!map.getLayer(projectionLayerIDs.uncertainty)) {
    map.addLayer({
      id: projectionLayerIDs.uncertainty,
      type: 'fill',
      source: projectionSourceID,
      filter: ['==', ['get', 'kind'], 'uncertainty'],
      paint: {
        'fill-color': '#a78bfa',
        'fill-opacity': [
          'interpolate',
          ['linear'],
          ['get', 'confidence'],
          0,
          0.08,
          1,
          0.22,
        ],
        'fill-outline-color': '#c4b5fd',
      },
    })
  }

  if (!map.getLayer(projectionLayerIDs.line)) {
    map.addLayer({
      id: projectionLayerIDs.line,
      type: 'line',
      source: projectionSourceID,
      filter: ['==', ['get', 'kind'], 'projection-line'],
      layout: { 'line-cap': 'round', 'line-join': 'round' },
      paint: {
        'line-color': '#c4b5fd',
        'line-width': 4,
        'line-opacity': 0.95,
        'line-dasharray': [1.5, 1.5],
      },
    })
  }

  if (!map.getLayer(projectionLayerIDs.points)) {
    map.addLayer({
      id: projectionLayerIDs.points,
      type: 'circle',
      source: projectionSourceID,
      filter: ['==', ['get', 'kind'], 'projection-point'],
      paint: {
        'circle-radius': 5,
        'circle-color': '#ddd6fe',
        'circle-stroke-color': '#6d28d9',
        'circle-stroke-width': 2,
        'circle-opacity': 0.95,
      },
    })
  }
}

function focusProjection(
  map: maplibregl.Map,
  featureCollection: ProjectionFeatureCollection
) {
  const bounds = new maplibregl.LngLatBounds()
  let count = 0

  for (const feature of featureCollection.features) {
    if (feature.geometry.type === 'Point') {
      bounds.extend(feature.geometry.coordinates)
      count++
    }
    if (feature.geometry.type === 'LineString') {
      for (const coordinate of feature.geometry.coordinates) {
        bounds.extend(coordinate)
        count++
      }
    }
  }

  if (count === 0) return
  map.fitBounds(bounds, {
    padding: { top: 72, right: 72, bottom: 72, left: 72 },
    duration: 700,
    maxZoom: aviationMapView.maxEvidenceZoom,
  })
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
  if (map.getLayer(layerID)) return

  map.addLayer({
    id: layerID,
    type: 'line',
    source: trajectorySourceID,
    filter: ['==', ['get', 'status'], status],
    layout: { 'line-cap': 'round', 'line-join': 'round' },
    paint: {
      'line-color': color,
      'line-width': width,
      'line-opacity': opacity,
      ...(dashArray ? { 'line-dasharray': dashArray } : {}),
    },
  })
}

function buildTrajectoryFeatureCollection(
  trajectory: AircraftTrajectory | undefined
): TrajectoryFeatureCollection {
  if (!trajectory) return emptyTrajectoryFeatureCollection()

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
          [segment.start_longitude, segment.start_latitude],
          [segment.end_longitude, segment.end_latitude],
        ],
      },
    }))

  return { type: 'FeatureCollection', features }
}

function emptyTrajectoryFeatureCollection(): TrajectoryFeatureCollection {
  return { type: 'FeatureCollection', features: [] }
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

  if (coordinateCount === 0) return
  map.fitBounds(bounds, {
    padding: { top: 72, right: 72, bottom: 72, left: 72 },
    duration: 700,
    maxZoom: aviationMapView.maxEvidenceZoom,
  })
}

function createMarkerRecord(
  item: TrafficAircraft,
  visualState: AircraftVisualState,
  onSelectAircraft: (icao24: string) => void
): AircraftMarkerRecord {
  const root = document.createElement('button')
  root.type = 'button'
  root.className = 'gfa-aircraft-marker'
  root.addEventListener('click', () => onSelectAircraft(item.icao24))

  const icon = document.createElement('span')
  icon.className = 'gfa-aircraft-marker__icon'
  icon.textContent = '✈'

  const label = document.createElement('span')
  label.className = 'gfa-aircraft-marker__label'
  root.append(icon, label)

  const popup = new maplibregl.Popup({
    closeButton: true,
    closeOnClick: true,
    maxWidth: '280px',
    offset: 28,
  })

  const marker = new maplibregl.Marker({ element: root })
    .setLngLat([visualState.longitude, visualState.latitude])
    .setPopup(popup)

  const record: AircraftMarkerRecord = {
    marker,
    popup,
    root,
    icon,
    label,
  }
  updateMarkerRecord(record, item, visualState)
  return record
}

function updateMarkerRecord(
  record: AircraftMarkerRecord,
  item: TrafficAircraft,
  visualState: AircraftVisualState
) {
  record.root.setAttribute(
    'aria-label',
    `Open aircraft details for ${visualState.label}`
  )
  record.root.setAttribute(
    'aria-pressed',
    visualState.isSelected ? 'true' : 'false'
  )
  record.root.dataset.motionState = visualState.motionState
  record.root.dataset.selected = visualState.isSelected ? 'true' : 'false'
  record.icon.style.transform = `rotate(${visualState.headingDegrees}deg)`
  record.label.textContent = visualState.label
  record.marker.setLngLat([visualState.longitude, visualState.latitude])
  record.popup.setDOMContent(createPopupContent(item))
}

function createPopupContent(item: TrafficAircraft): HTMLElement {
  const container = document.createElement('div')
  container.className = 'gfa-aircraft-popup'

  const title = document.createElement('div')
  title.className = 'gfa-aircraft-popup__title'
  title.textContent = cleanText(item.callsign) || item.icao24.toUpperCase()

  const subtitle = document.createElement('div')
  subtitle.className = 'gfa-aircraft-popup__subtitle'
  subtitle.textContent = `ICAO24 ${item.icao24.toUpperCase()}`

  const telemetry = document.createElement('div')
  telemetry.className = 'gfa-aircraft-popup__telemetry'

  appendOptionalDetail(telemetry, 'Altitude', formatObservedAltitude(item), true)
  appendOptionalDetail(telemetry, 'Speed', formatObservedSpeed(item.velocity_mps), true)
  appendOptionalDetail(telemetry, 'Heading', formatObservedHeading(item.heading_degrees), true)
  appendOptionalDetail(
    telemetry,
    'Status',
    item.on_ground ? 'On ground' : 'Airborne',
    true
  )

  const details = document.createElement('div')
  details.className = 'gfa-aircraft-popup__details'
  appendOptionalDetail(details, 'Airline', cleanText(item.airline))
  appendOptionalDetail(details, 'Aircraft', cleanText(item.aircraft_model))
  appendOptionalDetail(details, 'Country', cleanText(item.origin_country))

  const observedAt = formatObservedAt(item.observed_at)
  const footer = observedAt ? document.createElement('div') : null
  if (footer) {
    footer.className = 'gfa-aircraft-popup__footer'
    footer.textContent = `Observed ${observedAt}`
  }

  container.append(title, subtitle)
  if (telemetry.childElementCount > 0) container.append(telemetry)
  if (details.childElementCount > 0) container.append(details)
  if (footer) container.append(footer)
  return container
}

function appendOptionalDetail(
  container: HTMLElement,
  label: string,
  value: string,
  primary = false
) {
  if (!value) return

  const row = document.createElement('div')
  row.className = primary
    ? 'gfa-aircraft-popup__metric'
    : 'gfa-aircraft-popup__detail'

  const labelElement = document.createElement('span')
  labelElement.className = 'gfa-aircraft-popup__label'
  labelElement.textContent = label

  const valueElement = document.createElement('strong')
  valueElement.className = 'gfa-aircraft-popup__value'
  valueElement.textContent = value

  row.append(labelElement, valueElement)
  container.appendChild(row)
}

function cleanText(value: string): string {
  return value.trim()
}

function formatObservedAltitude(item: TrafficAircraft): string {
  if (item.altitude_status === 'ground') return formatTrafficAltitude(item)
  if (
    item.altitude_status !== 'observed' ||
    item.altitude_m === null ||
    !Number.isFinite(item.altitude_m)
  ) {
    return ''
  }

  return formatTrafficAltitude(item)
}

function formatObservedSpeed(value: number): string {
  if (!Number.isFinite(value) || value < 0) return ''
  return `${new Intl.NumberFormat(undefined, {
    maximumFractionDigits: 0,
  }).format(value * 3.6)} km/h`
}

function formatObservedHeading(value: number): string {
  if (!Number.isFinite(value)) return ''
  return `${normalizeAircraftHeading(value)}°`
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

function formatObservedAt(observedAt: string): string {
  const date = new Date(observedAt)
  if (Number.isNaN(date.getTime())) return ''
  return date.toLocaleString()
}
