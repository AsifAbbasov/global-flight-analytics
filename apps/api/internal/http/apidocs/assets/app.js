'use strict'

const state = {
  spec: null,
  operations: [],
  selected: null,
}

const baseURLInput = document.querySelector('#base-url')
const filterInput = document.querySelector('#operation-filter')
const statusNode = document.querySelector('#contract-status')
const listNode = document.querySelector('#operation-list')
const panelNode = document.querySelector('#operation-panel')

baseURLInput.value = window.location.origin
filterInput.addEventListener('input', renderOperationList)

function element(tag, options = {}) {
  const node = document.createElement(tag)
  if (options.className) node.className = options.className
  if (options.text !== undefined) node.textContent = options.text
  return node
}

function operationRows(spec) {
  const methods = new Set(['get', 'post', 'put', 'patch', 'delete'])
  const rows = []
  for (const [path, pathItem] of Object.entries(spec.paths ?? {})) {
    for (const [method, operation] of Object.entries(pathItem ?? {})) {
      if (!methods.has(method) || !operation || typeof operation !== 'object') continue
      rows.push({
        method: method.toUpperCase(),
        path,
        operation,
        pathItem,
        protected: Array.isArray(operation.security) && operation.security.some((entry) => entry && Object.hasOwn(entry, 'InternalAPIKey')),
      })
    }
  }
  return rows.sort((left, right) => left.path.localeCompare(right.path) || left.method.localeCompare(right.method))
}

function renderOperationList() {
  const query = filterInput.value.trim().toLowerCase()
  listNode.replaceChildren()
  const visible = state.operations.filter((entry) => {
    const text = `${entry.method} ${entry.path} ${entry.operation.operationId ?? ''} ${entry.operation.summary ?? ''}`.toLowerCase()
    return text.includes(query)
  })
  for (const entry of visible) {
    const button = element('button', { className: 'operation-button' })
    button.type = 'button'
    button.setAttribute('aria-current', String(state.selected === entry))
    button.append(element('span', { className: 'method', text: entry.method }))
    button.append(element('span', { className: 'path', text: entry.path }))
    button.addEventListener('click', () => selectOperation(entry))
    listNode.append(button)
  }
  statusNode.textContent = `${visible.length} of ${state.operations.length} operations`
}

function allParameters(entry) {
  const parameters = [...(entry.pathItem.parameters ?? []), ...(entry.operation.parameters ?? [])]
  return parameters.filter((parameter) => parameter && typeof parameter === 'object' && !parameter.$ref)
}

function selectOperation(entry) {
  state.selected = entry
  renderOperationList()
  renderOperationPanel(entry)
  panelNode.focus()
}

function renderOperationPanel(entry) {
  panelNode.replaceChildren()
  const title = element('h2', { text: entry.operation.summary ?? entry.operation.operationId ?? `${entry.method} ${entry.path}` })
  title.id = 'operation-title'
  panelNode.append(title)

  const metadata = element('div', { className: 'operation-meta' })
  metadata.append(element('span', { className: 'badge', text: entry.method }))
  metadata.append(element('code', { text: entry.path }))
  if (entry.protected) metadata.append(element('span', { className: 'badge protected', text: 'Protected mutation' }))
  panelNode.append(metadata)

  if (entry.operation.description) panelNode.append(element('p', { text: entry.operation.description }))
  if (entry.protected) {
    panelNode.append(element('p', {
      className: 'warning',
      text: 'Protected mutation: execution is disabled in the browser documentation. Use the generated TypeScript client from a trusted server-side environment and provide the internal API key there.',
    }))
  }

  const form = element('div', { className: 'parameter-grid' })
  for (const parameter of allParameters(entry)) {
    const wrapper = element('label', { className: 'parameter' })
    const name = `${parameter.name} (${parameter.in})${parameter.required ? ' — required' : ''}`
    wrapper.append(element('span', { text: name }))
    const input = element('input')
    input.dataset.parameterName = parameter.name
    input.dataset.parameterLocation = parameter.in
    input.required = parameter.required === true
    input.placeholder = parameter.schema?.default !== undefined ? String(parameter.schema.default) : ''
    wrapper.append(input)
    if (parameter.description) wrapper.append(element('small', { text: parameter.description }))
    form.append(wrapper)
  }
  panelNode.append(form)

  const actions = element('div', { className: 'actions' })
  const sendButton = element('button', { className: 'action', text: 'Send request' })
  sendButton.type = 'button'
  sendButton.disabled = entry.method !== 'GET' || entry.protected
  sendButton.addEventListener('click', () => executeOperation(entry, form))
  actions.append(sendButton)

  const curlButton = element('button', { className: 'action secondary', text: 'Copy curl' })
  curlButton.type = 'button'
  curlButton.addEventListener('click', () => copyCurl(entry, form))
  actions.append(curlButton)
  panelNode.append(actions)

  const requestPreview = element('pre', { text: 'Request preview will appear here.' })
  requestPreview.id = 'request-preview'
  panelNode.append(requestPreview)
  const responsePreview = element('pre', { text: 'Response will appear here.' })
  responsePreview.id = 'response-preview'
  responsePreview.setAttribute('aria-live', 'polite')
  panelNode.append(responsePreview)
  updateRequestPreview(entry, form)
  form.addEventListener('input', () => updateRequestPreview(entry, form))
}

function parameterValues(form) {
  const values = { path: {}, query: {}, header: {} }
  for (const input of form.querySelectorAll('input[data-parameter-name]')) {
    const value = input.value.trim()
    if (value !== '') values[input.dataset.parameterLocation][input.dataset.parameterName] = value
  }
  return values
}

function buildRequest(entry, form) {
  const values = parameterValues(form)
  let path = entry.path
  for (const [name, value] of Object.entries(values.path)) path = path.replace(`{${name}}`, encodeURIComponent(value))
  if (/\{[^}]+\}/.test(path)) throw new Error('Complete every required path parameter.')
  const url = new URL(path, baseURLInput.value)
  for (const [name, value] of Object.entries(values.query)) url.searchParams.set(name, value)
  return { url, values }
}

function curlCommand(entry, form) {
  const { url, values } = buildRequest(entry, form)
  const parts = [`curl --fail-with-body --silent --show-error -X ${entry.method}`, `'${url.toString()}'`, "-H 'Accept: application/json'"]
  for (const [name, value] of Object.entries(values.header)) parts.push(`-H '${name}: ${value.replaceAll("'", "'\\''")}'`)
  if (entry.protected) parts.push("-H 'X-Internal-API-Key: <trusted-server-key>'")
  return parts.join(' \\\n  ')
}

function updateRequestPreview(entry, form) {
  const preview = panelNode.querySelector('#request-preview')
  try {
    preview.textContent = curlCommand(entry, form)
  } catch (error) {
    preview.textContent = error instanceof Error ? error.message : String(error)
  }
}

async function copyCurl(entry, form) {
  const command = curlCommand(entry, form)
  await navigator.clipboard.writeText(command)
  statusNode.textContent = 'curl command copied'
}

async function executeOperation(entry, form) {
  const responsePreview = panelNode.querySelector('#response-preview')
  try {
    const { url, values } = buildRequest(entry, form)
    if (url.origin !== window.location.origin) {
      throw new Error('Browser execution is limited to the documentation origin. Use the generated curl command for another deployment.')
    }
    responsePreview.textContent = 'Loading…'
    const headers = new Headers({ Accept: 'application/json' })
    for (const [name, value] of Object.entries(values.header)) headers.set(name, value)
    const response = await fetch(url, { method: entry.method, headers })
    const text = await response.text()
    let body = text
    try { body = JSON.stringify(JSON.parse(text), null, 2) } catch { /* preserve non-JSON text */ }
    responsePreview.textContent = `HTTP ${response.status}\nX-Request-ID: ${response.headers.get('x-request-id') ?? 'not provided'}\n\n${body}`
  } catch (error) {
    responsePreview.textContent = error instanceof Error ? error.message : String(error)
  }
}

async function initialize() {
  try {
    const response = await fetch('/api/docs/openapi.json', { headers: { Accept: 'application/json' } })
    if (!response.ok) throw new Error(`OpenAPI request failed with status ${response.status}`)
    state.spec = await response.json()
    state.operations = operationRows(state.spec)
    renderOperationList()
    if (state.operations.length > 0) selectOperation(state.operations[0])
  } catch (error) {
    statusNode.textContent = error instanceof Error ? error.message : String(error)
  }
}

initialize()
