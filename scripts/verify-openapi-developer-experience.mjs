#!/usr/bin/env node
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { generateFromFiles } from './generate-openapi-client.mjs'

const files = Object.freeze({
  rootSpec: 'openapi/openapi.json',
  embeddedSpec: 'apps/api/internal/http/apidocs/openapi.json',
  handler: 'apps/api/internal/http/apidocs/handler.go',
  handlerTest: 'apps/api/internal/http/apidocs/handler_test.go',
  html: 'apps/api/internal/http/apidocs/assets/index.html',
  javascript: 'apps/api/internal/http/apidocs/assets/app.js',
  stylesheet: 'apps/api/internal/http/apidocs/assets/app.css',
  server: 'apps/api/internal/server/server.go',
  generated: 'packages/api-client/src/generated.ts',
  client: 'packages/api-client/src/client.ts',
  packageJSON: 'packages/api-client/package.json',
  packageTSConfig: 'packages/api-client/tsconfig.json',
  packageTest: 'packages/api-client/tests/client.test.mjs',
  rootPackageJSON: 'package.json',
  lockfile: 'pnpm-lock.yaml',
  workflow: '.github/workflows/api-contract.yml',
  release: 'scripts/verify-release.sh',
  document: 'docs/181_OPENAPI_DEVELOPER_EXPERIENCE.md',
  index: 'docs/DOCUMENT_INDEX.md',
})

function read(root, relativePath) {
  return fs.readFileSync(path.join(root, relativePath), 'utf8')
}

function exists(root, relativePath) {
  return fs.existsSync(path.join(root, relativePath))
}

function operationCount(spec) {
  const methods = new Set(['get', 'post', 'put', 'patch', 'delete'])
  let count = 0
  for (const pathItem of Object.values(spec.paths ?? {})) {
    for (const method of Object.keys(pathItem ?? {})) if (methods.has(method)) count += 1
  }
  return count
}

export function validateRepository(root) {
  const errors = []
  for (const relativePath of Object.values(files)) {
    if (!exists(root, relativePath)) errors.push(`missing required file: ${relativePath}`)
  }
  if (errors.length > 0) return errors

  const rootSpecBytes = fs.readFileSync(path.join(root, files.rootSpec))
  const embeddedSpecBytes = fs.readFileSync(path.join(root, files.embeddedSpec))
  if (!rootSpecBytes.equals(embeddedSpecBytes)) {
    errors.push('embedded OpenAPI contract must be byte-identical to openapi/openapi.json')
  }

  const spec = JSON.parse(rootSpecBytes.toString('utf8'))
  if (spec.openapi !== '3.1.0') errors.push(`OpenAPI version must remain 3.1.0; found ${spec.openapi}`)
  if (operationCount(spec) !== 38) errors.push(`OpenAPI operation count must remain 38; found ${operationCount(spec)}`)

  const server = read(root, files.server)
  if (!server.includes('internal/http/apidocs')) errors.push('server must import the API documentation package')
  if (!/apidocs\.Register\s*\(\s*app\s*,?\s*\)/.test(server)) errors.push('server must register the API documentation surface')
  if (server.indexOf('apidocs.Register') > server.indexOf('v1 := app.Group')) {
    errors.push('API documentation must be registered outside the /api/v1 production operation group')
  }

  const handler = read(root, files.handler)
  for (const route of ['/api/docs', '/api/docs/openapi.json', '/api/docs/assets/app.js', '/api/docs/assets/app.css']) {
    if (!handler.includes(route)) errors.push(`documentation handler is missing route ${route}`)
  }
  if (!handler.includes("script-src 'self'")) errors.push('documentation CSP must allow only same-origin scripts')
  if (!handler.includes("connect-src 'self'")) errors.push('documentation CSP must explicitly bound API requests')
  if (!handler.includes('X-Robots-Tag')) errors.push('documentation surface must remain noindex')
  if (!handler.includes('IfNoneMatch')) errors.push('OpenAPI and static assets must support conditional requests')

  const browser = read(root, files.javascript)
  for (const forbidden of ['localStorage', 'sessionStorage', 'document.cookie', 'internalApiKeyInput']) {
    if (browser.includes(forbidden)) errors.push(`browser explorer must not persist or collect credentials: ${forbidden}`)
  }
  if (!browser.includes('Protected mutation')) errors.push('browser explorer must explain the protected mutation boundary')
  if (!browser.includes("entry.method !== 'GET' || entry.protected")) {
    errors.push('browser explorer must execute only public GET operations')
  }

  const client = read(root, files.client)
  if (!client.includes("headers.set('X-Internal-API-Key', this.#internalApiKey)")) {
    errors.push('generated client runtime must attach the internal API key only for protected operations')
  }
  if (!client.includes('operationDefinitions[operationId]')) errors.push('client must execute from generated operation metadata')
  for (const forbidden of ['localStorage', 'sessionStorage', 'document.cookie']) {
    if (client.includes(forbidden)) errors.push(`client must not persist credentials: ${forbidden}`)
  }

  const generated = generateFromFiles({ root }).generated
  if (read(root, files.generated) !== generated) errors.push('generated TypeScript client contract is stale')
  const generatedSource = read(root, files.generated)
  const operationIds = Object.values(spec.paths ?? {}).flatMap((pathItem) =>
    Object.entries(pathItem ?? {})
      .filter(([method]) => ['get', 'post', 'put', 'patch', 'delete'].includes(method))
      .map(([, operation]) => operation.operationId),
  )
  for (const operationId of operationIds) {
    if (typeof operationId === 'string' && !generatedSource.includes(`${operationId}:`)) {
      errors.push(`generated client is missing operationId ${operationId}`)
    }
  }

  const packageJSON = JSON.parse(read(root, files.rootPackageJSON))
  for (const script of [
    'generate:openapi-client',
    'test:openapi-developer-experience',
    'verify:openapi-developer-experience',
    'typecheck:api-client',
    'test:api-client',
  ]) {
    if (!packageJSON.scripts?.[script]) errors.push(`root package.json is missing script ${script}`)
  }

  const lockfile = read(root, files.lockfile)
  if (!/\n  packages\/api-client:\n/.test(lockfile)) errors.push('pnpm lockfile is missing packages/api-client importer')
  if (!lockfile.includes('version: 5.9.3')) errors.push('api-client importer must resolve the existing TypeScript 5.9.3 lock entry')

  const workflow = read(root, files.workflow)
  for (const needle of [
    'verify-openapi-developer-experience.test.mjs',
    'verify-openapi-developer-experience.mjs',
    'generate-openapi-client.mjs --check',
    'pnpm --dir packages/api-client typecheck',
    'pnpm --dir packages/api-client test',
    'docs/181_OPENAPI_DEVELOPER_EXPERIENCE.md',
  ]) {
    if (!workflow.includes(needle)) errors.push(`OpenAPI workflow is missing ${needle}`)
  }

  const release = read(root, files.release)
  if (!release.includes('test:openapi-developer-experience')) errors.push('release verification must run API developer experience tests')
  if (!release.includes('verify:openapi-developer-experience')) errors.push('release verification must run API developer experience verification')
  if (!release.includes('typecheck:api-client')) errors.push('release verification must typecheck the generated API client')
  if (!release.includes('test:api-client')) errors.push('release verification must run generated API client runtime tests')

  if (!read(root, files.index).includes('OPENAPI-DEVELOPER-EXPERIENCE:DOCUMENT-INDEX')) {
    errors.push('documentation index is missing Document 181 marker')
  }

  return errors
}

function main() {
  const errors = validateRepository(process.cwd())
  if (errors.length > 0) {
    for (const error of errors) console.error(`OPENAPI_DEVELOPER_EXPERIENCE=FAIL reason=${error}`)
    process.exit(1)
  }
  console.log('OPENAPI_DEVELOPER_DOCS_ROUTE=/api/docs')
  console.log('OPENAPI_DEVELOPER_SPEC_ROUTE=/api/docs/openapi.json')
  console.log('OPENAPI_GENERATED_CLIENT_OPERATIONS=38')
  console.log('OPENAPI_BROWSER_MUTATION_EXECUTION=DISABLED')
  console.log('OPENAPI_EMBEDDED_SPEC_DRIFT=PASS')
  console.log('OPENAPI_GENERATED_CLIENT_DRIFT=PASS')
  console.log('OPENAPI_DEVELOPER_EXPERIENCE=PASS')
}

const currentFile = fileURLToPath(import.meta.url)
if (process.argv[1] && path.resolve(process.argv[1]) === currentFile) main()
