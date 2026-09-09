import { APIRequestError, requestAPIData } from '@/lib/api/client'
import type {
  ETAReliabilityMetrics,
  ETAReliabilityRequest,
  ETAReliabilityResponse,
  ETAReliabilityStatus,
} from '@/types/eta-reliability'

const statuses = new Set<ETAReliabilityStatus>(['unavailable', 'limited', 'complete'])
const evidenceClass = 'historically_recomputed_from_persisted_observations_with_endpoint_proxy' as const

export async function getETAReliability(request: ETAReliabilityRequest): Promise<ETAReliabilityResponse> {
  const trajectoryID = uuid(request.trajectoryID, 'trajectoryID')
  const asOfTime = timestamp(request.asOfTime, 'asOfTime')
  if (!Number.isInteger(request.durationSeconds) || request.durationSeconds < 1) {
    throw new APIRequestError('ETA Reliability duration must be a positive whole number of seconds.')
  }
  const searchParams = new URLSearchParams({
    as_of_time: asOfTime,
    duration_seconds: String(request.durationSeconds),
  })
  const value = await requestAPIData<unknown>(
    `/api/v1/trajectories/${encodeURIComponent(trajectoryID)}/eta-reliability`,
    { signal: request.signal, searchParams, timeoutMilliseconds: 20_000 }
  )
  return parseETAReliabilityResponse(value)
}

export function parseETAReliabilityResponse(value: unknown): ETAReliabilityResponse {
  const root = record(value, 'response')
  const status = text(root.status, 'status') as ETAReliabilityStatus
  if (!statuses.has(status)) invalid('status is unsupported.')
  const route = record(root.route, 'route')
  const method = record(root.method, 'method')
  const receivedEvidenceClass = text(root.evidence_class, 'evidence_class')
  if (receivedEvidenceClass !== evidenceClass) invalid('evidence_class is unsupported.')
  const metrics = root.metrics == null ? undefined : parseMetrics(root.metrics)
  const result: ETAReliabilityResponse = {
    version: text(root.version, 'version'),
    status,
    trajectory_id: uuid(text(root.trajectory_id, 'trajectory_id'), 'trajectory_id'),
    route: {
      origin_icao_code: icao(route.origin_icao_code, 'route.origin_icao_code'),
      destination_icao_code: icao(route.destination_icao_code, 'route.destination_icao_code'),
    },
    method: {
      name: text(method.name, 'method.name'),
      version: text(method.version, 'method.version'),
      decision_class: text(method.decision_class, 'method.decision_class'),
    },
    target_lead_seconds: positiveWhole(root.target_lead_seconds, 'target_lead_seconds'),
    lead_tolerance_seconds: positiveWhole(root.lead_tolerance_seconds, 'lead_tolerance_seconds'),
    endpoint_radius_km: positive(root.endpoint_radius_km, 'endpoint_radius_km'),
    candidate_count: whole(root.candidate_count, 'candidate_count'),
    eligible_sample_count: whole(root.eligible_sample_count, 'eligible_sample_count'),
    ...(metrics ? { metrics } : {}),
    evidence_class: evidenceClass,
    limitations: list(root.limitations, 'limitations').map((item, index) => {
      const entry = record(item, `limitations[${index}]`)
      return { code: text(entry.code, `limitations[${index}].code`), message: text(entry.message, `limitations[${index}].message`) }
    }),
    input_fingerprint: fingerprint(root.input_fingerprint, 'input_fingerprint'),
    generated_at: timestamp(root.generated_at, 'generated_at'),
  }
  if (result.eligible_sample_count > result.candidate_count) invalid('eligible_sample_count exceeds candidate_count.')
  if (status === 'unavailable' && metrics) invalid('unavailable status must not publish metrics.')
  if (status !== 'unavailable' && (!metrics || metrics.sample_count !== result.eligible_sample_count)) invalid('published metrics do not match eligible samples.')
  return result
}

function parseMetrics(value: unknown): ETAReliabilityMetrics {
  const entry = record(value, 'metrics')
  const median = nonNegative(entry.median_absolute_error_seconds, 'metrics.median_absolute_error_seconds')
  const p80 = nonNegative(entry.p80_absolute_error_seconds, 'metrics.p80_absolute_error_seconds')
  if (p80 < median) invalid('metrics.p80_absolute_error_seconds must not be below the median.')
  return {
    sample_count: positiveWhole(entry.sample_count, 'metrics.sample_count'),
    median_absolute_error_seconds: median,
    p80_absolute_error_seconds: p80,
    within_five_minutes_ratio: ratio(entry.within_five_minutes_ratio, 'metrics.within_five_minutes_ratio'),
    within_ten_minutes_ratio: ratio(entry.within_ten_minutes_ratio, 'metrics.within_ten_minutes_ratio'),
    interval_coverage_ratio: ratio(entry.interval_coverage_ratio, 'metrics.interval_coverage_ratio'),
  }
}

function record(value: unknown, field: string): Record<string, unknown> { if (typeof value !== 'object' || value === null || Array.isArray(value)) invalid(`${field} must be an object.`); return value as Record<string, unknown> }
function list(value: unknown, field: string): unknown[] { if (!Array.isArray(value)) invalid(`${field} must be an array.`); return value }
function text(value: unknown, field: string): string { if (typeof value !== 'string' || value.trim() === '') invalid(`${field} must be a non-empty string.`); return value }
function number(value: unknown, field: string): number { if (typeof value !== 'number' || !Number.isFinite(value)) invalid(`${field} must be finite.`); return value }
function nonNegative(value: unknown, field: string): number { const result = number(value, field); if (result < 0) invalid(`${field} must not be negative.`); return result }
function positive(value: unknown, field: string): number { const result = number(value, field); if (result <= 0) invalid(`${field} must be positive.`); return result }
function whole(value: unknown, field: string): number { const result = nonNegative(value, field); if (!Number.isInteger(result)) invalid(`${field} must be a whole number.`); return result }
function positiveWhole(value: unknown, field: string): number { const result = positive(value, field); if (!Number.isInteger(result)) invalid(`${field} must be a whole number.`); return result }
function ratio(value: unknown, field: string): number { const result = number(value, field); if (result < 0 || result > 1) invalid(`${field} must be between zero and one.`); return result }
function timestamp(value: unknown, field: string): string { const result = text(value, field); if (Number.isNaN(Date.parse(result))) invalid(`${field} must be a valid timestamp.`); return result }
function uuid(value: string, field: string): string { const result = value.trim().toLowerCase(); if (!/^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/.test(result)) invalid(`${field} must be a valid UUID.`); return result }
function icao(value: unknown, field: string): string { const result = text(value, field).trim().toUpperCase(); if (!/^[A-Z0-9]{4}$/.test(result)) invalid(`${field} must be a valid ICAO code.`); return result }
function fingerprint(value: unknown, field: string): string { const result = text(value, field); if (!/^sha256:[0-9a-f]{64}$/.test(result)) invalid(`${field} must be a SHA-256 fingerprint.`); return result }
function invalid(message: string): never { throw new APIRequestError(`The ETA Reliability response is invalid: ${message}`) }
