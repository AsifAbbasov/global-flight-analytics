# Document 27 — Engineering Principles

Status: Architecture Baseline v1.1  
Project: Global Flight Analytics  
Scope: Engineering principles for readable, simple, predictable, testable, and explainable implementation

---

## 1. Purpose

This document defines the engineering principles that must guide implementation in Global Flight Analytics.

The project must be technically strong, but it must not become complex for the sake of complexity.

The goal is not to look enterprise-grade.

The goal is to build a readable, simple, predictable, testable, and explainable aviation analytics platform.

---

## 2. Main Engineering Rule

```text
Do not add complexity for its own sake.
```

Complexity is allowed only when it clearly improves at least one of the following:

```text
1. readability
2. duplication removal
3. error risk reduction
4. testability
5. documentation alignment
```

If a new abstraction, package, interface, service, policy, or helper does not improve one of these, it must not be added.

---

## 3. Quality Order

Implementation quality must be evaluated in this order:

```text
1. readable
2. simple
3. predictable
4. testable
5. explainable
```

This order matters.

The project must not sacrifice readability just to look architecturally advanced.

The project must not introduce indirection unless that indirection makes the code easier to understand, safer to change, or easier to test.

---

## 4. Simple First Rule

Start with the simplest implementation that is correct, readable, and testable.

Do not start with:

```text
generic frameworks
unnecessary interfaces
abstract factories
deep inheritance-like composition
configuration for values that are not yet variable
premature optimization
premature distributed architecture
```

Prefer:

```text
plain functions
small structs
clear package boundaries
explicit data flow
table-driven tests
small repositories
small services
clear domain models
```

---

## 5. Abstraction Gate

Before adding a new abstraction, answer these questions:

```text
1. Does it make the code easier to read?
2. Does it remove real duplication?
3. Does it reduce the risk of mistakes?
4. Does it make a rule easier to test independently?
5. Does it follow the documentation baseline?
```

If the answer is not clearly yes for at least one item, do not add the abstraction.

---

## 6. Magic Number Policy

The project must not use unexplained numbers in analytical logic.

A number is acceptable when it is one of these:

```text
a physical or geographical boundary
a protocol or provider format rule
a documented product rule
a named analytical threshold
a named scoring penalty
a local test fixture value
```

A number is not acceptable when its meaning must be guessed from context.

Bad example:

```text
score >= 0.85
```

Better example:

```text
score >= HighConfidenceMinimumScore
```

Bad example:

```text
speed > 420
```

Better example:

```text
speed > DefaultMaxGroundSpeedMetersPerSecond
```

Bad example:

```text
heading > 360
```

Better example:

```text
IsHeadingDegreesInclusive(heading)
```

---

## 7. Analytical Threshold Policy

Every analytical threshold must have:

```text
1. a clear name
2. one owner package
3. at least one boundary test
4. a known consumer
5. an explanation when it affects product output
```

Thresholds that affect confidence, data quality, trust gates, trajectory gaps, route confidence, or displayed analytics must not be hidden inside business logic.

They must live in a policy or constraints package.

---

## 8. Policy Layer Rule

Policy packages are allowed only when they make the project simpler to reason about.

Allowed policy packages:

```text
constraints
qualitypolicy
trajectorypolicy
source trust policy
route confidence policy
```

Forbidden policy packages:

```text
policy packages for one trivial local if statement
policy packages with no tests
policy packages that hide simple logic
policy packages that exist only to look advanced
```

A policy layer must make rules more visible, not more mysterious.

---

## 9. Data Trust Rule

No analytics may use raw provider data directly.

The required flow is:

```text
external provider response
↓
normalizer
↓
validator
↓
quality score
↓
trust gate
↓
repository
↓
service
↓
API response
```

If data is incomplete, suspicious, stale, or estimated, this must be reflected through:

```text
confidence
quality score
warnings
missing fields
source name
calculated_at
limitations
```

The system must not present weak assumptions as verified facts.

---

## 10. Unit Test Rule

Important logic must be covered by unit tests.

Required test types:

```text
boundary tests
table-driven tests
invalid input tests
normalization tests
scoring tests
confidence threshold tests
repository validation tests
handler response tests
```

Unit tests should be fast and deterministic.

Unit tests must not depend on live external APIs.

External API behavior should be tested through local test servers or mocked responses.

Real external API checks belong to smoke tests, not unit tests.

---

## 11. Runtime Smoke Test Rule

Runtime smoke tests are allowed for verifying full integration chains.

Example:

```text
Open-Meteo API
↓
Open-Meteo client
↓
Current weather snapshot
↓
Weather repository
↓
PostgreSQL
```

Smoke tests must be temporary unless they are converted into controlled integration tests.

Temporary smoke test files must not be committed.

---

## 12. Documentation Alignment Rule

The implementation must follow the documentation baseline.

When implementation changes the architecture, one of these must happen:

```text
1. update the relevant document
2. add a numbered document
3. add a clearly marked amendment
```

Code must not silently drift away from documentation.

Documentation must not describe architecture that the code is clearly not following.

---

## 13. No Overengineering Rule

The project must avoid:

```text
interfaces without multiple real implementations
configuration for values that are not actually configurable
services that only call one function without adding meaning
repositories that hide simple queries without value
deep package nesting without clear responsibility
premature caching
premature concurrency
premature machine learning
```

Engineering maturity means choosing the simplest correct design, not the most complicated design.

---

## 14. Accepted Complexity

Complexity is acceptable when it protects the project from real risk.

Accepted reasons:

```text
data quality protection
source isolation
provider response normalization
database migration safety
analytical threshold visibility
confidence explanation
testability
runtime verification
```

Rejected reasons:

```text
it looks senior
it looks enterprise
it might be useful someday
it makes the architecture diagram bigger
it copies patterns from large systems without current need
```

---

## 15. Engineering Review Checklist

Before committing new code, check:

```text
1. Is the code readable?
2. Is the simplest correct design used?
3. Are analytical thresholds named?
4. Are important rules tested?
5. Are provider-specific details isolated?
6. Are invalid inputs handled?
7. Are errors explicit?
8. Does the code follow the current documentation?
9. Does the change avoid unnecessary abstraction?
10. Does go test ./... pass?
```

A change that fails this checklist should be simplified or better explained before commit.

---

## 16. Final Principle

Global Flight Analytics must be built as a serious engineering project, but serious engineering does not mean unnecessary complexity.

The correct standard is:

```text
as simple as possible,
as strict as necessary,
as explainable as the analytics require.
```
