import { APIRequestError, requestAPIData } from '@/lib/api/client'
import { parseProjectionIntelligenceResponse } from '@/lib/api/projection-intelligence'
import type {
  StabilityIntelligenceAnalysis,
  StabilityIntelligenceAnalysisMetrics,
  StabilityIntelligenceConfidenceSummary,
  StabilityIntelligenceFailure,
  StabilityIntelligenceFailureExplanation,
  StabilityIntelligenceInterventionGuard,
  StabilityIntelligenceRequest,
  StabilityIntelligenceResponse,
  StabilityIntelligenceScopeEnforcement,
  StabilityIntelligenceScopeViolation,
  StabilityIntelligenceTransition,
  StabilityIntelligenceTransitionMetrics,
  StabilityIntelligenceVersion,
} from '@/types/stability-intelligence'

const minimumAsOfTimeCount = 2
const maximumAsOfTimeCount = 8

export async function getStabilityIntelligence(
  request: StabilityIntelligenceRequest
): Promise<StabilityIntelligenceResponse> {
  const trajectoryID = uuid(request.trajectoryID, 'trajectoryID')
  const asOfTimes = orderedTimestamps(request.asOfTimes, 'asOfTimes')

  if (!Number.isInteger(request.durationSeconds) || request.durationSeconds < 1) {
    throw new APIRequestError(
      'Stability Intelligence duration must be a positive whole number of seconds.'
    )
  }

  const searchParams = new URLSearchParams({
    as_of_times: asOfTimes.join(','),
    duration_seconds: String(request.durationSeconds),
  })

  const value = await requestAPIData<unknown>(
    `/api/v1/trajectories/${encodeURIComponent(trajectoryID)}/stability-intelligence`,
    {
      signal: request.signal,
      searchParams,
      timeoutMilliseconds: 45_000,
    }
  )

  const result = parseStabilityIntelligenceResponse(value)
  if (result.trajectory_id !== trajectoryID) {
    invalid('trajectory_id does not match the requested trajectory.')
  }
  if (!timestampsEqual(result.as_of_times, asOfTimes)) {
    invalid('as_of_times do not match the requested analytical history.')
  }

  return result
}

export function parseStabilityIntelligenceResponse(
  value: unknown
): StabilityIntelligenceResponse {
  const root = record(value, 'response')
  const trajectoryID = uuid(
    text(root.trajectory_id, 'trajectory_id'),
    'trajectory_id'
  )
  const asOfTimes = orderedTimestamps(
    list(root.as_of_times, 'as_of_times').map((item, index) =>
      timestamp(item, `as_of_times[${index}]`)
    ),
    'as_of_times'
  )

  const projections = list(root.projections, 'projections').map((item) =>
    parseProjectionIntelligenceResponse(item)
  )
  const versions = list(root.forecast_versions, 'forecast_versions').map(
    parseVersion
  )
  const transitions = list(root.transitions, 'transitions').map(parseTransition)

  if (projections.length !== asOfTimes.length) {
    invalid('projections must contain one result for every as-of time.')
  }
  if (versions.length !== asOfTimes.length) {
    invalid('forecast_versions must contain one version for every as-of time.')
  }
  if (transitions.length !== Math.max(0, versions.length - 1)) {
    invalid('transitions must connect every consecutive forecast version.')
  }

  projections.forEach((projection, index) => {
    if (projection.projection.trajectory_id !== trajectoryID) {
      invalid(`projections[${index}] belongs to another trajectory.`)
    }
    if (
      Date.parse(projection.projection.horizon.as_of_time) !==
      Date.parse(asOfTimes[index])
    ) {
      invalid(`projections[${index}] does not match as_of_times[${index}].`)
    }
  })

  versions.forEach((version, index) => {
    if (version.ordinal !== index + 1) {
      invalid(`forecast_versions[${index}].ordinal is not sequential.`)
    }
    if (index === 0 && version.parent_version_id) {
      invalid('the first forecast version must not have a parent version.')
    }
    if (
      index > 0 &&
      version.parent_version_id !== versions[index - 1].version_id
    ) {
      invalid(`forecast_versions[${index}] has an invalid parent version.`)
    }
  })

  transitions.forEach((transition, index) => {
    if (
      transition.baseline_version_id !== versions[index].version_id ||
      transition.candidate_version_id !== versions[index + 1].version_id
    ) {
      invalid(`transitions[${index}] does not match the forecast lineage.`)
    }
  })

  const analysis = parseAnalysis(root.forecast_analysis)
  if (
    analysis.metrics.version_count !== versions.length ||
    analysis.metrics.transition_count !== transitions.length
  ) {
    invalid('forecast_analysis counts do not match the returned history.')
  }

  return {
    version: text(root.version, 'version'),
    trajectory_id: trajectoryID,
    as_of_times: asOfTimes,
    projections,
    forecast_versions: versions,
    transitions,
    forecast_analysis: analysis,
    propagated_confidence: parseConfidenceSummary(root.propagated_confidence),
    failure_explanation: parseFailureExplanation(root.failure_explanation),
    unknown_intervention: parseInterventionGuard(root.unknown_intervention),
    scope_enforcement: parseScopeEnforcement(root.scope_enforcement),
    scope_guards: list(root.scope_guards, 'scope_guards').map((item, index) =>
      text(item, `scope_guards[${index}]`)
    ),
    input_fingerprint: fingerprint(root.input_fingerprint, 'input_fingerprint'),
    generated_at: timestamp(root.generated_at, 'generated_at'),
  }
}

function parseVersion(value: unknown, index: number): StabilityIntelligenceVersion {
  const field = `forecast_versions[${index}]`
  const entry = record(value, field)
  const parentVersionID = optionalText(
    entry.parent_version_id,
    `${field}.parent_version_id`
  )

  return {
    version_id: text(entry.version_id, `${field}.version_id`),
    ordinal: positiveWhole(entry.ordinal, `${field}.ordinal`),
    ...(parentVersionID ? { parent_version_id: parentVersionID } : {}),
    method_name: text(entry.method_name, `${field}.method_name`),
    method_version: text(entry.method_version, `${field}.method_version`),
    policy_version: text(entry.policy_version, `${field}.policy_version`),
    implementation_version: text(
      entry.implementation_version,
      `${field}.implementation_version`
    ),
    input_fingerprint: fingerprint(
      entry.input_fingerprint,
      `${field}.input_fingerprint`
    ),
    output_fingerprint: fingerprint(
      entry.output_fingerprint,
      `${field}.output_fingerprint`
    ),
    decision_fingerprint: fingerprint(
      entry.decision_fingerprint,
      `${field}.decision_fingerprint`
    ),
    created_at: timestamp(entry.created_at, `${field}.created_at`),
  }
}

function parseTransition(
  value: unknown,
  index: number
): StabilityIntelligenceTransition {
  const field = `transitions[${index}]`
  const entry = record(value, field)

  return {
    baseline_version_id: text(
      entry.baseline_version_id,
      `${field}.baseline_version_id`
    ),
    candidate_version_id: text(
      entry.candidate_version_id,
      `${field}.candidate_version_id`
    ),
    level: text(entry.level, `${field}.level`),
    score: ratio(entry.score, `${field}.score`),
    metrics: parseTransitionMetrics(entry.metrics, `${field}.metrics`),
    input_fingerprint: fingerprint(
      entry.input_fingerprint,
      `${field}.input_fingerprint`
    ),
    evaluated_at: timestamp(entry.evaluated_at, `${field}.evaluated_at`),
  }
}

function parseTransitionMetrics(
  value: unknown,
  field: string
): StabilityIntelligenceTransitionMetrics {
  const entry = record(value, field)

  return {
    aligned_point_count: whole(
      entry.aligned_point_count,
      `${field}.aligned_point_count`
    ),
    aligned_point_share: ratio(
      entry.aligned_point_share,
      `${field}.aligned_point_share`
    ),
    mean_horizontal_shift_kilometers: nonNegative(
      entry.mean_horizontal_shift_kilometers,
      `${field}.mean_horizontal_shift_kilometers`
    ),
    maximum_horizontal_shift_kilometers: nonNegative(
      entry.maximum_horizontal_shift_kilometers,
      `${field}.maximum_horizontal_shift_kilometers`
    ),
    aggregate_confidence_delta: number(
      entry.aggregate_confidence_delta,
      `${field}.aggregate_confidence_delta`
    ),
    mean_relative_horizontal_uncertainty_change: number(
      entry.mean_relative_horizontal_uncertainty_change,
      `${field}.mean_relative_horizontal_uncertainty_change`
    ),
    arrival_comparable: boolean(
      entry.arrival_comparable,
      `${field}.arrival_comparable`
    ),
    arrival_shift_seconds: number(
      entry.arrival_shift_seconds,
      `${field}.arrival_shift_seconds`
    ),
    method_changed: boolean(entry.method_changed, `${field}.method_changed`),
    policy_changed: boolean(entry.policy_changed, `${field}.policy_changed`),
    implementation_changed: boolean(
      entry.implementation_changed,
      `${field}.implementation_changed`
    ),
    input_changed: boolean(entry.input_changed, `${field}.input_changed`),
    output_changed: boolean(entry.output_changed, `${field}.output_changed`),
  }
}

function parseAnalysis(value: unknown): StabilityIntelligenceAnalysis {
  const field = 'forecast_analysis'
  const entry = record(value, field)

  return {
    status: text(entry.status, `${field}.status`),
    trend: text(entry.trend, `${field}.trend`),
    health: text(entry.health, `${field}.health`),
    metrics: parseAnalysisMetrics(entry.metrics, `${field}.metrics`),
    confidence_score: ratio(
      entry.confidence_score,
      `${field}.confidence_score`
    ),
    confidence_level: text(
      entry.confidence_level,
      `${field}.confidence_level`
    ),
    input_fingerprint: fingerprint(
      entry.input_fingerprint,
      `${field}.input_fingerprint`
    ),
  }
}

function parseAnalysisMetrics(
  value: unknown,
  field: string
): StabilityIntelligenceAnalysisMetrics {
  const entry = record(value, field)

  return {
    version_count: whole(entry.version_count, `${field}.version_count`),
    transition_count: whole(
      entry.transition_count,
      `${field}.transition_count`
    ),
    comparable_transition_count: whole(
      entry.comparable_transition_count,
      `${field}.comparable_transition_count`
    ),
    stable_transition_share: ratio(
      entry.stable_transition_share,
      `${field}.stable_transition_share`
    ),
    comparable_transition_share: ratio(
      entry.comparable_transition_share,
      `${field}.comparable_transition_share`
    ),
    material_change_share: ratio(
      entry.material_change_share,
      `${field}.material_change_share`
    ),
    mean_stability_score: ratio(
      entry.mean_stability_score,
      `${field}.mean_stability_score`
    ),
    minimum_stability_score: ratio(
      entry.minimum_stability_score,
      `${field}.minimum_stability_score`
    ),
    score_standard_deviation: nonNegative(
      entry.score_standard_deviation,
      `${field}.score_standard_deviation`
    ),
    longest_stable_run: whole(
      entry.longest_stable_run,
      `${field}.longest_stable_run`
    ),
    method_change_count: whole(
      entry.method_change_count,
      `${field}.method_change_count`
    ),
    policy_change_count: whole(
      entry.policy_change_count,
      `${field}.policy_change_count`
    ),
    implementation_change_count: whole(
      entry.implementation_change_count,
      `${field}.implementation_change_count`
    ),
    input_change_count: whole(
      entry.input_change_count,
      `${field}.input_change_count`
    ),
    output_change_count: whole(
      entry.output_change_count,
      `${field}.output_change_count`
    ),
    mean_horizontal_shift_kilometers: nonNegative(
      entry.mean_horizontal_shift_kilometers,
      `${field}.mean_horizontal_shift_kilometers`
    ),
    maximum_horizontal_shift_kilometers: nonNegative(
      entry.maximum_horizontal_shift_kilometers,
      `${field}.maximum_horizontal_shift_kilometers`
    ),
    latest_level: text(entry.latest_level, `${field}.latest_level`),
  }
}

function parseConfidenceSummary(
  value: unknown
): StabilityIntelligenceConfidenceSummary {
  const field = 'propagated_confidence'
  const entry = record(value, field)
  const limitingDependencyID = optionalText(
    entry.limiting_dependency_id,
    `${field}.limiting_dependency_id`
  )

  return {
    status: text(entry.status, `${field}.status`),
    score: ratio(entry.score, `${field}.score`),
    level: text(entry.level, `${field}.level`),
    target_node_id: text(entry.target_node_id, `${field}.target_node_id`),
    ...(limitingDependencyID
      ? { limiting_dependency_id: limitingDependencyID }
      : {}),
    input_fingerprint: fingerprint(
      entry.input_fingerprint,
      `${field}.input_fingerprint`
    ),
  }
}

function parseFailureExplanation(
  value: unknown
): StabilityIntelligenceFailureExplanation {
  const field = 'failure_explanation'
  const entry = record(value, field)

  return {
    status: text(entry.status, `${field}.status`),
    primary_code: text(entry.primary_code, `${field}.primary_code`),
    blocking_count: whole(entry.blocking_count, `${field}.blocking_count`),
    warning_count: whole(entry.warning_count, `${field}.warning_count`),
    unknown_cause_count: whole(
      entry.unknown_cause_count,
      `${field}.unknown_cause_count`
    ),
    confidence_score: ratio(
      entry.confidence_score,
      `${field}.confidence_score`
    ),
    confidence_level: text(
      entry.confidence_level,
      `${field}.confidence_level`
    ),
    failures: list(entry.failures, `${field}.failures`).map(parseFailure),
    input_fingerprint: fingerprint(
      entry.input_fingerprint,
      `${field}.input_fingerprint`
    ),
  }
}

function parseFailure(value: unknown, index: number): StabilityIntelligenceFailure {
  const field = `failure_explanation.failures[${index}]`
  const entry = record(value, field)

  return {
    rank: positiveWhole(entry.rank, `${field}.rank`),
    code: text(entry.code, `${field}.code`),
    category: text(entry.category, `${field}.category`),
    severity: text(entry.severity, `${field}.severity`),
    classification: text(entry.classification, `${field}.classification`),
    summary: text(entry.summary, `${field}.summary`),
    detail: text(entry.detail, `${field}.detail`),
    source: text(entry.source, `${field}.source`),
    blocks_use: boolean(entry.blocks_use, `${field}.blocks_use`),
    priority_score: number(entry.priority_score, `${field}.priority_score`),
    evidence_fingerprints: list(
      entry.evidence_fingerprints,
      `${field}.evidence_fingerprints`
    ).map((item, fingerprintIndex) =>
      fingerprint(item, `${field}.evidence_fingerprints[${fingerprintIndex}]`)
    ),
  }
}

function parseInterventionGuard(
  value: unknown
): StabilityIntelligenceInterventionGuard {
  const field = 'unknown_intervention'
  const entry = record(value, field)

  return {
    status: text(entry.status, `${field}.status`),
    claim_kind: text(entry.claim_kind, `${field}.claim_kind`),
    decision: text(entry.decision, `${field}.decision`),
    confidence_score: ratio(
      entry.confidence_score,
      `${field}.confidence_score`
    ),
    evidence_count: whole(entry.evidence_count, `${field}.evidence_count`),
    unknown_evidence_count: whole(
      entry.unknown_evidence_count,
      `${field}.unknown_evidence_count`
    ),
    estimated_evidence_count: whole(
      entry.estimated_evidence_count,
      `${field}.estimated_evidence_count`
    ),
    evidence_completeness: ratio(
      entry.evidence_completeness,
      `${field}.evidence_completeness`
    ),
    input_fingerprint: fingerprint(
      entry.input_fingerprint,
      `${field}.input_fingerprint`
    ),
  }
}

function parseScopeEnforcement(
  value: unknown
): StabilityIntelligenceScopeEnforcement {
  const field = 'scope_enforcement'
  const entry = record(value, field)

  return {
    status: text(entry.status, `${field}.status`),
    decision: text(entry.decision, `${field}.decision`),
    claim_count: whole(entry.claim_count, `${field}.claim_count`),
    allowed_count: whole(entry.allowed_count, `${field}.allowed_count`),
    limited_count: whole(entry.limited_count, `${field}.limited_count`),
    blocked_count: whole(entry.blocked_count, `${field}.blocked_count`),
    violations: list(entry.violations, `${field}.violations`).map(
      parseScopeViolation
    ),
    input_fingerprint: fingerprint(
      entry.input_fingerprint,
      `${field}.input_fingerprint`
    ),
  }
}

function parseScopeViolation(
  value: unknown,
  index: number
): StabilityIntelligenceScopeViolation {
  const field = `scope_enforcement.violations[${index}]`
  const entry = record(value, field)

  return {
    code: text(entry.code, `${field}.code`),
    claim_code: text(entry.claim_code, `${field}.claim_code`),
    message: text(entry.message, `${field}.message`),
    blocking: boolean(entry.blocking, `${field}.blocking`),
  }
}

function orderedTimestamps(value: string[], field: string): string[] {
  if (
    value.length < minimumAsOfTimeCount ||
    value.length > maximumAsOfTimeCount
  ) {
    throw new APIRequestError(
      'Stability Intelligence requires two to eight analytical timestamps.'
    )
  }

  const result = value.map((item, index) =>
    timestamp(item, `${field}[${index}]`)
  )
  for (let index = 1; index < result.length; index += 1) {
    if (Date.parse(result[index]) <= Date.parse(result[index - 1])) {
      throw new APIRequestError(
        'Stability Intelligence timestamps must be unique and strictly increasing.'
      )
    }
  }

  return result
}

function timestampsEqual(left: string[], right: string[]): boolean {
  return (
    left.length === right.length &&
    left.every((value, index) => Date.parse(value) === Date.parse(right[index]))
  )
}

function record(value: unknown, field: string): Record<string, unknown> {
  if (typeof value !== 'object' || value === null || Array.isArray(value)) {
    invalid(`${field} must be an object.`)
  }
  return value as Record<string, unknown>
}

function list(value: unknown, field: string): unknown[] {
  if (!Array.isArray(value)) invalid(`${field} must be an array.`)
  return value
}

function text(value: unknown, field: string): string {
  if (typeof value !== 'string' || value.trim() === '') {
    invalid(`${field} must be a non-empty string.`)
  }
  return value
}

function optionalText(value: unknown, field: string): string | undefined {
  if (value == null || value === '') return undefined
  if (typeof value !== 'string') invalid(`${field} must be a string.`)
  return value
}

function number(value: unknown, field: string): number {
  if (typeof value !== 'number' || !Number.isFinite(value)) {
    invalid(`${field} must be finite.`)
  }
  return value
}

function boolean(value: unknown, field: string): boolean {
  if (typeof value !== 'boolean') invalid(`${field} must be boolean.`)
  return value
}

function nonNegative(value: unknown, field: string): number {
  const result = number(value, field)
  if (result < 0) invalid(`${field} must not be negative.`)
  return result
}

function whole(value: unknown, field: string): number {
  const result = nonNegative(value, field)
  if (!Number.isInteger(result)) invalid(`${field} must be a whole number.`)
  return result
}

function positiveWhole(value: unknown, field: string): number {
  const result = whole(value, field)
  if (result < 1) invalid(`${field} must be positive.`)
  return result
}

function ratio(value: unknown, field: string): number {
  const result = number(value, field)
  if (result < 0 || result > 1) {
    invalid(`${field} must be between zero and one.`)
  }
  return result
}

function timestamp(value: unknown, field: string): string {
  const result = text(value, field)
  if (Number.isNaN(Date.parse(result))) {
    invalid(`${field} must be a valid timestamp.`)
  }
  return result
}

function uuid(value: string, field: string): string {
  const result = value.trim().toLowerCase()
  if (
    !/^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/.test(
      result
    )
  ) {
    invalid(`${field} must be a valid UUID.`)
  }
  return result
}

function fingerprint(value: unknown, field: string): string {
  const result = text(value, field)
  if (!/^sha256:[0-9a-f]{64}$/.test(result)) {
    invalid(`${field} must be a SHA-256 fingerprint.`)
  }
  return result
}

function invalid(message: string): never {
  throw new APIRequestError(
    `The Stability Intelligence response is invalid: ${message}`
  )
}
