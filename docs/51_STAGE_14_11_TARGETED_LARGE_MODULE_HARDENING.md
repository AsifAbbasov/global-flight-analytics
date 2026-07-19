# Document 51 — Stage 14.11 Targeted Large-Module Hardening

Status: Implementation Baseline v1.0
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
