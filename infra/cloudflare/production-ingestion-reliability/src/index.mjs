const SERVICE_NAME =
  'global-flight-analytics-production-ingestion-reliability'

const DEFAULTS = Object.freeze({
  githubAPIBaseURL: 'https://api.github.com/',
  githubAPIVersion: '2022-11-28',
  githubOwner: 'AsifAbbasov',
  githubRepository: 'global-flight-analytics',
  githubWorkflow: 'production-traffic-ingestion.yml',
  githubRef: 'main',
  trafficAPIURL:
    'https://global-flight-analytics-api.onrender.com/api/v1/traffic/current',
  primaryCron: '3,13,23,33,43,53 * * * *',
  watchdogCron: '*/5 * * * *',
  maxTrafficAgeSeconds: 1800,
  maxFutureSkewSeconds: 60,
  deduplicationWindowSeconds: 480,
  recentFailureCooldownSeconds: 21600,
  dispatchEnabled: false,
  requestTimeoutMilliseconds: 10000,
  workflowRunsPerPage: 20,
})

const DISPATCH_SOURCES = Object.freeze({
  primary: 'cloudflare-primary',
  watchdog: 'cloudflare-watchdog',
})

function requiredString(environment, name) {
  const value = environment[name]
  if (typeof value !== 'string' || value.trim() === '') {
    throw new Error(`${name} is required`)
  }
  return value.trim()
}

function optionalString(environment, name, defaultValue) {
  const value = environment[name]
  if (value === undefined || value === null || String(value).trim() === '') {
    return defaultValue
  }
  return String(value).trim()
}

function strictBoolean(environment, name, defaultValue) {
  const rawValue = optionalString(environment, name, String(defaultValue))
  if (rawValue === 'true') {
    return true
  }
  if (rawValue === 'false') {
    return false
  }
  throw new Error(`${name} must be true or false`)
}

function positiveInteger(environment, name, defaultValue, maximum) {
  const rawValue = optionalString(environment, name, String(defaultValue))
  const value = Number(rawValue)
  if (!Number.isInteger(value) || value <= 0 || value > maximum) {
    throw new Error(`${name} must be an integer between 1 and ${maximum}`)
  }
  return value
}

function nonNegativeInteger(environment, name, defaultValue, maximum) {
  const rawValue = optionalString(environment, name, String(defaultValue))
  const value = Number(rawValue)
  if (!Number.isInteger(value) || value < 0 || value > maximum) {
    throw new Error(`${name} must be an integer between 0 and ${maximum}`)
  }
  return value
}

function httpsURL(rawValue, name) {
  let parsed
  try {
    parsed = new URL(rawValue)
  } catch {
    throw new Error(`${name} must be a valid URL`)
  }

  if (
    parsed.protocol !== 'https:' ||
    parsed.username !== '' ||
    parsed.password !== '' ||
    parsed.hash !== ''
  ) {
    throw new Error(
      `${name} must be an HTTPS URL without credentials or fragments`
    )
  }
  return parsed
}

export function loadConfig(environment) {
  const githubToken = requiredString(environment, 'GITHUB_ACTIONS_TOKEN')
  const githubAPIBaseURL = httpsURL(
    optionalString(
      environment,
      'GITHUB_API_BASE_URL',
      DEFAULTS.githubAPIBaseURL
    ),
    'GITHUB_API_BASE_URL'
  )
  const trafficAPIURL = httpsURL(
    optionalString(environment, 'TRAFFIC_API_URL', DEFAULTS.trafficAPIURL),
    'TRAFFIC_API_URL'
  )
  const primaryCron = optionalString(
    environment,
    'PRIMARY_CRON',
    DEFAULTS.primaryCron
  )
  const watchdogCron = optionalString(
    environment,
    'WATCHDOG_CRON',
    DEFAULTS.watchdogCron
  )
  if (primaryCron === watchdogCron) {
    throw new Error('PRIMARY_CRON and WATCHDOG_CRON must be different')
  }

  return Object.freeze({
    githubToken,
    githubAPIBaseURL,
    githubAPIVersion: optionalString(
      environment,
      'GITHUB_API_VERSION',
      DEFAULTS.githubAPIVersion
    ),
    githubOwner: optionalString(
      environment,
      'GITHUB_OWNER',
      DEFAULTS.githubOwner
    ),
    githubRepository: optionalString(
      environment,
      'GITHUB_REPOSITORY',
      DEFAULTS.githubRepository
    ),
    githubWorkflow: optionalString(
      environment,
      'GITHUB_WORKFLOW',
      DEFAULTS.githubWorkflow
    ),
    githubRef: optionalString(environment, 'GITHUB_REF', DEFAULTS.githubRef),
    trafficAPIURL,
    primaryCron,
    watchdogCron,
    dispatchEnabled: strictBoolean(
      environment,
      'DISPATCH_ENABLED',
      DEFAULTS.dispatchEnabled
    ),
    maxTrafficAgeSeconds: positiveInteger(
      environment,
      'MAX_TRAFFIC_AGE_SECONDS',
      DEFAULTS.maxTrafficAgeSeconds,
      86400
    ),
    maxFutureSkewSeconds: nonNegativeInteger(
      environment,
      'MAX_FUTURE_SKEW_SECONDS',
      DEFAULTS.maxFutureSkewSeconds,
      3600
    ),
    deduplicationWindowSeconds: positiveInteger(
      environment,
      'DEDUPLICATION_WINDOW_SECONDS',
      DEFAULTS.deduplicationWindowSeconds,
      3600
    ),
    recentFailureCooldownSeconds: positiveInteger(
      environment,
      'RECENT_FAILURE_COOLDOWN_SECONDS',
      DEFAULTS.recentFailureCooldownSeconds,
      86400
    ),
    requestTimeoutMilliseconds: positiveInteger(
      environment,
      'REQUEST_TIMEOUT_MILLISECONDS',
      DEFAULTS.requestTimeoutMilliseconds,
      30000
    ),
    workflowRunsPerPage: positiveInteger(
      environment,
      'WORKFLOW_RUNS_PER_PAGE',
      DEFAULTS.workflowRunsPerPage,
      100
    ),
  })
}

async function fetchWithTimeout(url, init, timeoutMilliseconds, fetchImpl) {
  const controller = new AbortController()
  const timeout = setTimeout(() => controller.abort(), timeoutMilliseconds)
  try {
    return await fetchImpl(url, {
      ...init,
      signal: controller.signal,
    })
  } finally {
    clearTimeout(timeout)
  }
}

function githubHeaders(config, includeJSON = false) {
  const headers = {
    Accept: 'application/vnd.github+json',
    Authorization: `Bearer ${config.githubToken}`,
    'X-GitHub-Api-Version': config.githubAPIVersion,
    'User-Agent': SERVICE_NAME,
  }
  if (includeJSON) {
    headers['Content-Type'] = 'application/json'
  }
  return headers
}

function workflowAPIURL(config, suffix) {
  const owner = encodeURIComponent(config.githubOwner)
  const repository = encodeURIComponent(config.githubRepository)
  const workflow = encodeURIComponent(config.githubWorkflow)
  return new URL(
    `/repos/${owner}/${repository}/actions/workflows/${workflow}${suffix}`,
    config.githubAPIBaseURL
  )
}

async function listWorkflowRuns(config, fetchImpl) {
  const url = workflowAPIURL(config, '/runs')
  url.searchParams.set('branch', config.githubRef)
  url.searchParams.set('per_page', String(config.workflowRunsPerPage))

  const response = await fetchWithTimeout(
    url,
    {
      method: 'GET',
      headers: githubHeaders(config),
    },
    config.requestTimeoutMilliseconds,
    fetchImpl
  )

  if (!response.ok) {
    throw new Error(
      `GitHub workflow-runs request returned HTTP ${response.status}`
    )
  }

  const payload = await response.json()
  if (
    payload === null ||
    typeof payload !== 'object' ||
    !Array.isArray(payload.workflow_runs)
  ) {
    throw new Error('GitHub workflow-runs response is missing workflow_runs')
  }
  return payload.workflow_runs
}

function parseRunTimestamp(run, field) {
  const rawValue = run[field]
  if (typeof rawValue !== 'string' || rawValue.trim() === '') {
    throw new Error(`GitHub workflow run is missing ${field}`)
  }
  const timestamp = Date.parse(rawValue)
  if (!Number.isFinite(timestamp)) {
    throw new Error(`GitHub workflow run has invalid ${field}`)
  }
  return timestamp
}

export function classifyWorkflowRuns(
  workflowRuns,
  nowMilliseconds,
  deduplicationWindowSeconds,
  recentFailureCooldownSeconds = deduplicationWindowSeconds
) {
  if (!Array.isArray(workflowRuns)) {
    throw new Error('workflow runs must be an array')
  }

  let activeRun = null
  let latestCompletedRun = null
  let latestCompletedTimestamp = Number.NEGATIVE_INFINITY

  for (const run of workflowRuns) {
    if (run === null || typeof run !== 'object') {
      throw new Error('GitHub workflow run must be an object')
    }

    const status = run.status
    if (typeof status !== 'string' || status.trim() === '') {
      throw new Error('GitHub workflow run is missing status')
    }

    if (status !== 'completed') {
      activeRun ??= {
        id: run.id ?? null,
        status,
        event: run.event ?? null,
      }
      continue
    }

    const createdAt = parseRunTimestamp(run, 'created_at')
    if (createdAt > latestCompletedTimestamp) {
      latestCompletedTimestamp = createdAt
      latestCompletedRun = {
        id: run.id ?? null,
        status,
        conclusion: run.conclusion ?? null,
        event: run.event ?? null,
        createdAt: new Date(createdAt).toISOString(),
      }
    }
  }

  let recentSuccessfulRun = null
  let recentFailedRun = null

  if (latestCompletedRun !== null) {
    const successBoundary =
      nowMilliseconds - deduplicationWindowSeconds * 1000
    const failureBoundary =
      nowMilliseconds - recentFailureCooldownSeconds * 1000

    if (
      latestCompletedRun.conclusion === 'success' &&
      latestCompletedTimestamp >= successBoundary
    ) {
      recentSuccessfulRun = latestCompletedRun
    } else if (
      latestCompletedRun.conclusion !== 'success' &&
      latestCompletedTimestamp >= failureBoundary
    ) {
      recentFailedRun = latestCompletedRun
    }
  }

  return {
    activeRun,
    latestCompletedRun,
    recentSuccessfulRun,
    recentFailedRun,
  }
}

async function dispatchWorkflow(config, dispatchSource, fetchImpl) {
  const url = workflowAPIURL(config, '/dispatches')
  const response = await fetchWithTimeout(
    url,
    {
      method: 'POST',
      headers: githubHeaders(config, true),
      body: JSON.stringify({
        ref: config.githubRef,
        inputs: {
          dispatch_source: dispatchSource,
        },
      }),
    },
    config.requestTimeoutMilliseconds,
    fetchImpl
  )

  if (response.status !== 204) {
    throw new Error(
      `GitHub workflow dispatch returned HTTP ${response.status}`
    )
  }
}

async function dispatchUnlessBlocked(
  config,
  dispatchSource,
  nowMilliseconds,
  fetchImpl
) {
  const workflowRuns = await listWorkflowRuns(config, fetchImpl)
  const classification = classifyWorkflowRuns(
    workflowRuns,
    nowMilliseconds,
    config.deduplicationWindowSeconds,
    config.recentFailureCooldownSeconds
  )

  if (classification.activeRun !== null) {
    return {
      action: 'skipped-active-run',
      decisionMarker: 'ACTIVE_RUN_DEDUPLICATION',
      run: classification.activeRun,
    }
  }

  if (classification.recentSuccessfulRun !== null) {
    return {
      action: 'skipped-recent-success',
      decisionMarker: 'RECENT_SUCCESS_DEDUPLICATION',
      run: classification.recentSuccessfulRun,
    }
  }

  if (classification.recentFailedRun !== null) {
    return {
      action: 'skipped-recent-failure',
      decisionMarker: 'RECENT_FAILURE_CIRCUIT_BREAKER',
      run: classification.recentFailedRun,
    }
  }

  await dispatchWorkflow(config, dispatchSource, fetchImpl)
  return {
    action: 'dispatched',
    decisionMarker: 'WORKFLOW_DISPATCH',
    dispatchSource,
  }
}

function unwrapAPIData(payload) {
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

export function summarizeTrafficFreshness(payload, nowMilliseconds) {
  const traffic = unwrapAPIData(payload)
  if (!Array.isArray(traffic)) {
    throw new Error('traffic response data must be an array')
  }

  if (traffic.length === 0) {
    return {
      state: 'empty',
      aircraftCount: 0,
      latestObservedAt: null,
      latestAgeSeconds: null,
    }
  }

  const timestamps = traffic.map((aircraft, index) => {
    if (aircraft === null || typeof aircraft !== 'object') {
      throw new Error(`traffic item ${index} must be an object`)
    }
    const observedAt = aircraft.observed_at
    if (typeof observedAt !== 'string' || observedAt.trim() === '') {
      throw new Error(`traffic item ${index} is missing observed_at`)
    }
    const timestamp = Date.parse(observedAt)
    if (!Number.isFinite(timestamp)) {
      throw new Error(`traffic item ${index} has invalid observed_at`)
    }
    return timestamp
  })

  const latestTimestamp = Math.max(...timestamps)
  return {
    state: 'observed',
    aircraftCount: traffic.length,
    latestObservedAt: new Date(latestTimestamp).toISOString(),
    latestAgeSeconds: Math.floor(
      (nowMilliseconds - latestTimestamp) / 1000
    ),
  }
}

export function classifyTrafficFreshness(summary, config) {
  if (summary.state === 'empty') {
    return 'stale'
  }
  if (summary.latestAgeSeconds < -config.maxFutureSkewSeconds) {
    throw new Error(
      `traffic observation is too far in the future: age=${summary.latestAgeSeconds}s`
    )
  }
  if (summary.latestAgeSeconds > config.maxTrafficAgeSeconds) {
    return 'stale'
  }
  return 'fresh'
}

async function inspectTraffic(config, nowMilliseconds, fetchImpl) {
  const response = await fetchWithTimeout(
    config.trafficAPIURL,
    {
      method: 'GET',
      headers: {
        Accept: 'application/json',
        'User-Agent': SERVICE_NAME,
      },
    },
    config.requestTimeoutMilliseconds,
    fetchImpl
  )

  if (!response.ok) {
    throw new Error(
      `production traffic request returned HTTP ${response.status}`
    )
  }

  const payload = await response.json()
  const summary = summarizeTrafficFreshness(payload, nowMilliseconds)
  return {
    classification: classifyTrafficFreshness(summary, config),
    summary,
  }
}

function scheduledTime(controller, nowMilliseconds) {
  if (
    controller !== null &&
    typeof controller === 'object' &&
    Number.isFinite(controller.scheduledTime)
  ) {
    return controller.scheduledTime
  }
  return nowMilliseconds
}

export async function executeScheduled(
  controller,
  environment,
  {
    fetchImpl = fetch,
    nowMilliseconds = Date.now(),
  } = {}
) {
  const config = loadConfig(environment)
  const effectiveNow = scheduledTime(controller, nowMilliseconds)
  const cron = controller?.cron

  if (cron !== config.primaryCron && cron !== config.watchdogCron) {
    throw new Error(`unsupported Cron Trigger ${JSON.stringify(cron)}`)
  }

  if (!config.dispatchEnabled) {
    return {
      marker: 'CLOUDFLARE_DISPATCH_DISABLED',
      status: 'PASS',
      cron,
      action: 'skipped-dispatch-disabled',
    }
  }

  if (cron === config.primaryCron) {
    const dispatch = await dispatchUnlessBlocked(
      config,
      DISPATCH_SOURCES.primary,
      effectiveNow,
      fetchImpl
    )
    return {
      marker: 'CLOUDFLARE_PRIMARY_SCHEDULE',
      status: 'PASS',
      cron,
      ...dispatch,
    }
  }

  if (cron === config.watchdogCron) {
    const freshness = await inspectTraffic(
      config,
      effectiveNow,
      fetchImpl
    )

    if (freshness.classification === 'fresh') {
      return {
        marker: 'CLOUDFLARE_WATCHDOG',
        status: 'PASS',
        cron,
        action: 'skipped-fresh-traffic',
        freshness: freshness.summary,
      }
    }

    const dispatch = await dispatchUnlessBlocked(
      config,
      DISPATCH_SOURCES.watchdog,
      effectiveNow,
      fetchImpl
    )
    return {
      marker: 'STALE_TRAFFIC_RECOVERY_DISPATCH',
      status: 'PASS',
      cron,
      freshness: freshness.summary,
      ...dispatch,
    }
  }

  throw new Error(`unsupported Cron Trigger ${JSON.stringify(cron)}`)
}

function safeErrorMessage(error) {
  if (error instanceof Error && error.message.trim() !== '') {
    return error.message
  }
  return 'unknown error'
}

function JSONResponse(payload, status) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      'Cache-Control': 'no-store',
      'Content-Type': 'application/json; charset=utf-8',
      'X-Content-Type-Options': 'nosniff',
    },
  })
}

export async function handleFetch(request, environment) {
  const url = new URL(request.url)
  if (request.method !== 'GET' || url.pathname !== '/health') {
    return JSONResponse({ error: 'NOT_FOUND' }, 404)
  }

  try {
    const config = loadConfig(environment)
    return JSONResponse(
      {
        service: SERVICE_NAME,
        status: 'ok',
        github_repository:
          `${config.githubOwner}/${config.githubRepository}`,
        github_workflow: config.githubWorkflow,
        github_ref: config.githubRef,
        primary_cron: config.primaryCron,
        watchdog_cron: config.watchdogCron,
        traffic_api_host: config.trafficAPIURL.host,
        github_token_configured: true,
      },
      200
    )
  } catch (error) {
    return JSONResponse(
      {
        service: SERVICE_NAME,
        status: 'misconfigured',
        error: safeErrorMessage(error),
      },
      503
    )
  }
}

export default {
  async scheduled(controller, environment) {
    try {
      const result = await executeScheduled(controller, environment)
      console.log(JSON.stringify(result))
    } catch (error) {
      console.error(
        JSON.stringify({
          marker: 'CLOUDFLARE_INGESTION_RELIABILITY',
          status: 'FAIL',
          error: safeErrorMessage(error),
        })
      )
      throw error
    }
  },

  fetch(request, environment) {
    return handleFetch(request, environment)
  },
}
