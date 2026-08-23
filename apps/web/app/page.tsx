// FRONTEND_APPLICATION_SHELL_V1
import { RegionalTrafficExperience } from '@/components/regional-traffic-experience'
import { ApplicationShell } from '@/components/product/application-shell'
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
    <ApplicationShell
      initialTrafficCount={initialTraffic.length}
      regionCount={regions.length}
      trafficUnavailable={initialError !== null}
      regionsUnavailable={regionsWarning !== null}
    >
      <RegionalTrafficExperience
        regions={regions}
        initialTraffic={initialTraffic}
        initialError={initialError}
        regionsWarning={regionsWarning}
      />
      <p className='mx-auto mt-6 max-w-7xl px-4 pb-6 text-xs text-slate-500 sm:px-6 lg:px-8'>
        Traffic data may include{' '}
        <a
          className='underline decoration-slate-600 underline-offset-2 hover:text-slate-300'
          href='https://adsb.lol/'
          rel='noreferrer'
          target='_blank'
        >
          ADSB.lol
        </a>{' '}
        data, licensed under the{' '}
        <a
          className='underline decoration-slate-600 underline-offset-2 hover:text-slate-300'
          href='https://opendatacommons.org/licenses/odbl/'
          rel='noreferrer'
          target='_blank'
        >
          ODbL
        </a>
        .
      </p>
    </ApplicationShell>
  )
}

function ensureWorldRegion(regions: Region[]): Region[] {
  if (regions.some(region => region.code === worldRegion.code)) {
    return regions
  }

  return [worldRegion, ...regions]
}
