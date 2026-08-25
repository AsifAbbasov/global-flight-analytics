# Document 73 — Stage 14.31 PostgreSQL Write Repository Decomposition

Status: Implemented
Project: Global Flight Analytics
Scope: decompose Airport Import and Flight State PostgreSQL write paths without changing their public contracts

## 1. Maintainability problem

Two PostgreSQL repository methods still owned several independent responsibilities inside
one coordinator:

```text
AirportRepository.UpsertImported
FlightStateRepository.SaveFlightStates
```

The Airport Import method owned transaction coordination, temporary staging-table creation,
batch staging, ICAO reconciliation, source-identity reconciliation, insert selection, commit,
and rollback.

The Flight State method owned transaction coordination, the complete insert statement,
altitude conversion, squawk normalization, position-source normalization, aircraft-category
validation, nullable telemetry mapping, row execution, commit, and rollback.

The behavior was correct, but the source layout made review and future changes unnecessarily
risky because a single method mixed orchestration, validation, mapping, and SQL ownership.

## 2. Preserved contracts

This increment does not change:

```text
AirportRepository.UpsertImported signature
FlightStateRepository.SaveFlightStates signature
transaction atomicity
rollback behavior
SQL statements
row ordering
normalization rules
error wording
nullable telemetry semantics
public HTTP contracts
database schema
```

No PostgreSQL migration is required.

## 3. Airport Import ownership

`airport_import_repository.go` now owns only:

```text
repository and empty-input validation
context normalization
transaction begin
rollback registration
delegation to executeAirportImport
transaction commit
inserted-record count result
```

Dedicated owner files now contain:

```text
airport_import_write_steps.go
  ordered write workflow

airport_import_staging_write.go
  temporary staging table
  batched staging inserts

airport_import_merge_write.go
  ICAO reconciliation
  source-identity reconciliation
  remaining-airport insertion
```

The write sequence remains explicit and deterministic.

## 4. Flight State ownership

`flightstate_repository.go` now owns only the public transaction boundary and delegates the
batch operation to `saveFlightStateBatch`.

`flightstate_write.go` owns:

```text
canonical Flight State insert SQL
per-row preparation
altitude persistence conversion
squawk normalization
position-source normalization
aircraft-category validation
nullable telemetry mapping
row execution and indexed error context
```

Read paths remain in `flightstate_repository.go`; this increment does not alter read
behavior.

## 5. Permanent anti-monolith gates

`write_repository_decomposition_test.go` parses the coordinator source with Go's parser and
protects:

```text
maximum coordinator method size
absence of delegated SQL in coordinator files
absence of domain preparation in SaveFlightStates
presence of dedicated owner functions
transaction begin, rollback, delegate, and commit ordering
```

The Stage 14 current-scope source audit independently requires the coordinator delegates,
owner files, architecture tests, and responsibility boundaries.

## 6. Verification

The permanent verification path includes:

```text
gofmt validation
repository package tests
Stage 14 source-audit tests
strict Stage 14 source audit
full backend correctness audit
full backend test suite
race detector
Go static analysis
vulnerability analysis
production migration catalog integration
frontend security, lint, type checking, and production build
backend container build and health verification
```

The success marker is:

```text
STAGE_14_31_WRITE_REPOSITORY_DECOMPOSITION=PASS
```

## 7. Completion boundary

Stage 14.31 closes the known Airport Import and Flight State write-method monoliths.

Stage 14 remains reopened. Airport catalog pagination requires a separate domain contract
because the current `airport.Repository.List` method returns the entire result as
`[]airport.Airport`. That contract migration is intentionally reserved for Stage 14.32
rather than being hidden inside this behavior-preserving decomposition.

## 8. Established backend audit ownership

The established Backend Final Correctness Audit now follows the decomposed ownership:

```text
flightstate_write.go
  nullable telemetry write mapping and availability helpers

flightstate_repository.go
  nullable telemetry read scanning and database-value restoration
```

The shadow preflight runs both the Backend Final Correctness Audit tests and its strict
repository audit before any real working-tree file is changed. This prevents a source
ownership refactor from passing the new Stage 14 audit while failing an older permanent
correctness gate later in the installation.

## 9. Canonical finding record — GFA-MAINT-018

### Finding / symptom

Airport Import and Flight State write coordinators mixed transaction orchestration, domain-value preparation, SQL ownership, mapping, and persistence steps inside large methods.

### Root cause

The repositories grew incrementally as correctness features were added. Each addition was locally reasonable, but ownership boundaries were not re-cut after the methods accumulated several independent responsibilities.

### Failure scenario

A future change to one concern—such as altitude conversion, staging reconciliation, or transaction cleanup—would require editing a coordinator that also owned unrelated SQL and lifecycle logic. The risk is not a known historical data-corruption event; it is increased review surface and a higher chance of accidental behavioral coupling.

### Impact

Changes become harder to reason about, code review becomes less local, architecture tests have weaker ownership targets, and future correctness fixes are more likely to touch unrelated behavior.

### Severity rationale

**P3 retrospective.** The historical document explicitly states that behavior was already correct. This is a maintainability/cohesion finding, not a production correctness incident.

### Existing guarantees violated

The issue violated the project's maintainability goal that transaction coordinators should express workflow while dedicated helpers own SQL, mapping, and domain preparation. It did not violate an external API or persisted-data guarantee.

### Considered solutions

1. leave the methods intact and rely on comments;
2. introduce new public repository interfaces/services;
3. split implementation ownership into package-private responsibility files while preserving public contracts.

### Chosen remediation and why

The implementation keeps the existing public repositories and signatures, but moves staging, merge steps, row preparation, and SQL into dedicated owners. This reduces review surface without introducing new architectural layers.

### Rejected alternatives

Comments were rejected because they do not reduce coupled change surface. New public interfaces/services were rejected as overengineering: callers did not need new abstractions, only clearer internal ownership.

### Trade-offs

The package contains more files and requires navigation between coordinator and owner files. The trade is accepted because each file has a smaller, testable responsibility and public composition remains unchanged.

### Regression tests

Parser-backed architecture tests cap coordinator responsibility, forbid delegated SQL from returning to coordinator files, and require transaction ordering plus dedicated owner functions.

### Adversarial review and remediation iterations

The decomposition was deliberately behavior-preserving. Later Stage 14 audits and the Backend Final Correctness Audit were run against the new ownership to catch a failure mode where a refactor could satisfy new structural rules while breaking older correctness gates.

### Residual risk / limitations

File decomposition does not automatically guarantee conceptual simplicity. Future contributors can still create indirect coupling across helpers, so architecture tests and ordinary review remain necessary.

### Operational / deployment consequences

None. No schema migration, API change, or deployment configuration change is required.

### Exact evidence

Implementation commit: `520779faef05b88fdeba4d9d244feb09f569010c` (`refactor: decompose postgres write repositories`). Historical PR/reviewer metadata is not asserted where it cannot be recovered reliably.

### Final canonical status

**CLOSED.** The documented monolithic write ownership was decomposed while preserving behavior.

## 10. Prevention / future guard

Future repository growth should trigger responsibility review before a coordinator becomes the owner of SQL text, normalization, mapping, transaction lifecycle, and workflow policy simultaneously. Structural tests should defend meaningful responsibilities rather than line counts alone.
