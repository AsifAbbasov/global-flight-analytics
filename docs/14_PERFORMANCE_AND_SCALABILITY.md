# DOCUMENT 14

# PERFORMANCE AND SCALABILITY

# Global Flight Analytics

Version: 1.1

Status: Approved

---

# 1. Purpose

This document defines the performance, scalability, and resilience requirements for the Global Flight Analytics platform.

The document establishes:

- performance targets;
- MVP constraints;
- the scaling strategy;
- acceptable workloads;
- optimization rules.

---

# 2. MVP Scope

Target MVP region:

```text
Caucasus
```

---

Primary airports:

- Baku;
- Tbilisi;
- Yerevan;
- Istanbul.

---

Primary MVP objective:

Validate the architecture and user interest without expensive infrastructure.

---

# 3. Performance Goals

## Home Page

Complete load time:

```text
up to 2 seconds
```

---

## Live Map

Initial map rendering:

```text
up to 3 seconds
```

---

## Airport Profile

Complete page load:

```text
up to 1 second
```

---

## Aircraft Profile

Complete page load:

```text
up to 1 second
```

---

# 4. Backend Performance Targets

Average response time:

```text
up to 300 milliseconds
```

---

Maximum response time:

```text
up to 1000 milliseconds
```

---

Target:

```text
95% of requests
below 500 milliseconds
```

---

# 5. Database Performance Targets

## Airport Lookup

```text
up to 100 milliseconds
```

---

## Aircraft Lookup

```text
up to 100 milliseconds
```

---

## Statistics Queries

```text
up to 500 milliseconds
```

---

All critical fields must be indexed.

---

# 6. Frontend Performance Targets

Requirements:

- minimize repeated renders;
- minimize the number of requests;
- use caching;
- use lazy page loading.

---

# 7. API Optimization Strategy

The platform uses:

- aggregated DTOs;
- server-side data aggregation;
- pagination;
- response-size limits.

---

The Frontend must not combine multiple data sources by itself.

---

# 8. Map Rendering

The platform uses:

```text
MapLibre
```

---

Reasons:

- WebGL;
- high performance;
- support for large numbers of objects.

---

# 9. Memory Strategy

Current aircraft states are stored in Backend memory.

---

Implementation:

```text
sync.Map
```

---

Purpose:

- fast aircraft-state lookup;
- reduced PostgreSQL load;
- Live Map operation.

---

# 10. Data Retention Strategy

PostgreSQL does not persist every coordinate for every aircraft.

---

The platform persists:

- aggregates;
- statistics;
- state snapshots;
- analytical data.

---

This limits database growth.

---

# 11. Scalability Stages

## Stage 1

Caucasus.

---

## Stage 2

Caucasus and Turkey.

---

## Stage 3

Europe.

---

## Stage 4

Global coverage.

---

# 12. Future Scaling Options

After workload evidence is available, the following components may be introduced:

- Redis;
- Object Storage;
- Dedicated Server;
- CDN for analytical data.

---

These components are not used in the MVP.

---

# 13. Bottlenecks

Potential bottlenecks:

- OpenSky limitations;
- Backend memory;
- rendering large numbers of objects;
- expensive analytical queries.

---

# 14. Optimization Strategy

Primary optimization methods:

- caching;
- aggregation;
- precomputation;
- database indexing;
- limiting the observation region.

---

# 15. Success Metrics

The platform must operate correctly on:

- laptops;
- tablets;
- mobile devices.

---

Supported browsers:

- Chrome;
- Safari;
- Edge;
- Firefox.

---

# 16. Cost-Aware Scaling

Scaling must not cause a sudden increase in costs.

---

Every new infrastructure component must have a workload-based justification.

---

# 17. Scalability Boundaries

The MVP does not use:

- Kubernetes;
- Kafka;
- RabbitMQ;
- distributed clusters;
- dedicated analytical servers.

---

Architectural complexity is allowed only after a real need appears.

---

# 18. Summary

Platform performance is achieved through:

- Go;
- PostgreSQL;
- MapLibre;
- TanStack Query;
- in-memory caching;
- server-side data aggregation.

The architecture allows the MVP to operate on free infrastructure and scale gradually toward regional and international coverage.
