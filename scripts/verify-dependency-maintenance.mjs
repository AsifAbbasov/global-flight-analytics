#!/usr/bin/env node
import fs from 'node:fs';

function fail(message) {
  console.error(`DEPENDENCY_MAINTENANCE_CONTRACT=FAIL reason=${message}`);
  process.exit(1);
}

function read(relative) {
  return fs.readFileSync(relative, 'utf8');
}

const web = JSON.parse(read('apps/web/package.json'));
const root = JSON.parse(read('package.json'));
const goMod = read('apps/api/go.mod');
const backendCI = read('.github/workflows/backend-ci.yml');
const frontendCI = read('.github/workflows/frontend-ci.yml');
const releaseScript = read('scripts/verify-release.sh');
const dependabot = read('.github/dependabot.yml');
const index = read('docs/DOCUMENT_INDEX.md');
const closure = read('docs/170_DEPENDENCY_MAINTENANCE_CLOSURE.md');

const expected = new Map([
  ['react', web.dependencies?.react],
  ['react-dom', web.dependencies?.['react-dom']],
  ['three', web.dependencies?.three],
  ['@types/three', web.devDependencies?.['@types/three']],
  ['@tailwindcss/postcss', web.devDependencies?.['@tailwindcss/postcss']],
  ['tailwindcss', web.devDependencies?.tailwindcss],
  ['typescript', web.devDependencies?.typescript],
]);
for (const [name, actual] of expected) {
  const targets = {
    react: '19.2.8',
    'react-dom': '19.2.8',
    three: '^0.185.1',
    '@types/three': '^0.185.1',
    '@tailwindcss/postcss': '^4.3.3',
    tailwindcss: '^4.3.3',
    typescript: '^5',
  };
  if (actual !== targets[name]) fail(`unexpected ${name} version: ${actual}`);
}
if (!goMod.includes('github.com/gofiber/fiber/v2 v2.52.14')) fail('Fiber v2.52.14 is missing');
if (backendCI.includes('actions/setup-go@v6') || !backendCI.includes('actions/setup-go@v7')) fail('setup-go v7 contract failed');
if (frontendCI.includes('actions/setup-node@v6') || !frontendCI.includes('actions/setup-node@v7')) fail('setup-node v7 contract failed');
if (!backendCI.includes('Verify dependency maintenance contract') || !frontendCI.includes('Verify dependency maintenance contract')) fail('dependency maintenance is not enforced by both CI workflows');
for (const literal of ['react-runtime:', 'tailwind-toolchain:', 'three-runtime:', 'fiber-runtime:', 'setup-actions:', 'package-ecosystem: "docker"', 'version-update:semver-major']) {
  if (!dependabot.includes(literal)) fail(`Dependabot contract is missing ${literal}`);
}
if (dependabot.includes('labels:')) fail('Dependabot still references repository labels');
if (!root.scripts?.['test:dependency-maintenance'] || !root.scripts?.['verify:dependency-maintenance']) fail('root dependency scripts are missing');
if (root.scripts?.['verify:release'] !== 'bash scripts/verify-release.sh') fail('stable release entry point changed');
if (!releaseScript.includes('pnpm run test:dependency-maintenance') || !releaseScript.includes('pnpm run verify:dependency-maintenance')) fail('release gate does not include dependency maintenance');
if (!index.includes('## Document 170 — Dependency Maintenance Closure')) fail('Document 170 index entry is missing');
if (!closure.includes('DEPENDENCY_MAINTENANCE_DEBT=CLOSED')) fail('dependency closure marker is missing');
console.log('DEPENDENCY_MAINTENANCE_CONTRACT=PASS');
