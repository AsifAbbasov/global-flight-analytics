'use client'

import type { ReactNode } from 'react'
import { APIRequestError, getRequestErrorMessage } from '@/lib/api/client'
import type {
  StabilityIntelligenceFailure,
  StabilityIntelligenceResponse,
  StabilityIntelligenceScopeViolation,
  StabilityIntelligenceTransition,
} from '@/types/stability-intelligence'

interface Props {
  selectedICAO24: string | null
  trajectoryID: string | null
  asOfTimes: string[]
  result: StabilityIntelligenceResponse | undefined
  isPending: boolean
  isFetching: boolean
  error: Error | null
  onRetry: () => void
}

export function StabilityIntelligencePanel({
  selectedICAO24,
  trajectoryID,
  asOfTimes,
  result,
  isPending,
  isFetching,
  error,
  onRetry,
}: Props) {
  if (selectedICAO24 === null) return null

  const unavailable =
    error instanceof APIRequestError &&
    (error.status === 400 || error.status === 404 || error.status === 422)

  return (
    <aside
      className='rounded-xl border border-slate-700 bg-slate-950/95 p-5'
      aria-labelledby='stability-intelligence-title'
    >
      <div className='flex items-start justify-between gap-4'>
        <div>
          <p className='text-xs font-semibold uppercase tracking-[0.18em] text-emerald-300'>
            Forecast consistency evidence
          </p>
          <h3
            id='stability-intelligence-title'
            className='mt-2 text-lg font-semibold text-white'
          >
            Stability and Explainability
          </h3>
          <p className='mt-1 text-xs leading-5 text-slate-400'>
            Consistency is not accuracy. Explanations remain non-causal and
            research-only.
          </p>
        </div>
        {isFetching ? (
          <span className='text-xs text-sky-300'>Updating…</span>
        ) : null}
      </div>

      {result ? <Content result={result} /> : null}

      {trajectoryID === null && !error ? (
        <Message>
          Waiting for a persisted trajectory before requesting Stability
          Intelligence.
        </Message>
      ) : null}

      {trajectoryID !== null && asOfTimes.length < 2 && !error ? (
        <Message>
          The persisted trajectory does not yet span two analytical timestamps.
        </Message>
      ) : null}

      {trajectoryID !== null && asOfTimes.length >= 2 && isPending && !error ? (
        <p className='mt-4 text-sm leading-6 text-slate-400'>
          Comparing bounded forecast versions and evaluating explanation guards…
        </p>
      ) : null}

      {unavailable ? (
        <Message>
          Stability Intelligence is unavailable or denied for the current
          trajectory history.
        </Message>
      ) : null}

      {error && !unavailable ? (
        <div className='mt-4 rounded-lg border border-amber-400/30 bg-amber-400/10 p-3'>
          <p className='text-sm leading-6 text-amber-100'>
            {getRequestErrorMessage(error)}
          </p>
          <button
            type='button'
            onClick={onRetry}
            disabled={isFetching}
            className='mt-3 rounded-md border border-amber-300/40 px-3 py-1.5 text-sm font-medium text-amber-100 disabled:opacity-60'
          >
            Retry Stability Intelligence
          </button>
        </div>
      ) : null}
    </aside>
  )
}

function Content({ result }: { result: StabilityIntelligenceResponse }) {
  const analysis = result.forecast_analysis
  const metrics = analysis.metrics
  const confidence = result.propagated_confidence
  const failure = result.failure_explanation
  const intervention = result.unknown_intervention
  const scope = result.scope_enforcement

  return (
    <>
      <div className='mt-4 rounded-lg border border-slate-800 bg-slate-900/70 p-3'>
        <div className='flex flex-wrap items-center justify-between gap-3'>
          <Badge value={`${analysis.health} · ${analysis.trend}`} />
          <Badge value={`${confidence.level} · ${percent(confidence.score)}`} />
        </div>
        <div
          className='mt-3 h-2 overflow-hidden rounded-full bg-slate-800'
          role='progressbar'
          aria-label='Mean forecast stability score'
          aria-valuemin={0}
          aria-valuemax={100}
          aria-valuenow={Math.round(metrics.mean_stability_score * 100)}
        >
          <div
            className='h-full rounded-full bg-emerald-400'
            style={{ width: `${metrics.mean_stability_score * 100}%` }}
          />
        </div>
        <dl className='mt-3 grid grid-cols-2 gap-x-4 gap-y-3 text-sm'>
          <Detail label='Analysis status' value={analysis.status} />
          <Detail label='Latest level' value={metrics.latest_level} />
          <Detail label='Forecast versions' value={String(metrics.version_count)} />
          <Detail label='Transitions' value={String(metrics.transition_count)} />
          <Detail
            label='Stable share'
            value={percent(metrics.stable_transition_share)}
          />
          <Detail
            label='Comparable share'
            value={percent(metrics.comparable_transition_share)}
          />
          <Detail
            label='Mean shift'
            value={kilometers(metrics.mean_horizontal_shift_kilometers)}
          />
          <Detail
            label='Maximum shift'
            value={kilometers(metrics.maximum_horizontal_shift_kilometers)}
          />
        </dl>
        <p className='mt-3 text-xs leading-5 text-amber-100'>
          A stable forecast may still be inaccurate. A changed forecast may be
          an improvement.
        </p>
      </div>

      <div className='mt-3 rounded-lg border border-emerald-400/20 bg-emerald-400/5 p-3'>
        <h4 className='text-xs font-semibold uppercase tracking-wide text-emerald-200'>
          Propagated confidence
        </h4>
        <dl className='mt-3 grid grid-cols-2 gap-x-4 gap-y-3 text-sm'>
          <Detail label='Status' value={confidence.status} />
          <Detail label='Score' value={percent(confidence.score)} />
          <Detail label='Target node' value={confidence.target_node_id} />
          <Detail
            label='Limiting dependency'
            value={confidence.limiting_dependency_id ?? 'Not reported'}
          />
        </dl>
        <p className='mt-3 text-xs leading-5 text-slate-400'>
          Confidence is a bounded analytical score, not a calibrated
          probability.
        </p>
      </div>

      <TransitionList transitions={result.transitions} />
      <FailureExplanation
        failures={failure.failures}
        primaryCode={failure.primary_code}
        blockingCount={failure.blocking_count}
        warningCount={failure.warning_count}
        unknownCauseCount={failure.unknown_cause_count}
      />

      <div className='mt-3 rounded-lg border border-violet-400/20 bg-violet-400/5 p-3'>
        <h4 className='text-xs font-semibold uppercase tracking-wide text-violet-200'>
          Attribution and scope guards
        </h4>
        <dl className='mt-3 grid grid-cols-2 gap-x-4 gap-y-3 text-sm'>
          <Detail label='Claim kind' value={intervention.claim_kind} />
          <Detail label='Attribution decision' value={intervention.decision} />
          <Detail
            label='Evidence completeness'
            value={percent(intervention.evidence_completeness)}
          />
          <Detail
            label='Estimated evidence'
            value={`${intervention.estimated_evidence_count} / ${intervention.evidence_count}`}
          />
          <Detail label='Scope decision' value={scope.decision} />
          <Detail
            label='Allowed, limited, blocked'
            value={`${scope.allowed_count}, ${scope.limited_count}, ${scope.blocked_count}`}
          />
        </dl>
        <ViolationList violations={scope.violations} />
      </div>

      <TextList
        title='Declared scope guards'
        items={result.scope_guards}
        empty='No scope guards were reported.'
        warning
      />

      <p className='mt-4 break-all font-mono text-[11px] leading-5 text-slate-500'>
        {date(result.as_of_times[0])} —{' '}
        {date(result.as_of_times[result.as_of_times.length - 1])}
      </p>
      <p className='mt-2 break-all font-mono text-[11px] leading-5 text-slate-500'>
        {result.input_fingerprint}
      </p>
    </>
  )
}

function TransitionList({
  transitions,
}: {
  transitions: StabilityIntelligenceTransition[]
}) {
  return (
    <div className='mt-3 rounded-lg border border-slate-800 bg-slate-900/60 p-3'>
      <h4 className='text-xs font-semibold uppercase tracking-wide text-slate-300'>
        Forecast transitions
      </h4>
      {transitions.length ? (
        <ul className='mt-3 space-y-3'>
          {transitions.slice(-4).map((transition) => (
            <li
              key={`${transition.baseline_version_id}:${transition.candidate_version_id}`}
              className='rounded-md border border-slate-800 bg-slate-950/60 p-3'
            >
              <div className='flex flex-wrap items-center justify-between gap-2'>
                <span className='text-sm font-medium text-slate-200'>
                  {transition.level}
                </span>
                <span className='text-xs text-emerald-200'>
                  {percent(transition.score)}
                </span>
              </div>
              <dl className='mt-2 grid grid-cols-2 gap-x-4 gap-y-2 text-sm'>
                <Detail
                  label='Mean shift'
                  value={kilometers(
                    transition.metrics.mean_horizontal_shift_kilometers
                  )}
                />
                <Detail
                  label='Aligned points'
                  value={`${transition.metrics.aligned_point_count} · ${percent(
                    transition.metrics.aligned_point_share
                  )}`}
                />
                <Detail
                  label='Confidence delta'
                  value={signedPercent(
                    transition.metrics.aggregate_confidence_delta
                  )}
                />
                <Detail
                  label='Arrival shift'
                  value={
                    transition.metrics.arrival_comparable
                      ? seconds(transition.metrics.arrival_shift_seconds)
                      : 'Not comparable'
                  }
                />
              </dl>
            </li>
          ))}
        </ul>
      ) : (
        <p className='mt-2 text-sm text-slate-400'>
          No forecast transitions were reported.
        </p>
      )}
    </div>
  )
}

function FailureExplanation({
  failures,
  primaryCode,
  blockingCount,
  warningCount,
  unknownCauseCount,
}: {
  failures: StabilityIntelligenceFailure[]
  primaryCode: string
  blockingCount: number
  warningCount: number
  unknownCauseCount: number
}) {
  return (
    <div className='mt-3 rounded-lg border border-amber-400/25 bg-amber-400/5 p-3'>
      <h4 className='text-xs font-semibold uppercase tracking-wide text-amber-200'>
        Failure and limitation explanation
      </h4>
      <dl className='mt-3 grid grid-cols-2 gap-x-4 gap-y-3 text-sm'>
        <Detail label='Primary code' value={primaryCode} />
        <Detail
          label='Blocking, warning, unknown'
          value={`${blockingCount}, ${warningCount}, ${unknownCauseCount}`}
        />
      </dl>
      {failures.length ? (
        <ul className='mt-3 space-y-2 text-sm leading-5 text-amber-100'>
          {failures.slice(0, 6).map((failure) => (
            <li key={`${failure.rank}:${failure.code}`}>
              <span className='font-medium'>{failure.summary}</span>
              <span className='text-amber-200/80'> — {failure.detail}</span>
            </li>
          ))}
        </ul>
      ) : (
        <p className='mt-2 text-sm text-slate-400'>
          No failure conditions were reported.
        </p>
      )}
      <p className='mt-3 text-xs leading-5 text-amber-100'>
        A detected limitation is not proof of pilot intent, air traffic control
        instruction, or exact maneuver cause.
      </p>
    </div>
  )
}

function ViolationList({
  violations,
}: {
  violations: StabilityIntelligenceScopeViolation[]
}) {
  if (!violations.length) {
    return (
      <p className='mt-3 text-sm text-slate-400'>
        No scope violations were reported.
      </p>
    )
  }

  return (
    <ul className='mt-3 space-y-2 text-sm leading-5 text-violet-100'>
      {violations.slice(0, 6).map((violation) => (
        <li key={`${violation.code}:${violation.claim_code}`}>
          {violation.message}
          {violation.blocking ? ' This claim was blocked.' : ''}
        </li>
      ))}
    </ul>
  )
}

function TextList({
  title,
  items,
  empty,
  warning = false,
}: {
  title: string
  items: string[]
  empty: string
  warning?: boolean
}) {
  return (
    <div
      className={`mt-4 rounded-lg border p-3 ${
        warning
          ? 'border-amber-400/25 bg-amber-400/5'
          : 'border-slate-800 bg-slate-900/60'
      }`}
    >
      <h4
        className={`text-xs font-semibold uppercase tracking-wide ${
          warning ? 'text-amber-200' : 'text-slate-300'
        }`}
      >
        {title}
      </h4>
      {items.length ? (
        <ul
          className={`mt-2 space-y-2 text-sm leading-5 ${
            warning ? 'text-amber-100' : 'text-slate-400'
          }`}
        >
          {items.slice(0, 8).map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      ) : (
        <p className='mt-2 text-sm text-slate-400'>{empty}</p>
      )}
    </div>
  )
}

function Message({ children }: { children: ReactNode }) {
  return (
    <p className='mt-4 rounded-lg border border-slate-800 bg-slate-900/60 p-3 text-sm leading-6 text-slate-400'>
      {children}
    </p>
  )
}

function Badge({ value }: { value: string }) {
  return (
    <span className='rounded-full border border-emerald-400/40 bg-emerald-400/10 px-2.5 py-1 text-xs font-semibold uppercase tracking-wide text-emerald-200'>
      {value}
    </span>
  )
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className='text-xs uppercase tracking-wide text-slate-500'>{label}</dt>
      <dd className='mt-1 break-words text-slate-200'>
        {value.trim() || 'Unknown'}
      </dd>
    </div>
  )
}

function percent(value: number) {
  return `${Math.round(value * 100)}%`
}

function signedPercent(value: number) {
  const sign = value > 0 ? '+' : ''
  return `${sign}${new Intl.NumberFormat(undefined, {
    maximumFractionDigits: 1,
  }).format(value * 100)}%`
}

function kilometers(value: number) {
  return `${new Intl.NumberFormat(undefined, {
    maximumFractionDigits: 2,
  }).format(value)} km`
}

function seconds(value: number) {
  const sign = value > 0 ? '+' : ''
  return `${sign}${new Intl.NumberFormat(undefined, {
    maximumFractionDigits: 0,
  }).format(value)} seconds`
}

function date(value: string) {
  const result = new Date(value)
  return Number.isNaN(result.getTime()) ? 'Unknown' : result.toLocaleString()
}
