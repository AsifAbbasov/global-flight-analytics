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
        <footer className='border-t border-white/10 px-4 py-3 text-center text-xs text-slate-400'>
          Live aircraft data ©{' '}
          <a
            className='underline underline-offset-2 hover:text-slate-200'
            href='https://www.adsb.lol/'
            rel='noreferrer'
            target='_blank'
          >
            ADSB.lol contributors
          </a>{' '}
          · licensed under{' '}
          <a
            className='underline underline-offset-2 hover:text-slate-200'
            href='https://opendatacommons.org/licenses/odbl/1-0/'
            rel='noreferrer'
            target='_blank'
          >
            ODbL 1.0
          </a>
        </footer>
      </body>
    </html>
  )
}
