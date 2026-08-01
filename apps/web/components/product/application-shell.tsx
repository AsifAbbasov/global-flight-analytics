// FRONTEND_APPLICATION_SHELL_V1
// FRONTEND_UNIFIED_AIRPORT_ANALYTICS_WORKSPACE_V1
import type { ReactNode } from 'react'

import {
  buildApplicationInitialStatus,
  type InitialSnapshotAvailability,
} from '@/lib/product/application-status'

interface ApplicationShellProps {
  children: ReactNode
  initialTrafficCount: number
  regionCount: number
  trafficUnavailable: boolean
  regionsUnavailable: boolean
}

const navigationItems = [
  { href: '#overview', label: 'Overview' },
  { href: '#airport-intelligence', label: 'Airport Intelligence' },
  { href: '#live-traffic', label: 'Live workspace' },
  { href: '#research-scope', label: 'Research scope' },
] as const

export function ApplicationShell({
  children,
  initialTrafficCount,
  regionCount,
  trafficUnavailable,
  regionsUnavailable,
}: ApplicationShellProps) {
  const initialStatus = buildApplicationInitialStatus({
    initialTrafficCount,
    regionCount,
    trafficUnavailable,
    regionsUnavailable,
  })

  return (
    <div className='relative min-h-screen overflow-hidden bg-slate-950 text-slate-100'>
      <div
        aria-hidden='true'
        className='pointer-events-none absolute inset-x-0 top-0 h-[46rem] bg-[radial-gradient(circle_at_15%_10%,rgba(14,165,233,0.16),transparent_34%),radial-gradient(circle_at_86%_4%,rgba(16,185,129,0.12),transparent_30%)]'
      />

      <header className='sticky top-0 z-50 border-b border-white/10 bg-slate-950/85 backdrop-blur-xl'>
        <div className='mx-auto flex max-w-[1600px] items-center justify-between gap-5 px-4 py-3 sm:px-8'>
          <a href='#top' className='flex min-w-0 items-center gap-3'>
            <span
              aria-hidden='true'
              className='relative flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-sky-400/30 bg-sky-400/10'
            >
              <span className='h-2.5 w-2.5 rounded-full bg-sky-300 shadow-[0_0_18px_rgba(125,211,252,0.9)]' />
              <span className='absolute h-px w-6 rotate-[-28deg] bg-gradient-to-r from-transparent via-sky-300 to-transparent' />
            </span>
            <span className='min-w-0'>
              <span className='block truncate text-sm font-semibold tracking-wide text-white'>
                Global Flight Analytics
              </span>
              <span className='block truncate text-[11px] uppercase tracking-[0.2em] text-slate-500'>
                Open aviation research
              </span>
            </span>
          </a>

          <nav
            aria-label='Primary navigation'
            className='hidden items-center gap-1 lg:flex'
          >
            {navigationItems.map(item => (
              <a
                key={item.href}
                href={item.href}
                className='rounded-lg px-3 py-2 text-sm text-slate-400 transition hover:bg-white/5 hover:text-white'
              >
                {item.label}
              </a>
            ))}
          </nav>

          <a
            href='#live-traffic'
            className='shrink-0 rounded-lg border border-sky-400/35 bg-sky-400/10 px-3 py-2 text-xs font-semibold text-sky-100 transition hover:border-sky-300/60 hover:bg-sky-400/15 sm:px-4 sm:text-sm'
          >
            Open workspace
          </a>
        </div>
      </header>

      <main id='top' className='relative'>
        <section className='mx-auto max-w-[1600px] px-4 pb-6 pt-14 sm:px-8 sm:pt-20'>
          <div className='grid items-end gap-10 xl:grid-cols-[minmax(0,1fr)_420px]'>
            <div>
              <p className='text-xs font-semibold uppercase tracking-[0.28em] text-sky-300'>
                Evidence-aware air traffic intelligence
              </p>
              <h1 className='mt-5 max-w-5xl text-4xl font-semibold tracking-[-0.04em] text-white sm:text-6xl xl:text-7xl'>
                Observe aviation data.
                <span className='block text-slate-400'>Understand its limits.</span>
              </h1>
              <p className='mt-6 max-w-3xl text-base leading-8 text-slate-400 sm:text-lg'>
                A research and visualization platform for open aviation data,
                regional traffic, trajectories and explainable analytical
                signals. Every published metric keeps confidence, eligibility
                and limitations visible.
              </p>

              <div className='mt-8 flex flex-wrap gap-3'>
                <a
                  href='#live-traffic'
                  className='rounded-xl bg-sky-300 px-5 py-3 text-sm font-semibold text-slate-950 transition hover:bg-sky-200'
                >
                  Explore live traffic
                </a>
                <a
                  href='#research-scope'
                  className='rounded-xl border border-white/15 bg-white/5 px-5 py-3 text-sm font-semibold text-slate-200 transition hover:border-white/25 hover:bg-white/10'
                >
                  Review research scope
                </a>
              </div>
            </div>

            <aside className='rounded-2xl border border-white/10 bg-slate-900/60 p-5 shadow-2xl shadow-black/20 backdrop-blur'>
              <p className='text-xs font-semibold uppercase tracking-[0.2em] text-emerald-300'>
                Platform boundary
              </p>
              <h2 className='mt-3 text-xl font-semibold text-white'>
                Research intelligence, not operational control
              </h2>
              <p className='mt-3 text-sm leading-6 text-slate-400'>
                The platform interprets open observations for research. It is
                not air traffic control, navigation guidance, ticketing or an
                authoritative flight-status service.
              </p>
              <dl className='mt-5 grid gap-3 text-sm'>
                <BoundaryItem label='Data path' value='Next.js → Go API → PostgreSQL' />
                <BoundaryItem label='Primary mode' value='Live regional observation' />
                <BoundaryItem label='Claims' value='Confidence-bounded and explainable' />
              </dl>
            </aside>
          </div>

          <div className='mt-10 grid overflow-hidden rounded-2xl border border-white/10 bg-slate-900/55 shadow-2xl shadow-black/20 backdrop-blur md:grid-cols-2 xl:grid-cols-4'>
            <StatusCard availability={initialStatus.availability} label={initialStatus.label} summary={initialStatus.summary} />
            <MetricCard
              eyebrow='Initial aircraft'
              value={formatInteger(initialStatus.initialTrafficCount)}
              description='Server-rendered world snapshot before client refresh.'
            />
            <MetricCard
              eyebrow='Available views'
              value={formatInteger(initialStatus.regionCount)}
              description='World plus regional scopes available at startup.'
            />
            <MetricCard
              eyebrow='Architecture'
              value='Modular monolith'
              description='A bounded Go backend with a typed Next.js client.'
            />
          </div>
        </section>

        <div className='mx-auto max-w-[1600px] px-4 pb-20 sm:px-8'>
          {children}
        </div>

        <section
          id='research-scope'
          className='scroll-mt-28 border-y border-white/10 bg-slate-900/40'
        >
          <div className='mx-auto max-w-[1600px] px-4 py-16 sm:px-8 sm:py-20'>
            <div className='max-w-3xl'>
              <p className='text-xs font-semibold uppercase tracking-[0.24em] text-emerald-300'>
                Responsible analytical scope
              </p>
              <h2 className='mt-4 text-3xl font-semibold tracking-tight text-white sm:text-4xl'>
                Strong engineering without pretending the data knows more than it does
              </h2>
              <p className='mt-4 text-base leading-7 text-slate-400'>
                The product separates observation, inference and limitation so
                that polished visuals never hide uncertainty in the underlying
                evidence.
              </p>
            </div>

            <div className='mt-10 grid gap-4 lg:grid-cols-3'>
              <ScopeCard
                index='01'
                title='Research, not control'
                description='No operational navigation, air traffic control or safety-critical decisions are delegated to this interface.'
              />
              <ScopeCard
                index='02'
                title='Evidence before claims'
                description='Metrics expose eligibility, confidence, source windows and limitations instead of presenting every result as equally reliable.'
              />
              <ScopeCard
                index='03'
                title='Open-data constraints'
                description='Coverage gaps, provider delays and inferred route context remain visible parts of the product rather than hidden implementation details.'
              />
            </div>
          </div>
        </section>
      </main>

      <footer className='relative border-t border-white/10 bg-slate-950'>
        <div className='mx-auto flex max-w-[1600px] flex-col gap-3 px-4 py-8 text-sm text-slate-500 sm:flex-row sm:items-center sm:justify-between sm:px-8'>
          <p>Global Flight Analytics · Open aviation research platform</p>
          <p>Research and visualization only · Not operational guidance</p>
        </div>
      </footer>
    </div>
  )
}

function BoundaryItem({ label, value }: { label: string; value: string }) {
  return (
    <div className='flex items-start justify-between gap-4 border-t border-white/10 pt-3'>
      <dt className='text-slate-500'>{label}</dt>
      <dd className='text-right font-medium text-slate-200'>{value}</dd>
    </div>
  )
}

function StatusCard({
  availability,
  label,
  summary,
}: {
  availability: InitialSnapshotAvailability
  label: string
  summary: string
}) {
  return (
    <article className='border-b border-white/10 p-5 md:border-r xl:border-b-0'>
      <div className='flex items-center gap-2'>
        <span
          aria-hidden='true'
          className={`h-2.5 w-2.5 rounded-full ${statusDotClassName(availability)}`}
        />
        <p className='text-xs font-semibold uppercase tracking-[0.16em] text-slate-500'>
          Startup status
        </p>
      </div>
      <p className='mt-3 font-semibold text-white'>{label}</p>
      <p className='mt-2 text-xs leading-5 text-slate-500'>{summary}</p>
    </article>
  )
}

function MetricCard({
  eyebrow,
  value,
  description,
}: {
  eyebrow: string
  value: string
  description: string
}) {
  return (
    <article className='border-b border-white/10 p-5 md:border-r md:[&:nth-child(2)]:border-r-0 xl:border-b-0 xl:[&:nth-child(2)]:border-r xl:last:border-r-0'>
      <p className='text-xs font-semibold uppercase tracking-[0.16em] text-slate-500'>
        {eyebrow}
      </p>
      <p className='mt-3 text-xl font-semibold text-white'>{value}</p>
      <p className='mt-2 text-xs leading-5 text-slate-500'>{description}</p>
    </article>
  )
}

function ScopeCard({
  index,
  title,
  description,
}: {
  index: string
  title: string
  description: string
}) {
  return (
    <article className='rounded-2xl border border-white/10 bg-slate-950/70 p-6'>
      <p className='font-mono text-xs text-sky-300'>{index}</p>
      <h3 className='mt-5 text-lg font-semibold text-white'>{title}</h3>
      <p className='mt-3 text-sm leading-6 text-slate-400'>{description}</p>
    </article>
  )
}

function statusDotClassName(
  availability: InitialSnapshotAvailability
): string {
  switch (availability) {
    case 'ready':
      return 'bg-emerald-300 shadow-[0_0_16px_rgba(110,231,183,0.8)]'
    case 'degraded':
      return 'bg-amber-300 shadow-[0_0_16px_rgba(252,211,77,0.7)]'
    case 'unavailable':
      return 'bg-rose-300 shadow-[0_0_16px_rgba(253,164,175,0.7)]'
  }
}

function formatInteger(value: number): string {
  return new Intl.NumberFormat().format(value)
}
