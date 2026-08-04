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
