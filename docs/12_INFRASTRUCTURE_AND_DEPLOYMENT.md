# DOCUMENT 12

# INFRASTRUCTURE AND DEPLOYMENT

# Global Flight Analytics

Version: 1.1

Status: Approved

---

# 1. Purpose

This document defines the infrastructure, environments, deployment, and operation of the Global Flight Analytics platform.

Goals:

- allow one developer to launch the MVP;
- minimize cost;
- preserve the ability to scale without a complete architecture rewrite.

---

# 2. Infrastructure Principles

## Free First

The MVP uses free or limited-free services.

---

## Simplicity First

The MVP does not use:

- Kubernetes;
- Redis;
- Kafka;
- RabbitMQ;
- microservices.

---

## Monolith First

The system is deployed as a modular monolith.

---

## Open Data First

Only open data sources are used.

---

# 3. Frontend Hosting

Platform:

```text
Vercel
```

---

Purpose:

- deploy Next.js;
- deliver static assets globally;
- deploy automatically through GitHub.

---

# 4. Backend Hosting

Platform:

```text
Render
```

---

Purpose:

- host the Go API;
- execute background tasks;
- run the Data Collection Pipeline.

---

Reasons:

- stable free plan;
- simple GitHub integration;
- Go support;
- fast MVP launch.

---

# 5. Database

Platform:

```text
Neon PostgreSQL
```

---

Purpose:

- reference data;
- airports;
- aircraft;
- routes;
- statistics;
- analytics.

---

# 6. Live Data Layer

Location:

```text
Go Application Memory
```

---

Implementation:

```text
sync.Map
```

---

Purpose:

- store current aircraft states;
- operate the live map;
- construct API responses quickly.

---

# 7. Source Layer

Primary sources:

- OpenSky Network;
- OurAirports;
- OpenStreetMap;
- Wikipedia;
- Wikidata.

---

# 8. Deployment Architecture

High-level structure:

```text
GitHub Repository

       ├── Vercel
       │      │
       │      ▼
       │   Frontend
       │
       ├── Render
       │      │
       │      ▼
       │   Go Backend
       │
       └── Neon
              │
              ▼
         PostgreSQL
```

---

# 9. Deployment Flow

Publication flow:

```text
Developer

↓

GitHub

↓

Automatic Deployment

↓

Vercel Frontend

↓

Render Backend

↓

Neon Database
```

---

# 10. Environment Variables

## Backend

```text
DATABASE_URL

OPENSKY_USERNAME

OPENSKY_PASSWORD

APP_ENV

API_PORT
```

---

## Frontend

```text
NEXT_PUBLIC_API_URL
```

---

Secrets must never be committed to the repository.

---

# 11. Monitoring

Minimum MVP monitoring:

- Application Logs;
- Health Endpoint;
- Error Logs;
- Database Logs.

---

Additional monitoring:

- Uptime Monitoring;
- Health Check Monitoring;
- Deployment Monitoring.

---

# 12. Backup Strategy

Source:

```text
Neon PostgreSQL
```

---

Strategy:

```text
Daily Backup
```

---

Retention:

```text
7 days
```

---

Purpose:

- protect against data loss;
- recover from failures;
- test migrations.

---

# 13. Security Rules

Mandatory rules:

- HTTPS Only;
- No Direct Database Access;
- No Client Secrets;
- Environment Variables Only;
- Backend API Only.

---

Additional rules:

- CORS Restrictions;
- Rate Limiting;
- Request Validation.

---

# 14. Failure Strategy

## OpenSky Unavailable

Action:

```text
Use the latest available in-memory data
```

---

The Frontend displays:

```text
Live Data Delayed
```

---

## Database Unavailable

Action:

```text
Read Only Mode
```

---

## Backend Restart

Action:

```text
Reload data
from OpenSky
```

---

# 15. Scaling Strategy

## Phase 1

Region:

```text
Caucasus
```

---

## Phase 2

Region:

```text
Europe
```

---

## Phase 3

Region:

```text
Global
```

---

Scaling does not change the architectural principles.

---

# 16. Cost Target

Target MVP budget:

```text
0–5 USD per month
```

---

The architecture must remain operational even when providers change their free plans.

---

# 17. Infrastructure Boundaries

The MVP does not use:

- Kubernetes;
- Redis;
- Kafka;
- RabbitMQ;
- Elasticsearch;
- dedicated analytical clusters.

---

Infrastructure is made more complex only after proven workload appears.

---

# 18. Summary

The platform infrastructure is built around:

- Vercel;
- Render;
- Neon PostgreSQL;
- OpenSky Network;
- OpenStreetMap;
- OurAirports;
- Wikidata.

The architecture allows one developer to launch and maintain the MVP at minimal cost while preserving future scalability.
