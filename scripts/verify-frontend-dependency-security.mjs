import { readFile } from "node:fs/promises";
import { pathToFileURL } from "node:url";

export const MINIMUM_SAFE_NANOID_VERSION = "3.3.18";
export const PINNED_NANOID_VERSION = "3.3.18";
export const MINIMUM_SAFE_POSTCSS_VERSION = "8.5.23";
export const PINNED_POSTCSS_VERSION = "8.5.23";
export const MINIMUM_SAFE_SHARP_VERSION = "0.35.0";
export const PINNED_SHARP_VERSION = "0.35.3";
export const MINIMUM_SAFE_NEXT_VERSION = "16.2.12";
export const PINNED_NEXT_VERSION = "16.2.12";
export const NEXT_VERSION = PINNED_NEXT_VERSION;

export function parseVersion(value) {
  const match = /^(\d+)\.(\d+)\.(\d+)$/.exec(value);
  if (!match) {
    throw new Error(`Invalid semantic version: ${value}`);
  }

  return match.slice(1).map((part) => Number.parseInt(part, 10));
}

export function compareVersions(left, right) {
  const leftParts = parseVersion(left);
  const rightParts = parseVersion(right);

  for (let index = 0; index < 3; index += 1) {
    if (leftParts[index] < rightParts[index]) {
      return -1;
    }
    if (leftParts[index] > rightParts[index]) {
      return 1;
    }
  }

  return 0;
}

function collectPackageVersions(lockfileText, packageName) {
  const versions = new Set();
  const escapedPackageName = packageName.replaceAll("/", "\\/");
  const pattern = new RegExp(
    `^\\s{2}['\"]?${escapedPackageName}@(\\d+\\.\\d+\\.\\d+)['\"]?:\\s*$`,
    "gm",
  );

  for (const match of lockfileText.matchAll(pattern)) {
    versions.add(match[1]);
  }

  return [...versions].sort(compareVersions);
}

export function collectNanoidVersions(lockfileText) {
  return collectPackageVersions(lockfileText, "nanoid");
}

export function collectPostCSSVersions(lockfileText) {
  return collectPackageVersions(lockfileText, "postcss");
}

export function collectSharpVersions(lockfileText) {
  return collectPackageVersions(lockfileText, "sharp");
}

function nextDependencyUsesVersion(lockfileText, dependencyName, version) {
  const escapedNextVersion = NEXT_VERSION.replaceAll(".", "\\.");
  const blockPattern = new RegExp(
    `^  next@${escapedNextVersion}[^\\n]*:\\n([\\s\\S]*?)(?=^  \\S|\\Z)`,
    "gm",
  );

  for (const match of lockfileText.matchAll(blockPattern)) {
    if (
      new RegExp(
        `^\\s{6}${dependencyName}:\\s+${version.replaceAll(".", "\\.")}\\s*$`,
        "m",
      ).test(match[1])
    ) {
      return true;
    }
  }

  return false;
}

export function nextUsesPinnedPostCSS(lockfileText) {
  return nextDependencyUsesVersion(
    lockfileText,
    "postcss",
    PINNED_POSTCSS_VERSION,
  );
}

export function postCSSUsesPinnedNanoid(lockfileText) {
  const escapedPostCSSVersion = PINNED_POSTCSS_VERSION.replaceAll(".", "\\.");
  const blockPattern = new RegExp(
    `^  postcss@${escapedPostCSSVersion}[^\\n]*:\\n([\\s\\S]*?)(?=^  \\S|\\Z)`,
    "gm",
  );

  for (const match of lockfileText.matchAll(blockPattern)) {
    if (
      new RegExp(
        `^\\s{6}nanoid:\\s+${PINNED_NANOID_VERSION.replaceAll(".", "\\.")}\\s*$`,
        "m",
      ).test(match[1])
    ) {
      return true;
    }
  }

  return false;
}

function importerDependencyUsesVersion(
  lockfileText,
  importerName,
  dependencyName,
  version,
) {
  const lines = lockfileText.split("\n");
  const importerHeader = `  ${importerName}:`;
  const dependencyHeader = `      ${dependencyName}:`;
  const versionLine = `        version: ${version}`;

  const importerIndex = lines.findIndex((line) => line === importerHeader);
  if (importerIndex < 0) {
    return false;
  }

  let importerEnd = lines.length;
  for (let index = importerIndex + 1; index < lines.length; index += 1) {
    const line = lines[index];
    if (/^[^\s]/.test(line) || /^  \S/.test(line)) {
      importerEnd = index;
      break;
    }
  }

  for (let index = importerIndex + 1; index < importerEnd; index += 1) {
    if (lines[index] !== dependencyHeader) {
      continue;
    }

    for (let nestedIndex = index + 1; nestedIndex < importerEnd; nestedIndex += 1) {
      const nestedLine = lines[nestedIndex];
      if (/^\s{6}\S/.test(nestedLine) || /^\s{4}\S/.test(nestedLine)) {
        break;
      }
      if (
        nestedLine === versionLine ||
        nestedLine.startsWith(`${versionLine}(`)
      ) {
        return true;
      }
    }
  }

  return false;
}

export function webImporterUsesPinnedSharp(lockfileText) {
  return importerDependencyUsesVersion(
    lockfileText,
    "apps/web",
    "sharp",
    PINNED_SHARP_VERSION,
  );
}

export function webImporterUsesPinnedNext(lockfileText) {
  return importerDependencyUsesVersion(
    lockfileText,
    "apps/web",
    "next",
    PINNED_NEXT_VERSION,
  );
}

export function webImporterUsesPinnedESLintConfigNext(lockfileText) {
  return importerDependencyUsesVersion(
    lockfileText,
    "apps/web",
    "eslint-config-next",
    PINNED_NEXT_VERSION,
  );
}

function workspaceHasOverride(workspaceText, packageName, minimum, pinned) {
  return new RegExp(
    `^\\s{2}['\"]?${packageName}@<${minimum.replaceAll(".", "\\.")}['\"]?:\\s+['\"]?${pinned.replaceAll(".", "\\.")}['\"]?\\s*$`,
    "m",
  ).test(workspaceText);
}

export function workspaceHasNanoidOverride(workspaceText) {
  return workspaceHasOverride(
    workspaceText,
    "nanoid",
    MINIMUM_SAFE_NANOID_VERSION,
    PINNED_NANOID_VERSION,
  );
}

export function workspaceHasTargetedOverride(workspaceText) {
  return workspaceHasOverride(
    workspaceText,
    "postcss",
    MINIMUM_SAFE_POSTCSS_VERSION,
    PINNED_POSTCSS_VERSION,
  );
}

export function workspaceHasSharpOverride(workspaceText) {
  return workspaceHasOverride(
    workspaceText,
    "sharp",
    MINIMUM_SAFE_SHARP_VERSION,
    PINNED_SHARP_VERSION,
  );
}

export function webPinsSharp(webPackageText) {
  const packageJSON = JSON.parse(webPackageText);
  return packageJSON.dependencies?.sharp === PINNED_SHARP_VERSION;
}

export function webPinsNextSecurityRelease(webPackageText) {
  const packageJSON = JSON.parse(webPackageText);
  return (
    packageJSON.dependencies?.next === PINNED_NEXT_VERSION &&
    packageJSON.devDependencies?.["eslint-config-next"] ===
      PINNED_NEXT_VERSION
  );
}

function validateResolvedVersions({
  failures,
  packageName,
  versions,
  minimumSafeVersion,
  pinnedVersion,
}) {
  if (versions.length === 0) {
    failures.push(
      `pnpm-lock.yaml contains no ${packageName} package resolution`,
    );
    return;
  }

  const vulnerableVersions = versions.filter(
    (version) => compareVersions(version, minimumSafeVersion) < 0,
  );
  if (vulnerableVersions.length > 0) {
    failures.push(
      `pnpm-lock.yaml still contains vulnerable ${packageName} versions: ${vulnerableVersions.join(", ")}`,
    );
  }

  if (!versions.includes(pinnedVersion)) {
    failures.push(
      `pnpm-lock.yaml does not contain pinned ${packageName} ${pinnedVersion}`,
    );
  }
}

export function verifyFrontendBuildDeterminism({
  layoutText,
  globalsText,
}) {
  const failures = [];
  const forbiddenLayoutTokens = [
    "next/font/google",
    "fonts.gstatic.com",
    "geistSans.variable",
    "geistMono.variable",
  ];

  for (const token of forbiddenLayoutTokens) {
    if (layoutText.includes(token)) {
      failures.push(
        `apps/web/app/layout.tsx must not depend on remote build-time font token ${token}`,
      );
    }
  }

  if (!layoutText.includes("className='h-full antialiased'")) {
    failures.push(
      "apps/web/app/layout.tsx must use deterministic local font classes",
    );
  }

  const requiredGlobalFontStacks = [
    "--font-sans: Arial, Helvetica, sans-serif;",
    "--font-mono: 'SFMono-Regular', Consolas, 'Liberation Mono', monospace;",
  ];
  for (const stack of requiredGlobalFontStacks) {
    if (!globalsText.includes(stack)) {
      failures.push(
        `apps/web/app/globals.css is missing deterministic system font stack ${stack}`,
      );
    }
  }
  if (globalsText.includes("var(--font-geist")) {
    failures.push(
      "apps/web/app/globals.css must not reference remote Geist font variables",
    );
  }

  if (failures.length > 0) {
    throw new Error(failures.join("\n"));
  }

  return { fontSource: "system-local" };
}

export function verifyDependencySecurity({
  lockfileText,
  workspaceText,
  webPackageText,
  layoutText,
  globalsText,
}) {
  const failures = [];

  try {
    verifyFrontendBuildDeterminism({ layoutText, globalsText });
  } catch (error) {
    failures.push(error.message);
  }

  if (!workspaceHasNanoidOverride(workspaceText)) {
    failures.push(
      `pnpm-workspace.yaml must override nanoid@<${MINIMUM_SAFE_NANOID_VERSION} to ${PINNED_NANOID_VERSION}`,
    );
  }

  if (!workspaceHasTargetedOverride(workspaceText)) {
    failures.push(
      `pnpm-workspace.yaml must override postcss@<${MINIMUM_SAFE_POSTCSS_VERSION} to ${PINNED_POSTCSS_VERSION}`,
    );
  }

  if (!workspaceHasSharpOverride(workspaceText)) {
    failures.push(
      `pnpm-workspace.yaml must override sharp@<${MINIMUM_SAFE_SHARP_VERSION} to ${PINNED_SHARP_VERSION}`,
    );
  }

  if (!webPinsSharp(webPackageText)) {
    failures.push(
      `apps/web/package.json must pin sharp ${PINNED_SHARP_VERSION}`,
    );
  }

  if (!webPinsNextSecurityRelease(webPackageText)) {
    failures.push(
      `apps/web/package.json must pin next and eslint-config-next ${PINNED_NEXT_VERSION}`,
    );
  }

  if (!webImporterUsesPinnedNext(lockfileText)) {
    failures.push(
      `apps/web importer does not resolve next ${PINNED_NEXT_VERSION}`,
    );
  }

  if (!webImporterUsesPinnedESLintConfigNext(lockfileText)) {
    failures.push(
      `apps/web importer does not resolve eslint-config-next ${PINNED_NEXT_VERSION}`,
    );
  }

  const nanoidVersions = collectNanoidVersions(lockfileText);
  validateResolvedVersions({
    failures,
    packageName: "nanoid",
    versions: nanoidVersions,
    minimumSafeVersion: MINIMUM_SAFE_NANOID_VERSION,
    pinnedVersion: PINNED_NANOID_VERSION,
  });

  const postcssVersions = collectPostCSSVersions(lockfileText);
  validateResolvedVersions({
    failures,
    packageName: "PostCSS",
    versions: postcssVersions,
    minimumSafeVersion: MINIMUM_SAFE_POSTCSS_VERSION,
    pinnedVersion: PINNED_POSTCSS_VERSION,
  });

  const sharpVersions = collectSharpVersions(lockfileText);
  validateResolvedVersions({
    failures,
    packageName: "sharp",
    versions: sharpVersions,
    minimumSafeVersion: MINIMUM_SAFE_SHARP_VERSION,
    pinnedVersion: PINNED_SHARP_VERSION,
  });

  if (!nextUsesPinnedPostCSS(lockfileText)) {
    failures.push(
      `Next.js ${NEXT_VERSION} does not resolve PostCSS ${PINNED_POSTCSS_VERSION}`,
    );
  }

  if (!postCSSUsesPinnedNanoid(lockfileText)) {
    failures.push(
      `PostCSS ${PINNED_POSTCSS_VERSION} does not resolve nanoid ${PINNED_NANOID_VERSION}`,
    );
  }

  if (!webImporterUsesPinnedSharp(lockfileText)) {
    failures.push(
      `apps/web importer does not resolve sharp ${PINNED_SHARP_VERSION}`,
    );
  }

  if (failures.length > 0) {
    throw new Error(failures.join("\n"));
  }

  return {
    nextVersion: NEXT_VERSION,
    minimumSafeNextVersion: MINIMUM_SAFE_NEXT_VERSION,
    pinnedNextVersion: PINNED_NEXT_VERSION,
    nanoidVersions,
    postcssVersions,
    sharpVersions,
    minimumSafeNanoidVersion: MINIMUM_SAFE_NANOID_VERSION,
    pinnedNanoidVersion: PINNED_NANOID_VERSION,
    minimumSafePostCSSVersion: MINIMUM_SAFE_POSTCSS_VERSION,
    minimumSafeSharpVersion: MINIMUM_SAFE_SHARP_VERSION,
    pinnedPostCSSVersion: PINNED_POSTCSS_VERSION,
    pinnedSharpVersion: PINNED_SHARP_VERSION,
    fontSource: "system-local",
  };
}

async function main() {
  const [workspaceText, lockfileText, webPackageText, layoutText, globalsText] = await Promise.all([
    readFile("pnpm-workspace.yaml", "utf8"),
    readFile("pnpm-lock.yaml", "utf8"),
    readFile("apps/web/package.json", "utf8"),
    readFile("apps/web/app/layout.tsx", "utf8"),
    readFile("apps/web/app/globals.css", "utf8"),
  ]);

  const result = verifyDependencySecurity({
    lockfileText,
    workspaceText,
    webPackageText,
    layoutText,
    globalsText,
  });

  console.log(
    `FRONTEND_DEPENDENCY_SECURITY=PASS next=${result.nextVersion} minimum_safe_next=${result.minimumSafeNextVersion} pinned_next=${result.pinnedNextVersion} nanoid=${result.nanoidVersions.join(",")} postcss=${result.postcssVersions.join(",")} sharp=${result.sharpVersions.join(",")} minimum_safe_nanoid=${result.minimumSafeNanoidVersion} minimum_safe_postcss=${result.minimumSafePostCSSVersion} minimum_safe_sharp=${result.minimumSafeSharpVersion} fonts=${result.fontSource}`,
  );
}

const isMain =
  process.argv[1] !== undefined &&
  import.meta.url === pathToFileURL(process.argv[1]).href;

if (isMain) {
  main().catch((error) => {
    console.error(`FRONTEND_DEPENDENCY_SECURITY=FAIL\n${error.message}`);
    process.exitCode = 1;
  });
}
