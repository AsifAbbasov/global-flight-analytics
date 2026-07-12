const defaultAPIBaseURL = 'http://localhost:8080'
const defaultRequestTimeoutMilliseconds = 10_000

export interface APIRequestOptions {
  signal?: AbortSignal
  searchParams?: URLSearchParams
  timeoutMilliseconds?: number
}

interface APIEnvelope<T> {
  success: boolean
  data: T
}

export class APIRequestError extends Error {
  readonly status: number | null

  constructor(message: string, status: number | null = null) {
    super(message)
    this.name = 'APIRequestError'
    this.status = status
  }
}

export function isAbortError(error: unknown): boolean {
  return error instanceof DOMException && error.name === 'AbortError'
}

export function getRequestErrorMessage(error: unknown): string {
  if (error instanceof APIRequestError) {
    return error.message
  }

  if (error instanceof Error && error.message.trim() !== '') {
    return error.message
  }

  return 'The request failed. Please try again.'
}

export async function requestAPIData<T>(
  path: string,
  options: APIRequestOptions = {}
): Promise<T> {
  const requestURL = buildAPIURL(path, options.searchParams)
  const timeoutMilliseconds =
    options.timeoutMilliseconds ?? defaultRequestTimeoutMilliseconds

  if (!Number.isFinite(timeoutMilliseconds) || timeoutMilliseconds <= 0) {
    throw new APIRequestError('The API request timeout is invalid.')
  }

  const requestController = new AbortController()
  let timedOut = false

  const handleCallerAbort = () => {
    requestController.abort()
  }

  if (options.signal?.aborted) {
    requestController.abort()
  } else {
    options.signal?.addEventListener('abort', handleCallerAbort, {
      once: true,
    })
  }

  const timeoutID = windowOrGlobalSetTimeout(() => {
    timedOut = true
    requestController.abort()
  }, timeoutMilliseconds)

  try {
    const response = await fetch(requestURL, {
      cache: 'no-store',
      headers: {
        Accept: 'application/json',
      },
      signal: requestController.signal,
    })

    if (!response.ok) {
      throw new APIRequestError(
        `The API returned HTTP ${response.status}.`,
        response.status
      )
    }

    const contentType = response.headers.get('content-type') ?? ''

    if (!contentType.toLowerCase().includes('application/json')) {
      throw new APIRequestError('The API returned a non-JSON response.')
    }

    const payload: unknown = await response.json()

    if (!isAPIEnvelope<T>(payload) || payload.success !== true) {
      throw new APIRequestError('The API response shape is invalid.')
    }

    return payload.data
  } catch (error) {
    if (timedOut) {
      throw new APIRequestError('The API request timed out.')
    }

    throw error
  } finally {
    clearTimeout(timeoutID)
    options.signal?.removeEventListener('abort', handleCallerAbort)
  }
}

function buildAPIURL(path: string, searchParams?: URLSearchParams): URL {
  const baseURL = resolveAPIBaseURL()
  const normalizedPath = path.startsWith('/') ? path : `/${path}`
  const requestURL = new URL(normalizedPath, `${baseURL}/`)

  if (searchParams) {
    requestURL.search = searchParams.toString()
  }

  return requestURL
}

function resolveAPIBaseURL(): string {
  const configuredBaseURL = process.env.NEXT_PUBLIC_API_BASE_URL?.trim()

  if (!configuredBaseURL && process.env.NODE_ENV === 'production') {
    throw new APIRequestError(
      'NEXT_PUBLIC_API_BASE_URL is required in production.'
    )
  }

  const candidate = configuredBaseURL || defaultAPIBaseURL
  const parsedURL = new URL(candidate)

  if (parsedURL.protocol !== 'http:' && parsedURL.protocol !== 'https:') {
    throw new APIRequestError(
      'NEXT_PUBLIC_API_BASE_URL must use HTTP or HTTPS.'
    )
  }

  return parsedURL.toString().replace(/\/+$/, '')
}

function isAPIEnvelope<T>(value: unknown): value is APIEnvelope<T> {
  return (
    typeof value === 'object' &&
    value !== null &&
    'success' in value &&
    typeof value.success === 'boolean' &&
    'data' in value
  )
}

function windowOrGlobalSetTimeout(
  callback: () => void,
  timeoutMilliseconds: number
): ReturnType<typeof setTimeout> {
  return setTimeout(callback, timeoutMilliseconds)
}
