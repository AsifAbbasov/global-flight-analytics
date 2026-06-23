# DOCUMENT 18

# TECHNICAL DECISIONS RECORD

# Global Flight Analytics

Version: 1.1

Status: Approved

---

# 1. Purpose

Документ фиксирует ключевые технические решения проекта.

Документ отвечает на вопросы:

- какое решение принято;
- почему оно принято;
- какие альтернативы рассматривались;
- какие последствия имеет решение.

---

# 2. Frontend Framework

## Decision

Использовать Next.js.

---

## Status

Accepted

---

## Alternatives

- React + Vite
- Remix
- Nuxt

---

## Why

- зрелость экосистемы;
- отличная поддержка TypeScript;
- удобный деплой;
- хорошая производительность;
- возможность дальнейшего использования Server Components.

---

## Consequences

- привязка к экосистеме Next.js;
- упрощение развертывания на Vercel.

---

# 3. Frontend Language

## Decision

Использовать TypeScript.

---

## Status

Accepted

---

## Alternatives

- JavaScript

---

## Why

- типизация;
- безопасность изменений;
- масштабируемость;
- удобство рефакторинга.

---

## Consequences

- небольшое увеличение сложности разработки;
- значительное снижение количества ошибок.

---

# 4. Backend Language

## Decision

Использовать Go.

---

## Status

Accepted

---

## Alternatives

- Python
- Node.js
- Java

---

## Why

- высокая производительность;
- низкое потребление памяти;
- простое развертывание;
- хорошая работа с сетевыми сервисами.

---

## Consequences

- более ограниченная экосистема по сравнению с JavaScript;
- высокий выигрыш по производительности.

---

# 5. Backend Framework

## Decision

Использовать Fiber.

---

## Status

Accepted

---

## Alternatives

- Gin
- Echo
- Chi

---

## Why

- высокая производительность;
- простота;
- низкие накладные расходы.

---

## Consequences

- меньшая экосистема по сравнению с Gin;
- высокая скорость разработки.

---

# 6. Database

## Decision

Использовать PostgreSQL.

---

## Status

Accepted

---

## Alternatives

- MySQL
- MariaDB
- MongoDB

---

## Why

- зрелость;
- надежность;
- поддержка аналитических запросов;
- совместимость с Neon.

---

## Consequences

- необходимость проектирования схемы данных;
- высокая надежность хранения.

---

# 7. Database Driver

## Decision

Использовать pgx.

---

## Status

Accepted

---

## Alternatives

- database/sql
- GORM

---

## Why

- высокая производительность;
- полный контроль над SQL;
- минимальные накладные расходы.

---

# 8. Mapping Engine

## Decision

Использовать MapLibre.

---

## Status

Accepted

---

## Alternatives

- Google Maps
- Mapbox
- Leaflet

---

## Why

- открытый исходный код;
- отсутствие лицензионных ограничений;
- высокая производительность через WebGL.

---

## Consequences

- больше контроля над картографией;
- отсутствие зависимости от коммерческого поставщика.

---

# 9. Live Aviation Data

## Decision

Использовать OpenSky Network.

---

## Status

Accepted

---

## Alternatives

- FlightRadar24
- FlightAware
- ADS-B Exchange

---

## Why

- бесплатный доступ;
- открытые данные;
- большое сообщество.

---

## Consequences

- ограничения покрытия;
- зависимость от доступности OpenSky.

---

# 10. Runtime Cache

## Decision

Использовать память процесса Go.

---

## Status

Accepted

---

## Alternatives

- Redis

---

## Why

- минимальная инфраструктура;
- отсутствие дополнительных расходов;
- достаточная производительность для MVP.

---

## Consequences

- данные исчезают после перезапуска;
- для MVP это допустимо.

---

# 11. Service Architecture

## Decision

Использовать модульный монолит.

---

## Status

Accepted

---

## Alternatives

- микросервисы

---

## Why

- проект разрабатывается одним разработчиком;
- простота сопровождения;
- минимальные расходы.

---

## Consequences

- более простой деплой;
- возможность выделения сервисов в будущем.

---

# 12. Analytics Storage

## Decision

Не использовать ClickHouse в MVP.

---

## Status

Accepted

---

## Alternatives

- ClickHouse

---

## Why

- преждевременное усложнение;
- дополнительные расходы;
- недостаточный объем данных.

---

## Consequences

- PostgreSQL остается единственной базой данных.

---

# 13. Container Orchestration

## Decision

Не использовать Kubernetes.

---

## Status

Accepted

---

## Alternatives

- Kubernetes

---

## Why

- избыточность для MVP;
- высокая сложность сопровождения.

---

## Consequences

- упрощение инфраструктуры;
- снижение стоимости поддержки.

---

# 14. Backend Computation

## Decision

Не использовать Python в MVP.

---

## Status

Accepted

---

## Alternatives

- отдельный Python-сервис

---

## Why

- уменьшение сложности;
- уменьшение потребления памяти;
- единый Backend.

---

## Consequences

- все вычисления реализуются на Go;
- Python может быть добавлен после появления задач машинного обучения.

---

# 15. Geographic Strategy

## Decision

Запуск с регионального покрытия.

---

## Status

Accepted

---

## Why

- меньше нагрузка;
- быстрее запуск;
- проще проверка гипотез.

---

## Consequences

- ограниченная география на старте;
- возможность постепенного масштабирования.

---

# 16. Future Technical Decisions

После завершения MVP допускается пересмотр следующих решений:

- Redis;
- ClickHouse;
- Python Services;
- Machine Learning Infrastructure;
- Object Storage;
- Mobile Applications.

---

# 17. Final Technical Position

Проект строится на принципах:

- простота;
- надежность;
- минимальные расходы;
- масштабируемость;
- инженерная обоснованность решений.

Каждое новое техническое решение должно иметь документированное обоснование и измеримую пользу для продукта.
