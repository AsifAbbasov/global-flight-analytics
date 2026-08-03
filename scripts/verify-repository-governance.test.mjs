import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';

const read = relative => fs.readFileSync(relative, 'utf8');

test('all external actions use immutable full SHAs with version comments', () => {
  const files = fs.readdirSync('.github/workflows').filter(name => /\.ya?ml$/.test(name));
  let count = 0;
  for (const file of files) {
    const content = read(`.github/workflows/${file}`);
    for (const match of content.matchAll(/^\s*uses:\s*([^\s#]+)(?:\s+#\s*(\S+))?\s*$/gm)) {
      if (match[1].startsWith('./')) continue;
      count += 1;
      assert.match(match[1], /^[^/@]+\/[^/@]+(?:\/[^@]+)?@[0-9a-f]{40}$/);
      assert.match(match[2] ?? '', /^v\d+(?:\.\d+\.\d+)?$/);
    }
  }
  assert.ok(count > 0);
});

test('required workflows expose stable gates and run for every pull request', () => {
  const backend = read('.github/workflows/backend-ci.yml');
  const frontend = read('.github/workflows/frontend-ci.yml');
  for (const workflow of [backend, frontend]) {
    const trigger = workflow.slice(0, workflow.indexOf('\npermissions:\n'));
    assert.match(trigger, /on:\n  pull_request:\n\n  push:/);
  }
  assert.match(backend, /name: Backend CI Gate/);
  assert.match(frontend, /name: Frontend CI Gate/);
});

test('CodeQL covers Go and JavaScript TypeScript with a stable gate', () => {
  const codeql = read('.github/workflows/codeql.yml');
  assert.match(codeql, /- language: go\n            build-mode: autobuild/);
  assert.match(codeql, /- language: javascript-typescript\n            build-mode: none/);
  assert.match(codeql, /build-mode:/);
  assert.match(codeql, /queries: security-extended/);
  assert.match(codeql, /name: CodeQL Security Gate/);
});

test('legacy workflow visibility contracts understand unconditional pull request CI', () => {
  const recruiter = read('scripts/verify-recruiter-quickstart.sh');
  const releasePortfolio = read('scripts/verify-release-portfolio.sh');
  const releasePortfolioTests = read('scripts/verify-release-portfolio.test.mjs');
  const backendOperationsTests = read('scripts/verify-backend-operations.test.mjs');

  assert.match(recruiter, /Backend CI must run for every pull request/);
  assert.match(recruiter, /push path filters must include README\.md exactly once/);
  assert.doesNotMatch(recruiter, /for both push and pull_request/);

  assert.match(releasePortfolio, /Backend CI must run for every pull request/);
  assert.match(releasePortfolio, /push path filters must track frontend workflow changes exactly once/);
  assert.doesNotMatch(releasePortfolio, /for push and pull_request/);

  for (const migratedTestSource of [releasePortfolioTests, backendOperationsTests]) {
    assert.ok(migratedTestSource.includes('const trigger ='));
    assert.ok(migratedTestSource.includes('assert.match(trigger, /on:'));
    assert.ok(migratedTestSource.includes('pull_request:'));
    assert.ok(migratedTestSource.includes('push:/)'));
  }
  assert.ok(releasePortfolioTests.includes('backend.split(trackedFrontendWorkflowPath).length - 1, 1'));
  assert.ok(!releasePortfolioTests.includes(']).length, 2'));
  assert.ok(backendOperationsTests.includes('source.split(trackedRenderBlueprintPath).length - 1, 1'));
  assert.ok(!backendOperationsTests.includes(']).length, 2'));
});

test('ownership security and settings automation are durable repository files', () => {
  assert.match(read('.github/CODEOWNERS'), /\* @AsifAbbasov/);
  assert.match(read('SECURITY.md'), /Reporting a vulnerability/);
  assert.match(read('scripts/configure-repository-governance.sh'), /MAIN_PROTECTION_RULESET_NAME/);
  assert.match(read('scripts/verify-repository-governance-settings.sh'), /REPOSITORY_GOVERNANCE_SETTINGS=PASS/);
});
