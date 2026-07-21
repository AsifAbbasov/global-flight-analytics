# Document 76 — Stage 14.34 PostgreSQL Contract Consolidation

Status: Implemented current-scope baseline
Project: Global Flight Analytics
Scope: consolidate migration-repair planning, repository nullable arguments, source evidence, and UUID array query semantics

## 1. Consolidated correctness scope

This increment deliberately combines related PostgreSQL contract debt instead of splitting it into several mechanical patches:

```text
migration repair embedded the historical versions 010, 011, and 012 in verifier and SQL logic
nullable UUID and text helpers returned pointer nil values without owning validation
missing source evidence was silently rewritten to the invented value "unknown"
UUID membership queries cast indexed UUID columns to text
```

The changes share one boundary: Go values must preserve meaning when they enter PostgreSQL, and PostgreSQL queries must preserve native column types.

## 2. Repository argument semantics

`nullableUUID` and `nullableText` now return concrete `database/sql/driver.Valuer` arguments rather than `*string` values.

```text
blank UUID       → SQL NULL
valid UUID       → canonical UUID text accepted by the UUID codec
malformed UUID   → ErrRepositoryUUIDArgumentInvalid
blank text       → SQL NULL
non-blank text   → trimmed text
```

This removes typed-nil ambiguity while preserving nullable database behavior.

## 3. Required source evidence

The former `sourceNameOrUnknown` helper is removed. Every former call site now uses `requiredSourceNameValue`.

```text
non-empty source → normalized persisted source
empty source     → ErrRepositorySourceNameRequired
```

The repository no longer manufactures provenance. Existing `NOT NULL` database constraints remain unchanged; missing evidence fails before persistence instead of being replaced by a believable but false source name.

## 4. Native UUID array membership

Queries shaped like:

```sql
id::text = ANY($1::text[])
```

cast the indexed UUID column and weaken type ownership. They are rewritten to:

```sql
id = ANY (
    SELECT candidate::uuid
    FROM unnest($1::text[]) AS candidates(candidate)
)
```

The incoming compatibility contract may remain a string slice, but PostgreSQL converts each candidate to UUID and compares it with the native UUID column. Invalid identifiers fail closed.

## 5. Migration repair plan

Migration repair no longer stores a duplicate checksum or a fixed `010/011/012` query.

The plan now:

```text
parses the canonical anchor migration file name
reads the anchor migration from MIGRATIONS_DIR
calculates its SHA-256 checksum from repository content
passes the plan into the PostgreSQL inspector
loads the anchor and every later applied migration through version >= anchor
blocks any applied version later than the anchor
```

The historical anchor file remains explicit because it defines the repair being verified. Version sequencing and checksum evidence are derived rather than repeated across verifier and SQL code.

## 6. Verification

Permanent tests protect:

```text
no pointer-returning nullable repository helpers
malformed UUID rejection
blank source evidence rejection
absence of sourceNameOrUnknown
absence of UUID-column text casts for array membership
real PostgreSQL UUID array membership behavior
repository-derived migration checksum
blocking of any later applied migration, not only 011 and 012
absence of the retired hard-coded migration sequence
```

The Stage 14 cross-stack command continues to execute backend, PostgreSQL, security, frontend, and container gates.

## 7. Completion boundary

This increment closes the combined PostgreSQL argument, source-evidence, UUID membership, and migration-repair generalization group.

Stage 14 remains reopened for evidence-backed trajectory query profiling and the final closure audit.
