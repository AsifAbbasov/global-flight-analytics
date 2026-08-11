// FRONTEND_PRODUCT_HARDENING_V1
// FRONTEND_MAP_FIRST_REDESIGN_V1
import type { Metadata, Viewport } from 'next'

import { RuntimeResilienceBoundary } from '@/components/product/runtime-resilience-boundary'
import { QueryProvider } from '@/providers/query-provider'

import './globals.css'

export const metadata: Metadata = {
  title: {
    default: 'Global Flight Analytics',
    template: '%s · Global Flight Analytics',
  },
  description:
    'Open aviation traffic research, visualization and explainable analytics.',
  applicationName: 'Global Flight Analytics',
  keywords: [
    'aviation analytics',
    'air traffic visualization',
    'open aviation data',
    'flight research',
  ],
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  colorScheme: 'dark',
  themeColor: '#111315',
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang='en' className='h-full antialiased'>
      <body className='flex min-h-full flex-col bg-[#111315] text-slate-100'>
        <QueryProvider>
          <RuntimeResilienceBoundary>
            {children}
          </RuntimeResilienceBoundary>
        </QueryProvider>
      </body>
    </html>
  )
}
