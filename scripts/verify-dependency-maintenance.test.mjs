import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';

test('dependency targets stay coordinated', () => {
  const web = JSON.parse(fs.readFileSync('apps/web/package.json', 'utf8'));
  assert.equal(web.dependencies['@tanstack/react-query'], '^5.101.4');
  assert.equal(web.dependencies.next, '16.2.12');
  assert.equal(web.dependencies.react, web.dependencies['react-dom']);
  assert.equal(web.dependencies.react, '19.2.8');
  assert.equal(web.dependencies.three, '^0.185.1');
  assert.equal(web.devDependencies['@types/react'], '^19.2.18');
  assert.equal(web.devDependencies['@types/react-dom'], '^19.2.4');
  assert.equal(web.devDependencies['@types/three'], '^0.185.1');
  assert.equal(web.devDependencies['eslint-config-next'], '16.2.12');
  assert.equal(web.devDependencies.typescript, '^5');
});

test('Dependabot groups related updates and blocks TypeScript major automation', () => {
  const config = fs.readFileSync('.github/dependabot.yml', 'utf8');
  assert.match(config, /react-runtime:/);
  assert.match(config, /next-toolchain:/);
  assert.match(config, /tailwind-toolchain:/);
  assert.match(config, /three-runtime:/);
  assert.match(config, /fiber-runtime:/);
  assert.match(config, /setup-actions:/);
  assert.match(config, /package-ecosystem: "docker"/);
  for (const dependency of ['typescript', 'eslint', '@types/node', 'maplibre-gl']) {
    const escaped = dependency.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    assert.match(config, new RegExp(`dependency-name: \"${escaped}\"[\\s\\S]*?version-update:semver-major`));
  }
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


test('follow-up reconciliation is documented and CI-visible', () => {
  const index = fs.readFileSync('docs/DOCUMENT_INDEX.md', 'utf8');
  const closure = fs.readFileSync('docs/170_DEPENDENCY_MAINTENANCE_CLOSURE.md', 'utf8');
  const followUp = fs.readFileSync('docs/171_DEPENDABOT_FOLLOW_UP_RECONCILIATION.md', 'utf8');
  const frontend = fs.readFileSync('.github/workflows/frontend-ci.yml', 'utf8');
  assert.match(index, /Document 171 — Dependabot Follow-up Reconciliation/);
  assert.match(closure, /DEPENDABOT_FOLLOW_UP_RECONCILIATION=PASS/);
  assert.match(followUp, /DEPENDENCY_MAINTENANCE_DEBT=CLOSED/);
  assert.match(frontend, /docs\/171_DEPENDABOT_FOLLOW_UP_RECONCILIATION\.md/);
  assert.match(frontend, /scripts\/verify-frontend-dependency-security\.mjs/);
  assert.match(frontend, /scripts\/verify-frontend-dependency-security\.test\.mjs/);
});
