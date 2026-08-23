#!/usr/bin/env node
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

export const repositoryFiles = Object.freeze({
  readme: 'README.md',
  implementationSequence: 'docs/25_IMPLEMENTATION_SEQUENCE.md',
  completionDocument:
    'docs/184_STAGE_13_FRONTEND_ANALYTICS_INTEGRATION_COMPLETION.md',
  documentIndex: 'docs/DOCUMENT_INDEX.md',
  dashboard: 'apps/web/components/traffic-dashboard.tsx',
  trafficMap: 'apps/web/components/map/traffic-map.tsx',
  packageJSON: 'package.json',
  releaseVerifier: 'scripts/verify-release.sh',
  frontendWorkflow: '.github/workflows/frontend-ci.yml',
})

function exists(root, relativePath) {
  return fs.existsSync(path.join(root, relativePath))
}

function read(root, relativePath) {
  return fs.readFileSync(path.join(root, relativePath), 'utf8')
}

function stage13Section(source) {
  const startMarker = '## 15. Stage 13 — Frontend Analytics Integration'
  const endMarker = '\n---\n\n## 16. First Implementation Slice'
  const start = source.indexOf(startMarker)
  const end = source.indexOf(endMarker, start)

  if (start < 0 || end < 0) {
    return null
  }

  return source.slice(start, end)
}

function requireText(errors, source, needle, message) {
  if (!source.includes(needle)) {
    errors.push(message)
  }
}

export function validateRepository(root) {
  const errors = []

  for (const relativePath of Object.values(repositoryFiles)) {
    if (!exists(root, relativePath)) {
      errors.push(`missing required file: ${relativePath}`)
    }
  }

  if (errors.length > 0) {
    return errors
  }

  const implementationSequence = read(
    root,
    repositoryFiles.implementationSequence
  )
  const stage13 = stage13Section(implementationSequence)

  if (stage13 === null) {
    errors.push('implementation sequence is missing the bounded Stage 13 section')
  } else {
    requireText(
      errors,
      stage13,
      'Status: COMPLETED on 2026-08-07.',
      'Stage 13 status must remain completed on 2026-08-07'
    )
    requireText(
      errors,
      stage13,
      'docs/184_STAGE_13_FRONTEND_ANALYTICS_INTEGRATION_COMPLETION.md',
      'Stage 13 must reference Document 184 completion evidence'
    )

    for (const slice of [
      'Stage 13.1 — Projection Intelligence Frontend Foundation',
      'Stage 13.2 — Projection Map Visualization',
      'Stage 13.3 — Weather Context Frontend Foundation',
      'Stage 13.4 — Stability and Explainability Frontend Foundation',
    ]) {
      requireText(
        errors,
        stage13,
        slice,
        `Stage 13 completion is missing ${slice}`
      )
    }

    if (stage13.includes('Status: IN PROGRESS')) {
      errors.push('Stage 13 must not regress to IN PROGRESS')
    }
  }

  const readme = read(root, repositoryFiles.readme)
  for (const [needle, message] of [
    [
      '<!-- STAGE-13-FRONTEND-ANALYTICS-CLOSURE-V1 -->',
      'README is missing the Stage 13 closure marker',
    ],
    [
      '## Frontend Analytics Integration',
      'README is missing the frontend analytics integration section',
    ],
    [
      'docs/184_STAGE_13_FRONTEND_ANALYTICS_INTEGRATION_COMPLETION.md',
      'README is missing the Document 184 link',
    ],
    [
      'estimated projection geometry remain separate',
      'README must preserve observed-versus-estimated geometry disclosure',
    ],
    [
      'later visual and interaction redesign',
      'README must record the later visual redesign without rewriting Stage 13 completion',
    ],
    [
      'Visual Polish V2',
      'README must record the post-closure visual polish without rewriting Stage 13 completion',
    ],
  ]) {
    requireText(errors, readme, needle, message)
  }

  const completionDocument = read(
    root,
    repositoryFiles.completionDocument
  )
  for (const [needle, message] of [
    [
      '<!-- STAGE-13-FRONTEND-ANALYTICS-CLOSURE-V1 -->',
      'Document 184 is missing the closure marker',
    ],
    [
      'Status: COMPLETED on 2026-08-07',
      'Document 184 must record the exact completion date',
    ],
    [
      'FRONTEND_ANALYTICAL_RECOMPUTATION=PROHIBITED',
      'Document 184 must preserve the frontend recomputation boundary',
    ],
    [
      'OBSERVED_PROJECTED_GEOMETRY_SEPARATION=PRESERVED',
      'Document 184 must preserve map evidence separation',
    ],
    [
      'FRONTEND_VISUAL_REDESIGN=SEPARATE_PHASE',
      'Document 184 must keep redesign as a separate phase',
    ],
  ]) {
    requireText(errors, completionDocument, needle, message)
  }

  const documentIndex = read(root, repositoryFiles.documentIndex)
  requireText(
    errors,
    documentIndex,
    '<!-- STAGE-13-FRONTEND-ANALYTICS-CLOSURE:DOCUMENT-INDEX -->',
    'documentation index is missing the Document 184 marker'
  )
  requireText(
    errors,
    documentIndex,
    '184_STAGE_13_FRONTEND_ANALYTICS_INTEGRATION_COMPLETION.md',
    'documentation index is missing Document 184'
  )

  const dashboard = read(root, repositoryFiles.dashboard)
  for (const token of [
    'ProjectionIntelligencePanel',
    'WeatherContextPanel',
    'StabilityIntelligencePanel',
    'useProjectionIntelligence',
    'useWeatherContext',
    'useStabilityIntelligence',
    '<ProjectionIntelligencePanel',
    '<WeatherContextPanel',
    '<StabilityIntelligencePanel',
    'projection={projectionQuery.data?.projection}',
  ]) {
    requireText(
      errors,
      dashboard,
      token,
      `traffic dashboard is missing Stage 13 runtime token ${token}`
    )
  }

  const trafficMap = read(root, repositoryFiles.trafficMap)
  const trajectorySource = trafficMap.match(
    /const trajectorySourceID = '([^']+)'/
  )?.[1]
  const projectionSource = trafficMap.match(
    /const projectionSourceID = '([^']+)'/
  )?.[1]

  if (!trajectorySource) {
    errors.push('traffic map is missing the observed trajectory source')
  }
  if (!projectionSource) {
    errors.push('traffic map is missing the estimated projection source')
  }
  if (
    trajectorySource &&
    projectionSource &&
    trajectorySource === projectionSource
  ) {
    errors.push(
      'observed trajectory and estimated projection sources must remain distinct'
    )
  }

  for (const token of [
    "uncertainty: 'selected-aircraft-projection-uncertainty'",
    "line: 'selected-aircraft-projection-line'",
    "points: 'selected-aircraft-projection-points'",
    'buildProjectionFeatureCollection',
    'Projected path',
    'Horizontal uncertainty',
  ]) {
    requireText(
      errors,
      trafficMap,
      token,
      `traffic map is missing projection evidence token ${token}`
    )
  }

  const packageJSON = JSON.parse(read(root, repositoryFiles.packageJSON))
  for (const scriptName of [
    'test:stage13-frontend-analytics-closure',
    'verify:stage13-frontend-analytics-closure',
  ]) {
    if (!packageJSON.scripts?.[scriptName]) {
      errors.push(`package.json is missing script ${scriptName}`)
    }
  }

  const releaseVerifier = read(root, repositoryFiles.releaseVerifier)
  for (const command of [
    'pnpm run test:stage13-frontend-analytics-closure',
    'pnpm run verify:stage13-frontend-analytics-closure',
  ]) {
    requireText(
      errors,
      releaseVerifier,
      command,
      `release verification is missing ${command}`
    )
  }

  const frontendWorkflow = read(root, repositoryFiles.frontendWorkflow)
  for (const token of [
    'scripts/verify-stage-13-frontend-analytics-closure.mjs',
    'scripts/verify-stage-13-frontend-analytics-closure.test.mjs',
    'docs/25_IMPLEMENTATION_SEQUENCE.md',
    'docs/184_STAGE_13_FRONTEND_ANALYTICS_INTEGRATION_COMPLETION.md',
    'docs/DOCUMENT_INDEX.md',
    'pnpm run test:stage13-frontend-analytics-closure',
    'pnpm run verify:stage13-frontend-analytics-closure',
  ]) {
    requireText(
      errors,
      frontendWorkflow,
      token,
      `Frontend CI is missing Stage 13 closure token ${token}`
    )
  }

  return errors
}

function main() {
  const errors = validateRepository(process.cwd())

  if (errors.length > 0) {
    for (const error of errors) {
      console.error(
        `STAGE_13_FRONTEND_ANALYTICS_CLOSURE=FAIL reason=${error}`
      )
    }
    process.exit(1)
  }

  console.log('STAGE_13_DOCUMENT_STATUS=COMPLETED')
  console.log('STAGE_13_FRONTEND_PANELS=PASS')
  console.log('STAGE_13_MAP_SOURCE_SEPARATION=PASS')
  console.log('STAGE_13_SCOPE_BOUNDARY=PASS')
  console.log('STAGE_13_FRONTEND_ANALYTICS_CLOSURE=PASS')
}

const currentFile = fileURLToPath(import.meta.url)
if (
  process.argv[1] &&
  path.resolve(process.argv[1]) === currentFile
) {
  main()
}
