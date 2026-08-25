# Document 41 — Stage 14.1 Architecture Consolidation Foundation

Status: Remediation History v1.1
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

<!-- STAGE-14-2-DEAD-CODE-CLASSIFICATION:DOCUMENT-41 -->

## Stage 14.2 Continuation

Document 42 applies the reachability evidence created here. It removes confirmed dead packages and converts every remaining non-runtime package from an unowned review item into an explicit release disposition.

## 12. Canonical finding decomposition

Stage 14.1 established several foundation guards in one increment. The modern Finding Register tracks the materially distinct problems separately:

```text
GFA-ARCH-031      duplicated cross-domain confidence vocabulary
GFA-CONTRACT-032  Go/TypeScript/runtime trajectory contract drift
GFA-REL-033       production reachability evidence gap
GFA-OPS-034       frontend package-manager execution instability
```

The security boundary identified in this document is remediated and registered separately by Document 45 rather than being duplicated here.

## 13. GFA-ARCH-031 — Duplicated cross-domain confidence vocabulary

### Finding / symptom

Several bounded contexts carried their own copies of the same ordinal confidence vocabulary (`none`, `low`, `medium`, `high`).

### Root cause

Confidence semantics evolved independently inside analytical modules. Although the surrounding confidence assessments remained domain-specific, the shared ordinal labels had no single owner.

### Failure scenario

One context adds, removes, renames, or reorders an ordinal value while another keeps the old set. Serialization, validation, comparisons, or frontend unions can then disagree even though both sides intend to express the same ordinal vocabulary.

### Impact

The immediate risk is semantic drift and duplicated maintenance rather than known production data corruption. Divergent vocabularies also make contract audits less reliable because equivalent concepts no longer share one canonical source.

### Severity rationale

**P3 retrospective.** The evidence shows duplicated ownership and drift risk, not a historical user-visible correctness incident.

### Existing guarantees violated

- shared primitive vocabulary should have one canonical owner;
- bounded contexts may add context-specific meaning without redefining identical primitive labels;
- compatibility aliases must not become independent sources of truth.

### Considered solutions

1. keep duplicated enums and synchronize manually;
2. merge all confidence models into one universal cross-domain structure;
3. centralize only the ordinal vocabulary while keeping domain-specific confidence evidence local.

### Chosen remediation and why

`internal/domain/confidence` owns the ordinal vocabulary. Data Quality and Metrics retain compatibility aliases and their richer domain-specific reasons, scores, provenance, and limitations. This removes duplicate primitive truth without flattening bounded contexts.

### Rejected alternatives

Manual synchronization leaves the original drift mechanism intact. A universal confidence result type was rejected because equal ordinal labels do not imply identical evidence semantics across analytical contexts.

### Trade-offs

Contexts gain a small dependency on a shared domain primitive. The dependency is intentionally narrow and avoids a larger shared-model package.

### Regression protection

Architecture audits detect duplicate vocabulary ownership and preserve the shared source. Existing domain tests continue to validate context-specific semantics.

### Adversarial review findings and remediation iterations

The consolidation explicitly rejected the tempting overcorrection of merging all confidence assessments. The final design centralizes only what is genuinely identical.

### Residual risk / limitations

A shared ordinal vocabulary does not calibrate confidence values across domains. `high` in one analytical context is not automatically numerically comparable with `high` in another.

### Operational / deployment consequences

None. This is an internal contract consolidation with compatibility aliases.

### Exact evidence

Implementation commit: `fc6c3dbafa302d061653587163457d72f08c7a77` (`refactor: establish architecture consolidation gates`). Historical PR/reviewer metadata is not asserted where it cannot be recovered reliably.

### Final canonical status

**CLOSED.**

### Prevention / future guard

Before adding a new shared-looking enum/value set, determine whether the primitive vocabulary is actually identical across contexts. Centralize the primitive only when justified; keep context-specific evidence and policy local.

## 14. GFA-CONTRACT-032 — Go/TypeScript/runtime trajectory contract drift

### Finding / symptom

Trajectory identity fields and enums could evolve in Go without equivalent updates to the TypeScript public interface or runtime parser. The contract initially lacked `identity_key`, `identity_basis`, and `split_reason` alignment across the complete response chain.

### Root cause

Compilation validates each language separately. No permanent cross-language audit proved that Go domain/DTO fields, TypeScript types, and runtime parsing rules represented the same trajectory contract.

### Failure scenario

The backend emits a new required field or enum value. TypeScript interfaces or the runtime parser remain stale. Depending on the mismatch, the frontend can reject valid backend responses, silently omit identity evidence, or accept a type declaration that the runtime parser does not actually enforce.

### Impact

Cross-stack contract drift can break production reads or, more subtly, lose trajectory identity/split evidence in the UI. The problem is especially important for fields that explain why a trajectory was segmented or how its identity was derived.

### Severity rationale

**P2 retrospective.** This is a production API compatibility/correctness risk. The evidence does not establish a historical outage, so it is not promoted to P1.

### Existing guarantees violated

- public backend and frontend contracts must agree on fields and enums;
- runtime parsing must enforce the same contract TypeScript claims statically;
- identity/explainability fields must not disappear between layers.

### Considered solutions

1. rely on manual review and TypeScript compilation;
2. generate the entire frontend client from one schema immediately;
3. add targeted Go↔TypeScript contract auditing plus runtime-parser coverage for the trajectory response chain.

### Chosen remediation and why

The project audit compares Go DTOs/enums with TypeScript interfaces and validation sets. Runtime parsing is included, so a type-only update cannot mask a stale parser. The audit checks semantic field/value ownership rather than one literal import formatting style.

### Rejected alternatives

Manual review cannot guarantee future synchronization. Full client/schema generation was rejected as disproportionate to the immediate drift problem and would introduce a larger tooling migration.

### Trade-offs

The audit contains cross-language contract knowledge that must evolve with deliberate API changes. This maintenance cost is accepted because it makes drift fail before merge.

### Regression protection

Strict project architecture audit checks trajectory DTO fields, enum sets, TypeScript interface fields, and runtime validation behavior. TypeScript compilation independently validates imported names.

### Adversarial review findings and remediation iterations

The first audit approach risked coupling correctness to exact TypeScript import syntax. A follow-up hardened it to inspect validation sets and parsed fields instead, avoiding false positives caused by formatting/import style while keeping semantic enforcement.

### Residual risk / limitations

The targeted audit protects the trajectory contract, not every API in the repository. Other cross-language contracts require their own generated or explicit guards.

### Operational / deployment consequences

None at runtime. CI becomes stricter when backend/frontend trajectory contracts change.

### Exact evidence

Foundation implementation: `fc6c3dbafa302d061653587163457d72f08c7a77`. The same historical document records follow-up runtime-parser and syntax-independence hardening. Exact historical review comments are not reconstructed without evidence.

### Final canonical status

**CLOSED for the protected trajectory contract.**

### Prevention / future guard

Cross-language contracts must be protected at all three layers that matter: backend DTO/domain values, frontend static types, and runtime decoding. Audits should verify semantics, not formatting trivia.

## 15. GFA-REL-033 — Production reachability evidence gap

### Finding / symptom

A package could compile and have tests while still being unreachable from every production or operational runtime root. Conversely, lack of `cmd/server` reachability could be misread as proof that code was dead.

### Root cause

The repository previously lacked a formal runtime-root dependency-graph classification. Compilation/test success and folder naming were being asked to answer a different question: whether code actually participates in production, verification, research, or no supported path.

### Failure scenario

A fully tested analytical package is described as a production capability even though no executable imports it; or a legitimate offline/reconciliation/verification package is deleted merely because `cmd/server` does not import it.

### Impact

Release claims can overstate implemented capabilities, while cleanup can remove valid non-server tooling. Both outcomes weaken architecture truth and make dead-code decisions subjective.

### Severity rationale

**P2 retrospective.** This is release/architecture correctness: it can produce false product claims or unsafe deletion decisions, but it is not direct persisted-data corruption.

### Existing guarantees violated

- production capability claims require runtime reachability evidence;
- compilation/tests do not prove operational integration;
- non-runtime code requires explicit disposition before release or deletion.

### Considered solutions

1. classify by directory/file names;
2. treat anything compiling as required production code;
3. compute dependency reachability from explicit runtime roots and classify non-runtime packages by supported purpose.

### Chosen remediation and why

`go list`-based auditing counts runtime-reachable, verification-only, and non-runtime packages from named roots (`cmd/server`, `cmd/ingest`, `cmd/reconcile`, materializers). Non-runtime analytical packages require an explicit disposition instead of automatic deletion.

### Rejected alternatives

Folder heuristics are not dependency evidence. Compilation-only classification cannot distinguish unused code from integrated code.

### Trade-offs

The project must maintain the list of supported runtime/verification roots and dispositions. This explicit bookkeeping is preferable to unowned analytical package trees.

### Regression protection

Strict project audit fails when required analytical contexts have no runtime reachability or when unknown non-runtime packages lack a disposition. Document 42 extends the rule into concrete deletion/classification.

### Adversarial review findings and remediation iterations

The design explicitly resists the common cleanup shortcut "not reachable from server = dead." Later stages use the evidence to integrate Airport Intelligence, Feature Pipeline, Transponder Evidence, and offline formula evaluation rather than deleting them mechanically.

### Residual risk / limitations

Static import reachability proves dependency inclusion, not that every route/job is exercised successfully in production. Runtime integration tests and deployment evidence remain separate requirements.

### Operational / deployment consequences

Release audits gain a stricter evidence boundary; unresolved `planned_production_integration`/equivalent dispositions become visible blockers rather than hidden code.

### Exact evidence

Implementation commit: `fc6c3dbafa302d061653587163457d72f08c7a77`. Follow-up classification/removal is in Document 42, commit `8bcc73ad1281d468fc17dc9f0628d54f79d7e2b0`.

### Final canonical status

**CLOSED as a governance/reachability foundation; individual unresolved packages were handled by later stage findings.**

### Prevention / future guard

Every new executable/research/verification package tree must have a named root or explicit non-runtime disposition. Product/release claims should cite runtime reachability plus behavioral evidence, not compilation alone.

## 16. GFA-OPS-034 — Frontend package-manager execution instability

### Finding / symptom

Local/CI validation depended on an experimental Corepack proxy path that could be broken by a globally cached shim even though the repository pinned pnpm 11.8.0.

### Root cause

The repository version contract and the actual package-manager invocation path were not the same thing. A host-level Corepack failure could prevent deterministic validation before project code was even evaluated.

### Failure scenario

A developer or CI runner has a broken/stale Corepack shim. Commands fail to resolve the pinned pnpm despite the repository declaring the correct version, producing environment failures unrelated to application code.

### Impact

Build/validation reliability suffers and dependency evidence becomes host-tool dependent.

### Severity rationale

**P3 retrospective.** This is developer/CI tooling reliability, not production runtime correctness.

### Existing guarantees violated

- repository-pinned tooling should be reproducibly invokable;
- CI should not depend on an experimental host proxy when an official setup action exists;
- local fallback must preserve the pinned version.

### Considered solutions

1. keep Corepack and document cache cleanup;
2. globally install an unpinned pnpm;
3. use official pnpm setup in CI and a pinned installed/npm-exec fallback locally.

### Chosen remediation and why

CI uses `pnpm/action-setup`; local verification uses pnpm 11.8.0 when present or `npm exec` for exactly 11.8.0. The root `packageManager` pin remains authoritative.

### Rejected alternatives

Manual Corepack cache repair is host-specific and fragile. An unpinned global install would trade one reproducibility problem for another.

### Trade-offs

There are two supported invocation mechanisms, but both enforce the same version. This small complexity isolates the project from Corepack shim state.

### Regression protection

Frontend CI and local verification commands exercise the pinned setup path. Dependency policy checks continue to verify lockfile/tooling contracts.

### Adversarial review findings and remediation iterations

The fix distinguishes repository version ownership from transport/tool invocation, avoiding a workaround that silently changes package-manager version.

### Residual risk / limitations

Node/npm availability is still required for the fallback path; package registry/network availability may matter when the pinned pnpm is not already installed.

### Operational / deployment consequences

None for deployed application behavior. CI/developer setup becomes more reproducible.

### Exact evidence

Recorded in the Stage 14.1 historical document following foundation commit `fc6c3dbafa302d061653587163457d72f08c7a77`. A separate exact historical follow-up commit is not asserted without recoverable evidence.

### Final canonical status

**CLOSED.**

### Prevention / future guard

Toolchain governance must distinguish the pinned project version from host-level launcher mechanisms. CI should prefer official deterministic setup actions and local fallbacks must preserve the exact repository pin.
