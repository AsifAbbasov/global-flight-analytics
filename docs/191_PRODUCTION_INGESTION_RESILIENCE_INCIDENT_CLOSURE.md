# Production Ingestion Resilience Incident Closure

Status: Incident containment complete; production provider recovery remains open
Date: 2026-08-20
Scope: Production traffic ingestion scheduling, provider fallback, and recovery control

---

## 1. Executive Summary

Global Flight Analytics experienced a production ingestion incident in which the
configured `airplanes.live` traffic provider returned HTTP `403 Unauthorized`.
The ingestion command correctly treated the request as failed, so successful
ingestion timestamps stopped advancing and production freshness could no longer
be maintained.

The incident was amplified by the recovery scheduler. The deployed Cloudflare
Worker could repeatedly dispatch `Production Traffic Ingestion` because its
deduplication logic protected against active runs and recent successful runs but
did not treat a recent failed run as a reason to stop retrying. A deterministic
provider authorization failure therefore became a repeated dispatch/failure loop.

The remediation deliberately did not convert the provider error into a false
success. Instead, the system now fails closed at the scheduling boundary, applies
a bounded recent-failure circuit breaker, and treats provider-level unauthorized
responses as recoverable only within a multi-provider orchestration chain while
preserving the original unauthorized evidence.

Production ingestion remains intentionally disabled until an approved provider
path is proven in production.

---

## 2. User-Visible and Operational Symptoms

Observed symptoms included:

- repeated failed `Production Traffic Ingestion` GitHub Actions runs;
- repeated email notifications for the same deterministic ingestion failure;
- ingestion runs ending during the bounded production ingestion cycle;
- `airplanes.live` health evidence reporting `unauthorized`;
- traffic freshness remaining stale because no successful ingestion completed.

Representative provider evidence:

```text
provider=airplanes.live
status=unavailable
latest_outcome=unauthorized
reasons=[
  provider_authentication_rejected
  provider_has_no_successful_requests
]
```

The failure was not caused by PostgreSQL connectivity, migration execution,
GitHub Actions setup, or the owner's local Mac. The production request reached
the provider and was rejected with HTTP `403`.

---

## 3. Root Cause

The incident had four distinct layers.

### 3.1 Provider failure

Production was configured with:

```text
TRAFFIC_PROVIDER=airplanes.live
```

The provider returned HTTP `403 Unauthorized`, so a real ingestion cycle could
not complete.

### 3.2 Single-provider production selection

Because production selected `airplanes.live` directly, no secondary provider
participated in that workflow run.

### 3.3 Unauthorized was terminal for the fallback chain

The fallback implementation previously classified
`ErrProviderUnauthorized` as a terminal orchestration error. That was too broad
for a multi-provider system: authorization failure is terminal for that provider
attempt, but it should not necessarily terminate the whole provider chain.

### 3.4 External dispatch loop had no recent-failure breaker

The Cloudflare reliability Worker suppressed:

- an already active workflow run;
- a recent successful workflow run.

It did not suppress a recent failed workflow run. Once the provider started
returning a fast deterministic `403`, the scheduler could repeatedly create
another GitHub `workflow_dispatch`.

This turned one provider incident into repeated automation noise.

### 3.5 Provider access policy was confirmed externally

On 2026-08-20, Airplanes.live responded directly to the project's REST API access
request and confirmed that its general free API had been taken down.

The provider stated that:

- feeders may access the API from the same IP as the feeder;
- API users are encouraged to operate a feeder;
- application, website, and other higher-volume users are directed to sponsorship;
- the quoted sponsorship options at the time of the response were USD 25/month and
  USD 50/month for developer sponsorship.

No free external API access exception or project-specific credential was granted to
Global Flight Analytics.

This changes the interpretation of the production `403`: it is not treated as a
transient provider outage or an application authentication bug. It is an externally
confirmed access-policy boundary. Under the project's current zero-cost production
constraint, Airplanes.live is therefore not an available general external production
provider path unless a compatible feeder-origin access model is introduced.

---

## 4. Containment

Containment was intentionally fail-closed.

### GitHub Actions gate

`Production Traffic Ingestion` was manually disabled and active/queued executions
were cancelled where applicable.

The workflow remains disabled while production ingestion has no approved working
provider path.

### Cloudflare dispatch gate

The Worker now exposes:

```text
DISPATCH_ENABLED=false
```

When this flag is false, both configured Cron Trigger paths remain deployed but
return without dispatching GitHub Actions.

This preserves the infrastructure definition while preventing hidden recovery
traffic during an unresolved provider incident.

---

## 5. Permanent Resilience Changes

### 5.1 Recent-failure circuit breaker

The Worker now tracks the latest completed workflow run. A recent failed run
opens a bounded circuit breaker:

```text
RECENT_FAILURE_COOLDOWN_SECONDS=21600
```

That is a six-hour cooldown.

A newer successful run supersedes an older failure, so the system does not remain
artificially blocked after real recovery.

### 5.2 Provider-level unauthorized fallback

`ErrProviderUnauthorized` was moved out of the terminal orchestration
classification and into the unavailable-provider candidate classification.

New semantics:

```text
primary provider
  -> Unauthorized
  -> preserve unauthorized attempt evidence
  -> mark that provider unavailable
  -> try the next permitted provider
```

If a secondary provider succeeds, the decision is `fallback_selected`.

If all permitted providers are unavailable, the result is
`no_provider_available`.

The original unauthorized evidence remains present in `Decision.Attempts[]`.
The system therefore does not hide or rewrite the provider failure.

### 5.3 Provider terms remain part of the runtime contract

The existing OpenSky operational-agreement gate was preserved.

The incident response does not set an approval flag merely to restore service.
A technically reachable provider is not automatically an approved production
provider.

---

## 6. Verification Evidence

The Cloudflare reliability changes passed:

```text
Worker tests: 14/14 PASS
repository reliability tests: 20/20 PASS
CLOUDFLARE_RELIABILITY_SOURCE_CONTRACT=PASS
ZERO_COST_INGESTION_RELIABILITY_FOUNDATION=PASS
PRODUCTION_INGESTION_RELIABILITY=PASS
```

The backend fallback changes passed:

```text
focused TrafficFallbackProvider tests=PASS
go test -count=1 ./...=PASS
go vet ./...=PASS
git diff --check=PASS
```

The resilience patch was published as:

```text
557636f3818b8ecb241b2320503a32660ca05aa2
fix: harden production traffic ingestion resilience
```

The fail-closed Cloudflare Worker was deployed with:

```text
DISPATCH_ENABLED=false
RECENT_FAILURE_COOLDOWN_SECONDS=21600
Worker Version ID=2dd252b1-40eb-4858-8905-e2e4508bd0dc
```

Post-deploy verification reported:

```text
PRODUCTION_TRAFFIC_INGESTION_WORKFLOW=DISABLED
NEW_INGESTION_RUN_AFTER_GATE=NO
CLOUDFLARE_WORKER_DEPLOY=PASS
CLOUDFLARE_DISPATCH_KILL_SWITCH=ACTIVE
```

A later GitHub run with ID `32311392438` was manually started by the owner while
the workflow was being inspected. It reproduced the known
`airplanes.live -> 403 Unauthorized` failure and is not evidence that the
automatic Cloudflare dispatch loop resumed.

---

## 7. Why This Is Not a False Green

The remediation does **not**:

- convert a `403` into a successful ingestion;
- update freshness timestamps without a successful ingestion;
- suppress provider health evidence;
- set an unconfirmed provider agreement flag;
- fabricate traffic observations;
- re-enable production ingestion before an operational source is proven.

The production state remains explicitly degraded rather than falsely healthy.

---

## 8. Reactivation Criteria

Production ingestion may be re-enabled only after all of the following are true:

1. at least one production provider is explicitly permitted for the intended use;
2. its production configuration is complete without bypassing agreement/contact gates;
3. a real production ingestion smoke test succeeds;
4. provider/fallback telemetry shows the actual provider selected;
5. traffic freshness verification passes;
6. a subsequent normal scheduled run succeeds;
7. the Cloudflare dispatch kill switch is intentionally changed from
   `DISPATCH_ENABLED=false`;
8. the GitHub workflow is intentionally re-enabled;
9. alerting remains quiet for the expected reason rather than because monitoring
   was disabled.

For a multi-provider path, a primary-provider failure must remain visible even
when a secondary provider restores ingestion.

---

## 9. Current State

```text
PRODUCTION_INGESTION_RESILIENCE_INCIDENT=CONTAINED
GITHUB_WORKFLOW_GATE=DISABLED
CLOUDFLARE_DISPATCH_KILL_SWITCH=ACTIVE
RECENT_FAILURE_CIRCUIT_BREAKER=DEPLOYED
UNAUTHORIZED_PROVIDER_FALLBACK=HARDENED
AIRPLANES_LIVE_HTTP_STATUS=403
AIRPLANES_LIVE_PROVIDER_POLICY_CONFIRMED=YES
AIRPLANES_LIVE_FREE_EXTERNAL_API=UNAVAILABLE
AIRPLANES_LIVE_FEEDER_ACCESS=SAME_IP_ONLY
AIRPLANES_LIVE_PROJECT_EXCEPTION=NOT_GRANTED
PRODUCTION_PROVIDER_RECOVERY=OPEN
PRODUCTION_INGESTION=INTENTIONALLY_OFFLINE
```

The resilience/containment increment is closed. Provider recovery and controlled
production reactivation are a separate follow-up increment.


<!-- ADSBLOL-RECOVERY-SUCCESSOR-V1 -->
## Provider recovery successor

The follow-on recovery implementation uses ADSB.lol as the default open-data
traffic provider candidate. The adapter, provider policy, access gates and
source-ready production workflow are recorded in Document 193.

This does not close `PRODUCTION_PROVIDER_RECOVERY` by itself. Production remains
intentionally offline until ADSB.lol production-use contact is confirmed and the
real production smoke, freshness and subsequent scheduled-run checks pass.

```text
ADSBLOL_IMPLEMENTATION_FOUNDATION=READY
ADSBLOL_PRODUCTION_CONTACT=PENDING
PRODUCTION_PROVIDER_RECOVERY=OPEN
PRODUCTION_INGESTION=INTENTIONALLY_OFFLINE
```

---

## Canonical remediation record — GFA-OPS-452

### 1. Finding / symptom
A deterministic provider failure could be amplified into repeated automatic `workflow_dispatch -> failure -> dispatch` cycles and repeated failure notifications.

### 2. Root cause
The Cloudflare scheduler suppressed active runs and recent successful runs, but its dispatch decision did not treat a recent completed failure as a bounded reason to stop automatic recovery attempts.

### 3. Failure scenario
A provider begins returning a fast deterministic authorization failure. Each GitHub ingestion run fails quickly; the external scheduler sees no active or recent successful run and dispatches another known-failing execution, repeating until an operator intervenes.

### 4. Impact
The system generated repeated failing automation and notification noise without any chance of restoring freshness, consuming CI/provider capacity and obscuring the single underlying incident.

### 5. Severity rationale
**P1 retrospective.** This was a real production automation-amplification incident in which one deterministic external failure drove repeated executions until fail-closed containment. The severity is reconstructed from the observed operational impact; no original severity label is claimed.

### 6. Existing guarantees violated
- deterministic failures must not create an unbounded recovery loop;
- production recovery automation must be bounded and fail closed;
- external/provider failures must not create uncontrolled GitHub dispatch volume;
- a known failed state must remain visible rather than being retried indefinitely.

### 7. Considered solutions
- rely on manual workflow disablement only;
- increase the existing recent-success deduplication window;
- stop all scheduling permanently;
- add a recent-failure circuit breaker plus an explicit dispatch kill switch.

### 8. Chosen remediation
Add a six-hour recent-failure circuit breaker keyed from the latest completed workflow run, allow a newer success to supersede an older failure, and add fail-closed `DISPATCH_ENABLED=false` containment while provider recovery is unresolved.

### 9. Why this solution was selected
It bounds deterministic failure amplification without erasing the failed run, preserves automatic recovery after a later real success, and provides an explicit owner-controlled production stop switch.

### 10. Rejected alternatives
Manual-only containment was insufficient as a permanent guard; enlarging success deduplication does not reason about failure; permanently removing scheduling would discard the validated reliability architecture.

### 11. Trade-offs
A long cooldown can delay automatic retry after an upstream issue is fixed, so controlled recovery requires an intentional owner action or a newer successful run. The kill switch also keeps production intentionally offline until recovery evidence exists.

### 12. Regression tests / protection
Worker and repository reliability tests cover recent-failure classification, cooldown bounds, newer-success supersession, kill-switch behavior and the requirement that disabled Cron handlers perform no dispatch work.

### 13. Adversarial review findings
The fix deliberately does not convert provider `403` into success, advance freshness, suppress provider health evidence, or bypass provider-access gates merely to stop alert noise.

### 14. Remediation iterations
1. Production incident exposed repeated failed dispatches.
2. The GitHub ingestion workflow was disabled and active/queued work contained.
3. The Worker gained the explicit kill switch and recent-failure circuit breaker.
4. Deployment evidence verified the kill switch active while provider recovery remained open.

### 15. Residual risks and limitations
A circuit breaker limits automation amplification but cannot restore service when every approved provider is unavailable. Cloudflare and GitHub remain external scheduling/execution dependencies.

### 16. Operational or deployment consequences
Production dispatch stays disabled while provider recovery is unresolved. Reactivation must explicitly enable both the GitHub workflow and Cloudflare dispatch only after provider smoke/freshness evidence passes.

### 17. Exact evidence
- remediation PR #79 head `4f4aab82e609baece3346240620b0bbf195fb7f5`, merge `f97e367667686e58429f26656f998537931236d7`;
- patch commit `557636f3818b8ecb241b2320503a32660ca05aa2`;
- Backend CI `32362621147` SUCCESS;
- Frontend CI `32362621102` SUCCESS;
- CodeQL `32362621146` SUCCESS;
- API Load Baseline `32362621013` SUCCESS;
- OpenAPI Contract `32362621127` SUCCESS;
- Playwright E2E `32362621113` SUCCESS;
- deployed `RECENT_FAILURE_COOLDOWN_SECONDS=21600`, `DISPATCH_ENABLED=false`, Worker version `2dd252b1-40eb-4858-8905-e2e4508bd0dc`;
- post-deploy markers `PRODUCTION_TRAFFIC_INGESTION_WORKFLOW=DISABLED`, `NEW_INGESTION_RUN_AFTER_GATE=NO`, `CLOUDFLARE_DISPATCH_KILL_SWITCH=ACTIVE`.

### 18. Final canonical status
**CLOSED.** The automation-amplification defect is contained and regression-protected; provider recovery is a separate open runtime state.

### 19. Prevention / future guard
Keep recent-failure suppression and the explicit kill switch in the production scheduler contract. Do not use empty retrigger or retry loops to recover deterministic provider failures; require a bounded recovery decision with preserved failure evidence.

---

## Canonical remediation record — GFA-REL-453

### 1. Finding / symptom
A provider-level `ErrProviderUnauthorized` terminated the complete traffic fallback orchestration instead of allowing another permitted provider candidate to be attempted.

### 2. Root cause
Unauthorized was classified as a terminal orchestration error rather than an unavailable result for the current provider attempt. Provider-scoped authentication failure and chain-wide terminal failure were conflated.

### 3. Failure scenario
The primary provider returns Unauthorized while a secondary provider is technically and policy-eligible. The fallback chain stops at the first attempt, records terminal failure and never calls the eligible secondary provider.

### 4. Impact
One provider's access failure could unnecessarily remove the entire multi-provider availability path and prevent service restoration through a permitted fallback, while production freshness continued degrading.

### 5. Severity rationale
**P1 retrospective.** The defect allowed a provider-scoped production availability failure to disable the whole fallback chain. It did not corrupt observations, but it materially weakened the system's intended provider-resilience boundary. The severity is retrospective, not a claimed historical label.

### 6. Existing guarantees violated
- provider attempt failures and orchestration terminal failures must remain distinct;
- fallback must try the next permitted candidate after a provider becomes unavailable;
- original provider failure evidence must be preserved through fallback;
- policy/access gates must still decide whether a secondary candidate is permitted.

### 7. Considered solutions
- keep Unauthorized terminal for the whole chain;
- convert Unauthorized into a generic server failure;
- hide the failed primary when fallback succeeds;
- classify Unauthorized as provider-unavailable while retaining explicit unauthorized attempt evidence.

### 8. Chosen remediation
Move `ErrProviderUnauthorized` from terminal error classification into unavailable-provider candidate classification; preserve `AttemptErrorClassUnauthorized`; continue to the next policy-permitted candidate; return `fallback_selected` on success or `no_provider_available` if all permitted candidates fail.

### 9. Why this solution was selected
It restores provider-level fault isolation without weakening authorization semantics or hiding the original `403`. The chain can recover only through candidates already allowed by provider policy.

### 10. Rejected alternatives
Keeping Unauthorized chain-terminal defeated multi-provider resilience; relabeling it as generic server failure lost semantics; suppressing primary failure evidence would create a false green; bypassing OpenSky/other access gates was prohibited.

### 11. Trade-offs
Fallback may create an additional permitted provider request after an unauthorized primary attempt. This is bounded by provider policy, budgets and health controls, and it still cannot recover if no secondary provider is approved or available.

### 12. Regression tests / protection
TrafficFallbackProvider tests cover primary Unauthorized followed by secondary success, mixed unavailable failures, `NoProviderAvailable`, ordered attempt evidence, `UsedFallback`, selected provider and preserved unauthorized error class.

### 13. Adversarial review findings
The remediation preserves the original Unauthorized attempt in `Decision.Attempts[]` and retains the OpenSky operational-agreement gate. It does not make an unapproved provider reachable merely because the primary failed.

### 14. Remediation iterations
1. Incident investigation separated the external Airplanes.live access failure from internal fallback behavior.
2. Unauthorized was removed from terminal classification.
3. Fallback tests were rewritten to require secondary execution and full ordered attempt evidence.
4. Production remained intentionally offline because a code-level fallback fix alone does not prove an approved working provider.

### 15. Residual risks and limitations
A correct fallback chain cannot recover when every permitted provider is unavailable or policy-blocked. Provider terms, quotas and authentication can change independently of repository code.

### 16. Operational or deployment consequences
No provider gate is relaxed. Production reactivation still requires an explicitly permitted provider, controlled smoke, freshness verification and a subsequent scheduled run.

### 17. Exact evidence
- PR #79 head `4f4aab82e609baece3346240620b0bbf195fb7f5`, merge `f97e367667686e58429f26656f998537931236d7`;
- patch commit `557636f3818b8ecb241b2320503a32660ca05aa2`;
- Backend CI `32362621147`, Frontend CI `32362621102`, CodeQL `32362621146`, API Load Baseline `32362621013`, OpenAPI Contract `32362621127`, Playwright E2E `32362621113` — all SUCCESS;
- PR #81 head `060347b8f7cf7c94008f0f618a5f8304902ce27e`, merge `40f6eeee1f3a40dc7d11409aa1c656f316004a45` later records Airplanes.live's confirmed access-policy boundary rather than reclassifying the external `403` as an application defect.

### 18. Final canonical status
**CLOSED.** Provider-level Unauthorized fallback semantics are corrected; `PRODUCTION_PROVIDER_RECOVERY` remains independently open.

### 19. Prevention / future guard
Maintain explicit provider-attempt error classes, ordered fallback evidence and policy-gated candidate selection. New provider-specific authorization errors must be reviewed for provider scope versus chain-wide terminal scope before being added to terminal orchestration classification.
