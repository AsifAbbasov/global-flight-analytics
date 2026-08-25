# Document 97 — Analytical Contributor Semantics Hardening

Status: Implemented Engineering Increment v1.0
Project: Global Flight Analytics
Baseline: `1ddb65c5e5471ce180314cc38a4b6d7baad80cd3`

## 1. Purpose

This increment closes three correctness findings from the original Analytical
Core Foundation review:

```text
eligibility must run before aircraft-level deduplication;
observations materially in the future must not contribute to analytics;
traffic-density arithmetic must reject non-finite area values.
```

## 2. Contributor ordering contract

For Active Aircraft and Traffic Density the production order is now:

```text
evaluate every trajectory for capability eligibility;
retain eligible contributors;
deduplicate eligible contributors by aircraft identity;
calculate the metric;
publish duplicate and exclusion evidence.
```

An ineligible newer trajectory can no longer replace an eligible older
trajectory for the same aircraft.

Denied contributors remain visible in the scope summary. Duplicate removal is
performed only among eligible contributors.

## 3. Future observation contract

Every analytical capability uses a bounded clock-skew tolerance:

```text
default maximum future observation skew: 30 seconds
```

A trajectory whose end time exceeds the evaluation time plus the configured
tolerance receives:

```text
future_observation
```

and is excluded from the affected capability.

A timestamp exactly at the tolerance boundary remains eligible.

## 4. Traffic Density finite-number contract

Traffic Density now rejects:

```text
negative active-aircraft counts;
zero or negative area;
NaN area;
positive or negative infinite area.
```

The domain calculator owns these checks independently of HTTP parsing.

## 5. Verification

The installer executes targeted analytical tests, targeted race tests, complete
backend tests, Go static analysis, architecture audits and whitespace checks.

## 6. Remaining Analytical Core review scope

This increment does not close the full review. Remaining accepted work includes:

```text
airport-owned Airport Activity classification;
geographically bound Traffic Density requests;
server-owned production Coverage Score and Data Freshness;
strict analytical provenance and safe public failures;
reference-time and UUID normalization;
obsolete analytical foundation classification;
metric identifier consolidation.
```

---

## Canonical remediation history

The original Analytical Core review identifiers are retained below so that later repository-level finding IDs remain traceable to the review that produced them.

### GFA-DATA-099 / AC-01 — eligibility ran after aircraft-level deduplication

1. **Finding / symptom.** Active Aircraft and Traffic Density could deduplicate trajectories by aircraft before capability eligibility was evaluated.
2. **Root cause.** Contributor identity reduction was performed on the raw trajectory set instead of on the eligible contributor set.
3. **Failure scenario.** A newer ineligible trajectory for an aircraft wins deduplication over an older eligible trajectory; the ineligible row is then denied, leaving that aircraft absent from the metric even though eligible evidence existed.
4. **Impact.** Active-aircraft counts and derived traffic density can be understated, while scope evidence no longer represents the best eligible observation for an aircraft.
5. **Severity rationale.** **P1 retrospective.** This changes published analytical values under ordinary mixed-quality data and can systematically exclude valid contributors.
6. **Existing guarantees violated.** Capability policy must decide eligibility before contributor reduction; denied evidence must not erase eligible evidence for the same aircraft; scope accounting must remain explainable.
7. **Considered solutions.** Keep pre-filter deduplication and alter tie-breaking; include all duplicate trajectories in calculation; filter first then deduplicate; move deduplication into every metric independently.
8. **Chosen remediation.** `executeTrajectoryMetric` filters for capability eligibility first and then applies an explicit contributor-preparation step; Active Aircraft and Traffic Density use `prepareUniqueAircraftContributors` only on eligible trajectories.
9. **Why this solution was selected.** It makes the ordering invariant structural and reusable while preserving denied-contributor evidence in the scope summary.
10. **Rejected alternatives.** More elaborate pre-filter tie-breaking cannot know future eligibility; counting duplicate eligible trajectories double-counts aircraft; duplicating the policy in each metric increases drift risk.
11. **Trade-offs.** Metric execution has an explicit preparation phase and duplicate warnings now describe eligible duplicates specifically; this added stage is intentional because eligibility and identity reduction are different policies.
12. **Regression tests / protection.** Tests construct an older eligible and newer denied trajectory for one aircraft and require the eligible aircraft to remain counted for Active Aircraft and Traffic Density. The final Analytical Core audit preserves eligibility-before-deduplication ordering.
13. **Adversarial review findings.** Scope input counts must still include denied contributors even though deduplication occurs only among eligible items; warning text must not imply denied duplicates were removed from the raw input.
14. **Remediation iterations.** The execution pipeline gained a `trajectoryMetricPreparation` seam so the fix was not a one-off reorder in two handlers; contributor counts were adjusted to preserve allowed-plus-denied evidence after preparation.
15. **Residual risks and limitations.** Deduplication remains aircraft-identity based and does not resolve higher-order track fusion; provider identity quality and trajectory eligibility are governed by separate contracts.
16. **Operational or deployment consequences.** No new infrastructure or configuration is required. Published counts may increase relative to the defective ordering when an eligible older trajectory coexists with a newer denied trajectory.
17. **Exact evidence.** Historical implementation commit `c5fd1f32273af9215df9d83d1d40c227d3740646` (`fix: harden analytical contributor semantics`). Original review ID: `AC-01`. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-DATA-099=CLOSED`.
19. **Prevention / future guard.** New contributor-reduction logic must run after capability eligibility unless a metric explicitly documents and tests a different semantic order; the permanent Analytical Core audit must continue to detect reversal.

### GFA-DATA-100 / AC-02 — materially future observations could contribute to analytics

1. **Finding / symptom.** Analytical eligibility did not consistently exclude trajectory observations whose timestamps were materially later than the metric evaluation time.
2. **Root cause.** Staleness/recency policy handled old evidence but had no symmetric bounded future-clock-skew rule across analytical capabilities.
3. **Failure scenario.** A provider, clock, replay or malformed timestamp places a trajectory minutes or hours in the future; the record is treated as fresh/current and contributes to live analytical metrics.
4. **Impact.** Future evidence can inflate activity, distort freshness and contaminate time-bounded analytical decisions with observations that could not have existed at evaluation time.
5. **Severity rationale.** **P1 retrospective.** The defect violates temporal truth of published analytics and can affect multiple metrics from one bad timestamp.
6. **Existing guarantees violated.** Analytical contributors must belong to a valid temporal window; bounded engineering clock skew may be tolerated, but materially future evidence must be explicit and excluded.
7. **Considered solutions.** Reject every timestamp after `now`; silently clamp future times to `now`; configure one bounded skew tolerance; leave future handling to provider adapters only.
8. **Chosen remediation.** All analytical capabilities use a maximum future observation skew, defaulting to 30 seconds; items later than `evaluatedAt + tolerance` receive the `future_observation` exclusion reason.
9. **Why this solution was selected.** A small tolerance handles realistic clock jitter without legitimizing materially future evidence, and keeping the rule in analytical eligibility protects replay and non-provider callers too.
10. **Rejected alternatives.** Zero tolerance is operationally brittle; timestamp clamping fabricates evidence time; provider-only enforcement leaves other analytical entry paths unprotected.
11. **Trade-offs.** Slightly future observations within the tolerance remain eligible by design. The tolerance is an engineering policy, not a claim that provider clocks are exact.
12. **Regression tests / protection.** Eligibility tests cover materially future observations and the exact tolerance boundary; the Analytical Core final audit requires the future-observation guard.
13. **Adversarial review findings.** Boundary semantics must be deterministic: exactly-at-tolerance remains eligible, while any timestamp beyond the bound is denied; the rule must operate from the execution reference time rather than wall-clock calls scattered through metrics.
14. **Remediation iterations.** The future-time rule was centralized at trajectory eligibility so individual metrics did not need separate timestamp checks.
15. **Residual risks and limitations.** A timestamp inside the allowed skew can still be slightly ahead of the evaluator. Upstream clock provenance and replay-time semantics remain separate concerns.
16. **Operational or deployment consequences.** No infrastructure change. Bad future timestamps become explicit analytical exclusions and may reduce affected metric contributor counts.
17. **Exact evidence.** Historical implementation commit `c5fd1f32273af9215df9d83d1d40c227d3740646`. Original review ID: `AC-02`. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-DATA-100=CLOSED`.
19. **Prevention / future guard.** Any new analytical capability using trajectory time must use the shared eligibility boundary or prove an equivalent bounded future-evidence rule with boundary tests.

### GFA-DATA-101 / AC-07 — Traffic Density accepted non-finite arithmetic inputs

1. **Finding / symptom.** Traffic Density validated only that area was greater than zero and therefore did not explicitly reject `NaN`, positive/negative infinity, or a negative active-aircraft count.
2. **Root cause.** Numeric validation assumed ordinary finite HTTP-originated values instead of treating the domain calculator as an independent trust boundary.
3. **Failure scenario.** An internal caller, test seam or future transport passes a non-finite area or negative count; invalid floating-point arithmetic reaches the metric calculation and may propagate an undefined analytical result or serialization failure.
4. **Impact.** The metric can violate its numeric domain and lose the guarantee that a published density is a finite, non-negative physical ratio.
5. **Severity rationale.** **P2 retrospective.** This is a real analytical correctness defect, but it requires invalid numeric input rather than ordinary valid provider evidence.
6. **Existing guarantees violated.** Metric calculators must validate their own domain independently of HTTP parsing; published metric values must be finite and semantically valid.
7. **Considered solutions.** Rely on request parsing; sanitize invalid values to zero; add finite-number and non-negative-count checks in the calculator; add generic reflection-based numeric validation.
8. **Chosen remediation.** `TrafficDensityMetric.Calculate` rejects negative aircraft counts and rejects area when `NaN`, infinite, zero or negative before division.
9. **Why this solution was selected.** The calculator is the canonical semantic owner and explicit checks are simpler and safer than transport assumptions or generic numeric machinery.
10. **Rejected alternatives.** Coercing invalid values would fabricate a valid-looking metric; HTTP-only validation leaves internal callers unsafe; generic validation obscures the exact domain rules.
11. **Trade-offs.** Invalid calculations now fail rather than returning a numeric placeholder. Callers must handle the typed execution failure path.
12. **Regression tests / protection.** Tests cover `NaN`, both infinities, zero/negative area and negative active-aircraft counts. The Analytical Core final audit preserves finite-number guards.
13. **Adversarial review findings.** `NaN <= 0` is false, so an ordinary positivity check is insufficient; infinity also passes a simple positive check. The order and explicit `math.IsNaN`/`math.IsInf` checks are therefore material.
14. **Remediation iterations.** Validation was moved into the metric domain rather than added only to HTTP handling, making the contract reusable for every caller.
15. **Residual risks and limitations.** Finite but physically implausible region sizes are governed by geographic scope configuration, not by this low-level calculator.
16. **Operational or deployment consequences.** None beyond invalid inputs now producing a controlled failed analytical execution.
17. **Exact evidence.** Historical implementation commit `c5fd1f32273af9215df9d83d1d40c227d3740646`. Original review ID: `AC-07`. Historical adversarial-review/PR evidence unavailable; reconstruction is limited to repository source, tests, commits and CI evidence.
18. **Final canonical status.** `GFA-DATA-101=CLOSED`.
19. **Prevention / future guard.** New floating-point analytical formulas must define finite-number and sign/range ownership in the domain calculator, with tests for `NaN` and infinity where applicable.