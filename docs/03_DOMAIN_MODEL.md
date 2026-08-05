# DOCUMENT 03

# DOMAIN MODEL

# Global Flight Analytics

Version: 1.1

Status: Approved

---

# 1. Purpose

This document describes the system domain.

The document aims to:

- define core entities;
- define relationships between entities;
- define the business meaning of each entity;
- establish a foundation for database and API design.

---

# 2. Domain Overview

The platform is built around air-traffic observation.

Primary system objects:

- aircraft;
- airlines;
- airports;
- flights;
- flight states;
- routes;
- regions;
- airport infrastructure.

---

# 3. Aircraft

A physical aircraft.

Represents one specific aircraft.

## Data Category

Enriched Data

## Business Meaning

Aircraft is the primary observation object.

Through an aircraft, the user accesses all related information.

## Examples

- Airbus A320
- Boeing 737
- Airbus A350
- Boeing 777

## Core Attributes

- ICAO24 Identifier
- Registration Number
- Aircraft Model
- Operator
- Country
- Manufacturer

## Relationships

Aircraft

→ AircraftModel

→ Airline

→ Flight

---

# 4. Aircraft Model

An aircraft type.

## Data Category

Enriched Data

## Business Meaning

Describes technical characteristics of an aircraft model.

## Examples

- Airbus A320-214
- Boeing 737-800
- Airbus A350-900

## Core Attributes

- Name
- Manufacturer
- Maximum Range
- Maximum Speed
- Passenger Capacity
- Cargo Capacity

## Relationships

AircraftModel

← Aircraft

---

# 5. Airline

An airline.

## Data Category

Enriched Data

## Business Meaning

Operates aircraft.

## Examples

- Azerbaijan Airlines
- Lufthansa
- Turkish Airlines
- Emirates

## Core Attributes

- Name
- ICAO Code
- IATA Code
- Country

## Relationships

Airline

→ Aircraft

→ Flight

---

# 6. Airport

An airport.

## Data Category

Enriched Data

## Business Meaning

An air-traffic hub.

## Examples

- GYD
- IST
- FRA
- LHR

## Core Attributes

- ICAO Code
- IATA Code
- Name
- Latitude
- Longitude
- Elevation

## Relationships

Airport

→ Runway

→ Route

→ AirportFacility

---

# 7. Runway

An airport runway.

## Data Category

Enriched Data

## Business Meaning

An airport infrastructure object.

## Core Attributes

- Identifier
- Length
- Width
- Surface Type

## Relationships

Runway

← Airport

---

# 8. Airport Facility

An airport infrastructure object.

## Data Category

Enriched Data

## Business Meaning

Contributes to the airport digital profile.

## Examples

- Railway Station
- Bus Terminal
- Parking Area
- Hotel
- Cargo Terminal

## Core Attributes

- Name
- Type
- Coordinates

## Relationships

AirportFacility

← Airport

---

# 9. Flight

An observed aircraft flight.

## Data Category

Real Data

## Business Meaning

Represents an observed flight performed by a specific aircraft.

## Important

Flight is not the same as Aircraft.

One aircraft performs many flights.

## Core Attributes

- Callsign
- Flight Identifier
- Start Timestamp
- End Timestamp

## Relationships

Flight

← Aircraft

→ Route

→ FlightState

---

# 10. Flight State

A snapshot of a flight state at a specific time.

## Data Category

Real Data

## Business Meaning

An observation received from OpenSky.

Used to construct flight history and analytics.

## Core Attributes

- Latitude
- Longitude
- Altitude
- Velocity
- Heading
- Timestamp

## Relationships

FlightState

← Flight

---

# 11. Route

A route between two airports.

## Data Category

Inferred Data

## Business Meaning

A computed entity.

## Core Attributes

- Origin Airport
- Destination Airport
- Confidence Level

## Relationships

Route

← Flight

→ Airport

---

# 12. Region

A geographic region.

## Data Category

Enriched Data

## Business Meaning

Used for air-traffic analytics.

## Examples

- Caucasus
- Europe
- Middle East
- Central Asia

## Core Attributes

- Name
- Boundaries

## Relationships

Region

→ Airport

→ Flight

---

# 13. Traffic Snapshot

An air-traffic snapshot.

## Data Category

Statistical Data

## Business Meaning

Used for statistics and historical analysis.

## Core Attributes

- Timestamp
- Region
- Flight Count
- Airport Count
- Route Count

## Relationships

TrafficSnapshot

→ Region

---

# 14. Airport Profile

An airport digital profile.

## Data Category

Aggregated View

## Business Meaning

Aggregates all airport information.

## Contains

### Static Information

- name;
- location;
- history;
- description.

### Infrastructure

- runways;
- terminals;
- transport.

### Statistics

- passenger traffic;
- cargo traffic;
- routes.

---

# 15. Aircraft Profile

An extended aircraft description.

## Data Category

Aggregated View

## Business Meaning

Combines data from several domain entities.

## Contains

### Aircraft Information

- registration;
- model;
- manufacturer.

### Flight Information

- velocity;
- altitude;
- direction.

### Route Information

- inferred route;
- confidence level.

---

# 16. Air Traffic Intelligence

The air-traffic analytics module.

## Data Category

Statistical Data

## Purpose

Understand airspace behavior.

## Metrics

- number of flights;
- active airports;
- active routes;
- traffic density;
- change dynamics.

---

# 17. Confidence Level

The confidence level of a calculation.

## High

The route is supported by several indicators.

## Medium

The route is probable.

## Low

Insufficient data is available.

---

# 18. Domain Rules

## Rule 1

One Aircraft can perform many Flights.

## Rule 2

Each Flight belongs to only one Aircraft.

## Rule 3

One Flight contains many Flight States.

## Rule 4

Route is a computed entity.

## Rule 5

Airport is an independent reference entity.

## Rule 6

Region is used only for analytics.

## Rule 7

The system must distinguish:

- Real Data
- Enriched Data
- Inferred Data
- Statistical Data

---

# 19. Data Provenance Rules

Every system entity must belong to one data category.

Allowed categories:

- Real Data
- Enriched Data
- Inferred Data
- Statistical Data

Mixing categories inside one domain entity is prohibited.

---

# 20. Domain Boundaries

The system works only with open aviation data.

The system does not store:

- air traffic controller flight plans;
- internal airline data;
- air traffic control data;
- closed aviation data.

---

# 21. Domain Summary

Primary system entities:

Aircraft

AircraftModel

Airline

Airport

Runway

AirportFacility

Flight

FlightState

Route

Region

TrafficSnapshot

All other modules are built around these objects.
