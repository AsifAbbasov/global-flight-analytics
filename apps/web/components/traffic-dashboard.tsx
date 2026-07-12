'use client'

import { useEffect, useRef, useState } from 'react'

import { TrafficGlobe } from '@/components/globe/traffic-globe'
import { TrafficMap } from '@/components/map/traffic-map'
import {
  getRequestErrorMessage,
  isAbortError,
} from '@/lib/api/client'
import { getCurrentTraffic } from '@/lib/api/traffic'
import type { Region } from '@/types/region'
import type { TrafficAircraft } from '@/types/traffic'

interface TrafficDashboardProps {
  regions: Region[]
  initialTraffic: TrafficAircraft[]
  initialError: string | null
  regionsWarning: string | null
}

export function TrafficDashboard({
  regions,
  initialTraffic,
  initialError,
  regionsWarning,
}: TrafficDashboardProps) {
  const [selectedRegion, setSelectedRegion] = useState('world')
  const [traffic, setTraffic] = useState(initialTraffic)
  const [isLoading, setIsLoading] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(
    initialError
  )
  const [retryVersion, setRetryVersion] = useState(0)

  const skipInitialRequestRef = useRef(initialError === null)
  const requestSequenceRef = useRef(0)

  useEffect(() => {
    if (skipInitialRequestRef.current) {
      skipInitialRequestRef.current = false
      return
    }

    const requestController = new AbortController()
    const requestSequence = ++requestSequenceRef.current

    async function loadTraffic() {
      setIsLoading(true)
      setErrorMessage(null)

      try {
        const nextTraffic = await getCurrentTraffic(selectedRegion, {
          signal: requestController.signal,
        })

        if (requestSequence === requestSequenceRef.current) {
          setTraffic(nextTraffic)
        }
      } catch (error) {
        if (
          requestController.signal.aborted ||
          isAbortError(error)
        ) {
          return
        }

        if (requestSequence === requestSequenceRef.current) {
          setErrorMessage(getRequestErrorMessage(error))
        }
      } finally {
        if (
          !requestController.signal.aborted &&
          requestSequence === requestSequenceRef.current
        ) {
          setIsLoading(false)
        }
      }
    }

    void loadTraffic()

    return () => {
      requestController.abort()
    }
  }, [retryVersion, selectedRegion])

  return (
    <>
      <section className='mt-6 rounded-xl border border-slate-800 bg-slate-900 p-4'>
        <label
          className='block text-sm font-medium text-slate-300'
          htmlFor='traffic-region'
        >
          Region
        </label>

        <select
          id='traffic-region'
          value={selectedRegion}
          onChange={event => {
            setSelectedRegion(event.target.value)
          }}
          className='mt-2 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-white'
        >
          {regions.map(region => (
            <option key={region.code} value={region.code}>
              {region.name}
            </option>
          ))}
        </select>

        <div
          aria-live='polite'
          className='mt-3 flex flex-wrap items-center gap-3 text-sm'
        >
          <span className='text-slate-300'>
            Aircraft: {traffic.length}
          </span>

          {regionsWarning ? (
            <span className='text-amber-300'>{regionsWarning}</span>
          ) : null}

          {isLoading ? (
            <span className='text-sky-300'>Loading current traffic…</span>
          ) : null}

          {errorMessage ? (
            <>
              <span className='text-amber-300'>{errorMessage}</span>
              <button
                type='button'
                onClick={() => {
                  setRetryVersion(version => version + 1)
                }}
                className='rounded-md border border-amber-400/50 px-3 py-1 font-medium text-amber-200 transition hover:bg-amber-400/10'
              >
                Retry
              </button>
            </>
          ) : null}
        </div>
      </section>

      <div className='mt-4' aria-busy={isLoading}>
        <TrafficGlobe aircraft={traffic} />
      </div>

      <section className='mt-8 rounded-xl border border-slate-800 bg-slate-900 p-4 sm:p-6'>
        <h2 className='text-xl font-semibold'>Current Traffic</h2>

        <div className='mt-4' aria-busy={isLoading}>
          <TrafficMap aircraft={traffic} />
        </div>

        {!isLoading && !errorMessage && traffic.length === 0 ? (
          <p className='mt-4 text-sm text-slate-400'>
            No aircraft were returned for the selected region.
          </p>
        ) : null}
      </section>
    </>
  )
}
