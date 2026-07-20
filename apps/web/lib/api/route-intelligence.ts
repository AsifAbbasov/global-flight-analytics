import {
  APIRequestError,
  requestAPIData,
  type APIRequestOptions,
} from '@/lib/api/client'
import type {
  RouteIntelligenceAirport,
  RouteIntelligenceConfidenceLevel,
  RouteIntelligenceEndpointRole,
  RouteIntelligenceHistory,
  RouteIntelligenceHistoryOptions,
  RouteIntelligenceRecord,
  RouteIntelligenceStatus,
} from '@/types/route-intelligence'

const statuses = new Set<RouteIntelligenceStatus>(['unavailable', 'partial', 'complete'])
const levels = new Set<RouteIntelligenceConfidenceLevel>(['none', 'low', 'medium', 'high'])
const roles = new Set<RouteIntelligenceEndpointRole>(['origin', 'destination'])
const elevationStatuses = new Set<RouteIntelligenceAirport['elevation_status']>([
  'observed',
  'unknown',
  'invalid',
])

export async function processRouteIntelligence(
  trajectoryID: string,
  options: APIRequestOptions = {}
): Promise<RouteIntelligenceRecord> {
  const id = normalizeTrajectoryID(trajectoryID)
  return parseRecord(
    await requestAPIData<unknown>(
      `/api/v1/trajectories/${encodeURIComponent(id)}/route-intelligence`,
      { ...options, method: 'POST', timeoutMilliseconds: options.timeoutMilliseconds ?? 20_000 }
    )
  )
}

export async function getLatestRouteIntelligence(
  trajectoryID: string,
  options: APIRequestOptions = {}
): Promise<RouteIntelligenceRecord> {
  const id = normalizeTrajectoryID(trajectoryID)
  return parseRecord(
    await requestAPIData<unknown>(
      `/api/v1/trajectories/${encodeURIComponent(id)}/route-intelligence/latest`,
      options
    )
  )
}

export async function getRouteIntelligenceHistory(
  trajectoryID: string,
  options: RouteIntelligenceHistoryOptions = {}
): Promise<RouteIntelligenceHistory> {
  const id = normalizeTrajectoryID(trajectoryID)
  const searchParams = new URLSearchParams()
  if (options.limit !== undefined) {
    if (!Number.isInteger(options.limit) || options.limit < 1 || options.limit > 100) {
      throw new APIRequestError(
        'Route Intelligence history limit must be between one and one hundred.'
      )
    }
    searchParams.set('limit', String(options.limit))
  }
  if (options.beforeAsOfTime !== undefined) {
    timestamp(options.beforeAsOfTime, 'beforeAsOfTime')
    searchParams.set('before_as_of_time', options.beforeAsOfTime)
  }
  return parseHistory(
    await requestAPIData<unknown>(
      `/api/v1/trajectories/${encodeURIComponent(id)}/route-intelligence/history`,
      { signal: options.signal, searchParams }
    )
  )
}

function parseHistory(value: unknown): RouteIntelligenceHistory {
  const r = record(value, 'history')
  const cursor = r.next_before_as_of_time == null
    ? undefined
    : timestamp(r.next_before_as_of_time, 'next_before_as_of_time')
  return {
    items: array(r.items, 'items').map((item, index) => parseRecord(item, `items[${index}]`)),
    has_more: booleanValue(r.has_more, 'has_more'),
    next_before_as_of_time: cursor,
  }
}

function parseRecord(value: unknown, field = 'record'): RouteIntelligenceRecord {
  const r = record(value, field)
  return {
    id: stringValue(r.id, `${field}.id`),
    input_fingerprint: fingerprint(r.input_fingerprint, `${field}.input_fingerprint`),
    stored_at: timestamp(r.stored_at, `${field}.stored_at`),
    result: parseResult(r.result, `${field}.result`),
  }
}

function parseResult(value: unknown, field: string): RouteIntelligenceRecord['result'] {
  const r = record(value, field)
  const status = stringValue(r.status, `${field}.status`) as RouteIntelligenceStatus
  if (!statuses.has(status)) invalid(`${field}.status is unsupported.`)
  const origin = endpoint(r.origin, 'origin', `${field}.origin`)
  const destination = endpoint(r.destination, 'destination', `${field}.destination`)
  return {
    schema_version: stringValue(r.schema_version, `${field}.schema_version`),
    status,
    trajectory_id: normalizeTrajectoryID(stringValue(r.trajectory_id, `${field}.trajectory_id`)),
    identity_key: stringValue(r.identity_key, `${field}.identity_key`, true),
    flight_id: stringValue(r.flight_id, `${field}.flight_id`, true),
    aircraft_id: stringValue(r.aircraft_id, `${field}.aircraft_id`, true),
    icao24: stringValue(r.icao24, `${field}.icao24`).toLowerCase(),
    callsign: stringValue(r.callsign, `${field}.callsign`, true),
    window: windowValue(r.window, `${field}.window`),
    ...(origin ? { origin } : {}),
    ...(destination ? { destination } : {}),
    summary: summaryValue(r.summary, `${field}.summary`),
    confidence: confidence(r.confidence, `${field}.confidence`),
    limitations: array(r.limitations, `${field}.limitations`).map((x,i)=>limitation(x,`${field}.limitations[${i}]`)),
    provenance: provenance(r.provenance, `${field}.provenance`),
    generated_at: timestamp(r.generated_at, `${field}.generated_at`),
  }
}

function endpoint(value: unknown, expected: RouteIntelligenceEndpointRole, field: string) {
  if (value == null) return undefined
  const r = record(value, field)
  const role = stringValue(r.role, `${field}.role`) as RouteIntelligenceEndpointRole
  if (!roles.has(role) || role !== expected) invalid(`${field}.role is invalid.`)
  return {
    role,
    airport: airport(r.airport, `${field}.airport`),
    distance_km: nonNegative(r.distance_km, `${field}.distance_km`),
    confidence: confidence(r.confidence, `${field}.confidence`),
    evidence: array(r.evidence, `${field}.evidence`).map((x,i)=>evidence(x,`${field}.evidence[${i}]`)),
    limitations: array(r.limitations, `${field}.limitations`).map((x,i)=>limitation(x,`${field}.limitations[${i}]`)),
  }
}

function airport(value: unknown, field: string): RouteIntelligenceAirport {
  const r=record(value,field)
  return {
    icao_code:stringValue(r.icao_code,`${field}.icao_code`),
    iata_code:stringValue(r.iata_code,`${field}.iata_code`,true),
    name:stringValue(r.name,`${field}.name`),
    city:stringValue(r.city,`${field}.city`,true),
    country:stringValue(r.country,`${field}.country`,true),
    latitude:bounded(r.latitude,`${field}.latitude`,-90,90),
    longitude:bounded(r.longitude,`${field}.longitude`,-180,180),
    ...airportElevation(
      r.elevation_m,
      r.elevation_status,
      field
    ),
    timezone:stringValue(r.timezone,`${field}.timezone`,true),
  }
}
function airportElevation(
  value: unknown,
  statusValue: unknown,
  field: string
): Pick<RouteIntelligenceAirport, 'elevation_m' | 'elevation_status'> {
  const status = stringValue(
    statusValue,
    `${field}.elevation_status`
  ) as RouteIntelligenceAirport['elevation_status']
  if (!elevationStatuses.has(status)) {
    invalid(`${field}.elevation_status is unsupported.`)
  }
  if (status === 'observed') {
    return {
      elevation_m: numberValue(value, `${field}.elevation_m`),
      elevation_status: status,
    }
  }
  if (value !== null) {
    invalid(
      `${field}.elevation_m must be null when elevation is not observed.`
    )
  }
  return { elevation_m: null, elevation_status: status }
}
function windowValue(value: unknown,field:string){const r=record(value,field);return{start_time:timestamp(r.start_time,`${field}.start_time`),end_time:timestamp(r.end_time,`${field}.end_time`),as_of_time:timestamp(r.as_of_time,`${field}.as_of_time`)}}
function summaryValue(value:unknown,field:string){const r=record(value,field);return{great_circle_distance_km:nonNegative(r.great_circle_distance_km,`${field}.great_circle_distance_km`),same_airport:booleanValue(r.same_airport,`${field}.same_airport`)}}
function confidence(value:unknown,field:string){const r=record(value,field);const level=stringValue(r.level,`${field}.level`) as RouteIntelligenceConfidenceLevel;if(!levels.has(level)) invalid(`${field}.level is unsupported.`);return{score:ratio(r.score,`${field}.score`),level,evidence_count:integer(r.evidence_count,`${field}.evidence_count`),reasons:array(r.reasons,`${field}.reasons`).map((x,i)=>{const a=record(x,`${field}.reasons[${i}]`);return{code:stringValue(a.code,`${field}.reasons[${i}].code`),message:stringValue(a.message,`${field}.reasons[${i}].message`),contribution:numberValue(a.contribution,`${field}.reasons[${i}].contribution`)}})}}
function evidence(value:unknown,field:string){const r=record(value,field);return{type:stringValue(r.type,`${field}.type`),source_name:stringValue(r.source_name,`${field}.source_name`),source_version:stringValue(r.source_version,`${field}.source_version`),score:ratio(r.score,`${field}.score`),weight:ratio(r.weight,`${field}.weight`),observed_at:timestamp(r.observed_at,`${field}.observed_at`),summary:stringValue(r.summary,`${field}.summary`),attributes:array(r.attributes,`${field}.attributes`).map((x,i)=>{const a=record(x,`${field}.attributes[${i}]`);return{key:stringValue(a.key,`${field}.attributes[${i}].key`),value:stringValue(a.value,`${field}.attributes[${i}].value`,true)}})}}
function limitation(value:unknown,field:string){const r=record(value,field);return{code:stringValue(r.code,`${field}.code`),message:stringValue(r.message,`${field}.message`),scope:stringValue(r.scope,`${field}.scope`)}}
function provenance(value:unknown,field:string){const r=record(value,field);return{resolver_version:stringValue(r.resolver_version,`${field}.resolver_version`),input_fingerprint:fingerprint(r.input_fingerprint,`${field}.input_fingerprint`),trajectory_updated_at:timestamp(r.trajectory_updated_at,`${field}.trajectory_updated_at`),source_names:array(r.source_names,`${field}.source_names`).map((x,i)=>stringValue(x,`${field}.source_names[${i}]`))}}

function normalizeTrajectoryID(value:string){const v=value.trim().toLowerCase();if(!/^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/.test(v))throw new APIRequestError('Trajectory identifier must be a valid UUID.');return v}
function record(value:unknown,field:string):Record<string,unknown>{if(typeof value!=='object'||value===null||Array.isArray(value))invalid(`${field} must be an object.`);return value as Record<string,unknown>}
function array(value:unknown,field:string):unknown[]{if(!Array.isArray(value))invalid(`${field} must be an array.`);return value}
function stringValue(value:unknown,field:string,empty=false):string{if(typeof value!=='string')invalid(`${field} must be a string.`);if(!empty&&value.trim()==='')invalid(`${field} must not be empty.`);return value}
function booleanValue(value:unknown,field:string):boolean{if(typeof value!=='boolean')invalid(`${field} must be a boolean.`);return value}
function numberValue(value:unknown,field:string):number{if(typeof value!=='number'||!Number.isFinite(value))invalid(`${field} must be finite.`);return value}
function nonNegative(value:unknown,field:string):number{const v=numberValue(value,field);if(v<0)invalid(`${field} must not be negative.`);return v}
function integer(value:unknown,field:string):number{const v=nonNegative(value,field);if(!Number.isInteger(v))invalid(`${field} must be an integer.`);return v}
function ratio(value:unknown,field:string):number{const v=numberValue(value,field);if(v<0||v>1)invalid(`${field} must be between zero and one.`);return v}
function bounded(value:unknown,field:string,min:number,max:number):number{const v=numberValue(value,field);if(v<min||v>max)invalid(`${field} is outside its allowed range.`);return v}
function timestamp(value:unknown,field:string):string{const v=stringValue(value,field);if(Number.isNaN(Date.parse(v)))invalid(`${field} must be a valid timestamp.`);return v}
function fingerprint(value:unknown,field:string):string{const v=stringValue(value,field);if(!/^sha256:[0-9a-f]{64}$/.test(v))invalid(`${field} must be a SHA-256 fingerprint.`);return v}
function invalid(message:string):never{throw new APIRequestError(`The Route Intelligence response is invalid: ${message}`)}
