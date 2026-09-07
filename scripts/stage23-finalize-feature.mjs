#!/usr/bin/env node
import fs from 'node:fs'

function read(path) { return fs.readFileSync(path, 'utf8') }
function write(path, content) { fs.writeFileSync(path, content) }
function mustReplace(content, needle, replacement, label) {
  if (!content.includes(needle)) throw new Error(`missing ${label}`)
  return content.replace(needle, replacement)
}

// 1. Canonical Playwright mock surface.
{
  const path = 'apps/web/e2e/mock-api.mjs'
  let source = read(path)
  source = mustReplace(
    source,
    "  '/api/v1/trajectories/{id}/projection-intelligence',\n",
    "  '/api/v1/trajectories/{id}/projection-intelligence',\n  '/api/v1/trajectories/{id}/eta-reliability',\n",
    'ETA Reliability OpenAPI mock path',
  )
  source = mustReplace(
    source,
    "  'intelligence-error',\n])",
    "  'intelligence-error',\n  'eta-reliability',\n])",
    'ETA Reliability scenario',
  )
  source = mustReplace(
    source,
    "const routeIntelligenceTrajectoryID =\n",
    `const etaReliability = {\n  version: 'eta-reliability-v1',\n  status: 'limited',\n  trajectory_id: advancedTrajectoryID,\n  route: { origin_icao_code: 'UBBB', destination_icao_code: 'LTFM' },\n  method: { name: 'continuation', version: 'v1', decision_class: 'derived' },\n  target_lead_seconds: 1800,\n  lead_tolerance_seconds: 300,\n  endpoint_radius_km: 25,\n  candidate_count: 8,\n  eligible_sample_count: 6,\n  metrics: {\n    sample_count: 6,\n    median_absolute_error_seconds: 258,\n    p80_absolute_error_seconds: 462,\n    within_five_minutes_ratio: 0.667,\n    within_ten_minutes_ratio: 0.833,\n    interval_coverage_ratio: 0.833,\n  },\n  evidence_class: 'historically_recomputed_from_persisted_observations_with_endpoint_proxy',\n  limitations: [\n    {\n      code: 'trajectory_endpoint_arrival_proxy',\n      message: 'Historical arrival truth uses the last persisted trajectory observation within 25 km of LTFM; it is not an official touchdown, gate or schedule timestamp.',\n    },\n    {\n      code: 'limited_sample_size',\n      message: 'Reliability is based on 6 samples; more evidence is required for complete status.',\n    },\n  ],\n  input_fingerprint: \`sha256:\${'f'.repeat(64)}\`,\n  generated_at: '2026-08-04T18:00:07Z',\n}\n\nconst routeIntelligenceTrajectoryID =\n`,
    'ETA Reliability fixture insertion point',
  )
  source = mustReplace(
    source,
    `  if (/^\\/api\\/v1\\/trajectories\\/[^/]+\\/projection-intelligence$/.test(pathname)) {\n    return '/api/v1/trajectories/{id}/projection-intelligence'\n  }\n`,
    `  if (/^\\/api\\/v1\\/trajectories\\/[^/]+\\/projection-intelligence$/.test(pathname)) {\n    return '/api/v1/trajectories/{id}/projection-intelligence'\n  }\n  if (/^\\/api\\/v1\\/trajectories\\/[^/]+\\/eta-reliability$/.test(pathname)) {\n    return '/api/v1/trajectories/{id}/eta-reliability'\n  }\n`,
    'ETA Reliability path normalization',
  )
  source = mustReplace(
    source,
    `  if (\n    method === 'GET' &&\n    normalizePath(pathname) ===\n      '/api/v1/trajectories/{id}/stability-intelligence'\n  ) {\n`,
    `  if (\n    method === 'GET' &&\n    normalizePath(pathname) === '/api/v1/trajectories/{id}/eta-reliability'\n  ) {\n    return success(etaReliability)\n  }\n  if (\n    method === 'GET' &&\n    normalizePath(pathname) ===\n      '/api/v1/trajectories/{id}/stability-intelligence'\n  ) {\n`,
    'ETA Reliability mock response',
  )
  write(path, source)
}

// 2. Dedicated browser product journey.
{
  const path = 'apps/web/e2e/tests/advanced-intelligence.spec.mjs'
  let source = read(path).trimEnd()
  if (!source.includes("test('ETA reliability exposes bounded historical endpoint-proxy evidence")) {
    source += `\n\n\ntest('ETA reliability exposes bounded historical endpoint-proxy evidence', async ({\n  page,\n  request,\n}) => {\n  await setScenario(request, 'eta-reliability')\n  await page.goto(\n    '/?region=az&aircraft=4b1801&view=intelligence#live-traffic',\n    { waitUntil: 'domcontentloaded' },\n  )\n\n  await expect(\n    page.getByText('Historical ETA Reliability', { exact: true }),\n  ).toBeVisible()\n  await expect(\n    page.getByRole('heading', {\n      name: 'How reliable have comparable ETA estimates been?',\n    }),\n  ).toBeVisible()\n  await expect(page.getByText('6 eligible / 8 checked', { exact: true })).toBeVisible()\n  await expect(page.getByText('Median ETA error', { exact: true })).toBeVisible()\n  await expect(page.getByText('4m 18s', { exact: true })).toBeVisible()\n  await expect(page.getByText('80% error threshold', { exact: true })).toBeVisible()\n  await expect(page.getByText('≤ 7m 42s', { exact: true })).toBeVisible()\n  await expect(page.getByText('Within ±5 min', { exact: true })).toBeVisible()\n  await expect(page.getByText('67%', { exact: true })).toBeVisible()\n  await expect(page.getByText('Within ±10 min', { exact: true })).toBeVisible()\n  await expect(page.getByText('83%', { exact: true }).first()).toBeVisible()\n  await expect(\n    page.getByText('ETA window covered endpoint', { exact: true }),\n  ).toBeVisible()\n  await expect(page.getByText('Evidence boundary', { exact: true })).toBeVisible()\n  await expect(\n    page.getByText(/observed endpoint proxy, not an official touchdown, gate or schedule timestamp/i),\n  ).toBeVisible()\n})\n`
  }
  write(path, source)
}

// 3. Playwright foundation contract.
{
  const path = 'scripts/verify-playwright-e2e.mjs'
  let source = read(path)
  source = mustReplace(source, 'export const expectedBrowserScenarioCount = 20', 'export const expectedBrowserScenarioCount = 21', 'browser scenario count')
  source = mustReplace(source, "  'intelligence-error',\n])", "  'intelligence-error',\n  'eta-reliability',\n])", 'expected mock scenario')
  source = mustReplace(
    source,
    "      'Attribution and scope guards',\n",
    "      'Attribution and scope guards',\n      'Historical ETA Reliability',\n      '6 eligible / 8 checked',\n      'Evidence boundary',\n",
    'advanced intelligence Stage 23 literals',
  )
  source = mustReplace(
    source,
    "      \"'/api/v1/trajectories/{id}/stability-intelligence'\",\n",
    "      \"'/api/v1/trajectories/{id}/stability-intelligence'\",\n      \"'/api/v1/trajectories/{id}/eta-reliability'\",\n      \"'eta-reliability'\",\n",
    'mock Stage 23 literals',
  )
  write(path, source)
}

// 4. Playwright verifier tests.
{
  const path = 'scripts/verify-playwright-e2e.test.mjs'
  let source = read(path)
  source = mustReplace(source, 'assert.equal(expectedBrowserScenarioCount, 20)', 'assert.equal(expectedBrowserScenarioCount, 21)', 'test expected browser count')
  source = mustReplace(source, 'assert.equal(countBrowserScenarios(process.cwd()), 20)', 'assert.equal(countBrowserScenarios(process.cwd()), 21)', 'test actual browser count')
  source = mustReplace(
    source,
    "      'intelligence-error',\n      'regions-error',",
    "      'intelligence-error',\n      'eta-reliability',\n      'regions-error',",
    'test supported scenarios',
  )
  const insertion = `\n\ntest('ETA Reliability fixture preserves bounded endpoint-proxy evidence', () => {\n  const result = resolveMockResponse({\n    method: 'GET',\n    requestURL:\n      'http://127.0.0.1:8091/api/v1/trajectories/11111111-1111-4111-8111-111111111111/eta-reliability?as_of_time=2026-08-04T18:00:00Z&duration_seconds=300',\n    scenario: 'eta-reliability',\n  })\n  assert.equal(result.status, 200)\n  assert.equal(result.body.data.status, 'limited')\n  assert.equal(result.body.data.candidate_count, 8)\n  assert.equal(result.body.data.eligible_sample_count, 6)\n  assert.equal(result.body.data.metrics.median_absolute_error_seconds, 258)\n  assert.equal(result.body.data.metrics.p80_absolute_error_seconds, 462)\n  assert.equal(\n    result.body.data.evidence_class,\n    'historically_recomputed_from_persisted_observations_with_endpoint_proxy',\n  )\n  assert.match(result.body.data.limitations[0].message, /not an official touchdown, gate or schedule timestamp/i)\n})\n`
  if (!source.includes("test('ETA Reliability fixture preserves bounded endpoint-proxy evidence'")) {
    source = mustReplace(
      source,
      "\ntest('Stability Intelligence fixture mirrors requested analytical timestamps'",
      `${insertion}\ntest('Stability Intelligence fixture mirrors requested analytical timestamps'`,
      'ETA Reliability fixture test insertion',
    )
  }
  write(path, source)
}

// 5. Implementation sequence: close V2 reconciliation and explicitly register Stage 23.
{
  const path = 'docs/25_IMPLEMENTATION_SEQUENCE.md'
  let source = read(path)
  source = mustReplace(source, 'Status: Implementation Baseline v1.7', 'Status: Implementation Baseline v1.8', 'implementation sequence version')
  source = mustReplace(
    source,
    'Status: RECONCILIATION CANDIDATE. Canonical closure requires exact-head validation, merge, and independent post-merge validation as defined by Document 212.',
    'Status: CLOSED / CI_VERIFIED. Documents 212 and 213 own the reconciliation decision and independent post-merge closure evidence.',
    'V2 reconciliation status',
  )
  source = mustReplace(source, 'VERSION_2_RECONCILIATION=CANDIDATE', 'VERSION_2_RECONCILIATION=CLOSED', 'V2 reconciliation marker')
  if (!source.includes('<!-- STAGE-23-ETA-RELIABILITY:IMPLEMENTATION -->')) {
    source = source.trimEnd() + `\n\n<!-- STAGE-23-ETA-RELIABILITY:IMPLEMENTATION -->\n\n## Stage 23 — ETA Reliability Intelligence\n\nStatus: IN PROGRESS on draft PR #171.\n\nStage 23 is an explicit post-Version-2 product decision; it was not created by the Version 2 reconciliation itself. It is allowed only as a zero-budget vertical slice with an existing frontend consumer.\n\n\`\`\`text\npersisted observations\n        ↓\nproduction Projection Intelligence recomputed at persisted historical as_of_time\n        ↓\nbounded ETA reliability aggregation (maximum 8 historical candidates)\n        ↓\nGET /api/v1/trajectories/{id}/eta-reliability\n        ↓\nOpenAPI + generated TypeScript client + TanStack Query\n        ↓\nAircraft Detail / Estimated Arrival / Historical ETA Reliability UI\n\`\`\`\n\nEvidence and cost boundaries:\n\n\`\`\`text\nSTAGE_23_ADDITIONAL_COST=0_RUB\nSTAGE_23_FRONTEND_CONSUMER=YES\nSTAGE_23_OFFICIAL_ARRIVAL_TRUTH=NONE\nSTAGE_23_HISTORICAL_ARRIVAL_EVIDENCE=PERSISTED_ENDPOINT_PROXY\nSTAGE_23_MAX_HISTORICAL_CANDIDATES=8\nSTAGE_23_FRONTEND_POLLING=NONE\nSTAGE_23_NEW_DATABASE_TABLE=NONE\nSTAGE_23_NEW_MIGRATION=NONE\nSTAGE_23_NEW_PAID_PROVIDER=NONE\n\`\`\`\n\nThe production server must not depend on the offline-only Projection Evaluation package. ETA Reliability reuses production Projection Intelligence and computes only the bounded arrival-error and interval-coverage definitions required by the user-facing feature.\n\nDocument 214 is the canonical Stage 23 pre-merge engineering record. No Stage 24 follows automatically; any later increment requires a new product decision with the same zero-cost and frontend-value gate.\n`
  }
  write(path, source)
}

// 6. README current feature-branch contract truth.
{
  const path = 'README.md'
  let source = read(path)
  source = source.replace(
    /Stage 23 reuses the existing Go\/PostgreSQL Projection Intelligence and Projection Evaluation\nstack, Next\.js and TanStack Query\./,
    'Stage 23 reuses the existing Go/PostgreSQL Production Projection Intelligence, Next.js and TanStack Query. The production server does not depend on the offline-only Projection Evaluation package.',
  )
  source = source.replace(
    /The current source\nroute is still being reconciled with the canonical OpenAPI contract, so the repository-wide\npublic OpenAPI surface remains the previously closed 39-operation contract until that sync\npasses exact-head validation\./,
    'The feature branch now carries a source-backed 40-operation OpenAPI candidate (39 public GET reads plus one protected Route Intelligence POST) and a regenerated TypeScript client. Canonical `main` remains on the previously closed 39-operation contract until PR #171 is merged and independently verified.',
  )
  source = source.replace(
    /The current \*\*closed canonical\*\* repository contract exposes 39 source-backed OpenAPI\noperations: 38 unauthenticated public read operations and one protected Route Intelligence\nmutation\. Stage 23 has added a source route while its OpenAPI reconciliation is still in\nprogress, so the feature must not be represented as a closed 40-operation public contract\nuntil source\/OpenAPI\/generated-client gates agree on one exact head\./,
    'Canonical `main` currently exposes the closed 39-operation contract (38 public GET reads and one protected Route Intelligence POST). The Stage 23 feature branch candidate exposes 40 source-backed operations (39 public GET reads and the same protected POST); that 40-operation surface is not canonical until the feature merges and independent post-merge validation succeeds.',
  )
  write(path, source)
}

// 7. Document 214: align with production-safe architecture and completed contract/E2E implementation.
{
  const path = 'docs/214_STAGE_23_ETA_RELIABILITY_INTELLIGENCE.md'
  let source = read(path)
  source = source.replace('- the existing Projection Evaluation package;\n', '')
  source = source.replace(
    'The service composes existing Projection Intelligence and Projection Evaluation behavior instead of introducing a second forecasting engine.',
    'The service reuses production Projection Intelligence historical recomputation and computes only the bounded ETA absolute-error and interval-coverage definitions needed by the product. The production runtime does not import the offline-only Projection Evaluation package.',
  )
  source = source.replace(
    'The public contract must be added to both canonical OpenAPI copies and the generated TypeScript client before Stage 23 can become review-ready.\n\nAt the current Stage 23 documentation state, that OpenAPI synchronization remains incomplete and therefore Stage 23 remains `IN_PROGRESS`.',
    'The feature branch now contains the source-backed 40-operation OpenAPI candidate, byte-identical root/embedded specifications and a regenerated TypeScript client exposing `getETAReliabilityByTrajectoryID`. Canonical `main` remains on the prior 39-operation contract until merge and independent post-merge verification. Stage 23 remains `IN_PROGRESS` until final exact-head validation succeeds.',
  )
  source = source.replace(
    'STAGE_23_ROADMAP=ALIGNED_V1_3\nSTAGE_23_DOCUMENTATION_REGRESSION_TEST=INSTALLED',
    'STAGE_23_ROADMAP=ALIGNED_V1_3\nSTAGE_23_IMPLEMENTATION_SEQUENCE=ALIGNED_V1_8\nSTAGE_23_OPENAPI_CANDIDATE_OPERATIONS=40\nSTAGE_23_OPENAPI_CANDIDATE_GET_OPERATIONS=39\nSTAGE_23_GENERATED_CLIENT=ALIGNED\nSTAGE_23_DEDICATED_PLAYWRIGHT_JOURNEY=INSTALLED\nSTAGE_23_DOCUMENTATION_REGRESSION_TEST=INSTALLED',
  )
  source = source.replace(
    'The Version 2 reconciliation subsection of Document 25 was written before Stage 23 existed and states that the reconciliation itself did not create Stage 23. That sentence remains historically true about the reconciliation operation. The later explicit product decision that authorizes Stage 23 is owned by Document 24 Section 21 and this Document 214; it must not be misread as a retroactive claim that Version 2 reconciliation created Stage 23.',
    'Document 25 is aligned to Implementation Baseline v1.8: Version 2 reconciliation is closed, the historical statement that reconciliation itself did not create Stage 23 is preserved, and a later append-only Stage 23 section records the explicit zero-cost frontend product decision.',
  )
  const oldRemaining = `The stage remains blocked on all of the following:\n\n1. run \`gofmt\`-equivalent formatting over every changed Go file and revalidate;\n2. synchronize canonical OpenAPI public route/schema contract;\n3. keep root and embedded OpenAPI byte-identical;\n4. regenerate the canonical TypeScript API client from OpenAPI;\n5. update route inventory/count assertions from the previous surface to the new source-backed surface;\n6. add dedicated production E2E/mock assertions for the ETA Reliability user path rather than relying only on unrelated Playwright success;\n7. complete a full exact-head GitHub CI matrix and Vercel check;\n8. verify review threads/reviews and mergeability on the exact final head.\n\nThe Stage 23 documentation/claim-boundary regression test, README alignment, Document Index registration and roadmap registration are already present and are no longer listed as unfinished engineering work.`
  const newRemaining = `The implementation, formatting, source/OpenAPI/generated-client contract synchronization and dedicated ETA Reliability browser journey are now present on the feature branch. The remaining review-ready work is evidence closure:\n\n1. complete a full exact-head GitHub CI matrix and Vercel check;\n2. remediate any SHA-specific failure without transferring earlier pass evidence;\n3. verify review threads/reviews and mergeability on the exact final head;\n4. align this document and the PR body to the final exact-head evidence, then revalidate that final documentation SHA.\n\nThe Stage 23 documentation/claim-boundary regression test, README, Document Index, roadmap and Implementation Sequence are aligned to the in-progress feature.`
  source = mustReplace(source, oldRemaining, newRemaining, 'Document 214 remaining work')
  write(path, source)
}

console.log('STAGE_23_FEATURE_FINALIZATION=PASS')
