import assert from "node:assert/strict";
import test from "node:test";

import {
  collectPostCSSVersions,
  compareVersions,
  nextUsesPinnedPostCSS,
  verifyDependencySecurity,
  workspaceHasTargetedOverride,
} from "./verify-frontend-dependency-security.mjs";

const secureWorkspace = `packages:
  - 'apps/*'
overrides:
  'postcss@<8.5.10': 8.5.15
`;

const secureLockfile = `lockfileVersion: '9.0'

packages:
  postcss@8.5.15:
    resolution: {integrity: sha512-example}

snapshots:
  next@16.2.9(react@19.2.4):
    dependencies:
      postcss: 8.5.15

  postcss@8.5.15: {}
`;

test("semantic versions are compared numerically", () => {
  assert.equal(compareVersions("8.5.9", "8.5.10"), -1);
  assert.equal(compareVersions("8.5.10", "8.5.10"), 0);
  assert.equal(compareVersions("8.6.0", "8.5.10"), 1);
});

test("PostCSS resolutions are collected deterministically", () => {
  assert.deepEqual(
    collectPostCSSVersions(
      `${secureLockfile}\n  postcss@8.5.12:\n    resolution: {integrity: sha512-second}\n`,
    ),
    ["8.5.12", "8.5.15"],
  );
});

test("targeted workspace override is recognized", () => {
  assert.equal(workspaceHasTargetedOverride(secureWorkspace), true);
});

test("Next.js pinned PostCSS resolution is recognized", () => {
  assert.equal(nextUsesPinnedPostCSS(secureLockfile), true);
});

test("secure dependency graph passes", () => {
  const result = verifyDependencySecurity({
    lockfileText: secureLockfile,
    workspaceText: secureWorkspace,
  });

  assert.deepEqual(result.postcssVersions, ["8.5.15"]);
});

test("vulnerable PostCSS resolution fails", () => {
  const vulnerableLockfile = secureLockfile.replaceAll(
    "8.5.15",
    "8.4.31",
  );

  assert.throws(
    () =>
      verifyDependencySecurity({
        lockfileText: vulnerableLockfile,
        workspaceText: secureWorkspace,
      }),
    /vulnerable PostCSS versions: 8\.4\.31/,
  );
});

test("missing override fails", () => {
  assert.throws(
    () =>
      verifyDependencySecurity({
        lockfileText: secureLockfile,
        workspaceText: "packages:\n  - 'apps/*'\n",
      }),
    /must override postcss/,
  );
});
