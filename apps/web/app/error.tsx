// FRONTEND_PRODUCT_HARDENING_V1
'use client'

import Link from 'next/link'

interface ApplicationErrorProps {
  error: Error & { digest?: string }
  reset: () => void
}

export default function ApplicationError({
  error,
  reset,
}: ApplicationErrorProps) {
  return (
    <main
      id='main-content'
      className='flex min-h-screen items-center justify-center bg-slate-950 px-4 py-16 text-slate-100'
    >
      <section
        role='alert'
        aria-labelledby='application-error-title'
        className='w-full max-w-2xl rounded-2xl border border-rose-300/25 bg-slate-900 p-6 shadow-2xl shadow-black/30 sm:p-8'
      >
        <p className='text-xs font-semibold uppercase tracking-[0.22em] text-rose-300'>
          Recoverable application error
        </p>
        <h1
          id='application-error-title'
          className='mt-3 text-3xl font-semibold tracking-tight text-white'
        >
          This research view could not be rendered
        </h1>
        <p className='mt-4 text-sm leading-7 text-slate-400'>
          The failure is isolated from the rest of the browser session. Retry the
          current view or return to the application entry point. No operational
          aviation claim should be inferred from this error state.
        </p>
        {error.digest ? (
          <p className='mt-4 font-mono text-xs text-slate-600'>
            Reference: {error.digest}
          </p>
        ) : null}
        <div className='mt-6 flex flex-wrap gap-3'>
          <button
            type='button'
            onClick={reset}
            className='min-h-11 rounded-lg bg-rose-200 px-4 py-2 text-sm font-semibold text-slate-950 transition hover:bg-rose-100'
          >
            Retry this view
          </button>
          <Link
            href='/'
            className='inline-flex min-h-11 items-center rounded-lg border border-white/15 px-4 py-2 text-sm font-semibold text-slate-200 transition hover:bg-white/5'
          >
            Return to home
          </Link>
        </div>
      </section>
    </main>
  )
}
