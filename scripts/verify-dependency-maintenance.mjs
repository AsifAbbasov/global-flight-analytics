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
const followUp = read('docs/171_DEPENDABOT_FOLLOW_UP_RECONCILIATION.md');

const expected = new Map([
  ['@tanstack/react-query', web.dependencies?.['@tanstack/react-query']],
  ['next', web.dependencies?.next],
  ['react', web.dependencies?.react],
  ['react-dom', web.dependencies?.['react-dom']],
  ['three', web.dependencies?.three],
  ['@types/react', web.devDependencies?.['@types/react']],
  ['@types/react-dom', web.devDependencies?.['@types/react-dom']],
  ['@types/three', web.devDependencies?.['@types/three']],
  ['eslint-config-next', web.devDependencies?.['eslint-config-next']],
  ['@tailwindcss/postcss', web.devDependencies?.['@tailwindcss/postcss']],
  ['tailwindcss', web.devDependencies?.tailwindcss],
  ['typescript', web.devDependencies?.typescript],
]);
for (const [name, actual] of expected) {
  const targets = {
    '@tanstack/react-query': '^5.101.4',
    next: '16.2.12',
    react: '19.2.8',
    'react-dom': '19.2.8',
    three: '^0.185.1',
    '@types/react': '^19.2.18',
    '@types/react-dom': '^19.2.4',
    '@types/three': '^0.185.1',
    'eslint-config-next': '16.2.12',
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
for (const literal of ['react-runtime:', 'next-toolchain:', 'tailwind-toolchain:', 'three-runtime:', 'fiber-runtime:', 'setup-actions:', 'package-ecosystem: "docker"', 'version-update:semver-major']) {
  if (!dependabot.includes(literal)) fail(`Dependabot contract is missing ${literal}`);
}
for (const dependency of ['typescript', 'eslint', '@types/node', 'maplibre-gl']) {
  const pattern = new RegExp(`dependency-name: \"${dependency.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\"[\\s\\S]*?version-update:semver-major`);
  if (!pattern.test(dependabot)) fail(`Dependabot major deferral is missing for ${dependency}`);
}
if (dependabot.includes('labels:')) fail('Dependabot still references repository labels');
if (!root.scripts?.['test:dependency-maintenance'] || !root.scripts?.['verify:dependency-maintenance']) fail('root dependency scripts are missing');
if (root.scripts?.['verify:release'] !== 'bash scripts/verify-release.sh') fail('stable release entry point changed');
if (!releaseScript.includes('pnpm run test:dependency-maintenance') || !releaseScript.includes('pnpm run verify:dependency-maintenance')) fail('release gate does not include dependency maintenance');
if (!index.includes('## Document 170 — Dependency Maintenance Closure')) fail('Document 170 index entry is missing');
if (!index.includes('## Document 171 — Dependabot Follow-up Reconciliation')) fail('Document 171 index entry is missing');
if (!closure.includes('DEPENDABOT_FOLLOW_UP_RECONCILIATION=PASS')) fail('dependency closure follow-up marker is missing');
if (!followUp.includes('DEPENDENCY_MAINTENANCE_DEBT=CLOSED')) fail('follow-up closure marker is missing');
if (!frontendCI.includes('docs/171_DEPENDABOT_FOLLOW_UP_RECONCILIATION.md')) fail('frontend CI does not watch Document 171');
if (!frontendCI.includes('scripts/verify-frontend-dependency-security.mjs') || !frontendCI.includes('scripts/verify-frontend-dependency-security.test.mjs')) fail('frontend CI does not watch dependency security contracts');
console.log('DEPENDENCY_MAINTENANCE_CONTRACT=PASS');
