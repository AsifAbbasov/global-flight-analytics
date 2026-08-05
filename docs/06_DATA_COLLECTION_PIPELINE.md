# DOCUMENT 06

# DATA COLLECTION PIPELINE

# Global Flight Analytics

Version: 1.1

Status: Approved

---

# 1. Purpose

This document describes how Global Flight Analytics retrieves, processes, enriches, stores, and serves aviation data.

The document defines:

- data flow;
- processing stages;
- storage rules;
- data-enrichment mechanisms;
- analytics-generation mechanisms;
- failure handling;
- scalability requirements.

---

# 2. Pipeline Overview

High-level data flow:

```text
OpenSky Network

        ↓

Validation Layer

        ↓

Flight Matching Layer

        ↓

Data Enrichment Layer

        ↓

In-Memory Store

        ↓

Snapshot Generator

        ↓

PostgreSQL

        ↓

Backend API

        ↓

Next.js
```

---

# 3. Collection Principles

## Principle 1

Store only useful data.

---

## Principle 2

Do not turn PostgreSQL into an unbounded telemetry store.

---

## Principle 3

Historical analytics must be based on aggregates.

---

## Principle 4

Live data must be served primarily from memory.

---

## Principle 5

Every record must preserve data provenance.

---

# 4. Live Collection Layer

Source:

OpenSky Network

---

The Backend sends requests to OpenSky.

Refresh interval:

- minimum: 15 seconds;
- maximum: 30 seconds.

---

Retrieved fields:

- ICAO24;
- Callsign;
- Latitude;
- Longitude;
- Velocity;
- Heading;
- Barometric Altitude;
- Geometric Altitude;
- Vertical Rate;
- On Ground Status;
- Observation Timestamp.

---

# 5. Validation Layer

After retrieval, data passes through mandatory validation.

---

Validation checks:

- valid coordinates;
- valid altitude;
- valid velocity;
- valid heading;
- ICAO24 presence;
- valid timestamp.

---

Invalid records:

- are rejected;
- are recorded in logs.

---

# 6. Flight Matching Layer

Goal:

Connect an observation to the system domain model.

---

Flow:

```text
ICAO24

↓

Aircraft

↓

Flight

↓

Flight State
```

---

Result:

Every observation is connected to:

- aircraft;
- flight;
- airline;
- aircraft_model.

---

# 7. Data Enrichment Layer

After matching, the data is enriched.

---

The layer uses:

- aircraft;
- aircraft_models;
- airlines;
- airports.

---

Result:

An enriched aircraft view is produced.

---

Flow:

```text
Flight State

↓

Aircraft

↓

Aircraft Model

↓

Airline

↓

Enriched Aircraft View
```

---

# 8. In-Memory Store

Application memory is used to serve the live map.

---

Implementation:

```text
sync.Map
```

---

Reasons:

- high read performance;
- no external dependency;
- no Redis in the MVP;
- compliance with the free-infrastructure principle.

---

Purpose:

- render the live map;
- search aircraft quickly;
- construct API responses.

---

# 9. Flight State Storage

Flight states are stored in:

```text
flight_states
```

---

Stored values:

- coordinates;
- velocity;
- altitude;
- heading;
- observation timestamp.

---

Retention Policy:

Minimum:

```text
24 hours
```

---

MVP target:

```text
7 days
```

---

Expired data is deleted.

---

# 10. Snapshot Generator

Every few minutes, the system creates an aggregated regional snapshot.

---

Source:

In-Memory Store

---

Result:

```text
traffic_snapshots
```

---

Contents:

- aircraft count;
- route count;
- active airports;
- traffic intensity;
- regional distribution.

---

# 11. Airport Statistics Pipeline

Flow:

```text
Flight

↓

Airport Detection

↓

Airport Statistics Update
```

---

Updates:

```text
airport_statistics
```

---

Metrics:

- arrivals;
- departures;
- total_flights.

---

# 12. Route Statistics Pipeline

Flow:

```text
Flight

↓

Route Prediction

↓

Route Statistics Update
```

---

Updates:

```text
route_statistics
```

---

Metrics:

- flight_count;
- route_activity.

---

# 13. Historical Replay Pipeline

Replay mode uses:

- flight_states;
- traffic_snapshots.

---

Goal:

Visualize recent movement history.

---

The MVP does not store months of raw telemetry.

---

History is bounded by the retention policy.

---

# 14. Ingestion Runs

Every data-ingestion execution creates a record in:

```text
ingestion_runs
```

---

Recorded values:

- source_name;
- started_at;
- finished_at;
- status;
- records_received;
- records_inserted;
- records_updated;
- error_message.

---

Purpose:

- monitoring;
- diagnostics;
- data-ingestion audit.

---

# 15. Failure Handling

When OpenSky is unavailable:

the latest available in-memory data is used.

---

The Frontend displays:

```text
Live Data Delayed
```

---

The system must not terminate unexpectedly.

---

# 16. Data Quality Rules

The system must:

- filter invalid coordinates;
- filter invalid altitudes;
- filter invalid velocities;
- prevent duplicate observations;
- record ingestion errors.

---

# 17. MVP Capacity Targets

MVP targets:

- up to 10,000 simultaneously observed aircraft;
- up to 100 API requests per second;
- one regional snapshot every 5 minutes.

---

Without:

- Redis;
- Kafka;
- RabbitMQ;
- Kubernetes;
- microservices.

---

# 18. Scalability Strategy

During the MVP:

```text
Go API

↓

sync.Map

↓

PostgreSQL
```

---

Architecture is made more complex only after proven workload appears.

---

# 19. Summary

The pipeline follows this principle:

```text
Retrieve data

↓

Validate data

↓

Connect data to the domain model

↓

Enrich data

↓

Store useful information

↓

Generate aggregates

↓

Present data to the user
```

This approach allows the MVP to operate on free infrastructure without Redis, message queues, or a microservice architecture.
