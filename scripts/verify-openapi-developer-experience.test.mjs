import assert from 'node:assert/strict'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import test from 'node:test'
import { generateClientSource } from './generate-openapi-client.mjs'
import { validateRepository } from './verify-openapi-developer-experience.mjs'


const repositoryFixtureFiles = [
  'openapi/openapi.json',
  'apps/api/internal/http/apidocs/openapi.json',
  'apps/api/internal/http/apidocs/handler.go',
  'apps/api/internal/http/apidocs/handler_test.go',
  'apps/api/internal/http/apidocs/assets/index.html',
  'apps/api/internal/http/apidocs/assets/app.js',
  'apps/api/internal/http/apidocs/assets/app.css',
  'apps/api/internal/server/server.go',
  'packages/api-client/src/generated.ts',
  'packages/api-client/src/client.ts',
  'packages/api-client/package.json',
  'packages/api-client/tsconfig.json',
  'packages/api-client/tests/client.test.mjs',
  'package.json',
  'pnpm-lock.yaml',
  '.github/workflows/api-contract.yml',
  'scripts/verify-release.sh',
  'docs/181_OPENAPI_DEVELOPER_EXPERIENCE.md',
  'docs/DOCUMENT_INDEX.md',
]

function copyRepositoryFixture() {
  const temporaryRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'gfa-openapi-dx-'))
  for (const relativePath of repositoryFixtureFiles) {
    const source = path.join(process.cwd(), relativePath)
    const target = path.join(temporaryRoot, relativePath)
    fs.mkdirSync(path.dirname(target), { recursive: true })
    fs.copyFileSync(source, target)
  }
  return temporaryRoot
}

function fixtureSpec() {
  return {
    openapi: '3.1.0',
    info: { title: 'Fixture', version: '1.0.0' },
    paths: {
      '/api/v1/health': {
        get: {
          operationId: 'getHealth',
          responses: {
            200: {
              content: {
                'application/json': { schema: { $ref: '#/components/schemas/HealthResponse' } },
              },
            },
          },
        },
      },
      '/api/v1/items/{id}': {
        post: {
          operationId: 'createItem',
          security: [{ InternalAPIKey: [] }],
          parameters: [
            { name: 'id', in: 'path', required: true, schema: { type: 'string' } },
            { name: 'limit', in: 'query', required: false, schema: { type: ['integer', 'null'] } },
          ],
          responses: {
            200: {
              content: {
                'application/json': { schema: { type: 'array', items: { $ref: '#/components/schemas/Item' } } },
              },
            },
          },
        },
      },
    },
    components: {
      schemas: {
        HealthResponse: {
          type: 'object',
          additionalProperties: false,
          required: ['status'],
          properties: { status: { type: 'string', enum: ['ok'] } },
        },
        Item: {
          type: 'object',
          additionalProperties: false,
          required: ['id'],
          properties: {
            id: { type: 'string' },
            elevation: { type: ['number', 'null'] },
          },
        },
      },
    },
  }
}

test('generator is deterministic and preserves schema nullability', () => {
  const spec = fixtureSpec()
  const bytes = Buffer.from(`${JSON.stringify(spec, null, 2)}\n`)
  const first = generateClientSource(spec, bytes)
  const second = generateClientSource(spec, bytes)
  assert.equal(first, second)
  assert.match(first, /export type Item = \{[\s\S]*readonly elevation\?: number \| null/)
  assert.match(first, /readonly createItem: \{[\s\S]*readonly path:/)
  assert.match(first, /protected: true/)
})

test('generator emits typed operation responses and metadata', () => {
  const output = generateClientSource(fixtureSpec())
  assert.match(output, /readonly getHealth: HealthResponse/)
  assert.match(output, /readonly createItem: ReadonlyArray<Item>/)
  assert.match(output, /path: "\/api\/v1\/items\/\{id\}"/)
})

test('repository verifier accepts the current checkout', () => {
  assert.deepEqual(validateRepository(process.cwd()), [])
})

test('repository verifier rejects embedded specification drift', () => {
  const temporaryRoot = copyRepositoryFixture()
  fs.appendFileSync(path.join(temporaryRoot, 'apps/api/internal/http/apidocs/openapi.json'), '\n')
  assert(validateRepository(temporaryRoot).some((error) => error.includes('byte-identical')))
})

test('repository verifier rejects browser credential persistence', () => {
  const temporaryRoot = copyRepositoryFixture()
  fs.appendFileSync(path.join(temporaryRoot, 'apps/api/internal/http/apidocs/assets/app.js'), '\nlocalStorage.setItem("key", "value")\n')
  assert(validateRepository(temporaryRoot).some((error) => error.includes('localStorage')))
})
