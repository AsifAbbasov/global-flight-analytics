# Document 54 — Stage 14.14 Composite Historical Pagination Cursor

Status: Implementation Baseline v2.0
Project: Global Flight Analytics
Scope: lossless store and HTTP keyset pagination for Historical Intelligence

## 1. Problem

Historical aggregate results are ordered by:

```text
window_end DESC
window_start DESC
as_of_time DESC
id ASC
```

The previous store and HTTP contracts carried only one timestamp boundary.

When more records shared the same `window_end` than fit on one page, records
after the visible boundary could be skipped permanently.

## 2. Store Contract

The pagination cursor now contains every term of the stable ordering:

```text
WindowEnd
WindowStart
AsOfTime
ID
```

A page contains:

```text
Records
HasMore
NextCursor
```

`NextCursor` is generated from the last record returned to the caller. It is
present only when the sentinel query proves that another record exists.

## 3. PostgreSQL Keyset Predicate

For the ordering:

```text
window_end DESC
window_start DESC
as_of_time DESC
id ASC
```

the next page begins where:

```text
window_end < cursor.window_end
OR
window_end = cursor.window_end
AND window_start < cursor.window_start
OR
window_end = cursor.window_end
AND window_start = cursor.window_start
AND as_of_time < cursor.as_of_time
OR
window_end = cursor.window_end
AND window_start = cursor.window_start
AND as_of_time = cursor.as_of_time
AND id > cursor.id
```

The identifier comparison is ascending because `id` is the final ascending
ordering field.

## 4. HTTP Cursor

The history endpoint now accepts:

```text
cursor=<opaque versioned token>
```

The response exposes:

```json
{
  "has_more": true,
  "next_cursor": "<opaque versioned token>"
}
```

The token contains all four store cursor fields and is encoded with unpadded
URL-safe Base64 over strict versioned JSON.

The HTTP token is intentionally opaque. Clients must return it unchanged and
must not construct or edit its internal values.

The decoder rejects:

```text
invalid Base64
invalid JSON
unknown JSON fields
trailing JSON values
unsupported cursor versions
missing cursor fields
invalid time ordering
oversized tokens
```

## 5. Removed Legacy Contract

The following names are removed from production Go and TypeScript code:

```text
before_window_end
next_before_window_end
BeforeWindowEnd
NextBeforeWindowEnd
```

This prevents store, HTTP, runtime verification, and future frontend clients
from accidentally restoring single-field pagination.

## 6. Cursor Validation

A cursor is valid only when:

```text
WindowEnd is present
WindowStart is present
AsOfTime is present
ID is non-empty
WindowStart is before WindowEnd
AsOfTime is not before WindowEnd
ID length is bounded
```

Times are normalized to UTC and identifiers are trimmed.

Partially populated cursors fail with `ErrInvalidListCursor`.

## 7. Preserved Behavior

This increment does not change:

```text
result ordering
default or maximum list limits
sentinel pagination
scope filtering
metric filtering
granularity filtering
stored result format
database schema
migrations
materialization formulas
analytical result payloads
frontend visualization behavior
```

Only pagination transport and boundary correctness change.

## 8. Recovery from Failed v1 Installation

The v2 installer supports the known partially installed Stage 14.14 v1 state.

It verifies that:

```text
HEAD remains 1f30bae
all dirty paths belong to the failed Stage 14.14 attempt
no unrelated user changes are present
```

It then restores only the known Stage 14.14 tracked paths, removes only the
known Stage 14.14 untracked paths, verifies a clean tree, and applies v2.

It does not use a repository-wide destructive clean against unverified work.

## 9. Regression Gates

Automated tests require:

```text
the store cursor to contain all four ordering values
the PostgreSQL predicate to use all ordering values
the next cursor to match the last visible record
partial store cursors to be rejected
HTTP cursor encode/decode round trips
malformed HTTP tokens to be rejected
DTO responses to emit next_cursor only when required
handlers to accept cursor and pass the complete decoded boundary
runtime HTTP verification to use next_cursor
legacy cursor names to be absent from production Go and TypeScript
page cloning not to share cursor state
```

## 10. Acceptance

The increment is accepted only after:

```text
recovery verification
Historical Aggregate Contract tests
Historical Aggregate Store tests
HTTP cursor codec tests
Historical Intelligence DTO tests
Historical Intelligence handler tests
runtime verifier tests
race detector
strict project architecture audit
complete Go build
go vet
complete Go test suite
frontend dependency security verification
frontend production dependency audit
ESLint
TypeScript validation
Next.js production build
backend Docker image build
git diff check
```
