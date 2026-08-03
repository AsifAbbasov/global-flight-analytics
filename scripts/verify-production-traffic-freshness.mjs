#!/usr/bin/env node

import { pathToFileURL } from 'node:url'

const DEFAULT_MAX_AGE_SECONDS = 30 * 60
const DEFAULT_MAX_FUTURE_SKEW_SECONDS = 60
const DEFAULT_ATTEMPTS = 12
const DEFAULT_RETRY_DELAY_MILLISECONDS = 10_000
const DEFAULT_REQUEST_TIMEOUT_MILLISECONDS = 30_000

export function unwrapAPIData(payload) {
  if (
    payload !== null &&
    typeof payload === 'object' &&
    !Array.isArray(payload) &&
    Object.hasOwn(payload, 'data')
  ) {
    return payload.data
  }

  return payload
}

export function summarizeTrafficFreshness(
  payload,
  now = new Date()
) {
  const traffic = unwrapAPIData(payload)
  if (!Array.isArray(traffic)) {
    throw new Error('production traffic response data must be an array')
  }
  if (traffic.length === 0) {
    throw new Error('production traffic response must contain at least one aircraft')
  }

  const observedTimestamps = traffic.map((aircraft, index) => {
    if (aircraft === null || typeof aircraft !== 'object') {
      throw new Error(`production traffic item ${index} must be an object`)
    }

    const observedAt = aircraft.observed_at
    if (typeof observedAt !== 'string' || observedAt.trim() === '') {
      throw new Error(`production traffic item ${index} is missing observed_at`)
    }

    const timestamp = Date.parse(observedAt)
    if (!Number.isFinite(timestamp)) {
      throw new Error(
        `production traffic item ${index} has invalid observed_at ${JSON.stringify(observedAt)}`
      )
    }

    return timestamp
  })

  const latestTimestamp = Math.max(...observedTimestamps)
  const latestAgeSeconds = Math.floor(
    (now.getTime() - latestTimestamp) / 1000
  )

  return {
    aircraftCount: traffic.length,
    latestObservedAt: new Date(latestTimestamp).toISOString(),
    latestAgeSeconds,
  }
}

export function assertTrafficFreshness(
  summary,
  {
    maxAgeSeconds = DEFAULT_MAX_AGE_SECONDS,
    maxFutureSkewSeconds = DEFAULT_MAX_FUTURE_SKEW_SECONDS,
  } = {}
) {
  if (!Number.isInteger(maxAgeSeconds) || maxAgeSeconds <= 0) {
    throw new Error('maximum traffic age must be a positive integer')
  }
  if (
    !Number.isInteger(maxFutureSkewSeconds) ||
    maxFutureSkewSeconds < 0
  ) {
    throw new Error('maximum future skew must be a non-negative integer')
  }
  if (summary.latestAgeSeconds > maxAgeSeconds) {
    throw new Error(
      `production traffic is stale: age=${summary.latestAgeSeconds}s maximum=${maxAgeSeconds}s`
    )
  }
  if (summary.latestAgeSeconds < -maxFutureSkewSeconds) {
    throw new Error(
      `production traffic is too far in the future: age=${summary.latestAgeSeconds}s allowed_skew=${maxFutureSkewSeconds}s`
    )
  }
}

export async function fetchJSONWithRetry(
  url,
  {
    attempts = DEFAULT_ATTEMPTS,
    retryDelayMilliseconds = DEFAULT_RETRY_DELAY_MILLISECONDS,
    requestTimeoutMilliseconds = DEFAULT_REQUEST_TIMEOUT_MILLISECONDS,
    fetchImpl = fetch,
    sleep = defaultSleep,
  } = {}
) {
  let latestError = null

  for (let attempt = 1; attempt <= attempts; attempt += 1) {
    const controller = new AbortController()
    const timeout = setTimeout(
      () => controller.abort(),
      requestTimeoutMilliseconds
    )

    try {
      const response = await fetchImpl(url, {
        headers: {
          Accept: 'application/json',
        },
        signal: controller.signal,
      })

      if (!response.ok) {
        throw new Error(
          `production traffic request returned HTTP ${response.status}`
        )
      }

      return await response.json()
    } catch (error) {
      latestError = error
      if (attempt === attempts) {
        break
      }
      await sleep(retryDelayMilliseconds)
    } finally {
      clearTimeout(timeout)
    }
  }

  throw new Error(
    `production traffic request failed after ${attempts} attempts: ${latestError?.message ?? 'unknown error'}`
  )
}

function defaultSleep(durationMilliseconds) {
  return new Promise((resolve) => {
    setTimeout(resolve, durationMilliseconds)
  })
}

function requiredPositiveIntegerEnvironmentVariable(
  name,
  defaultValue
) {
  const rawValue = process.env[name]?.trim()
  if (!rawValue) {
    return defaultValue
  }

  const value = Number(rawValue)
  if (!Number.isInteger(value) || value <= 0) {
    throw new Error(`${name} must be a positive integer`)
  }

  return value
}

async function main() {
  const apiBaseURL = process.env.API_BASE_URL?.trim()
  if (!apiBaseURL) {
    throw new Error('API_BASE_URL is required')
  }

  const maxAgeSeconds = requiredPositiveIntegerEnvironmentVariable(
    'MAX_TRAFFIC_AGE_SECONDS',
    DEFAULT_MAX_AGE_SECONDS
  )
  const attempts = requiredPositiveIntegerEnvironmentVariable(
    'TRAFFIC_FRESHNESS_ATTEMPTS',
    DEFAULT_ATTEMPTS
  )
  const requestTimeoutMilliseconds =
    requiredPositiveIntegerEnvironmentVariable(
      'TRAFFIC_FRESHNESS_REQUEST_TIMEOUT_MILLISECONDS',
      DEFAULT_REQUEST_TIMEOUT_MILLISECONDS
    )

  const endpoint = new URL(
    '/api/v1/traffic/current',
    apiBaseURL
  ).toString()

  const payload = await fetchJSONWithRetry(endpoint, {
    attempts,
    requestTimeoutMilliseconds,
  })
  const summary = summarizeTrafficFreshness(payload)
  assertTrafficFreshness(summary, {
    maxAgeSeconds,
  })

  console.log(
    `PRODUCTION_TRAFFIC_FRESHNESS=PASS count=${summary.aircraftCount}` +
      ` latest_observed_at=${summary.latestObservedAt}` +
      ` age_seconds=${summary.latestAgeSeconds}` +
      ` maximum_age_seconds=${maxAgeSeconds}`
  )
}

const isMain =
  process.argv[1] !== undefined &&
  import.meta.url === pathToFileURL(process.argv[1]).href

if (isMain) {
  main().catch((error) => {
    console.error(`PRODUCTION_TRAFFIC_FRESHNESS=FAIL error=${JSON.stringify(error.message)}`)
    process.exitCode = 1
  })
}
