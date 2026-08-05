# DOCUMENT 11

# FRONTEND ARCHITECTURE

# Global Flight Analytics

Version: 1.1

Status: Approved

---

# 1. Purpose

This document defines the architecture of the Global Flight Analytics Frontend application.

The document describes:

- application structure;
- routing;
- data loading;
- state management;
- map architecture;
- Backend API integration.

---

# 2. Technology Stack

Framework:

- Next.js

---

Language:

- TypeScript

---

Server State:

- TanStack Query

---

Maps:

- MapLibre

---

Styling:

- Tailwind CSS

---

# 3. Frontend Principles

## API First

The Frontend receives data only through the Backend API.

---

The Frontend does not access the following sources directly:

- OpenSky;
- OurAirports;
- OpenStreetMap;
- Wikidata.

---

## Mobile First

All screens are designed for mobile devices first.

---

## Data First

The primary value of the interface is data and analytics visualization.

---

# 4. Routing Structure

Primary routes:

```text
/

/map

/aircraft/[icao24]

/airports/[icao]

/analytics

/replay

/about
```

---

# 5. Main Pages

## Home

Contains:

- project description;
- activity map;
- quick search;
- primary system metrics.

---

## Live Map

The primary platform screen.

---

Displays:

- aircraft;
- airports;
- routes;
- air traffic.

---

## Aircraft Page

Displays:

- aircraft;
- airline;
- model;
- route;
- flight parameters.

---

## Airport Page

Displays:

- airport digital profile;
- statistics;
- infrastructure;
- routes.

---

## Analytics Page

Displays:

- charts;
- heat maps;
- statistics;
- comparative analysis.

---

## Replay Page

Displays:

- historical airspace state.

---

## About Page

Contains:

- project description;
- data sources;
- platform limitations.

---

# 6. State Management Strategy

The Frontend uses two state levels.

---

## Server State

Uses:

```text
TanStack Query
```

---

Purpose:

- retrieve data;
- cache data;
- refresh data;
- synchronize with the Backend.

---

## Client State

Uses:

```text
React State
```

---

Purpose:

- interface state;
- filters;
- modal windows;
- local settings.

---

Redux is not used in the MVP.

---

# 7. Feature Modules

The Frontend is divided into modules.

---

## Map Module

Responsible for:

- map;
- aircraft rendering;
- airport rendering.

---

## Aircraft Module

Responsible for:

- aircraft card;
- flight parameters;
- route.

---

## Airport Module

Responsible for:

- airport digital profile;
- infrastructure;
- statistics.

---

## Analytics Module

Responsible for:

- charts;
- heat maps;
- analytical panels.

---

## Replay Module

Responsible for:

- historical replay.

---

## Search Module

Responsible for:

- search;
- autocomplete;
- object navigation.

---

# 8. Project Structure

```text
src/

app/

widgets/

features/

entities/

shared/
```

---

Detailed structure:

```text
src

app

widgets

features
  search
  replay
  filters

entities
  aircraft
  airport
  route
  flight

shared
  api
  config
  lib
  types
  ui
```

---

# 9. Core Widgets

## Map Widget

Primary map widget.

---

## Aircraft Card

Aircraft card.

---

## Airport Card

Airport card.

---

## Analytics Charts

Chart collection.

---

## Replay Controls

Replay controls.

---

## Search Widget

Global search.

---

# 10. Map Architecture

The MVP uses only:

```text
MapLibre
```

---

The MVP does not use:

- Leaflet;
- Google Maps;
- Mapbox.

---

MapLibre is the only mapping engine used by the MVP.

---

# 11. Data Loading

Uses:

```text
TanStack Query
```

---

Every screen manages its own data loading.

---

Aggregated Backend DTOs are preferred.

---

The Frontend must not combine data from several external sources.

---

# 12. API Integration Rules

The Frontend communicates exclusively through:

```text
/api/v1/*
```

---

The contract is defined by:

```text
10_API_SPECIFICATION.md
```

---

Every API change requires contract updates.

---

# 13. Performance Rules

Requirements:

- minimize the number of requests;
- use TanStack Query caching;
- use lazy page loading;
- avoid unnecessary repeated requests.

---

Aggregated Backend DTOs are preferred.

---

# 14. Mobile Support

Mobile-device support is mandatory.

---

Minimum width:

```text
320px
```

---

Primary workflows must operate on mobile devices without restrictions.

---

# 15. Future Extensions

After the MVP, the following may be added:

- dark theme;
- user settings;
- saved filters;
- favorite airports;
- favorite aircraft.

---

# 16. Summary

The Frontend is the aviation-data and analytics visualization layer.

It receives data exclusively through the Backend API and provides users with tools for exploring airspace, airports, routes, and air traffic.
