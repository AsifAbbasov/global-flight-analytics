#!/usr/bin/env node
import fs from 'node:fs';

function fail(message) {
  console.error(`REPOSITORY_GOVERNANCE_CONTRACT=FAIL reason=${message}`);
  process.exit(1);
}
function read(relative) {
  return fs.readFileSync(relative, 'utf8');
}

const workflowDirectory = '.github/workflows';
const workflows = fs.readdirSync(workflowDirectory)
  .filter(name => /\.ya?ml$/.test(name))
  .sort();

const externalUses = [];
for (const name of workflows) {
  const content = read(`${workflowDirectory}/${name}`);
  for (const match of content.matchAll(/^\s*uses:\s*([^\s#]+)(?:\s+#\s*(\S+))?\s*$/gm)) {
    const reference = match[1];
    const versionComment = match[2] ?? '';
    if (reference.startsWith('./')) continue;
    externalUses.push({ name, reference, versionComment });
    if (!/^[^/@]+\/[^/@]+(?:\/[^@]+)?@[0-9a-f]{40}$/.test(reference)) {
      fail(`external action is not pinned to a full SHA: ${name}: ${reference}`);
    }
    if (!/^v\d+(?:\.\d+\.\d+)?$/.test(versionComment)) {
      fail(`pinned action lacks a readable version comment: ${name}: ${reference}`);
    }
  }
}
if (externalUses.length === 0) fail('no external action references were found');

const backend = read('.github/workflows/backend-ci.yml');
const frontend = read('.github/workflows/frontend-ci.yml');
const codeql = read('.github/workflows/codeql.yml');

for (const [name, workflow] of [['Backend CI', backend], ['Frontend CI', frontend]]) {
  const trigger = workflow.slice(0, workflow.indexOf('\npermissions:\n'));
  if (!/on:\n  pull_request:\n\n  push:/.test(trigger)) {
    fail(`${name} does not run for every pull request`);
  }
}
if (!backend.includes('name: Backend CI Gate')) fail('Backend CI Gate is missing');
if (!frontend.includes('name: Frontend CI Gate')) fail('Frontend CI Gate is missing');
if (!codeql.includes('name: CodeQL Security Gate')) fail('CodeQL Security Gate is missing');
if (!codeql.includes('- language: go\n            build-mode: autobuild')) {
  fail('CodeQL Go autobuild mode is missing');
}
if (!codeql.includes('- language: javascript-typescript\n            build-mode: none')) {
  fail('CodeQL JavaScript TypeScript no-build mode is missing');
}
if (!codeql.includes('build-mode: ${{ matrix.build-mode }}')) {
  fail('CodeQL matrix build mode is not wired into initialization');
}
if (!codeql.includes('queries: security-extended')) fail('CodeQL security-extended suite is missing');

const codeowners = read('.github/CODEOWNERS');
if (!codeowners.includes('* @AsifAbbasov')) fail('CODEOWNERS baseline is missing');
const security = read('SECURITY.md');
if (!security.includes('Reporting a vulnerability')) fail('SECURITY policy is missing');

const root = JSON.parse(read('package.json'));
if (!root.scripts?.['test:repository-governance']) fail('governance test script is missing');
if (!root.scripts?.['verify:repository-governance']) fail('governance verifier script is missing');
const release = read('scripts/verify-release.sh');
if (!release.includes('pnpm run test:repository-governance')) fail('release gate misses governance tests');
if (!release.includes('pnpm run verify:repository-governance')) fail('release gate misses governance verifier');

const recruiterQuickstart = read('scripts/verify-recruiter-quickstart.sh');
if (!recruiterQuickstart.includes('Backend CI must run for every pull request')) {
  fail('recruiter quickstart verifier still assumes duplicated pull request path filters');
}
if (!recruiterQuickstart.includes('push path filters must include README.md exactly once')) {
  fail('recruiter quickstart verifier does not preserve the push README contract');
}

const releasePortfolio = read('scripts/verify-release-portfolio.sh');
if (!releasePortfolio.includes('Backend CI must run for every pull request')) {
  fail('release portfolio verifier still assumes duplicated pull request path filters');
}
if (!releasePortfolio.includes('push path filters must track frontend workflow changes exactly once')) {
  fail('release portfolio verifier does not preserve the push workflow visibility contract');
}
const releasePortfolioTests = read('scripts/verify-release-portfolio.test.mjs');
if (!releasePortfolioTests.includes('trackedFrontendWorkflowPath')) fail('release portfolio Node single-path contract missing');
if (releasePortfolioTests.split('\n').some(line => line.includes('frontend-ci') && /length,\s*2/.test(line))) fail('release portfolio duplicate path contract remains');
const backendOperationsTests = read('scripts/verify-backend-operations.test.mjs');
if (!backendOperationsTests.includes('trackedRenderBlueprintPath')) fail('backend operations Node single-path contract missing');
if (backendOperationsTests.split('\n').some(line => line.includes('render') && /length,\s*2/.test(line))) fail('backend operations duplicate path contract remains');

const document = read('docs/172_REPOSITORY_GOVERNANCE_AND_SECURITY_AUTOMATION.md');
if (!document.includes('REPOSITORY_GOVERNANCE_STAGE=READY_FOR_SETTINGS')) {
  fail('Document 172 readiness marker is missing');
}
for (const marker of [
  'INSTALLER_POSITIVE_FIXTURES=PASS',
  'INSTALLER_NEGATIVE_FIXTURES=PASS',
  'INSTALLER_ROLLBACK_SELF_TEST=PASS',
  'GOFMT_BEFORE_AND_AFTER=PASS',
  'FULL_REPOSITORY_VALIDATION=PASS',
]) {
  if (!document.includes(marker)) {
    fail('Document 172 installer quality marker is missing: ' + marker);
  }
}
console.log(`PINNED_EXTERNAL_ACTION_REFERENCES=${externalUses.length}`);
console.log('REPOSITORY_GOVERNANCE_CONTRACT=PASS');
