# DOCUMENT 05

# DATA SOURCES

# Global Flight Analytics

Version: 1.1

Status: Approved

---

# 1. Purpose

This document defines all data sources used by the Global Flight Analytics platform.

The document answers:

- where the data comes from;
- which data is considered primary;
- which data is considered supplementary;
- how often data is refreshed;
- which data can be trusted;
- which limitations exist.

---

# 2. Data Source Strategy

The project uses only:

- open data;
- legally available data;
- free data;
- publicly available data.

The project does not use:

- compromised databases;
- closed aviation systems;
- paid corporate aviation data;
- air traffic control data.

---

# 3. Source Categories

Sources are divided into four groups.

## Category 1

Live Aviation Data

Real-time data.

## Category 2

Airport Data

Airport data.

## Category 3

Aircraft Data

Aircraft data.

## Category 4

Infrastructure Data

Infrastructure data.

---

# 4. Primary Live Source

OpenSky Network

Primary MVP data source.

Provides:

- aircraft coordinates;
- altitude;
- velocity;
- heading;
- vertical rate;
- ICAO24 identifier;
- callsign;
- signal timestamp;
- on-ground status.

Does not provide:

- passenger data;
- ticket prices;
- aircraft load;
- commercial airline data;
- official flight plans.

---

# 5. OpenSky Limitations

The following limitations must be considered:

- coverage is uneven;
- data quality is stronger over Europe;
- gaps may occur over oceans;
- aircraft may temporarily disappear over some regions;
- some aircraft have incomplete attributes.

---

# 6. Airport Source

OurAirports

Primary airport data source.

Provides:

- ICAO code;
- IATA code;
- name;
- coordinates;
- elevation;
- airport type;
- runway data.

Usage:

Initial airport reference-data import.

Refresh frequency:

Monthly.

---

# 7. Aircraft Source

OpenSky Aircraft Database

Primary aircraft data source.

Provides:

- ICAO24;
- registration number;
- airline;
- registration country;
- aircraft type.

Usage:

Enrichment of OpenSky observation data.

---

# 8. Aircraft Specification Sources

Sources:

- Wikidata;
- Wikipedia;
- official manufacturer websites.

Provide:

- manufacturer;
- model;
- capacity;
- range;
- maximum speed;
- cargo capacity.

Usage:

Populate the aircraft_models reference table.

---

# 9. Airport Infrastructure Source

OpenStreetMap

Provides:

- roads;
- railway stations;
- bus stops;
- parking areas;
- hotels;
- infrastructure objects.

Usage:

Build the airport digital profile.

Important:

The application does not issue direct OpenStreetMap requests during normal operation.

Data is imported in advance and stored in PostgreSQL.

---

# 10. Geographic Source

OpenStreetMap

Used for:

- coordinates;
- object geometry;
- regional boundaries;
- spatial analytics.

---

# 11. Knowledge Sources

Wikipedia

Wikidata

Used for:

- airport history;
- object descriptions;
- reference information;
- notable facts.

---

# 12. Data Trust Levels

Every source receives a trust level.

## HIGH

Very high reliability.

Examples:

- OpenSky coordinates;
- ICAO code;
- IATA code;
- airport coordinates.

## MEDIUM

Sufficient reliability.

Examples:

- airline;
- aircraft model;
- registration number.

## LOW

Estimated data.

Examples:

- probable route;
- inferred destination airport;
- analytical forecasts.

---

# 13. Derived Data

Some data is calculated by the system.

Example:

Route Prediction.

The system determines:

- High
- Medium
- Low

Based on:

- heading;
- velocity;
- altitude;
- aircraft position;
- distance to airports;
- observation history.

All computed data must contain:

- confidence_level;
- data_source;
- calculated_at.

---

# 14. Data Refresh Policy

OpenSky

Every 15–30 seconds.

---

OurAirports

Monthly.

---

OpenSky Aircraft Database

Monthly.

---

OpenStreetMap

Weekly.

---

Wikipedia

On demand.

---

Wikidata

On demand.

---

# 15. Unsupported Data

The platform does not present the following as fact:

- passenger count;
- baggage count;
- ticket price;
- business-class price;
- internal airline data;
- air traffic control data;
- flight plans.

If such data appears in the future, it must be labeled explicitly as estimated.

---

# 16. Data Ownership

The platform does not own the source data.

The platform owns:

- the data structure;
- aggregated statistics;
- historical snapshots;
- analytical representations;
- route-calculation algorithms.

---

# 17. Source Reliability Matrix

| Source                    | Category            | Refresh Rate  | Trust Level | Criticality |
| ------------------------- | ------------------- | ------------- | ----------- | ----------- |
| OpenSky Network           | Live Aviation Data  | 15–30 seconds | High        | Critical    |
| OurAirports               | Airport Data        | Monthly       | High        | Critical    |
| OpenSky Aircraft Database | Aircraft Data       | Monthly       | Medium      | High        |
| OpenStreetMap             | Infrastructure Data | Weekly        | High        | Medium      |
| Wikidata                  | Knowledge Data      | On Demand     | Medium      | Medium      |
| Wikipedia                 | Knowledge Data      | On Demand     | Medium      | Low         |

---

# 18. Summary

Primary platform sources:

- OpenSky Network
- OurAirports
- OpenSky Aircraft Database
- OpenStreetMap
- Wikidata
- Wikipedia

This source set is sufficient to build a complete air-traffic analysis and visualization platform without paid aviation services.
