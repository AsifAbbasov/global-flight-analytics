# Document 99 — Provenance and Analytical Trust Hardening

Status: Implemented Engineering Increment v1.0
Project: Global Flight Analytics
Baseline: `0ae85ccbff7584a993030c0adcdee3290dd4b7bd`

## 1. Purpose

This increment closes the accepted Analytical Core findings concerning
incomplete source provenance, placeholder source attribution, public failure
detail leakage, and overstated confidence for request-parameter snapshot
metrics.

## 2. Strict source provenance

Every published analytical source must now contain:

```text
a non-placeholder source name;
a known source role;
a complete observation window when an observation window is present;
a non-zero retrieval timestamp;
a retrieval timestamp that does not follow result calculation time.
```

The reserved placeholder name `unknown` is rejected by the analytical source
contract.

Trajectory metadata no longer converts a blank provider name into `unknown`.
Unattributed observations are omitted from the source list and disclosed by the
`unattributed_source_observations` limitation.

All trajectory and request-parameter sources receive a server-owned UTC
`RetrievedAt` timestamp.

## 3. Safe failure publication

The default metric failure path preserves the typed machine code and retry
classification but replaces the operation error text with the stable public
message:

```text
Analytical metric calculation failed.
```

An explicitly supplied custom FailureMapper remains responsible for its own
public contract.

## 4. Honest confidence for request snapshots

Coverage Score and Data Freshness remain deterministic formulas, but their
current public endpoints accept snapshot values supplied by the caller.

Their confidence factors now distinguish:

```text
high formula stability;
low independently verified source coverage.
```

This produces low confidence rather than high confidence for direct
request-parameter calculations. The existing request-parameter limitation
continues to make endpoint results limited rather than complete.

## 5. Verification

The installer performs all changes first in a detached temporary Git worktree
and runs:

```text
complete backend compilation;
targeted tests;
the complete backend test suite.
```

Only after that preflight passes is the working `main` branch changed. The
working tree then runs targeted compilation, targeted tests, race tests, the
complete backend test suite, Go vet, architecture audits, static contract
checks, documentation checks, and whitespace validation.

## 6. Remaining Analytical Core review scope

```text
replace public request-parameter snapshots with server-owned production snapshots;
reject zero analytical reference time;
canonicalize accepted UUID identifiers;
classify and consolidate obsolete analytical foundation packages;
unify metric identifiers across the analytical stack.
```
