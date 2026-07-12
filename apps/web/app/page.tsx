import { TrafficDashboard } from '@/components/traffic-dashboard'
import { getRegions } from '@/lib/api/regions'
import { getCurrentTraffic } from '@/lib/api/traffic'
import type { Region } from '@/types/region'

const worldRegion: Region = {
  code: 'world',
  name: 'World',
  description: 'Global air traffic',
  bounds: {
    min_latitude: -90,
    max_latitude: 90,
    min_longitude: -180,
    max_longitude: 180,
  },
}

export default async function Home() {
  const [regionsResult, trafficResult] = await Promise.allSettled([
    getRegions(),
    getCurrentTraffic(worldRegion.code),
  ])

  const hasRegions =
    regionsResult.status === 'fulfilled' &&
    regionsResult.value.length > 0

  const regions = hasRegions
    ? ensureWorldRegion(regionsResult.value)
    : [worldRegion]

  const initialTraffic =
    trafficResult.status === 'fulfilled' ? trafficResult.value : []

  const initialError =
    trafficResult.status === 'rejected'
      ? 'Initial traffic data is temporarily unavailable. Use Retry to request it again.'
      : null

  const regionsWarning = hasRegions
    ? null
    : 'The region list is temporarily unavailable. World view remains available; reload the page to retry.'

  return (
    <main className='min-h-screen bg-slate-950 p-4 text-white sm:p-8'>
      <h1 className='text-3xl font-bold'>Global Flight Analytics</h1>

      <p className='mt-2 text-slate-400'>
        Current air traffic from the Go API and PostgreSQL.
      </p>

      <TrafficDashboard
        regions={regions}
        initialTraffic={initialTraffic}
        initialError={initialError}
        regionsWarning={regionsWarning}
      />
    </main>
  )
}

function ensureWorldRegion(regions: Region[]): Region[] {
  if (regions.some(region => region.code === worldRegion.code)) {
    return regions
  }

  return [worldRegion, ...regions]
}
