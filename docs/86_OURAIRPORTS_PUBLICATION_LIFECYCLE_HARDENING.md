# OurAirports Publication Lifecycle Hardening

Status: Implemented Engineering Contract v1.0
Project: Global Flight Analytics

## Purpose

This document records the durable publication lifecycle used by the production
OurAirports import command.

The lifecycle prevents a publication from being marked as processed before its
database import succeeds. It also prevents two import processes from reconciling
the same publication concurrently.

## Publication identity

Every downloaded OurAirports CSV response receives a deterministic identifier:

```text
sha256:<hexadecimal content digest>
```

The identifier is based on the complete bounded response body. HTTP validators
remain responsible for conditional retrieval, while the content digest is the
canonical import publication identity.

## Durable lifecycle

The PostgreSQL table `provider_publications` records one row for each provider
and publication identifier.

The supported lifecycle is:

```text
ReservePublication
→ execute idempotent airport reconciliation
→ CommitPublication
```

When reconciliation fails or is cancelled before commit:

```text
ReleasePublication
→ the same publication may be retried
```

A reservation contains a unique ownership token. Only the owner may commit or
release it.

## Lease and crash recovery

Reservations use a thirty-minute lease in the production command.

A second process receives an in-progress decision while the lease is active.
After the lease expires, another process may atomically reclaim the publication.
This allows recovery after process termination without permitting concurrent
active owners.

## Validator ordering

The source HTTP validator is persisted only after one of these outcomes:

1. the publication import completed and the publication was committed;
2. PostgreSQL already contained a committed record for the same publication.

The validator is not persisted after an import failure. Therefore a failed
publication cannot be hidden behind a later HTTP `304 Not Modified` response.

If validator persistence fails after publication commit, the next command run
may download the content again, observe that the publication is already
committed, skip database reconciliation, and repair the validator record.

## Concurrency and idempotency

The database reservation uses a transaction and row-level locking.

The contract guarantees:

- one active owner per provider and publication identifier;
- retry after explicit release;
- retry after lease expiry;
- idempotent commit by the same reservation token;
- rejection of stale or foreign reservation tokens;
- no duplicate airport reconciliation after a committed publication;
- safe retry when reconciliation completed but publication commit was not
  recorded, because airport reconciliation is implemented as an idempotent
  upsert.

## Verification

The implementation includes:

- process-local publication lifecycle tests;
- orchestration release, commit, duplicate, and coalescing tests;
- deterministic OurAirports publication identity tests;
- production import workflow tests;
- optional PostgreSQL integration tests for reserve, release, commit, lease
  reclaim, and concurrent reservation;
- race detector coverage;
- full backend tests;
- `go vet`;
- migration and existing policy gates.

## Scope boundary

The lifecycle is publication-driven and currently used by OurAirports.
It does not convert live traffic providers into publication sources and does
not replace their fixed-window or provider-reported budget policies.
