# Document 54 — Stage 14.14 Composite Historical Pagination Cursor

Status: Remediation History v2.1
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

## 11. Canonical remediation history

### Finding / symptom

Historical Intelligence exposed a pagination cursor that represented only `window_end` while the actual stable result order used four fields. Pages could therefore lose records when multiple results shared the visible timestamp boundary.

### Root cause

The transport/store cursor contract did not model the complete PostgreSQL ordering tuple. Pagination semantics were designed around a convenient timestamp rather than the full deterministic sort key.

### Failure scenario

```text
more than page-size records share window_end
↓
first page ends inside that equal-window_end group
↓
client receives only before_window_end
↓
next predicate jumps to rows with an earlier window_end
↓
remaining same-window_end records are never returned
```

### Impact

The API could silently omit valid historical analytical results across page boundaries. Because the response still looked well-formed and pagination continued, the loss could remain unnoticed by clients and downstream analysis.

### Severity rationale

Historical severity was not explicitly recorded. Retrospective classification: **P1 data retrieval correctness** because valid persisted records could become permanently unreachable through the public paginated contract without an error.

### Existing guarantees violated

```text
pagination must be lossless across every stable-order tie
cursor semantics must match the complete database ordering
next_cursor must start strictly after the last visible record
client transport must not rely on constructing internal ordering state
```

### Considered solutions

1. Keep a single timestamp cursor and increase page size.
2. Use offset pagination.
3. Add ad-hoc secondary timestamp parameters.
4. Encode the complete ordering tuple as one opaque, versioned cursor.

### Chosen remediation

Option 4: use `(WindowEnd, WindowStart, AsOfTime, ID)` throughout the store cursor and keyset predicate, expose a versioned opaque HTTP token, and generate the next cursor from the last visible record only when a sentinel row proves more data exists.

### Why this solution was selected

It directly mirrors the canonical ordering, remains deterministic under ties, avoids offset drift, and hides storage-order details from public clients while retaining an evolvable token version.

### Rejected alternatives

Increasing limits was rejected because it only moves the failure threshold. Offset pagination was rejected because concurrent changes and large offsets make it less stable and potentially more expensive. Multiple public cursor query parameters were rejected because they expose internal ordering details and make partial/malformed cursor states easier to construct.

### Trade-offs

```text
+ pagination becomes lossless under complete ordering ties
+ clients handle one opaque token
+ keyset semantics remain index-friendly and deterministic
- cursor tokens are more complex to encode/debug manually
- ordering changes require cursor-version compatibility decisions
- legacy single-field clients must migrate
```

### Regression tests / protection

Protection covers the complete four-field predicate, next-cursor derivation, partial cursor rejection, strict codec decoding, malformed/oversized token rejection, runtime pagination verification, absence of legacy cursor names, and ownership-safe page cloning.

### Adversarial review findings

Review exposed two non-obvious requirements: the final `id` term must use ascending comparison because the canonical order is mixed-direction, and a cursor must be generated from the last **visible** record rather than the sentinel record. The failed v1 installer also required bounded recovery that touched only known Stage 14.14 paths.

Historical PR/reviewer evidence for the original July remediation is unavailable; reconstruction is limited to repository source, tests, commits, and this stage document.

### Remediation iterations

A first Stage 14.14 installation attempt failed and left a known partial state based on HEAD `1f30bae...`. The v2 remediation added guarded recovery and landed the final lossless implementation in commit `6a78070499ec0cbe9f905fa94d4b0995d41f2a40` (`fix: make historical pagination lossless`).

### Residual risks and limitations

The cursor is valid only for the ordering/version contract it encodes. A future ordering change must either preserve version-1 semantics for in-flight tokens or introduce a new cursor version. Pagination correctness does not by itself guarantee completeness of the underlying materialized historical dataset.

### Operational or deployment consequences

No schema migration is required. API consumers must treat cursor tokens as opaque and return them unchanged. Deployments that change cursor order/version must include backward-compatibility or explicit invalidation policy.

### Exact evidence

```text
implementation commit:
6a78070499ec0cbe9f905fa94d4b0995d41f2a40

permanent evidence:
Historical Aggregate Contract tests
PostgreSQL keyset/store tests
HTTP cursor codec and handler tests
runtime pagination verifier
backend final correctness audit
```

### Final canonical status

```text
FINDING_GFA_DATA_052_HISTORICAL_PAGINATION_LOSS=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/54_STAGE_14_14_COMPOSITE_HISTORICAL_PAGINATION_CURSOR.md
IMPLEMENTATION_COMMIT=6a78070499ec0cbe9f905fa94d4b0995d41f2a40
```

### Prevention / future guard

Any paginated query must derive its cursor from the complete stable ordering tuple. Architecture/review checks must reject a cursor that omits a sort term or reverses a mixed-direction comparison. Public clients must never be required to synthesize cursor internals.
