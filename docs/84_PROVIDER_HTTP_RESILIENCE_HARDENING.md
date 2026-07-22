# Document 84 — Provider HTTP Resilience Hardening

Status: Implemented
Project: Global Flight Analytics
Scope: external provider error preservation, response body limits, fallback classification, and non-destructive observation

## 1. Problem

Provider response observation previously participated in the primary data-plane result. For Airplanes.live and Open-Meteo, an observation failure could replace a real HTTP 429 or 500 error. OpenSky could reject an otherwise valid response when its observer failed. This made automatic fallback dependent on an auxiliary telemetry path.

External JSON and CSV response bodies were also decoded without an explicit byte limit. HTTP timeouts bound duration but do not bound decoded response size.

## 2. Error preservation contract

The provider HTTP status remains the primary error classification.

For failed HTTP responses:

```text
provider status error + observer error = errors.Join
```

This preserves `errors.Is` and `errors.As` matching for rate limiting, provider server failure, authorization failure, and other typed provider outcomes.

For successful responses, observation is best-effort. An auxiliary health or budget evidence failure does not discard a valid provider payload.

Transport and response parsing failures remain visible as primary failures. Observation failures may be joined to those failures but cannot replace them.

## 3. Response size contract

A shared integration helper now enforces response limits before decoding:

```text
Airplanes.live state response: 8 MiB
OpenSky state response: 16 MiB
Open-Meteo weather response: 1 MiB
OurAirports CSV response: 32 MiB
```

The helper rejects:

- a declared `Content-Length` above the limit;
- a streamed response that exceeds the limit even when `Content-Length` is absent or incorrect.

The canonical typed error is:

```text
integrations/common.ErrProviderResponseTooLarge
```

No partial provider payload is published after the limit is exceeded.

## 4. Fallback guarantee

Automatic traffic fallback continues to recognize the original provider failure when an observer failure is joined to it. A primary Airplanes.live server failure still permits OpenSky fallback.

## 5. Verification

The permanent regression coverage includes:

```text
exact-limit response acceptance
declared oversized response rejection
streamed oversized response rejection
Airplanes.live server classification with observer failure
OpenSky server classification with observer failure
OpenSky successful payload with observer failure
Open-Meteo server classification with observer failure
oversized response wiring for all four provider adapters
fallback after a joined provider and observer failure
targeted race tests
full backend tests
Go static analysis
git diff validation
```

## 6. Completion boundary

This document closes the provider observer error replacement and unbounded response body findings. It does not close daemon retry scheduling, publication reservation lifecycle, or full per-attempt fallback evidence. Those remain separate orchestration contracts.
