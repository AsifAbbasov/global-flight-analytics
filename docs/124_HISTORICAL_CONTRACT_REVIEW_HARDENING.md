# Historical Contract Review Hardening

## Baseline

`39549504bbeff1a6c272153bf3dcde469b766202`

## Corrected integrity boundaries

- One production metric catalog owns metric name, unit, aggregation, value kind, builder family, and allowed scopes.
- Only the sixteen metrics materialized by Traffic, Airport, and Route builders are advertised as supported.
- Four reserved names remain source-compatible constants but are not accepted as materializable metrics.
- Count values remain in the heterogeneous float64 transport field but must be exact non-negative safe integers.
- Ratio values use an explicit absolute tolerance; continuous values use a dimensionless relative tolerance; count comparisons are exact.
- Unavailable bucket confidence is canonical zero evidence.
- Partial series require at least one represented partial or complete bucket.
- A zero-coverage result with no represented bucket is canonical unavailable and does not expose unavailable points as analytical observations.
- Partial and unavailable evidence require explicit limitations.
- Comparison current values are bound to the aggregation-selected current summary.
- Confidence reason contributions must reconcile to the declared score.
- The schema registry now describes every semantic field already present in the result model.
- Aggregate region normalization is lowercase and matches the contract.
- Zero-event complete buckets remain valid because source coverage and event count are independent concepts.

## Deliberately retained contracts

- `Point.Value` remains `float64` because one versioned series model carries count, ratio, rate, and distance metrics.
- Optional comparison fields remain Go pointers; nil is the idiomatic representation of absent optional comparison evidence.
- Custom granularity remains supported because the production window planner and materializer already use it.
- Region scope remains a structural type but no current production metric catalog entry allows it.

## Versions

```text
Historical Contract implementation: historical-intelligence-contract-v2
Historical Contract validation: historical-intelligence-contract-validation-v2
Schema: historical-intelligence-v1
PostgreSQL migration: not required
```

## Closure

The original pull-request-triggered GitHub Actions run for implementation commit
`fc254881fa446c7e80f94a959e2a9d5609874821` is not recoverable through the
available repository evidence. Closure is therefore not represented as
historical exact-commit Continuous Integration evidence. Current repository-state
regression evidence is available on PR #123 exact head
`bbe20b1e2a9da873e0b5400aac136bd0b0c006c8`: Backend CI run `32835560290`
completed successfully and Backend Quality executed
`historicalcontractreviewaudit -strict` successfully.

```text
OPEN_CONFIRMED_FINDINGS=0
UNCLASSIFIED_FINDINGS=0
DEFERRED_FINDINGS=0
HISTORICAL_CONTRACT_REVIEW_STATUS=CLOSED
```

## Canonical remediation history

Historical adversarial-review identities/comments and original exact-commit Continuous Integration evidence are unavailable; reconstruction is limited to repository source, tests, implementation commit history, the permanent strict review audit, and later exact-head regression evidence. Severity labels are retrospective.

### GFA-CONTRACT-214 — Historical metric identity, aggregation and scope policy was fragmented across builders

1. **Finding / symptom:** metric name, unit, aggregation, value kind, builder family and allowed scope were not owned by one production catalog, while reserved source-compatible metric constants could appear indistinguishable from materializable metrics.
2. **Root cause:** metric policy was duplicated across builders and validation rather than represented by one authoritative specification.
3. **Failure scenario:** a builder or consumer accepts a reserved/unsupported metric, associates the wrong unit or aggregation, or permits a scope that no production builder can materialize.
4. **Impact:** durable Historical Results can carry contradictory metric identity or claim unsupported analytical capability.
5. **Severity rationale:** **P1 retrospective** because metric identity and scope determine the meaning of every persisted value.
6. **Existing guarantees violated:** one deterministic contract must bind metric identity, numerical kind, builder ownership and permitted scopes.
7. **Considered solutions:** preserve per-builder switches, remove reserved constants, or centralize materialization policy while retaining source compatibility.
8. **Chosen remediation:** `MetricSpecFor` becomes the single production catalog; only sixteen actually materialized metrics are supported and four reserved names remain constants without materialization permission.
9. **Why this solution was selected:** it removes policy drift without a breaking source-level deletion of reserved names.
10. **Rejected alternatives:** allowing arbitrary metric/unit combinations and treating constant existence as production support.
11. **Trade-offs:** adding a new production metric now deliberately requires a catalog change and contract review.
12. **Regression tests / protection:** catalog validation tests, builder-family/scope tests, and `historicalcontractreviewaudit -strict`.
13. **Adversarial review findings:** heterogeneous `Point.Value` transport and structural Region scope were retained because neither implies that every metric/scope combination is supported.
14. **Remediation iterations:** implemented in `fc254881fa446c7e80f94a959e2a9d5609874821`.
15. **Residual risks and limitations:** the catalog remains a finite versioned policy and must be updated deliberately when product scope expands.
16. **Operational or deployment consequences:** no PostgreSQL migration; contract generation advances to version two.
17. **Exact evidence:** implementation commit `fc254881fa446c7e80f94a959e2a9d5609874821`; permanent `historicalcontractreviewaudit`; later exact-head regression evidence PR #123 head `bbe20b1e...`, Backend CI run `32835560290`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every new Historical metric must enter through the canonical metric catalog and prove builder/scope reachability before being advertised.

### GFA-DATA-215 — Count metrics could accept fractional, negative, non-finite or non-exact float64 values

1. **Finding / symptom:** the shared float64 transport did not by itself guarantee that count metrics represented exact non-negative integers.
2. **Root cause:** value validation was generic instead of being driven by metric value kind.
3. **Failure scenario:** a count bucket carries `1.5`, a negative value, or a magnitude beyond the exact IEEE-754 integer boundary and is treated as a valid count.
4. **Impact:** mathematically impossible count evidence can enter summaries, comparisons and persistence.
5. **Severity rationale:** **P1 retrospective** because the defect can publish invalid analytical values as factual event counts.
6. **Existing guarantees violated:** count metrics must be exact, finite, non-negative integers within the transport's exact domain.
7. **Considered solutions:** replace all float64 transport with integer/decimal unions, rely on builders, or validate according to catalog value kind.
8. **Chosen remediation:** retain heterogeneous float64 transport but require exact non-negative safe integers for count metrics.
9. **Why this solution was selected:** it preserves the version-one result schema while enforcing mathematically correct count semantics.
10. **Rejected alternatives:** cross-contract numeric-union migration without a demonstrated need for non-integer count transport.
11. **Trade-offs:** extremely large counts beyond exact float64 integer representation are rejected rather than approximated.
12. **Regression tests / protection:** fractional/negative/non-finite/exact-boundary count tests and strict contract audit.
13. **Adversarial review findings:** float64 remains appropriate as a transport for the mixed metric model only because metric-kind validation constrains count values.
14. **Remediation iterations:** `fc254881fa446c7e80f94a959e2a9d5609874821`.
15. **Residual risks and limitations:** consumers must not reinterpret count-valued float64 fields without consulting the metric catalog.
16. **Operational or deployment consequences:** malformed historical count results now fail validation; no migration.
17. **Exact evidence:** implementation commit, metric-value validation tests, permanent contract audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every metric value must be validated through its catalog-declared value kind before aggregation or persistence.

### GFA-DATA-216 — Historical numerical tolerance was not coherent across counts, ratios and continuous measurements

1. **Finding / symptom:** one generic comparison tolerance could not correctly represent exact counts, bounded ratios and unit-bearing continuous aviation values.
2. **Root cause:** numerical comparison policy was not specialized by metric kind.
3. **Failure scenario:** a tolerance appropriate for a ratio is applied to a large distance or exact count, hiding a real mismatch or creating a false failure.
4. **Impact:** validation and comparison results become dependent on units rather than analytical meaning.
5. **Severity rationale:** **P1 retrospective** because mathematically contradictory results can be accepted or valid results rejected.
6. **Existing guarantees violated:** tolerances must be dimensionally and semantically appropriate for the values being compared.
7. **Considered solutions:** zero tolerance everywhere, one absolute epsilon, per-metric constants, or value-kind-specific policy.
8. **Chosen remediation:** count comparisons are exact, ratio comparisons use explicit absolute tolerance, and continuous values use dimensionless relative tolerance.
9. **Why this solution was selected:** it matches the mathematical behavior of each metric family without introducing arbitrary unit-specific magic numbers.
10. **Rejected alternatives:** adding one tolerance directly to values of incompatible dimensions.
11. **Trade-offs:** future near-zero continuous relationships may require an explicitly documented absolute floor if evidence shows a need.
12. **Regression tests / protection:** count, ratio and continuous comparison boundary tests plus strict contract audit.
13. **Adversarial review findings:** full decimal arithmetic was rejected because these are non-monetary analytical metrics and exact-count semantics are already enforceable.
14. **Remediation iterations:** `fc254881fa446c7e80f94a959e2a9d5609874821`.
15. **Residual risks and limitations:** relative tolerance behavior around zero remains governed by the shared comparison helper and its tests.
16. **Operational or deployment consequences:** stricter deterministic validation only; no persistence migration.
17. **Exact evidence:** implementation commit and value-kind tolerance regression tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** any new numerical comparison must declare whether its tolerance is exact, absolute, relative, or expressed in a named physical unit.

### GFA-DATA-217 — Unavailable or zero-coverage buckets could retain analytical value/confidence evidence

1. **Finding / symptom:** unavailable evidence did not have one canonical zero payload/confidence rule, and zero-coverage results could still expose bucket points as if they were analytical observations.
2. **Root cause:** availability status, point payload, coverage and confidence were validated as partially independent fields.
3. **Failure scenario:** a bucket claims no usable evidence but retains a value, samples, coverage-derived point, or non-zero confidence.
4. **Impact:** absence can masquerade as measured analytical evidence.
5. **Severity rationale:** **P1 retrospective** because consumers may publish or compare values that the result itself declares unavailable.
6. **Existing guarantees violated:** unavailable evidence must have canonical zero analytical payload and confidence; zero coverage is not an observation.
7. **Considered solutions:** let consumers ignore payload when status is unavailable, sanitize in storage, or reject contradictory evidence at the contract boundary.
8. **Chosen remediation:** unavailable buckets require zero value/samples/coverage/confidence; zero-coverage series with no represented evidence are canonical unavailable and expose no analytical points.
9. **Why this solution was selected:** fail-closed validation prevents stale/default values from becoming durable evidence.
10. **Rejected alternatives:** consumer-side convention and silent payload repair.
11. **Trade-offs:** legacy malformed fixtures must be corrected explicitly.
12. **Regression tests / protection:** unavailable bucket confidence/payload tests and zero-coverage representation tests.
13. **Adversarial review findings:** complete zero-event buckets remain valid because complete source coverage with zero events is fundamentally different from unavailable evidence.
14. **Remediation iterations:** `fc254881fa446c7e80f94a959e2a9d5609874821`.
15. **Residual risks and limitations:** consumers must still respect result status when presenting analytical availability.
16. **Operational or deployment consequences:** contradictory unavailable results fail closed; no migration.
17. **Exact evidence:** implementation commit, unavailable/zero-coverage tests, strict contract audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every unavailable state must define and test canonical residual payload and confidence semantics.

### GFA-DATA-218 — Partial and unavailable Historical evidence could be structurally unexplained

1. **Finding / symptom:** partial/unavailable results and buckets could lack limitations, and a partial series could contain no actually represented partial or complete bucket.
2. **Root cause:** status/count validation did not require evidence-producing components to explain why data was incomplete.
3. **Failure scenario:** a result claims partial availability but every point is unavailable, or an unavailable result provides no reason for the loss of evidence.
4. **Impact:** silent data loss is indistinguishable from a known limitation and downstream confidence/explainability becomes unreliable.
5. **Severity rationale:** **P1 retrospective** because incomplete evidence can be persisted as legitimate without auditable cause.
6. **Existing guarantees violated:** non-complete Historical evidence must be explainable and internally consistent with represented buckets.
7. **Considered solutions:** synthesize generic limitations in Validator, accept status alone, or require domain-owned limitations and represented evidence.
8. **Chosen remediation:** partial/unavailable states require explicit limitations; partial series require at least one represented partial or complete bucket.
9. **Why this solution was selected:** it keeps provenance ownership with the component that actually excluded or lost evidence.
10. **Rejected alternatives:** fabricated generic limitation text and permissive status-only acceptance.
11. **Trade-offs:** producers must carry more precise diagnostic metadata.
12. **Regression tests / protection:** missing limitation, all-unavailable partial series and represented-bucket tests.
13. **Adversarial review findings:** nil optional comparison pointers remain unrelated because they represent mathematically absent optional fields, not unexplained evidence loss.
14. **Remediation iterations:** `fc254881fa446c7e80f94a959e2a9d5609874821`.
15. **Residual risks and limitations:** explanation presence does not replace source-specific correctness tests for the explanation itself.
16. **Operational or deployment consequences:** malformed incomplete results are rejected earlier.
17. **Exact evidence:** implementation commit, status/limitation regression tests, strict contract audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every non-complete status added in future must define required evidence and limitation invariants in the contract validator.

### GFA-DATA-219 — Comparison current values were not bound to the aggregation-selected current summary

1. **Finding / symptom:** a comparison could carry a current value that did not equal the canonical summary field selected by the metric's aggregation.
2. **Root cause:** comparison payload validation checked finite shape without reconciling it to aggregation semantics.
3. **Failure scenario:** an average metric comparison uses `Summary.Total` or another stale value while the result advertises `AggregationAverage`.
4. **Impact:** percentage/direction evidence can disagree with the current Historical Result it claims to compare.
5. **Severity rationale:** **P1 retrospective** because derived comparison output may be mathematically valid but semantically attached to the wrong source value.
6. **Existing guarantees violated:** comparison current value must be a deterministic projection of the current result's declared aggregation.
7. **Considered solutions:** trust the comparison producer, remove the mirrored current value, or validate it against catalog-selected summary semantics.
8. **Chosen remediation:** current comparison values are reconciled to the aggregation-selected current summary.
9. **Why this solution was selected:** preserves the existing schema while making the mirror a checked invariant.
10. **Rejected alternatives:** allowing each builder to choose its own current-value interpretation.
11. **Trade-offs:** previously tolerated inconsistent derived records now fail validation.
12. **Regression tests / protection:** aggregation-selector/current-value mismatch tests and strict contract audit.
13. **Adversarial review findings:** generic `Summary.Total` remains descriptive and is not redefined to fit ratio/average comparison semantics.
14. **Remediation iterations:** `fc254881fa446c7e80f94a959e2a9d5609874821`.
15. **Residual risks and limitations:** comparison construction still depends on the metric catalog remaining the single aggregation owner.
16. **Operational or deployment consequences:** invalid derived comparisons fail closed; no migration.
17. **Exact evidence:** implementation commit and comparison-summary reconciliation tests.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** any mirrored derived field must be validated against its canonical source field/formula before persistence.

### GFA-DATA-220 — Confidence reason contributions were not required to reconcile to the declared confidence score

1. **Finding / symptom:** confidence could publish a score whose explanatory reason contributions were absent or summed to a different value.
2. **Root cause:** reason entries were range-checked individually but not reconciled as a complete explanation of the score.
3. **Failure scenario:** confidence score `0.8` is persisted with no reasons or with contributions totaling `0.3`.
4. **Impact:** explainability evidence contradicts the numerical trust score and cannot be audited deterministically.
5. **Severity rationale:** **P1 retrospective** because confidence is a first-class trust claim and its explanation could be mathematically false.
6. **Existing guarantees violated:** declared confidence and reason contributions must represent one reproducible calculation.
7. **Considered solutions:** treat reasons as decorative, require at least one reason only, or reconcile compensated contribution sum to score.
8. **Chosen remediation:** positive confidence requires explanatory reasons and compensated contribution sum must match the declared score.
9. **Why this solution was selected:** turns reasons into verifiable provenance rather than prose metadata.
10. **Rejected alternatives:** accepting unexplained positive confidence.
11. **Trade-offs:** confidence producers must preserve contribution decomposition explicitly.
12. **Regression tests / protection:** missing-reason and contribution-mismatch tests plus strict audit.
13. **Adversarial review findings:** compensated accumulation is used to avoid misclassifying ordinary floating-point summation noise.
14. **Remediation iterations:** `fc254881fa446c7e80f94a959e2a9d5609874821`.
15. **Residual risks and limitations:** reason correctness beyond arithmetic reconciliation remains producer/domain-policy responsibility.
16. **Operational or deployment consequences:** contradictory confidence payloads are rejected before storage/API use.
17. **Exact evidence:** implementation commit, confidence validation tests, strict contract audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** any explainability decomposition must be mathematically reconciled to the aggregate score it explains.

### GFA-CONTRACT-221 — Historical schema registry omitted semantic fields already present in the result contract

1. **Finding / symptom:** the versioned schema registry did not describe every semantic field produced by the Historical Result model.
2. **Root cause:** schema metadata and Go result evolution were maintained separately without a completeness guard.
3. **Failure scenario:** tooling or consumers rely on the registry and miss a semantic field that runtime results already persist or expose.
4. **Impact:** contract documentation/code generation/audit can drift from actual durable payload semantics.
5. **Severity rationale:** **P2 retrospective** because runtime values can remain correct while the declared versioned schema is incomplete.
6. **Existing guarantees violated:** a versioned schema registry must completely describe the semantic result it claims to govern.
7. **Considered solutions:** remove the registry, document missing fields informally, or make the registry complete and audit it.
8. **Chosen remediation:** register every semantic field already present in the version-one result model.
9. **Why this solution was selected:** preserves the existing schema version while eliminating declaration drift.
10. **Rejected alternatives:** silently relying on Go struct reflection as the public schema authority.
11. **Trade-offs:** future result-field changes require synchronized registry maintenance.
12. **Regression tests / protection:** schema completeness tests and strict historical contract audit.
13. **Adversarial review findings:** no new schema version was required because the runtime fields already existed; the defect was registry incompleteness.
14. **Remediation iterations:** `fc254881fa446c7e80f94a959e2a9d5609874821`.
15. **Residual risks and limitations:** semantic descriptions still require review for meaning, not only field presence.
16. **Operational or deployment consequences:** no PostgreSQL migration and no wire-field change.
17. **Exact evidence:** implementation commit, schema registry changes/tests, permanent audit.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** schema audits must fail when semantic runtime fields and registry definitions diverge.

### GFA-CONTRACT-222 — Region scope normalization disagreed across Historical Contract and aggregate/builder boundaries

1. **Finding / symptom:** region identifiers could be normalized with different casing rules across Historical components.
2. **Root cause:** scope normalization policy was duplicated and had not converged on one canonical representation.
3. **Failure scenario:** semantically identical region scopes produce different keys/fingerprints or fail equality depending on which component normalized them.
4. **Impact:** duplicate identities, failed lookups and inconsistent scope comparison.
5. **Severity rationale:** **P2 retrospective** because the defect primarily affects canonical identity and compatibility rather than inventing metric values directly.
6. **Existing guarantees violated:** one semantic scope must have one deterministic canonical key representation.
7. **Considered solutions:** uppercase everywhere, lowercase everywhere, preserve caller case, or introduce case-insensitive keys.
8. **Chosen remediation:** aggregate and contract region normalization use lowercase canonical region codes.
9. **Why this solution was selected:** it aligns current contract semantics and avoids case-dependent identity.
10. **Rejected alternatives:** preserving mixed-case caller identity and silent dual-key compatibility.
11. **Trade-offs:** persisted rows created under incompatible legacy region keys require explicit migration/rematerialization when database constraints are corrected later.
12. **Regression tests / protection:** region normalization/scope-key tests and strict contract audit.
13. **Adversarial review findings:** Region scope remains structurally defined even though no current production metric catalog entry permits materialization; structural capability is not production support.
14. **Remediation iterations:** `fc254881fa446c7e80f94a959e2a9d5609874821`; database persistence alignment is later hardened in Document 131.
15. **Residual risks and limitations:** pre-existing incompatible persisted region identifiers are outside this no-migration contract increment and require governed remediation at the store layer.
16. **Operational or deployment consequences:** no migration in this increment; deterministic lowercase scope identity for new contract operations.
17. **Exact evidence:** implementation commit, scope normalization/key tests, permanent audit; later persistence remediation documented separately in Document 131.
18. **Final canonical status:** **CLOSED** for the contract/builder normalization scope.
19. **Prevention / future guard:** scope normalization must have a single documented canonical form and storage migrations must validate against it before new persisted scope types are enabled.
