import assert from "node:assert/strict";
import test from "node:test";

import {
  collectPostCSSVersions,
  collectSharpVersions,
  compareVersions,
  nextUsesPinnedPostCSS,
  webImporterUsesPinnedSharp,
  verifyDependencySecurity,
  verifyFrontendBuildDeterminism,
  webPinsSharp,
  workspaceHasSharpOverride,
  workspaceHasTargetedOverride,
} from "./verify-frontend-dependency-security.mjs";

const secureWorkspace = `packages:
  - 'apps/*'
overrides:
  'postcss@<8.5.10': 8.5.15
  'sharp@<0.35.0': 0.35.3
`;

const secureWebPackage = JSON.stringify({
  dependencies: {
    next: "16.2.9",
    sharp: "0.35.3",
  },
});


const secureLayout = `import type { Metadata } from 'next'

export default function RootLayout() {
  return <html lang='en' className='h-full antialiased' />
}
`;

const secureGlobals = `@theme inline {
  --font-sans: Arial, Helvetica, sans-serif;
  --font-mono: 'SFMono-Regular', Consolas, 'Liberation Mono', monospace;
}
`;

const secureLockfile = `lockfileVersion: '9.0'

importers:

  apps/web:
    dependencies:
      next:
        specifier: 16.2.9
        version: 16.2.9(react@19.2.4)
      sharp:
        specifier: 0.35.3
        version: 0.35.3

packages:
  postcss@8.5.15:
    resolution: {integrity: sha512-postcss}

  sharp@0.35.3:
    resolution: {integrity: sha512-sharp}

snapshots:
  next@16.2.9(react@19.2.4):
    dependencies:
      postcss: 8.5.15
  postcss@8.5.15: {}

  sharp@0.35.3: {}
`;

function verify({
  lockfileText = secureLockfile,
  workspaceText = secureWorkspace,
  webPackageText = secureWebPackage,
  layoutText = secureLayout,
  globalsText = secureGlobals,
} = {}) {
  return verifyDependencySecurity({
    lockfileText,
    workspaceText,
    webPackageText,
    layoutText,
    globalsText,
  });
}

test("semantic versions are compared numerically", () => {
  assert.equal(compareVersions("0.34.9", "0.35.0"), -1);
  assert.equal(compareVersions("0.35.0", "0.35.0"), 0);
  assert.equal(compareVersions("0.35.3", "0.35.0"), 1);
});

test("PostCSS resolutions are collected deterministically", () => {
  assert.deepEqual(
    collectPostCSSVersions(
      `${secureLockfile}\n  postcss@8.5.12:\n    resolution: {integrity: sha512-second}\n`,
    ),
    ["8.5.12", "8.5.15"],
  );
});

test("sharp resolutions are collected deterministically", () => {
  assert.deepEqual(
    collectSharpVersions(
      `${secureLockfile}\n  sharp@0.35.1:\n    resolution: {integrity: sha512-second}\n`,
    ),
    ["0.35.1", "0.35.3"],
  );
});

test("targeted workspace overrides are recognized", () => {
  assert.equal(workspaceHasTargetedOverride(secureWorkspace), true);
  assert.equal(workspaceHasSharpOverride(secureWorkspace), true);
});

test("web application pins sharp directly", () => {
  assert.equal(webPinsSharp(secureWebPackage), true);
});

test("Next.js PostCSS and web Sharp resolutions are recognized", () => {
  assert.equal(nextUsesPinnedPostCSS(secureLockfile), true);
  assert.equal(webImporterUsesPinnedSharp(secureLockfile), true);
});

test("secure dependency graph passes", () => {
  const result = verify();
  assert.deepEqual(result.postcssVersions, ["8.5.15"]);
  assert.deepEqual(result.sharpVersions, ["0.35.3"]);
});

test("vulnerable PostCSS resolution fails", () => {
  const vulnerableLockfile = secureLockfile.replaceAll(
    "8.5.15",
    "8.4.31",
  );

  assert.throws(
    () => verify({ lockfileText: vulnerableLockfile }),
    /vulnerable PostCSS versions: 8\.4\.31/,
  );
});

test("vulnerable sharp resolution fails", () => {
  const vulnerableLockfile = secureLockfile.replaceAll(
    "0.35.3",
    "0.34.5",
  );

  assert.throws(
    () => verify({ lockfileText: vulnerableLockfile }),
    /vulnerable sharp versions: 0\.34\.5/,
  );
});

test("missing sharp override fails", () => {
  const workspaceText = secureWorkspace.replace(
    "  'sharp@<0.35.0': 0.35.3\n",
    "",
  );

  assert.throws(
    () => verify({ workspaceText }),
    /must override sharp/,
  );
});

test("missing direct sharp pin fails", () => {
  assert.throws(
    () =>
      verify({
        webPackageText: JSON.stringify({
          dependencies: { next: "16.2.9" },
        }),
      }),
    /must pin sharp 0\.35\.3/,
  );
});

test("missing web importer sharp resolution fails", () => {
  const lockfileText = secureLockfile.replace(
    "      sharp:\n        specifier: 0.35.3\n        version: 0.35.3\n",
    "",
  );

  assert.throws(
    () => verify({ lockfileText }),
    /apps\/web importer does not resolve sharp 0\.35\.3/,
  );
});

test("remote Google font import fails", () => {
  const layoutText = secureLayout.replace(
    "import type { Metadata } from 'next'",
    "import type { Metadata } from 'next'\nimport { Geist } from 'next/font/google'",
  );

  assert.throws(
    () => verify({ layoutText }),
    /must not depend on remote build-time font token next\/font\/google/,
  );
});

test("missing deterministic system font stack fails", () => {
  const globalsText = secureGlobals.replace(
    "--font-mono: 'SFMono-Regular', Consolas, 'Liberation Mono', monospace;",
    "--font-mono: var(--font-geist-mono);",
  );

  assert.throws(
    () => verify({ globalsText }),
    /missing deterministic system font stack/,
  );
});

test("frontend build determinism contract passes", () => {
  assert.deepEqual(
    verifyFrontendBuildDeterminism({
      layoutText: secureLayout,
      globalsText: secureGlobals,
    }),
    { fontSource: "system-local" },
  );
});
