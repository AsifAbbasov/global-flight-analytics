# Frontend Application Shell

Status: CLOSED — feature implementation and Continuous Integration evidence reconciled
Project: Global Flight Analytics
Reviewed baseline: `9e9a10e93fecec21d07e395df486a4f76d48c9db`
Implementation commit: `cd99532aa40fe7a61fb19580455ac2d1ba5f650c`
Frontend CI: `30707691429` — SUCCESS
Backend CI: `30707691414` — SUCCESS
Date: 2026-08-01
Canonical reconciliation: 2026-08-25

## 1. Purpose

This increment turns the existing analytical workspace into a coherent product entry
surface without changing backend contracts, analytical semantics or live query
ownership.

## 2. Product structure

The page provides:

- a sticky application header and primary anchor navigation;
- a product hero that explains the platform value and non-operational scope;
- an honest startup-status strip based only on the server-rendered initial snapshot;
- stable overview, live-workspace and research-scope anchors;
- a visible architecture boundary and research-only disclaimer;
- a structured research-scope section and product footer;
- deterministic dark-mode global styling, focus visibility and reduced-motion support.

## 3. Status semantics

The shell deliberately does not claim to be a continuously updated online indicator.
It describes only the initial server-rendered API snapshot. Client-owned React Query
refresh, errors and retry behavior remain inside the live traffic workspace.

The startup state is classified as:

- `ready` when the initial traffic and region requests completed;
- `degraded` when the world snapshot exists but the region catalog failed;
- `unavailable` when the initial traffic snapshot failed.

A valid empty traffic response remains `ready`; absence of aircraft is not treated as
transport failure.

## 4. Verification

The increment adds five dependency-free tests covering ready, empty, degraded,
unavailable and invalid-count normalization semantics. Existing frontend tests,
ESLint, TypeScript validation, dependency policy and production build gates remain
required.

The exact implementation commit is:

```text
cd99532aa40fe7a61fb19580455ac2d1ba5f650c
feat: add frontend application shell
```

GitHub Actions evidence for that exact commit is:

```text
Frontend CI 30707691429 = SUCCESS
Backend CI  30707691414 = SUCCESS
```

Frontend Quality on run `30707691429` successfully completed dependency policy,
production dependency audit, ESLint, TypeScript validation, frontend contract tests and
the production frontend build. The old document wording that exact-commit CI closure
was still pending is therefore historical drift rather than an open present-day state.

## 5. Scope boundary

This increment did not add authentication, user accounts, persistent preferences,
new backend endpoints, remote fonts, telemetry vendors, a design-system dependency or
a lockfile change.

It also did not claim that the initial server-rendered status represented later client
refresh health. That separation is intentional and prevents a product shell from
fabricating live availability semantics it does not own.

## 6. Canonical classification

This document is a **feature/product implementation record**, not a remediation
finding record.

The implementation establishes a product shell, truthful initial-snapshot semantics,
accessibility-oriented styling and tests. Although it improves the earlier presentation,
repository evidence does not establish a separate defect lifecycle requiring a
synthetic finding ID.

```text
Canonical finding ID: none by design
Classification: feature / product-shell implementation and closure evidence
Historical implementation: CLOSED
Exact-commit Frontend CI: CLOSED
Open remediation findings owned by this document: 0
```

## 7. Residual boundaries and prevention

Later frontend product hardening and visual redesigns may supersede the exact shell
layout. The historical closure here means the increment was implemented and CI-verified;
it does not freeze the August 1 visual design as the final UI.

Future shell changes should preserve the core evidence rule introduced here: initial
server-rendered status must not be mislabeled as continuously current client health.
Actual failures of that contract, accessibility or navigation may be registered as new
findings when evidence supports them; ordinary visual evolution should remain feature
history rather than finding inflation.
