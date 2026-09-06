'use client'

import {
  buildFlightReplayEvidenceQualityProfile,
  type FlightReplayEvidenceQualityProfile,
} from '@/lib/replay/flight-replay-evidence-quality-model'
import type { FlightReplay } from '@/types/flight-replay'

export function FlightReplayEvidenceQuality({ replay }: { replay: FlightReplay }) {
  const profile = buildFlightReplayEvidenceQualityProfile(replay)

  return (
    <div
      className='mt-4 rounded-lg border border-violet-400/15 bg-slate-950/45 px-3 py-3'
      aria-label='Replay evidence quality profile'
      data-flight-replay-evidence-quality='descriptive-only'
      data-flight-replay-evidence-score='none'
    >
      <div className='flex flex-wrap items-start justify-between gap-3'>
        <div>
          <p className='text-[10px] font-bold uppercase tracking-[0.12em] text-slate-500'>
            Replay evidence quality
          </p>
          <p className='mt-1 max-w-3xl text-[11px] leading-relaxed text-slate-500'>
            Descriptive sampling evidence only. No universal good/bad score is assigned, and gaps
            remain unobserved rather than being reconstructed.
          </p>
        </div>
        <span className='rounded-full border border-violet-400/25 bg-violet-400/10 px-2 py-0.5 text-[9px] font-bold uppercase tracking-[0.1em] text-violet-200'>
          No synthetic score
        </span>
      </div>

      <div className='mt-3 grid gap-2 sm:grid-cols-2 lg:grid-cols-4'>
        <EvidenceDatum label='Observations' value={String(profile.sampleCount)} />
        <EvidenceDatum label='Sampling intervals' value={String(profile.intervalCount)} />
        <EvidenceDatum
          label='Mean gap'
          value={formatOptionalDuration(profile.meanGapSeconds)}
        />
        <EvidenceDatum
          label='Median gap'
          value={formatOptionalDuration(profile.medianGapSeconds)}
        />
        <EvidenceDatum
          label='P90 gap'
          value={formatOptionalDuration(profile.p90GapSeconds)}
        />
        <EvidenceDatum
          label='Largest gap'
          value={formatOptionalDuration(profile.largestGapSeconds)}
        />
        <EvidenceDatum
          label='Largest-gap share'
          value={formatOptionalPercent(profile.largestGapSharePercent, ' of observed span')}
        />
        <EvidenceDatum
          label='Observation density'
          value={formatObservationDensity(profile.observationDensityPerHour)}
        />
        <EvidenceDatum
          label='Altitude evidence'
          value={`${profile.altitudeCoveragePercent}%`}
        />
      </div>

      <LargestGapEvidence profile={profile} />

      <p className='mt-3 text-[11px] leading-relaxed text-slate-500'>
        These values describe the persisted sample set only. They are not calibrated aviation
        quality grades and must not be interpreted as ATC-, navigation- or safety-grade coverage.
      </p>
    </div>
  )
}

function LargestGapEvidence({
  profile,
}: {
  profile: FlightReplayEvidenceQualityProfile
}) {
  if (
    profile.largestGapStartObservedAt === null ||
    profile.largestGapEndObservedAt === null ||
    profile.largestGapSeconds === null
  ) {
    return (
      <p className='mt-3 rounded-md border border-white/10 bg-[#111315] px-3 py-2 text-xs text-slate-400'>
        Only one persisted observation is available, so sampling density and gap distribution
        cannot be measured.
      </p>
    )
  }

  return (
    <div className='mt-3 rounded-md border border-amber-300/15 bg-amber-300/5 px-3 py-2'>
      <p className='text-[9px] font-bold uppercase tracking-[0.12em] text-amber-200/70'>
        Largest unobserved interval
      </p>
      <p className='mt-1 font-mono text-[11px] text-amber-100'>
        {formatTimestamp(profile.largestGapStartObservedAt)} →{' '}
        {formatTimestamp(profile.largestGapEndObservedAt)} ·{' '}
        {formatDuration(profile.largestGapSeconds)}
      </p>
      <p className='mt-1 text-[11px] leading-relaxed text-slate-500'>
        No persisted position exists between these endpoint observations.
      </p>
    </div>
  )
}

function EvidenceDatum({ label, value }: { label: string; value: string }) {
  return (
    <div className='rounded-lg border border-white/10 bg-[#111315] px-2.5 py-2'>
      <p className='text-[9px] font-bold uppercase tracking-[0.12em] text-slate-600'>{label}</p>
      <p className='mt-1 break-words font-mono text-[11px] text-slate-300'>{value}</p>
    </div>
  )
}

function formatOptionalDuration(value: number | null): string {
  return value === null ? 'Unavailable' : formatDuration(value)
}

function formatOptionalPercent(value: number | null, suffix: string): string {
  return value === null ? 'Unavailable' : `${value}%${suffix}`
}

function formatObservationDensity(value: number | null): string {
  return value === null ? 'Unavailable' : `${value.toFixed(1)} observations/hour`
}

function formatTimestamp(value: string): string {
  const timestamp = Date.parse(value)
  if (Number.isNaN(timestamp)) return 'Unavailable'
  return new Date(timestamp).toISOString().replace('.000Z', 'Z')
}

function formatDuration(value: number): string {
  const totalSeconds = Math.max(0, Math.round(value))
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60

  if (hours > 0) return `${hours}h ${minutes}m ${seconds}s`
  if (minutes > 0) return `${minutes}m ${seconds}s`
  return `${seconds}s`
}