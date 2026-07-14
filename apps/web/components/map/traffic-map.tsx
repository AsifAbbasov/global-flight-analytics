'use client'

import 'maplibre-gl/dist/maplibre-gl.css'

import { useEffect, useRef } from 'react'
import maplibregl from 'maplibre-gl'

import { buildRegionView } from '@/lib/geo/region-view'
import type { Region } from '@/types/region'
import type { TrafficAircraft } from '@/types/traffic'

interface TrafficMapProps {
  aircraft: TrafficAircraft[]
  region: Region
}

interface AircraftMarkerRecord {
  marker: maplibregl.Marker
  popup: maplibregl.Popup
  root: HTMLButtonElement
  icon: HTMLSpanElement
  label: HTMLSpanElement
}

export function TrafficMap({
  aircraft,
  region,
}: TrafficMapProps) {
  const mapContainerRef = useRef<HTMLDivElement | null>(null)
  const mapRef = useRef<maplibregl.Map | null>(null)
  const markersRef = useRef<Map<string, AircraftMarkerRecord>>(
    new Map()
  )

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

    const nextAircraftKeys = new Set<string>()

    for (const item of aircraft) {
      if (!hasValidCoordinates(item)) {
        continue
      }

      const key = item.icao24.trim()

      if (!key) {
        continue
      }

      nextAircraftKeys.add(key)

      const existingRecord = markersRef.current.get(key)

      if (existingRecord) {
        updateMarkerRecord(existingRecord, item)
        continue
      }

      const nextRecord = createMarkerRecord(item)
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
  }, [aircraft])

  return (
    <div
      className='h-[600px] w-full overflow-hidden rounded-xl'
      ref={mapContainerRef}
      aria-label={`Current traffic map focused on ${region.name}`}
      data-region-code={region.code}
    />
  )
}

function createMarkerRecord(
  item: TrafficAircraft
): AircraftMarkerRecord {
  const root = document.createElement('button')
  root.type = 'button'
  root.className =
    'flex items-center gap-2 rounded-full border border-sky-400/40 bg-slate-950/95 px-3 py-1 text-xs font-semibold text-white shadow-xl'
  root.setAttribute(
    'aria-label',
    `Open aircraft details for ${displayAircraftName(item)}`
  )

  const icon = document.createElement('span')
  icon.textContent = '✈'
  icon.style.display = 'inline-block'
  icon.style.color = '#38bdf8'
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

  updateMarkerRecord(record, item)

  return record
}

function updateMarkerRecord(
  record: AircraftMarkerRecord,
  item: TrafficAircraft
) {
  const name = displayAircraftName(item)

  record.root.setAttribute(
    'aria-label',
    `Open aircraft details for ${name}`
  )
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
