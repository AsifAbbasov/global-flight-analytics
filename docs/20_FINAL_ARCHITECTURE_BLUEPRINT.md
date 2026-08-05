# DOCUMENT 20

# FINAL ARCHITECTURE BLUEPRINT

# Global Flight Analytics

Version: 1.1

Status: FINAL

---

# 1. Mission

Create a platform for air-traffic research, analysis, and visualization based on open aviation data.

---

# 2. Product Vision

The platform combines aviation data, airport data, and analytical calculations into one airspace research system.

---

# 3. What The Product Is

The platform displays:

- real aircraft;
- real airports;
- real routes;
- real statistics;
- air-traffic changes;
- analytical views of airspace.

---

# 4. What The Product Is Not

The platform is not:

- an air-traffic control system;
- a dispatch system;
- a military system;
- a ticket-sales system;
- a navigation system;
- an aviation-safety system;
- an internal airport system.

---

# 5. Technology Stack

## Frontend

- Next.js
- TypeScript
- MapLibre
- TanStack Query
- Tailwind CSS

---

## Backend

- Go
- Fiber
- pgx

---

## Database

- PostgreSQL
- Neon

---

## Infrastructure

- GitHub
- Vercel
- Render

---

# 6. High Level Architecture

```text
OpenSky
OurAirports
OpenStreetMap
Wikidata
        │
        ▼

Data Collection Pipeline

        │
        ▼

Go Backend

        │
        ├── Aircraft Enrichment
        │
        ├── Airport Intelligence
        │
        ├── Route Detection Engine
        │
        ├── Traffic Analytics
        │
        └── Historical Replay

        │
        ▼

PostgreSQL

        │
        ▼

REST API

        │
        ▼

Next.js Frontend
```

---

# 7. Core Modules

## Aircraft Module

Responsible for:

- aircraft;
- aircraft models;
- airlines.

---

## Airport Intelligence Module

Responsible for:

- airport digital passport;
- infrastructure;
- statistics;
- transport connections.

---

## Route Detection Engine

Responsible for:

- route determination;
- Confidence Level calculation;
- aircraft-movement analysis.

---

## Traffic Analytics Module

Responsible for:

- statistics;
- heat maps;
- analytical panels.

---

## Historical Replay Module

Responsible for:

- historical snapshots;
- air-traffic replay.

---

# 8. Core Data Sources

## OpenSky

Primary air-traffic data source.

---

## OurAirports

Primary airport data source.

---

## OpenStreetMap

Infrastructure data source.

---

## Wikidata

Reference information source.

---

# 9. Architectural Principles

## Principle 1

Keep It Simple.

---

## Principle 2

Build For Today.

Prepare For Tomorrow.

---

## Principle 3

No Premature Optimization.

---

## Principle 4

Open Data First.

---

## Principle 5

Backend Owns The Data.

---

## Principle 6

One Source Of Truth.

---

# 10. Architectural Boundaries

The MVP prohibits:

- microservices;
- Kubernetes;
- Redis;
- ClickHouse;
- a separate Python Backend;
- paid infrastructure.

---

Architectural complexity is allowed only after real workload appears.

---

# 11. Security Boundaries

The Frontend never accesses data sources directly.

---

All requests pass through the Backend API.

---

All external data is treated as untrusted until validation succeeds.

---

# 12. Legal Boundaries

The platform uses only open data.

---

The platform must comply with:

- data-source terms of use;
- license restrictions;
- user-agreement requirements.

---

The platform does not provide official aviation information.

---

# 13. First Public Version

Version 1.0 must support:

- aircraft display;
- airport display;
- search;
- object cards;
- route determination;
- analytics;
- historical snapshots.

---

# 14. Product Value

The main product value is:

Not the map.

---

The main product value is:

Intelligent integration of fragmented aviation data into one analytical system.

---

# 15. Engineering Value

The project demonstrates:

- geospatial data processing;
- mapping-system integration;
- streaming-data processing;
- architecture design;
- analytical-system construction;
- open aviation data processing.

---

# 16. Future Evolution

Platform evolution:

```text
Live Tracking

↓

Airport Intelligence

↓

Route Intelligence

↓

Traffic Analytics

↓

Traffic Forecasting

↓

Machine Learning

↓

Predictive Aviation Intelligence
```

---

# 17. Final Product Statement

Global Flight Analytics is a research platform for air-traffic visualization, analysis, and forecasting that combines open aviation data, airport data, and spatial analytics into one airspace observation system.
