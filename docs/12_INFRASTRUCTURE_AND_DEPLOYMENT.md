# DOCUMENT 12

# INFRASTRUCTURE AND DEPLOYMENT

# Global Flight Analytics

Version: 1.1

Status: Approved

---

# 1. Purpose

Документ определяет инфраструктуру, окружения, развертывание и эксплуатацию платформы Global Flight Analytics.

Цель:

- обеспечить запуск MVP одним разработчиком;
- минимизировать расходы;
- сохранить возможность масштабирования без полной переработки архитектуры.

---

# 2. Infrastructure Principles

## Free First

На этапе MVP используются бесплатные или условно-бесплатные сервисы.

---

## Simplicity First

Не используются:

- Kubernetes;
- Redis;
- Kafka;
- RabbitMQ;
- микросервисы.

---

## Monolith First

Система развертывается как модульный монолит.

---

## Open Data First

Используются только открытые источники данных.

---

# 3. Frontend Hosting

Platform:

```text
Vercel
```

---

Purpose:

- развертывание Next.js;
- глобальная доставка статических ресурсов;
- автоматический деплой через GitHub.

---

# 4. Backend Hosting

Platform:

```text
Render
```

---

Purpose:

- размещение Go API;
- выполнение фоновых задач;
- работа Data Collection Pipeline.

---

Причины выбора:

- стабильный бесплатный тариф;
- простая интеграция с GitHub;
- поддержка Go;
- быстрый запуск MVP.

---

# 5. Database

Platform:

```text
Neon PostgreSQL
```

---

Purpose:

- справочники;
- аэропорты;
- самолеты;
- маршруты;
- статистика;
- аналитика.

---

# 6. Live Data Layer

Location:

```text
Go Application Memory
```

---

Implementation:

```text
sync.Map
```

---

Назначение:

- хранение текущих состояний самолетов;
- работа live-карты;
- быстрое формирование ответов API.

---

# 7. Source Layer

Основные источники:

- OpenSky Network;
- OurAirports;
- OpenStreetMap;
- Wikipedia;
- Wikidata.

---

# 8. Deployment Architecture

Высокоуровневая схема:

```text
GitHub Repository

       ├── Vercel
       │      │
       │      ▼
       │   Frontend
       │
       ├── Render
       │      │
       │      ▼
       │   Go Backend
       │
       └── Neon
              │
              ▼
         PostgreSQL
```

---

# 9. Deployment Flow

Поток публикации:

```text
Developer

↓

GitHub

↓

Automatic Deployment

↓

Vercel Frontend

↓

Render Backend

↓

Neon Database
```

---

# 10. Environment Variables

## Backend

```text
DATABASE_URL

OPENSKY_USERNAME

OPENSKY_PASSWORD

APP_ENV

API_PORT
```

---

## Frontend

```text
NEXT_PUBLIC_API_URL
```

---

Секреты никогда не должны попадать в репозиторий.

---

# 11. Monitoring

Минимальный набор мониторинга MVP:

- Application Logs;
- Health Endpoint;
- Error Logs;
- Database Logs.

---

Дополнительно:

- Uptime Monitoring;
- Health Check Monitoring;
- Deployment Monitoring.

---

# 12. Backup Strategy

Источник:

```text
Neon PostgreSQL
```

---

Стратегия:

```text
Daily Backup
```

---

Retention:

```text
7 days
```

---

Назначение:

- защита от потери данных;
- восстановление после ошибок;
- тестирование миграций.

---

# 13. Security Rules

Обязательные правила:

- HTTPS Only;
- No Direct Database Access;
- No Client Secrets;
- Environment Variables Only;
- Backend API Only.

---

Дополнительно:

- CORS Restrictions;
- Rate Limiting;
- Request Validation.

---

# 14. Failure Strategy

## OpenSky Unavailable

Действие:

```text
Использовать последние доступные данные из памяти
```

---

Frontend отображает:

```text
Live Data Delayed
```

---

## Database Unavailable

Действие:

```text
Read Only Mode
```

---

## Backend Restart

Действие:

```text
Повторная загрузка данных
из OpenSky
```

---

# 15. Scaling Strategy

## Phase 1

Регион:

```text
Кавказ
```

---

## Phase 2

Регион:

```text
Европа
```

---

## Phase 3

Регион:

```text
Global
```

---

Масштабирование выполняется без изменения архитектурных принципов.

---

# 16. Cost Target

Целевой бюджет MVP:

```text
0 – 5 USD в месяц
```

---

Архитектура должна сохранять работоспособность даже при изменении бесплатных тарифов провайдеров.

---

# 17. Infrastructure Boundaries

В MVP не используются:

- Kubernetes;
- Redis;
- Kafka;
- RabbitMQ;
- Elasticsearch;
- отдельные аналитические кластеры.

---

Усложнение инфраструктуры допускается только после появления подтвержденной нагрузки.

---

# 18. Summary

Инфраструктура платформы построена вокруг следующих компонентов:

- Vercel;
- Render;
- Neon PostgreSQL;
- OpenSky Network;
- OpenStreetMap;
- OurAirports;
- Wikidata.

Архитектура обеспечивает запуск и поддержку MVP одним разработчиком с минимальными затратами и возможностью дальнейшего масштабирования.
