# DOCUMENT 01

# PRODUCT VISION

# Global Flight Analytics

Version: 1.0

Status: Approved

---

# 1. Project Overview

Global Flight Analytics is a research platform for analyzing and visualizing air traffic using open aviation data.

The platform collects data from open sources, combines it with aviation reference data, and provides users with a unified environment for exploring aircraft, airports, routes, and air traffic.

The project is not an air traffic control system.

The project is not a flight planning system.

The project is not a ticket-purchasing service.

The project is not an official source of aeronautical information.

The primary objective of the project is to help users explore and understand what is happening in airspace through visualization, statistics, and open-data analysis.

---

# 2. Problem Statement

Aviation information is currently fragmented.

To study a single flight, a user often has to use several different services:

- an aircraft tracking service;
- an airport directory;
- aviation encyclopedias;
- mapping services;
- airline websites.

Most existing platforms focus exclusively on displaying aircraft coordinates.

They provide limited explanations of:

- why an aircraft follows a particular route;
- how air traffic changes;
- which airports are key hubs;
- how air corridors change during major events.

---

# 3. Product Vision

The platform must answer not only:

“Where is the aircraft now?”

It must also answer:

- What is happening with air traffic in the region?
- Which routes are currently the most active?
- Which airports handle the highest traffic volume?
- How do routes change over time?
- How does the structure of air traffic change after major events?

---

# 4. Target Audience

Primary audience:

- aviation enthusiasts;
- travelers;
- aviation students;
- transport researchers;
- journalists;
- developers;
- educators;
- transport-system analysts.

---

# 5. Core Value

The platform's primary value is the integration of aviation data into one research system.

The user receives:

- real aircraft on the map;
- aircraft information;
- airport information;
- inferred routes;
- air-flow visualization;
- air-traffic statistics.

---

# 6. Product Principles

## Transparency

The system must distinguish real data from computed data.

The user must always understand the origin of the information.

## Data First

The platform is built around data and analytics rather than visual effects.

## Explainability

Every computed metric must include an explanation of how it was produced.

## Open Sources

The platform uses only open data sources.

---

# 7. Data Classification

All data is divided into four categories.

## Real Data

Obtained directly from open sources.

Examples:

- coordinates;
- altitude;
- velocity;
- heading;
- callsign;
- data-retrieval timestamp.

## Enriched Data

Obtained by matching reference datasets.

Examples:

- aircraft model;
- manufacturer;
- airline;
- aircraft specifications;
- airport specifications.

## Inferred Data

Produced by platform algorithms.

Examples:

- inferred route;
- inferred destination airport;
- inferred departure airport;
- route confidence level.

## Statistical Data

Produced from accumulated observations.

Examples:

- number of flights;
- airport activity;
- route distribution;
- air-traffic trends.

---

# 8. Core Entities

The platform is built around the following entities.

## Aircraft

A physical aircraft.

## Flight

An observed flight.

## Airport

An airport.

## Airline

An airline.

## Route

A route between airports.

## Region

A geographic region.

---

# 9. MVP Features

## Live Flight Map

Real-time aircraft display.

## Aircraft Profile

Aircraft information card.

Contains:

- model;
- manufacturer;
- operator;
- specifications;
- current state.

## Airport Profile

Airport information card.

Contains:

- description;
- specifications;
- infrastructure;
- statistics.

## Route Intelligence Engine

Determines the probable route.

Confidence levels:

- High
- Medium
- Low

## Search

Search for aircraft, airports, and airlines.

## Air Traffic Intelligence

Air-traffic analysis.

Contains:

- active routes;
- active airports;
- traffic intensity;
- heat maps.

---

# 10. Regional Analytics

The user can:

- select a region;
- view the number of observed flights;
- view major air corridors;
- view the most active airports;
- compare activity by day;
- explore changes in air traffic.

---

# 11. Non Goals

The MVP does not include:

- ticket purchases;
- hotel reservations;
- taxi ordering;
- air traffic control;
- dispatch functions;
- flight operations management;
- investment analytics;
- stock-market forecasting;
- mobile applications;
- a payment system;
- user subscriptions.

---

# 12. Data Sources

Primary data sources:

- OpenSky Network;
- OurAirports;
- OpenStreetMap;
- Wikidata.

---

# 13. Technology Stack

Frontend:

- Next.js;
- TypeScript;
- MapLibre;
- TanStack Query.

Backend:

- Go;
- Fiber;
- pgx.

Database:

- PostgreSQL.

Infrastructure:

- Vercel;
- a free cloud server for the Go API;
- a free PostgreSQL database.

---

# 14. Success Criteria

The MVP is successful when a user can:

1. open the map;
2. see real aircraft;
3. open an aircraft card;
4. view an inferred route;
5. open an airport card;
6. obtain airport information;
7. explore regional air traffic;
8. search for aviation objects.
