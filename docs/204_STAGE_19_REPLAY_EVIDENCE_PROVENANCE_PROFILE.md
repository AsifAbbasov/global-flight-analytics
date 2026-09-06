# Stage 19 Replay Evidence Provenance Profile

Status: implementation in progress; pre-merge validation pending

```text
STAGE_19_REPLAY_EVIDENCE_PROVENANCE=IN_PROGRESS
STAGE_19_BASE_MAIN=cbe6905a96f0656d868e328ba7c6ab38e756cfd8
STAGE_19_EVIDENCE_CLASS=PERSISTED_SOURCE_LABELS_ONLY
STAGE_19_PROVIDER_RANKING=NONE
STAGE_19_PROVIDER_ACCURACY_SCORE=NONE
STAGE_19_PROVIDER_SWITCH_TIME_INFERENCE=NONE
STAGE_19_NEW_BACKEND_DATA=NONE
STAGE_19_ADDITIONAL_COST=0_RUB
STAGE_19_PRE_MERGE_CI=PENDING
STAGE_19_POST_MERGE_CI=PENDING
```

## Product need

Stages 15 through 18 made Historical Flight Replay evidence-aware:

```text
Stage 15 — observed-only replay, gap timeline, analytics, shareable observations
Stage 16 — real elapsed-time navigation without position interpolation
Stage 17 — arbitrary observed interval comparison
Stage 18 — descriptive replay sampling-density and gap-quality profile
```

After Stage 18, a user can understand how dense or sparse the persisted replay evidence is, but one important evidence question is still hidden:

> Which persisted sources contributed the observations in this replay, and do adjacent observations ever carry different source labels?

The existing replay detail view shows `source_name` for only the currently selected observation. That is insufficient for understanding the provenance composition of the entire replay window.

This matters because the production ingestion architecture can select different permitted free traffic providers over time. Document 38 explicitly preserves the actual selected source on successful ingestion. Historical replay already transports `source_name` on every `FlightReplayPoint`, so the browser possesses the evidence required for a replay-wide provenance description.

## Source data and evidence boundary

Stage 19 uses only fields already present in the loaded replay payload:

```text
FlightReplayPoint.source_name
FlightReplayPoint.observed_at
FlightReplayPoint.id
ordered persisted replay points
```

Evidence classification:

```text
source label                       OBSERVED / PERSISTED METADATA
sample count by source             DERIVED FROM PERSISTED LABELS
sample share by source             DERIVED FROM PERSISTED LABELS
adjacent source-label transition   DERIVED FROM TWO PERSISTED LABELS
transition endpoint timestamps     OBSERVED / PERSISTED METADATA
transition elapsed interval        DERIVED FROM PERSISTED TIMESTAMPS
provider quality or accuracy       NOT CLAIMED
exact provider switch instant      NOT CLAIMED
```

No source ownership is projected across an unobserved interval. If two adjacent observations carry different source labels, the system can truthfully say only that the persisted labels differ at those two observed endpoints. It cannot know when the upstream provider changed between them.

## Product behavior

Stage 19 adds a Replay Evidence Provenance Profile that exposes:

- total persisted sample count;
- number of identified source labels;
- number of observations without a usable source label;
- source composition by persisted sample count;
- sample-share percentage for each identified source;
- first and last persisted observation timestamps carrying each source label;
- adjacent persisted observation pairs whose identified source labels differ;
- exact endpoint timestamps for each such source-label transition;
- elapsed time between those persisted endpoints when timestamps are valid;
- an interactive jump to the later persisted observation of a source-label transition.

The profile explicitly states that source percentages are shares of persisted samples, not shares of elapsed time.

## Problem and root cause

The replay payload already preserved per-observation source provenance, but product presentation remained point-local. A user had to manually advance through every observation and remember source labels to detect source mixing.

The root cause is therefore not missing backend data. It is missing replay-wide composition over data that is already loaded and validated.

## Failure and adversarial scenarios

### Scenario A — provider ranking disguised as provenance

Bad implementation:

```text
ADSB.lol = 92 / 100
OpenSky = 84 / 100
BEST SOURCE = ADSB.lol
```

Why rejected: the project has no calibrated dataset proving a universal provider-accuracy ranking for this replay. A source label is provenance, not a quality grade.

### Scenario B — elapsed-time share inferred from sample share

Bad implementation:

```text
70% of samples are source A
therefore source A covered 70% of replay time
```

Why rejected: sparse and irregular sampling makes sample share different from elapsed-time ownership. Stage 19 reports sample share only.

### Scenario C — exact switch time fabricated inside a gap

Observed evidence:

```text
17:46:00 source A
17:51:00 source B
```

Bad claim:

```text
provider switched at 17:48:30
```

Why rejected: no persisted observation proves the switch instant. The only defensible statement is that the adjacent observed endpoints carry different source labels.

### Scenario D — unattributed point bridges two providers

Observed labels:

```text
source A
unavailable
source B
```

Bad implementation: count this as a direct A → B transition.

Chosen behavior: unattributed evidence breaks transition continuity. No direct source transition is fabricated across unknown provenance.

### Scenario E — second network path for already loaded data

Bad implementation: fetch provider metadata or replay provenance from a new endpoint merely to compute counts already available from `FlightReplayPoint[]`.

Why rejected: increases latency, failure surface, request volume and architectural duplication without adding evidence.

## Considered solutions

### Option 1 — keep only the current observation source

Rejected. It preserves truth but does not solve replay-wide provenance visibility.

### Option 2 — add a backend provenance aggregate endpoint

Rejected for Stage 19. The browser already has the exact bounded replay points and source labels. A second server aggregate would duplicate a deterministic presentation calculation and add contract and operational cost.

### Option 3 — assign provider grades or preferred-provider ranking

Rejected. No calibration evidence supports such a claim.

### Option 4 — infer source ownership continuously between observations

Rejected. It would convert sparse endpoint metadata into unsupported continuous provenance.

### Option 5 — deterministic frontend provenance profile over persisted replay points

Selected.

It is the smallest truthful vertical slice and preserves the existing data and cost boundaries.

## Chosen architecture

```text
existing Historical Replay HTTP response
        ↓
FlightReplayPoint[]
        ↓
source_name + observed_at
        ↓
flight-replay-evidence-provenance-model.ts
        ↓
FlightReplayEvidenceProvenance
        ↓
existing FlightReplayTimeNavigation
```

No provider, backend, PostgreSQL, ingestion or cache change is required.

The model owns only deterministic provenance composition. The component owns presentation and optional navigation to an already persisted observation. Existing elapsed-time replay semantics remain authoritative for cursor movement.

## Missing-data policy

A source label is considered identified only when `source_name.trim()` is non-empty.

Unattributed observations:

- remain part of total persisted sample count;
- remain part of the denominator used for identified source sample shares;
- increment `unattributedSampleCount`;
- do not create a named source summary;
- do not create a direct transition through unknown provenance.

This prevents the UI from inventing a provider name or silently deleting unknown provenance from the evidence denominator.

## Percentage semantics

For each identified source:

```text
sampleSharePercent = source persisted samples / all persisted replay samples × 100
```

The displayed percentage is rounded for presentation.

It is explicitly not:

```text
elapsed-time coverage
provider uptime
request success rate
provider accuracy
provider reliability score
traffic completeness
```

## Transition semantics

A transition is recorded only when two adjacent persisted replay observations both have identified source labels and those labels differ.

For a transition:

```text
from source = earlier adjacent persisted point source_name
to source   = later adjacent persisted point source_name
start time  = earlier persisted observed_at
end time    = later persisted observed_at
elapsed     = timestamp difference when both timestamps parse successfully
```

The feature does not assign a switch timestamp between those endpoints.

## Regression protection

Stage 19 adds:

- `apps/web/tests/flight-replay-evidence-provenance-model.test.mjs` executing the real compiled production model;
- `apps/web/tests/stage-19-replay-evidence-provenance-contract.test.mjs` protecting source-only, no-ranking and no-second-data-path semantics;
- explicit inclusion of the production provenance model in `apps/web/tsconfig.test.json`;
- browser coverage in the existing aircraft-intelligence Playwright journey;
- a documentation contract protecting product purpose, evidence boundaries, zero-budget scope and residual limitations.

Critical regression guards include:

```text
SECOND_FETCH_PATH=FORBIDDEN
PROVIDER_RANKING=FORBIDDEN_WITHOUT_CALIBRATION
PROVIDER_ACCURACY_SCORE=FORBIDDEN_WITHOUT_CALIBRATION
ELAPSED_TIME_SHARE_FROM_SAMPLE_SHARE=FORBIDDEN
EXACT_SWITCH_TIME_INFERENCE=FORBIDDEN
UNKNOWN_SOURCE_BRIDGING=FORBIDDEN
```

## Expected result

A user inspecting Historical Replay can now distinguish, without manually stepping through every observation:

- a replay whose persisted observations all come from one identified source;
- a replay whose observations are composed from multiple identified sources;
- observations with unavailable source attribution;
- exact adjacent observed endpoints where source labels differ.

The feature improves evidence explainability without converting provenance into a provider-quality claim.

## Infrastructure and monetary impact

```text
New aviation provider       NO
New backend endpoint        NO
New PostgreSQL table        NO
New database migration      NO
New ingestion path          NO
New server                  NO
New cache                   NO
New persistent storage      NO
New runtime dependency      NO
Additional external calls   NO
Additional cost             0 RUB
```

## Residual limitations

Stage 19 intentionally does not solve or claim:

- provider-level accuracy measurement;
- calibrated provider reliability ranking;
- request-level fallback reason reconstruction from replay points;
- exact provider switch time between persisted observations;
- continuous source ownership across an unobserved interval;
- identification of a source when persisted `source_name` is absent;
- proof that one provider observed every aircraft state available in reality;
- denser historical sampling than the project actually persisted;
- commercial or satellite aviation coverage.

A source label says where the persisted observation is attributed. It does not certify that observation as more accurate than another provider's observation.

## Continuous Integration evidence

Pre-merge Continuous Integration evidence does not exist yet and must not be fabricated.

Current status:

```text
Frontend CI       PENDING
Backend CI        PENDING
CodeQL            PENDING
API Load Baseline PENDING
Playwright E2E    PENDING
Vercel            PENDING
```

Real failures, root causes, rejected fixes, remediation commits and successful reruns must be appended here only after they actually occur.

## Future guard

Any future replay provenance feature must declare before implementation:

1. exact persisted provenance fields it consumes;
2. whether the output describes samples, requests, ingestion runs or elapsed time;
3. missing-provenance behavior;
4. whether source transitions are observed directly or merely bounded by adjacent observations;
5. whether any ranking or quality claim is calibrated;
6. what evidence supports that calibration;
7. whether a new backend or provider data path is actually required;
8. regression protection;
9. infrastructure impact;
10. monetary cost.

Provider provenance must never be silently converted into provider quality, accuracy, completeness or preference.

## Current stage status

```text
STAGE_19_PRODUCT_IMPLEMENTATION=IN_PROGRESS
STAGE_19_DOCUMENTATION=PREMERGE_DRAFT
STAGE_19_PRE_MERGE_CI=PENDING
STAGE_19_POST_MERGE_CI=PENDING
STAGE_19_REPLAY_EVIDENCE_PROVENANCE=IN_PROGRESS
```
