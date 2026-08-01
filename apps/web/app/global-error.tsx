// FRONTEND_PRODUCT_HARDENING_V1
'use client'

interface GlobalErrorProps {
  error: Error & { digest?: string }
  reset: () => void
}

export default function GlobalError({ error, reset }: GlobalErrorProps) {
  return (
    <html lang='en'>
      <body
        style={{
          margin: 0,
          background: '#020617',
          color: '#e2e8f0',
          fontFamily: 'Arial, Helvetica, sans-serif',
        }}
      >
        <main
          style={{
            minHeight: '100vh',
            display: 'grid',
            placeItems: 'center',
            padding: '2rem',
          }}
        >
          <section
            role='alert'
            aria-labelledby='global-error-title'
            style={{
              width: 'min(100%, 40rem)',
              border: '1px solid rgba(253, 164, 175, 0.35)',
              borderRadius: '1rem',
              background: '#0f172a',
              padding: '2rem',
            }}
          >
            <p style={{ color: '#fda4af', fontSize: '0.75rem' }}>
              GLOBAL APPLICATION FALLBACK
            </p>
            <h1 id='global-error-title'>Global Flight Analytics is unavailable</h1>
            <p style={{ color: '#94a3b8', lineHeight: 1.7 }}>
              A root rendering failure prevented the normal interface from loading.
              Retry the application before relying on any displayed analytical state.
            </p>
            {error.digest ? (
              <p style={{ color: '#64748b', fontFamily: 'monospace' }}>
                Reference: {error.digest}
              </p>
            ) : null}
            <button
              type='button'
              onClick={reset}
              style={{
                minHeight: '44px',
                marginTop: '1rem',
                border: 0,
                borderRadius: '0.5rem',
                background: '#bae6fd',
                color: '#020617',
                cursor: 'pointer',
                fontWeight: 700,
                padding: '0.7rem 1rem',
              }}
            >
              Retry application
            </button>
          </section>
        </main>
      </body>
    </html>
  )
}
