import {
  operationDefinitions,
  type OperationDefinition,
  type OperationId,
  type OperationParameters,
  type OperationResponses,
} from './generated.js'

export interface APIClientOptions {
  readonly baseURL: string
  readonly fetch?: typeof globalThis.fetch
  readonly internalApiKey?: string
  readonly defaultHeaders?: Readonly<Record<string, string>>
}

export interface APIRequestOptions {
  readonly signal?: AbortSignal
  readonly headers?: Readonly<Record<string, string>>
}

export interface APIErrorPayload {
  readonly code?: string
  readonly message?: string
  readonly details?: unknown
}

export class APIError extends Error {
  readonly status: number
  readonly code?: string
  readonly requestId?: string
  readonly details?: unknown

  constructor(
    message: string,
    options: {
      readonly status: number
      readonly code?: string
      readonly requestId?: string
      readonly details?: unknown
    },
  ) {
    super(message)
    this.name = 'APIError'
    this.status = options.status
    this.code = options.code
    this.requestId = options.requestId
    this.details = options.details
  }
}

function normalizeBaseURL(value: string): URL {
  const normalized = new URL(value)
  if (normalized.username || normalized.password) {
    throw new TypeError('baseURL must not contain credentials')
  }
  if (normalized.protocol !== 'http:' && normalized.protocol !== 'https:') {
    throw new TypeError('baseURL must use http or https')
  }
  if (normalized.pathname !== '/') {
    throw new TypeError('baseURL must be an origin without a path')
  }
  if (normalized.search || normalized.hash) {
    throw new TypeError('baseURL must not contain a query string or fragment')
  }
  return normalized
}

function mergeHeaders(
  defaults: Readonly<Record<string, string>>,
  overrides: Readonly<Record<string, string>> | undefined,
): Headers {
  const headers = new Headers(defaults)
  for (const [name, value] of Object.entries(overrides ?? {})) headers.set(name, value)
  return headers
}

function objectSection(
  parameters: unknown,
  section: 'path' | 'query' | 'header',
): Readonly<Record<string, unknown>> {
  if (!parameters || typeof parameters !== 'object') return {}
  const value = (parameters as Record<string, unknown>)[section]
  return value && typeof value === 'object' ? value as Readonly<Record<string, unknown>> : {}
}

async function parseResponseBody(response: Response): Promise<unknown> {
  if (response.status === 204) return undefined
  const contentType = response.headers.get('content-type') ?? ''
  if (!contentType.toLowerCase().includes('application/json')) return response.text()
  const text = await response.text()
  return text.length === 0 ? undefined : JSON.parse(text)
}

function errorPayload(value: unknown): APIErrorPayload {
  if (!value || typeof value !== 'object') return {}
  const root = value as Record<string, unknown>
  const nested = root.error && typeof root.error === 'object'
    ? root.error as Record<string, unknown>
    : root
  return {
    code: typeof nested.code === 'string' ? nested.code : undefined,
    message: typeof nested.message === 'string' ? nested.message : undefined,
    details: nested.details,
  }
}

export class GlobalFlightAnalyticsClient {
  readonly #baseURL: URL
  readonly #fetch: typeof globalThis.fetch
  readonly #internalApiKey?: string
  readonly #defaultHeaders: Readonly<Record<string, string>>

  constructor(options: APIClientOptions) {
    this.#baseURL = normalizeBaseURL(options.baseURL)
    this.#fetch = options.fetch ?? globalThis.fetch
    this.#internalApiKey = options.internalApiKey
    this.#defaultHeaders = Object.freeze({
      Accept: 'application/json',
      ...options.defaultHeaders,
    })
  }

  async request<K extends OperationId>(
    operationId: K,
    parameters: OperationParameters[K],
    options: APIRequestOptions = {},
  ): Promise<OperationResponses[K]> {
    const definition: OperationDefinition = operationDefinitions[operationId]
    let routePath: string = definition.path
    const pathParameters = objectSection(parameters, 'path')

    for (const parameter of definition.parameters) {
      if (parameter.in !== 'path') continue
      const value = pathParameters[parameter.name]
      if (value === undefined || value === null || value === '') {
        throw new TypeError(`missing path parameter: ${parameter.name}`)
      }
      routePath = routePath.replace(`{${parameter.name}}`, encodeURIComponent(String(value)))
    }

    if (routePath.includes('{') || routePath.includes('}')) {
      throw new TypeError(`unresolved path parameter in ${routePath}`)
    }

    const requestURL = new URL(routePath, this.#baseURL)
    const queryParameters = objectSection(parameters, 'query')
    for (const [name, value] of Object.entries(queryParameters)) {
      if (value === undefined || value === null) continue
      if (Array.isArray(value)) {
        for (const entry of value) requestURL.searchParams.append(name, String(entry))
      } else {
        requestURL.searchParams.set(name, String(value))
      }
    }

    const headers = mergeHeaders(this.#defaultHeaders, options.headers)
    for (const [name, value] of Object.entries(objectSection(parameters, 'header'))) {
      if (value !== undefined && value !== null) headers.set(name, String(value))
    }

    if (definition.protected) {
      if (!this.#internalApiKey) {
        throw new TypeError(`operation ${String(operationId)} requires internalApiKey`)
      }
      headers.set('X-Internal-API-Key', this.#internalApiKey)
    }

    const requestInit: RequestInit = {
      method: definition.method,
      headers,
      signal: options.signal,
    }

    if (definition.hasBody) {
      const body = parameters && typeof parameters === 'object'
        ? (parameters as Record<string, unknown>).body
        : undefined
      if (body !== undefined) {
        headers.set('Content-Type', 'application/json')
        requestInit.body = JSON.stringify(body)
      }
    }

    const response = await this.#fetch(requestURL, requestInit)
    const body = await parseResponseBody(response)
    if (!response.ok) {
      const payload = errorPayload(body)
      throw new APIError(payload.message ?? `request failed with status ${response.status}`, {
        status: response.status,
        code: payload.code,
        requestId: response.headers.get('x-request-id') ?? undefined,
        details: payload.details,
      })
    }

    return body as OperationResponses[K]
  }
}
