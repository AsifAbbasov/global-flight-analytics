import { APIRequestError, requestAPIData } from '@/lib/api/client'
import type {
  ProjectionArrivalEstimate,
  ProjectionConfidence,
  ProjectionConfidenceLevel,
  ProjectionEvidence,
  ProjectionIntelligenceRequest,
  ProjectionIntelligenceResponse,
  ProjectionLimitation,
  ProjectionPoint,
} from '@/types/projection-intelligence'

const confidenceLevels = new Set<ProjectionConfidenceLevel>(['none', 'low', 'medium', 'high'])

export async function getProjectionIntelligence(request: ProjectionIntelligenceRequest): Promise<ProjectionIntelligenceResponse> {
  const trajectoryID = uuid(request.trajectoryID, 'trajectoryID')
  const asOfTime = timestamp(request.asOfTime, 'asOfTime')
  if (!Number.isInteger(request.durationSeconds) || request.durationSeconds < 1) {
    throw new APIRequestError('Projection duration must be a positive whole number of seconds.')
  }
  const searchParams = new URLSearchParams({
    as_of_time: asOfTime,
    duration_seconds: String(request.durationSeconds),
  })
  const value = await requestAPIData<unknown>(
    `/api/v1/trajectories/${encodeURIComponent(trajectoryID)}/projection-intelligence`,
    { signal: request.signal, searchParams, timeoutMilliseconds: 20_000 }
  )
  return parseResponse(value)
}

function parseResponse(value: unknown): ProjectionIntelligenceResponse {
  const root = record(value, 'response')
  const projection = record(root.projection, 'projection')
  const method = record(projection.method, 'projection.method')
  const horizon = record(projection.horizon, 'projection.horizon')
  const provenance = record(projection.provenance, 'projection.provenance')
  const arrival = projection.arrival == null ? undefined : parseArrival(projection.arrival)
  return {
    version: text(root.version, 'version'),
    strategy: text(root.strategy, 'strategy'),
    fallback_reason: optionalText(root.fallback_reason, 'fallback_reason') ?? '',
    arrival_status: text(root.arrival_status, 'arrival_status'),
    projection: {
      schema_version: text(projection.schema_version, 'projection.schema_version'),
      status: text(projection.status, 'projection.status'),
      trajectory_id: uuid(text(projection.trajectory_id, 'projection.trajectory_id'), 'projection.trajectory_id'),
      flight_id: optionalText(projection.flight_id, 'projection.flight_id') ?? '',
      aircraft_id: optionalText(projection.aircraft_id, 'projection.aircraft_id') ?? '',
      icao24: optionalText(projection.icao24, 'projection.icao24') ?? '',
      callsign: optionalText(projection.callsign, 'projection.callsign') ?? '',
      method: { name: text(method.name, 'projection.method.name'), version: text(method.version, 'projection.method.version'), decision_class: text(method.decision_class, 'projection.method.decision_class') },
      horizon: { as_of_time: timestamp(horizon.as_of_time, 'projection.horizon.as_of_time'), end_time: timestamp(horizon.end_time, 'projection.horizon.end_time'), step_seconds: whole(horizon.step_seconds, 'projection.horizon.step_seconds'), duration_seconds: whole(horizon.duration_seconds, 'projection.horizon.duration_seconds') },
      points: list(projection.points, 'projection.points').map(parsePoint),
      ...(arrival ? { arrival } : {}),
      confidence: parseConfidence(projection.confidence, 'projection.confidence'),
      limitations: list(projection.limitations, 'projection.limitations').map((item, index) => parseLimitation(item, `projection.limitations[${index}]`)),
      explanations: list(projection.explanations, 'projection.explanations').map((item, index) => { const entry=record(item,`projection.explanations[${index}]`); return {code:text(entry.code,`projection.explanations[${index}].code`),message:text(entry.message,`projection.explanations[${index}].message`)} }),
      scope_guard: text(projection.scope_guard, 'projection.scope_guard'),
      provenance: {
        input_fingerprint: fingerprint(provenance.input_fingerprint, 'projection.provenance.input_fingerprint'),
        inputs: list(provenance.inputs, 'projection.provenance.inputs').map((item,index)=>{const entry=record(item,`projection.provenance.inputs[${index}]`);const limitation=optionalText(entry.limitation,`projection.provenance.inputs[${index}].limitation`);return{name:text(entry.name,`projection.provenance.inputs[${index}].name`),classification:text(entry.classification,`projection.provenance.inputs[${index}].classification`),source_name:text(entry.source_name,`projection.provenance.inputs[${index}].source_name`),observed_at:timestamp(entry.observed_at,`projection.provenance.inputs[${index}].observed_at`),retrieved_at:timestamp(entry.retrieved_at,`projection.provenance.inputs[${index}].retrieved_at`),...(limitation?{limitation}:{})}}),
        latest_input_observed_at: timestamp(provenance.latest_input_observed_at, 'projection.provenance.latest_input_observed_at'),
      },
      generated_at: timestamp(projection.generated_at, 'projection.generated_at'),
    },
    evidence: parseEvidence(root.evidence),
    notices: list(root.notices, 'notices').map((item,index)=>{const entry=record(item,`notices[${index}]`);return{code:text(entry.code,`notices[${index}].code`),message:text(entry.message,`notices[${index}].message`)}}),
    input_fingerprint: fingerprint(root.input_fingerprint, 'input_fingerprint'),
    generated_at: timestamp(root.generated_at, 'generated_at'),
  }
}

function parsePoint(value: unknown, index: number): ProjectionPoint { const field=`projection.points[${index}]`;const entry=record(value,field);const position=record(entry.position,`${field}.position`);const uncertainty=record(entry.uncertainty,`${field}.uncertainty`);const altitude=optionalNumber(position.altitude_m,`${field}.position.altitude_m`);const vertical=optionalNumber(uncertainty.vertical_radius_m,`${field}.uncertainty.vertical_radius_m`);return{sequence:whole(entry.sequence,`${field}.sequence`),forecast_time:timestamp(entry.forecast_time,`${field}.forecast_time`),position:{latitude:bounded(position.latitude,`${field}.position.latitude`,-90,90),longitude:bounded(position.longitude,`${field}.position.longitude`,-180,180),...(altitude===undefined?{}:{altitude_m:altitude})},uncertainty:{horizontal_radius_m:nonNegative(uncertainty.horizontal_radius_m,`${field}.uncertainty.horizontal_radius_m`),...(vertical===undefined?{}:{vertical_radius_m:nonNegative(vertical,`${field}.uncertainty.vertical_radius_m`)})},confidence:parseConfidence(entry.confidence,`${field}.confidence`)}}
function parseArrival(value:unknown):ProjectionArrivalEstimate{const entry=record(value,'projection.arrival');return{airport_icao_code:text(entry.airport_icao_code,'projection.arrival.airport_icao_code'),earliest_time:timestamp(entry.earliest_time,'projection.arrival.earliest_time'),estimated_time:timestamp(entry.estimated_time,'projection.arrival.estimated_time'),latest_time:timestamp(entry.latest_time,'projection.arrival.latest_time'),confidence:parseConfidence(entry.confidence,'projection.arrival.confidence'),limitations:list(entry.limitations,'projection.arrival.limitations').map((item,index)=>parseLimitation(item,`projection.arrival.limitations[${index}]`))}}
function parseConfidence(value:unknown,field:string):ProjectionConfidence{const entry=record(value,field);const level=text(entry.level,`${field}.level`) as ProjectionConfidenceLevel;if(!confidenceLevels.has(level))invalid(`${field}.level is unsupported.`);return{score:ratio(entry.score,`${field}.score`),level,reasons:list(entry.reasons,`${field}.reasons`).map((item,index)=>{const reason=record(item,`${field}.reasons[${index}]`);return{code:text(reason.code,`${field}.reasons[${index}].code`),message:text(reason.message,`${field}.reasons[${index}].message`),contribution:number(reason.contribution,`${field}.reasons[${index}].contribution`)}})}}
function parseLimitation(value:unknown,field:string):ProjectionLimitation{const entry=record(value,field);return{code:text(entry.code,`${field}.code`),message:text(entry.message,`${field}.message`),scope:text(entry.scope,`${field}.scope`)}}
function parseEvidence(value:unknown):ProjectionEvidence{const entry=record(value,'evidence');return{...optionalRecord(entry.neighbor_selection,'evidence.neighbor_selection','neighbor_selection'),...optionalRecord(entry.pattern_confidence,'evidence.pattern_confidence','pattern_confidence'),...optionalRecord(entry.freshness,'evidence.freshness','freshness'),...optionalRecord(entry.route_frequency,'evidence.route_frequency','route_frequency')}}
function optionalRecord(value:unknown,field:string,key:keyof ProjectionEvidence):ProjectionEvidence{if(value==null)return{};return{[key]:record(value,field)}}
function record(value:unknown,field:string):Record<string,unknown>{if(typeof value!=='object'||value===null||Array.isArray(value))invalid(`${field} must be an object.`);return value as Record<string,unknown>}
function list(value:unknown,field:string):unknown[]{if(!Array.isArray(value))invalid(`${field} must be an array.`);return value}
function text(value:unknown,field:string):string{if(typeof value!=='string'||value.trim()==='')invalid(`${field} must be a non-empty string.`);return value}
function optionalText(value:unknown,field:string):string|undefined{if(value==null)return undefined;if(typeof value!=='string')invalid(`${field} must be a string.`);return value}
function number(value:unknown,field:string):number{if(typeof value!=='number'||!Number.isFinite(value))invalid(`${field} must be finite.`);return value}
function optionalNumber(value:unknown,field:string):number|undefined{if(value==null)return undefined;return number(value,field)}
function nonNegative(value:unknown,field:string):number{const result=number(value,field);if(result<0)invalid(`${field} must not be negative.`);return result}
function whole(value:unknown,field:string):number{const result=nonNegative(value,field);if(!Number.isInteger(result))invalid(`${field} must be a whole number.`);return result}
function ratio(value:unknown,field:string):number{const result=number(value,field);if(result<0||result>1)invalid(`${field} must be between zero and one.`);return result}
function bounded(value:unknown,field:string,min:number,max:number):number{const result=number(value,field);if(result<min||result>max)invalid(`${field} is outside its allowed range.`);return result}
function timestamp(value:unknown,field:string):string{const result=text(value,field);if(Number.isNaN(Date.parse(result)))invalid(`${field} must be a valid timestamp.`);return result}
function uuid(value:string,field:string):string{const result=value.trim().toLowerCase();if(!/^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/.test(result))invalid(`${field} must be a valid UUID.`);return result}
function fingerprint(value:unknown,field:string):string{const result=text(value,field);if(!/^sha256:[0-9a-f]{64}$/.test(result))invalid(`${field} must be a SHA-256 fingerprint.`);return result}
function invalid(message:string):never{throw new APIRequestError(`The Projection Intelligence response is invalid: ${message}`)}
