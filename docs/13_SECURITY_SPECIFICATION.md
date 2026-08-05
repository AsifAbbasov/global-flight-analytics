# DOCUMENT 13

# SECURITY SPECIFICATION

# Global Flight Analytics

Version: 1.2

Status: Approved

---

# 1. Purpose

This document defines the security requirements of the Global Flight Analytics platform.

The document establishes protection rules for:

- infrastructure;
- Backend API;
- database;
- external integrations;
- user requests.

---

# 2. Security Goals

Primary goals:

- protect infrastructure;
- protect the API;
- protect the database;
- protect user requests;
- protect system secrets;
- ensure stable platform operation.

---

# 3. Security Scope

The platform does not process:

- banking data;
- payment-card data;
- government systems;
- military systems;
- air traffic control systems;
- personal user data.

---

# 4. Security Principles

Primary principles:

- Least Privilege;
- Defense In Depth;
- Secure By Default;
- Validation First;
- Backend Only Access.

---

# 5. HTTPS

All connections must use HTTPS.

---

HTTP is prohibited.

---

Only secure TLS connections are allowed.

---

# 6. Secrets Management

Secrets must not be stored:

- in GitHub;
- in source code;
- in client-side code;
- in logs.

---

Secrets may be stored only:

- in environment variables;
- in hosting-platform settings.

---

Examples:

```text
DATABASE_URL

OPENSKY_USERNAME

OPENSKY_PASSWORD
```

---

# 7. Database Security

The database is accessible only by the Backend.

---

The Frontend has no direct access.

---

Access is provided only through:

```text
/api/v1/*
```

---

# 8. Input Validation

The system validates:

- ICAO codes;
- IATA codes;
- coordinates;
- route identifiers;
- filter parameters;
- search parameters.

---

Invalid data is rejected.

---

# 9. Rate Limiting

Request-rate limiting is mandatory.

---

MVP:

```text
100 requests per minute
per IP address
```

---

When the limit is exceeded, the system returns:

```text
429 Too Many Requests
```

---

# 10. Logging

The system logs:

- errors;
- exceptions;
- unavailable data sources;
- critical security events.

---

The system does not log:

- passwords;
- secrets;
- tokens;
- connection strings.

---

# 11. API Security

All parameters pass through:

- validation;
- normalization;
- type checking.

---

The following are prohibited:

- arbitrary command execution;
- unsafe serialization;
- user-code execution.

---

# 12. SQL Security

The system uses:

- prepared statements;
- parameterized queries.

---

The following are prohibited:

- SQL-string concatenation;
- dynamic SQL generation from user input.

---

SQL injection must be impossible.

---

# 13. Cross-Origin Policy

Only trusted domains are allowed.

---

Example:

```text
https://global-flight-analytics.vercel.app
```

---

All other domains are blocked.

---

# 14. Security Headers

The Backend must use the following HTTP security headers:

- X-Content-Type-Options;
- X-Frame-Options;
- Referrer-Policy;
- Content-Security-Policy.

---

Purpose:

- protect against Clickjacking;
- protect against MIME Sniffing;
- control content sources;
- limit data leakage.

---

# 15. Dependency Security

All dependencies:

- are updated regularly;
- are checked for known vulnerabilities;
- pass an audit before release.

---

The platform uses:

```text
npm audit

govulncheck
```

---

# 16. Supply Chain Security

All dependencies must be pinned through lock files.

---

The project uses:

```text
pnpm-lock.yaml

go.sum
```

---

Removing lock files from the repository is prohibited.

---

# 17. Security Events

Security events include:

- rate-limit violations;
- bulk requests;
- suspicious activity;
- large error volumes;
- attempts to access nonexistent resources.

---

# 18. External Source Security

All external sources are treated as untrusted.

---

Before processing, data is:

- validated;
- checked for correctness;
- normalized.

---

Sources:

- OpenSky;
- OurAirports;
- OpenStreetMap;
- Wikipedia;
- Wikidata.

---

# 19. Infrastructure Security

Mandatory requirements:

- HTTPS Only;
- Private Database Access;
- Environment Variables Only;
- Health Monitoring;
- Daily Backups.

---

# 20. Security Review

Every change to the following components must be checked against this document:

- API;
- infrastructure;
- dependencies;
- data-storage mechanisms;
- external integrations.

---

# 21. Security Boundaries

The platform does not guarantee:

- accuracy of external sources;
- OpenSky availability;
- uninterrupted operation of third-party services.

---

Security coverage applies only to platform infrastructure and platform data.

---

# 22. Summary

Platform security is built around:

- least privilege;
- component isolation;
- secret protection;
- input validation;
- workload control;
- secure operation with external sources;
- reproducible builds;
- regular dependency audits.

All platform components must comply with this document.
