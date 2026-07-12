import {
  APIRequestError,
  requestAPIData,
  type APIRequestOptions,
} from '@/lib/api/client'
import type { TrafficAircraft } from '@/types/traffic'

export async function getCurrentTraffic(
  regionCode?: string,
  options: APIRequestOptions = {}
): Promise<TrafficAircraft[]> {
  const searchParams = new URLSearchParams()

  if (regionCode?.trim()) {
    searchParams.set('region', regionCode.trim())
  }

  const data = await requestAPIData<unknown>(
    '/api/v1/traffic/current',
    {
      ...options,
      searchParams,
    }
  )

  if (!Array.isArray(data)) {
    throw new APIRequestError('The traffic response is not an array.')
  }

  return data as TrafficAircraft[]
}
