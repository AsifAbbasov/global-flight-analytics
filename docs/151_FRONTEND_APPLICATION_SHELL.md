# Frontend Application Shell

Status: Implementation prepared; exact-commit Continuous Integration closure pending
Project: Global Flight Analytics
Reviewed baseline: `9e9a10e93fecec21d07e395df486a4f76d48c9db`
Date: 2026-08-01

## 1. Purpose

This increment turns the existing analytical workspace into a coherent product entry
surface without changing backend contracts, analytical semantics or live query
ownership.

## 2. Product structure

The page now provides:

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

## 5. Scope boundary

This increment does not add authentication, user accounts, persistent preferences,
new backend endpoints, remote fonts, telemetry vendors, a design-system dependency or
a lockfile change.
