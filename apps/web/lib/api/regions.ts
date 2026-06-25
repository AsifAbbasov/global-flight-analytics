import type { Region } from '@/types/region'

interface RegionsResponse {
  success: boolean
  data: Region[]
}

export async function getRegions(): Promise<Region[]> {
  const response = await fetch('http://localhost:8080/api/v1/regions', {
    cache: 'no-store',
  })

  if (!response.ok) {
    throw new Error('Failed to load regions')
  }

  const result: RegionsResponse = await response.json()

  return result.data
}
