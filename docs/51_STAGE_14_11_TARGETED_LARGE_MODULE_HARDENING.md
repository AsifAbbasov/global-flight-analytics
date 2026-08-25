# Document 51 — Stage 14.11 Targeted Large-Module Hardening

Status: Remediation History v1.1
Project: Global Flight Analytics
Scope: confirmed cohesion and orchestration problems in four backend modules

## 1. Audit Scope

The audit examined:

```text
Historical Intelligence contract validation
Route Intelligence contract validation
Historical-neighbor projection continuation
Estimated-arrival projection
```

The audit did not use a repository-wide line-count rule.

Changes were accepted only where a source unit mixed multiple independent
responsibilities or one public operation coordinated detailed computation
directly.

## 2. Confirmed Findings

The Historical Intelligence validation source contained contract identity,
scope, time, series, summary, comparison, confidence, limitation, provenance,
and numerical helper rules in one source file.

The Route Intelligence validation source contained identity, window, endpoint,
airport, evidence, assessment, confidence, limitation, provenance, and
numerical helper rules in one source file.

The historical-neighbor continuation `Project` method directly performed:

```text
horizon planning
neighbor selection
pattern confidence evaluation
current endpoint preparation
candidate indexing
per-horizon neighbor translation
sample combination
limitation construction
result contract assembly
fallback selection
```

The estimated-arrival `Estimate` and `computeArrival` methods directly
performed:

```text
input contract validation
cross-contract identity validation
future-evidence prevention
destination eligibility
position sample construction
arrival-radius crossing interpolation
inside-radius handling
bounded extrapolation
confidence construction
provenance construction
result mutation
```

These are cohesion and test-isolation findings, not merely large line counts.

## 3. Validation Decomposition

The two contract validation sources are split by top-level responsibility.

Historical validation becomes:

```text
core
identifiers
identity
scope and time
series
summary
evidence and provenance
```

Route validation becomes:

```text
core
identifiers
identity and window
endpoint and evidence
assessment
provenance and support
```

The public `Validate` functions, issue codes, severity, ordering, and contract
behavior remain unchanged.

No validation rule is deleted or weakened.

## 4. Projection Continuation Decomposition

`Project` remains the public operation and now coordinates:

```text
request identity validation
horizon planning
preparation
forecast point production
fallback dispatch
result construction
contract validation
```

Detailed work is isolated into:

```text
continuation preparation
neighbor sample translation
forecast point production
sample combination
fallback handling
result assembly
geometry and interpolation
evidence and provenance
fingerprinting
```

Fallback reason codes and conservative kinematic fallback behavior remain
unchanged.

## 5. Estimated-Arrival Decomposition

`Estimate` remains the public operation and now coordinates:

```text
request validation
availability gates
position sample selection
arrival computation
result attachment
```

Arrival computation is separated into:

```text
distance calculation
arrival-radius crossing
already-inside-radius handling
bounded post-horizon extrapolation
```

Confidence, limitations, provenance, and unavailable-result behavior remain
separate from geometric arrival computation.

## 6. Intentionally Rejected Changes

The increment does not:

```text
change analytical formulas
change confidence weights
change projection thresholds
change route validation semantics
change Historical Intelligence validation issue codes
replace optional floating point values
replace domain-state booleans
rename functions solely because they contain With or And
add a dependency injection framework
change HTTP contracts
change SQL or migrations
change provider behavior
change frontend behavior
```

## 7. Regression Gates

Automated architecture tests require:

```text
the four former monolithic source files to be absent
responsibility-specific validation files to exist
targeted production source files to remain below the bounded audit size
Project to remain a narrow coordinator
Estimate to remain a narrow coordinator
computeArrival to remain a narrow dispatcher
```

Existing package tests remain the primary behavioral regression evidence.

## 8. Acceptance

The increment is accepted only after:

```text
source transformation verification
focused package tests
focused architecture tests
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

## 9. Canonical finding decomposition

```text
GFA-MAINT-046  Historical Intelligence validation responsibility concentration
GFA-MAINT-047  Route Intelligence validation responsibility concentration
GFA-MAINT-048  Historical-neighbor Project orchestration concentration
GFA-MAINT-049  Estimated-arrival orchestration/computation concentration
```

All four are **P3 retrospective**: the historical evidence explicitly states that formulas, issue codes, thresholds, HTTP contracts, persistence, and provider behavior were preserved. These were cohesion/test-isolation findings, not discovered production miscalculations.

## 10. GFA-MAINT-046 — Historical Intelligence validation responsibility concentration

### Finding / symptom

One Historical Intelligence validation source owned identity, scope/time, series, summary/comparison, confidence/limitations, provenance, identifiers, and numeric helpers.

### Root cause

Validation rules accumulated as the Historical contract expanded, without recutting source ownership around distinct rule families.

### Failure scenario

A change to one validation concern requires editing a file containing many unrelated rule groups, making review/test isolation harder and increasing the chance of accidentally changing issue ordering, severity, or another contract rule.

### Impact

Maintainability and review precision degrade; future correctness changes become harder to localize.

### Severity rationale

**P3 retrospective.** Behavior was already correct; the defect was cohesion/test-isolation.

### Existing guarantees violated

Validation ownership should reflect top-level contract responsibilities, while public `Validate`, issue codes/severity/order remain stable.

### Considered solutions

Keep the file with comments/regions; introduce a generic validation framework; split rule families into responsibility-specific same-package files.

### Chosen remediation / why selected

Rules were split into core, identifiers, identity, scope/time, series, summary, and evidence/provenance owners while preserving the public validator. This reduces coupled edits without adding a framework.

### Rejected alternatives

Comments do not reduce source coupling. A generic rule engine would add indirection without a need for dynamic validation composition.

### Trade-offs

More files and navigation, but smaller review units and clearer rule ownership.

### Regression tests / protection

Architecture tests require the former monolith absent and responsibility files present; existing contract tests protect issue semantics and ordering.

### Adversarial review findings

The audit rejected line-count-only refactoring and required evidence of independent responsibilities; no rule was deleted merely to shrink files.

### Remediation iterations

One targeted Stage 14.11 decomposition after the architecture audit identified the concrete cohesion problem.

### Residual risks / limitations

Rule families can still become internally large; future splits require actual independent concerns rather than numeric thresholds alone.

### Operational / deployment consequences

None.

### Exact evidence

Implementation commit: `d1fc34b6f25b5d7e8c18ac287709241e42617000` (`refactor: harden large backend modules`). Historical PR/reviewer metadata unavailable unless recoverable from repository evidence.

### Final canonical status

**CLOSED.**

### Prevention / future guard

Validation sources should be reviewed when unrelated rule families begin changing together; preserve public semantics and split only on meaningful responsibilities.

## 11. GFA-MAINT-047 — Route Intelligence validation responsibility concentration

### Finding / symptom

One Route Intelligence validation source owned identity/window, endpoint/airport/evidence, assessment/confidence, limitations/provenance, identifiers, and numeric support.

### Root cause

Contract growth accumulated independent validation concerns in one source unit.

### Failure scenario

A change to endpoint evidence validation shares a diff with unrelated assessment/provenance/identity rules, increasing review blast radius and making regression localization harder.

### Impact

Maintainability and test isolation suffer; future Route contract changes become unnecessarily risky.

### Severity rationale

**P3 retrospective.** No historical route-validation behavior defect is asserted.

### Existing guarantees violated

Distinct validation concerns should have explicit owners while `Validate`, issue codes, severities, and ordering stay stable.

### Considered solutions

Comments/regions; generic framework; same-package responsibility split.

### Chosen remediation / why selected

Route validation was decomposed into core, identifiers, identity/window, endpoint/evidence, assessment, and provenance/support owners. This preserves compile-time behavior and avoids new abstractions.

### Rejected alternatives

Comments retain coupled source ownership. A generic framework would obscure domain-specific rule ordering and add unnecessary machinery.

### Trade-offs

Additional files versus smaller, domain-oriented change surfaces.

### Regression tests / protection

Architecture gates require responsibility files; existing Route validation tests preserve issue semantics/order.

### Adversarial review findings

The refactor deliberately preserves every validation rule and avoids unrelated renaming or optional-value redesign.

### Remediation iterations

Single Stage 14.11 behavior-preserving decomposition.

### Residual risks / limitations

Responsibility files can still drift in cohesion; future review remains necessary.

### Operational / deployment consequences

None.

### Exact evidence

Implementation commit: `d1fc34b6f25b5d7e8c18ac287709241e42617000`.

### Final canonical status

**CLOSED.**

### Prevention / future guard

Keep validation rule ownership domain-specific and require semantic justification for further splitting or consolidation.

## 12. GFA-MAINT-048 — Historical-neighbor `Project` orchestration concentration

### Finding / symptom

The public historical-neighbor `Project` method directly owned horizon planning, neighbor selection, pattern confidence, endpoint preparation, candidate indexing, translation, sample combination, limitations, result assembly, and fallback selection.

### Root cause

Detailed algorithmic steps accumulated in the public operation as projection functionality matured, turning the entry point into both orchestrator and implementation.

### Failure scenario

A change to one detail (geometry, fallback, sample combination, provenance) forces modification of the public coordinator and makes unit isolation of that concern harder, increasing the chance of unintended cross-step behavior changes.

### Impact

Reviewability/test isolation degrade for a correctness-sensitive projection path, even when formulas themselves are unchanged.

### Severity rationale

**P3 retrospective.** The stage explicitly preserved formulas, confidence weights, thresholds, fallback reason codes, and behavior.

### Existing guarantees violated

Public operations should coordinate major steps; detailed computations and evidence construction should have dedicated owners.

### Considered solutions

Keep monolith; replace with a generic pipeline framework; extract preparation/translation/combination/fallback/result helpers while retaining one public operation.

### Chosen remediation / why selected

`Project` becomes a narrow coordinator; detailed continuation, samples, geometry, fallback, evidence/provenance, and fingerprinting move to dedicated owners. This improves isolation without formula changes.

### Rejected alternatives

A generic pipeline/framework was unnecessary and could make deterministic control flow harder to inspect. Pure line splitting without responsibility ownership was also rejected.

### Trade-offs

More internal functions/files, but clearer algorithm boundaries and focused tests.

### Regression tests / protection

Architecture tests keep `Project` narrow; existing projection tests/race tests preserve behavior and fallback semantics.

### Adversarial review findings

The review explicitly prohibited formula/threshold/confidence changes in the same patch, preventing a structural refactor from hiding analytical changes.

### Remediation iterations

Targeted Stage 14.11 decomposition; later projection correctness stages build on the same public operation without re-monolithizing it.

### Residual risks / limitations

A narrow coordinator does not itself prove formula quality; benchmark/calibration governance remains separate.

### Operational / deployment consequences

None.

### Exact evidence

Implementation commit: `d1fc34b6f25b5d7e8c18ac287709241e42617000`.

### Final canonical status

**CLOSED.**

### Prevention / future guard

Keep public analytical operations orchestration-focused; detailed geometry, evidence, fallback, and combination logic should remain independently testable and formula changes should use separate reviewed increments.

## 13. GFA-MAINT-049 — Estimated-arrival orchestration/computation concentration

### Finding / symptom

`Estimate` and `computeArrival` jointly owned validation, identity/future-evidence checks, destination eligibility, sample construction, crossing interpolation, inside-radius handling, bounded extrapolation, confidence/provenance, and result mutation.

### Root cause

ETA functionality expanded around one public operation without separating availability/evidence policy from geometric arrival calculation.

### Failure scenario

A geometry change can inadvertently affect validation/confidence/provenance/result mutation because the responsibilities share the same implementation surface; focused testing of arrival-radius crossing versus unavailable-result policy becomes harder.

### Impact

Maintainability and correctness-review isolation are weakened on a user-visible analytical feature.

### Severity rationale

**P3 retrospective.** No ETA formula bug was established; all analytical behavior was preserved by design.

### Existing guarantees violated

Validation/availability/evidence policy should be distinct from geometric arrival computation, while the public `Estimate` operation remains stable.

### Considered solutions

Keep combined methods; introduce a general forecasting framework; split request/availability/sample selection/arrival geometry/result attachment into explicit internal responsibilities.

### Chosen remediation / why selected

`Estimate` coordinates validation, gates, sample selection, arrival computation, and attachment. Arrival calculation is separated into distance, radius crossing, already-inside, and bounded post-horizon extrapolation owners; confidence/provenance remain separate.

### Rejected alternatives

A general framework was disproportionate. Mechanical renaming or optional-value redesign was rejected because it did not address the confirmed cohesion issue.

### Trade-offs

More internal code units versus clearer separation between geometry and evidence/contract policy.

### Regression tests / protection

Architecture tests keep `Estimate`/`computeArrival` narrow; existing ETA tests protect interpolation, inside-radius, extrapolation, unavailable paths, confidence, and provenance behavior.

### Adversarial review findings

Structural change was explicitly barred from modifying projection thresholds, confidence weights, or formulas.

### Remediation iterations

Targeted Stage 14.11 behavior-preserving decomposition.

### Residual risks / limitations

The split improves reviewability but does not calibrate arrival predictions; formula evaluation remains governed by Document 46.

### Operational / deployment consequences

None.

### Exact evidence

Implementation commit: `d1fc34b6f25b5d7e8c18ac287709241e42617000`.

### Final canonical status

**CLOSED.**

### Prevention / future guard

Keep availability/validation/evidence policy separate from numerical arrival geometry; require separate evidence-backed changes for formula or threshold modifications.
