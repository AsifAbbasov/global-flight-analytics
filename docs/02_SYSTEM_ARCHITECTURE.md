# DOCUMENT 02

# SYSTEM ARCHITECTURE

# Global Flight Analytics

Version: 1.1

Status: Approved

---

# 1. Purpose

This document describes the architecture of the Global Flight Analytics platform.

The document defines:

- system architecture boundaries;
- core components;
- data flows;
- module responsibilities;
- interactions between components.

---

# 2. Architecture Overview

The system consists of four primary layers:

```text
External Data Sources

        ↓

Backend API / Ingestion Layer

        ↓

PostgreSQL

        ↑

Backend API

        ↓

Frontend Application
```

The data flow works as follows:

1. External sources provide data to the Backend.
2. The Backend normalizes and processes the data.
3. The Backend stores required data in PostgreSQL.
4. The Frontend never accesses the database directly.
5. The Frontend receives data exclusively through the Backend API.

---

# 3. High Level Architecture

```text
                OpenSky Network
                       │
                       ▼

                Go Backend API
                       │

      ┌────────┬────────┬────────┬────────┬────────┐
      │        │        │        │        │
      ▼        ▼        ▼        ▼        ▼

 Aircraft   Flights  Airports  Routes  Analytics

                       │
                       ▼

                  PostgreSQL

                       ▲
                       │

                Go Backend API

                       │
                       ▼

                    Next.js

                       │
                       ▼

                 User Browser
```

---

# 4. System Components

## Frontend Layer

Technologies:

- Next.js
- TypeScript
- MapLibre
- TanStack Query

Responsibilities:

- map rendering;
- aircraft rendering;
- airport rendering;
- statistics rendering;
- search;
- user workflows.

The Frontend does not communicate directly with external APIs.

The Frontend communicates only through the Backend API.

---

## Backend Layer

Technologies:

- Go
- Fiber
- pgx

Responsibilities:

- retrieve data from OpenSky;
- process client requests;
- calculate routes;
- work with reference datasets;
- construct API responses;
- cache data.

The Backend is the only data-access point.

---

## Database Layer

Technology:

- PostgreSQL

Responsibilities:

- store airports;
- store aircraft;
- store airlines;
- store route-detection results;
- store statistics;
- store reference data.

The database is not used to store millions of coordinates without limits.

Only required data is persisted.

---

## External Data Layer

Sources:

### OpenSky Network

Used to retrieve:

- coordinates;
- velocity;
- altitude;
- heading;
- callsign;
- flight state.

### OurAirports

Used to retrieve:

- airports;
- ICAO codes;
- IATA codes;
- coordinates;
- runway information.

### OpenStreetMap

Used to retrieve:

- transport infrastructure;
- roads;
- railway stations;
- bus routes.

### Wikidata

Used to retrieve:

- descriptions;
- historical information;
- reference information.

---

# 5. Data Flow

## Aircraft Flow

```text
OpenSky

   ↓

Go Backend

   ↓

Aircraft Module

   ↓

Route Intelligence Module

   ↓

API Response

   ↓

Next.js
```

Stages:

1. The Backend receives OpenSky data.
2. The data is normalized.
3. The aircraft is matched.
4. The route is determined.
5. The client response is constructed.

---

# 6. Route Intelligence Flow

Goal:

Determine the probable aircraft route.

Input data:

- coordinates;
- altitude;
- velocity;
- heading;
- callsign.

Sources:

- OpenSky;
- Airports Database.

Result:

```text
Origin Airport

Destination Airport

Confidence Level
```

Confidence levels:

- High
- Medium
- Low

---

# 7. Data Provenance

The system must distinguish:

- Real Data
- Enriched Data
- Inferred Data
- Statistical Data

Every computed object must contain:

- data_source
- confidence_level
- last_updated_at

The user must be able to identify the origin of displayed data.

---

# 8. Airport Intelligence Flow

Goal:

Build an airport digital profile.

Sources:

- OurAirports
- OpenStreetMap
- Wikidata

Result:

- airport characteristics;
- infrastructure;
- statistics;
- popular routes.

---

# 9. Air Traffic Intelligence Flow

Goal:

Analyze regional air traffic.

Computed indicators:

- number of aircraft;
- number of routes;
- active airports;
- air-traffic density;
- traffic intensity.

Presentation:

- heat maps;
- charts;
- statistical panels.

---

# 10. Live Mode

Real-time mode.

Source:

OpenSky.

Characteristics:

- current data;
- periodic updates;
- live traffic display.

---

# 11. Historical Replay Mode

Historical replay mode.

Purpose:

Review accumulated historical data.

The user can:

- select a date;
- select a region;
- replay traffic;
- analyze changes.

The MVP stores a limited observation history sufficient for basic analytics and replay of recent events.

---

# 12. Caching Strategy

The Backend uses application memory for frequently accessed data.

Cached data:

- airports;
- airlines;
- aircraft models;
- latest OpenSky data.

Goal:

Minimize PostgreSQL access.

---

# 13. Security Principles

The Frontend never accesses OpenSky directly.

All requests pass through the Backend.

The Backend performs:

- data validation;
- data filtering;
- request-rate control.

---

# 14. Scalability Strategy

During the MVP:

```text
Next.js

    ↓

Go API

    ↓

PostgreSQL
```

Without:

- microservices;
- message queues;
- Kubernetes;
- Redis.

Architecture is made more complex only after real workload appears.

---

# 15. Architectural Boundaries

The system is NOT:

- an air traffic control system;
- a dispatch system;
- a route-planning system;
- a flight-safety system;
- an official aviation service.

The system is a research platform for analyzing and visualizing air traffic using open aviation data.
