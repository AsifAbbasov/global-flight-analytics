# Grafana Cloud production metrics transport

This directory contains the Grafana Alloy configuration used by the scheduled external
production metrics scraper.

The scraper runs outside Render on a GitHub-hosted runner, authenticates to the protected
`/internal/metrics` endpoint, and forwards the Prometheus-compatible samples to Grafana Cloud
through `prometheus.remote_write`.

The configuration intentionally carries only bounded application labels plus three external
labels:

- `environment="production"`;
- `service="global-flight-analytics-api"`;
- `deployment_revision=<exact Render SHA>`.

Secrets are read only from workflow environment variables. Never write credentials into this
file or commit generated Alloy storage.

## SLO dashboard and alerts

`folder.json`, `dashboard.json`, and `alert-rules.json` define the production observability
resources. `scripts/provision-grafana-observability.sh` renders the Prometheus datasource UID,
upserts the Grafana folder and dashboard through the current `/apis` resources, and idempotently
replaces the bounded SLO alert group through Grafana's provisioning endpoint.

The alert group uses the existing default notification policy instead of overwriting the stack's
policy tree. Provisioning verifies that a receiver exists. Delivery must still be proven by a
Grafana contact-point test and one controlled test alert; repository configuration alone is not
delivery evidence.
