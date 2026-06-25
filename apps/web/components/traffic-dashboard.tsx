'use client'

import { useEffect, useState } from 'react'

import { TrafficGlobe } from '@/components/globe/traffic-globe'
import { TrafficMap } from '@/components/map/traffic-map'
import { getCurrentTraffic } from '@/lib/api/traffic'
import type { Region } from '@/types/region'
import type { TrafficAircraft } from '@/types/traffic'

interface TrafficDashboardProps {
  regions: Region[]
  initialTraffic: TrafficAircraft[]
}

export function TrafficDashboard({
  regions,
  initialTraffic,
}: TrafficDashboardProps) {
  const [selectedRegion, setSelectedRegion] = useState('world')
  const [traffic, setTraffic] = useState(initialTraffic)

  useEffect(() => {
    async function loadTraffic() {
      const nextTraffic = await getCurrentTraffic(selectedRegion)
      setTraffic(nextTraffic)
    }

    loadTraffic()
  }, [selectedRegion])

  return (
    <>
      <div className='mt-6 rounded-xl border border-slate-800 bg-slate-900 p-4'>
        <label className='block text-sm font-medium text-slate-300'>
          Region
        </label>

        <select
          value={selectedRegion}
          onChange={event => setSelectedRegion(event.target.value)}
          className='mt-2 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-white'
        >
          {regions.map(region => (
            <option key={region.code} value={region.code}>
              {region.name}
            </option>
          ))}
        </select>
      </div>

      <div className='mt-4'>
        <TrafficGlobe aircraft={traffic} />
      </div>

      <section className='mt-8 rounded-xl border border-slate-800 bg-slate-900 p-6'>
        <h2 className='text-xl font-semibold'>Current Traffic</h2>

        <div className='mt-4'>
          <TrafficMap aircraft={traffic} />
        </div>

        <pre className='mt-4 overflow-auto rounded-lg bg-black p-4 text-sm text-green-400'>
          {JSON.stringify(traffic, null, 2)}
        </pre>
      </section>
    </>
  )
}
