import type { TrafficAircraft } from '../../types/traffic'

export type VerticalMotionStatus =
  | 'climbing'
  | 'descending'
  | 'level'
  | 'ground'
  | 'unavailable'

export const defaultVerticalMotionDeadbandMPS = 0.5

export interface VerticalMotionEvidence {
  status: VerticalMotionStatus
  label: string
  verticalRateMPS: number | null
  verticalRateFeetPerMinute: number | null
  displayRate: string
  description: string
}

export function buildVerticalMotionEvidence(
  aircraft: TrafficAircraft,
  deadbandMPS = defaultVerticalMotionDeadbandMPS
): VerticalMotionEvidence {
  const verticalRateMPS = finiteNumberOrNull(aircraft.vertical_rate_mps)
  const validDeadband =
    Number.isFinite(deadbandMPS) && deadbandMPS >= 0
      ? deadbandMPS
      : defaultVerticalMotionDeadbandMPS

  if (aircraft.on_ground) {
    return evidence(
      'ground',
      'On ground',
      verticalRateMPS,
      'Ground state is observed; vertical-rate evidence is shown separately when available.'
    )
  }

  if (verticalRateMPS === null) {
    return evidence(
      'unavailable',
      'Unavailable',
      null,
      'No usable vertical-rate evidence is available for this observation.'
    )
  }

  if (verticalRateMPS > validDeadband) {
    return evidence(
      'climbing',
      '↑ Climbing',
      verticalRateMPS,
      'Positive observed vertical rate exceeds the presentation deadband.'
    )
  }

  if (verticalRateMPS < -validDeadband) {
    return evidence(
      'descending',
      '↓ Descending',
      verticalRateMPS,
      'Negative observed vertical rate exceeds the presentation deadband.'
    )
  }

  return evidence(
    'level',
    'Level',
    verticalRateMPS,
    'Observed vertical rate is within the ±' +
      validDeadband.toFixed(1) +
      ' m/s presentation deadband.'
  )
}

function evidence(
  status: VerticalMotionStatus,
  label: string,
  verticalRateMPS: number | null,
  description: string
): VerticalMotionEvidence {
  const verticalRateFeetPerMinute =
    verticalRateMPS === null ? null : verticalRateMPS * 196.8503937007874

  return {
    status,
    label,
    verticalRateMPS,
    verticalRateFeetPerMinute,
    displayRate: formatVerticalRate(verticalRateMPS, verticalRateFeetPerMinute),
    description,
  }
}

export function formatVerticalRate(
  verticalRateMPS: number | null,
  verticalRateFeetPerMinute =
    verticalRateMPS === null ? null : verticalRateMPS * 196.8503937007874
): string {
  if (verticalRateMPS === null || verticalRateFeetPerMinute === null) {
    return 'Unavailable'
  }

  const mpsPrefix = verticalRateMPS > 0 ? '+' : ''
  const fpmRounded = Math.round(verticalRateFeetPerMinute)
  const fpmPrefix = fpmRounded > 0 ? '+' : ''

  return [mpsPrefix, verticalRateMPS.toFixed(1), ' m/s (', fpmPrefix, fpmRounded.toLocaleString('en-US'), ' ft/min)'].join('')
}

function finiteNumberOrNull(value: unknown): number | null {
  return typeof value === 'number' && Number.isFinite(value) ? value : null
}
