# DOCUMENT 03

# DOMAIN MODEL

# Global Flight Analytics

Version: 1.1

Status: Approved

---

# 1. Purpose

Документ описывает предметную область системы.

Цель документа:

- определить основные сущности;
- определить связи между сущностями;
- определить бизнес-смысл каждой сущности;
- создать основу для проектирования базы данных и API.

---

# 2. Domain Overview

Платформа строится вокруг наблюдения за воздушным движением.

Основные объекты системы:

- самолеты;
- авиакомпании;
- аэропорты;
- рейсы;
- состояния рейсов;
- маршруты;
- регионы;
- инфраструктура аэропортов.

---

# 3. Aircraft

Воздушное судно.

Представляет конкретный физический самолет.

## Data Category

Enriched Data

## Business Meaning

Самолет является главным объектом наблюдения.

Через самолет пользователь получает доступ ко всей связанной информации.

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

Тип самолета.

## Data Category

Enriched Data

## Business Meaning

Описывает технические характеристики модели.

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

Авиакомпания.

## Data Category

Enriched Data

## Business Meaning

Эксплуатирует самолеты.

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

Аэропорт.

## Data Category

Enriched Data

## Business Meaning

Узел воздушного движения.

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

Взлетно-посадочная полоса.

## Data Category

Enriched Data

## Business Meaning

Инфраструктурный объект аэропорта.

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

Объект инфраструктуры аэропорта.

## Data Category

Enriched Data

## Business Meaning

Помогает формировать цифровой паспорт аэропорта.

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

Наблюдаемый полет самолета.

## Data Category

Real Data

## Business Meaning

Представляет факт выполнения полета конкретным самолетом.

## Important

Flight не равен Aircraft.

Один самолет выполняет множество рейсов.

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

Снимок состояния полета в конкретный момент времени.

## Data Category

Real Data

## Business Meaning

Является наблюдением, полученным из OpenSky.

Используется для построения истории полета и аналитики.

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

Маршрут между двумя аэропортами.

## Data Category

Inferred Data

## Business Meaning

Является вычисляемой сущностью.

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

Географический регион.

## Data Category

Enriched Data

## Business Meaning

Используется для аналитики воздушного движения.

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

Снимок воздушного движения.

## Data Category

Statistical Data

## Business Meaning

Используется для статистики и исторического анализа.

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

Цифровой паспорт аэропорта.

## Data Category

Aggregated View

## Business Meaning

Агрегирует всю информацию об аэропорте.

## Contains

### Static Information

- название;
- местоположение;
- история;
- описание.

### Infrastructure

- полосы;
- терминалы;
- транспорт.

### Statistics

- пассажиропоток;
- грузопоток;
- маршруты.

---

# 15. Aircraft Profile

Расширенное описание самолета.

## Data Category

Aggregated View

## Business Meaning

Объединяет данные из нескольких доменных сущностей.

## Contains

### Aircraft Information

- регистрация;
- модель;
- производитель.

### Flight Information

- скорость;
- высота;
- направление.

### Route Information

- предполагаемый маршрут;
- уровень уверенности.

---

# 16. Air Traffic Intelligence

Модуль аналитики воздушного движения.

## Data Category

Statistical Data

## Purpose

Понимание поведения воздушного пространства.

## Metrics

- количество рейсов;
- активные аэропорты;
- активные маршруты;
- плотность трафика;
- динамика изменений.

---

# 17. Confidence Level

Уровень уверенности вычислений.

## High

Маршрут подтвержден несколькими признаками.

## Medium

Маршрут вероятен.

## Low

Недостаточно данных.

---

# 18. Domain Rules

## Rule 1

Один Aircraft может выполнять множество Flight.

## Rule 2

Каждый Flight принадлежит только одному Aircraft.

## Rule 3

Один Flight содержит множество Flight State.

## Rule 4

Route является вычисляемой сущностью.

## Rule 5

Airport является независимой справочной сущностью.

## Rule 6

Region используется только для аналитики.

## Rule 7

Система обязана разделять:

- Real Data
- Enriched Data
- Inferred Data
- Statistical Data

---

# 19. Data Provenance Rules

Каждая сущность системы должна принадлежать одной категории данных.

Допускаются категории:

- Real Data
- Enriched Data
- Inferred Data
- Statistical Data

Смешивание категорий внутри одной доменной сущности запрещено.

---

# 20. Domain Boundaries

Система работает только с открытыми авиационными данными.

Система не хранит:

- планы полетов диспетчеров;
- внутренние данные авиакомпаний;
- данные управления воздушным движением;
- закрытые авиационные данные.

---

# 21. Domain Summary

Главные сущности системы:

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

Все остальные модули строятся вокруг этих объектов.
