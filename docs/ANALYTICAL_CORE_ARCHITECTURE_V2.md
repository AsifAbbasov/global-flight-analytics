# Analytical Core Architecture V2

## Цель

Построить аналитическое ядро, которое позволяет добавлять новые метрики без изменения существующей архитектуры.

---

## Принципы

- Один пакет — одна ответственность.
- Минимальная связанность.
- Максимальная расширяемость.
- Домен не зависит от HTTP.
- Домен не зависит от PostgreSQL.
- Домен не зависит от внешних поставщиков данных.
- Все вычисления детерминированы.

---

## Основные сущности

### Metric

Единица аналитики.

Примеры:

- Active Aircraft
- Traffic Density
- Airport Activity
- Arrivals Proxy
- Departures Proxy
- Coverage Score
- Data Freshness

---

### Snapshot

Полное вычисленное состояние аналитики в определённый момент времени.

---

### Time Window

Интервал вычисления.

Примеры:

- Now
- 5 minutes
- 15 minutes
- 1 hour
- 24 hours

---

### Calculator

Компонент, вычисляющий одну конкретную метрику.

---

### Aggregation

Объединяет результаты Calculator в Snapshot.

---

### Projection

Модель данных для Frontend.

---

### Query

Описание требуемых данных без вычислений.

---

## Зависимости

Frontend

↓

HTTP

↓

Application

↓

Analytical Core

↓

Repository

↓

Database

Обратные зависимости запрещены.

---

## Правила

Calculator:

- не знает HTTP;
- не знает PostgreSQL;
- не знает JSON;
- получает данные;
- возвращает результат.

Repository:

- только получает данные;
- ничего не вычисляет.

Projection:

- не содержит бизнес-логики.

---

## Добавление новой метрики

Допускается изменение только:

- нового Calculator;
- регистрации Calculator;
- тестов.

Существующие Calculator изменяться не должны.

---

## План разработки

### Этап 1

Analytical Core Foundation

### Этап 2

Traffic Metrics

- Active Aircraft
- Traffic Density
- Airport Activity
- Data Freshness
- Coverage Score

### Этап 3

Airport Intelligence

### Этап 4

Route Intelligence

### Этап 5

Airspace Intelligence

### Этап 6

Frontend Integration
