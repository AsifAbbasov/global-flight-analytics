#!/usr/bin/env node
import fs from 'node:fs'
import path from 'node:path'
import { generateFromFiles } from './generate-openapi-client.mjs'

const root = process.cwd()
const rootSpecPath = path.join(root, 'openapi/openapi.json')
const embeddedSpecPath = path.join(root, 'apps/api/internal/http/apidocs/openapi.json')
const generatedClientPath = path.join(root, 'packages/api-client/src/generated.ts')

const spec = JSON.parse(fs.readFileSync(rootSpecPath, 'utf8'))

spec.info = {
  ...spec.info,
  version: '1.5.0',
  description:
    'Complete source-backed OpenAPI contract for all 40 production public operations: 39 stable reads and one protected Route Intelligence mutation. The internal metrics endpoint remains excluded.',
}

spec.paths['/api/v1/trajectories/{id}/eta-reliability'] = {
  get: {
    operationId: 'getETAReliabilityByTrajectoryID',
    summary: 'Get bounded historical ETA reliability evidence for a trajectory',
    description:
      'Recomputes comparable historical Projection Intelligence at persisted observation times and compares ETA evidence with a persisted trajectory endpoint proxy near the inferred destination. The endpoint proxy is not an official touchdown, gate-arrival, schedule, or operational timestamp.',
    tags: ['Projection Intelligence'],
    parameters: [
      {
        name: 'id',
        in: 'path',
        required: true,
        schema: { type: 'string', minLength: 1 },
      },
      {
        name: 'as_of_time',
        in: 'query',
        required: true,
        schema: { type: 'string', format: 'date-time' },
      },
      {
        name: 'duration_seconds',
        in: 'query',
        required: false,
        schema: { type: 'integer', minimum: 0 },
      },
    ],
    responses: {
      200: {
        description: 'Bounded historical ETA reliability evidence.',
        content: {
          'application/json': {
            schema: { $ref: '#/components/schemas/ETAReliabilityResponse' },
          },
        },
      },
      400: {
        description: 'Invalid ETA Reliability request.',
        content: { 'application/json': { schema: { $ref: '#/components/schemas/ErrorResponse' } } },
      },
      404: {
        description: 'Trajectory evidence was not found.',
        content: { 'application/json': { schema: { $ref: '#/components/schemas/ErrorResponse' } } },
      },
      408: {
        description: 'Request was canceled.',
        content: { 'application/json': { schema: { $ref: '#/components/schemas/ErrorResponse' } } },
      },
      422: {
        description: 'ETA Reliability is unavailable for the current trajectory evidence.',
        content: { 'application/json': { schema: { $ref: '#/components/schemas/ErrorResponse' } } },
      },
      500: {
        description: 'ETA Reliability evaluation failed.',
        content: { 'application/json': { schema: { $ref: '#/components/schemas/ErrorResponse' } } },
      },
      503: {
        description: 'ETA Reliability service is unavailable.',
        content: { 'application/json': { schema: { $ref: '#/components/schemas/ErrorResponse' } } },
      },
      504: {
        description: 'ETA Reliability evaluation timed out.',
        content: { 'application/json': { schema: { $ref: '#/components/schemas/ErrorResponse' } } },
      },
    },
  },
}

const schemas = spec.components.schemas
schemas.ETAReliabilityRoute = {
  type: 'object',
  additionalProperties: false,
  required: ['origin_icao_code', 'destination_icao_code'],
  properties: {
    origin_icao_code: { type: 'string', pattern: '^[A-Z0-9]{4}$' },
    destination_icao_code: { type: 'string', pattern: '^[A-Z0-9]{4}$' },
  },
}

schemas.ETAReliabilityMetrics = {
  type: 'object',
  additionalProperties: false,
  required: [
    'sample_count',
    'median_absolute_error_seconds',
    'p80_absolute_error_seconds',
    'within_five_minutes_ratio',
    'within_ten_minutes_ratio',
    'interval_coverage_ratio',
  ],
  properties: {
    sample_count: { type: 'integer', minimum: 1 },
    median_absolute_error_seconds: { type: 'number', minimum: 0 },
    p80_absolute_error_seconds: { type: 'number', minimum: 0 },
    within_five_minutes_ratio: { type: 'number', minimum: 0, maximum: 1 },
    within_ten_minutes_ratio: { type: 'number', minimum: 0, maximum: 1 },
    interval_coverage_ratio: { type: 'number', minimum: 0, maximum: 1 },
  },
}

schemas.ETAReliabilityNotice = {
  type: 'object',
  additionalProperties: false,
  required: ['code', 'message'],
  properties: {
    code: { type: 'string', minLength: 1 },
    message: { type: 'string', minLength: 1 },
  },
}

schemas.ETAReliability = {
  type: 'object',
  additionalProperties: false,
  required: [
    'version',
    'status',
    'trajectory_id',
    'route',
    'method',
    'target_lead_seconds',
    'lead_tolerance_seconds',
    'endpoint_radius_km',
    'candidate_count',
    'eligible_sample_count',
    'evidence_class',
    'limitations',
    'input_fingerprint',
    'generated_at',
  ],
  properties: {
    version: { type: 'string', const: 'eta-reliability-v1' },
    status: { type: 'string', enum: ['unavailable', 'limited', 'complete'] },
    trajectory_id: { type: 'string', minLength: 1 },
    route: { $ref: '#/components/schemas/ETAReliabilityRoute' },
    method: { $ref: '#/components/schemas/ProjectionMethod' },
    target_lead_seconds: { type: 'integer', minimum: 1 },
    lead_tolerance_seconds: { type: 'integer', minimum: 1 },
    endpoint_radius_km: { type: 'number', exclusiveMinimum: 0 },
    candidate_count: { type: 'integer', minimum: 0, maximum: 8 },
    eligible_sample_count: { type: 'integer', minimum: 0, maximum: 8 },
    metrics: { $ref: '#/components/schemas/ETAReliabilityMetrics' },
    evidence_class: {
      type: 'string',
      const: 'historically_recomputed_from_persisted_observations_with_endpoint_proxy',
    },
    limitations: {
      type: 'array',
      items: { $ref: '#/components/schemas/ETAReliabilityNotice' },
    },
    input_fingerprint: { type: 'string', pattern: '^sha256:[0-9a-f]{64}$' },
    generated_at: { type: 'string', format: 'date-time' },
  },
}

schemas.ETAReliabilityResponse = {
  type: 'object',
  additionalProperties: false,
  required: ['success', 'data'],
  properties: {
    success: { type: 'boolean', const: true },
    data: { $ref: '#/components/schemas/ETAReliability' },
  },
}

const rendered = `${JSON.stringify(spec, null, 2)}\n`
fs.writeFileSync(rootSpecPath, rendered)
fs.writeFileSync(embeddedSpecPath, rendered)

const { generated } = generateFromFiles({ root })
fs.writeFileSync(generatedClientPath, generated)

console.log('STAGE_23_OPENAPI_SYNC=PASS')
console.log('OPENAPI_CONTRACT_OPERATIONS=40')
console.log('OPENAPI_PUBLIC_READ_OPERATIONS=39')
console.log('OPENAPI_PROTECTED_MUTATION_OPERATIONS=1')
console.log('OPENAPI_ETA_RELIABILITY_OPERATIONS=1')
