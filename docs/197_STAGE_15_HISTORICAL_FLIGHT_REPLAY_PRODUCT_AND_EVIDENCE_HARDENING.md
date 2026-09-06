# Stage 15 Historical Flight Replay Product and Evidence Hardening

Status: closed

```text
STAGE_15_HISTORICAL_FLIGHT_REPLAY=CLOSED
STAGE_15_IMPLEMENTATION_SHA=6ed51f563818a619bb2346e6c4ad50d4256d60ec
HISTORICAL_FLIGHT_REPLAY_EVIDENCE=OBSERVED_ONLY
HISTORICAL_FLIGHT_REPLAY_INTERPOLATION=NONE
HISTORICAL_FLIGHT_REPLAY_COST_BOUNDARY=0_RUB
STAGE_15_DOCUMENTATION=CLOSED
```

## Purpose

This document is the canonical engineering history for the post-v1.0.0 Historical Flight Replay product slice delivered through pull requests #149-#153.

It records the product problem, evidence boundary, zero-budget architecture, real remediation history, adversarial scenarios, regression protection, CI evidence and residual limitations. It deliberately distinguishes repository-proven events from retrospective engineering reasoning. A rejected alternative is not described as a review rejection unless repository evidence shows that a review or CI gate actually rejected it.

## Classification and finding ownership

Stage 15 is a product/engineering closure document rather than a new incident register. It follows the same governance pattern used by the frontend product closure documents 195-196: the stage can contain detailed remediation and adversarial records without manufacturing synthetic global finding IDs for every product decision.

The records below therefore use local identifiers `STAGE15-R1` through `STAGE15-R6`. They are permanent engineering decision/remediation records, not additions to the canonical `GFA-*` finding namespace. Existing repository findings, including unrelated open governance findings, are not reclassified by this document.

## Scope

Stage 15 adds a browser-visible Historical Flight Replay workspace for persisted flight observations and then hardens that workspace through four incremental product extensions:

1. observed historical replay foundation;
2. gap-aware telemetry and evidence timeline;
3. observed replay analytics summary;
4. shareable observation deep links;
5. adjacent-observation change inspection.

The stage does not add a new ingestion source, a new backend service, a new database, a new persistence model, a new map provider or a new paid dependency.

## Zero-budget architecture

The stage is intentionally implemented on data and infrastructure already present in Global Flight Analytics:

```text
persisted flight_states
        ↓
existing GET /api/v1/flights/{flightID}/states
        ↓
frontend API loader
        ↓
React Query
        ↓
deterministic replay model
        ↓
MapLibre + replay controls + evidence panels
```

Cost boundary:

```text
New paid aviation provider = NO
New backend endpoint        = NO
New server                  = NO
New database                = NO
New cache                    = NO
New persistent storage      = NO
New paid map service        = NO
New runtime dependency      = NO
Additional stage cost       = 0 RUB
```

The zero-cost statement means that Stage 15 itself introduces no new paid requirement. It does not promise that third-party free tiers, provider policies or hosting limits will remain unchanged forever.

## Canonical replay evidence contract

Historical Replay is bounded by the following rules:

- source records are already-persisted Flight State observations;
- replay loading requires a durable `flight_id`;
- states are constrained to the selected trajectory identity and time window;
- ICAO24 identity is checked rather than assumed from list position;
- observations are deterministically sorted and duplicate state identities are removed;
- the user-visible replay advances only between persisted samples;
- map replay samples are represented as observations, not invented continuous positions;
- gaps are evidence that the intermediate position is unknown;
- observed values, derived aggregates and projection outputs remain distinct concepts;
- no intermediate position is inferred;
- endpoint displacement is not travelled path distance.

Canonical markers:

```text
EVIDENCE_CLASS=OBSERVED
INTERPOLATION_POLICY=NONE
```

## Product increment ledger

| PR | Product increment | Exact final head | Resulting main SHA | Required PR evidence |
|---|---|---|---|---|
| #149 | Observed Historical Flight Replay | `9483d05dbb1828f1efd19f7d6e782c8b16f92939` | `30cee776abbabe8122908f92ddf2b7f777f4d110` | Frontend #401 / `34022210485`; Backend #742 / `34022210486`; CodeQL #381 / `34022210475`; API Load #279 / `34022210438`; Playwright #178 / `34022210410` — SUCCESS |
| #150 | Gap-aware replay telemetry | `f3db7c5045915b13c17797969d15896f1d206e74` | `d93cbb8537db93b1d8e069d1aa82c40080ecc95a` | Frontend #404 / `34024855683`; Backend #745 / `34024855676`; CodeQL #384 / `34024855744`; API Load #281 / `34024855777`; Playwright #181 / `34024855653` — SUCCESS |
| #151 | Observed replay analytics summary | `9c1496445d124137eccbc68b36ecc3ac4318f8c7` | `ce62d3612761a1c10986923d8ade29d859c1e243` | Frontend #406; Backend #747; CodeQL #386; API Load #282; Playwright #183; Vercel — SUCCESS |
| #152 | Shareable replay observation links | `3f7d06edfb7ff125b007fb6c822e6c9e87ea4dff` | `1c0eae7d9162d587ed17b8db2957070af1f94478` | Frontend #408; Backend #748; CodeQL #388; API Load #283; Playwright #185; Vercel — SUCCESS |
| #153 | Observed replay change inspector | `cc783b28410b029309eed5afc819608ab36441ee` | `6ed51f563818a619bb2346e6c4ad50d4256d60ec` | Frontend #410 / `34029175551`; Backend #749 / `34029175562`; CodeQL #390 / `34029175545`; API Load #284 / `34029175554`; Playwright #187 / `34029175541`; Vercel — SUCCESS |

## Real CI/remediation history

The stage had two repository-proven remediation events that must not be flattened into a generic “all green on first attempt” story.

### PR #149 — replay lifecycle ownership

The first replay implementation synchronously reset replay state inside a React `useEffect`. Frontend CI/ESLint rejected that lifecycle pattern. The rule was not disabled. Replay state was moved behind a keyed component boundary so React remount semantics own reset when the selected trajectory revision changes.

During the same feature review, a second edge case was identified before merge: the existing Trail visibility toggle could indirectly remove the trajectory context used by replay. The data source was separated from presentation visibility: raw trajectory data remains available to replay while `mapTrajectory` controls only the static trail presentation.

### PR #150 — strict E2E locator ownership

A Playwright run failed because the broad locator for `0.0 m/s` matched both the current velocity presentation (`230.0 m/s · 828 km/h`) and the vertical-rate value (`0.0 m/s`). The product UI was not weakened to satisfy the test. The test was corrected to use exact/scoped semantic ownership, and the final Playwright #181 run `34024855653` completed successfully.

For PRs #151-#153, no repository evidence shows an architectural rejection that required a second implementation. Their rejected alternatives below are retrospective/adversarial engineering analysis, not fabricated review events.

---

## STAGE15-R1 — Historical playback could visually invent intermediate aircraft positions

1. **Finding / symptom:** a conventional animated replay can imply a continuous historical path even when persistence contains only sparse observations.
2. **Root cause:** animation and map-line conventions naturally fill visual space between known endpoints unless the evidence model explicitly forbids it.
3. **Failure scenario:** persisted samples exist at 18:01 at point A and 18:04 at point B; a smooth marker or synthesized line makes the user believe the aircraft position at 18:02 and 18:03 is known.
4. **Impact:** the frontend can convert missing evidence into apparently observed history and undermine the project-wide observed/derived/projected trust boundary.
5. **Severity rationale:** **P1-equivalent retrospective product-evidence risk** because an invented historical position is a false factual claim, although no production incident was recorded.
6. **Existing guarantees violated:** Historical Intelligence must preserve evidence availability and must not present estimation as observation.
7. **Considered solutions:** client interpolation; server-generated intermediate points; a LineString connecting samples; or discrete persisted observations with explicit gaps.
8. **Chosen remediation:** replay advances only through persisted Flight State samples, uses observation points for historical positions, declares `OBSERVED` evidence and `NONE` interpolation, and later adds an explicit evidence-gap timeline.
9. **Why this solution was selected:** it is the only option that preserves source truth, reuses existing data, costs 0 RUB and requires no synthetic position model.
10. **Rejected alternatives:** interpolation, route reconstruction and unqualified lines between samples were rejected because they visually claim information that is not present in persistence.
11. **Trade-offs:** playback is less visually smooth and sparse provider coverage remains visible to the user.
12. **Regression tests / protection:** replay model tests; source-contract tests requiring observed-only/no-interpolation markers; Point-only replay map representation; aircraft-intelligence Playwright coverage.
13. **Adversarial review findings:** large time gaps are intentionally not filled; a visually attractive continuous path is considered less correct than a discontinuous evidence display.
14. **Remediation iterations:** PR #149 established observed-only playback; PR #150 made missing intervals first-class visible gap evidence.
15. **Residual risks and limitations:** replay quality is bounded by persisted sampling density and provider/source availability; missing observations cannot be reconstructed truthfully.
16. **Operational or deployment consequences:** frontend-only behavior over an existing endpoint; no new deployment service or persistence workload.
17. **Exact evidence:** PR #149 head `9483d05d...`, Playwright #178 `34022210410`; PR #150 head `f3db7c50...`, Playwright #181 `34024855653`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every future replay visualization must explicitly declare evidence class and interpolation policy before any continuous path or generated point can be introduced.

## STAGE15-R2 — Replay state lifecycle and Trail visibility could couple data truth to UI lifecycle

1. **Finding / symptom:** the initial replay state reset used synchronous state mutation in `useEffect`, and trajectory availability for replay could be accidentally coupled to the separate static Trail visibility toggle.
2. **Root cause:** lifecycle reset ownership and map presentation ownership were both initially located too close to presentation effects instead of immutable selection identity/data ownership.
3. **Failure scenario:** React selection changes trigger effect-driven reset behavior that becomes harder to reason about, or a user turns Trail off and unintentionally loses replay context although persisted states still exist.
4. **Impact:** replay behavior can become state-order dependent and an unrelated presentation preference can alter evidence availability.
5. **Severity rationale:** **P2 retrospective** because this is primarily correctness/maintainability risk rather than source-data corruption.
6. **Existing guarantees violated:** selection identity must deterministically own replay state; visual toggles must not mutate or hide underlying analytical evidence ownership.
7. **Considered solutions:** suppress the lint rule; retain reset effects with guards; manually synchronize multiple state variables; or let a keyed component remount own state reset and separate raw replay trajectory from map-trail visibility.
8. **Chosen remediation:** keyed remount by trajectory revision for replay state; raw trajectory remains the replay input while `mapTrajectory` represents only trail visibility.
9. **Why this solution was selected:** React remount semantics make reset deterministic without effect-driven state synchronization, and separating data from presentation prevents a UI toggle from becoming an evidence gate.
10. **Rejected alternatives:** disabling ESLint or adding effect exceptions was rejected because it would preserve the lifecycle smell rather than fix ownership.
11. **Trade-offs:** remount resets local playback controls when trajectory identity changes, which is intentional.
12. **Regression tests / protection:** Frontend lint/type checks, replay source contracts and browser journey coverage.
13. **Adversarial review findings:** Trail OFF must still allow replay data to exist; trajectory replacement must not leak the previous cursor into the new replay.
14. **Remediation iterations:** the initial PR #149 candidate was corrected after Frontend CI/ESLint objection; the Trail coupling was then removed before the final head `9483d05d...`.
15. **Residual risks and limitations:** local UI playback state is intentionally ephemeral and is not persisted across unrelated trajectory selection changes.
16. **Operational or deployment consequences:** no server or persistence changes.
17. **Exact evidence:** final PR #149 head `9483d05dbb1828f1efd19f7d6e782c8b16f92939`; Frontend CI #401 `34022210485` SUCCESS; final merge `30cee776abbabe8122908f92ddf2b7f777f4d110`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** new replay controls must keep evidence/data ownership independent from purely visual map toggles and must not bypass React/ESLint lifecycle rules for convenience.

## STAGE15-R3 — Telemetry E2E assertions could confuse identical textual values

1. **Finding / symptom:** the first PR #150 Playwright assertion for `0.0 m/s` matched more than one element after richer telemetry was added.
2. **Root cause:** the test encoded a textual coincidence instead of the semantic owner of the value.
3. **Failure scenario:** velocity text contains a substring that also equals vertical-rate text; Playwright strict mode rejects the ambiguous locator or, under a weaker runner, a test could assert the wrong card.
4. **Impact:** browser evidence becomes flaky or falsely proves the wrong telemetry field.
5. **Severity rationale:** **P2 retrospective testing-evidence risk** because the product data was correct but the acceptance evidence was ambiguous.
6. **Existing guarantees violated:** deterministic browser tests must identify semantic UI ownership rather than depend on incidental text uniqueness.
7. **Considered solutions:** disable strict locator behavior; use an arbitrary `nth()` selector; change production copy to make tests pass; or scope/exact-match the intended datum.
8. **Chosen remediation:** exact/scoped locator ownership, including `getByText('0.0 m/s', { exact: true })` for the intended value.
9. **Why this solution was selected:** it fixes the test contract without distorting product UI or weakening Playwright strictness.
10. **Rejected alternatives:** arbitrary index selectors and product-code workarounds were rejected because DOM order and copy are not semantic identity.
11. **Trade-offs:** tests are slightly more explicit and must be updated when semantic UI ownership intentionally changes.
12. **Regression tests / protection:** the corrected aircraft-intelligence browser journey runs in the permanent Playwright suite.
13. **Adversarial review findings:** repeated numeric values are expected in telemetry; tests must remain correct when two different fields legitimately show the same formatted number.
14. **Remediation iterations:** the initial Playwright #180 failure exposed ambiguity; the final locator-only correction produced Playwright #181 SUCCESS.
15. **Residual risks and limitations:** accessibility/semantic selectors still depend on intentional UI labels remaining stable.
16. **Operational or deployment consequences:** test-only correction; no runtime behavior changed for the remediation.
17. **Exact evidence:** PR #150 final head `f3db7c5045915b13c17797969d15896f1d206e74`; Playwright #181 run `34024855653` SUCCESS; Frontend #404 `34024855683` SUCCESS.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** E2E telemetry assertions must scope values by semantic owner or exact accessible text and must not weaken strict-mode behavior to hide ambiguity.

## STAGE15-R4 — Sparse observations could tempt unsupported flight-phase or predictive claims

1. **Finding / symptom:** once replay analytics existed, it would be easy to convert climb/descent, speed and gap aggregates into labels such as “climb phase”, “cruise” or predictive motion without enough evidence.
2. **Root cause:** derived telemetry summaries resemble inputs commonly used by flight-phase classifiers, but the free persisted dataset does not guarantee sufficient cadence or completeness for such classification.
3. **Failure scenario:** two or three sparse samples with changing altitude are labeled as a stable flight phase even though a large unobserved interval contains unknown behavior.
4. **Impact:** the product overstates analytical certainty and blurs the boundary between observed summaries and inferred aviation state.
5. **Severity rationale:** **P1-equivalent retrospective trust risk** because unsupported inference could be presented as factual flight behavior.
6. **Existing guarantees violated:** analytical outputs must expose their evidence basis and limitations; sparse source data must not silently become a stronger semantic claim.
7. **Considered solutions:** heuristic flight-phase labels; thresholds based only on current replay samples; model-based prediction; or descriptive aggregates only.
8. **Chosen remediation:** Stage 15 analytics remain evidence-derived descriptive aggregates: sample count/span, gaps, altitude coverage/range, peak observed velocity, climb/descent extrema and ground/airborne sample counts.
9. **Why this solution was selected:** these values are reproducible from persisted observations, useful to the user and require neither paid data nor invented classification confidence.
10. **Rejected alternatives:** automatic flight-phase and predictive labels are deferred until a separately governed evidence/calibration contract exists.
11. **Trade-offs:** the UI gives fewer high-level labels and requires the user to interpret descriptive evidence.
12. **Regression tests / protection:** replay analytics unit/source-contract tests and existing aircraft-intelligence browser journey.
13. **Adversarial review findings:** a high climb delta across a long gap is not proof that the aircraft continuously climbed throughout that gap.
14. **Remediation iterations:** no historical CI/review rejection occurred for this decision; PR #151 implemented the conservative aggregate-only design directly.
15. **Residual risks and limitations:** aggregates can still be unrepresentative when sample coverage is sparse; coverage/gap metrics must be read alongside extrema.
16. **Operational or deployment consequences:** frontend computation only; no materialization job or paid analytical service.
17. **Exact evidence:** PR #151 head `9c1496445d124137eccbc68b36ecc3ac4318f8c7`; Frontend #406, Backend #747, CodeQL #386, API Load #282 and Playwright #183 SUCCESS; merge `ce62d3612761a1c10986923d8ade29d859c1e243`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** any future flight-phase/prediction feature requires an explicit inference policy, calibration/evaluation evidence and limitation contract; descriptive replay analytics alone are not sufficient authorization.

## STAGE15-R5 — Shareable replay links must bind durable observation identity, not slider ordinal

1. **Finding / symptom:** a share link based on cursor index such as “sample 7” would not identify a durable evidence record.
2. **Root cause:** replay order is a presentation derived from sorting, filtering, deduplication and the set of states currently returned.
3. **Failure scenario:** new observations are added or ordering/deduplication changes, causing an old `cursor=7` link to restore a different observation than the one originally shared.
4. **Impact:** a research/debug link can silently change meaning and point to the wrong evidence.
5. **Severity rationale:** **P1-equivalent retrospective provenance risk** because a shareable evidence reference must not silently rebind to another record.
6. **Existing guarantees violated:** durable evidence references must use stable source identity rather than presentation position.
7. **Considered solutions:** slider ordinal; observed timestamp; compound timestamp/coordinate key; or persisted Flight State `id`.
8. **Chosen remediation:** use `replay_observation=<state.id>` in the URL and restore the cursor by matching that durable state identity.
9. **Why this solution was selected:** `state.id` already exists, costs nothing, survives replay reordering and maps directly to the persisted observation contract.
10. **Rejected alternatives:** ordinal links were rejected as unstable; timestamp-only links can be ambiguous and are weaker than the existing durable ID.
11. **Trade-offs:** the link can no longer resolve if the underlying persisted state becomes unavailable.
12. **Regression tests / protection:** replay deep-link model tests, URL/source contracts and browser assertion that the selected state ID is exposed/restored.
13. **Adversarial review findings:** adding an earlier observation must not cause an existing shared link to point at a different sample.
14. **Remediation iterations:** no historical review rejection occurred; PR #152 implemented durable ID binding as the first shareable-link contract.
15. **Residual risks and limitations:** links are only as durable as retained persistence and compatible trajectory/filter context; the project does not promise permanent archival URLs.
16. **Operational or deployment consequences:** browser URL/clipboard behavior only; no link-shortening service or new storage.
17. **Exact evidence:** PR #152 head `3f7d06edfb7ff125b007fb6c822e6c9e87ea4dff`; Frontend #408, Backend #748, CodeQL #388, API Load #283 and Playwright #185 SUCCESS; merge `1c0eae7d9162d587ed17b8db2957070af1f94478`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** any future share/export reference to replay evidence must use durable observation identity and must define behavior when the referenced record is unavailable.

## STAGE15-R6 — Endpoint displacement could be misrepresented as travelled route distance

1. **Finding / symptom:** comparing two persisted coordinates allows a great-circle endpoint distance to be calculated, but users can easily interpret a distance value as how far the aircraft actually travelled.
2. **Root cause:** two endpoints constrain straight/geodesic separation but do not reveal the continuous path taken between them.
3. **Failure scenario:** the aircraft turns or follows an airway between two observations, while the UI reports the shorter endpoint separation as if it were the flown path length.
4. **Impact:** historical movement analytics become mathematically precise-looking but semantically false.
5. **Severity rationale:** **P1-equivalent retrospective analytical-trust risk** because the derived value could be mistaken for a measured route quantity.
6. **Existing guarantees violated:** derived geometry must not be promoted to observed route evidence and missing intermediate positions must remain explicit.
7. **Considered solutions:** call the value distance travelled; connect samples and sum segments; infer a route from heading; or expose only endpoint-to-endpoint great-circle displacement with an explicit caveat.
8. **Chosen remediation:** the Observed Change Inspector labels the value as endpoint displacement and explicitly states that it is great-circle distance between two persisted positions, not travelled path distance; no intermediate position is inferred.
9. **Why this solution was selected:** it provides a useful reproducible comparison while preserving the exact limits of the available data and requiring no external route service.
10. **Rejected alternatives:** path-distance language, route inference and interpolation were rejected because the evidence needed to support them is absent.
11. **Trade-offs:** endpoint displacement cannot answer how far the aircraft actually flew between observations.
12. **Regression tests / protection:** model tests for displacement and deltas; absent-altitude handling; shortest heading-change wraparound test where `350° → 10°` equals `20°`; source-contract and E2E visibility checks.
13. **Adversarial review findings:** two identical endpoints can still hide movement between observations; therefore endpoint displacement must never be interpreted as complete route length.
14. **Remediation iterations:** no historical review rejection occurred; PR #153 implemented the qualified endpoint metric directly.
15. **Residual risks and limitations:** displacement reflects only endpoint geometry; altitude/velocity deltas are unavailable when either required observation lacks that evidence; heading delta does not prove the path continuously turned by that amount.
16. **Operational or deployment consequences:** deterministic frontend calculation only; no routing API or paid geospatial service.
17. **Exact evidence:** PR #153 head `cc783b28410b029309eed5afc819608ab36441ee`; Frontend #410 `34029175551`, Backend #749 `34029175562`, CodeQL #390 `34029175545`, API Load #284 `34029175554`, Playwright #187 `34029175541` — SUCCESS; merge `6ed51f563818a619bb2346e6c4ad50d4256d60ec`.
18. **Final canonical status:** **CLOSED**.
19. **Prevention / future guard:** every movement metric must state whether it is observed, endpoint-derived, interpolated or projected; path-distance claims require actual path evidence rather than endpoint geometry.

---

## Regression protection matrix

| Contract | Permanent protection |
|---|---|
| Replay is observed-only | `flight-replay-contract.test.mjs`, UI evidence attributes and E2E journey |
| No interpolation | source-contract assertions and Point-based replay representation |
| Replay ordering/cursor behavior | `flight-replay-model.test.mjs` |
| Gap semantics | model/source-contract tests and visible evidence-gap timeline |
| Telemetry ownership | TypeScript parser/model contract plus exact/scoped E2E assertions |
| Analytics remain descriptive | replay analytics model/source-contract tests |
| Share links bind durable ID | deep-link model/source-contract tests and browser journey |
| Endpoint displacement is not path distance | observed-change model/source-contract tests and explanatory UI copy |
| Heading wraparound | model test for `350° → 10° = 20°` |

## Final post-merge evidence

The completed Stage 15 implementation is present on canonical `main` at:

```text
6ed51f563818a619bb2346e6c4ad50d4256d60ec
```

Post-merge checks on that exact SHA:

```text
Frontend CI #411   run 34029950587   SUCCESS
CodeQL #391        run 34029950613   SUCCESS
Playwright E2E #188 run 34029950618  SUCCESS
Vercel                                   SUCCESS
```

Backend CI and API Load Baseline do not run on this frontend-only push under the repository path filters; both were part of the full PR #153 matrix before merge and completed successfully.

## Residual limitations

Stage 15 intentionally retains the following limitations:

- only persisted observations can be replayed;
- sparse provider/source coverage remains sparse rather than being synthesized;
- there is no guarantee of continuous aircraft position between samples;
- replay is not ATC-grade, navigation-grade or safety-critical evidence;
- endpoint displacement is not route/path distance;
- descriptive replay analytics do not establish flight phase or intent;
- a shared observation link cannot resolve if its underlying state is no longer available;
- browser clipboard behavior remains subject to browser permission/security context;
- the stage does not buy commercial historical coverage or satellite data;
- future free-tier/provider/retention limits can reduce data availability without changing the correctness of the evidence semantics.

## Cost ceiling policy

Stage 15 remains inside the project-wide zero-budget rule.

If a future replay feature cannot be implemented truthfully using existing persisted data, open/free data, client/server computation on the current free infrastructure, or available free tiers, that feature must stop at design review. The documentation must record which missing capability requires money, what is lost by staying free, and the approximate paid dependency before implementation is authorized.

## Future guard for Stage 15 extensions

Every future Historical Flight Replay feature must document, at minimum:

1. source data used;
2. evidence class (`OBSERVED`, `DERIVED`, `ESTIMATED`, or `PROJECTED` as applicable);
3. interpolation policy;
4. identity/provenance key;
5. behavior under missing/sparse evidence;
6. new API/storage/infrastructure requirements;
7. incremental monetary cost;
8. regression protection;
9. residual limitations.

A feature that cannot answer these questions is not ready to merge.

## Closure statement

Historical Flight Replay PRs #149-#153 are now covered by a canonical product/evidence engineering history that records real CI remediations, explicitly labels retrospective adversarial analysis, preserves the zero-budget boundary, and protects the distinction between observed evidence and unsupported inference.

```text
STAGE_15_HISTORICAL_FLIGHT_REPLAY=CLOSED
STAGE_15_DOCUMENTATION=CLOSED
```
