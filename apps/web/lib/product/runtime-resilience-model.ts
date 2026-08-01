// FRONTEND_PRODUCT_HARDENING_V1

export type RuntimeConnectivityState =
  | 'unknown'
  | 'online'
  | 'offline'
  | 'recovered'

export interface RuntimeConnectivityView {
  visible: boolean
  title: string
  message: string
  liveMode: 'polite' | 'assertive'
}

export interface FrontendQueryRetryInput {
  failureCount: number
  status: number | null
  online: boolean
}

export const maximumFrontendQueryRetries = 2
export const maximumFrontendRetryDelayMilliseconds = 4_000

const transientHTTPStatuses = new Set([408, 425, 429])

export function shouldRetryFrontendQuery({
  failureCount,
  status,
  online,
}: FrontendQueryRetryInput): boolean {
  if (
    !Number.isInteger(failureCount) ||
    failureCount < 0 ||
    failureCount >= maximumFrontendQueryRetries ||
    !online
  ) {
    return false
  }

  if (status === null) {
    return true
  }

  return transientHTTPStatuses.has(status) || status >= 500
}

export function frontendRetryDelayMilliseconds(
  failureCount: number
): number {
  const normalizedFailureCount =
    Number.isInteger(failureCount) && failureCount > 0 ? failureCount : 0
  return Math.min(
    1_000 * 2 ** normalizedFailureCount,
    maximumFrontendRetryDelayMilliseconds
  )
}

export function runtimeConnectivityState(
  online: boolean | null,
  recoveredFromOffline: boolean
): RuntimeConnectivityState {
  if (online === null) {
    return 'unknown'
  }
  if (!online) {
    return 'offline'
  }
  return recoveredFromOffline ? 'recovered' : 'online'
}

export function buildRuntimeConnectivityView(
  state: RuntimeConnectivityState
): RuntimeConnectivityView {
  switch (state) {
    case 'offline':
      return {
        visible: true,
        title: 'Connection unavailable',
        message:
          'Live requests are paused while the browser is offline. Existing evidence remains visible.',
        liveMode: 'assertive',
      }
    case 'recovered':
      return {
        visible: true,
        title: 'Connection restored',
        message:
          'The browser is online again. Active queries may refresh their evidence.',
        liveMode: 'polite',
      }
    case 'unknown':
    case 'online':
      return {
        visible: false,
        title: '',
        message: '',
        liveMode: 'polite',
      }
  }
}
