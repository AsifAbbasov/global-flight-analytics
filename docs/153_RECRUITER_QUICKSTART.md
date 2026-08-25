# Recruiter Quickstart Contract

Status: CLOSED — operability implementation and Continuous Integration evidence reconciled  
Project: Global Flight Analytics  
Reviewed baseline: `fa222650ef5425dcf60a05fb949c17cafc36c1a8`  
Implementation commit: `c34264a122913272e3083a4f64397b0e8470c4f3`  
Backend CI: `30709750384` — SUCCESS  
Date: 2026-08-01
Canonical reconciliation: 2026-08-25

## 1. Purpose

This increment turns the repository entry point into a reproducible reviewer workflow.
It does not add a second Continuous Integration file or a second Docker Compose file.
The quickstart is built on the existing hardened `compose.yaml`, migration service,
healthcheck binary, lifecycle endpoints and pnpm workspace.

## 2. Reviewer path

The top-level README now provides one bounded flow:

```text
validate Compose
→ start PostgreSQL
→ complete migrations
→ start the API
→ verify health, readiness and build information
→ install the frozen pnpm workspace
→ start the Next.js frontend
```

The commands use the real endpoint paths and the real root package scripts.

## 3. Local mutation protection

Database-backed server configuration requires `API_MUTATION_KEY_SHA256`. Compose now
provides an overridable local-only digest so the read-oriented demo starts without a
manual secret-preparation step. No corresponding raw key is distributed. Mutation
routes therefore remain unavailable by default. Developers who need those routes must
generate their own raw key and override the digest explicitly.

This default is local development infrastructure and is not production secret
management.

## 4. Permanent verification

`scripts/verify-recruiter-quickstart.sh` checks that:

1. README uses the real Compose commands;
2. health, readiness and version endpoints remain exact;
3. frontend installation remains frozen and pnpm-based;
4. Compose preserves migration-before-API ordering;
5. the local mutation digest remains overridable and structurally valid;
6. Backend Continuous Integration runs this contract;
7. README changes trigger Backend Continuous Integration.

The existing container job still executes `docker compose config`, builds the hardened
image and performs the real PostgreSQL-backed smoke test.

## 5. Scope boundary

This increment does not add production deployment, publish credentials, expose a raw
mutation key, replace the existing Compose topology, add frontend containers, change
runtime ports or weaken any existing Continuous Integration gate.

## 6. Historical closure evidence

The exact implementation owner is:

```text
c34264a122913272e3083a4f64397b0e8470c4f3
docs: add verified recruiter quickstart
```

The implementation adds the README reviewer path, overridable local-only mutation-key
digest, permanent quickstart verifier and Backend CI reachability for README changes.

GitHub Actions evidence for that exact commit is:

```text
Backend CI 30709750384 = SUCCESS
```

The run completed Backend Quality, Backend Race Safety, PostgreSQL 16 Integration and
Backend Container successfully. Backend Container explicitly executed `Verify recruiter
quickstart contract`, validated Docker Compose, built the backend image, verified the
non-root runtime user and completed the PostgreSQL-backed container health smoke test.

The original `exact-commit Continuous Integration closure pending` status is therefore
historical drift and is closed by recovered repository evidence.

## 7. Canonical classification

This document is **operability / reviewer-workflow implementation and closure evidence**,
not a remediation finding record.

The increment made the repository easier and safer to evaluate reproducibly. Source
evidence does not establish a separate pre-existing engineering defect with a distinct
root-cause/remediation lifecycle that should receive a synthetic `GFA-*` finding ID.

```text
Canonical finding ID: none by design
Classification: operability / reviewer-workflow implementation and closure evidence
Historical implementation: CLOSED
Exact-commit Backend CI: CLOSED
Open remediation findings owned by this document: 0
```

## 8. Residual boundaries and prevention

This historical closure does not turn the local Compose digest into production secret
management and does not claim public-deployment availability. Those remain governed by
later deployment and release evidence.

Regression ownership remains with `scripts/verify-recruiter-quickstart.sh`, Backend CI,
Compose validation and the container smoke test. Future quickstart failures should be
registered as findings only when source-backed correctness or operability evidence
establishes a real defect rather than ordinary documentation evolution.
