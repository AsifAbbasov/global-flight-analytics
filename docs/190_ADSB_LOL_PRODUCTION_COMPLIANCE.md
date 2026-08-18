# ADSB.lol Production Compliance

Status: implemented on `feat/adsblol-production-compliance` after maintainer guidance received on 2026-08-18.

## Provider guidance received

The ADSB.lol maintainer confirmed that Global Flight Analytics may use the public API and asked production consumers to:

- display visible ODbL attribution;
- send an identifiable User-Agent with responsible-party contact information;
- respect HTTP errors and rate limiting, including `Retry-After` when present;
- cache and deduplicate where practical;
- keep geographic queries appropriately scoped;
- avoid rate-limit evasion through multiple IP addresses;
- design for future API authentication and endpoint changes;
- avoid treating ADSB.lol as a guaranteed or safety-critical service.

## Compliance mapping

### Visible attribution

The web application renders an application-wide footer crediting ADSB.lol contributors and linking to ODbL 1.0.

### Identifiable User-Agent

Production live-traffic requests use:

`GlobalFlightAnalytics/1.0 (+https://github.com/AsifAbbasov/global-flight-analytics; contact: aassifabbasov@gmail.com)`

### Rate limiting and HTTP error behavior

The existing provider orchestration is retained. Global Flight Analytics applies a conservative application-side ADSB.lol budget and does not retry failed requests in a tight loop. The live collector increases delay after failures. Standard `Retry-After` values are parsed by the provider-response controller and converted into provider cooldown state.

### Geographic scope

ADSB.lol live requests use bounded point/radius queries. Radius validation is fail-closed and capped at 250 nautical miles.

### Cache and deduplication boundary

Browsers do not call ADSB.lol directly. A central backend collector obtains a regional snapshot and writes canonical flight states into the shared current-state store. Subsequent application consumers read from that shared state rather than multiplying upstream requests per browser session. The current-state store applies canonical upsert semantics, preventing duplicate state records from accumulating as independent live objects.

### Availability and graceful degradation

ADSB.lol remains an external, volunteer-operated data provider and is not treated as safety-critical infrastructure. Provider health, budget decisions, failures and cooldowns remain observable through the existing orchestration and observability layers. Collector failures back off rather than increasing request pressure.

### Future API keys and API changes

The provider is isolated behind the `internal/integrations/adsblol` adapter and configuration-owned base URL. Future authentication can therefore be added at the adapter/configuration boundary without changing the canonical flight-state model or browser clients.

## Production activation

The previous external contact dependency is now resolved. `ADSB_LOL_PRODUCTION_CONTACT_CONFIRMED=true` is documented in the API environment example, while `LIVE_TRAFFIC_COLLECTOR_ENABLED=false` remains the safe default so production activation is still an explicit deployment decision.

## Non-goals

This compliance increment does not claim an ADSB.lol service-level agreement, fixed provider quota, guaranteed accuracy, guaranteed availability, or exclusive API access.
