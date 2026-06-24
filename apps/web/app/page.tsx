import { getCurrentTraffic } from '@/lib/api/traffic'
import { TrafficMap } from '@/components/map/traffic-map'

export default async function Home() {
  const traffic = await getCurrentTraffic()

  return (
    <main className='min-h-screen bg-slate-950 p-8 text-white'>
      <h1 className='text-3xl font-bold'>Global Flight Analytics</h1>

      <p className='mt-2 text-slate-400'>
        Current air traffic from Go API and Neon PostgreSQL.
      </p>

      <section className='mt-8 rounded-xl border border-slate-800 bg-slate-900 p-6'>
        <h2 className='text-xl font-semibold'>Current Traffic</h2>

        <div className='mt-4'>
          <TrafficMap aircraft={traffic} />
        </div>

        <pre className='mt-4 overflow-auto rounded-lg bg-black p-4 text-sm text-green-400'>
          {JSON.stringify(traffic, null, 2)}
        </pre>
      </section>
    </main>
  )
}
