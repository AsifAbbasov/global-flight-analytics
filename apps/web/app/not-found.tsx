// FRONTEND_PRODUCT_HARDENING_V1

import Link from 'next/link'

export default function NotFound() {
  return (
    <main
      id='main-content'
      className='flex min-h-screen items-center justify-center bg-slate-950 px-4 py-16 text-slate-100'
    >
      <section
        aria-labelledby='not-found-title'
        className='w-full max-w-xl rounded-2xl border border-white/10 bg-slate-900 p-6 text-center shadow-2xl shadow-black/30 sm:p-8'
      >
        <p className='font-mono text-sm text-sky-300'>404</p>
        <h1
          id='not-found-title'
          className='mt-3 text-3xl font-semibold tracking-tight text-white'
        >
          Research view not found
        </h1>
        <p className='mt-4 text-sm leading-7 text-slate-400'>
          The requested route is not part of the published research interface.
          Return to the main workspace to inspect current and historical evidence.
        </p>
        <Link
          href='/'
          className='mt-6 inline-flex min-h-11 items-center rounded-lg bg-sky-200 px-4 py-2 text-sm font-semibold text-slate-950 transition hover:bg-sky-100'
        >
          Open Global Flight Analytics
        </Link>
      </section>
    </main>
  )
}
