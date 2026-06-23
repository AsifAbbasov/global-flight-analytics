# DOCUMENT 10

# API SPECIFICATION

# Global Flight Analytics

Version: 1.1

Status: Approved

---

# 1. Purpose

Документ определяет публичные и внутренние интерфейсы платформы Global Flight Analytics.

Документ является контрактом между:

- Frontend;
- Backend;
- Database;
- Data Collection Pipeline.

---

# 2. API Principles

Используется:

- REST API;
- JSON;
- HTTPS.

---

Все ответы возвращаются в формате UTF-8.

---

# 3. API Versioning

Текущая версия:

```text
/api/v1
```

---

Все будущие изменения должны добавляться через новую версию API.

---

# 4. Response Envelope

Все успешные ответы используют единый формат.

---

## Success Response

```json
{
  "success": true,
  "data": {}
}
```

---

## Success Response With Metadata

```json
{
  "success": true,
  "data": [],
  "meta": {}
}
```

---

# 5. Error Contract

Все ошибки используют единый формат.

---

## Error Response

```json
{
  "success": false,
  "error": {
    "code": "AIRPORT_NOT_FOUND",
    "message": "Airport not found"
  }
}
```

---

# 6. HTTP Status Codes

Используются:

```text
200 OK

400 Bad Request

404 Not Found

429 Too Many Requests

500 Internal Server Error

503 Service Unavailable
```

---

# 7. Pagination Contract

Все списочные эндпоинты поддерживают пагинацию.

---

Параметры:

```text
page
limit
```

---

Пример:

```text
/api/v1/airports?page=1&limit=50
```

---

Ответ:

```json
{
  "success": true,
  "data": [],
  "meta": {
    "page": 1,
    "limit": 50,
    "total": 1000
  }
}
```

---

# 8. Flights API

## Get Live Flights

```http
GET /api/v1/flights/live
```

---

Query Parameters:

```text
region
limit
```

---

Response:

- ICAO24;
- callsign;
- latitude;
- longitude;
- altitude;
- velocity;
- heading.

---

## Get Flight

```http
GET /api/v1/flights/{flightId}
```

---

Response:

- flight;
- aircraft;
- current_state;
- route_prediction.

---

# 9. Aircraft API

## Get Aircraft

```http
GET /api/v1/aircraft/{icao24}
```

---

Response:

- registration;
- model;
- manufacturer;
- airline;
- country.

---

# 10. Airports API

## Get Airports

```http
GET /api/v1/airports
```

---

Parameters:

```text
page
limit
country
```

---

## Get Airport

```http
GET /api/v1/airports/{icao}
```

---

Response:

- profile;
- runways;
- facilities;
- statistics;
- activity_score.

---

# 11. Routes API

## Get Routes

```http
GET /api/v1/routes
```

---

## Get Route

```http
GET /api/v1/routes/{id}
```

---

Response:

- origin;
- destination;
- confidence_level;
- confidence_score;
- statistics.

---

# 12. Analytics API

## Regional Analytics

```http
GET /api/v1/analytics/regions
```

---

## Airport Analytics

```http
GET /api/v1/analytics/airports
```

---

## Route Analytics

```http
GET /api/v1/analytics/routes
```

---

## Heat Map Analytics

```http
GET /api/v1/analytics/heatmap
```

---

Parameters:

```text
region
```

---

# 13. Search API

## Global Search

```http
GET /api/v1/search
```

---

Parameters:

```text
query
```

---

Поиск выполняется по:

- аэропортам;
- самолетам;
- авиакомпаниям.

---

# 14. Replay API

## Historical Replay

```http
GET /api/v1/replay
```

---

Parameters:

```text
date
region
```

---

Response:

- flight_states;
- traffic_snapshot.

---

# 15. Health API

## Health Check

```http
GET /api/v1/health
```

---

Response:

```json
{
  "success": true,
  "data": {
    "status": "ok"
  }
}
```

---

# 16. Version API

## Get Version

```http
GET /api/v1/version
```

---

Response:

```json
{
  "success": true,
  "data": {
    "version": "1.0.0"
  }
}
```

---

# 17. Internal Ingestion API

Используется только внутри Backend.

Не доступен Frontend.

---

## Trigger OpenSky Sync

```http
POST /internal/opensky/sync
```

---

## Trigger Airport Sync

```http
POST /internal/airports/sync
```

---

## Trigger Statistics Generation

```http
POST /internal/statistics/generate
```

---

# 18. Rate Limiting

MVP ограничения:

```text
60 запросов в минуту
на IP адрес
```

---

# 19. Performance Targets

Live Flights:

```text
< 300 ms
```

---

Airport Profile:

```text
< 500 ms
```

---

Analytics:

```text
< 1000 ms
```

---

# 20. Summary

API предоставляет единый интерфейс для:

- самолетов;
- аэропортов;
- маршрутов;
- аналитики;
- исторического воспроизведения;
- поиска;
- мониторинга состояния системы.

Данный документ является официальным контрактом между Frontend и Backend платформы Global Flight Analytics.
