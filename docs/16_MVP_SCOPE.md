# DOCUMENT 16

# MVP SCOPE

# Global Flight Analytics

Version: 1.1

Status: Approved

---

# 1. Purpose

This document defines the boundaries of the first Global Flight Analytics release.

The document establishes:

- mandatory MVP functionality;
- MVP constraints;
- MVP success criteria;
- capabilities excluded from the first release.

---

# 2. Main Goal

Create an operational air-traffic visualization platform based on open aviation data.

---

# 3. MVP Success Criteria

A user must be able to:

- open the map;
- see real aircraft;
- open an aircraft card;
- open an airport card;
- perform a search;
- view basic analytics;
- view an estimated aircraft route.

---

The MVP is successful when all listed scenarios operate reliably.

---

# 4. Core User Scenarios

## Scenario 1

Explore airspace.

---

The user:

- opens the map;
- views aircraft;
- explores regional activity.

---

## Scenario 2

Explore an aircraft.

---

The user:

- selects an aircraft;
- receives aircraft-model information;
- receives airline information;
- receives route information.

---

## Scenario 3

Explore an airport.

---

The user:

- opens an airport;
- explores infrastructure;
- explores statistics;
- explores routes.

---

# 5. Included Features

## Live Aircraft Map

Real-time aircraft display.

---

## Aircraft Profile

Aircraft card.

Contains:

- model;
- airline;
- registration;
- flight parameters.

---

## Airport Profile

Airport card.

Contains:

- description;
- runways;
- infrastructure;
- statistics.

---

## Route Detection

Basic route determination.

---

## Traffic Analytics

Basic regional analytics.

---

## Search

Search for:

- airports;
- aircraft;
- airlines.

---

# 6. Nice To Have Features

The following may be implemented in the MVP when time permits:

- heat maps;
- airport ranking;
- additional charts;
- advanced filters.

---

The absence of these capabilities does not affect MVP readiness.

---

# 7. Excluded Features

The MVP does not include:

- a mobile application;
- notifications;
- user accounts;
- authorization;
- favorites;
- user settings;
- subscriptions;
- payments;
- traffic forecasting;
- artificial intelligence;
- machine learning;
- anomaly detection.

---

# 8. Geographic Scope

The first release covers:

- Azerbaijan;
- the Caucasus;
- Turkey.

---

After a successful launch, coverage may expand to Europe.

---

# 9. Data Scope

The platform uses:

- OpenSky Network;
- OurAirports;
- OpenStreetMap;
- Wikidata.

---

Other data sources are not mandatory for the MVP.

---

# 10. Infrastructure Scope

The MVP uses only:

- Vercel;
- Render;
- Neon PostgreSQL.

---

Without:

- Kubernetes;
- Redis;
- Kafka;
- RabbitMQ.

---

# 11. MVP Constraints

The first release must remain:

- simple;
- inexpensive;
- maintainable by one developer.

---

Any capability that complicates the architecture without proven value is moved outside the MVP.

---

# 12. MVP Success Metrics

The MVP is successful when:

- data is displayed correctly;
- the map operates reliably;
- search operates correctly;
- routes are determined correctly;
- analytics are generated without errors.

---

# 13. MVP Summary

The first release must prove that:

The platform can collect, connect, analyze, and visualize aviation data from open sources.

After product viability is confirmed, forecasting, machine learning, and intelligent analytics may be developed.
