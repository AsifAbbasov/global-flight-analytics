#!/usr/bin/env node

import { readdir, readFile } from 'node:fs/promises'
import path from 'node:path'
import process from 'node:process'

const repositoryRoot = path.resolve(
  path.dirname(new URL(import.meta.url).pathname),
  '..'
)

const excludedDirectories = new Set([
  '.git',
  '.next',
  'build',
  'coverage',
  'dist',
  'node_modules',
  'vendor',
])

const cyrillicPattern = /[\u0400-\u04FF]/
const findings = []

async function visit(directory) {
  const entries = await readdir(directory, { withFileTypes: true })

  for (const entry of entries) {
    if (entry.name.startsWith('.') && entry.name !== '.github') {
      if (entry.isDirectory()) {
        continue
      }
    }

    const absolutePath = path.join(directory, entry.name)
    const relativePath = path.relative(repositoryRoot, absolutePath)

    if (entry.isDirectory()) {
      if (!excludedDirectories.has(entry.name)) {
        await visit(absolutePath)
      }
      continue
    }

    if (!entry.isFile() || !entry.name.endsWith('.md')) {
      continue
    }

    const content = await readFile(absolutePath, 'utf8')
    const lines = content.split(/\r?\n/)

    lines.forEach((line, index) => {
      if (cyrillicPattern.test(line)) {
        findings.push(`${relativePath}:${index + 1}:${line}`)
      }
    })
  }
}

await visit(repositoryRoot)

if (findings.length > 0) {
  console.error('DOCUMENTATION_ENGLISH_AUDIT=FAIL')
  for (const finding of findings) {
    console.error(finding)
  }
  process.exit(1)
}

console.log('DOCUMENTATION_ENGLISH_AUDIT=PASS')
