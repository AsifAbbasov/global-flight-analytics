# Document 218 — Live Vertical Rate and Motion

Status: R3 engineering implementation complete; canonical stacked-PR validation pending.

## Purpose

R3 surfaces the vertical-rate evidence that GFA already persists for flight states and turns it into a bounded live presentation for the Aircraft Detail experience. The feature does not add a new provider, polling path, database table, paid service or background worker.

## Evidence source

The canonical stored value is `flight_states.vertical_rate_mps`. Existing provider adapters already normalize usable vertical-rate telemetry into metres per second before persistence:

- OpenSky supplies vertical rate directly in metres per second when available.
- readsb-compatible providers normalize barometric rate from feet per minute into metres per second.

Current Traffic now exposes that persisted value as nullable `vertical_rate_mps`. Null means no usable vertical-rate evidence is available for that observation.

## Frontend interpretation

The frontend preserves the numeric value and applies a presentation-only deadband of ±0.5 m/s:

- greater than +0.5 m/s → `↑ Climbing`
- less than -0.5 m/s → `↓ Descending`
- within the inclusive ±0.5 m/s band → `Level`
- observed `on_ground=true` → `On ground` takes presentation precedence
- missing/non-finite evidence → `Unavailable`

The UI also shows an approximate feet-per-minute conversion for aviation readability while metres per second remains the canonical API evidence.

## Claim boundary

The movement label is not a flight-phase detector and does not claim pilot intent, cleared altitude, autopilot mode, climb/descent target, turbulence state or future motion. A single observed vertical-rate sample can fluctuate and must not be interpreted as a complete trajectory trend.

## API contract

`CurrentTrafficItem.vertical_rate_mps` is required as a response property but nullable. This makes absence explicit rather than silently coercing missing evidence to zero.

## Persistence and cost

No migration is required because `vertical_rate_mps` already exists in the canonical flight-state schema. R3 only connects existing persisted evidence to Current Traffic and the frontend. Infrastructure cost remains zero.

## Validation boundary

Branch-local engineering validation covers Go tests/vet, Stage14 strict audit, OpenAPI contract and generated-client drift, dependency policy checks, API client checks, frontend lint/typecheck/tests and production build. Canonical CI, Vercel, review and merge evidence remain separate stacked-PR gates.
