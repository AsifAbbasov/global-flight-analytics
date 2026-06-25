import { TrafficDashboard } from '@/components/traffic-dashboard'
import { getRegions } from '@/lib/api/regions'
import { getCurrentTraffic } from '@/lib/api/traffic'

export default async function Home() {
  const regions = await getRegions()
  const initialTraffic = await getCurrentTraffic('world')

  return (
    <main className='min-h-screen bg-slate-950 p-8 text-white'>
      <h1 className='text-3xl font-bold'>Global Flight Analytics</h1>

      <p className='mt-2 text-slate-400'>
        Current air traffic from Go API and Neon PostgreSQL.
      </p>

      <TrafficDashboard regions={regions} initialTraffic={initialTraffic} />
    </main>
  )
}
