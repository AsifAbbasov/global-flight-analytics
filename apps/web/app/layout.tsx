import type { Metadata } from 'next'

import { QueryProvider } from '@/providers/query-provider'

import './globals.css'


export const metadata: Metadata = {
  title: 'Global Flight Analytics',
  description:
    'Open aviation traffic research, visualization and explainable analytics.',
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode
}>) {
  return (
    <html
      lang='en'
      className='h-full antialiased'
    >
      <body className='flex min-h-full flex-col'>
        <QueryProvider>{children}</QueryProvider>
      </body>
    </html>
  )
}
