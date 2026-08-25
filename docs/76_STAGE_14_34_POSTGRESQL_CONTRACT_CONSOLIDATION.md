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

## 8. Canonical finding decomposition

```text
GFA-DB-023    hard-coded migration-repair sequence and duplicated checksum evidence
GFA-DB-024    nullable repository argument typed-nil / validation ambiguity
GFA-DATA-025  fabricated source provenance through "unknown"
GFA-PERF-026  UUID-column text casts in membership queries
```

## 9. GFA-DB-023 — Hard-coded migration-repair sequence and duplicated checksum evidence

### Finding / symptom

The migration-repair verifier encoded a fixed historical sequence around versions 010, 011, and 012 and duplicated checksum/identity information that already existed in the repository migration catalog.

### Root cause

A one-time repair safety check was implemented around the exact migrations known at that moment. The verifier therefore copied repository facts into code instead of deriving the repair plan from the canonical migration file and current applied history.

### Failure scenario

A later migration version can become applied after the repair anchor without being named by the old fixed query. The verifier may then approve a repair under assumptions that are no longer true, or stale duplicated checksum constants can drift from repository content.

### Impact

Migration repair is a high-trust operational path. Stale assumptions can weaken the proof that a historical repair is safe against the actual deployed migration sequence.

### Severity rationale

**P2 retrospective.** The issue weakens a safety verifier and can become correctness-significant during operational repair, but no evidence shows that an unsafe repair was executed.

### Existing guarantees violated

- repair evidence must be derived from the canonical migration catalog;
- later applied migrations must block a repair whose assumptions predate them;
- checksums should have one source of truth.

### Considered solutions

1. keep extending fixed version lists;
2. move the same constants to configuration;
3. derive the plan from one canonical anchor migration and inspect every applied version from that anchor forward.

### Chosen remediation and why

The repair plan parses the canonical anchor filename, reads repository content, computes the checksum, and asks PostgreSQL for the anchor plus every later applied version. Any later applied migration blocks the repair.

### Rejected alternatives

Extending fixed lists only postpones the same drift. Configuration would relocate duplicated truth rather than eliminate it.

### Trade-offs

Repair verification now depends on access to `MIGRATIONS_DIR` and repository migration content. That dependency is intentional because repository content is the evidence being verified.

### Regression tests

Tests verify repository-derived checksum ownership, generic blocking of later migrations, and absence of the retired fixed sequence.

### Adversarial review and remediation iterations

Earlier migration hardening centralized filename parsing and catalog integrity. Stage 14.34 applies the same SSOT principle to the repair subsystem so repair cannot maintain a separate interpretation of migration history.

### Residual risk / limitations

The verifier proves the repository/history preconditions it encodes; it does not independently validate arbitrary manual schema changes outside migration history.

### Operational / deployment consequences

Repair tooling becomes more conservative: unexpected later migrations block execution until operators review the situation explicitly.

### Exact evidence and final status

Implementation commit: `e850eeb0b29c9a83fb1e1f8ee2215fe80828e969` (`refactor: consolidate PostgreSQL contracts`). Historical PR/reviewer metadata not asserted where unrecoverable. **Canonical finding status: CLOSED.**

## 10. GFA-DB-024 — Nullable repository argument typed-nil and validation ambiguity

### Finding / symptom

Nullable UUID/text helpers returned pointer values, including pointer nil, and did not make UUID validation ownership explicit before the value entered PostgreSQL/pgx encoding.

### Root cause

The helper API encoded SQL nullability through Go pointer shape instead of a concrete database argument contract. This mixed absence semantics with representation details and left malformed identifier handling less explicit.

### Failure scenario

A caller passes a malformed non-blank UUID. Depending on encoding path, validation may occur later than intended and typed-nil values can be harder to reason about when passed through interfaces.

### Impact

The issue increases ambiguity at the Go→PostgreSQL boundary, weakens fail-fast diagnostics, and makes nullable behavior less type-transparent.

### Severity rationale

**P2 retrospective.** This is a persistence-boundary validation and representation issue. It can cause invalid input to fail late, but there is no evidence of silent accepted corruption.

### Existing guarantees violated

- blank and malformed values must have distinct semantics;
- malformed UUIDs should fail before persistence;
- SQL NULL representation should not depend on typed-nil pointer behavior.

### Considered solutions

1. retain pointer helpers and add comments;
2. make all domain identifiers pointer types;
3. return concrete `driver.Valuer` arguments that own validation/null semantics.

### Chosen remediation and why

Helpers return concrete database arguments. Blank values map to SQL NULL, valid UUIDs are canonicalized/validated, malformed UUIDs return a typed repository error, and text normalization is explicit.

### Rejected alternatives

Comments do not remove interface/typed-nil ambiguity. Changing domain types would broaden the refactor beyond a repository-boundary issue.

### Trade-offs

The helpers become slightly more explicit and error-returning, requiring callers/tests to handle validation. This is preferred to implicit codec failures.

### Regression tests

Tests forbid pointer-returning nullable helpers and cover blank, valid, malformed UUID, and normalized text cases.

### Adversarial review and remediation iterations

The change is deliberately limited to the persistence adapter, preserving domain/public contracts while tightening the database argument boundary.

### Residual risk / limitations

Other repository helpers can still introduce similar ambiguity unless they follow the same concrete-argument rule.

### Operational / deployment consequences

Malformed inputs fail earlier; no schema change.

### Exact evidence and final status

Implementation commit: `e850eeb0b29c9a83fb1e1f8ee2215fe80828e969`. **Canonical finding status: CLOSED.**

## 11. GFA-DATA-025 — Fabricated source provenance through `"unknown"`

### Finding / symptom

When required source evidence was missing, the repository could persist the literal value `"unknown"`, turning absence of provenance into a believable source name.

### Root cause

A convenience helper attempted to satisfy non-null persistence requirements by inventing a fallback string rather than treating missing provenance as invalid evidence.

### Failure scenario

A write path omits the source. The repository persists `"unknown"`. Downstream analytics, debugging, quality scoring, or operator SQL can no longer distinguish "source explicitly reported as unknown" from "the application lost source provenance before persistence."

### Impact

This corrupts provenance meaning. Because provenance is used to interpret open-data quality and provider behavior, fabricated evidence can mislead diagnostics and analytical confidence.

### Severity rationale

**P1 retrospective.** The system manufactures durable evidence rather than failing on missing required provenance. This is a data-integrity issue even if the numeric observation itself is otherwise valid.

### Existing guarantees violated

- provenance must reflect actual evidence;
- missing required evidence must fail closed;
- the application must not fabricate plausible source identities.

### Considered solutions

1. keep `"unknown"` fallback;
2. store NULL by relaxing schema constraints;
3. require a non-empty normalized source and reject missing evidence before SQL.

### Chosen remediation and why

`requiredSourceNameValue` rejects blank source names with `ErrRepositorySourceNameRequired`. Existing database non-null semantics remain, but the application can no longer manufacture provenance to satisfy them.

### Rejected alternatives

The fallback was rejected because it destroys evidence fidelity. Relaxing the schema was rejected because the affected records require source provenance by contract; making it nullable would weaken rather than clarify the model.

### Trade-offs

Previously tolerated incomplete writes now fail. Callers must preserve provenance end to end.

### Regression tests

Tests require blank source rejection and forbid reintroduction of `sourceNameOrUnknown`.

### Adversarial review and remediation iterations

The remediation aligns with earlier GFA quality/provenance work: uncertainty belongs in explicit status/confidence fields, not in invented source identities.

### Residual risk / limitations

A non-empty but incorrect source supplied by an upstream bug remains syntactically valid. Cross-checking truthful source identity still depends on provider-adapter ownership and tests.

### Operational / deployment consequences

Missing source data becomes a visible write failure and may surface latent caller bugs during deployment.

### Exact evidence and final status

Implementation commit: `e850eeb0b29c9a83fb1e1f8ee2215fe80828e969`. **Canonical finding status: CLOSED.**

## 12. GFA-PERF-026 — UUID-column text casts in membership queries

### Finding / symptom

Membership queries cast UUID columns to text before comparing them with text-array inputs.

### Root cause

The query adapted the indexed database column to the caller's string-slice representation rather than adapting input candidates to PostgreSQL's native UUID type.

### Failure scenario

For large tables, `id::text = ANY(...)` can weaken straightforward use of UUID index/type semantics and accepts a query shape where input validation belongs to text comparison rather than native UUID conversion.

### Impact

The query is less type-safe and can become less index-friendly on hot paths. Invalid identifiers are also not expressed as native UUID conversion failures at the intended boundary.

### Severity rationale

**P2 retrospective.** This is query-contract/performance debt with type-safety implications. No production latency incident is asserted.

### Existing guarantees violated

- indexed columns should remain in their native type in predicates when possible;
- compatibility conversion should occur on inputs, not by transforming indexed columns;
- malformed identifiers should fail closed.

### Considered solutions

1. keep text casts;
2. change every caller to pass PostgreSQL UUID arrays immediately;
3. retain string-slice compatibility but cast each input candidate to UUID before native-column comparison.

### Chosen remediation and why

The query unnests input text, casts candidates to UUID, and compares them with the UUID column unchanged. This keeps compatibility while restoring native type ownership.

### Rejected alternatives

Column casts were rejected for type/index reasons. A broad caller contract rewrite was unnecessary for this increment.

### Trade-offs

The SQL is more verbose, but the type boundary is explicit and invalid candidates fail predictably.

### Regression tests

Source tests forbid UUID-column text casts, and PostgreSQL integration verifies native UUID membership behavior.

### Adversarial review and remediation iterations

The same native-type principle is later preserved in Trajectory identifier-list queries in Document 77, showing the contract was generalized rather than treated as a one-off string rewrite.

### Residual risk / limitations

Actual planner behavior still depends on query shape, data distribution, and PostgreSQL statistics; type-correct SQL is not a substitute for profiling hot paths.

### Operational / deployment consequences

No schema change. Invalid text identifiers now fail during UUID conversion instead of participating in text comparison.

### Exact evidence and final status

Implementation commit: `e850eeb0b29c9a83fb1e1f8ee2215fe80828e969`. **Canonical finding status: CLOSED.**

## 13. Prevention / future guard

Repository boundaries must preserve semantics rather than manufacture compatibility values. Migration/verifier logic must derive catalog facts from one canonical source, required provenance must fail closed, nullable arguments must own validation explicitly, and query compatibility conversions should adapt inputs while leaving indexed PostgreSQL columns in native types.
