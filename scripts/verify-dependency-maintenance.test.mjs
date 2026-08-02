import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';

test('dependency targets stay coordinated', () => {
  const web = JSON.parse(fs.readFileSync('apps/web/package.json', 'utf8'));
  assert.equal(web.dependencies.react, web.dependencies['react-dom']);
  assert.equal(web.dependencies.react, '19.2.8');
  assert.equal(web.dependencies.three, '^0.185.1');
  assert.equal(web.devDependencies['@types/three'], '^0.185.1');
  assert.equal(web.devDependencies.typescript, '^5');
});

test('Dependabot groups related updates and blocks TypeScript major automation', () => {
  const config = fs.readFileSync('.github/dependabot.yml', 'utf8');
  assert.match(config, /react-runtime:/);
  assert.match(config, /tailwind-toolchain:/);
  assert.match(config, /three-runtime:/);
  assert.match(config, /fiber-runtime:/);
  assert.match(config, /setup-actions:/);
  assert.match(config, /package-ecosystem: "docker"/);
  assert.match(config, /dependency-name: "typescript"[\s\S]*version-update:semver-major/);
  assert.doesNotMatch(config, /labels:/);
});

test('stable release entry point executes dependency maintenance', () => {
  const root = JSON.parse(fs.readFileSync('package.json', 'utf8'));
  const release = fs.readFileSync('scripts/verify-release.sh', 'utf8');
  assert.equal(root.scripts['verify:release'], 'bash scripts/verify-release.sh');
  assert.match(release, /pnpm run test:dependency-maintenance/);
  assert.match(release, /pnpm run verify:dependency-maintenance/);
});

test('CI setup actions use version 7 only', () => {
  const backend = fs.readFileSync('.github/workflows/backend-ci.yml', 'utf8');
  const frontend = fs.readFileSync('.github/workflows/frontend-ci.yml', 'utf8');
  assert.match(backend, /actions\/setup-go@v7/);
  assert.doesNotMatch(backend, /actions\/setup-go@v6/);
  assert.match(frontend, /actions\/setup-node@v7/);
  assert.doesNotMatch(frontend, /actions\/setup-node@v6/);
});
