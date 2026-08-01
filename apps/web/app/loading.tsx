// FRONTEND_PRODUCT_HARDENING_V1

export default function Loading() {
  return (
    <main
      id='main-content'
      aria-busy='true'
      className='min-h-screen bg-slate-950 px-4 py-16 text-slate-100 sm:px-8'
    >
      <section
        role='status'
        aria-live='polite'
        aria-label='Loading Global Flight Analytics'
        className='mx-auto max-w-[1600px]'
      >
        <span className='sr-only'>Loading aviation research workspace.</span>
        <div className='h-4 w-56 animate-pulse rounded bg-slate-800' />
        <div className='mt-6 h-14 max-w-4xl animate-pulse rounded-xl bg-slate-800/80' />
        <div className='mt-4 h-5 max-w-2xl animate-pulse rounded bg-slate-800/60' />
        <div className='mt-10 grid gap-4 md:grid-cols-2 xl:grid-cols-4'>
          {Array.from({ length: 4 }, (_, index) => (
            <div
              key={index}
              aria-hidden='true'
              className='h-32 animate-pulse rounded-2xl border border-white/5 bg-slate-900'
            />
          ))}
        </div>
      </section>
    </main>
  )
}
