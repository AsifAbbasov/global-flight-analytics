import type {
  TrafficAircraft,
  TrafficPositionSource,
} from '../../types/traffic'

export type PositionProvenanceStatus = 'observed' | 'unavailable'

export interface PositionProvenanceEvidence {
  status: PositionProvenanceStatus
  method: TrafficPositionSource | null
  methodLabel: string
  description: string
  sourceName: string | null
}

const presentations: Record<
  TrafficPositionSource,
  { label: string; description: string }
> = {
  adsb: {
    label: 'ADS-B',
    description: 'Position reported by an ADS-B-equipped emitter.',
  },
  adsr: {
    label: 'ADS-R',
    description: 'ADS-B position rebroadcast through ADS-R.',
  },
  tisb: {
    label: 'TIS-B',
    description: 'Traffic-information position rebroadcast through TIS-B.',
  },
  adsc: {
    label: 'ADS-C',
    description: 'Position reported through aircraft communication data.',
  },
  mlat: {
    label: 'MLAT',
    description:
      'Position derived by multilateration; accuracy can vary with receiver geometry.',
  },
  asterix: {
    label: 'ASTERIX',
    description: 'Position supplied through an ASTERIX surveillance feed.',
  },
  flarm: {
    label: 'FLARM',
    description: 'Position supplied through a FLARM-compatible feed.',
  },
}

export function buildPositionProvenanceEvidence(
  aircraft: TrafficAircraft
): PositionProvenanceEvidence {
  const method = normalizePositionSource(aircraft.position_source)
  const sourceName = normalizeSourceName(aircraft.source_name)

  if (method === null) {
    return {
      status: 'unavailable',
      method: null,
      methodLabel: 'Unavailable',
      description:
        'No position-method evidence is available for this observation.',
      sourceName,
    }
  }

  return {
    status: 'observed',
    method,
    methodLabel: presentations[method].label,
    description: presentations[method].description,
    sourceName,
  }
}

function normalizePositionSource(value: unknown): TrafficPositionSource | null {
  return typeof value === 'string' && Object.hasOwn(presentations, value)
    ? (value as TrafficPositionSource)
    : null
}

function normalizeSourceName(value: unknown): string | null {
  if (typeof value !== 'string') return null
  const normalized = value.trim()
  return normalized === '' ? null : normalized
}
