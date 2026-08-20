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
AIRPLANES_LIVE=UNAVAILABLE_403
PRODUCTION_PROVIDER_RECOVERY=OPEN
PRODUCTION_INGESTION=INTENTIONALLY_OFFLINE
```

The resilience/containment increment is closed. Provider recovery and controlled
production reactivation are a separate follow-up increment.
