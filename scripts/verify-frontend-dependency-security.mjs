import { readFile } from "node:fs/promises";
import { pathToFileURL } from "node:url";

export const MINIMUM_SAFE_POSTCSS_VERSION = "8.5.10";
export const PINNED_POSTCSS_VERSION = "8.5.15";
export const NEXT_VERSION = "16.2.9";

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

export function collectPostCSSVersions(lockfileText) {
  const versions = new Set();
  const pattern = /^\s{2}postcss@(\d+\.\d+\.\d+):\s*$/gm;

  for (const match of lockfileText.matchAll(pattern)) {
    versions.add(match[1]);
  }

  return [...versions].sort(compareVersions);
}

export function nextUsesPinnedPostCSS(lockfileText) {
  const escapedNextVersion = NEXT_VERSION.replaceAll(".", "\\.");
  const blockPattern = new RegExp(
    `^  next@${escapedNextVersion}[^\\n]*:\\n([\\s\\S]*?)(?=^  \\S|\\Z)`,
    "gm",
  );

  for (const match of lockfileText.matchAll(blockPattern)) {
    if (
      new RegExp(
        `^\\s{6}postcss:\\s+${PINNED_POSTCSS_VERSION}\\s*$`,
        "m",
      ).test(match[1])
    ) {
      return true;
    }
  }

  return false;
}

export function workspaceHasTargetedOverride(workspaceText) {
  return new RegExp(
    `^\\s{2}['"]?postcss@<${MINIMUM_SAFE_POSTCSS_VERSION.replaceAll(".", "\\.")}['"]?:\\s+['"]?${PINNED_POSTCSS_VERSION.replaceAll(".", "\\.")}['"]?\\s*$`,
    "m",
  ).test(workspaceText);
}

export function verifyDependencySecurity({
  lockfileText,
  workspaceText,
}) {
  const failures = [];

  if (!workspaceHasTargetedOverride(workspaceText)) {
    failures.push(
      `pnpm-workspace.yaml must override postcss@<${MINIMUM_SAFE_POSTCSS_VERSION} to ${PINNED_POSTCSS_VERSION}`,
    );
  }

  const versions = collectPostCSSVersions(lockfileText);
  if (versions.length === 0) {
    failures.push("pnpm-lock.yaml contains no PostCSS package resolution");
  }

  const vulnerableVersions = versions.filter(
    (version) =>
      compareVersions(version, MINIMUM_SAFE_POSTCSS_VERSION) < 0,
  );
  if (vulnerableVersions.length > 0) {
    failures.push(
      `pnpm-lock.yaml still contains vulnerable PostCSS versions: ${vulnerableVersions.join(", ")}`,
    );
  }

  if (!versions.includes(PINNED_POSTCSS_VERSION)) {
    failures.push(
      `pnpm-lock.yaml does not contain pinned PostCSS ${PINNED_POSTCSS_VERSION}`,
    );
  }

  if (!nextUsesPinnedPostCSS(lockfileText)) {
    failures.push(
      `Next.js ${NEXT_VERSION} does not resolve PostCSS ${PINNED_POSTCSS_VERSION}`,
    );
  }

  if (failures.length > 0) {
    throw new Error(failures.join("\n"));
  }

  return {
    nextVersion: NEXT_VERSION,
    postcssVersions: versions,
    minimumSafeVersion: MINIMUM_SAFE_POSTCSS_VERSION,
    pinnedVersion: PINNED_POSTCSS_VERSION,
  };
}

async function main() {
  const [workspaceText, lockfileText] = await Promise.all([
    readFile("pnpm-workspace.yaml", "utf8"),
    readFile("pnpm-lock.yaml", "utf8"),
  ]);

  const result = verifyDependencySecurity({
    lockfileText,
    workspaceText,
  });

  console.log(
    `FRONTEND_DEPENDENCY_SECURITY=PASS next=${result.nextVersion} postcss=${result.postcssVersions.join(",")} minimum_safe=${result.minimumSafeVersion}`,
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
