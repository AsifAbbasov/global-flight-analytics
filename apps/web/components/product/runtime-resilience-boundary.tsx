// FRONTEND_PRODUCT_HARDENING_V1
'use client'

import { useEffect, useRef, useState, type ReactNode } from 'react'

import {
  buildRuntimeConnectivityView,
  runtimeConnectivityState,
  type RuntimeConnectivityState,
} from '@/lib/product/runtime-resilience-model'

interface RuntimeResilienceBoundaryProps {
  children: ReactNode
}

const recoveryNoticeDurationMilliseconds = 5_000

export function RuntimeResilienceBoundary({
  children,
}: RuntimeResilienceBoundaryProps) {
  const [connectivityState, setConnectivityState] =
    useState<RuntimeConnectivityState>('unknown')
  const wasOfflineRef = useRef(false)
  const recoveryTimerRef = useRef<number | null>(null)

  useEffect(() => {
    const clearRecoveryTimer = () => {
      if (recoveryTimerRef.current !== null) {
        window.clearTimeout(recoveryTimerRef.current)
        recoveryTimerRef.current = null
      }
    }

    const publishConnectivity = (online: boolean) => {
      clearRecoveryTimer()

      if (!online) {
        wasOfflineRef.current = true
        setConnectivityState(runtimeConnectivityState(false, false))
        return
      }

      const recoveredFromOffline = wasOfflineRef.current
      wasOfflineRef.current = false
      setConnectivityState(
        runtimeConnectivityState(true, recoveredFromOffline)
      )

      if (recoveredFromOffline) {
        recoveryTimerRef.current = window.setTimeout(() => {
          setConnectivityState('online')
          recoveryTimerRef.current = null
        }, recoveryNoticeDurationMilliseconds)
      }
    }

    const handleOnline = () => publishConnectivity(true)
    const handleOffline = () => publishConnectivity(false)

    publishConnectivity(window.navigator.onLine)
    window.addEventListener('online', handleOnline)
    window.addEventListener('offline', handleOffline)

    return () => {
      clearRecoveryTimer()
      window.removeEventListener('online', handleOnline)
      window.removeEventListener('offline', handleOffline)
    }
  }, [])

  const connectivityView = buildRuntimeConnectivityView(connectivityState)

  return (
    <>
      {children}
      {connectivityView.visible ? (
        <aside
          role={
            connectivityView.liveMode === 'assertive' ? 'alert' : 'status'
          }
          aria-live={connectivityView.liveMode}
          aria-atomic='true'
          className={`fixed inset-x-4 bottom-4 z-[80] mx-auto max-w-xl rounded-xl border p-4 shadow-2xl backdrop-blur sm:inset-x-auto sm:right-6 ${
            connectivityState === 'offline'
              ? 'border-amber-300/40 bg-amber-950/95 text-amber-50'
              : 'border-emerald-300/40 bg-emerald-950/95 text-emerald-50'
          }`}
        >
          <p className='text-sm font-semibold'>{connectivityView.title}</p>
          <p className='mt-1 text-xs leading-5 opacity-80'>
            {connectivityView.message}
          </p>
        </aside>
      ) : null}
    </>
  )
}
