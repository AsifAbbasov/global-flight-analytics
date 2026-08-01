// FRONTEND_RESEARCH_SNAPSHOT_EXPORT_V1
import type { TrafficAircraft } from '../../types/traffic'

export interface TrafficSnapshotExportContext {
  regionCode: string
  regionName: string
  snapshotUpdatedAt: number
  generatedAt: string
  selectedAircraftICAO24?: string | null
}

export interface TrafficSnapshotExportFile {
  filename: string
  mediaType: string
  content: string
}

export interface TrafficSnapshotExportBundle {
  csv: TrafficSnapshotExportFile
  geoJSON: TrafficSnapshotExportFile
  totalAircraftCount: number
  geoJSONFeatureCount: number
  excludedInvalidCoordinateCount: number
}

interface GeoJSONFeature {
  type: 'Feature'
  id: string
  geometry: {
    type: 'Point'
    coordinates: [number, number]
  }
  properties: Record<string, string | number | boolean | null>
}

const exportSchemaVersion = '1.0'
const evidenceBoundary =
  'Current API snapshot only; not historical trend, operational guidance, or authoritative flight status.'

const csvColumns = [
  'schema_version',
  'region_code',
  'region_name',
  'snapshot_updated_at',
  'export_generated_at',
  'selected_aircraft_icao24',
  'icao24',
  'callsign',
  'latitude',
  'longitude',
  'altitude_m',
  'altitude_status',
  'altitude_source',
  'velocity_mps',
  'heading_degrees',
  'on_ground',
  'observed_at',
  'aircraft_model',
  'airline',
  'origin_country',
] as const

export function buildTrafficSnapshotExport(
  aircraft: TrafficAircraft[],
  context: TrafficSnapshotExportContext
): TrafficSnapshotExportBundle {
  const normalizedContext = normalizeContext(context)
  const orderedAircraft = [...aircraft].sort(compareAircraft)
  const features = orderedAircraft
    .filter(hasValidCoordinates)
    .map(toGeoJSONFeature)
  const excludedInvalidCoordinateCount =
    orderedAircraft.length - features.length
  const filenameStem = buildFilenameStem(normalizedContext)

  return {
    csv: {
      filename: `${filenameStem}.csv`,
      mediaType: 'text/csv;charset=utf-8',
      content: buildCSV(orderedAircraft, normalizedContext),
    },
    geoJSON: {
      filename: `${filenameStem}.geojson`,
      mediaType: 'application/geo+json;charset=utf-8',
      content: JSON.stringify(
        {
          type: 'FeatureCollection',
          metadata: {
            schema_version: exportSchemaVersion,
            region_code: normalizedContext.regionCode,
            region_name: normalizedContext.regionName,
            snapshot_updated_at: normalizedContext.snapshotUpdatedAtISO,
            export_generated_at: normalizedContext.generatedAt,
            selected_aircraft_icao24:
              normalizedContext.selectedAircraftICAO24,
            aircraft_count: orderedAircraft.length,
            feature_count: features.length,
            excluded_invalid_coordinate_count:
              excludedInvalidCoordinateCount,
            evidence_boundary: evidenceBoundary,
          },
          features,
        },
        null,
        2
      ),
    },
    totalAircraftCount: orderedAircraft.length,
    geoJSONFeatureCount: features.length,
    excludedInvalidCoordinateCount,
  }
}

interface NormalizedContext {
  regionCode: string
  regionName: string
  snapshotUpdatedAtISO: string | null
  generatedAt: string
  selectedAircraftICAO24: string | null
}

function normalizeContext(
  context: TrafficSnapshotExportContext
): NormalizedContext {
  const generatedAt = normalizeISODate(context.generatedAt)
  const snapshotUpdatedAtISO =
    Number.isFinite(context.snapshotUpdatedAt) && context.snapshotUpdatedAt > 0
      ? new Date(context.snapshotUpdatedAt).toISOString()
      : null

  return {
    regionCode: normalizeText(context.regionCode),
    regionName: normalizeText(context.regionName),
    snapshotUpdatedAtISO,
    generatedAt,
    selectedAircraftICAO24: normalizeOptionalICAO24(
      context.selectedAircraftICAO24
    ),
  }
}

function buildCSV(
  aircraft: TrafficAircraft[],
  context: NormalizedContext
): string {
  const rows = [csvColumns.join(',')]

  for (const item of aircraft) {
    const row: Record<(typeof csvColumns)[number], unknown> = {
      schema_version: exportSchemaVersion,
      region_code: context.regionCode,
      region_name: context.regionName,
      snapshot_updated_at: context.snapshotUpdatedAtISO,
      export_generated_at: context.generatedAt,
      selected_aircraft_icao24: context.selectedAircraftICAO24,
      icao24: normalizeICAO24(item.icao24),
      callsign: item.callsign,
      latitude: finiteNumberOrNull(item.latitude),
      longitude: finiteNumberOrNull(item.longitude),
      altitude_m: finiteNumberOrNull(item.altitude_m),
      altitude_status: item.altitude_status,
      altitude_source: item.altitude_source,
      velocity_mps: finiteNumberOrNull(item.velocity_mps),
      heading_degrees: finiteNumberOrNull(item.heading_degrees),
      on_ground: item.on_ground,
      observed_at: item.observed_at,
      aircraft_model: item.aircraft_model,
      airline: item.airline,
      origin_country: item.origin_country,
    }

    rows.push(
      csvColumns.map(column => escapeCSVValue(row[column])).join(',')
    )
  }

  return `${rows.join('\n')}\n`
}

function toGeoJSONFeature(item: TrafficAircraft): GeoJSONFeature {
  return {
    type: 'Feature',
    id: normalizeICAO24(item.icao24),
    geometry: {
      type: 'Point',
      coordinates: [item.longitude, item.latitude],
    },
    properties: {
      icao24: normalizeICAO24(item.icao24),
      callsign: item.callsign,
      altitude_m: finiteNumberOrNull(item.altitude_m),
      altitude_status: item.altitude_status,
      altitude_source: item.altitude_source,
      velocity_mps: finiteNumberOrNull(item.velocity_mps),
      heading_degrees: finiteNumberOrNull(item.heading_degrees),
      on_ground: item.on_ground,
      observed_at: item.observed_at,
      aircraft_model: item.aircraft_model,
      airline: item.airline,
      origin_country: item.origin_country,
    },
  }
}

function hasValidCoordinates(item: TrafficAircraft): boolean {
  return (
    Number.isFinite(item.latitude) &&
    Number.isFinite(item.longitude) &&
    item.latitude >= -90 &&
    item.latitude <= 90 &&
    item.longitude >= -180 &&
    item.longitude <= 180
  )
}

function compareAircraft(
  left: TrafficAircraft,
  right: TrafficAircraft
): number {
  const icaoOrder = normalizeICAO24(left.icao24).localeCompare(
    normalizeICAO24(right.icao24)
  )
  if (icaoOrder !== 0) {
    return icaoOrder
  }

  const observationOrder = left.observed_at.localeCompare(right.observed_at)
  if (observationOrder !== 0) {
    return observationOrder
  }

  return left.callsign.localeCompare(right.callsign)
}

function buildFilenameStem(context: NormalizedContext): string {
  const regionSlug = slugify(context.regionCode || context.regionName) || 'region'
  const timestamp = compactTimestamp(
    context.snapshotUpdatedAtISO ?? context.generatedAt
  )
  return `global-flight-analytics-${regionSlug}-${timestamp}`
}

function slugify(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

function compactTimestamp(value: string): string {
  return normalizeISODate(value)
    .replace(/[-:]/g, '')
    .replace(/\.\d{3}Z$/, 'Z')
}

function normalizeISODate(value: string): string {
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) {
    throw new Error('export generatedAt must be a valid ISO-compatible date')
  }
  return parsed.toISOString()
}

function normalizeICAO24(value: string): string {
  return value.trim().toLowerCase()
}

function normalizeOptionalICAO24(
  value: string | null | undefined
): string | null {
  const normalized = value?.trim().toLowerCase() ?? ''
  return normalized.length > 0 ? normalized : null
}

function normalizeText(value: string): string {
  return value.trim()
}

function finiteNumberOrNull(
  value: number | null
): number | null {
  return value !== null && Number.isFinite(value) ? value : null
}

function escapeCSVValue(value: unknown): string {
  if (value === null || value === undefined) {
    return ''
  }

  const text = String(value)
  if (!/[",\r\n]/.test(text)) {
    return text
  }

  return `"${text.replace(/"/g, '""')}"`
}
