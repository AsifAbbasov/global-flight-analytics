#!/usr/bin/env python3
from __future__ import annotations

import argparse
import base64
import hashlib
import json
import subprocess
import zlib
from pathlib import Path

SCRIPT_DIR = Path(__file__).resolve().parent
PART_NAMES = ['payload.00.b64', 'payload.01.b64', 'payload.02.b64', 'payload.03.b64', 'payload.04.b64']
EXPECTED_ENCODED_SHA256 = 'bc778a1b6f29f734fd7250b0cb99f614d83f241879f61aa502dae600ddc67359'
EXPECTED_CANONICAL_SHA256 = '2924879280d1105a6474a0deb0a7926c03cbd27cc4fd07e213dd9fc91a50eb7a'
PAYLOAD_BYTES = b"".join((SCRIPT_DIR / name).read_bytes() for name in PART_NAMES)
if hashlib.sha256(PAYLOAD_BYTES).hexdigest() != EXPECTED_ENCODED_SHA256:
    raise SystemExit("OBSERVABILITY_PATCH_ERROR=encoded payload checksum mismatch")
CANONICAL_BYTES = zlib.decompress(base64.b64decode(PAYLOAD_BYTES, validate=True))
if hashlib.sha256(CANONICAL_BYTES).hexdigest() != EXPECTED_CANONICAL_SHA256:
    raise SystemExit("OBSERVABILITY_PATCH_ERROR=canonical payload checksum mismatch")
PAYLOAD = json.loads(CANONICAL_BYTES.decode("utf-8"))


def fail(message: str) -> None:
    raise SystemExit(f"OBSERVABILITY_PATCH_ERROR={message}")


def run(*args: str) -> str:
    completed = subprocess.run(
        args,
        check=True,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    return completed.stdout.strip()


def repository_root() -> Path:
    current = Path.cwd().resolve()
    for candidate in (current, *current.parents):
        if (candidate / "apps/api/go.mod").is_file() and (candidate / ".git").exists():
            return candidate
    fail("repository root not found")
    raise AssertionError


def replace_once(path: Path, old: str, new: str) -> None:
    content = path.read_text()
    count = content.count(old)
    if count != 1:
        fail(f"replacement count for {path} is {count}, expected 1")
    path.write_text(content.replace(old, new, 1))


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--allow-nonbaseline", action="store_true")
    args = parser.parse_args()

    root = repository_root()
    if not args.allow_nonbaseline:
        head = run("git", "-C", str(root), "rev-parse", "HEAD")
        if head != PAYLOAD["baseline"]:
            fail(f"baseline {head} does not match {PAYLOAD['baseline']}")

    for relative_path, content in PAYLOAD["new_files"].items():
        destination = root / relative_path
        if destination.exists():
            fail(f"new file already exists: {relative_path}")
        destination.parent.mkdir(parents=True, exist_ok=True)
        destination.write_text(content)

    for replacement in PAYLOAD["replacements"]:
        replace_once(
            root / replacement["path"],
            replacement["old"],
            replacement["new"],
        )

    for relative_path, suffix in PAYLOAD["appends"].items():
        destination = root / relative_path
        content = destination.read_text().rstrip("\n") + suffix
        destination.write_text(content.rstrip("\n") + "\n")

    subprocess.run(
        ["python", str(SCRIPT_DIR / "apply_observability_fixes.py")],
        check=True,
    )
    subprocess.run(["gofmt", "-w", str(root / "apps/api")], check=True)
    subprocess.run(["git", "-C", str(root), "diff", "--check"], check=True)
    print("OBSERVABILITY_PATCH_APPLICATION=PASS")


if __name__ == "__main__":
    main()
