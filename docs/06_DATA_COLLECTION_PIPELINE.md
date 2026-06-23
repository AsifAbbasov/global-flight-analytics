# DOCUMENT 06

# DATA COLLECTION PIPELINE

# Global Flight Analytics

Version: 1.1

Status: Approved

---

# 1. Purpose

Документ описывает процесс получения, обработки, обогащения, хранения и предоставления авиационных данных в системе Global Flight Analytics.

Документ определяет:

- поток данных;
- этапы обработки;
- правила хранения;
- механизм обогащения данных;
- механизм генерации аналитики;
- обработку ошибок;
- требования масштабируемости.

---

# 2. Pipeline Overview

Высокоуровневый поток данных:

```text
OpenSky Network

        ↓

Validation Layer

        ↓

Flight Matching Layer

        ↓

Data Enrichment Layer

        ↓

In-Memory Store

        ↓

Snapshot Generator

        ↓

PostgreSQL

        ↓

Backend API

        ↓

Next.js
```

---

# 3. Collection Principles

## Principle 1

Хранить только полезные данные.

---

## Principle 2

Не превращать PostgreSQL в бесконечное хранилище телеметрии.

---

## Principle 3

Историческая аналитика должна строиться на агрегатах.

---

## Principle 4

Live-данные должны обслуживаться преимущественно из памяти.

---

## Principle 5

Каждая запись должна иметь источник происхождения данных.

---

# 4. Live Collection Layer

Источник:

OpenSky Network

---

Backend выполняет запросы к OpenSky.

Интервал обновления:

- минимум 15 секунд;
- максимум 30 секунд.

---

Получаем:

- ICAO24;
- Callsign;
- Latitude;
- Longitude;
- Velocity;
- Heading;
- Barometric Altitude;
- Geometric Altitude;
- Vertical Rate;
- On Ground Status;
- Observation Timestamp.

---

# 5. Validation Layer

После получения данные проходят обязательную валидацию.

---

Проверяется:

- корректность координат;
- корректность высоты;
- корректность скорости;
- корректность курса;
- наличие ICAO24;
- корректность временной метки.

---

Некорректные записи:

- отбрасываются;
- фиксируются в логах.

---

# 6. Flight Matching Layer

Цель:

Связать наблюдение с доменной моделью системы.

---

Поток:

```text
ICAO24

↓

Aircraft

↓

Flight

↓

Flight State
```

---

Результат:

Каждое наблюдение получает связь с:

- aircraft;
- flight;
- airline;
- aircraft_model.

---

# 7. Data Enrichment Layer

После сопоставления выполняется обогащение.

---

Используются:

- aircraft;
- aircraft_models;
- airlines;
- airports.

---

Результат:

Формируется расширенное представление самолета.

---

Поток:

```text
Flight State

↓

Aircraft

↓

Aircraft Model

↓

Airline

↓

Enriched Aircraft View
```

---

# 8. In-Memory Store

Для обслуживания live-карты используется оперативная память.

---

Implementation:

```text
sync.Map
```

---

Причины выбора:

- высокая скорость чтения;
- отсутствие внешних зависимостей;
- отсутствие Redis в MVP;
- соответствие принципу бесплатной инфраструктуры.

---

Назначение:

- отображение live-карты;
- быстрый поиск самолетов;
- формирование API-ответов.

---

# 9. Flight State Storage

Состояния полетов сохраняются в таблицу:

```text
flight_states
```

---

Хранятся:

- координаты;
- скорость;
- высота;
- курс;
- время наблюдения.

---

Retention Policy:

Минимум:

```text
24 часа
```

---

Целевой срок MVP:

```text
7 дней
```

---

После истечения срока хранения данные удаляются.

---

# 10. Snapshot Generator

Каждые несколько минут система создает агрегированный снимок региона.

---

Источник:

In-Memory Store

---

Результат:

```text
traffic_snapshots
```

---

Содержимое:

- количество самолетов;
- количество маршрутов;
- активные аэропорты;
- интенсивность движения;
- распределение по регионам.

---

# 11. Airport Statistics Pipeline

Поток:

```text
Flight

↓

Airport Detection

↓

Airport Statistics Update
```

---

Обновляется:

```text
airport_statistics
```

---

Метрики:

- arrivals;
- departures;
- total_flights.

---

# 12. Route Statistics Pipeline

Поток:

```text
Flight

↓

Route Prediction

↓

Route Statistics Update
```

---

Обновляется:

```text
route_statistics
```

---

Метрики:

- flight_count;
- route_activity.

---

# 13. Historical Replay Pipeline

Для режима воспроизведения используются:

- flight_states;
- traffic_snapshots.

---

Цель:

Визуализация недавней истории движения.

---

MVP не хранит многомесячную телеметрию.

---

История ограничена retention policy.

---

# 14. Ingestion Runs

Каждый запуск загрузки данных создает запись:

```text
ingestion_runs
```

---

Фиксируются:

- source_name;
- started_at;
- finished_at;
- status;
- records_received;
- records_inserted;
- records_updated;
- error_message.

---

Назначение:

- мониторинг;
- диагностика;
- аудит загрузки данных.

---

# 15. Failure Handling

Если OpenSky недоступен:

используются последние доступные данные из памяти.

---

На фронтенде отображается статус:

```text
Live Data Delayed
```

---

Система не должна аварийно завершать работу.

---

# 16. Data Quality Rules

Система обязана:

- фильтровать некорректные координаты;
- фильтровать некорректные высоты;
- фильтровать некорректные скорости;
- предотвращать дублирование наблюдений;
- фиксировать ошибки загрузки.

---

# 17. MVP Capacity Targets

Целевые показатели MVP:

- до 10 000 одновременно наблюдаемых самолетов;
- до 100 API-запросов в секунду;
- один региональный снимок каждые 5 минут.

---

Без:

- Redis;
- Kafka;
- RabbitMQ;
- Kubernetes;
- микросервисов.

---

# 18. Scalability Strategy

На этапе MVP используется:

```text
Go API

↓

sync.Map

↓

PostgreSQL
```

---

Усложнение архитектуры допускается только после появления подтвержденной нагрузки.

---

# 19. Summary

Конвейер построен вокруг принципа:

```text
Получить данные

↓

Проверить данные

↓

Связать с доменной моделью

↓

Обогатить данные

↓

Сохранить полезную информацию

↓

Сформировать агрегаты

↓

Отобразить пользователю
```

Данный подход позволяет обслуживать MVP на бесплатной инфраструктуре без Redis, очередей сообщений и микросервисной архитектуры.
