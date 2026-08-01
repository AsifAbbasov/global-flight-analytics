import type { Metadata } from 'next'

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

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html lang='en' className='h-full antialiased'>
      <body className='flex min-h-full flex-col bg-slate-950 text-slate-100'>
        <QueryProvider>{children}</QueryProvider>
      </body>
    </html>
  )
}
