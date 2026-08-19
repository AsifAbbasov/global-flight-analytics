// FRONTEND_APPLICATION_SHELL_V1
// FRONTEND_PRODUCT_HARDENING_V1
// FRONTEND_FLIGHT_TRACKER_CLEAN_REBUILD_V1
// FRONTEND_AIRCRAFT_ONLY_SCOPE_V1
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
    <div id='top' className='gfa-flight-app'>
      <a className='skip-link' href='#main-content'>
        Skip to main content
      </a>

      <header className='gfa-flight-header'>
        <a
          href='#live-traffic'
          className='gfa-flight-brand'
          aria-label='Global Flight Analytics live flight tracker'
        >
          <span className='gfa-flight-brand-mark' aria-hidden='true'>
            <svg viewBox='0 0 24 24' focusable='false'>
              <path d='M12 2.25c.61 0 .98.45.98 1.05v5.58l6.62 3.59v1.67l-6.62-1.8v4.8l2 1.65v1.11L12 19l-2.98.9v-1.11l2-1.65v-4.8l-6.62 1.8v-1.67l6.62-3.59V3.3c0-.6.37-1.05.98-1.05Z' />
            </svg>
          </span>
          <span className='gfa-flight-brand-copy'>
            <strong>GFA</strong>
            <span>Flight tracker</span>
          </span>
        </a>

        <a
          href='#traffic-workspace-panel'
          className='gfa-flight-search'
          aria-label='Open aircraft search'
        >
          <span className='gfa-flight-search-icon' aria-hidden='true'>⌕</span>
          <span className='gfa-flight-search-copy'>
            Search aircraft or ICAO24
          </span>
          <kbd>/</kbd>
        </a>

        <nav className='gfa-flight-header-tools' aria-label='Tracker tools'>
          <a href='#overview' aria-label='Open analytics overview' title='Analytics'>
            <span aria-hidden='true'>▦</span>
          </a>
          <a href='#historical-analytics' aria-label='Open historical analytics' title='History'>
            <span aria-hidden='true'>◷</span>
          </a>
          <a href='#research-scope' aria-label='Open research scope' title='About'>
            <span aria-hidden='true'>i</span>
          </a>
        </nav>

        <details className='gfa-mobile-menu'>
          <summary>Navigate</summary>
          <nav
            className='gfa-mobile-navigation'
            aria-label='Mobile primary navigation'
          >
            <a href='#live-traffic'>Live map</a>
            <a href='#overview'>Analytics</a>
            <a href='#historical-analytics'>History</a>
            <a href='#research-scope'>About</a>
          </nav>
        </details>

        <div className='gfa-flight-header-status' aria-label='Initial platform status'>
          <span className='gfa-flight-header-count'>
            <strong>{formatInteger(initialStatus.initialTrafficCount)}</strong>
            <span>aircraft</span>
          </span>
          <StatusPill
            availability={initialStatus.availability}
            label={initialStatus.label}
          />
        </div>
      </header>

      <main id='main-content' tabIndex={-1} className='gfa-flight-main'>
        {children}

        <section id='research-scope' className='gfa-overlay-panel gfa-about-panel'>
          <header className='gfa-overlay-header'>
            <div>
              <p>Research boundary</p>
              <h2>Global Flight Analytics</h2>
            </div>
            <a href='#live-traffic' aria-label='Close research scope'>×</a>
          </header>

          <div className='gfa-about-copy'>
            <p>
              Open aviation observations, visualization and explainable analytics.
              This product is not air traffic control, navigation guidance,
              ticketing or an authoritative operational flight-status service.
            </p>
          </div>

          <div className='gfa-about-grid'>
            <BoundaryCard
              label='Evidence'
              value='Observation and inference stay visibly separate.'
            />
            <BoundaryCard
              label='Confidence'
              value='Eligibility and limitations stay attached to analytical claims.'
            />
            <BoundaryCard
              label='Architecture'
              value='Next.js → Go API → PostgreSQL modular monolith.'
            />
          </div>
        </section>
      </main>
    </div>
  )
}

function StatusPill({
  availability,
  label,
}: {
  availability: InitialSnapshotAvailability
  label: string
}) {
  return (
    <span className={`gfa-flight-health ${statusPillClassName(availability)}`}>
      <span aria-hidden='true' />
      {label}
    </span>
  )
}

function BoundaryCard({ label, value }: { label: string; value: string }) {
  return (
    <article className='gfa-about-card'>
      <p>{label}</p>
      <span>{value}</span>
    </article>
  )
}

function statusPillClassName(
  availability: InitialSnapshotAvailability
): string {
  switch (availability) {
    case 'ready':
      return 'is-ready'
    case 'degraded':
      return 'is-degraded'
    case 'unavailable':
      return 'is-unavailable'
  }
}

function formatInteger(value: number): string {
  return new Intl.NumberFormat().format(value)
}
