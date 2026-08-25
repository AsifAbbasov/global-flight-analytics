# Document 46 — Stage 14.6 Formula Benchmark and Calibration Gate

Status: Remediation History v1.1
Project: Global Flight Analytics
Scope: reproducible offline evaluation of projection formulas without automatic calibration

## 1. Decision

Stage 14.6 does not change production formulas.

The repository already contains a Projection Evaluation engine capable of
calculating position truth coverage, horizontal and altitude errors,
uncertainty coverage, arrival-time error, and arrival interval coverage.

This increment connects those existing results to a bounded external research
manifest, explicit release-gate policy, deterministic report, and manual
decision boundary.

## 2. Offline Command

```text
benchmark-projection-formulas
```

Usage:

```bash
go run ./cmd/benchmark-projection-formulas \
  --input request.json \
  --output report.json
```

The command does not download datasets, access production PostgreSQL, modify
source code, write formula configuration, or alter production weights.

## 3. Dataset Boundary

The benchmark plan uses the adopted bounded regional subset of:

```text
opensky-weekly-state-vectors-2017-2022
```

The source remains offline only, non-production, region-filtered,
licence-reviewed, attributed, and bounded by files, bytes, and maximum records.

Monday-only data cannot support general weekly seasonality claims.

## 4. Report Status

Each projection method and the complete report receive one of:

```text
insufficient_evidence
benchmark_failed
benchmark_passed
```

Evidence gates include minimum evaluation count, complete evaluation ratio,
truth point coverage, altitude evidence coverage, and arrival evidence
coverage.

Performance gates include horizontal error, uncertainty coverage, altitude
error, arrival-time error, and arrival interval coverage.

## 5. Calibration Boundary

Every report always contains:

```text
calibration_allowed = false
automatic_formula_changes_allowed = false
manual_review_required = true
maximum_claim = bounded_offline_benchmark_evidence_only
```

A passing report allows engineering review. It does not prove universal model
accuracy and does not authorize formula modification.

A production formula change requires a separate reviewed increment containing
the immutable benchmark report, exact formula change, before-and-after results,
scope statement, rollback plan, and manual approval evidence.

## 6. Exit Codes

```text
0 — benchmark passed
1 — invalid request or execution failure
2 — insufficient evidence
3 — benchmark failed
```

The report is written before a non-zero evidence or threshold exit code is
returned.

## 7. Architecture Gate

`projectaudit -mode formulas -strict` verifies that the offline command imports
research benchmark governance, dataset governance, Formula Benchmark, and
Projection Evaluation; that none enter production runtime roots; and that the
command is not included in the production Docker image.

The formula audit also runs through:

```text
projectaudit -mode all -strict
```

## 8. Limitations

The default policy is a conservative project release gate, not a scientific
constant.

No external benchmark dataset is included in the repository.

No model is described as calibrated until a separate reviewed formula-change
increment records evidence and approval.

## 9. Canonical finding record — GFA-GOV-040

### Finding / symptom

Projection Evaluation code could compute accuracy/coverage metrics, but the repository lacked a reproducible bounded dataset manifest, evidence thresholds, immutable report boundary, and explicit rule preventing a benchmark result from automatically changing production formulas or being described as universal calibration.

### Root cause

Evaluation mechanics were implemented before research/release governance. The code could answer "what errors did this sample produce?" without a canonical contract for sample scope, evidence sufficiency, allowed claims, or formula-change approval.

### Failure scenario

An engineer runs a convenient local sample, sees favorable errors, changes projection weights/thresholds, and describes the result as calibrated. The sample may be too small, geographically narrow, temporally biased (for example Monday-only), or incomplete in altitude/arrival truth, and the change may have no preserved before/after evidence or rollback plan.

### Impact

The main risk is scientific/product overclaim and formula regression: production behavior can be changed on weak or irreproducible evidence while future reviewers cannot reconstruct the benchmark scope that justified it.

### Severity rationale

**P2 retrospective.** This is analytical/release governance around production formulas. No unsafe automatic calibration was proven to have occurred, but the missing gate could materially affect projection quality and credibility.

### Existing guarantees violated

- evaluation evidence must be bounded, reproducible, attributed, and licence-reviewed;
- insufficient evidence must be distinguishable from benchmark failure/success;
- offline evaluation must remain isolated from live production runtime;
- passing a benchmark is not authorization to mutate formulas;
- calibration claims require a separate reviewed change with immutable evidence.

### Considered solutions

1. let Projection Evaluation remain an internal library and rely on engineer judgment;
2. automatically optimize formula parameters from benchmark outputs;
3. expose a deterministic offline benchmark command with explicit evidence/performance gates and a hard manual-review boundary.

### Chosen remediation

`benchmark-projection-formulas` consumes a bounded manifest, produces a deterministic report with evidence/performance status, always disables automatic calibration/formula changes, and requires a separate reviewed increment before any production formula modification.

### Why selected

The solution adds the missing evidence/governance layer while preserving the existing formulas and evaluation engine. It enables repeatable research without pretending the project has enough data for autonomous model calibration.

### Rejected alternatives

Engineer judgment alone was rejected because it is not reproducible. Automatic optimization was rejected because the bounded open-data sample does not justify general calibration and could overfit incomplete regional/temporal evidence. Shipping the benchmark inside the production image was rejected because later-truth evaluation must not enter live forecast generation.

### Trade-offs

Formula improvement becomes deliberately slower: a passing benchmark still requires human review and a separate change package. External datasets are not bundled, so benchmark execution requires separately obtained/licence-reviewed data. These costs protect evidence quality and project claims.

### Regression tests / protection

`projectaudit -mode formulas -strict` requires the offline dependency graph, excludes evaluation/benchmark packages from production roots and Docker, and enforces the benchmark governance contract. Command tests cover exit statuses and deterministic report behavior.

### Adversarial review findings

The document explicitly records that Monday-only OpenSky weekly files cannot support general weekly seasonality claims and that the default policy thresholds are project release gates, not scientific constants. These limitations prevent benchmark machinery itself from becoming a source of overclaim.

### Remediation iterations

Stage 14.2 classified Projection Evaluation as offline evaluation requiring a real entry point. Stage 14.6 added that entry point and the manual calibration boundary without moving evaluation into production.

### Residual risks / limitations

Benchmark quality is still limited by dataset coverage, truth quality, regional selection, historical period, and chosen thresholds. Passing the gate does not prove universal prediction accuracy or future-distribution stability.

### Operational / deployment consequences

None for live runtime: the command stays offline and outside the production image/database path. Research operators must supply the external dataset manifest and preserve resulting reports for reviewed formula changes.

### Exact evidence

Implementation commit: `f817bad2d6d12fe1619bb5c3bba3238d94d4c620` (`feat: add offline formula benchmark gate`). Historical PR/reviewer identities are not invented when unavailable.

### Final canonical status

**CLOSED.** The repository has a reproducible offline benchmark/review boundary; production formulas remain separately governed.

### Prevention / future guard

Any formula/calibration change must cite an immutable benchmark report, exact dataset scope, evidence sufficiency, before/after metrics, limitations, rollback plan, and explicit approval. Offline evaluation packages must remain structurally excluded from live production roots unless a new architecture decision redefines their purpose.
