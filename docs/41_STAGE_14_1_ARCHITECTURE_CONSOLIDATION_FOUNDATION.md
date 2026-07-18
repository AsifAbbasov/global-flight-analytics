# Document 41 — Stage 14.1 Architecture Consolidation Foundation

Status: Implementation Baseline v1.0
Project: Global Flight Analytics
Scope: shared confidence vocabulary, contract drift prevention, production reachability evidence, and supply-chain quality gates

## 1. Purpose

Stage 14 starts a backend feature freeze and architectural consolidation.

This increment fixes confirmed defects without deleting packages on the basis of folder names or file counts alone.

## 2. Shared Confidence Vocabulary

The ordinal values:

```text
none
low
medium
high
```

now have one source of truth:

```text
internal/domain/confidence
```

Data Quality and Metrics retain context-local compatibility aliases. Their surrounding structures continue to define the meaning, reasons, score, provenance, and limitations of confidence.

The shared package does not merge domain-specific confidence assessments into one universal result type.

## 3. Go and TypeScript Contract Drift

The trajectory contract now includes:

```text
identity_key
identity_basis
split_reason
```

The TypeScript unions are checked against Go domain enums.

The project audit compares:

```text
Go DTO Trajectory              ↔ TypeScript AircraftTrajectory
Go DTO TrajectorySegment       ↔ TypeScript TrajectorySegment
Go DTO CoverageGap             ↔ TypeScript CoverageGap
```

Field additions, removals, type changes, and enum drift fail the audit.

## 4. Production Reachability

The consolidated project audit uses `go list` rather than folder-name assumptions.

Runtime roots:

```text
cmd/server
cmd/ingest
cmd/reconcile
cmd/materialize-historical-intelligence
```

Analytical contexts receive factual counts for:

```text
runtime reachable
verification only
not reachable from runtime roots
```

A package marked `REVIEW_NOT_RUNTIME_REACHABLE` is not automatically deleted. It must be classified as one of:

```text
offline research
verification support
test support
obsolete implementation
unfinished feature
genuinely dead code
```

Required analytical contexts fail strict audit when none of their packages are reachable from a runtime root.

## 5. Compilation and Behavioral Evidence

`go test ./...` compiles every Go package included by the module and runs all tests.

This proves compilation and test execution. It does not by itself prove scientific calibration or operational aviation correctness.

The reachability audit proves whether analytical packages participate in runtime dependency graphs.

Existing PostgreSQL and HTTP runtime verifiers remain behavioral evidence for their respective completed stages.

## 6. Supply-Chain Gates

Backend continuous integration now runs pinned `govulncheck`.

Frontend continuous integration runs a production dependency audit.

Dependabot monitors:

```text
Go modules
pnpm/npm dependencies
Docker base images
GitHub Actions
```

The existing backend container job already builds the Dockerfile, verifies the non-root runtime user, starts the image, waits for health, and calls the health endpoint.

## 7. Authentication Boundary

Public read-only endpoints continue to expose open research data without user accounts.

Authentication is not coupled to frontend styling.

Any route that triggers computation, persistence, administration, or private user data must be protected before deployment. The existing route that processes Route Intelligence through HTTP is part of the next security consolidation slice.

## 8. Non-Goals

This increment does not:

```text
delete packages solely because they are not reached by cmd/server
claim that all analytical formulas are calibrated
introduce user accounts
merge all bounded contexts into a shared package
generate every frontend contract
replace runtime integration tests
```

<!-- STAGE-14-1-FRONTEND-TOOLCHAIN-FIX -->

## 9. Frontend Package Manager Execution

Local and continuous integration validation no longer depends on the experimental Corepack proxy.

GitHub Actions installs the repository-pinned pnpm version through the official `pnpm/action-setup` action.

The local verification command first uses an already installed pnpm 11.8.0 binary when available. Otherwise it invokes pnpm 11.8.0 through `npm exec`. This bypasses a broken globally cached Corepack shim while preserving the version fixed in the root `packageManager` field.

<!-- STAGE-14-1-TRAJECTORY-RUNTIME-PARSER-FIX -->

## 10. Runtime Parser Contract

The contract drift gate now covers the complete trajectory response chain:

```text
Go domain enums
Go HTTP DTO
TypeScript public interface
TypeScript runtime parser
```

The frontend parser validates `identity_basis` and `split_reason` against the
same value sets exported by the Go domain. It also requires `identity_key`.

A future change that updates only the DTO or only the TypeScript interface
will fail the project architecture audit before merge.

<!-- STAGE-14-1-AUDIT-FALSE-POSITIVE-FIX -->

## 11. TypeScript Import Syntax Independence

The runtime parser audit does not depend on a literal import formatting style.
The audit verifies runtime validation sets and parsed response fields. TypeScript
compilation independently verifies that imported type names exist and are used.
