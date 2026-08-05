# DOCUMENT 19

# RISK ANALYSIS

# Global Flight Analytics

Version: 1.2

Status: Approved

---

# 1. Purpose

This document describes the project's primary risks and their mitigations.

---

# 2. Project Philosophy

Primary principle:

Do not start managing risks only after launch.

---

Primary principle:

Build risk protection into the architecture.

---

# 3. Risk Classification

Risks are divided into:

- technical;
- infrastructure;
- product;
- legal;
- operational;
- strategic.

---

# 4. Technical Risks

## Risk 1

OpenSky is unavailable.

---

Consequences:

- no new data;
- map updates stop.

---

Mitigation:

- cache the latest data;
- notify the user;
- reconnect automatically.

---

## Risk 2

The OpenSky API structure changes.

---

Consequences:

- data import fails.

---

Mitigation:

- adapter layer;
- integration isolation;
- contract tests.

---

## Risk 3

Route-determination errors.

---

Consequences:

- false routes;
- reduced trust in the system.

---

Mitigation:

- Confidence Level;
- present assumptions as assumptions;
- avoid categorical claims.

---

## Risk 4

Historical data volume grows.

---

Consequences:

- increased database size;
- reduced performance.

---

Mitigation:

- data aggregation;
- retention policy;
- limited historical detail.

---

# 5. Infrastructure Risks

## Risk 5

The free server enters sleep mode.

---

Consequences:

- slow first response.

---

Mitigation:

- inform the user;
- use caching;
- remain ready to move to a paid plan.

---

## Risk 6

The database limit is exceeded.

---

Consequences:

- writes become impossible.

---

Mitigation:

- aggregation;
- archiving;
- data-growth control.

---

## Risk 7

Insufficient memory.

---

Consequences:

- process restart;
- loss of in-memory state.

---

Mitigation:

- regional limits;
- object-count control;
- memory-usage monitoring.

---

# 6. Product Risks

## Risk 8

Attempting to compete directly with Flightradar24.

---

Consequences:

- loss of product uniqueness;
- failure against larger competitors.

---

Mitigation:

Focus on analytics, routes, airports, and air-traffic research.

---

## Risk 9

Scope expansion.

---

Consequences:

- endless development;
- no release.

---

Mitigation:

Strictly follow the MVP Scope.

---

## Risk 10

No user value.

---

Consequences:

- no audience;
- no product growth.

---

Mitigation:

- launch the MVP quickly;
- collect feedback;
- analyze user behavior.

---

# 7. Legal Risks

## Risk 11

Using data outside its license terms.

---

Consequences:

- legal claims;
- blocked data access.

---

Mitigation:

Use only open data sources.

---

## Risk 12

Users misinterpret analytics.

---

Consequences:

A user treats a forecast or analytical estimate as a confirmed fact.

---

Mitigation:

- use Confidence Level;
- label estimates;
- display system limitations.

---

# 8. Legal And Compliance Risks

## Risk 13

Violating data-provider licenses.

---

Consequences:

- restricted data access;
- blocked API access;
- legal claims;
- required data deletion.

---

Mitigation:

- document every source license;
- review terms of use regularly;
- identify data sources on the website.

---

## Risk 14

Violating OpenSky terms of use.

---

Consequences:

- access restrictions;
- temporary blocking;
- permanent account blocking.

---

Mitigation:

- comply with request limits;
- use the Backend API;
- cache data;
- control workload.

---

## Risk 15

Using data in ways prohibited by providers.

---

Consequences:

- license violations;
- loss of access.

---

Mitigation:

Review licensing terms before integrating a new source.

---

## Risk 16

No user agreement.

---

Consequences:

- legal uncertainty;
- no limitation of platform liability;
- increased legal risk.

---

Mitigation:

Prepare the following before public launch:

- Terms Of Service;
- Privacy Policy;
- Disclaimer.

---

## Risk 17

Using analytics as official aviation information.

---

Consequences:

A user makes decisions based on estimated data.

---

Mitigation:

Display the following notice on all platform pages:

The platform is a research tool and does not provide official aviation information.

---

## Risk 18

Personal user data appears in future versions.

---

Consequences:

Data-protection requirements become applicable.

---

Mitigation:

Do not store personal data before user accounts are introduced.

Perform a separate legal requirements audit after user accounts are introduced.

---

## Risk 19

User claims caused by inaccurate data.

---

Consequences:

Reduced trust in the platform and legal claims.

---

Mitigation:

Label all computed data as:

- Predicted;
- Estimated;
- Confidence Based;
- Unofficial.

---

Computed data must never be presented as confirmed fact.

---

## Risk 20

Using the platform outside its intended purpose.

---

Examples:

- aviation navigation;
- dispatch operations;
- flight-safety operations.

---

Consequences:

A user makes safety-critical decisions based on platform analytics.

---

Mitigation:

All user-facing documents must include the following statement:

The platform is intended exclusively for research, visualization, and analysis of open aviation data.

---

# 9. Operational Risks

## Risk 21

Workload grows faster than infrastructure.

---

Mitigation:

Scale only after real workload appears.

---

## Risk 22

Data loss after a failure.

---

Mitigation:

- backups;
- persist critical data in PostgreSQL.

---

## Risk 23

An external data provider fails.

---

Consequences:

Partial loss of functionality.

---

Mitigation:

Prepare alternative data providers in future versions.

---

# 10. Strategic Risks

## Risk 24

Loss of project focus.

---

Consequences:

The product becomes a set of unrelated capabilities.

---

Mitigation:

Maintain one product strategy.

---

## Risk 25

Premature machine-learning adoption.

---

Consequences:

High complexity without real value.

---

Mitigation:

Introduce machine learning only after sufficient historical data is accumulated.

---

## Risk 26

Premature architectural complexity.

---

Consequences:

Higher development and maintenance cost.

---

Examples:

- Kubernetes;
- microservices;
- Redis;
- ClickHouse.

---

Mitigation:

Follow this principle:

```text
Build For Today

Prepare For Tomorrow
```

---

# 11. Main Strategic Risk

The most dangerous project risk:

Losing product focus.

---

The platform must remain:

- an air-traffic research system;
- an aviation analytics system;
- an airspace forecasting system.

---

The platform must not become:

- a travel portal;
- a ticket-sales system;
- a dispatch system;
- an aviation ERP platform.

---

# 12. Risk Summary

Current MVP risk level:

```text
Low To Medium
```

---

Reasons:

- simple architecture;
- limited scope;
- no complex infrastructure;
- incremental product development.

---

The main success factors remain product focus, license compliance, legal-risk protection, and disciplined platform development.
