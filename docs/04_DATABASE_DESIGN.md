# DOCUMENT 04

# DATABASE DESIGN

# Global Flight Analytics

Version: 1.1

Status: Approved

---

# 1. Purpose

This document defines the PostgreSQL database structure for the Global Flight Analytics MVP.

Primary database responsibilities:

- store reference data;
- store regions;
- store airports;
- store airlines;
- store aircraft;
- store observed flights;
- store flight states;
- store computed routes;
- store statistical air-traffic snapshots;
- store data-ingestion run history.

The database is not intended to store the complete unbounded real-time coordinate stream of every aircraft.

---

# 2. Database Principles

## Single Source Of Truth

Every domain entity has one primary storage location.

---

## Read Optimized

The database is optimized for read operations.

Primary user workflows:

- map;
- search;
- aircraft card;
- airport card;
- regional analytics.

---

## Controlled Denormalization

Some denormalization is allowed only for read-optimized aggregates.

Examples of allowed aggregates:

- traffic_snapshots;
- airport_statistics;
- route_statistics.

Reference entities must not be duplicated without control.

---

## Data Provenance

Stored data must preserve its origin.

Computed and aggregated data must include:

- data_source;
- confidence_level;
- calculated_at or last_updated_at.

---

## Open Data Only

The database stores only data from open sources.

---

# 3. Data Categories

The system divides data into categories:

## Real Data

Data obtained directly from open observation sources.

Example:

- flight_states from OpenSky.

## Enriched Data

Data obtained from reference datasets and enrichment.

Examples:

- airports;
- aircraft;
- aircraft_models;
- airlines.

## Inferred Data

Data computed by the system.

Example:

- route_predictions.

## Statistical Data

Data calculated from accumulated observations.

Examples:

- traffic_snapshots;
- airport_statistics;
- route_statistics.

---

# 4. countries

Country reference data.

## Data Category

Enriched Data

## Columns

id

uuid

primary key

---

name

text

not null

---

iso2

varchar(2)

unique

not null

---

iso3

varchar(3)

unique

not null

---

continent

text

nullable

---

created_at

timestamp

not null

---

updated_at

timestamp

not null

---

# 5. regions

Geographic regions used for filtering and analytics.

## Data Category

Enriched Data

## Columns

id

uuid

primary key

---

name

text

unique

not null

---

slug

text

unique

not null

---

description

text

nullable

---

min_latitude

numeric

not null

---

max_latitude

numeric

not null

---

min_longitude

numeric

not null

---

max_longitude

numeric

not null

---

created_at

timestamp

not null

---

updated_at

timestamp

not null

---

## MVP Regions

- Azerbaijan
- Caucasus
- Turkey

---

# 6. airlines

Airline reference data.

## Data Category

Enriched Data

## Columns

id

uuid

primary key

---

name

text

not null

---

icao_code

varchar(3)

unique

nullable

---

iata_code

varchar(2)

nullable

---

country_id

uuid

foreign key to countries.id

nullable

---

website

text

nullable

---

source_name

text

not null

---

last_synced_at

timestamp

nullable

---

created_at

timestamp

not null

---

updated_at

timestamp

not null

---

# 7. aircraft_models

Aircraft-model reference data.

## Data Category

Enriched Data

## Columns

id

uuid

primary key

---

manufacturer

text

not null

---

model

text

not null

---

aircraft_type

text

nullable

---

max_speed_kmh

integer

nullable

---

max_range_km

integer

nullable

---

passenger_capacity

integer

nullable

---

cargo_capacity_kg

integer

nullable

---

source_name

text

not null

---

last_synced_at

timestamp

nullable

---

created_at

timestamp

not null

---

updated_at

timestamp

not null

---

## Constraints

manufacturer + model must be unique.

---

# 8. aircraft

Specific physical aircraft.

## Data Category

Enriched Data

## Columns

id

uuid

primary key

---

icao24

varchar(10)

unique

not null

---

registration

text

nullable

---

model_id

uuid

foreign key to aircraft_models.id

nullable

---

airline_id

uuid

foreign key to airlines.id

nullable

---

country_id

uuid

foreign key to countries.id

nullable

---

source_name

text

not null

---

first_seen_at

timestamp

nullable

---

last_seen_at

timestamp

nullable

---

created_at

timestamp

not null

---

updated_at

timestamp

not null

---

# 9. airports

Airports.

## Data Category

Enriched Data

## Columns

id

uuid

primary key

---

icao_code

varchar(4)

unique

nullable

---

iata_code

varchar(3)

nullable

---

name

text

not null

---

city

text

nullable

---

country_id

uuid

foreign key to countries.id

nullable

---

latitude

numeric

not null

---

longitude

numeric

not null

---

elevation_ft

integer

nullable

---

timezone

text

nullable

---

source_name

text

not null

---

last_synced_at

timestamp

nullable

---

created_at

timestamp

not null

---

updated_at

timestamp

not null

---

## Constraints

At least one of icao_code or iata_code should exist when available from source data.

---

# 10. runways

Airport runways.

## Data Category

Enriched Data

## Columns

id

uuid

primary key

---

airport_id

uuid

foreign key to airports.id

not null

---

identifier

text

not null

---

length_m

integer

nullable

---

width_m

integer

nullable

---

surface

text

nullable

---

source_name

text

not null

---

created_at

timestamp

not null

---

updated_at

timestamp

not null

---

# 11. airport_facilities

Airport infrastructure.

## Data Category

Enriched Data

## Columns

id

uuid

primary key

---

airport_id

uuid

foreign key to airports.id

not null

---

facility_type

text

not null

---

name

text

nullable

---

latitude

numeric

nullable

---

longitude

numeric

nullable

---

source_name

text

not null

---

created_at

timestamp

not null

---

updated_at

timestamp

not null

---

# 12. airport_profiles

Airport digital profiles.

## Data Category

Aggregated View

## Columns

id

uuid

primary key

---

airport_id

uuid

foreign key to airports.id

unique

not null

---

description

text

nullable

---

history

text

nullable

---

passenger_traffic

bigint

nullable

---

cargo_traffic_tons

bigint

nullable

---

terminals_count

integer

nullable

---

runways_count

integer

nullable

---

metadata_json

jsonb

nullable

---

source_name

text

nullable

---

last_updated_at

timestamp

nullable

---

created_at

timestamp

not null

---

updated_at

timestamp

not null

---

# 13. flights

Observed flights.

## Data Category

Real Data

## Meaning

Flight represents an observed aircraft flight.

Flight does not store coordinates, altitude, velocity, or heading.

These values are stored in flight_states.

## Columns

id

uuid

primary key

---

aircraft_id

uuid

foreign key to aircraft.id

nullable

---

callsign

text

nullable

---

first_seen_at

timestamp

not null

---

last_seen_at

timestamp

not null

---

status

text

not null

---

created_at

timestamp

not null

---

updated_at

timestamp

not null

---

# 14. flight_states

Flight states over time.

## Data Category

Real Data

## Meaning

Flight State is a snapshot of a flight at a specific time.

MVP source:

OpenSky.

## Columns

id

uuid

primary key

---

flight_id

uuid

foreign key to flights.id

nullable

---

aircraft_id

uuid

foreign key to aircraft.id

nullable

---

icao24

varchar(10)

not null

---

callsign

text

nullable

---

latitude

numeric

nullable

---

longitude

numeric

nullable

---

barometric_altitude_m

integer

nullable

---

geometric_altitude_m

integer

nullable

---

velocity_mps

numeric

nullable

---

heading_degrees

numeric

nullable

---

vertical_rate_mps

numeric

nullable

---

on_ground

boolean

nullable

---

origin_country

text

nullable

---

observed_at

timestamp

not null

---

source_name

text

not null

---

ingestion_run_id

uuid

foreign key to ingestion_runs.id

nullable

---

created_at

timestamp

not null

---

## Retention Rule

The MVP stores a limited flight_states history.

Recommended MVP limit:

- 24 hours to 7 days for live replay;
- aggregates are retained longer in traffic_snapshots.

---

# 15. route_predictions

Computed routes.

## Data Category

Inferred Data

## Meaning

Route Prediction represents the probable route of a flight.

It is not an official route or flight plan.

## Columns

id

uuid

primary key

---

flight_id

uuid

foreign key to flights.id

nullable

---

aircraft_id

uuid

foreign key to aircraft.id

nullable

---

origin_airport_id

uuid

foreign key to airports.id

nullable

---

destination_airport_id

uuid

foreign key to airports.id

nullable

---

confidence_level

varchar(20)

not null

---

confidence_score

numeric

nullable

---

method_name

text

not null

---

data_source

text

not null

---

calculated_at

timestamp

not null

---

created_at

timestamp

not null

---

## Constraints

confidence_level must be one of:

- High
- Medium
- Low

---

# 16. traffic_snapshots

Air-traffic snapshots.

## Data Category

Statistical Data

## Meaning

Used for regional analytics and historical analysis.

## Columns

id

uuid

primary key

---

region_id

uuid

foreign key to regions.id

not null

---

snapshot_time

timestamp

not null

---

flight_count

integer

not null

---

airport_count

integer

not null

---

route_count

integer

not null

---

payload_json

jsonb

nullable

---

calculated_at

timestamp

not null

---

created_at

timestamp

not null

---

# 17. route_statistics

Route statistics.

## Data Category

Statistical Data

## Columns

id

uuid

primary key

---

route_prediction_id

uuid

foreign key to route_predictions.id

nullable

---

origin_airport_id

uuid

foreign key to airports.id

nullable

---

destination_airport_id

uuid

foreign key to airports.id

nullable

---

observation_date

date

not null

---

flight_count

integer

not null

---

created_at

timestamp

not null

---

# 18. airport_statistics

Airport statistics.

## Data Category

Statistical Data

## Columns

id

uuid

primary key

---

airport_id

uuid

foreign key to airports.id

not null

---

observation_date

date

not null

---

arrivals

integer

not null

---

departures

integer

not null

---

total_flights

integer

not null

---

created_at

timestamp

not null

---

# 19. ingestion_runs

Data-ingestion run history.

## Data Category

Operational Data

## Meaning

The table supports debugging, monitoring, and control of data ingestion from external sources.

## Columns

id

uuid

primary key

---

source_name

text

not null

---

region_id

uuid

foreign key to regions.id

nullable

---

started_at

timestamp

not null

---

finished_at

timestamp

nullable

---

status

text

not null

---

records_received

integer

not null

---

records_inserted

integer

not null

---

records_updated

integer

not null

---

error_message

text

nullable

---

created_at

timestamp

not null

---

## Constraints

status must be one of:

- running
- success
- failed
- partial

---

# 20. Relationships

Country

→ Airlines

→ Airports

→ Aircraft

---

Region

→ TrafficSnapshots

→ IngestionRuns

---

Airline

→ Aircraft

---

AircraftModel

→ Aircraft

---

Aircraft

→ Flights

→ FlightStates

→ RoutePredictions

---

Airport

→ Runways

→ AirportFacilities

→ AirportProfile

→ AirportStatistics

---

Flight

→ FlightStates

→ RoutePredictions

---

RoutePrediction

→ RouteStatistics

---

# 21. Index Strategy

## Unique Indexes

countries.iso2

countries.iso3

airlines.icao_code

aircraft_models.manufacturer + aircraft_models.model

aircraft.icao24

airports.icao_code

regions.slug

---

## Lookup Indexes

airports.iata_code

airports.country_id

aircraft.registration

aircraft.airline_id

flights.callsign

flights.aircraft_id

flight_states.icao24

flight_states.flight_id

flight_states.aircraft_id

flight_states.observed_at

flight_states.ingestion_run_id

route_predictions.flight_id

route_predictions.aircraft_id

traffic_snapshots.region_id

traffic_snapshots.snapshot_time

airport_statistics.airport_id

airport_statistics.observation_date

route_statistics.observation_date

ingestion_runs.source_name

ingestion_runs.started_at

---

## Composite Indexes

flight_states.icao24 + flight_states.observed_at

traffic_snapshots.region_id + traffic_snapshots.snapshot_time

airport_statistics.airport_id + airport_statistics.observation_date

route_statistics.origin_airport_id + route_statistics.destination_airport_id + route_statistics.observation_date

---

# 22. Retention Strategy

## flight_states

Retained for a limited period.

MVP retention:

- minimum: 24 hours;
- target: 7 days;
- longer storage only after validation of cost.

---

## traffic_snapshots

Retained longer than flight_states.

Used for analytics without storing the full coordinate stream.

---

## ingestion_runs

Retained for diagnostics and data-quality control.

---

# 23. Future Extensions

The following may be added:

- users;
- favorite airports;
- favorite aircraft;
- notifications;
- historical archives;
- materialized views;
- PostGIS.

Without changing the current architectural foundation.

---

# 24. Database Boundaries

The database does not store:

- official flight plans;
- air traffic control system data;
- military aviation data;
- closed commercial airline data;
- complete unbounded telemetry for all aircraft.

---

# 25. Summary

The database consists of the following primary domain groups:

- Countries
- Regions
- Airlines
- Aircraft Models
- Aircraft
- Airports
- Runways
- Airport Facilities
- Airport Profiles
- Flights
- Flight States
- Route Predictions
- Traffic Snapshots
- Route Statistics
- Airport Statistics
- Ingestion Runs

This schema is the foundation of the Global Flight Analytics MVP.
