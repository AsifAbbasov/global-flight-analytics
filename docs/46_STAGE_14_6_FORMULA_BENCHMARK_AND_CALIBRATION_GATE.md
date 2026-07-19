# Document 46 — Stage 14.6 Formula Benchmark and Calibration Gate

Status: Implementation Baseline v1.0
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
