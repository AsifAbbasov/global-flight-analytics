# Analytical Core Architecture V2

## Goal

Build an analytical core that allows new metrics to be added without changing the existing architecture.

---

## Principles

- One package has one responsibility.
- Minimal coupling.
- Maximum extensibility.
- The domain does not depend on HTTP.
- The domain does not depend on PostgreSQL.
- The domain does not depend on external data providers.
- All calculations are deterministic.

---

## Core Entities

### Metric

One unit of analytics.

Examples:

- Active Aircraft
- Traffic Density
- Airport Activity
- Arrivals Proxy
- Departures Proxy
- Coverage Score
- Data Freshness

---

### Snapshot

The complete calculated analytical state at a specific point in time.

---

### Time Window

The calculation interval.

Examples:

- Now
- 5 minutes
- 15 minutes
- 1 hour
- 24 hours

---

### Calculator

A component that calculates one specific metric.

---

### Aggregation

Combines Calculator results into a Snapshot.

---

### Projection

A data model for the Frontend.

---

### Query

A description of required data without calculations.

---

## Dependencies

Frontend

↓

HTTP

↓

Application

↓

Analytical Core

↓

Repository

↓

Database

Reverse dependencies are prohibited.

---

## Rules

Calculator:

- does not know HTTP;
- does not know PostgreSQL;
- does not know JSON;
- receives data;
- returns a result.

Repository:

- only retrieves data;
- performs no calculations.

Projection:

- contains no business logic.

---

## Adding a New Metric

Only the following additions are allowed:

- a new Calculator;
- Calculator registration;
- tests.

Existing Calculator implementations must not be changed.

---

## Development Plan

### Stage 1

Analytical Core Foundation

### Stage 2

Traffic Metrics

- Active Aircraft
- Traffic Density
- Airport Activity
- Data Freshness
- Coverage Score

### Stage 3

Airport Intelligence

### Stage 4

Route Intelligence

### Stage 5

Airspace Intelligence

### Stage 6

Frontend Integration
