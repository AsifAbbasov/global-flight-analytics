import {
  APIRequestError,
  requestAPIData,
  type APIRequestOptions,
} from '@/lib/api/client'
import type { Region } from '@/types/region'

export async function getRegions(
  options: APIRequestOptions = {}
): Promise<Region[]> {
  const data = await requestAPIData<unknown>('/api/v1/regions', options)

  if (!Array.isArray(data)) {
    throw new APIRequestError('The regions response is not an array.')
  }

  return data as Region[]
}
