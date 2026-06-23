# DOCUMENT 04

# DATABASE DESIGN

# Global Flight Analytics

Version: 1.1

Status: Approved

---

# 1. Purpose

Документ определяет структуру базы данных PostgreSQL для MVP Global Flight Analytics.

Основные задачи базы данных:

- хранение справочников;
- хранение регионов;
- хранение аэропортов;
- хранение авиакомпаний;
- хранение самолетов;
- хранение наблюдаемых рейсов;
- хранение состояний рейсов;
- хранение вычисленных маршрутов;
- хранение статистических снимков воздушного движения;
- хранение истории запусков загрузки данных.

База данных не предназначена для хранения полного бесконечного потока координат каждого самолета в реальном времени.

---

# 2. Database Principles

## Single Source Of Truth

Каждая доменная сущность имеет одно основное место хранения.

---

## Read Optimized

База оптимизируется под чтение.

Основные пользовательские сценарии:

- карта;
- поиск;
- карточка самолета;
- карточка аэропорта;
- региональная аналитика.

---

## Controlled Denormalization

Некоторая денормализация допускается только для read-optimized агрегатов.

Примеры допустимых агрегатов:

- traffic_snapshots;
- airport_statistics;
- route_statistics.

Справочные сущности не должны дублироваться бесконтрольно.

---

## Data Provenance

Данные должны сохранять происхождение.

Для вычисляемых и агрегированных данных обязательны:

- data_source;
- confidence_level;
- calculated_at или last_updated_at.

---

## Open Data Only

В базе хранятся только данные из открытых источников.

---

# 3. Data Categories

Система разделяет данные на категории:

## Real Data

Данные, полученные напрямую из открытых источников наблюдения.

Пример:

- flight_states из OpenSky.

## Enriched Data

Данные, полученные из справочников и обогащения.

Пример:

- airports;
- aircraft;
- aircraft_models;
- airlines.

## Inferred Data

Данные, вычисленные системой.

Пример:

- route_predictions.

## Statistical Data

Данные, рассчитанные на основе накопленных наблюдений.

Пример:

- traffic_snapshots;
- airport_statistics;
- route_statistics.

---

# 4. countries

Справочник стран.

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

Географические регионы для фильтрации и аналитики.

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

Справочник авиакомпаний.

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

Справочник моделей самолетов.

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

Конкретные воздушные суда.

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

Аэропорты.

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

Взлетно-посадочные полосы.

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

Инфраструктура аэропорта.

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

Цифровой паспорт аэропорта.

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

Наблюдаемые рейсы.

## Data Category

Real Data

## Meaning

Flight представляет наблюдаемый полет самолета.

Flight не хранит координаты, высоту, скорость и курс.

Эти данные хранятся в flight_states.

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

Состояния рейсов во времени.

## Data Category

Real Data

## Meaning

Flight State является снимком состояния полета в конкретный момент времени.

Источник MVP:

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

MVP хранит ограниченную историю flight_states.

Рекомендуемое ограничение MVP:

- от 24 часов до 7 дней для live replay;
- агрегаты сохраняются дольше в traffic_snapshots.

---

# 15. route_predictions

Вычисленные маршруты.

## Data Category

Inferred Data

## Meaning

Route Prediction представляет вероятный маршрут рейса.

Это не официальный маршрут и не план полета.

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

Снимки воздушного движения.

## Data Category

Statistical Data

## Meaning

Используются для региональной аналитики и исторического анализа.

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

Статистика маршрутов.

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

Статистика аэропортов.

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

История запусков загрузки данных.

## Data Category

Operational Data

## Meaning

Таблица нужна для отладки, мониторинга и контроля загрузки данных из внешних источников.

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

Хранятся ограниченно.

MVP retention:

- minimum: 24 hours;
- target: 7 days;
- longer storage only after validation of cost.

---

## traffic_snapshots

Хранятся дольше, чем flight_states.

Используются для аналитики без хранения полного координатного потока.

---

## ingestion_runs

Хранятся для диагностики и контроля качества данных.

---

# 23. Future Extensions

Допускается добавление:

- пользователей;
- избранных аэропортов;
- избранных самолетов;
- уведомлений;
- исторических архивов;
- materialized views;
- PostGIS.

Без изменения текущей архитектурной основы.

---

# 24. Database Boundaries

База данных не хранит:

- официальные планы полетов;
- данные диспетчерских систем;
- данные военной авиации;
- закрытые коммерческие данные авиакомпаний;
- полную бесконечную телеметрию всех самолетов.

---

# 25. Summary

База данных состоит из основных доменных групп:

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

Данная схема является фундаментом MVP Global Flight Analytics.
