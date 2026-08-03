#!/usr/bin/env node
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { createHash } from 'node:crypto';
import { execFileSync, spawnSync } from 'node:child_process';

const BASELINE = '72fc2be736ec4de386a055d826592fa832dcc0fc';
const PROJECT_ROOT = process.env.PROJECT_ROOT
  ?? path.join(os.homedir(), 'Documents', 'global-flight-analytics');
const PACKAGE_ROOT = path.dirname(new URL(import.meta.url).pathname);
const VALIDATED_ACTION_REFS = Object.freeze({
  checkout: '3d3c42e5aac5ba805825da76410c181273ba90b1',
  setupGo: 'b7ad1dad31e06c5925ef5d2fc7ad053ef454303e',
  setupNode: '820762786026740c76f36085b0efc47a31fe5020',
  pnpmSetup: '0ebf47130e4866e96fce0953f49152a61190b271',
  codeql: 'f205ea1c3313d32999d8d6a48b4f6530d4437b38',
});

function fail(message) {
  console.error(`REPOSITORY_GOVERNANCE_INSTALL=FAIL reason=${message}`);
  process.exit(1);
}

function run(command, args, options = {}) {
  const result = spawnSync(command, args, {
    cwd: options.cwd ?? PROJECT_ROOT,
    encoding: 'utf8',
    stdio: options.capture ? 'pipe' : 'inherit',
    env: { ...process.env, ...(options.env ?? {}) },
  });
  if (result.status !== 0) {
    if (options.capture) {
      process.stderr.write(result.stdout ?? '');
      process.stderr.write(result.stderr ?? '');
    }
    throw new Error(`${command} ${args.join(' ')} failed`);
  }
  return options.capture ? (result.stdout ?? '').trim() : '';
}

function git(args, cwd = PROJECT_ROOT) {
  return run('git', args, { cwd, capture: true });
}

function read(relative, cwd = PROJECT_ROOT) {
  return fs.readFileSync(path.join(cwd, relative), 'utf8');
}

function write(relative, content, cwd = PROJECT_ROOT) {
  const destination = path.join(cwd, relative);
  fs.mkdirSync(path.dirname(destination), { recursive: true });
  fs.writeFileSync(destination, content);
  if (relative.endsWith('.sh')) fs.chmodSync(destination, 0o755);
}

const EXPECTED_CHANGED_FILES = [
  '.github/CODEOWNERS',
  '.github/workflows/backend-ci.yml',
  '.github/workflows/codeql.yml',
  '.github/workflows/frontend-ci.yml',
  'SECURITY.md',
  'docs/172_REPOSITORY_GOVERNANCE_AND_SECURITY_AUTOMATION.md',
  'docs/DOCUMENT_INDEX.md',
  'package.json',
  'scripts/configure-repository-governance.sh',
  'scripts/verify-backend-operations.test.mjs',
  'scripts/verify-dependency-maintenance.mjs',
  'scripts/verify-dependency-maintenance.test.mjs',
  'scripts/verify-recruiter-quickstart.sh',
  'scripts/verify-release-portfolio.sh',
  'scripts/verify-release-portfolio.test.mjs',
  'scripts/verify-release.sh',
  'scripts/verify-repository-governance-settings.sh',
  'scripts/verify-repository-governance.mjs',
  'scripts/verify-repository-governance.test.mjs',
].sort();

function expectFailure(label, callback, expectedPattern = null) {
  let failure = null;
  try {
    callback();
  } catch (error) {
    failure = error;
  }

  if (failure === null) {
    throw new Error(`${label}: expected failure but operation succeeded`);
  }

  const message = failure instanceof Error ? failure.message : String(failure);
  if (expectedPattern !== null && !expectedPattern.test(message)) {
    throw new Error(
      `${label}: unexpected failure message: ${message}`
    );
  }
}

function listFilesRecursively(directory, predicate) {
  const files = [];

  function visit(current) {
    const entries = fs.readdirSync(current, { withFileTypes: true })
      .sort((left, right) => left.name.localeCompare(right.name));

    for (const entry of entries) {
      const absolute = path.join(current, entry.name);
      if (entry.isDirectory()) {
        visit(absolute);
      } else if (entry.isFile() && predicate(absolute)) {
        files.push(absolute);
      }
    }
  }

  visit(directory);
  return files;
}

function hashGoTree(cwd) {
  const goRoot = path.join(cwd, 'apps', 'api');
  const files = listFilesRecursively(
    goRoot,
    absolute => absolute.endsWith('.go')
  );
  if (files.length === 0) {
    throw new Error('Go source tree is empty');
  }

  const digest = createHash('sha256');
  for (const absolute of files) {
    const relative = path.relative(cwd, absolute).split(path.sep).join('/');
    digest.update(relative);
    digest.update('\0');
    digest.update(fs.readFileSync(absolute));
    digest.update('\0');
  }
  return digest.digest('hex');
}

function gofmtFiles(cwd) {
  const goRoot = path.join(cwd, 'apps', 'api');
  return listFilesRecursively(
    goRoot,
    absolute => absolute.endsWith('.go')
  );
}

function verifyGoFormattingBaseline(cwd, marker) {
  const files = gofmtFiles(cwd);
  if (files.length === 0) throw new Error('Go source tree is empty');

  const result = spawnSync(
    'gofmt',
    ['-l', ...files],
    {
      cwd,
      encoding: 'utf8',
      stdio: 'pipe',
    }
  );
  if (result.status !== 0) {
    process.stderr.write(result.stdout ?? '');
    process.stderr.write(result.stderr ?? '');
    throw new Error('gofmt baseline inspection failed');
  }

  const unformatted = (result.stdout ?? '').trim();
  if (unformatted !== '') {
    throw new Error(
      `Go source is not formatted before patch: ${unformatted.replace(/\s+/g, ', ')}`
    );
  }

  const digest = hashGoTree(cwd);
  console.log(`${marker}=PASS hash=${digest}`);
  return digest;
}

function formatAndVerifyGoTree(cwd, expectedDigest, marker) {
  const files = gofmtFiles(cwd);
  const result = spawnSync(
    'gofmt',
    ['-w', ...files],
    {
      cwd,
      encoding: 'utf8',
      stdio: 'pipe',
    }
  );
  if (result.status !== 0) {
    process.stderr.write(result.stdout ?? '');
    process.stderr.write(result.stderr ?? '');
    throw new Error('gofmt write pass failed');
  }

  const actualDigest = hashGoTree(cwd);
  if (actualDigest !== expectedDigest) {
    throw new Error(
      `Go source changed during patch formatting: expected ${expectedDigest}, got ${actualDigest}`
    );
  }
  console.log(`${marker}=PASS hash=${actualDigest}`);
}

function changedFiles(cwd) {
  const tracked = git(['diff', '--name-only'], cwd)
    .split('\n')
    .filter(Boolean);
  const untracked = git(['ls-files', '--others', '--exclude-standard'], cwd)
    .split('\n')
    .filter(Boolean);
  return [...new Set([...tracked, ...untracked])].sort();
}

function verifyChangedManifest(cwd, expected = EXPECTED_CHANGED_FILES) {
  const actual = changedFiles(cwd);
  const normalizedExpected = [...expected].sort();
  if (JSON.stringify(actual) !== JSON.stringify(normalizedExpected)) {
    throw new Error(
      `changed file manifest mismatch expected=${normalizedExpected.join(',')} actual=${actual.join(',')}`
    );
  }
}

function rollbackRepository(cwd, baseline) {
  run('git', ['reset', '--hard', baseline], { cwd });
  for (const relative of EXPECTED_CHANGED_FILES) {
    const absolute = path.join(cwd, relative);
    const tracked = spawnSync(
      'git',
      ['ls-files', '--error-unmatch', '--', relative],
      {
        cwd,
        encoding: 'utf8',
        stdio: 'pipe',
      }
    );
    if (tracked.status !== 0) {
      fs.rmSync(absolute, { recursive: true, force: true });
    }
  }
  if (git(['status', '--porcelain'], cwd) !== '') {
    throw new Error('rollback did not restore a clean repository');
  }
  console.log('REAL_REPOSITORY_ROLLBACK=PASS');
}

function applyWithRollback(cwd, baseline, patch, validation) {
  try {
    patch();
    validation();
  } catch (error) {
    rollbackRepository(cwd, baseline);
    throw error;
  }
}

function replaceExactly(source, needle, replacement, label) {
  const count = source.split(needle).length - 1;
  if (count !== 1) throw new Error(`${label}: expected one anchor, found ${count}`);
  return source.replace(needle, replacement);
}

function replacePatternExactly(source, pattern, replacement, label) {
  const match = source.match(pattern);
  if (match === null || match.index === undefined) {
    throw new Error(`${label}: expected one semantic match, found 0`);
  }

  const remainder = source.slice(match.index + match[0].length);
  if (pattern.test(remainder)) {
    pattern.lastIndex = 0;
    throw new Error(`${label}: expected one semantic match, found more than 1`);
  }
  pattern.lastIndex = 0;

  return source.replace(pattern, replacement);
}

function migrateRecruiterQuickstartContract(source) {
  const replacement = `trigger_header="$(
  awk '
    /^on:/ { capture = 1 }
    /^permissions:/ { exit }
    capture { print }
  ' "$WORKFLOW_FILE"
)"

pull_request_count="$(
  printf '%s\\n' "$trigger_header" |
    awk '
      $0 == "  pull_request:" { count += 1 }
      END { print count + 0 }
    '
)"

if [ "$pull_request_count" -ne 1 ]; then
  fail 'Backend CI must run for every pull request'
fi

pull_request_block="$(
  printf '%s\\n' "$trigger_header" |
    awk '
      $0 == "  pull_request:" { capture = 1; next }
      $0 == "  push:" { exit }
      capture { print }
    '
)"

pull_request_paths_count="$(
  printf '%s\\n' "$pull_request_block" |
    awk '
      {
        line = $0
        sub(/^[[:space:]]*/, "", line)
        if (line == "paths:") {
          count += 1
        }
      }
      END { print count + 0 }
    '
)"

if [ "$pull_request_paths_count" -ne 0 ]; then
  fail 'Backend CI pull_request trigger must not use path filters'
fi

readme_path_count="$(
  grep -F -c -- "- 'README.md'" "$WORKFLOW_FILE"
)"
if [ "$readme_path_count" -ne 1 ]; then
  fail 'Backend CI push path filters must include README.md exactly once'
fi
`;

  return replacePatternExactly(
    source,
    /readme_path_count="\$\([\s\S]*?Backend CI must trigger on README\.md for both push and pull_request'\nfi\n/,
    replacement,
    'recruiter quickstart governance trigger migration'
  );
}

function migrateReleasePortfolioContract(source) {
  const replacement = `backend_trigger_header="$(
  awk '
    /^on:/ { capture = 1 }
    /^permissions:/ { exit }
    capture { print }
  ' "$REPOSITORY_ROOT/.github/workflows/backend-ci.yml"
)"

backend_pull_request_count="$(
  printf '%s\\n' "$backend_trigger_header" |
    awk '
      $0 == "  pull_request:" { count += 1 }
      END { print count + 0 }
    '
)"

if [ "$backend_pull_request_count" -ne 1 ]; then
  fail 'Backend CI must run for every pull request'
fi

backend_pull_request_block="$(
  printf '%s\\n' "$backend_trigger_header" |
    awk '
      $0 == "  pull_request:" { capture = 1; next }
      $0 == "  push:" { exit }
      capture { print }
    '
)"

backend_pull_request_paths_count="$(
  printf '%s\\n' "$backend_pull_request_block" |
    awk '
      {
        line = $0
        sub(/^[[:space:]]*/, "", line)
        if (line == "paths:") {
          count += 1
        }
      }
      END { print count + 0 }
    '
)"

if [ "$backend_pull_request_paths_count" -ne 0 ]; then
  fail 'Backend CI pull_request trigger must not use path filters'
fi

backend_frontend_workflow_count="$(
  grep -F -c -- "- '.github/workflows/frontend-ci.yml'" \\
    "$REPOSITORY_ROOT/.github/workflows/backend-ci.yml"
)"
[ "$backend_frontend_workflow_count" -eq 1 ] || \\
  fail 'Backend CI push path filters must track frontend workflow changes exactly once'
`;

  return replacePatternExactly(
    source,
    /backend_frontend_workflow_count="\$\([\s\S]*?Backend CI must track frontend workflow changes for push and pull_request'\n/,
    replacement,
    'release portfolio governance trigger migration'
  );
}

function migrateReleasePortfolioTests(source) {
  return replacePatternExactly(
    source,
    /^\s*assert\.equal\(\(backend\.match\([^\n]*frontend-ci[^\n]*\)\s*\?\?\s*\[\]\)\.length,\s*2\)\s*$/m,
    `  const trigger = backend.slice(0, backend.indexOf('\\npermissions:\\n'))
  assert.match(trigger, /on:\\n  pull_request:\\n\\n  push:/)
  const trackedFrontendWorkflowPath = "- '.github/workflows/frontend-ci.yml'"
  assert.equal(backend.split(trackedFrontendWorkflowPath).length - 1, 1)`,
    'release portfolio test governance trigger migration'
  );
}

function migrateBackendOperationsTests(source) {
  return replacePatternExactly(
    source,
    /^\s*assert\.equal\(\(source\.match\([^\n]*render[^\n]*yaml[^\n]*\)\s*\?\?\s*\[\]\)\.length,\s*2\)\s*$/m,
    `  const trigger = source.slice(0, source.indexOf('\\npermissions:\\n'))
  assert.match(trigger, /on:\\n  pull_request:\\n\\n  push:/)
  const trackedRenderBlueprintPath = "- 'render.yaml'"
  assert.equal(source.split(trackedRenderBlueprintPath).length - 1, 1)`,
    'backend operations test governance trigger migration'
  );
}

function assertLegacyWorkflowContractsRemoved(cwd) {
  const recruiter = read('scripts/verify-recruiter-quickstart.sh', cwd);
  const releaseShell = read('scripts/verify-release-portfolio.sh', cwd);
  const releaseTests = read('scripts/verify-release-portfolio.test.mjs', cwd);
  const backendOperationsTests = read('scripts/verify-backend-operations.test.mjs', cwd);
  const stale = [
    recruiter.includes('Backend CI must trigger on README.md for both push and pull_request'),
    releaseShell.includes('Backend CI must track frontend workflow changes for push and pull_request'),
    releaseTests.split('\\n').some(line => line.includes('frontend-ci') && /length,\s*2/.test(line)),
    backendOperationsTests.split('\\n').some(line => line.includes('render') && /length,\s*2/.test(line)),
  ];
  if (stale.some(Boolean)) throw new Error('stale workflow visibility contract remains');
}

function resolveActionRef(repository, ref) {
  let object = JSON.parse(run('gh', [
    'api',
    `repos/${repository}/git/ref/tags/${ref}`,
  ], { capture: true, cwd: PROJECT_ROOT })).object;

  for (let depth = 0; depth < 4; depth += 1) {
    if (object.type === 'commit') {
      if (!/^[0-9a-f]{40}$/.test(object.sha)) {
        throw new Error(`invalid commit SHA for ${repository}@${ref}`);
      }
      return object.sha;
    }
    if (object.type !== 'tag') {
      throw new Error(`unsupported Git object type ${object.type} for ${repository}@${ref}`);
    }
    object = JSON.parse(run('gh', [
      'api',
      `repos/${repository}/git/tags/${object.sha}`,
    ], { capture: true, cwd: PROJECT_ROOT })).object;
  }
  throw new Error(`tag dereference depth exceeded for ${repository}@${ref}`);
}

function patchTrigger(workflow, name) {
  const onMarker = 'on:\n';
  const permissionsMarker = '\npermissions:\n';
  const onIndex = workflow.indexOf(onMarker);
  const permissionsIndex = workflow.indexOf(permissionsMarker);
  if (onIndex < 0 || permissionsIndex < 0 || permissionsIndex <= onIndex) {
    throw new Error(`${name}: workflow trigger boundaries are missing`);
  }
  if (!workflow.startsWith(`name: ${name}\n\n`)) {
    throw new Error(`${name}: unexpected workflow header`);
  }

  const existingTrigger = workflow.slice(
    onIndex + onMarker.length,
    permissionsIndex
  );
  const pushIndex = existingTrigger.indexOf('  push:\n');
  if (pushIndex < 0) throw new Error(`${name}: push trigger is missing`);

  const pushBlock = existingTrigger.slice(pushIndex).trimEnd();
  return `${workflow.slice(0, onIndex)}on:\n  pull_request:\n\n${pushBlock}\n  workflow_dispatch:\n${workflow.slice(permissionsIndex)}`;
}

function pinUses(workflow, refs) {
  const replacements = [
    ['actions/checkout@v7', `actions/checkout@${refs.checkout} # v7`],
    ['actions/setup-go@v7', `actions/setup-go@${refs.setupGo} # v7`],
    ['actions/setup-node@v7', `actions/setup-node@${refs.setupNode} # v7`],
    ['pnpm/action-setup@v6', `pnpm/action-setup@${refs.pnpmSetup} # v6`],
  ];
  for (const [before, after] of replacements) {
    workflow = workflow.split(before).join(after);
  }
  return workflow;
}

function patchRepository(cwd, refs) {
  let backend = read('.github/workflows/backend-ci.yml', cwd);
  backend = patchTrigger(backend, 'Backend CI');
  backend = pinUses(backend, refs);
  const backendGate = `

  backend-ci-gate:
    name: Backend CI Gate
    if: \${{ always() }}
    needs:
      - backend-quality
      - backend-container
      - backend-race
      - postgres-integration
    runs-on: ubuntu-latest
    timeout-minutes: 5
    steps:
      - name: Verify backend workflow result
        shell: bash
        run: |
          test "\${{ needs.backend-quality.result }}" = "success"
          test "\${{ needs.backend-container.result }}" = "success"
          test "\${{ needs.backend-race.result }}" = "success"
          test "\${{ needs.postgres-integration.result }}" = "success"
`;
  backend = replaceExactly(
    backend,
    '\n# STAGE-14-1-ARCHITECTURE-CONSOLIDATION-V1-1',
    `${backendGate}\n# STAGE-14-1-ARCHITECTURE-CONSOLIDATION-V1-1`,
    'backend gate'
  );
  write('.github/workflows/backend-ci.yml', backend, cwd);

  let frontend = read('.github/workflows/frontend-ci.yml', cwd);
  frontend = patchTrigger(frontend, 'Frontend CI');
  frontend = pinUses(frontend, refs);
  const frontendGate = `

  frontend-ci-gate:
    name: Frontend CI Gate
    if: \${{ always() }}
    needs:
      - frontend-quality
    runs-on: ubuntu-latest
    timeout-minutes: 5
    steps:
      - name: Verify frontend workflow result
        shell: bash
        run: |
          test "\${{ needs.frontend-quality.result }}" = "success"
`;
  frontend = replaceExactly(
    frontend,
    '\n# STAGE-14-1-ARCHITECTURE-CONSOLIDATION-V1-1',
    `${frontendGate}\n# STAGE-14-1-ARCHITECTURE-CONSOLIDATION-V1-1`,
    'frontend gate'
  );
  write('.github/workflows/frontend-ci.yml', frontend, cwd);

  const codeql = `name: CodeQL

on:
  pull_request:
  push:
    branches:
      - main
  schedule:
    - cron: '17 3 * * 1'
  workflow_dispatch:

permissions:
  contents: read
  packages: read
  security-events: write

concurrency:
  group: codeql-\${{ github.workflow }}-\${{ github.ref }}
  cancel-in-progress: true

jobs:
  analyze:
    name: CodeQL Analyze (\${{ matrix.language }})
    runs-on: ubuntu-latest
    timeout-minutes: 30
    strategy:
      fail-fast: false
      matrix:
        include:
          - language: go
            build-mode: autobuild
          - language: javascript-typescript
            build-mode: none
    steps:
      - name: Checkout repository
        uses: actions/checkout@${refs.checkout} # v7

      - name: Initialize CodeQL
        uses: github/codeql-action/init@${refs.codeql} # v4
        with:
          languages: \${{ matrix.language }}
          build-mode: \${{ matrix.build-mode }}
          queries: security-extended

      - name: Analyze repository
        uses: github/codeql-action/analyze@${refs.codeql} # v4
        with:
          category: "/language:\${{ matrix.language }}"

  codeql-security-gate:
    name: CodeQL Security Gate
    if: \${{ always() }}
    needs:
      - analyze
    runs-on: ubuntu-latest
    timeout-minutes: 5
    steps:
      - name: Verify CodeQL analysis result
        shell: bash
        run: test "\${{ needs.analyze.result }}" = "success"
`;
  write('.github/workflows/codeql.yml', codeql, cwd);

  write('.github/CODEOWNERS', `# Repository ownership baseline
* @AsifAbbasov
`, cwd);

  write('SECURITY.md', `# Security Policy

## Supported version

Security fixes are applied to the current default branch.

## Reporting a vulnerability

Do not open a public issue containing exploit details, credentials, tokens, private
connection strings, or personal information. Use the repository's private security
advisory reporting flow when available.

Include:

- the affected component and revision;
- reproduction steps;
- expected and observed behavior;
- impact and realistic attack prerequisites;
- any suggested mitigation.

## Scope

This repository is a research and visualization platform. Reports about exposed
credentials, authorization boundaries, dependency vulnerabilities, workflow supply-chain
risk, injection, data integrity, and denial of service are in scope.

Operational aviation guidance and safety certification are outside the product scope.
`, cwd);

  const governanceVerifier = `#!/usr/bin/env node
import fs from 'node:fs';

function fail(message) {
  console.error(\`REPOSITORY_GOVERNANCE_CONTRACT=FAIL reason=\${message}\`);
  process.exit(1);
}
function read(relative) {
  return fs.readFileSync(relative, 'utf8');
}

const workflowDirectory = '.github/workflows';
const workflows = fs.readdirSync(workflowDirectory)
  .filter(name => /\\.ya?ml$/.test(name))
  .sort();

const externalUses = [];
for (const name of workflows) {
  const content = read(\`\${workflowDirectory}/\${name}\`);
  for (const match of content.matchAll(/^\\s*uses:\\s*([^\\s#]+)(?:\\s+#\\s*(\\S+))?\\s*$/gm)) {
    const reference = match[1];
    const versionComment = match[2] ?? '';
    if (reference.startsWith('./')) continue;
    externalUses.push({ name, reference, versionComment });
    if (!/^[^/@]+\\/[^/@]+(?:\\/[^@]+)?@[0-9a-f]{40}$/.test(reference)) {
      fail(\`external action is not pinned to a full SHA: \${name}: \${reference}\`);
    }
    if (!/^v\\d+(?:\\.\\d+\\.\\d+)?$/.test(versionComment)) {
      fail(\`pinned action lacks a readable version comment: \${name}: \${reference}\`);
    }
  }
}
if (externalUses.length === 0) fail('no external action references were found');

const backend = read('.github/workflows/backend-ci.yml');
const frontend = read('.github/workflows/frontend-ci.yml');
const codeql = read('.github/workflows/codeql.yml');

for (const [name, workflow] of [['Backend CI', backend], ['Frontend CI', frontend]]) {
  const trigger = workflow.slice(0, workflow.indexOf('\\npermissions:\\n'));
  if (!/on:\\n  pull_request:\\n\\n  push:/.test(trigger)) {
    fail(\`\${name} does not run for every pull request\`);
  }
}
if (!backend.includes('name: Backend CI Gate')) fail('Backend CI Gate is missing');
if (!frontend.includes('name: Frontend CI Gate')) fail('Frontend CI Gate is missing');
if (!codeql.includes('name: CodeQL Security Gate')) fail('CodeQL Security Gate is missing');
if (!codeql.includes('- language: go\\n            build-mode: autobuild')) {
  fail('CodeQL Go autobuild mode is missing');
}
if (!codeql.includes('- language: javascript-typescript\\n            build-mode: none')) {
  fail('CodeQL JavaScript TypeScript no-build mode is missing');
}
if (!codeql.includes('build-mode: \${{ matrix.build-mode }}')) {
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
if (releasePortfolioTests.split('\\n').some(line => line.includes('frontend-ci') && /length,\\s*2/.test(line))) fail('release portfolio duplicate path contract remains');
const backendOperationsTests = read('scripts/verify-backend-operations.test.mjs');
if (!backendOperationsTests.includes('trackedRenderBlueprintPath')) fail('backend operations Node single-path contract missing');
if (backendOperationsTests.split('\\n').some(line => line.includes('render') && /length,\\s*2/.test(line))) fail('backend operations duplicate path contract remains');

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
console.log(\`PINNED_EXTERNAL_ACTION_REFERENCES=\${externalUses.length}\`);
console.log('REPOSITORY_GOVERNANCE_CONTRACT=PASS');
`;
  write('scripts/verify-repository-governance.mjs', governanceVerifier, cwd);

  const governanceTests = `import test from 'node:test';
import assert from 'node:assert/strict';
import fs from 'node:fs';

const read = relative => fs.readFileSync(relative, 'utf8');

test('all external actions use immutable full SHAs with version comments', () => {
  const files = fs.readdirSync('.github/workflows').filter(name => /\\.ya?ml$/.test(name));
  let count = 0;
  for (const file of files) {
    const content = read(\`.github/workflows/\${file}\`);
    for (const match of content.matchAll(/^\\s*uses:\\s*([^\\s#]+)(?:\\s+#\\s*(\\S+))?\\s*$/gm)) {
      if (match[1].startsWith('./')) continue;
      count += 1;
      assert.match(match[1], /^[^/@]+\\/[^/@]+(?:\\/[^@]+)?@[0-9a-f]{40}$/);
      assert.match(match[2] ?? '', /^v\\d+(?:\\.\\d+\\.\\d+)?$/);
    }
  }
  assert.ok(count > 0);
});

test('required workflows expose stable gates and run for every pull request', () => {
  const backend = read('.github/workflows/backend-ci.yml');
  const frontend = read('.github/workflows/frontend-ci.yml');
  for (const workflow of [backend, frontend]) {
    const trigger = workflow.slice(0, workflow.indexOf('\\npermissions:\\n'));
    assert.match(trigger, /on:\\n  pull_request:\\n\\n  push:/);
  }
  assert.match(backend, /name: Backend CI Gate/);
  assert.match(frontend, /name: Frontend CI Gate/);
});

test('CodeQL covers Go and JavaScript TypeScript with a stable gate', () => {
  const codeql = read('.github/workflows/codeql.yml');
  assert.match(codeql, /- language: go\\n            build-mode: autobuild/);
  assert.match(codeql, /- language: javascript-typescript\\n            build-mode: none/);
  assert.match(codeql, /build-mode: \\${{ matrix\\.build-mode }}/);
  assert.match(codeql, /queries: security-extended/);
  assert.match(codeql, /name: CodeQL Security Gate/);
});

test('legacy workflow visibility contracts understand unconditional pull request CI', () => {
  const recruiter = read('scripts/verify-recruiter-quickstart.sh');
  const releasePortfolio = read('scripts/verify-release-portfolio.sh');
  const releasePortfolioTests = read('scripts/verify-release-portfolio.test.mjs');
  const backendOperationsTests = read('scripts/verify-backend-operations.test.mjs');

  assert.match(recruiter, /Backend CI must run for every pull request/);
  assert.match(recruiter, /push path filters must include README\\.md exactly once/);
  assert.doesNotMatch(recruiter, /for both push and pull_request/);

  assert.match(releasePortfolio, /Backend CI must run for every pull request/);
  assert.match(releasePortfolio, /push path filters must track frontend workflow changes exactly once/);
  assert.doesNotMatch(releasePortfolio, /for push and pull_request/);

  assert.match(releasePortfolioTests, /pull_request:\\n\\n  push:/);
  assert.ok(releasePortfolioTests.includes('backend.split(trackedFrontendWorkflowPath).length - 1, 1'));
  assert.doesNotMatch(releasePortfolioTests, /frontend-ci[^\\n]*length,\\s*2/);
  assert.match(backendOperationsTests, /pull_request:\\n\\n  push:/);
  assert.ok(backendOperationsTests.includes('source.split(trackedRenderBlueprintPath).length - 1, 1'));
  assert.doesNotMatch(backendOperationsTests, /render[^\\n]*length,\\s*2/);
});

test('ownership security and settings automation are durable repository files', () => {
  assert.match(read('.github/CODEOWNERS'), /\\* @AsifAbbasov/);
  assert.match(read('SECURITY.md'), /Reporting a vulnerability/);
  assert.match(read('scripts/configure-repository-governance.sh'), /MAIN_PROTECTION_RULESET_NAME/);
  assert.match(read('scripts/verify-repository-governance-settings.sh'), /REPOSITORY_GOVERNANCE_SETTINGS=PASS/);
});
`;
  write('scripts/verify-repository-governance.test.mjs', governanceTests, cwd);

  const settingsVerifier = `#!/usr/bin/env bash
set -euo pipefail

REPOSITORY="\${REPOSITORY:-AsifAbbasov/global-flight-analytics}"
EXPECTED_SHA="\${EXPECTED_GOVERNANCE_SHA:-}"
API_VERSION="2026-03-10"
RULESET_NAME="Global Flight Analytics main protection"

if [ -z "$EXPECTED_SHA" ]; then
  printf '%s\\n' 'EXPECTED_GOVERNANCE_SHA is required' >&2
  exit 1
fi

repository_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY")"
actions_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/actions/permissions")"
workflow_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/actions/permissions/workflow")"
selected_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/actions/permissions/selected-actions")"
rulesets_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/rulesets?per_page=100")"
rules_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/rules/branches/main")"
analyses_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/code-scanning/analyses?ref=refs/heads/main&per_page=100")"
dependabot_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/dependabot/alerts?state=open&per_page=100")"
secrets_json="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/secret-scanning/alerts?state=open&per_page=100")"

REPOSITORY_JSON="$repository_json" ACTIONS_JSON="$actions_json" WORKFLOW_JSON="$workflow_json" \\
SELECTED_JSON="$selected_json" RULESETS_JSON="$rulesets_json" RULES_JSON="$rules_json" \\
ANALYSES_JSON="$analyses_json" DEPENDABOT_JSON="$dependabot_json" SECRETS_JSON="$secrets_json" \\
EXPECTED_SHA="$EXPECTED_SHA" RULESET_NAME="$RULESET_NAME" node <<'NODE'
const repository = JSON.parse(process.env.REPOSITORY_JSON);
const actions = JSON.parse(process.env.ACTIONS_JSON);
const workflow = JSON.parse(process.env.WORKFLOW_JSON);
const selected = JSON.parse(process.env.SELECTED_JSON);
const rulesets = JSON.parse(process.env.RULESETS_JSON);
const rules = JSON.parse(process.env.RULES_JSON);
const analyses = JSON.parse(process.env.ANALYSES_JSON);
JSON.parse(process.env.DEPENDABOT_JSON);
JSON.parse(process.env.SECRETS_JSON);

function assert(condition, message) {
  if (!condition) throw new Error(message);
}
assert(repository.default_branch === 'main', 'default branch is not main');
assert(repository.allow_merge_commit === false, 'merge commits are still allowed');
assert(repository.allow_rebase_merge === false, 'rebase merges are still allowed');
assert(repository.allow_squash_merge === true, 'squash merge is not allowed');
assert(repository.delete_branch_on_merge === true, 'delete branch on merge is disabled');
assert(repository.allow_update_branch === true, 'update branch is disabled');
assert(actions.enabled === true, 'Actions are disabled');
assert(actions.allowed_actions === 'selected', 'Actions are not restricted to selected');
assert(actions.sha_pinning_required === true, 'Action SHA pinning policy is disabled');
assert(workflow.default_workflow_permissions === 'read', 'workflow token is not read-only');
assert(workflow.can_approve_pull_request_reviews === false, 'workflow token can approve PRs');
assert(selected.github_owned_allowed === true, 'GitHub-owned actions are not allowed');
assert(selected.verified_allowed === false, 'all verified actions are unexpectedly allowed');
assert((selected.patterns_allowed ?? []).includes('pnpm/action-setup@*'), 'pnpm action allowlist missing');

const ruleset = rulesets.find(item => item.name === process.env.RULESET_NAME);
assert(ruleset?.enforcement === 'active', 'active main ruleset missing');
const types = new Set(rules.map(rule => rule.type));
for (const type of ['deletion', 'non_fast_forward', 'required_linear_history', 'pull_request', 'required_status_checks', 'code_scanning']) {
  assert(types.has(type), `active main rule missing: ${type}`);
}
for (const context of ['Backend CI Gate', 'Frontend CI Gate', 'CodeQL Security Gate']) {
  assert(JSON.stringify(rules).includes(context), `required status context missing: ${context}`);
}
assert(analyses.some(item => item.commit_sha === process.env.EXPECTED_SHA && item.tool?.name === 'CodeQL'), 'CodeQL analysis for expected SHA missing');
console.log('DEPENDABOT_ALERTS_API=PASS');
console.log('SECRET_SCANNING_ALERTS_API=PASS');
console.log('CODEQL_ANALYSIS=PASS');
console.log('MAIN_RULESET=PASS');
console.log('ACTIONS_POLICY=PASS');
console.log('REPOSITORY_GOVERNANCE_SETTINGS=PASS');
NODE
`;
  write('scripts/verify-repository-governance-settings.sh', settingsVerifier, cwd);

  const configure = `#!/usr/bin/env bash
set -euo pipefail

REPOSITORY="\${REPOSITORY:-AsifAbbasov/global-flight-analytics}"
EXPECTED_SHA="\${EXPECTED_GOVERNANCE_SHA:-}"
API_VERSION="2026-03-10"
MAIN_PROTECTION_RULESET_NAME="Global Flight Analytics main protection"
ARCHIVE_BRANCH="docs/open-aviation-metrics-positioning"
ARCHIVE_TAG="archive/open-aviation-metrics-positioning-2026-08-03"
MERGED_BRANCH="feature/active-aircraft-metric"

if [ -z "$EXPECTED_SHA" ]; then
  printf '%s\\n' 'EXPECTED_GOVERNANCE_SHA is required' >&2
  exit 1
fi

test "$(git symbolic-ref --quiet --short HEAD)" = "main"
test -z "$(git status --porcelain)"
git fetch --prune origin
test "$(git rev-parse HEAD)" = "$EXPECTED_SHA"
test "$(git rev-parse origin/main)" = "$EXPECTED_SHA"
auth_status="$(gh auth status --hostname github.com 2>&1)"
printf '%s\\n' "$auth_status" | grep -F 'admin:repo_hook' >/dev/null || {
  printf '%s\\n' 'GitHub CLI token requires admin:repo_hook scope.' >&2
  printf '%s\\n' 'Run: gh auth refresh -h github.com -s admin:repo_hook' >&2
  exit 1
}

node --test scripts/verify-repository-governance.test.mjs
node scripts/verify-repository-governance.mjs

runs_url="https://api.github.com/repos/$REPOSITORY/actions/runs?head_sha=$EXPECTED_SHA&per_page=100"
attempt=1
while [ "$attempt" -le 120 ]; do
  payload="$(curl --fail --silent --show-error --location "$runs_url")"
  result="$(RUNS_PAYLOAD="$payload" EXPECTED_SHA="$EXPECTED_SHA" node <<'NODE'
const data = JSON.parse(process.env.RUNS_PAYLOAD);
const required = ['Backend CI', 'Frontend CI', 'CodeQL'];
const selected = required.map(name => (data.workflow_runs ?? []).find(run =>
  run.name === name &&
  run.head_sha === process.env.EXPECTED_SHA &&
  run.event === 'push'
));
if (selected.some(run => !run || run.status !== 'completed')) {
  process.stdout.write('PENDING');
} else if (selected.every(run => run.conclusion === 'success')) {
  for (const run of selected) console.log(`${run.name} run_id=${run.id} conclusion=success`);
  console.log('SUCCESS');
} else {
  for (const run of selected) console.log(`${run.name} run_id=${run.id} conclusion=${run.conclusion}`);
  console.log('FAILURE');
}
NODE
)"
  case "$result" in
    *SUCCESS) printf '%s\\n' "$result"; break ;;
    *FAILURE) printf '%s\\n' "$result" >&2; exit 1 ;;
  esac
  sleep 5
  attempt=$((attempt + 1))
done
if [ "$attempt" -gt 120 ]; then
  printf '%s\\n' 'GOVERNANCE_WORKFLOWS=TIMEOUT' >&2
  exit 1
fi

gh api --method PUT -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/vulnerability-alerts" >/dev/null
gh api --method PUT -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/automated-security-fixes" >/dev/null

repository_payload="$(mktemp)"
actions_payload="$(mktemp)"
selected_payload="$(mktemp)"
ruleset_payload="$(mktemp)"
trap 'rm -f "$repository_payload" "$actions_payload" "$selected_payload" "$ruleset_payload"' EXIT

cat >"$repository_payload" <<'JSON'
{
  "allow_merge_commit": false,
  "allow_squash_merge": true,
  "allow_rebase_merge": false,
  "allow_auto_merge": true,
  "delete_branch_on_merge": true,
  "allow_update_branch": true,
  "security_and_analysis": {
    "secret_scanning": { "status": "enabled" },
    "secret_scanning_push_protection": { "status": "enabled" }
  }
}
JSON
gh api --method PATCH -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY" --input "$repository_payload" >/dev/null

cat >"$actions_payload" <<'JSON'
{
  "enabled": true,
  "allowed_actions": "selected",
  "sha_pinning_required": true
}
JSON
gh api --method PUT -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/actions/permissions" --input "$actions_payload" >/dev/null

cat >"$selected_payload" <<'JSON'
{
  "github_owned_allowed": true,
  "verified_allowed": false,
  "patterns_allowed": ["pnpm/action-setup@*"]
}
JSON
gh api --method PUT -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/actions/permissions/selected-actions" --input "$selected_payload" >/dev/null

gh api --method PUT -H "X-GitHub-Api-Version: $API_VERSION" \\
  "repos/$REPOSITORY/actions/permissions/workflow" \\
  -f default_workflow_permissions=read \\
  -F can_approve_pull_request_reviews=false >/dev/null

repository_id="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY" --jq '.id')"
cat >"$ruleset_payload" <<JSON
{
  "name": "$MAIN_PROTECTION_RULESET_NAME",
  "target": "branch",
  "enforcement": "active",
  "conditions": {
    "ref_name": {
      "include": ["~DEFAULT_BRANCH"],
      "exclude": []
    }
  },
  "rules": [
    { "type": "deletion" },
    { "type": "non_fast_forward" },
    { "type": "required_linear_history" },
    {
      "type": "pull_request",
      "parameters": {
        "allowed_merge_methods": ["squash"],
        "dismiss_stale_reviews_on_push": true,
        "require_code_owner_review": false,
        "require_last_push_approval": false,
        "required_approving_review_count": 0,
        "required_review_thread_resolution": true
      }
    },
    {
      "type": "required_status_checks",
      "parameters": {
        "do_not_enforce_on_create": false,
        "strict_required_status_checks_policy": true,
        "required_status_checks": [
          { "context": "Backend CI Gate" },
          { "context": "Frontend CI Gate" },
          { "context": "CodeQL Security Gate" }
        ]
      }
    },
    {
      "type": "code_scanning",
      "parameters": {
        "code_scanning_tools": [
          {
            "tool": "CodeQL",
            "alerts_threshold": "errors",
            "security_alerts_threshold": "high_or_higher"
          }
        ]
      }
    }
  ]
}
JSON

existing_ruleset_id="$(gh api -H "X-GitHub-Api-Version: $API_VERSION" "repos/$REPOSITORY/rulesets?per_page=100" --jq ".[] | select(.name == \\"$MAIN_PROTECTION_RULESET_NAME\\") | .id" | head -n 1)"
if [ -n "$existing_ruleset_id" ]; then
  gh api --method PUT -H "X-GitHub-Api-Version: $API_VERSION" \\
    "repos/$REPOSITORY/rulesets/$existing_ruleset_id" --input "$ruleset_payload" >/dev/null
else
  gh api --method POST -H "X-GitHub-Api-Version: $API_VERSION" \\
    "repos/$REPOSITORY/rulesets" --input "$ruleset_payload" >/dev/null
fi

git fetch --prune origin
if git show-ref --verify --quiet "refs/remotes/origin/$ARCHIVE_BRANCH"; then
  archive_head="$(git rev-parse "origin/$ARCHIVE_BRANCH")"
  branch_only="$(git rev-list --left-right --count "origin/main...origin/$ARCHIVE_BRANCH" | awk '{print $2}')"
  test "$branch_only" -gt 0
  if git ls-remote --exit-code --tags origin "refs/tags/$ARCHIVE_TAG" >/dev/null 2>&1; then
    test "$(git ls-remote --tags origin "refs/tags/$ARCHIVE_TAG^{}" "refs/tags/$ARCHIVE_TAG" | tail -n 1 | awk '{print $1}')" = "$archive_head"
  else
    git tag -a "$ARCHIVE_TAG" "$archive_head" -m "Archive abandoned Open Aviation Metrics positioning branch"
    git push origin "refs/tags/$ARCHIVE_TAG"
  fi
  git push origin --delete "$ARCHIVE_BRANCH"
fi

git fetch --prune origin
if git show-ref --verify --quiet "refs/remotes/origin/$MERGED_BRANCH"; then
  branch_only="$(git rev-list --left-right --count "origin/main...origin/$MERGED_BRANCH" | awk '{print $2}')"
  test "$branch_only" = "0"
  git push origin --delete "$MERGED_BRANCH"
fi

EXPECTED_GOVERNANCE_SHA="$EXPECTED_SHA" bash scripts/verify-repository-governance-settings.sh
printf '%s\\n' 'REPOSITORY_GOVERNANCE_CONFIGURATION=PASS'
`;
  write('scripts/configure-repository-governance.sh', configure, cwd);

  let root = JSON.parse(read('package.json', cwd));
  root.scripts['test:repository-governance'] = 'node --test scripts/verify-repository-governance.test.mjs';
  root.scripts['verify:repository-governance'] = 'node scripts/verify-repository-governance.mjs';
  write('package.json', `${JSON.stringify(root, null, 2)}\n`, cwd);

  let release = read('scripts/verify-release.sh', cwd);
  release = replaceExactly(
    release,
    'pnpm run verify:dependency-maintenance\n',
    'pnpm run verify:dependency-maintenance\npnpm run test:repository-governance\npnpm run verify:repository-governance\n',
    'release governance integration'
  );
  release = replaceExactly(
    release,
    "printf '%s\\n' 'RELEASE_RECRUITER_QUICKSTART=PASS'\n",
    "printf '%s\\n' 'RELEASE_REPOSITORY_GOVERNANCE=PASS'\nprintf '%s\\n' 'RELEASE_RECRUITER_QUICKSTART=PASS'\n",
    'release marker'
  );
  write('scripts/verify-release.sh', release, cwd);

  let dependencyVerifier = read('scripts/verify-dependency-maintenance.mjs', cwd);
  dependencyVerifier = dependencyVerifier.replace(
    "if (backendCI.includes('actions/setup-go@v6') || !backendCI.includes('actions/setup-go@v7')) fail('setup-go v7 contract failed');",
    "if (!/actions\\/setup-go@[0-9a-f]{40}\\s+# v7/.test(backendCI)) fail('setup-go v7 immutable pin contract failed');"
  );
  dependencyVerifier = dependencyVerifier.replace(
    "if (frontendCI.includes('actions/setup-node@v6') || !frontendCI.includes('actions/setup-node@v7')) fail('setup-node v7 contract failed');",
    "if (!/actions\\/setup-node@[0-9a-f]{40}\\s+# v7/.test(frontendCI)) fail('setup-node v7 immutable pin contract failed');"
  );
  write('scripts/verify-dependency-maintenance.mjs', dependencyVerifier, cwd);

  let dependencyTests = read('scripts/verify-dependency-maintenance.test.mjs', cwd);
  dependencyTests = dependencyTests.replace(
    "test('CI setup actions use version 7 only', () => {",
    "test('CI setup actions use immutable version 7 pins', () => {"
  );
  dependencyTests = dependencyTests.replace(
    "  assert.match(backend, /actions\\/setup-go@v7/);\n  assert.doesNotMatch(backend, /actions\\/setup-go@v6/);\n  assert.match(frontend, /actions\\/setup-node@v7/);\n  assert.doesNotMatch(frontend, /actions\\/setup-node@v6/);",
    "  assert.match(backend, /actions\\/setup-go@[0-9a-f]{40}\\s+# v7/);\n  assert.match(frontend, /actions\\/setup-node@[0-9a-f]{40}\\s+# v7/);"
  );
  write('scripts/verify-dependency-maintenance.test.mjs', dependencyTests, cwd);

  let recruiterQuickstart = read('scripts/verify-recruiter-quickstart.sh', cwd);
  recruiterQuickstart = migrateRecruiterQuickstartContract(recruiterQuickstart);
  write('scripts/verify-recruiter-quickstart.sh', recruiterQuickstart, cwd);

  let releasePortfolio = read('scripts/verify-release-portfolio.sh', cwd);
  releasePortfolio = migrateReleasePortfolioContract(releasePortfolio);
  write('scripts/verify-release-portfolio.sh', releasePortfolio, cwd);

  let releasePortfolioTests = read('scripts/verify-release-portfolio.test.mjs', cwd);
  releasePortfolioTests = migrateReleasePortfolioTests(releasePortfolioTests);
  write('scripts/verify-release-portfolio.test.mjs', releasePortfolioTests, cwd);

  let backendOperationsTests = read('scripts/verify-backend-operations.test.mjs', cwd);
  backendOperationsTests = migrateBackendOperationsTests(backendOperationsTests);
  write('scripts/verify-backend-operations.test.mjs', backendOperationsTests, cwd);
  assertLegacyWorkflowContractsRemoved(cwd);

  const document = `# Repository Governance and Security Automation

<!-- REPOSITORY-GOVERNANCE-SECURITY-AUTOMATION-V1 -->

Status: Repository patch prepared; GitHub settings activation pending exact-commit CI
Date: 2026-08-03
Baseline: \`${BASELINE}\`

## Audit finding

The public repository had no active ruleset or legacy branch protection. Direct pushes,
force pushes and branch deletion were not blocked by GitHub. Actions accepted mutable
major tags, all third-party actions were allowed, CodeQL had no analysis, secret scanning
was disabled, Dependabot alerts were disabled, CODEOWNERS was absent and two stale remote
branches remained.

## Repository patch

The repository patch establishes:

- full immutable SHA pins for every external GitHub Action with readable release comments;
- Backend CI and Frontend CI on every pull request targeting the default branch;
- stable Backend CI Gate and Frontend CI Gate contexts;
- CodeQL security-extended analysis for Go and JavaScript/TypeScript;
- a stable CodeQL Security Gate context;
- CODEOWNERS and SECURITY policy files;
- permanent governance contract tests in the release gate;
- migrated recruiter, release, and backend-operations verifiers that recognize unconditional pull request CI while preserving explicit push visibility paths;
- idempotent GitHub settings configuration and verification scripts.

## GitHub settings phase

Settings activation occurs only after the patch commit has passed Backend CI, Frontend CI
and CodeQL on the exact commit SHA. The configuration then restricts Actions, enables
security alerts and secret protection, activates the main branch ruleset and reconciles
stale branches.

The solo-owner repository requires pull requests but does not require an approving review
from another person. Review conversations must be resolved. The ruleset allows squash merge
only, requires the three stable gates, enforces CodeQL high-severity security results,
prevents force pushes and deletion, and requires linear history.

## Branch reconciliation

The fully merged \`feature/active-aircraft-metric\` branch has no unique commits and is
deleted after verification.

The \`docs/open-aviation-metrics-positioning\` branch contains two unique commits for an
abandoned Open Aviation Metrics API product pivot. Its head is preserved as the annotated
tag \`archive/open-aviation-metrics-positioning-2026-08-03\` before the active branch is
deleted.

## Installer release quality contract

Every package that changes this repository must be validated as a complete transaction.
A package must not be released after only one discovered assertion or one-line failure is
fixed. Before a package may be provided, its installer must:

- execute positive transformation and validation fixtures;
- execute negative fixtures that prove malformed input and stale contracts are rejected;
- execute a rollback fixture that deliberately fails after mutation and proves restoration;
- prove that the Go source tree is formatted before the patch;
- run \`gofmt\` after the patch and prove the Go source digest is unchanged;
- run the complete repository release validation in a detached worktree;
- verify the exact changed-file manifest;
- roll the real repository back to the exact baseline after any patch,
  formatting, validation or manifest failure.

## Closure sequence

\`\`\`text
INSTALLER_POSITIVE_FIXTURES=PASS
INSTALLER_NEGATIVE_FIXTURES=PASS
INSTALLER_ROLLBACK_SELF_TEST=PASS
GOFMT_BEFORE_AND_AFTER=PASS
FULL_REPOSITORY_VALIDATION=PASS
EXACT_BASELINE_GITHUB_ACTIONS_VALIDATION=REQUIRED_BEFORE_PACKAGE_RELEASE
REPOSITORY_GOVERNANCE_PATCH=PASS
BACKEND_CI_GATE=PASS
FRONTEND_CI_GATE=PASS
CODEQL_SECURITY_GATE=PASS
DEPENDABOT_ALERTS=ENABLED
SECRET_SCANNING=ENABLED
SECRET_PUSH_PROTECTION=ENABLED
ACTIONS_SHA_PINNING=ENFORCED
MAIN_RULESET=ACTIVE
STALE_BRANCH_HISTORY=PRESERVED
REPOSITORY_GOVERNANCE_STAGE=READY_FOR_SETTINGS
\`\`\`
`;
  write('docs/172_REPOSITORY_GOVERNANCE_AND_SECURITY_AUTOMATION.md', document, cwd);

  let index = read('docs/DOCUMENT_INDEX.md', cwd);
  if (!index.includes('## Document 172 — Repository Governance and Security Automation')) {
    index = `${index.trimEnd()}

## Document 172 — Repository Governance and Security Automation

- File: \`172_REPOSITORY_GOVERNANCE_AND_SECURITY_AUTOMATION.md\`
- Status: \`PATCH PREPARED; SETTINGS PENDING EXACT-COMMIT CI\`
- Purpose: establishes immutable Action pins, stable required CI gates, CodeQL, ownership and security policy files, reproducible GitHub settings automation, protected main-branch governance, and history-preserving stale-branch reconciliation.
`;
  }
  write('docs/DOCUMENT_INDEX.md', index, cwd);
}

function validate(cwd, expectedGoDigest, marker) {
  formatAndVerifyGoTree(
    cwd,
    expectedGoDigest,
    `${marker}_GOFMT_IDENTITY`
  );
  verifyChangedManifest(cwd);
  run('node', ['--test', 'scripts/verify-repository-governance.test.mjs'], { cwd });
  run('node', ['scripts/verify-repository-governance.mjs'], { cwd });
  run('node', ['--test', 'scripts/verify-dependency-maintenance.test.mjs'], { cwd });
  run('node', ['scripts/verify-dependency-maintenance.mjs'], { cwd });
  run('bash', ['-n', 'scripts/configure-repository-governance.sh'], { cwd });
  run('bash', ['-n', 'scripts/verify-repository-governance-settings.sh'], { cwd });
  run('bash', ['scripts/verify-release.sh'], { cwd });
  run('git', ['diff', '--check'], { cwd });
  verifyChangedManifest(cwd);
  console.log(`${marker}_FULL_REPOSITORY_VALIDATION=PASS`);
}

function cleanupInstallerWorktrees() {
  const temporaryRoot = `${path.resolve(os.tmpdir())}${path.sep}`;
  const listing = git(['worktree', 'list', '--porcelain']);

  for (const block of listing.split(/\n\n+/)) {
    const firstLine = block
      .split('\n')
      .find(line => line.startsWith('worktree '));
    if (!firstLine) continue;

    const worktreePath = firstLine.slice('worktree '.length);
    if (
      path.resolve(worktreePath).startsWith(temporaryRoot) &&
      path.basename(worktreePath).startsWith('gfa-governance-')
    ) {
      fs.rmSync(worktreePath, { recursive: true, force: true });
    }
  }

  const prune = spawnSync(
    'git',
    ['worktree', 'prune', '--expire', 'now'],
    {
      cwd: PROJECT_ROOT,
      encoding: 'utf8',
      stdio: 'pipe',
    }
  );
  if (prune.status !== 0) {
    process.stderr.write(prune.stdout ?? '');
    process.stderr.write(prune.stderr ?? '');
    throw new Error('Git 2.15 worktree prune failed');
  }

  console.log('STALE_INSTALLER_WORKTREE_CLEANUP=PASS');
}

function validateExactBaselineOnly() {
  selfTest();
  if (!fs.existsSync(PROJECT_ROOT)) throw new Error(`project root not found: ${PROJECT_ROOT}`);
  if (spawnSync('git', ['cat-file', '-e', `${BASELINE}^{commit}`], { cwd: PROJECT_ROOT, encoding: 'utf8', stdio: 'pipe' }).status !== 0) throw new Error(`exact baseline commit is unavailable: ${BASELINE}`);
  const validationRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'gfa-governance-exact-baseline-'));
  let validationError = null; let cleanupError = null;
  try {
    run('git', ['worktree', 'add', '--detach', validationRoot, BASELINE], { cwd: PROJECT_ROOT });
    const goDigest = verifyGoFormattingBaseline(validationRoot, 'EXACT_BASELINE_GO_FORMATTING_BEFORE_PATCH');
    patchRepository(validationRoot, VALIDATED_ACTION_REFS);
    verifyChangedManifest(validationRoot);
    console.log('EXACT_BASELINE_CHANGED_FILE_MANIFEST=PASS');
    validate(validationRoot, goDigest, 'EXACT_BASELINE');
    console.log('EXACT_BASELINE_REPOSITORY_VALIDATION=PASS');
    console.log('EXACT_BASELINE_GITHUB_ACTIONS_VALIDATION=PASS');
  } catch (error) { validationError = error; }
  finally {
    fs.rmSync(validationRoot, { recursive: true, force: true });
    const prune = spawnSync('git', ['worktree', 'prune', '--expire', 'now'], { cwd: PROJECT_ROOT, encoding: 'utf8', stdio: 'pipe' });
    if (prune.status !== 0) cleanupError = new Error('exact baseline worktree cleanup failed');
    else console.log('EXACT_BASELINE_WORKTREE_CLEANUP=PASS');
  }
  if (validationError) throw validationError;
  if (cleanupError) throw cleanupError;
}

function selfTest() {
  const fake = {
    checkout: '1'.repeat(40),
    setupGo: '2'.repeat(40),
    setupNode: '3'.repeat(40),
    pnpmSetup: '4'.repeat(40),
    codeql: '5'.repeat(40),
  };
  for (const value of Object.values(fake)) {
    if (!/^[0-9a-f]{40}$/.test(value)) throw new Error('fixture SHA invalid');
  }

  const gofmtFixture = fs.mkdtempSync(
    path.join(os.tmpdir(), 'gfa-governance-gofmt-')
  );
  try {
    const goFile = path.join(
      gofmtFixture,
      'apps',
      'api',
      'main.go'
    );
    fs.mkdirSync(path.dirname(goFile), { recursive: true });
    fs.writeFileSync(goFile, 'package main\n\nfunc main() {}\n');
    const formattedDigest = verifyGoFormattingBaseline(
      gofmtFixture,
      'GOFMT_POSITIVE_FIXTURE'
    );
    formatAndVerifyGoTree(
      gofmtFixture,
      formattedDigest,
      'GOFMT_IDENTITY_POSITIVE_FIXTURE'
    );

    fs.writeFileSync(goFile, 'package main\nfunc main(){println("x")}\n');
    expectFailure(
      'unformatted Go negative fixture',
      () => verifyGoFormattingBaseline(
        gofmtFixture,
        'GOFMT_NEGATIVE_FIXTURE'
      ),
      /not formatted before patch/
    );

    const beforeFormatting = hashGoTree(gofmtFixture);
    expectFailure(
      'Go identity mutation negative fixture',
      () => formatAndVerifyGoTree(
        gofmtFixture,
        beforeFormatting,
        'GOFMT_MUTATION_NEGATIVE_FIXTURE'
      ),
      /Go source changed during patch formatting/
    );
  } finally {
    fs.rmSync(gofmtFixture, { recursive: true, force: true });
  }

  const sample = 'uses: actions/checkout@v7\nuses: pnpm/action-setup@v6\n';
  const pinned = pinUses(sample, fake);
  if (!pinned.includes(`actions/checkout@${fake.checkout} # v7`)) {
    throw new Error('checkout pin positive fixture failed');
  }
  if (!pinned.includes(`pnpm/action-setup@${fake.pnpmSetup} # v6`)) {
    throw new Error('pnpm pin positive fixture failed');
  }

  const workflowFixture = `name: Frontend CI

on:
  pull_request:
    paths:
      - 'apps/web/**'
      - 'docs/171_DEPENDABOT_FOLLOW_UP_RECONCILIATION.md'

  push:
    branches:
      - main
    paths:
      - 'apps/web/**'
      - 'docs/171_DEPENDABOT_FOLLOW_UP_RECONCILIATION.md'

permissions:
  contents: read
`;
  const patchedTrigger = patchTrigger(workflowFixture, 'Frontend CI');
  if (!patchedTrigger.includes('on:\n  pull_request:\n\n  push:')) {
    throw new Error('global pull request trigger positive fixture failed');
  }
  if (
    !patchedTrigger.includes(
      'docs/171_DEPENDABOT_FOLLOW_UP_RECONCILIATION.md'
    )
  ) {
    throw new Error('push path preservation positive fixture failed');
  }

  expectFailure(
    'missing push trigger negative fixture',
    () => patchTrigger(
      `name: Frontend CI

on:
  pull_request:

permissions:
  contents: read
`,
      'Frontend CI'
    ),
    /push trigger is missing/
  );
  expectFailure(
    'exact replacement zero-match negative fixture',
    () => replaceExactly('alpha', 'beta', 'gamma', 'fixture'),
    /found 0/
  );
  expectFailure(
    'exact replacement multi-match negative fixture',
    () => replaceExactly('beta beta', 'beta', 'gamma', 'fixture'),
    /found 2/
  );
  expectFailure(
    'semantic replacement zero-match negative fixture',
    () => replacePatternExactly('alpha', /beta/, 'gamma', 'fixture'),
    /found 0/
  );
  expectFailure(
    'semantic replacement multi-match negative fixture',
    () => replacePatternExactly('beta beta', /beta/, 'gamma', 'fixture'),
    /more than 1/
  );

  const recruiterLegacyFixture = `require_literal "$WORKFLOW_FILE" "- 'README.md'" \\
  'Backend CI path filters do not include README.md'

readme_path_count="$(
  grep -F -c -- "- 'README.md'" "$WORKFLOW_FILE"
)"
if [ "$readme_path_count" -ne 2 ]; then
  fail 'Backend CI must trigger on README.md for both push and pull_request'
fi

printf '%s\n' "RECRUITER_QUICKSTART_CONTRACT=PASS"
`;
  const migratedRecruiter = migrateRecruiterQuickstartContract(
    recruiterLegacyFixture
  );
  if (
    !migratedRecruiter.includes(
      'Backend CI push path filters must include README.md exactly once'
    ) ||
    migratedRecruiter.includes(
      'Backend CI must trigger on README.md for both push and pull_request'
    )
  ) {
    throw new Error('recruiter semantic migration positive fixture failed');
  }
  expectFailure(
    'recruiter already-migrated negative fixture',
    () => migrateRecruiterQuickstartContract(migratedRecruiter),
    /found 0/
  );

  const releaseLegacyFixture = `backend_frontend_workflow_count="$(
  grep -F -c -- "- '.github/workflows/frontend-ci.yml'" \\
    "$REPOSITORY_ROOT/.github/workflows/backend-ci.yml"
)"
[ "$backend_frontend_workflow_count" -eq 2 ] || \\
  fail 'Backend CI must track frontend workflow changes for push and pull_request'

require_literal apps/api/.env.example 'Use a direct PostgreSQL connection string for migrations' \\
  'API environment example does not distinguish migration connection semantics'
`;
  const migratedRelease = migrateReleasePortfolioContract(
    releaseLegacyFixture
  );
  if (
    !migratedRelease.includes(
      'Backend CI push path filters must track frontend workflow changes exactly once'
    ) ||
    migratedRelease.includes(
      'Backend CI must track frontend workflow changes for push and pull_request'
    )
  ) {
    throw new Error('release semantic migration positive fixture failed');
  }
  expectFailure(
    'release duplicate-anchor negative fixture',
    () => migrateReleasePortfolioContract(
      `${releaseLegacyFixture}\n${releaseLegacyFixture}`
    ),
    /more than 1/
  );

  const releaseTestLegacyFixture = `  assert.equal((backend.match(/- '\\.github\\/workflows\\/frontend-ci\\.yml'/g) ?? []).length, 2)
  assert.match(backend, /Verify release and portfolio contract/)
`;
  const migratedReleaseTest = migrateReleasePortfolioTests(
    releaseTestLegacyFixture
  );
  if (
    !migratedReleaseTest.includes(
      'backend.split(trackedFrontendWorkflowPath).length - 1, 1'
    )
  ) {
    throw new Error('release Node test migration positive fixture failed');
  }
  expectFailure(
    'release Node test missing-anchor negative fixture',
    () => migrateReleasePortfolioTests(
      'assert.match(backend, /Verify release and portfolio contract/)\n'
    ),
    /found 0/
  );
  const backendOperationsLegacyFixture = `test('Backend CI permanently enforces operations contracts and Blueprint changes', () => {
  const source = read('.github/workflows/backend-ci.yml')
  assert.equal((source.match(/- 'render\\.yaml'/g) ?? []).length, 2)
})
`;
  const migratedBackendOperationsTest = migrateBackendOperationsTests(backendOperationsLegacyFixture);
  if (!migratedBackendOperationsTest.includes('source.split(trackedRenderBlueprintPath).length - 1, 1')) throw new Error('backend operations migration fixture failed');
  expectFailure('backend operations missing-anchor negative fixture', () => migrateBackendOperationsTests('assert.match(source, /Verify backend operations contract/)\n'), /found 0/);

  const fixtureRoot = fs.mkdtempSync(
    path.join(os.tmpdir(), 'gfa-governance-fixtures-')
  );
  try {
    const goodWorkflow = `name: Backend CI

on:
  pull_request:

  push:
    branches:
      - main
    paths:
      - 'README.md'
      - '.github/workflows/frontend-ci.yml'

permissions:
  contents: read
`;
    const badPullRequestWorkflow = `name: Backend CI

on:
  pull_request:
    paths:
      - 'README.md'

  push:
    branches:
      - main
    paths:
      - 'README.md'
      - '.github/workflows/frontend-ci.yml'

permissions:
  contents: read
`;
    const missingReadmeWorkflow = `name: Backend CI

on:
  pull_request:

  push:
    branches:
      - main
    paths:
      - '.github/workflows/frontend-ci.yml'

permissions:
  contents: read
`;
    const missingFrontendWorkflow = `name: Backend CI

on:
  pull_request:

  push:
    branches:
      - main
    paths:
      - 'README.md'

permissions:
  contents: read
`;

    const workflowPath = path.join(fixtureRoot, 'backend-ci.yml');
    fs.writeFileSync(workflowPath, goodWorkflow);

    const recruiterScript = path.join(
      fixtureRoot,
      'verify-recruiter-fixture.sh'
    );
    const fullRecruiterFixture = migrateRecruiterQuickstartContract(`#!/usr/bin/env bash
set -euo pipefail

fail() {
  printf '%s\\n' "$1" >&2
  exit 1
}

require_literal() {
  file="$1"
  literal="$2"
  message="$3"
  grep -F -- "$literal" "$file" >/dev/null || fail "$message"
}

WORKFLOW_FILE="$1"

require_literal "$WORKFLOW_FILE" "- 'README.md'" \\
  'Backend CI path filters do not include README.md'

readme_path_count="$(
  grep -F -c -- "- 'README.md'" "$WORKFLOW_FILE"
)"
if [ "$readme_path_count" -ne 2 ]; then
  fail 'Backend CI must trigger on README.md for both push and pull_request'
fi

printf '%s\\n' 'RECRUITER_FIXTURE=PASS'
`);
    fs.writeFileSync(recruiterScript, fullRecruiterFixture);
    fs.chmodSync(recruiterScript, 0o755);

    const repositoryRoot = path.join(fixtureRoot, 'repository');
    fs.mkdirSync(
      path.join(repositoryRoot, '.github', 'workflows'),
      { recursive: true }
    );
    const repositoryWorkflowPath = path.join(
      repositoryRoot,
      '.github',
      'workflows',
      'backend-ci.yml'
    );
    fs.writeFileSync(repositoryWorkflowPath, goodWorkflow);

    const releaseScript = path.join(
      fixtureRoot,
      'verify-release-portfolio-fixture.sh'
    );
    const fullReleaseFixture = migrateReleasePortfolioContract(`#!/usr/bin/env bash
set -euo pipefail

fail() {
  printf '%s\\n' "$1" >&2
  exit 1
}

REPOSITORY_ROOT="$1"

backend_frontend_workflow_count="$(
  grep -F -c -- "- '.github/workflows/frontend-ci.yml'" \\
    "$REPOSITORY_ROOT/.github/workflows/backend-ci.yml"
)"
[ "$backend_frontend_workflow_count" -eq 2 ] || \\
  fail 'Backend CI must track frontend workflow changes for push and pull_request'

printf '%s\\n' 'RELEASE_PORTFOLIO_FIXTURE=PASS'
`);
    fs.writeFileSync(releaseScript, fullReleaseFixture);
    fs.chmodSync(releaseScript, 0o755);

    function expectProcessStatus(label, command, args, expectedStatus) {
      const result = spawnSync(
        command,
        args,
        {
          cwd: fixtureRoot,
          encoding: 'utf8',
          stdio: 'pipe',
        }
      );
      const actualStatus = result.status ?? 1;
      if (
        (expectedStatus === 0 && actualStatus !== 0) ||
        (expectedStatus !== 0 && actualStatus === 0)
      ) {
        process.stderr.write(result.stdout ?? '');
        process.stderr.write(result.stderr ?? '');
        throw new Error(
          `${label}: expected status ${expectedStatus === 0 ? 'success' : 'failure'}, got ${actualStatus}`
        );
      }
      return result;
    }

    expectProcessStatus(
      'recruiter positive process fixture',
      'bash',
      [recruiterScript, workflowPath],
      0
    );
    expectProcessStatus(
      'release positive process fixture',
      'bash',
      [releaseScript, repositoryRoot],
      0
    );

    fs.writeFileSync(workflowPath, badPullRequestWorkflow);
    fs.writeFileSync(repositoryWorkflowPath, badPullRequestWorkflow);
    expectProcessStatus(
      'recruiter pull request paths negative process fixture',
      'bash',
      [recruiterScript, workflowPath],
      1
    );
    expectProcessStatus(
      'release pull request paths negative process fixture',
      'bash',
      [releaseScript, repositoryRoot],
      1
    );

    fs.writeFileSync(workflowPath, missingReadmeWorkflow);
    expectProcessStatus(
      'recruiter missing README negative process fixture',
      'bash',
      [recruiterScript, workflowPath],
      1
    );

    fs.writeFileSync(repositoryWorkflowPath, missingFrontendWorkflow);
    expectProcessStatus(
      'release missing workflow negative process fixture',
      'bash',
      [releaseScript, repositoryRoot],
      1
    );

    const nodeTestPath = path.join(
      fixtureRoot,
      'release-portfolio-migrated.test.mjs'
    );
    const migratedAssertion = migrateReleasePortfolioTests(
      `  assert.equal((backend.match(/- '\\.github\\/workflows\\/frontend-ci\\.yml'/g) ?? []).length, 2)`
    );
    fs.writeFileSync(
      nodeTestPath,
      `import test from 'node:test'
import assert from 'node:assert/strict'

function verify(backend) {
${migratedAssertion}
}

test('positive migrated Node assertion', () => {
  verify(\`name: Backend CI

on:
  pull_request:

  push:
    paths:
      - '.github/workflows/frontend-ci.yml'

permissions:
  contents: read
\`)
})
`);
    expectProcessStatus(
      'release Node assertion positive process fixture',
      'node',
      ['--test', nodeTestPath],
      0
    );

    fs.writeFileSync(
      nodeTestPath,
      `import test from 'node:test'
import assert from 'node:assert/strict'

function verify(backend) {
${migratedAssertion}
}

test('negative migrated Node assertion', () => {
  verify(\`name: Backend CI

on:
  pull_request:
    paths:
      - '.github/workflows/frontend-ci.yml'

  push:
    paths:
      - '.github/workflows/frontend-ci.yml'

permissions:
  contents: read
\`)
})
`);
    expectProcessStatus(
      'release Node assertion negative process fixture',
      'node',
      ['--test', nodeTestPath],
      1
    );

    const gitFixture = path.join(fixtureRoot, 'git-fixture');
    fs.mkdirSync(gitFixture, { recursive: true });
    run('git', ['init'], { cwd: gitFixture });
    run('git', ['config', 'user.name', 'Installer Fixture'], { cwd: gitFixture });
    run('git', ['config', 'user.email', 'installer@example.invalid'], { cwd: gitFixture });
    fs.writeFileSync(path.join(gitFixture, 'tracked.txt'), 'baseline\n');
    run('git', ['add', 'tracked.txt'], { cwd: gitFixture });
    run('git', ['commit', '-m', 'fixture baseline'], { cwd: gitFixture });
    const fixtureBaseline = git(['rev-parse', 'HEAD'], gitFixture);

    expectFailure(
      'rollback transaction negative fixture',
      () => applyWithRollback(
        gitFixture,
        fixtureBaseline,
        () => {
          fs.writeFileSync(path.join(gitFixture, 'tracked.txt'), 'changed\n');
          fs.writeFileSync(path.join(gitFixture, 'untracked.txt'), 'new\n');
        },
        () => {
          throw new Error('deliberate validation failure');
        }
      ),
      /deliberate validation failure/
    );
    if (
      fs.readFileSync(
        path.join(gitFixture, 'tracked.txt'),
        'utf8'
      ) !== 'baseline\n' ||
      fs.existsSync(path.join(gitFixture, 'untracked.txt')) ||
      git(['status', '--porcelain'], gitFixture) !== ''
    ) {
      throw new Error('rollback transaction fixture did not restore baseline');
    }

    applyWithRollback(
      gitFixture,
      fixtureBaseline,
      () => {
        fs.writeFileSync(path.join(gitFixture, 'tracked.txt'), 'accepted\n');
      },
      () => {
        if (
          fs.readFileSync(
            path.join(gitFixture, 'tracked.txt'),
            'utf8'
          ) !== 'accepted\n'
        ) {
          throw new Error('positive transaction validation failed');
        }
      }
    );
    if (
      fs.readFileSync(
        path.join(gitFixture, 'tracked.txt'),
        'utf8'
      ) !== 'accepted\n'
    ) {
      throw new Error('positive transaction fixture failed');
    }
    rollbackRepository(gitFixture, fixtureBaseline);

    const patchFixture = path.join(fixtureRoot, 'patch-fixture');
    fs.mkdirSync(patchFixture, { recursive: true });
    run('git', ['init'], { cwd: patchFixture });
    run('git', ['config', 'user.name', 'Installer Fixture'], { cwd: patchFixture });
    run('git', ['config', 'user.email', 'installer@example.invalid'], { cwd: patchFixture });

    write('.github/workflows/backend-ci.yml', `name: Backend CI

on:
  pull_request:
    paths:
      - 'README.md'
      - '.github/workflows/frontend-ci.yml'
      - 'render.yaml'

  push:
    branches:
      - main
    paths:
      - 'README.md'
      - '.github/workflows/frontend-ci.yml'
      - 'render.yaml'

permissions:
  contents: read

jobs:
  backend-quality:
    name: Backend Quality
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7

  backend-container:
    name: Backend Container
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-node@v7

  backend-race:
    name: Backend Race Safety
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7

  postgres-integration:
    name: PostgreSQL 16 Integration
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: actions/setup-go@v7

# STAGE-14-1-ARCHITECTURE-CONSOLIDATION-V1-1
`, patchFixture);

    write('.github/workflows/frontend-ci.yml', `name: Frontend CI

on:
  pull_request:
    paths:
      - 'apps/web/**'
      - 'docs/171_DEPENDABOT_FOLLOW_UP_RECONCILIATION.md'
      - 'scripts/verify-frontend-dependency-security.mjs'
      - 'scripts/verify-frontend-dependency-security.test.mjs'

  push:
    branches:
      - main
    paths:
      - 'apps/web/**'
      - 'docs/171_DEPENDABOT_FOLLOW_UP_RECONCILIATION.md'
      - 'scripts/verify-frontend-dependency-security.mjs'
      - 'scripts/verify-frontend-dependency-security.test.mjs'

permissions:
  contents: read

jobs:
  frontend-quality:
    name: Frontend Quality
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
      - uses: pnpm/action-setup@v6
      - uses: actions/setup-node@v7

# STAGE-14-1-ARCHITECTURE-CONSOLIDATION-V1-1
`, patchFixture);

    write('package.json', `${JSON.stringify({
      name: 'fixture',
      private: true,
      scripts: {
        'verify:release': 'bash scripts/verify-release.sh',
      },
    }, null, 2)}\n`, patchFixture);

    write('scripts/verify-release.sh', `#!/usr/bin/env bash
set -euo pipefail
pnpm run test:dependency-maintenance
pnpm run verify:dependency-maintenance
printf '%s\\n' 'RELEASE_RECRUITER_QUICKSTART=PASS'
`, patchFixture);

    write('scripts/verify-dependency-maintenance.mjs', `const backendCI = ''
const frontendCI = ''
if (backendCI.includes('actions/setup-go@v6') || !backendCI.includes('actions/setup-go@v7')) fail('setup-go v7 contract failed');
if (frontendCI.includes('actions/setup-node@v6') || !frontendCI.includes('actions/setup-node@v7')) fail('setup-node v7 contract failed');
`, patchFixture);

    write('scripts/verify-dependency-maintenance.test.mjs', `import test from 'node:test'
import assert from 'node:assert/strict'

test('CI setup actions use version 7 only', () => {
  const backend = ''
  const frontend = ''
  assert.match(backend, /actions\\/setup-go@v7/);
  assert.doesNotMatch(backend, /actions\\/setup-go@v6/);
  assert.match(frontend, /actions\\/setup-node@v7/);
  assert.doesNotMatch(frontend, /actions\\/setup-node@v6/);
});
`, patchFixture);

    write(
      'scripts/verify-recruiter-quickstart.sh',
      `#!/usr/bin/env bash
set -euo pipefail
WORKFLOW_FILE="$1"
fail() { exit 1; }
require_literal() { grep -F -- "$2" "$1" >/dev/null || fail "$3"; }
require_literal "$WORKFLOW_FILE" "- 'README.md'" 'missing'
readme_path_count="$(
  grep -F -c -- "- 'README.md'" "$WORKFLOW_FILE"
)"
if [ "$readme_path_count" -ne 2 ]; then
  fail 'Backend CI must trigger on README.md for both push and pull_request'
fi
`,
      patchFixture
    );

    write(
      'scripts/verify-release-portfolio.sh',
      `#!/usr/bin/env bash
set -euo pipefail
REPOSITORY_ROOT="$1"
fail() { exit 1; }
backend_frontend_workflow_count="$(
  grep -F -c -- "- '.github/workflows/frontend-ci.yml'" \\
    "$REPOSITORY_ROOT/.github/workflows/backend-ci.yml"
)"
[ "$backend_frontend_workflow_count" -eq 2 ] || \\
  fail 'Backend CI must track frontend workflow changes for push and pull_request'
`,
      patchFixture
    );

    write('scripts/verify-release-portfolio.test.mjs', `import test from 'node:test'
import assert from 'node:assert/strict'

test('backend and frontend CI both enforce the release contract', () => {
  const backend = ''
  assert.equal((backend.match(/- '\\.github\\/workflows\\/frontend-ci\\.yml'/g) ?? []).length, 2)
})
`, patchFixture);

    write('scripts/verify-backend-operations.test.mjs', `import test from 'node:test'
import assert from 'node:assert/strict'
test('Backend CI permanently enforces operations contracts and Blueprint changes', () => {
  const source = ''
  assert.equal((source.match(/- 'render\\.yaml'/g) ?? []).length, 2)
})
`, patchFixture);

    write('docs/DOCUMENT_INDEX.md', '# Document Index\n', patchFixture);
    write('apps/api/main.go', 'package main\n\nfunc main() {}\n', patchFixture);

    run('git', ['add', '.'], { cwd: patchFixture });
    run('git', ['commit', '-m', 'fixture baseline'], { cwd: patchFixture });

    patchRepository(patchFixture, fake);
    verifyChangedManifest(patchFixture);
    run(
      'node',
      ['--check', 'scripts/verify-repository-governance.mjs'],
      { cwd: patchFixture }
    );
    run(
      'node',
      ['--check', 'scripts/verify-repository-governance.test.mjs'],
      { cwd: patchFixture }
    );
    run(
      'node',
      ['--test', 'scripts/verify-repository-governance.test.mjs'],
      { cwd: patchFixture }
    );
    run(
      'node',
      ['scripts/verify-repository-governance.mjs'],
      { cwd: patchFixture }
    );
    run(
      'bash',
      ['-n', 'scripts/configure-repository-governance.sh'],
      { cwd: patchFixture }
    );
    run(
      'bash',
      ['-n', 'scripts/verify-repository-governance-settings.sh'],
      { cwd: patchFixture }
    );
    expectFailure(
      'second patch application negative fixture',
      () => patchRepository(patchFixture, fake),
      /expected|missing|unexpected|found 0/
    );

    const manifestFixture = path.join(fixtureRoot, 'manifest-fixture');
    fs.mkdirSync(manifestFixture, { recursive: true });
    run('git', ['init'], { cwd: manifestFixture });
    run('git', ['config', 'user.name', 'Installer Fixture'], { cwd: manifestFixture });
    run('git', ['config', 'user.email', 'installer@example.invalid'], { cwd: manifestFixture });
    fs.writeFileSync(path.join(manifestFixture, 'base.txt'), 'base\n');
    run('git', ['add', 'base.txt'], { cwd: manifestFixture });
    run('git', ['commit', '-m', 'fixture baseline'], { cwd: manifestFixture });
    fs.writeFileSync(path.join(manifestFixture, 'new.txt'), 'new\n');
    verifyChangedManifest(manifestFixture, ['new.txt']);
    expectFailure(
      'manifest mismatch negative fixture',
      () => verifyChangedManifest(manifestFixture, ['other.txt']),
      /manifest mismatch/
    );
  } finally {
    fs.rmSync(fixtureRoot, { recursive: true, force: true });
  }

  const ownSource = fs.readFileSync(new URL(import.meta.url), 'utf8');
  const forbiddenCommands = [
    "['worktree', '" + "remove'",
    "['" + "switch'",
    "['" + "restore'",
    '--show-' + 'current',
  ];
  for (const forbidden of forbiddenCommands) {
    if (ownSource.includes(forbidden)) {
      throw new Error(`Git 2.15 forbidden command found: ${forbidden}`);
    }
  }

  console.log('INSTALLER_POSITIVE_FIXTURES=PASS');
  console.log('INSTALLER_NEGATIVE_FIXTURES=PASS');
  console.log('INSTALLER_ROLLBACK_SELF_TEST=PASS');
  console.log('FULL_PATCH_GENERATION_FIXTURE=PASS');
  console.log('GOFMT_BEFORE_AND_AFTER_SELF_TEST=PASS');
  console.log('MANIFEST_SELF_TEST=PASS');
  console.log('SEMANTIC_WORKFLOW_CONTRACT_MIGRATION_SELF_TEST=PASS');
  console.log('BACKEND_OPERATIONS_CONTRACT_MIGRATION_SELF_TEST=PASS');
  console.log('GIT_2_15_COMPATIBILITY_SELF_TEST=PASS');
  console.log('INSTALLER_SELF_TEST=PASS');
}

if (process.argv.includes('--self-test')) {
  selfTest();
  process.exit(0);
}
if (process.argv.includes('--validate-baseline-only')) {
  try { validateExactBaselineOnly(); }
  catch (error) { fail(error instanceof Error ? error.message.replace(/\s+/g, ' ') : String(error)); }
  process.exit(0);
}

try {
  selfTest();
  if (!fs.existsSync(PROJECT_ROOT)) fail(`project root not found: ${PROJECT_ROOT}`);
  if (git(['symbolic-ref', '--quiet', '--short', 'HEAD']) !== 'main') fail('current branch is not main');
  if (git(['status', '--porcelain']) !== '') fail('working tree is not clean');
  run('git', ['fetch', '--prune', 'origin'], { cwd: PROJECT_ROOT });
  if (git(['rev-parse', 'HEAD']) !== BASELINE) fail('local HEAD does not match baseline');
  if (git(['rev-parse', 'origin/main']) !== BASELINE) fail('origin/main does not match baseline');
  run('gh', ['auth', 'status', '--hostname', 'github.com'], { cwd: PROJECT_ROOT });
  cleanupInstallerWorktrees();

  const refs = {
    checkout: resolveActionRef('actions/checkout', 'v7'),
    setupGo: resolveActionRef('actions/setup-go', 'v7'),
    setupNode: resolveActionRef('actions/setup-node', 'v7'),
    pnpmSetup: resolveActionRef('pnpm/action-setup', 'v6'),
    codeql: resolveActionRef('github/codeql-action', 'v4'),
  };
  console.log(`ACTION_CHECKOUT_SHA=${refs.checkout}`);
  console.log(`ACTION_SETUP_GO_SHA=${refs.setupGo}`);
  console.log(`ACTION_SETUP_NODE_SHA=${refs.setupNode}`);
  console.log(`ACTION_PNPM_SETUP_SHA=${refs.pnpmSetup}`);
  console.log(`ACTION_CODEQL_SHA=${refs.codeql}`);

  const realGoDigest = verifyGoFormattingBaseline(
    PROJECT_ROOT,
    'REAL_GO_FORMATTING_BEFORE_PATCH'
  );

  const worktree = fs.mkdtempSync(
    path.join(os.tmpdir(), 'gfa-governance-')
  );
  let isolatedError = null;
  let cleanupError = null;

  try {
    run(
      'git',
      ['worktree', 'add', '--detach', worktree, BASELINE],
      { cwd: PROJECT_ROOT }
    );

    const isolatedGoDigest = verifyGoFormattingBaseline(
      worktree,
      'ISOLATED_GO_FORMATTING_BEFORE_PATCH'
    );

    patchRepository(worktree, refs);
    verifyChangedManifest(worktree);
    console.log('ISOLATED_CHANGED_FILE_MANIFEST=PASS');

    validate(
      worktree,
      isolatedGoDigest,
      'ISOLATED'
    );
    console.log('ISOLATED_REPOSITORY_VALIDATION=PASS');
  } catch (error) {
    isolatedError = error;
  } finally {
    fs.rmSync(worktree, { recursive: true, force: true });
    const prune = spawnSync(
      'git',
      ['worktree', 'prune', '--expire', 'now'],
      {
        cwd: PROJECT_ROOT,
        encoding: 'utf8',
        stdio: 'pipe',
      }
    );
    if (prune.status !== 0) {
      process.stderr.write(prune.stdout ?? '');
      process.stderr.write(prune.stderr ?? '');
      cleanupError = new Error(
        'Git 2.15 detached worktree cleanup failed'
      );
    } else {
      console.log('ISOLATED_WORKTREE_CLEANUP=PASS');
    }
  }

  if (isolatedError) throw isolatedError;
  if (cleanupError) throw cleanupError;

  applyWithRollback(
    PROJECT_ROOT,
    BASELINE,
    () => {
      patchRepository(PROJECT_ROOT, refs);
    },
    () => {
      verifyChangedManifest(PROJECT_ROOT);
      console.log('CHANGED_FILE_MANIFEST=PASS');

      validate(
        PROJECT_ROOT,
        realGoDigest,
        'REAL'
      );
    }
  );

  console.log('REAL_REPOSITORY_VALIDATION=PASS');
  console.log('REPOSITORY_GOVERNANCE_PATCH_INSTALL=PASS');
  console.log('REPOSITORY_GOVERNANCE_STAGE_READY_FOR_REVIEW=PASS');
  console.log('GITHUB_SETTINGS_NOT_CHANGED=PASS');
  console.log('REMOTE_BRANCHES_NOT_CHANGED=PASS');
  console.log('COMMIT_AND_PUSH_NOT_PERFORMED=PASS');
} catch (error) {
  fail(error instanceof Error ? error.message.replace(/\s+/g, ' ') : String(error));
}
