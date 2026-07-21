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
