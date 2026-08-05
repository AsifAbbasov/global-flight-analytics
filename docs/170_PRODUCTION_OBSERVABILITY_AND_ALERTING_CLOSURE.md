# Production Observability and Alerting Closure

Status: Closed  
Closure date: 2026-08-05  
Project: Global Flight Analytics

---

## Purpose

This document records the completed production observability and alert-delivery
boundary for Global Flight Analytics. It separates repository-owned configuration,
cloud provisioning, production metric evidence, notification delivery, security
controls, and remaining operational limitations.

This is an immutable engineering evidence record. It does not claim that mutable
production aliases will always serve the same application revision.

---

## Closed Scope

The completed observability path is:

```text
Render production API
  -> protected Prometheus metrics endpoint
  -> GitHub Actions production metrics workflow
  -> Grafana Alloy validation and remote write
  -> Grafana Cloud Prometheus
  -> repository-provisioned SLO dashboard
  -> nine Grafana-managed alert rules
  -> Grafana notification policy
  -> owner-controlled email contact point
```

The provisioning implementation was merged through pull request `#49` at source
revision:

```text
3b2a101256d85e0f178492db8bee3e5cb3cfbb63
```

The production metric evidence used during closure was tied to the explicitly
verified application revision:

```text
5735d4eb530a8e670d676d6e654a02d6d7dd71bf
```

These two revisions represent different evidence boundaries: the first identifies
the repository provisioning implementation, while the second identifies the
application revision represented by the verified production metric series.

---

## Production Metric Evidence

GitHub Actions run `30957340540` successfully completed the protected production
metric scrape and Grafana Cloud remote-write path.

Verified markers:

```text
PRODUCTION_METRICS_INPUT=PASS
PRODUCTION_METRICS_SOURCE_PREFLIGHT=PASS
GRAFANA_ALLOY_CONFIG=PASS
GRAFANA_CLOUD_REMOTE_WRITE=PASS
GRAFANA_CLOUD_QUERY_EVIDENCE=PASS
```

The production API exposes Prometheus-format telemetry only through the configured
metrics protection boundary. The raw metrics credential is stored only as a GitHub
Actions secret. Render stores only the SHA-256 digest used by the API for credential
verification.

No raw metric credential, Grafana token, service-account token, database credential,
or notification address is recorded in this document.

---

## Grafana Cloud Provisioning

The Grafana Cloud stack uses the stack-scoped Kubernetes-style API namespace:

```text
Stack slug: maroontanager1621
Stack ID: 1749941
API namespace: stacks-1749941
Prometheus datasource UID: grafanacloud-prom
```

The first complete stack-scoped provisioning run was GitHub Actions run
`30963255727`. The final receiver-validated provisioning evidence is GitHub Actions
run `30963876652`.

The final run verified:

```text
GRAFANA_PROVISION_INPUT=PASS
GRAFANA_OBSERVABILITY_CONTRACT=PASS
GRAFANA_OBSERVABILITY_RENDER=PASS
GRAFANA_NAMESPACE=stacks-1749941
GRAFANA_FOLDER=PASS
GRAFANA_SLO_DASHBOARD=PASS
GRAFANA_ALERT_RULES=PASS
GRAFANA_NOTIFICATION_POLICY_RECEIVER=global-flight-analytics-production-email
GRAFANA_NOTIFICATION_POLICY=PASS
GRAFANA_OBSERVABILITY_PROVISION=PASS
GRAFANA_OBSERVABILITY_WORKFLOW=PASS
```

Provisioning is repository-owned and idempotent. Re-running the workflow updates the
same stable folder, dashboard, and alert group rather than creating duplicate
resources. The workflow validates the configured receiver but does not overwrite the
owner-managed notification policy tree.

---

## Provisioned Resources

### Folder and dashboard

```text
Folder UID: gfa-observability
Dashboard resource name: gfa-production-slo
Dashboard title: Global Flight Analytics — Production SLO
```

The dashboard contains production panels for request availability, latency, server
errors, ingestion freshness and failure state, PostgreSQL pool pressure,
reconciliation backlog, collector health, and metric continuity. It also exposes an
explicit deployment-revision variable so evidence can be filtered by application
revision.

### Alert rules

The managed alert group contains exactly nine rules:

```text
gfa-api-availability
gfa-api-p95-latency
gfa-api-server-errors
gfa-ingestion-freshness
gfa-ingestion-failures
gfa-postgres-pool
gfa-reconciliation-backlog
gfa-collector-health
gfa-metrics-missing
```

Rules use Grafana-managed Math conditions with bounded production labels. Ordinary
rules treat scheduled scrape gaps conservatively to avoid false pages, while the
explicit metric-continuity rule owns the missing-metrics condition.

---

## Notification Delivery Evidence

The default Grafana notification policy targets the contact point:

```text
global-flight-analytics-production-email
```

A controlled Grafana test alert was sent successfully and delivered to the
owner-controlled mailbox on 2026-08-05. The received message identified one firing
`TestAlert` instance and the `Notification test` summary.

The mailbox address is intentionally excluded from repository documentation. The
successful test proves the configured Grafana-to-email delivery path; it is not
presented as evidence of a real production incident.

---

## Security and Ownership Boundaries

- GitHub repository variables contain non-secret stack metadata and expected resource identifiers.
- GitHub Actions secrets contain the raw production metrics credential and Grafana tokens.
- Render stores only the production metrics credential digest.
- Grafana service-account access is scoped to repository-driven provisioning.
- Notification addresses remain owner-controlled Grafana configuration and are not committed.
- Provisioning reads and validates the notification policy but does not replace the policy tree.
- Revoked or superseded tokens are not valid evidence and are not recorded.

---

## Remaining Operational Limitations

- The current collection path is GitHub Actions-driven rather than a continuously running dedicated telemetry agent.
- Email is the only verified notification channel; there is no secondary escalation channel.
- The project does not provide a staffed 24/7 on-call rotation or incident-response service-level agreement.
- Alert thresholds are engineering portfolio controls, not regulated aviation safety thresholds.
- Grafana Cloud account plan continuity, billing, and future quota changes remain owner responsibilities.
- Future production application revisions must repeat revision-specific smoke and metric evidence.

---

## Closure Statement

The production observability increment is closed because protected production metrics
reach Grafana Cloud, the stack-scoped provisioning workflow is green and idempotent,
the SLO dashboard and nine alert rules exist, the notification policy resolves to the
expected receiver, and controlled email delivery has been demonstrated.

```text
PRODUCTION_METRICS_REMOTE_WRITE=CLOSED
GRAFANA_STACK_NAMESPACE=CLOSED
GRAFANA_SLO_DASHBOARD=CLOSED
GRAFANA_ALERT_RULES=CLOSED
GRAFANA_NOTIFICATION_POLICY=CLOSED
ALERT_NOTIFICATION_DELIVERY=CLOSED
PRODUCTION_OBSERVABILITY_CLOSURE=PASS
```
