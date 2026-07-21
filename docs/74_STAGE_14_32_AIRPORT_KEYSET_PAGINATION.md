# Stage 14.32 — Airport Keyset Pagination and Canonical Row Scanning

## Status

Completed as an independently verified Stage 14 increment. Stage 14 remains reopened for the remaining explicitly recorded maintainability and query-contract backlog.

## Problem

The PostgreSQL Airport Repository previously exposed one unbounded `List` query ordered only by airport name. The query loaded the entire catalog in one database result set, had no stable tie-breaker for duplicate names, and duplicated the same row-scanning sequence in `List` and `GetByICAO`.

## Implemented contract

The airport domain now defines:

- `ListRequest` with an explicit bounded limit and optional cursor;
- `ListCursor` containing the complete `(name, id)` ordering key;
- `ListPage` with items and an optional next cursor;
- default and maximum page sizes;
- deterministic request normalization and typed validation errors;
- a separate `PagedRepository` contract while preserving the existing `Repository` interface.

## PostgreSQL keyset pagination

`AirportRepository.ListPage` uses keyset pagination with this total order:

```text
airport name ascending
airport identifier ascending
```

The continuation predicate is equivalent to:

```text
name > cursor.name
OR (name = cursor.name AND id > cursor.id)
```

No `OFFSET` pagination is permitted. Each query requests `limit + 1` rows, returns at most `limit`, and emits a next cursor only when the lookahead row proves that more data exists.

## Backward compatibility

The existing `List(context.Context)` method remains available. It no longer owns an unbounded SQL query. It acts as a compatibility adapter that repeatedly invokes bounded `ListPage` reads using the maximum approved page size and combines the pages for callers that still require the complete catalog.

## Canonical row scanner

`ListPage` and `GetByICAO` now share one `scanAirportRecord` owner. Nullable elevation evidence is converted in exactly one place, preserving observed zero elevation and unknown elevation semantics.

## Verification

Permanent verification includes:

- domain normalization tests for default limits, maximum limits, cursor cloning, and invalid cursor shape;
- page-builder tests proving that the next cursor uses the last returned row rather than the lookahead row;
- source-contract tests forbidding `OFFSET`, the old unbounded ordering query, and duplicated scanning ownership;
- PostgreSQL integration tests with duplicate airport names, multiple pages, stable ordering, no duplicates, and no omissions;
- Stage 14 source-audit ownership rules;
- the unified Stage 14 backend, PostgreSQL, frontend, vulnerability, and container audit.

## Non-goals

This increment does not change the public HTTP API, frontend airport browsing, airport ranking algorithms, import semantics, or database schema. Those areas remain separate contracts.

## Acceptance marker

```text
STAGE_14_32_AIRPORT_PAGINATION=PASS
```
