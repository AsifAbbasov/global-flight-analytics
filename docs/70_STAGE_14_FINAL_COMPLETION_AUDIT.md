# Document 70 — Stage 14 Final Completion Audit

Status: Remediation History / Closed v2.1
Project: Global Flight Analytics
Scope: own the cross-stack gate and the independently verified final Stage 14 closure decision

## 1. Audit reason

Stage 14 accumulated permanent package tests and evidence documents across architecture,
security, analytical reachability, PostgreSQL correctness, data semantics, and repository
maintainability. The existing backend final-correctness script remained useful, but its
name and implementation were backend-specific while Document 56 described a wider
acceptance boundary.

The final audit found two closure gaps:

```text
no single command executed backend, PostgreSQL integration, frontend, security, and
container checks together

Flight Feature timestamp integration tests used TEST_DATABASE_URL but the PostgreSQL
continuous integration job executed only internal/repository/postgres
```

These gaps did not invalidate the successful installer evidence for Documents 41–69.
They meant that the same evidence was not yet permanently reproducible through one
repository-owned gate.

The first execution of this unified gate then found a third evidence-backed blocker:
`govulncheck` confirmed five reachable standard-library vulnerabilities because
`apps/api/go.mod` pinned Go 1.26.2. The Docker builder already used Go 1.26.5, but local
and continuous integration toolchain selection followed `go.mod`. This amendment pins
Go 1.26.5 everywhere and makes the final audit verify the effective toolchain before any
acceptance claim.

## 2. Final source audit

The repository now contains:

```text
apps/api/tools/stage14finalaudit
```

The tool verifies:

```text
Documents 41 through 70 exist in one contiguous register
DOCUMENT_INDEX.md references every Stage 14 document exactly once
one cross-stack verification script is reachable from package.json
backend continuous integration runs the Stage 14 source audit
PostgreSQL continuous integration runs repository and Flight Feature timestamp tests
frontend continuous integration preserves dependency, lint, type, and build gates
the decomposed Trajectory Repository cannot collapse back into two monolithic files
Traffic altitude, Airport elevation, and Flight Feature timestamp semantics remain explicit
all complete FlightStateRepository integration fixtures contain current evidence columns
all direct terminal ingestion-run integration fixtures include finished_at
Document 58 records closure of the known PostgreSQL debt register
```

The tool is source-level governance. Behavioral correctness remains owned by the package,
integration, race, frontend, and container checks that the final script executes.

## 3. Unified completion command

From the repository root:

```bash
scripts/verify-stage-14-completion.sh
```

The same command is also exposed through:

```bash
pnpm run verify:stage14
```

The script executes:

```text
repository diff validation
exact Go 1.26.5 toolchain selection
Go module and Docker builder pin validation
Go formatting validation
Stage 14 source-audit tests and strict source audit
established backend final-correctness audit
pinned Go vulnerability analysis with the patched Go 1.26.5 standard library
isolated PostgreSQL 16 integration database
PostgreSQL repository integration
Flight Feature timestamp integration
frontend dependency policy tests
frontend production dependency audit
frontend lint
frontend TypeScript validation
frontend production build
Docker Compose validation
backend container build
non-root runtime-user verification
backend container health smoke test
final source audit and diff validation
```

Successful completion ends with:

```text
STAGE_14_CURRENT_SCOPE_AUDIT=PASS
```

## 4. Go standard-library security correction

The repository language version is pinned to:

```text
go 1.26.5
```

The backend builder remains pinned to:

```text
golang:1.26.5-alpine3.24
```

The unified audit exports `GOTOOLCHAIN=go1.26.5+auto`, verifies the effective
`go env GOVERSION`, and refuses to continue unless it is exactly `go1.26.5`. This allows
an older Go 1.21-or-newer host command to download the required patched toolchain while
keeping local, continuous integration, and container builds aligned.

The failed Go 1.26.2 audit and its automatic repository restoration are retained as
negative security evidence. Vulnerability analysis remains mandatory and is not muted,
filtered, or converted into a warning.

The next PostgreSQL execution exposed two additional blockers that source-only and
skipped integration tests had not revealed:

```text
migration 018 used an unqualified point_count aggregate while joining the parent and
segment tables, so a clean PostgreSQL migration failed with SQLSTATE 42702

the Flight State altitude integration fixture predated the transponder and aircraft
evidence columns now written by FlightStateRepository, so repository inserts failed with
SQLSTATE 42703 before altitude behavior could be tested
```

Migration 018 is corrected in place because the ambiguous statement executes inside its
opening transaction before any schema change or migration-history record can commit. The
broken form therefore cannot exist as a successfully applied checksum baseline. The test
fixture is expanded to match the current repository persistence contract; production
schema semantics are unchanged.

The following complete PostgreSQL run exposed one final fixture-parity defect: the derived
identity integration fixture inserted an ingestion run directly with terminal `success`
status but omitted `finished_at`. Migration 017 correctly rejected that row through
`ingestion_runs_lifecycle_check`. The fixture now records `started_at` and `finished_at`
for the terminal row, and the source audit scans every PostgreSQL integration test for
direct terminal ingestion-run inserts that omit finish evidence. This correction changes
no production lifecycle semantics; it makes test evidence obey Document 62.

## 5. PostgreSQL integration correction

The backend continuous integration PostgreSQL job now runs:

```text
internal/repository/postgres
internal/features/featurestore
```

This makes the exact Unix-nanosecond and PostgreSQL timestamp mirror contract from
Document 68 a real PostgreSQL continuous integration boundary rather than a test that
normally skipped without `TEST_DATABASE_URL`.

The local final audit starts an isolated PostgreSQL 16 container, waits for readiness,
runs both packages with an explicit test database URL, and removes the container after
the tests finish.

## 6. Cross-stack acceptance boundary

The backend final-correctness script remains the owner of backend source invariants,
focused race testing, command compilation, static analysis, and the complete Go test
suite.

The new Stage 14 script composes that backend gate with:

```text
production dependency security
frontend validation and production build
real PostgreSQL integration
backend container construction and health
final documentation and ownership audit
```

This resolves the mismatch between the former backend-only script and the broader
acceptance language in Document 56 without weakening or duplicating the established
backend checks.

## 7. Container evidence

The final gate validates the same production boundary represented in backend continuous
integration:

```text
Docker Compose configuration parses
backend image builds from apps/api/Dockerfile
runtime user is exactly 10001:10001
container health reaches healthy
GET /api/v1/health succeeds through the published local port
```

Temporary PostgreSQL and backend containers and the temporary audit image are removed
through an exit trap on both success and failure.

## 8. Preserved boundaries

This increment does not change:

```text
production API behavior
analytical formulas
provider behavior
PostgreSQL schema shape or applied migration history
domain contracts
frontend visual behavior
Trajectory Repository public interfaces
Stage 14 implementation documents 41–69
```

No new PostgreSQL migration number is required. The previously unapplyable migration 018 source is repaired before Stage 14 closure.

## 9. Completion decision

Stage 14 is accepted only when the unified command completes and prints:

```text
STAGE_14_CURRENT_SCOPE_AUDIT=PASS
```

A passing marker proves the repository state passed the source, backend, PostgreSQL,
security, frontend, and backend container gates defined in this document. It does not
claim deployed Render or Neon availability, production traffic load evidence, browser
end-to-end coverage, or completion of Stage 15 frontend visual design.

The marker proves only the checks implemented by this script. It does not prove that the
complete Stage 14 debt register is closed. Stage 14 is reopened until the remaining
evidence-backed correctness and maintainability findings are either fixed or explicitly
accepted as documented non-blocking design decisions.

## 18. Scoped PostgreSQL fixture classification

The first broad fixture-parity rule treated every test-local `flight_states` table as if
it were a complete production repository schema. That was incorrect. Data Quality
migration tests, active-aircraft metric tests, and Traffic query tests intentionally own
minimal tables containing only the columns used by the behavior under test.

The permanent rule now applies full evidence-column parity only when an integration test
both creates `flight_states` and instantiates `NewFlightStateRepository`. The
`flightstate_reconciliation_repository_integration_test.go` fixture is upgraded because
it exercises that complete repository. Purpose-built minimal schemas remain narrow and
are not padded with unrelated columns.

## 19. Reopening decision

The committed `eb37e03` state contained two SQL files with migration version `016`.
`migrator.Runner.ListMigrations` rejects duplicate versions, but the former unified gate
never executed the production migration runner against the full repository catalog.
Therefore its completion marker was not valid evidence of deployability.

Document 71 introduces a repository-catalog test, moves Data Quality Parent Integrity to
version `019`, and runs `cmd/migrate` twice against a clean PostgreSQL database. The
former completion marker is retired. The current script ends with:

```text
STAGE_14_CURRENT_SCOPE_AUDIT=PASS
STAGE_14_OVERALL_STATUS=REOPENED
```

<!-- STAGE-14-30-POSTGRES-CORRECTNESS-HARDENING:DOCUMENT-70 -->

## Stage 14.30 current-scope amendment

The PostgreSQL gate now applies migration 020 and runs repository, Feature Store, Route
Store, and Historical Aggregate tests. It verifies Ingestion Run evidence constraints,
timestamp mirrors, and independent rollback ownership. `STAGE_14_CURRENT_SCOPE_AUDIT=PASS`
remains a current-scope marker only; `STAGE_14_OVERALL_STATUS=REOPENED` remains authoritative.

<!-- STAGE-14-31-POSTGRES-WRITE-REPOSITORY-DECOMPOSITION:DOCUMENT-70 -->

## Stage 14.31 write-repository decomposition amendment

The Airport Import and Flight State write coordinators no longer own embedded SQL and
preparation workflows. Dedicated owner files and parser-backed architecture tests are part
of the current-scope gate. The Stage 14 overall status remains reopened because Airport
catalog pagination requires a separate domain contract migration.

<!-- STAGE-14-32-AIRPORT-KEYSET-PAGINATION:FINAL-AUDIT -->

## Reopened Scope Update: Airport Pagination

Stage 14.32 removes the unbounded Airport catalog query by introducing keyset `ListPage`
reads ordered by `(name, id)`. The legacy `List` contract is preserved as a bounded-page
adapter, and all Airport read paths use one scanner. The overall Stage 14 status remains
reopened for the remaining explicitly recorded backlog.

<!-- STAGE-14-33-EXPLICIT-REPOSITORY-CONTEXT-AND-TRAJECTORY-WRITE-MODE:FINAL-AUDIT -->

## Reopened Scope Update: Explicit Context and Trajectory Write Mode

Stage 14.33 removes silent nil-context replacement from the selected PostgreSQL repository
boundaries and replaces the Trajectory empty-task-identifier mode sentinel with a validated
typed request. The independent rollback background context remains intentional cleanup.
The overall Stage 14 status remains reopened for the remaining recorded backlog.

<!-- STAGE-14-34-POSTGRESQL-CONTRACT-CONSOLIDATION:FINAL-AUDIT -->

## Reopened Scope Update: PostgreSQL Contract Consolidation

Stage 14.34 closes the combined migration-repair, nullable-argument, source-provenance, and
UUID membership query group. The cross-stack audit now protects these boundaries. The overall
Stage 14 status remains reopened for evidence-backed trajectory index profiling and final
closure.

<!-- STAGE-14-35-TRAJECTORY-QUERY-CONSOLIDATION-AND-PROFILING:FINAL-AUDIT -->

## Stage 14.35 current-scope evidence

Trajectory query ownership, row mapping, caller context, index migration, and EXPLAIN ANALYZE evidence are now enforced by permanent source and PostgreSQL integration gates. A successful current run emits `STAGE_14_35_TRAJECTORY_QUERY_PROFILING=PASS`. This is not the Stage 14 closure statement; Stage 14 remains reopened until the independent final closure audit.

<!-- STAGE-14-36-FINAL-CLOSURE:DOCUMENT-70 -->

## 20. Final closure amendment

Stage 14.35 completed the final recorded technical backlog: Trajectory query ownership,
row-scanner consolidation, caller-context consistency, migration 021, duplicate-index removal,
and permanent `EXPLAIN ANALYZE` evidence.

Stage 14.36 changes no application behavior. It independently reruns the complete authoritative
command after Stage 14.35 is committed, adds Document 78 to the contiguous register, and makes
the current status declaration machine-readable.

A successful run now ends with:

```text
STAGE_14_36_FINAL_CLOSURE_AUDIT=PASS
STAGE_14_CURRENT_SCOPE_AUDIT=PASS
STAGE_14_OVERALL_STATUS=CLOSED
```

Historical reopening sections above remain evidence of prior invalid closure attempts. They do
not override the final marker. Stage 14 is closed; later feature development belongs to a new
stage.

## 21. Canonical finding decomposition

Document 70 owns only the unique audit/closure defects below. The later reopening amendments in
this file are chronology and integration evidence, not duplicate finding ownership.

```text
GFA-GOV-056   Missing unified cross-stack Stage 14 acceptance gate
GFA-GOV-057   Flight Feature timestamp PostgreSQL CI coverage gap
GFA-SEC-058   Reachable Go standard-library vulnerabilities and toolchain drift
GFA-DB-059    Migration 018 ambiguous aggregate blocked clean migration
GFA-TEST-060  Full FlightStateRepository integration fixture parity drift
GFA-TEST-061  Terminal ingestion-run test fixture omitted finished_at
GFA-GOV-062   Over-broad PostgreSQL fixture-parity audit rule
```

### Deduplication / canonical ownership rule

The duplicate migration-version blocker discovered after `eb37e03` is owned by
`GFA-DB-013` / Document 71. Stage 14.30–35 findings are owned by Documents 72–77 and their
registered findings `GFA-DB-014` through `GFA-PERF-029`. The post-closure migrator context
finding is owned by `GFA-DB-030` / Document 79. Document 78 remains stage-level final closure
evidence. They are intentionally **not** registered again here.

## 22. GFA-GOV-056 — Missing unified cross-stack Stage 14 acceptance gate

### Finding / symptom

Stage 14 had strong subsystem verification, but no single repository-owned command composed
backend correctness, real PostgreSQL integration, frontend validation/security, and production
container evidence into one acceptance boundary.

### Root cause

Verification grew stage by stage. Document 56 introduced a backend-specific final gate, while
later closure language implicitly depended on additional PostgreSQL, frontend and container
checks that were not one executable contract.

### Failure scenario

```text
backend final audit passes
↓
separate PostgreSQL/frontend/container checks are skipped or run inconsistently
↓
repository is declared Stage 14-complete
↓
one required stack boundary was never part of the reproducible closure command
```

### Impact

Closure evidence could overstate what had actually been executed. This is a release/governance
integrity problem and can hide defects that only appear in real integration or container paths.

### Severity rationale

Historical severity was not recorded. Retrospective classification: **P2 governance/reliability**.
The gap did not itself alter production behavior, but it weakened the evidence required to close a
stage containing multiple P1/P2 correctness and security remediations.

### Existing guarantees violated

```text
stage closure must be reproducible through repository-owned commands
closure claims must match the checks actually executed
backend-only evidence cannot silently stand in for PostgreSQL/frontend/container evidence
```

### Considered solutions

1. Keep a manual checklist of independent commands.
2. Expand `verify-backend-final-correctness.sh` until it owned the entire stack.
3. Create a thin Stage 14 orchestration gate that composes the existing owners.

### Chosen remediation

Option 3: `scripts/verify-stage-14-completion.sh` and `pnpm run verify:stage14` compose the
existing backend audit with PostgreSQL 16 integration, dependency/frontend validation,
container build/non-root/health checks, and a dedicated `stage14finalaudit` source guard.

### Why this solution was selected

It creates one authoritative stage-level command without duplicating subsystem test logic or
misnaming the backend-specific verifier as a whole-stack gate.

### Rejected alternatives

Manual checklists were rejected as non-reproducible. Expanding the backend script into a generic
release framework was rejected because it would blur existing ownership and grow a monolithic
verification script.

### Trade-offs

```text
+ one reproducible closure command
+ explicit cross-stack ownership
+ existing subsystem gates remain independent and reusable
- full local execution is heavier than backend-only verification
- orchestration must be maintained when required stack gates change
```

### Regression tests / protection

`apps/api/tools/stage14finalaudit` verifies command reachability and CI ownership. Backend CI runs
the Stage 14 source audit. The unified command performs beginning/end diff validation and fails on
any component failure.

### Adversarial review findings

Later review proved that even the first unified gate was not sufficient for total Stage 14 closure
because it did not execute the production migration runner against the full catalog. That blocker is
separately owned by Document 71. The lesson is that a unified gate has a bounded claim; “unified”
does not mean omniscient.

Historical PR/reviewer metadata for the July remediation is unavailable; reconstruction is limited
to repository source, tests, commit `eb37e03...`, and subsequent closure documents.

### Remediation iterations

1. Backend-specific final gate existed in Document 56.
2. Commit `eb37e03c6793314e446cdb048ae9584e38f2567c` added the cross-stack Stage 14 gate.
3. Document 71 later reopened Stage 14 after discovering migration-catalog coverage was absent.
4. Document 78 established the later independent final closure after the remaining backlog closed.

### Residual risks and limitations

The command proves only the checks it owns. It does not prove deployed Render/Neon availability,
production traffic/load behavior, or unrelated future-stage acceptance requirements.

### Operational or deployment consequences

Stage 14 closure/revalidation requires the cross-stack command, including temporary PostgreSQL and
backend containers locally where applicable. Temporary resources are cleaned through failure-safe
traps.

### Exact evidence

```text
implementation commit:
eb37e03c6793314e446cdb048ae9584e38f2567c

primary command:
scripts/verify-stage-14-completion.sh

source guard:
apps/api/tools/stage14finalaudit

later final closure:
202c00cabb352b50a6d3a2a6002ad3401c1ad23e
```

### Final canonical status

```text
FINDING_GFA_GOV_056_UNIFIED_STAGE14_GATE=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/70_STAGE_14_FINAL_COMPLETION_AUDIT.md
IMPLEMENTATION_COMMIT=eb37e03c6793314e446cdb048ae9584e38f2567c
```

### Prevention / future guard

Stage/release closure claims must point to one explicit orchestration boundary and list the
subsystem gates it composes. Adding a new mandatory acceptance surface requires updating both the
orchestrator and its source-level governance check.

## 23. GFA-GOV-057 — Flight Feature timestamp PostgreSQL CI coverage gap

### Finding / symptom

The Flight Feature timestamp mirror integration tests required `TEST_DATABASE_URL`, but the
PostgreSQL CI job executed only `internal/repository/postgres`. The exact timestamp contract from
Document 68 could therefore exist without real PostgreSQL execution in ordinary CI.

### Root cause

The PostgreSQL job was scoped around repository integration before Feature Store gained its own
PostgreSQL-backed integrity contract. Test existence was mistaken for permanent CI execution.

### Failure scenario

```text
featurestore PostgreSQL integration test exists
↓
CI omits that package / TEST_DATABASE_URL path
↓
timestamp mirror regression compiles and non-DB tests pass
↓
Stage 14 closure cites evidence that normal CI does not actually execute
```

### Impact

The permanent evidence for a P1 timestamp consistency remediation was weaker than documented and
could allow mirror corruption to regress without PostgreSQL CI detection.

### Severity rationale

Historical severity was not recorded. Retrospective classification: **P2 test/evidence governance**.
The underlying timestamp corruption finding is P1 (`GFA-DB-011`); this finding concerns missing
continuous execution of its integration evidence.

### Existing guarantees violated

```text
regression tests for closed database-integrity findings must run in CI
TEST_DATABASE_URL-dependent tests cannot count as CI evidence when the job never supplies it
closure documentation and workflow execution must agree
```

### Considered solutions

1. Treat package-local tests as sufficient even when skipped without a database.
2. Add a separate Feature Store PostgreSQL workflow.
3. Extend the existing PostgreSQL 16 integration job to run both repository and featurestore packages.

### Chosen remediation

Option 3. Backend PostgreSQL CI now executes both:

```text
./internal/repository/postgres
./internal/features/featurestore
```

with the integration database available.

### Why this solution was selected

The existing PostgreSQL job already owns database lifecycle, so extending its package list adds the
missing evidence with minimal workflow duplication.

### Rejected alternatives

Skipped tests were rejected as false closure evidence. A second database workflow was rejected as
unnecessary duplication of PostgreSQL setup and cleanup.

### Trade-offs

CI performs additional database-backed tests and therefore takes slightly longer, but the timestamp
contract becomes continuously executable rather than documentary.

### Regression tests / protection

`stage14finalaudit` verifies that backend CI includes both PostgreSQL repository and Feature Store
integration packages. The unified Stage 14 command also runs them against an isolated PostgreSQL 16
instance.

### Adversarial review findings

The key review insight was that “test exists” and “test runs in required CI” are different evidence
states. Any database test whose environment normally causes a skip must have explicit workflow
ownership before it can support a closure claim.

Historical PR/reviewer metadata is unavailable; reconstruction is limited to repository workflows,
tests, Document 68/70, and commit `eb37e03...`.

### Remediation iterations

The timestamp consistency implementation closed in Document 68. Document 70's cross-stack audit
then discovered the CI coverage gap and commit `eb37e03...` wired Feature Store into the PostgreSQL
job.

### Residual risks and limitations

This guarantees CI execution for the registered package, not every future database-backed test in
the repository. New packages still need explicit integration-job ownership.

### Operational or deployment consequences

No production behavior changes. CI requires the PostgreSQL 16 service to remain available for the
Feature Store integration package.

### Exact evidence

```text
implementation commit:
eb37e03c6793314e446cdb048ae9584e38f2567c

underlying timestamp remediation:
GFA-DB-011 / Document 68

workflow evidence:
.github/workflows/backend-ci.yml PostgreSQL 16 Integration
```

### Final canonical status

```text
FINDING_GFA_GOV_057_FEATURESTORE_POSTGRES_CI=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/70_STAGE_14_FINAL_COMPLETION_AUDIT.md
IMPLEMENTATION_COMMIT=eb37e03c6793314e446cdb048ae9584e38f2567c
```

### Prevention / future guard

Every database-backed regression test used as closure evidence must identify the CI job and
environment that executes it without skipping. Stage-level audits should verify this ownership for
high-risk contracts.

## 24. GFA-SEC-058 — Reachable Go standard-library vulnerabilities and toolchain drift

### Finding / symptom

The first unified Stage 14 run executed `govulncheck` and found five reachable standard-library
vulnerabilities because `apps/api/go.mod` selected Go 1.26.2 while the Docker builder already used
Go 1.26.5.

### Root cause

Go toolchain ownership was split: the module version controlled local/CI selection, while the
container builder had already advanced to the patched release. There was no final gate asserting the
effective toolchain used by vulnerability analysis and tests.

### Failure scenario

```text
container Dockerfile uses patched Go
↓
local/CI follows go.mod and runs Go 1.26.2
↓
govulncheck reports reachable standard-library vulnerabilities
↓
source/tests may be accepted under a vulnerable toolchain despite patched container intent
```

### Impact

Reachable vulnerable standard-library code remained in the backend build/test toolchain selected by
the repository. No exploitation evidence is claimed, but Stage 14 could not be closed while
reachable vulnerabilities were confirmed.

### Severity rationale

Historical severity was not recorded. Retrospective classification: **P1 security/release blocker**
because `govulncheck` reported reachable vulnerabilities in the selected standard library. This
classification does not assert exploitation or a known production incident.

### Existing guarantees violated

```text
known reachable production/backend vulnerabilities must block closure
module, CI and container toolchains should not drift across security-fix releases
vulnerability analysis must fail closed rather than be muted or filtered
```

### Considered solutions

1. Ignore/suppress the stdlib findings because the Docker builder was newer.
2. Pin only CI to 1.26.5.
3. Align module and Docker builder to 1.26.5 and make the Stage 14 gate verify effective GOVERSION.

### Chosen remediation

Option 3. The repository pins `go 1.26.5`, retains `golang:1.26.5-alpine3.24`, exports
`GOTOOLCHAIN=go1.26.5+auto`, verifies `go env GOVERSION == go1.26.5`, and keeps pinned
`govulncheck` mandatory.

### Why this solution was selected

Security is tied to the effective toolchain, not only the Dockerfile. Aligning every owner removes
ambiguity and makes local, CI and container evidence consistent.

### Rejected alternatives

Suppressing the findings was rejected because they were reachable. CI-only pinning was rejected
because local/repository module semantics would still disagree and reproducibility would remain
weak.

### Trade-offs

Hosts may need Go's toolchain auto-download path, and the repository must deliberately update all
pins when future security releases are adopted. In return, toolchain provenance is explicit.

### Regression tests / protection

`stage14finalaudit` checks `apps/api/go.mod`, the Docker builder image and Stage 14 toolchain
selection. The unified script refuses to proceed under the wrong effective Go version and runs
pinned `govulncheck` without warning-only conversion.

### Adversarial review findings

The failed Go 1.26.2 audit is retained as negative evidence. Review explicitly rejected the tempting
argument that a patched Docker builder automatically makes local/CI vulnerability findings
irrelevant.

Historical PR/reviewer metadata is unavailable; reconstruction is limited to commit `eb37e03...`,
Document 70, toolchain files and audit code.

### Remediation iterations

The Docker builder had already moved to 1.26.5. The unified audit exposed the remaining `go.mod`
selection at 1.26.2; `eb37e03...` aligned the module and made exact effective-version verification
part of closure.

### Residual risks and limitations

A patched Go version does not prove third-party dependencies are vulnerability-free; separate
module/dependency scans remain required. New advisories can also appear after closure and require a
new security update.

### Operational or deployment consequences

Backend builds/tests/audits require Go 1.26.5 semantics. Older compatible hosts may use the Go
toolchain auto-download mechanism configured by the audit.

### Exact evidence

```text
implementation commit:
eb37e03c6793314e446cdb048ae9584e38f2567c

patched module:
apps/api/go.mod -> go 1.26.5

patched builder:
apps/api/Dockerfile -> golang:1.26.5-alpine3.24

mandatory scanner:
govulncheck@v1.1.4
```

### Final canonical status

```text
FINDING_GFA_SEC_058_GO_STDLIB_TOOLCHAIN=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/70_STAGE_14_FINAL_COMPLETION_AUDIT.md
IMPLEMENTATION_COMMIT=eb37e03c6793314e446cdb048ae9584e38f2567c
```

### Prevention / future guard

Security-sensitive toolchain versions must have one declared repository version and matching
container/CI evidence. Reachable vulnerability findings cannot be waived merely because another
build surface happens to use a patched version.

## 25. GFA-DB-059 — Migration 018 ambiguous aggregate blocked clean PostgreSQL migration

### Finding / symptom

A clean PostgreSQL execution of migration 018 failed with `SQLSTATE 42702` because its aggregate
query used unqualified `point_count` (and related unqualified segment columns) while joining
trajectory parent and segment scopes.

### Root cause

Migration SQL relied on column names without consistently qualifying the intended
`trajectory_segments AS segment` owner in a joined/PLpgSQL context.

### Failure scenario

```text
clean database applies migration 018
↓
PostgreSQL resolves aggregate query containing SUM(point_count)
↓
column reference is ambiguous
↓
transaction aborts with SQLSTATE 42702
↓
clean repository bootstrap cannot complete
```

### Impact

The repository's production migration sequence was not clean-bootstrap deployable. The failure was
caught before successful application, but it blocked database deployment and invalidated closure
evidence based only on source-level checks.

### Severity rationale

Historical severity was not recorded. Retrospective classification: **P1 deployment/correctness**
because a committed production migration failed on a clean PostgreSQL database.

### Existing guarantees violated

```text
committed production migrations must apply from a clean supported database
migration SQL must have unambiguous column ownership
failed migrations must not produce a successful schema_migrations baseline
```

### Considered solutions

1. Add a new follow-up migration to repair migration 018.
2. Change migration 018 in place to qualify segment-owned columns.
3. Relax/skip the clean migration path in the final audit.

### Chosen remediation

Option 2: qualify the segment aggregate/ordering columns (`SUM(segment.point_count)`,
`MIN(segment.sequence_number)`, `MAX(segment.sequence_number)`, etc.) and make the source audit
reject the ambiguous form.

### Why this solution was selected

The broken statement executes inside migration 018's opening transaction and fails before schema
changes or migration-history recording can commit. Therefore the broken form cannot be a valid,
successfully applied checksum baseline that needs a later repair migration.

### Rejected alternatives

A follow-up migration was rejected because there is no successful broken state to migrate from.
Skipping the clean run was rejected because that would preserve false deployability evidence.

### Trade-offs

The committed migration source changes in place, which is normally sensitive. Here the transactional
failure-before-history property is the explicit safety condition that permits it.

### Regression tests / protection

`stage14finalaudit` requires qualified segment expressions and has a negative test that mutates
`SUM(segment.point_count)` back to `SUM(point_count)` and expects audit failure. Clean production
migration execution remains part of later PostgreSQL gates.

### Adversarial review findings

The remediation explicitly asked whether editing an existing migration was safe. It was accepted
only after establishing that the ambiguous statement cannot successfully commit any schema/history
record in the broken form.

Historical PR/reviewer metadata is unavailable; reconstruction is limited to migration source,
audit tests, commit `eb37e03...` and Document 70.

### Remediation iterations

The source-only audit initially passed. Real PostgreSQL execution then surfaced `42702`, leading to
the in-place qualification and permanent source/integration guard in `eb37e03...`.

### Residual risks and limitations

Qualification fixes this ambiguity; it does not prove every statement in every migration is
semantically correct. Full migration catalog execution remains independently required and later
found the duplicate-version blocker owned by Document 71.

### Operational or deployment consequences

Clean database bootstrap requires the corrected migration 018 source. No new migration number or
schema shape is introduced by this correction.

### Exact evidence

```text
implementation commit:
eb37e03c6793314e446cdb048ae9584e38f2567c

migration:
database/migrations/018_trajectory_relational_integrity.sql

observed failure:
SQLSTATE 42702
```

### Final canonical status

```text
FINDING_GFA_DB_059_MIGRATION018_AMBIGUITY=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/70_STAGE_14_FINAL_COMPLETION_AUDIT.md
IMPLEMENTATION_COMMIT=eb37e03c6793314e446cdb048ae9584e38f2567c
```

### Prevention / future guard

Production migrations with joined or procedural query scopes must qualify columns whenever
ownership could be ambiguous, and clean PostgreSQL migration execution remains mandatory closure
evidence rather than relying on source review alone.

## 26. GFA-TEST-060 — Full FlightStateRepository integration fixture parity drift

### Finding / symptom

The Flight State altitude integration fixture predated repository evidence columns such as
`squawk_code`, `special_purpose_indicator`, `position_source`, `aircraft_category`, and
`aircraft_category_available`. `FlightStateRepository` inserts therefore failed with
`SQLSTATE 42703` before the intended behavior could be tested. A reconciliation fixture exercising
the same full repository contract required the same parity treatment.

### Root cause

Test-local full repository schemas were copied before the production persistence contract expanded
and had no explicit parity owner tying them to `NewFlightStateRepository` writes.

### Failure scenario

```text
integration test constructs local flight_states table
↓
fixture omits columns currently written by FlightStateRepository
↓
repository INSERT references missing column
↓
test fails before reaching altitude/reconciliation assertion
```

### Impact

Integration evidence became stale and could neither validate the intended behavior nor accurately
model the current repository persistence contract. This is a test-evidence defect, not evidence of a
production-schema defect.

### Severity rationale

Historical severity was not recorded. Retrospective classification: **P2 test/evidence integrity**
because high-value repository integration tests were unable to execute their target behavior and
could misrepresent production parity.

### Existing guarantees violated

```text
full repository integration fixtures must match columns written by that repository
integration failures should reach the behavior under test rather than fail on stale fixture shape
purpose-built minimal schemas must not be conflated with full-repository fixtures
```

### Considered solutions

1. Make every test-local `flight_states` table mirror the complete production schema.
2. Update only fixtures that instantiate the full `FlightStateRepository` and keep minimal
   behavior-specific schemas narrow.
3. Stop using local fixtures and run all tests only against the entire production migration catalog.

### Chosen remediation

Option 2: upgrade complete repository fixtures with the current evidence columns and scope the
source audit to tests that both create `flight_states` and instantiate `NewFlightStateRepository`.

### Why this solution was selected

Parity is required where the real repository contract is exercised, while minimal schemas are useful
for isolated behavior tests and should not accumulate unrelated columns.

### Rejected alternatives

Global full-schema padding was rejected after producing false positives and needless fixture
coupling. Requiring the entire production catalog for every focused test was rejected as slow and
less isolated.

### Trade-offs

Full repository fixtures require maintenance when repository writes change. Minimal fixtures remain
lean but require classification so the audit knows which parity standard applies.

### Regression tests / protection

`stage14finalaudit` checks current evidence columns in the altitude and reconciliation fixtures and
contains negative fixture tests. The scoped classification rule prevents unrelated minimal schemas
from being treated as full repository copies.

### Adversarial review findings

The initial parity rule was intentionally challenged and found over-broad; this directly led to the
separate `GFA-GOV-062` scope correction. That iteration is preserved rather than hidden.

Historical PR/reviewer metadata is unavailable; reconstruction is limited to test fixtures, source
audit code, commit `eb37e03...` and Document 70.

### Remediation iterations

1. PostgreSQL execution surfaced missing-column `42703` in the altitude fixture.
2. Full repository fixtures were updated.
3. A broad parity audit produced false positives on intentional minimal schemas.
4. The rule was narrowed to true `NewFlightStateRepository` fixtures.

### Residual risks and limitations

Future full repository fixtures can still drift if created outside the guarded patterns. The source
audit must evolve when repository construction patterns legitimately change.

### Operational or deployment consequences

None for production. CI fixtures now model the current persistence contract accurately where the
full repository is exercised.

### Exact evidence

```text
implementation commit:
eb37e03c6793314e446cdb048ae9584e38f2567c

observed failure:
SQLSTATE 42703

protected fixtures include:
flightstate_altitude_integration_test.go
flightstate_reconciliation_repository_integration_test.go
```

### Final canonical status

```text
FINDING_GFA_TEST_060_FLIGHTSTATE_FIXTURE_PARITY=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/70_STAGE_14_FINAL_COMPLETION_AUDIT.md
IMPLEMENTATION_COMMIT=eb37e03c6793314e446cdb048ae9584e38f2567c
```

### Prevention / future guard

Any integration fixture that exercises a complete repository must declare/maintain parity with the
repository's write contract. Minimal behavior fixtures remain explicitly exempt unless they use the
full repository constructor.

## 27. GFA-TEST-061 — Terminal ingestion-run fixture omitted `finished_at`

### Finding / symptom

A derived-identity integration fixture inserted an ingestion run directly with terminal `success`
status but omitted `finished_at`; migration 017 correctly rejected the row through
`ingestion_runs_lifecycle_check`.

### Root cause

The test fixture retained an older direct-insert shape and bypassed the current terminal lifecycle
evidence requirements introduced by Document 62.

### Failure scenario

```text
test inserts ingestion_runs(status='success') directly
↓
finished_at is NULL
↓
production lifecycle CHECK constraint rejects fixture row
↓
downstream integration scenario never executes
```

### Impact

The fixture no longer represented a valid production lifecycle state and blocked integration
evidence. Production behavior was correct; the test was stale.

### Severity rationale

Historical severity was not recorded. Retrospective classification: **P2 test/evidence integrity**
because a correctness integration scenario could not execute and the fixture contradicted a P1
lifecycle invariant.

### Existing guarantees violated

```text
terminal ingestion-run rows require finish evidence
integration fixtures must obey production lifecycle constraints
direct SQL test setup cannot silently bypass newer invariants
```

### Considered solutions

1. Weaken the production check for tests.
2. Special-case the fixture through disabled constraints.
3. Update the fixture with valid `started_at`/`finished_at` evidence and audit direct terminal inserts.

### Chosen remediation

Option 3. The fixture records both timestamps for the terminal row, and the Stage 14 source audit
scans PostgreSQL integration tests for direct terminal inserts missing finish evidence.

### Why this solution was selected

The production constraint was correct. Test data must move forward to the canonical lifecycle rather
than weakening the invariant to preserve an obsolete fixture.

### Rejected alternatives

Constraint weakening and test-only bypasses were rejected because they would destroy the value of
integration evidence and reopen the underlying lifecycle risk.

### Trade-offs

Direct fixture setup becomes slightly more verbose, but it now communicates terminal lifecycle
semantics explicitly.

### Regression tests / protection

The source audit detects terminal ingestion-run integration fixtures that omit `finished_at`, while
PostgreSQL integration naturally enforces `ingestion_runs_lifecycle_check`.

### Adversarial review findings

The complete PostgreSQL run surfaced this only after earlier migration/fixture blockers were fixed,
showing why iterative end-to-end integration execution matters: the next valid boundary can reveal
a deeper stale fixture.

Historical PR/reviewer metadata is unavailable; reconstruction is limited to integration fixtures,
migration 017 constraints, audit code, commit `eb37e03...` and Document 70.

### Remediation iterations

The fixture defect was the final PostgreSQL parity issue found by that audit sequence after migration
018 ambiguity and FlightState fixture columns had been corrected.

### Residual risks and limitations

The static scan protects known direct-insert forms. New fixture helpers or generated SQL may require
the audit to evolve, while PostgreSQL constraints remain the final behavioral guard.

### Operational or deployment consequences

None for production. CI setup now obeys the same terminal lifecycle rule as production data.

### Exact evidence

```text
implementation commit:
eb37e03c6793314e446cdb048ae9584e38f2567c

production invariant owner:
Document 62 / migration 017 ingestion_runs_lifecycle_check
```

### Final canonical status

```text
FINDING_GFA_TEST_061_TERMINAL_RUN_FIXTURE=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/70_STAGE_14_FINAL_COMPLETION_AUDIT.md
IMPLEMENTATION_COMMIT=eb37e03c6793314e446cdb048ae9584e38f2567c
```

### Prevention / future guard

Direct SQL fixtures must satisfy the current production lifecycle constraints. When constraints
become stricter, integration setup is updated rather than granted a test-only exception.

## 28. GFA-GOV-062 — Over-broad PostgreSQL fixture-parity audit rule

### Finding / symptom

The first fixture-parity source rule treated every test-local `flight_states` table as if it were a
complete production repository schema. This incorrectly flagged Data Quality migration tests,
active-aircraft metric tests and Traffic query tests that intentionally owned minimal schemas.

### Root cause

The initial static rule classified fixtures by table name alone instead of by the behavior being
exercised—specifically whether the test instantiated the full `FlightStateRepository` contract.

### Failure scenario

```text
focused test creates minimal flight_states table
↓
source audit sees table name
↓
audit demands unrelated repository evidence columns
↓
valid isolated test is rejected or padded with irrelevant schema
↓
source audit creates architectural noise instead of protecting real parity
```

### Impact

A well-intended governance check could force needless coupling into focused tests, increase fixture
maintenance, and produce false-positive CI failures.

### Severity rationale

Historical severity was not recorded. Retrospective classification: **P3 governance/test
maintainability** because the defect was in audit precision, not runtime correctness.

### Existing guarantees violated

```text
source audits must model semantic ownership rather than superficial text matches
focused tests may use purpose-built minimal schemas
full parity is required only when the full repository contract is exercised
```

### Considered solutions

1. Require every `flight_states` fixture to be full schema.
2. Delete the parity audit entirely.
3. Scope full parity to tests that create `flight_states` and instantiate `NewFlightStateRepository`.

### Chosen remediation

Option 3. Complete-repository fixtures retain strict evidence-column parity, while narrow behavior
fixtures remain intentionally minimal.

### Why this solution was selected

It preserves the useful regression guard exactly where repository writes depend on schema parity
without converting focused tests into copies of the whole production schema.

### Rejected alternatives

Global full-schema enforcement was rejected as over-coupled. Removing the audit was rejected because
real stale full-repository fixtures had already caused integration failures.

### Trade-offs

The audit is semantically more precise but slightly more complex: it must recognize repository
construction in addition to table creation.

### Regression tests / protection

`stage14finalaudit` includes fixture-classification tests covering complete repository fixtures and
intentional minimal schemas.

### Adversarial review findings

This finding itself came from adversarial review of the first remediation: the initial rule solved a
real problem too broadly. The final design preserves both strict parity and test isolation.

Historical PR/reviewer metadata is unavailable; reconstruction is limited to audit tests, fixture
source, commit `eb37e03...` and Document 70.

### Remediation iterations

The rule evolved from table-name-wide parity to repository-usage-scoped parity during the same final
audit remediation sequence.

### Residual risks and limitations

Static classification can still miss future alternate constructors or helpers. Such architectural
changes require a corresponding audit update rather than silently weakening the rule.

### Operational or deployment consequences

None. The correction affects source governance and test-maintenance scope only.

### Exact evidence

```text
implementation commit:
eb37e03c6793314e446cdb048ae9584e38f2567c

source audit:
apps/api/tools/stage14finalaudit

canonical classification:
full parity only for fixtures exercising NewFlightStateRepository
```

### Final canonical status

```text
FINDING_GFA_GOV_062_SCOPED_FIXTURE_PARITY=CLOSED
CANONICAL_FINDING_DOCUMENT=docs/70_STAGE_14_FINAL_COMPLETION_AUDIT.md
IMPLEMENTATION_COMMIT=eb37e03c6793314e446cdb048ae9584e38f2567c
```

### Prevention / future guard

Audit rules should key off semantic ownership whenever practical. A static rule that starts
flagging intentional narrow fixtures must be refined to the production contract it protects rather
than expanding test schemas mechanically.
