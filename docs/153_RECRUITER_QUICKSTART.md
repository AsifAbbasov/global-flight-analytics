# Recruiter Quickstart Contract

Status: Implementation prepared; exact-commit Continuous Integration closure pending  
Project: Global Flight Analytics  
Reviewed baseline: `fa222650ef5425dcf60a05fb949c17cafc36c1a8`  
Date: 2026-08-01

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
