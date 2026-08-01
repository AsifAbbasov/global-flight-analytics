// FRONTEND_TRAFFIC_DATA_QUALITY_LENS_V1
import type { TrafficAircraft } from '../../types/traffic'

export type TrafficDataQualitySeverity =
  | 'critical'
  | 'warning'
  | 'information'

export type TrafficDataQualityIssueKey =
  | 'invalid-identifier'
  | 'invalid-coordinate'
  | 'invalid-observation-time'
  | 'stale-observation'
  | 'future-observation'
  | 'duplicate-identifier'
  | 'invalid-motion'
  | 'missing-airborne-altitude'
  | 'missing-callsign'
  | 'missing-aircraft-model'
  | 'missing-airline'
  | 'missing-origin-country'

export type TrafficIdentityFieldKey =
  | 'callsign'
  | 'aircraft-model'
  | 'airline'
  | 'origin-country'

export interface TrafficDataQualityIssue {
  key: TrafficDataQualityIssueKey
  label: string
  description: string
  severity: TrafficDataQualitySeverity
  count: number
  denominator: number
  share: number
}

export interface TrafficIdentityCompleteness {
  key: TrafficIdentityFieldKey
  label: string
  presentCount: number
  missingCount: number
  share: number
}

export interface TrafficDataQualityModel {
  totalCount: number
  uniqueAircraftCount: number
  validIdentifierCount: number
  duplicateIdentifierRecordCount: number
  coreUsableCount: number
  coreUsableShare: number
  validCoordinateCount: number
  validCoordinateShare: number
  validMotionCount: number
  validMotionShare: number
  validObservationTimestampCount: number
  validObservationTimestampShare: number
  referenceTimeISO: string | null
  recentObservationCount: number
  recentObservationShare: number | null
  staleObservationCount: number
  futureObservationCount: number
  airborneCount: number
  usableAirborneAltitudeCount: number
  usableAirborneAltitudeShare: number
  identityCompleteness: TrafficIdentityCompleteness[]
  issues: TrafficDataQualityIssue[]
}

export interface TrafficDataQualityOptions {
  recentObservationWindowMilliseconds?: number
  futureClockSkewToleranceMilliseconds?: number
}

const defaultRecentObservationWindowMilliseconds = 5 * 60 * 1000
const defaultFutureClockSkewToleranceMilliseconds = 60 * 1000
const icao24Pattern = /^[0-9a-f]{6}$/

const issueDefinitions: Record<
  TrafficDataQualityIssueKey,
  Pick<TrafficDataQualityIssue, 'label' | 'description' | 'severity'>
> = {
  'invalid-identifier': {
    label: 'Invalid aircraft identifiers',
    description: 'ICAO24 must contain exactly six hexadecimal characters.',
    severity: 'critical',
  },
  'invalid-coordinate': {
    label: 'Invalid map coordinates',
    description: 'Latitude or longitude cannot be represented as a valid map point.',
    severity: 'critical',
  },
  'invalid-observation-time': {
    label: 'Invalid observation timestamps',
    description: 'The provider observation time cannot be parsed.',
    severity: 'critical',
  },
  'stale-observation': {
    label: 'Older observations',
    description: 'Observation time is older than the published five-minute browser window.',
    severity: 'warning',
  },
  'future-observation': {
    label: 'Future-dated observations',
    description: 'Observation time exceeds the allowed one-minute clock-skew tolerance.',
    severity: 'warning',
  },
  'duplicate-identifier': {
    label: 'Duplicate ICAO24 records',
    description: 'Additional records reuse an ICAO24 already present in this snapshot.',
    severity: 'warning',
  },
  'invalid-motion': {
    label: 'Invalid motion values',
    description: 'Velocity must be non-negative and heading must remain within 0–359.999 degrees.',
    severity: 'critical',
  },
  'missing-airborne-altitude': {
    label: 'Airborne altitude unavailable',
    description: 'Airborne records lack a usable observed altitude value.',
    severity: 'warning',
  },
  'missing-callsign': {
    label: 'Missing callsigns',
    description: 'No callsign label is available for these records.',
    severity: 'information',
  },
  'missing-aircraft-model': {
    label: 'Missing aircraft models',
    description: 'No aircraft-model attribution is available for these records.',
    severity: 'information',
  },
  'missing-airline': {
    label: 'Missing airlines',
    description: 'No airline attribution is available for these records.',
    severity: 'information',
  },
  'missing-origin-country': {
    label: 'Missing origin countries',
    description: 'No provider-supplied origin-country label is available.',
    severity: 'information',
  },
}

export function buildTrafficDataQualityModel(
  aircraft: TrafficAircraft[],
  referenceTimeMilliseconds: number,
  options: TrafficDataQualityOptions = {}
): TrafficDataQualityModel {
  const totalCount = aircraft.length
  const recentWindow = normalizePositiveDuration(
    options.recentObservationWindowMilliseconds,
    defaultRecentObservationWindowMilliseconds
  )
  const futureTolerance = normalizeNonNegativeDuration(
    options.futureClockSkewToleranceMilliseconds,
    defaultFutureClockSkewToleranceMilliseconds
  )
  const referenceTime = normalizeReferenceTime(referenceTimeMilliseconds)

  let validIdentifierCount = 0
  let duplicateIdentifierRecordCount = 0
  let validCoordinateCount = 0
  let validMotionCount = 0
  let validObservationTimestampCount = 0
  let recentObservationCount = 0
  let staleObservationCount = 0
  let futureObservationCount = 0
  let airborneCount = 0
  let usableAirborneAltitudeCount = 0
  let coreUsableCount = 0

  const seenIdentifiers = new Set<string>()
  const identityPresence = {
    callsign: 0,
    'aircraft-model': 0,
    airline: 0,
    'origin-country': 0,
  } satisfies Record<TrafficIdentityFieldKey, number>

  for (const item of aircraft) {
    const identifier = normalizeICAO24(item.icao24)
    const identifierValid = identifier !== null
    if (identifierValid) {
      validIdentifierCount++
      if (seenIdentifiers.has(identifier)) {
        duplicateIdentifierRecordCount++
      } else {
        seenIdentifiers.add(identifier)
      }
    }

    const coordinateValid = hasValidCoordinates(item)
    if (coordinateValid) {
      validCoordinateCount++
    }

    const motionValid = hasValidMotion(item)
    if (motionValid) {
      validMotionCount++
    }

    const observationTime = parseTimestamp(item.observed_at)
    const observationTimeValid = observationTime !== null
    if (observationTimeValid) {
      validObservationTimestampCount++
      if (referenceTime !== null) {
        const age = referenceTime - observationTime
        if (age < -futureTolerance) {
          futureObservationCount++
        } else if (age <= recentWindow) {
          recentObservationCount++
        } else {
          staleObservationCount++
        }
      }
    }

    const altitudeUsable = item.on_ground || hasUsableAirborneAltitude(item)
    if (!item.on_ground) {
      airborneCount++
      if (altitudeUsable) {
        usableAirborneAltitudeCount++
      }
    }

    if (
      identifierValid &&
      coordinateValid &&
      observationTimeValid &&
      motionValid &&
      altitudeUsable
    ) {
      coreUsableCount++
    }

    if (hasText(item.callsign)) {
      identityPresence.callsign++
    }
    if (hasText(item.aircraft_model)) {
      identityPresence['aircraft-model']++
    }
    if (hasText(item.airline)) {
      identityPresence.airline++
    }
    if (hasText(item.origin_country)) {
      identityPresence['origin-country']++
    }
  }

  const identityCompleteness = buildIdentityCompleteness(
    identityPresence,
    totalCount
  )
  const issueInputs: Array<{
    key: TrafficDataQualityIssueKey
    count: number
    denominator: number
  }> = [
    {
      key: 'invalid-identifier',
      count: totalCount - validIdentifierCount,
      denominator: totalCount,
    },
    {
      key: 'invalid-coordinate',
      count: totalCount - validCoordinateCount,
      denominator: totalCount,
    },
    {
      key: 'invalid-observation-time',
      count: totalCount - validObservationTimestampCount,
      denominator: totalCount,
    },
    {
      key: 'stale-observation',
      count: staleObservationCount,
      denominator: validObservationTimestampCount,
    },
    {
      key: 'future-observation',
      count: futureObservationCount,
      denominator: validObservationTimestampCount,
    },
    {
      key: 'duplicate-identifier',
      count: duplicateIdentifierRecordCount,
      denominator: validIdentifierCount,
    },
    {
      key: 'invalid-motion',
      count: totalCount - validMotionCount,
      denominator: totalCount,
    },
    {
      key: 'missing-airborne-altitude',
      count: airborneCount - usableAirborneAltitudeCount,
      denominator: airborneCount,
    },
    {
      key: 'missing-callsign',
      count: totalCount - identityPresence.callsign,
      denominator: totalCount,
    },
    {
      key: 'missing-aircraft-model',
      count: totalCount - identityPresence['aircraft-model'],
      denominator: totalCount,
    },
    {
      key: 'missing-airline',
      count: totalCount - identityPresence.airline,
      denominator: totalCount,
    },
    {
      key: 'missing-origin-country',
      count: totalCount - identityPresence['origin-country'],
      denominator: totalCount,
    },
  ]

  return {
    totalCount,
    uniqueAircraftCount: seenIdentifiers.size,
    validIdentifierCount,
    duplicateIdentifierRecordCount,
    coreUsableCount,
    coreUsableShare: ratio(coreUsableCount, totalCount),
    validCoordinateCount,
    validCoordinateShare: ratio(validCoordinateCount, totalCount),
    validMotionCount,
    validMotionShare: ratio(validMotionCount, totalCount),
    validObservationTimestampCount,
    validObservationTimestampShare: ratio(
      validObservationTimestampCount,
      totalCount
    ),
    referenceTimeISO:
      referenceTime === null ? null : new Date(referenceTime).toISOString(),
    recentObservationCount,
    recentObservationShare:
      referenceTime === null
        ? null
        : ratio(recentObservationCount, validObservationTimestampCount),
    staleObservationCount,
    futureObservationCount,
    airborneCount,
    usableAirborneAltitudeCount,
    usableAirborneAltitudeShare: ratio(
      usableAirborneAltitudeCount,
      airborneCount
    ),
    identityCompleteness,
    issues: issueInputs
      .filter(issue => issue.count > 0)
      .map(issue => ({
        ...issueDefinitions[issue.key],
        ...issue,
        share: ratio(issue.count, issue.denominator),
      }))
      .sort(compareIssues),
  }
}

function buildIdentityCompleteness(
  presence: Record<TrafficIdentityFieldKey, number>,
  totalCount: number
): TrafficIdentityCompleteness[] {
  const definitions: ReadonlyArray<{
    key: TrafficIdentityFieldKey
    label: string
  }> = [
    { key: 'callsign', label: 'Callsign' },
    { key: 'aircraft-model', label: 'Aircraft model' },
    { key: 'airline', label: 'Airline' },
    { key: 'origin-country', label: 'Origin country' },
  ]

  return definitions.map(definition => ({
    ...definition,
    presentCount: presence[definition.key],
    missingCount: totalCount - presence[definition.key],
    share: ratio(presence[definition.key], totalCount),
  }))
}

function compareIssues(
  left: TrafficDataQualityIssue,
  right: TrafficDataQualityIssue
): number {
  const severityOrder = severityWeight(left.severity) - severityWeight(right.severity)
  if (severityOrder !== 0) {
    return severityOrder
  }

  const countOrder = right.count - left.count
  if (countOrder !== 0) {
    return countOrder
  }

  return left.key.localeCompare(right.key)
}

function severityWeight(severity: TrafficDataQualitySeverity): number {
  switch (severity) {
    case 'critical':
      return 0
    case 'warning':
      return 1
    case 'information':
      return 2
  }
}

function normalizeICAO24(value: string): string | null {
  const normalized = value.trim().toLowerCase()
  return icao24Pattern.test(normalized) ? normalized : null
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

function hasValidMotion(item: TrafficAircraft): boolean {
  return (
    Number.isFinite(item.velocity_mps) &&
    item.velocity_mps >= 0 &&
    Number.isFinite(item.heading_degrees) &&
    item.heading_degrees >= 0 &&
    item.heading_degrees < 360
  )
}

function hasUsableAirborneAltitude(item: TrafficAircraft): boolean {
  return (
    item.altitude_status === 'observed' &&
    item.altitude_m !== null &&
    Number.isFinite(item.altitude_m) &&
    item.altitude_m >= 0
  )
}

function parseTimestamp(value: string): number | null {
  const parsed = Date.parse(value)
  return Number.isFinite(parsed) ? parsed : null
}

function normalizeReferenceTime(value: number): number | null {
  return Number.isFinite(value) && value > 0 ? value : null
}

function normalizePositiveDuration(
  value: number | undefined,
  fallback: number
): number {
  return value !== undefined && Number.isFinite(value) && value > 0
    ? value
    : fallback
}

function normalizeNonNegativeDuration(
  value: number | undefined,
  fallback: number
): number {
  return value !== undefined && Number.isFinite(value) && value >= 0
    ? value
    : fallback
}

function hasText(value: string): boolean {
  return value.trim().length > 0
}

function ratio(numerator: number, denominator: number): number {
  return denominator > 0 ? numerator / denominator : 0
}
