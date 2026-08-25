# Document 100 — Query and Architecture Consolidation

Status: Implemented Engineering Increment v1.0
Project: Global Flight Analytics
Baseline: `a31fd8ce3fb6f42a9c90a5153f902c37e7b0f111`

## 1. Purpose

This increment closes the accepted Analytical Core findings concerning zero
query reference time, non-canonical UUID identifiers, inconsistent Metric IDs,
stored legacy calculator state, exposed executor internals, and the concrete
executor dependency owned by the metric service.

## 2. Query contracts

`metricquery.RecentRequest.Normalize` now rejects a zero reference time with
`ErrReferenceTimeRequired`.

`NormalizeTrajectoryIDs` now:

```text
trims input;
parses each UUID;
converts it to the canonical lowercase hyphenated representation;
deduplicates by the canonical representation;
returns only canonical identifiers.
```

Equivalent UUID spellings can no longer reach the repository as distinct IDs.

## 3. Canonical Metric IDs

The `metrics` package is the authoritative owner of these identifiers:

```text
traffic.active_aircraft
traffic.traffic_density
traffic.airport_activity
traffic.coverage_score
traffic.data_freshness
```

The `metricexecution` package exposes compatibility aliases to the same
constants and uses those aliases for every execution result.

## 4. Executor boundary

The legacy `calculator` constructor argument remains accepted to avoid an
unnecessary breaking change, but it is no longer retained by `Executor`.

The following internal dependency getters are removed:

```text
Executor.Calculator
Executor.ScopeGuard
Executor.ConfidenceEvaluator
Service.Executor
```

Confidence evaluation is exposed as the narrow behavior
`Executor.EvaluateConfidence`, rather than exposing the evaluator object.

`metricexecution.Service` now depends on an unexported behavioral interface
containing only:

```text
FilterTrajectories
EvaluateConfidence
```

## 5. Legacy package classification

`analytics/calculator` and `analytics/registry` remain compiled and tested as
compatibility foundation packages. They are not composed by the production
server, not stored by the runtime executor, and not exposed through the metric
service.

A later breaking-version cleanup may remove their compatibility constructor
type completely. That deletion is not required for the current runtime
architecture.

## 6. Permanent regression coverage

The increment adds tests for:

```text
zero reference-time rejection;
canonical UUID output and canonical deduplication;
canonical Metric IDs and execution aliases;
absence of stored calculator state;
absence of internal dependency getter methods;
narrow executor behavior.
```

## 7. Verification

Before changing `main`, the installer validates all patch anchors and symbol
usage, applies the patch in a detached temporary Git worktree, and runs:

```text
complete backend compilation;
targeted tests;
complete backend tests.
```

The working tree then runs targeted compilation, targeted tests, race tests,
the complete backend test suite, Go vet, architecture audits, static contracts,
documentation checks, and whitespace validation.

## 8. Remaining Analytical Core review scope

```text
replace request-parameter Coverage Score and Data Freshness endpoints with
server-owned production snapshots or explicitly separated calculator routes;

perform the final Analytical Core closure audit and evidence register.
```
<!-- ANALYTICAL-CORE-REVIEW-CLOSURE:QUERY-ARCHITECTURE -->

## 9. Post-closure resolution

Document 102 completes the original Analytical Core finding register.

```text
ANALYTICAL_CORE_REVIEW_STATUS=CLOSED
Open Query and Architecture findings: 0
```

The compatibility calculator and registry packages remain classified exactly as
recorded in Section 5. They do not form a production runtime architecture.

---

## Canonical remediation history

### GFA-DATA-106 / AC-10 — analytical recent queries accepted a zero reference time

1. **Finding / symptom.** `metricquery.RecentRequest.Normalize` could accept a zero `ReferenceTime` and derive a recent query window from an undefined temporal anchor.
2. **Root cause.** Query normalization validated window/limit structure but treated Go's zero `time.Time` as an ordinary timestamp instead of a missing required field.
3. **Failure scenario.** An internal or future HTTP caller omits the analytical reference time; normalization constructs a window near year one rather than failing, causing empty/misleading repository reads and non-obvious analytical output.
4. **Impact.** Time-window semantics become invalid and failures manifest as misleading no-data results instead of explicit invalid-request evidence.
5. **Severity rationale.** **P2 retrospective.** The defect corrupts query semantics but requires a missing/invalid caller reference time rather than ordinary valid traffic evidence.
6. **Existing guarantees violated.** Every recent analytical query must have an explicit evaluation/reference time; normalization must fail closed on missing temporal ownership.
7. **Considered solutions.** Treat zero as `time.Now`; allow repositories to return empty results; require non-zero reference time during normalization; add a pointer type solely for presence detection.
8. **Chosen remediation.** `RecentRequest.Normalize` rejects zero time with `ErrReferenceTimeRequired` before constructing query bounds.
9. **Why this solution was selected.** It keeps time ownership explicit and deterministic without introducing wall-clock behavior into a pure normalization boundary.
10. **Rejected alternatives.** Implicit `time.Now` makes tests/replay nondeterministic; empty reads hide invalid input; a pointer refactor is unnecessary when zero time already represents missing input in the current API.
11. **Trade-offs.** Callers must always supply a reference time; this is an intentional requirement for reproducible analytics.
12. **Regression tests / protection.** Query normalization tests require zero-time rejection; the Analytical Core final audit checks the contract.
13. **Adversarial review findings.** The reference time must be validated before subtracting the window so invalid temporal arithmetic cannot masquerade as a valid normalized range.
14. **Remediation iterations.** The fix was implemented in the shared query normalizer so all consumers inherit one fail-closed rule.
15. **Residual risks and limitations.** A non-zero but incorrect caller time is still syntactically valid; production handlers own the choice of server evaluation time.
16. **Operational or deployment consequences.** None. Invalid callers now receive a controlled error instead of a misleading empty analytical result.
17. **Exact evidence.** Historical implementation commit `b8ccbf590ef3b9ffc221d72e0274e1d78da6c650` (`refactor: consolidate analytical query architecture`). Original review ID: `AC-10`. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-DATA-106=CLOSED`.
19. **Prevention / future guard.** Query value objects with required temporal anchors must reject zero/missing time during normalization and must not synthesize wall-clock defaults silently.

### GFA-DATA-107 / AC-11 — accepted UUID identifiers were not canonicalized before deduplication

1. **Finding / symptom.** Trajectory UUID strings could be parsed as valid but remain in different equivalent textual spellings, allowing semantically identical identifiers to survive string-level deduplication as distinct values.
2. **Root cause.** Validation and deduplication operated on trimmed input text rather than on the canonical representation of the parsed UUID value.
3. **Failure scenario.** The same UUID appears using uppercase or another accepted spelling; both values reach repository membership queries, causing duplicate query inputs and inconsistent identity evidence.
4. **Impact.** Query identity sets are not canonical, unnecessary repository work is performed, and downstream assumptions that one UUID value has one textual representation can fail.
5. **Severity rationale.** **P2 retrospective.** This is an identity-normalization correctness defect with bounded query impact; it does not by itself mutate persisted data.
6. **Existing guarantees violated.** Accepted identifiers must be canonical before equality/deduplication; repository boundaries should receive normalized domain identities.
7. **Considered solutions.** Deduplicate raw strings case-insensitively; let PostgreSQL UUID casts normalize them; parse and reserialize each UUID canonically before deduplication; accept duplicates as harmless.
8. **Chosen remediation.** `NormalizeTrajectoryIDs` trims, parses, serializes to canonical lowercase hyphenated UUID form, deduplicates canonical values and returns only canonical identifiers.
9. **Why this solution was selected.** Parsing gives semantic identity rather than text heuristics and keeps repository behavior independent of database coercion.
10. **Rejected alternatives.** Case-insensitive text comparison does not define all accepted UUID spelling equivalence; database normalization occurs too late; duplicate inputs waste work and weaken contract clarity.
11. **Trade-offs.** Returned identifiers may differ textually from caller input even when valid. That is intentional canonicalization.
12. **Regression tests / protection.** Tests cover canonical output and deduplication of equivalent UUID spellings; the final source audit preserves canonical UUID normalization.
13. **Adversarial review findings.** Deduplication must happen after parsing/reserialization, not merely after trimming; invalid UUIDs must still fail rather than be carried through normalization.
14. **Remediation iterations.** UUID ownership was consolidated with query normalization instead of relying on repository-specific casts.
15. **Residual risks and limitations.** Canonical UUID normalization does not prove the referenced trajectory exists; repository lookup owns existence.
16. **Operational or deployment consequences.** None; some duplicate query inputs disappear before database access.
17. **Exact evidence.** Historical implementation commit `b8ccbf590ef3b9ffc221d72e0274e1d78da6c650`. Original review ID: `AC-11`. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-DATA-107=CLOSED`.
19. **Prevention / future guard.** New UUID-list query contracts must parse to UUID values and deduplicate canonical identities before repository execution.

### GFA-CONTRACT-108 / AC-17 — analytical Metric IDs were inconsistent across packages and clients

1. **Finding / symptom.** Metric identifiers were owned in multiple packages and could diverge between domain metric implementations, execution results and frontend consumers.
2. **Root cause.** The analytical stack duplicated string constants instead of designating one package as the authoritative Metric ID namespace.
3. **Failure scenario.** A metric implementation changes one identifier while execution or frontend code retains an older string; results, query keys, UI routing or downstream comparisons refer to different logical metric IDs.
4. **Impact.** Cross-layer contracts drift despite successful local compilation, making analytical results harder to correlate and potentially breaking consumers silently.
5. **Severity rationale.** **P2 retrospective.** This is a public/cross-package contract-integrity defect with real integration risk but no direct data corruption.
6. **Existing guarantees violated.** One logical metric must have one stable identifier across domain, execution and client layers; aliases may exist only when they resolve to the same canonical owner.
7. **Considered solutions.** Keep duplicate constants with tests; generate IDs from names; use `metrics` as the canonical owner with execution compatibility aliases; move strings into a generic shared package detached from metric implementations.
8. **Chosen remediation.** The `metrics` package owns all `traffic.*` identifiers and `metricexecution` exposes aliases to those exact constants; execution results and frontend contracts use the canonical namespace.
9. **Why this solution was selected.** Ownership stays next to the metric domain while compatibility aliases avoid unnecessary breaking churn.
10. **Rejected alternatives.** Duplicate constants remain drift-prone; generated names can change through display-label edits; a generic shared package would add indirection without stronger ownership.
11. **Trade-offs.** `metricexecution` depends on metric identifier constants, but not on concrete calculation internals; this narrow dependency is intentional.
12. **Regression tests / protection.** `TestMetricIDsUseOneCanonicalNamespace` compares execution aliases, metric constants and `ID()` methods. Frontend reconciliation and the Analytical Core final audit enforce the same `traffic.*` values.
13. **Adversarial review findings.** Compatibility aliases are acceptable only if they are direct references, not copied strings; tests must cover all metric IDs rather than one representative value.
14. **Remediation iterations.** Backend constants were consolidated first; Document 101 reconciled frontend analytical request/types/query consumers to the same namespace.
15. **Residual risks and limitations.** External consumers outside the repository must still treat Metric IDs as API contract values and version breaking changes deliberately.
16. **Operational or deployment consequences.** None. Existing canonical `traffic.*` identifiers remain stable.
17. **Exact evidence.** Historical implementation commit `b8ccbf590ef3b9ffc221d72e0274e1d78da6c650`. Original review ID: `AC-17`. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-CONTRACT-108=CLOSED`.
19. **Prevention / future guard.** A new metric ID must be defined once in the canonical metrics owner and referenced by aliases/clients; copied literal identifiers should be rejected by contract tests or the final audit.

### GFA-ARCH-109 / AC-06 — metric execution exposed concrete dependencies and retained legacy calculator state

1. **Finding / symptom.** Runtime `Executor` retained the legacy calculator object, exposed concrete dependency getters, and `metricexecution.Service` depended on the concrete executor rather than the narrow behavior it required.
2. **Root cause.** Compatibility construction evolved into runtime dependency exposure, allowing callers/tests to reach internal objects instead of depending on analytical behavior.
3. **Failure scenario.** New code couples to `Executor.Calculator`, `ScopeGuard`, `ConfidenceEvaluator` or `Service.Executor`, making internal refactors breaking changes and encouraging a parallel calculator/registry runtime path.
4. **Impact.** Architectural boundaries weaken, implementation objects escape their owners, and the Analytical Core becomes harder to change/test without expanding dependency surface.
5. **Severity rationale.** **P3 retrospective.** The existing runtime behavior was largely correct; the defect is maintainability and architecture coupling rather than an observed wrong metric.
6. **Existing guarantees violated.** Services should depend on required behavior, not concrete implementation objects; compatibility packages must not become an accidental second runtime architecture.
7. **Considered solutions.** Remove the compatibility constructor entirely; retain all getters for tests; keep the constructor argument but do not store it and narrow runtime behavior; introduce a broad public executor interface.
8. **Chosen remediation.** The legacy calculator constructor argument remains accepted but is ignored/not retained; concrete dependency getters are removed; confidence is exposed as behavior; `metricexecution.Service` depends on an unexported interface containing only `FilterTrajectories` and `EvaluateConfidence`.
9. **Why this solution was selected.** It closes runtime coupling without forcing a breaking-version cleanup of compatibility constructors that are not part of production composition.
10. **Rejected alternatives.** Immediate deletion creates avoidable churn; getters perpetuate coupling; a broad public interface merely relocates the same oversized dependency surface.
11. **Trade-offs.** Some tests use internal package access to verify configured dependencies, while external callers no longer receive concrete objects. Compatibility packages remain compiled until a deliberate breaking cleanup.
12. **Regression tests / protection.** Reflection tests assert absence of stored calculator state and dependency getter methods; service tests use narrow behavior; the Analytical Core final audit verifies compatibility-package runtime isolation.
13. **Adversarial review findings.** Accepting the legacy constructor parameter is safe only while it is provably not stored or used by runtime execution; compatibility source presence must not be confused with production composition reachability.
14. **Remediation iterations.** Concrete getters were replaced with `EvaluateConfidence`; the service executor dependency became an unexported behavioral interface; runtime calculator storage was removed while constructor compatibility was retained.
15. **Residual risks and limitations.** The compatibility constructor/type remains source debt for a future breaking release, but it is not part of production runtime state or composition.
16. **Operational or deployment consequences.** None.
17. **Exact evidence.** Historical implementation commit `b8ccbf590ef3b9ffc221d72e0274e1d78da6c650`. Original review ID: `AC-06`. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-ARCH-109=CLOSED`.
19. **Prevention / future guard.** Runtime services must expose behavior rather than internal dependency objects; compatibility constructors may remain only when tests prove they do not recreate a parallel production dependency graph.

### AC-05 deliberately retained classification

The original review observation that `analytics/calculator` and `analytics/registry` appeared to form a parallel runtime architecture is **not** registered as a defect. Repository evidence shows they remain compatibility foundation packages that are compiled/tested but not composed by the production server, not retained by runtime `Executor`, and not required by the metric service interface. Document 102 therefore classifies `AC-05=DELIBERATELY_RETAINED`, not `FIXED`, `DEFERRED` or `ACCEPTED_RISK`; no synthetic GFA finding ID is created.