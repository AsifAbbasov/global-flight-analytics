#!/usr/bin/env node
import crypto from 'node:crypto'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const DEFAULT_SPEC_PATH = 'openapi/openapi.json'
const DEFAULT_OUTPUT_PATH = 'packages/api-client/src/generated.ts'
const HTTP_METHODS = new Set(['get', 'post', 'put', 'patch', 'delete'])

function readJSON(filePath) {
  return JSON.parse(fs.readFileSync(filePath, 'utf8'))
}

function stableJSONString(value) {
  return JSON.stringify(value)
}

function localRefName(reference) {
  const prefix = '#/components/schemas/'
  if (!reference.startsWith(prefix)) {
    throw new Error(`unsupported non-schema reference: ${reference}`)
  }
  return decodeURIComponent(reference.slice(prefix.length).replace(/~1/g, '/').replace(/~0/g, '~'))
}

function resolveLocalReference(spec, reference) {
  if (!reference.startsWith('#/')) {
    throw new Error(`external references are not supported: ${reference}`)
  }
  let current = spec
  for (const rawSegment of reference.slice(2).split('/')) {
    const segment = rawSegment.replace(/~1/g, '/').replace(/~0/g, '~')
    if (!current || typeof current !== 'object' || !(segment in current)) {
      throw new Error(`unresolved local reference: ${reference}`)
    }
    current = current[segment]
  }
  return current
}

function propertyKey(name) {
  return /^[A-Za-z_$][A-Za-z0-9_$]*$/.test(name) ? name : stableJSONString(name)
}

function indent(text, spaces = 2) {
  const prefix = ' '.repeat(spaces)
  return text.split('\n').map((line) => `${prefix}${line}`).join('\n')
}

function unique(values) {
  return [...new Set(values)]
}

export function renderSchemaType(schema, spec) {
  if (!schema || typeof schema !== 'object') return 'unknown'
  if (schema.$ref) return localRefName(schema.$ref)
  if (Object.hasOwn(schema, 'const')) return stableJSONString(schema.const)

  if (Array.isArray(schema.enum) && schema.enum.length > 0) {
    return unique(schema.enum.map((value) => stableJSONString(value))).join(' | ')
  }

  for (const keyword of ['oneOf', 'anyOf']) {
    if (Array.isArray(schema[keyword]) && schema[keyword].length > 0) {
      return unique(schema[keyword].map((entry) => renderSchemaType(entry, spec))).join(' | ')
    }
  }

  if (Array.isArray(schema.allOf) && schema.allOf.length > 0) {
    return unique(schema.allOf.map((entry) => renderSchemaType(entry, spec))).join(' & ')
  }

  const declaredTypes = Array.isArray(schema.type)
    ? schema.type
    : schema.type
      ? [schema.type]
      : []

  if (declaredTypes.length > 1) {
    return unique(declaredTypes.map((type) => renderSchemaType({ ...schema, type }, spec))).join(' | ')
  }

  const type = declaredTypes[0]
  let rendered
  switch (type) {
    case 'string':
      rendered = 'string'
      break
    case 'integer':
    case 'number':
      rendered = 'number'
      break
    case 'boolean':
      rendered = 'boolean'
      break
    case 'null':
      rendered = 'null'
      break
    case 'array':
      rendered = `ReadonlyArray<${renderSchemaType(schema.items, spec)}>`
      break
    case 'object':
    case undefined: {
      if (!schema.properties && !Object.hasOwn(schema, 'additionalProperties')) {
        rendered = 'unknown'
        break
      }
      const required = new Set(Array.isArray(schema.required) ? schema.required : [])
      const lines = []
      for (const [name, propertySchema] of Object.entries(schema.properties ?? {})) {
        const optional = required.has(name) ? '' : '?'
        lines.push(`readonly ${propertyKey(name)}${optional}: ${renderSchemaType(propertySchema, spec)}`)
      }
      if (schema.additionalProperties === true) {
        lines.push('readonly [key: string]: unknown')
      } else if (schema.additionalProperties && typeof schema.additionalProperties === 'object') {
        lines.push(`readonly [key: string]: ${renderSchemaType(schema.additionalProperties, spec)}`)
      }
      rendered = lines.length === 0
        ? 'Record<string, never>'
        : `{\n${indent(lines.join('\n'))}\n}`
      break
    }
    default:
      rendered = 'unknown'
  }

  if (schema.nullable === true && rendered !== 'null') return `${rendered} | null`
  return rendered
}

function resolveParameter(spec, parameter) {
  return parameter?.$ref ? resolveLocalReference(spec, parameter.$ref) : parameter
}

function collectParameters(spec, pathItem, operation) {
  const collected = new Map()
  for (const candidate of [...(pathItem.parameters ?? []), ...(operation.parameters ?? [])]) {
    const parameter = resolveParameter(spec, candidate)
    if (!parameter || typeof parameter !== 'object') continue
    const key = `${parameter.in}:${parameter.name}`
    collected.set(key, parameter)
  }
  return [...collected.values()]
}

function resolveResponse(spec, response) {
  return response?.$ref ? resolveLocalReference(spec, response.$ref) : response
}

function successResponseSchema(spec, operation) {
  const entries = Object.entries(operation.responses ?? {})
    .filter(([status]) => /^2\d\d$/.test(status))
    .sort(([left], [right]) => Number(left) - Number(right))
  if (entries.length === 0) return undefined
  const response = resolveResponse(spec, entries[0][1])
  return response?.content?.['application/json']?.schema
}

function requestBodySchema(spec, operation) {
  const requestBody = operation.requestBody?.$ref
    ? resolveLocalReference(spec, operation.requestBody.$ref)
    : operation.requestBody
  return requestBody?.content?.['application/json']?.schema
}

function groupParameterType(parameters, location, spec) {
  const selected = parameters.filter((parameter) => parameter.in === location)
  if (selected.length === 0) return undefined
  const lines = selected.map((parameter) => {
    const optional = parameter.required === true ? '' : '?'
    return `readonly ${propertyKey(parameter.name)}${optional}: ${renderSchemaType(parameter.schema, spec)}`
  })
  return `{\n${indent(lines.join('\n'))}\n}`
}

function operationParametersType(spec, pathItem, operation) {
  const parameters = collectParameters(spec, pathItem, operation)
  const fields = []
  for (const location of ['path', 'query', 'header']) {
    const type = groupParameterType(parameters, location, spec)
    if (!type) continue
    const required = parameters.some((parameter) => parameter.in === location && parameter.required === true)
    fields.push(`readonly ${location}${required ? '' : '?'}: ${type}`)
  }
  const bodySchema = requestBodySchema(spec, operation)
  if (bodySchema) {
    const required = operation.requestBody?.required === true
    fields.push(`readonly body${required ? '' : '?'}: ${renderSchemaType(bodySchema, spec)}`)
  }
  if (fields.length === 0) return '{}'
  return `{\n${indent(fields.join('\n'))}\n}`
}

function operationIsProtected(operation) {
  return Array.isArray(operation.security) && operation.security.some((requirement) =>
    requirement && typeof requirement === 'object' && Object.hasOwn(requirement, 'InternalAPIKey'),
  )
}

function operationMetadata(spec) {
  const operations = []
  for (const [routePath, pathItem] of Object.entries(spec.paths ?? {})) {
    for (const [method, operation] of Object.entries(pathItem ?? {})) {
      if (!HTTP_METHODS.has(method) || !operation || typeof operation !== 'object') continue
      if (typeof operation.operationId !== 'string' || operation.operationId.length === 0) {
        throw new Error(`${method.toUpperCase()} ${routePath} is missing operationId`)
      }
      const parameters = collectParameters(spec, pathItem, operation)
      operations.push({
        operationId: operation.operationId,
        method: method.toUpperCase(),
        path: routePath,
        protected: operationIsProtected(operation),
        parameters,
        parametersType: operationParametersType(spec, pathItem, operation),
        responseType: renderSchemaType(successResponseSchema(spec, operation), spec),
        hasBody: Boolean(requestBodySchema(spec, operation)),
      })
    }
  }
  operations.sort((left, right) => left.operationId.localeCompare(right.operationId))
  const duplicate = operations.find((entry, index) => operations.findIndex((candidate) => candidate.operationId === entry.operationId) !== index)
  if (duplicate) throw new Error(`duplicate operationId: ${duplicate.operationId}`)
  return operations
}

function renderSchemas(spec) {
  const schemas = Object.entries(spec.components?.schemas ?? {}).sort(([left], [right]) => left.localeCompare(right))
  return schemas.map(([name, schema]) => `export type ${name} = ${renderSchemaType(schema, spec)}`).join('\n\n')
}

function renderOperationMap(name, operations, select) {
  const lines = operations.map((operation) => `readonly ${propertyKey(operation.operationId)}: ${select(operation)}`)
  return `export interface ${name} {\n${indent(lines.join('\n'))}\n}`
}

function renderDefinitions(operations) {
  const entries = operations.map((operation) => {
    const parameters = operation.parameters.map((parameter) => ({
      name: parameter.name,
      in: parameter.in,
      required: parameter.required === true,
    }))
    return `${propertyKey(operation.operationId)}: {\n${indent([
      `method: ${stableJSONString(operation.method)},`,
      `path: ${stableJSONString(operation.path)},`,
      `protected: ${operation.protected ? 'true' : 'false'},`,
      `hasBody: ${operation.hasBody ? 'true' : 'false'},`,
      `parameters: ${JSON.stringify(parameters)},`,
    ].join('\n'), 4)}\n  }`
  })
  return `export const operationDefinitions = {\n${indent(entries.join(',\n'))}\n} as const satisfies Record<OperationId, OperationDefinition>`
}

export function generateClientSource(spec, sourceBytes = Buffer.from(JSON.stringify(spec))) {
  const operations = operationMetadata(spec)
  const specSHA256 = crypto.createHash('sha256').update(sourceBytes).digest('hex')
  const schemas = renderSchemas(spec)
  const parameterMap = renderOperationMap('OperationParameters', operations, (operation) => operation.parametersType)
  const responseMap = renderOperationMap('OperationResponses', operations, (operation) => operation.responseType)
  const definitions = renderDefinitions(operations)

  return `/* eslint-disable */\n// This file is generated from ${DEFAULT_SPEC_PATH}. Do not edit manually.\n// OpenAPI SHA-256: ${specSHA256}\n\n${schemas}\n\nexport interface OperationDefinition {\n  readonly method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'\n  readonly path: string\n  readonly protected: boolean\n  readonly hasBody: boolean\n  readonly parameters: ReadonlyArray<{\n    readonly name: string\n    readonly in: string\n    readonly required: boolean\n  }>\n}\n\n${parameterMap}\n\n${responseMap}\n\nexport type OperationId = keyof OperationParameters\n\n${definitions}\n`
}

export function generateFromFiles({ root = process.cwd(), specPath = DEFAULT_SPEC_PATH, outputPath = DEFAULT_OUTPUT_PATH } = {}) {
  const absoluteSpec = path.resolve(root, specPath)
  const absoluteOutput = path.resolve(root, outputPath)
  const sourceBytes = fs.readFileSync(absoluteSpec)
  const spec = JSON.parse(sourceBytes.toString('utf8'))
  const generated = generateClientSource(spec, sourceBytes)
  return { absoluteOutput, generated }
}

function parseArguments(argv) {
  const result = { mode: 'write', specPath: DEFAULT_SPEC_PATH, outputPath: DEFAULT_OUTPUT_PATH }
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index]
    if (argument === '--check') result.mode = 'check'
    else if (argument === '--write') result.mode = 'write'
    else if (argument === '--spec') result.specPath = argv[++index]
    else if (argument === '--output') result.outputPath = argv[++index]
    else throw new Error(`unknown argument: ${argument}`)
  }
  return result
}

function main() {
  const options = parseArguments(process.argv.slice(2))
  const { absoluteOutput, generated } = generateFromFiles(options)
  if (options.mode === 'check') {
    const current = fs.existsSync(absoluteOutput) ? fs.readFileSync(absoluteOutput, 'utf8') : ''
    if (current !== generated) {
      console.error(`OPENAPI_GENERATED_CLIENT=FAIL reason=${path.relative(process.cwd(), absoluteOutput)} is stale`)
      process.exit(1)
    }
    console.log('OPENAPI_GENERATED_CLIENT=PASS')
    return
  }
  fs.mkdirSync(path.dirname(absoluteOutput), { recursive: true })
  fs.writeFileSync(absoluteOutput, generated)
  console.log(`OPENAPI_GENERATED_CLIENT_WRITTEN=${path.relative(process.cwd(), absoluteOutput)}`)
}

const currentFile = fileURLToPath(import.meta.url)
if (process.argv[1] && path.resolve(process.argv[1]) === currentFile) main()
