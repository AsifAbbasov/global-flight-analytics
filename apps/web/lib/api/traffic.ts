import type { TrafficAircraft } from '@/types/traffic'

interface TrafficResponse {
  success: boolean
  data: TrafficAircraft[]
}

export async function getCurrentTraffic(): Promise<TrafficAircraft[]> {
  const response = await fetch('http://localhost:8080/api/v1/traffic/current', {
    cache: 'no-store',
  })

  if (!response.ok) {
    throw new Error('Failed to load traffic')
  }

  const result: TrafficResponse = await response.json()

  return result.data
}
