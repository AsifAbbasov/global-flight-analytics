// FRONTEND_APPLICATION_SHELL_V1
// FRONTEND_UNIFIED_AIRPORT_ANALYTICS_WORKSPACE_V1
// FRONTEND_HISTORICAL_ANALYTICS_COMPARISON_V1
// FRONTEND_PRODUCT_HARDENING_V1
// FRONTEND_MAP_FIRST_PRODUCT_SHELL_V1
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
  { href: '#live-traffic', label: 'Live workspace' },
  { href: '#overview', label: 'Overview' },
  { href: '#airport-intelligence', label: 'Airport Intelligence' },
  { href: '#historical-analytics', label: 'Historical Analytics' },
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
    <div
      id='top'
      className='relative min-h-screen overflow-x-hidden bg-[#111315] text-slate-100'
    >
      <a className='skip-link' href='#main-content'>
        Skip to main content
      </a>
      <a className='skip-link skip-link-secondary' href='#live-traffic'>
        Skip to live workspace
      </a>

      <header className='sticky top-0 z-50 border-b border-white/10 bg-[#202225]/95 shadow-lg shadow-black/20 backdrop-blur-xl'>
        <div className='flex min-h-14 items-center gap-3 px-3 sm:px-4'>
          <a href='#live-traffic' className='flex min-w-0 shrink-0 items-center gap-2.5'>
            <span
              aria-hidden='true'
              className='relative flex h-9 w-9 items-center justify-center rounded-md bg-amber-400 text-[#17191b] shadow-sm shadow-black/30'
            >
              <span className='absolute h-1.5 w-6 rounded-full bg-current' />
              <span className='absolute h-6 w-1.5 rounded-full bg-current' />
              <span className='absolute h-3 w-3 rotate-45 border-b-2 border-r-2 border-current' />
            </span>
            <span className='hidden min-w-0 sm:block'>
              <span className='block truncate text-sm font-bold tracking-tight text-white'>
                Global Flight Analytics
              </span>
              <span className='block truncate text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-400'>
                Open aviation intelligence
              </span>
            </span>
            <span className='text-sm font-bold text-white sm:hidden'>GFA</span>
          </a>

          <nav
            aria-label='Primary navigation'
            className='ml-2 hidden min-w-0 flex-1 items-center gap-1 lg:flex'
          >
            {navigationItems.slice(0, 4).map(item => (
              <a
                key={item.href}
                href={item.href}
                className='rounded-md px-3 py-2 text-xs font-semibold text-slate-300 transition hover:bg-white/10 hover:text-white'
              >
                {item.label}
              </a>
            ))}
          </nav>

          <div className='ml-auto hidden items-center gap-2 md:flex'>
            <div className='flex items-center gap-2 rounded-full border border-white/10 bg-black/20 px-3 py-1.5'>
              <span
                aria-hidden='true'
                className={`h-2 w-2 rounded-full ${statusDotClassName(initialStatus.availability)}`}
              />
              <span className='text-xs font-semibold text-slate-200'>
                {initialStatus.label}
              </span>
            </div>
            <div className='rounded-full border border-white/10 bg-black/20 px-3 py-1.5 text-xs font-semibold text-slate-300'>
              {formatInteger(initialStatus.initialTrafficCount)} aircraft
            </div>
          </div>

          <details className='relative lg:hidden'>
            <summary className='flex min-h-10 cursor-pointer list-none items-center rounded-md border border-white/10 bg-white/5 px-3 py-2 text-xs font-semibold text-slate-200 transition hover:bg-white/10 [&::-webkit-details-marker]:hidden'>
              Navigate
            </summary>
            <nav
              aria-label='Mobile primary navigation'
              className='absolute right-0 top-[calc(100%+0.5rem)] z-20 w-64 rounded-lg border border-white/10 bg-[#202225]/98 p-2 shadow-2xl shadow-black/50'
            >
              {navigationItems.map(item => (
                <a
                  key={item.href}
                  href={item.href}
                  className='block min-h-11 rounded-md px-3 py-3 text-sm font-medium text-slate-300 transition hover:bg-white/10 hover:text-white'
                >
                  {item.label}
                </a>
              ))}
            </nav>
          </details>
        </div>
      </header>

      <main id='main-content' tabIndex={-1} className='relative outline-none'>
        <section
          aria-label='Application status'
          className='border-b border-white/10 bg-[#17191c]'
        >
          <h1 className='sr-only'>
            Observe aviation data. Understand its limits.
          </h1>
          <div className='grid md:grid-cols-2 xl:grid-cols-4'>
            <StatusCard
              availability={initialStatus.availability}
              label={initialStatus.label}
              summary={initialStatus.summary}
            />
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
              eyebrow='Data boundary'
              value='Research only'
              description='Open observations and explainable analytics; not operational guidance.'
            />
          </div>
        </section>

        <div className='px-3 pb-12 sm:px-4'>{children}</div>

        <section
          id='research-scope'
          className='scroll-mt-20 border-y border-white/10 bg-[#17191c]'
        >
          <div className='mx-auto max-w-[1600px] px-4 py-12 sm:px-8 sm:py-16'>
            <div className='max-w-3xl'>
              <p className='text-xs font-semibold uppercase tracking-[0.22em] text-amber-300'>
                Responsible analytical scope
              </p>
              <h2 className='mt-3 text-2xl font-semibold tracking-tight text-white sm:text-3xl'>
                Evidence-first aviation intelligence
              </h2>
              <p className='mt-3 text-sm leading-6 text-slate-400 sm:text-base'>
                GFA separates observed state, inferred context and analytical
                confidence so a polished interface never implies evidence that
                the source data does not contain.
              </p>
            </div>

            <div className='mt-8 grid gap-3 lg:grid-cols-3'>
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

      <footer className='relative border-t border-white/10 bg-[#111315]'>
        <div className='mx-auto flex max-w-[1600px] flex-col gap-2 px-4 py-6 text-xs text-slate-500 sm:flex-row sm:items-center sm:justify-between sm:px-8'>
          <p>Global Flight Analytics · Open aviation research platform</p>
          <p>Research and visualization only · Not operational guidance</p>
        </div>
      </footer>
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
    <article className='border-b border-white/10 px-4 py-3 md:border-r xl:border-b-0'>
      <div className='flex items-center gap-2'>
        <span
          aria-hidden='true'
          className={`h-2 w-2 rounded-full ${statusDotClassName(availability)}`}
        />
        <p className='text-[10px] font-bold uppercase tracking-[0.16em] text-slate-500'>
          Startup status
        </p>
      </div>
      <p className='mt-1.5 text-sm font-semibold text-white'>{label}</p>
      <p className='mt-1 line-clamp-1 text-[11px] leading-5 text-slate-500'>{summary}</p>
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
    <article className='border-b border-white/10 px-4 py-3 md:border-r md:[&:nth-child(2)]:border-r-0 xl:border-b-0 xl:[&:nth-child(2)]:border-r xl:last:border-r-0'>
      <p className='text-[10px] font-bold uppercase tracking-[0.16em] text-slate-500'>
        {eyebrow}
      </p>
      <p className='mt-1.5 text-sm font-semibold text-white'>{value}</p>
      <p className='mt-1 line-clamp-1 text-[11px] leading-5 text-slate-500'>
        {description}
      </p>
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
    <article className='rounded-lg border border-white/10 bg-[#202225] p-5 shadow-lg shadow-black/10'>
      <p className='font-mono text-xs text-amber-300'>{index}</p>
      <h3 className='mt-3 text-base font-semibold text-white'>{title}</h3>
      <p className='mt-2 text-sm leading-6 text-slate-400'>{description}</p>
    </article>
  )
}

function statusDotClassName(
  availability: InitialSnapshotAvailability
): string {
  switch (availability) {
    case 'ready':
      return 'bg-emerald-300 shadow-[0_0_12px_rgba(110,231,183,0.8)]'
    case 'degraded':
      return 'bg-amber-300 shadow-[0_0_12px_rgba(252,211,77,0.7)]'
    case 'unavailable':
      return 'bg-rose-300 shadow-[0_0_12px_rgba(253,164,175,0.7)]'
  }
}

function formatInteger(value: number): string {
  return new Intl.NumberFormat().format(value)
}
