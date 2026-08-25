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

## Canonical finding decomposition

This increment closes two separate findings:

```text
GFA-PERF-019   unbounded Airport catalog listing and incomplete pagination order
GFA-MAINT-020  duplicated Airport row-scanning ownership
```

## GFA-PERF-019 — Unbounded Airport catalog listing and incomplete pagination order

### Finding / symptom

The repository exposed a full-catalog `List` SQL query with no bounded page contract. Ordering used airport name alone, so duplicate names did not have a complete deterministic ordering key.

### Root cause

The original repository contract was designed for a small catalog and returned `[]airport.Airport` directly. As the data layer matured, the same convenience API became both a scalability boundary and an ordering ambiguity.

### Failure scenario

As the airport catalog grows, one call loads the entire result set and forces memory, network, and query work proportional to total catalog size. If multiple airports share the same name, a later attempt to paginate by name alone can repeat or skip rows because the ordering is not total.

### Impact

The unbounded query creates avoidable database/client work and prevents a reliable continuation contract. The incomplete order can produce unstable page boundaries in a paginated implementation.

### Severity rationale

**P2 retrospective.** This is a scalability and deterministic-query contract issue with plausible correctness effects once pagination is introduced. It is not evidence of existing persisted-data corruption.

### Existing guarantees violated

- list operations over growing catalogs should be bounded;
- cursor pagination requires a stable total ordering;
- continuation must not duplicate or omit rows.

### Considered solutions

1. keep unbounded `List` indefinitely;
2. use `OFFSET/LIMIT` pagination;
3. introduce keyset pagination with a complete `(name, id)` cursor while retaining `List` as a compatibility adapter.

### Chosen remediation and why

`ListPage` uses bounded keyset pagination over `(name, id)`. The full-catalog `List` remains for compatibility but internally walks bounded pages. This migrates ownership without forcing an immediate public HTTP/API rewrite.

### Rejected alternatives

Unbounded listing was rejected because cost grows with total rows. `OFFSET` was rejected because large offsets become progressively expensive and concurrent inserts can shift page boundaries. Name-only cursors were rejected because airport names are not unique.

### Trade-offs

The domain contract gains request/page/cursor types and compatibility `List` may issue multiple queries. In return, every database read is bounded and page traversal is deterministic.

### Regression tests

Domain normalization tests, source rules forbidding `OFFSET`, duplicate-name integration cases, multi-page no-duplicate/no-omission tests, and lookahead cursor tests protect the contract.

### Adversarial review and remediation iterations

The implementation deliberately retains legacy `List` only as a bounded-page adapter. This addresses the adversarial scenario where a new paged method exists but old callers still trigger the original unbounded SQL path.

### Residual risk / limitations

The compatibility adapter can still materialize the entire catalog in application memory for legacy callers. The database work is bounded per page, but callers that truly need streaming or UI pagination should migrate to `ListPage` rather than rely on full aggregation.

### Operational / deployment consequences

No schema migration is required. Query patterns change from one potentially large result set to bounded keyset reads.

### Exact evidence

Implementation commit: `06f87ba4adfa3202cfe4f68232712a97e6812630` (`feat: add airport keyset pagination`). Historical PR/reviewer metadata is not reconstructed without repository evidence.

### Final canonical status

**CLOSED.** The production repository no longer owns an unbounded airport-list SQL query and the page order is total.

## GFA-MAINT-020 — Duplicated Airport row-scanning ownership

### Finding / symptom

`List` and `GetByICAO` duplicated the database-to-domain scan sequence for Airport rows.

### Root cause

Read methods evolved independently before nullable elevation semantics and pagination increased the cost of keeping mappings synchronized.

### Failure scenario

A future column or semantic change can be added to one scanner and omitted from the other, producing path-dependent domain mapping for the same database row.

### Impact

The immediate issue is maintainability and semantic drift risk. It is especially relevant after the elevation-semantics fix because nullable elevation must be mapped identically in every Airport read path.

### Severity rationale

**P3 retrospective.** No historical behavior mismatch is asserted; the finding is duplicated mapping ownership that makes future drift easier.

### Existing guarantees violated

One database record shape should have one canonical scanner/mapping owner.

### Considered solutions

1. keep duplicate scanners and add comments/tests;
2. generate row scanners;
3. extract one package-local `scanAirportRecord` shared by all Airport reads.

### Chosen remediation and why

A single scanner owns Airport row mapping. This is the smallest solution and preserves repository interfaces.

### Rejected alternatives

Comments do not remove duplication. Scanner generation was unnecessary for one stable row shape and would introduce tooling complexity without proportional value.

### Trade-offs

Read methods become dependent on one shared helper, which is desirable because the row mapping is intentionally one contract.

### Regression tests

Source-contract tests require canonical scanner ownership and protect nullable elevation mapping.

### Adversarial review and remediation iterations

The scanner consolidation was delivered together with pagination specifically to avoid creating a new `ListPage` scanner while leaving the older duplicate mappings in place.

### Residual risk / limitations

A shared scanner can still grow if the Airport row shape expands substantially; future decomposition should be driven by actual mapping complexity rather than file-count goals.

### Operational / deployment consequences

None.

### Exact evidence

Implementation commit: `06f87ba4adfa3202cfe4f68232712a97e6812630`.

### Final canonical status

**CLOSED.**

## Prevention / future guard

Future list-style repository contracts must define bounded result size and a complete deterministic order before introducing cursors. Database-to-domain mapping for one row shape should have one canonical owner unless a deliberate alternative representation is documented.
