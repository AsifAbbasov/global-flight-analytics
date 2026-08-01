// FRONTEND_PRODUCT_HARDENING_V1
'use client'

import {
  QueryClient,
  QueryClientProvider,
} from '@tanstack/react-query'
import { useState, type ReactNode } from 'react'

import { APIRequestError } from '@/lib/api/client'
import {
  frontendRetryDelayMilliseconds,
  shouldRetryFrontendQuery,
} from '@/lib/product/runtime-resilience-model'

interface QueryProviderProps {
  children: ReactNode
}

export function QueryProvider({ children }: QueryProviderProps) {
  const [queryClient] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            refetchOnWindowFocus: false,
            refetchOnReconnect: true,
            staleTime: 30_000,
            gcTime: 5 * 60_000,
            networkMode: 'online',
            retry: (failureCount: number, error: Error) =>
              shouldRetryFrontendQuery({
                failureCount,
                status:
                  error instanceof APIRequestError ? error.status : null,
                online:
                  typeof navigator === 'undefined' || navigator.onLine,
              }),
            retryDelay: frontendRetryDelayMilliseconds,
          },
          mutations: {
            networkMode: 'online',
            retry: false,
          },
        },
      })
  )

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  )
}
