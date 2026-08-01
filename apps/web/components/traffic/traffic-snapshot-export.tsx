// FRONTEND_RESEARCH_SNAPSHOT_EXPORT_V1
'use client'

import { useState } from 'react'

import { buildTrafficSnapshotExport } from '@/lib/traffic/traffic-snapshot-export'
import type { TrafficAircraft } from '@/types/traffic'

interface TrafficSnapshotExportProps {
  aircraft: TrafficAircraft[]
  regionCode: string
  regionName: string
  snapshotUpdatedAt: number
  selectedAircraftICAO24: string | null
}

type ExportStatus = 'idle' | 'csv' | 'geojson' | 'unavailable'

export function TrafficSnapshotExport({
  aircraft,
  regionCode,
  regionName,
  snapshotUpdatedAt,
  selectedAircraftICAO24,
}: TrafficSnapshotExportProps) {
  const [status, setStatus] = useState<ExportStatus>('idle')
  const exportAvailable = snapshotUpdatedAt > 0

  const download = (format: 'csv' | 'geojson') => {
    if (!exportAvailable || typeof window === 'undefined') {
      setStatus('unavailable')
      return
    }

    const bundle = buildTrafficSnapshotExport(aircraft, {
      regionCode,
      regionName,
      snapshotUpdatedAt,
      generatedAt: new Date().toISOString(),
      selectedAircraftICAO24,
    })
    const file = format === 'csv' ? bundle.csv : bundle.geoJSON

    try {
      downloadTextFile(file.filename, file.mediaType, file.content)
      setStatus(format)
    } catch {
      setStatus('unavailable')
    }
  }

  return (
    <section
      aria-label='Research snapshot export'
      className='mt-4 rounded-xl border border-slate-800 bg-slate-950/55 p-4'
    >
      <div className='flex flex-wrap items-start justify-between gap-4'>
        <div>
          <p className='text-[11px] font-semibold uppercase tracking-[0.16em] text-emerald-300'>
            Research snapshot export
          </p>
          <h3 className='mt-2 text-sm font-semibold text-slate-100'>
            Download the current {regionName} evidence set
          </h3>
          <p className='mt-1 max-w-3xl text-xs leading-5 text-slate-500'>
            CSV preserves the complete typed record set. GeoJSON includes only
            valid point coordinates and reports every excluded record in
            provenance metadata. Exports describe one API snapshot, not a
            historical trend or authoritative flight status.
          </p>
        </div>

        <div className='flex flex-wrap gap-2'>
          <button
            type='button'
            onClick={() => download('csv')}
            disabled={!exportAvailable}
            className='rounded-lg border border-emerald-400/35 bg-emerald-400/5 px-3 py-2 text-sm font-medium text-emerald-100 transition hover:bg-emerald-400/10 disabled:cursor-not-allowed disabled:opacity-50'
          >
            Download CSV
          </button>
          <button
            type='button'
            onClick={() => download('geojson')}
            disabled={!exportAvailable}
            className='rounded-lg border border-sky-400/35 bg-sky-400/5 px-3 py-2 text-sm font-medium text-sky-100 transition hover:bg-sky-400/10 disabled:cursor-not-allowed disabled:opacity-50'
          >
            Download GeoJSON
          </button>
        </div>
      </div>

      <div aria-live='polite' className='mt-3 text-xs text-slate-500'>
        {!exportAvailable
          ? 'Export becomes available after the first successful traffic snapshot.'
          : status === 'csv'
            ? `CSV exported · ${formatInteger(aircraft.length)} records.`
            : status === 'geojson'
              ? 'GeoJSON exported with snapshot provenance metadata.'
              : status === 'unavailable'
                ? 'Browser download is unavailable for this request.'
                : `${formatInteger(aircraft.length)} records ready for reproducible export.`}
      </div>
    </section>
  )
}

function downloadTextFile(
  filename: string,
  mediaType: string,
  content: string
): void {
  if (
    typeof document === 'undefined' ||
    typeof URL === 'undefined' ||
    typeof URL.createObjectURL !== 'function'
  ) {
    throw new Error('browser download APIs are unavailable')
  }

  const blob = new Blob([content], { type: mediaType })
  const objectURL = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = objectURL
  anchor.download = filename
  anchor.hidden = true
  document.body.append(anchor)

  try {
    anchor.click()
  } finally {
    anchor.remove()
    URL.revokeObjectURL(objectURL)
  }
}

function formatInteger(value: number): string {
  return new Intl.NumberFormat().format(value)
}
