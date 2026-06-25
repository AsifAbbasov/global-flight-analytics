import type { TrafficAircraft } from '@/types/traffic'

interface TrafficResponse {
  success: boolean
  data: TrafficAircraft[]
}

export async function getCurrentTraffic(
  regionCode?: string
): Promise<TrafficAircraft[]> {
  const url = new URL('http://localhost:8080/api/v1/traffic/current')

  if (regionCode) {
    url.searchParams.set('region', regionCode)
  }

  const response = await fetch(url.toString(), {
    cache: 'no-store',
  })

  if (!response.ok) {
    throw new Error('Failed to load traffic')
  }

  const result: TrafficResponse = await response.json()

  return result.data
}
