# DOCUMENT 18

# TECHNICAL DECISIONS RECORD

# Global Flight Analytics

Version: 1.1

Status: Approved

---

# 1. Purpose

This document records the project's key technical decisions.

The document answers:

- which decision was made;
- why it was made;
- which alternatives were considered;
- which consequences the decision creates.

---

# 2. Frontend Framework

## Decision

Use Next.js.

---

## Status

Accepted

---

## Alternatives

- React + Vite
- Remix
- Nuxt

---

## Why

- mature ecosystem;
- excellent TypeScript support;
- convenient deployment;
- strong performance;
- future use of Server Components.

---

## Consequences

- dependency on the Next.js ecosystem;
- simplified Vercel deployment.

---

# 3. Frontend Language

## Decision

Use TypeScript.

---

## Status

Accepted

---

## Alternatives

- JavaScript

---

## Why

- static typing;
- safer changes;
- scalability;
- convenient refactoring.

---

## Consequences

- a small increase in development complexity;
- a significant reduction in errors.

---

# 4. Backend Language

## Decision

Use Go.

---

## Status

Accepted

---

## Alternatives

- Python
- Node.js
- Java

---

## Why

- high performance;
- low memory consumption;
- simple deployment;
- strong support for network services.

---

## Consequences

- a more limited ecosystem than JavaScript;
- substantial performance benefits.

---

# 5. Backend Framework

## Decision

Use Fiber.

---

## Status

Accepted

---

## Alternatives

- Gin
- Echo
- Chi

---

## Why

- high performance;
- simplicity;
- low overhead.

---

## Consequences

- a smaller ecosystem than Gin;
- high development speed.

---

# 6. Database

## Decision

Use PostgreSQL.

---

## Status

Accepted

---

## Alternatives

- MySQL
- MariaDB
- MongoDB

---

## Why

- maturity;
- reliability;
- support for analytical queries;
- compatibility with Neon.

---

## Consequences

- the data schema must be designed explicitly;
- high storage reliability.

---

# 7. Database Driver

## Decision

Use pgx.

---

## Status

Accepted

---

## Alternatives

- database/sql
- GORM

---

## Why

- high performance;
- full SQL control;
- minimal overhead.

---

# 8. Mapping Engine

## Decision

Use MapLibre.

---

## Status

Accepted

---

## Alternatives

- Google Maps
- Mapbox
- Leaflet

---

## Why

- open source;
- no commercial license restrictions;
- high WebGL performance.

---

## Consequences

- greater control over cartography;
- no dependency on a commercial provider.

---

# 9. Live Aviation Data

## Decision

Use OpenSky Network.

---

## Status

Accepted

---

## Alternatives

- FlightRadar24
- FlightAware
- ADS-B Exchange

---

## Why

- free access;
- open data;
- a large community.

---

## Consequences

- coverage limitations;
- dependency on OpenSky availability.

---

# 10. Runtime Cache

## Decision

Use Go process memory.

---

## Status

Accepted

---

## Alternatives

- Redis

---

## Why

- minimal infrastructure;
- no additional cost;
- sufficient MVP performance.

---

## Consequences

- data disappears after a process restart;
- this is acceptable for the MVP.

---

# 11. Service Architecture

## Decision

Use a modular monolith.

---

## Status

Accepted

---

## Alternatives

- microservices

---

## Why

- the project is developed by one developer;
- simple maintenance;
- minimal cost.

---

## Consequences

- simpler deployment;
- services may be extracted later.

---

# 12. Analytics Storage

## Decision

Do not use ClickHouse in the MVP.

---

## Status

Accepted

---

## Alternatives

- ClickHouse

---

## Why

- premature complexity;
- additional cost;
- insufficient data volume.

---

## Consequences

- PostgreSQL remains the only database.

---

# 13. Container Orchestration

## Decision

Do not use Kubernetes.

---

## Status

Accepted

---

## Alternatives

- Kubernetes

---

## Why

- excessive complexity for the MVP;
- high maintenance cost.

---

## Consequences

- simpler infrastructure;
- reduced support cost.

---

# 14. Backend Computation

## Decision

Do not use Python in the MVP.

---

## Status

Accepted

---

## Alternatives

- a separate Python service

---

## Why

- reduced complexity;
- reduced memory consumption;
- a single Backend.

---

## Consequences

- all calculations are implemented in Go;
- Python may be added when machine-learning tasks appear.

---

# 15. Geographic Strategy

## Decision

Launch with regional coverage.

---

## Status

Accepted

---

## Why

- lower workload;
- faster launch;
- easier hypothesis validation.

---

## Consequences

- limited initial geography;
- gradual scaling remains possible.

---

# 16. Future Technical Decisions

The following decisions may be reconsidered after MVP completion:

- Redis;
- ClickHouse;
- Python Services;
- Machine Learning Infrastructure;
- Object Storage;
- Mobile Applications.

---

# 17. Final Technical Position

The project follows these principles:

- simplicity;
- reliability;
- minimal cost;
- scalability;
- engineering justification for decisions.

Every new technical decision must have documented justification and measurable product value.
