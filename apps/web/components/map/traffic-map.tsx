'use client'

import 'maplibre-gl/dist/maplibre-gl.css'

import { useEffect, useRef } from 'react'
import maplibregl from 'maplibre-gl'

import type { TrafficAircraft } from '@/types/traffic'

interface TrafficMapProps {
  aircraft: TrafficAircraft[]
}

function createAircraftMarker(item: TrafficAircraft): HTMLElement {
  const marker = document.createElement('div')

  marker.className =
    'flex items-center gap-2 rounded-full border border-sky-400/40 bg-slate-950/95 px-3 py-1 text-xs font-semibold text-white shadow-xl'

  marker.innerHTML = `
    <span
      style="
        display: inline-block;
        transform: rotate(${item.heading_degrees}deg);
        color: #38bdf8;
        font-size: 18px;
        line-height: 1;
      "
    >
      ✈
    </span>
    <span>${item.callsign || item.icao24}</span>
  `

  return marker
}

function createPopupContent(item: TrafficAircraft): string {
  return `
    <div
      style="
        width: 260px;
        max-width: 260px;
        padding: 14px;
        border: 1px solid rgba(56, 189, 248, 0.45);
        border-radius: 14px;
        background: rgba(15, 23, 42, 0.98);
        color: #e5e7eb;
        font-family: Arial, Helvetica, sans-serif;
        font-size: 13px;
        line-height: 1.55;
        box-shadow: 0 18px 45px rgba(0, 0, 0, 0.55);
      "
    >
      <div style="font-size: 16px; font-weight: 700; color: #38bdf8;">
        ${item.callsign || 'Unknown callsign'}
      </div>

      <div style="margin-top: 10px; display: grid; gap: 4px;">
        <div><span style="color: #94a3b8;">ICAO24:</span> ${item.icao24}</div>
        <div><span style="color: #94a3b8;">Airline:</span> ${
          item.airline || 'Unknown'
        }</div>
        <div><span style="color: #94a3b8;">Aircraft:</span> ${
          item.aircraft_model || 'Unknown'
        }</div>
        <div><span style="color: #94a3b8;">Altitude:</span> ${
          item.altitude_m
        } m</div>
        <div><span style="color: #94a3b8;">Speed:</span> ${
          item.velocity_mps
        } m/s</div>
        <div><span style="color: #94a3b8;">Heading:</span> ${
          item.heading_degrees
        }°</div>
        <div><span style="color: #94a3b8;">Status:</span> ${
          item.on_ground ? 'On ground' : 'In air'
        }</div>
        <div><span style="color: #94a3b8;">Country:</span> ${
          item.origin_country || 'Unknown'
        }</div>
      </div>

      <div style="margin-top: 10px; border-top: 1px solid rgba(148, 163, 184, 0.25); padding-top: 8px; color: #94a3b8;">
        Observed: ${new Date(item.observed_at).toLocaleString()}
      </div>
    </div>
  `
}

export function TrafficMap({ aircraft }: TrafficMapProps) {
  const mapContainerRef = useRef<HTMLDivElement | null>(null)
  const mapRef = useRef<maplibregl.Map | null>(null)

  useEffect(() => {
    if (!mapContainerRef.current || mapRef.current) {
      return
    }

    mapRef.current = new maplibregl.Map({
      container: mapContainerRef.current,
      style: 'https://demotiles.maplibre.org/style.json',
      center: [50.0467, 40.4675],
      zoom: 6,
    })

    aircraft.forEach(item => {
      new maplibregl.Marker({
        element: createAircraftMarker(item),
      })
        .setLngLat([item.longitude, item.latitude])
        .setPopup(
          new maplibregl.Popup({
            closeButton: true,
            closeOnClick: true,
            maxWidth: '280px',
            offset: 28,
          }).setHTML(createPopupContent(item))
        )
        .addTo(mapRef.current!)
    })

    return () => {
      mapRef.current?.remove()
      mapRef.current = null
    }
  }, [aircraft])

  return (
    <div
      className='h-[600px] w-full overflow-hidden rounded-xl'
      ref={mapContainerRef}
    />
  )
}
