# Document 173 — Core Flight Data Ingestion Production Closure

Status: CLOSED — canonical remediation history reconciled  
Project: Global Flight Analytics  
Scope: free production scheduling, bounded ingestion execution, freshness evidence, and honest runtime boundaries  
Initial production-ingestion PR: #35, head `64e4d17d914371f789992065e0ececf9eaa25161`, merge `eea53a2ac7636c024903522047d03660e1db86dd`  
Coverage / live-quality repair PR: #36, head `f6ca94f40115a75c8fb4698b336265ef46890b94`, merge `f5f7e9cb4e4bb1075a61b284641788984f0a2a67`

---

## 1. Audit Finding

The source implementation already contained a mature ingestion pipeline:

```text
provider policy and budget
provider health and fallback
canonical FlightState mapping
normalization and validation
exact duplicate removal
track construction
PostgreSQL persistence
durable ingestion runs
partial and failed terminal status
stale-run recovery
reconciliation for derived-write failures
```

The production audit on 2026-08-03 found a deployment gap instead of a missing domain implementation.

```text
PRODUCTION_API_REVISION=5c1c0862581842a78c323f5581c1425641b2b363
PRODUCTION_API_SOURCE_REVISION_MATCH=YES
PRODUCTION_TRAFFIC_SAMPLE_1 count=3
PRODUCTION_TRAFFIC_SAMPLE_2 count=3
PRODUCTION_TRAFFIC_ADVANCEMENT=NOT_OBSERVED
RENDER_INGEST_SERVICE_DECLARED=NO
STAGE_5_PRODUCTION_INGESTION_RUNTIME=NOT_DECLARED
```

The newest production observation was from 2026-07-14, approximately nineteen and one-half days before the audit. The API was healthy and revision-correct, but the database was not receiving new traffic observations.

---

## 2. Root Cause

The production container includes `/app/ingest`, but the Render Blueprint starts only `/app/server`.

A second continuously running Render service would require a non-free service type. The project permanently excludes paid production infrastructure from this portfolio closure, so the fix did not add a Render background worker or paid cron job.

---

## 3. Free Production Runtime

The first free production ingestion runtime was:

```text
GitHub Actions schedule
↓
serialized production ingestion job
↓
go run ./cmd/ingest --once
↓
free external traffic provider
↓
canonical processing pipeline
↓
Neon PostgreSQL
↓
Render read-only API
↓
public freshness verification
```

The scheduled workflow ran every ten minutes and could also be dispatched manually. Only one production ingestion run could execute at a time; an active database write was not cancelled by a newer scheduled event.

Later production reliability work replaced/hardened this scheduling boundary with Cloudflare primary/watchdog orchestration plus GitHub fallback. That later evolution does not reopen the historical missing-runtime finding recorded here.

---

## 4. Bounded Command Contract

The existing `ingest` command keeps daemon mode as its default behavior.

```text
go run ./cmd/ingest
```

The production scheduler uses explicit one-shot mode:

```text
go run ./cmd/ingest --once
```

One-shot mode:

```text
loads the same production configuration
uses the same provider policy, budget and health controls
recovers stale durable runs
executes exactly one ingestion cycle
records the same cycle and provider evidence
returns a non-zero exit code on cycle failure
schedules no retry delay and starts no long-running metrics listener
exits after the cycle reaches a terminal result
```

It does not create an alternate ingestion implementation.

---

## 5. Secret Boundary

The scheduled workflow requires one GitHub Actions repository secret:

```text
PRODUCTION_INGESTION_DATABASE_URL
```

The value is the owner-controlled Neon PostgreSQL connection string used for bounded production writes. It must never be committed, printed, stored in artifacts, or passed through pull-request events.

The workflow has read-only repository permissions. It runs only from `schedule` and `workflow_dispatch`; it is not exposed to untrusted pull-request code.

---

## 6. Freshness Gate

After one-shot ingestion completes, the workflow queries the public production endpoint:

```text
GET /api/v1/traffic/current
```

The gate requires:

```text
an HTTP success response
a valid application response envelope
at least one aircraft record
a valid observed_at timestamp on every returned item
the newest observation to be no older than thirty minutes
future timestamp skew to remain bounded
```

Stable success evidence:

```text
PRODUCTION_TRAFFIC_FRESHNESS=PASS
```

A successful ingestion process without fresh public data is not accepted as production closure.

### Validated regional query radius

The first production runtime activation on 2026-08-03 proved that the original one-hundred-nautical-mile query completed successfully but returned zero aircraft around Baku. A direct provider diagnostic immediately afterward returned:

```text
radius 100 nautical miles: total=0
radius 250 nautical miles: total=1
provider HTTP status: 200
provider message: No error
```

The production workflow therefore moved to a two-hundred-fifty-nautical-mile radius. This was a coverage correction based on observed provider behavior, not a claim of guaranteed regional availability.

### Live quality parent identity

The first two-hundred-fifty-nautical-mile runtime attempt exposed a separate persistence contract defect. Live provider states do not carry application-assigned UUID values. PostgreSQL generates the canonical `flight_states.id`, while the original quality write attempted to find the parent only through the empty incoming identifier.

The live quality repository now resolves the canonical parent through either:

```text
persisted flight state UUID when one is already known
or
the unique source_name + icao24 + observed_at observation identity
```

The resulting report still stores the actual canonical `flight_states.id` in both parent columns. Foreign-key enforcement and rejected-state separation remain unchanged.

---

## 7. Schedule Limitations

GitHub scheduled workflows are best-effort infrastructure. They may start later than the nominal cron time, and inactive public repositories may have scheduled workflows disabled by the platform.

Therefore the product must not claim:

```text
continuous ten-minute execution
guaranteed real-time coverage
guaranteed provider availability
operational surveillance continuity
safety-critical freshness
```

The thirty-minute freshness threshold is an explicit portfolio-health boundary, not an operational aviation service-level agreement.

---

## 8. Verification

Source verification:

```bash
pnpm run test:production-ingestion-contract
pnpm run verify:production-ingestion-contract
```

Focused Go verification:

```bash
cd apps/api
go test ./cmd/ingest
```

PR #35 exact-head CI on `64e4d17d914371f789992065e0ececf9eaa25161`:

```text
Backend CI   30814963325 = SUCCESS
Frontend CI  30814963347 = SUCCESS
CodeQL       30814963330 = SUCCESS
```

PR #36 exact-head CI on `f6ca94f40115a75c8fb4698b336265ef46890b94`:

```text
Backend CI   30834469585 = SUCCESS
Frontend CI  30834473086 = SUCCESS
CodeQL       30834472732 = SUCCESS
```

Later live production closure evidence in Document 183 records successful ingestion runs, `PRODUCTION_TRAFFIC_FRESHNESS=PASS`, and final exact-revision runtime validation on `7dfc66685247a5a1aaea87b1391624d1014d7013`.

---

## 9. Completion Boundary

The historical source/runtime defects in this document are finding-level CLOSED. That status means the repository acquired a bounded free ingestion execution path, the validated regional coverage correction, and a live Data Quality parent-resolution path with exact source/CI/runtime evidence.

It does **not** mean production ingestion can never be intentionally stopped or later require provider/infrastructure recovery. Later Documents 182–195 own subsequent scheduling, free-tier reliability, provider recovery, and current runtime state. A later operational outage or provider recovery state does not automatically reopen these historical implementation findings unless the same implementation defect recurs.

---

## 10. Canonical remediation record — GFA-OPS-447

### 1. Finding / symptom

Production deployed the read-only API but no production ingestion runtime was declared, so live traffic observations stopped advancing while the API remained healthy.

### 2. Root cause

The production image contained the ingestion command, but the Render Blueprint started only `/app/server`. The free-infrastructure constraint also ruled out adding a second continuously running paid Render worker.

### 3. Failure scenario

The API could stay revision-correct and return HTTP success while PostgreSQL retained increasingly stale aircraft observations because no production process executed the existing ingestion pipeline.

### 4. Impact

Live-traffic functionality degraded into stale historical output, invalidating freshness expectations and any product surface that depended on advancing current observations.

### 5. Severity rationale

**P1 retrospective.** The defect disabled the production data-refresh path while leaving the public API apparently healthy, creating a material production correctness/availability gap. The classification is retrospective; no original severity label is claimed.

### 6. Existing guarantees violated

- production current-traffic data must have an explicit execution owner;
- healthy API process state must not be treated as evidence that source observations are advancing;
- free-tier architecture must still provide a bounded supported ingestion path;
- production closure requires public freshness evidence, not command existence alone.

### 7. Considered solutions

- add a paid continuously running Render worker;
- embed background ingestion inside the public API process;
- run the existing pipeline from a serialized free GitHub Actions schedule using one-shot mode;
- leave ingestion manual-only.

### 8. Chosen remediation

Add explicit `ingest --once` execution, a serialized scheduled/workflow-dispatch GitHub Actions production ingestion workflow, a dedicated repository-secret boundary, and a post-run public freshness gate. Reuse the existing ingestion pipeline rather than creating a second implementation.

### 9. Why this solution was selected

It restored a production execution owner within the project's free-infrastructure constraint, preserved separation between read API and write ingestion, and made freshness externally testable.

### 10. Rejected alternatives

- a paid Render worker was rejected by the zero-cost portfolio constraint;
- background ingestion inside the API was rejected because it couples read-serving lifecycle and scheduled writes;
- manual-only execution was rejected because it does not provide a production refresh contract.

### 11. Trade-offs

GitHub scheduling is best effort and can be delayed or disabled by platform behavior. The workflow therefore cannot claim hard real-time cadence or operational surveillance continuity.

### 12. Regression tests / protection

Production-ingestion contract tests verify one-shot execution, serialized scheduling, secret ownership, source configuration, and public freshness verification. Backend CI executes the production-ingestion contract verifier.

### 13. Adversarial review findings

The review distinguished a missing production execution path from a missing ingestion domain implementation. It rejected duplicating the pipeline or adding paid infrastructure and required freshness evidence after execution.

### 14. Remediation iterations

PR #35 introduced the free production runtime and one-shot command. PR #36 then corrected runtime-discovered regional coverage and a separate live Data Quality persistence defect. Later reliability work added Cloudflare orchestration and GitHub fallback without removing the bounded ingestion command.

### 15. Residual risks and limitations

Provider availability, GitHub/Cloudflare scheduling, Neon quotas, and later free-tier operational recovery remain external runtime concerns. Scheduled execution is not safety-critical or guaranteed real-time service.

### 16. Operational or deployment consequences

Production requires the owner-controlled database secret and an enabled scheduler/orchestrator. Freshness must be verified after activation and after material runtime changes.

### 17. Exact evidence

- production audit revision: `5c1c0862581842a78c323f5581c1425641b2b363`;
- PR #35 head: `64e4d17d914371f789992065e0ececf9eaa25161`;
- PR #35 merge: `eea53a2ac7636c024903522047d03660e1db86dd`;
- PR #35 Backend CI `30814963325` — SUCCESS;
- PR #35 Frontend CI `30814963347` — SUCCESS;
- PR #35 CodeQL `30814963330` — SUCCESS;
- later live closure revision: `7dfc66685247a5a1aaea87b1391624d1014d7013`;
- Document 183 records successful production ingestion and `PRODUCTION_TRAFFIC_FRESHNESS=PASS`.

### 18. Final canonical status

**CLOSED.**

### 19. Prevention / future guard

The production-ingestion contract verifier, public freshness gate, and later production-reliability validators prevent source/runtime closure from being inferred merely from API liveness or the existence of an ingestion binary.

---

## 11. Canonical remediation record — GFA-OPS-448

### 1. Finding / symptom

The initial production ingestion query used a 100-nautical-mile Baku radius that completed successfully at the provider transport level but returned zero aircraft, causing the public freshness gate to fail.

### 2. Root cause

The original production coverage radius was selected before live regional provider behavior was validated. Transport success was possible even when the chosen geographic scope contained no current provider observations.

### 3. Failure scenario

A scheduled ingestion cycle could complete without provider error and persist no current aircraft. Without the separate freshness gate this could look operationally successful while current traffic remained stale.

### 4. Impact

Production traffic availability for the target region failed despite healthy provider transport and ingestion execution. The freshness gate correctly prevented false closure, but the configured coverage policy was unusable for observed conditions.

### 5. Severity rationale

**P2 retrospective.** The defect caused production data unavailability, but the freshness gate detected it and prevented a false success claim. No evidence shows corrupted persisted observations or a safety-critical impact.

### 6. Existing guarantees violated

- production coverage configuration must be validated against observed provider behavior;
- successful provider HTTP execution is not equivalent to usable regional data;
- production closure requires at least one fresh public observation within the documented portfolio-health boundary.

### 7. Considered solutions

- keep the 100-NM radius and accept empty runs;
- increase the radius without diagnostic evidence;
- run direct provider diagnostics at multiple bounded radii and adopt the smallest observed working scope;
- remove the freshness gate.

### 8. Chosen remediation

Use direct provider evidence showing 100 NM returned zero aircraft and 250 NM returned one aircraft with HTTP 200/no provider error, then change the production workflow to 250 NM and lock that value into production-ingestion contract tests.

### 9. Why this solution was selected

It corrected the concrete observed coverage failure while retaining an explicit bounded regional query and preserving the independent freshness gate.

### 10. Rejected alternatives

- accepting empty successful cycles was rejected because transport success does not satisfy production data freshness;
- removing the freshness gate was rejected because it would hide recurrence;
- claiming 250 NM as guaranteed coverage was rejected because provider availability remains variable.

### 11. Trade-offs

A larger radius increases the geographic scope and potentially returned aircraft volume. It still cannot guarantee aircraft availability at every run.

### 12. Regression tests / protection

Production-ingestion contract tests require `TRAFFIC_INGESTION_RADIUS: '250'`; public freshness verification remains a separate post-ingestion gate.

### 13. Adversarial review findings

The review treated the zero-aircraft response as a production configuration failure rather than a provider HTTP failure and preserved the distinction between validated observed coverage and guaranteed service coverage.

### 14. Remediation iterations

The 100-NM runtime failure occurred after PR #35 activation. PR #36 changed the production radius to 250 NM and carried the runtime evidence into the closure document and contract tests.

### 15. Residual risks and limitations

The 250-NM value is historical validated policy, not a permanent guarantee. Provider coverage and regional aircraft presence can change; later provider recovery/policy documents own subsequent runtime configuration changes.

### 16. Operational or deployment consequences

Production ingestion must use a region/provider configuration that produces usable observations and must continue checking public freshness after each bounded run.

### 17. Exact evidence

- PR #36 head: `f6ca94f40115a75c8fb4698b336265ef46890b94`;
- PR #36 merge: `f5f7e9cb4e4bb1075a61b284641788984f0a2a67`;
- PR #36 body records 100 NM = zero aircraft and 250 NM = one aircraft with provider HTTP 200/no error;
- PR #36 Backend CI `30834469585` — SUCCESS;
- PR #36 Frontend CI `30834473086` — SUCCESS;
- PR #36 CodeQL `30834472732` — SUCCESS;
- later Document 183 live closure records fresh production traffic PASS.

### 18. Final canonical status

**CLOSED.**

### 19. Prevention / future guard

Keep production coverage configuration in the contract verifier and preserve the independent end-to-end freshness gate so a future successful-but-empty provider run cannot be mistaken for production closure.

---

## 12. Canonical remediation record — GFA-DB-449

### 1. Finding / symptom

Live Data Quality persistence could not associate a provider state with its canonical `flight_states` parent when PostgreSQL generated the parent UUID and the incoming live state had no application-assigned ID.

### 2. Root cause

The original quality write path resolved the parent only through `state.ID`. Live provider states were persisted first with database-generated UUIDs, so the in-memory state used for the subsequent quality write could still contain an empty identifier.

### 3. Failure scenario

A valid live state could be persisted successfully, then its Data Quality report write could fail because the repository searched for a parent by an empty application identifier even though the canonical parent row existed.

### 4. Impact

Production ingestion could produce incomplete derived quality evidence or fail the ingestion cycle after the primary observation write, reducing consistency between canonical flight states and their quality reports.

### 5. Severity rationale

**P1 retrospective.** The defect crossed a durable parent/derived-write integrity boundary during live production ingestion. It could leave the canonical observation persisted while its associated Data Quality evidence failed to persist.

### 6. Existing guarantees violated

- every accepted live quality report must reference the actual canonical persisted `flight_states.id`;
- derived writes must support the identity lifecycle used by live provider persistence;
- parent lookup must not require an application-assigned UUID when the database owns UUID generation;
- both quality parent columns must resolve to the same canonical parent.

### 7. Considered solutions

- assign application UUIDs to every live provider state before persistence;
- return/mutate the persisted UUID through every upstream object;
- retain ID lookup when known and add deterministic lookup by the existing unique observation identity `source_name + icao24 + observed_at` when the ID is absent;
- weaken the foreign-key/parent contract.

### 8. Chosen remediation

Keep direct UUID lookup for already-identified states and add a fallback insert/select path that resolves the persisted parent through the unique source/ICAO24/observation-time identity. Persist the resulting canonical UUID into both Data Quality parent columns.

### 9. Why this solution was selected

It matches the existing database identity contract, avoids inventing a second UUID owner, and uses an already-enforced unique observation identity to recover the canonical persisted parent safely.

### 10. Rejected alternatives

- forcing application UUID assignment was rejected because PostgreSQL already owns canonical ID generation for this live path;
- weakening foreign keys was rejected because it would reduce integrity;
- blindly mutating upstream objects after insert would increase coupling across repository boundaries.

### 11. Trade-offs

The fallback depends on the uniqueness and canonical semantics of `source_name + icao24 + observed_at`. If that identity contract changes, parent resolution must change with it.

### 12. Regression tests / protection

PR #36 adds a PostgreSQL integration fixture that inserts a live parent with a database-generated UUID, saves quality from an in-memory state with no ID, and asserts both stored parent UUID columns equal the actual persisted parent.

### 13. Adversarial review findings

The runtime failure demonstrated that source-level tests based only on application-assigned IDs were insufficient. The remediation was validated against the exact database-generated identity lifecycle used by live ingestion.

### 14. Remediation iterations

The defect was exposed by the first 250-NM production runtime attempt after the coverage correction. The parent-resolution fix landed in the same PR #36 merge `f5f7e9cb4e4bb1075a61b284641788984f0a2a67`.

### 15. Residual risks and limitations

The fallback assumes the observation identity remains unique and uses the same canonical source-name/ICAO24/timestamp semantics as the parent insert. Later schema changes must preserve or explicitly replace that guarantee.

### 16. Operational or deployment consequences

Live ingestion can persist Data Quality evidence for database-generated parent IDs without requiring a second application-owned UUID lifecycle. PostgreSQL uniqueness and foreign-key enforcement remain mandatory.

### 17. Exact evidence

- PR #36 head: `f6ca94f40115a75c8fb4698b336265ef46890b94`;
- PR #36 merge: `f5f7e9cb4e4bb1075a61b284641788984f0a2a67`;
- PR #36 diff adds `TestDataQualityRepositoryResolvesDatabaseGeneratedLiveParentByObservationIdentity` and the observation-identity repository path;
- PR #36 Backend CI `30834469585` — SUCCESS;
- PR #36 Frontend CI `30834473086` — SUCCESS;
- PR #36 CodeQL `30834472732` — SUCCESS;
- later Document 183 production ingestion runs completed successfully with fresh public traffic evidence.

### 18. Final canonical status

**CLOSED.**

### 19. Prevention / future guard

Retain the PostgreSQL integration test for database-generated live parent identity and the production freshness/reliability gates so future identity-lifecycle drift is detected before release closure.
